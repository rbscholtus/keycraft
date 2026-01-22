package keycraft

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ViewInput contains parameters for viewing layout analysis.
// This is pure computational input - no display/rendering concerns.
type ViewInput struct {
	LayoutFiles []string     // Full filepaths to layout files to view
	Corpus      *Corpus      // Text corpus for analysis
	Targets     *TargetLoads // User target loads
}

// ViewResult contains the analysis results for viewing layouts.
// Display-agnostic - just the data.
type ViewResult struct {
	Analysers []*Analyser // Analysis results for each layout
}

// ViewLayouts performs layout analysis for viewing.
// Pure computation - no I/O, no rendering, no display logic.
func ViewLayouts(input ViewInput) (*ViewResult, error) {
	analysers := make([]*Analyser, 0, len(input.LayoutFiles))
	for _, path := range input.LayoutFiles {
		// Extract layout name from filename (remove directory and extension)
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		layout, err := NewLayoutFromFile(name, path)
		if err != nil {
			return nil, fmt.Errorf("could not create new layout from file: %w", err)
		}
		analyser := NewAnalyser(layout, input.Corpus, input.Targets)
		analysers = append(analysers, analyser)
	}

	return &ViewResult{
		Analysers: analysers,
	}, nil
}
