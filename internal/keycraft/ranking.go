package keycraft

import (
	"fmt"
	"path/filepath"
	"strings"
)

// RankingInput encapsulates all configuration for layout ranking computation.
// All layouts in LayoutsDir are analyzed for normalization, then filtered to LayoutFiles.
type RankingInput struct {
	LayoutsDir  string       // Used to load all layouts for calculating medians/IQRs for normalization
	LayoutFiles []string     // Full filepaths for specific layouts to rank.
	Corpus      *Corpus      // The corpus that ranking is based on
	Targets     *TargetLoads // Load targets (row, finger, pinky penalties)
	Weights     *Weights     // Metric weights for weighted scoring
}

// RankingResult provides ranked layouts with normalization statistics.
// Medians and IQRs are computed across all layouts for robust scaling.
type RankingResult struct {
	Scores  []LayoutScore      // Ranked layouts, not in sorted order
	Medians map[string]float64 // Median values for each metric across all layouts
	IQRs    map[string]float64 // Interquartile ranges for normalization
}

// ComputeRankings performs pure computation without I/O or rendering.
// It loads layouts, computes statistics, filters, scores, and returns results.
func ComputeRankings(input RankingInput) (*RankingResult, error) {
	// Load and analyze all layouts (needed for normalization even if we filter later)
	analysers, err := LoadAnalysers(input.LayoutsDir, input.Corpus, input.Targets)
	if err != nil {
		return nil, fmt.Errorf("could not load analysers: %w", err)
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
		// Extract layout name from full filepath (basename without extension)
		layoutName := strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname))
		analyser, ok := analyserMap[layoutName]
		if !ok {
			// // Layout not in the reference set, load it explicitly
			// layout, err := NewLayoutFromFile(layoutName, fname)
			// if err != nil {
			// 	return nil, fmt.Errorf("could not load layout %s: %v", fname, err)
			// }
			// analyser = NewAnalyser(layout, input.Corpus, input.Targets)
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
