// Package layout provides functionality to evaluate and rank keyboard layouts
// based on various ergonomic and usage metrics, producing penalty scores that
// help determine the relative quality of each layout.

package layout

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// metricNames defines which metrics will be included in the ranking.
var metricNames = []string{"SFB", "LSB", "FSB", "HSB", "SFS", "LSS", "FSS", "HSS", "ALT", "ROL", "ONE", "RED"}

// LayoutScore holds the name of a layout, its computed penalty score,
// and the analyser used to generate metrics.
type LayoutScore struct {
	Name     string    // Name is the identifier or filename of the layout.
	Penalty  float64   // Penalty is the total weighted score used for ranking.
	Analyser *Analyser // Analyser contains detailed metric evaluations for the layout.
}

// loadAnalysers loads keyboard layout files from the specified directory,
// analyzes each layout against the given corpus and style, and returns a slice
// of analysers containing the computed metrics.
func loadAnalysers(layoutsDir string, corpus *Corpus, style string) ([]*Analyser, error) {
	var analysers = make([]*Analyser, 0)

	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, fmt.Errorf("error finding layout files in %v: %v", layoutsDir, err)
	}

	for _, file := range layoutFiles {
		if !strings.HasSuffix(file.Name(), ".klf") {
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

// computeMediansAndIQR calculates median and interquartile range (IQR) values
// for each metric across all analysers, which are used as normalization factors
// to scale metric values during scoring.
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

// computePenalties calculates a weighted penalty score for each layout by
// normalizing metric values using medians and IQR, applying weights, and then
// sorts the layouts by their total penalty to produce a ranked list.
func computePenalties(analysers []*Analyser, medians, iqr map[string]float64, weights map[string]float64) []LayoutScore {
	var layoutPenalties []LayoutScore
	for _, analyser := range analysers {
		penalty := 0.0
		for metric, value := range analyser.Metrics {
			weight, ok := weights[metric]
			if !ok {
				weight = 1.0
			}
			var scaledValue float64
			if iqr[metric] == 0 {
				scaledValue = 0
			} else {
				scaledValue = (value - medians[metric]) / iqr[metric]
			}
			penalty += weight * scaledValue
		}
		layoutPenalties = append(layoutPenalties, LayoutScore{analyser.Layout.Name, penalty, analyser})
	}

	sort.Slice(layoutPenalties, func(i, j int) bool {
		return layoutPenalties[i].Penalty < layoutPenalties[j].Penalty
	})

	return layoutPenalties
}

// renderTable outputs the ranked layout scores in a nicely formatted table
// including layout names, penalty scores, and individual metric values.
func renderTable(scores []LayoutScore) {
	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.SetTitle("Layout Ranks")
	tw.Style().Title.Align = text.AlignCenter

	configs := []table.ColumnConfig{
		{Name: "Name", Align: text.AlignLeft},
		{Name: "Penalty", Align: text.AlignRight, Transformer: func(a any) string {
			return fmt.Sprintf("%.2f", a)
		}},
	}
	for _, metricName := range metricNames {
		configs = append(configs, table.ColumnConfig{
			Name:  metricName,
			Align: text.AlignRight,
			Transformer: func(a any) string {
				return fmt.Sprintf("%.2f%%", a)
			},
		})
	}
	tw.SetColumnConfigs(configs)

	headerRow := []any{"Name", "Penalty"}
	for _, metricName := range metricNames {
		headerRow = append(headerRow, metricName)
	}
	tw.AppendHeader(headerRow)

	for _, scores := range scores {
		row := []any{scores.Name, scores.Penalty}
		for _, metricName := range metricNames {
			val, ok := scores.Analyser.Metrics[metricName]
			if !ok {
				continue
			}
			row = append(row, val)
		}
		tw.AppendRow(row)
	}

	fmt.Println(tw.Render())
}

// DoRankings orchestrates the entire ranking process: loading layouts,
// computing normalization factors, scoring layouts with weighted metrics,
// and rendering the final ranked table.
func DoRankings(corpus *Corpus, layoutsDir string, weights map[string]float64, style string) error {
	//fmt.Printf("Ranking layouts in data/layouts with %s and weights: %v\n", corpus.Name, weights)

	analysers, err := loadAnalysers(layoutsDir, corpus, style)
	if err != nil {
		return err
	}

	medians, iqr := computeMediansAndIQR(analysers)

	layoutPenalties := computePenalties(analysers, medians, iqr, weights)

	renderTable(layoutPenalties)

	return nil
}
