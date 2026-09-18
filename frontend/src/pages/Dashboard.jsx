import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { pollApi } from "../services/api";
import { useAuth } from "../hooks/useAuth";
import { useToast } from "../components/Toast";
import PollCard from "../components/PollCard";
import Loading from "../components/Loading";
import ErrorMessage from "../components/ErrorMessage";
import ConfirmDialog from "../components/ConfirmDialog";

export default function Dashboard() {
  const { user } = useAuth();
  const showToast = useToast();
  const [polls, setPolls] = useState(null);
  const [error, setError] = useState("");
  const [pendingDelete, setPendingDelete] = useState(null);

  async function load() {
    try {
      const data = await pollApi.listMine();
      setPolls(data);
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleClose(id) {
    try {
      await pollApi.updateStatus(id, "closed");
      showToast("Poll closed", "success");
      load();
    } catch (err) {
      showToast(err.message, "error");
    }
  }

  async function handleDelete(id) {
    try {
      await pollApi.remove(id);
      showToast("Poll deleted", "success");
      setPendingDelete(null);
      load();
    } catch (err) {
      showToast(err.message, "error");
    }
  }

  function handleShare(url) {
    navigator.clipboard?.writeText(url);
    showToast("Link copied to clipboard", "success");
  }

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <div>
          <h1>Dashboard</h1>
          <p>Welcome back{user?.email ? `, ${user.email}` : ""}!</p>
        </div>
        <Link to="/create" className="btn btn-primary">
          Create Poll
        </Link>
      </div>

      <h2 className="section-title">Your Polls</h2>
      <ErrorMessage>{error}</ErrorMessage>

      {polls === null && !error && <Loading label="Loading your polls…" />}

      {polls && polls.length === 0 && (
        <div className="empty-state">
          <p>You haven't created any polls yet.</p>
          <Link to="/create" className="btn btn-primary">
            Create your first poll
          </Link>
        </div>
      )}

      {polls && polls.length > 0 && (
        <div className="poll-grid">
          {polls.map((poll) => (
            <PollCard
              key={poll.id}
              poll={poll}
              onClose={handleClose}
              onDelete={(id) => setPendingDelete(id)}
              onShare={handleShare}
            />
          ))}
        </div>
      )}

      <ConfirmDialog
        open={pendingDelete !== null}
        title="Delete this poll?"
        message="This permanently deletes the poll. Existing votes are not recoverable. This cannot be undone."
        confirmLabel="Delete"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => handleDelete(pendingDelete)}
      />
    </div>
  );
}
