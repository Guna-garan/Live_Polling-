export default function LiveIndicator({ status }) {
  const isLive = status === "live";
  return (
    <span className={`live-indicator ${isLive ? "is-live" : "is-reconnecting"}`}>
      <span className="live-dot" aria-hidden="true">
        {isLive ? "●" : "○"}
      </span>
      {isLive ? "Live" : status === "connecting" ? "Connecting…" : "Reconnecting…"}
    </span>
  );
}
