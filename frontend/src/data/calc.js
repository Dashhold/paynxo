// Commission Calculation Engine
// Implements the rules from the spec.

export function inr(n) {
  if (n === null || n === undefined || isNaN(n)) return '₹0';
  const neg = n < 0;
  const v = Math.abs(Number(n));
  const formatted = v.toLocaleString('en-IN', { maximumFractionDigits: 2 });
  return (neg ? '-₹' : '₹') + formatted;
}

export function num(n) {
  const v = Number(n);
  return isNaN(v) ? 0 : v;
}

/**
 * Calculate commissions and company net income for a single transaction.
 *
 * @param {object} txn   transaction record
 * @param {object} ctx   { company, merchant, affiliate }
 * @returns full breakdown object
 */
export function calcTransaction(txn, ctx) {
  const { company, merchant, affiliate } = ctx;
  const txnAmount = num(txn.txnAmount);
  const settlementAmount = num(txn.settlementAmount);
  const txnCharges = num(txn.txnCharges);
  const otherCharges = num(txn.otherCharges);

  // Gateway commission config from the company assignment
  const gwAssign = (company?.gateways || []).find((g) => g.gatewayId === txn.gatewayId);
  const gatewayCommissionPct = gwAssign ? num(gwAssign.commission) : 0;
  const chargeBearer = gwAssign ? gwAssign.chargeBearer : 'Admin';

  // Admin commission = gateway % on TRANSACTION amount
  const adminCommission = (txnAmount * gatewayCommissionPct) / 100;

  // Merchant / Affiliate commission
  let beneficiary = 'Merchant';
  let beneficiaryPct = 0;
  let beneficiaryBase = 'Settlement Amount';

  if (merchant?.affiliateId && affiliate) {
    beneficiary = 'Affiliate';
    beneficiaryPct = num(affiliate.commissionPct);
    beneficiaryBase = affiliate.commissionBase || 'Settlement Amount';
  } else if (merchant) {
    beneficiary = 'Merchant';
    beneficiaryPct = num(merchant.commissionPct);
    beneficiaryBase = merchant.commissionBase || 'Settlement Amount';
  }

  const baseAmountForBeneficiary =
    beneficiaryBase === 'Transaction Amount' ? txnAmount : settlementAmount;
  const beneficiaryCommission = (baseAmountForBeneficiary * beneficiaryPct) / 100;

  // ── Commission model ──────────────────────────────────────────────
  // 1. Admin earns the gateway commission FROM the company.
  //    Company Net Income = Settlement Amount − Admin Commission
  //    (Merchant/Affiliate commission is NOT charged to the company.)
  // 2. The Merchant/Affiliate commission is paid OUT OF the Admin's commission.
  //    Admin Net Income = Admin Commission − Merchant/Affiliate Commission
  // 3. Transaction/other charges are deducted from whoever is the charge bearer.

  let companyChargesDeducted = 0;
  let companyNetIncome = settlementAmount - adminCommission;
  if (chargeBearer === 'Company') {
    companyChargesDeducted = txnCharges + otherCharges;
    companyNetIncome -= companyChargesDeducted;
  }

  let adminChargesDeducted = 0;
  let adminNetCommission = adminCommission - beneficiaryCommission;
  if (chargeBearer === 'Admin') {
    adminChargesDeducted = txnCharges + otherCharges;
    adminNetCommission -= adminChargesDeducted;
  }

  return {
    txnAmount,
    settlementAmount,
    txnCharges,
    otherCharges,
    gatewayCommissionPct,
    chargeBearer,
    adminCommission,
    beneficiary,
    beneficiaryPct,
    beneficiaryBase,
    beneficiaryCommission,
    chargesDeducted: companyChargesDeducted, // kept for backward compatibility
    companyChargesDeducted,
    adminChargesDeducted,
    adminNetCommission,
    companyNetIncome,
  };
}

/**
 * Build a company ledger: receivable (company net income) vs payments made.
 */
export function companyLedger(companyId, db) {
  const company = db.companies.find((c) => c.id === companyId);
  const txns = db.transactions.filter((t) => t.companyId === companyId);
  let receivable = 0;
  txns.forEach((t) => {
    const r = calcTransaction(t, {
      company,
      merchant: db.merchants.find((m) => m.id === t.merchantId),
      affiliate: db.affiliates.find((a) => a.id === (db.merchants.find((m) => m.id === t.merchantId)?.affiliateId)),
    });
    receivable += r.companyNetIncome;
  });
  const paid = db.settlements
    .filter((s) => s.companyId === companyId)
    .reduce((sum, s) => sum + num(s.amount), 0);

  const balance = receivable - paid;
  return { receivable, paid, balance };
}

/**
 * Affiliate ledger: total commission earned across all their merchants' txns.
 */
export function affiliateLedger(affiliateId, db) {
  const affiliate = db.affiliates.find((a) => a.id === affiliateId);
  const merchantIds = db.merchants.filter((m) => m.affiliateId === affiliateId).map((m) => m.id);
  const txns = db.transactions.filter((t) => merchantIds.includes(t.merchantId));
  let earned = 0;
  txns.forEach((t) => {
    const merchant = db.merchants.find((m) => m.id === t.merchantId);
    const company = db.companies.find((c) => c.id === t.companyId);
    const r = calcTransaction(t, { company, merchant, affiliate });
    earned += r.beneficiaryCommission;
  });
  const paid = (db.affiliatePayments || [])
    .filter((p) => p.affiliateId === affiliateId)
    .reduce((sum, p) => sum + num(p.amount), 0);
  return { earned, paid, balance: earned - paid };
}

/**
 * Merchant ledger: commission earned (only for direct merchants).
 */
export function merchantLedger(merchantId, db) {
  const merchant = db.merchants.find((m) => m.id === merchantId);
  const txns = db.transactions.filter((t) => t.merchantId === merchantId);
  let earned = 0;
  // Only direct merchants earn merchant commission
  if (merchant && !merchant.affiliateId) {
    txns.forEach((t) => {
      const company = db.companies.find((c) => c.id === t.companyId);
      const r = calcTransaction(t, { company, merchant, affiliate: null });
      earned += r.beneficiaryCommission;
    });
  }
  const paid = (db.merchantPayments || [])
    .filter((p) => p.merchantId === merchantId)
    .reduce((sum, p) => sum + num(p.amount), 0);
  return { earned, paid, balance: earned - paid };
}
