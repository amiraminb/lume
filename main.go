package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lume/internal/report"
	"lume/internal/report/build"
	"lume/internal/report/render"
	"lume/internal/timewarrior"
)

func main() {
	cfg, entries, err := timewarrior.ParseStdin(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lume: %v\n", err)
		os.Exit(1)
	}

	if cfg.HasTag("generate") {
		if err := runGenerate(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "lume: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runReport(cfg, entries); err != nil {
		fmt.Fprintf(os.Stderr, "lume: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(cfg timewarrior.TimewConfig) error {
	outputDir := cfg.Get("reports.lume.output")
	if outputDir == "" {
		return fmt.Errorf("reports.lume.output not set in timewarrior config")
	}
	outputDir = expandTilde(outputDir)

	dataDir := resolveDataDir(cfg)
	if dataDir == "" {
		return fmt.Errorf("cannot determine timewarrior data directory")
	}

	fmt.Fprintf(os.Stderr, "Loading timewarrior data from: %s\n", dataDir)
	entries, err := timewarrior.ParseDataDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Found %d time entries\n", len(entries))

	if err := report.Generate(entries, outputDir, 0); err != nil {
		return fmt.Errorf("failed to generate reports: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Reports generated under: %s/\n", outputDir)
	return nil
}

func runReport(cfg timewarrior.TimewConfig, entries []timewarrior.Entry) error {
	start, hasStart := cfg.ReportStart()
	end, hasEnd := cfg.ReportEnd()

	if !hasStart || !hasEnd {
		// No date range provided — show a range report of all data.
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
		render.RangeReport(os.Stdout, data, earliest, latest)
		return nil
	}

	days := int(end.Sub(start).Hours()/24 + 0.5)

	switch {
	case days <= 1:
		data := build.DayReport(entries, start)
		render.DayReport(os.Stdout, data)
	case days <= 7:
		data := build.WeekReport(entries, start)
		render.WeekReport(os.Stdout, data)
	case days <= 31:
		data := build.MonthReport(entries, start.Month(), start.Year())
		render.MonthReport(os.Stdout, data, start.Year())
	default:
		data := build.RangeReport(entries, start, end)
		render.RangeReport(os.Stdout, data, start, end)
	}

	return nil
}

func resolveDataDir(cfg timewarrior.TimewConfig) string {
	// temp.db is the data directory path provided by timewarrior.
	if db := cfg.Get("temp.db"); db != "" {
		return db
	}
	// Fallback to default location.
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "timewarrior", "data")
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
