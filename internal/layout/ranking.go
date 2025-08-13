// listing.go provides functionality to evaluate and rank keyboard layouts
// based on various ergonomic and usage metrics. It calculates scores
// that help determine the relative quality of each layout.

package layout

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var MetricsMap = map[string][]string{
	"metrics": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "2RL", "3RL", "IN:OUT", "RED",
	},
	"ext-metrics": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "ALT-SFS", "ALT-OTH",
		"2RL", "2RL-IN", "2RL-OUT", "2RL-SF",
		"3RL", "3RL-IN", "3RL-OUT", "3RL-SF",
		"IN:OUT",
		"RED", "RED-BAD", "RED-SFS", "RED-OTH",
	},
	"columns": {
		"C0", "C1", "C2", "C3", "C4", "C5",
		"C6", "C7", "C8", "C9", "C10", "C11",
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

// DefaultMetrics defines metrics where a higher value is considered better,
// and thus their color coding for deltas is reversed.
var DefaultMetrics = map[string]float64{
	"SFB": -1.0,
}

// NewWeights creates a Weights struct with positive metrics from PositiveMetrics set to 1.0.
func NewWeights() *Weights {
	weights := make(map[string]float64)
	maps.Copy(weights, DefaultMetrics)
	return &Weights{weights}
}

// NewWeightsFromString parses a string of metric-weight pairs into a Weights struct.
// Input format: "metric1=value1,metric2=value2,...", case-insensitive.
// Returns an error if the format is invalid or weights cannot be parsed.
func NewWeightsFromString(weightsStr string) (*Weights, error) {
	w := NewWeights()
	err := w.AddWeightsFromString(weightsStr)
	return w, err
}

// AddWeightsFromString adds or overrides weights from a string of metric-weight pairs.
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

// Get returns the weight assigned to a given metric.
// Returns -1.0 if the metric is not found in the weights map.
func (w *Weights) Get(metric string) float64 {
	if val, ok := w.weights[metric]; ok {
		return val
	}
	return 0.0
}

// LayoutScore represents a keyboard layout's name, its total score,
// and the analyser that computed its metrics.
type LayoutScore struct {
	Name     string    // Layout identifier or filename.
	Score    float64   // Weighted score for ranking.
	Analyser *Analyser // Analyser with detailed metric values.
}

// loadAnalysers reads layout files from the given directory, creates analysers
// by evaluating each layout against the specified corpus and style, and returns them.
// Only files ending with ".klf" are considered.
func loadAnalysers(layoutsDir string, corpus *Corpus, style string) ([]*Analyser, error) {
	var analysers []*Analyser

	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading layout files from %v: %v", layoutsDir, err)
	}

	for _, file := range layoutFiles {
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".klf") {
			continue
		}
		layoutPath := filepath.Join(layoutsDir, file.Name())
		layout, err := NewLayoutFromFile(file.Name(), layoutPath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		analyser := NewAnalyser(layout, corpus, style)
		analysers = append(analysers, analyser)
	}

	return analysers, nil
}

// computeMediansAndIQR calculates the median and interquartile range (IQR)
// for each metric across all analysers. These statistics are used to normalize
// metric values during scoring.
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

// computeScores calculates weighted scores for each layout.
// Metrics are normalized by subtracting the median and dividing by the IQR.
// The weighted sum of these normalized metrics produces the layout's score.
func computeScores(analysers []*Analyser, medians, iqr map[string]float64, weights *Weights) []LayoutScore {
	var layoutScores []LayoutScore

	for _, analyser := range analysers {
		score := 0.0
		for metric, value := range analyser.Metrics {
			weight := weights.Get(metric)
			var scaledValue float64
			if iqr[metric] == 0 {
				scaledValue = 0
			} else {
				scaledValue = (value - medians[metric]) / iqr[metric]
			}
			score += weight * scaledValue
		}
		layoutScores = append(layoutScores, LayoutScore{analyser.Layout.Name, score, analyser})
	}

	return layoutScores
}

// renderTable prints the ranked layout scores as a formatted table.
// Includes layout names, scores, individual metric values,
// and optionally, delta rows showing changes compared to the previous layout.
func renderTable(scores []LayoutScore, metrics []string, weights *Weights, deltas string, base *LayoutScore) {
	tw := table.NewWriter()
	if base == nil {
		tw.SetTitle("Layout Ranking")
	} else {
		tw.SetTitle(fmt.Sprintf("Layout Ranking (Compare to %s)", base.Name))
	}
	tw.Style().Title.Align = text.AlignCenter

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

// formatDelta returns a color-coded string for the delta value of a metric.
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
// - metricsMode: string indicating which metric set to use ("metrics", "ext-metrics", "columns", or "fingers").
// - deltas: whether to show delta rows between layouts in the output.
func DoLayoutRankings(corpus *Corpus, layoutsDir string, layouts []string, weights *Weights, metricsSet string, deltas string) error {
	// Choose metric list based on metricsMode flag
	metrics, ok := MetricsMap[metricsSet]
	if !ok {
		opts := slices.Collect(maps.Keys(MetricsMap))
		return fmt.Errorf("invalid metrics mode %q; must be one of %v", metricsSet, opts)
	}

	// Load all analysers for layouts in the directory
	analysers, err := loadAnalysers(layoutsDir, corpus, "layoutsdoc")
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
	// godump.Dump(base.Name)
	renderTable(layoutScores, metrics, weights, deltas, base)

	return nil
}
