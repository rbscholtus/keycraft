package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// Config holds paths + policy for a SessionManager.
type Config struct {
	DataDir        string        // root data directory (contains corpus/, layouts/, config/)
	SessionTimeout time.Duration // idle timeout before a session is evicted
	MaxSessions    int           // hard cap on concurrent sessions
}

// DefaultConfig returns a Config using the CLI-compatible relative "data/" dir.
func DefaultConfig() Config {
	return Config{
		DataDir:        "data",
		SessionTimeout: 30 * time.Minute,
		MaxSessions:    4,
	}
}

// CorpusDir returns the absolute path to the corpus subdirectory.
func (c Config) CorpusDir() string { return filepath.Join(c.DataDir, "corpus") }

// LayoutDir returns the absolute path to the layouts subdirectory.
func (c Config) LayoutDir() string { return filepath.Join(c.DataDir, "layouts") }

// ConfigDir returns the absolute path to the config subdirectory.
func (c Config) ConfigDir() string { return filepath.Join(c.DataDir, "config") }

// Session holds per-connection state. Read/write paths acquire mu; the run
// semaphore blocks concurrent long-running operations on the same session.
type Session struct {
	ID         string
	createdAt  time.Time
	lastUsedAt time.Time
	runSem     chan struct{} // capacity 1: held while a long-running op is in flight
	mu         sync.RWMutex
}

func newSession() *Session {
	return &Session{
		ID:         newSessionID(),
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		runSem:     make(chan struct{}, 1),
	}
}

// TryAcquireRun reserves the session's one run slot. Returns false immediately
// if another run is in flight; release via ReleaseRun when done.
func (s *Session) TryAcquireRun() bool {
	select {
	case s.runSem <- struct{}{}:
		return true
	default:
		return false
	}
}

// ReleaseRun frees the run slot acquired by TryAcquireRun. Safe to call
// from the same goroutine that acquired it.
func (s *Session) ReleaseRun() {
	select {
	case <-s.runSem:
	default:
	}
}

// Touch updates lastUsedAt so the session avoids eviction.
func (s *Session) Touch() {
	s.mu.Lock()
	s.lastUsedAt = time.Now()
	s.mu.Unlock()
}

// corpusCacheEntry holds a loaded corpus plus its reference count. When
// refCount reaches zero the entry may be evicted.
type corpusCacheEntry struct {
	corpus   *kc.Corpus
	refCount int
}

// SessionManager tracks sessions and a process-wide corpus cache.
type SessionManager struct {
	cfg      Config
	mu       sync.Mutex
	sessions map[string]*Session
	corpora  map[string]*corpusCacheEntry // key = "name|coverage"
}

// NewSessionManager constructs a SessionManager with the provided config.
func NewSessionManager(cfg Config) *SessionManager {
	return &SessionManager{
		cfg:      cfg,
		sessions: make(map[string]*Session),
		corpora:  make(map[string]*corpusCacheEntry),
	}
}

// Config returns the manager's configuration (paths + policy).
func (m *SessionManager) Config() Config { return m.cfg }

// CreateSession opens a new session. Returns ErrSessionLimit when MaxSessions
// is reached and no idle session can be evicted.
func (m *SessionManager) CreateSession() (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.evictExpiredLocked()

	if m.cfg.MaxSessions > 0 && len(m.sessions) >= m.cfg.MaxSessions {
		return nil, ErrSessionLimit
	}

	sess := newSession()
	m.sessions[sess.ID] = sess
	return sess, nil
}

// GetSession returns an active session and refreshes its lastUsedAt. Returns
// ErrSessionNotFound if the ID is unknown or the session was evicted.
func (m *SessionManager) GetSession(id string) (*Session, error) {
	m.mu.Lock()
	sess, ok := m.sessions[id]
	m.mu.Unlock()
	if !ok {
		return nil, ErrSessionNotFound
	}
	sess.Touch()
	return sess, nil
}

// CloseSession removes a session. It's a no-op if the ID is unknown.
func (m *SessionManager) CloseSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

// Sessions returns a snapshot of live session IDs (for diagnostics/tests).
func (m *SessionManager) Sessions() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		out = append(out, id)
	}
	return out
}

// LoadCorpus returns the *Corpus for the given ref, loading from disk on
// cache miss. Cached corpora are shared across sessions and reference-counted;
// callers must pair this with ReleaseCorpus (typically deferred) to allow
// eventual eviction.
func (m *SessionManager) LoadCorpus(ref CorpusRef) (*kc.Corpus, error) {
	if ref.Name == "" {
		return nil, fmt.Errorf("%w: corpus name is required", ErrInvalidInput)
	}
	if ref.Coverage <= 0 {
		ref.Coverage = 98.0 // CLI default
	}

	key := corpusKey(ref)

	m.mu.Lock()
	if entry, ok := m.corpora[key]; ok {
		entry.refCount++
		m.mu.Unlock()
		return entry.corpus, nil
	}
	m.mu.Unlock()

	path := filepath.Join(m.cfg.CorpusDir(), ref.Name)
	corpusName := trimExt(ref.Name)
	corpus, err := kc.NewCorpusFromFile(corpusName, path, false, ref.Coverage)
	if err != nil {
		return nil, fmt.Errorf("could not load corpus: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.corpora[key]; ok { // another goroutine raced us
		entry.refCount++
		return entry.corpus, nil
	}
	m.corpora[key] = &corpusCacheEntry{corpus: corpus, refCount: 1}
	return corpus, nil
}

// ReleaseCorpus decrements the ref count on a cached corpus.
func (m *SessionManager) ReleaseCorpus(ref CorpusRef) {
	key := corpusKey(ref)
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.corpora[key]
	if !ok {
		return
	}
	entry.refCount--
	if entry.refCount <= 0 {
		delete(m.corpora, key)
	}
}

func (m *SessionManager) evictExpiredLocked() {
	if m.cfg.SessionTimeout <= 0 {
		return
	}
	cutoff := time.Now().Add(-m.cfg.SessionTimeout)
	for id, sess := range m.sessions {
		sess.mu.RLock()
		last := sess.lastUsedAt
		sess.mu.RUnlock()
		if last.Before(cutoff) {
			delete(m.sessions, id)
		}
	}
}

func corpusKey(ref CorpusRef) string {
	return fmt.Sprintf("%s|%.4f", ref.Name, ref.Coverage)
}

func trimExt(name string) string {
	if ext := filepath.Ext(name); ext != "" {
		return name[:len(name)-len(ext)]
	}
	return name
}

func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failure is fatal in practice; fall back to a time-based
		// token rather than panicking in a server hot path.
		return fmt.Sprintf("sess-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
