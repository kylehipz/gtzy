---
name: verify-gtzy
description: Full-stack verification for the gtzy repo (Go backend + React frontend) — build, vet, test, and frontend build in one pass. Use after any change to internal/, main.go, or web/ before considering the work done.
---

Run these in order and report failures with the exact command that failed:

1. `go build -o /tmp/gtzy-verify-build .` — must succeed with no CGO required (don't set CGO_ENABLED=0 explicitly unless double-checking; the default build must already be CGO-free since the driver is `modernc.org/sqlite`).
2. `go vet ./...` — must produce no output.
3. `go test ./...` — all packages must pass. `internal/timer` and `internal/store` carry the real test suite (timer state machine, recurrence firing rules, stats aggregation); a failure there is a regression, not flakiness.
4. `npm --prefix web run build` — must succeed (`tsc -b && vite build`). A TypeScript error here means a type mismatch was introduced, not a build-tool issue.

Clean up the build artifact from step 1 (`rm -f /tmp/gtzy-verify-build`) when done — don't leave a `gtzy` binary in the repo root from ad-hoc builds.

If steps 1-3 pass but you changed anything under `internal/api/` or `internal/timer/`, also do a quick smoke test against a temp DB before declaring done — the plan's one-active-task invariant and recurrence idempotency are the two things unit tests can miss if a new code path bypasses `internal/timer.Service` or the `ON CONFLICT` insert:

```sh
export GTZY_DB=$(mktemp -d)/verify.db
./gtzy serve --port 8499 &
sleep 1
curl -s localhost:8499/api/health
# ... exercise whatever changed via curl ...
kill %1
```

Don't claim a UI change works without actually looking at it — this repo has no browser tool bundled, so either run the frontend through Playwright/chromium-cli manually (see the `run` skill) or tell the user you verified the build only, not the rendered page.
