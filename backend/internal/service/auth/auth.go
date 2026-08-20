// Package auth implements the Auth_Service: credential verification, stateless
// Session_Token issuance/validation, and server-side token revocation for
// logout.
//
// Passwords are verified against one-way bcrypt hashes (Req 6.3, 6.4). A
// successful login yields a JWT signed with HMAC-SHA256 using the configured
// TOKEN_SECRET. The token carries the principal's identity (account id, role,
// tenant, owner scope), a unique token id (jti), and an expiry derived from the
// configured TOKEN_TTL (Req 6.1). Validation is stateless for the common path;
// logout records a token's jti in the revoked_tokens table so it is rejected
// thereafter (Req 6.7).
//
// Login never reveals whether the user id or the password was wrong: both an
// unknown account and a bad password return the same generic invalid-credentials
// error (Req 6.2). Missing, malformed, expired, or revoked tokens are rejected
// as unauthenticated (Req 6.5, 6.6).
//
// Leased Admin accounts are additionally gated on their lease's effective
// status: a non-active lease (expired, suspended, or revoked) denies both login
// and token use with a 403 lease-inactive error (Req 14.2, 14.3, 15.4, 15.6).
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
)

// Stored lease intent and the derived "expired" status. These mirror the
// values written to model.Lease.Status (Active | Suspended | Revoked); Expired
// is never stored and is derived from now vs ExpiryDate.
//
// NOTE: this is a minimal local derivation of the lease state machine so the
// Auth_Service can gate access now. Task 12.1 introduces a shared
// Lease_Manager.EffectiveStatus with the same precedence
// (Revoked > Suspended > Expired > Active); this helper is expected to be
// refactored to share that implementation.
const (
	leaseActive    = "Active"
	leaseSuspended = "Suspended"
	leaseRevoked   = "Revoked"
	leaseExpired   = "Expired"
)

// AuthService is the behavior the HTTP layer and middleware depend on. *Service
// is the concrete implementation.
type AuthService interface {
	// Login verifies credentials and issues a Session_Token on success.
	Login(userID, password string) (token string, p service.Principal, err error)
	// Authenticate validates a Session_Token and returns its principal.
	Authenticate(token string) (service.Principal, error)
	// Logout invalidates the given Session_Token.
	Logout(token string) error
}

// Service implements AuthService over a GORM database and an HMAC signing
// secret.
type Service struct {
	db     *gorm.DB
	secret []byte
	ttl    time.Duration
	// now returns the current time; overridable in tests for deterministic
	// expiry and lease derivation.
	now func() time.Time
}

// Ensure *Service satisfies the interface.
var _ AuthService = (*Service)(nil)

// New constructs an Auth_Service bound to a database handle, the token signing
// secret, and the token lifetime (typically config.TokenSecret and
// config.TokenTTL).
func New(db *gorm.DB, tokenSecret string, tokenTTL time.Duration) *Service {
	return &Service{
		db:     db,
		secret: []byte(tokenSecret),
		ttl:    tokenTTL,
		now:    time.Now,
	}
}

