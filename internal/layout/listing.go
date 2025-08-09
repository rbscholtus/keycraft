// listing.go provides functionality to evaluate and rank keyboard layouts
// based on various ergonomic and usage metrics. It calculates penalty scores
// that help determine the relative quality of each layout.

package layout

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Weights holds metric weights used for scoring layouts.
// Metrics not explicitly set in the input string default to predefined values.
// ALT, ROL, and ONE have default negative weights since they represent positive aspects.
type Weights struct {
	weights map[string]float64
}

// NewWeightsFromString parses a string of metric-weight pairs into a Weights struct.
// Input format: "metric1=value1,metric2=value2,...", case-insensitive.
// Returns an error if the format is invalid or weights cannot be parsed.
func NewWeightsFromString(weightsStr string) (*Weights, error) {
	weights := map[string]float64{
		"ALT": -1,
		"ROL": -1,
		"ONE": -1,
	}

	if weightsStr == "" {
		return &Weights{weights}, nil
	}
	weightsStr = strings.ToUpper(strings.TrimSpace(weightsStr))

	for pair := range strings.SplitSeq(weightsStr, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weights format: %s", pair)
		}
		metric := strings.TrimSpace(parts[0])
		weight, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid weight value for metric %s", metric)
		}
		weights[metric] = weight
	}

	return &Weights{weights}, nil
}

// Get returns the weight assigned to a given metric.
// Returns 1.0 if the metric is not found in the weights map.
func (w *Weights) Get(metric string) float64 {
	if val, ok := w.weights[metric]; ok {
		return val
	}
	return 1.0
}

// MetricNames lists the metrics included in the layout ranking calculations.
var MetricNames = []string{"SFB", "LSB", "FSB", "HSB", "SFS", "LSS", "FSS", "HSS", "ALT", "ROL", "ONE", "RED"}

// ReversedMetrics defines metrics where a higher value is considered better,
// and thus their color coding for deltas is reversed.
var ReversedMetrics = map[string]bool{"ALT": true, "ROL": true, "ONE": true}

// LayoutScore represents a keyboard layout's name, its total penalty score,
// and the analyser that computed its metrics.
type LayoutScore struct {
	Name     string    // Layout identifier or filename.
	Penalty  float64   // Weighted penalty score for ranking.
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

// computeScores calculates weighted penalty scores for each layout.
// Metrics are normalized by subtracting the median and dividing by the IQR.
// The weighted sum of these normalized metrics produces the layout's penalty score.
func computeScores(analysers []*Analyser, medians, iqr map[string]float64, weights *Weights) []LayoutScore {
	var layoutScores []LayoutScore

	for _, analyser := range analysers {
		penalty := 0.0
		for metric, value := range analyser.Metrics {
			weight := weights.Get(metric)
			var scaledValue float64
			if iqr[metric] == 0 {
				scaledValue = 0
			} else {
				scaledValue = (value - medians[metric]) / iqr[metric]
			}
			penalty += weight * scaledValue
		}
		layoutScores = append(layoutScores, LayoutScore{analyser.Layout.Name, penalty, analyser})
	}

	return layoutScores
}

// renderTable prints the ranked layout scores as a formatted table.
// Includes layout names, penalty scores, individual metric values,
// and optionally, delta rows showing changes compared to the previous layout.
func renderTable(scores []LayoutScore, weights *Weights, showDeltas bool, title string) {
	tw := table.NewWriter()
	tw.SetTitle(title)
	tw.Style().Title.Align = text.AlignCenter

	// Configure columns: index, name, penalty, then each metric
	colConfigs := []table.ColumnConfig{
		{Name: "Index", Align: text.AlignRight},
		{Name: "Name", Align: text.AlignLeft},
		{Name: "Penalty", Align: text.AlignRight},
	}
	for _, metric := range MetricNames {
		colConfigs = append(colConfigs, table.ColumnConfig{Name: metric, Align: text.AlignRight})
	}
	tw.SetColumnConfigs(colConfigs)

	// Header row with column names
	header := table.Row{"#", "Name", "Penalty"}
	for _, metric := range MetricNames {
		header = append(header, metric)
	}
	tw.AppendHeader(header)

	// Weight row displayed after header with blank index and penalty columns
	weightRow := table.Row{"", "Weight", ""}
	for _, metric := range MetricNames {
		weight := weights.Get(metric)
		weightRow = append(weightRow, fmt.Sprintf("%.2f", weight))
	}
	tw.AppendHeader(weightRow)

	var prevMetrics []float64
	index := 1

	for i, score := range scores {
		// Main row for each layout
		row := table.Row{index, score.Name, fmt.Sprintf("%+.2f", score.Penalty)}

		currMetrics := make([]float64, 0, len(MetricNames))
		for _, metric := range MetricNames {
			val, ok := score.Analyser.Metrics[metric]
			if !ok {
				val = 0.0
			}
			row = append(row, fmt.Sprintf("%.2f%%", val))
			currMetrics = append(currMetrics, val)
		}

		// Append delta row if enabled and not the first layout
		if showDeltas && i > 0 {
			deltaRow := table.Row{"", "", ""}
			for idx, currVal := range currMetrics {
				metric := MetricNames[idx]
				delta := currVal - prevMetrics[idx]
				deltaRow = append(deltaRow, formatDelta(metric, delta))
			}
			tw.AppendRow(deltaRow)
		}

		tw.AppendRow(row)
		prevMetrics = currMetrics
		index++
	}

	fmt.Println(tw.Render())
}

// formatDelta returns a color-coded string for the delta value of a metric.
// For metrics flagged as reversed, the colors for positive and negative deltas are swapped.
func formatDelta(metric string, delta float64) string {
	reversed := ReversedMetrics[metric]
	var c text.Color

	switch {
	case delta > 0:
		c = IfThen(reversed, text.FgGreen, text.FgRed)
	case delta < 0:
		c = IfThen(reversed, text.FgRed, text.FgGreen)
	default:
		c = text.Reset
	}
	return c.Sprintf("%+.2f%%", delta)
}

// DoLayoutList evaluates and ranks keyboard layouts based on the given parameters.
// It loads all layouts from the specified directory, computes normalization stats,
// filters layouts if a subset is specified, computes weighted penalty scores,
// optionally sorts by rank, and renders the results including delta rows if requested.
// Parameters:
// - corpus: text corpus used for layout analysis.
// - layoutsDir: directory containing keyboard layout files (.klf).
// - weights: metric weights to apply during scoring.
// - style: analysis style parameter.
// - layouts: list of layout names to include; if empty, no filtering is applied.
// - orderOption: if "rank", results are sorted by penalty score ascending.
// - showDeltas: whether to show delta rows between layouts in the output.
func DoLayoutList(corpus *Corpus, layoutsDir string, weights *Weights, style string, layouts []string, orderOption string, showDeltas bool) error {
	// Load all analysers for layouts in the directory
	analysers, err := loadAnalysers(layoutsDir, corpus, style)
	if err != nil {
		return err
	}

	// Compute median and IQR for each metric for normalization
	medians, iqrs := computeMediansAndIQR(analysers)

	// Map analysers by layout name for quick lookup
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

	// Compute weighted penalty scores for filtered layouts
	layoutScores := computeScores(filteredAnalysers, medians, iqrs, weights)

	// Sort layouts by penalty if requested
	if orderOption == "rank" {
		sort.Slice(layoutScores, func(i, j int) bool {
			return layoutScores[i].Penalty < layoutScores[j].Penalty
		})
	}

	// Render results in a formatted table
	renderTable(layoutScores, weights, showDeltas, "Layout List")

	return nil
}
