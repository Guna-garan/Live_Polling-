export function percentage(count, total) {
  if (!total) return 0;
  return Math.round((count / total) * 100);
}

export function formatDate(dateString) {
  if (!dateString) return "";
  return new Date(dateString).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export function pollShareUrl(pollId) {
  return `${window.location.origin}/#/poll/${pollId}`;
}

const EXPIRY_OPTIONS = [
  { label: "No expiration", minutes: null },
  { label: "5 minutes", minutes: 5 },
  { label: "30 minutes", minutes: 30 },
  { label: "1 hour", minutes: 60 },
  { label: "24 hours", minutes: 60 * 24 },
];

export function expiryOptions() {
  return EXPIRY_OPTIONS;
}

export function expiresAtFromMinutes(minutes) {
  if (!minutes) return null;
  return new Date(Date.now() + minutes * 60_000).toISOString();
}
