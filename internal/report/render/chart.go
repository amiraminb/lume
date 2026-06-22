package render

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/amiraminb/lume/internal/report/model"
)

const barWidth = 24

// chartHeight is the number of text rows a vertical chart's plot area spans.
// Combined with eighth-block resolution it yields chartHeight*8 distinct levels.
const chartHeight = 8

// maxVerticalWidth caps how wide a vertical chart may grow before it would wrap
// in a typical terminal; past this the caller falls back to horizontal bars.
const maxVerticalWidth = 76

// partialVBlocks indexes vertical partial-block runes by eighths (1–7); index 0
// is a blank cell. A full cell uses fullBlock.
var partialVBlocks = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇'}

// eighthBlocks indexes the sub-cell partial block runes by eighths (0–7); the
// full cell uses fullBlock. Together they give bars ~8x finer resolution than
// whole characters, so a 53% and a 57% bar look visibly different.
var eighthBlocks = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}

const (
	fullBlock  = '█'
	emptyBlock = '░'
)

// renderBar draws a single proportional bar of fixed width using eighth-block
// resolution. ratio is clamped to [0,1].
func renderBar(ratio float64) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	totalEighths := int(math.Round(ratio * float64(barWidth) * 8))
	full := totalEighths / 8
	rem := totalEighths % 8

	bar := make([]rune, 0, barWidth)
	for range full {
		bar = append(bar, fullBlock)
	}
	if rem > 0 && full < barWidth {
		bar = append(bar, eighthBlocks[rem])
	}
	for len(bar) < barWidth {
		bar = append(bar, emptyBlock)
	}

	return string(bar)
}

type chartRow struct {
	label string
	hours float64
}

type chartColumn struct {
	top    string // label printed above the column (e.g. "Sun")
	bottom string // value printed below the column (e.g. "2h42m"), "" when empty
	ratio  float64
}

// verticalChartWidth reports the rendered width (in cells) of a vertical chart
// with the given columns, so callers can decide whether it fits before drawing.
func verticalChartWidth(columns []chartColumn) int {
	colWidth := 3
	for _, c := range columns {
		colWidth = max(colWidth, len([]rune(c.top)), len([]rune(c.bottom)))
	}
	const axisWidth = 4 // peak label + "┤"
	return axisWidth + len(columns)*(colWidth+1)
}

// writeVerticalChart draws a column chart inside a fenced code block: each
// column rises from a baseline using vertical eighth-blocks for sub-row
// precision, with a y-axis showing the peak value and per-column labels above
// and values below. peakLabel annotates the top gridline.
func writeVerticalChart(file *os.File, title string, columns []chartColumn, peakLabel string) {
	colWidth := 3
	for _, c := range columns {
		colWidth = max(colWidth, len([]rune(c.top)), len([]rune(c.bottom)))
	}

	// Each column's height in eighths across the whole plot area.
	levels := make([]int, len(columns))
	for i, c := range columns {
		r := c.ratio
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		levels[i] = int(math.Round(r * float64(chartHeight) * 8))
	}

	axisPad := len([]rune(peakLabel))
	cell := func(filled bool, s string) string {
		if filled {
			return fmt.Sprintf("%-*s", colWidth+1, s)
		}
		return strings.Repeat(" ", colWidth+1)
	}

	fmt.Fprintf(file, "**%s**\n\n", title)
	fmt.Fprintf(file, "```\n")

	for row := chartHeight - 1; row >= 0; row-- {
		// y-axis gutter: peak label on the top row, blanks elsewhere.
		if row == chartHeight-1 {
			fmt.Fprintf(file, "%s ┤", peakLabel)
		} else {
			fmt.Fprintf(file, "%s │", strings.Repeat(" ", axisPad))
		}

		lo := row * 8
		for _, eighths := range levels {
			switch {
			case eighths >= lo+8:
				fmt.Fprint(file, cell(true, string(fullBlock)))
			case eighths > lo:
				fmt.Fprint(file, cell(true, string(partialVBlocks[eighths-lo])))
			default:
				fmt.Fprint(file, cell(false, ""))
			}
		}
		fmt.Fprint(file, "\n")
	}

	// Baseline axis.
	fmt.Fprintf(file, "%s └", strings.Repeat(" ", axisPad))
	fmt.Fprint(file, strings.Repeat("─", len(columns)*(colWidth+1)))
	fmt.Fprint(file, "\n")

	// Top labels and bottom values, aligned to the columns (after the axis gutter).
	gutter := strings.Repeat(" ", axisPad+2)
	fmt.Fprint(file, gutter)
	for _, c := range columns {
		fmt.Fprintf(file, "%-*s", colWidth+1, c.top)
	}
	fmt.Fprint(file, "\n")
	fmt.Fprint(file, gutter)
	for _, c := range columns {
		fmt.Fprintf(file, "%-*s", colWidth+1, c.bottom)
	}
	fmt.Fprint(file, "\n")

	fmt.Fprintf(file, "```\n")
}

// compactDuration formats hours like "2h42m" / "45m" / "" (for zero) so values
// fit under a narrow vertical column.
func compactDuration(hours float64) string {
	if hours <= 0 {
		return ""
	}
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

// writeShareChart renders a horizontal bar chart of labelled shares inside a
// fenced code block so glow preserves alignment. Bars are scaled to the
// largest value (not the total) so the leader fills the track and differences
// stay legible; the share percentage is taken against total.
func writeShareChart(file *os.File, title string, values map[string]float64, total float64) {
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

	fmt.Fprintf(file, "**%s**\n\n", title)
	fmt.Fprintf(file, "```\n")
	for _, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = (r.hours / total) * 100
		}
		fmt.Fprintf(file, "%-*s  %s  %7s  %3.0f%%\n",
			labelWidth, r.label,
			renderBar(r.hours/max),
			formatDuration(r.hours),
			pct)
	}
	fmt.Fprintf(file, "```\n")
}

// writeWeekdayChart renders a Sun–Sat bar chart of daily totals for a single
// week so the within-week rhythm is visible at a glance.
func writeWeekdayChart(file *os.File, week model.WeekData) {
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

	writeVerticalChart(file, "Daily Trend", columns, formatDuration(max))
}

// writeWeekTrend renders a week-over-week bar chart for a month/range report so
// the shape of effort over time is visible at a glance. Weeks are assumed
// chronological (build.groupByWeek sorts them).
func writeWeekTrend(file *os.File, weeks []model.WeekData, birthdayMonth time.Month, birthdayDay int) {
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
		writeVerticalChart(file, "Weekly Trend", columns, formatDuration(max))
		return
	}

	// Too many weeks to fit vertically (e.g. a full-year range); fall back to
	// horizontal bars where long label lists wrap gracefully.
	labels := make([]string, len(weeks))
	labelWidth := 0
	for i, w := range weeks {
		labels[i] = fmt.Sprintf("W%d %s", birthdayWeekNumber(w.Start, birthdayMonth, birthdayDay), w.Start.Format("Jan 2"))
		if n := len([]rune(labels[i])); n > labelWidth {
			labelWidth = n
		}
	}

	fmt.Fprintf(file, "**Weekly Trend**\n\n")
	fmt.Fprintf(file, "```\n")
	for i, w := range weeks {
		fmt.Fprintf(file, "%-*s  %s  %7s\n",
			labelWidth, labels[i],
			renderBar(w.Total/max),
			formatDuration(w.Total))
	}
	fmt.Fprintf(file, "```\n")
}
