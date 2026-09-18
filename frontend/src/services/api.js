const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

/**
 * ApiError carries the backend's structured { code, message } so callers
 * can branch on `err.code` (e.g. "ALREADY_VOTED") instead of parsing text.
 */
export class ApiError extends Error {
  constructor(status, code, message) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request(path, options = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    credentials: "include", // send the HttpOnly auth cookie
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  if (res.status === 204) return null;

  let body = null;
  try {
    body = await res.json();
  } catch {
    // no JSON body (e.g. network-level failure) — fall through
  }

  if (!res.ok) {
    const code = body?.error?.code || "INTERNAL_ERROR";
    const message = body?.error?.message || "Something went wrong";
    throw new ApiError(res.status, code, message);
  }

  return body;
}

export const authApi = {
  signup: (email, password, confirmPassword) =>
    request("/api/auth/signup", {
      method: "POST",
      body: JSON.stringify({ email, password, confirmPassword }),
    }),
  login: (email, password) =>
    request("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request("/api/auth/logout", { method: "POST" }),
  me: () => request("/api/auth/me"),
};

export const pollApi = {
  create: (payload) =>
    request("/api/polls", { method: "POST", body: JSON.stringify(payload) }),
  listMine: () => request("/api/polls"),
  get: (id) => request(`/api/polls/${id}`),
  updateStatus: (id, status) =>
    request(`/api/polls/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    }),
  remove: (id) => request(`/api/polls/${id}`, { method: "DELETE" }),
};

export const voteApi = {
  cast: (pollId, optionId, voterId) =>
    request(`/api/polls/${pollId}/vote`, {
      method: "POST",
      body: JSON.stringify({ optionId, voterId }),
    }),
};

export const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
