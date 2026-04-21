# keycraft-wails

Phase 4 scaffold for the keycraft desktop application. The Go-side adapter
(`app.go`) wraps `internal/api.Service` and is the binding target for a
future Wails v2 frontend.

## What's here

- `main.go` — entrypoint that constructs an `App` from `api.Service`. Runs
  standalone today (no Wails dependency); prints a one-line note and exits.
- `app.go` — the `App` struct exposing every API surface as a method
  (ListLayouts, Analyse, OptimiseFast, etc.). Each method is the future
  Wails binding target.
- `app_test.go` — three-tier coverage of the App adapter (lazy session,
  delegation to Service, ctx cancellation).
- `frontend/` — placeholder for the Vite/Svelte/etc. frontend project.

## Wiring up Wails

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd cmd/keycraft-wails
wails init -n keycraft-wails -t vanilla -ide vscode
# Merge the generated main.go entry with the existing one. The Bind block
# should reference NewApp(...) from app.go.
wails dev    # hot-reload during development
wails build  # produces a packaged distributable
```

## Why scaffold-only

The MCP and HTTP transports already exercise the full `api.Service` surface,
so the Go side is fully tested. Wails integration is mostly a
frontend/toolchain problem and was deferred per the project plan to keep v1
focused on the backend.
