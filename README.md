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

## Usage

```bash
timew lume :day                           # Today
timew lume :week                          # This week
timew lume :month                         # This month
timew lume 2025-01-15 - 2025-01-16        # Specific day
timew lume 2025-01 - 2025-02              # Specific month
timew lume 2025-01-01 - 2025-06-01        # Custom range
```

Report type is auto-detected from the date span:

| Span      | Report                                     |
|:-----     |:-------                                    |
| 1 day     | Day report with task breakdown by category |
| 2–7 days  | Week report with daily trend chart         |
| 8–31 days | Month report with weekly sections          |
| 32+ days  | Range report with weekly sections          |

### Output formats

Lume renders in two formats:

- `color` (default): styled terminal output with colored bars, trend charts, and tables. Print it straight to the terminal (no pager needed).
- `markdown`: plain Markdown, suited for piping into a Markdown renderer such as [`glow`](https://github.com/charmbracelet/glow) or saving to a file.

Select the format with the `LUME_FORMAT` environment variable (per invocation) or the `reports.lume.format` config key (persistent default). The environment variable wins when both are set, and an unrecognized value falls back to `color`.

```bash
timew lume :week                          # color (default)
LUME_FORMAT=markdown timew lume :week | glow
```

Because lume runs as a timewarrior extension, its stdout is always a pipe, so color is forced on rather than auto-detected. Set `NO_COLOR` to disable it.

### Configuration

Configure options in timewarrior's own config file (`~/.config/timewarrior/timewarrior.cfg`):

```
reports.lume.birthday = 04-14
reports.lume.format = color
```

No separate config file is needed.

- `reports.lume.birthday` is optional and accepts `MM-DD` or `YYYY-MM-DD`. Default is `04-14` if not set.
- `reports.lume.format` is optional and accepts `color` or `markdown`. Default is `color` if not set.

## Requirements

- Go 1.22+
- Timewarrior with existing time entries
