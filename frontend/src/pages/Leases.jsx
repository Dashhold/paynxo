import { useEffect, useState } from 'react';
import { get, post, put, ApiError } from '../data/apiClient';
import { PageHead, Empty, Field, Confirm } from '../components/ui';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

// Leases.jsx — SuperAdmin-only lease management screen (Req 16.3).
//
// Lists every lease and lets the SuperAdmin create new leased Admins and run the
// lifecycle operations (extend / suspend / reactivate / revoke) against the
// `/api/leases` endpoints. Lease data is not part of the store.jsx collections,
// so this screen talks to the API directly through apiClient (get/post) and
// surfaces validation (400) and conflict (409) errors from ApiError.

const blankCreate = { adminUserId: '', adminName: '', password: '', startDate: '', expiryDate: '' };

// Lease statuses are Active | Expired | Suspended | Revoked. Reuse the shared
// badge classes: Active reads as "solid", everything else as "muted".
function LeaseBadge({ status }) {
  const active = status === 'Active';
  return <span className={`badge ${active ? 'solid' : 'muted'}`}>{status}</span>;
}

// Render a date that may arrive as a full ISO timestamp or a yyyy-mm-dd string.
function fmtDate(d) {
  if (!d) return '—';
  const s = String(d);
  return s.length >= 10 ? s.slice(0, 10) : s;
}

// Turn an ApiError into a human-readable message. 400 may carry field-level
// details (Req 18.3); 409 carries a conflict message (Req 13.5).
function errMessage(e) {
  if (e instanceof ApiError) {
    if (e.fields && typeof e.fields === 'object') {
      const parts = Object.entries(e.fields).map(([k, v]) => `${k}: ${v}`);
      if (parts.length) return parts.join('; ');
    }
    return e.message || `Request failed (${e.status}).`;
  }
  return e?.message || 'Something went wrong.';
}

