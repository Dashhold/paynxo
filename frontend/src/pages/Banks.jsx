import { useState } from 'react';
import { useStore, uid } from '../data/store';
import { PageHead, Empty } from '../components/ui';
import Modal from '../components/Modal';
import BankForm from '../components/BankForm';
import { BankView } from './MerchantBankAccounts';
import Pagination from '../components/Pagination';

const blankBank = {
  bankName: '', accountName: '', accountNumber: '', ifsc: '',
  netbankingLink: '', username: '', loginPassword: '', txnPassword: '',
  customerId: '', mobile: '', email: '',
  mobileBanking: 'No', mobileLoginId: '', mpin: '',
  atmCards: [], custom: [],
};

export default function Banks() {
  const { db, update } = useStore();
  const [view, setView] = useState(null); // bank record for detail
  const [edit, setEdit] = useState(null); // {merchantId, index|'new'}
  const [form, setForm] = useState(blankBank);
  const [reveal, setReveal] = useState(false);
  const [addFor, setAddFor] = useState('');

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  // Flatten all banks across merchants
  const rows = [];
  db.merchants.forEach((m) => {
    (m.banks || []).forEach((b, idx) => {
      rows.push({ ...b, merchantId: m.id, merchantName: m.name, index: idx });
    });
  });

  // Paginated data
  const totalBanks = rows.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedRows = rows.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const openEdit = (row) => {
    setForm(JSON.parse(JSON.stringify({ ...blankBank, ...row })));
    setEdit({ merchantId: row.merchantId, index: row.index });
  };

  const openAdd = () => {
    if (!addFor) { alert('Select a merchant to add a bank to.'); return; }
    setForm({ ...blankBank, id: uid('bk') });
    setEdit({ merchantId: addFor, index: 'new' });
  };

  const save = () => {
    if (!form.bankName.trim() || !form.accountNumber.trim()) { alert('Bank Name and Account Number are required.'); return; }
    const m = db.merchants.find((x) => x.id === edit.merchantId);
    const banks = [...(m.banks || [])];
    const clean = { ...form };
    delete clean.merchantId; delete clean.merchantName; delete clean.index;
    if (edit.index === 'new') banks.push(clean);
    else banks[edit.index] = clean;
    update('merchants', m.id, { banks });
    setEdit(null);
  };

  const del = (row) => {
    if (!window.confirm('Delete this bank account?')) return;
    const m = db.merchants.find((x) => x.id === row.merchantId);
    const banks = (m.banks || []).filter((_, i) => i !== row.index);
    update('merchants', m.id, { banks });
  };

  return (
    <div>
      <PageHead
        title="Banks"
        sub="All merchant bank accounts with netbanking, mobile banking and ATM cards."
        actions={
          <div className="btn-row">
            <select value={addFor} onChange={(e) => setAddFor(e.target.value)} style={{ width: 200 }}>
              <option value="">— Select Merchant —</option>
              {db.merchants.map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
            </select>
            <button className="btn primary" onClick={openAdd}>+ Add Bank</button>
          </div>
        }
      />

      <div className="panel">
        {rows.length === 0 ? (
          <div className="panel-body"><Empty icon="▤" text="No bank accounts yet." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Bank</th>
                  <th>Merchant</th>
                  <th>Account No.</th>
                  <th>IFSC</th>
                  <th>Mobile Banking</th>
                  <th>Cards</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedRows.map((r) => (
                  <tr key={r.merchantId + r.index}>
                    <td className="bold">{r.bankName}<br /><span className="muted">{r.accountName}</span></td>
                    <td>{r.merchantName}</td>
                    <td className="mono">{r.accountNumber}</td>
                    <td className="mono">{r.ifsc || '—'}</td>
                    <td>{r.mobileBanking === 'Yes' ? <span className="badge solid">Enabled</span> : <span className="badge muted">No</span>}</td>
                    <td>{(r.atmCards || []).length}</td>
                    <td className="center">
                      <div className="row-actions" style={{ justifyContent: 'center' }}>
                        <button className="btn sm" onClick={() => { setView(r); setReveal(false); }}>View</button>
                        <button className="btn sm ghost" onClick={() => openEdit(r)}>Edit</button>
                        <button className="btn sm danger" onClick={() => del(r)}>Delete</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {totalBanks > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalBanks}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={(newSize) => { setItemsPerPage(newSize); setCurrentPage(1); }}
              />
            )}
          </>
        )}
      </div>

      {view && (
        <BankView bank={view} reveal={reveal} setReveal={setReveal} onClose={() => setView(null)} />
      )}

      {edit && (
        <Modal
          wide
          title={edit.index === 'new' ? 'Add Bank Account' : 'Edit Bank Account'}
          onClose={() => setEdit(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setEdit(null)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Bank</button>
          </>}
        >
          <BankForm value={form} onChange={setForm} />
        </Modal>
      )}
    </div>
  );
}
