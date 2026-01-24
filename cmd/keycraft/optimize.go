package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// optimizeFlags are flags specific to the optimize command
var optimizeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage: "File specifying keys to pin during optimization. " +
			"Defaults to pinning '~' and '_'.",
		Category: "Optimization",
	},
	&cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage: "Additional characters to pin (e.g., 'aeiouy'). " +
			"Combined with pins-file.",
		Category: "Optimization",
	},
	&cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage: "Characters free to move during optimization. " +
			"All others are pinned.",
		Category: "Optimization",
	},
	&cli.UintFlag{
		Name:     "generations",
		Aliases:  []string{"gens", "g"},
		Usage:    "Number of optimization iterations to run.",
		Value:    1000,
		Category: "Optimization",
	},
	&cli.UintFlag{
		Name:     "maxtime",
		Aliases:  []string{"mt"},
		Usage:    "Maximum optimization time in minutes.",
		Value:    5,
		Category: "Optimization",
	},
	&cli.Int64Flag{
		Name:     "seed",
		Aliases:  []string{"s"},
		Usage:    "Random seed for reproducible results. Uses current timestamp if 0.",
		Value:    0,
		Category: "Optimization",
	},
	&cli.StringFlag{
		Name:     "log-file",
		Aliases:  []string{"lf"},
		Usage:    "JSONL log file path for detailed optimization metrics.",
		Category: "Optimization",
	},
}

// optimizeFlagsSlice returns all flags for the optimize command
func optimizeFlagsSlice() []cli.Flag {
	commonFlags := flagsSlice("corpus", "load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties", "weights-file", "weights")
	return append(commonFlags, optimizeFlags...)
}

// optimizeCommand defines the "optimize" CLI command for running Breakout Local Search (BLS)
// optimization on a keyboard layout.
var optimizeCommand = &cli.Command{
	Name:          "optimize",
	Aliases:       []string{"o"},
	Usage:         "Optimize a keyboard layout using Breakout Local Search (BLS)",
	Flags:         optimizeFlagsSlice(),
	ArgsUsage:     "<layout>",
	Before:        validateOptFlags,
	Action:        optimizeAction,
	ShellComplete: layoutShellComplete,
}

// validateOptFlags validates CLI flags before running the optimize command.
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

// optimizeAction manages the full optimization workflow: it builds the
// optimization input from CLI flags, executes the BLS algorithm, persists the
// best discovered layout, and generates a comparative ranking against the
// original layout.
func optimizeAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	input, err := buildOptimizeInput(c)
	if err != nil {
		return fmt.Errorf("could not parse user input: %w", err)
	}

	// Open log file if requested
	logFilePath := c.String("log-file")
	if logFilePath != "" {
		f, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("could not create log file %s: %w", logFilePath, err)
		}
		defer kc.CloseFile(f)
		input.LogFile = f
	}

	optResult, err := kc.OptimizeLayout(input, os.Stdout)
	if err != nil {
		return fmt.Errorf("could not optimize layout: %w", err)
	}

	origPath := filepath.Join(layoutDir, optResult.OriginalLayout.Name+".klf")
	bestPath := filepath.Join(layoutDir, optResult.BestLayout.Name+".klf")
	if err := optResult.BestLayout.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("could not save best layout to %s: %w", bestPath, err)
	}

	layoutsToCompare := []string{origPath, bestPath}
	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layoutsToCompare,
		Corpus:      input.Corpus,
		Targets:     input.Targets,
	})
	if err != nil {
		return fmt.Errorf("could not perform layout analysis: %w", err)
	}

	if err := tui.RenderView(viewResult); err != nil {
		return fmt.Errorf("could not render view: %w", err)
	}

	rankingInput := kc.RankingInput{
		LayoutsDir:  layoutDir,
		LayoutFiles: layoutsToCompare,
		Corpus:      input.Corpus,
		Targets:     input.Targets,
		Weights:     input.Weights,
	}

	rankingResult, err := kc.ComputeRankings(rankingInput)
	if err != nil {
		return fmt.Errorf("could not compute layout rankings: %w", err)
	}

	displayOpts := tui.RankingDisplayOptions{
		OutputFormat:   tui.OutputTable,
		MetricsOption:  tui.MetricsWeighted,
		ShowWeights:    true,
		Weights:        input.Weights,
		DeltasOption:   tui.DeltasCustom,
		BaseLayoutName: optResult.OriginalLayout.Name,
	}

	if err := tui.RenderRankingTable(rankingResult, displayOpts); err != nil {
		return fmt.Errorf("could not render layout rankings: %w", err)
	}

	return nil
}

// buildOptimizeInput gathers all input parameters for layout optimization.
func buildOptimizeInput(c *cli.Command) (kc.OptimizeInput, error) {
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.OptimizeInput{}, fmt.Errorf("could not load corpus: %w", err)
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return kc.OptimizeInput{}, fmt.Errorf("could not load target loads: %w", err)
	}

	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return kc.OptimizeInput{}, fmt.Errorf("could not load weights: %w", err)
	}

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return kc.OptimizeInput{}, fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	maxTime := c.Uint("maxtime")
	if maxTime <= 0 {
		return kc.OptimizeInput{}, fmt.Errorf("maximum time must be above 0. Got: %d", maxTime)
	}

	layout, err := loadLayout(c.Args().First())
	if err != nil {
		return kc.OptimizeInput{}, fmt.Errorf("could not load layout: %w", err)
	}

	pinsPath := c.String("pins-file")
	if pinsPath != "" {
		pinsPath = filepath.Join(configDir, pinsPath)
	}
	pinned, err := kc.LoadPinsFromParams(pinsPath, c.String("pins"), c.String("free"), layout)
	if err != nil {
		return kc.OptimizeInput{}, fmt.Errorf("could not load pins: %w", err)
	}

	return kc.OptimizeInput{
		Layout:         layout,
		LayoutsDir:     layoutDir,
		Corpus:         corpus,
		Targets:        targets,
		Weights:        weights,
		Pinned:         pinned,
		NumGenerations: int(numGenerations),
		MaxTime:        int(maxTime),
		Seed:           c.Int64("seed"),
	}, nil
}
