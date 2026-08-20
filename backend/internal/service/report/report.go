// This file implements task 10.2: the six report types with inclusive date
// filtering. It extends the report Service (see ledger.go) with Generate, which
// produces a column/row ReportResult for each report type the frontend
// Reports.jsx screen renders: company-, merchant-, affiliate-, gateway-,
// settlement-, and outstanding-wise (Req 11.1).
//
// Every report is built from records read through the principal's tenant +
// owner scope (repo.ScopeTenant via Service.scoped), so a report only ever
// aggregates the requester's own data (Req 11.6). Computed figures reuse the
// Commission_Engine (commission.Calc) and the same ledger arithmetic as
// ledger.go.
//
// Date filtering (Req 11.2): when StartDate and/or EndDate are supplied, only
// records whose date falls within the inclusive [start, end] range are
// included. Transaction, settlement, and payment dates are ISO yyyy-mm-dd
// strings, so the bound is compared as a yyyy-mm-dd string (lexicographic order
// on that format matches chronological order). Unlike the original frontend —
// which date-filtered transactions but not the ledger/settlement figures — the
// filter here is applied uniformly to every record a report aggregates, so the
// whole report reflects the selected range.
//
// ReportResult is intentionally a plain column/row model carrying raw values
// (numbers stay numbers): task 10.3 renders the same Columns/Rows to CSV and
// PDF without re-deriving anything, and the ?format= handling lives there.
package report

import (
	"sort"
	"time"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/commission"
)

// ReportType identifies one of the six reports the Report_Service produces
// (Req 11.1). The values match the {type} path segment of
// GET /api/reports/{type}.
type ReportType string

// The six supported report types (Req 11.1).
const (
	ReportCompany     ReportType = "company"
	ReportMerchant    ReportType = "merchant"
	ReportAffiliate   ReportType = "affiliate"
	ReportGateway     ReportType = "gateway"
	ReportSettlement  ReportType = "settlement"
	ReportOutstanding ReportType = "outstanding"
)

// dateLayout is the ISO yyyy-mm-dd layout used by every record date and by the
// report start/end bounds.
const dateLayout = "2006-01-02"

// Column describes one report column. Key names the field in each row map;
// Label is the human-readable header; Num marks numeric columns so a renderer
// (table, CSV, PDF) can right-align/format them.
type Column struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Num   bool   `json:"num,omitempty"`
}

// ReportRequest is the resolved request for a single report. StartDate and
// EndDate are optional inclusive bounds (nil means unbounded on that side).
type ReportRequest struct {
	Type      ReportType
	StartDate *time.Time
	EndDate   *time.Time
}

