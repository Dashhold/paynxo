import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, Empty } from '../components/ui';
import { companyLedger, affiliateLedger, merchantLedger, calcTransaction, inr } from '../data/calc';
import Pagination from '../components/Pagination';

export default function Ledgers() {
  const { db } = useStore();
  const [tab, setTab] = useState('company');
  const [selected, setSelected] = useState(null);

  return (
    <div>
      <PageHead title="Ledgers" sub="Running balances for companies, affiliates and merchants." />

      <div className="tabs">
        <div className={`tab ${tab === 'company' ? 'active' : ''}`} onClick={() => { setTab('company'); setSelected(null); }}>Company Ledger</div>
        <div className={`tab ${tab === 'affiliate' ? 'active' : ''}`} onClick={() => { setTab('affiliate'); setSelected(null); }}>Affiliate Ledger</div>
        <div className={`tab ${tab === 'merchant' ? 'active' : ''}`} onClick={() => { setTab('merchant'); setSelected(null); }}>Merchant Ledger</div>
      </div>

      {tab === 'company' && <CompanyLedgers db={db} selected={selected} setSelected={setSelected} />}
      {tab === 'affiliate' && <AffiliateLedgers db={db} />}
      {tab === 'merchant' && <MerchantLedgers db={db} />}
    </div>
  );
}

