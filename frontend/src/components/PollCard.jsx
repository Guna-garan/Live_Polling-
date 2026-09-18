import { Link } from "react-router-dom";
import Button from "./Button";
import { formatDate, pollShareUrl } from "../utils/formatters";

export default function PollCard({ poll, onClose, onDelete, onShare }) {
  return (
    <div className="poll-card">
      <h3 className="poll-card-question">{poll.question}</h3>
      <div className="poll-card-meta">
        <span className={`badge badge-${poll.status}`}>{poll.status}</span>
        <span>{poll.totalVotes} votes</span>
        <span>Created {formatDate(poll.createdAt)}</span>
        {poll.expiresAt && <span>Expires {formatDate(poll.expiresAt)}</span>}
      </div>
      <div className="poll-card-actions">
        <Link to={`/poll/${poll.id}`} className="btn btn-secondary btn-sm">
          View
        </Link>
        <Button variant="secondary" className="btn-sm" onClick={() => onShare(pollShareUrl(poll.id))}>
          Share
        </Button>
        {poll.status === "active" && (
          <Button variant="secondary" className="btn-sm" onClick={() => onClose(poll.id)}>
            Close
          </Button>
        )}
        <Button variant="danger" className="btn-sm" onClick={() => onDelete(poll.id)}>
          Delete
        </Button>
      </div>
    </div>
  );
}
