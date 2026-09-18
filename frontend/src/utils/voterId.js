const STORAGE_KEY = "livepoll_voter_id";

/**
 * Returns a stable per-browser anonymous voter ID, creating one on first
 * use. This is a lightweight anonymous-voting mechanism, NOT a strong
 * identity check — see the README's "Voter Identity" section. The
 * backend enforces the real duplicate-vote guard via a unique MongoDB
 * index on (pollId, voterId); this ID is just what gets checked against it.
 */
export function getVoterId() {
  let id = localStorage.getItem(STORAGE_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(STORAGE_KEY, id);
  }
  return id;
}
