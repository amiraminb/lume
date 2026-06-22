package render

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/amiraminb/lume/internal/report/model"
	"github.com/charmbracelet/lipgloss"
)

const ansiBarWidth = 28

// renderColorBar draws a horizontal bar in the given color using eighth-block
// resolution, padded to ansiBarWidth with a dim track.
func renderColorBar(ratio float64, color lipgloss.Color) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	totalEighths := int(ratio*float64(ansiBarWidth)*8 + 0.5)
	full := totalEighths / 8
	rem := totalEighths % 8

	bar := make([]rune, 0, ansiBarWidth)
	for range full {
		bar = append(bar, fullBlock)
	}
	if rem > 0 && full < ansiBarWidth {
		bar = append(bar, eighthBlocks[rem])
	}
	filled := lipgloss.NewStyle().Foreground(color).Render(string(bar))

	trackLen := ansiBarWidth - len(bar)
	track := ""
	if trackLen > 0 {
		dots := make([]rune, trackLen)
		for i := range dots {
			dots[i] = emptyBlock
		}
		track = emptyStyle.Render(string(dots))
	}
	return filled + track
}

// writeColorShareChart renders a colored horizontal bar chart of labelled
// shares. Bars scale to the largest value; the percentage is of total.
func writeColorShareChart(file *os.File, title string, values map[string]float64, total float64) {
	rows := make([]chartRow, 0, len(values))
	var max float64
	for label, hours := range values {
		rows = append(rows, chartRow{label: label, hours: hours})
		if hours > max {
			max = hours
		}
	}
	if len(rows) == 0 || max <= 0 {
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].hours != rows[j].hours {
			return rows[i].hours > rows[j].hours
		}
		return rows[i].label < rows[j].label
	})

	labelWidth := 0
	for _, r := range rows {
		if n := len([]rune(r.label)); n > labelWidth {
			labelWidth = n
		}
	}

	fmt.Fprintln(file, headerStyle.Render(title))
	for _, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = (r.hours / total) * 100
		}
		label := lipgloss.NewStyle().Foreground(colorFor(r.label)).Render(fmt.Sprintf("%-*s", labelWidth, r.label))
		fmt.Fprintf(file, "%s  %s  %s  %s\n",
			label,
			renderColorBar(r.hours/max, colorFor(r.label)),
			fmt.Sprintf("%7s", formatDuration(r.hours)),
			subtleStyle.Render(fmt.Sprintf("%3.0f%%", pct)))
	}
	fmt.Fprintln(file)
}

var accentColor = lipgloss.Color("39")

// writeColorVerticalChart draws a colored column chart: columns rise from a
// baseline using vertical eighth-blocks, with a y-axis peak label and per-column
// labels above and values below. All columns share accentColor.
func writeColorVerticalChart(file *os.File, title string, columns []chartColumn, peakLabel string) {
	colWidth := 3
	for _, c := range columns {
		colWidth = max(colWidth, len([]rune(c.top)), len([]rune(c.bottom)))
	}

	levels := make([]int, len(columns))
	for i, c := range columns {
		r := c.ratio
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		levels[i] = int(r*float64(chartHeight)*8 + 0.5)
	}

	axisPad := len([]rune(peakLabel))
	barStyle := lipgloss.NewStyle().Foreground(accentColor)

	fmt.Fprintln(file, headerStyle.Render(title))

	for row := chartHeight - 1; row >= 0; row-- {
		if row == chartHeight-1 {
			fmt.Fprintf(file, "%s %s", subtleStyle.Render(peakLabel), subtleStyle.Render("┤"))
		} else {
			fmt.Fprintf(file, "%s %s", strings.Repeat(" ", axisPad), subtleStyle.Render("│"))
		}

		lo := row * 8
		for _, eighths := range levels {
			var glyph string
			switch {
			case eighths >= lo+8:
				glyph = string(fullBlock)
			case eighths > lo:
				glyph = string(partialVBlocks[eighths-lo])
			default:
				glyph = " "
			}
			cellText := fmt.Sprintf("%-*s", colWidth+1, glyph)
			if glyph == " " {
				fmt.Fprint(file, cellText)
			} else {
				fmt.Fprint(file, barStyle.Render(cellText))
			}
		}
		fmt.Fprintln(file)
	}

	fmt.Fprintf(file, "%s %s\n",
		strings.Repeat(" ", axisPad),
		subtleStyle.Render("└"+strings.Repeat("─", len(columns)*(colWidth+1))))

	gutter := strings.Repeat(" ", axisPad+2)
	fmt.Fprint(file, gutter)
	for _, c := range columns {
		fmt.Fprint(file, headerStyle.Render(fmt.Sprintf("%-*s", colWidth+1, c.top)))
	}
	fmt.Fprintln(file)
	fmt.Fprint(file, gutter)
	for _, c := range columns {
		fmt.Fprint(file, subtleStyle.Render(fmt.Sprintf("%-*s", colWidth+1, c.bottom)))
	}
	fmt.Fprintln(file)
	fmt.Fprintln(file)
}

