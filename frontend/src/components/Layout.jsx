import { useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useStore } from '../data/store';

// Admin navigation. Reused as-is for the SuperAdmin role (which runs its own
// business exactly like an Admin) plus the SuperAdmin-only "Leases" entry.
const ADMIN_NAV = [
  { section: 'Overview' },
  { to: '/', label: 'Dashboard', ico: '▦', end: true },
  { section: 'Masters' },
  {
    label: 'Company', ico: '▣', children: [
      { to: '/companies/new', label: 'Add New Company' },
      { to: '/companies', label: 'View / Edit Company', end: true },
      { to: '/companies/payment', label: 'Payment to Company' },
    ],
  },
  {
    label: 'Merchant', ico: '◈', children: [
      { to: '/merchants/new', label: 'Add New Merchant' },
      { to: '/merchants', label: 'View / Edit Merchant', end: true },
      { to: '/merchants/payment-gateway', label: 'Add Payment Gateway' },
      { to: '/merchants/bank-account', label: 'Add Bank Account' },
    ],
  },
  { to: '/affiliates', label: 'Affiliates', ico: '⬡' },
  { to: '/gateways', label: 'Payment Gateways', ico: '⬢' },
  { to: '/banks', label: 'Banks', ico: '▤' },
  { section: 'Operations' },
  { to: '/transactions', label: 'Transactions', ico: '⇄' },
  { to: '/settlements', label: 'Settlements', ico: '⇣' },
  { section: 'Ledgers & Reports' },
  { to: '/ledgers', label: 'Ledgers', ico: '≣' },
  { to: '/reports', label: 'Reports', ico: '◫' },
];

const NAV_BY_ROLE = {
  Admin: ADMIN_NAV,
  // SuperAdmin sees the full Admin nav plus a "Leases" entry positioned
  // immediately below "Reports" (Req 16.1). Built by reusing ADMIN_NAV to stay
  // DRY; other roles never include this entry (Req 16.2, 16.4).
  SuperAdmin: [
    ...ADMIN_NAV,
    { to: '/leases', label: 'Leases', ico: '⧉' },
  ],
  Company: [
    { section: 'Company Portal' },
    { to: '/', label: 'Dashboard', ico: '▦', end: true },
    { to: '/merchants', label: 'My Merchants', ico: '◈' },
    { to: '/transactions', label: 'Transactions', ico: '⇄' },
    { to: '/settlements', label: 'Settlements', ico: '⇣' },
    { to: '/ledger', label: 'My Ledger', ico: '≣' },
  ],
  Affiliate: [
    { section: 'Affiliate Portal' },
    { to: '/', label: 'Dashboard', ico: '▦', end: true },
    { to: '/merchants', label: 'My Merchants', ico: '◈' },
    { to: '/transactions', label: 'Transactions', ico: '⇄' },
    { to: '/ledger', label: 'Commission Ledger', ico: '≣' },
  ],
  Merchant: [
    { section: 'Merchant Portal' },
    { to: '/', label: 'Dashboard', ico: '▦', end: true },
    { to: '/transactions', label: 'My Transactions', ico: '⇄' },
    { to: '/banks', label: 'My Banks', ico: '▤' },
    { to: '/ledger', label: 'Commission Ledger', ico: '≣' },
  ],
};

const TITLES = {
  '/': 'Dashboard',
  '/companies': 'View / Edit Company',
  '/companies/new': 'Add New Company',
  '/companies/payment': 'Payment to Company',
  '/merchants': 'View / Edit Merchant',
  '/merchants/new': 'Add New Merchant',
  '/merchants/payment-gateway': 'Add Payment Gateway',
  '/merchants/bank-account': 'Add Bank Account',
  '/affiliates': 'Affiliate Management',
  '/gateways': 'Payment Gateway Master',
  '/banks': 'Banks',
  '/transactions': 'Transactions',
  '/settlements': 'Settlements',
  '/ledgers': 'Ledgers',
  '/ledger': 'Ledger',
  '/reports': 'Reports',
  '/leases': 'Lease Management',
};

