import { percentage } from "../utils/formatters";

/**
 * A single option's result bar. The width transition (see index.css) is
 * what gives the "animated horizontal bar" effect from the spec — no
 * animation library needed, just a CSS transition on width driven by
 * React re-rendering with new percentages as WebSocket events arrive.
 */
export default function PollOption({
  option,
  count,
  total,
  selected,
  selectable,
  onSelect,
  colorIndex = 0,
}) {
  const pct = percentage(count, total);

  const colors = [
    "var(--grad-1)",
    "var(--grad-2)",
    "var(--grad-3)",
    "var(--grad-4)",
    "var(--grad-5)",
  ];
  const fillGradient = colors[colorIndex % colors.length];

  return (
    <div
      className={`poll-option ${selectable ? "is-selectable" : ""} ${selected ? "is-selected" : ""}`}
      onClick={selectable ? () => onSelect(option.id) : undefined}
      role={selectable ? "radio" : undefined}
      aria-checked={selectable ? selected : undefined}
      tabIndex={selectable ? 0 : undefined}
      onKeyDown={
        selectable
          ? (e) => {
              if (e.key === " " || e.key === "Enter") {
                e.preventDefault();
                onSelect(option.id);
              }
            }
          : undefined
      }
    >
      <div className="poll-option-row">
        <div className="poll-option-left">
          {selectable && (
            <span className={`radio-indicator ${selected ? "is-checked" : ""}`}>
              <span className="radio-dot" />
            </span>
          )}
          <span className="poll-option-text">{option.text}</span>
        </div>
        <span className="poll-option-pct">{pct}%</span>
      </div>
      <div className="poll-option-bar-track">
        <div
          className="poll-option-bar-fill"
          style={{ width: `${pct}%`, background: fillGradient }}
        />
      </div>
      <div className="poll-option-footer">
        <span className="poll-option-count">
          {count} {count === 1 ? "vote" : "votes"}
        </span>
      </div>
    </div>
  );
}
