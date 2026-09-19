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
  const closedByUsRef = useRef(false);

  useEffect(() => {
    if (!pollId) return;
    closedByUsRef.current = false;

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
        backoffRef.current = 1000;
        setStatus("live");
        resync();
      };

      sock.onmessage = (evt) => {
        try {
          applyEvent(JSON.parse(evt.data));
        } catch {
          // ignore malformed frames
        }
      };

      sock.onclose = () => {
        if (closedByUsRef.current) return;
        setStatus("reconnecting");
        const delay = backoffRef.current;
        backoffRef.current = Math.min(delay * 2, MAX_BACKOFF_MS);
        setTimeout(connect, delay);
      };

      sock.onerror = () => {
        sock.close();
      };
    }

    connect();

    return () => {
      closedByUsRef.current = true;
      sockRef.current?.close();
    };
  }, [pollId]);

  const updateResults = useCallback((newResults, newTotalVotes) => {
    if (newResults) setResults(newResults);
    if (typeof newTotalVotes === "number") setTotalVotes(newTotalVotes);
  }, []);

  return { results, totalVotes, activeWatchers, status, setResults, setTotalVotes, updateResults };
}

