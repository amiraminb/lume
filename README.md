# Lume

A [Timewarrior](https://timewarrior.net/) extension that generates beautiful time reports from your tracked entries.

## Installation

```bash
go build -o lume .
ln -s "$(pwd)/lume" ~/.config/timewarrior/extensions/lume
```

Add to `~/.config/timewarrior/timewarrior.cfg`:

```
reports.lume.output = ~/wiki/report
```

## Usage

```bash
timew lume :day                           # Today
timew lume :week                          # This week
timew lume :month                         # This month
timew lume 2025-01-15 - 2025-01-16        # Specific day
timew lume 2025-01 - 2025-02              # Specific month
timew lume 2025-01-01 - 2025-06-01        # Custom range
timew lume generate                       # Write markdown files to disk
```

Report type is auto-detected from the date span:

| Span | Report |
|:-----|:-------|
| 1 day | Day report with task breakdown by category |
| 2–7 days | Week report with per-day columns |
| 8–31 days | Month report with weekly sections |
| 32+ days | Range report with weekly sections |

### Configuration

The output directory for `generate` is configured in timewarrior's own config file (`~/.config/timewarrior/timewarrior.cfg`):

```
reports.lume.output = ~/wiki/report
```

No separate config file is needed.

## Output Structure

```
report/
└── 2025/
    ├── index.md
    ├── 01-january.md
    ├── 02-february.md
    └── ...
```

Each month file contains:
- Monthly overview with tag breakdown
- Weekly sections with task tables
- Time tracked per task with session counts

## Requirements

- Go 1.21+
- Timewarrior with existing time entries
