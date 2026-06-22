package render

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// lume runs as a timewarrior extension, so its stdout is always a pipe into
// timew (never a TTY). Auto-detection would therefore strip color, so we force
// a color profile explicitly. NO_COLOR (https://no-color.org) still disables it.
func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		lipgloss.SetColorProfile(termenv.Ascii)
		return
	}
	lipgloss.SetColorProfile(termenv.ANSI256)
}

const (
	colorTitle       = lipgloss.Color("231") // bright white — report title
	colorHeader      = lipgloss.Color("67")  // dim steel blue — section headers
	colorTableHeader = lipgloss.Color("130") // dim orange — table column headers
	colorTotal       = lipgloss.Color("72")  // green — totals (matches share)
	colorShare       = lipgloss.Color("72")  // green — share percentages and chart values
	colorAccent      = lipgloss.Color("39")  // blue — trend bars
	colorProject     = lipgloss.Color("253") // whitish — project/category labels, bars, "Total:" label
	colorDate        = lipgloss.Color("252") // light gray — date subtitles
	colorBorder      = lipgloss.Color("240") // gray — table borders
	colorSubtle      = lipgloss.Color("245") // gray — labels, axes
	colorEmpty       = lipgloss.Color("246") // gray — "no entries" (readable, italic)
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorTitle)
	subtleStyle  = lipgloss.NewStyle().Foreground(colorSubtle)
	dateStyle    = lipgloss.NewStyle().Foreground(colorDate)
	totalStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorTotal)
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorHeader)
	projectStyle = lipgloss.NewStyle().Foreground(colorProject)
	shareStyle   = lipgloss.NewStyle().Foreground(colorShare)
	emptyStyle   = lipgloss.NewStyle().Italic(true).Foreground(colorEmpty)
)
