# Lume

A CLI tool that generates beautiful time reports from your [Timewarrior](https://timewarrior.net/) entries.

## Installation

```bash
go build -o lume .
```

## Usage

```bash
./lume [flags]
```

### Flags

| Flag | Short | Default | Description |
|:-----|:-----:|:--------|:------------|
| `--output` | `-o` | `~/wiki/report` | Output directory for reports |
| `--timewarrior` | `-t` | `~/.config/timewarrior/data` | Timewarrior data directory |
| `--year` | `-y` | Current year | Year to generate report for |

### Commands

- `generate` - Build markdown reports for the configured year.
- `report day` - Show the current day report (use `-t YYYY-MM-DD` to override).
- `report week` - Show the current week report (use `-t YYYY-MM-DD` to override).
- `report month` - Show the current month report (use `-t YYYY-MM` to override).
- `report range` - Show a range report (use `--from` and `--to`).
- `config` - Interactive config for default paths.

### Examples

```bash
# Generate report for current year
./lume generate

# Generate report for 2024
./lume generate --year 2024

# Custom output directory
./lume generate -o ~/documents/reports

# Full custom
./lume generate -t ~/.timewarrior/data -o ~/reports -y 2024

# Day report (today)
./lume report day

# Day report for a specific date
./lume report day -t 2026-01-18

# Week report (current week)
./lume report week

# Month report (current month)
./lume report month

# Range report
./lume report range --from 2026-01-01 --to 2026-01-31

# Configure defaults interactively
./lume config
```

### Zsh Autocomplete

```zsh
./lume completion zsh > "${fpath[1]}/_lume"
```

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
