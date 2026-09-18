export default function ErrorMessage({ children }) {
  if (!children) return null;
  return (
    <div className="error-message" role="alert">
      {children}
    </div>
  );
}