export default function Leases() {
  const [leases, setLeases] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(null);

  // Create modal state
  const [creating, setCreating] = useState(false);
  const [createForm, setCreateForm] = useState(blankCreate);
  const [createErr, setCreateErr] = useState(null);
  const [createBusy, setCreateBusy] = useState(false);

  // Extend modal state
  const [extendTarget, setExtendTarget] = useState(null); // lease being extended
  const [extendDate, setExtendDate] = useState('');
  const [extendErr, setExtendErr] = useState(null);
  const [extendBusy, setExtendBusy] = useState(false);

  // Per-row action state (id currently processing) + inline action errors
  const [actingId, setActingId] = useState(null);
  const [actionErr, setActionErr] = useState(null);
  // Confirmation prompt for the suspend / reactivate / revoke actions.
  const [confirmAction, setConfirmAction] = useState(null); // { lease, action, label }

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  // View Admin modal state
  const [viewingAccount, setViewingAccount] = useState(null); // account object
  const [editing, setEditing] = useState(false);
  const [editForm, setEditForm] = useState({ userId: '', password: '', name: '' });
  const [showPassword, setShowPassword] = useState(false);
  const [showViewPassword, setShowViewPassword] = useState(false);
  const [editErr, setEditErr] = useState(null);
  const [editBusy, setEditBusy] = useState(false);

  const asList = (data) => (Array.isArray(data) ? data : data?.items || []);

  const load = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const data = await get('/leases');
      setLeases(asList(data));
    } catch (e) {
      setLoadError(errMessage(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  // Paginated data
  const totalLeases = leases.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedLeases = leases.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // ---- Create ----
  const openCreate = () => { setCreateForm(blankCreate); setCreateErr(null); setCreating(true); };

  const submitCreate = async () => {
    setCreateErr(null);
    const { adminUserId, password, startDate, expiryDate } = createForm;
    if (!adminUserId.trim() || !password.trim() || !startDate || !expiryDate) {
      setCreateErr('Admin user id, password, start date and expiry date are all required.');
      return;
    }
    setCreateBusy(true);
    try {
      await post('/leases', {
        adminUserId: adminUserId.trim(),
        adminName: createForm.adminName.trim(),
        password,
        startDate,
        expiryDate,
      });
      setCreating(false);
      await load();
    } catch (e) {
      setCreateErr(errMessage(e));
    } finally {
      setCreateBusy(false);
    }
  };

  // ---- Extend ----
  const openExtend = (lease) => {
    setExtendTarget(lease);
    setExtendDate(fmtDate(lease.expiryDate));
    setExtendErr(null);
  };

  const submitExtend = async () => {
    setExtendErr(null);
    if (!extendDate) { setExtendErr('A new expiry date is required.'); return; }
    setExtendBusy(true);
    try {
      await post(`/leases/${extendTarget.id}/extend`, { expiryDate: extendDate });
      setExtendTarget(null);
      await load();
    } catch (e) {
      setExtendErr(errMessage(e));
    } finally {
      setExtendBusy(false);
    }
  };

  // ---- Suspend / Reactivate / Revoke ----
  const runAction = async (lease, action) => {
    setActionErr(null);
    setActingId(lease.id);
    try {
      await post(`/leases/${lease.id}/${action}`);
      await load();
    } catch (e) {
      setActionErr(errMessage(e));
    } finally {
      setActingId(null);
    }
  };

  // ---- View Admin ----
  const openViewAdmin = async (lease) => {
    setEditErr(null);
    setEditing(false);
    setShowPassword(false);
    setShowViewPassword(false);
    try {
      const account = await get(`/accounts/${lease.accountId}`);
      console.log('Fetched account:', account);
      console.log('Password field:', account.password);
      setViewingAccount(account);
      setEditForm({ userId: account.userId, password: '', name: account.name });
    } catch (e) {
      setActionErr(errMessage(e));
    }
  };

  const submitEditAdmin = async () => {
    setEditErr(null);
    const { userId, password, name } = editForm;
    if (!userId.trim() || !name.trim()) {
      setEditErr('User ID and name are required.');
      return;
    }
    setEditBusy(true);
    try {
      const updated = await put(`/accounts/${viewingAccount.id}`, { userId: userId.trim(), password, name: name.trim() });
      setViewingAccount(updated);
      setEditForm({ ...editForm, password: '', userId: updated.userId, name: updated.name });
      setEditing(false);
      setShowPassword(false);
    } catch (e) {
      setEditErr(errMessage(e));
    } finally {
      setEditBusy(false);
    }
  };

  return (
    <div>
      <PageHead
        title="Lease Management"
        sub="Create leased Admin instances and manage their tenure and access."
        actions={<button className="btn primary" onClick={openCreate}>+ Create Lease</button>}
      />

      {actionErr && <div className="panel" style={{ marginBottom: 12 }}><div className="panel-body" style={{ color: 'var(--danger, #c0392b)' }}>{actionErr}</div></div>}

      <div className="panel">
        {loading ? (
          <div className="panel-body"><Empty icon="⧉" text="Loading leases…" /></div>
        ) : loadError ? (
          <div className="panel-body" style={{ color: 'var(--danger, #c0392b)' }}>{loadError}</div>
        ) : leases.length === 0 ? (
          <div className="panel-body"><Empty icon="⧉" text="No leases yet." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>User ID</th>
                  <th>Name</th>
                  <th>Start Date</th>
                  <th>Expiry Date</th>
                  <th>Status</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedLeases.map((l) => {
                  const busy = actingId === l.id;
                  const isActive = l.status === 'Active';
                  const isSuspended = l.status === 'Suspended';
                  const isRevoked = l.status === 'Revoked';
                  
                  return (
                    <tr key={l.id}>
                      <td className="bold">{l.adminUserId}</td>
                      <td>{l.adminName || '—'}</td>
                      <td className="mono">{fmtDate(l.startDate)}</td>
                      <td className="mono">{fmtDate(l.expiryDate)}</td>
                      <td><LeaseBadge status={l.status} /></td>
                      <td className="center">
                        <div className="row-actions" style={{ justifyContent: 'center' }}>
                          <button className="btn sm ghost" disabled={busy} onClick={() => openViewAdmin(l)}>View</button>
                          {!isRevoked && <button className="btn sm ghost" disabled={busy} onClick={() => openExtend(l)}>Extend</button>}
                          {(isActive || l.status === 'Expired') && <button className="btn sm ghost" disabled={busy} onClick={() => setConfirmAction({ lease: l, action: 'suspend', label: 'Suspend', remarks: 'This will temporarily block access for this admin.' })}>Suspend</button>}
                          {isSuspended && <button className="btn sm success" disabled={busy} onClick={() => setConfirmAction({ lease: l, action: 'reactivate', label: 'Reactivate', remarks: 'This will restore access for this admin.' })}>Reactivate</button>}
                          {!isRevoked && <button className="btn sm danger" disabled={busy} onClick={() => setConfirmAction({ lease: l, action: 'revoke', label: 'Revoke', remarks: 'This will permanently disable access. This action cannot be undone.' })}>Revoke</button>}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalLeases > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalLeases}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={(newSize) => { setItemsPerPage(newSize); setCurrentPage(1); }}
              />
            )}
          </>
        )}
      </div>

      {creating && (
        <Modal
          title="Create Lease"
          onClose={() => setCreating(false)}
          footer={<>
            <button className="btn ghost" onClick={() => setCreating(false)} disabled={createBusy}>Cancel</button>
            <button className="btn primary" onClick={submitCreate} disabled={createBusy}>
              {createBusy ? 'Creating…' : 'Create Lease'}
            </button>
          </>}
        >
          {createErr && <div className="form-error" style={{ color: 'var(--danger, #c0392b)', marginBottom: 12 }}>{createErr}</div>}
          <div className="section-title">New Leased Admin</div>
          <div className="form-grid">
            <Field label="Admin User ID" span={2}>
              <input value={createForm.adminUserId} onChange={(e) => setCreateForm({ ...createForm, adminUserId: e.target.value })} autoFocus />
            </Field>
            <Field label="Admin Name" span={2}>
              <input value={createForm.adminName} onChange={(e) => setCreateForm({ ...createForm, adminName: e.target.value })} placeholder="Display name for this admin" />
            </Field>
            <Field label="Password" span={2}>
              <input type="password" value={createForm.password} onChange={(e) => setCreateForm({ ...createForm, password: e.target.value })} />
            </Field>
            <Field label="Start Date">
              <input type="date" value={createForm.startDate} onChange={(e) => setCreateForm({ ...createForm, startDate: e.target.value })} />
            </Field>
            <Field label="Expiry Date">
              <input type="date" value={createForm.expiryDate} onChange={(e) => setCreateForm({ ...createForm, expiryDate: e.target.value })} />
            </Field>
          </div>
        </Modal>
      )}

      {extendTarget && (
        <Modal
          title={`Extend Lease — ${extendTarget.adminUserId}`}
          onClose={() => setExtendTarget(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setExtendTarget(null)} disabled={extendBusy}>Cancel</button>
            <button className="btn primary" onClick={submitExtend} disabled={extendBusy}>
              {extendBusy ? 'Saving…' : 'Extend Lease'}
            </button>
          </>}
        >
          {extendErr && <div className="form-error" style={{ color: 'var(--danger, #c0392b)', marginBottom: 12 }}>{extendErr}</div>}
          <Field label="New Expiry Date">
            <input type="date" value={extendDate} onChange={(e) => setExtendDate(e.target.value)} autoFocus />
          </Field>
        </Modal>
      )}

      {confirmAction && (
        <Confirm
          message={
            <>
              <div style={{ marginBottom: 12 }}>
                <strong>{confirmAction.label}</strong> the lease for "<strong>{confirmAction.lease.adminName || confirmAction.lease.adminUserId}</strong>"?
              </div>
              {confirmAction.remarks && <div style={{ color: 'var(--text-muted, #666)', fontSize: 14 }}>{confirmAction.remarks}</div>}
            </>
          }
          onCancel={() => setConfirmAction(null)}
          onConfirm={() => { const ca = confirmAction; setConfirmAction(null); runAction(ca.lease, ca.action); }}
        />
      )}

      {viewingAccount && (
        <Modal
          title={`Admin Account — ${viewingAccount.name || viewingAccount.userId}`}
          onClose={() => setViewingAccount(null)}
          footer={editing ? <>
            <button className="btn ghost" onClick={() => { setEditing(false); setShowPassword(false); setShowViewPassword(false); setEditErr(null); }} disabled={editBusy}>Cancel</button>
            <button className="btn primary" onClick={submitEditAdmin} disabled={editBusy}>
              {editBusy ? 'Saving…' : 'Save Credentials'}
            </button>
          </> : <>
            <button className="btn ghost" onClick={() => setViewingAccount(null)}>Close</button>
            <button className="btn primary" onClick={() => setEditing(true)}>Edit Credentials</button>
          </>}
        >
          {editErr && <div className="form-error" style={{ color: 'var(--danger, #c0392b)', marginBottom: 12 }}>{editErr}</div>}
          
          {!editing ? (
            <>
              <div className="section-title">Account Details</div>
              <div className="form-grid">
                <Field label="User ID" span={2}>
                  <input value={viewingAccount.userId} readOnly />
                </Field>
                <Field label="Name" span={2}>
                  <input value={viewingAccount.name} readOnly />
                </Field>
                <Field label="Password" span={2}>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <input
                      type={showViewPassword ? 'text' : 'password'}
                      value={viewingAccount.password || '(not available)'}
                      readOnly
                      style={{ flex: 1 }}
                    />
                    <button
                      className="btn ghost sm"
                      type="button"
                      onClick={() => {
                        console.log('Toggle clicked. Current state:', showViewPassword);
                        setShowViewPassword(!showViewPassword);
                      }}
                      style={{ minWidth: 80 }}
                    >
                      {showViewPassword ? 'Hide' : 'Show'}
                    </button>
                  </div>
                </Field>
              </div>
            </>
          ) : (
            <>
              <div className="section-title">Edit Admin Credentials</div>
              <div className="form-grid">
                <Field label="User ID" span={2}>
                  <input value={editForm.userId} onChange={(e) => setEditForm({ ...editForm, userId: e.target.value })} autoFocus />
                </Field>
                <Field label="Name" span={2}>
                  <input value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} />
                </Field>
                <Field label="New Password (leave blank to keep current)" span={2}>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={editForm.password}
                      onChange={(e) => setEditForm({ ...editForm, password: e.target.value })}
                      placeholder="Leave blank to keep current password"
                      style={{ flex: 1 }}
                    />
                    <button
                      className="btn ghost sm"
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      style={{ minWidth: 80 }}
                    >
                      {showPassword ? 'Hide' : 'Show'}
                    </button>
                  </div>
                </Field>
              </div>
            </>
          )}
        </Modal>
      )}
    </div>
  );
}
