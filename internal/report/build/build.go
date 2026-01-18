package build

import (
	"sort"
	"time"

	"lume/internal/report/model"
	"lume/internal/timewarrior"
)

func YearReport(entries []timewarrior.Entry, year int) model.YearReport {
	filtered := filterByYear(entries, year)
	byMonth := groupByMonth(filtered)

	var months []model.MonthData
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

		months = append(months, model.MonthData{
			Month: month,
			Weeks: weeks,
			Total: monthTotal,
		})
	}

	return model.YearReport{
		Year:   year,
		Months: months,
		Total:  yearTotal,
	}
}

func WeekReport(entries []timewarrior.Entry, date time.Time) model.WeekData {
	start := weekStart(date)
	end := start.AddDate(0, 0, 7)

	var weekEntries []timewarrior.Entry
	for _, e := range entries {
		if !e.Start.Before(start) && e.Start.Before(end) {
			weekEntries = append(weekEntries, e)
		}
	}

	tasks := aggregateByDescription(weekEntries)
	byTag := aggregateByTag(weekEntries)
	weekStartDate, weekEndDate := weekBounds(start)

	var total float64
	for _, e := range weekEntries {
		total += e.Duration().Hours()
	}

	return model.WeekData{
		WeekNum: weekNumber(start),
		Start:   weekStartDate,
		End:     weekEndDate,
		Tasks:   tasks,
		ByTag:   byTag,
		Total:   total,
	}
}

func MonthReport(entries []timewarrior.Entry, month time.Month, year int) model.MonthData {
	var monthEntries []timewarrior.Entry
	for _, e := range entries {
		if e.Start.Year() == year && e.Start.Month() == month {
			monthEntries = append(monthEntries, e)
		}
	}

	weeks := groupByWeek(monthEntries)
	var total float64
	for _, w := range weeks {
		total += w.Total
	}

	return model.MonthData{
		Month: month,
		Weeks: weeks,
		Total: total,
	}
}

func DayReport(entries []timewarrior.Entry, date time.Time) model.DayReport {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	var dayEntries []timewarrior.Entry
	for _, e := range entries {
		if !e.Start.Before(start) && e.Start.Before(end) {
			dayEntries = append(dayEntries, e)
		}
	}

	tasks := aggregateByDescription(dayEntries)
	byTag := aggregateByTag(dayEntries)
	var total float64
	for _, e := range dayEntries {
		total += e.Duration().Hours()
	}

	return model.DayReport{
		Date:  start,
		Tasks: tasks,
		ByTag: byTag,
		Total: total,
	}
}

func RangeReport(entries []timewarrior.Entry, start time.Time, end time.Time) model.MonthData {
	var rangeEntries []timewarrior.Entry
	for _, e := range entries {
		if !e.Start.Before(start) && e.Start.Before(end) {
			rangeEntries = append(rangeEntries, e)
		}
	}

	weeks := groupByWeek(rangeEntries)
	var total float64
	for _, w := range weeks {
		total += w.Total
	}

	return model.MonthData{
		Weeks: weeks,
		Total: total,
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

func groupByWeek(entries []timewarrior.Entry) []model.WeekData {
	weekMap := make(map[time.Time][]timewarrior.Entry)

	for _, e := range entries {
		start := weekStart(e.Start)
		weekMap[start] = append(weekMap[start], e)
	}

	var weeks []model.WeekData
	for weekStartDate, weekEntries := range weekMap {
		tasks := aggregateByDescription(weekEntries)
		byTag := aggregateByTag(weekEntries)
		start, end := weekBounds(weekStartDate)

		var total float64
		for _, e := range weekEntries {
			total += e.Duration().Hours()
		}

		weeks = append(weeks, model.WeekData{
			WeekNum: weekNumber(weekStartDate),
			Start:   start,
			End:     end,
			Tasks:   tasks,
			ByTag:   byTag,
			Total:   total,
		})
	}

	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].Start.Before(weeks[j].Start)
	})

	return weeks
}

func aggregateByDescription(entries []timewarrior.Entry) []model.TaskSummary {
	taskMap := make(map[string]*model.TaskSummary)

	for _, e := range entries {
		desc := e.Description
		if desc == "" {
			desc = "(no description)"
		}

		if _, exists := taskMap[desc]; !exists {
			taskMap[desc] = &model.TaskSummary{
				Description: desc,
				Tags:        make(map[string]bool),
				DayTotals:   make(map[time.Weekday]float64),
			}
		}
		taskMap[desc].TotalTime += e.Duration().Hours()
		taskMap[desc].Sessions++
		weekday := e.Start.Weekday()
		taskMap[desc].DayTotals[weekday] += e.Duration().Hours()
		for _, tag := range e.Tags {
			taskMap[desc].Tags[tag] = true
		}
	}

	var tasks []model.TaskSummary
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

func weekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start.AddDate(0, 0, -weekday)
}

func weekBounds(start time.Time) (time.Time, time.Time) {
	end := start.AddDate(0, 0, 6).Add(time.Hour*23 + time.Minute*59 + time.Second*59)
	return start, end
}

func weekNumber(start time.Time) int {
	startOfYear := weekStart(time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, start.Location()))
	weeks := int(start.Sub(startOfYear).Hours() / 24 / 7)
	return weeks + 1
}
