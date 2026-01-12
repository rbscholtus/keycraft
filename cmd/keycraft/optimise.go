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

// optimiseFlags are flags specific to the optimise command
var optimiseFlags = []cli.Flag{
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

// optimiseFlagsSlice returns all flags for the optimise command
func optimiseFlagsSlice() []cli.Flag {
	commonFlags := flagsSlice("corpus", "load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties", "weights-file", "weights")
	return append(commonFlags, optimiseFlags...)
}

// optimiseCommand defines the "optimise" CLI command for running Breakout Local Search (BLS)
// optimization on a keyboard layout.
var optimiseCommand = &cli.Command{
	Name:          "optimise",
	Aliases:       []string{"o"},
	Usage:         "Optimise a keyboard layout using Breakout Local Search (BLS)",
	Flags:         optimiseFlagsSlice(),
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

// optimiseAction manages the full optimization workflow: it builds the
// optimization input from CLI flags, executes the BLS algorithm, persists the
// best discovered layout, and generates a comparative ranking against the
// original layout.
func optimiseAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	input, err := buildOptimiseInput(c)
	if err != nil {
		return err
	}

	// Open log file if requested
	logFilePath := c.String("log-file")
	if logFilePath != "" {
		f, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to create log file %s: %v", logFilePath, err)
		}
		defer kc.CloseFile(f)
		input.LogFile = f
	}

	optResult, err := kc.OptimizeLayout(input, os.Stdout)
	if err != nil {
		return err
	}

	bestPath := filepath.Join(layoutDir, optResult.BestLayout.Name+".klf")
	if err := optResult.BestLayout.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
	}

	layoutsToCompare := []string{ensureKlf(optResult.OriginalLayout.Name), ensureKlf(optResult.BestLayout.Name)}

	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layoutsToCompare,
		Corpus:      input.Corpus,
		Targets:     input.Targets,
	})
	if err != nil {
		return fmt.Errorf("failed to perform layout analysis: %v", err)
	}

	if err := tui.RenderView(viewResult); err != nil {
		return fmt.Errorf("failed to render view: %v", err)
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
		return fmt.Errorf("failed to compute layout rankings: %v", err)
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
		return fmt.Errorf("failed to render layout rankings: %v", err)
	}

	return nil
}

// buildOptimiseInput gathers all input parameters for layout optimization.
func buildOptimiseInput(c *cli.Command) (kc.OptimiseInput, error) {
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.OptimiseInput{}, err
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return kc.OptimiseInput{}, err
	}

	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return kc.OptimiseInput{}, err
	}

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return kc.OptimiseInput{}, fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	maxTime := c.Uint("maxtime")
	if maxTime <= 0 {
		return kc.OptimiseInput{}, fmt.Errorf("maximum time must be above 0. Got: %d", maxTime)
	}

	layout, err := loadLayout(c.Args().First())
	if err != nil {
		return kc.OptimiseInput{}, err
	}

	pinsPath := c.String("pins-file")
	if pinsPath != "" {
		pinsPath = filepath.Join(configDir, pinsPath)
	}
	pinned, err := kc.LoadPinsFromParams(pinsPath, c.String("pins"), c.String("free"), layout)
	if err != nil {
		return kc.OptimiseInput{}, err
	}

	return kc.OptimiseInput{
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
