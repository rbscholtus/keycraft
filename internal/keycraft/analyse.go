package keycraft

import (
	"path/filepath"
	"strings"
)

// AnalyseInput contains parameters needed for layout analysis computation.
// This is pure computational input - no display/rendering concerns.
type AnalyseInput struct {
	LayoutFiles []string     // Full filepaths to layout files to analyse
	Corpus      *Corpus      // Text corpus for analysis
	TargetLoads *TargetLoads // User target loads
}

// AnalyseResult contains the computational results of layout analysis.
// Display-agnostic - just the data.
type AnalyseResult struct {
	Analysers []*Analyser // Analysis results for each layout
}

// AnalyseDisplayOptions contains rendering/display preferences.
// These don't affect the computation, only how results are presented.
type AnalyseDisplayOptions struct {
	MaxRows         int  // Maximum rows to show in detail tables
	CompactTrigrams bool // Whether to use compact trigram display
	TrigramRows     int  // Number of trigram rows to display
}

// AnalyseLayouts performs detailed layout analysis.
// Pure computation - no I/O, no rendering, no display logic.
func AnalyseLayouts(input AnalyseInput) (*AnalyseResult, error) {
	analysers := make([]*Analyser, 0, len(input.LayoutFiles))
	for _, path := range input.LayoutFiles {
		// Extract layout name from filename (remove directory and extension)
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		layout, err := NewLayoutFromFile(name, path)
		if err != nil {
			return nil, err
		}
		analyser := NewAnalyser(layout, input.Corpus, input.TargetLoads)
		analysers = append(analysers, analyser)
	}

	return &AnalyseResult{
		Analysers: analysers,
	}, nil
}
