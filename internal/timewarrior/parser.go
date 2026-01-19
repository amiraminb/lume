package timewarrior

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Entry struct {
	Start       time.Time
	End         time.Time
	Description string
	Tags        []string
}

func (e Entry) Duration() time.Duration {
	return e.End.Sub(e.Start)
}

var entryRegex = regexp.MustCompile(`^inc (\d{8}T\d{6}Z) - (\d{8}T\d{6}Z) # (.*)$`)

func ParseDataDir(dataDir string) ([]Entry, error) {
	files, err := filepath.Glob(filepath.Join(dataDir, "*.data"))
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, file := range files {
		if strings.Contains(file, "undo.data") || strings.Contains(file, "tags.data") {
			continue
		}
		fileEntries, err := parseFile(file)
		if err != nil {
			return nil, err
		}
		entries = append(entries, fileEntries...)
	}
	return entries, nil
}

func parseFile(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if entry, ok := parseLine(line); ok {
			entries = append(entries, entry)
		}
	}
	return entries, scanner.Err()
}

func parseLine(line string) (Entry, bool) {
	matches := entryRegex.FindStringSubmatch(line)
	if matches == nil {
		return Entry{}, false
	}

	start, err := time.Parse("20060102T150405Z", matches[1])
	if err != nil {
		return Entry{}, false
	}

	end, err := time.Parse("20060102T150405Z", matches[2])
	if err != nil {
		return Entry{}, false
	}

	start = start.In(time.Local)
	end = end.In(time.Local)

	desc, tags := parseAnnotation(matches[3])

	return Entry{
		Start:       start,
		End:         end,
		Description: desc,
		Tags:        tags,
	}, true
}

func parseAnnotation(annotation string) (string, []string) {
	var description string
	var tags []string

	parts := splitAnnotation(annotation)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "\"desc:") {
			description = strings.TrimSuffix(strings.TrimPrefix(part, "\"desc:"), "\"")
		} else if strings.HasPrefix(part, "desc:") {
			description = part[5:]
		} else if part != "" && !strings.HasPrefix(part, "#") {
			tags = append(tags, part)
		}
	}

	return description, tags
}

func splitAnnotation(annotation string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, r := range annotation {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ' ':
			if inQuotes {
				current.WriteRune(r)
			} else {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