// ReportResult is the column/row model returned for every report. Title is a
// display/export caption; Columns defines the ordered headers; each Rows entry
// maps a Column.Key to its value (numbers stay numeric so exporters and the UI
// format consistently).
type ReportResult struct {
	Title   string           `json:"title"`
	Columns []Column         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// Generate produces the requested report within the principal's tenant + owner
// scope (Req 11.1, 11.6) with inclusive date filtering applied to the records
// each report aggregates (Req 11.2). An unrecognized report type is a client
// error and yields a 400 validation error identifying the offending value.
func (s *Service) Generate(p service.Principal, req ReportRequest) (ReportResult, error) {
	switch req.Type {
	case ReportCompany:
		return s.companyReport(p, req)
	case ReportMerchant:
		return s.merchantReport(p, req)
	case ReportAffiliate:
		return s.affiliateReport(p, req)
	case ReportGateway:
		return s.gatewayReport(p, req)
	case ReportSettlement:
		return s.settlementReport(p, req)
	case ReportOutstanding:
		return s.outstandingReport(p, req)
	default:
		return ReportResult{}, apierr.ValidationField("type",
			"unknown report type; expected one of company, merchant, affiliate, gateway, settlement, outstanding")
	}
}

// inRange reports whether an ISO yyyy-mm-dd date string falls within the
// inclusive [start, end] bounds (Req 11.2). A nil bound is open on that side.
// The comparison is lexicographic on the yyyy-mm-dd form, which orders
// chronologically.
func inRange(date string, start, end *time.Time) bool {
	if start != nil && date < start.Format(dateLayout) {
		return false
	}
	if end != nil && date > end.Format(dateLayout) {
		return false
	}
	return true
}

// reportData holds the scoped records and lookup maps a report needs. Loading
// it once per request keeps each report builder simple and avoids repeated
// queries.
type reportData struct {
	companies  map[string]*model.Company
	merchants  map[string]*model.Merchant
	affiliates map[string]*model.Affiliate
	gateways   map[string]*model.Gateway
	txns       []model.Transaction
}

// load reads the companies (with gateway assignments), merchants, affiliates,
// gateways, and transactions within the principal's scope, indexing the lookup
// entities by id for the per-transaction Commission_Engine context.
func (s *Service) load(p service.Principal) (*reportData, error) {
	var companies []model.Company
	if err := s.scoped(p).Preload("Gateways").Find(&companies).Error; err != nil {
		return nil, err
	}
	var merchants []model.Merchant
	if err := s.scoped(p).Find(&merchants).Error; err != nil {
		return nil, err
	}
	var affiliates []model.Affiliate
	if err := s.scoped(p).Find(&affiliates).Error; err != nil {
		return nil, err
	}
	var gateways []model.Gateway
	if err := s.scoped(p).Find(&gateways).Error; err != nil {
		return nil, err
	}
	var txns []model.Transaction
	if err := s.scoped(p).Find(&txns).Error; err != nil {
		return nil, err
	}

	d := &reportData{
		companies:  make(map[string]*model.Company, len(companies)),
		merchants:  make(map[string]*model.Merchant, len(merchants)),
		affiliates: make(map[string]*model.Affiliate, len(affiliates)),
		gateways:   make(map[string]*model.Gateway, len(gateways)),
		txns:       txns,
	}
	for i := range companies {
		d.companies[companies[i].ID] = &companies[i]
	}
	for i := range merchants {
		d.merchants[merchants[i].ID] = &merchants[i]
	}
	for i := range affiliates {
		d.affiliates[affiliates[i].ID] = &affiliates[i]
	}
	for i := range gateways {
		d.gateways[gateways[i].ID] = &gateways[i]
	}
	return d, nil
}

// affiliateFor returns the affiliate assigned to the merchant, or nil for a
// direct merchant or an unknown affiliate.
func (d *reportData) affiliateFor(m *model.Merchant) *model.Affiliate {
	if m == nil || m.AffiliateID == nil || *m.AffiliateID == "" {
		return nil
	}
	return d.affiliates[*m.AffiliateID]
}

// calc computes the commission breakdown for a transaction using its related
// company/merchant/affiliate from the loaded lookup maps.
func (d *reportData) calc(t model.Transaction) commission.Breakdown {
	m := d.merchants[t.MerchantID]
	return commission.Calc(t, commission.TxnContext{
		Company:   d.companies[t.CompanyID],
		Merchant:  m,
		Affiliate: d.affiliateFor(m),
	})
}

// nameOr returns the entity name or the "—" placeholder used by the frontend
// when an entity is missing.
func nameOr(name string) string {
	if name == "" {
		return "—"
	}
	return name
}

// companyReport sums, per company, the transaction count, transaction and
// settlement amounts, admin commission, and company net income over the
// in-range transactions, plus paid (in-range settlements) and outstanding
// (net − paid) (Req 11.1; mirrors Reports.jsx "Company Wise").
func (s *Service) companyReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	d, err := s.load(p)
	if err != nil {
		return ReportResult{}, err
	}

	type agg struct {
		name      string
		count     int
		txnAmount float64
		settle    float64
		adminComm float64
		net       float64
	}
	byCompany := map[string]*agg{}
	order := []string{}
	for _, t := range d.txns {
		if !inRange(t.Date, req.StartDate, req.EndDate) {
			continue
		}
		a := byCompany[t.CompanyID]
		if a == nil {
			name := ""
			if c := d.companies[t.CompanyID]; c != nil {
				name = c.Name
			}
			a = &agg{name: nameOr(name)}
			byCompany[t.CompanyID] = a
			order = append(order, t.CompanyID)
		}
		bd := d.calc(t)
		a.count++
		a.txnAmount += t.TxnAmount
		a.settle += t.SettlementAmount
		a.adminComm += bd.AdminCommission
		a.net += bd.CompanyNetIncome
	}

	// Paid per company from in-range settlements.
	paidByCompany, err := s.settlementTotals(p, req)
	if err != nil {
		return ReportResult{}, err
	}

	sort.Slice(order, func(i, j int) bool { return byCompany[order[i]].name < byCompany[order[j]].name })
	rows := make([]map[string]any, 0, len(order))
	for _, id := range order {
		a := byCompany[id]
		paid := paidByCompany[id]
		rows = append(rows, map[string]any{
			"company":     a.name,
			"count":       a.count,
			"txnAmount":   a.txnAmount,
			"settlement":  a.settle,
			"adminComm":   a.adminComm,
			"net":         a.net,
			"paid":        paid,
			"outstanding": a.net - paid,
		})
	}
	return ReportResult{
		Title: "Company Wise Report",
		Columns: []Column{
			{Key: "company", Label: "Company"},
			{Key: "count", Label: "Txns", Num: true},
			{Key: "txnAmount", Label: "Txn Amount", Num: true},
			{Key: "settlement", Label: "Settlement", Num: true},
			{Key: "adminComm", Label: "Admin Comm.", Num: true},
			{Key: "net", Label: "Company Net", Num: true},
			{Key: "paid", Label: "Paid", Num: true},
			{Key: "outstanding", Label: "Outstanding", Num: true},
		},
		Rows: rows,
	}, nil
}

