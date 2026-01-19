package tui

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// OutputFormat determines how results are rendered.
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputHTML  OutputFormat = "html"
	OutputCSV   OutputFormat = "csv"
)

// MetricsOption determines which metrics to display.
type MetricsOption string

const (
	MetricsBasic    MetricsOption = "basic"
	MetricsExtended MetricsOption = "extended"
	MetricsFingers  MetricsOption = "fingers"
	MetricsWeighted MetricsOption = "weighted"
	MetricsAll      MetricsOption = "all"
	MetricsCustom   MetricsOption = "custom"
)

// DeltasOption determines how deltas are displayed.
type DeltasOption string

const (
	DeltasNone   DeltasOption = "none"
	DeltasRows   DeltasOption = "rows"
	DeltasMedian DeltasOption = "median"
	DeltasCustom DeltasOption = "custom" // Compare to specific layout
)

// RankingDisplayOptions configures presentation and comparison behavior.
// Supports predefined metric sets or custom lists, various delta modes, and multiple output formats.
type RankingDisplayOptions struct {
	OutputFormat   OutputFormat
	MetricsOption  MetricsOption
	CustomMetrics  []string     // Used when MetricsOption == MetricsCustom
	ShowWeights    bool         // Display weight row in output
	Weights        *kc.Weights  // Metric weights for display and delta coloring
	DeltasOption   DeltasOption // "none", "rows", "median", "custom"
	BaseLayoutName string       // Name of reference layout when DeltasOption == DeltasCustom
	CorpusName     string       // Name of the corpus used for ranking
	// baseLayoutScores *kc.LayoutScore // Cached reference to base layout scores (set during rendering)
}

// GetMetrics returns the list of metrics to display based on options.
func (opts RankingDisplayOptions) GetMetrics() []string {
	if opts.MetricsOption == MetricsCustom {
		return opts.CustomMetrics
	}
	if opts.MetricsOption == MetricsWeighted {
		// Return all metrics with absolute weight >= 0.01
		allMetrics := kc.MetricsMap["all"]
		var weightedMetrics []string
		for _, metric := range allMetrics {
			if weight := opts.Weights.Get(metric); weight >= 0.01 || weight <= -0.01 {
				weightedMetrics = append(weightedMetrics, metric)
			}
		}
		return weightedMetrics
	}
	return kc.MetricsMap[string(opts.MetricsOption)]
}

// RenderRankingTable formats and prints ranking results.
func RenderRankingTable(result *kc.RankingResult, opts RankingDisplayOptions) error {
	metrics := opts.GetMetrics()

	scores := result.Scores

	// Optionally add median reference row before sorting
	if opts.DeltasOption == DeltasMedian {
		medianScore := kc.ComputeMedianScore(result.Medians, opts.Weights)
		scores = append(scores, medianScore)
	}

	// Sort scores once (higher is better)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Render based on output format
	switch opts.OutputFormat {
	case OutputTable:
		renderTableTerminal(scores, metrics, opts)
		return nil
	case OutputHTML:
		renderTableHTML(scores, metrics, opts)
		return nil
	case OutputCSV:
		return renderCSV(os.Stdout, scores, metrics, opts)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.OutputFormat)
	}
}

// renderTableTerminal renders to terminal with colors.
func renderTableTerminal(scores []kc.LayoutScore, metrics []string, opts RankingDisplayOptions) {
	tw := buildTable(scores, metrics, opts)
	fmt.Println(tw.Render())
}

// renderTableHTML uses go-pretty's built-in HTML rendering.
func renderTableHTML(scores []kc.LayoutScore, metrics []string, opts RankingDisplayOptions) {
	tw := buildTable(scores, metrics, opts)
	tw.SetHTMLCSSClass("keycraft-ranking-table")
	fmt.Println(tw.RenderHTML())
}

