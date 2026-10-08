package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fdats/internal/api"
	"fdats/internal/domain"
	"fdats/internal/report"
)

// reportTitles maps report kinds to their Excel titles (SRS §29).
var reportTitles = map[string]string{
	api.ReportRegister: "FD Register",
	api.ReportMaturity: "Maturity Report",
	api.ReportActive:   "Active FD Report",
	api.ReportClosed:   "Closed FD Report",
	api.ReportMembers:  "Member-wise FD Summary",
}

// ExportReport gathers the requested dataset and writes it to an .xlsx file
// at the user-chosen path (SRS §29).
func (s *FDService) ExportReport(req api.ReportRequest) (api.ReportResult, error) {
	title, ok := reportTitles[req.Kind]
	if !ok {
		return api.ReportResult{}, domain.ErrInvalidReport
	}
	path := strings.TrimSpace(req.Path)
	if path == "" {
		return api.ReportResult{}, domain.ErrExportPathRequired
	}
	if !strings.HasSuffix(strings.ToLower(path), ".xlsx") {
		path += ".xlsx"
	}

	generatedAt := s.nowUTC().In(time.Local)

	opts, err := s.reportOptions(req, title, generatedAt)
	if err != nil {
		return api.ReportResult{}, err
	}
	opts.Path = path

	if err := report.Export(opts); err != nil {
		return api.ReportResult{}, err
	}

	return api.ReportResult{Path: path, RowCount: reportRowCount(opts)}, nil
}

// previewRowLimit bounds the rows returned for the on-screen preview; the
// exported file always contains every row.
const previewRowLimit = 100

// PreviewReport returns the headings and first rows of a report so the user
// can see exactly what will be exported.
func (s *FDService) PreviewReport(req api.ReportRequest) (api.ReportPreview, error) {
	title, ok := reportTitles[req.Kind]
	if !ok {
		return api.ReportPreview{}, domain.ErrInvalidReport
	}

	generatedAt := s.nowUTC().In(time.Local)
	opts, err := s.reportOptions(req, title, generatedAt)
	if err != nil {
		return api.ReportPreview{}, err
	}

	headers, rows := report.Sheet(opts)

	preview := make([][]string, 0, min(len(rows), previewRowLimit))
	for i, row := range rows {
		if i >= previewRowLimit {
			break
		}
		converted := make([]string, len(row))
		for j, cell := range row {
			converted[j] = formatPreviewCell(cell)
		}
		preview = append(preview, converted)
	}

	return api.ReportPreview{
		Kind:    req.Kind,
		Headers: headers,
		Rows:    preview,
		Total:   len(rows),
	}, nil
}

// reportOptions loads the dataset for a report kind into export options
// (SRS §29). The member-wise summary reads members instead of FD rows.
func (s *FDService) reportOptions(req api.ReportRequest, title string, generatedAt time.Time) (report.Options, error) {
	opts := report.Options{
		Kind:        req.Kind,
		Title:       title,
		GeneratedAt: generatedAt,
	}
	if req.Kind == api.ReportMembers {
		members, err := s.members.AllWithStats()
		if err != nil {
			return opts, err
		}
		count, amount, err := s.fds.UnassignedActiveSummary()
		if err != nil {
			return opts, err
		}
		opts.Members = members
		opts.UnassignedActiveFDCount = count
		opts.UnassignedActiveAmount = amount
		return opts, nil
	}
	fds, err := s.reportData(req, generatedAt)
	if err != nil {
		return opts, err
	}
	opts.FDs = fds
	return opts, nil
}

// reportRowCount is the number of data rows a report will contain.
func reportRowCount(opts report.Options) int {
	if opts.Kind == api.ReportMembers {
		n := len(opts.Members)
		if opts.UnassignedActiveFDCount > 0 {
			n++
		}
		return n
	}
	return len(opts.FDs)
}

// reportData loads the FD dataset for a report kind (SRS §29).
func (s *FDService) reportData(req api.ReportRequest, generatedAt time.Time) ([]domain.FixedDeposit, error) {
	var (
		fds []domain.FixedDeposit
		err error
	)
	switch req.Kind {
	case api.ReportRegister:
		fds, err = s.fds.ListAll()
	case api.ReportActive:
		fds, err = s.fds.ListAllByStatus(string(domain.StatusActive))
	case api.ReportClosed:
		fds, err = s.fds.ListAllByStatus(string(domain.StatusClosed))
	case api.ReportMaturity:
		fds, err = s.maturityRange(req, generatedAt)
	default:
		err = domain.ErrInvalidReport
	}
	if err != nil {
		return nil, err
	}
	if fds == nil {
		fds = []domain.FixedDeposit{}
	}
	return fds, nil
}

// formatPreviewCell renders one Excel cell value as display text.
func formatPreviewCell(v any) string {
	switch t := v.(type) {
	case time.Time:
		if t.IsZero() {
			return ""
		}
		return domain.FormatDate(t)
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

// maturityRange resolves the date window for the maturity report. An empty
// range defaults to today .. today + 90 days (SRS §28).
func (s *FDService) maturityRange(req api.ReportRequest, generatedAt time.Time) ([]domain.FixedDeposit, error) {
	from := strings.TrimSpace(req.FromDate)
	if from == "" {
		from = domain.FormatDate(generatedAt)
	}
	to := strings.TrimSpace(req.ToDate)
	if to == "" {
		to = domain.FormatDate(domain.AddDays(generatedAt, 90))
	}
	if _, err := domain.ParseDate(from); err != nil {
		return nil, domain.ErrInvalidDate
	}
	if _, err := domain.ParseDate(to); err != nil {
		return nil, domain.ErrInvalidDate
	}
	return s.fds.ListByMaturityRange(from, to)
}
