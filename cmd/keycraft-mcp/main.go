// keycraft-mcp is a Model Context Protocol (MCP) server exposing the
// keycraft API to LLM clients (Claude Desktop, etc) over stdio. It speaks
// newline-delimited JSON-RPC 2.0.
//
// MCP clients launch this binary as a subprocess and communicate via
// stdin/stdout. Because the working directory is unpredictable, the
// --data-dir flag (defaulting to $XDG_DATA_HOME/keycraft, falling back to
// ./data) controls where layouts, corpora, and gen configs are read from.
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
	timeout := flag.Duration("session-timeout", 30*time.Minute, "idle timeout before a session is evicted")
	flag.Parse()

	cfg := api.Config{
		DataDir:        *dataDir,
		SessionTimeout: *timeout,
		MaxSessions:    0,
	}
	svc := api.NewService(api.NewSessionManager(cfg))

	srv := newServer(svc)
	if err := srv.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

// defaultDataDir resolves the layout/corpus/config directory.
// XDG_DATA_HOME/keycraft if set; $HOME/.local/share/keycraft otherwise. If
// neither path exists, falls back to "data" relative to CWD (developer mode).
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
