package api

// CorpusRef identifies a corpus by filename (within the data-dir's corpus/
// directory) and coverage percentage. Two refs with identical values share a
// cached *Corpus across sessions.
type CorpusRef struct {
	Name     string  `json:"name"`
	Coverage float64 `json:"coverage"`
}

// TargetLoadsDTO mirrors keycraft.TargetLoads settings a client can override.
// Empty strings fall back to defaults.
type TargetLoadsDTO struct {
	LoadTargetsFile string `json:"load_targets_file,omitempty"`
	HandLoad        string `json:"hand_load,omitempty"`
	FingerLoad      string `json:"finger_load,omitempty"`
	RowLoad         string `json:"row_load,omitempty"`
	PinkyPenalties  string `json:"pinky_penalties,omitempty"`
}

// WeightsDTO captures metric weight overrides for ranking/scoring.
// Either WeightsFile, Weights, or both may be provided; an empty DTO uses
// default weights.
type WeightsDTO struct {
	WeightsFile string `json:"weights_file,omitempty"`
	Weights     string `json:"weights,omitempty"`
}

// SessionDTO is the response body for POST /v1/sessions.
type SessionDTO struct {
	ID string `json:"id"`
}

// LayoutDTO is the wire representation of a SplitLayout: three main rows of
// 12 runes each plus a 6-rune thumb row.
type LayoutDTO struct {
	Name       string   `json:"name"`
	LayoutType string   `json:"layout_type"`
	Rows       []string `json:"rows"`
	ThumbRow   string   `json:"thumb_row"`
}

// CorpusSummaryDTO is the summary returned by corpus load/describe endpoints.
type CorpusSummaryDTO struct {
	Name                string `json:"name"`
	Coverage            float64 `json:"coverage"`
	UniqueUnigrams      int    `json:"unique_unigrams"`
	UniqueBigrams       int    `json:"unique_bigrams"`
	UniqueTrigrams      int    `json:"unique_trigrams"`
	UniqueSkipgrams     int    `json:"unique_skipgrams"`
	UniqueWords         int    `json:"unique_words"`
	TotalUnigramsCount  uint64 `json:"total_unigrams_count"`
	TotalBigramsCount   uint64 `json:"total_bigrams_count"`
	TotalTrigramsCount  uint64 `json:"total_trigrams_count"`
	TotalSkipgramsCount uint64 `json:"total_skipgrams_count"`
	TotalWordsCount     uint64 `json:"total_words_count"`
}

// AnalyseRequest selects layouts and a corpus for detailed analysis.
type AnalyseRequest struct {
	Layouts []string       `json:"layouts"`
	Corpus  CorpusRef      `json:"corpus"`
	Targets TargetLoadsDTO `json:"targets,omitempty"`
}

// LayoutMetricsDTO pairs a layout name with its computed metric map.
type LayoutMetricsDTO struct {
	Name    string             `json:"name"`
	Metrics map[string]float64 `json:"metrics"`
}

// AnalyseResponse returns per-layout metrics in the requested order.
type AnalyseResponse struct {
	Results []LayoutMetricsDTO `json:"results"`
}

// RankRequest asks for weighted scoring over a list of layouts.
type RankRequest struct {
	Layouts []string       `json:"layouts"`
	Corpus  CorpusRef      `json:"corpus"`
	Targets TargetLoadsDTO `json:"targets,omitempty"`
	Weights WeightsDTO     `json:"weights,omitempty"`
}

// LayoutScoreDTO is the ranked-score entry for one layout.
type LayoutScoreDTO struct {
	Name    string             `json:"name"`
	Score   float64            `json:"score"`
	Metrics map[string]float64 `json:"metrics"`
}

// RankResponse mirrors keycraft.RankingResult in DTO form.
type RankResponse struct {
	Scores  []LayoutScoreDTO   `json:"scores"`
	Medians map[string]float64 `json:"medians"`
	IQRs    map[string]float64 `json:"iqrs"`
}

// FlipResponse returns the flipped layout in-memory. The client chooses
// whether to persist it via a separate save endpoint/tool.
type FlipResponse struct {
	Layout LayoutDTO `json:"layout"`
}

// ViewRequest is the same shape as AnalyseRequest; distinct for clarity.
type ViewRequest = AnalyseRequest

// ViewResponse is the same shape as AnalyseResponse.
type ViewResponse = AnalyseResponse

// OptimiseFastRequest runs a single BLS steepest-descent pass on a loaded
// layout. Perturbation is skipped entirely, so the call returns quickly.
type OptimiseFastRequest struct {
	Layout   string         `json:"layout"`
	Corpus   CorpusRef      `json:"corpus"`
	Targets  TargetLoadsDTO `json:"targets,omitempty"`
	Weights  WeightsDTO     `json:"weights,omitempty"`
	PinsFile string         `json:"pins_file,omitempty"`
	Pins     string         `json:"pins,omitempty"`
	Free     string         `json:"free,omitempty"`
	Seed     int64          `json:"seed,omitempty"`
}

// OptimiseFastResponse carries the best layout plus its metric map.
type OptimiseFastResponse struct {
	Original LayoutDTO          `json:"original"`
	Best     LayoutDTO          `json:"best"`
	Metrics  map[string]float64 `json:"metrics"`
}

// GenerateFastRequest produces layouts from a .gen config WITHOUT an
// optimisation pass. Layouts are persisted under the data-dir's layouts
// subdirectory.
type GenerateFastRequest struct {
	Config     string `json:"config"`
	MaxLayouts int    `json:"max_layouts,omitempty"`
	Seed       uint64 `json:"seed,omitempty"`
}

// GenerateFastResponse reports which layouts were produced.
type GenerateFastResponse struct {
	Layouts    []string `json:"layouts"`
	TotalPerms int      `json:"total_perms"`
	Generated  int      `json:"generated"`
	Warnings   []string `json:"warnings,omitempty"`
}
