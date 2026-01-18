package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"lume/internal/report/model"
)

func YearIndex(file *os.File, report model.YearReport) {
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
}

func MonthFile(file *os.File, month model.MonthData, year int) {
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
		WeekSection(file, week)
	}
}

func DayReport(file *os.File, report model.DayReport) {
	fmt.Fprintf(file, "# %s\n\n", report.Date.Format("Monday, Jan 2, 2006"))
	fmt.Fprintf(file, "> **Daily Total:** %s\n\n", formatDuration(report.Total))

	if len(report.ByTag) > 0 {
		fmt.Fprintf(file, "## Overview\n\n")
		writeTagSummary(file, report.ByTag, report.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(report.Tasks) == 0 {
		fmt.Fprintf(file, "No entries found for this day.\n")
		return
	}

	fmt.Fprintf(file, "| # | Task | Time | Sessions |\n")
	fmt.Fprintf(file, "|--:|:-----|-----:|---------:|\n")
	for i, t := range report.Tasks {
		fmt.Fprintf(file, "| %d | %s | %s | %d |\n",
			i+1,
			truncate(t.Description, 55),
			formatDuration(t.TotalTime),
			t.Sessions)
	}
	fmt.Fprintf(file, "\n")
}

func WeekReport(file *os.File, week model.WeekData) {
	fmt.Fprintf(file, "# Week %d\n", week.WeekNum)
	fmt.Fprintf(file, "> %s → %s\n\n",
		week.Start.Format("Mon, Jan 2"),
		week.End.Format("Mon, Jan 2"))

	fmt.Fprintf(file, "**Total:** %s\n\n", formatDuration(week.Total))

	if len(week.ByTag) > 0 {
		fmt.Fprintf(file, "## Overview\n\n")
		writeTagSummary(file, week.ByTag, week.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(week.Tasks) == 0 {
		fmt.Fprintf(file, "No entries found for this week.\n")
		return
	}

	writeWeekTasks(file, week.Tasks)
}

func MonthReport(file *os.File, month model.MonthData, year int) {
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

	if len(month.Weeks) == 0 {
		fmt.Fprintf(file, "No entries found for this month.\n")
		return
	}

	for _, week := range month.Weeks {
		WeekSection(file, week)
	}
}

func WeekSection(file *os.File, week model.WeekData) {
	fmt.Fprintf(file, "## Week %d\n", week.WeekNum)
	fmt.Fprintf(file, "> %s → %s\n\n",
		week.Start.Format("Mon, Jan 2"),
		week.End.Format("Mon, Jan 2"))

	fmt.Fprintf(file, "**Total:** %s\n\n", formatDuration(week.Total))

	if len(week.ByTag) > 0 {
		fmt.Fprintf(file, "## Overview\n\n")
		writeTagSummary(file, week.ByTag, week.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(week.Tasks) > 0 {
		writeWeekTasks(file, week.Tasks)
	}

	fmt.Fprintf(file, "---\n\n")
}

func writeWeekTasks(file *os.File, tasks []model.TaskSummary) {
	tasksByTag := groupTasksByTag(tasks)

	var tags []string
	for tag := range tasksByTag {
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool {
		var totalI, totalJ float64
		for _, t := range tasksByTag[tags[i]] {
			totalI += t.TotalTime
		}
		for _, t := range tasksByTag[tags[j]] {
			totalJ += t.TotalTime
		}
		return totalI > totalJ
	})

	for _, tag := range tags {
		tasks := tasksByTag[tag]
		var tagTotal float64
		for _, t := range tasks {
			tagTotal += t.TotalTime
		}

		fmt.Fprintf(file, "### %s\n", tag)
		fmt.Fprintf(file, "**Subtotal:** %s\n\n", formatDuration(tagTotal))

		fmt.Fprintf(file, "| # | Task | Time | Sessions |\n")
		fmt.Fprintf(file, "|--:|:-----|-----:|---------:|\n")

		for i, t := range tasks {
			fmt.Fprintf(file, "| %d | %s | %s | %d |\n",
				i+1,
				truncate(t.Description, 55),
				formatDuration(t.TotalTime),
				t.Sessions)
		}
		fmt.Fprintf(file, "\n")
	}
}

func groupTasksByTag(tasks []model.TaskSummary) map[string][]model.TaskSummary {
	grouped := make(map[string][]model.TaskSummary)

	for _, t := range tasks {
		if len(t.Tags) == 0 {
			grouped["untagged"] = append(grouped["untagged"], t)
		} else {
			for tag := range t.Tags {
				grouped[tag] = append(grouped[tag], t)
			}
		}
	}

	for tag := range grouped {
		sort.Slice(grouped[tag], func(i, j int) bool {
			return grouped[tag][i].TotalTime > grouped[tag][j].TotalTime
		})
	}

	return grouped
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

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
