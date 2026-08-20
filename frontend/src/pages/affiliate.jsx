import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Stat, Empty, StatusBadge } from '../components/ui';
import { calcTransaction, affiliateLedger, inr } from '../data/calc';
import Pagination from '../components/Pagination';

function useAffiliate() {
  const { db, auth } = useStore();
  const affiliate = db.affiliates.find((a) => a.id === auth.id);
  const merchantIds = db.merchants.filter((m) => m.affiliateId === auth.id).map((m) => m.id);
  return { db, affiliate, merchantIds };
}

export function AffiliateDashboard() {
  const { db, affiliate, merchantIds } = useAffiliate();
  if (!affiliate) return <Empty text="Affiliate not found." />;
  const txns = db.transactions.filter((t) => merchantIds.includes(t.merchantId));
  const led = affiliateLedger(affiliate.id, db);
  const vol = txns.reduce((s, t) => s + t.txnAmount, 0);

  return (
    <div>
      <PageHead title={`Welcome, ${affiliate.name}`} sub="Your merchants and commission earnings." />
      <div className="stat-grid">
        <Stat label="My Merchants" value={merchantIds.length} meta={`${txns.length} transactions`} />
        <Stat label="Transaction Volume" value={inr(vol)} />
        <Stat label="Commission Earned" value={inr(led.earned)} meta={`${affiliate.commissionPct}% on ${affiliate.commissionBase}`} />
        <Stat label="Outstanding" value={inr(led.balance)} meta="Yet to be paid" invert />
      </div>

      <div className="panel">
        <div className="panel-head"><h2>Recent Transactions</h2></div>
        {txns.length === 0 ? <div className="panel-body"><Empty text="No transactions yet." /></div> : (
          <table className="data">
            <thead><tr><th>Date</th><th>Merchant</th><th>Gateway</th><th className="num">Txn Amount</th><th className="num">My Commission</th></tr></thead>
            <tbody>
              {txns.slice(-6).reverse().map((t) => {
                const merchant = db.merchants.find((m) => m.id === t.merchantId);
                const company = db.companies.find((c) => c.id === t.companyId);
                const calc = calcTransaction(t, { company, merchant, affiliate });
                return (
                  <tr key={t.id}>
                    <td className="nowrap">{t.date}</td>
                    <td>{merchant?.name || '—'}</td>
                    <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                    <td className="num mono">{inr(t.txnAmount)}</td>
                    <td className="num mono bold">{inr(calc.beneficiaryCommission)}</td>
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

export function AffiliateMerchants() {
  const { db, merchantIds } = useAffiliate();
  const merchants = db.merchants.filter((m) => merchantIds.includes(m.id));
  return (
    <div>
      <PageHead title="My Merchants" sub="Merchants operating under you." />
      <div className="panel">
        {merchants.length === 0 ? <div className="panel-body"><Empty icon="◈" text="No merchants under you." /></div> : (
          <table className="data">
            <thead><tr><th>Merchant</th><th>Contact</th><th>Status</th></tr></thead>
            <tbody>
              {merchants.map((m) => (
                <tr key={m.id}>
                  <td className="bold">{m.name}</td>
                  <td className="mono">{m.contact}</td>
                  <td><StatusBadge status={m.status} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export function AffiliateTransactions() {
  const { db, affiliate, merchantIds } = useAffiliate();
  const txns = db.transactions.filter((t) => merchantIds.includes(t.merchantId)).slice().reverse();
  
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
      <PageHead title="Transactions" sub="Transactions from your merchants." />
      <div className="panel">
        {txns.length === 0 ? <div className="panel-body"><Empty icon="⇄" text="No transactions." /></div> : (
          <>
            <table className="data">
              <thead><tr><th>Date</th><th>Merchant</th><th>Gateway</th><th className="num">Txn Amt</th><th className="num">Settlement</th><th className="num">My Commission</th></tr></thead>
              <tbody>
                {paginatedTxns.map((t) => {
                  const merchant = db.merchants.find((m) => m.id === t.merchantId);
                  const company = db.companies.find((c) => c.id === t.companyId);
                  const calc = calcTransaction(t, { company, merchant, affiliate });
                  return (
                    <tr key={t.id}>
                      <td className="nowrap">{t.date}</td>
                      <td>{merchant?.name || '—'}</td>
                      <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                      <td className="num mono">{inr(t.txnAmount)}</td>
                      <td className="num mono">{inr(t.settlementAmount)}</td>
                      <td className="num mono bold">{inr(calc.beneficiaryCommission)}</td>
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

export function AffiliateLedgerPage() {
  const { db, affiliate, merchantIds } = useAffiliate();
  if (!affiliate) return <Empty text="Affiliate not found." />;
  const led = affiliateLedger(affiliate.id, db);
  const txns = db.transactions.filter((t) => merchantIds.includes(t.merchantId));
  const events = [
    ...txns.map((t) => {
      const merchant = db.merchants.find((m) => m.id === t.merchantId);
      const company = db.companies.find((c) => c.id === t.companyId);
      const calc = calcTransaction(t, { company, merchant, affiliate });
      return { date: t.date, label: `Commission · ${merchant?.name || ''}`, earned: calc.beneficiaryCommission, paid: 0 };
    }),
    ...(db.affiliatePayments || []).filter((p) => p.affiliateId === affiliate.id).map((p) => ({ date: p.date, label: `Payment · ${p.paymentMode}`, earned: 0, paid: p.amount })),
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
      <PageHead title="Commission Ledger" sub="Earnings vs payments received." />
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
