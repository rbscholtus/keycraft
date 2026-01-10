package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// optimiseCommand defines the "optimise" CLI command for running Breakout Local Search (BLS)
// optimization on a keyboard layout.
var optimiseCommand = &cli.Command{
	Name:          "optimise",
	Aliases:       []string{"o"},
	Usage:         "Optimise a keyboard layout using Breakout Local Search (BLS)",
	Flags:         flagsSlice("corpus", "row-load", "finger-load", "pinky-penalties", "weights-file", "weights", "pins-file", "pins", "free", "generations", "maxtime", "seed", "log-file"),
	ArgsUsage:     "<layout>",
	Before:        validateOptFlags,
	Action:        optimiseAction,
	ShellComplete: layoutShellComplete,
}

// validateOptFlags validates CLI flags before running the optimise command.
func validateOptFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
	// Skip validation during shell completion
	// Check os.Args directly since -- prevents flag parsing
	if isShellCompletion() {
		return ctx, nil
	}

	if c.Args().Len() != 1 {
		return ctx, fmt.Errorf("expected exactly 1 layout, got %d", c.Args().Len())
	}
	return ctx, nil
}

// optimiseAction performs layout optimization using Breakout Local Search (BLS),
// then analyzes and ranks the original vs optimized layouts.
func optimiseAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if isShellCompletion() {
		return nil
	}

	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	prefs, err := loadPreferredLoadsFromFlags(c)
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

	seed := c.Int64("seed")

	layoutFile := c.Args().First()
	layout, err := loadLayout(layoutFile)
	if err != nil {
		return err
	}

	pinsPath := c.String("pins-file")
	if pinsPath != "" {
		pinsPath = filepath.Join(configDir, pinsPath)
	}
	pinned, err := kc.LoadPinsFromParams(pinsPath, c.String("pins"), c.String("free"), layout)
	if err != nil {
		return err
	}

	// Set up log file if specified
	var logFile io.Writer
	logFilePath := c.String("log-file")
	if logFilePath != "" {
		f, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to create log file %s: %v", logFilePath, err)
		}
		defer kc.CloseFile(f)
		logFile = f
	}

	// Run optimization with specified preferences
	best, err := kc.OptimizeLayoutBLS(
		layout,
		layoutDir,
		corpus,
		weights,
		prefs,               // load preferences
		pinned,              // pinned keys
		int(numGenerations), // max iterations
		int(maxTime),        // max time in minutes
		seed,                // random seed
		os.Stdout,           // console output
		logFile,             // JSONL log file
	)
	if err != nil {
		return err
	}

	// Save best layout to file
	bestPath := filepath.Join(layoutDir, best.Name+".klf")
	if err := best.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
	}

	// Prepare layouts for comparison view
	layoutsToCompare := []string{ensureKlf(layout.Name), ensureKlf(best.Name)}

	// View the before/after layouts
	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layoutsToCompare,
		Corpus:      corpus,
		Prefs:       prefs,
	})
	if err != nil {
		return fmt.Errorf("failed to perform layout analysis: %v", err)
	}

	if err := tui.RenderView(viewResult); err != nil {
		return fmt.Errorf("failed to render view: %v", err)
	}

	// Perform layout ranking using new architecture
	input := kc.RankingInput{
		LayoutsDir:  layoutDir,
		LayoutFiles: layoutsToCompare,
		Corpus:      corpus,
		Prefs:       prefs,
		Weights:     weights,
	}

	rankingResult, err := kc.ComputeRankings(input)
	if err != nil {
		return fmt.Errorf("failed to compute layout rankings: %v", err)
	}

	displayOpts := tui.RankingDisplayOptions{
		OutputFormat:   tui.OutputTable,
		MetricsOption:  tui.MetricsWeighted,
		ShowWeights:    true,
		Weights:        weights,
		DeltasOption:   tui.DeltasCustom,
		BaseLayoutName: layout.Name,
	}

	if err := tui.RenderRankingTable(rankingResult, displayOpts); err != nil {
		return fmt.Errorf("failed to render layout rankings: %v", err)
	}

	return nil
}
