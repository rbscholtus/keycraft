package main

import (
	"fmt"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var analyseCommand = &cli.Command{
	Name:      "analyse",
	Aliases:   []string{"a"},
	Usage:     "Analyse one or more layouts in detail",
	ArgsUsage: "<layout1.klf> <layout2.klf> ...",
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

	if err := DoAnalysis(corp, c.Args().Slice(), true); err != nil {
		return err
	}
	return nil
}

func DoAnalysis(corpus *layout.Corpus, layoutFilenames []string, dataTables bool) error {
	// load an analyser for each layout
	analysers := make([]*layout.Analyser, 0, len(layoutFilenames))
	for _, fn := range layoutFilenames {
		lay, err := loadLayout(fn)
		if err != nil {
			return err
		}
		an := layout.NewAnalyser(lay, corpus, "")
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
		h = append(h, an.Layout)
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
		details := make([][]*layout.MetricDetails, 0, len(layoutFilenames))
		for _, an := range analysers {
			details = append(details, an.AllMetricsDetails())
		}

		metrics := details[0] // get the first entry to get the metrics
		for i, ma := range metrics {
			data := table.Row{ma.Metric}
			for _, mas := range details {
				data = append(data, mas[i])
			}
			twOuter.AppendRow(data)
		}
	}

	// Print layout(s) in the table
	fmt.Println(twOuter.Render())
	return nil
}

func HandUsageString(an *layout.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""

	// Define column headers
	fingerAbbreviations := table.Row{"LP", "LP", "LR", "LM", "LI", "LI", "RI", "RI", "RM", "RR", "RP", "RP"}
	colConfigs := make([]table.ColumnConfig, 0, len(fingerAbbreviations))
	for i := range len(fingerAbbreviations) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i,
			AlignHeader: text.AlignCenter, Align: text.AlignCenter})
	}
	tw.SetColumnConfigs(colConfigs)
	tw.AppendHeader(fingerAbbreviations, table.RowConfig{AutoMerge: true})

	// Append row with ColumnUsage data
	columnUsageRow := make(table.Row, 12)
	for i := range columnUsageRow {
		key := "C" + strconv.Itoa(i)
		columnUsageRow[i] = fmt.Sprintf("%.1f", an.Metrics[key])
	}
	tw.AppendRow(columnUsageRow)

	// Append row with FingerUsage data and merged cells
	fingerUsageRow := make(table.Row, 12)
	fi := [12]int{0, 0, 1, 2, 3, 3, 6, 6, 7, 8, 9, 9}
	for i, v := range fi {
		key := "F" + strconv.Itoa(v)
		fingerUsageRow[i] = fmt.Sprintf("%.1f", an.Metrics[key])
	}
	tw.AppendRow(fingerUsageRow, table.RowConfig{AutoMerge: true})

	// Append row with HandUsage data in merged cells
	handUsageRow := make(table.Row, 12)
	hi := [12]int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}
	for i, v := range hi {
		key := "H" + strconv.Itoa(v)
		handUsageRow[i] = fmt.Sprintf("%.1f%%", an.Metrics[key])
	}
	tw.AppendRow(handUsageRow, table.RowConfig{AutoMerge: true})

	return tw.Render()
}

func RowUsageString(an *layout.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 2, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 3, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 4, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
	})

	tw.AppendRow(table.Row{"Top", "Home", "Bottom", "Thumb"})
	tw.AppendRow(table.Row{
		fmt.Sprintf("%.1f%%", an.Metrics["R0"]), fmt.Sprintf("%.1f%%", an.Metrics["R1"]),
		fmt.Sprintf("%.1f%%", an.Metrics["R2"]), fmt.Sprintf("%.1f%%", an.Metrics["R3"]),
	})

	return tw.Render()
}

func MetricsString(an *layout.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignJustify},
		{Number: 2, Align: text.AlignJustify},
		{Number: 3, Align: text.AlignJustify},
		{Number: 4, Align: text.AlignJustify},
	})

	data := []table.Row{
		{
			fmt.Sprintf("SFB: %.2f%%", an.Metrics["SFB"]),
			fmt.Sprintf("LSB: %.2f%%", an.Metrics["LSB"]),
			fmt.Sprintf("FSB: %.2f%%", an.Metrics["FSB"]),
			fmt.Sprintf("HSB: %.2f%%", an.Metrics["HSB"]),
		},
		{
			fmt.Sprintf("SFS: %.2f%%", an.Metrics["SFS"]),
			fmt.Sprintf("LSS: %.2f%%", an.Metrics["LSS"]),
			fmt.Sprintf("FSS: %.2f%%", an.Metrics["FSS"]),
			fmt.Sprintf("HSS: %.2f%%", an.Metrics["HSS"]),
		},
		{
			fmt.Sprintf("ALT: %.2f%%", an.Metrics["ALT"]),
			fmt.Sprintf("2RL: %.2f%%", an.Metrics["2RL"]),
			fmt.Sprintf("3RL: %.2f%%", an.Metrics["3RL"]),
			fmt.Sprintf("RED: %.2f%%", an.Metrics["RED"]),
		},
		{
			fmt.Sprintf("I:O: %.2f", an.Metrics["IN:OUT"]),
			fmt.Sprintf("FBL: %.2f%%", an.Metrics["FBL"]),
			fmt.Sprintf("POH: %.2f%%", an.Metrics["POH"]),
			fmt.Sprintf("BAD: %.2f%%", an.Metrics["RED-BAD"]),
		},
	}
	tw.AppendRows(data)

	return tw.Render()
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
