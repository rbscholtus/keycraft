package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	l "github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var validAcceptFunctions = []string{"always", "drop-slow", "linear", "drop-fast", "never"}

var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Usage:     "Optimise a layout file with a corpus, pins, weights, generations, and accept function",
	ArgsUsage: "<layout file>",
	Flags:     flagsSlice("corpus", "weights-file", "weights", "pins-file", "pins", "generations", "accept-func"),
	Action:    optimiseAction,
}

func optimiseAction(c *cli.Context) error {
	// Load the corpus used for analyzing layouts.
	corpus, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
	}
	weights, err := l.NewWeightsFromParams(weightsPath, c.String("weights"))
	if err != nil {
		return err
	}

	acceptFunction := c.String("accept-func")
	if !slices.Contains(validAcceptFunctions, acceptFunction) {
		return fmt.Errorf("invalid accept function: %s. Must be one of: %v", acceptFunction, validAcceptFunctions)
	}

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	layoutFile := c.Args().First()
	if layoutFile == "" {
		return fmt.Errorf("layout file is required")
	}

	layout, err := loadLayout(layoutFile)
	if err != nil {
		return err
	}

	pinsPath := c.String("pins-file")
	if pinsPath != "" {
		pinsPath = filepath.Join(pinsDir, pinsPath)
	}
	if err := layout.LoadPinsFromParams(pinsPath, c.String("pins")); err != nil {
		return err
	}

	// fmt.Printf("Optimising layout: %s with corpus: %s, pins: %s, weights: %s, generations: %d, accept function: %s\n",
	// 	lay.Name, corpus.Name, pinsFile, weightsConfig, numGenerations, acceptFunction)
	// fmt.Println(layout.Pinned)
	best := layout.Optimise(corpus, weights, numGenerations, acceptFunction)

	// Save best layout to file
	name := filepath.Base(layout.Name)
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".klf" {
		name = name[:len(name)-len(ext)]
	}
	bestFilename := fmt.Sprintf("%s-opt.klf", name)
	bestPath := filepath.Join(layoutDir, bestFilename)
	if err := best.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
	}

	// Prepare layouts for ranking
	layoutsToRank := []string{layoutFile, bestFilename}

	// Call DoLayoutRankings with the layouts
	if err := l.DoLayoutRankings(corpus, layoutDir, layoutsToRank, weights, "basic", "rows"); err != nil {
		return fmt.Errorf("failed to perform layout rankings: %v", err)
	}

	fmt.Println(layout)
	// an := layout.NewAnalyser(lay, corpus, "layoutsdoc")
	// fmt.Println(an.MetricsString())
	fmt.Println(best)
	// anBest := layout.NewAnalyser(best, corpus, "layoutsdoc")
	// fmt.Println(anBest.MetricsString())

	return nil
}
