package main

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var validAcceptFunctions = []string{"always", "drop-slow", "gradual", "drop-fast", "never"}

var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Usage:     "Optimise a layout file with a corpus, pins, weights, generations, and accept function",
	ArgsUsage: "<layout file>",
	Action:    optimiseAction,
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
			Required: false,
		},
		&cli.StringFlag{
			Name:     "weights",
			Aliases:  []string{"w"},
			Usage:    "specify the weights configuration",
			Required: false,
		},
		&cli.UintFlag{
			Name:     "generations",
			Aliases:  []string{"g"},
			Usage:    "specify the number of generations",
			Required: false,
			Value:    99,
		},
		&cli.StringFlag{
			Name:     "accept",
			Aliases:  []string{"a"},
			Usage:    "specify the accept function",
			Required: false,
			Value:    "drop-slow",
		},
	},
}

func optimiseAction(c *cli.Context) error {
	layoutFile := c.Args().First()
	corpusFile := c.String("corpus")
	pinsFile := c.String("pins")
	weightsConfig := c.String("weights")
	numGenerations := c.Uint("generations")
	acceptFunction := c.String("accept")

	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	if corpusFile == "" {
		return fmt.Errorf("corpus file is required")
	}

	if !slices.Contains(validAcceptFunctions, acceptFunction) {
		return fmt.Errorf("invalid accept function: %s. Must be one of: %v", acceptFunction, validAcceptFunctions)
	}

	if numGenerations <= 0 {
		return fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
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

	doOptimisation(lay, corp, pinsFile, weightsConfig, numGenerations, acceptFunction)
	return nil
}

func doOptimisation(lay *layout.SplitLayout, corp *layout.Corpus, pinsFile string, weightsConfig string, numGenerations uint, acceptFunction string) {
	fmt.Printf("Optimising layout: %s with corpus: %s, pins: %s, weights: %s, generations: %d, accept function: %s\n",
		lay.Name, corp.Name, pinsFile, weightsConfig, numGenerations, acceptFunction)

	if pinsFile != "" {
		err := lay.LoadPins("data/pins/" + pinsFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	best := lay.Optimise(corp, numGenerations, acceptFunction)
	fmt.Println(best)
	doHandUsage(best, corp)
	doSfbs(best, corp)
}

func doHandUsage(lay *layout.SplitLayout, corp *layout.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	fmt.Println(handInfo)
	// godump.Dump(handInfo)
}

func doSfbs(lay *layout.SplitLayout, corp *layout.Corpus) {
	sfbInfo := lay.AnalyzeSfbs(corp)
	fmt.Println(sfbInfo)
}
