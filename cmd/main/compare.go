package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rbscholtus/kb/internal/layout"
)

var compareCommand = &cli.Command{
	Name:      "compare",
	Usage:     "Compare multiple layout files with a corpus file",
	ArgsUsage: "<layout files...>",
	Action:    compareAction,
}
var ReversedMetrics = map[string]bool{"ALT": true, "ROL": true, "ONE": true}

func compareAction(c *cli.Context) error {
	layoutFiles := c.Args().Slice()

	if len(layoutFiles) < 2 || len(layoutFiles) > 5 {
		return fmt.Errorf("please provide between 2 and 5 layout files")
	}

	corpus, err := loadCorpus(c)
	if err != nil {
		return fmt.Errorf("failed to load corpus file: %w", err)
	}

	style := c.String("style")

	layouts := make([]*layout.SplitLayout, len(layoutFiles))
	analysers := make([]*layout.Analyser, len(layoutFiles))
	for i, lf := range layoutFiles {
		l, err := loadLayout(lf)
		if err != nil {
			return fmt.Errorf("failed to load layout file %s: %w", lf, err)
		}
		layouts[i] = l
		analysers[i] = layout.NewAnalyser(l, corpus, style)
	}

	t := table.NewWriter()
	header := table.Row{"Metric"}
	header = append(header, layoutFiles[0])
	for i := 1; i < len(layoutFiles); i++ {
		header = append(header, "Δ")
		header = append(header, layoutFiles[i])
	}
	t.AppendHeader(header)

	for _, metric := range layout.MetricNames {
		row := table.Row{metric}
		values := make([]float64, len(layoutFiles))
		for i, analyser := range analysers {
			values[i] = analyser.Metrics[metric]
		}
		row = append(row, fmt.Sprintf("%.2f%%", values[0]))
		for i := 1; i < len(values); i++ {
			delta := values[i] - values[i-1]
			var color text.Color
			if delta > 0 {
				color = layout.IfThen(ReversedMetrics[metric], text.FgGreen, text.FgRed)
			} else if delta < 0 {
				color = layout.IfThen(ReversedMetrics[metric], text.FgRed, text.FgGreen)
			} else {
				color = text.Reset
			}
			row = append(row, text.Colors{color}.Sprint(fmt.Sprintf("%+0.2f%%", delta)))
			row = append(row, fmt.Sprintf("%.2f%%", values[i]))
		}
		t.AppendRow(row)
	}

	fmt.Println(t.Render())

	return nil
}
