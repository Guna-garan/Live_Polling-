import { percentage } from "../utils/formatters";

const CHART_COLORS = [
  "#6366f1", // Indigo
  "#10b981", // Emerald
  "#f59e0b", // Amber
  "#ec4899", // Pink
  "#06b6d4", // Cyan
  "#8b5cf6", // Violet
  "#3b82f6", // Blue
  "#f97316", // Orange
  "#14b8a6", // Teal
  "#ef4444", // Red
];

export default function DonutChart({ options, results, totalVotes }) {
  if (!totalVotes || totalVotes === 0) {
    return (
      <div className="donut-chart-empty">
        <p>No votes recorded yet. Be the first to vote!</p>
      </div>
    );
  }

  // Calculate SVG arc paths
  const radius = 70;
  const strokeWidth = 24;
  const circumference = 2 * Math.PI * radius;

  let accumulatedPercent = 0;

  const slices = options.map((opt, i) => {
    const count = results[opt.id] || 0;
    const pct = percentage(count, totalVotes);
    const strokeDasharray = `${(pct / 100) * circumference} ${circumference}`;
    const strokeDashoffset = -((accumulatedPercent / 100) * circumference);

    accumulatedPercent += pct;
    const color = CHART_COLORS[i % CHART_COLORS.length];

    return {
      id: opt.id,
      text: opt.text,
      count,
      pct,
      color,
      strokeDasharray,
      strokeDashoffset,
    };
  });

  return (
    <div className="donut-chart-container">
      <div className="donut-chart-graphic">
        <svg viewBox="0 0 200 200" className="donut-svg">
          <circle
            cx="100"
            cy="100"
            r={radius}
            fill="transparent"
            stroke="var(--line)"
            strokeWidth={strokeWidth}
          />
          {slices.map(
            (s) =>
              s.pct > 0 && (
                <circle
                  key={s.id}
                  cx="100"
                  cy="100"
                  r={radius}
                  fill="transparent"
                  stroke={s.color}
                  strokeWidth={strokeWidth}
                  strokeDasharray={s.strokeDasharray}
                  strokeDashoffset={s.strokeDashoffset}
                  transform="rotate(-90 100 100)"
                  className="donut-slice"
                >
                  <title>{`${s.text}: ${s.count} votes (${s.pct}%)`}</title>
                </circle>
              )
          )}
        </svg>
        <div className="donut-center-text">
          <span className="donut-total-num">{totalVotes}</span>
          <span className="donut-total-label">{totalVotes === 1 ? "Vote" : "Votes"}</span>
        </div>
      </div>

      <div className="donut-legend">
        {slices.map((s) => (
          <div className="legend-item" key={s.id}>
            <span className="legend-color-dot" style={{ background: s.color }} />
            <span className="legend-text">{s.text}</span>
            <span className="legend-count">
              {s.count} ({s.pct}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
