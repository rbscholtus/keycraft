package keycraft

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// MetricsMap groups named metric sets used for different ranking views.
var MetricsMap = map[string][]string{
	"basic": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "2RL", "3RL", "RED", "RED-WEAK",
		"IN:OUT", "FBL", "POH", "FLW",
	},
	"extended": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "ALT-SFS", "ALT-OTH",
		"2RL", "2RL-IN", "2RL-OUT", "2RL-SFB",
		"3RL", "3RL-IN", "3RL-OUT", "3RL-SFS",
		"RED", "RED-WEAK", "RED-SFS", "RED-OTH",
		"IN:OUT", "FBL", "POH", "FLW",
	},
	"fingers": {
		"H0", "F0", "F1", "F2", "F3", "F4",
		"F5", "F6", "F7", "F8", "F9", "H1",
	},
}

// Weights holds metric weights used for scoring layouts.
// Metrics not explicitly set in the input string default to predefined values.
// ALT, ROL, and ONE have default negative weights since they represent positive aspects.
type Weights struct {
	weights map[string]float64
}

// DefaultMetrics contains built-in metric weights used as defaults when no custom weight is provided.
var DefaultMetrics = map[string]float64{
	"SFB": -1.0,
}

// NewWeights creates an empty Weights structure ready to be populated.
func NewWeights() *Weights {
	weights := make(map[string]float64)
	// maps.Copy(weights, DefaultMetrics)
	return &Weights{weights}
}

// NewWeightsFromString parses a comma-separated `metric=weight` string into a Weights instance.
// Returns an error if the format is invalid or weights cannot be parsed.
func NewWeightsFromString(weightsStr string) (*Weights, error) {
	w := Weights{}
	err := w.AddWeightsFromString(weightsStr)
	return &w, err
}

// NewWeightsFromParams constructs weights from an optional file and CLI string.
func NewWeightsFromParams(path, weightsStr string) (*Weights, error) {
	weights := NewWeights()

	// Load weights from a file if specified.
	if path != "" {
		if err := weights.AddWeightsFromFile(path); err != nil {
			return nil, err
		}
	}

	// Override or add weights from the --weights string flag.
	if err := weights.AddWeightsFromString(weightsStr); err != nil {
		return nil, fmt.Errorf("failed to parse weights: %v", err)
	}

	return weights, nil
}

// AddWeightsFromFile reads weights from a file (ignoring comments/blanks) and applies them to the receiver.
func (w *Weights) AddWeightsFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read weights file %q: %v", path, err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") && line != "" {
			if err := w.AddWeightsFromString(line); err != nil {
				return fmt.Errorf("failed to parse weights from file %q: %v", path, err)
			}
		}
	}
	return nil
}

// AddWeightsFromString parses and applies a comma-separated `metric=weight` string.
// If weightsStr is empty, returns the existing Weights unchanged.
func (w *Weights) AddWeightsFromString(weightsStr string) error {
	if weightsStr == "" {
		return nil
	}

	weightsStr = strings.ToUpper(strings.TrimSpace(weightsStr))
	for pair := range strings.SplitSeq(weightsStr, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid weights format: %s", pair)
		}
		metric := strings.TrimSpace(parts[0])
		weight, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return fmt.Errorf("invalid weight value for metric %s", metric)
		}
		w.weights[metric] = weight
	}

	return nil
}

// Get returns the weight for a metric or 0 if not present.
func (w *Weights) Get(metric string) float64 {
	if val, ok := w.weights[metric]; ok {
		return val
	}
	return 0.0
}

// LayoutScore represents a layout name, its computed score and the analyser providing metrics.
type LayoutScore struct {
	Name     string    // Layout identifier or filename.
	Score    float64   // Weighted score for ranking.
	Analyser *Analyser // Analyser with detailed metric values.
}

