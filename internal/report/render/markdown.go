package render

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/amiraminb/lume/internal/report/model"
)

func YearIndex(file *os.File, report model.YearReport) {
	fmt.Fprintf(file, "# Time Report %d\n\n", report.Year)
	fmt.Fprintf(file, "> **Total Tracked:** %s\n\n", formatDuration(report.Total))

	yearTags := make(map[string]float64)
	yearProjects := make(map[string]float64)
	for _, month := range report.Months {
		for _, week := range month.Weeks {
			for tag, hours := range week.ByTag {
				yearTags[tag] += hours
			}
			for project, hours := range week.ByProject {
				yearProjects[project] += hours
			}
		}
	}

	if len(yearProjects) > 0 {
		writeProjectSummary(file, yearProjects, report.Total)
		fmt.Fprintf(file, "\n")
	}

	if len(yearTags) > 0 {
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

func MonthFile(file *os.File, month model.MonthData, year int, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "# %s %d\n\n", month.Month.String(), year)
	fmt.Fprintf(file, "> **Monthly Total:** %s\n\n", formatDuration(month.Total))
	fmt.Fprintf(file, "---\n\n")

	monthTags := make(map[string]float64)
	monthProjects := make(map[string]float64)
	for _, week := range month.Weeks {
		for tag, hours := range week.ByTag {
			monthTags[tag] += hours
		}
		for project, hours := range week.ByProject {
			monthProjects[project] += hours
		}
	}

	if len(monthProjects) > 0 {
		writeProjectSummary(file, monthProjects, month.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(monthTags) > 0 {
		writeTagSummary(file, monthTags, month.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	for _, week := range month.Weeks {
		WeekSection(file, week, birthdayMonth, birthdayDay)
	}
}

func DayReport(file *os.File, report model.DayReport, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "# Day %d\n", birthdayDayNumber(report.Date, birthdayMonth, birthdayDay))
	fmt.Fprintf(file, "> %s\n\n", report.Date.Format("Monday, Jan 2, 2006"))
	fmt.Fprintf(file, "> **Daily Total:** %s\n\n", formatDuration(report.Total))

	if len(report.ByProject) > 0 {
		writeProjectSummary(file, report.ByProject, report.Total)
		fmt.Fprintf(file, "\n")
	}

	if len(report.ByTag) > 0 {
		writeTagSummary(file, report.ByTag, report.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(report.Tasks) == 0 {
		fmt.Fprintf(file, "No entries found for this day.\n")
		return
	}

	categorized := groupTasksByCategory(report.Tasks)
	writeCategoryTable(file, "Dev", categorized[categoryDev])
	writeCategoryTable(file, "Meetings", categorized[categoryMeetings])
	writeCategoryTable(file, "Knowledge", categorized[categoryKnowledge])
	writeCategoryTable(file, "Misc", categorized[categoryMisc])
}

func WeekReport(file *os.File, week model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "# Week %d\n", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))
	fmt.Fprintf(file, "> %s → %s\n\n",
		week.Start.Format("Mon, Jan 2"),
		week.End.Format("Mon, Jan 2"))

	fmt.Fprintf(file, "**Total:** %s\n\n", formatDuration(week.Total))

	writeDailyTotals(file, week)

	if len(week.ByProject) > 0 {
		writeProjectSummary(file, week.ByProject, week.Total)
		fmt.Fprintf(file, "\n")
	}

	if len(week.ByTag) > 0 {
		writeTagSummary(file, week.ByTag, week.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(week.Tasks) == 0 {
		fmt.Fprintf(file, "No entries found for this week.\n")
		return
	}

	categorized := groupTasksByCategory(week.Tasks)
	writeCategoryWeekTable(file, "Dev", categorized[categoryDev])
	writeCategoryWeekTable(file, "Meetings", categorized[categoryMeetings])
	writeCategoryWeekTable(file, "Knowledge", categorized[categoryKnowledge])
	writeCategoryWeekTable(file, "Misc", categorized[categoryMisc])
}

func MonthReport(file *os.File, month model.MonthData, year int, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "# %s %d\n\n", month.Month.String(), year)
	fmt.Fprintf(file, "> **Monthly Total:** %s\n\n", formatDuration(month.Total))
	fmt.Fprintf(file, "---\n\n")

	monthTags := make(map[string]float64)
	monthProjects := make(map[string]float64)
	for _, week := range month.Weeks {
		for tag, hours := range week.ByTag {
			monthTags[tag] += hours
		}
		for project, hours := range week.ByProject {
			monthProjects[project] += hours
		}
	}

	if len(monthProjects) > 0 {
		writeProjectSummary(file, monthProjects, month.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(monthTags) > 0 {
		writeTagSummary(file, monthTags, month.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(month.Weeks) == 0 {
		fmt.Fprintf(file, "No entries found for this month.\n")
		return
	}

	for _, week := range month.Weeks {
		WeekSection(file, week, birthdayMonth, birthdayDay)
	}
}

func RangeReport(file *os.File, report model.MonthData, start time.Time, end time.Time, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "# %s → %s\n\n", start.Format("Jan 2, 2006"), end.AddDate(0, 0, -1).Format("Jan 2, 2006"))
	fmt.Fprintf(file, "> **Range Total:** %s\n\n", formatDuration(report.Total))
	fmt.Fprintf(file, "---\n\n")

	rangeTags := make(map[string]float64)
	rangeProjects := make(map[string]float64)
	for _, week := range report.Weeks {
		for tag, hours := range week.ByTag {
			rangeTags[tag] += hours
		}
		for project, hours := range week.ByProject {
			rangeProjects[project] += hours
		}
	}

	if len(rangeProjects) > 0 {
		writeProjectSummary(file, rangeProjects, report.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(rangeTags) > 0 {
		writeTagSummary(file, rangeTags, report.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(report.Weeks) == 0 {
		fmt.Fprintf(file, "No entries found for this range.\n")
		return
	}

	for _, week := range report.Weeks {
		WeekSection(file, week, birthdayMonth, birthdayDay)
	}
}

func WeekSection(file *os.File, week model.WeekData, birthdayMonth time.Month, birthdayDay int) {
	fmt.Fprintf(file, "## Week %d\n", birthdayWeekNumber(week.Start, birthdayMonth, birthdayDay))
	fmt.Fprintf(file, "> %s → %s\n\n",
		week.Start.Format("Mon, Jan 2"),
		week.End.Format("Mon, Jan 2"))

	fmt.Fprintf(file, "**Total:** %s\n\n", formatDuration(week.Total))

	writeDailyTotals(file, week)

	if len(week.ByProject) > 0 {
		writeProjectSummary(file, week.ByProject, week.Total)
		fmt.Fprintf(file, "\n")
	}

	if len(week.ByTag) > 0 {
		writeTagSummary(file, week.ByTag, week.Total)
		fmt.Fprintf(file, "\n---\n\n")
	}

	if len(week.Tasks) > 0 {
		categorized := groupTasksByCategory(week.Tasks)
		writeCategoryTable(file, "Dev", categorized[categoryDev])
		writeCategoryTable(file, "Meetings", categorized[categoryMeetings])
		writeCategoryTable(file, "Knowledge", categorized[categoryKnowledge])
		writeCategoryTable(file, "Misc", categorized[categoryMisc])
	}

	fmt.Fprintf(file, "---\n\n")
}

func writeDailyTotals(file *os.File, week model.WeekData) {
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

	fmt.Fprintf(file, "**Daily Totals:**\n\n")

	fmt.Fprintf(file, "|")
	for i, day := range days {
		date := week.Start.AddDate(0, 0, i)
		fmt.Fprintf(file, " %s %s |", day.String()[:3], date.Format("Jan 2"))
	}
	fmt.Fprintf(file, "\n|")
	for range days {
		fmt.Fprintf(file, "--:|")
	}
	fmt.Fprintf(file, "\n|")

	for _, day := range days {
		hours := dayTotals[day]
		if hours <= 0 {
			fmt.Fprintf(file, " — |")
		} else {
			fmt.Fprintf(file, " %s |", formatDuration(hours))
		}
	}
	fmt.Fprintf(file, "\n\n")
}

func writeWeekTasks(file *os.File, tasks []model.TaskSummary) {
	tasksByTag := groupTasksByTag(tasks)

	var tags []string
	for tag := range tasksByTag {
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool {
		var totalI, totalJ float64
		for _, groupedTasks := range tasksByTag[tags[i]] {
			totalI += sumTaskHours(groupedTasks)
		}
		for _, groupedTasks := range tasksByTag[tags[j]] {
			totalJ += sumTaskHours(groupedTasks)
		}
		return totalI > totalJ
	})

	for _, tag := range tags {
		projectGroups := tasksByTag[tag]
		var tagTotal float64
		for _, groupedTasks := range projectGroups {
			for _, t := range groupedTasks {
				tagTotal += t.TotalTime
			}
		}

		fmt.Fprintf(file, "### %s\n", tag)
		fmt.Fprintf(file, "**Subtotal:** %s\n\n", formatDuration(tagTotal))

		projects := sortedProjects(projectGroups)
		for _, project := range projects {
			projectTasks := projectGroups[project]
			projectTotal := sumTaskHours(projectTasks)

			fmt.Fprintf(file, "#### project:%s\n", project)
			fmt.Fprintf(file, "**Project Subtotal:** %s\n\n", formatDuration(projectTotal))

			fmt.Fprintf(file, "| Task | Time | Sessions |\n")
			fmt.Fprintf(file, "|:-----|-----:|---------:|\n")

			for _, t := range projectTasks {
				fmt.Fprintf(file, "| %s | %s | %d |\n",
					truncate(t.Description, 55),
					formatDuration(t.TotalTime),
					t.Sessions)
			}
			fmt.Fprintf(file, "\n")
		}
	}
}

func groupTasksByTag(tasks []model.TaskSummary) map[string]map[string][]model.TaskSummary {
	grouped := make(map[string]map[string][]model.TaskSummary)

	for _, t := range tasks {
		tags := taskNonProjectTags(t)
		project := t.Project
		if project == "" {
			project = "unknown"
		}

		for _, tag := range tags {
			if _, ok := grouped[tag]; !ok {
				grouped[tag] = make(map[string][]model.TaskSummary)
			}
			grouped[tag][project] = append(grouped[tag][project], t)
		}
	}

	for tag := range grouped {
		for project := range grouped[tag] {
			sort.Slice(grouped[tag][project], func(i, j int) bool {
				return grouped[tag][project][i].TotalTime > grouped[tag][project][j].TotalTime
			})
		}
	}

	return grouped
}

func taskNonProjectTags(task model.TaskSummary) []string {
	var tags []string
	for tag := range task.Tags {
		if strings.HasPrefix(tag, "project:") {
			continue
		}
		tags = append(tags, tag)
	}
	if len(tags) == 0 {
		return []string{"untagged"}
	}
	sort.Strings(tags)
	return tags
}

func sortedProjects(projectGroups map[string][]model.TaskSummary) []string {
	var projects []string
	for project := range projectGroups {
		projects = append(projects, project)
	}
	sort.Slice(projects, func(i, j int) bool {
		return sumTaskHours(projectGroups[projects[i]]) > sumTaskHours(projectGroups[projects[j]])
	})
	return projects
}

func sumTaskHours(tasks []model.TaskSummary) float64 {
	var total float64
	for _, t := range tasks {
		total += t.TotalTime
	}
	return total
}

type taskCategory string

const (
	categoryDev       taskCategory = "dev"
	categoryMeetings  taskCategory = "meetings"
	categoryKnowledge taskCategory = "knowledge"
	categoryMisc      taskCategory = "misc"
)

func groupTasksByCategory(tasks []model.TaskSummary) map[taskCategory][]model.TaskSummary {
	categorized := map[taskCategory][]model.TaskSummary{
		categoryDev:       {},
		categoryMeetings:  {},
		categoryKnowledge: {},
		categoryMisc:      {},
	}

	for _, task := range tasks {
		matched := false
		for tag := range task.Tags {
			switch strings.ToLower(tag) {
			case string(categoryDev):
				categorized[categoryDev] = append(categorized[categoryDev], task)
				matched = true
			case string(categoryMeetings):
				categorized[categoryMeetings] = append(categorized[categoryMeetings], task)
				matched = true
			case string(categoryKnowledge):
				categorized[categoryKnowledge] = append(categorized[categoryKnowledge], task)
				matched = true
			case string(categoryMisc):
				categorized[categoryMisc] = append(categorized[categoryMisc], task)
				matched = true
			}
		}
		if !matched {
			categorized[categoryMisc] = append(categorized[categoryMisc], task)
		}
	}

	for category := range categorized {
		sort.Slice(categorized[category], func(i, j int) bool {
			return categorized[category][i].TotalTime > categorized[category][j].TotalTime
		})
	}

	return categorized
}

func writeCategoryTable(file *os.File, title string, tasks []model.TaskSummary) {
	fmt.Fprintf(file, "## %s\n\n", title)

	if len(tasks) == 0 {
		fmt.Fprintf(file, "No entries found.\n\n")
		return
	}

	sorted := sortTasksByProject(tasks)

	fmt.Fprintf(file, "| Project | Task | Time | Sessions |\n")
	fmt.Fprintf(file, "|:--------|:-----|-----:|---------:|\n")
	for _, t := range sorted {
		fmt.Fprintf(file, "| %s | %s | %s | %d |\n",
			truncate(projectName(t), 24),
			truncate(t.Description, 55),
			formatDuration(t.TotalTime),
			t.Sessions)
	}
	fmt.Fprintf(file, "\n")
}

func writeCategoryWeekTable(file *os.File, title string, tasks []model.TaskSummary) {
	fmt.Fprintf(file, "## %s\n\n", title)

	if len(tasks) == 0 {
		fmt.Fprintf(file, "No entries found.\n\n")
		return
	}

	sorted := sortTasksByProject(tasks)

	fmt.Fprintf(file, "| Project | Task | Time | Sun | Mon | Tue | Wed | Thu | Fri | Sat |\n")
	fmt.Fprintf(file, "|:--------|:-----|-----:|----:|----:|----:|----:|----:|----:|----:|\n")
	for _, t := range sorted {
		fmt.Fprintf(file, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			truncate(projectName(t), 24),
			truncate(t.Description, 55),
			formatDuration(t.TotalTime),
			formatDayHours(t, time.Sunday),
			formatDayHours(t, time.Monday),
			formatDayHours(t, time.Tuesday),
			formatDayHours(t, time.Wednesday),
			formatDayHours(t, time.Thursday),
			formatDayHours(t, time.Friday),
			formatDayHours(t, time.Saturday))
	}
	fmt.Fprintf(file, "\n")
}

func formatDayHours(task model.TaskSummary, day time.Weekday) string {
	hours := task.DayTotals[day]
	if hours <= 0 {
		return ""
	}
	return formatDuration(hours)
}

func projectName(task model.TaskSummary) string {
	if strings.TrimSpace(task.Project) == "" {
		return "unknown"
	}
	return task.Project
}

func sortTasksByProject(tasks []model.TaskSummary) []model.TaskSummary {
	sorted := append([]model.TaskSummary(nil), tasks...)
	sort.Slice(sorted, func(i, j int) bool {
		pi := projectName(sorted[i])
		pj := projectName(sorted[j])
		if pi != pj {
			return pi < pj
		}
		if sorted[i].TotalTime != sorted[j].TotalTime {
			return sorted[i].TotalTime > sorted[j].TotalTime
		}
		return sorted[i].Description < sorted[j].Description
	})
	return sorted
}

func writeTagSummary(file *os.File, tags map[string]float64, total float64) {
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tags[tagList[i]] > tags[tagList[j]]
	})

	fmt.Fprintf(file, "| Category | Time | Share |\n")
	fmt.Fprintf(file, "|:---------|-----:|------:|\n")

	for _, tag := range tagList {
		hours := tags[tag]
		pct := 0.0
		if total > 0 {
			pct = (hours / total) * 100
		}
		fmt.Fprintf(file, "| %s | %s | %.0f%% |\n", strings.ReplaceAll(tag, "|", "\\|"), formatDuration(hours), pct)
	}
	fmt.Fprintf(file, "\n")
}

func writeProjectSummary(file *os.File, projects map[string]float64, total float64) {
	var projectList []string
	for project := range projects {
		projectList = append(projectList, project)
	}
	sort.Slice(projectList, func(i, j int) bool {
		return projects[projectList[i]] > projects[projectList[j]]
	})

	fmt.Fprintf(file, "| Project | Time | Share |\n")
	fmt.Fprintf(file, "|:--------|-----:|------:|\n")

	for _, project := range projectList {
		hours := projects[project]
		pct := 0.0
		if total > 0 {
			pct = (hours / total) * 100
		}
		fmt.Fprintf(file, "| %s | %s | %.0f%% |\n", strings.ReplaceAll(project, "|", "\\|"), formatDuration(hours), pct)
	}
	fmt.Fprintf(file, "\n")
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

func birthdayCycleStart(t time.Time, month time.Month, day int) time.Time {
	start := time.Date(t.Year(), month, day, 0, 0, 0, 0, t.Location())
	if t.Before(start) {
		start = time.Date(t.Year()-1, month, day, 0, 0, 0, 0, t.Location())
	}
	return start
}

func birthdayDayNumber(t time.Time, month time.Month, day int) int {
	start := birthdayCycleStart(t, month, day)
	days := int(t.Sub(start).Hours() / 24)
	return days + 1
}

func birthdayWeekNumber(t time.Time, month time.Month, day int) int {
	start := birthdayCycleStart(t, month, day)
	days := int(t.Sub(start).Hours() / 24)
	return (days / 7) + 1
}
