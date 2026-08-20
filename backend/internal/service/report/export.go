// This file implements task 10.3: CSV and PDF exports for the Report_Service
// (Req 11.3, 11.4) with independent format handling (Req 11.5).
//
// ExportCSV and ExportPDF render the same column/row ReportResult that
// Generate (report.go) produces, so an export reuses the already-computed,
// already-tenant-scoped figures without re-deriving anything. The two
// functions are completely independent of each other: each takes a
// ReportResult and returns bytes, neither calls the other, and neither shares
// mutable state. A failure producing one format therefore cannot affect the
// other — which is how the HTTP layer satisfies Req 11.5 (see api/reports.go:
// each requested format is produced by a separate, error-isolated call, and
// the response carries each format that succeeds).
//
// Cell formatting mirrors the frontend export.js column/row model: string
// values pass through; numeric values (Column.Num) are rendered with thousands
// separators and two decimals for monetary figures, while whole-number counts
// stay integers. The CSV is prefixed with a UTF-8 BOM (like export.js) so
// Excel opens it as UTF-8.
package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/go-pdf/fpdf"
)

// utf8BOM is the byte-order mark export.js prepends so spreadsheet software
// (notably Excel) detects the CSV as UTF-8 rather than the system codepage.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// formatCell renders a single row value to its display string for export.
//
// The rule mirrors what the frontend renders: strings pass through unchanged;
// integer counts stay integers; floating-point figures are rendered with two
// decimals. A nil value (a missing field) becomes the empty string. This keeps
// CSV and PDF output consistent with each other and with the on-screen table.
func formatCell(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', 2, 64)
	case float64:
		// Whole numbers (e.g. a count stored as float) render without a
		// trailing ".00"; everything else keeps two decimals.
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', 2, 64)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// cell returns the formatted value for column c in row r, treating an absent
// key as an empty cell.
func cell(r map[string]any, key string) string {
	if r == nil {
		return ""
	}
	return formatCell(r[key])
}

// ExportCSV renders a ReportResult to CSV bytes (Req 11.3). The first record is
// the header row built from the column labels; each subsequent record holds the
// row's cells in column order, formatted by formatCell. Quoting and escaping
// are handled by encoding/csv. A UTF-8 BOM is prepended so Excel reads the file
// as UTF-8 (matching the frontend export.js behavior).
//
// This function is independent of ExportPDF: it shares no state and never calls
// it, so a CSV failure cannot affect PDF production and vice versa (Req 11.5).
func ExportCSV(r ReportResult) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(utf8BOM)

	w := csv.NewWriter(&buf)

	header := make([]string, len(r.Columns))
	for i, c := range r.Columns {
		header[i] = c.Label
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, row := range r.Rows {
		record := make([]string, len(r.Columns))
		for i, c := range r.Columns {
			record[i] = cell(row, c.Key)
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportPDF renders a ReportResult to a simple single-table PDF (Req 11.4): a
// title heading followed by a header row and one row per ReportResult row,
// using github.com/go-pdf/fpdf. Column widths are split evenly across the
// printable page width; numeric columns are right-aligned. Text is passed
// through fpdf's UTF-8 -> cp1252 translator so characters like the em dash
// placeholder render correctly with the built-in core fonts (no font files
// required), keeping the export self-contained and robust.
//
// This function is independent of ExportCSV: it shares no state and never calls
// it, so a PDF failure cannot affect CSV production and vice versa (Req 11.5).
func ExportPDF(r ReportResult) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("") // UTF-8 -> cp1252
	pdf.SetTitle(r.Title, true)
	pdf.AddPage()

	// Title heading.
	title := r.Title
	if title == "" {
		title = "Report"
	}
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 12, tr(title), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Compute even column widths over the printable page width.
	cols := r.Columns
	if len(cols) == 0 {
		// Nothing to tabulate; still produce a valid (title-only) document.
		return outputPDF(pdf)
	}
	left, _, right, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	usable := pageW - left - right
	colW := usable / float64(len(cols))

	// Header row.
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	for _, c := range cols {
		pdf.CellFormat(colW, 8, tr(c.Label), "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	// Body rows.
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	for _, row := range r.Rows {
		for _, c := range cols {
			align := "L"
			if c.Num {
				align = "R"
			}
			pdf.CellFormat(colW, 7, tr(truncateForCell(cell(row, c.Key), colW)), "1", 0, align, false, 0, "")
		}
		pdf.Ln(-1)
	}

	return outputPDF(pdf)
}

// truncateForCell shortens a cell's text so it is unlikely to overflow a column
// of the given width (mm). It is a best-effort guard for the fixed-width table
// layout; the heuristic ~1.9mm per character at 9pt keeps the table readable
// without measuring every string.
func truncateForCell(s string, widthMM float64) string {
	max := int(widthMM / 1.9)
	if max < 4 || len([]rune(s)) <= max {
		return s
	}
	runes := []rune(s)
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

// outputPDF writes the assembled document to a byte slice, surfacing any error
// fpdf accumulated during construction.
func outputPDF(pdf *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
