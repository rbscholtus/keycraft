package keycraft

import (
	"fmt"
	"io"
)

// OptimizeInput encapsulates parameters for BLS optimization.
type OptimizeInput struct {
	Layout          *SplitLayout
	LayoutsDir      string
	Corpus          *Corpus
	Targets         *TargetLoads
	Weights         *Weights
	Pinned          *PinnedKeys
	NumGenerations  int
	MaxTime         int // minutes
	Seed            int64
	LogFile         io.Writer
	Medians         map[string]float64 // Optional: pre-computed filtered medians (skip LoadAnalysers)
	IQRs            map[string]float64 // Optional: pre-computed filtered IQRs (skip LoadAnalysers)
	FilteredWeights map[string]float64 // Optional: pre-computed filtered weights (used with Medians/IQRs)
	UseParallel     bool               // Enable parallel evaluation in BLS steepest descent
}

// OptimizeResult contains optimization results.
type OptimizeResult struct {
	OriginalLayout *SplitLayout
	BestLayout     *SplitLayout
}

// OptimizeLayout performs BLS optimization.
// This is the pure computation function that doesn't handle I/O or rendering.
func OptimizeLayout(input OptimizeInput, consoleWriter io.Writer) (*OptimizeResult, error) {
	best, err := OptimizeLayoutBLS(input, consoleWriter)
	if err != nil {
		return nil, fmt.Errorf("could not optimize layout: %w", err)
	}

	return &OptimizeResult{
		OriginalLayout: input.Layout,
		BestLayout:     best,
	}, nil
}
