// Package deps anchors the project's intended third-party dependencies during
// scaffolding (task 1.1) so they remain declared as direct requirements in
// go.mod until the real implementation imports them in later tasks.
//
// The standard-library CSV writer (encoding/csv) and PDF generation
// (github.com/go-pdf/fpdf) back report exports; GORM and the Postgres driver
// back persistence; golang-jwt and bcrypt back authentication; rapid backs the
// property-based tests.
package deps

import (
	// CSV export uses the standard library.
	_ "encoding/csv"

	_ "github.com/go-pdf/fpdf"
	_ "github.com/golang-jwt/jwt/v5"
	_ "golang.org/x/crypto/bcrypt"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
	_ "pgregory.net/rapid"
)
