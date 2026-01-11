package tui

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// RenderAnalyse renders detailed analysis results to stdout.
// Displays board, hand/finger/row load, stats overview, and detailed metric tables.
func RenderAnalyse(result *kc.AnalyseResult, opts kc.AnalyseDisplayOptions) error {
	if len(result.Analysers) < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// Make a table with a column for each layout
	twOuter := table.NewWriter()
	twOuter.SetStyle(EmptyStyle())
	twOuter.Style().Options.SeparateRows = true
	colConfigs := make([]table.ColumnConfig, 0, len(result.Analysers))
	for i := range len(result.Analysers) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i + 2,
			AlignHeader: text.AlignCenter, Align: text.AlignCenter})
	}
	twOuter.SetColumnConfigs(colConfigs)

	// Add header
	h := table.Row{""}
	for _, an := range result.Analysers {
		h = append(h, an.Layout.Name)
	}
	twOuter.AppendHeader(h)

	// Layout picture
	h = table.Row{"Board"}
	for _, an := range result.Analysers {
		h = append(h, SplitLayoutString(an.Layout))
	}
	twOuter.AppendRow(h)

	// Hand load distribution
	h = table.Row{"Hand"}
	for _, an := range result.Analysers {
		h = append(h, HandUsageString(an))
	}
	twOuter.AppendRow(h)

	// Row load distribution
	h = table.Row{"Row"}
	for _, an := range result.Analysers {
		h = append(h, RowUsageString(an))
	}
	twOuter.AppendRow(h)

	// Metrics overview
	h = table.Row{"Stats"}
	for _, an := range result.Analysers {
		h = append(h, MetricsString(an))
	}
	twOuter.AppendRow(h)

	// Add detailed data rows
	details := make([][]*kc.MetricDetails, 0, len(result.Analysers))
	for _, an := range result.Analysers {
		details = append(details, an.AllMetricsDetails())
	}

	metrics := details[0] // get the first entry to get the metrics
	for i, ma := range metrics {
		data := table.Row{ma.Metric}
		for _, mas := range details {
			data = append(data, MetricDetailsString(mas[i], opts.MaxRows))
		}
		twOuter.AppendRow(data)
	}

	// Add trigram table row
	h = table.Row{"Trigr"}
	for _, an := range result.Analysers {
		h = append(h, TopTrigramsString(an, opts.CompactTrigrams, opts.TrigramRows))
	}
	twOuter.AppendRow(h)

	// Print layout(s) in the table
	fmt.Println(twOuter.Render())
	return nil
}

// MetricDetailsString renders metric details as a paginated table.
func MetricDetailsString(ma *kc.MetricDetails, nrows int) string {
	t := createSimpleTable()

	// Collect unique custom keys
	customKeys := []string{}
	customKeySet := make(map[string]bool)
	for _, fields := range ma.Custom {
		for k := range fields {
			if !customKeySet[k] {
				customKeySet[k] = true
				customKeys = append(customKeys, k)
			}
		}
		break // we don't need to visit all rows - they all have the same custom fields
	}

	// Header
	header := table.Row{"orderby", ma.Metric, "Count", "%", "Dist"}
	for _, ck := range customKeys {
		header = append(header, ck)
	}
	t.AppendHeader(header)

	// Rows
	for ngram := range ma.NGramCount {
		row := []any{
			ma.NGramCount[ngram],
			ngram,
			ma.NGramCount[ngram],
			float64(ma.NGramCount[ngram]) / float64(ma.CorpusNGramC),
			ma.NGramDist[ngram],
		}
		for _, ck := range customKeys {
			if fields, ok := ma.Custom[ngram]; ok {
				row = append(row, fields[ck])
			} else {
				row = append(row, "")
			}
		}
		t.AppendRow(row)
	}

	footer := table.Row{"", "", ma.TotalNGrams, float64(ma.TotalNGrams) / float64(ma.CorpusNGramC)}
	for range customKeys {
		footer = append(footer, "")
	}
	t.AppendFooter(footer)

	return t.Pager(table.PageSize(nrows)).Render()
}

