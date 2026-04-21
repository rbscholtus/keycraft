package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rbscholtus/keycraft/internal/api"
)

// tempDataDir mirrors the helper used by the api and e2e packages: builds a
// temp data-dir seeded with a small set of reference layouts plus the
// monkeyracer corpus JSON cache.
func tempDataDir(t *testing.T, layouts, configs []string) string {
	t.Helper()
	tmp := t.TempDir()
	for _, sub := range []string{"corpus", "layouts", "config"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	cp := func(rel string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("..", "..", "data", rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, rel), data, 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	cp("corpus/monkeyracer.txt.json")
	baseline := map[string]bool{
		"qwerty": true, "colemak": true, "colemak-dh": true, "dvorak": true,
	}
	for _, l := range layouts {
		baseline[l] = true
	}
	for l := range baseline {
		cp("layouts/" + l + ".klf")
	}
	for _, c := range configs {
		cp("config/" + c)
	}
	return tmp
}

func newTestApp(t *testing.T, layouts, configs []string) *App {
	t.Helper()
	cfg := api.Config{
		DataDir:        tempDataDir(t, layouts, configs),
		SessionTimeout: 5 * time.Minute,
		MaxSessions:    0,
	}
	return NewApp(api.NewService(api.NewSessionManager(cfg)))
}

func TestAppSessionLazyOpen(t *testing.T) {
	app := newTestApp(t, nil, nil)
	if app.sessionID != "" {
		t.Error("session should not exist before first use")
	}
	if _, err := app.ListLayouts(); err != nil {
		t.Fatalf("ListLayouts: %v", err)
	}
	// ListLayouts doesn't need a session — sessionID still empty.
	if app.sessionID != "" {
		t.Error("ListLayouts should not open a session")
	}
	// Analyse opens one.
	_, _ = app.Analyse(api.AnalyseRequest{
		Layouts: []string{"qwerty"},
		Corpus:  api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if app.sessionID == "" {
		t.Error("Analyse should open a session")
	}
}

func TestAppGetLayout(t *testing.T) {
	app := newTestApp(t, nil, nil)
	dto, err := app.GetLayout("qwerty")
	if err != nil {
		t.Fatalf("GetLayout: %v", err)
	}
	if dto.Name != "qwerty" {
		t.Errorf("name: %q", dto.Name)
	}
}

func TestAppFlipLayout(t *testing.T) {
	app := newTestApp(t, nil, nil)
	resp, err := app.FlipLayout("qwerty")
	if err != nil {
		t.Fatalf("FlipLayout: %v", err)
	}
	if resp.Layout.Name != "qwerty-flipped" {
		t.Errorf("name: %q", resp.Layout.Name)
	}
}

func TestAppOptimiseFast(t *testing.T) {
	app := newTestApp(t, []string{"qwerty"}, nil)
	resp, err := app.OptimiseFast(context.Background(), api.OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
		Seed:   42,
	})
	if err != nil {
		t.Fatalf("OptimiseFast: %v", err)
	}
	if resp.Best.Name != "qwerty-opt" {
		t.Errorf("best: %q", resp.Best.Name)
	}
}

func TestAppOptimiseFastCancel(t *testing.T) {
	app := newTestApp(t, []string{"qwerty"}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := app.OptimiseFast(ctx, api.OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestAppGenerateFast(t *testing.T) {
	app := newTestApp(t, nil, []string{"vowels.gen"})
	resp, err := app.GenerateFast(context.Background(), api.GenerateFastRequest{
		Config:     "vowels.gen",
		MaxLayouts: 2,
		Seed:       42,
	})
	if err != nil {
		t.Fatalf("GenerateFast: %v", err)
	}
	if resp.Generated == 0 {
		t.Error("no layouts generated")
	}
}
