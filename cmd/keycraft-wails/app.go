package main

import (
	"context"

	"github.com/rbscholtus/keycraft/internal/api"
)

// App is the Wails binding target. Each exported method becomes callable from
// the frontend via Wails' generated TypeScript bindings. The implementation
// is a thin adapter that:
//
//  1. Maintains an implicit per-window session (Wails apps are typically
//     single-window, so the session lifecycle matches the binary).
//  2. Forwards calls to api.Service so the same code path serves Wails,
//     keycraft-serve (HTTP), and keycraft-mcp (MCP).
//
// Long-running calls (OptimiseFast, GenerateFast) accept a context.Context
// from the Wails runtime, which is cancelled when the window closes.
type App struct {
	svc       *api.Service
	sessionID string
}

// NewApp constructs an App and lazily opens a session on first use.
func NewApp(svc *api.Service) *App {
	return &App{svc: svc}
}

// ensureSession is called by every method that needs a session ID.
func (a *App) ensureSession() error {
	if a.sessionID != "" {
		return nil
	}
	dto, err := a.svc.OpenSession()
	if err != nil {
		return err
	}
	a.sessionID = dto.ID
	return nil
}

// ListLayouts returns all .klf layout names in the data dir.
func (a *App) ListLayouts() ([]string, error) {
	return a.svc.ListLayouts()
}

// GetLayout returns one layout's grid.
func (a *App) GetLayout(name string) (api.LayoutDTO, error) {
	return a.svc.GetLayout(name)
}

// FlipLayout returns the horizontally flipped layout in memory.
func (a *App) FlipLayout(name string) (api.FlipResponse, error) {
	return a.svc.FlipLayout(name)
}

// ListCorpora returns all .txt corpora in the data dir.
func (a *App) ListCorpora() ([]string, error) {
	return a.svc.ListCorpora()
}

// DescribeCorpus loads (or reuses cached) a corpus and returns its summary.
func (a *App) DescribeCorpus(ref api.CorpusRef) (api.CorpusSummaryDTO, error) {
	if err := a.ensureSession(); err != nil {
		return api.CorpusSummaryDTO{}, err
	}
	return a.svc.DescribeCorpus(a.sessionID, ref)
}

// Analyse computes detailed metrics for the supplied layouts.
func (a *App) Analyse(req api.AnalyseRequest) (api.AnalyseResponse, error) {
	if err := a.ensureSession(); err != nil {
		return api.AnalyseResponse{}, err
	}
	return a.svc.Analyse(a.sessionID, req)
}

// Rank computes weighted scoring across the supplied layouts.
func (a *App) Rank(req api.RankRequest) (api.RankResponse, error) {
	if err := a.ensureSession(); err != nil {
		return api.RankResponse{}, err
	}
	return a.svc.Rank(a.sessionID, req)
}

// OptimiseFast runs BLS local-search on the named layout. ctx originates
// from the Wails runtime so window closure cancels the run.
func (a *App) OptimiseFast(ctx context.Context, req api.OptimiseFastRequest) (api.OptimiseFastResponse, error) {
	if err := a.ensureSession(); err != nil {
		return api.OptimiseFastResponse{}, err
	}
	return a.svc.OptimiseFast(ctx, a.sessionID, req)
}

// GenerateFast produces layouts from a .gen config without optimising.
func (a *App) GenerateFast(ctx context.Context, req api.GenerateFastRequest) (api.GenerateFastResponse, error) {
	if err := a.ensureSession(); err != nil {
		return api.GenerateFastResponse{}, err
	}
	return a.svc.GenerateFast(ctx, a.sessionID, req)
}
