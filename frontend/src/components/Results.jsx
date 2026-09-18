import PollOption from "./PollOption";

export default function Results({
  options,
  results,
  totalVotes,
  selectable = false,
  selectedOptionId,
  onSelect,
}) {
  return (
    <div className="results" role={selectable ? "radiogroup" : undefined}>
      {options.map((opt) => (
        <PollOption
          key={opt.id}
          option={opt}
          count={results[opt.id] || 0}
          total={totalVotes}
          selectable={selectable}
          selected={selectedOptionId === opt.id}
          onSelect={onSelect}
        />
      ))}
      <p className="results-total">
        {totalVotes} total {totalVotes === 1 ? "vote" : "votes"}
      </p>
    </div>
  );
}
