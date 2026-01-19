package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func PrintDayReport(entries []timewarrior.Entry, day string) error {
	date, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return fmt.Errorf("invalid day %q (expected YYYY-MM-DD): %w", day, err)
	}

	report := build.DayReport(entries, date)
	return writeToStdout(func(file *os.File) {
		render.DayReport(file, report)
	})
}

func PrintWeekReport(entries []timewarrior.Entry, date string) error {
	parsed, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return fmt.Errorf("invalid week date %q (expected YYYY-MM-DD): %w", date, err)
	}

	report := build.WeekReport(entries, parsed)
	return writeToStdout(func(file *os.File) {
		render.WeekReport(file, report)
	})
}

func PrintMonthReport(entries []timewarrior.Entry, month string) error {
	parsed, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return fmt.Errorf("invalid month %q (expected YYYY-MM): %w", month, err)
	}

	report := build.MonthReport(entries, parsed.Month(), parsed.Year())
	return writeToStdout(func(file *os.File) {
		render.MonthReport(file, report, parsed.Year())
	})
}

func PrintRangeReport(entries []timewarrior.Entry, start, end string) error {
	startDate, err := time.ParseInLocation("2006-01-02", start, time.Local)
	if err != nil {
		return fmt.Errorf("invalid from date %q (expected YYYY-MM-DD): %w", start, err)
	}
	endDate, err := time.ParseInLocation("2006-01-02", end, time.Local)
	if err != nil {
		return fmt.Errorf("invalid to date %q (expected YYYY-MM-DD): %w", end, err)
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("range end must be on or after range start")
	}

	endExclusive := endDate.AddDate(0, 0, 1)
	report := build.RangeReport(entries, startDate, endExclusive)
	return writeToStdout(func(file *os.File) {
		render.RangeReport(file, report, startDate, endDate)
	})
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

func writeToStdout(write func(file *os.File)) error {
	write(os.Stdout)
	return nil
}
