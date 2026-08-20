package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service/report"
)

// reportDateLayout is the ISO yyyy-mm-dd layout accepted for the start/end
// query parameters, matching the date format stored on records.
const reportDateLayout = "2006-01-02"

// Supported export formats for the ?format= query parameter.
const (
	formatJSON = "json"
	formatCSV  = "csv"
	formatPDF  = "pdf"
)

// reportHandlers serves the read-only report endpoint (Req 11). Each report is
// produced by the Report_Service over tenant + owner scoped records, so a
// report only ever aggregates the requester's own data (Req 11.6).
type reportHandlers struct {
	svc *report.Service
}

// newReportHandlers builds the report handlers over a database handle.
func newReportHandlers(deps Deps) *reportHandlers {
	return &reportHandlers{svc: report.NewService(deps.DB)}
}

// generate handles GET /api/reports/{type}?start&end&format. The {type} path
// value selects the report; the optional start and end query parameters
// (yyyy-mm-dd) apply inclusive date filtering (Req 11.1, 11.2). The optional
// format parameter selects the export representation: json (default), csv (Req
// 11.3), or pdf (Req 11.4). An unknown report type, an unparseable date, or an
// unknown format yields a 400 validation error.
//
// Independent format handling (Req 11.5). format may name a single format or a
// comma-separated preference list (e.g. "csv,pdf"). Each requested format is
// produced by its own error-isolated call to an independent exporter
// (report.ExportCSV / report.ExportPDF, which share no state), so a failure
// producing one format never prevents producing another that succeeds: the
// handler walks the requested formats in order and responds with the first one
// that is produced successfully, skipping any that fail. Only when every
// requested format fails does the handler surface an error.
func (h *reportHandlers) generate(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}

	start, err := parseReportDate(r.URL.Query().Get("start"), "start")
	if err != nil {
		return err
	}
	end, err := parseReportDate(r.URL.Query().Get("end"), "end")
	if err != nil {
		return err
	}

	formats, err := parseReportFormats(r.URL.Query().Get("format"))
	if err != nil {
		return err
	}

	reportType := r.PathValue("type")
	result, err := h.svc.Generate(p, report.ReportRequest{
		Type:      report.ReportType(reportType),
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return err
	}

	// Produce each requested format independently and respond with the first
	// that succeeds; a failure in one format does not block another (Req 11.5).
	var lastErr error
	for _, f := range formats {
		switch f {
		case formatJSON:
			middleware.WriteJSON(w, http.StatusOK, result)
			return nil
		case formatCSV:
			data, exportErr := report.ExportCSV(result)
			if exportErr != nil {
				lastErr = exportErr
				continue
			}
			writeAttachment(w, "text/csv; charset=utf-8", reportType, "csv", data)
			return nil
		case formatPDF:
			data, exportErr := report.ExportPDF(result)
			if exportErr != nil {
				lastErr = exportErr
				continue
			}
			writeAttachment(w, "application/pdf", reportType, "pdf", data)
			return nil
		}
	}

	// Every requested format failed to render. Surface the last failure as an
	// internal error (untyped errors map to a safe 500 per Req 18.4).
	return lastErr
}

// writeAttachment writes a binary export body with the given content type as a
// downloadable attachment named "<reportType>-report.<ext>".
func writeAttachment(w http.ResponseWriter, contentType, reportType, ext string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s-report.%s"`, reportType, ext))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// parseReportFormats parses the optional format query parameter into an ordered
// list of formats to attempt. An empty value defaults to JSON. The value may be
// a single format or a comma-separated preference list; each entry must be one
// of json, csv, or pdf, otherwise a 400 validation error identifies the
// offending value (Req 18.3).
func parseReportFormats(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{formatJSON}, nil
	}
	parts := strings.Split(raw, ",")
	formats := make([]string, 0, len(parts))
	for _, part := range parts {
		f := strings.ToLower(strings.TrimSpace(part))
		if f == "" {
			continue
		}
		switch f {
		case formatJSON, formatCSV, formatPDF:
			formats = append(formats, f)
		default:
			return nil, apierr.ValidationField("format",
				"must be one of json, csv, pdf")
		}
	}
	if len(formats) == 0 {
		return []string{formatJSON}, nil
	}
	return formats, nil
}

// parseReportDate parses an optional yyyy-mm-dd query parameter. An empty value
// means the bound is unset (nil); a malformed value is a 400 validation error
// identifying the offending parameter (Req 18.3).
func parseReportDate(raw, field string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(reportDateLayout, raw)
	if err != nil {
		return nil, apierr.ValidationField(field, "must be a date in yyyy-mm-dd format")
	}
	return &t, nil
}
