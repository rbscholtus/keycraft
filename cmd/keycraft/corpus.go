package main

import (
	"context"
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// corpusFlags defines flags specific to the corpus command.
var corpusFlags = []cli.Flag{
	&cli.IntFlag{
		Name:     "corpus-rows",
		Aliases:  []string{"cr"},
		Usage:    "Maximum number of rows to display in corpus data tables.",
		Value:    100,
		Category: "Display",
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if isShellCompletion() {
				return nil
			}
			if value < 1 {
				return fmt.Errorf("--corpus-rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	&cli.Float64Flag{
		Name: "coverage",
		Usage: "Corpus word coverage percentage (0.1-100.0). Filters " +
			"low-frequency words. Forces cache rebuild.",
		Value:    98.0,
		Category: "",
		Action: func(ctx context.Context, c *cli.Command, value float64) error {
			if isShellCompletion() {
				return nil
			}
			if value < 0.1 || value > 100.0 {
				return fmt.Errorf("--coverage must be 0.1-100 (got %f)", value)
			}
			return nil
		},
	},
}

// corpusFlagsSlice returns all flags for the corpus command.
func corpusFlagsSlice() []cli.Flag {
	commonFlags := flagsSlice("corpus")
	return append(commonFlags, corpusFlags...)
}

// corpusCommand defines the CLI command for displaying corpus statistics.
var corpusCommand = &cli.Command{
	Name:          "corpus",
	Aliases:       []string{"c"},
	Usage:         "Display statistics for a text corpus",
	Flags:         corpusFlagsSlice(),
	Before:        validateCorpusFlags,
	Action:        corpusAction,
	ShellComplete: layoutShellComplete,
}

// validateCorpusFlags validates CLI flags before running the corpus command.
func validateCorpusFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
	// Skip validation during shell completion
	// Check os.Args directly since -- prevents flag parsing
	if isShellCompletion() {
		return ctx, nil
	}

	// Skip validation if help is requested
	// The framework handles help display, but we need to allow it through validation
	if c.NArg() == 1 && c.Args().First() == "help" {
		return ctx, nil
	}

	// Corpus command takes no arguments, only flags
	if c.NArg() != 0 {
		return ctx, fmt.Errorf("corpus command takes no arguments, got %d. Did you mean: '--corpus %s'?", c.NArg(), c.Args().First())
	}

	return ctx, nil
}

// corpusAction processes a text corpus to extract and display n-gram frequency
// statistics, optionally applying word coverage filtering to prune low-frequency
// vocabulary.
func corpusAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if isShellCompletion() {
		return nil
	}

	// 1. Build input from CLI flags
	input, err := buildCorpusInput(c)
	if err != nil {
		return fmt.Errorf("could not parse user input: %w", err)
	}

	// 2. Process (business logic in internal/keycraft/)
	result, err := kc.DisplayCorpus(input)
	if err != nil {
		return fmt.Errorf("could not display corpus: %w", err)
	}

	// 3. Render (presentation layer in tui package)
	return tui.RenderCorpus(result)
}

// buildCorpusInput gathers all input parameters for corpus display.
func buildCorpusInput(c *cli.Command) (kc.CorpusInput, error) {
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.CorpusInput{}, fmt.Errorf("could not load corpus: %w", err)
	}

	nrows := c.Int("corpus-rows")

	return kc.CorpusInput{
		Corpus: corpus,
		NRows:  nrows,
	}, nil
}
