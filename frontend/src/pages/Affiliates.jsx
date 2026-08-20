import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, StatusBadge, Empty, Field, Confirm } from '../components/ui';
import { COMMISSION_BASES } from '../data/seed';
import { affiliateLedger, inr } from '../data/calc';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

const blank = {
  name: '', contact: '', altContact: '', email: '',
  commissionPct: 0, commissionBase: 'Settlement Amount',
  userId: '', password: '', status: 'Active',
};

export default function Affiliates() {
  const { db, add, update, remove } = useStore();
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);
  const [showPwd, setShowPwd] = useState(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const openNew = () => { setForm(blank); setShowPwd(false); setEditing('new'); };
  const openEdit = (a) => { setForm({ ...blank, ...a, password: '' }); setShowPwd(false); setEditing(a); };

  const save = () => {
    if (!form.name.trim()) { alert('Affiliate name is required.'); return; }
    if (!form.userId.trim()) { alert('Login User ID is required.'); return; }
    if (editing === 'new' && !form.password.trim()) { alert('Login Password is required.'); return; }
    
    // Check for duplicate userId across all user types
    const userIdLower = form.userId.trim().toLowerCase();
    const isDuplicate = [
      ...db.companies,
      ...db.merchants,
      ...db.affiliates.filter(a => editing === 'new' || a.id !== editing.id)
    ].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);
    
    if (isDuplicate) {
      alert('This User ID is already taken. Please choose a different User ID.');
      return;
    }
    
    const payload = { ...form, commissionPct: Number(form.commissionPct) || 0 };
    if (editing !== 'new' && !payload.password.trim()) delete payload.password;
    if (editing === 'new') add('affiliates', payload);
    else update('affiliates', editing.id, payload);
    setEditing(null);
  };

  // Paginated data
  const totalAffiliates = db.affiliates.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedAffiliates = db.affiliates.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1); // Reset to first page
  };

  return (
    <div>
      <PageHead
        title="Affiliate Management"
        sub="Create affiliates with commission terms and login credentials."
        actions={<button className="btn primary" onClick={openNew}>+ Add Affiliate</button>}
      />

      <div className="panel">
        {db.affiliates.length === 0 ? (
          <div className="panel-body"><Empty icon="⬡" text="No affiliates yet." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Affiliate</th>
                  <th>Contact</th>
                  <th>Commission</th>
                  <th>Merchants</th>
                  <th className="num">Earned</th>
                  <th className="num">Balance</th>
                  <th>Login</th>
                  <th>Status</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedAffiliates.map((a) => {
                  const merchants = db.merchants.filter((m) => m.affiliateId === a.id).length;
                  const led = affiliateLedger(a.id, db);
                  return (
                    <tr key={a.id}>
                      <td className="bold">{a.name}</td>
                      <td className="mono">{a.contact}</td>
                      <td className="mono">{a.commissionPct}% · {a.commissionBase === 'Transaction Amount' ? 'Txn' : 'Settle'}</td>
                      <td>{merchants}</td>
                      <td className="num mono">{inr(led.earned)}</td>
                      <td className="num mono">{inr(led.balance)}</td>
                      <td className="mono">{a.userId}</td>
                      <td><StatusBadge status={a.status} /></td>
                      <td className="center">
                        <div className="row-actions" style={{ justifyContent: 'center' }}>
                          <button className="btn sm ghost" onClick={() => openEdit(a)}>Edit</button>
                          <button className="btn sm danger" onClick={() => setConfirmId(a.id)}>Delete</button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalAffiliates > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalAffiliates}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={handleItemsPerPageChange}
              />
            )}
          </>
        )}
      </div>

      {editing && (
        <Modal
          title={editing === 'new' ? 'Add Affiliate' : `Edit — ${editing.name}`}
          onClose={() => setEditing(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setEditing(null)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Affiliate</button>
          </>}
        >
          <div className="section-title">Affiliate Details</div>
          <div className="form-grid">
            <Field label="Affiliate Name" span={2}>
              <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} autoFocus />
            </Field>
            <Field label="Contact Details"><input value={form.contact} onChange={(e) => setForm({ ...form, contact: e.target.value })} /></Field>
            <Field label="Alternative Contact"><input value={form.altContact} onChange={(e) => setForm({ ...form, altContact: e.target.value })} /></Field>
            <Field label="Email"><input value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></Field>
            <Field label="Status">
              <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                <option>Active</option><option>Inactive</option>
              </select>
            </Field>
          </div>

          <div className="section-title">Commission</div>
          <div className="form-grid">
            <Field label="Commission Percentage">
              <input type="number" step="0.01" value={form.commissionPct} onChange={(e) => setForm({ ...form, commissionPct: e.target.value })} />
            </Field>
            <Field label="Commission Base">
              <select value={form.commissionBase} onChange={(e) => setForm({ ...form, commissionBase: e.target.value })}>
                {COMMISSION_BASES.map((b) => <option key={b}>{b}</option>)}
              </select>
            </Field>
          </div>

          <div className="section-title">Login Credentials</div>
          <div className="form-grid">
            <Field label="User ID"><input value={form.userId} onChange={(e) => setForm({ ...form, userId: e.target.value })} /></Field>
            <Field label="Password" help={editing === 'new' ? 'Set the portal password.' : 'Leave blank to keep current password.'}>
              <div style={{ position: 'relative' }}>
                <input
                  type={showPwd ? 'text' : 'password'}
                  value={form.password}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                  placeholder={editing === 'new' ? 'Enter password' : 'Leave blank to keep current'}
                  style={{ width: '100%', paddingRight: 60 }}
                />
                <button type="button" onClick={() => setShowPwd((s) => !s)}
                  style={{ position: 'absolute', top: '50%', right: 8, transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', fontSize: 11, fontWeight: 700, color: 'var(--accent, #1f9d55)' }}>
                  {showPwd ? 'HIDE' : 'SHOW'}
                </button>
              </div>
            </Field>
          </div>
        </Modal>
      )}

      {confirmId && (
        <Confirm
          message="Delete this affiliate?"
          onCancel={() => setConfirmId(null)}
          onConfirm={() => { remove('affiliates', confirmId); setConfirmId(null); }}
        />
      )}
    </div>
  );
}
