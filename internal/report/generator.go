package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"lume/internal/timewarrior"
)

type TaskSummary struct {
	Description string
	TotalTime   float64
	Sessions    int
	Tags        map[string]bool
}

type WeekData struct {
	WeekNum int
	Start   time.Time
	End     time.Time
	Tasks   []TaskSummary
	ByTag   map[string]float64
	Total   float64
}

type MonthData struct {
	Month time.Month
	Weeks []WeekData
	Total float64
}

type YearReport struct {
	Year   int
	Months []MonthData
	Total  float64
}

func Generate(entries []timewarrior.Entry, outputDir string, year int) error {
	yearDir := filepath.Join(outputDir, fmt.Sprintf("%d", year))
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		return err
	}

	report := buildYearReport(entries, year)

	if err := writeYearIndex(report, yearDir); err != nil {
		return err
	}

	for _, month := range report.Months {
		if err := writeMonthFile(month, yearDir, year); err != nil {
			return err
		}
	}

	return nil
}

func buildYearReport(entries []timewarrior.Entry, year int) YearReport {
	filtered := filterByYear(entries, year)
	byMonth := groupByMonth(filtered)

	var months []MonthData
	var yearTotal float64

	for month := time.January; month <= time.December; month++ {
		monthEntries := byMonth[month]
		if len(monthEntries) == 0 {
			continue
		}

		weeks := groupByWeek(monthEntries)
		var monthTotal float64
		for _, w := range weeks {
			monthTotal += w.Total
		}
		yearTotal += monthTotal

		months = append(months, MonthData{
			Month: month,
			Weeks: weeks,
			Total: monthTotal,
		})
	}

	return YearReport{
		Year:   year,
		Months: months,
		Total:  yearTotal,
	}
}

func filterByYear(entries []timewarrior.Entry, year int) []timewarrior.Entry {
	var filtered []timewarrior.Entry
	for _, e := range entries {
		if e.Start.Year() == year {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func groupByMonth(entries []timewarrior.Entry) map[time.Month][]timewarrior.Entry {
	grouped := make(map[time.Month][]timewarrior.Entry)
	for _, e := range entries {
		month := e.Start.Month()
		grouped[month] = append(grouped[month], e)
	}
	return grouped
}

func groupByWeek(entries []timewarrior.Entry) []WeekData {
	weekMap := make(map[int][]timewarrior.Entry)

	for _, e := range entries {
		_, week := e.Start.ISOWeek()
		weekMap[week] = append(weekMap[week], e)
	}

	var weeks []WeekData
	for weekNum, weekEntries := range weekMap {
		tasks := aggregateByDescription(weekEntries)
		byTag := aggregateByTag(weekEntries)
		start, end := weekBounds(weekEntries)

		var total float64
		for _, e := range weekEntries {
			total += e.Duration().Hours()
		}

		weeks = append(weeks, WeekData{
			WeekNum: weekNum,
			Start:   start,
			End:     end,
			Tasks:   tasks,
			ByTag:   byTag,
			Total:   total,
		})
	}

	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].WeekNum < weeks[j].WeekNum
	})

	return weeks
}

func aggregateByDescription(entries []timewarrior.Entry) []TaskSummary {
	taskMap := make(map[string]*TaskSummary)

	for _, e := range entries {
		desc := e.Description
		if desc == "" {
			desc = "(no description)"
		}

		if _, exists := taskMap[desc]; !exists {
			taskMap[desc] = &TaskSummary{
				Description: desc,
				Tags:        make(map[string]bool),
			}
		}
		taskMap[desc].TotalTime += e.Duration().Hours()
		taskMap[desc].Sessions++
		for _, tag := range e.Tags {
			taskMap[desc].Tags[tag] = true
		}
	}

	var tasks []TaskSummary
	for _, t := range taskMap {
		tasks = append(tasks, *t)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].TotalTime > tasks[j].TotalTime
	})

	return tasks
}

func aggregateByTag(entries []timewarrior.Entry) map[string]float64 {
	tagTime := make(map[string]float64)
	for _, e := range entries {
		for _, tag := range e.Tags {
			tagTime[tag] += e.Duration().Hours()
		}
		if len(e.Tags) == 0 {
			tagTime["untagged"] += e.Duration().Hours()
		}
	}
	return tagTime
}

