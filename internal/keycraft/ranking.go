package keycraft

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

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
func DoLayoutRankings(layoutsDir string, layoutFiles []string, corpus *Corpus, idealRowLoad *[3]float64, idealfgrLoad *[10]float64, pinkyWeights *[12]float64, weights *Weights, metricsSet string, deltas string) error {
	// Select the appropriate metric set
	metrics, ok := MetricsMap[metricsSet]
	if !ok {
		opts := slices.Collect(maps.Keys(MetricsMap))
		return fmt.Errorf("invalid metrics mode %q; must be one of %v", metricsSet, opts)
	}

	// Load and analyze all layouts (needed for normalization even if we filter later)
	analysers, err := LoadAnalysers(layoutsDir, corpus, idealRowLoad, idealfgrLoad, pinkyWeights)
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
	renderTable(layoutScores, metrics, weights, metricsSet != "fingers", deltas, base)

	return nil
}
