import { useState } from 'react';
import { useStore, uid } from '../data/store';
import { PageHead, Empty, Confirm } from '../components/ui';
import Modal from '../components/Modal';
import BankForm from '../components/BankForm';

const blankBank = {
  bankName: '', accountName: '', accountNumber: '', ifsc: '',
  netbankingLink: '', username: '', loginPassword: '', txnPassword: '',
  customerId: '', mobile: '', email: '',
  mobileBanking: 'No', mobileLoginId: '', mpin: '',
  atmCards: [], custom: [],
};

export default function MerchantBankAccounts() {
  const { db, update } = useStore();
  const [merchantId, setMerchantId] = useState('');
  const [edit, setEdit] = useState(null); // 'new' | index
  const [form, setForm] = useState(blankBank);
  const [view, setView] = useState(null);
  const [reveal, setReveal] = useState(false);
  const [confirmIdx, setConfirmIdx] = useState(null);

  const merchant = db.merchants.find((m) => m.id === merchantId);
  const list = merchant?.banks || [];

  const openNew = () => { setForm({ ...blankBank, id: uid('bk') }); setEdit('new'); };
  const openEdit = (b, idx) => { setForm(JSON.parse(JSON.stringify({ ...blankBank, ...b }))); setEdit(idx); };

  const save = () => {
    if (!form.bankName.trim() || !form.accountNumber.trim()) { alert('Bank Name and Account Number are required.'); return; }
    const next = [...list];
    if (edit === 'new') next.push(form);
    else next[edit] = form;
    update('merchants', merchant.id, { banks: next });
    setEdit(null);
  };

  const del = (idx) => {
    update('merchants', merchant.id, { banks: list.filter((_, i) => i !== idx) });
    setConfirmIdx(null);
  };

  return (
    <div>
      <PageHead title="Add Bank Account" sub="Bank accounts with netbanking, mobile banking and multiple ATM cards." />

      <div className="toolbar">
        <div className="filters">
          <label className="field" style={{ marginBottom: 0, minWidth: 280 }}>
            <span>Select Merchant</span>
            <select value={merchantId} onChange={(e) => setMerchantId(e.target.value)}>
              <option value="">— Select Merchant —</option>
              {db.merchants.map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
            </select>
          </label>
        </div>
        {merchant && <button className="btn primary" style={{ alignSelf: 'flex-end' }} onClick={openNew}>+ Add Bank Account</button>}
      </div>

      {!merchant ? (
        <div className="panel"><div className="panel-body"><Empty icon="▤" text="Select a merchant to manage its bank accounts." /></div></div>
      ) : (
        <div className="panel">
          {list.length === 0 ? (
            <div className="panel-body"><Empty icon="▤" text="No bank accounts added for this merchant." /></div>
          ) : (
            <table className="data">
              <thead>
                <tr><th>Bank</th><th>Account No.</th><th>IFSC</th><th>Mobile Banking</th><th>Cards</th><th className="center">Actions</th></tr>
              </thead>
              <tbody>
                {list.map((b, idx) => (
                  <tr key={b.id || idx}>
                    <td className="bold">{b.bankName}<br /><span className="muted">{b.accountName}</span></td>
                    <td className="mono">{b.accountNumber}</td>
                    <td className="mono">{b.ifsc || '—'}</td>
                    <td>{b.mobileBanking === 'Yes' ? <span className="badge solid">Enabled</span> : <span className="badge muted">No</span>}</td>
                    <td>{(b.atmCards || []).length}</td>
                    <td className="center">
                      <div className="row-actions" style={{ justifyContent: 'center' }}>
                        <button className="btn sm" onClick={() => { setView(b); setReveal(false); }}>View</button>
                        <button className="btn sm ghost" onClick={() => openEdit(b, idx)}>Edit</button>
                        <button className="btn sm danger" onClick={() => setConfirmIdx(idx)}>Delete</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {edit !== null && (
        <Modal
          wide
          title={edit === 'new' ? `Add Bank Account — ${merchant.name}` : `Edit Bank Account — ${merchant.name}`}
          onClose={() => setEdit(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setEdit(null)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Bank</button>
          </>}
        >
          <BankForm value={form} onChange={setForm} />
        </Modal>
      )}

      {view && <BankView bank={view} reveal={reveal} setReveal={setReveal} onClose={() => setView(null)} />}

      {confirmIdx !== null && (
        <Confirm message="Delete this bank account?" onCancel={() => setConfirmIdx(null)} onConfirm={() => del(confirmIdx)} />
      )}
    </div>
  );
}

export function BankView({ bank, reveal, setReveal, onClose }) {
  return (
    <Modal title={`${bank.bankName} — Account Details`} onClose={onClose} footer={<button className="btn ghost" onClick={onClose}>Close</button>}>
      <div className="btn-row" style={{ marginBottom: 14 }}>
        <button className="btn sm" onClick={() => setReveal(!reveal)}>{reveal ? 'Hide' : 'Reveal'} Credentials</button>
      </div>
      <table className="data">
        <tbody>
          <Row k="Account Name" v={bank.accountName} />
          <Row k="Account Number" v={bank.accountNumber} />
          <Row k="IFSC Code" v={bank.ifsc} />
          <Row k="Netbanking Link" v={bank.netbankingLink} />
          <Row k="Username" v={bank.username} />
          <Row k="Login Password" v={reveal ? bank.loginPassword : '••••••••'} />
          <Row k="Transaction Password" v={reveal ? bank.txnPassword : '••••••••'} />
          <Row k="Customer ID" v={bank.customerId} />
          <Row k="Registered Mobile" v={bank.mobile} />
          <Row k="Registered Email" v={bank.email} />
          <Row k="Mobile Banking" v={bank.mobileBanking || 'No'} />
          {bank.mobileBanking === 'Yes' && <Row k="Mobile Login ID" v={bank.mobileLoginId} />}
          {bank.mobileBanking === 'Yes' && <Row k="MPIN" v={reveal ? bank.mpin : '••••'} />}
          {(bank.custom || []).map((c, i) => <Row key={i} k={c.key} v={c.value} />)}
        </tbody>
      </table>

      {(bank.atmCards || []).length > 0 && (
        <>
          <div className="section-title">ATM / Debit Cards</div>
          {bank.atmCards.map((c, i) => (
            <div key={c.id || i} className="bank-card">
              <div className="bc-head"><span className="bank-name">Card {i + 1} — {c.nameOnCard || '—'}</span></div>
              <table className="data">
                <tbody>
                  <Row k="Card Number" v={c.cardNumber} />
                  <Row k="Expiry" v={c.expiry} />
                  <Row k="CVV" v={reveal ? c.cvv : '•••'} />
                  <Row k="ATM PIN" v={reveal ? c.atmPin : '••••'} />
                </tbody>
              </table>
            </div>
          ))}
        </>
      )}
    </Modal>
  );
}

function Row({ k, v }) {
  return <tr><td className="bold" style={{ width: '40%' }}>{k}</td><td className="mono">{v || '—'}</td></tr>;
}