// writeColorWeekdayChart renders a colored Sun–Sat column chart of daily totals.
func writeColorWeekdayChart(file *os.File, week model.WeekData) {
	dayTotals := make(map[time.Weekday]float64)
	for _, task := range week.Tasks {
		for day, hours := range task.DayTotals {
			dayTotals[day] += hours
		}
	}

	days := []time.Weekday{
		time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
		time.Thursday, time.Friday, time.Saturday,
	}

	var max float64
	for _, day := range days {
		if dayTotals[day] > max {
			max = dayTotals[day]
		}
	}
	if max <= 0 {
		return
	}

	columns := make([]chartColumn, len(days))
	for i, day := range days {
		hours := dayTotals[day]
		columns[i] = chartColumn{
			top:    day.String()[:3],
			bottom: compactDuration(hours),
			ratio:  hours / max,
		}
	}
	writeColorVerticalChart(file, "Daily Trend", columns, formatDuration(max))
}

// writeColorCategoryTable prints a category's tasks as a colored, aligned table.
func writeColorCategoryTable(file *os.File, title string, tasks []model.TaskSummary) {
	fmt.Fprintln(file, headerStyle.Render(title))
	if len(tasks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found."))
		fmt.Fprintln(file)
		return
	}

	sorted := sortTasksByProject(tasks)

	projWidth, descWidth := len("Project"), len("Task")
	for _, t := range sorted {
		projWidth = max(projWidth, len([]rune(truncate(projectName(t), 24))))
		descWidth = max(descWidth, len([]rune(truncate(t.Description, 55))))
	}

	fmt.Fprintf(file, "%s  %s  %s  %s\n",
		subtleStyle.Render(fmt.Sprintf("%-*s", projWidth, "Project")),
		subtleStyle.Render(fmt.Sprintf("%-*s", descWidth, "Task")),
		subtleStyle.Render(fmt.Sprintf("%7s", "Time")),
		subtleStyle.Render("Sessions"))
	for _, t := range sorted {
		proj := projectName(t)
		projCol := lipgloss.NewStyle().Foreground(colorFor(proj)).Render(fmt.Sprintf("%-*s", projWidth, truncate(proj, 24)))
		fmt.Fprintf(file, "%s  %-*s  %s  %8d\n",
			projCol,
			descWidth, truncate(t.Description, 55),
			fmt.Sprintf("%7s", formatDuration(t.TotalTime)),
			t.Sessions)
	}
	fmt.Fprintln(file)
}

func writeColorCategories(file *os.File, tasks []model.TaskSummary) {
	categorized := groupTasksByCategory(tasks)
	writeColorCategoryTable(file, "Dev", categorized[categoryDev])
	writeColorCategoryTable(file, "Meetings", categorized[categoryMeetings])
	writeColorCategoryTable(file, "Knowledge", categorized[categoryKnowledge])
	writeColorCategoryTable(file, "Misc", categorized[categoryMisc])
}

// WeekReportANSI renders a week report as styled terminal output.
func WeekReportANSI(file *os.File, week model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintln(file, titleStyle.Render(fmt.Sprintf("Week %d", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))))
	fmt.Fprintln(file, subtleStyle.Render(fmt.Sprintf("%s → %s",
		week.Start.Format("Mon, Jan 2"), week.End.Format("Mon, Jan 2"))))
	fmt.Fprintf(file, "%s %s\n\n", subtleStyle.Render("Total:"), totalStyle.Render(formatDuration(week.Total)))

	writeColorWeekdayChart(file, week)

	if len(week.ByProject) > 0 {
		writeColorShareChart(file, "Projects", week.ByProject, week.Total)
	}
	if len(week.ByTag) > 0 {
		writeColorShareChart(file, "Categories", week.ByTag, week.Total)
	}

	if len(week.Tasks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found for this week."))
		return
	}

	writeColorCategories(file, week.Tasks)
}

// writeColorWeekTrend renders a week-over-week chart: vertical columns when
// they fit, otherwise colored horizontal bars (e.g. a full-year range).
func writeColorWeekTrend(file *os.File, weeks []model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	if len(weeks) < 2 {
		return
	}

	var max float64
	for _, w := range weeks {
		if w.Total > max {
			max = w.Total
		}
	}
	if max <= 0 {
		return
	}

	columns := make([]chartColumn, len(weeks))
	for i, w := range weeks {
		columns[i] = chartColumn{
			top:    fmt.Sprintf("W%d", birthdayWeekNumber(w.Start, birthdayMonth, birthdayDay)),
			bottom: compactDuration(w.Total),
			ratio:  w.Total / max,
		}
	}

	if verticalChartWidth(columns) <= maxVerticalWidth {
		writeColorVerticalChart(file, "Weekly Trend", columns, formatDuration(max))
		return
	}

	labels := make([]string, len(weeks))
	labelWidth := 0
	for i, w := range weeks {
		labels[i] = fmt.Sprintf("W%d %s", birthdayWeekNumber(w.Start, birthdayMonth, birthdayDay), w.Start.Format("Jan 2"))
		if n := len([]rune(labels[i])); n > labelWidth {
			labelWidth = n
		}
	}

	fmt.Fprintln(file, headerStyle.Render("Weekly Trend"))
	for i, w := range weeks {
		fmt.Fprintf(file, "%s  %s  %s\n",
			subtleStyle.Render(fmt.Sprintf("%-*s", labelWidth, labels[i])),
			renderColorBar(w.Total/max, accentColor),
			fmt.Sprintf("%7s", formatDuration(w.Total)))
	}
	fmt.Fprintln(file)
}

