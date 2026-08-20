import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Stat, Empty, StatusBadge } from '../components/ui';
import { calcTransaction, companyLedger, inr } from '../data/calc';
import Pagination from '../components/Pagination';

function useCompany() {
  const { db, auth } = useStore();
  const company = db.companies.find((c) => c.id === auth.id);
  return { db, company };
}

function enrich(db, t, company) {
  const merchant = db.merchants.find((m) => m.id === t.merchantId);
  const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
  return { merchant, affiliate, calc: calcTransaction(t, { company, merchant, affiliate }) };
}

export function CompanyDashboard() {
  const { db, company } = useCompany();
  if (!company) return <Empty text="Company not found." />;
  const txns = db.transactions.filter((t) => t.companyId === company.id);
  const totals = txns.reduce((acc, t) => {
    const { calc } = enrich(db, t, company);
    acc.vol += t.txnAmount; acc.settle += t.settlementAmount; acc.net += calc.companyNetIncome;
    return acc;
  }, { vol: 0, settle: 0, net: 0 });
  const led = companyLedger(company.id, db);
  const merchants = db.merchants.filter((m) => m.companyId === company.id);

  return (
    <div>
      <PageHead title={`Welcome, ${company.name}`} sub="Your transactions, settlements and balance at a glance." />
      <div className="stat-grid">
        <Stat label="My Merchants" value={merchants.length} meta={`${txns.length} transactions`} />
        <Stat label="Transaction Volume" value={inr(totals.vol)} />
        <Stat label="My Net Income" value={inr(totals.net)} meta="After commissions & charges" />
        <Stat label="Outstanding Balance" value={inr(Math.abs(led.balance))} meta={led.balance >= 0 ? 'You should receive' : 'You owe Admin'} invert />
      </div>

      <div className="panel">
        <div className="panel-head"><h2>Recent Transactions</h2></div>
        {txns.length === 0 ? <div className="panel-body"><Empty text="No transactions yet." /></div> : (
          <table className="data">
            <thead><tr><th>Date</th><th>Merchant</th><th>Gateway</th><th className="num">Txn Amount</th><th className="num">My Net Income</th></tr></thead>
            <tbody>
              {txns.slice(-6).reverse().map((t) => {
                const { merchant, calc } = enrich(db, t, company);
                return (
                  <tr key={t.id}>
                    <td className="nowrap">{t.date}</td>
                    <td>{merchant?.name || '—'}</td>
                    <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                    <td className="num mono">{inr(t.txnAmount)}</td>
                    <td className="num mono bold">{inr(calc.companyNetIncome)}</td>
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

export function CompanyMerchants() {
  const { db, company } = useCompany();
  const merchants = db.merchants.filter((m) => m.companyId === company?.id);
  
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalMerchants = merchants.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedMerchants = merchants.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div>
      <PageHead title="My Merchants" sub="Merchants assigned to your company." />
      <div className="panel">
        {merchants.length === 0 ? <div className="panel-body"><Empty icon="◈" text="No merchants assigned." /></div> : (
          <>
            <table className="data">
              <thead><tr><th>Merchant</th><th>Contact</th><th>Type</th><th>Commission</th><th>Banks</th><th>Status</th></tr></thead>
              <tbody>
                {paginatedMerchants.map((m) => (
                  <tr key={m.id}>
                    <td className="bold">{m.name}</td>
                    <td className="mono">{m.contact}</td>
                    <td>{m.affiliateId ? <span className="badge">Affiliate</span> : <span className="badge muted">Direct</span>}</td>
                    <td className="mono">{m.affiliateId ? <span className="muted">via affiliate</span> : `${m.commissionPct}%`}</td>
                    <td>{(m.banks || []).length}</td>
                    <td><StatusBadge status={m.status} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
            {totalMerchants > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalMerchants}
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

export function CompanyTransactions() {
  const { db, company } = useCompany();
  const txns = db.transactions.filter((t) => t.companyId === company?.id).slice().reverse();
  
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
      <PageHead title="Transactions" sub="All transactions recorded for your company." />
      <div className="panel">
        {txns.length === 0 ? <div className="panel-body"><Empty icon="⇄" text="No transactions." /></div> : (
          <>
            <table className="data">
              <thead><tr><th>Date</th><th>Merchant</th><th>Gateway</th><th className="num">Txn Amt</th><th className="num">Settlement</th><th className="num">Admin Comm.</th><th className="num">My Net</th></tr></thead>
              <tbody>
                {paginatedTxns.map((t) => {
                  const { merchant, calc } = enrich(db, t, company);
                  return (
                    <tr key={t.id}>
                      <td className="nowrap">{t.date}</td>
                      <td>{merchant?.name || '—'}</td>
                      <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                      <td className="num mono">{inr(t.txnAmount)}</td>
                      <td className="num mono">{inr(t.settlementAmount)}</td>
                      <td className="num mono">{inr(calc.adminCommission)}</td>
                      <td className="num mono bold">{inr(calc.companyNetIncome)}</td>
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

export function CompanySettlements() {
  const { db, company } = useCompany();
  const pays = db.settlements.filter((s) => s.companyId === company?.id).slice().reverse();
  
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalPays = pays.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedPays = pays.slice(startIndex, startIndex + itemsPerPage);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div>
      <PageHead title="Settlements" sub="Payments received from Admin." />
      <div className="panel">
        {pays.length === 0 ? <div className="panel-body"><Empty icon="⇣" text="No settlement payments yet." /></div> : (
          <>
            <table className="data">
              <thead><tr><th>Date</th><th className="num">Amount</th><th>Mode</th><th>Reference</th><th>Remarks</th></tr></thead>
              <tbody>
                {paginatedPays.map((s) => (
                  <tr key={s.id}>
                    <td className="nowrap">{s.date}</td>
                    <td className="num mono">{inr(s.amount)}</td>
                    <td><span className="badge muted">{s.paymentMode}</span></td>
                    <td className="mono">{s.refNumber || '—'}</td>
                    <td className="muted">{s.remarks || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {totalPays > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalPays}
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

export function CompanyLedger() {
  const { db, company } = useCompany();
  if (!company) return <Empty text="Company not found." />;
  const led = companyLedger(company.id, db);
  const txns = db.transactions.filter((t) => t.companyId === company.id);
  const events = [
    ...txns.map((t) => {
      const { merchant, calc } = enrich(db, t, company);
      return { date: t.date, label: `Txn · ${merchant?.name || ''}`, debit: calc.companyNetIncome, credit: 0, ref: t.remarks };
    }),
    ...db.settlements.filter((s) => s.companyId === company.id).map((s) => ({ date: s.date, label: `Payment · ${s.paymentMode}`, debit: 0, credit: s.amount, ref: s.refNumber })),
  ].sort((a, b) => a.date.localeCompare(b.date));

  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalEvents = events.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedEvents = events.slice(startIndex, startIndex + itemsPerPage);

  // Calculate running balance considering pagination
  let runningStart = 0;
  for (let i = 0; i < startIndex; i++) {
    runningStart += events[i].debit - events[i].credit;
  }

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div>
      <PageHead title="My Ledger" sub="Running balance of receivables and payments." />
      <div className="stat-grid" style={{ gridTemplateColumns: 'repeat(3,1fr)' }}>
        <Stat label="Total Receivable" value={inr(led.receivable)} />
        <Stat label="Total Received" value={inr(led.paid)} />
        <Stat label="Outstanding" value={inr(Math.abs(led.balance))} meta={led.balance >= 0 ? 'You should receive' : 'You owe Admin'} invert />
      </div>
      <div className="panel">
        <div className="panel-head"><h2>Running Ledger</h2></div>
        <table className="data">
          <thead><tr><th>Date</th><th>Particulars</th><th>Reference</th><th className="num">Receivable (+)</th><th className="num">Received (−)</th><th className="num">Balance</th></tr></thead>
          <tbody>
            {events.length === 0 && <tr><td colSpan={6}><Empty text="No entries." /></td></tr>}
            {paginatedEvents.map((e, i) => {
              const running = runningStart + paginatedEvents.slice(0, i + 1).reduce((sum, evt) => sum + evt.debit - evt.credit, 0);
              return (
                <tr key={i}>
                  <td className="nowrap">{e.date}</td>
                  <td>{e.label}</td>
                  <td className="muted mono">{e.ref || '—'}</td>
                  <td className="num mono">{e.debit ? inr(e.debit) : '—'}</td>
                  <td className="num mono">{e.credit ? inr(e.credit) : '—'}</td>
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