// createSimpleTable returns a configured table writer with rounded style and common settings.
func createSimpleTable() table.Writer {
	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.SetStyle(table.StyleRounded)
	tw.Style().Title.Align = text.AlignCenter
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Distance", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Dist", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Row", Transformer: Fraction},
		{Name: "Count", Transformer: Thousands, TransformerFooter: Thousands},
		{Name: "%", Transformer: Percentage, TransformerFooter: Percentage},
		{Name: "Cum%", Transformer: Percentage, TransformerFooter: Percentage},
		{Name: "Δrow", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Δcol", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Angle", Transformer: Angle, TransformerFooter: Angle},
		{Name: "Length", Align: text.AlignRight},
	})
	tw.SortBy([]table.SortBy{{Name: "orderby", Mode: table.DscNumeric}})
	return tw
}

// TopTrigramsString generates a table showing the top N trigrams with their
// classifications (ALT, 2RL, 3RL, RED) and their specific categories.
func TopTrigramsString(an *kc.Analyser, compactTrigrams bool, trigramRows int) string {
	t := createSimpleTable()

	// Get trigram classifications from TrigramDetails
	alt, rl2, rl3, red := an.TrigramDetails()

	// Helper to get classification for a trigram
	getClassification := func(triStr string) string {
		// Check each metric category
		if _, ok := alt.NGramCount[triStr]; ok {
			if cat, hasCat := alt.Custom[triStr]["Dir"]; hasCat {
				return fmt.Sprintf("ALT-%v", cat)
			}
			return "ALT"
		}
		if _, ok := rl2.NGramCount[triStr]; ok {
			if cat, hasCat := rl2.Custom[triStr]["Dir"]; hasCat {
				return fmt.Sprintf("2RL-%v", cat)
			}
			return "2RL"
		}
		if _, ok := rl3.NGramCount[triStr]; ok {
			if cat, hasCat := rl3.Custom[triStr]["Dir"]; hasCat {
				return fmt.Sprintf("3RL-%v", cat)
			}
			return "3RL"
		}
		if _, ok := red.NGramCount[triStr]; ok {
			if cat, hasCat := red.Custom[triStr]["Dir"]; hasCat {
				return fmt.Sprintf("RED-%v", cat)
			}
			return "RED"
		}
		return "OTHER"
	}

	// Common classifications that should not be highlighted
	commonClassifications := map[string]bool{
		"ALT-NML": true,
		"2RL-IN":  true,
		"2RL-OUT": true,
		"3RL-IN":  true,
		"3RL-OUT": true,
	}

	// Get top N trigrams from corpus
	topTrigrams := an.Corpus.TopTrigrams(trigramRows)

	// Header
	header := table.Row{"orderby", "Tri", "Count", "%", "Cum%", "Class"}
	t.AppendHeader(header)

	// Calculate cumulative percentages and build rows
	totalTrigramCount := an.Corpus.TotalTrigramsCount
	cumulativeCount := uint64(0)
	rowNum := 0

	for _, pair := range topTrigrams {
		triStr := pair.Key.String()
		count := pair.Count
		classification := getClassification(triStr)

		// Skip if compact mode and category is common
		if compactTrigrams && commonClassifications[classification] {
			continue
		}

		cumulativeCount += count
		percentage := float64(count) / float64(totalTrigramCount)
		cumulativePercentage := float64(cumulativeCount) / float64(totalTrigramCount)

		rowNum++

		// Color the entire row if it's a non-common classification
		isNonCommon := !commonClassifications[classification]
		var row table.Row
		if isNonCommon {
			row = table.Row{
				count,                              // orderby (hidden, for sorting)
				text.FgHiRed.Sprintf("%s", triStr), // Tri (colored)
				text.FgHiRed.Sprintf("%d", count),  // Count (colored)
				text.FgHiRed.Sprintf("%.2f%%", percentage*100),           // % (colored)
				text.FgHiRed.Sprintf("%.2f%%", cumulativePercentage*100), // Cum% (colored)
				text.FgHiRed.Sprintf("%s", classification),               // Classification (colored)
			}
		} else {
			row = table.Row{
				count,                // orderby (hidden, for sorting)
				triStr,               // Tri
				count,                // Count
				percentage,           // %
				cumulativePercentage, // Cum%
				classification,       // Classification
			}
		}
		t.AppendRow(row)
	}

	// Footer with totals
	finalPercentage := float64(cumulativeCount) / float64(totalTrigramCount)
	footer := table.Row{"", "", cumulativeCount, finalPercentage, finalPercentage, ""}
	t.AppendFooter(footer)

	return t.Render()
}
