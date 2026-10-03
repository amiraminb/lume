package render

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/amiraminb/lume/internal/report/model"
)

// CompactDayReport renders the useful totals and task breakdown without charts
// or bordered tables.
func CompactDayReport(file *os.File, report model.DayReport, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "%s %s %s %s\n",
		titleStyle.Render(fmt.Sprintf("Day %d", birthdayDayNumber(report.Date, birthdayMonth, birthdayDay))),
		subtleStyle.Render("·"),
		dateStyle.Render(report.Date.Format("Mon, Jan 2")),
		totalStyle.Render(formatDuration(report.Total)))

	writeCompactBreakdownLines(file, "Projects", report.ByProject, report.Total)
	writeCompactBreakdownLines(file, "Categories", report.ByTag, report.Total)
	writeCompactTaskGroups(file, report.Tasks)

	if len(report.Tasks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found."))
	}
}

// CompactWeekReport renders daily, project, category, and task totals as
// compact lines instead of charts and tables.
func CompactWeekReport(file *os.File, week model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "%s %s %s %s\n",
		titleStyle.Render(fmt.Sprintf("Week %d", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))),
		subtleStyle.Render("·"),
		dateStyle.Render(weekDateRange(week.Start, week.End)),
		totalStyle.Render(formatDuration(week.Total)))

	writeCompactDailyTotals(file, week)
	writeCompactBreakdownLines(file, "Projects", week.ByProject, week.Total)
	writeCompactBreakdownLines(file, "Categories", week.ByTag, week.Total)
	writeCompactTaskGroups(file, week.Tasks)

	if len(week.Tasks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found."))
	}
}

// CompactMonthReport renders a month as a one-line summary followed by weekly
// totals and aggregate breakdowns.
func CompactMonthReport(file *os.File, month model.MonthData, year int, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "%s %s\n",
		titleStyle.Render(fmt.Sprintf("%s %d", month.Month.String(), year)),
		totalStyle.Render(formatDuration(month.Total)))

	writeCompactWeekTotals(file, month.Weeks, birthdayMonth, birthdayDay)
	tags, projects := aggregateWeeks(month.Weeks)
	writeCompactBreakdownLines(file, "Projects", projects, month.Total)
	writeCompactBreakdownLines(file, "Categories", tags, month.Total)

	if len(month.Weeks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found."))
	}
}

// CompactRangeReport renders a custom range without the weekly matrix.
func CompactRangeReport(file *os.File, report model.MonthData, start, end time.Time, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "%s %s\n",
		dateStyle.Render(fmt.Sprintf("%s → %s", start.Format("Jan 2, 2006"), end.AddDate(0, 0, -1).Format("Jan 2, 2006"))),
		totalStyle.Render(formatDuration(report.Total)))

	writeCompactWeekTotals(file, report.Weeks, birthdayMonth, birthdayDay)
	tags, projects := aggregateWeeks(report.Weeks)
	writeCompactBreakdownLines(file, "Projects", projects, report.Total)
	writeCompactBreakdownLines(file, "Categories", tags, report.Total)

	if len(report.Weeks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found."))
	}
}

type compactMetric struct {
	label string
	hours float64
}

func writeCompactBreakdownLines(file *os.File, title string, values map[string]float64, total float64) {
	metrics := compactMetrics(values)
	if len(metrics) == 0 {
		return
	}

	fmt.Fprintln(file, headerStyle.Render(title+":"))
	labelWidth := 0
	durationWidth := 0
	for _, metric := range metrics {
		labelWidth = max(labelWidth, len([]rune(metric.label)))
		durationWidth = max(durationWidth, len(formatDuration(metric.hours)))
	}
	for _, metric := range metrics {
		share := 0.0
		if total > 0 {
			share = metric.hours / total * 100
		}
		label := fmt.Sprintf("%-*s", labelWidth, metric.label)
		duration := fmt.Sprintf("%*s", durationWidth, formatDuration(metric.hours))
		percentage := fmt.Sprintf("(%3.0f%%)", share)
		fmt.Fprintf(file, "  %s  %s  %s\n",
			projectStyle.Render(label),
			shareStyle.Render(duration),
			shareStyle.Render(percentage))
	}
}

