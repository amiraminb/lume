# Lume

A [Timewarrior](https://timewarrior.net/) extension that generates beautiful time reports from your tracked entries.

## Installation

### Install script (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/amiraminb/lume/main/install.sh | bash
```

Downloads the latest release for your OS/arch and places the binary at `~/.config/timewarrior/extensions/lume`.

### go install

```bash
go install github.com/amiraminb/lume@latest
ln -sf "$(go env GOBIN)/lume" ~/.config/timewarrior/extensions/lume
```

(If `GOBIN` is unset, the binary lands in `$(go env GOPATH)/bin`.)

### From source

```bash
git clone https://github.com/amiraminb/lume
cd lume
make install
```

Add to `~/.config/timewarrior/timewarrior.cfg`:

```
reports.lume.output = ~/wiki/report
reports.lume.birthday = 04-14
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
reports.lume.birthday = 04-14
```

No separate config file is needed.

- `reports.lume.birthday` is optional and accepts `MM-DD` or `YYYY-MM-DD`.
- Default birthday is `04-14` if not set.

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