// LoadAnalysers loads and analyses all .klf layout files from a directory.
func LoadAnalysers(layoutsDir string, corpus *Corpus) ([]*Analyser, error) {
	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading layout files from %v: %v", layoutsDir, err)
	}

	var (
		analysers []*Analyser
		mu        sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, runtime.GOMAXPROCS(0)) // bound concurrency
	)

	for _, file := range layoutFiles {
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".klf") {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(f os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }() // release

			layoutPath := filepath.Join(layoutsDir, f.Name())
			layout, err := NewLayoutFromFile(f.Name(), layoutPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			analyser := NewAnalyser(layout, corpus)

			mu.Lock()
			analysers = append(analysers, analyser)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return analysers, nil
}

// computeMediansAndIQR computes median and IQR for each metric across analysers for normalization.
func computeMediansAndIQR(analysers []*Analyser) (map[string]float64, map[string]float64) {
	metrics := make(map[string][]float64)
	for _, analyser := range analysers {
		for metric, value := range analyser.Metrics {
			metrics[metric] = append(metrics[metric], value)
		}
	}

	medians := make(map[string]float64)
	iqr := make(map[string]float64)
	for metric, values := range metrics {
		sort.Float64s(values)
		medians[metric] = Median(values)
		q1, q3 := Quartiles(values)
		iqr[metric] = q3 - q1
	}

	return medians, iqr
}

// computeScores normalizes metrics by median/IQR and computes weighted layout scores.
func computeScores(analysers []*Analyser, medians, iqr map[string]float64, weights *Weights) []LayoutScore {
	var layoutScores []LayoutScore

	for _, analyser := range analysers {
		score := 0.0
		for metric, value := range analyser.Metrics {
			if iqr[metric] == 0 {
				continue
			}
			weight := weights.Get(metric)
			if weight == 0 {
				continue
			}
			scaledValue := (value - medians[metric]) / iqr[metric]
			score += weight * scaledValue
		}
		layoutScores = append(layoutScores, LayoutScore{
			Name:     analyser.Layout.Name,
			Score:    score,
			Analyser: analyser,
		})
	}

	return layoutScores
}

// renderTable formats and prints the ranking table including optional delta rows.
func renderTable(scores []LayoutScore, metrics []string, weights *Weights, deltas string, base *LayoutScore) {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.Style().Title.Align = text.AlignCenter
	if base == nil {
		tw.SetTitle("Layout Ranking")
	} else {
		tw.SetTitle(fmt.Sprintf("Layout Ranking (Compare to %s)", base.Name))
	}

	// Configure columns: index, name, score, then each metric
	colConfigs := []table.ColumnConfig{
		{Name: "Index", Align: text.AlignRight},
		{Name: "Name", Align: text.AlignLeft},
		{Name: "Score", Align: text.AlignRight},
	}
	for _, metric := range metrics {
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:        metric,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
	}
	tw.SetColumnConfigs(colConfigs)

	// Header row with column names
	header := table.Row{"#", "Name", "Score"}
	for _, metric := range metrics {
		header = append(header, metric)
	}
	tw.AppendHeader(header)

	// Weight row displayed after header with blank index and score columns
	weightRow := table.Row{"", "Weight", ""}
	for _, metric := range metrics {
		weight := weights.Get(metric)
		weightRow = append(weightRow, fmt.Sprintf("%.2f", weight))
	}
	tw.AppendHeader(weightRow)

	rowIdx := 1
	if base != nil {
		rowIdx -= 1 + slices.IndexFunc(scores, func(ls LayoutScore) bool { return ls.Name == base.Name })
	}
	var prevMetrics, refMetrics []float64
	if base != nil {
		refMetrics = make([]float64, 0, len(metrics))
		for _, metric := range metrics {
			val := WithDefault(base.Analyser.Metrics, metric, 0.0)
			refMetrics = append(refMetrics, val)
		}
	}

	for i, score := range scores {
		// one dataRow for each layout
		dataRow := table.Row{rowIdx, score.Name, fmt.Sprintf("%+.2f", score.Score)}
		currMetrics := make([]float64, 0, len(metrics))
		for _, metric := range metrics {
			val := WithDefault(score.Analyser.Metrics, metric, 0.0)
			dataRow = append(dataRow, formatMetricValue(metric, val))
			currMetrics = append(currMetrics, val)
		}

		// delta row if enabled and not the first layout
		if i > 0 && deltas != "none" {
			deltaRow := table.Row{"", "", ""}
			for idx, currMetric := range currMetrics {
				var delta float64
				if base == nil {
					delta = currMetric - prevMetrics[idx]
				} else if rowIdx <= 0 {
					delta = prevMetrics[idx] - refMetrics[idx]
				} else {
					delta = currMetric - refMetrics[idx]
				}
				deltaRow = append(deltaRow, formatDelta(metrics[idx], delta, weights))
			}
			tw.AppendRow(deltaRow)
		}
		tw.AppendRow(dataRow)

		prevMetrics = currMetrics
		rowIdx++
	}

	fmt.Println(tw.Render())
}

// Helper
func formatMetricValue(metric string, val float64) string {
	if metric == "IN:OUT" {
		return fmt.Sprintf("%.2f", val)
	}
	return fmt.Sprintf("%.2f%%", val)
}

// formatDelta formats the delta between metrics with color according to weight polarity.
// For metrics flagged as reversed, the colors for positive and negative deltas are swapped.
func formatDelta(metric string, delta float64, weights *Weights) string {
	positive := weights.Get(metric) >= 0
	var c text.Color

	switch {
	case delta >= 0.01:
		c = IfThen(positive, text.FgGreen, text.FgRed)
	case delta <= -0.01:
		c = IfThen(positive, text.FgRed, text.FgGreen)
	default:
		c = text.Reset
	}

	if metric == "IN:OUT" {
		return c.Sprintf("%.2f", delta)
	}
	return c.Sprintf("%+.2f%%", delta)
}

// DoLayoutRankings evaluates and ranks keyboard layouts based on the given parameters.
// It loads all layouts from the specified directory, computes normalization stats,
// filters layouts if a subset is specified, computes weighted scores,
// optionally sorts by rank, and renders the results including delta rows if requested.
// Parameters:
// - corpus: text corpus used for layout analysis.
// - layoutsDir: directory containing keyboard layout files (.klf).
// - layouts: list of layout names to include; if empty, no filtering is applied.
// - weights: metric weights to apply during scoring.
// - metricsSet: string indicating which metric set to use ("basic", "extended", or "fingers").
// - deltas: whether to show delta rows between layouts in the output.
func DoLayoutRankings(corpus *Corpus, layoutsDir string, layouts []string, weights *Weights, metricsSet string, deltas string) error {
	// Choose metric list based on metricsMode flag
	metrics, ok := MetricsMap[metricsSet]
	if !ok {
		opts := slices.Collect(maps.Keys(MetricsMap))
		return fmt.Errorf("invalid metrics mode %q; must be one of %v", metricsSet, opts)
	}

	// Load all analysers for layouts in the directory
	analysers, err := LoadAnalysers(layoutsDir, corpus)
	if err != nil {
		return err
	}

	// Map analysers by layout name for fast lookup
	analyserMap := make(map[string]*Analyser, len(analysers))
	for _, analyser := range analysers {
		analyserMap[analyser.Layout.Name] = analyser
	}

	// Filter analysers to only include specified layouts (if any)
	filteredAnalysers := make([]*Analyser, 0, len(layouts))
	for _, name := range layouts {
		if analyser, ok := analyserMap[name]; ok {
			filteredAnalysers = append(filteredAnalysers, analyser)
		}
	}

	// Compute median and IQR for each metric for normalization
	medians, iqrs := computeMediansAndIQR(analysers)
	// maps.Copy(medians, idealFingerLoad)

	// Compute weighted scores for filtered layouts
	layoutScores := computeScores(filteredAnalysers, medians, iqrs, weights)

	// Add a Median row if needed
	var base *LayoutScore
	if deltas == "median" {
		lss := computeScores([]*Analyser{{Layout: &SplitLayout{Name: "median"}, Metrics: medians}}, medians, iqrs, weights)
		layoutScores = append(layoutScores, lss[0])
	}

	// Sort layouts by rank
	sort.Slice(layoutScores, func(i, j int) bool {
		return layoutScores[i].Score > layoutScores[j].Score
	})

	// get a ptr to the layout we compare to if relevant
	if deltas != "none" && deltas != "rows" {
		if i := slices.IndexFunc(layoutScores, func(ls LayoutScore) bool { return ls.Name == deltas }); i < 0 {
			panic(fmt.Sprintf("can't find %s", deltas))
		} else {
			base = &layoutScores[i]
		}
	}

	// Render results in a formatted table
	renderTable(layoutScores, metrics, weights, deltas, base)

	return nil
}
