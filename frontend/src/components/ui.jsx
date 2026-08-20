// Small shared UI helpers

export function Field({ label, children, span, help }) {
  const cls = span === 2 ? 'span-2' : span === 'full' ? 'span-full' : '';
  return (
    <label className={`field ${cls}`}>
      <span>{label}</span>
      {children}
      {help && <div className="help">{help}</div>}
    </label>
  );
}

export function StatusBadge({ status }) {
  const active = status === 'Active';
  return <span className={`badge ${active ? 'solid' : 'muted'}`}>{status}</span>;
}

export function PageHead({ title, sub, actions }) {
  return (
    <div className="page-head">
      <div>
        <div className="ph-title">{title}</div>
        {sub && <div className="ph-sub">{sub}</div>}
      </div>
      {actions && <div className="btn-row">{actions}</div>}
    </div>
  );
}

export function Empty({ icon = '∅', text = 'No records yet.' }) {
  return (
    <div className="empty">
      <div className="big">{icon}</div>
      <div>{text}</div>
    </div>
  );
}

export function Stat({ label, value, meta, invert }) {
  return (
    <div className={`stat ${invert ? 'invert' : ''}`}>
      <div className="label">{label}</div>
      <div className="value">{value}</div>
      {meta && <div className="meta">{meta}</div>}
    </div>
  );
}

export function Confirm({ message, onConfirm, onCancel }) {
  return (
    <div className="modal-overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onCancel(); }}>
      <div className="modal" style={{ maxWidth: 420 }}>
        <div className="modal-head"><h3>Confirm</h3><button className="x" onClick={onCancel}>×</button></div>
        <div className="modal-body">{message}</div>
        <div className="modal-foot">
          <button className="btn ghost" onClick={onCancel}>Cancel</button>
          <button className="btn primary" onClick={onConfirm}>Delete</button>
        </div>
      </div>
    </div>
  );
}
