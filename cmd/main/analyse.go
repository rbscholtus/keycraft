package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var analyseCommand = &cli.Command{
	Name:      "analyse",
	Usage:     "Analyse a layout file with a corpus file and style",
	ArgsUsage: "<layout file>",
	Flags:     flagsSlice("corpus"),
	Action:    analyseAction,
}

func analyseAction(c *cli.Context) error {
	corp, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// lay, err := loadLayout(c.Args().First())
	// // fmt.Println(lay)
	// sfb := lay.AnalyzeSfbs(corp)
	// fmt.Println(sfb)

	if err := doAnalysis(corp, c.Args().Slice()); err != nil {
		return err
	}
	return nil
}

func doAnalysis(corpus *layout.Corpus, layoutFilenames []string) error {
	// load an analyser for each layout
	analysers := make([]*layout.Analyser, 0, len(layoutFilenames))
	details := make([][]*layout.MetricAnalysis, 0, len(layoutFilenames))
	for _, fn := range layoutFilenames {
		lay, err := loadLayout(fn)
		if err != nil {
			return err
		}
		an := layout.NewAnalyser(lay, corpus, "")
		analysers = append(analysers, an)
		details = append(details, an.AllMetricsDetails())
	}

	if len(details) < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// Print the layout(s)
	for _, an := range analysers {
		fmt.Println(an.Layout)
	}

	// Make a table with a column for each layout
	twOuter := table.NewWriter()
	twOuter.SetStyle(EmptyStyle())
	twOuter.Style().Options.SeparateRows = true
	colConfigs := make([]table.ColumnConfig, 0, len(layoutFilenames))
	for i := range len(layoutFilenames) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i + 2, AlignHeader: text.AlignCenter})
	}
	twOuter.SetColumnConfigs(colConfigs)

	// Add header
	h := table.Row{"Metric"}
	for _, an := range analysers {
		h = append(h, an.Layout.Name)
	}
	twOuter.AppendHeader(h)

	// h := table.Row{"Layout"}
	// for _, an := range analysers {
	// 	h = append(h, an.Layout)
	// }
	// twOuter.AppendHeader(h)

	// Add data rows
	metrics := details[0]
	for i, ma := range metrics {
		data := table.Row{ma.Metric}
		for _, mas := range details {
			data = append(data, mas[i])
		}
		twOuter.AppendRow(data)
	}

	// Print layout(s) in the table
	fmt.Println(twOuter.Render())
	return nil
}

func EmptyStyle() table.Style {
	s := table.StyleDefault
	s.Box = table.StyleBoxRounded
	s.Box = table.BoxStyle{
		BottomLeft:       s.Box.BottomLeft,
		BottomRight:      s.Box.BottomRight,
		BottomSeparator:  s.Box.BottomSeparator,
		Left:             " ",
		LeftSeparator:    s.Box.LeftSeparator,
		MiddleHorizontal: " ",
		MiddleSeparator:  s.Box.MiddleSeparator,
		MiddleVertical:   " ",
		Right:            " ",
		RightSeparator:   s.Box.RightSeparator,
		TopLeft:          s.Box.TopLeft,
		TopRight:         s.Box.TopRight,
		TopSeparator:     s.Box.TopSeparator,
		UnfinishedRow:    " ",
	}
	return s
}
