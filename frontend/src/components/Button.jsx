export default function Button({
  children,
  variant = "primary",
  disabled = false,
  loading = false,
  type = "button",
  onClick,
  className = "",
}) {
  return (
    <button
      type={type}
      disabled={disabled || loading}
      onClick={onClick}
      className={`btn btn-${variant} ${className}`}
    >
      {loading ? "…" : children}
    </button>
  );
}
