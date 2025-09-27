package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// validAcceptFuncs lists supported strategies for the accept-worse decision
// used during optimisation passes.
var validAcceptFuncs = []string{"always", "drop-slow", "linear", "drop-fast", "never"}

var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Aliases:   []string{"o"},
	Usage:     "Optimise a keyboard layout",
	ArgsUsage: "<layout.klf>",
	Flags:     flagsSlice("corpus", "weights-file", "weights", "pins-file", "pins", "free", "generations", "accept-worse"),
	Action:    optimiseAction,
}

// optimiseAction performs optimisation for a single layout file:
//   - loads corpus, weights and pins
//   - validates accept-function and generation count
//   - runs optimisation and persists the best layout
//   - runs analysis and ranking on original vs optimized layouts
func optimiseAction(c *cli.Context) error {
	// Load the corpus used for analysing layouts.
	corpus, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
	}
	weights, err := kc.NewWeightsFromParams(weightsPath, c.String("weights"))
	if err != nil {
		return err
	}

	acceptFunction := c.String("accept-worse")
	if !slices.Contains(validAcceptFuncs, acceptFunction) {
		return fmt.Errorf("invalid accept function: %s. Must be one of: %v", acceptFunction, validAcceptFuncs)
	}

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	if c.Args().Len() != 1 {
		return fmt.Errorf("expected exactly 1 layout file, got %d", c.Args().Len())
	}
	layoutFile := c.Args().First()
	layout, err := loadLayout(layoutFile)
	if err != nil {
		return err
	}

	pinsPath := c.String("pins-file")
	if pinsPath != "" {
		pinsPath = filepath.Join(pinsDir, pinsPath)
	}
	if err := layout.LoadPinsFromParams(pinsPath, c.String("pins"), c.String("free")); err != nil {
		return err
	}

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
	layoutsToCompare := []string{layoutFile, bestFilename}

	// Call DoAnalysis with the layouts
	if err := DoAnalysis(corpus, layoutsToCompare, false, 10); err != nil {
		return fmt.Errorf("failed to perform layout analysis: %v", err)
	}

	// Call DoLayoutRankings with the layouts
	if err := kc.DoLayoutRankings(corpus, layoutDir, layoutsToCompare, weights, "basic", "rows"); err != nil {
		return fmt.Errorf("failed to perform layout rankings: %v", err)
	}

	return nil
}
