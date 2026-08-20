// Package migration implements the Migration_Service: establishing the GORM
// PostgreSQL connection from configuration and applying schema migrations
// (tables, columns, indexes, and foreign keys) for all model structs.
//
// Connecting and migrating run at startup before the API_Server accepts
// requests (Req 1.4). A failure to connect (Req 1.5) or to migrate (Req 1.6)
// is surfaced to the caller so startup can log a descriptive error and exit.
// Secret configuration values (the database password) are never logged: the
// data source name produced here is used only to open the connection and is
// never returned or logged in plaintext (Req 2.5).
package migration

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pgcs/backend/internal/config"
	"pgcs/backend/internal/model"
)

// AllModels returns every GORM model struct that the schema is composed of, in
// an order suitable for AutoMigrate. GORM resolves foreign-key relationships
// across the full set, so all tables, columns, indexes, and foreign keys are
// created (Req 5.1, 1.3).
func AllModels() []any {
	return []any{
		&model.Tenant{},
		&model.Account{},
		&model.Gateway{},
		&model.Bank{},
		&model.Company{},
		&model.CompanyGateway{},
		&model.Affiliate{},
		&model.Merchant{},
		&model.MerchantBank{},
		&model.AtmCard{},
		&model.MerchantGatewayCredential{},
		&model.CustomField{},
		&model.Transaction{},
		&model.Settlement{},
		&model.AffiliatePayment{},
		&model.MerchantPayment{},
		&model.Lease{},
		&model.RevokedToken{},
	}
}

// dsn builds the PostgreSQL data source name from configuration. It embeds the
// database password and therefore MUST NOT be logged; callers use it only to
// open a connection (Req 2.5).
func dsn(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
}

// Connect opens a GORM PostgreSQL connection using the settings in cfg and
// verifies it is reachable. On failure it returns an error that describes the
// problem without exposing the database password (Req 1.5, 2.5).
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn(cfg)), &gorm.Config{})
	if err != nil {
		// err originates from the driver and may reference host/port/user/dbname
		// but never the password, which is not echoed back by the pgx driver.
		return nil, fmt.Errorf("connect to postgres at %s:%s/%s as %q: %w",
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, err)
	}

	// Verify the connection is actually usable before proceeding to migrations.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obtain underlying database handle: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres at %s:%s/%s as %q: %w",
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, err)
	}

	return db, nil
}

// AutoMigrate applies GORM auto-migration for the given models, creating or
// evolving all tables, columns, indexes, and foreign keys (Req 1.4, 5.1). When
// no models are supplied it migrates the full model set from AllModels.
func AutoMigrate(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		models = AllModels()
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate schema: %w", err)
	}
	return nil
}
