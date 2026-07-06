# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / test / dev

- `make build` — `npm --prefix web install && npm --prefix web run build && go build -o gtzy .` (single self-contained binary, no CGO)
- `make test` — `go test ./...`
- `make dev-server` — `go run . serve` (backend on :8420)
- `make dev-web` — `npm --prefix web run dev` (Vite dev server on :5173, proxies `/api` to `GTZY_URL` or `http://localhost:8420`)
- `make clean` — removes the binary and resets `web/dist/` to just the `.gitkeep` placeholder
- Web lint: `npm --prefix web run lint` (oxlint)
- Go lint: `golangci-lint run ./...` (`.golangci.yml`; errcheck on deferred `Close`/`Parse` is intentionally excluded — not actionable on read paths and cleanup here)

## Env vars

- `GTZY_DB` — sqlite db path (default `~/.local/share/gtzy/gtzy.db`)
- `GTZY_URL` — CLI/dev-proxy base URL for the running server (default `http://localhost:8420`)
- `GTZY_AI_MODEL` — Claude model for AI summaries (default `claude-opus-4-8`)
- `ANTHROPIC_API_KEY` — presence alone enables the AI summary feature (`internal/ai/claude.go`); unset means it no-ops cleanly, don't add fallback logic around this

## Architecture invariants

- **One active task at a time.** All timer mutations (start/pause/complete/next) must go through `internal/timer.Service` — it enforces that at most one `tasks` row has `active_started_at` non-null. Don't write `active_started_at` anywhere else.
- **Recurrence materialization is idempotent by a DB constraint, not app logic.** The partial unique index `idx_tasks_recur_day ON tasks(recurrence_id, scheduled_date) WHERE recurrence_id IS NOT NULL` is what makes re-running `EnsureInstancesForDate` safe to call on every read. If you change how recurring instances are inserted, keep the `ON CONFLICT` targeting this index.
- **`web/dist/.gitkeep` must stay committed.** `//go:embed all:dist` in `web/embed.go` requires `web/dist/` to exist with at least one file at Go build time. `npm run build` (Vite) wipes the directory contents on every build, including `.gitkeep` — that's expected and fine locally, but never `git rm` the committed `.gitkeep` or a fresh clone's `go build` breaks before the frontend is built.
