import { useEffect, useState } from "react";
import { useParams, useLocation } from "react-router-dom";
import { QRCodeSVG } from "qrcode.react";
import { pollApi, voteApi, ApiError, API_URL } from "../services/api";
import { usePollSocket } from "../hooks/usePollSocket";
import { getVoterId } from "../utils/voterId";
import { pollShareUrl } from "../utils/formatters";
import { useToast } from "../components/Toast";
import Results from "../components/Results";
import LiveIndicator from "../components/LiveIndicator";
import Button from "../components/Button";
import Loading from "../components/Loading";
import ErrorMessage from "../components/ErrorMessage";
import Confetti from "../components/Confetti";

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
  const [showShareModal, setShowShareModal] = useState(false);

  const { results, totalVotes, activeWatchers, status, updateResults } = usePollSocket(id);

  useEffect(() => {
    const voterId = getVoterId();
    setHasVoted(!!localStorage.getItem(votedKey(id)));
    setJustVoted(false);
    setSelectedOptionId(null);

    if (voterId) {
      voteApi
        .status(id, voterId)
        .then((res) => {
          if (res && res.hasVoted) {
            localStorage.setItem(votedKey(id), "1");
            setHasVoted(true);
          }
        })
        .catch(() => {});
    }

    pollApi
      .get(id)
      .then((data) => {
        setPoll(data);
        if (data && data.results) {
          updateResults(data.results, data.totalVotes || 0);
        }
      })
      .catch((err) => setLoadError(err.message));
  }, [id, updateResults]);

  async function handleVote() {
    if (!selectedOptionId || voting) return;
    setVoting(true);
    setVoteError("");
    try {
      const res = await voteApi.cast(id, selectedOptionId, getVoterId());
      localStorage.setItem(votedKey(id), "1");
      setHasVoted(true);
      setJustVoted(true);
      if (res && res.results) {
        updateResults(res.results, res.totalVotes || 0);
      }
      showToast("Vote recorded live!", "success");
    } catch (err) {
      if (err instanceof ApiError && err.code === "ALREADY_VOTED") {
        localStorage.setItem(votedKey(id), "1");
        setHasVoted(true);
        if (err.results) {
          updateResults(err.results, err.totalVotes || 0);
        } else {
          pollApi
            .get(id)
            .then((data) => {
              if (data && data.results) updateResults(data.results, data.totalVotes || 0);
            })
            .catch(() => {});
        }
        showToast("You have already voted in this poll", "error");
      } else {
        setVoteError(err.message);
      }
    } finally {
      setVoting(false);
    }
  }

  function handleShare() {
    navigator.clipboard?.writeText(pollShareUrl(id));
    showToast("Share link copied to clipboard!", "success");
  }

  function handleDownloadQR(svgId = "poll-qr-svg") {
    const svgEl = document.getElementById(svgId) || document.getElementById("poll-qr-svg");
    if (!svgEl) return;
    const svgData = new XMLSerializer().serializeToString(svgEl);
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    const img = new Image();
    img.onload = () => {
      canvas.width = img.width + 40;
      canvas.height = img.height + 40;
      ctx.fillStyle = "#ffffff";
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.drawImage(img, 20, 20);
      const pngUrl = canvas.toDataURL("image/png");
      const downloadLink = document.createElement("a");
      downloadLink.href = pngUrl;
      downloadLink.download = `poll-${id}-qr.png`;
      document.body.appendChild(downloadLink);
      downloadLink.click();
      document.body.removeChild(downloadLink);
      showToast("QR Code downloaded!", "success");
    };
    img.src = "data:image/svg+xml;base64," + btoa(unescape(encodeURIComponent(svgData)));
  }

  if (loadError) {
    return (
      <div className="poll-page">
        <ErrorMessage>{loadError}</ErrorMessage>
      </div>
    );
  }

  if (!poll) return <Loading label="Loading poll details…" />;

  const isClosed = poll.status === "closed";
  const isExpired = poll.expiresAt && new Date(poll.expiresAt) < new Date();
  const votingLocked = isClosed || isExpired || hasVoted;

  return (
    <div className="poll-page">
      {justVoted && <Confetti duration={3000} />}

      {location.state?.justCreated && (
        <div className="share-banner glass-card">
          <div className="share-banner-content">
            <p className="share-title">🎉 Poll created successfully!</p>
            <p className="share-sub">Share this link or QR code with your audience to gather live votes:</p>
            <div className="share-row margin-v">
              <code className="share-link-code">{pollShareUrl(id)}</code>
              <Button variant="secondary" className="btn-sm" onClick={handleShare}>
                Copy Link
              </Button>
            </div>
            <div className="share-qr-section">
              <div className="qr-center">
                <QRCodeSVG id="poll-qr-banner-svg" value={pollShareUrl(id)} size={160} className="qr-code-img" />
              </div>
              <div className="qr-banner-actions">
                <Button variant="secondary" className="btn-sm" onClick={() => handleDownloadQR("poll-qr-banner-svg")}>
                  Download QR PNG
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}


      <div className="poll-top-bar">
        <LiveIndicator status={status} activeWatchers={activeWatchers} />
        <div className="poll-action-buttons">
          <Button variant="ghost" className="btn-sm" onClick={() => setShowShareModal(true)}>
            Share Poll
          </Button>
          <a
            href={`${API_URL}/api/polls/${id}/export`}
            download
            className="btn btn-ghost btn-sm"
            title="Download CSV report"
          >
            Export CSV
          </a>
        </div>
      </div>

      <div className="poll-card-container glass-card">
        <h1 className="poll-question">{poll.question}</h1>

        {isClosed && <p className="poll-status-note is-closed"> This poll is closed. Final live results below.</p>}
        {!isClosed && isExpired && <p className="poll-status-note is-expired"> This poll has expired. Final results below.</p>}

        {hasVoted && !isClosed && !isExpired && (
          <p className="poll-status-note poll-status-success">
            {justVoted
              ? "✓ Your vote was submitted — watching results update live below."
              : "✓ You have voted in this poll. Real-time updates active below."}
          </p>
        )}

        {!votingLocked ? (
          <div className="voting-section">
            <p className="poll-instruction">Select your answer:</p>
            <Results
              options={poll.options}
              results={results}
              totalVotes={totalVotes}
              selectable
              selectedOptionId={selectedOptionId}
              onSelect={setSelectedOptionId}
            />
            <ErrorMessage>{voteError}</ErrorMessage>
            <Button
              onClick={handleVote}
              disabled={!selectedOptionId || voting}
              loading={voting}
              className="btn-block btn-lg btn-primary vote-btn-pulse"
            >
              Submit Vote
            </Button>
          </div>
        ) : (
          <Results options={poll.options} results={results} totalVotes={totalVotes} />
        )}
      </div>

      {showShareModal && (
        <div className="modal-backdrop" onClick={() => setShowShareModal(false)}>
          <div className="modal glass-card" onClick={(e) => e.stopPropagation()}>
            <h3>Share this Poll</h3>
            <p>Anyone with this link can participate and view results live in real time.</p>
            <div className="share-row margin-v">
              <code className="share-link-code">{pollShareUrl(id)}</code>
              <Button variant="secondary" className="btn-sm" onClick={handleShare}>
                Copy
              </Button>
            </div>
            <div className="qr-center">
              <QRCodeSVG id="poll-qr-svg" value={pollShareUrl(id)} size={160} className="qr-code-img" />
            </div>
            <div className="modal-actions">
              <Button variant="secondary" onClick={handleDownloadQR}>
                 Download QR PNG
              </Button>
              <Button variant="secondary" onClick={() => setShowShareModal(false)}>
                Close
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
