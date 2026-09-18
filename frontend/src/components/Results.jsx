import { useState } from "react";
import PollOption from "./PollOption";
import DonutChart from "./DonutChart";
import { percentage } from "../utils/formatters";

export default function Results({
  options,
  results,
  totalVotes,
  selectable = false,
  selectedOptionId,
  onSelect,
}) {
  const [viewMode, setViewMode] = useState("bars"); // "bars" | "donut" | "table"

  return (
    <div className="results-wrapper">
      {!selectable && (
        <div className="view-mode-tabs">
          <button
            type="button"
            className={`tab-btn ${viewMode === "bars" ? "is-active" : ""}`}
            onClick={() => setViewMode("bars")}
          >
            📊 Bar Chart
          </button>
          <button
            type="button"
            className={`tab-btn ${viewMode === "donut" ? "is-active" : ""}`}
            onClick={() => setViewMode("donut")}
          >
            🍩 Donut Chart
          </button>
          <button
            type="button"
            className={`tab-btn ${viewMode === "table" ? "is-active" : ""}`}
            onClick={() => setViewMode("table")}
          >
            📋 Breakdown
          </button>
        </div>
      )}

      {viewMode === "bars" || selectable ? (
        <div className="results" role={selectable ? "radiogroup" : undefined}>
          {options.map((opt, i) => (
            <PollOption
              key={opt.id}
              option={opt}
              count={results[opt.id] || 0}
              total={totalVotes}
              selectable={selectable}
              selected={selectedOptionId === opt.id}
              onSelect={onSelect}
              colorIndex={i}
            />
          ))}
          <p className="results-total">
            {totalVotes} total {totalVotes === 1 ? "vote" : "votes"}
          </p>
        </div>
      ) : viewMode === "donut" ? (
        <DonutChart options={options} results={results} totalVotes={totalVotes} />
      ) : (
        <div className="table-breakdown">
          <table className="breakdown-table">
            <thead>
              <tr>
                <th>Option</th>
                <th>Votes</th>
                <th>Percentage</th>
              </tr>
            </thead>
            <tbody>
              {options.map((opt) => {
                const count = results[opt.id] || 0;
                const pct = percentage(count, totalVotes);
                return (
                  <tr key={opt.id}>
                    <td className="opt-text">{opt.text}</td>
                    <td className="opt-count">{count}</td>
                    <td className="opt-pct">{pct}%</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <p className="results-total">
            Total: <strong>{totalVotes}</strong> {totalVotes === 1 ? "vote" : "votes"}
          </p>
        </div>
      )}
    </div>
  );
}
