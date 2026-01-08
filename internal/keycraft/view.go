package keycraft

// ViewInput contains parameters for viewing layout analysis.
// This is pure computational input - no display/rendering concerns.
type ViewInput struct {
	LayoutFiles []string        // Layout filenames to view
	Corpus      *Corpus         // Text corpus for analysis
	Prefs       *PreferredLoads // User preferences for ideal loads
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
	for _, filename := range input.LayoutFiles {
		layout, err := NewLayoutFromFile(filename, "data/layouts/"+filename)
		if err != nil {
			return nil, err
		}
		analyser := NewAnalyser(layout, input.Corpus, input.Prefs)
		analysers = append(analysers, analyser)
	}

	return &ViewResult{
		Analysers: analysers,
	}, nil
}
