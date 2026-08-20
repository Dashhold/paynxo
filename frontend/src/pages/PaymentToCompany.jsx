import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Empty, Field, Stat, Confirm } from '../components/ui';
import { companyLedger, inr } from '../data/calc';

const blank = {
  companyId: '', date: new Date().toISOString().slice(0, 10),
  amount: '', from: '', to: '',
};

export default function PaymentToCompany() {
  const { db, add, remove } = useStore();
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);

  const company = db.companies.find((c) => c.id === form.companyId);
  const led = form.companyId ? companyLedger(form.companyId, db) : null;

  const save = () => {
    if (!form.companyId) { alert('Please select a company.'); return; }
    if (!form.amount || Number(form.amount) <= 0) { alert('Enter a valid amount.'); return; }
    if (!form.date) { alert('Please select a date.'); return; }
    add('settlements', {
      companyId: form.companyId,
      date: form.date,
      amount: Number(form.amount),
      from: form.from,
      to: form.to,
      paymentMode: 'Payment to Company',
      refNumber: '',
      remarks: form.from || form.to ? `From: ${form.from || '—'} → To: ${form.to || '—'}` : '',
    });
    setForm({ ...blank, companyId: form.companyId });
  };

  const payments = db.settlements.filter((s) => s.companyId === form.companyId).slice().reverse();

  return (
    <div>
      <PageHead title="Payment to Company" sub="Record a payment to a company. Balance is shown instantly on selection." />

      <div className="split-2">
        <div className="panel">
          <div className="panel-head"><h2>Payment Details</h2></div>
          <div className="panel-body">
            <Field label="Select Company">
              <select value={form.companyId} onChange={(e) => setForm({ ...form, companyId: e.target.value })}>
                <option value="">— Select Company —</option>
                {db.companies.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </Field>

            {company && led && (
              <div className="stat-grid" style={{ gridTemplateColumns: '1fr', gap: 12, marginBottom: 18 }}>
                <Stat
                  label="Current Company Balance"
                  value={inr(Math.abs(led.balance))}
                  meta={led.balance > 0.0001 ? `${company.name} should receive` : led.balance < -0.0001 ? `${company.name} owes Admin` : 'Settled'}
                  invert
                />
              </div>
            )}

            <div className="form-grid">
              <Field label="Date"><input type="date" value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} /></Field>
              <Field label="Amount"><input type="number" value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} placeholder="0.00" /></Field>
              <Field label="From"><input value={form.from} onChange={(e) => setForm({ ...form, from: e.target.value })} placeholder="Paying account / party" /></Field>
              <Field label="To"><input value={form.to} onChange={(e) => setForm({ ...form, to: e.target.value })} placeholder="Receiving account / party" /></Field>
            </div>

            <div className="btn-row" style={{ marginTop: 8 }}>
              <button className="btn primary" onClick={save} disabled={!form.companyId}>Save Payment</button>
            </div>
          </div>
        </div>

        <div className="panel">
          <div className="panel-head"><h2>{company ? `${company.name} — Payments` : 'Payment History'}</h2></div>
          {!company ? (
            <div className="panel-body"><Empty icon="⇣" text="Select a company to view its payment history." /></div>
          ) : payments.length === 0 ? (
            <div className="panel-body"><Empty icon="⇣" text="No payments recorded yet." /></div>
          ) : (
            <table className="data">
              <thead>
                <tr><th>Date</th><th className="num">Amount</th><th>From</th><th>To</th><th className="center">Actions</th></tr>
              </thead>
              <tbody>
                {payments.map((p) => (
                  <tr key={p.id}>
                    <td className="nowrap">{p.date}</td>
                    <td className="num mono">{inr(p.amount)}</td>
                    <td>{p.from || '—'}</td>
                    <td>{p.to || '—'}</td>
                    <td className="center"><button className="btn sm danger" onClick={() => setConfirmId(p.id)}>Delete</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {confirmId && (
        <Confirm message="Delete this payment?" onCancel={() => setConfirmId(null)} onConfirm={() => { remove('settlements', confirmId); setConfirmId(null); }} />
      )}
    </div>
  );
}
