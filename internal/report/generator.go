package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lume/internal/report/build"
	"lume/internal/report/model"
	"lume/internal/report/render"
	"lume/internal/timewarrior"
)

func Generate(entries []timewarrior.Entry, outputDir string, year int) error {
	reports := build.YearReports(entries)
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		yearDir := filepath.Join(outputDir, fmt.Sprintf("%d", report.Year))
		if err := os.MkdirAll(yearDir, 0755); err != nil {
			return err
		}

		if err := writeYearIndex(report, yearDir); err != nil {
			return err
		}

		for _, month := range report.Months {
			if err := writeMonthFile(month, yearDir, report.Year); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeYearIndex(report model.YearReport, yearDir string) error {
	filename := filepath.Join(yearDir, "index.md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	render.YearIndex(file, report)
	return nil
}

func writeMonthFile(month model.MonthData, yearDir string, year int) error {
	filename := filepath.Join(yearDir, fmt.Sprintf("%02d-%s.md", month.Month, strings.ToLower(month.Month.String())))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	render.MonthFile(file, month, year)
	return nil
}