// buildTable creates the table structure (shared by both HTML and terminal rendering).
func buildTable(scores []kc.LayoutScore, metrics []string, opts RankingDisplayOptions) table.Writer {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.Style().Title.Align = text.AlignLeft

	// Set title based on delta mode
	switch opts.DeltasOption {
	case DeltasCustom:
		tw.SetTitle(fmt.Sprintf("Layout Ranking (Compare to %s)", opts.BaseLayoutName))
	case DeltasMedian:
		tw.SetTitle("Layout Ranking (Compare to median)")
	default:
		if opts.CorpusName != "" {
			tw.SetTitle(fmt.Sprintf("Layout Ranking - %s", opts.CorpusName))
		} else {
			tw.SetTitle("Layout Ranking")
		}
	}

	// Configure column alignment
	colConfigs := []table.ColumnConfig{
		{Name: "Index", Align: text.AlignRight},
		{Name: "Name", Align: text.AlignLeft},
		{Name: "Th", Align: text.AlignCenter},
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
	header := table.Row{"#", "Name", "Th", "Score"}
	for _, metric := range metrics {
		header = append(header, metric)
	}
	tw.AppendHeader(header)

	// Add weight row if requested
	if opts.ShowWeights {
		weightRow := table.Row{"", "Weight", "", ""}
		for _, metric := range metrics {
			weight := opts.Weights.Get(metric)
			weightRow = append(weightRow, fmt.Sprintf("%.2f", weight))
		}
		tw.AppendHeader(weightRow)
	}

	// Add data rows
	addDataRows(tw, scores, metrics, opts)

	return tw
}

// addDataRows populates the table with data (shared logic).
func addDataRows(tw table.Writer, scores []kc.LayoutScore, metrics []string, opts RankingDisplayOptions) {
	rowIdx := 1
	var baseLayout *kc.LayoutScore

	// Find reference layout for custom or median delta modes
	if opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian {
		refName := kc.IfThen(opts.DeltasOption == DeltasMedian, "median", opts.BaseLayoutName)
		if idx := slices.IndexFunc(scores, func(ls kc.LayoutScore) bool { return ls.Name == refName }); idx >= 0 {
			baseLayout = &scores[idx]
			rowIdx -= 1 + idx
		}
	}

	var prevMetrics, refMetrics []float64
	if (opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian) && baseLayout != nil {
		refMetrics = extractMetrics(baseLayout, metrics)
	}

	for i, score := range scores {
		// Build data row for this layout
		thumbChars := getThumbChars(&score)
		dataRow := table.Row{rowIdx, score.Name, thumbChars, fmt.Sprintf("%+.2f", score.Score)}
		currMetrics := extractMetrics(&score, metrics)
		for j, metric := range metrics {
			dataRow = append(dataRow, formatMetricValue(metric, currMetrics[j]))
		}

		// Add delta row showing differences from previous, median, or base layout
		if i > 0 && opts.DeltasOption != DeltasNone {
			deltaRow := table.Row{"", "", "", ""}
			for idx, currMetric := range currMetrics {
				var delta float64
				if opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian {
					if rowIdx <= 0 {
						delta = prevMetrics[idx] - refMetrics[idx]
					} else {
						delta = currMetric - refMetrics[idx]
					}
				} else {
					delta = currMetric - prevMetrics[idx]
				}
				deltaRow = append(deltaRow, formatDelta(metrics[idx], delta, opts.Weights))
			}
			tw.AppendRow(deltaRow)
		}
		tw.AppendRow(dataRow)

		prevMetrics = currMetrics
		rowIdx++
	}
}

// extractMetrics extracts metric values in the specified order.
func extractMetrics(score *kc.LayoutScore, metrics []string) []float64 {
	result := make([]float64, len(metrics))
	for i, metric := range metrics {
		result[i] = kc.WithDefault(score.Analyser.Metrics, metric, 0.0)
	}
	return result
}

// getThumbChars extracts thumb characters (indexes 36-41) with special formatting.
// Non-printable characters become spaces, spaces become ␣, and outer empty positions are trimmed.
func getThumbChars(score *kc.LayoutScore) string {
	if score.Analyser.Layout == nil {
		return ""
	}

	var thumb [6]rune

	for i := 36; i <= 41; i++ {
		r := score.Analyser.Layout.Runes[i]
		if r < ' ' {
			thumb[i-36] = ' '
		} else if r == ' ' {
			thumb[i-36] = '␣'
		} else {
			thumb[i-36] = r
		}
	}

	// Trim outer empty positions for cleaner display
	if thumb[0] == ' ' && thumb[5] == ' ' {
		if thumb[1] == ' ' && thumb[4] == ' ' {
			return string(thumb[2:4])
		}
		return string(thumb[1:5])
	}
	return string(thumb[:])
}

// renderCSV outputs rankings in CSV format.
func renderCSV(w io.Writer, scores []kc.LayoutScore, metrics []string, opts RankingDisplayOptions) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header row
	header := []string{"Rank", "Name", "Th", "Score"}
	header = append(header, metrics...)
	if err := writer.Write(header); err != nil {
		return err
	}

	// Optionally write weight row
	if opts.ShowWeights {
		weightRow := []string{"", "Weight", "", ""}
		for _, metric := range metrics {
			weight := opts.Weights.Get(metric)
			weightRow = append(weightRow, fmt.Sprintf("%.2f", weight))
		}
		if err := writer.Write(weightRow); err != nil {
			return err
		}
	}

	rowIdx := 1
	var baseLayout *kc.LayoutScore

	// Find reference layout for custom or median delta modes
	if opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian {
		refName := opts.BaseLayoutName
		if opts.DeltasOption == DeltasMedian {
			refName = "median"
		}
		if idx := slices.IndexFunc(scores, func(ls kc.LayoutScore) bool { return ls.Name == refName }); idx >= 0 {
			baseLayout = &scores[idx]
			rowIdx -= 1 + idx
		}
	}

	var prevMetrics, refMetrics []float64
	if (opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian) && baseLayout != nil {
		refMetrics = extractMetrics(baseLayout, metrics)
	}

	for i, score := range scores {
		// Build data row
		thumbChars := getThumbChars(&score)
		dataRow := []string{
			fmt.Sprintf("%d", rowIdx),
			score.Name,
			thumbChars,
			fmt.Sprintf("%.2f", score.Score),
		}
		currMetrics := extractMetrics(&score, metrics)
		for j, metric := range metrics {
			dataRow = append(dataRow, formatMetricValueCSV(metric, currMetrics[j]))
		}

		// Write delta row if needed
		if i > 0 && opts.DeltasOption != DeltasNone {
			deltaRow := []string{"", "", "", ""}
			for idx, currMetric := range currMetrics {
				var delta float64
				if opts.DeltasOption == DeltasCustom || opts.DeltasOption == DeltasMedian {
					if rowIdx <= 0 {
						delta = prevMetrics[idx] - refMetrics[idx]
					} else {
						delta = currMetric - refMetrics[idx]
					}
				} else {
					delta = currMetric - prevMetrics[idx]
				}
				deltaRow = append(deltaRow, formatDeltaCSV(metrics[idx], delta))
			}
			if err := writer.Write(deltaRow); err != nil {
				return err
			}
		}

		if err := writer.Write(dataRow); err != nil {
			return err
		}

		prevMetrics = currMetrics
		rowIdx++
	}

	return nil
}

