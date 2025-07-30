package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var viewCommand = &cli.Command{
	Name:      "view",
	Usage:     "View a layout file with a corpus file",
	ArgsUsage: "<layout file>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
	Action: viewAction,
}

func viewAction(c *cli.Context) error {
	layoutFile := c.Args().First() // Assuming the first argument is the layout file
	corpusFile := c.String("corpus")

	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	fmt.Printf("Viewing layout: %s with corpus: %s\n", layoutFile, corpusFile)
	return nil
}