func weekBounds(entries []timewarrior.Entry) (time.Time, time.Time) {
	if len(entries) == 0 {
		return time.Time{}, time.Time{}
	}
	start, end := entries[0].Start, entries[0].End
	for _, e := range entries {
		if e.Start.Before(start) {
			start = e.Start
		}
		if e.End.After(end) {
			end = e.End
		}
	}
	return start, end
}

func formatDuration(hours float64) string {
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60

	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

func progressBar(value, max float64, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int((value / max) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func writeYearIndex(report YearReport, yearDir string) error {
	filename := filepath.Join(yearDir, "index.md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# Time Report %d\n\n", report.Year)
	fmt.Fprintf(file, "> **Total Tracked:** %s\n\n", formatDuration(report.Total))

	yearTags := make(map[string]float64)
	for _, month := range report.Months {
		for _, week := range month.Weeks {
			for tag, hours := range week.ByTag {
				yearTags[tag] += hours
			}
		}
	}

	if len(yearTags) > 0 {
		fmt.Fprintf(file, "## Year Overview\n\n")
		writeTagSummary(file, yearTags, report.Total)
		fmt.Fprintf(file, "\n")
	}

	fmt.Fprintf(file, "---\n\n")
	fmt.Fprintf(file, "## Months\n\n")

	for _, month := range report.Months {
		monthFile := fmt.Sprintf("%02d-%s.md", month.Month, strings.ToLower(month.Month.String()))
		fmt.Fprintf(file, "- [%s](%s) — %s\n", month.Month.String(), monthFile, formatDuration(month.Total))
	}

	return nil
}

func writeMonthFile(month MonthData, yearDir string, year int) error {
	filename := filepath.Join(yearDir, fmt.Sprintf("%02d-%s.md", month.Month, strings.ToLower(month.Month.String())))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# %s %d\n\n", month.Month.String(), year)
	fmt.Fprintf(file, "> **Monthly Total:** %s\n\n", formatDuration(month.Total))
	fmt.Fprintf(file, "---\n\n")

	monthTags := make(map[string]float64)
	for _, week := range month.Weeks {
		for tag, hours := range week.ByTag {
			monthTags[tag] += hours
		}
	}

	if len(monthTags) > 0 {
		fmt.Fprintf(file, "## Overview\n\n")
		writeTagSummary(file, monthTags, month.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	for _, week := range month.Weeks {
		writeWeek(file, week)
	}

	return nil
}

func writeTagSummary(file *os.File, tags map[string]float64, total float64) {
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tags[tagList[i]] > tags[tagList[j]]
	})

	maxTagTime := 0.0
	for _, hours := range tags {
		if hours > maxTagTime {
			maxTagTime = hours
		}
	}

	for _, tag := range tagList {
		hours := tags[tag]
		pct := (hours / total) * 100
		bar := progressBar(hours, maxTagTime, 20)
		fmt.Fprintf(file, "`%s` %s **%s** (%.0f%%)\n\n", tag, bar, formatDuration(hours), pct)
	}
}

func writeWeek(file *os.File, week WeekData) {
	fmt.Fprintf(file, "## Week %d\n", week.WeekNum)
	fmt.Fprintf(file, "> %s → %s\n\n",
		week.Start.Format("Mon, Jan 2"),
		week.End.Format("Mon, Jan 2"))

	fmt.Fprintf(file, "**Total:** %s\n\n", formatDuration(week.Total))

	if len(week.Tasks) > 0 {
		fmt.Fprintf(file, "| # | Task | Tags | Time | Sessions |\n")
		fmt.Fprintf(file, "|--:|:-----|:-----|-----:|---------:|\n")

		for i, t := range week.Tasks {
			tags := formatTags(t.Tags)
			if tags == "" {
				tags = "—"
			}
			fmt.Fprintf(file, "| %d | %s | %s | %s | %d |\n",
				i+1,
				truncate(t.Description, 50),
				tags,
				formatDuration(t.TotalTime),
				t.Sessions)
		}
		fmt.Fprintf(file, "\n")
	}

	fmt.Fprintf(file, "---\n\n")
}

func formatTags(tags map[string]bool) string {
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	sort.Strings(tagList)
	return strings.Join(tagList, ", ")
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