function CompanyLedgers({ db, selected, setSelected }) {
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  if (selected) {
    const co = db.companies.find((c) => c.id === selected);
    const txns = db.transactions.filter((t) => t.companyId === selected).slice().sort((a, b) => a.date.localeCompare(b.date));
    const pays = db.settlements.filter((s) => s.companyId === selected).map((s) => ({ ...s, kind: 'payment' }));
    // Build chronological running balance
    const events = [
      ...txns.map((t) => {
        const merchant = db.merchants.find((m) => m.id === t.merchantId);
        const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
        const r = calcTransaction(t, { company: co, merchant, affiliate });
        return { date: t.date, kind: 'txn', label: `Txn · ${merchant?.name || ''}`, debit: r.companyNetIncome, credit: 0, ref: t.remarks };
      }),
      ...pays.map((s) => ({ date: s.date, kind: 'pay', label: `Payment · ${s.paymentMode}`, debit: 0, credit: s.amount, ref: s.refNumber })),
    ].sort((a, b) => a.date.localeCompare(b.date));

    let running = 0;
    const led = companyLedger(selected, db);

    // Paginate events
    const totalEvents = events.length;
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const paginatedEvents = events.slice(startIndex, endIndex);

    return (
      <div>
        <button className="btn sm ghost" style={{ marginBottom: 16 }} onClick={() => setSelected(null)}>← Back to companies</button>
        <div className="stat-grid" style={{ gridTemplateColumns: 'repeat(3,1fr)' }}>
          <Stat label="Total Receivable" value={inr(led.receivable)} />
          <Stat label="Total Paid" value={inr(led.paid)} />
          <StatStatus balance={led.balance} who={co?.name} />
        </div>
        <div className="panel">
          <div className="panel-head"><h2>{co?.name} — Running Ledger</h2></div>
          <table className="data">
            <thead>
              <tr><th>Date</th><th>Particulars</th><th>Reference</th><th className="num">Receivable (+)</th><th className="num">Paid (−)</th><th className="num">Balance</th></tr>
            </thead>
            <tbody>
              {events.length === 0 && <tr><td colSpan={6}><Empty text="No ledger entries." /></td></tr>}
              {paginatedEvents.map((e, i) => {
                // Calculate running balance from start to current position
                const actualIndex = startIndex + i;
                running = events.slice(0, actualIndex + 1).reduce((sum, ev) => sum + ev.debit - ev.credit, 0);
                return (
                  <tr key={actualIndex}>
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
              onPageChange={(page) => { setCurrentPage(page); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
              onItemsPerPageChange={(newItems) => { setItemsPerPage(newItems); setCurrentPage(1); }}
            />
          )}
        </div>
      </div>
    );
  }

  // List view
  const totalCompanies = db.companies.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedCompanies = db.companies.slice(startIndex, endIndex);

  return (
    <div className="panel">
      <table className="data">
        <thead>
          <tr><th>Company</th><th className="num">Receivable</th><th className="num">Paid</th><th className="num">Balance</th><th>Status</th><th className="center">Ledger</th></tr>
        </thead>
        <tbody>
          {paginatedCompanies.map((c) => {
            const l = companyLedger(c.id, db);
            return (
              <tr key={c.id}>
                <td className="bold">{c.name}</td>
                <td className="num mono">{inr(l.receivable)}</td>
                <td className="num mono">{inr(l.paid)}</td>
                <td className="num mono bold">{inr(Math.abs(l.balance))}</td>
                <td>{statusText(l.balance, c.name)}</td>
                <td className="center"><button className="btn sm" onClick={() => setSelected(c.id)}>View</button></td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {totalCompanies > 10 && (
        <Pagination
          currentPage={currentPage}
          totalItems={totalCompanies}
          itemsPerPage={itemsPerPage}
          onPageChange={(page) => { setCurrentPage(page); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
          onItemsPerPageChange={(newItems) => { setItemsPerPage(newItems); setCurrentPage(1); }}
        />
      )}
    </div>
  );
}

function AffiliateLedgers({ db }) {
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalAffiliates = db.affiliates.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedAffiliates = db.affiliates.slice(startIndex, endIndex);

  return (
    <div className="panel">
      <table className="data">
        <thead>
          <tr><th>Affiliate</th><th className="num">Commission Earned</th><th className="num">Total Paid</th><th className="num">Outstanding</th></tr>
        </thead>
        <tbody>
          {db.affiliates.length === 0 && <tr><td colSpan={4}><Empty text="No affiliates." /></td></tr>}
          {paginatedAffiliates.map((a) => {
            const l = affiliateLedger(a.id, db);
            return (
              <tr key={a.id}>
                <td className="bold">{a.name}</td>
                <td className="num mono">{inr(l.earned)}</td>
                <td className="num mono">{inr(l.paid)}</td>
                <td className="num mono bold">{inr(l.balance)}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {totalAffiliates > 10 && (
        <Pagination
          currentPage={currentPage}
          totalItems={totalAffiliates}
          itemsPerPage={itemsPerPage}
          onPageChange={(page) => { setCurrentPage(page); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
          onItemsPerPageChange={(newItems) => { setItemsPerPage(newItems); setCurrentPage(1); }}
        />
      )}
    </div>
  );
}

function MerchantLedgers({ db }) {
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const totalMerchants = db.merchants.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedMerchants = db.merchants.slice(startIndex, endIndex);

  return (
    <div className="panel">
      <table className="data">
        <thead>
          <tr><th>Merchant</th><th>Type</th><th className="num">Commission Earned</th><th className="num">Total Paid</th><th className="num">Outstanding</th></tr>
        </thead>
        <tbody>
          {paginatedMerchants.map((m) => {
            const l = merchantLedger(m.id, db);
            return (
              <tr key={m.id}>
                <td className="bold">{m.name}</td>
                <td>{m.affiliateId ? <span className="badge">Affiliate</span> : <span className="badge muted">Direct</span>}</td>
                <td className="num mono">{m.affiliateId ? <span className="muted">via affiliate</span> : inr(l.earned)}</td>
                <td className="num mono">{inr(l.paid)}</td>
                <td className="num mono bold">{m.affiliateId ? '—' : inr(l.balance)}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {totalMerchants > 10 && (
        <Pagination
          currentPage={currentPage}
          totalItems={totalMerchants}
          itemsPerPage={itemsPerPage}
          onPageChange={(page) => { setCurrentPage(page); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
          onItemsPerPageChange={(newItems) => { setItemsPerPage(newItems); setCurrentPage(1); }}
        />
      )}
    </div>
  );
}

function Stat({ label, value }) {
  return <div className="stat"><div className="label">{label}</div><div className="value">{value}</div></div>;
}
function StatStatus({ balance, who }) {
  return (
    <div className="stat invert">
      <div className="label">Status</div>
      <div className="value" style={{ fontSize: 18 }}>{inr(Math.abs(balance))}</div>
      <div className="meta">{statusMessage(balance, who)}</div>
    </div>
  );
}

function statusMessage(balance, who) {
  if (balance > 0.0001) return `${who} should receive`;
  if (balance < -0.0001) return `${who} owes Admin`;
  return 'Settled';
}
function statusText(balance, who) {
  if (balance > 0.0001) return <span className="badge">Receivable</span>;
  if (balance < -0.0001) return <span className="badge solid">Owes Admin</span>;
  return <span className="badge muted">Settled</span>;
}
