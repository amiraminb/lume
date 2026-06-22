package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amiraminb/lume/internal/report/build"
	"github.com/amiraminb/lume/internal/report/render"
	"github.com/amiraminb/lume/internal/timewarrior"
)

func main() {
	cfg, entries, err := timewarrior.ParseStdin(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lume: %v\n", err)
		os.Exit(1)
	}

	if err := runReport(cfg, entries); err != nil {
		fmt.Fprintf(os.Stderr, "lume: %v\n", err)
		os.Exit(1)
	}
}

func runReport(cfg timewarrior.TimewConfig, entries []timewarrior.Entry) error {
	birthdayMonth, birthdayDay, err := cfg.Birthday()
	if err != nil {
		return err
	}

	format := resolveFormat(cfg)

	start, hasStart := cfg.ReportStart()
	end, hasEnd := cfg.ReportEnd()

	if !hasStart || !hasEnd {
		if len(entries) == 0 {
			fmt.Println("No entries found.")
			return nil
		}
		earliest := entries[0].Start
		latest := entries[0].End
		for _, e := range entries[1:] {
			if e.Start.Before(earliest) {
				earliest = e.Start
			}
			if e.End.After(latest) {
				latest = e.End
			}
		}
		data := build.RangeReport(entries, earliest, latest)
		if format == formatColor {
			render.RangeReportANSI(os.Stdout, data, earliest, latest, birthdayMonth, birthdayDay)
		} else {
			render.RangeReport(os.Stdout, data, earliest, latest, birthdayMonth, birthdayDay)
		}
		return nil
	}

	days := int(end.Sub(start).Hours()/24 + 0.5)

	nextMonth := start.AddDate(0, 1, 0)
	isFullMonth := start.Day() == 1 && end.Year() == nextMonth.Year() && end.Month() == nextMonth.Month()

	switch {
	case days <= 1:
		data := build.DayReport(entries, start)
		if format == formatColor {
			render.DayReportANSI(os.Stdout, data, birthdayMonth, birthdayDay)
		} else {
			render.DayReport(os.Stdout, data, birthdayMonth, birthdayDay)
		}
	case days <= 7:
		allEntries, err := loadAllEntries(cfg)
		if err != nil {
			return err
		}
		data := build.WeekReport(allEntries, start)
		if format == formatColor {
			render.WeekReportANSI(os.Stdout, data, birthdayMonth, birthdayDay)
		} else {
			render.WeekReport(os.Stdout, data, birthdayMonth, birthdayDay)
		}
	case isFullMonth:
		data := build.MonthReport(entries, start.Month(), start.Year())
		if format == formatColor {
			render.MonthReportANSI(os.Stdout, data, start.Year(), birthdayMonth, birthdayDay)
		} else {
			render.MonthReport(os.Stdout, data, start.Year(), birthdayMonth, birthdayDay)
		}
	default:
		data := build.RangeReport(entries, start, end)
		if format == formatColor {
			render.RangeReportANSI(os.Stdout, data, start, end, birthdayMonth, birthdayDay)
		} else {
			render.RangeReport(os.Stdout, data, start, end, birthdayMonth, birthdayDay)
		}
	}

	return nil
}

const (
	formatMarkdown = "markdown"
	formatColor    = "color"
)

// resolveFormat picks the output format by precedence: LUME_FORMAT env var,
// then the reports.lume.format config key, then the default (color). Unknown
// values fall back to the default rather than erroring.
func resolveFormat(cfg timewarrior.TimewConfig) string {
	candidates := []string{os.Getenv("LUME_FORMAT"), cfg.Format()}
	for _, c := range candidates {
		switch strings.ToLower(strings.TrimSpace(c)) {
		case formatMarkdown:
			return formatMarkdown
		case formatColor:
			return formatColor
		}
	}
	return formatColor
}

func loadAllEntries(cfg timewarrior.TimewConfig) ([]timewarrior.Entry, error) {
	dataDir := resolveDataDir(cfg)
	if dataDir == "" {
		return nil, fmt.Errorf("cannot determine timewarrior data directory")
	}
	return timewarrior.ParseDataDir(dataDir)
}

func resolveDataDir(cfg timewarrior.TimewConfig) string {
	if db := cfg.Get("temp.db"); db != "" {
		return filepath.Join(db, "data")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "timewarrior", "data")
}
