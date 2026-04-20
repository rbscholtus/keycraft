package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// Service is the transport-agnostic facade exposing keycraft operations to
// HTTP, MCP, or Wails adapters. All methods accept primitive/DTO inputs and
// return DTOs.
type Service struct {
	mgr *SessionManager
}

// NewService wires a Service on top of a SessionManager.
func NewService(mgr *SessionManager) *Service {
	return &Service{mgr: mgr}
}

// Manager exposes the underlying SessionManager (for tests / diagnostics).
func (s *Service) Manager() *SessionManager { return s.mgr }

// OpenSession creates a new session and returns its DTO.
func (s *Service) OpenSession() (SessionDTO, error) {
	sess, err := s.mgr.CreateSession()
	if err != nil {
		return SessionDTO{}, err
	}
	return SessionDTO{ID: sess.ID}, nil
}

// CloseSession removes the named session.
func (s *Service) CloseSession(id string) error {
	s.mgr.CloseSession(id)
	return nil
}

// ListCorpora returns filenames in the configured corpus directory that look
// like corpora (".txt" files).
func (s *Service) ListCorpora() ([]string, error) {
	entries, err := os.ReadDir(s.mgr.cfg.CorpusDir())
	if err != nil {
		return nil, fmt.Errorf("read corpus dir: %w", err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".txt") {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// DescribeCorpus loads (or reuses cached) the corpus for the given session
// and returns a summary. The cache reference is released once we have the
// snapshot — future calls re-load from cache instantly.
func (s *Service) DescribeCorpus(sessionID string, ref CorpusRef) (CorpusSummaryDTO, error) {
	if _, err := s.mgr.GetSession(sessionID); err != nil {
		return CorpusSummaryDTO{}, err
	}
	corpus, err := s.mgr.LoadCorpus(ref)
	if err != nil {
		return CorpusSummaryDTO{}, err
	}
	defer s.mgr.ReleaseCorpus(ref)

	return corpusSummary(corpus, ref.Coverage), nil
}

// ListLayouts returns .klf filenames in the configured layouts directory.
func (s *Service) ListLayouts() ([]string, error) {
	entries, err := os.ReadDir(s.mgr.cfg.LayoutDir())
	if err != nil {
		return nil, fmt.Errorf("read layout dir: %w", err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".klf") {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// GetLayout loads a single layout by name (with or without .klf suffix).
func (s *Service) GetLayout(name string) (LayoutDTO, error) {
	layout, err := s.loadLayout(name)
	if err != nil {
		return LayoutDTO{}, err
	}
	return layoutToDTO(layout), nil
}

// FlipLayout returns the horizontally flipped layout in-memory. The caller
// decides whether/where to persist it.
func (s *Service) FlipLayout(name string) (FlipResponse, error) {
	path, err := s.resolveLayoutPath(name)
	if err != nil {
		return FlipResponse{}, err
	}
	res, err := kc.FlipLayout(kc.FlipInput{LayoutFile: path})
	if err != nil {
		return FlipResponse{}, err
	}
	return FlipResponse{Layout: layoutToDTO(res.Layout)}, nil
}

// Analyse computes detailed metrics for each requested layout.
func (s *Service) Analyse(sessionID string, req AnalyseRequest) (AnalyseResponse, error) {
	if _, err := s.mgr.GetSession(sessionID); err != nil {
		return AnalyseResponse{}, err
	}
	if len(req.Layouts) == 0 {
		return AnalyseResponse{}, fmt.Errorf("%w: at least one layout required", ErrInvalidInput)
	}

	corpus, err := s.mgr.LoadCorpus(req.Corpus)
	if err != nil {
		return AnalyseResponse{}, err
	}
	defer s.mgr.ReleaseCorpus(req.Corpus)

	targets, err := s.buildTargets(req.Targets)
	if err != nil {
		return AnalyseResponse{}, err
	}

	layoutFiles, err := s.resolveLayoutPaths(req.Layouts)
	if err != nil {
		return AnalyseResponse{}, err
	}

	result, err := kc.AnalyseLayouts(kc.AnalyseInput{
		LayoutFiles: layoutFiles,
		Corpus:      corpus,
		TargetLoads: targets,
	})
	if err != nil {
		return AnalyseResponse{}, err
	}

	resp := AnalyseResponse{Results: make([]LayoutMetricsDTO, 0, len(result.Analysers))}
	for _, a := range result.Analysers {
		resp.Results = append(resp.Results, LayoutMetricsDTO{
			Name:    a.Layout.Name,
			Metrics: a.Metrics,
		})
	}
	return resp, nil
}

// View is the lighter summary variant; shares the AnalyseRequest/Response
// shape and reuses the analyse path.
func (s *Service) View(sessionID string, req ViewRequest) (ViewResponse, error) {
	return s.Analyse(sessionID, req)
}

// Rank computes weighted scoring across the supplied layouts.
func (s *Service) Rank(sessionID string, req RankRequest) (RankResponse, error) {
	if _, err := s.mgr.GetSession(sessionID); err != nil {
		return RankResponse{}, err
	}
	if len(req.Layouts) == 0 {
		return RankResponse{}, fmt.Errorf("%w: at least one layout required", ErrInvalidInput)
	}

	corpus, err := s.mgr.LoadCorpus(req.Corpus)
	if err != nil {
		return RankResponse{}, err
	}
	defer s.mgr.ReleaseCorpus(req.Corpus)

	targets, err := s.buildTargets(req.Targets)
	if err != nil {
		return RankResponse{}, err
	}

	weights, err := s.buildWeights(req.Weights)
	if err != nil {
		return RankResponse{}, err
	}

	layoutFiles, err := s.resolveLayoutPaths(req.Layouts)
	if err != nil {
		return RankResponse{}, err
	}

	result, err := kc.ComputeRankings(kc.RankingInput{
		LayoutsDir:  s.mgr.cfg.LayoutDir(),
		LayoutFiles: layoutFiles,
		Corpus:      corpus,
		Targets:     targets,
		Weights:     weights,
	})
	if err != nil {
		return RankResponse{}, err
	}

	resp := RankResponse{
		Scores:  make([]LayoutScoreDTO, 0, len(result.Scores)),
		Medians: result.Medians,
		IQRs:    result.IQRs,
	}
	for _, sc := range result.Scores {
		resp.Scores = append(resp.Scores, LayoutScoreDTO{
			Name:    sc.Name,
			Score:   sc.Score,
			Metrics: sc.Analyser.Metrics,
		})
	}
	return resp, nil
}

// -------- helpers --------

func (s *Service) loadLayout(name string) (*kc.SplitLayout, error) {
	path, err := s.resolveLayoutPath(name)
	if err != nil {
		return nil, err
	}
	layoutName := trimExt(filepath.Base(path))
	return kc.NewLayoutFromFile(layoutName, path)
}

func (s *Service) resolveLayoutPath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("%w: layout name is required", ErrInvalidInput)
	}
	if !strings.EqualFold(filepath.Ext(name), ".klf") {
		name = name + ".klf"
	}
	path := filepath.Join(s.mgr.cfg.LayoutDir(), name)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: layout %q", ErrNotFound, name)
		}
		return "", fmt.Errorf("stat layout: %w", err)
	}
	return path, nil
}

func (s *Service) resolveLayoutPaths(names []string) ([]string, error) {
	paths := make([]string, 0, len(names))
	for _, n := range names {
		p, err := s.resolveLayoutPath(n)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (s *Service) buildTargets(dto TargetLoadsDTO) (*kc.TargetLoads, error) {
	var targets *kc.TargetLoads
	var err error

	if dto.LoadTargetsFile == "" {
		targets = kc.NewTargetLoads()
	} else {
		path := filepath.Join(s.mgr.cfg.ConfigDir(), dto.LoadTargetsFile)
		targets, err = kc.NewTargetLoadsFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("load targets: %w", err)
		}
	}

	if dto.HandLoad != "" {
		if err := targets.SetHandLoad(dto.HandLoad); err != nil {
			return nil, fmt.Errorf("hand load: %w", err)
		}
	}
	if dto.FingerLoad != "" {
		if err := targets.SetFingerLoad(dto.FingerLoad); err != nil {
			return nil, fmt.Errorf("finger load: %w", err)
		}
	}
	if dto.RowLoad != "" {
		if err := targets.SetRowLoad(dto.RowLoad); err != nil {
			return nil, fmt.Errorf("row load: %w", err)
		}
	}
	if dto.PinkyPenalties != "" {
		if err := targets.SetPinkyPenalties(dto.PinkyPenalties); err != nil {
			return nil, fmt.Errorf("pinky penalties: %w", err)
		}
	}

	return targets, nil
}

func (s *Service) buildWeights(dto WeightsDTO) (*kc.Weights, error) {
	path := dto.WeightsFile
	if path != "" {
		path = filepath.Join(s.mgr.cfg.ConfigDir(), path)
	}
	w, err := kc.NewWeightsFromParams(path, dto.Weights)
	if err != nil {
		return nil, fmt.Errorf("load weights: %w", err)
	}
	return w, nil
}

func layoutToDTO(l *kc.SplitLayout) LayoutDTO {
	rows := make([]string, 3)
	for row := range 3 {
		rows[row] = runesToString(l.Runes[row*12 : row*12+12])
	}
	thumbs := runesToString(l.Runes[36:42])
	return LayoutDTO{
		Name:       l.Name,
		LayoutType: kc.LayoutTypeStrings[l.LayoutType],
		Rows:       rows,
		ThumbRow:   thumbs,
	}
}

func runesToString(rs []rune) string {
	out := make([]rune, len(rs))
	for i, r := range rs {
		if r == 0 {
			out[i] = '~'
		} else {
			out[i] = r
		}
	}
	return string(out)
}

func corpusSummary(c *kc.Corpus, coverage float64) CorpusSummaryDTO {
	return CorpusSummaryDTO{
		Name:                c.Name,
		Coverage:            coverage,
		UniqueUnigrams:      len(c.Unigrams),
		UniqueBigrams:       len(c.Bigrams),
		UniqueTrigrams:      len(c.Trigrams),
		UniqueSkipgrams:     len(c.Skipgrams),
		UniqueWords:         len(c.Words),
		TotalUnigramsCount:  c.TotalUnigramsCount,
		TotalBigramsCount:   c.TotalBigramsCount,
		TotalTrigramsCount:  c.TotalTrigramsCount,
		TotalSkipgramsCount: c.TotalSkipgramsCount,
		TotalWordsCount:     c.TotalWordsCount,
	}
}
