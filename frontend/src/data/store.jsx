import { createContext, useContext, useEffect, useRef, useState } from 'react';
import {
  get, post, put, del,
  setToken, clearToken, getToken,
  setUnauthorizedHandler,
  ApiError,
} from './apiClient';

// store.jsx
//
// The StoreProvider keeps the SAME `useStore()` contract the existing screens
// rely on (`db`, `auth`, `add`, `update`, `remove`, `login`, `logout`,
// `resetData`, `setCollection`, `setDb`) but is now backed by the Go API_Server
// through `apiClient` instead of browser localStorage (Req 17.1).
//
// What changed vs. the localStorage implementation:
//   - `db` is an in-memory cache hydrated from the `/api/*` collection endpoints
//     when a session is established (login or token restore on mount). Pages keep
//     reading `db.<collection>` exactly as before, and `calc.js` keeps computing
//     display values off that same `db` shape.
//   - `login` posts to `/api/auth/login`, stores the returned Session_Token, and
//     derives the `{ role, name, id }` auth shape from the returned principal
//     (Req 17.2). The `expectedRole` role-mismatch behavior is preserved.
//   - `logout` revokes the token server-side (best effort), clears the token and
//     wipes the in-memory db (Req 17.5).
//   - `add`/`update`/`remove` call the matching REST endpoint and then patch the
//     in-memory db so screens re-render with the server's view.
//   - A 401 from any request clears the session so <Login /> re-renders (Req 17.6).
//
// Server-computed values (Req 17.3): transactions returned by the API already
// embed their commission breakdown, and the ledger/report endpoints are
// authoritative. `calc.js` is retained for display formatting (`inr`, `num`) and
// for the screens that still derive ledger/breakdown values from the cached `db`
// (those operate on the same server-sourced records, so the figures match the
// server's). Wiring individual screens to the dedicated ledger/report endpoints
// is a follow-up that does not change the `useStore()` contract.

const StoreContext = createContext(null);

// Map in-memory db collection keys <-> API path segments. The API uses kebab-case
// collection names; the existing db keys are camelCase (Req 17.1).
const COLLECTION_PATHS = {
  gateways: 'gateways',
  companies: 'companies',
  affiliates: 'affiliates',
  merchants: 'merchants',
  transactions: 'transactions',
  settlements: 'settlements',
  affiliatePayments: 'affiliate-payments',
  merchantPayments: 'merchant-payments',
};

const COLLECTION_KEYS = Object.keys(COLLECTION_PATHS);

function emptyDb() {
  return COLLECTION_KEYS.reduce((acc, key) => {
    acc[key] = [];
    return acc;
  }, {});
}

// Normalize a collection response to a plain array. The API may return a bare
// array or an envelope like { items: [...] }.
function asList(data) {
  if (Array.isArray(data)) return data;
  if (data && Array.isArray(data.items)) return data.items;
  return [];
}

// uid() is still used by screens to mint client-side ids for nested rows
// (bank accounts, ATM cards, gateway credentials) before they are sent to the
// server as part of a merchant update. Kept for backward compatibility.
let counter = Date.now();
export function uid(prefix = 'id') {
  counter += 1;
  return `${prefix}${counter}`;
}

