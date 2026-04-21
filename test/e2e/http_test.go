// Package e2e exercises keycraft API transports end-to-end. HTTP tests spin
// up a real HTTP server via httptest and drive it with a real net/http
// client; MCP tests will spawn the compiled binary and speak JSON-RPC over
// its stdio.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rbscholtus/keycraft/internal/api"
)

const dummyToken = "test-token-abc123"

// httpFixture mirrors cmd/keycraft-serve's wiring for use from tests.
type httpFixture struct {
	server *httptest.Server
	token  string
}

func (f *httpFixture) Close() { f.server.Close() }

func (f *httpFixture) do(t *testing.T, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		r = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, f.server.URL+path, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// readJSON decodes the response body into v and returns the status code.
func readJSON(t *testing.T, resp *http.Response, v any) int {
	t.Helper()
	defer resp.Body.Close()
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil && err != io.EOF {
			t.Fatalf("decode: %v", err)
		}
	}
	return resp.StatusCode
}

// newFixture spins up the HTTP server backed by the repo's data/ directory.
// It requires a bearer token so auth-failure paths are exercised alongside
// the happy ones.
func newFixture(t *testing.T) *httpFixture {
	t.Helper()

	dataDir := findDataDir(t)

	cfg := api.Config{
		DataDir:        dataDir,
		SessionTimeout: 10 * time.Minute,
		MaxSessions:    0,
	}
	svc := api.NewService(api.NewSessionManager(cfg))

	handler := api.RequireBearer(dummyToken, api.NewRouter(svc))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return &httpFixture{server: srv, token: dummyToken}
}

// findDataDir walks up from the test directory until it finds a dir
// containing both "layouts" and "corpus" subdirectories.
func findDataDir(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for range 8 {
		candidate := filepath.Join(dir, "data")
		if _, err := os.Stat(filepath.Join(candidate, "layouts")); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate data/ ancestor of %s", cwd)
	return ""
}

// newFixtureWithTempDataDir spins up the HTTP server backed by a fresh
// temp data-dir seeded from the repo's data/. Use for tests that mutate the
// layouts directory (optimise-fast saves -opt files; generate-fast persists
// generated layouts). Always includes a small set of reference layouts so
// the scorer's medians/IQRs computation has enough data points.
func newFixtureWithTempDataDir(t *testing.T, layouts, configs []string) (*httpFixture, string) {
	t.Helper()
	src := findDataDir(t)
	tmp := t.TempDir()
	for _, sub := range []string{"corpus", "layouts", "config"} {
		if err := os.MkdirAll(filepath.Join(tmp, sub), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}

	cp := func(rel string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(src, rel))
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

	cfg := api.Config{
		DataDir:        tmp,
		SessionTimeout: 10 * time.Minute,
		MaxSessions:    0,
	}
	svc := api.NewService(api.NewSessionManager(cfg))
	handler := api.RequireBearer(dummyToken, api.NewRouter(svc))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return &httpFixture{server: srv, token: dummyToken}, tmp
}

// ---------- auth ----------

func TestHTTPAuthMissingToken(t *testing.T) {
	f := newFixture(t)
	resp, err := http.Get(f.server.URL + "/v1/corpora")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("missing token: got %d, want 401", resp.StatusCode)
	}
}

func TestHTTPAuthWrongToken(t *testing.T) {
	f := newFixture(t)
	req, _ := http.NewRequest("GET", f.server.URL+"/v1/corpora", nil)
	req.Header.Set("Authorization", "Bearer nope")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong token: got %d, want 401", resp.StatusCode)
	}
}

// ---------- sessions ----------

func TestHTTPSessionLifecycle(t *testing.T) {
	f := newFixture(t)

	var sess api.SessionDTO
	resp := f.do(t, "POST", "/v1/sessions", nil, nil)
	if status := readJSON(t, resp, &sess); status != http.StatusOK {
		t.Fatalf("POST /v1/sessions: status %d", status)
	}
	if sess.ID == "" {
		t.Fatal("empty session ID")
	}

	del := f.do(t, "DELETE", "/v1/sessions/"+sess.ID, nil, nil)
	del.Body.Close()
	if del.StatusCode != http.StatusNoContent {
		t.Errorf("DELETE /v1/sessions: got %d, want 204", del.StatusCode)
	}
}

