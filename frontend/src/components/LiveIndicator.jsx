export default function LiveIndicator({ status, activeWatchers }) {
  const isLive = status === "live";
  return (
    <span className={`live-indicator ${isLive ? "is-live" : "is-reconnecting"}`}>
      <span className="live-dot" aria-hidden="true">
        {isLive ? "●" : "○"}
      </span>
      {isLive ? (
        <span>
          LIVE {typeof activeWatchers === "number" && activeWatchers > 0 ? `• ${activeWatchers} watching` : ""}
        </span>
      ) : status === "connecting" ? (
        "Connecting…"
      ) : (
        "Reconnecting…"
      )}
    </span>
  );
}
