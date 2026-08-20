package report

import (
	"bytes"
	"strings"
	"testing"
)

// sampleResult is a small ReportResult exercising string, integer, and
// floating-point cells plus the em-dash placeholder used by the reports.
func sampleResult() ReportResult {
	return ReportResult{
		Title: "Company Wise Report",
		Columns: []Column{
			{Key: "company", Label: "Company"},
			{Key: "count", Label: "Txns", Num: true},
			{Key: "net", Label: "Company Net", Num: true},
		},
		Rows: []map[string]any{
			{"company": "Acme, Inc.", "count": 3, "net": 1234.5},
			{"company": "—", "count": 0, "net": 0.0},
		},
	}
}

func TestExportCSV(t *testing.T) {
	data, err := ExportCSV(sampleResult())
	if err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ExportCSV returned no bytes")
	}

	// UTF-8 BOM prefix (matches the frontend export.js behavior).
	if !bytes.HasPrefix(data, utf8BOM) {
		t.Error("CSV is missing the UTF-8 BOM prefix")
	}

	body := string(bytes.TrimPrefix(data, utf8BOM))
	lines := strings.Split(strings.ReplaceAll(strings.TrimRight(body, "\r\n"), "\r\n", "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("CSV line count = %d, want 3 (header + 2 rows); body=%q", len(lines), body)
	}

	// Header row comes from the column labels.
	if got, want := lines[0], "Company,Txns,Company Net"; got != want {
		t.Errorf("header = %q, want %q", got, want)
	}

	// First data row: the comma in the company name must be quoted, the count
	// stays an integer, the amount keeps two decimals.
	if got, want := lines[1], `"Acme, Inc.",3,1234.50`; got != want {
		t.Errorf("row 1 = %q, want %q", got, want)
	}

	// Whole-number float renders without trailing decimals.
	if got, want := lines[2], "—,0,0"; got != want {
		t.Errorf("row 2 = %q, want %q", got, want)
	}
}

func TestExportPDF(t *testing.T) {
	data, err := ExportPDF(sampleResult())
	if err != nil {
		t.Fatalf("ExportPDF error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("ExportPDF returned no bytes")
	}
	// A well-formed PDF starts with the %PDF- signature.
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("PDF is missing the %%PDF- signature; first bytes = %q", data[:min(8, len(data))])
	}
}

// TestExportsAreIndependent confirms CSV and PDF are produced by independent
// calls that share no state: producing one does not depend on or alter the
// other, which is how the handler returns each format that succeeds even if
// another were to fail (Req 11.5).
func TestExportsAreIndependent(t *testing.T) {
	r := sampleResult()

	csv1, err := ExportCSV(r)
	if err != nil {
		t.Fatalf("ExportCSV error: %v", err)
	}
	if _, err := ExportPDF(r); err != nil {
		t.Fatalf("ExportPDF error: %v", err)
	}
	csv2, err := ExportCSV(r)
	if err != nil {
		t.Fatalf("ExportCSV (second) error: %v", err)
	}
	// CSV output is unaffected by an interleaved PDF export.
	if !bytes.Equal(csv1, csv2) {
		t.Error("CSV output changed after a PDF export; functions are not independent")
	}
}

func TestExportEmptyReport(t *testing.T) {
	empty := ReportResult{Title: "Empty", Columns: nil, Rows: nil}

	csvData, err := ExportCSV(empty)
	if err != nil {
		t.Fatalf("ExportCSV(empty) error: %v", err)
	}
	if len(csvData) == 0 {
		t.Error("ExportCSV(empty) returned no bytes (expected at least the BOM)")
	}

	pdfData, err := ExportPDF(empty)
	if err != nil {
		t.Fatalf("ExportPDF(empty) error: %v", err)
	}
	if !bytes.HasPrefix(pdfData, []byte("%PDF-")) {
		t.Error("ExportPDF(empty) did not produce a valid PDF")
	}
}
