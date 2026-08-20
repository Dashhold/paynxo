import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStore, uid } from '../data/store';
import { PageHead, StatusBadge, Empty, Field, Confirm } from '../components/ui';
import { COMMISSION_BASES } from '../data/seed';
import Modal from '../components/Modal';
import BankForm from '../components/BankForm';
import Pagination from '../components/Pagination';

const blankMerchant = {
  name: '', contact: '', altContact: '', email: '',
  companyId: '', affiliateId: null,
  commissionPct: 0, commissionBase: 'Settlement Amount',
  userId: '', password: '', status: 'Active', banks: [],
};

const blankBank = {
  bankName: '', accountName: '', accountNumber: '', ifsc: '',
  netbankingLink: '', username: '', loginPassword: '', txnPassword: '',
  customerId: '', mobile: '', email: '',
  mobileBanking: 'No', mobileLoginId: '', mpin: '',
  atmCards: [], custom: [],
};

export default function Merchants({ startNew = false }) {
  const { db, add, update, remove } = useStore();
  const navigate = useNavigate();
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState(blankMerchant);
  const [confirmId, setConfirmId] = useState(null);
  const [search, setSearch] = useState('');
  const [bankEdit, setBankEdit] = useState(null); // {index} or 'new'
  const [bankForm, setBankForm] = useState(blankBank);
  const [showPwd, setShowPwd] = useState(false);
  const [merchantType, setMerchantType] = useState('direct');

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  useEffect(() => {
    if (startNew) { setForm(blankMerchant); setEditing('new'); }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startNew]);

  const closeModal = () => {
    setEditing(null);
    if (startNew) navigate('/merchants');
  };

  const openNew = () => { setForm(blankMerchant); setShowPwd(false); setMerchantType('direct'); setEditing('new'); };
  const openEdit = (m) => { setForm(JSON.parse(JSON.stringify({ ...blankMerchant, ...m, password: '' }))); setShowPwd(false); setMerchantType(m.affiliateId ? 'affiliate' : 'direct'); setEditing(m); };

  const companyName = (id) => db.companies.find((c) => c.id === id)?.name || '—';
  const affiliateName = (id) => db.affiliates.find((a) => a.id === id)?.name || '';

  // Validation: a merchant belongs to only one company. The select enforces single
  // company. When changing to a different company while already assigned, warn.
  const changeCompany = (newId) => {
    if (editing !== 'new' && form.companyId && newId && newId !== form.companyId) {
      const ok = window.confirm(
        'Merchant is already assigned to another company. Remove existing company assignment before assigning a new company.\n\nProceed and reassign?'
      );
      if (!ok) return;
    }
    setForm({ ...form, companyId: newId });
  };

  const save = () => {
    if (!form.name.trim()) { alert('Merchant name is required.'); return; }
    if (!form.companyId) { alert('Please assign the merchant to a company.'); return; }
    if (merchantType === 'affiliate' && !form.affiliateId) { alert('Please select an affiliate, or choose "Direct".'); return; }
    if (!form.userId.trim()) { alert('Login User ID is required.'); return; }
    if (editing === 'new' && !form.password.trim()) { alert('Login Password is required.'); return; }
    
    // Check for duplicate userId across all user types
    const userIdLower = form.userId.trim().toLowerCase();
    const isDuplicate = [
      ...db.companies,
      ...db.merchants.filter(m => editing === 'new' || m.id !== editing.id),
      ...db.affiliates
    ].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);
    
    if (isDuplicate) {
      alert('This User ID is already taken. Please choose a different User ID.');
      return;
    }
    
    const payload = {
      ...form,
      affiliateId: merchantType === 'affiliate' ? form.affiliateId : null,
      commissionPct: Number(form.commissionPct) || 0,
    };
    if (payload.affiliateId) payload.commissionPct = 0; // affiliate receives commission
    if (editing !== 'new' && !payload.password.trim()) delete payload.password;
    if (editing === 'new') add('merchants', payload);
    else update('merchants', editing.id, payload);
    closeModal();
  };

  // ---- Bank sub-management (inline within merchant form) ----
  const openNewBank = () => { setBankForm({ ...blankBank, id: uid('bk') }); setBankEdit('new'); };
  const openEditBank = (bank, idx) => { setBankForm(JSON.parse(JSON.stringify({ ...blankBank, ...bank }))); setBankEdit(idx); };

  const saveBank = () => {
    if (!bankForm.bankName.trim() || !bankForm.accountNumber.trim()) { alert('Bank Name and Account Number are required.'); return; }
    setForm((f) => {
      const banks = [...(f.banks || [])];
      if (bankEdit === 'new') banks.push(bankForm);
      else banks[bankEdit] = bankForm;
      return { ...f, banks };
    });
    setBankEdit(null);
  };

  const deleteBank = (idx) => {
    setForm((f) => ({ ...f, banks: f.banks.filter((_, i) => i !== idx) }));
  };

  const filtered = db.merchants.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));
  const isAffiliate = merchantType === 'affiliate';

  // Paginated data
  const totalMerchants = filtered.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedMerchants = filtered.slice(startIndex, endIndex);

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

  const changeMerchantType = (t) => {
    setMerchantType(t);
    if (t === 'direct') setForm((f) => ({ ...f, affiliateId: null }));
    else setForm((f) => ({ ...f, affiliateId: f.affiliateId || '' }));
  };

  return (
    <div>
      <PageHead
        title="Merchant Management"
        sub="Create merchants, link to a company, set commission or affiliate, and manage bank accounts."
        actions={<button className="btn primary" onClick={openNew}>+ Add Merchant</button>}
      />

      <div className="toolbar">
        <input className="search" placeholder="Search merchants…" value={search} onChange={(e) => handleSearchChange(e.target.value)} />
        <span className="muted">{totalMerchants} of {db.merchants.length} merchants</span>
      </div>

      <div className="panel">
        {filtered.length === 0 ? (
          <div className="panel-body"><Empty icon="◈" text="No merchants found." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Merchant</th>
                  <th>Company</th>
                  <th>Type</th>
                  <th>Commission</th>
                  <th>Banks</th>
                  <th>Login</th>
                  <th>Status</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedMerchants.map((m) => (
                  <tr key={m.id}>
                    <td className="bold">{m.name}<br /><span className="muted mono">{m.contact}</span></td>
                    <td>{companyName(m.companyId)}</td>
                    <td>
                      {m.affiliateId
                        ? <span className="badge">Affiliate · {affiliateName(m.affiliateId)}</span>
                        : <span className="badge muted">Direct</span>}
                    </td>
                    <td className="mono">
                      {m.affiliateId ? <span className="muted">via affiliate</span> : `${m.commissionPct}% · ${m.commissionBase === 'Transaction Amount' ? 'Txn' : 'Settle'}`}
                    </td>
                    <td>{(m.banks || []).length}</td>
                    <td className="mono">{m.userId}</td>
                    <td><StatusBadge status={m.status} /></td>
                    <td className="center">
                      <div className="row-actions" style={{ justifyContent: 'center' }}>
                        <button className="btn sm ghost" onClick={() => openEdit(m)}>Edit</button>
                        <button className="btn sm danger" onClick={() => setConfirmId(m.id)}>Delete</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {totalMerchants > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalMerchants}
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
          title={editing === 'new' ? 'Add Merchant' : `Edit — ${editing.name}`}
          onClose={closeModal}
          footer={<>
            <button className="btn ghost" onClick={closeModal}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Merchant</button>
          </>}
        >
          <div className="section-title">Merchant Details</div>
          <div className="form-grid">
            <Field label="Merchant Name" span={2}>
              <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} autoFocus />
            </Field>
            <Field label="Contact Details">
              <input value={form.contact} onChange={(e) => setForm({ ...form, contact: e.target.value })} />
            </Field>
            <Field label="Alternative Contact Details">
              <input value={form.altContact} onChange={(e) => setForm({ ...form, altContact: e.target.value })} />
            </Field>
            <Field label="Email">
              <input value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            </Field>
            <Field label="Status">
              <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                <option>Active</option><option>Inactive</option>
              </select>
            </Field>
          </div>

          <div className="section-title">Assignment</div>
          <div className="form-grid">
            <Field label="Company (only one allowed)">
              <select value={form.companyId} onChange={(e) => changeCompany(e.target.value)}>
                <option value="">— Select Company —</option>
                {db.companies.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </Field>
            <Field label="Merchant Type">
              <select value={merchantType} onChange={(e) => changeMerchantType(e.target.value)}>
                <option value="direct">Direct (under Admin)</option>
                <option value="affiliate">Under Affiliate</option>
              </select>
            </Field>
            {isAffiliate && (
              <Field label="Select Affiliate">
                <select value={form.affiliateId || ''} onChange={(e) => setForm({ ...form, affiliateId: e.target.value })}>
                  <option value="">— Select Affiliate —</option>
                  {db.affiliates.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
                </select>
                {db.affiliates.length === 0 && <p className="help" style={{ color: 'var(--danger, #c0392b)' }}>No affiliates exist yet. Create an affiliate first, or choose Direct.</p>}
              </Field>
            )}
          </div>

          {/* Merchant commission hidden when under an affiliate */}
          {!isAffiliate ? (
            <>
              <div className="section-title">Merchant Commission</div>
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
            </>
          ) : (
            <div className="calc-box" style={{ margin: '14px 0' }}>
              <div className="calc-row"><span className="lbl">Commission</span><span>Paid to affiliate — merchant commission hidden</span></div>
            </div>
          )}

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

          <div className="section-title">Bank Accounts</div>
          <div className="btn-row" style={{ marginBottom: 12 }}>
            <button className="btn sm primary" onClick={openNewBank}>+ Add Bank</button>
          </div>
          {(form.banks || []).length === 0 && <p className="help">No bank accounts added.</p>}
          {(form.banks || []).map((b, idx) => (
            <div key={b.id || idx} className="bank-card">
              <div className="bc-head">
                <span className="bank-name">{b.bankName} · <span className="mono">{b.accountNumber}</span></span>
                <div className="row-actions">
                  <button className="btn sm ghost" onClick={() => openEditBank(b, idx)}>Edit</button>
                  <button className="btn sm danger" onClick={() => deleteBank(idx)}>Delete</button>
                </div>
              </div>
              <div className="muted" style={{ fontSize: 12 }}>
                {b.accountName} · IFSC {b.ifsc || '—'} · Cust ID {b.customerId || '—'}
              </div>
            </div>
          ))}
        </Modal>
      )}

      {bankEdit !== null && (
        <Modal
          wide
          title={bankEdit === 'new' ? 'Add Bank Account' : 'Edit Bank Account'}
          onClose={() => setBankEdit(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setBankEdit(null)}>Cancel</button>
            <button className="btn primary" onClick={saveBank}>Save Bank</button>
          </>}
        >
          <BankForm value={bankForm} onChange={setBankForm} />
        </Modal>
      )}

      {confirmId && (
        <Confirm
          message="Delete this merchant? Their transactions will remain in records."
          onCancel={() => setConfirmId(null)}
          onConfirm={() => { remove('merchants', confirmId); setConfirmId(null); }}
        />
      )}
    </div>
  );
}
