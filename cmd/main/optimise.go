package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Usage:     "Optimise a layout file with a corpus, pins, weights, generations, and accept function",
	ArgsUsage: "<layout file>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "pins",
			Aliases:  []string{"p"},
			Usage:    "specify the pins file",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "weights",
			Aliases:  []string{"w"},
			Usage:    "specify the weights configuration",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "generations",
			Aliases:  []string{"g"},
			Usage:    "specify the number of generations",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "accept",
			Aliases:  []string{"a"},
			Usage:    "specify the accept function",
			Required: true,
		},
	},
	Action: optimiseAction,
}

func optimiseAction(c *cli.Context) error {
	layoutFile := c.Args().First()
	corpusFile := c.String("corpus")
	pinsFile := c.String("pins")
	weightsConfig := c.String("weights")
	numGenerations := c.Int("generations")
	acceptFunction := c.String("accept")

	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	fmt.Printf("Optimising layout: %s with corpus: %s, pins: %s, weights: %s, generations: %d, accept function: %s\n", layoutFile, corpusFile, pinsFile, weightsConfig, numGenerations, acceptFunction)
	return nil
}
