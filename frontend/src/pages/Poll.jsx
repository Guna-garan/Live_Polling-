import { useEffect, useState } from "react";
import { useParams, useLocation } from "react-router-dom";
import { QRCodeSVG } from "qrcode.react";
import { pollApi, voteApi, ApiError } from "../services/api";
import { usePollSocket } from "../hooks/usePollSocket";
import { getVoterId } from "../utils/voterId";
import { pollShareUrl } from "../utils/formatters";
import { useToast } from "../components/Toast";
import Results from "../components/Results";
import LiveIndicator from "../components/LiveIndicator";
import Button from "../components/Button";
import Loading from "../components/Loading";
import ErrorMessage from "../components/ErrorMessage";

function votedKey(pollId) {
  return `livepoll_voted:${pollId}`;
}

export default function Poll() {
  const { id } = useParams();
  const location = useLocation();
  const showToast = useToast();

  const [poll, setPoll] = useState(null);
  const [loadError, setLoadError] = useState("");
  const [selectedOptionId, setSelectedOptionId] = useState(null);
  const [voting, setVoting] = useState(false);
  const [voteError, setVoteError] = useState("");
  const [hasVoted, setHasVoted] = useState(() => !!localStorage.getItem(votedKey(id)));
  const [justVoted, setJustVoted] = useState(false);

  const { results, totalVotes, status } = usePollSocket(id);

  useEffect(() => {
    pollApi
      .get(id)
      .then(setPoll)
      .catch((err) => setLoadError(err.message));
  }, [id]);

  async function handleVote() {
    if (!selectedOptionId) return;
    setVoting(true);
    setVoteError("");
    try {
      await voteApi.cast(id, selectedOptionId, getVoterId());
      localStorage.setItem(votedKey(id), "1");
      setHasVoted(true);
      setJustVoted(true);
    } catch (err) {
      if (err instanceof ApiError && err.code === "ALREADY_VOTED") {
        localStorage.setItem(votedKey(id), "1");
        setHasVoted(true);
      } else {
        setVoteError(err.message);
      }
    } finally {
      setVoting(false);
    }
  }

  function handleShare() {
    navigator.clipboard?.writeText(pollShareUrl(id));
    showToast("Link copied to clipboard", "success");
  }

  if (loadError) {
    return (
      <div className="poll-page">
        <ErrorMessage>{loadError}</ErrorMessage>
      </div>
    );
  }

  if (!poll) return <Loading label="Loading poll…" />;

  const isClosed = poll.status === "closed";
  const isExpired = poll.expiresAt && new Date(poll.expiresAt) < new Date();
  const votingLocked = isClosed || isExpired || hasVoted;

  return (
    <div className="poll-page">
      {location.state?.justCreated && (
        <div className="share-banner">
          <p>Poll created! Share it:</p>
          <div className="share-row">
            <code>{pollShareUrl(id)}</code>
            <Button variant="secondary" className="btn-sm" onClick={handleShare}>
              Copy Link
            </Button>
          </div>
          <div className="share-qr">
            <QRCodeSVG value={pollShareUrl(id)} size={96} />
          </div>
        </div>
      )}

      <div className="poll-header">
        <LiveIndicator status={status} />
      </div>

      <h1 className="poll-question">{poll.question}</h1>

      {isClosed && <p className="poll-status-note">This poll is closed. Final Results</p>}
      {!isClosed && isExpired && <p className="poll-status-note">This poll has ended. Final Results</p>}

      {!votingLocked && (
        <>
          <p className="poll-instruction">Select an option</p>
          <Results
            options={poll.options}
            results={results}
            totalVotes={totalVotes}
            selectable
            selectedOptionId={selectedOptionId}
            onSelect={setSelectedOptionId}
          />
          <ErrorMessage>{voteError}</ErrorMessage>
          <Button onClick={handleVote} disabled={!selectedOptionId} loading={voting} className="btn-block">
            Vote
          </Button>
        </>
      )}

      {votingLocked && (
        <>
          {hasVoted && !isClosed && !isExpired && (
            <p className="poll-status-note poll-status-success">
              {justVoted
                ? "✓ Vote recorded — results are updating live."
                : "You've already voted in this poll."}
            </p>
          )}
          <Results options={poll.options} results={results} totalVotes={totalVotes} />
        </>
      )}
    </div>
  );
}
