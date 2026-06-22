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
var projectPalette = []lipgloss.Color{
	lipgloss.Color("39"),  // blue
	lipgloss.Color("208"), // orange
	lipgloss.Color("42"),  // green
	lipgloss.Color("170"), // magenta
	lipgloss.Color("220"), // yellow
	lipgloss.Color("45"),  // cyan
	lipgloss.Color("203"), // red
	lipgloss.Color("141"), // purple
	lipgloss.Color("114"), // light green
	lipgloss.Color("215"), // peach
	lipgloss.Color("75"),  // sky
	lipgloss.Color("180"), // tan
}

// colorFor returns a stable palette color for a label via FNV hashing.
func colorFor(label string) lipgloss.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(label))
	return projectPalette[h.Sum32()%uint32(len(projectPalette))]
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	totalStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	emptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
