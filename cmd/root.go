package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"lume/internal/report"
	"lume/internal/timewarrior"
)

var (
	timewDataDir string
	outputDir    string
	year         int
)

var rootCmd = &cobra.Command{
	Use:   "lume",
	Short: "Generate time reports from timewarrior entries",
	Long:  `Lume analyzes your timewarrior entries and generates comprehensive time reports organized by year, month, and week.`,
	RunE:  run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()

	rootCmd.Flags().StringVarP(&timewDataDir, "timewarrior", "t", filepath.Join(home, ".config", "timewarrior", "data"), "timewarrior data directory")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", filepath.Join(home, "wiki", "report"), "output directory for reports")
	rootCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "year to generate report for")
}

func run(cmd *cobra.Command, args []string) error {
	fmt.Printf("Loading timewarrior data from: %s\n", timewDataDir)
	entries, err := timewarrior.ParseDataDir(timewDataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}
	fmt.Printf("Found %d time entries\n", len(entries))

	fmt.Printf("Generating report for year %d...\n", year)
	if err := report.Generate(entries, outputDir, year); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	reportPath := filepath.Join(outputDir, fmt.Sprintf("%d", year))
	fmt.Printf("Report generated: %s/\n", reportPath)

	return nil
}
