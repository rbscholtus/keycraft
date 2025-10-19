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
// Each metric set defines which columns appear in the ranking table output.
var MetricsMap = map[string][]string{
	"basic": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", // "LSS", "FSS", "HSS",
		"ALT", "2RL", "3RL", "RED", "RED-WEAK",
		"IN:OUT", "RBL", "FBL", "POH", "FLW",
	},
	"extended": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "ALT-NML", "ALT-SFS",
		"2RL", "2RL-IN", "2RL-OUT", "2RL-SFB",
		"3RL", "3RL-IN", "3RL-OUT", "3RL-SFB",
		"RED", "RED-NML", "RED-WEAK", "RED-SFS",
		"IN:OUT", "RBL", "FBL", "POH", "FLW",
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

// LoadAnalysers loads and analyses all .klf layout files from a directory in parallel.
// Uses bounded concurrency based on GOMAXPROCS to avoid overloading the system.
func LoadAnalysers(layoutsDir string, corpus *Corpus, idealRowLoad *[3]float64, idealfgrLoad *[10]float64) ([]*Analyser, error) {
	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading layout files from %v: %v", layoutsDir, err)
	}

	var (
		analysers = make([]*Analyser, 0, len(layoutFiles)) // Pre-allocate
		mu        sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, runtime.GOMAXPROCS(0)) // Semaphore to limit concurrent goroutines
	)

	for _, file := range layoutFiles {
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".klf") {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			layoutName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			layoutPath := filepath.Join(layoutsDir, f.Name())
			layout, err := NewLayoutFromFile(layoutName, layoutPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			analyser := NewAnalyser(layout, corpus, idealRowLoad, idealfgrLoad)

			mu.Lock()
			analysers = append(analysers, analyser)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return analysers, nil
}

// computeMediansAndIQR computes median and interquartile range (IQR) for each metric
// across all analysers. These values are used for robust normalization of layout scores.
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

// computeScores normalizes metrics using median and IQR, then computes weighted layout scores.
// Only metrics with non-zero weights are included in the final score calculation.
func computeScores(analysers []*Analyser, medians, iqr map[string]float64, weights *Weights) []LayoutScore {
	var layoutScores []LayoutScore

	for _, analyser := range analysers {
		score := 0.0
		for metric, value := range analyser.Metrics {
			// Skip metrics with zero IQR (all values identical)
			if iqr[metric] == 0 {
				continue
			}
			weight := weights.Get(metric)
			if weight == 0 {
				continue
			}
			// Apply robust normalization and weight
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

// renderTable formats and prints the ranking table with optional weight and delta rows.
// Delta rows show differences between consecutive layouts or compared to a base layout.
func renderTable(scores []LayoutScore, metrics []string, weights *Weights, showWeights bool, deltas string, base *LayoutScore) {
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

	// Configure column alignment
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

	// Build header row
	header := table.Row{"#", "Name", "Score"}
	for _, metric := range metrics {
		header = append(header, metric)
	}
	tw.AppendHeader(header)

	// Add weight row if requested
	if showWeights {
		weightRow := table.Row{"", "Weight", ""}
		for _, metric := range metrics {
			weight := weights.Get(metric)
			weightRow = append(weightRow, fmt.Sprintf("%.2f", weight))
		}
		tw.AppendHeader(weightRow)
	}

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
		// Build data row for this layout
		dataRow := table.Row{rowIdx, score.Name, fmt.Sprintf("%+.2f", score.Score)}
		currMetrics := make([]float64, 0, len(metrics))
		for _, metric := range metrics {
			val := WithDefault(score.Analyser.Metrics, metric, 0.0)
			dataRow = append(dataRow, formatMetricValue(metric, val))
			currMetrics = append(currMetrics, val)
		}

		// Add delta row showing differences from previous or base layout
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

// formatMetricValue formats a metric value for display.
// IN:OUT ratio is displayed as a plain number, others as percentages.
func formatMetricValue(metric string, val float64) string {
	if metric == "IN:OUT" {
		return fmt.Sprintf("%.2f", val)
	}
	return fmt.Sprintf("%.2f%%", val)
}

// formatDelta formats the delta between metrics with color based on weight polarity.
// Green indicates improvement (positive delta for positive weight, or vice versa).
// Red indicates degradation. Negligible changes (< 0.005) are shown in default color.
func formatDelta(metric string, delta float64, weights *Weights) string {
	positive := weights.Get(metric) >= 0
	var c text.Color

	switch {
	case delta >= 0.005:
		c = IfThen(positive, text.FgGreen, text.FgRed)
	case delta <= -0.005:
		c = IfThen(positive, text.FgRed, text.FgGreen)
	default:
		c = text.Reset
	}

	if metric == "IN:OUT" {
		return c.Sprintf("%.2f", delta)
	}
	return c.Sprintf("%+.2f%%", delta)
}

// DoLayoutRankings evaluates and ranks keyboard layouts using weighted scoring.
//
// The process:
//  1. Load all .klf layouts from layoutsDir
//  2. Compute median and IQR for each metric across all layouts
//  3. Filter to specified layouts (if layoutFiles is not empty)
//  4. Normalize metrics using robust scaling and compute weighted scores
//  5. Sort by score and render results table with optional delta rows
//
// Parameters:
//   - layoutsDir: directory containing .klf layout files
//   - layoutFiles: specific layout files to rank (empty means all)
//   - corpus: text corpus for frequency analysis
//   - idealRowLoad: ideal row load distribution for balance metrics
//   - idealfgrLoad: ideal finger load distribution for balance metrics
//   - weights: metric weights for scoring
//   - metricsSet: which metric set to display ("basic", "extended", or "fingers")
//   - deltas: how to display deltas ("none", "rows", "median", or a layout name)
func DoLayoutRankings(layoutsDir string, layoutFiles []string, corpus *Corpus, idealRowLoad *[3]float64, idealfgrLoad *[10]float64, weights *Weights, metricsSet string, deltas string) error {
	// Select the appropriate metric set
	metrics, ok := MetricsMap[metricsSet]
	if !ok {
		opts := slices.Collect(maps.Keys(MetricsMap))
		return fmt.Errorf("invalid metrics mode %q; must be one of %v", metricsSet, opts)
	}

	// Load and analyze all layouts (needed for normalization even if we filter later)
	analysers, err := LoadAnalysers(layoutsDir, corpus, idealRowLoad, idealfgrLoad)
	if err != nil {
		return err
	}
	medians, iqrs := computeMediansAndIQR(analysers)

	// Build lookup map for filtering
	analyserMap := make(map[string]*Analyser, len(analysers))
	for _, analyser := range analysers {
		analyserMap[analyser.Layout.Name] = analyser
	}

	// Filter to requested layouts
	filteredAnalysers := make([]*Analyser, 0, len(layoutFiles))
	for _, fname := range layoutFiles {
		layoutName := strings.TrimSuffix(fname, filepath.Ext(fname))
		analyser, ok := analyserMap[layoutName]
		if !ok {
			return fmt.Errorf("layout file %s was not found", fname)
		}
		filteredAnalysers = append(filteredAnalysers, analyser)
	}

	// Compute scores using normalized metrics
	layoutScores := computeScores(filteredAnalysers, medians, iqrs, weights)

	// Optionally add a median reference row
	var base *LayoutScore
	if deltas == "median" {
		lss := computeScores([]*Analyser{{Layout: &SplitLayout{Name: "median"}, Metrics: medians}}, medians, iqrs, weights)
		layoutScores = append(layoutScores, lss[0])
	}

	// Sort by score (higher is better)
	sort.Slice(layoutScores, func(i, j int) bool {
		return layoutScores[i].Score > layoutScores[j].Score
	})

	// Find base layout for delta comparison if specified
	if deltas != "none" && deltas != "rows" {
		if i := slices.IndexFunc(layoutScores, func(ls LayoutScore) bool { return ls.Name == deltas }); i < 0 {
			return fmt.Errorf("can't find %s", deltas)
		} else {
			base = &layoutScores[i]
		}
	}

	// Display the ranking table
	renderTable(layoutScores, metrics, weights, metricsSet == "extended", deltas, base)

	return nil
}