// Claims is the JWT payload carried by a Session_Token. The registered claims
// supply the unique token id (jti) and expiry (exp); the custom claims carry
// the principal so the common validation path stays stateless.
type Claims struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
	OwnerType string `json:"owner_type,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	jwt.RegisteredClaims
}

// Login looks up the account by user id, verifies the password against the
// stored bcrypt hash, and on success issues a Session_Token (Req 6.1, 6.3).
//
// An unknown user id and an incorrect password both return
// apierr.ErrInvalidCredentials so the response never reveals which field was
// wrong (Req 6.2). For a leased Admin whose lease is not active, login is
// denied with a 403 lease-inactive error (Req 14.2, 15.4, 15.6).
func (s *Service) Login(userID, password string) (string, service.Principal, error) {
	var acct model.Account
	// Case-insensitive, deterministic lookup so user ids cannot collide by case
	// (e.g. "Admin" vs "admin") and login always resolves to the single account
	// that owns the id.
	err := s.db.Where("LOWER(user_id) = LOWER(?)", userID).First(&acct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Unknown account: indistinguishable from a wrong password (Req 6.2).
		return "", service.Principal{}, apierr.ErrInvalidCredentials
	}
	if err != nil {
		return "", service.Principal{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(acct.PasswordHash), []byte(password)) != nil {
		return "", service.Principal{}, apierr.ErrInvalidCredentials
	}

	// Gate leased Admins on their lease's effective status before issuing a
	// token (Req 14.2, 15.4, 15.6).
	if err := s.checkLease(acct.ID, acct.Role); err != nil {
		return "", service.Principal{}, err
	}

	p := principalFromAccount(acct)
	token, err := s.issueToken(p)
	if err != nil {
		return "", service.Principal{}, err
	}
	return token, p, nil
}

// Authenticate validates a Session_Token's signature and expiry, rejects tokens
// whose jti has been revoked, re-derives lease status for leased Admins, and
// returns the principal carried by the token.
//
// Malformed or expired tokens map to apierr.ErrUnauthenticated (401, Req 6.6); a
// non-active lease maps to apierr.ErrLeaseInactive (403, Req 14.3, 15.4, 15.6).
func (s *Service) Authenticate(token string) (service.Principal, error) {
	claims, err := s.parse(token)
	if err != nil {
		// Never disclose why the token failed (Req 6.6).
		return service.Principal{}, apierr.ErrUnauthenticated
	}

	// Reject tokens that have been logged out (Req 6.7).
	if claims.ID != "" {
		var rt model.RevokedToken
		err := s.db.Where("jti = ?", claims.ID).First(&rt).Error
		switch {
		case err == nil:
			return service.Principal{}, apierr.ErrUnauthenticated
		case errors.Is(err, gorm.ErrRecordNotFound):
			// not revoked: continue
		default:
			return service.Principal{}, err
		}
	}

	// Re-derive lease status on every request so an expired/suspended/revoked
	// lease denies access even for a still-valid token (Req 14.3, 15.4, 15.6).
	if err := s.checkLease(claims.AccountID, claims.Role); err != nil {
		return service.Principal{}, err
	}

	return service.Principal{
		AccountID: claims.AccountID,
		Role:      claims.Role,
		TenantID:  claims.TenantID,
		OwnerType: claims.OwnerType,
		OwnerID:   claims.OwnerID,
	}, nil
}

// Logout records the token's jti in the revoked_tokens table (keyed by jti,
// with the token's own expiry retained for later cleanup), invalidating the
// active Session_Token (Req 6.7). Re-logging-out the same token is a no-op.
func (s *Service) Logout(token string) error {
	claims, err := s.parse(token)
	if err != nil {
		return apierr.ErrUnauthenticated
	}
	if claims.ID == "" {
		return apierr.ErrUnauthenticated
	}

	var exp time.Time
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}

	rt := model.RevokedToken{Jti: claims.ID, ExpiresAt: exp}
	// Idempotent insert: a second logout of the same token must not error.
	if err := s.db.
		Where(model.RevokedToken{Jti: claims.ID}).
		Attrs(model.RevokedToken{ExpiresAt: exp}).
		FirstOrCreate(&rt).Error; err != nil {
		return err
	}
	return nil
}

// checkLease denies access when a leased Admin's lease is not active. Accounts
// that are not Admins, or Admins with no lease (e.g. the seeded demo Admin),
// pass through. The effective status precedence is
// Revoked > Suspended > Expired(now>expiry) > Active.
func (s *Service) checkLease(accountID, role string) error {
	if role != service.RoleAdmin {
		return nil
	}

	var lease model.Lease
	err := s.db.Where("account_id = ?", accountID).First(&lease).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil // Admin without a lease (its own tenant) is unrestricted.
	}
	if err != nil {
		return err
	}

	switch effectiveLeaseStatus(lease, s.now()) {
	case leaseActive:
		return nil
	case leaseExpired:
		return apierr.LeaseInactive("lease has expired")
	case leaseSuspended:
		return apierr.LeaseInactive("lease is suspended")
	case leaseRevoked:
		return apierr.LeaseInactive("lease has been revoked")
	default:
		return apierr.ErrLeaseInactive
	}
}

// effectiveLeaseStatus resolves the visible status of a lease at time now from
// its stored administrative intent. Revoked and Suspended take precedence over
// expiry; otherwise a lease is Expired once now is past its ExpiryDate.
func effectiveLeaseStatus(l model.Lease, now time.Time) string {
	switch l.Status {
	case leaseRevoked:
		return leaseRevoked
	case leaseSuspended:
		return leaseSuspended
	}
	if now.After(l.ExpiryDate) {
		return leaseExpired
	}
	return leaseActive
}

// issueToken builds and signs a Session_Token for the principal with a fresh
// unique jti and an expiry of now + ttl.
func (s *Service) issueToken(p service.Principal) (string, error) {
	jti, err := newJTI()
	if err != nil {
		return "", err
	}
	now := s.now()
	claims := Claims{
		AccountID: p.AccountID,
		Role:      p.Role,
		TenantID:  p.TenantID,
		OwnerType: p.OwnerType,
		OwnerID:   p.OwnerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

// parse validates the token's signature and standard claims (including expiry)
// and returns the typed claims. The signing method is pinned to HMAC to reject
// algorithm-substitution attempts.
func (s *Service) parse(token string) (*Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithTimeFunc(s.now))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

// principalFromAccount projects an Account into the Principal carried by a
// Session_Token and consumed by tenant-scope enforcement.
func principalFromAccount(a model.Account) service.Principal {
	return service.Principal{
		AccountID: a.ID,
		Role:      a.Role,
		TenantID:  a.TenantID,
		OwnerType: a.OwnerType,
		OwnerID:   a.OwnerID,
	}
}

// newJTI returns a random 128-bit token id encoded as hex, unique per issued
// token so logout can target an individual token.
func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
