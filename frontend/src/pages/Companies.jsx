import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore } from '../data/store';
import { PageHead, StatusBadge, Empty, Field, Confirm } from '../components/ui';
import { CHARGE_BEARERS } from '../data/seed';
import { companyLedger, inr } from '../data/calc';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

const blank = {
  name: '', contactPerson: '', contactNumber: '', whatsapp: '', telegram: '',
  email: '', altContactPerson: '', altContactNumber: '', address: '', status: 'Active',
  userId: '', password: '', gateways: [],
};

export default function Companies({ startNew = false }) {
  const { db, add, update, remove } = useStore();
  const navigate = useNavigate();
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);
  const [search, setSearch] = useState('');
  const [showPwd, setShowPwd] = useState(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  // When reached via "Add New Company", open the create modal immediately.
  useEffect(() => {
    if (startNew) { setForm(blank); setEditing('new'); }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startNew]);

  const closeModal = () => {
    setEditing(null);
    if (startNew) navigate('/companies');
  };

  const openNew = () => { setForm(blank); setShowPwd(false); setEditing('new'); };
  const openEdit = (c) => { setForm({ ...blank, ...JSON.parse(JSON.stringify(c)), password: '' }); setShowPwd(false); setEditing(c); };

  const save = () => {
    if (!form.name.trim()) { alert('Company name is required.'); return; }
    if (!form.userId.trim()) { alert('Login User ID is required.'); return; }
    if (editing === 'new' && !form.password.trim()) { alert('Login Password is required.'); return; }
    
    // Check for duplicate userId across all user types
    const userIdLower = form.userId.trim().toLowerCase();
    const isDuplicate = [
      ...db.companies.filter(c => editing === 'new' || c.id !== editing.id),
      ...db.merchants,
      ...db.affiliates
    ].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);
    
    if (isDuplicate) {
      alert('This User ID is already taken. Please choose a different User ID.');
      return;
    }
    
    // Coerce numeric gateway commission (inputs yield strings) so the API
    // receives JSON numbers. Omit an empty password on edit (keeps current).
    const payload = {
      ...form,
      gateways: (form.gateways || []).map((g) => ({ ...g, commission: Number(g.commission) || 0 })),
    };
    if (editing !== 'new' && !payload.password.trim()) delete payload.password;
    if (editing === 'new') add('companies', payload);
    else update('companies', editing.id, payload);
    closeModal();
  };

  const toggleGateway = (gwId) => {
    setForm((f) => {
      const exists = f.gateways.find((g) => g.gatewayId === gwId);
      if (exists) return { ...f, gateways: f.gateways.filter((g) => g.gatewayId !== gwId) };
      return { ...f, gateways: [...f.gateways, { gatewayId: gwId, commission: 0, chargeBearer: 'Admin' }] };
    });
  };

  const setGw = (gwId, patch) => {
    setForm((f) => ({
      ...f,
      gateways: f.gateways.map((g) => (g.gatewayId === gwId ? { ...g, ...patch } : g)),
    }));
  };

  const filtered = db.companies.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()));

  // Paginated data
  const totalCompanies = filtered.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedCompanies = filtered.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1);
  };

  const handleSearchChange = (value) => {
    setSearch(value);
    setCurrentPage(1); // Reset to first page on search
  };

  return (
    <div>
      <PageHead
        title="Company Management"
        sub="Create companies, assign payment gateways, and issue login credentials."
        actions={<button className="btn primary" onClick={openNew}>+ Add Company</button>}
      />

      <div className="toolbar">
        <input className="search" placeholder="Search companies…" value={search} onChange={(e) => handleSearchChange(e.target.value)} />
        <span className="muted">{totalCompanies} of {db.companies.length} companies</span>
      </div>

      <div className="panel">
        {filtered.length === 0 ? (
          <div className="panel-body"><Empty icon="▣" text="No companies found." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Company</th>
                  <th>Contact</th>
                  <th>Gateways</th>
                  <th>Merchants</th>
                  <th className="num">Balance</th>
                  <th>Login</th>
                  <th>Status</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedCompanies.map((c) => {
                  const merchants = db.merchants.filter((m) => m.companyId === c.id).length;
                  const led = companyLedger(c.id, db);
                  return (
                    <tr key={c.id}>
                      <td className="bold">{c.name}</td>
                      <td>{c.contactPerson}<br /><span className="muted mono">{c.contactNumber}</span></td>
                      <td>{(c.gateways || []).length}</td>
                      <td>{merchants}</td>
                      <td className="num mono">{inr(led.balance)}</td>
                      <td className="mono">{c.userId}</td>
                      <td><StatusBadge status={c.status} /></td>
                      <td className="center">
                        <div className="row-actions" style={{ justifyContent: 'center' }}>
                          <button className="btn sm ghost" onClick={() => openEdit(c)}>Edit</button>
                          <button className="btn sm danger" onClick={() => setConfirmId(c.id)}>Delete</button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalCompanies > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalCompanies}
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
          wide
          title={editing === 'new' ? 'Add Company' : `Edit — ${editing.name}`}
          onClose={closeModal}
          footer={<>
            <button className="btn ghost" onClick={closeModal}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Company</button>
          </>}
        >
          <div className="section-title">Company Details</div>
          <div className="form-grid">
            <Field label="Company Name" span={2}>
              <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} autoFocus />
            </Field>
            <Field label="Contact Person Name">
              <input value={form.contactPerson} onChange={(e) => setForm({ ...form, contactPerson: e.target.value })} />
            </Field>
            <Field label="Contact Number">
              <input value={form.contactNumber} onChange={(e) => setForm({ ...form, contactNumber: e.target.value })} />
            </Field>
            <Field label="WhatsApp">
              <input value={form.whatsapp} onChange={(e) => setForm({ ...form, whatsapp: e.target.value })} />
            </Field>
            <Field label="Telegram">
              <input value={form.telegram} onChange={(e) => setForm({ ...form, telegram: e.target.value })} />
            </Field>
            <Field label="Email">
              <input value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            </Field>
            <Field label="Alternative Contact Person">
              <input value={form.altContactPerson} onChange={(e) => setForm({ ...form, altContactPerson: e.target.value })} />
            </Field>
            <Field label="Alternative Contact Number">
              <input value={form.altContactNumber} onChange={(e) => setForm({ ...form, altContactNumber: e.target.value })} />
            </Field>
            <Field label="Status">
              <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                <option>Active</option><option>Inactive</option>
              </select>
            </Field>
            <Field label="Address" span="full">
              <textarea value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
            </Field>
          </div>

          <div className="section-title">Login Credentials</div>
          <div className="form-grid">
            <Field label="User ID" help="Used by the company to log in.">
              <input value={form.userId} onChange={(e) => setForm({ ...form, userId: e.target.value })} />
            </Field>
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

          <div className="section-title">Payment Gateway Assignment</div>
          <p className="help" style={{ marginBottom: 12 }}>Tick gateways to assign, then set commission % and who bears the charges.</p>
          {db.gateways.map((g) => {
            const assigned = form.gateways.find((x) => x.gatewayId === g.id);
            return (
              <div key={g.id} className="bank-card">
                <div className="checkbox-line" style={{ marginBottom: assigned ? 12 : 0 }}>
                  <input type="checkbox" checked={!!assigned} onChange={() => toggleGateway(g.id)} id={`gw-${g.id}`} />
                  <label htmlFor={`gw-${g.id}`} className="bold" style={{ cursor: 'pointer' }}>{g.name}</label>
                  {g.status === 'Inactive' && <span className="badge muted">Inactive</span>}
                </div>
                {assigned && (
                  <div className="form-grid">
                    <Field label="Company Commission %">
                      <input type="number" step="0.01" value={assigned.commission}
                        onChange={(e) => setGw(g.id, { commission: e.target.value })} />
                    </Field>
                    <Field label="Charge Bearer">
                      <select value={assigned.chargeBearer} onChange={(e) => setGw(g.id, { chargeBearer: e.target.value })}>
                        {CHARGE_BEARERS.map((b) => <option key={b}>{b}</option>)}
                      </select>
                    </Field>
                  </div>
                )}
              </div>
            );
          })}
        </Modal>
      )}

      {confirmId && (
        <Confirm
          message="Delete this company? Linked merchants will remain but lose their company assignment."
          onCancel={() => setConfirmId(null)}
          onConfirm={() => { remove('companies', confirmId); setConfirmId(null); }}
        />
      )}
    </div>
  );
}
