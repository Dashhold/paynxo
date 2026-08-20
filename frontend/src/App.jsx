import { Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from './data/store';
import Layout from './components/Layout';
import Login from './pages/Login';

// Admin pages
import Dashboard from './pages/Dashboard';
import Companies from './pages/Companies';
import Merchants from './pages/Merchants';
import Affiliates from './pages/Affiliates';
import Gateways from './pages/Gateways';
import Banks from './pages/Banks';
import Transactions from './pages/Transactions';
import Settlements from './pages/Settlements';
import Ledgers from './pages/Ledgers';
import Reports from './pages/Reports';

// Company portal
import { CompanyDashboard, CompanyMerchants, CompanyTransactions, CompanySettlements, CompanyLedger } from './pages/company';
// Affiliate portal
import { AffiliateDashboard, AffiliateMerchants, AffiliateTransactions, AffiliateLedgerPage } from './pages/affiliate';
// Merchant portal
import { MerchantDashboard, MerchantTransactions, MerchantBanks, MerchantLedgerPage } from './pages/merchant';
// Admin sub-screens
import MerchantGateways from './pages/MerchantGateways';
import MerchantBankAccounts from './pages/MerchantBankAccounts';
import PaymentToCompany from './pages/PaymentToCompany';
// SuperAdmin lease management
import Leases from './pages/Leases';

function AdminRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Dashboard />} />
      <Route path="/companies" element={<Companies />} />
      <Route path="/companies/new" element={<Companies startNew />} />
      <Route path="/companies/payment" element={<PaymentToCompany />} />
      <Route path="/merchants" element={<Merchants />} />
      <Route path="/merchants/new" element={<Merchants startNew />} />
      <Route path="/merchants/payment-gateway" element={<MerchantGateways />} />
      <Route path="/merchants/bank-account" element={<MerchantBankAccounts />} />
      <Route path="/affiliates" element={<Affiliates />} />
      <Route path="/gateways" element={<Gateways />} />
      <Route path="/banks" element={<Banks />} />
      <Route path="/transactions" element={<Transactions />} />
      <Route path="/settlements" element={<Settlements />} />
      <Route path="/ledgers" element={<Ledgers />} />
      <Route path="/reports" element={<Reports />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

// SuperAdmin runs its own business exactly like an Admin (Req 12), so it gets
// the full set of Admin routes plus the SuperAdmin-only lease management screen.
function SuperAdminRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Dashboard />} />
      <Route path="/companies" element={<Companies />} />
      <Route path="/companies/new" element={<Companies startNew />} />
      <Route path="/companies/payment" element={<PaymentToCompany />} />
      <Route path="/merchants" element={<Merchants />} />
      <Route path="/merchants/new" element={<Merchants startNew />} />
      <Route path="/merchants/payment-gateway" element={<MerchantGateways />} />
      <Route path="/merchants/bank-account" element={<MerchantBankAccounts />} />
      <Route path="/affiliates" element={<Affiliates />} />
      <Route path="/gateways" element={<Gateways />} />
      <Route path="/banks" element={<Banks />} />
      <Route path="/transactions" element={<Transactions />} />
      <Route path="/settlements" element={<Settlements />} />
      <Route path="/ledgers" element={<Ledgers />} />
      <Route path="/reports" element={<Reports />} />
      <Route path="/leases" element={<Leases />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function CompanyRoutes() {
  return (
    <Routes>
      <Route path="/" element={<CompanyDashboard />} />
      <Route path="/merchants" element={<CompanyMerchants />} />
      <Route path="/transactions" element={<CompanyTransactions />} />
      <Route path="/settlements" element={<CompanySettlements />} />
      <Route path="/ledger" element={<CompanyLedger />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function AffiliateRoutes() {
  return (
    <Routes>
      <Route path="/" element={<AffiliateDashboard />} />
      <Route path="/merchants" element={<AffiliateMerchants />} />
      <Route path="/transactions" element={<AffiliateTransactions />} />
      <Route path="/ledger" element={<AffiliateLedgerPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function MerchantRoutes() {
  return (
    <Routes>
      <Route path="/" element={<MerchantDashboard />} />
      <Route path="/transactions" element={<MerchantTransactions />} />
      <Route path="/banks" element={<MerchantBanks />} />
      <Route path="/ledger" element={<MerchantLedgerPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  const { auth, ready } = useStore();

  // While a stored session token is being validated on first load, avoid
  // flashing the Login screen for an already-authenticated user.
  if (!ready && !auth) return null;

  if (!auth) return <Login />;

  const routesByRole = {
    Admin: <AdminRoutes />,
    SuperAdmin: <SuperAdminRoutes />,
    Company: <CompanyRoutes />,
    Affiliate: <AffiliateRoutes />,
    Merchant: <MerchantRoutes />,
  };

  return <Layout>{routesByRole[auth.role] || <AdminRoutes />}</Layout>;
}
