package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// tempDataDir builds an isolated copy of the repo's data/ directory containing
// only the requested layouts, the monkeyracer corpus JSON cache, and the
// supplied configs. Generated .klf files (from optimise / generate) end up
// inside the temp dir and are cleaned up automatically.
func tempDataDir(t *testing.T, layouts []string, configs []string) string {
	t.Helper()
	tmp := t.TempDir()
	for _, sub := range []string{"corpus", "layouts", "config"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}

	cp := func(rel, dst string) {
		t.Helper()
		src := filepath.Join("..", "..", "data", rel)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}

	// Corpus JSON cache: the .txt may not exist in the repo; loader falls back
	// to the .json cache when the source file is missing.
	cp("corpus/monkeyracer.txt.json", filepath.Join(tmp, "corpus", "monkeyracer.txt.json"))

	// NewScorer needs >=2 layouts in the dir to compute medians/IQRs. Always
	// include a baseline of well-known reference layouts so the scorer is
	// happy regardless of what individual tests asked for.
	baseline := []string{"qwerty", "colemak", "colemak-dh", "dvorak"}
	requested := make(map[string]bool)
	for _, l := range layouts {
		requested[l] = true
	}
	for _, b := range baseline {
		requested[b] = true
	}
	for l := range requested {
		cp("layouts/"+l+".klf", filepath.Join(tmp, "layouts", l+".klf"))
	}
	for _, c := range configs {
		cp("config/"+c, filepath.Join(tmp, "config", c))
	}
	return tmp
}

func newServiceWithDir(t *testing.T, dataDir string) *Service {
	t.Helper()
	cfg := DefaultConfig()
	cfg.DataDir = dataDir
	cfg.SessionTimeout = 10 * time.Minute
	cfg.MaxSessions = 0
	return NewService(NewSessionManager(cfg))
}

// ---------- OptimiseFast ----------

func TestServiceOptimiseFastHappy(t *testing.T) {
	dir := tempDataDir(t, []string{"qwerty"}, nil)
	svc := newServiceWithDir(t, dir)
	sess, _ := svc.OpenSession()

	resp, err := svc.OptimiseFast(context.Background(), sess.ID, OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
		Seed:   42,
	})
	if err != nil {
		t.Fatalf("OptimiseFast: %v", err)
	}
	if resp.Original.Name != "qwerty" {
		t.Errorf("original name: got %q", resp.Original.Name)
	}
	// Best layout name should carry the -opt suffix.
	if resp.Best.Name != "qwerty-opt" {
		t.Errorf("best name: got %q, want qwerty-opt", resp.Best.Name)
	}
	if _, ok := resp.Metrics["SFB"]; !ok {
		t.Error("expected SFB metric in response")
	}
}

func TestServiceOptimiseFastBadSession(t *testing.T) {
	dir := tempDataDir(t, []string{"qwerty"}, nil)
	svc := newServiceWithDir(t, dir)

	_, err := svc.OptimiseFast(context.Background(), "nope", OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if err != ErrSessionNotFound {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestServiceOptimiseFastMissingLayout(t *testing.T) {
	dir := tempDataDir(t, []string{"qwerty"}, nil)
	svc := newServiceWithDir(t, dir)
	sess, _ := svc.OpenSession()

	_, err := svc.OptimiseFast(context.Background(), sess.ID, OptimiseFastRequest{
		Layout: "no-such-layout",
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want wrapped ErrNotFound", err)
	}
}

func TestServiceOptimiseFastBusy(t *testing.T) {
	dir := tempDataDir(t, []string{"qwerty"}, nil)
	svc := newServiceWithDir(t, dir)
	dto, _ := svc.OpenSession()
	sess, _ := svc.Manager().GetSession(dto.ID)

	// Manually hold the run slot so OptimiseFast immediately sees busy.
	if !sess.TryAcquireRun() {
		t.Fatal("could not acquire run slot")
	}
	defer sess.ReleaseRun()

	_, err := svc.OptimiseFast(context.Background(), dto.ID, OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	if err != ErrBusy {
		t.Fatalf("got %v, want ErrBusy", err)
	}
}

func TestServiceOptimiseFastCancel(t *testing.T) {
	dir := tempDataDir(t, []string{"qwerty"}, nil)
	svc := newServiceWithDir(t, dir)
	dto, _ := svc.OpenSession()
	sess, _ := svc.Manager().GetSession(dto.ID)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before invoking — first ctx check trips

	start := time.Now()
	_, err := svc.OptimiseFast(ctx, dto.ID, OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	if elapsed > 3*time.Second {
		t.Errorf("cancel slow: %v", elapsed)
	}

	// Run slot must be released after cancellation, otherwise the next call hangs.
	if !sess.TryAcquireRun() {
		t.Error("run slot not released after cancellation")
	}
	sess.ReleaseRun()
}

// ---------- GenerateFast ----------

func TestServiceGenerateFastHappy(t *testing.T) {
	dir := tempDataDir(t, nil, []string{"vowels.gen"})
	svc := newServiceWithDir(t, dir)
	sess, _ := svc.OpenSession()

	resp, err := svc.GenerateFast(context.Background(), sess.ID, GenerateFastRequest{
		Config:     "vowels.gen",
		MaxLayouts: 3,
		Seed:       42,
	})
	if err != nil {
		t.Fatalf("GenerateFast: %v", err)
	}
	if resp.Generated == 0 {
		t.Fatal("no layouts generated")
	}
	if len(resp.Layouts) != resp.Generated {
		t.Errorf("count mismatch: layouts=%d generated=%d", len(resp.Layouts), resp.Generated)
	}
	// Verify the files actually landed in the temp layouts dir.
	for _, name := range resp.Layouts {
		path := filepath.Join(dir, "layouts", name+".klf")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s on disk: %v", path, err)
		}
	}
}

func TestServiceGenerateFastBadSession(t *testing.T) {
	dir := tempDataDir(t, nil, []string{"vowels.gen"})
	svc := newServiceWithDir(t, dir)

	_, err := svc.GenerateFast(context.Background(), "nope", GenerateFastRequest{
		Config: "vowels.gen", MaxLayouts: 1,
	})
	if err != ErrSessionNotFound {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestServiceGenerateFastEmptyConfig(t *testing.T) {
	dir := tempDataDir(t, nil, nil)
	svc := newServiceWithDir(t, dir)
	sess, _ := svc.OpenSession()

	_, err := svc.GenerateFast(context.Background(), sess.ID, GenerateFastRequest{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}

func TestServiceGenerateFastBusy(t *testing.T) {
	dir := tempDataDir(t, nil, []string{"vowels.gen"})
	svc := newServiceWithDir(t, dir)
	dto, _ := svc.OpenSession()
	sess, _ := svc.Manager().GetSession(dto.ID)
	if !sess.TryAcquireRun() {
		t.Fatal("could not acquire run slot")
	}
	defer sess.ReleaseRun()

	_, err := svc.GenerateFast(context.Background(), dto.ID, GenerateFastRequest{
		Config: "vowels.gen", MaxLayouts: 1,
	})
	if err != ErrBusy {
		t.Fatalf("got %v, want ErrBusy", err)
	}
}

