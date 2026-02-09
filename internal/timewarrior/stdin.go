package timewarrior

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"
)

type TimewConfig struct {
	Values map[string]string
}

func (c TimewConfig) Get(key string) string {
	return c.Values[key]
}

func (c TimewConfig) ReportStart() (time.Time, bool) {
	return c.parseTimestamp("temp.report.start")
}

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

type timewInterval struct {
	Start string   `json:"start"`
	End   string   `json:"end"`
	Tags  []string `json:"tags"`
}

func ParseStdin(r io.Reader) (TimewConfig, []Entry, error) {
	cfg := TimewConfig{Values: make(map[string]string)}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

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