// ---------- corpora ----------

func TestHTTPCorporaList(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "GET", "/v1/corpora", nil, nil)
	var body struct {
		Corpora []string `json:"corpora"`
	}
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	// body.Corpora may be empty in CI (only .json caches checked in); the
	// call still has to succeed.
	_ = body
}

func TestHTTPCorpusLoadRequiresSession(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "POST", "/v1/corpora/load", api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0}, nil)
	readJSON(t, resp, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("no session header: got %d, want 400", resp.StatusCode)
	}
}

func TestHTTPCorpusLoad(t *testing.T) {
	f := newFixture(t)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/corpora/load",
		api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
		map[string]string{"X-Keycraft-Session": sess.ID})
	var summary api.CorpusSummaryDTO
	if status := readJSON(t, resp, &summary); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if summary.Name == "" {
		t.Error("summary.Name empty")
	}
	if summary.TotalUnigramsCount == 0 {
		t.Error("expected non-zero total unigrams")
	}
}

// ---------- layouts ----------

func TestHTTPLayoutsList(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "GET", "/v1/layouts", nil, nil)
	var body struct {
		Layouts []string `json:"layouts"`
	}
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if len(body.Layouts) == 0 {
		t.Fatal("expected at least one layout")
	}
}

func TestHTTPGetLayout(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "GET", "/v1/layouts/qwerty", nil, nil)
	var dto api.LayoutDTO
	if status := readJSON(t, resp, &dto); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if dto.Name != "qwerty" {
		t.Errorf("name: got %q", dto.Name)
	}
	if len(dto.Rows) != 3 {
		t.Errorf("rows: got %d, want 3", len(dto.Rows))
	}
}

func TestHTTPGetLayoutMissing(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "GET", "/v1/layouts/not-a-layout", nil, nil)
	readJSON(t, resp, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("missing layout: got %d, want 404", resp.StatusCode)
	}
}

func TestHTTPFlipLayout(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "POST", "/v1/layouts/qwerty/flip", nil, nil)
	var body api.FlipResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if body.Layout.Name != "qwerty-flipped" {
		t.Errorf("name: got %q, want %q", body.Layout.Name, "qwerty-flipped")
	}
}

// ---------- analyse / view / rank ----------

func TestHTTPAnalyse(t *testing.T) {
	f := newFixture(t)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/analyse", api.AnalyseRequest{
		Layouts: []string{"qwerty", "colemak-dh"},
		Corpus:  api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	}, map[string]string{"X-Keycraft-Session": sess.ID})

	var body api.AnalyseResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if len(body.Results) != 2 {
		t.Fatalf("results: got %d, want 2", len(body.Results))
	}
	if _, ok := body.Results[0].Metrics["SFB"]; !ok {
		t.Error("expected SFB metric")
	}
}

func TestHTTPAnalyseBadSession(t *testing.T) {
	f := newFixture(t)
	resp := f.do(t, "POST", "/v1/analyse", api.AnalyseRequest{
		Layouts: []string{"qwerty"},
		Corpus:  api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	}, map[string]string{"X-Keycraft-Session": "unknown-id"})
	readJSON(t, resp, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown session: got %d, want 404", resp.StatusCode)
	}
}

func TestHTTPAnalyseEmptyLayouts(t *testing.T) {
	f := newFixture(t)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/analyse", api.AnalyseRequest{
		Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	}, map[string]string{"X-Keycraft-Session": sess.ID})
	readJSON(t, resp, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty layouts: got %d, want 400", resp.StatusCode)
	}
}

func TestHTTPView(t *testing.T) {
	f := newFixture(t)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/view", api.ViewRequest{
		Layouts: []string{"qwerty"},
		Corpus:  api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	}, map[string]string{"X-Keycraft-Session": sess.ID})
	var body api.ViewResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if len(body.Results) != 1 {
		t.Fatalf("results: got %d, want 1", len(body.Results))
	}
}

