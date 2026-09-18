const STORAGE_KEY = "livepoll_voter_id";

function generateUUID() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    try {
      return crypto.randomUUID();
    } catch {
      // Fall through to fallback
    }
  }

  const getRandomValues = (arr) => {
    if (typeof crypto !== "undefined" && typeof crypto.getRandomValues === "function") {
      return crypto.getRandomValues(arr);
    }
    for (let i = 0; i < arr.length; i++) {
      arr[i] = Math.floor(Math.random() * 256);
    }
    return arr;
  };

  const bytes = new Uint8Array(16);
  getRandomValues(bytes);
  bytes[6] = (bytes[6] & 0x0f) | 0x40; // Version 4
  bytes[8] = (bytes[8] & 0x3f) | 0x80; // Variant 10xx

  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

/**
 * Returns a stable per-browser anonymous voter ID, creating one on first
 * use. Safe against missing crypto.randomUUID in non-secure or legacy contexts.
 */
export function getVoterId() {
  let id = localStorage.getItem(STORAGE_KEY);
  if (!id || id === "undefined" || id === "null" || id.trim() === "") {
    id = generateUUID();
    localStorage.setItem(STORAGE_KEY, id);
  }
  return id;
}
