package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/rbscholtus/keycraft/internal/api"
)

// JSON-RPC 2.0 error codes (subset relevant to MCP).
const (
	codeParseError     = -32700
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// jsonRPCRequest is a parsed JSON-RPC 2.0 request frame.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // may be missing (notification)
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// server pairs a Service with the MCP wire protocol.
type server struct {
	svc       *api.Service
	sessionID string // implicit session — one per stdio connection
	tools     []toolDef

	writeMu sync.Mutex
}

func newServer(svc *api.Service) *server {
	s := &server{svc: svc}
	s.tools = s.buildTools()
	return s
}

// Run reads newline-delimited JSON-RPC requests from r, dispatches each, and
// writes the response (if any) to w. Blocks until r returns EOF. Notifications
// (no id) get no reply.
func (s *server) Run(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		req, err := parseRequest(line)
		if err != nil {
			s.writeErr(w, nil, codeParseError, err.Error(), nil)
			continue
		}
		s.handle(w, req)
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func parseRequest(line []byte) (*jsonRPCRequest, error) {
	var req jsonRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return nil, err
	}
	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("unsupported jsonrpc version: %q", req.JSONRPC)
	}
	if req.Method == "" {
		return nil, errors.New("missing method")
	}
	return &req, nil
}

func (s *server) handle(w io.Writer, req *jsonRPCRequest) {
	switch req.Method {
	case "initialize":
		s.writeOK(w, req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "keycraft-mcp", "version": "0.1.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		})
	case "notifications/initialized", "initialized":
		// Notifications: no response.
	case "tools/list":
		s.writeOK(w, req.ID, map[string]any{"tools": s.toolListPayload()})
	case "tools/call":
		s.handleToolsCall(w, req)
	case "ping":
		s.writeOK(w, req.ID, map[string]any{})
	default:
		// Notifications (no id) get silently ignored; requests get an error.
		if len(req.ID) > 0 && string(req.ID) != "null" {
			s.writeErr(w, req.ID, codeMethodNotFound, "method not found: "+req.Method, nil)
		}
	}
}

// handleToolsCall pulls {name, arguments} out of params, looks up the tool,
// invokes it, and wraps the structured result in MCP's content-block
// envelope.
func (s *server) handleToolsCall(w io.Writer, req *jsonRPCRequest) {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &call); err != nil {
		s.writeErr(w, req.ID, codeInvalidParams, "could not parse params: "+err.Error(), nil)
		return
	}

	tool := s.findTool(call.Name)
	if tool == nil {
		s.writeErr(w, req.ID, codeMethodNotFound, "unknown tool: "+call.Name, nil)
		return
	}

	if err := s.ensureSession(); err != nil {
		s.writeErr(w, req.ID, codeInternalError, "could not open session: "+err.Error(), nil)
		return
	}

	ctx := context.Background()
	result, err := tool.handler(ctx, s, call.Arguments)
	if err != nil {
		// Tool errors are reported as a successful tools/call with isError=true,
		// per MCP spec — that's how the LLM client learns of the failure.
		s.writeOK(w, req.ID, mcpToolError(err))
		return
	}

	s.writeOK(w, req.ID, mcpToolResult(result))
}

// ensureSession lazily creates the implicit per-connection session.
func (s *server) ensureSession() error {
	if s.sessionID != "" {
		return nil
	}
	dto, err := s.svc.OpenSession()
	if err != nil {
		return err
	}
	s.sessionID = dto.ID
	return nil
}

// mcpToolResult wraps a structured Go value into the MCP content envelope.
// The result is JSON-encoded and sent back as a single text content block,
// which is how MCP clients consume structured tool output.
func mcpToolResult(v any) map[string]any {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		body = []byte(fmt.Sprintf("internal: marshal failed: %v", err))
	}
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(body)},
		},
	}
}

// mcpToolError wraps an error into the MCP isError envelope.
func mcpToolError(err error) map[string]any {
	return map[string]any{
		"isError": true,
		"content": []map[string]any{
			{"type": "text", "text": err.Error()},
		},
	}
}

// ---------- write helpers ----------

func (s *server) writeOK(w io.Writer, id json.RawMessage, result any) {
	s.write(w, &jsonRPCResponse{JSONRPC: "2.0", ID: idOrNull(id), Result: result})
}

func (s *server) writeErr(w io.Writer, id json.RawMessage, code int, msg string, data any) {
	s.write(w, &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      idOrNull(id),
		Error:   &jsonRPCError{Code: code, Message: msg, Data: data},
	})
}

func (s *server) write(w io.Writer, resp *jsonRPCResponse) {
	if len(resp.ID) == 0 {
		return // notification — no response
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(resp)
}

func idOrNull(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return id
}
