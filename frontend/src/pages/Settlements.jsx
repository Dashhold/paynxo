import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Empty, Field, Confirm } from '../components/ui';
import { PAYMENT_MODES } from '../data/seed';
import { companyLedger, inr } from '../data/calc';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

const blank = {
  companyId: '', date: new Date().toISOString().slice(0, 10),
  amount: '', paymentMode: 'Bank Transfer', refNumber: '', remarks: '',
};

export default function Settlements() {
  const { db, add, remove } = useStore();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const nameOf = (id) => db.companies.find((c) => c.id === id)?.name || '—';

  const openNew = () => { setForm(blank); setOpen(true); };

  const save = () => {
    if (!form.companyId) { alert('Select a company.'); return; }
    if (!form.amount) { alert('Enter an amount.'); return; }
    add('settlements', { ...form, amount: Number(form.amount) });
    setOpen(false);
    setCurrentPage(1); // Reset to first page after adding
  };

  const selectedLedger = form.companyId ? companyLedger(form.companyId, db) : null;
  const rows = db.settlements.slice().reverse();

  // Paginated data
  const totalSettlements = rows.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedRows = rows.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1);
  };

  return (
    <div>
      <PageHead
        title="Settlement Management"
        sub="Record payments made to companies. Ledger balances update automatically."
        actions={<button className="btn primary" onClick={openNew}>+ Record Payment</button>}
      />

      <div className="panel">
        {rows.length === 0 ? (
          <div className="panel-body"><Empty icon="⇣" text="No settlement payments recorded." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Company</th>
                  <th className="num">Amount</th>
                  <th>Payment Mode</th>
                  <th>Reference No.</th>
                  <th>Remarks</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedRows.map((s) => (
                  <tr key={s.id}>
                    <td className="nowrap">{s.date}</td>
                    <td className="bold">{nameOf(s.companyId)}</td>
                    <td className="num mono">{inr(s.amount)}</td>
                    <td><span className="badge muted">{s.paymentMode}</span></td>
                    <td className="mono">{s.refNumber || '—'}</td>
                    <td className="muted">{s.remarks || '—'}</td>
                    <td className="center">
                      <button className="btn sm danger" onClick={() => setConfirmId(s.id)}>Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {totalSettlements > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalSettlements}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={handleItemsPerPageChange}
              />
            )}
          </>
        )}
      </div>

      {open && (
        <Modal
          title="Record Settlement Payment"
          onClose={() => setOpen(false)}
          footer={<>
            <button className="btn ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Payment</button>
          </>}
        >
          <div className="form-grid">
            <Field label="Company" span={2}>
              <select value={form.companyId} onChange={(e) => setForm({ ...form, companyId: e.target.value })}>
                <option value="">— Select Company —</option>
                {db.companies.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </Field>
            <Field label="Date"><input type="date" value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} /></Field>
            <Field label="Amount"><input type="number" value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} /></Field>
            <Field label="Payment Mode">
              <select value={form.paymentMode} onChange={(e) => setForm({ ...form, paymentMode: e.target.value })}>
                {PAYMENT_MODES.map((p) => <option key={p}>{p}</option>)}
              </select>
            </Field>
            <Field label="Transaction Reference Number"><input value={form.refNumber} onChange={(e) => setForm({ ...form, refNumber: e.target.value })} /></Field>
            <Field label="Remarks" span={2}><textarea value={form.remarks} onChange={(e) => setForm({ ...form, remarks: e.target.value })} /></Field>
          </div>

          {selectedLedger && (
            <div className="calc-box">
              <div className="calc-row"><span className="lbl">Total Receivable (Company Net Income)</span><span className="mono">{inr(selectedLedger.receivable)}</span></div>
              <div className="calc-row"><span className="lbl">Already Paid</span><span className="mono">{inr(selectedLedger.paid)}</span></div>
              <div className="calc-row total"><span className="lbl">Current Outstanding</span><span className="mono">{inr(selectedLedger.balance)}</span></div>
            </div>
          )}
        </Modal>
      )}

      {confirmId && (
        <Confirm message="Delete this settlement payment?" onCancel={() => setConfirmId(null)} onConfirm={() => { remove('settlements', confirmId); setConfirmId(null); }} />
      )}
    </div>
  );
}
