package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// dashboardHandlers serves the GET /api/dashboard endpoint. The dashboard is a
// read-only, tenant-scoped summary: it counts the business entities visible to
// the authenticated principal and returns them as labelled metric cards plus a
// small recent-activity feed derived from the most recent transactions.
type dashboardHandlers struct {
	deps Deps
}

// newDashboardHandlers builds the dashboard handlers over the router deps.
func newDashboardHandlers(deps Deps) *dashboardHandlers {
	return &dashboardHandlers{deps: deps}
}

// metric is a single dashboard metric card.
type metric struct {
	Label string `json:"label"`
	Value int    `json:"value"`
	Icon  string `json:"icon,omitempty"`
}

// dashboardResponse mirrors the mobile app's DashboardMetrics shape: a list of
// metric cards, a recent-activity feed, and a free-form summary map.
type dashboardResponse struct {
	Role           string           `json:"role"`
	Metrics        []metric         `json:"metrics"`
	RecentActivity []activityItem   `json:"recentActivity"`
	Summary        map[string]any   `json:"summary"`
}

// activityItem is one entry in the recent-activity feed.
type activityItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
}

// generate handles GET /api/dashboard. It aggregates counts within the
// principal's tenant/owner scope (so every role sees only the data it is
// permitted to) and returns role-appropriate metric cards. All counting reuses
// the generic tenant-scoped CRUD services, so scope enforcement is identical to
// the per-entity list endpoints (Req 8.5).
func (h *dashboardHandlers) generate(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}

	db := h.deps.DB

	companies, _ := crud.NewService[model.Company](db, nil).List(p)
	merchants, _ := crud.NewService[model.Merchant](db, nil).List(p)
	affiliates, _ := crud.NewService[model.Affiliate](db, nil).List(p)
	gateways, _ := crud.NewService[model.Gateway](db, nil).List(p)
	transactions, _ := crud.NewService[model.Transaction](db, nil).List(p)
	settlements, _ := crud.NewService[model.Settlement](db, nil).List(p)

	// Build metric cards. The set is intentionally broad; the app renders
	// whatever cards are returned, and each role only ever receives counts for
	// data within its own scope.
	metrics := []metric{
		{Label: "Companies", Value: len(companies), Icon: "domain"},
		{Label: "Merchants", Value: len(merchants), Icon: "store"},
		{Label: "Affiliates", Value: len(affiliates), Icon: "account-group"},
		{Label: "Gateways", Value: len(gateways), Icon: "credit-card"},
		{Label: "Transactions", Value: len(transactions), Icon: "swap-horizontal"},
		{Label: "Settlements", Value: len(settlements), Icon: "bank-transfer"},
	}

	// Recent activity: up to the five most recent transactions by slice order.
	recent := make([]activityItem, 0, 5)
	for i := len(transactions) - 1; i >= 0 && len(recent) < 5; i-- {
		t := transactions[i]
		recent = append(recent, activityItem{
			ID:          t.ID,
			Description: "Transaction " + t.ID,
			Date:        t.Date,
			Amount:      t.TxnAmount,
		})
	}

	resp := dashboardResponse{
		Role:           string(p.Role),
		Metrics:        metrics,
		RecentActivity: recent,
		Summary: map[string]any{
			"totalTransactions": len(transactions),
			"totalSettlements":  len(settlements),
		},
	}

	middleware.WriteJSON(w, http.StatusOK, resp)
	return nil
}
