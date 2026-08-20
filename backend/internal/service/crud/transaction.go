package crud

import (
	"errors"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/repo"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/commission"
)

// TransactionWithBreakdown embeds a transaction together with its computed
// commission breakdown (Req 9). The embedded Transaction is anonymous, so its
// fields are promoted into the top-level JSON object while the breakdown is
// nested under "breakdown" — producing the {...transaction, breakdown: {...}}
// shape the design specifies for transaction reads.
type TransactionWithBreakdown struct {
	model.Transaction
	Breakdown commission.Breakdown `json:"breakdown"`
}

// TransactionService is the tenant-scoped CRUD service for transactions. Unlike
// the simple entities served by the generic Service, every read enriches the
// transaction with the commission breakdown computed by the Commission_Engine
// (Req 9). To do so it loads the related Company (with its gateway
// assignments), Merchant, and — when the merchant is assigned to one —
// Affiliate, all constrained to the principal's tenant so the supporting
// lookups can never cross a tenant boundary (Req 4.1, 8.5).
//
// Writes reuse the generic, tenant-scoped repository so tenant_id is stamped
// from the principal on create and out-of-scope updates/deletes return
// not-found, exactly as the other entities (Req 4.2, 4.3).
type TransactionService struct {
	db *gorm.DB
}

// NewTransactionService constructs the transaction CRUD service bound to a
// database handle.
func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

// repoFor builds the generic tenant-scoped repository used for writes.
func (s *TransactionService) repoFor(p service.Principal) *repo.Repository[model.Transaction, *model.Transaction] {
	return repo.New[model.Transaction, *model.Transaction](s.db, p)
}

// List returns every transaction within the principal's scope, each enriched
// with its computed commission breakdown (Req 8.5, 9).
func (s *TransactionService) List(p service.Principal) ([]TransactionWithBreakdown, error) {
	var txns []model.Transaction
	if err := s.db.Scopes(repo.ScopeTenant(p)).Find(&txns).Error; err != nil {
		return nil, err
	}
	out := make([]TransactionWithBreakdown, 0, len(txns))
	for i := range txns {
		bd, err := s.breakdown(p, txns[i])
		if err != nil {
			return nil, err
		}
		out = append(out, TransactionWithBreakdown{Transaction: txns[i], Breakdown: bd})
	}
	return out, nil
}

// Get returns the transaction with the given id within the principal's scope,
// enriched with its computed commission breakdown (Req 9). A transaction that
// does not exist within the scope is reported as apierr.ErrNotFound so
// cross-tenant reads are indistinguishable from "does not exist" (Req 4.3,
// 18.2).
func (s *TransactionService) Get(p service.Principal, id string) (*TransactionWithBreakdown, error) {
	var txn model.Transaction
	err := s.db.Scopes(repo.ScopeTenant(p)).Where("id = ?", id).First(&txn).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	bd, err := s.breakdown(p, txn)
	if err != nil {
		return nil, err
	}
	return &TransactionWithBreakdown{Transaction: txn, Breakdown: bd}, nil
}

// Create validates and inserts a transaction, assigning its tenant from the
// principal (Req 4.2, 8.4). A missing id is filled with a generated one.
func (s *TransactionService) Create(p service.Principal, t *model.Transaction) error {
	if err := ValidateTransaction(t); err != nil {
		return err
	}
	if t.ID == "" {
		t.ID = GenID()
	}
	return s.repoFor(p).Create(t)
}

// Update validates and persists changes to a transaction. The record is
// resolved through the tenant scope first, so updating a transaction outside
// the principal's scope returns apierr.ErrNotFound and changes nothing
// (Req 4.3).
func (s *TransactionService) Update(p service.Principal, t *model.Transaction) error {
	if err := ValidateTransaction(t); err != nil {
		return err
	}
	return s.repoFor(p).Update(t)
}

// Delete removes the transaction with the given id within the principal's
// scope, or returns apierr.ErrNotFound when no in-scope record matches
// (Req 4.3).
func (s *TransactionService) Delete(p service.Principal, id string) error {
	return s.repoFor(p).Delete(id)
}

// breakdown loads the related records a transaction needs and computes its
// commission breakdown (Req 9). Every related lookup is constrained to the
// principal's tenant so the supporting reads stay within the tenant boundary.
// Missing related records degrade gracefully: the engine defaults to a 0%
// gateway commission with no beneficiary, mirroring calc.js optional chaining.
func (s *TransactionService) breakdown(p service.Principal, txn model.Transaction) (commission.Breakdown, error) {
	var ctx commission.TxnContext

	if txn.CompanyID != "" {
		var c model.Company
		err := s.db.Preload("Gateways").
			Where("tenant_id = ? AND id = ?", p.TenantID, txn.CompanyID).
			First(&c).Error
		switch {
		case err == nil:
			ctx.Company = &c
		case errors.Is(err, gorm.ErrRecordNotFound):
			// leave Company nil; engine defaults apply
		default:
			return commission.Breakdown{}, err
		}
	}

	if txn.MerchantID != "" {
		var m model.Merchant
		err := s.db.Where("tenant_id = ? AND id = ?", p.TenantID, txn.MerchantID).
			First(&m).Error
		switch {
		case err == nil:
			ctx.Merchant = &m
			if m.AffiliateID != nil && *m.AffiliateID != "" {
				var af model.Affiliate
				aerr := s.db.Where("tenant_id = ? AND id = ?", p.TenantID, *m.AffiliateID).
					First(&af).Error
				switch {
				case aerr == nil:
					ctx.Affiliate = &af
				case errors.Is(aerr, gorm.ErrRecordNotFound):
					// leave Affiliate nil; merchant becomes the beneficiary
				default:
					return commission.Breakdown{}, aerr
				}
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			// leave Merchant nil; engine defaults apply
		default:
			return commission.Breakdown{}, err
		}
	}

	return commission.Calc(txn, ctx), nil
}
