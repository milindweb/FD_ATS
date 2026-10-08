// Package report exports Fixed Deposit data to Excel (.xlsx) files (SRS §29).
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"fdats/internal/domain"
)

// Options configures a single report export.
type Options struct {
	Kind        string
	Title       string
	FDs         []domain.FixedDeposit
	Members     []domain.MemberWithStats // member-wise summary (SRS §29.6)
	GeneratedAt time.Time
	Path        string

	// UnassignedActiveFDCount/UnassignedActiveAmount aggregate active FDs
	// with no member link, reported as the "Unassigned" row (SRS §53).
	UnassignedActiveFDCount int64
	UnassignedActiveAmount  int64
}

const sheetName = "Report"

const (
	headerRow  = 4 // row 1 = title, row 2 = generated date, row 4 = headings
	dataOffset = 1
)

// Export writes the requested report to opts.Path as an .xlsx file with a
// title, generation date, styled headings and Excel autofilter (SRS §29.5).
func Export(opts Options) error {
	if strings.TrimSpace(opts.Path) == "" {
		return domain.ErrExportPathRequired
	}
	if !strings.HasSuffix(strings.ToLower(opts.Path), ".xlsx") {
		opts.Path += ".xlsx"
	}

	headers, rows := buildSheet(opts)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return err
	}

	styles, err := newStyles(f)
	if err != nil {
		return err
	}

	lastCol := len(headers)

	// Row 1: report title.
	titleCell, _ := excelize.CoordinatesToCellName(1, 1)
	lastTitleCell, _ := excelize.CoordinatesToCellName(lastCol, 1)
	if err := f.MergeCell(sheetName, titleCell, lastTitleCell); err != nil {
		return err
	}
	if err := f.SetCellStr(sheetName, titleCell, opts.Title); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, titleCell, lastTitleCell, styles.title); err != nil {
		return err
	}

	// Row 2: generation date.
	genCell, _ := excelize.CoordinatesToCellName(1, 2)
	if err := f.SetCellStr(sheetName, genCell, "Generated: "+opts.GeneratedAt.Format("02/01/2006")); err != nil {
		return err
	}

	// Row 4: column headings.
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		if err := f.SetCellStr(sheetName, cell, h); err != nil {
			return err
		}
	}
	firstHeader, _ := excelize.CoordinatesToCellName(1, headerRow)
	lastHeader, _ := excelize.CoordinatesToCellName(lastCol, headerRow)
	if err := f.SetCellStyle(sheetName, firstHeader, lastHeader, styles.header); err != nil {
		return err
	}

	// Data rows with per-cell number/date formats.
	for r, row := range rows {
		excelRow := headerRow + dataOffset + r
		for c, cell := range row {
			name, _ := excelize.CoordinatesToCellName(c+1, excelRow)
			var err error
			switch v := cell.(type) {
			case time.Time:
				err = f.SetCellValue(sheetName, name, v)
				if err == nil {
					err = f.SetCellStyle(sheetName, name, name, styles.date)
				}
			case int64:
				err = f.SetCellValue(sheetName, name, v)
				if err == nil && isMoneyColumn(headers[c]) {
					err = f.SetCellStyle(sheetName, name, name, styles.money)
				}
			case float64:
				err = f.SetCellValue(sheetName, name, v)
				if err == nil {
					err = f.SetCellStyle(sheetName, name, name, styles.rate)
				}
			default:
				err = f.SetCellStr(sheetName, name, fmt.Sprintf("%v", v))
			}
			if err != nil {
				return err
			}
		}
	}

	// Excel filtering over the heading row (SRS §29.5).
	filterRef := fmt.Sprintf("%s:%s", firstHeader, lastHeader)
	if err := f.AutoFilter(sheetName, filterRef, nil); err != nil {
		return err
	}

	// Column widths.
	for i := 1; i <= lastCol; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		if err := f.SetColWidth(sheetName, col, col, columnWidth(headers[i-1])); err != nil {
			return err
		}
	}

	// Freeze the heading rows so scrolling keeps them visible.
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:   true,
		XSplit:   0,
		YSplit:   headerRow,
		TopLeftCell: fmt.Sprintf("A%d", headerRow+1),
	}); err != nil {
		return err
	}

	return f.SaveAs(opts.Path)
}

type sheetStyles struct {
	title  int
	header int
	date   int
	money  int
	rate   int
}

func newStyles(f *excelize.File) (sheetStyles, error) {
	var s sheetStyles
	var err error

	if s.title, err = f.NewStyle(&styleTitle); err != nil {
		return s, err
	}
	if s.header, err = f.NewStyle(&styleHeader); err != nil {
		return s, err
	}
	if s.date, err = f.NewStyle(&styleDate); err != nil {
		return s, err
	}
	if s.money, err = f.NewStyle(&styleMoney); err != nil {
		return s, err
	}
	if s.rate, err = f.NewStyle(&styleRate); err != nil {
		return s, err
	}
	return s, nil
}

var styleTitle = excelize.Style{
	Font:      &excelize.Font{Bold: true, Size: 14, Color: "1F2937"},
	Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
}

var styleHeader = excelize.Style{
	Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
	Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E79"}, Pattern: 1},
	Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	Border: []excelize.Border{
		{Type: "left", Color: "143A5A", Style: 1},
		{Type: "right", Color: "143A5A", Style: 1},
		{Type: "top", Color: "143A5A", Style: 1},
		{Type: "bottom", Color: "143A5A", Style: 1},
	},
}

var styleDate = excelize.Style{
	NumFmt:   14, // dd/mm/yyyy
	Alignment: &excelize.Alignment{Horizontal: "center"},
}

var styleMoney = excelize.Style{
	NumFmt: 37, // accounting-style thousands separator
}

var styleRate = excelize.Style{
	NumFmt:   10, // 0.00
	Alignment: &excelize.Alignment{Horizontal: "right"},
}

func isMoneyColumn(header string) bool {
	switch header {
	case "Days Remaining", "Tenure", "Status", "Active FD Count":
		return false
	}
	return true
}

func columnWidth(header string) float64 {
	switch header {
	case "FD Number":
		return 14
	case "Member Name", "Remark", "Closure Type":
		return 24
	case "GEN No.":
		return 16
	case "Designation":
		return 18
	case "Active FD Count":
		return 16
	case "Total Active FD Amount":
		return 22
	case "Status":
		return 12
	case "Start Date", "Maturity Date", "Closure Date":
		return 14
	case "Tenure", "Days Remaining":
		return 15
	}
	return 18
}