// merchantReport sums, per merchant, the transaction count, transaction and
// settlement amounts, and commission over the in-range transactions.
// Every merchant accrues its own commission; for an affiliate-assigned merchant
// that commission is carved out of the affiliate's cut (Req 11.1; mirrors
// Reports.jsx "Merchant Wise").
func (s *Service) merchantReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	d, err := s.load(p)
	if err != nil {
		return ReportResult{}, err
	}

	type agg struct {
		name      string
		count     int
		txnAmount float64
		settle    float64
		comm      float64
	}
	byMerchant := map[string]*agg{}
	order := []string{}
	for _, t := range d.txns {
		if !inRange(t.Date, req.StartDate, req.EndDate) {
			continue
		}
		m := d.merchants[t.MerchantID]
		a := byMerchant[t.MerchantID]
		if a == nil {
			name := ""
			if m != nil {
				name = m.Name
			}
			a = &agg{name: nameOr(name)}
			byMerchant[t.MerchantID] = a
			order = append(order, t.MerchantID)
		}
		a.count++
		a.txnAmount += t.TxnAmount
		a.settle += t.SettlementAmount
		// Every merchant earns its own commission. For an affiliate-assigned
		// merchant the engine carves it out of the affiliate's cut, so
		// MerchantCommission is correct for both cases (Req 10.4 parity).
		a.comm += d.calc(t).MerchantCommission
	}

	sort.Slice(order, func(i, j int) bool { return byMerchant[order[i]].name < byMerchant[order[j]].name })
	rows := make([]map[string]any, 0, len(order))
	for _, id := range order {
		a := byMerchant[id]
		rows = append(rows, map[string]any{
			"merchant":   a.name,
			"count":      a.count,
			"txnAmount":  a.txnAmount,
			"settlement": a.settle,
			"comm":       a.comm,
		})
	}
	return ReportResult{
		Title: "Merchant Wise Report",
		Columns: []Column{
			{Key: "merchant", Label: "Merchant"},
			{Key: "count", Label: "Txns", Num: true},
			{Key: "txnAmount", Label: "Txn Amount", Num: true},
			{Key: "settlement", Label: "Settlement", Num: true},
			{Key: "comm", Label: "Commission", Num: true},
		},
		Rows: rows,
	}, nil
}

