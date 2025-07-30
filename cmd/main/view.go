package main

import (
	"fmt"
	"path/filepath"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

const (
	layoutDir = "data/layouts/"
	corpusDir = "data/corpus/"
)

var viewCommand = &cli.Command{
	Name:      "view",
	Usage:     "View a layout file with a corpus file",
	ArgsUsage: "<layout file>",
	Action:    viewAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
}

func viewAction(c *cli.Context) error {
	layoutFile := c.Args().First()
	corpusFile := c.String("corpus")

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

	doViewLayout(lay, corp)

	return nil
}

func doViewLayout(lay *layout.SplitLayout, corp *layout.Corpus) {
	an := layout.NewAnalyser(lay, corp)
	fmt.Println(an.HandUsageString())
	fmt.Println(an.MetricsString())
}
