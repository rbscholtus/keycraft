package keycraft

// AnalyseInput contains parameters needed for layout analysis computation.
// This is pure computational input - no display/rendering concerns.
type AnalyseInput struct {
	LayoutFiles []string        // Layout filenames to analyse
	Corpus      *Corpus         // Text corpus for analysis
	Prefs       *PreferredLoads // User preferences for ideal loads
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
	for _, filename := range input.LayoutFiles {
		layout, err := NewLayoutFromFile(filename, "data/layouts/"+filename)
		if err != nil {
			return nil, err
		}
		analyser := NewAnalyser(layout, input.Corpus, input.Prefs)
		analysers = append(analysers, analyser)
	}

	return &AnalyseResult{
		Analysers: analysers,
	}, nil
}
