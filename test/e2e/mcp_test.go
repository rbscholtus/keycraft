package e2e

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// mcpClient drives a spawned keycraft-mcp subprocess via JSON-RPC over its
// stdin/stdout. One client per test — the binary is shared but sessions are
// per-connection, and each test gets its own subprocess.
type mcpClient struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	scanner *bufio.Scanner
	mu      sync.Mutex
	nextID  int
	dataDir string
}

// pendingResponse is the parsed JSON-RPC envelope, with raw result bytes for
// test-specific decoding.
type pendingResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func startMCP(t *testing.T, layouts, configs []string) *mcpClient {
	t.Helper()

	dir := newTempDataDir(t, layouts, configs)

	binPath := filepath.Join(t.TempDir(), "keycraft-mcp")
	build := exec.Command("go", "build", "-o", binPath, "../../cmd/keycraft-mcp")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build: %v", err)
	}

	cmd := exec.Command(binPath, "--data-dir", dir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout: %v", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	c := &mcpClient{
		cmd:     cmd,
		in:      stdin,
		scanner: bufio.NewScanner(stdout),
		dataDir: dir,
	}
	c.scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	t.Cleanup(func() {
		_ = stdin.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	})
	return c
}

// call sends a JSON-RPC request and returns the parsed response. Returns the
// raw error if one came back from the server.
func (c *mcpClient) call(t *testing.T, method string, params any) *pendingResponse {
	t.Helper()
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.mu.Unlock()

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, err := c.in.Write(append(body, '\n')); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !c.scanner.Scan() {
		t.Fatalf("no response (server exited?): %v", c.scanner.Err())
	}
	var resp pendingResponse
	if err := json.Unmarshal(c.scanner.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v\nbody: %s", err, c.scanner.Text())
	}
	return &resp
}

// callTool wraps a tools/call invocation, decoding the text-content envelope
// back into a structured payload. Returns (resp, isError).
func (c *mcpClient) callTool(t *testing.T, name string, args any) (json.RawMessage, bool) {
	t.Helper()
	resp := c.call(t, "tools/call", map[string]any{"name": name, "arguments": args})
	if resp.Error != nil {
		t.Fatalf("rpc error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}
	var env struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &env); err != nil {
		t.Fatalf("decode tool envelope: %v\nresult: %s", err, string(resp.Result))
	}
	if len(env.Content) == 0 {
		t.Fatal("empty content")
	}
	return json.RawMessage(env.Content[0].Text), env.IsError
}

// newTempDataDir mirrors newFixtureWithTempDataDir's seeding for use without
// the HTTP server.
func newTempDataDir(t *testing.T, layouts, configs []string) string {
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
	return tmp
}

// ---------- protocol-level tests ----------

func TestMCPInitialize(t *testing.T) {
	c := startMCP(t, nil, nil)
	resp := c.call(t, "initialize", map[string]any{})
	if resp.Error != nil {
		t.Fatalf("initialize error: %v", resp.Error)
	}
	var body struct {
		ProtocolVersion string         `json:"protocolVersion"`
		ServerInfo      map[string]any `json:"serverInfo"`
		Capabilities    map[string]any `json:"capabilities"`
	}
	if err := json.Unmarshal(resp.Result, &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ProtocolVersion == "" {
		t.Error("empty protocolVersion")
	}
	if body.ServerInfo["name"] != "keycraft-mcp" {
		t.Errorf("server name: %v", body.ServerInfo["name"])
	}
}

func TestMCPToolsList(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})
	resp := c.call(t, "tools/list", nil)

	var body struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{
		"list_corpora", "describe_corpus",
		"list_layouts", "load_layout", "flip_layout", "save_layout",
		"analyse", "rank",
		"optimise_fast", "generate_fast",
	}
	got := map[string]bool{}
	for _, x := range body.Tools {
		got[x.Name] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing tool: %s", w)
		}
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	c := startMCP(t, nil, nil)
	resp := c.call(t, "no/such/method", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("got code %d, want -32601", resp.Error.Code)
	}
}

// ---------- tool tests ----------

