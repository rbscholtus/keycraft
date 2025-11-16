package main

import (
	"fmt"
	"os"
	"path/filepath"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// optimiseCommand defines the "optimise" CLI command for running simulated annealing
// optimization on a keyboard layout.
var optimiseCommand = &cli.Command{
	Name:      "optimise",
	Aliases:   []string{"o"},
	Usage:     "Optimise a keyboard layout using simulated annealing",
	Flags:     flagsSlice("corpus", "row-load", "finger-load", "weights-file", "weights", "pins-file", "pins", "free", "generations", "maxtime"),
	ArgsUsage: "<layout>",
	Before:    validateOptFlags,
	Action:    optimiseAction,
}

// validateOptFlags validates CLI flags before running the optimise command.
func validateOptFlags(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("expected exactly 1 layout, got %d", c.Args().Len())
	}
	return nil
}

// optimiseAction performs layout optimization using simulated annealing,
// then analyzes and ranks the original vs optimized layouts.
func optimiseAction(c *cli.Context) error {
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	rowBal, err := getRowLoadFromFlag(c)
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

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	maxTime := c.Uint("maxtime")
	if maxTime <= 0 {
		return fmt.Errorf("maximum time must be above 0. Got: %d", maxTime)
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
	pinned, err := kc.LoadPinsFromParams(pinsPath, c.String("pins"), c.String("free"), layout)
	if err != nil {
		return err
	}

	// Run optimization with specified row and finger balance
	best, err := kc.OptimizeLayoutBLS(
		layout,
		layoutDir,
		corpus,
		weights,
		rowBal,              // ideal row load distribution
		fingerBal,           // ideal finger load distribution
		pinned,              // pinned keys
		int(numGenerations), // max iterations
		int(maxTime),        // max time in minutes
		os.Stdout,           // progress output
	)
	if err != nil {
		return err
	}

	// Save best layout to file
	bestPath := filepath.Join(layoutDir, best.Name+".klf")
	if err := best.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
	}

	// Prepare layouts for ranking
	layoutsToCompare := []string{layout.Name, best.Name}

	// Call DoAnalysis with the layouts
	if err := DoAnalysis(layoutsToCompare, corpus, rowBal, fingerBal, false, 0); err != nil {
		return fmt.Errorf("failed to perform layout analysis: %v", err)
	}

	// Call DoLayoutRankings with the layouts
	if err := kc.DoLayoutRankings(layoutDir, layoutsToCompare, corpus, rowBal, fingerBal, weights, "extended", layout.Name); err != nil {
		return fmt.Errorf("failed to perform layout rankings: %v", err)
	}

	return nil
}
