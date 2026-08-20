// Package config loads and validates runtime configuration (database settings,
// HTTP port, token secret/lifetime) from the environment.
//
// Configuration is sourced from environment variables with documented defaults
// for non-secret values. Secret values (database password and token signing
// secret) have no default and are required; Load reports the first missing
// required value. Redacted produces a representation safe for logging that
// masks secret values.
package config

import (
	"fmt"
	"os"
	"time"
)

// Default values for non-secret settings. These mirror the documented defaults
// in the design configuration table.
const (
	defaultDBHost           = "localhost"
	defaultDBPort           = "5432"
	defaultDBName           = "pgcs"
	defaultDBUser           = "postgres"
	defaultHTTPPort         = "8080"
	defaultTokenTTL         = "24h"
	defaultSuperAdminUserID = "superadmin"

	redactedValue = "***"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string // secret
	HTTPPort   string

	TokenSecret string // secret
	TokenTTL    time.Duration

	// Bootstrap SuperAdmin credentials. The SuperAdmin is the only account
	// created at startup; every other account (Admins via leasing, and the
	// business entities they manage) is created at runtime. The password is a
	// required secret with no default so a real credential must be supplied.
	SuperAdminUserID   string
	SuperAdminPassword string // secret
}

// Load reads configuration from environment variables, applying documented
// defaults for non-secret values.
//
// The required secret values DB_PASSWORD and TOKEN_SECRET have no default. If a
// required value is missing, Load returns an error naming the first missing
// value, checked in the order DB_PASSWORD then TOKEN_SECRET. TOKEN_TTL is parsed
// as a Go duration (e.g. "24h", "30m"); an unparseable value yields an error.
func Load() (*Config, error) {
	c := &Config{
		DBHost:             getEnvDefault("DB_HOST", defaultDBHost),
		DBPort:             getEnvDefault("DB_PORT", defaultDBPort),
		DBName:             getEnvDefault("DB_NAME", defaultDBName),
		DBUser:             getEnvDefault("DB_USER", defaultDBUser),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		HTTPPort:           getEnvDefault("PORT", defaultHTTPPort),
		TokenSecret:        os.Getenv("TOKEN_SECRET"),
		SuperAdminUserID:   getEnvDefault("SUPERADMIN_USER_ID", defaultSuperAdminUserID),
		SuperAdminPassword: os.Getenv("SUPERADMIN_PASSWORD"),
	}

	// Required secrets, checked in documented order: DB_PASSWORD, TOKEN_SECRET,
	// then SUPERADMIN_PASSWORD.
	if c.DBPassword == "" {
		return nil, fmt.Errorf("missing required configuration value: DB_PASSWORD")
	}
	if c.TokenSecret == "" {
		return nil, fmt.Errorf("missing required configuration value: TOKEN_SECRET")
	}
	if c.SuperAdminPassword == "" {
		return nil, fmt.Errorf("missing required configuration value: SUPERADMIN_PASSWORD")
	}

	ttlRaw := getEnvDefault("TOKEN_TTL", defaultTokenTTL)
	ttl, err := time.ParseDuration(ttlRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid TOKEN_TTL %q: %w", ttlRaw, err)
	}
	c.TokenTTL = ttl

	return c, nil
}

// Redacted returns a representation of the configuration that is safe to log.
// Secret values (DBPassword and TokenSecret) are masked as "***" and never
// appear in plaintext.
func (c *Config) Redacted() string {
	return fmt.Sprintf(
		"Config{DBHost:%q, DBPort:%q, DBName:%q, DBUser:%q, DBPassword:%q, HTTPPort:%q, TokenSecret:%q, TokenTTL:%q, SuperAdminUserID:%q, SuperAdminPassword:%q}",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, redactedValue, c.HTTPPort, redactedValue, c.TokenTTL, c.SuperAdminUserID, redactedValue,
	)
}

// getEnvDefault returns the value of the environment variable named by key, or
// fallback when the variable is unset or empty.
func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
