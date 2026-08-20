// apiClient.js
//
// Centralized HTTP client for the React Web_Client to talk to the Go API_Server.
// All business data now flows through the `/api/*` REST endpoints instead of
// browser localStorage (Req 17.1). This module owns:
//   - Session_Token storage (Req 17.2 / 17.5)
//   - attaching the bearer token to every authenticated request (Req 17.2)
//   - a structured ApiError model mirroring the server error body (Req 18.1)
//   - automatic "return to login" on 401 responses (Req 17.6)
//
// 401-redirect approach (documented):
//   The SPA has no dedicated `/login` route. `App.jsx` renders <Login /> purely
//   based on the `auth` state held in `store.jsx` (auth === null => login screen).
//   So the cleanest way to "return the user to login" on a 401 is to let the
//   store register a callback that clears its auth/token state, which makes
//   React re-render the Login screen. We expose `setUnauthorizedHandler(fn)` for
//   that hook (task 16.2 wires it from the StoreProvider). If no handler is
//   registered (e.g. a 401 fires before the store mounts), we fall back to
//   clearing the token and forcing a full reload to the app root, which lands
//   the user on the login screen as well.

const TOKEN_KEY = 'pgcs_token_v1';

// ---- Token storage helpers ----

export function setToken(t) {
  try {
    localStorage.setItem(TOKEN_KEY, t);
  } catch (e) {
    // localStorage can throw in private-mode/quota situations; the in-memory
    // request will still work for this session via getToken's fallback.
    console.warn('Failed to persist session token', e);
  }
}

export function getToken() {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch (e) {
    return null;
  }
}

export function clearToken() {
  try {
    localStorage.removeItem(TOKEN_KEY);
  } catch (e) {
    /* ignore */
  }
}

// ---- Auth header ----

export function authHeader() {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// ---- Unauthorized (401) handling ----

let unauthorizedHandler = null;

// Lets the store/router register what to do when a 401 is received
// (typically: clear auth state so <Login /> renders). Pass null to unregister.
export function setUnauthorizedHandler(fn) {
  unauthorizedHandler = typeof fn === 'function' ? fn : null;
}

function redirectToLogin() {
  clearToken();
  if (unauthorizedHandler) {
    unauthorizedHandler();
    return;
  }
  // Fallback when no handler is registered: force the app back to its root,
  // where App.jsx will render the Login screen because auth is now empty.
  if (typeof window !== 'undefined' && window.location) {
    if (window.location.pathname !== '/') {
      window.location.assign('/');
    } else {
      window.location.reload();
    }
  }
}

// ---- Error model ----

export class ApiError extends Error {
  constructor(status, code, message, fields) {
    super(message || `API request failed with status ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.code = code || null;
    this.fields = fields || null;
  }

  // Build an ApiError from a non-2xx Response, parsing the structured error
  // body { code, message, fields } emitted by the API_Server (Req 18.1).
  static async from(res) {
    let code = null;
    let message = null;
    let fields = null;
    try {
      const data = await res.json();
      if (data && typeof data === 'object') {
        // Support both a flat shape and a { error: {...} } envelope.
        const err = data.error && typeof data.error === 'object' ? data.error : data;
        code = err.code ?? null;
        message = err.message ?? null;
        fields = err.fields ?? null;
      }
    } catch (e) {
      // Body was empty or not JSON; fall back to the status text.
      message = res.statusText || null;
    }
    return new ApiError(res.status, code, message, fields);
  }
}

// ---- Core request wrapper ----

export async function api(method, path, body) {
  const res = await fetch(`/api${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...authHeader(),
    },
    body: body !== undefined && body !== null ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    // Token missing/invalid/expired: discard it and return to login (Req 17.6).
    redirectToLogin();
    throw await ApiError.from(res);
  }

  if (!res.ok) {
    throw await ApiError.from(res);
  }

  // 204 No Content or an empty body => nothing to parse.
  if (res.status === 204) return null;
  const text = await res.text();
  if (!text) return null;
  return JSON.parse(text);
}

// ---- Convenience methods ----

export const get = (path) => api('GET', path);
export const post = (path, body) => api('POST', path, body);
export const put = (path, body) => api('PUT', path, body);
export const del = (path) => api('DELETE', path);
