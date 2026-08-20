import { useState } from 'react';
import { useStore } from '../data/store';

// Login.jsx — a single credential form for every role. The backend authenticates
// the user id / password and returns the principal's role; the app then routes
// to the matching portal (SuperAdmin, Admin, Company, Affiliate, Merchant). No
// demo accounts or role tabs are shown.
export default function Login() {
  const { login } = useStore();
  const [userId, setUserId] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [err, setErr] = useState('');
  const [busy, setBusy] = useState(false);

  const submit = async (e) => {
    e.preventDefault();
    setErr('');
    setBusy(true);
    try {
      // No expectedRole: the server resolves the role and the app routes on it.
      const res = await login(userId.trim(), password);
      if (res?.error === 'invalid') {
        setErr('Invalid credentials. Please check your user id and password.');
      } else if (res?.error === 'network') {
        setErr(res.message || 'Cannot reach the server. Please try again.');
      } else if (res?.error) {
        setErr(res.message || 'Unable to sign in. Please try again.');
      }
      // On success the store sets auth and the app renders the matching portal.
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="login-wrap">
      <div className="login-card">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 4 }}>
          <div style={{ width: 40, height: 40, borderRadius: 10, background: 'var(--accent)', color: '#04140a', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 22, fontWeight: 800 }}>◎</div>
          <div className="lc-logo">Payment Gateway Ops</div>
        </div>
        <div className="lc-sub">Commission &amp; Settlement System</div>

        <div className="lc-sub" style={{ margin: '14px 0 18px' }}>Sign in to your account</div>

        {err && <div className="err">{err}</div>}

        <form onSubmit={submit}>
          <label className="field">
            <span>User ID</span>
            <input value={userId} onChange={(e) => setUserId(e.target.value)} placeholder="Enter user id" autoFocus autoComplete="username" />
          </label>
          <label className="field">
            <span>Password</span>
            <div style={{ position: 'relative' }}>
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password"
                autoComplete="current-password"
                style={{ width: '100%', paddingRight: 64 }}
              />
              <button
                type="button"
                onClick={() => setShowPassword((s) => !s)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                style={{
                  position: 'absolute', top: '50%', right: 8, transform: 'translateY(-50%)',
                  background: 'none', border: 'none', cursor: 'pointer', padding: '4px 6px',
                  fontSize: 12, fontWeight: 700, letterSpacing: 0.5, color: 'var(--accent, #1f9d55)',
                }}
              >
                {showPassword ? 'HIDE' : 'SHOW'}
              </button>
            </div>
          </label>
          <button className="btn primary" style={{ width: '100%' }} type="submit" disabled={busy}>
            {busy ? 'Signing in…' : 'Sign In'}
          </button>
        </form>
      </div>

      {/* Dashhold Branding - Lower Left */}
      <div style={{ 
        position: 'fixed', 
        bottom: 20, 
        left: 20, 
        display: 'flex', 
        alignItems: 'center', 
        gap: 8,
        opacity: 0.7,
        transition: 'opacity 0.2s'
      }}
      onMouseEnter={(e) => e.currentTarget.style.opacity = '1'}
      onMouseLeave={(e) => e.currentTarget.style.opacity = '0.7'}>
        <span style={{ fontSize: 13, color: 'var(--text-muted, #666)' }}>Made by</span>
        <a 
          href="https://Dashhold.com" 
          target="_blank" 
          rel="noopener noreferrer"
          style={{ 
            display: 'flex', 
            alignItems: 'center', 
            textDecoration: 'none',
            fontSize: 15,
            fontWeight: 600,
            color: 'var(--accent, #22c55e)',
            letterSpacing: '0.02em'
          }}
        >
          Dashhold
        </a>
      </div>
    </div>
  );
}
