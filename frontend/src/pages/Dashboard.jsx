import { useNavigate } from 'react-router-dom';
import { useStore } from '../data/store';
import { PageHead, Stat } from '../components/ui';
import { calcTransaction, companyLedger, inr } from '../data/calc';

export default function Dashboard() {
  const { db, resetData } = useStore();
  const nav = useNavigate();

  const totals = db.transactions.reduce((acc, t) => {
    const merchant = db.merchants.find((m) => m.id === t.merchantId);
    const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
    const company = db.companies.find((c) => c.id === t.companyId);
    const r = calcTransaction(t, { company, merchant, affiliate });
    acc.txnAmount += t.txnAmount;
    acc.settlement += t.settlementAmount;
    acc.adminComm += r.adminCommission;
    acc.adminNet += r.adminNetCommission;
    acc.benefComm += r.beneficiaryCommission;
    acc.net += r.companyNetIncome;
    return acc;
  }, { txnAmount: 0, settlement: 0, adminComm: 0, adminNet: 0, benefComm: 0, net: 0 });

  const totalPaid = db.settlements.reduce((s, x) => s + Number(x.amount), 0);
  const outstanding = db.companies.reduce((s, c) => s + companyLedger(c.id, db).balance, 0);

  const recent = db.transactions.slice(-5).reverse();

  return (
    <div>
      <PageHead
        title="Dashboard"
        sub="Operational overview of commissions and settlements."
        actions={<button className="btn ghost" onClick={() => { if (window.confirm('Reset all demo data to defaults?')) resetData(); }}>Reset Demo Data</button>}
      />

      <div className="stat-grid">
        <Stat label="Companies" value={db.companies.length} meta={`${db.merchants.length} merchants · ${db.affiliates.length} affiliates`} />
        <Stat label="Transaction Volume" value={inr(totals.txnAmount)} meta={`${db.transactions.length} transactions`} />
        <Stat label="Admin Net Income" value={inr(totals.adminNet)} meta="Gateway commission after merchant/affiliate payout" />
        <Stat label="Outstanding" value={inr(outstanding)} meta="Across all companies" invert />
      </div>

      <div className="stat-grid">
        <Stat label="Total Settlement Amount" value={inr(totals.settlement)} />
        <Stat label="Company Net Income" value={inr(totals.net)} />
        <Stat label="Merchant/Affiliate Comm." value={inr(totals.benefComm)} />
        <Stat label="Payments Made" value={inr(totalPaid)} />
      </div>

      <div className="panel" style={{ marginBottom: 24 }}>
        <div className="panel-head">
          <h2>Recent Transactions</h2>
          <button className="btn sm" onClick={() => nav('/transactions')}>View All</button>
        </div>
        {recent.length === 0 ? (
          <div className="panel-body muted">No transactions yet.</div>
        ) : (
          <table className="data">
            <thead>
              <tr><th>Date</th><th>Company</th><th>Merchant</th><th>Gateway</th><th className="num">Txn Amount</th><th className="num">Company Net</th></tr>
            </thead>
            <tbody>
              {recent.map((t) => {
                const merchant = db.merchants.find((m) => m.id === t.merchantId);
                const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
                const company = db.companies.find((c) => c.id === t.companyId);
                const r = calcTransaction(t, { company, merchant, affiliate });
                return (
                  <tr key={t.id}>
                    <td className="nowrap">{t.date}</td>
                    <td>{company?.name || '—'}</td>
                    <td>{merchant?.name || '—'}</td>
                    <td>{db.gateways.find((g) => g.id === t.gatewayId)?.name || '—'}</td>
                    <td className="num mono">{inr(t.txnAmount)}</td>
                    <td className="num mono bold">{inr(r.companyNetIncome)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      <div className="panel">
        <div className="panel-head"><h2>Company Outstanding Balances</h2><button className="btn sm" onClick={() => nav('/ledgers')}>Open Ledgers</button></div>
        <table className="data">
          <thead><tr><th>Company</th><th className="num">Receivable</th><th className="num">Paid</th><th className="num">Balance</th><th>Status</th></tr></thead>
          <tbody>
            {db.companies.map((c) => {
              const l = companyLedger(c.id, db);
              return (
                <tr key={c.id}>
                  <td className="bold">{c.name}</td>
                  <td className="num mono">{inr(l.receivable)}</td>
                  <td className="num mono">{inr(l.paid)}</td>
                  <td className="num mono bold">{inr(Math.abs(l.balance))}</td>
                  <td>
                    {l.balance > 0.0001 ? <span className="badge">Should receive</span>
                      : l.balance < -0.0001 ? <span className="badge solid">Owes Admin</span>
                      : <span className="badge muted">Settled</span>}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
