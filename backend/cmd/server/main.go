// Command server is the entrypoint for the PGCS multi-tenant backend API.
//
// Startup sequence (design "Startup sequence"): load and validate configuration,
// connect to PostgreSQL, then apply schema migrations — all before the HTTP
// server accepts any requests (Req 1.4). A missing required configuration value
// (Req 2.4), an unreachable database (Req 1.5), or a failed migration (Req 1.6)
// each cause the process to log a descriptive error and exit without serving.
// Secret values are never logged in plaintext (Req 2.5).
package main

import (
	"log"
	"net/http"
	"os"

	"pgcs/backend/internal/api"
	"pgcs/backend/internal/config"
	"pgcs/backend/internal/service/auth"
	"pgcs/backend/internal/service/migration"
)

func main() {
	// 1. Load and validate configuration. On a missing required value, log which
	//    value is missing (the error names it) and exit without serving (Req 2.4).
	cfg, err := config.Load()
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	// 2. Connect to PostgreSQL. On failure, log a descriptive error using the
	//    redacted configuration (secrets masked) and exit without serving
	//    (Req 1.5, 2.5).
	db, err := migration.Connect(cfg)
	if err != nil {
		log.Printf("database connection failed: %v; config: %s", err, cfg.Redacted())
		os.Exit(1)
	}

	// 3. Apply schema migrations for all models before serving. On failure, log
	//    and exit without serving (Req 1.4, 1.6, 5.1).
	if err := migration.AutoMigrate(db); err != nil {
		log.Printf("schema migration failed: %v", err)
		os.Exit(1)
	}

	log.Printf("pgcs backend: configuration loaded, database connected, schema migrated; config: %s", cfg.Redacted())

	// 4. Bootstrap the root SuperAdmin account on first startup. SeedIfEmpty is
	//    idempotent: it creates the SuperAdmin (credentials from config) and its
	//    tenant on an empty database, and leaves an already-bootstrapped database
	//    unchanged. Every other account is created at runtime (Admins via
	//    leasing). Bootstrap completes before the server serves.
	if err := migration.SeedIfEmpty(db, cfg.SuperAdminUserID, cfg.SuperAdminPassword); err != nil {
		log.Printf("superadmin bootstrap failed: %v", err)
		os.Exit(1)
	}

	// 4. Construct the Auth_Service and build the HTTP router (Recover + Error
	//    behavior globally; Auth + TenantScope on protected route groups), then
	//    start serving. The server only begins accepting requests now that
	//    configuration, the database connection, and migrations have all
	//    succeeded (Req 1.4).
	authSvc := auth.New(db, cfg.TokenSecret, cfg.TokenTTL)
	router := api.NewRouter(api.Deps{Auth: authSvc, DB: db})

	addr := ":" + cfg.HTTPPort
	log.Printf("pgcs backend: startup complete; listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		// ListenAndServe only returns on a fatal error (e.g. the port is
		// already in use or the listener is closed); log and exit non-zero.
		log.Printf("http server error: %v", err)
		os.Exit(1)
	}
}
