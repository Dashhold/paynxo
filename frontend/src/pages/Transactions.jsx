import { useState, useMemo } from 'react';
import { useStore } from '../data/store';
import { PageHead, Empty, Field, Confirm } from '../components/ui';
import { calcTransaction, inr } from '../data/calc';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

const blank = {
  companyId: '', merchantId: '', gatewayId: '',
  date: new Date().toISOString().slice(0, 10),
  txnAmount: '', settlementAmount: '', txnCharges: '', otherCharges: '', remarks: '',
};

export default function Transactions() {
  const { db, add, remove } = useStore();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);
  const [filterCompany, setFilterCompany] = useState('');

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const nameOf = (coll, id) => db[coll].find((x) => x.id === id)?.name || '—';

  // Merchants restricted to selected company
  const availableMerchants = db.merchants.filter((m) => m.companyId === form.companyId);
  // Gateways restricted to those assigned to the company
  const company = db.companies.find((c) => c.id === form.companyId);
  const availableGateways = (company?.gateways || [])
    .map((g) => db.gateways.find((x) => x.id === g.gatewayId))
    .filter(Boolean);

  const preview = useMemo(() => {
    if (!form.companyId || !form.merchantId || !form.gatewayId) return null;
    const merchant = db.merchants.find((m) => m.id === form.merchantId);
    const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
    return calcTransaction(form, { company, merchant, affiliate });
  }, [form, db, company]);

  const openNew = () => { setForm(blank); setOpen(true); };

  const save = () => {
    if (!form.companyId || !form.merchantId || !form.gatewayId) { alert('Select company, merchant and gateway.'); return; }
    if (!form.txnAmount || !form.settlementAmount) { alert('Enter transaction and settlement amounts.'); return; }
    add('transactions', {
      ...form,
      txnAmount: Number(form.txnAmount),
      settlementAmount: Number(form.settlementAmount),
      txnCharges: Number(form.txnCharges || 0),
      otherCharges: Number(form.otherCharges || 0),
    });
    setOpen(false);
    setCurrentPage(1); // Reset to first page after adding
  };

  const rows = db.transactions
    .filter((t) => !filterCompany || t.companyId === filterCompany)
    .slice().reverse();

  // Paginated data
  const totalTransactions = rows.length;
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

  const handleFilterChange = (companyId) => {
    setFilterCompany(companyId);
    setCurrentPage(1); // Reset to first page when filter changes
  };

  return (
    <div>
      <PageHead
        title="Transaction Entry"
        sub="Record daily transactions. Commissions calculate automatically."
        actions={<button className="btn primary" onClick={openNew}>+ Add Transaction</button>}
      />

      <div className="toolbar">
        <div className="filters">
          <select value={filterCompany} onChange={(e) => handleFilterChange(e.target.value)}>
            <option value="">All Companies</option>
            {db.companies.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </div>
        <span className="muted">{totalTransactions} transactions</span>
      </div>

      <div className="panel">
        {rows.length === 0 ? (
          <div className="panel-body"><Empty icon="⇄" text="No transactions recorded." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Company</th>
                  <th>Merchant</th>
                  <th>Gateway</th>
                  <th className="num">Txn Amt</th>
                  <th className="num">Settlement</th>
                  <th className="num">Admin Comm.</th>
                  <th className="num">Benef. Comm.</th>
                  <th className="num">Admin Net</th>
                  <th className="num">Company Net</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedRows.map((t) => {
                  const merchant = db.merchants.find((m) => m.id === t.merchantId);
                  const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
                  const co = db.companies.find((c) => c.id === t.companyId);
                  const r = calcTransaction(t, { company: co, merchant, affiliate });
                  return (
                    <tr key={t.id}>
                      <td className="nowrap">{t.date}</td>
                      <td>{nameOf('companies', t.companyId)}</td>
                      <td>{nameOf('merchants', t.merchantId)}</td>
                      <td>{nameOf('gateways', t.gatewayId)}</td>
                      <td className="num mono">{inr(t.txnAmount)}</td>
                      <td className="num mono">{inr(t.settlementAmount)}</td>
                      <td className="num mono">{inr(r.adminCommission)}</td>
                      <td className="num mono">{inr(r.beneficiaryCommission)}<br /><span className="muted" style={{ fontSize: 10 }}>{r.beneficiary}</span></td>
                      <td className="num mono">{inr(r.adminNetCommission)}</td>
                      <td className="num mono bold">{inr(r.companyNetIncome)}</td>
                      <td className="center">
                        <button className="btn sm danger" onClick={() => setConfirmId(t.id)}>Delete</button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalTransactions > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalTransactions}
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
          wide
          title="Add Transaction"
          onClose={() => setOpen(false)}
          footer={<>
            <button className="btn ghost" onClick={() => setOpen(false)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Transaction</button>
          </>}
        >
          <div className="section-title">Step 1–4 · Selection</div>
          <div className="form-grid cols-3">
            <Field label="① Company">
              <select value={form.companyId} onChange={(e) => setForm({ ...form, companyId: e.target.value, merchantId: '', gatewayId: '' })}>
                <option value="">— Select —</option>
                {db.companies.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </Field>
            <Field label="② Merchant" help={form.companyId ? `${availableMerchants.length} under this company` : 'Select a company first'}>
              <select value={form.merchantId} onChange={(e) => setForm({ ...form, merchantId: e.target.value })} disabled={!form.companyId}>
                <option value="">— Select —</option>
                {availableMerchants.map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
              </select>
            </Field>
            <Field label="③ Payment Gateway" help={form.companyId ? '' : 'Select a company first'}>
              <select value={form.gatewayId} onChange={(e) => setForm({ ...form, gatewayId: e.target.value })} disabled={!form.companyId}>
                <option value="">— Select —</option>
                {availableGateways.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
              </select>
            </Field>
          </div>

          <div className="section-title">Step 5 · Details</div>
          <div className="form-grid cols-3">
            <Field label="Transaction Date">
              <input type="date" value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} />
            </Field>
            <Field label="Transaction Amount">
              <input type="number" value={form.txnAmount} onChange={(e) => setForm({ ...form, txnAmount: e.target.value })} placeholder="50000" />
            </Field>
            <Field label="Settlement Amount">
              <input type="number" value={form.settlementAmount} onChange={(e) => setForm({ ...form, settlementAmount: e.target.value })} placeholder="49500" />
            </Field>
            <Field label="Transaction Charges">
              <input type="number" value={form.txnCharges} onChange={(e) => setForm({ ...form, txnCharges: e.target.value })} placeholder="500" />
            </Field>
            <Field label="Other Charges (Optional)">
              <input type="number" value={form.otherCharges} onChange={(e) => setForm({ ...form, otherCharges: e.target.value })} placeholder="0" />
            </Field>
            <Field label="Remarks">
              <input value={form.remarks} onChange={(e) => setForm({ ...form, remarks: e.target.value })} />
            </Field>
          </div>

          {preview && (
            <>
              <div className="section-title">Live Commission Calculation</div>
              <div className="calc-box" style={{ marginBottom: 14 }}>
                <div className="calc-row"><span className="lbl">Settlement Amount</span><span className="mono">{inr(preview.settlementAmount)}</span></div>
                <div className="calc-row"><span className="lbl">Admin Commission ({preview.gatewayCommissionPct}% on Txn Amount)</span><span className="mono">− {inr(preview.adminCommission)}</span></div>
                {preview.chargeBearer === 'Company' && (
                  <div className="calc-row"><span className="lbl">Charges (borne by Company)</span><span className="mono">− {inr(preview.companyChargesDeducted)}</span></div>
                )}
                <div className="calc-row total"><span className="lbl">Company Net Income</span><span className="mono">{inr(preview.companyNetIncome)}</span></div>
              </div>
              <div className="calc-box">
                <div className="calc-row"><span className="lbl">Admin Commission (from Company)</span><span className="mono">{inr(preview.adminCommission)}</span></div>
                <div className="calc-row"><span className="lbl">{preview.beneficiary} Commission ({preview.beneficiaryPct}% on {preview.beneficiaryBase}) — paid by Admin</span><span className="mono">− {inr(preview.beneficiaryCommission)}</span></div>
                {preview.chargeBearer === 'Admin' && (
                  <div className="calc-row"><span className="lbl">Charges (borne by Admin)</span><span className="mono">− {inr(preview.adminChargesDeducted)}</span></div>
                )}
                <div className="calc-row total"><span className="lbl">Admin Net Income</span><span className="mono">{inr(preview.adminNetCommission)}</span></div>
              </div>
            </>
          )}
        </Modal>
      )}

      {confirmId && (
        <Confirm message="Delete this transaction?" onCancel={() => setConfirmId(null)} onConfirm={() => { remove('transactions', confirmId); setConfirmId(null); }} />
      )}
    </div>
  );
}