func compactMetrics(values map[string]float64) []compactMetric {
	metrics := make([]compactMetric, 0, len(values))
	for label, hours := range values {
		if hours > 0 {
			metrics = append(metrics, compactMetric{label: label, hours: hours})
		}
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].hours != metrics[j].hours {
			return metrics[i].hours > metrics[j].hours
		}
		return metrics[i].label < metrics[j].label
	})
	return metrics
}

func writeCompactDailyTotals(file *os.File, week model.WeekData) {
	totals := make(map[time.Weekday]float64)
	for _, task := range week.Tasks {
		for day, hours := range task.DayTotals {
			totals[day] += hours
		}
	}

	if len(totals) == 0 {
		return
	}

	fmt.Fprintln(file, headerStyle.Render("Days:"))
	durationWidth := 0
	for _, hours := range totals {
		durationWidth = max(durationWidth, len(formatDuration(hours)))
	}
	for _, day := range []time.Weekday{
		time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
		time.Thursday, time.Friday, time.Saturday,
	} {
		if totals[day] > 0 {
			duration := fmt.Sprintf("%*s", durationWidth, formatDuration(totals[day]))
			fmt.Fprintf(file, "  %-3s  %s\n", day.String()[:3], shareStyle.Render(duration))
		}
	}
}

func writeCompactWeekTotals(file *os.File, weeks []model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	if len(weeks) == 0 {
		return
	}

	weekWidth := 0
	dateWidth := 0
	durationWidth := 0
	for _, week := range weeks {
		weekWidth = max(weekWidth, len(fmt.Sprintf("W%d", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))))
		dateWidth = max(dateWidth, len([]rune(weekDateRange(week.Start, week.End))))
		durationWidth = max(durationWidth, len(formatDuration(week.Total)))
	}

	fmt.Fprintln(file, headerStyle.Render("Weeks:"))
	for _, week := range weeks {
		weekLabel := fmt.Sprintf("%-*s", weekWidth, fmt.Sprintf("W%d", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay)))
		dateLabel := fmt.Sprintf("%-*s", dateWidth, weekDateRange(week.Start, week.End))
		duration := fmt.Sprintf("%*s", durationWidth, formatDuration(week.Total))
		fmt.Fprintf(file, "  %s  %s  %s\n",
			weekLabel,
			dateStyle.Render(dateLabel),
			shareStyle.Render(duration))
	}
}

func writeCompactTaskGroups(file *os.File, tasks []model.TaskSummary) {
	categories := []struct {
		title taskCategory
		name  string
	}{
		{categoryDev, "Dev"},
		{categoryMeetings, "Meetings"},
		{categoryKnowledge, "Knowledge"},
		{categoryMisc, "Misc"},
	}

	categorized := groupTasksByCategory(tasks)
	for _, category := range categories {
		group := categorized[category.title]
		if len(group) == 0 {
			continue
		}

		fmt.Fprintln(file, categoryHeaderStyle.Render(category.name+":"))
		durationWidth := 0
		projectWidth := 0
		descriptionWidth := 0
		for _, task := range group {
			durationWidth = max(durationWidth, len(formatDuration(task.TotalTime)))
			projectWidth = max(projectWidth, len([]rune(projectName(task))))
			descriptionWidth = max(descriptionWidth, len([]rune(truncate(task.Description, 72))))
		}
		for _, task := range sortTasksByProject(group) {
			duration := fmt.Sprintf("%*s", durationWidth, formatDuration(task.TotalTime))
			project := fmt.Sprintf("%-*s", projectWidth, projectName(task))
			description := fmt.Sprintf("%-*s", descriptionWidth, truncate(task.Description, 72))
			fmt.Fprintf(file, "  %s  %s  %s\n",
				projectStyle.Render(project),
				description,
				totalStyle.Render(duration))
		}
	}
}