const BRAND_SUB = {
  Admin: 'Admin Console',
  SuperAdmin: 'SuperAdmin Console',
  Company: 'Company Portal',
  Affiliate: 'Affiliate Portal',
  Merchant: 'Merchant Portal',
};

function NavGroup({ item }) {
  const loc = useLocation();
  const childActive = item.children.some((c) => (c.end ? loc.pathname === c.to : loc.pathname.startsWith(c.to)));
  const [open, setOpen] = useState(childActive);

  return (
    <div>
      <div className={`nav-item ${childActive && !open ? 'active' : ''}`} onClick={() => setOpen(!open)}>
        <span className="ico">{item.ico}</span>
        {item.label}
        <span style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--muted-2)' }}>{open ? '▾' : '▸'}</span>
      </div>
      {open && item.children.map((c) => (
        <NavLink
          key={c.to}
          to={c.to}
          end={c.end}
          className={({ isActive }) => `nav-item nav-sub ${isActive ? 'active' : ''}`}
        >
          {c.label}
        </NavLink>
      ))}
    </div>
  );
}

export default function Layout({ children }) {
  const { auth, logout } = useStore();
  const loc = useLocation();
  const title = TITLES[loc.pathname] || 'Dashboard';
  const nav = NAV_BY_ROLE[auth?.role] || NAV_BY_ROLE.Admin;
  const [mobileOpen, setMobileOpen] = useState(false);

  // Close the drawer whenever the route changes.
  const closeDrawer = () => setMobileOpen(false);

  return (
    <div className="shell">
      {mobileOpen && <div className="sidebar-overlay" onClick={closeDrawer} />}
      <aside className={`sidebar ${mobileOpen ? 'open' : ''}`}>
        <div className="brand">
          <div className="brand-mark">◎</div>
          <div>
            <div className="logo">Payment Gateway</div>
            <div className="sub">{BRAND_SUB[auth?.role] || 'Settlement System'}</div>
          </div>
        </div>
        <nav onClick={(e) => { if (e.target.closest('a')) closeDrawer(); }}>
          {nav.map((item, i) =>
            item.section ? (
              <div className="nav-section" key={`s${i}`}>{item.section}</div>
            ) : item.children ? (
              <NavGroup key={item.label} item={item} />
            ) : (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
              >
                <span className="ico">{item.ico}</span>
                {item.label}
              </NavLink>
            )
          )}
        </nav>

        {/* Dashhold Branding - Bottom of Sidebar */}
        <div style={{ 
          marginTop: 'auto', 
          padding: '16px 20px',
          borderTop: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          opacity: 0.6,
          transition: 'opacity 0.2s'
        }}
        onMouseEnter={(e) => e.currentTarget.style.opacity = '1'}
        onMouseLeave={(e) => e.currentTarget.style.opacity = '0.6'}>
          <span style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.7)' }}>Made by</span>
          <a 
            href="https://Dashhold.com" 
            target="_blank" 
            rel="noopener noreferrer"
            style={{ 
              display: 'flex', 
              alignItems: 'center', 
              textDecoration: 'none',
              fontSize: 14,
              fontWeight: 600,
              color: 'rgba(255, 255, 255, 0.9)',
              letterSpacing: '0.02em'
            }}
          >
            Dashhold
          </a>
        </div>
      </aside>

      <div className="main">
        <header className="topbar">
          <div className="topbar-left">
            <button className="menu-toggle" onClick={() => setMobileOpen(true)} aria-label="Open menu">☰</button>
            <div className="page-title">{title}</div>
          </div>
          <div className="user-box">
            <div className="who">
              <div className="name">{auth?.name}</div>
              <div className="role">{auth?.role}</div>
            </div>
            <div className="avatar">{(auth?.name || 'A').charAt(0).toUpperCase()}</div>
            <button className="btn sm" onClick={logout}>Logout</button>
          </div>
        </header>
        <main className="content">{children}</main>
      </div>
    </div>
  );
}
