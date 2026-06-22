package render

import (
	"hash/fnv"
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

// projectPalette is a fixed set of ANSI-256 colors assigned to projects by hash
// so the mapping is stable across runs and any project (added or removed over
// time) lands on a consistent color without a hardcoded name list.
//
// These colors are deliberately disjoint from the chrome colors below (titles,
// headers, totals, trend bars, borders) so a project bar can never be confused
// with the report's structural styling.
var projectPalette = []lipgloss.Color{
	lipgloss.Color("208"), // orange
	lipgloss.Color("42"),  // green
	lipgloss.Color("170"), // magenta
	lipgloss.Color("45"),  // cyan
	lipgloss.Color("203"), // red
	lipgloss.Color("141"), // purple
	lipgloss.Color("114"), // light green
	lipgloss.Color("215"), // peach
	lipgloss.Color("180"), // tan
	lipgloss.Color("105"), // indigo
	lipgloss.Color("213"), // pink
	lipgloss.Color("100"), // olive
}

// colorFor returns a stable palette color for a label via FNV hashing.
func colorFor(label string) lipgloss.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(label))
	return projectPalette[h.Sum32()%uint32(len(projectPalette))]
}

// Chrome colors: structural styling kept disjoint from projectPalette so
// headers, totals, and trend bars never share a hue with a project/category bar.
const (
	colorTitle  = lipgloss.Color("231") // bright white — report title
	colorHeader = lipgloss.Color("81")  // bright cyan — section headers
	colorTotal  = lipgloss.Color("220") // gold — totals
	colorAccent = lipgloss.Color("39")  // blue — trend bars
	colorBorder = lipgloss.Color("240") // gray — table borders
	colorSubtle = lipgloss.Color("245") // gray — labels, axes
	colorEmpty  = lipgloss.Color("240") // gray — "no entries"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorTitle)
	subtleStyle = lipgloss.NewStyle().Foreground(colorSubtle)
	totalStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorTotal)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(colorHeader)
	emptyStyle  = lipgloss.NewStyle().Foreground(colorEmpty)
)
