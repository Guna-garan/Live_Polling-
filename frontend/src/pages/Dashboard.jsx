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
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all"); // "all" | "active" | "closed"

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
    showToast("Link copied to clipboard!", "success");
  }

  // Filter polls
  const filteredPolls = (polls || []).filter((p) => {
    const matchesSearch = p.question.toLowerCase().includes(searchQuery.toLowerCase());
    const isExpired = p.expiresAt && new Date(p.expiresAt) < new Date();
    const effectiveStatus = p.status === "closed" || isExpired ? "closed" : "active";
    if (statusFilter === "active") return matchesSearch && effectiveStatus === "active";
    if (statusFilter === "closed") return matchesSearch && effectiveStatus === "closed";
    return matchesSearch;
  });

  const totalPollsCount = polls?.length || 0;
  const activePollsCount = (polls || []).filter(
    (p) => p.status === "active" && (!p.expiresAt || new Date(p.expiresAt) > new Date())
  ).length;
  const totalVotesCount = (polls || []).reduce((acc, p) => acc + (p.totalVotes || 0), 0);

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <div>
          <h1 className="dash-title">Creator Dashboard</h1>
          <p className="dash-welcome">Welcome back{user?.email ? `, ${user.email}` : ""}!</p>
        </div>
        <Link to="/create" className="btn btn-primary btn-lg">
          + Create New Poll
        </Link>
      </div>

      {polls && (
        <div className="dashboard-stats-grid">
          <div className="stat-card glass-card">
            <span className="stat-label">Total Polls</span>
            <span className="stat-value">{totalPollsCount}</span>
          </div>
          <div className="stat-card glass-card">
            <span className="stat-label">Active Polls</span>
            <span className="stat-value glow-active">{activePollsCount}</span>
          </div>
          <div className="stat-card glass-card">
            <span className="stat-label">Total Votes Collected</span>
            <span className="stat-value glow-confirm">{totalVotesCount}</span>
          </div>
        </div>
      )}

      <div className="dashboard-controls">
        <div className="search-bar">
          <input
            type="text"
            className="input search-input"
            placeholder="Search your polls by question..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <div className="filter-tabs">
          <button
            type="button"
            className={`filter-btn ${statusFilter === "all" ? "is-active" : ""}`}
            onClick={() => setStatusFilter("all")}
          >
            All Polls ({totalPollsCount})
          </button>
          <button
            type="button"
            className={`filter-btn ${statusFilter === "active" ? "is-active" : ""}`}
            onClick={() => setStatusFilter("active")}
          >
            Active ({activePollsCount})
          </button>
          <button
            type="button"
            className={`filter-btn ${statusFilter === "closed" ? "is-active" : ""}`}
            onClick={() => setStatusFilter("closed")}
          >
            Closed ({totalPollsCount - activePollsCount})
          </button>
        </div>
      </div>

      <ErrorMessage>{error}</ErrorMessage>

      {polls === null && !error && <Loading label="Loading your polls…" />}

      {polls && polls.length === 0 && (
        <div className="empty-state glass-card">
          <p className="empty-title">You haven't created any polls yet</p>
          <p className="empty-sub">Create your first live poll in seconds and share it with your audience.</p>
          <Link to="/create" className="btn btn-primary">
            Create your first poll
          </Link>
        </div>
      )}

      {polls && polls.length > 0 && filteredPolls.length === 0 && (
        <div className="empty-state glass-card">
          <p>No polls match your search or filter criteria.</p>
        </div>
      )}

      {filteredPolls.length > 0 && (
        <div className="poll-grid">
          {filteredPolls.map((poll) => (
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
        message="This permanently deletes the poll and all live vote data. This action cannot be undone."
        confirmLabel="Delete Poll"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => handleDelete(pendingDelete)}
      />
    </div>
  );
}
