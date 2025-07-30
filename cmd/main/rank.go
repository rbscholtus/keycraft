package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var rankCommand = &cli.Command{
	Name:      "rank",
	Usage:     "Rank multiple layout files with a corpus file",
	ArgsUsage: "<layout files...>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
	Action: rankAction,
}

func rankAction(c *cli.Context) error {
	layoutFiles := c.Args().Slice()
	corpusFile := c.String("corpus")

	if len(layoutFiles) < 1 {
		return fmt.Errorf("at least one layout file is required")
	}

	fmt.Printf("Ranking layouts: %v with corpus: %s\n", layoutFiles, corpusFile)
	return nil
}
