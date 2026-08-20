import { Field } from './ui';
import { useStore } from '../data/store';

// Controlled merchant payment-gateway credential form.
export default function PaymentGatewayForm({ value, onChange }) {
  const { db } = useStore();
  const set = (patch) => onChange({ ...value, ...patch });

  const custom = value.custom || [];
  const addCustom = () => onChange({ ...value, custom: [...custom, { key: '', value: '' }] });
  const setCustom = (i, patch) => onChange({ ...value, custom: custom.map((c, idx) => (idx === i ? { ...c, ...patch } : c)) });
  const removeCustom = (i) => onChange({ ...value, custom: custom.filter((_, idx) => idx !== i) });

  return (
    <div>
      <div className="section-title">Gateway Credentials</div>
      <div className="form-grid">
        <Field label="Name of Payment Gateway" span={2}>
          <select value={value.gatewayId || ''} onChange={(e) => set({ gatewayId: e.target.value })} autoFocus>
            <option value="">— Select Gateway —</option>
            {db.gateways.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
          </select>
        </Field>
        <Field label="Link to Login" span={2}><input value={value.loginLink || ''} onChange={(e) => set({ loginLink: e.target.value })} placeholder="https://dashboard.gateway.com" /></Field>
        <Field label="Merchant ID"><input value={value.merchantId || ''} onChange={(e) => set({ merchantId: e.target.value })} /></Field>
        <Field label="Username"><input value={value.username || ''} onChange={(e) => set({ username: e.target.value })} /></Field>
        <Field label="Password"><input value={value.password || ''} onChange={(e) => set({ password: e.target.value })} /></Field>
        <Field label="Registered Mobile Number"><input value={value.mobile || ''} onChange={(e) => set({ mobile: e.target.value })} /></Field>
        <Field label="Registered Email" span={2}><input value={value.email || ''} onChange={(e) => set({ email: e.target.value })} /></Field>
      </div>

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
