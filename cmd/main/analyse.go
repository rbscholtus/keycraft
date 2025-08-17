package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
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

	lay, err := loadLayout(c.Args().First())
	if err != nil {
		return err
	}

	style := c.String("style")
	doAnalysis(lay, corp, style)
	return nil
}

func doAnalysis(lay *layout.SplitLayout, corp *layout.Corpus, _ string) {
	fmt.Println(lay)
	sfb := lay.AnalyzeSfbs(corp)
	sfs := lay.AnalyzeSfss(corp)
	lsb := lay.AnalyzeLsbs(corp)
	fsb := lay.AnalyzeScissors(corp)

	twOuter := table.NewWriter()
	twOuter.AppendRow(table.Row{sfb, sfs})
	twOuter.AppendRow(table.Row{lsb, "LSS"})
	twOuter.AppendRow(table.Row{fsb, "FSS"})
	twOuter.SetStyle(table.StyleLight)
	twOuter.Style().Options.SeparateRows = true
	fmt.Println(twOuter.Render())
}
