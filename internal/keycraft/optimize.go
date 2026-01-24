package keycraft

import (
	"fmt"
	"io"
)

// OptimizeInput encapsulates parameters for BLS optimization.
type OptimizeInput struct {
	Layout         *SplitLayout
	LayoutsDir     string
	Corpus         *Corpus
	Targets        *TargetLoads
	Weights        *Weights
	Pinned         *PinnedKeys
	NumGenerations int
	MaxTime        int // minutes
	Seed           int64
	LogFile        io.Writer
}

// OptimizeResult contains optimization results.
type OptimizeResult struct {
	OriginalLayout *SplitLayout
	BestLayout     *SplitLayout
}

// OptimizeLayout performs BLS optimization.
// This is the pure computation function that doesn't handle I/O or rendering.
func OptimizeLayout(input OptimizeInput, consoleWriter io.Writer) (*OptimizeResult, error) {
	// Run optimization
	best, err := OptimizeLayoutBLS(
		input.Layout,
		input.LayoutsDir,
		input.Corpus,
		input.Weights,
		input.Targets,
		input.Pinned,
		input.NumGenerations,
		input.MaxTime,
		input.Seed,
		consoleWriter,
		input.LogFile,
	)
	if err != nil {
		return nil, fmt.Errorf("could not optimize layout: %w", err)
	}

	return &OptimizeResult{
		OriginalLayout: input.Layout,
		BestLayout:     best,
	}, nil
}