// writeColorWeekSection renders one week as a sub-section inside a month or
// range report.
func writeColorWeekSection(file *os.File, week model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintln(file, headerStyle.Render(fmt.Sprintf("Week %d", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))))
	fmt.Fprintln(file, subtleStyle.Render(fmt.Sprintf("%s → %s  ·  %s",
		week.Start.Format("Mon, Jan 2"), week.End.Format("Mon, Jan 2"), formatDuration(week.Total))))
	fmt.Fprintln(file)

	writeColorWeekdayChart(file, week)

	if len(week.ByProject) > 0 {
		writeColorShareChart(file, "Projects", week.ByProject, week.Total)
	}
	if len(week.ByTag) > 0 {
		writeColorShareChart(file, "Categories", week.ByTag, week.Total)
	}
	if len(week.Tasks) > 0 {
		writeColorCategories(file, week.Tasks)
	}
}

// aggregateWeeks rolls per-week tag/project maps up to a parent total.
func aggregateWeeks(weeks []model.WeekData) (tags, projects map[string]float64) {
	tags = make(map[string]float64)
	projects = make(map[string]float64)
	for _, week := range weeks {
		for tag, hours := range week.ByTag {
			tags[tag] += hours
		}
		for project, hours := range week.ByProject {
			projects[project] += hours
		}
	}
	return tags, projects
}

// MonthReportANSI renders a month report as styled terminal output.
func MonthReportANSI(file *os.File, month model.MonthData, year int, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintln(file, titleStyle.Render(fmt.Sprintf("%s %d", month.Month.String(), year)))
	fmt.Fprintf(file, "%s %s\n\n", subtleStyle.Render("Total:"), totalStyle.Render(formatDuration(month.Total)))

	tags, projects := aggregateWeeks(month.Weeks)

	if len(month.Weeks) > 0 {
		writeColorWeekTrend(file, month.Weeks, birthdayMonth, birthdayDay)
	}
	if len(projects) > 0 {
		writeColorShareChart(file, "Projects", projects, month.Total)
	}
	if len(tags) > 0 {
		writeColorShareChart(file, "Categories", tags, month.Total)
	}

	if len(month.Weeks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found for this month."))
		return
	}

	for _, week := range month.Weeks {
		writeColorWeekSection(file, week, birthdayMonth, birthdayDay)
	}
}

// RangeReportANSI renders a custom date-range report as styled terminal output.
func RangeReportANSI(file *os.File, report model.MonthData, start, end time.Time, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintln(file, titleStyle.Render(fmt.Sprintf("%s → %s",
		start.Format("Jan 2, 2006"), end.AddDate(0, 0, -1).Format("Jan 2, 2006"))))
	fmt.Fprintf(file, "%s %s\n\n", subtleStyle.Render("Total:"), totalStyle.Render(formatDuration(report.Total)))

	tags, projects := aggregateWeeks(report.Weeks)

	if len(report.Weeks) > 0 {
		writeColorWeekTrend(file, report.Weeks, birthdayMonth, birthdayDay)
	}
	if len(projects) > 0 {
		writeColorShareChart(file, "Projects", projects, report.Total)
	}
	if len(tags) > 0 {
		writeColorShareChart(file, "Categories", tags, report.Total)
	}

	if len(report.Weeks) == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found for this range."))
		return
	}

	for _, week := range report.Weeks {
		writeColorWeekSection(file, week, birthdayMonth, birthdayDay)
	}
}

// DayReportANSI renders a single-day report as styled terminal output.
func DayReportANSI(file *os.File, report model.DayReport, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintln(file, titleStyle.Render(fmt.Sprintf("Day %d", birthdayDayNumber(report.Date, birthdayMonth, birthdayDay))))
	fmt.Fprintln(file, subtleStyle.Render(report.Date.Format("Monday, Jan 2, 2006")))
	fmt.Fprintf(file, "%s %s\n\n", subtleStyle.Render("Total:"), totalStyle.Render(formatDuration(report.Total)))

	if len(report.ByProject) > 0 {
		writeColorShareChart(file, "Projects", report.ByProject, report.Total)
	}

	if report.Total == 0 {
		fmt.Fprintln(file, emptyStyle.Render("No entries found for this day."))
	}
}