// ---------- optimise-fast / generate-fast ----------

func TestHTTPOptimiseFast(t *testing.T) {
	f, _ := newFixtureWithTempDataDir(t, []string{"qwerty"}, nil)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/optimise-fast", api.OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
		Seed:   42,
	}, map[string]string{"X-Keycraft-Session": sess.ID})

	var body api.OptimiseFastResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if body.Original.Name != "qwerty" {
		t.Errorf("original: got %q", body.Original.Name)
	}
	if body.Best.Name != "qwerty-opt" {
		t.Errorf("best: got %q", body.Best.Name)
	}
	if _, ok := body.Metrics["SFB"]; !ok {
		t.Error("expected SFB metric in response")
	}
}

func TestHTTPOptimiseFastBusy(t *testing.T) {
	f, _ := newFixtureWithTempDataDir(t, []string{"qwerty"}, nil)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	// Fire two concurrent calls. One should succeed (eventually); the other
	// should return 429 immediately because the session's run slot is held.
	statusCh := make(chan int, 2)
	for range 2 {
		go func() {
			resp := f.do(t, "POST", "/v1/optimise-fast", api.OptimiseFastRequest{
				Layout: "qwerty",
				Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
				Seed:   42,
			}, map[string]string{"X-Keycraft-Session": sess.ID})
			defer resp.Body.Close()
			statusCh <- resp.StatusCode
		}()
	}

	statuses := []int{}
	for range 2 {
		statuses = append(statuses, <-statusCh)
	}

	hasOK, hasBusy := false, false
	for _, s := range statuses {
		switch s {
		case http.StatusOK:
			hasOK = true
		case http.StatusTooManyRequests:
			hasBusy = true
		}
	}
	if !hasOK || !hasBusy {
		t.Errorf("expected one 200 + one 429, got %v", statuses)
	}
}

func TestHTTPOptimiseFastCancel(t *testing.T) {
	f, _ := newFixtureWithTempDataDir(t, []string{"qwerty"}, nil)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	// Cancel ctx before the request hits the wire — server returns context.Canceled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	body, _ := json.Marshal(api.OptimiseFastRequest{
		Layout: "qwerty",
		Corpus: api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", f.server.URL+"/v1/optimise-fast", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.token)
	req.Header.Set("X-Keycraft-Session", sess.ID)

	_, err := http.DefaultClient.Do(req)
	if err == nil {
		t.Fatal("expected client-side error from cancelled ctx")
	}
}

func TestHTTPGenerateFast(t *testing.T) {
	f, dir := newFixtureWithTempDataDir(t, nil, []string{"vowels.gen"})
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/generate-fast", api.GenerateFastRequest{
		Config:     "vowels.gen",
		MaxLayouts: 3,
		Seed:       42,
	}, map[string]string{"X-Keycraft-Session": sess.ID})

	var body api.GenerateFastResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if body.Generated == 0 {
		t.Fatal("no layouts generated")
	}
	for _, name := range body.Layouts {
		if _, err := os.Stat(filepath.Join(dir, "layouts", name+".klf")); err != nil {
			t.Errorf("expected %s on disk: %v", name, err)
		}
	}
}

func TestHTTPRank(t *testing.T) {
	f := newFixture(t)
	var sess api.SessionDTO
	readJSON(t, f.do(t, "POST", "/v1/sessions", nil, nil), &sess)

	resp := f.do(t, "POST", "/v1/rank", api.RankRequest{
		Layouts: []string{"qwerty", "colemak-dh"},
		Corpus:  api.CorpusRef{Name: "monkeyracer.txt", Coverage: 98.0},
	}, map[string]string{"X-Keycraft-Session": sess.ID})
	var body api.RankResponse
	if status := readJSON(t, resp, &body); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if len(body.Scores) != 2 {
		t.Fatalf("scores: got %d, want 2", len(body.Scores))
	}
	if len(body.Medians) == 0 {
		t.Error("medians empty")
	}
}