// formatMetricValue formats a metric value for table display.
// IN:OUT ratio is displayed as a plain number, others as percentages.
func formatMetricValue(metric string, val float64) string {
	if metric == "IN:OUT" {
		return fmt.Sprintf("%.2f", val)
	}
	return fmt.Sprintf("%.2f%%", val)
}

// formatMetricValueCSV formats a metric value for CSV (no color, plain numbers).
func formatMetricValueCSV(_ string, val float64) string {
	// CSV stores raw values without percentage symbols
	return fmt.Sprintf("%.2f", val)
}

// formatDeltaCSV formats delta for CSV (no color codes).
func formatDeltaCSV(metric string, delta float64) string {
	if metric == "IN:OUT" {
		return fmt.Sprintf("%.2f", delta)
	}
	return fmt.Sprintf("%+.2f", delta)
}

// formatDelta formats the delta between metrics with color based on weight polarity.
// Green indicates improvement (positive delta for positive weight, or vice versa).
// Red indicates degradation. Negligible changes (< 0.005) are shown in default color.
func formatDelta(metric string, delta float64, weights *kc.Weights) string {
	positive := weights.Get(metric) >= 0
	var c text.Color

	switch {
	case delta >= 0.005:
		c = kc.IfThen(positive, text.FgGreen, text.FgRed)
	case delta <= -0.005:
		c = kc.IfThen(positive, text.FgRed, text.FgGreen)
	default:
		c = text.Reset
	}

	if metric == "IN:OUT" {
		return c.Sprintf("%.2f", delta)
	}
	return c.Sprintf("%+.2f%%", delta)
}
