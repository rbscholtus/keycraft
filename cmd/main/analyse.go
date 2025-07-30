package main

import (
	"fmt"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var analyseCommand = &cli.Command{
	Name:      "analyse",
	Usage:     "Analyse a layout file with a corpus file and style",
	ArgsUsage: "<layout file>",
	Action:    analyseAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "style",
			Usage:    "specify the style",
			Required: false,
		},
	},
}

func analyseAction(c *cli.Context) error {
	layoutFile := c.Args().First()
	corpusFile := c.String("corpus")
	style := c.String("style")

	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	if corpusFile == "" {
		return fmt.Errorf("corpus file is required")
	}

	layoutPath := filepath.Join(layoutDir, layoutFile)
	lay, err := layout.NewLayoutFromFile(layoutFile, layoutPath)
	if err != nil {
		return fmt.Errorf("failed to load layout from %s: %v", layoutPath, err)
	}

	corpusPath := filepath.Join(corpusDir, corpusFile)
	corp, err := layout.NewCorpusFromFile(corpusFile, corpusPath)
	if err != nil {
		return fmt.Errorf("failed to load corpus from %s: %v", corpusPath, err)
	}

	doAnalysis(lay, corp, style)
	return nil
}

func doAnalysis(lay *layout.SplitLayout, corp *layout.Corpus, style string) {
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
