package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// analyseCommand defines the "analyse" CLI command which prints detailed
// analysis for one or more layouts (including data tables when requested).
var analyseCommand = &cli.Command{
	Name:      "analyse",
	Aliases:   []string{"a"},
	Usage:     "Analyse one or more keyboard layouts in detail",
	Flags:     flagsSlice("rows", "corpus", "finger-load"),
	ArgsUsage: "<layout1> <layout2> ...",
	Before:    validateAnalyseFlags,
	Action:    analyseAction,
}

// validateViewFlags validates CLI flags before running the view command.
func validateAnalyseFlags(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}
	return nil
}

// analyseAction loads the specified corpus, finger load, and layouts,
// then executes the analysis process. Returns an error if loading fails.
func analyseAction(c *cli.Context) error {
	corpus, err := getCorpusFromFlag(c)
	if err != nil {
		return err
	}

	fingerBal, err := getFingerLoadFromFlag(c)
	if err != nil {
		return err
	}

	layouts := getLayoutArgs(c)

	// Run detailed analysis on all specified layouts with the given corpus and finger load.
	if err := DoAnalysis(layouts, corpus, fingerBal, true, c.Int("rows")); err != nil {
		return err
	}

	return nil
}

// DoAnalysis loads analysers for the provided layouts, produces overview
// rows (board, hand, row, stats) and optionally appends detailed metric
// tables. The rendered table output is printed to stdout.
func DoAnalysis(layoutFilenames []string, corpus *kc.Corpus, fgrLoad *[10]float64, dataTables bool, nRows int) error {
	// load an analyser for each layout
	analysers := make([]*kc.Analyser, 0, len(layoutFilenames))
	for _, fn := range layoutFilenames {
		layout, err := loadLayout(fn)
		if err != nil {
			return err
		}
		an := kc.NewAnalyser(layout, corpus, fgrLoad)
		analysers = append(analysers, an)
	}

	if len(analysers) < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// Make a table with a column for each layout
	twOuter := table.NewWriter()
	twOuter.SetStyle(EmptyStyle())
	twOuter.Style().Options.SeparateRows = true
	colConfigs := make([]table.ColumnConfig, 0, len(layoutFilenames))
	for i := range len(layoutFilenames) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i + 2,
			AlignHeader: text.AlignCenter, Align: text.AlignCenter})
	}
	twOuter.SetColumnConfigs(colConfigs)

	// Add header
	h := table.Row{""}
	for _, an := range analysers {
		h = append(h, an.Layout.Name)
	}
	twOuter.AppendHeader(h)

	// Layout picture
	h = table.Row{"Board"}
	for _, an := range analysers {
		// use cmd-level formatter instead of relying on an.Layout.String()
		h = append(h, SplitLayoutString(an.Layout))
	}
	twOuter.AppendRow(h)

	// Hand balance
	h = table.Row{"Hand"}
	for _, an := range analysers {
		h = append(h, HandUsageString(an))
	}
	twOuter.AppendRow(h)

	// Row balance
	h = table.Row{"Row"}
	for _, an := range analysers {
		h = append(h, RowUsageString(an))
	}
	twOuter.AppendRow(h)

	// Metrics overview
	h = table.Row{"Stats"}
	for _, an := range analysers {
		h = append(h, MetricsString(an))
	}
	twOuter.AppendRow(h)

	// Add data rows
	if dataTables {
		details := make([][]*kc.MetricDetails, 0, len(layoutFilenames))
		for _, an := range analysers {
			details = append(details, an.AllMetricsDetails())
		}

		metrics := details[0] // get the first entry to get the metrics
		for i, ma := range metrics {
			data := table.Row{ma.Metric}
			for _, mas := range details {
				// render MetricDetails via the cmd formatter
				data = append(data, MetricDetailsString(mas[i], nRows))
			}
			twOuter.AppendRow(data)
		}
	}

	// Print layout(s) in the table
	fmt.Println(twOuter.Render())
	return nil
}
