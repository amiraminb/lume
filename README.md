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

### Examples

```bash
# Generate report for current year
./lume

# Generate report for 2024
./lume --year 2024

# Custom output directory
./lume -o ~/documents/reports

# Full custom
./lume -t ~/.timewarrior/data -o ~/reports -y 2024
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
