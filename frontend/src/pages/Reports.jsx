import { useState, useMemo } from 'react';
import { useStore } from '../data/store';
import { PageHead, Empty } from '../components/ui';
import { calcTransaction, companyLedger, affiliateLedger, inr } from '../data/calc';
import { exportCSV, exportPDF } from '../data/export';
import Pagination from '../components/Pagination';

const TABS = [
  { id: 'company', label: 'Company Wise' },
  { id: 'merchant', label: 'Merchant Wise' },
  { id: 'affiliate', label: 'Affiliate Wise' },
  { id: 'gateway', label: 'Gateway Wise' },
  { id: 'settlement', label: 'Settlement' },
  { id: 'outstanding', label: 'Outstanding' },
];

export default function Reports() {
  const { db } = useStore();
  const [tab, setTab] = useState('company');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const txns = useMemo(() => db.transactions.filter((t) => {
    if (from && t.date < from) return false;
    if (to && t.date > to) return false;
    return true;
  }), [db.transactions, from, to]);

  const enrich = (t) => {
    const merchant = db.merchants.find((m) => m.id === t.merchantId);
    const affiliate = db.affiliates.find((a) => a.id === merchant?.affiliateId);
    const company = db.companies.find((c) => c.id === t.companyId);
    return { t, merchant, affiliate, company, calc: calcTransaction(t, { company, merchant, affiliate }) };
  };

  const { columns, rows, title } = useMemo(() => buildReport(tab, db, txns, enrich), [tab, db, txns]);

  // Paginated data
  const totalRows = rows.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedRows = rows.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1);
  };

  const handleTabChange = (newTab) => {
    setTab(newTab);
    setCurrentPage(1); // Reset to first page when tab changes
  };

  return (
    <div>
      <PageHead
        title="Reports"
        sub="Filter, view and export operational reports."
        actions={
          <div className="btn-row">
            <button className="btn ghost" onClick={() => exportCSV(title, columns, rows)}>⬇ Excel (CSV)</button>
            <button className="btn primary" onClick={() => exportPDF(title, columns, rows)}>⬇ PDF</button>
          </div>
        }
      />

      <div className="toolbar">
        <div className="filters">
          <label className="field" style={{ marginBottom: 0 }}>
            <span>From Date</span>
            <input type="date" value={from} onChange={(e) => { setFrom(e.target.value); setCurrentPage(1); }} />
          </label>
          <label className="field" style={{ marginBottom: 0 }}>
            <span>To Date</span>
            <input type="date" value={to} onChange={(e) => { setTo(e.target.value); setCurrentPage(1); }} />
          </label>
          {(from || to) && <button className="btn sm ghost" style={{ alignSelf: 'flex-end' }} onClick={() => { setFrom(''); setTo(''); setCurrentPage(1); }}>Clear</button>}
        </div>
        <span className="muted">{totalRows} rows</span>
      </div>

      <div className="tabs">
        {TABS.map((t) => (
          <div key={t.id} className={`tab ${tab === t.id ? 'active' : ''}`} onClick={() => handleTabChange(t.id)}>{t.label}</div>
        ))}
      </div>

      <div className="panel">
        {rows.length === 0 ? (
          <div className="panel-body"><Empty icon="◫" text="No data for the selected report / date range." /></div>
        ) : (
          <>
            <table className="data">
              <thead><tr>{columns.map((c) => <th key={c.key} className={c.num ? 'num' : ''}>{c.label}</th>)}</tr></thead>
              <tbody>
                {paginatedRows.map((r, i) => (
                  <tr key={i}>{columns.map((c) => <td key={c.key} className={c.num ? 'num mono' : ''}>{r[c.key]}</td>)}</tr>
                ))}
              </tbody>
            </table>
            {totalRows > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalRows}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={handleItemsPerPageChange}
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}

function buildReport(tab, db, txns, enrich) {
  if (tab === 'company') {
    const map = {};
    txns.forEach((t) => {
      const { calc } = enrich(t);
      const co = db.companies.find((c) => c.id === t.companyId);
      const name = co?.name || '—';
      map[name] = map[name] || { company: name, count: 0, txnAmount: 0, settlement: 0, adminComm: 0, net: 0 };
      map[name].count += 1;
      map[name].txnAmount += t.txnAmount;
      map[name].settlement += t.settlementAmount;
      map[name].adminComm += calc.adminCommission;
      map[name].net += calc.companyNetIncome;
    });
    const rows = Object.values(map).map((r) => {
      const co = db.companies.find((c) => c.name === r.company);
      const led = co ? companyLedger(co.id, db) : { paid: 0, balance: 0 };
      return {
        company: r.company, count: r.count,
        txnAmount: inr(r.txnAmount), settlement: inr(r.settlement),
        adminComm: inr(r.adminComm), net: inr(r.net),
        paid: inr(led.paid), outstanding: inr(led.balance),
      };
    });
    return {
      title: 'Company Wise Report',
      columns: [
        { key: 'company', label: 'Company' },
        { key: 'count', label: 'Txns', num: true },
        { key: 'txnAmount', label: 'Txn Amount', num: true },
        { key: 'settlement', label: 'Settlement', num: true },
        { key: 'adminComm', label: 'Admin Comm.', num: true },
        { key: 'net', label: 'Company Net', num: true },
        { key: 'paid', label: 'Paid', num: true },
        { key: 'outstanding', label: 'Outstanding', num: true },
      ],
      rows,
    };
  }

  if (tab === 'merchant') {
    const map = {};
    txns.forEach((t) => {
      const { calc, merchant } = enrich(t);
      const name = merchant?.name || '—';
      map[name] = map[name] || { merchant: name, count: 0, txnAmount: 0, settlement: 0, comm: 0 };
      map[name].count += 1;
      map[name].txnAmount += t.txnAmount;
      map[name].settlement += t.settlementAmount;
      if (!merchant?.affiliateId) map[name].comm += calc.beneficiaryCommission;
    });
    const rows = Object.values(map).map((r) => ({
      merchant: r.merchant, count: r.count,
      txnAmount: inr(r.txnAmount), settlement: inr(r.settlement), comm: inr(r.comm),
    }));
    return {
      title: 'Merchant Wise Report',
      columns: [
        { key: 'merchant', label: 'Merchant' },
        { key: 'count', label: 'Txns', num: true },
        { key: 'txnAmount', label: 'Txn Amount', num: true },
        { key: 'settlement', label: 'Settlement', num: true },
        { key: 'comm', label: 'Commission', num: true },
      ],
      rows,
    };
  }

  if (tab === 'affiliate') {
    const rows = db.affiliates.map((a) => {
      const led = affiliateLedger(a.id, db);
      const count = db.transactions.filter((t) => db.merchants.find((m) => m.id === t.merchantId)?.affiliateId === a.id).length;
      return { affiliate: a.name, count, earned: inr(led.earned), paid: inr(led.paid), balance: inr(led.balance) };
    });
    return {
      title: 'Affiliate Wise Report',
      columns: [
        { key: 'affiliate', label: 'Affiliate' },
        { key: 'count', label: 'Txns', num: true },
        { key: 'earned', label: 'Commission', num: true },
        { key: 'paid', label: 'Paid', num: true },
        { key: 'balance', label: 'Balance', num: true },
      ],
      rows,
    };
  }

  if (tab === 'gateway') {
    const map = {};
    txns.forEach((t) => {
      const { calc } = enrich(t);
      const gw = db.gateways.find((g) => g.id === t.gatewayId);
      const name = gw?.name || '—';
      map[name] = map[name] || { gateway: name, count: 0, txnAmount: 0, adminComm: 0 };
      map[name].count += 1;
      map[name].txnAmount += t.txnAmount;
      map[name].adminComm += calc.adminCommission;
    });
    const rows = Object.values(map).map((r) => ({
      gateway: r.gateway, count: r.count, txnAmount: inr(r.txnAmount), adminComm: inr(r.adminComm),
    }));
    return {
      title: 'Gateway Wise Report',
      columns: [
        { key: 'gateway', label: 'Gateway' },
        { key: 'count', label: 'Txns', num: true },
        { key: 'txnAmount', label: 'Txn Amount', num: true },
        { key: 'adminComm', label: 'Admin Commission', num: true },
      ],
      rows,
    };
  }

  if (tab === 'settlement') {
    const rows = db.settlements
      .filter((s) => (!txns.length && !s) ? false : true)
      .filter((s) => true)
      .map((s) => ({
        date: s.date,
        company: db.companies.find((c) => c.id === s.companyId)?.name || '—',
        amount: inr(s.amount),
        mode: s.paymentMode,
        ref: s.refNumber || '—',
      }));
    return {
      title: 'Settlement Report',
      columns: [
        { key: 'date', label: 'Date' },
        { key: 'company', label: 'Company' },
        { key: 'amount', label: 'Amount', num: true },
        { key: 'mode', label: 'Mode' },
        { key: 'ref', label: 'Reference' },
      ],
      rows,
    };
  }

  // outstanding
  const rows = db.companies.map((c) => {
    const led = companyLedger(c.id, db);
    let status = 'Settled';
    if (led.balance > 0.0001) status = `${c.name} should receive`;
    else if (led.balance < -0.0001) status = `${c.name} owes Admin`;
    return {
      company: c.name, receivable: inr(led.receivable), paid: inr(led.paid),
      outstanding: inr(led.balance), status,
    };
  });
  return {
    title: 'Outstanding Report',
    columns: [
      { key: 'company', label: 'Company' },
      { key: 'receivable', label: 'Receivable', num: true },
      { key: 'paid', label: 'Paid', num: true },
      { key: 'outstanding', label: 'Outstanding', num: true },
      { key: 'status', label: 'Status' },
    ],
    rows,
  };
}
