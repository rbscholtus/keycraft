package api

import (
	"sync"
	"testing"
	"time"
)

func testConfig(t *testing.T) Config {
	t.Helper()
	cfg := DefaultConfig()
	cfg.DataDir = "../../data"
	cfg.SessionTimeout = 10 * time.Minute
	cfg.MaxSessions = 0 // unlimited for tests
	return cfg
}

func TestSessionLifecycle(t *testing.T) {
	mgr := NewSessionManager(testConfig(t))

	sess, err := mgr.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("empty session ID")
	}

	got, err := mgr.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.ID != sess.ID {
		t.Fatalf("GetSession returned wrong session")
	}

	mgr.CloseSession(sess.ID)
	if _, err := mgr.GetSession(sess.ID); err != ErrSessionNotFound {
		t.Fatalf("GetSession after close: got %v, want ErrSessionNotFound", err)
	}
}

func TestSessionLimit(t *testing.T) {
	cfg := testConfig(t)
	cfg.MaxSessions = 2
	mgr := NewSessionManager(cfg)

	if _, err := mgr.CreateSession(); err != nil {
		t.Fatalf("1st session: %v", err)
	}
	if _, err := mgr.CreateSession(); err != nil {
		t.Fatalf("2nd session: %v", err)
	}
	if _, err := mgr.CreateSession(); err != ErrSessionLimit {
		t.Fatalf("3rd session: got %v, want ErrSessionLimit", err)
	}
}

func TestSessionRunSemaphore(t *testing.T) {
	mgr := NewSessionManager(testConfig(t))
	sess, err := mgr.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if !sess.TryAcquireRun() {
		t.Fatal("first acquire should succeed")
	}
	if sess.TryAcquireRun() {
		t.Fatal("second acquire should fail (busy)")
	}
	sess.ReleaseRun()
	if !sess.TryAcquireRun() {
		t.Fatal("acquire after release should succeed")
	}
	sess.ReleaseRun()
}

func TestCorpusCacheReuse(t *testing.T) {
	mgr := NewSessionManager(testConfig(t))
	ref := CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0}

	c1, err := mgr.LoadCorpus(ref)
	if err != nil {
		t.Fatalf("1st load: %v", err)
	}
	c2, err := mgr.LoadCorpus(ref)
	if err != nil {
		t.Fatalf("2nd load: %v", err)
	}
	if c1 != c2 {
		t.Fatal("cache miss: different pointers for identical ref")
	}

	// two releases drop the entry
	mgr.ReleaseCorpus(ref)
	mgr.ReleaseCorpus(ref)

	mgr.mu.Lock()
	if _, ok := mgr.corpora[corpusKey(ref)]; ok {
		t.Error("corpus entry not evicted after refCount hit zero")
	}
	mgr.mu.Unlock()
}

func TestCorpusCacheConcurrentLoad(t *testing.T) {
	mgr := NewSessionManager(testConfig(t))
	ref := CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0}

	const N = 16
	var wg sync.WaitGroup
	wg.Add(N)
	results := make([]any, N)
	for i := range N {
		go func(i int) {
			defer wg.Done()
			c, err := mgr.LoadCorpus(ref)
			if err != nil {
				results[i] = err
				return
			}
			results[i] = c
		}(i)
	}
	wg.Wait()

	for range N {
		mgr.ReleaseCorpus(ref)
	}
	for _, r := range results {
		if err, ok := r.(error); ok {
			t.Fatalf("concurrent load error: %v", err)
		}
	}
}

func TestSessionEviction(t *testing.T) {
	cfg := testConfig(t)
	cfg.SessionTimeout = 10 * time.Millisecond
	mgr := NewSessionManager(cfg)

	sess, err := mgr.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	// Next CreateSession triggers eviction
	if _, err := mgr.CreateSession(); err != nil {
		t.Fatalf("CreateSession after sleep: %v", err)
	}

	if _, err := mgr.GetSession(sess.ID); err != ErrSessionNotFound {
		t.Fatalf("evicted session still findable: %v", err)
	}
}
