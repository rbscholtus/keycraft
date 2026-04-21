// keycraft-wails is the (future) Wails desktop binary. The Phase 4 scaffold
// only contains the Go-side App adapter that exposes api.Service to a Wails
// frontend. The actual Wails framework dependency, embedded JS bundle, and
// runtime wiring are intentionally omitted from v1 — this directory is the
// integration point for when the desktop UI is built out.
//
// To turn this into a Wails app:
//
//  1. Install the Wails CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest
//  2. From this directory, run `wails init -n keycraft-wails -t vanilla -ide vscode`
//     and merge the generated main.go entry point with App below — the Bind
//     section should reference the App struct.
//  3. Run `wails dev` for hot-reload, `wails build` for the distributable.
//
// Until that's wired, this binary builds and prints a one-line note.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rbscholtus/keycraft/internal/api"
)

func main() {
	dataDir := flag.String("data-dir", defaultDataDir(), "directory containing corpus/, layouts/, config/")
	flag.Parse()

	cfg := api.Config{
		DataDir:        *dataDir,
		SessionTimeout: 30 * time.Minute,
		MaxSessions:    1, // one window = one session
	}
	app := NewApp(api.NewService(api.NewSessionManager(cfg)))
	_ = app // wired into Wails Bind() once the framework is integrated

	fmt.Fprintln(os.Stderr, "keycraft-wails: Phase 4 scaffold — not yet wired to Wails. Run keycraft-serve for HTTP access.")
}

// defaultDataDir mirrors keycraft-mcp's resolution: XDG location, then a
// CWD-relative ./data fallback for development.
func defaultDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		candidate := filepath.Join(xdg, "keycraft")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, ".local", "share", "keycraft")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "data"
}