// affiliateReport lists every affiliate with the count and earned commission
// over its merchants' in-range transactions, paid (in-range affiliate
// payments), and balance (earned − paid) (Req 11.1; mirrors Reports.jsx
// "Affiliate Wise").
func (s *Service) affiliateReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	d, err := s.load(p)
	if err != nil {
		return ReportResult{}, err
	}

	type agg struct {
		count  int
		earned float64
	}
	byAffiliate := map[string]*agg{}
	for id := range d.affiliates {
		byAffiliate[id] = &agg{}
	}
	for _, t := range d.txns {
		if !inRange(t.Date, req.StartDate, req.EndDate) {
			continue
		}
		m := d.merchants[t.MerchantID]
		if m == nil || m.AffiliateID == nil || *m.AffiliateID == "" {
			continue
		}
		a := byAffiliate[*m.AffiliateID]
		if a == nil {
			continue // transaction's affiliate is out of scope
		}
		a.count++
		// The affiliate keeps its gross cut less the merchant share carved out
		// of it.
		a.earned += d.calc(t).AffiliateCommission
	}

	// Paid per affiliate from in-range affiliate payments.
	paidByAffiliate, err := s.affiliatePaymentTotals(p, req)
	if err != nil {
		return ReportResult{}, err
	}

	ids := make([]string, 0, len(d.affiliates))
	for id := range d.affiliates {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return d.affiliates[ids[i]].Name < d.affiliates[ids[j]].Name })
	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		a := byAffiliate[id]
		paid := paidByAffiliate[id]
		rows = append(rows, map[string]any{
			"affiliate": nameOr(d.affiliates[id].Name),
			"count":     a.count,
			"earned":    a.earned,
			"paid":      paid,
			"balance":   a.earned - paid,
		})
	}
	return ReportResult{
		Title: "Affiliate Wise Report",
		Columns: []Column{
			{Key: "affiliate", Label: "Affiliate"},
			{Key: "count", Label: "Txns", Num: true},
			{Key: "earned", Label: "Commission", Num: true},
			{Key: "paid", Label: "Paid", Num: true},
			{Key: "balance", Label: "Balance", Num: true},
		},
		Rows: rows,
	}, nil
}

// gatewayReport sums, per gateway, the transaction count, transaction amount,
// and admin commission over the in-range transactions (Req 11.1; mirrors
// Reports.jsx "Gateway Wise").
func (s *Service) gatewayReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	d, err := s.load(p)
	if err != nil {
		return ReportResult{}, err
	}

	type agg struct {
		name      string
		count     int
		txnAmount float64
		adminComm float64
	}
	byGateway := map[string]*agg{}
	order := []string{}
	for _, t := range d.txns {
		if !inRange(t.Date, req.StartDate, req.EndDate) {
			continue
		}
		a := byGateway[t.GatewayID]
		if a == nil {
			name := ""
			if g := d.gateways[t.GatewayID]; g != nil {
				name = g.Name
			}
			a = &agg{name: nameOr(name)}
			byGateway[t.GatewayID] = a
			order = append(order, t.GatewayID)
		}
		a.count++
		a.txnAmount += t.TxnAmount
		a.adminComm += d.calc(t).AdminCommission
	}

	sort.Slice(order, func(i, j int) bool { return byGateway[order[i]].name < byGateway[order[j]].name })
	rows := make([]map[string]any, 0, len(order))
	for _, id := range order {
		a := byGateway[id]
		rows = append(rows, map[string]any{
			"gateway":   a.name,
			"count":     a.count,
			"txnAmount": a.txnAmount,
			"adminComm": a.adminComm,
		})
	}
	return ReportResult{
		Title: "Gateway Wise Report",
		Columns: []Column{
			{Key: "gateway", Label: "Gateway"},
			{Key: "count", Label: "Txns", Num: true},
			{Key: "txnAmount", Label: "Txn Amount", Num: true},
			{Key: "adminComm", Label: "Admin Commission", Num: true},
		},
		Rows: rows,
	}, nil
}

// settlementReport lists each in-range settlement with its date, company,
// amount, mode, and reference (Req 11.1; mirrors Reports.jsx "Settlement").
func (s *Service) settlementReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	companyNames, err := s.companyNames(p)
	if err != nil {
		return ReportResult{}, err
	}
	bankNames, err := s.bankNames(p)
	if err != nil {
		return ReportResult{}, err
	}
	var settlements []model.Settlement
	if err := s.scoped(p).Find(&settlements).Error; err != nil {
		return ReportResult{}, err
	}

	// Stable order: by date then company name.
	sort.SliceStable(settlements, func(i, j int) bool {
		if settlements[i].Date != settlements[j].Date {
			return settlements[i].Date < settlements[j].Date
		}
		return companyNames[settlements[i].CompanyID] < companyNames[settlements[j].CompanyID]
	})

	rows := make([]map[string]any, 0, len(settlements))
	for _, st := range settlements {
		if !inRange(st.Date, req.StartDate, req.EndDate) {
			continue
		}
		rows = append(rows, map[string]any{
			"date":    st.Date,
			"company": nameOr(companyNames[st.CompanyID]),
			"amount":  st.Amount,
			"type":    nameOr(st.PaymentType),
			"mode":    st.PaymentMode,
			"bank":    nameOr(bankNames[st.BankID]),
			"ref":     nameOr(st.RefNumber),
		})
	}
	return ReportResult{
		Title: "Settlement Report",
		Columns: []Column{
			{Key: "date", Label: "Date"},
			{Key: "company", Label: "Company"},
			{Key: "amount", Label: "Amount", Num: true},
			{Key: "type", Label: "Cash/Bank"},
			{Key: "mode", Label: "Mode"},
			{Key: "bank", Label: "Bank"},
			{Key: "ref", Label: "Reference"},
		},
		Rows: rows,
	}, nil
}

