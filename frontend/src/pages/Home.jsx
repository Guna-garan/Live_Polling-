import { Link } from "react-router-dom";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import Button from "../components/Button";
import Input from "../components/Input";
import { useAuth } from "../hooks/useAuth";

export default function Home() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [joinId, setJoinId] = useState("");

  function handleJoin(e) {
    e.preventDefault();
    const id = joinId.trim().split("/").pop();
    if (id) navigate(`/poll/${id}`);
  }

  return (
    <div className="landing-page">
      <section className="hero">
        <div className="hero-copy">
          <div className="hero-badge">
            <span className="badge-pulse" />
            <span>Realtime Polling Engine</span>
          </div>
          <h1>Create. Share. Vote.<br /><span className="gradient-text">Watch it happen live.</span></h1>
          <p className="hero-sub">
            Instant, zero-refresh audience polling powered by Go, Redis Pub/Sub, MongoDB, and WebSockets.
          </p>
          <div className="hero-actions">
            <Link to={user ? "/create" : "/signup"} className="btn btn-primary btn-lg btn-glow">
              Create a Poll
            </Link>
          </div>
          <form className="join-form glass-card" onSubmit={handleJoin}>
            <Input
              placeholder="Paste poll link or ID to join live vote..."
              value={joinId}
              onChange={(e) => setJoinId(e.target.value)}
            />
            <Button type="submit" variant="secondary">
              Join Poll
            </Button>
          </form>
        </div>

        <div className="hero-visual glass-card">
          <div className="visual-header">
            <span className="visual-dot red" />
            <span className="visual-dot yellow" />
            <span className="visual-dot green" />
            <span className="visual-title">Live Results Stream</span>
          </div>
          <div className="equalizer-box">
            <div className="hero-bar" style={{ "--target-h": "65%" }}>
              <span className="bar-label">Option A</span>
            </div>
            <div className="hero-bar bar-emerald" style={{ "--target-h": "92%" }}>
              <span className="bar-label">Option B</span>
            </div>
            <div className="hero-bar bar-cyan" style={{ "--target-h": "45%" }}>
              <span className="bar-label">Option C</span>
            </div>
            <div className="hero-bar bar-amber" style={{ "--target-h": "80%" }}>
              <span className="bar-label">Option D</span>
            </div>
          </div>
        </div>
      </section>

      <section className="features-grid">
        <div className="feature-card glass-card">
          <div className="feature-icon"></div>
          <h3>Zero-Refresh Realtime</h3>
          <p>Votes flow over high-speed WebSockets backed by Redis Pub/Sub. Screen updates in milliseconds for everyone.</p>
        </div>
        <div className="feature-card glass-card">
          <div className="feature-icon"></div>
          <h3>Frictionless Voting</h3>
          <p>Audience members vote instantly with one tap. Unique DB indexes guard against duplicate votes without requiring logins.</p>
        </div>
        <div className="feature-card glass-card">
          <div className="feature-icon"></div>
          <h3>Backend Validated & Secure</h3>
          <p>Strict server-side validation, HttpOnly session cookies, bcrypt auth, and full ownership verification.</p>
        </div>
      </section>
    </div>
  );
}