func TestMCPListLayouts(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "list_layouts", map[string]any{})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var out struct {
		Layouts []string `json:"layouts"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Layouts) == 0 {
		t.Fatal("expected at least one layout")
	}
}

func TestMCPLoadLayout(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "load_layout", map[string]any{"name": "qwerty"})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var dto struct {
		Name string   `json:"name"`
		Rows []string `json:"rows"`
	}
	if err := json.Unmarshal(body, &dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dto.Name != "qwerty" {
		t.Errorf("name: %q", dto.Name)
	}
	if len(dto.Rows) != 3 {
		t.Errorf("rows: %d", len(dto.Rows))
	}
}

func TestMCPLoadLayoutMissing(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "load_layout", map[string]any{"name": "no-such-layout"})
	if !isErr {
		t.Fatalf("expected isError=true, got body: %s", string(body))
	}
}

func TestMCPFlipLayout(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "flip_layout", map[string]any{"name": "qwerty"})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var resp struct {
		Layout struct {
			Name string `json:"name"`
		} `json:"layout"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Layout.Name != "qwerty-flipped" {
		t.Errorf("got %q", resp.Layout.Name)
	}
}

func TestMCPSaveLayout(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "save_layout", map[string]any{
		"source": "qwerty", "dest": "qwerty-flip-saved", "flip": true,
	})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	if _, err := os.Stat(filepath.Join(c.dataDir, "layouts", "qwerty-flip-saved.klf")); err != nil {
		t.Errorf("expected file on disk: %v", err)
	}
}

func TestMCPSaveLayoutRefusesPath(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	// Path traversal must be refused.
	_, isErr := c.callTool(t, "save_layout", map[string]any{
		"source": "qwerty", "dest": "../escape.klf",
	})
	if !isErr {
		t.Error("expected isError=true for path traversal in dest")
	}
}

func TestMCPAnalyse(t *testing.T) {
	c := startMCP(t, nil, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "analyse", map[string]any{
		"layouts": []string{"qwerty", "colemak"},
		"corpus":  map[string]any{"name": "monkeyracer.txt", "coverage": 98.0},
	})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var resp struct {
		Results []struct {
			Name    string             `json:"name"`
			Metrics map[string]float64 `json:"metrics"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("results: %d", len(resp.Results))
	}
}

func TestMCPOptimiseFast(t *testing.T) {
	c := startMCP(t, []string{"qwerty"}, nil)
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "optimise_fast", map[string]any{
		"layout": "qwerty",
		"corpus": map[string]any{"name": "monkeyracer.txt", "coverage": 98.0},
		"seed":   42,
	})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var resp struct {
		Best struct {
			Name string `json:"name"`
		} `json:"best"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Best.Name != "qwerty-opt" {
		t.Errorf("best name: %q", resp.Best.Name)
	}
}

func TestMCPGenerateFast(t *testing.T) {
	c := startMCP(t, nil, []string{"vowels.gen"})
	c.call(t, "initialize", map[string]any{})

	body, isErr := c.callTool(t, "generate_fast", map[string]any{
		"config":      "vowels.gen",
		"max_layouts": 3,
		"seed":        42,
	})
	if isErr {
		t.Fatalf("isError=true: %s", string(body))
	}
	var resp struct {
		Layouts   []string `json:"layouts"`
		Generated int      `json:"generated"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Generated == 0 {
		t.Fatal("no layouts generated")
	}
}

// TestMCPDataDirIsolation verifies the --data-dir flag actually controls
// where layouts come from — the same MCP binary launched against a *different*
// data dir sees a different layout list.
func TestMCPDataDirIsolation(t *testing.T) {
	c1 := startMCP(t, nil, nil)
	c1.call(t, "initialize", map[string]any{})
	body1, _ := c1.callTool(t, "list_layouts", map[string]any{})

	c2 := startMCP(t, []string{"focal"}, nil) // different data dir
	c2.call(t, "initialize", map[string]any{})
	body2, _ := c2.callTool(t, "list_layouts", map[string]any{})

	// c2 has focal.klf, c1 doesn't
	if !strings.Contains(string(body2), "focal.klf") {
		t.Error("expected focal.klf in c2 layouts")
	}
	if strings.Contains(string(body1), "focal.klf") {
		t.Error("did not expect focal.klf in c1 layouts")
	}
}

