import { useState } from 'react';
import { useStore, uid } from '../data/store';
import { PageHead, Empty, Confirm } from '../components/ui';
import Modal from '../components/Modal';
import PaymentGatewayForm from '../components/PaymentGatewayForm';

const blank = { gatewayId: '', loginLink: '', merchantId: '', username: '', password: '', mobile: '', email: '', custom: [] };

export default function MerchantGateways() {
  const { db, update } = useStore();
  const [merchantId, setMerchantId] = useState('');
  const [edit, setEdit] = useState(null); // 'new' | index
  const [form, setForm] = useState(blank);
  const [view, setView] = useState(null);
  const [reveal, setReveal] = useState(false);
  const [confirmIdx, setConfirmIdx] = useState(null);

  const merchant = db.merchants.find((m) => m.id === merchantId);
  const list = merchant?.paymentGateways || [];
  const gwName = (id) => db.gateways.find((g) => g.id === id)?.name || '—';

  const openNew = () => { setForm({ ...blank, id: uid('pg') }); setEdit('new'); };
  const openEdit = (pg, idx) => { setForm(JSON.parse(JSON.stringify({ ...blank, ...pg }))); setEdit(idx); };

  const save = () => {
    if (!form.gatewayId) { alert('Please select a payment gateway.'); return; }
    const next = [...list];
    if (edit === 'new') next.push(form);
    else next[edit] = form;
    update('merchants', merchant.id, { paymentGateways: next });
    setEdit(null);
  };

  const del = (idx) => {
    update('merchants', merchant.id, { paymentGateways: list.filter((_, i) => i !== idx) });
    setConfirmIdx(null);
  };

  return (
    <div>
      <PageHead title="Add Payment Gateway" sub="Store payment-gateway login credentials per merchant. Multiple gateways supported." />

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
        {merchant && <button className="btn primary" style={{ alignSelf: 'flex-end' }} onClick={openNew}>+ Add Payment Gateway</button>}
      </div>

      {!merchant ? (
        <div className="panel"><div className="panel-body"><Empty icon="⬢" text="Select a merchant to manage its payment gateways." /></div></div>
      ) : (
        <div className="panel">
          {list.length === 0 ? (
            <div className="panel-body"><Empty icon="⬢" text="No payment gateways added for this merchant." /></div>
          ) : (
            <table className="data">
              <thead>
                <tr><th>Gateway</th><th>Merchant ID</th><th>Username</th><th>Login Link</th><th className="center">Actions</th></tr>
              </thead>
              <tbody>
                {list.map((pg, idx) => (
                  <tr key={pg.id || idx}>
                    <td className="bold">{gwName(pg.gatewayId)}</td>
                    <td className="mono">{pg.merchantId || '—'}</td>
                    <td className="mono">{pg.username || '—'}</td>
                    <td className="muted">{pg.loginLink || '—'}</td>
                    <td className="center">
                      <div className="row-actions" style={{ justifyContent: 'center' }}>
                        <button className="btn sm" onClick={() => { setView(pg); setReveal(false); }}>View</button>
                        <button className="btn sm ghost" onClick={() => openEdit(pg, idx)}>Edit</button>
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
          title={edit === 'new' ? `Add Payment Gateway — ${merchant.name}` : `Edit Payment Gateway — ${merchant.name}`}
          onClose={() => setEdit(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setEdit(null)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save Gateway</button>
          </>}
        >
          <PaymentGatewayForm value={form} onChange={setForm} />
        </Modal>
      )}

      {view && (
        <Modal title={`${gwName(view.gatewayId)} Credentials`} onClose={() => setView(null)} footer={<button className="btn ghost" onClick={() => setView(null)}>Close</button>}>
          <div className="btn-row" style={{ marginBottom: 14 }}>
            <button className="btn sm" onClick={() => setReveal(!reveal)}>{reveal ? 'Hide' : 'Reveal'} Password</button>
          </div>
          <table className="data">
            <tbody>
              <Row k="Gateway" v={gwName(view.gatewayId)} />
              <Row k="Login Link" v={view.loginLink} />
              <Row k="Merchant ID" v={view.merchantId} />
              <Row k="Username" v={view.username} />
              <Row k="Password" v={reveal ? view.password : '••••••••'} />
              <Row k="Registered Mobile" v={view.mobile} />
              <Row k="Registered Email" v={view.email} />
              {(view.custom || []).map((c, i) => <Row key={i} k={c.key} v={c.value} />)}
            </tbody>
          </table>
        </Modal>
      )}

      {confirmIdx !== null && (
        <Confirm message="Delete this payment gateway entry?" onCancel={() => setConfirmIdx(null)} onConfirm={() => del(confirmIdx)} />
      )}
    </div>
  );
}

function Row({ k, v }) {
  return <tr><td className="bold" style={{ width: '40%' }}>{k}</td><td className="mono">{v || '—'}</td></tr>;
}
