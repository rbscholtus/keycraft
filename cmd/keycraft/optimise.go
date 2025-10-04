package main

import (
	"fmt"
	"path/filepath"
	"slices"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// validAcceptFuncs lists supported strategies for the accept-worse decision
// used during optimisation passes.
var validAcceptFuncs = []string{"always", "drop-slow", "linear", "drop-fast", "never"}

var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Aliases:   []string{"o"},
	Usage:     "Optimise a keyboard layout using simulated annealing",
	Flags:     flagsSlice("corpus", "finger-load", "weights-file", "weights", "pins-file", "pins", "free", "generations", "accept-worse"),
	ArgsUsage: "<layout>",
	Before:    validateOptFlags,
	Action:    optimiseAction,
}

// validateViewFlags validates CLI flags before running the view command.
func validateOptFlags(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("expected exactly 1 layout, got %d", c.Args().Len())
	}
	return nil
}

// optimiseAction performs optimisation for a single layout file:
//   - loads corpus, weights and pins
//   - validates accept-function and generation count
//   - runs optimisation and persists the best layout
//   - runs analysis and ranking on original vs optimized layouts
func optimiseAction(c *cli.Context) error {
	corpus, err := getCorpusFromFlag(c)
	if err != nil {
		return err
	}

	fingerBal, err := getFingerLoadFromFlag(c)
	if err != nil {
		return err
	}

	weights, err := loadWeightsFromFlags(c)
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

	best := layout.Optimise(corpus, fingerBal, weights, numGenerations, acceptFunction)

	// Save best layout to file
	name := ensureNoKlf(filepath.Base(layout.Name))
	bestFilename := fmt.Sprintf("%s-opt.klf", name)
	bestPath := filepath.Join(layoutDir, bestFilename)
	if err := best.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
	}

	// Prepare layouts for ranking
	layoutsToCompare := []string{layoutFile, bestFilename}

	// Call DoAnalysis with the layouts
	if err := DoAnalysis(layoutsToCompare, corpus, fingerBal, false, 0); err != nil {
		return fmt.Errorf("failed to perform layout analysis: %v", err)
	}

	// Call DoLayoutRankings with the layouts
	if err := kc.DoLayoutRankings(layoutDir, layoutsToCompare, corpus, fingerBal, weights, "basic", "rows"); err != nil {
		return fmt.Errorf("failed to perform layout rankings: %v", err)
	}

	return nil
}