export function StoreProvider({ children }) {
  const [db, setDb] = useState(emptyDb);
  const [auth, setAuth] = useState(null);
  // `ready` flips true once the initial session-restore attempt completes, so the
  // app can avoid flashing the Login screen while a stored token is validated.
  const [ready, setReady] = useState(false);

  // Keep a ref to the latest db so async ops (e.g. update's merge) never read a
  // stale closure value.
  const dbRef = useRef(db);
  useEffect(() => { dbRef.current = db; }, [db]);

  // ---- Collection loading ----

  const loadCollection = async (key) => {
    const data = await get(`/${COLLECTION_PATHS[key]}`);
    return asList(data);
  };

  // Load every collection into the in-memory db. Portal roles may not be
  // permitted to read some collections (403) or a collection may be empty;
  // those degrade to [] so the app keeps working.
  const loadAllCollections = async () => {
    const entries = await Promise.all(
      COLLECTION_KEYS.map(async (key) => {
        try {
          return [key, await loadCollection(key)];
        } catch (e) {
          if (e instanceof ApiError && (e.status === 403 || e.status === 404)) {
            return [key, []];
          }
          throw e;
        }
      })
    );
    const next = emptyDb();
    entries.forEach(([key, list]) => { next[key] = list; });
    setDb(next);
    return next;
  };

  // ---- Auth shape derivation ----
  //
  // Pages expect `auth` to look like { role, name, id } where `id` matches the
  // owning entity id for portal roles (company/affiliate/merchant pages do
  // `db.<collection>.find(r => r.id === auth.id)`). The login/me principal
  // carries { accountId, role, tenantId, ownerType, ownerId }; we map ownerId ->
  // id for portal roles and derive a display name from the loaded collections.
  const deriveName = (principal, dbState) => {
    const { role, ownerId } = principal;
    if (role === 'Admin') return 'Administrator';
    if (role === 'SuperAdmin') return 'Super Admin';
    const find = (coll) => (dbState[coll] || []).find((r) => r.id === ownerId);
    if (role === 'Company') return find('companies')?.name || 'Company';
    if (role === 'Affiliate') return find('affiliates')?.name || 'Affiliate';
    if (role === 'Merchant') return find('merchants')?.name || 'Merchant';
    return principal.name || role;
  };

  const authFromPrincipal = (principal, dbState) => ({
    role: principal.role,
    id: principal.ownerId || principal.accountId,
    name: deriveName(principal, dbState),
    // keep the full principal around for any future use without breaking the
    // existing { role, name, id } contract.
    principal,
  });

  // ---- Auth ----

  // login(userId, password, expectedRole?)
  // Returns the account on success, or { error } describing the issue, matching
  // the previous contract. NOTE: this is now async — Login.jsx awaits it.
  const login = async (userId, password, expectedRole) => {
    let resp;
    try {
      resp = await post('/auth/login', { userId, password });
    } catch (e) {
      if (e instanceof ApiError && e.status === 403) {
        // e.g. a leased Admin whose lease is expired/suspended/revoked.
        return { error: 'forbidden', message: e.message };
      }
      if (e instanceof ApiError && e.status === 401) {
        // Wrong user id or password (server does not reveal which).
        return { error: 'invalid' };
      }
      if (e instanceof ApiError) {
        // Some other HTTP error (4xx/5xx) — surface its message.
        return { error: 'server', message: e.message || `Request failed (${e.status}).` };
      }
      // Not an ApiError => the request never got a response: the API is
      // unreachable (backend down, or the dev server proxy is not forwarding
      // /api). Distinguish this from bad credentials so it is actionable.
      return { error: 'network', message: 'Cannot reach the server. Is the backend running and the /api proxy configured?' };
    }

    const principal = resp?.principal;
    const token = resp?.token;
    if (!token || !principal) return { error: 'invalid' };

    // Preserve the role-mismatch behavior: when a specific role tab is used, a
    // matching-credential-but-wrong-role login is reported, and we do NOT
    // establish a session for it.
    if (expectedRole && principal.role !== expectedRole) {
      return { error: 'role-mismatch', actualRole: principal.role };
    }

    setToken(token);
    const dbState = await loadAllCollections();
    const account = authFromPrincipal(principal, dbState);
    setAuth(account);
    return account;
  };

  const logout = async () => {
    try {
      await post('/auth/logout');
    } catch (e) {
      // best-effort: even if the server call fails we still drop the session
    }
    clearToken();
    setAuth(null);
    setDb(emptyDb());
  };

  // ---- Generic collection ops ----

  // Preserved for signature compatibility: mutate a single collection in the
  // in-memory cache. (No screen calls this directly anymore, but it remains part
  // of the contract.)
  const setCollection = (key, updater) => {
    setDb((prev) => ({ ...prev, [key]: updater(prev[key] || []) }));
  };

  // add(key, record) -> created record (with server-assigned id)
  const add = async (key, record) => {
    const created = await post(`/${COLLECTION_PATHS[key]}`, record);
    const saved = created || record;
    setDb((prev) => ({ ...prev, [key]: [...(prev[key] || []), saved] }));
    return saved;
  };

  // update(key, id, patch). Screens pass partial patches (including nested
  // collections like merchant.banks/paymentGateways); we merge with the cached
  // record so the PUT carries the full entity, then store the server's response.
  const update = async (key, id, patch) => {
    const existing = (dbRef.current[key] || []).find((r) => r.id === id) || {};
    const body = { ...existing, ...patch, id };
    const saved = await put(`/${COLLECTION_PATHS[key]}/${id}`, body);
    const result = saved || body;
    setDb((prev) => ({
      ...prev,
      [key]: (prev[key] || []).map((r) => (r.id === id ? result : r)),
    }));
    return result;
  };

  const remove = async (key, id) => {
    await del(`/${COLLECTION_PATHS[key]}/${id}`);
    setDb((prev) => ({ ...prev, [key]: (prev[key] || []).filter((r) => r.id !== id) }));
  };

  // resetData previously reseeded the local localStorage db. The server now owns
  // all data, so there is nothing to reset client-side; we re-fetch the
  // collections from the server instead (a safe refresh). Signature preserved.
  const resetData = async () => {
    if (!auth) return;
    await loadAllCollections();
  };

  // ---- 401 handling: return the user to the Login screen (Req 17.6) ----
  useEffect(() => {
    setUnauthorizedHandler(() => {
      clearToken();
      setAuth(null);
      setDb(emptyDb());
    });
    return () => setUnauthorizedHandler(null);
  }, []);

  // ---- Restore an existing session on mount ----
  useEffect(() => {
    let cancelled = false;
    const restore = async () => {
      const token = getToken();
      if (!token) {
        if (!cancelled) setReady(true);
        return;
      }
      try {
        const data = await get('/me');
        const principal = data?.principal || data;
        const dbState = await loadAllCollections();
        if (!cancelled && principal) setAuth(authFromPrincipal(principal, dbState));
      } catch (e) {
        clearToken();
        if (!cancelled) {
          setAuth(null);
          setDb(emptyDb());
        }
      } finally {
        if (!cancelled) setReady(true);
      }
    };
    restore();
    return () => { cancelled = true; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const value = {
    db, auth, ready,
    add, update, remove, setCollection, setDb,
    login, logout, resetData,
  };

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore() {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error('useStore must be used within StoreProvider');
  return ctx;
}
