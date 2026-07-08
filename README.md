# gtzy

A single-user, locally-run productivity app: task manager with a strict
one-task-at-a-time timer, a calendar-based progress tracker, a journal, and
an analytics dashboard with an optional AI growth summary. One Go binary
serves a REST API, the embedded React SPA, and doubles as a CLI — built for
scripting into Hyprland keybinds, a Wofi task picker, and a Waybar module.

## Build

```sh
make build
```

This runs `npm install` + `npm run build` in `web/`, then `go build -o gtzy .`
— producing a single self-contained binary (no CGO) with the frontend
embedded.

For iterating on just the backend: `go build -o gtzy .` works fine without
touching `web/` — the binary falls back to API-only mode if `web/dist` hasn't
been built yet (a placeholder keeps the embed valid either way).

## Install

```sh
make install
```

Builds the frontend and runs `go install .`, which puts the `gtzy` binary at
`$(go env GOPATH)/bin/gtzy` (usually `~/go/bin`). Make sure that directory is
on your `PATH` so `gtzy` is runnable from anywhere.

## Run

```sh
./gtzy serve
```

Serves the API and the web UI at `http://localhost:8420` (override the port
with `--port`, or the database path with `--db`). The SQLite database lives
at `~/.local/share/gtzy/gtzy.db` by default (override with `GTZY_DB`).

### Development

Two processes, hot reload on the frontend:

```sh
make dev-server   # go run . serve --port 8420
make dev-web      # vite dev server on 5173, proxies /api to :8420
```

## CLI

All subcommands besides `serve` are thin HTTP clients against the running
server (default `http://localhost:8420`, override with `GTZY_URL`):

| Command | Action |
|---|---|
| `gtzy serve [--port 8420] [--db path]` | run the server |
| `gtzy next` | advance to the next eligible task, pausing the current one |
| `gtzy start <id>` | start a task (pauses whatever else is running) |
| `gtzy pause` | pause whatever is currently running |
| `gtzy complete [<id>]` | complete the current task, or a given id |
| `gtzy current` | print the current task + elapsed time |
| `gtzy add "<title>" [--priority p] [--category name] [--est minutes] [--date YYYY-MM-DD] [--at HH:MM] [--repeat daily\|weekly:1,3,5\|monthly:15]` | create a task, or a recurring rule with `--repeat` |
| `gtzy list [--today] [--wofi] [--json]` | list tasks |
| `gtzy waybar` | print Waybar module JSON |

Run `gtzy --help` or `gtzy <command> --help` for full usage and flag
descriptions (the CLI is built on Cobra).

## AI growth summaries (optional)

Set `ANTHROPIC_API_KEY` to enable the Dashboard's AI summary panel (uses
`claude-opus-4-8` by default; override with `GTZY_AI_MODEL`). Without a key,
the feature no-ops cleanly — the dashboard shows a placeholder card instead.

## Blood-sugar tracking

The Blood Sugar tab stores glucose readings (mg/dL), charts them against a
target range, and — with `ANTHROPIC_API_KEY` set — generates an AI summary that
interprets the trend. Readings can be entered by hand or pulled straight off a
Bluetooth glucose meter (Accu-Chek Guide or any device implementing the standard
Bluetooth Glucose Profile) via `gtzy sync` / the **Sync meter** button.

Bluetooth sync requires a one-time pairing and runs on the machine hosting
`gtzy serve`. See [docs/bluetooth-sync.md](docs/bluetooth-sync.md) for setup,
how records map to the database, and troubleshooting.

## Desktop integration (Hyprland / Waybar / Wofi)

These are documented here, not shipped as config — wire them up to taste.

### Waybar module

Add a `custom/gtzy` module to your Waybar config:

```jsonc
"custom/gtzy": {
  "exec": "gtzy waybar",
  "return-type": "json",
  "interval": 1,
  "on-click": "gtzy pause"
}
```

`gtzy waybar` always exits 0 and prints `{"text", "tooltip", "class", "percentage"}`
— `class` is `running`, `paused`, or `idle`.

### Wofi task picker

```sh
gtzy list --today --wofi | wofi --dmenu | cut -f1 | xargs gtzy start
```

Bind that to a Hyprland keybind, e.g. in `hyprland.conf`:

```
bind = $mainMod, T, exec, gtzy list --today --wofi | wofi --dmenu | cut -f1 | xargs gtzy start
bind = $mainMod SHIFT, T, exec, gtzy next
bind = $mainMod, P, exec, gtzy pause
```

## Tests

```sh
make test
```

Covers the timer state machine (start/pause accumulation, one-task-at-a-time,
`next` ordering), recurrence firing rules (daily/weekly/monthly + interval,
month-length clamping), and stats aggregation.
