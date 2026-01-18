package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"lume/internal/config"
	"lume/internal/report"
	"lume/internal/timewarrior"
)

var (
	timewDataDir string
	outputDir    string
	year         int
)

var reportTime string

var (
	reportFrom string
	reportTo   string
)

var rootCmd = &cobra.Command{
	Use:   "lume",
	Short: "Generate time reports from timewarrior entries",
	Long:  `Lume analyzes your timewarrior entries and generates comprehensive time reports organized by year, month, and week.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate reports",
	RunE:  runGenerate,
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show a report for a day, week, or month",
	RunE:  runReport,
}

var reportDayCmd = &cobra.Command{
	Use:   "day",
	Short: "Show a report for a day",
	RunE:  runReportDay,
}

var reportWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show a report for a week",
	RunE:  runReportWeek,
}

var reportMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Show a report for a month",
	RunE:  runReportMonth,
}

var reportRangeCmd = &cobra.Command{
	Use:   "range",
	Short: "Show a report for a date range",
	RunE:  runReportRange,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure default paths",
	RunE:  runConfigWizard,
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

	if cfg, err := config.Load(); err == nil {
		if cfg.TimewarriorDir != "" {
			timewDataDir = cfg.TimewarriorDir
		}
		if cfg.OutputDir != "" {
			outputDir = cfg.OutputDir
		}
	}

	reportDayCmd.Flags().StringVarP(&reportTime, "time", "t", "", "day to report (YYYY-MM-DD)")
	reportWeekCmd.Flags().StringVarP(&reportTime, "time", "t", "", "week to report (YYYY-MM-DD in that week)")
	reportMonthCmd.Flags().StringVarP(&reportTime, "time", "t", "", "month to report (YYYY-MM)")

	reportRangeCmd.Flags().StringVar(&reportFrom, "from", "", "range start (YYYY-MM-DD)")
	reportRangeCmd.Flags().StringVar(&reportTo, "to", "", "range end (YYYY-MM-DD)")

	reportCmd.AddCommand(reportDayCmd)
	reportCmd.AddCommand(reportWeekCmd)
	reportCmd.AddCommand(reportMonthCmd)
	reportCmd.AddCommand(reportRangeCmd)

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(configCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
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

func runReport(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("use one of the report subcommands: day, week, or month")
}

func runReportDay(cmd *cobra.Command, args []string) error {
	entries, err := timewarrior.ParseDataDir(timewDataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}

	day := reportTime
	if day == "" {
		day = time.Now().Format("2006-01-02")
	}

	return report.PrintDayReport(entries, day)
}

func runReportWeek(cmd *cobra.Command, args []string) error {
	entries, err := timewarrior.ParseDataDir(timewDataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}

	week := reportTime
	if week == "" {
		week = time.Now().Format("2006-01-02")
	}

	return report.PrintWeekReport(entries, week)
}

func runReportMonth(cmd *cobra.Command, args []string) error {
	entries, err := timewarrior.ParseDataDir(timewDataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}

	month := reportTime
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	return report.PrintMonthReport(entries, month)
}

func runReportRange(cmd *cobra.Command, args []string) error {
	if reportFrom == "" || reportTo == "" {
		return fmt.Errorf("both --from and --to are required")
	}

	entries, err := timewarrior.ParseDataDir(timewDataDir)
	if err != nil {
		return fmt.Errorf("failed to parse timewarrior data: %w", err)
	}

	return report.PrintRangeReport(entries, reportFrom, reportTo)
}

func runConfigWizard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Press Enter to keep the current value.")

	timewarrior, err := promptPath("Timewarrior data directory", cfg.TimewarriorDir)
	if err != nil {
		return err
	}
	output, err := promptPath("Output directory for reports", cfg.OutputDir)
	if err != nil {
		return err
	}

	cfg.TimewarriorDir = timewarrior
	cfg.OutputDir = output

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved.")
	return nil
}

func promptPath(label, current string) (string, error) {
	prompt := label
	if current != "" {
		prompt = fmt.Sprintf("%s [%s]", label, current)
	}
	fmt.Printf("%s: ", prompt)

	var input string
	if _, err := fmt.Fscanln(os.Stdin, &input); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	if input == "" {
		return current, nil
	}

	expanded, err := expandTilde(input)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(expanded); err != nil {
		return "", fmt.Errorf("path does not exist: %s", expanded)
	}

	return expanded, nil
}

func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
