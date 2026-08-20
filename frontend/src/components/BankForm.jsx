import { Field } from './ui';
import { uid } from '../data/store';

// Controlled bank form. Parent holds the bank object and passes value + onChange.
export default function BankForm({ value, onChange }) {
  const set = (patch) => onChange({ ...value, ...patch });

  const atmCards = value.atmCards || [];
  const addCard = () => onChange({ ...value, atmCards: [...atmCards, { id: uid('atm'), nameOnCard: '', cardNumber: '', expiry: '', cvv: '', atmPin: '' }] });
  const setCard = (i, patch) => onChange({ ...value, atmCards: atmCards.map((c, idx) => (idx === i ? { ...c, ...patch } : c)) });
  const removeCard = (i) => onChange({ ...value, atmCards: atmCards.filter((_, idx) => idx !== i) });

  const custom = value.custom || [];
  const addCustom = () => onChange({ ...value, custom: [...custom, { key: '', value: '' }] });
  const setCustom = (i, patch) => onChange({ ...value, custom: custom.map((c, idx) => (idx === i ? { ...c, ...patch } : c)) });
  const removeCustom = (i) => onChange({ ...value, custom: custom.filter((_, idx) => idx !== i) });

  const mobileBanking = value.mobileBanking || 'No';

  return (
    <div>
      <div className="section-title">Account Details</div>
      <div className="form-grid">
        <Field label="Bank Name"><input value={value.bankName || ''} onChange={(e) => set({ bankName: e.target.value })} autoFocus /></Field>
        <Field label="Account Name"><input value={value.accountName || ''} onChange={(e) => set({ accountName: e.target.value })} /></Field>
        <Field label="Account Number"><input value={value.accountNumber || ''} onChange={(e) => set({ accountNumber: e.target.value })} /></Field>
        <Field label="IFSC Code"><input value={value.ifsc || ''} onChange={(e) => set({ ifsc: e.target.value })} /></Field>
        <Field label="Netbanking Login Link" span={2}><input value={value.netbankingLink || ''} onChange={(e) => set({ netbankingLink: e.target.value })} /></Field>
        <Field label="Username"><input value={value.username || ''} onChange={(e) => set({ username: e.target.value })} /></Field>
        <Field label="Login Password"><input value={value.loginPassword || ''} onChange={(e) => set({ loginPassword: e.target.value })} /></Field>
        <Field label="Transaction Password"><input value={value.txnPassword || ''} onChange={(e) => set({ txnPassword: e.target.value })} /></Field>
        <Field label="Customer ID"><input value={value.customerId || ''} onChange={(e) => set({ customerId: e.target.value })} /></Field>
        <Field label="Registered Mobile"><input value={value.mobile || ''} onChange={(e) => set({ mobile: e.target.value })} /></Field>
        <Field label="Registered Email"><input value={value.email || ''} onChange={(e) => set({ email: e.target.value })} /></Field>
      </div>

      <div className="section-title">Mobile Banking</div>
      <div className="checkbox-line" style={{ gap: 20 }}>
        <label className="checkbox-line" style={{ marginBottom: 0 }}>
          <input type="radio" name={`mb-${value.id || 'new'}`} checked={mobileBanking === 'Yes'} onChange={() => set({ mobileBanking: 'Yes' })} />
          <span>Yes</span>
        </label>
        <label className="checkbox-line" style={{ marginBottom: 0 }}>
          <input type="radio" name={`mb-${value.id || 'new'}`} checked={mobileBanking === 'No'} onChange={() => set({ mobileBanking: 'No' })} />
          <span>No</span>
        </label>
      </div>
      {mobileBanking === 'Yes' && (
        <div className="form-grid" style={{ marginTop: 12 }}>
          <Field label="Login ID"><input value={value.mobileLoginId || ''} onChange={(e) => set({ mobileLoginId: e.target.value })} /></Field>
          <Field label="MPIN"><input value={value.mpin || ''} onChange={(e) => set({ mpin: e.target.value })} /></Field>
        </div>
      )}

      <div className="section-title">ATM / Debit Cards</div>
      {atmCards.length === 0 && <p className="help" style={{ marginBottom: 10 }}>No cards added. A bank account can hold multiple cards.</p>}
      {atmCards.map((c, i) => (
        <div key={c.id || i} className="bank-card">
          <div className="bc-head">
            <span className="bank-name">Card {i + 1}</span>
            <button className="btn sm danger" onClick={() => removeCard(i)}>Remove</button>
          </div>
          <div className="form-grid cols-3">
            <Field label="Name on Card" span={2}><input value={c.nameOnCard} onChange={(e) => setCard(i, { nameOnCard: e.target.value })} /></Field>
            <Field label="Card Number"><input value={c.cardNumber} onChange={(e) => setCard(i, { cardNumber: e.target.value })} placeholder="•••• •••• •••• ••••" /></Field>
            <Field label="Expiry"><input value={c.expiry} onChange={(e) => setCard(i, { expiry: e.target.value })} placeholder="MM/YY" /></Field>
            <Field label="CVV"><input value={c.cvv} onChange={(e) => setCard(i, { cvv: e.target.value })} placeholder="•••" /></Field>
            <Field label="ATM PIN"><input value={c.atmPin} onChange={(e) => setCard(i, { atmPin: e.target.value })} placeholder="••••" /></Field>
          </div>
        </div>
      ))}
      <button className="btn sm ghost" onClick={addCard}>+ Add Card</button>

      <div className="section-title">Custom Fields</div>
      {(custom || []).map((c, i) => (
        <div key={i} className="form-grid" style={{ gridTemplateColumns: '1fr 1fr auto', alignItems: 'end', gap: 10 }}>
          <Field label="Field Name"><input value={c.key} onChange={(e) => setCustom(i, { key: e.target.value })} /></Field>
          <Field label="Value"><input value={c.value} onChange={(e) => setCustom(i, { value: e.target.value })} /></Field>
          <button className="btn sm danger" style={{ marginBottom: 16 }} onClick={() => removeCustom(i)}>Remove</button>
        </div>
      ))}
      <button className="btn sm ghost" onClick={addCustom}>+ Add Custom Field</button>
    </div>
  );
}
