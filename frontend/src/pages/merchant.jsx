import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Stat, Empty } from '../components/ui';
import { calcTransaction, merchantLedger, inr } from '../data/calc';
import Modal from '../components/Modal';
import Pagination from '../components/Pagination';

function useMerchant() {
  const { db, auth } = useStore();
  const merchant = db.merchants.find((m) => m.id === auth.id);
  return { db, merchant };
}

export function MerchantDashboard() {
  const { db, merchant } = useMerchant();
  if (!merchant) return <Empty text="Merchant not found." />;
  const txns = db.transactions.filter((t) => t.merchantId === merchant.id);
  const company = db.companies.find((c) => c.id === merchant.companyId);
  const led = merchantLedger(merchant.id, db);
  const vol = txns.reduce((s, t) => s + t.txnAmount, 0);
  const settle = txns.reduce((s, t) => s + t.settlementAmount, 0);

  return (
    <div>
      <PageHead title={`Welcome, ${merchant.name}`} sub="Your transactions and commission at a glance." />
      <div className="stat-grid">
        <Stat label="My Transactions" value={txns.length} />
        <Stat label="Transaction Volume" value={inr(vol)} />
        <Stat label="Settlement Total" value={inr(settle)} />
        {merchant.affiliateId
          ? <Stat label="Commission" value="Via Affiliate" meta="Paid to your affiliate" invert />
          : <Stat label="Commission Earned" value={inr(led.earned)} meta={`${merchant.commissionPct}% on ${merchant.commissionBase}`} invert />}
      </div>

      <div className="panel">
        <div className="panel-head"><h2>Recent Transactions</h2></div>
        {txns.length === 0 ? <div className="panel-body"><Empty text="No transactions yet." /></div> : (
          <table className="data">
            <thead><tr><th>Date</th><th>Gateway</th><th className="num">Txn Amount</th><th className="num">Settlement</th><th className="num">My Commission</th></tr></thead>
            <tbody>
              {txns.slice(-6).reverse().map((t) => {
                const calc = calcTransaction(t, { company, merchant, affiliate: null });
                return (
                  <tr key={t.id}>
                    <td className="nowrap">{t.date}</td>
                    <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                    <td className="num mono">{inr(t.txnAmount)}</td>
                    <td className="num mono">{inr(t.settlementAmount)}</td>
                    <td className="num mono bold">{merchant.affiliateId ? '—' : inr(calc.beneficiaryCommission)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export function MerchantTransactions() {
  const { db, merchant } = useMerchant();
  const txns = db.transactions.filter((t) => t.merchantId === merchant?.id).slice().reverse();
  const company = db.companies.find((c) => c.id === merchant?.companyId);
  
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalTxns = txns.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedTxns = txns.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div>
      <PageHead title="My Transactions" sub="All transactions recorded for you." />
      <div className="panel">
        {txns.length === 0 ? <div className="panel-body"><Empty icon="⇄" text="No transactions." /></div> : (
          <>
            <table className="data">
              <thead><tr><th>Date</th><th>Gateway</th><th className="num">Txn Amt</th><th className="num">Settlement</th><th className="num">Charges</th><th className="num">My Commission</th><th>Remarks</th></tr></thead>
              <tbody>
                {paginatedTxns.map((t) => {
                  const calc = calcTransaction(t, { company, merchant, affiliate: null });
                  return (
                    <tr key={t.id}>
                      <td className="nowrap">{t.date}</td>
                      <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                      <td className="num mono">{inr(t.txnAmount)}</td>
                      <td className="num mono">{inr(t.settlementAmount)}</td>
                      <td className="num mono">{inr(t.txnCharges + (t.otherCharges || 0))}</td>
                      <td className="num mono bold">{merchant.affiliateId ? '—' : inr(calc.beneficiaryCommission)}</td>
                      <td className="muted">{t.remarks || '—'}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalTxns > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalTxns}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={(newSize) => { setItemsPerPage(newSize); setCurrentPage(1); }}
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}

export function MerchantBanks() {
  const { merchant } = useMerchant();
  const [view, setView] = useState(null);
  const [reveal, setReveal] = useState(false);
  const banks = merchant?.banks || [];

  return (
    <div>
      <PageHead title="My Banks" sub="Bank accounts on file for your business." />
      <div className="panel">
        {banks.length === 0 ? <div className="panel-body"><Empty icon="▤" text="No bank accounts on file." /></div> : (
          <table className="data">
            <thead><tr><th>Bank</th><th>Account Name</th><th>Account No.</th><th>IFSC</th><th className="center">Details</th></tr></thead>
            <tbody>
              {banks.map((b, i) => (
                <tr key={b.id || i}>
                  <td className="bold">{b.bankName}</td>
                  <td>{b.accountName}</td>
                  <td className="mono">{b.accountNumber}</td>
                  <td className="mono">{b.ifsc || '—'}</td>
                  <td className="center"><button className="btn sm" onClick={() => { setView(b); setReveal(false); }}>View</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {view && (
        <Modal title={view.bankName} onClose={() => setView(null)} footer={<button className="btn ghost" onClick={() => setView(null)}>Close</button>}>
          <div className="btn-row" style={{ marginBottom: 14 }}>
            <button className="btn sm" onClick={() => setReveal(!reveal)}>{reveal ? 'Hide' : 'Reveal'} Credentials</button>
          </div>
          <table className="data">
            <tbody>
              <Row k="Account Name" v={view.accountName} />
              <Row k="Account Number" v={view.accountNumber} />
              <Row k="IFSC Code" v={view.ifsc} />
              <Row k="Netbanking Link" v={view.netbankingLink} />
              <Row k="Username" v={view.username} />
              <Row k="Login Password" v={reveal ? view.loginPassword : '••••••••'} />
              <Row k="Transaction Password" v={reveal ? view.txnPassword : '••••••••'} />
              <Row k="Customer ID" v={view.customerId} />
              <Row k="Registered Mobile" v={view.mobile} />
              <Row k="Registered Email" v={view.email} />
              {(view.custom || []).map((c, i) => <Row key={i} k={c.key} v={c.value} />)}
            </tbody>
          </table>
        </Modal>
      )}
    </div>
  );
}

export function MerchantLedgerPage() {
  const { db, merchant } = useMerchant();
  if (!merchant) return <Empty text="Merchant not found." />;
  if (merchant.affiliateId) {
    return (
      <div>
        <PageHead title="Commission Ledger" sub="Your commission is handled by your affiliate." />
        <div className="panel"><div className="panel-body muted">You operate under an affiliate, so commission is paid to your affiliate. No direct commission ledger is maintained.</div></div>
      </div>
    );
  }
  const led = merchantLedger(merchant.id, db);
  const company = db.companies.find((c) => c.id === merchant.companyId);
  const txns = db.transactions.filter((t) => t.merchantId === merchant.id);
  const events = [
    ...txns.map((t) => {
      const calc = calcTransaction(t, { company, merchant, affiliate: null });
      return { date: t.date, label: `Commission · ${db.gateways.find((g) => g.id === t.gatewayId)?.name || ''}`, earned: calc.beneficiaryCommission, paid: 0 };
    }),
    ...(db.merchantPayments || []).filter((p) => p.merchantId === merchant.id).map((p) => ({ date: p.date, label: `Payment · ${p.paymentMode}`, earned: 0, paid: p.amount })),
  ].sort((a, b) => a.date.localeCompare(b.date));

  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalEvents = events.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedEvents = events.slice(startIndex, startIndex + itemsPerPage);

  // Calculate running balance considering pagination
  let runningStart = 0;
  for (let i = 0; i < startIndex; i++) {
    runningStart += events[i].earned - events[i].paid;
  }

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div>
      <PageHead title="Commission Ledger" sub="Your commission earnings vs payments received." />
      <div className="stat-grid" style={{ gridTemplateColumns: 'repeat(2,1fr)' }}>
        <Stat label="Total Earned" value={inr(led.earned)} />
        <Stat label="Outstanding" value={inr(led.balance)} invert />
      </div>
      <div className="panel">
        <div className="panel-head"><h2>Running Ledger</h2></div>
        <table className="data">
          <thead><tr><th>Date</th><th>Particulars</th><th className="num">Earned (+)</th><th className="num">Paid (−)</th><th className="num">Balance</th></tr></thead>
          <tbody>
            {events.length === 0 && <tr><td colSpan={5}><Empty text="No entries." /></td></tr>}
            {paginatedEvents.map((e, i) => {
              const running = runningStart + paginatedEvents.slice(0, i + 1).reduce((sum, evt) => sum + evt.earned - evt.paid, 0);
              return (
                <tr key={i}>
                  <td className="nowrap">{e.date}</td>
                  <td>{e.label}</td>
                  <td className="num mono">{e.earned ? inr(e.earned) : '—'}</td>
                  <td className="num mono">{e.paid ? inr(e.paid) : '—'}</td>
                  <td className="num mono bold">{inr(running)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {totalEvents > 10 && (
          <Pagination
            currentPage={currentPage}
            totalItems={totalEvents}
            itemsPerPage={itemsPerPage}
            onPageChange={handlePageChange}
            onItemsPerPageChange={(newSize) => { setItemsPerPage(newSize); setCurrentPage(1); }}
          />
        )}
      </div>
    </div>
  );
}

function Row({ k, v }) {
  return (
    <tr>
      <td className="bold" style={{ width: '40%' }}>{k}</td>
      <td className="mono">{v || '—'}</td>
    </tr>
  );
}
