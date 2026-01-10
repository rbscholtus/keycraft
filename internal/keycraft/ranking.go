package keycraft

import (
	"fmt"
	"path/filepath"
	"strings"
)

// RankingInput encapsulates all configuration for layout ranking computation.
// All layouts in LayoutsDir are analyzed for normalization, then filtered to LayoutFiles.
type RankingInput struct {
	LayoutsDir     string
	LayoutFiles    []string      // Specific layouts to rank; if empty, all layouts in LayoutsDir
	Corpus         *Corpus
	IdealRowLoad   *[3]float64   // Target distribution for row balance metrics
	IdealFgrLoad   *[10]float64  // Target distribution for finger balance metrics
	PinkyPenalties *[12]float64  // Penalty weights for pinky off-home positions
	Weights        *Weights      // Metric weights for weighted scoring
}

// RankingResult provides ranked layouts with normalization statistics.
// Medians and IQRs are computed across all layouts for robust scaling.
type RankingResult struct {
	Scores  []LayoutScore
	Medians map[string]float64 // Median values for each metric across all layouts
	IQRs    map[string]float64 // Interquartile ranges for normalization
}

// ComputeRankings performs pure computation without I/O or rendering.
// It loads layouts, computes statistics, filters, scores, and returns results.
func ComputeRankings(input RankingInput) (*RankingResult, error) {
	// Load and analyze all layouts (needed for normalization even if we filter later)
	analysers, err := LoadAnalysers(input.LayoutsDir, input.Corpus, input.IdealRowLoad, input.IdealFgrLoad, input.PinkyPenalties)
	if err != nil {
		return nil, err
	}
	medians, iqrs := computeMediansAndIQR(analysers)

	// Build lookup map for filtering
	analyserMap := make(map[string]*Analyser, len(analysers))
	for _, analyser := range analysers {
		analyserMap[analyser.Layout.Name] = analyser
	}

	// Filter to requested layouts
	filteredAnalysers := make([]*Analyser, 0, len(input.LayoutFiles))
	for _, fname := range input.LayoutFiles {
		layoutName := strings.TrimSuffix(fname, filepath.Ext(fname))
		analyser, ok := analyserMap[layoutName]
		if !ok {
			return nil, fmt.Errorf("layout file %s was not found", fname)
		}
		filteredAnalysers = append(filteredAnalysers, analyser)
	}

	// Compute scores using normalized metrics
	layoutScores := computeScores(filteredAnalysers, medians, iqrs, input.Weights)

	return &RankingResult{
		Scores:  layoutScores,
		Medians: medians,
		IQRs:    iqrs,
	}, nil
}

// ComputeMedianScore creates a synthetic LayoutScore from median values.
// The score is always 0.0 because normalized median values are (median - median) / IQR = 0.
func ComputeMedianScore(medians map[string]float64, weights *Weights) LayoutScore {
	analyser := &Analyser{
		Layout:  &SplitLayout{Name: "median"},
		Metrics: medians,
	}

	// The median score is 0.0 by definition of robust normalization.
	// Each metric's normalized value is (median - median) / IQR = 0.
	return LayoutScore{
		Name:     "median",
		Score:    0.0,
		Analyser: analyser,
	}
}
