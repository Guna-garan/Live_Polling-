export default function Loading({ label = "Loading…" }) {
  return (
    <div className="loading-state">
      <div className="spinner" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}
