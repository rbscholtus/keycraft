// keycraft-serve runs a localhost HTTP server exposing the keycraft API.
// It is intended for web-UI development (Vite proxies to it) or any
// same-machine client that wants to drive keycraft programmatically.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rbscholtus/keycraft/internal/api"
)

func main() {
	dataDir := flag.String("data-dir", "data", "directory containing corpus/, layouts/, config/")
	addr := flag.String("addr", "127.0.0.1:8787", "listen address (loopback only; refuses non-loopback binds)")
	timeout := flag.Duration("session-timeout", 30*time.Minute, "idle timeout before a session is evicted")
	maxSessions := flag.Int("max-sessions", 4, "maximum concurrent sessions")
	noAuth := flag.Bool("no-auth", false, "disable bearer-token auth (dev only)")
	flag.Parse()

	if err := refuseNonLoopback(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cfg := api.Config{
		DataDir:        *dataDir,
		SessionTimeout: *timeout,
		MaxSessions:    *maxSessions,
	}
	svc := api.NewService(api.NewSessionManager(cfg))

	token := ""
	if !*noAuth {
		token = api.NewAuthToken()
	}

	handler := api.RequireBearer(token, api.NewRouter(svc))

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("keycraft-serve listening on http://%s\n", *addr)
	fmt.Printf("data-dir: %s\n", *dataDir)
	if token != "" {
		fmt.Printf("auth token: %s\n", token)
	} else {
		fmt.Println("auth: DISABLED (--no-auth)")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil {
			fmt.Fprintln(os.Stderr, "server error:", err)
			os.Exit(1)
		}
	}
}

// refuseNonLoopback errors out on obviously-public binds like 0.0.0.0 or a
// public-looking host. We want to make it hard to accidentally expose the
// API to the network.
func refuseNonLoopback(addr string) error {
	host := addr
	if i := lastIndex(addr, ':'); i >= 0 {
		host = addr[:i]
	}
	switch host {
	case "", "127.0.0.1", "::1", "localhost":
		return nil
	}
	return fmt.Errorf("refusing non-loopback bind %q: use 127.0.0.1, ::1, or localhost", host)
}

func lastIndex(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
