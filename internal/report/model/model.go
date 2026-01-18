package model

import "time"

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

type DayReport struct {
	Date  time.Time
	Tasks []TaskSummary
	ByTag map[string]float64
	Total float64
}
