package timewarrior

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"
)

// TimewConfig holds key-value pairs from the timewarrior config block.
type TimewConfig struct {
	Values map[string]string
}

// Get returns the value for a config key, or empty string if not found.
func (c TimewConfig) Get(key string) string {
	return c.Values[key]
}

// ReportStart parses the temp.report.start timestamp.
func (c TimewConfig) ReportStart() (time.Time, bool) {
	return c.parseTimestamp("temp.report.start")
}

// ReportEnd parses the temp.report.end timestamp.
func (c TimewConfig) ReportEnd() (time.Time, bool) {
	return c.parseTimestamp("temp.report.end")
}

func (c TimewConfig) parseTimestamp(key string) (time.Time, bool) {
	v := c.Values[key]
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("20060102T150405Z", v)
	if err != nil {
		return time.Time{}, false
	}
	return t.In(time.Local), true
}

// HasTag checks whether a tag is present in temp.report.tags.
func (c TimewConfig) HasTag(tag string) bool {
	v := c.Values["temp.report.tags"]
	if v == "" {
		return false
	}
	for _, t := range strings.Split(v, ",") {
		if strings.TrimSpace(t) == tag {
			return true
		}
	}
	return v == tag
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// timewInterval represents one JSON interval from timewarrior stdin.
type timewInterval struct {
	Start string   `json:"start"`
	End   string   `json:"end"`
	Tags  []string `json:"tags"`
}

// ParseStdin reads the timewarrior extension stdin protocol:
// config lines (key: value), blank line, then JSON array of intervals.
func ParseStdin(r io.Reader) (TimewConfig, []Entry, error) {
	cfg := TimewConfig{Values: make(map[string]string)}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	// Read config lines until blank line.
	var jsonLines strings.Builder
	inConfig := true
	for scanner.Scan() {
		line := scanner.Text()
		if inConfig {
			if line == "" {
				inConfig = false
				continue
			}
			if idx := strings.Index(line, ": "); idx >= 0 {
				cfg.Values[line[:idx]] = line[idx+2:]
			}
		} else {
			jsonLines.WriteString(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, nil, err
	}

	// Parse JSON intervals.
	jsonData := strings.TrimSpace(jsonLines.String())
	if jsonData == "" || jsonData == "[]" {
		return cfg, nil, nil
	}

	var intervals []timewInterval
	if err := json.Unmarshal([]byte(jsonData), &intervals); err != nil {
		return cfg, nil, err
	}

	entries := make([]Entry, 0, len(intervals))
	for _, iv := range intervals {
		start, err := time.Parse("20060102T150405Z", iv.Start)
		if err != nil {
			continue
		}
		start = start.In(time.Local)

		var end time.Time
		if iv.End == "" {
			end = time.Now().In(time.Local)
		} else {
			parsed, err := time.Parse("20060102T150405Z", iv.End)
			if err != nil {
				continue
			}
			end = parsed.In(time.Local)
		}

		desc, tags := parseTimewTags(iv.Tags)
		entries = append(entries, Entry{
			Start:       start,
			End:         end,
			Description: desc,
			Tags:        tags,
		})
	}

	return cfg, entries, nil
}

// parseTimewTags extracts description and tags from a timewarrior tags array.
// Tags with "desc:" prefix become the description; the rest are tags.
func parseTimewTags(raw []string) (string, []string) {
	var description string
	var tags []string

	for _, t := range raw {
		if strings.HasPrefix(t, "desc:") {
			description = t[5:]
		} else {
			tags = append(tags, t)
		}
	}

	return description, tags
}