// outstandingReport lists every company with receivable (Σ company net income
// over in-range transactions), paid (in-range settlements), outstanding
// (receivable − paid), and a settlement status (Req 11.1; mirrors Reports.jsx
// "Outstanding").
func (s *Service) outstandingReport(p service.Principal, req ReportRequest) (ReportResult, error) {
	d, err := s.load(p)
	if err != nil {
		return ReportResult{}, err
	}

	receivable := map[string]float64{}
	for _, t := range d.txns {
		if !inRange(t.Date, req.StartDate, req.EndDate) {
			continue
		}
		receivable[t.CompanyID] += d.calc(t).CompanyNetIncome
	}
	paidByCompany, err := s.settlementTotals(p, req)
	if err != nil {
		return ReportResult{}, err
	}

	ids := make([]string, 0, len(d.companies))
	for id := range d.companies {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return d.companies[ids[i]].Name < d.companies[ids[j]].Name })
	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		name := nameOr(d.companies[id].Name)
		recv := receivable[id]
		paid := paidByCompany[id]
		bal := recv - paid
		status := "Settled"
		switch {
		case bal > 0.0001:
			status = name + " should receive"
		case bal < -0.0001:
			status = name + " owes Admin"
		}
		rows = append(rows, map[string]any{
			"company":     name,
			"receivable":  recv,
			"paid":        paid,
			"outstanding": bal,
			"status":      status,
		})
	}
	return ReportResult{
		Title: "Outstanding Report",
		Columns: []Column{
			{Key: "company", Label: "Company"},
			{Key: "receivable", Label: "Receivable", Num: true},
			{Key: "paid", Label: "Paid", Num: true},
			{Key: "outstanding", Label: "Outstanding", Num: true},
			{Key: "status", Label: "Status"},
		},
		Rows: rows,
	}, nil
}

// settlementTotals returns the sum of in-range settlement amounts per company
// within the principal's scope.
func (s *Service) settlementTotals(p service.Principal, req ReportRequest) (map[string]float64, error) {
	var settlements []model.Settlement
	if err := s.scoped(p).Find(&settlements).Error; err != nil {
		return nil, err
	}
	totals := map[string]float64{}
	for _, st := range settlements {
		if !inRange(st.Date, req.StartDate, req.EndDate) {
			continue
		}
		totals[st.CompanyID] += st.Amount
	}
	return totals, nil
}

// affiliatePaymentTotals returns the sum of in-range affiliate-payment amounts
// per affiliate within the principal's scope.
func (s *Service) affiliatePaymentTotals(p service.Principal, req ReportRequest) (map[string]float64, error) {
	var payments []model.AffiliatePayment
	if err := s.scoped(p).Find(&payments).Error; err != nil {
		return nil, err
	}
	totals := map[string]float64{}
	for _, pay := range payments {
		if !inRange(pay.Date, req.StartDate, req.EndDate) {
			continue
		}
		totals[pay.AffiliateID] += pay.Amount
	}
	return totals, nil
}

// companyNames returns a company-id -> name map within the principal's scope.
func (s *Service) companyNames(p service.Principal) (map[string]string, error) {
	var companies []model.Company
	if err := s.scoped(p).Find(&companies).Error; err != nil {
		return nil, err
	}
	names := make(map[string]string, len(companies))
	for i := range companies {
		names[companies[i].ID] = companies[i].Name
	}
	return names, nil
}

// bankNames returns the id -> display-name map of the banks within the
// principal's scope, used to label the Bank a payment was routed through.
func (s *Service) bankNames(p service.Principal) (map[string]string, error) {
	var banks []model.Bank
	if err := s.scoped(p).Find(&banks).Error; err != nil {
		return nil, err
	}
	names := make(map[string]string, len(banks))
	for i := range banks {
		names[banks[i].ID] = banks[i].Name
	}
	return names, nil
}
