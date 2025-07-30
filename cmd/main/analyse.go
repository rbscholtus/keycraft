package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var analyseCommand = &cli.Command{
	Name:      "analyse",
	Usage:     "Analyse a layout file with a corpus file and style",
	ArgsUsage: "<layout file>",
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
			Required: true,
		},
	},
	Action: analyseAction,
}

func analyseAction(c *cli.Context) error {
	layoutFile := c.Args().First()
	corpusFile := c.String("corpus")
	style := c.String("style")

	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	fmt.Printf("Analysing layout: %s with corpus: %s and style: %s\n", layoutFile, corpusFile, style)
	return nil
}
