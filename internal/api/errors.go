// Package api provides a transport-agnostic service layer for keycraft.
// It wraps the domain logic in internal/keycraft/ with session management,
// corpus caching, and typed request/response types suitable for HTTP, MCP,
// or Wails adapters.
package api

import "errors"

// Sentinel errors returned by Service methods. Transport adapters translate
// them into status codes or protocol-specific errors.
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionLimit    = errors.New("session limit reached")
	ErrBusy            = errors.New("session is busy with another run")
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidInput    = errors.New("invalid input")
)
