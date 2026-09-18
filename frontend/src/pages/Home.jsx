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
    <div className="landing">
      <section className="hero">
        <h1>LIVEPOLL</h1>
        <p className="hero-tagline">Create. Share. Vote. Watch it happen live.</p>
        <p className="hero-sub">Real-time polling for real audiences.</p>
        <div className="hero-actions">
          <Link to={user ? "/create" : "/signup"} className="btn btn-primary btn-lg">
            Create a Poll
          </Link>
        </div>
        <form className="join-form" onSubmit={handleJoin}>
          <Input
            placeholder="Paste a poll link or ID to join"
            value={joinId}
            onChange={(e) => setJoinId(e.target.value)}
          />
          <Button type="submit" variant="secondary">
            Join a Poll
          </Button>
        </form>
      </section>

      <section className="features">
        <div className="feature">
          <h3>Real-time results</h3>
          <p>Every vote updates every viewer's screen instantly — no refresh, ever.</p>
        </div>
        <div className="feature">
          <h3>Anonymous voting</h3>
          <p>Your audience votes with one click. No account, no signup, no friction.</p>
        </div>
        <div className="feature">
          <h3>Shareable links</h3>
          <p>One link, any number of viewers, all watching the same live results.</p>
        </div>
      </section>
    </div>
  );
}
