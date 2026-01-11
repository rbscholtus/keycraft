package keycraft

import (
	"io"
)

// OptimiseInput encapsulates parameters for BLS optimization.
type OptimiseInput struct {
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

// OptimiseResult contains optimization results.
type OptimiseResult struct {
	OriginalLayout *SplitLayout
	BestLayout     *SplitLayout
}

// OptimizeLayout performs BLS optimization.
// This is the pure computation function that doesn't handle I/O or rendering.
func OptimizeLayout(input OptimiseInput, consoleWriter io.Writer) (*OptimiseResult, error) {
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
		return nil, err
	}

	return &OptimiseResult{
		OriginalLayout: input.Layout,
		BestLayout:     best,
	}, nil
}
