import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import Button from "./Button";

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate("/");
  }

  return (
    <header className="navbar glass-header">
      <div className="navbar-container">
        <Link to="/" className="navbar-brand">
          <span className="brand-text">LIVEPOLL</span>
        </Link>
        <nav className="navbar-links">
          {user ? (
            <>
              <Link to="/dashboard" className="nav-link">
                Dashboard
              </Link>
              <Link to="/create" className="btn btn-primary btn-sm">
                + New Poll
              </Link>
              <div className="user-badge-pill">
                <span className="user-email">{user.email?.split("@")[0]}</span>
                <Button variant="ghost" className="btn-sm logout-btn" onClick={handleLogout}>
                  Log out
                </Button>
              </div>
            </>
          ) : (
            <>
              <Link to="/login" className="nav-link">
                Log in
              </Link>
              <Link to="/signup" className="btn btn-primary btn-sm btn-glow">
                Sign up
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
