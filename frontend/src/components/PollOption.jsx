import { percentage } from "../utils/formatters";

/**
 * A single option's result bar. The width transition (see index.css) is
 * what gives the "animated horizontal bar" effect from the spec — no
 * animation library needed, just a CSS transition on width driven by
 * React re-rendering with new percentages as WebSocket events arrive.
 */
export default function PollOption({ option, count, total, selected, selectable, onSelect }) {
  const pct = percentage(count, total);

  return (
    <div
      className={`poll-option ${selectable ? "is-selectable" : ""} ${selected ? "is-selected" : ""}`}
      onClick={selectable ? () => onSelect(option.id) : undefined}
      role={selectable ? "radio" : undefined}
      aria-checked={selectable ? selected : undefined}
    >
      <div className="poll-option-row">
        <span className="poll-option-text">{option.text}</span>
        <span className="poll-option-pct">{pct}%</span>
      </div>
      <div className="poll-option-bar-track">
        <div className="poll-option-bar-fill" style={{ width: `${pct}%` }} />
      </div>
      <span className="poll-option-count">
        {count} {count === 1 ? "vote" : "votes"}
      </span>
    </div>
  );
}
