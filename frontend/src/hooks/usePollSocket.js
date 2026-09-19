import { useEffect, useRef, useState, useCallback } from "react";
import { WS_URL, pollApi } from "../services/api";

const MAX_BACKOFF_MS = 15_000;

/**
 * Opens a live WebSocket connection to a poll's results feed.
 *
 * - Seeds `results`/`totalVotes` from the server's initial snapshot
 *   message, then applies every subsequent "poll.results.updated" event.
 * - Tracks connection status ("connecting" | "live" | "reconnecting")
 *   for the UI's live indicator.
 * - On disconnect, reconnects automatically with exponential backoff,
 *   and re-fetches the latest results via the REST API immediately on
 *   reconnect so the client never stays stale after a network blip
 *   (spec section 39).
 */
export function usePollSocket(pollId) {
  const [results, setResults] = useState({});
  const [totalVotes, setTotalVotes] = useState(0);
  const [activeWatchers, setActiveWatchers] = useState(1);
  const [status, setStatus] = useState("connecting");

  const backoffRef = useRef(1000);
  const sockRef = useRef(null);

  useEffect(() => {
    if (!pollId) return;
    let cancelled = false;

    async function resync() {
      try {
        const poll = await pollApi.get(pollId);
        setResults(poll.results || {});
        setTotalVotes(poll.totalVotes || 0);
      } catch {
        // best-effort
      }
    }

    resync();

    function applyEvent(evt) {
      if (evt.type === "poll.results.snapshot") {
        setResults(evt.results || {});
        setTotalVotes(evt.totalVotes || 0);
        if (typeof evt.activeWatchers === "number") {
          setActiveWatchers(evt.activeWatchers);
        }
      } else if (evt.type === "poll.results.updated") {
        setResults(evt.results || {});
        setTotalVotes(evt.totalVotes || 0);
      } else if (evt.type === "poll.watchers.updated") {
        if (typeof evt.activeWatchers === "number") {
          setActiveWatchers(evt.activeWatchers);
        }
      }
    }

    function connect() {
      setStatus((s) => (s === "connecting" ? "connecting" : "reconnecting"));
      const wsUrl = `${WS_URL}/api/polls/${pollId}/ws`;
      const sock = new WebSocket(wsUrl);
      sockRef.current = sock;

      sock.onopen = () => {
        if (cancelled || sock !== sockRef.current) return;
        backoffRef.current = 1000;
        setStatus("live");
        resync();
      };

      sock.onmessage = (evt) => {
        if (sock !== sockRef.current) return;
        try {
          applyEvent(JSON.parse(evt.data));
        } catch {
          // ignore malformed frames
        }
      };

      sock.onclose = () => {
        // Ignore close events from a socket that's no longer the active
        // one for this hook instance (superseded by a newer connect()
        // call, e.g. during React's dev-mode double-effect cycle, or a
        // fast pollId change). Comparing identity here — instead of a
        // shared "did we close this on purpose" boolean — is what makes
        // this correct even when this effect's cleanup and a fresh
        // connect() interleave: a boolean flag gets reset by the new
        // connection before the old socket's async close event arrives,
        // which previously caused a spurious extra reconnect (and an
        // inflated "watching" count server-side, since the hub still
        // thought that phantom client was connected).
        if (cancelled || sock !== sockRef.current) return;
        setStatus("reconnecting");
        const delay = backoffRef.current;
        backoffRef.current = Math.min(delay * 2, MAX_BACKOFF_MS);
        setTimeout(() => {
          if (!cancelled) connect();
        }, delay);
      };

      sock.onerror = () => {
        sock.close();
      };
    }

    connect();

    return () => {
      cancelled = true;
      sockRef.current?.close();
      sockRef.current = null;
    };
  }, [pollId]);

  const updateResults = useCallback((newResults, newTotalVotes) => {
    if (newResults) setResults(newResults);
    if (typeof newTotalVotes === "number") setTotalVotes(newTotalVotes);
  }, []);

  return { results, totalVotes, activeWatchers, status, setResults, setTotalVotes, updateResults };
}

