package main

import (
	"context"
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// analyseFlags defines flags specific to the analyse command.
var analyseFlags = []cli.Flag{
	&cli.IntFlag{
		Name:     "rows",
		Aliases:  []string{"r"},
		Usage:    "Maximum number of rows to display in data tables.",
		Value:    10,
		Category: "Display",
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if isShellCompletion() {
				return nil
			}
			if value < 1 {
				return fmt.Errorf("--rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	&cli.BoolFlag{
		Name:     "compact-trigrams",
		Usage:    "Omit common trigram categories (ALT-NML, 2RL-IN, 2RL-OUT, 3RL-IN, 3RL-OUT) from trigram table.",
		Value:    false,
		Category: "Display",
	},
	&cli.IntFlag{
		Name:     "trigram-rows",
		Usage:    "Maximum number of trigrams to display in trigram table.",
		Value:    50,
		Category: "Display",
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if isShellCompletion() {
				return nil
			}
			if value < 1 {
				return fmt.Errorf("--trigram-rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
}

// analyseFlagsSlice returns all flags for the analyse command.
func analyseFlagsSlice() []cli.Flag {
	commonFlags := flagsSlice("corpus", "load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties")
	return append(commonFlags, analyseFlags...)
}

// analyseCommand defines the "analyse" CLI command.
// It prints detailed analysis for one or more layouts,
// optionally including data tables.
var analyseCommand = &cli.Command{
	Name:          "analyse",
	Aliases:       []string{"a"},
	Usage:         "Analyse one or more keyboard layouts in detail",
	Flags:         analyseFlagsSlice(),
	ArgsUsage:     "<layout1> <layout2> ...",
	Before:        validateAnalyseFlags,
	Action:        analyseAction,
	ShellComplete: layoutShellComplete,
}

// validateAnalyseFlags validates CLI flags before running the analyse command.
func validateAnalyseFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
	// Skip validation during shell completion
	// Check os.Args directly since -- prevents flag parsing
	if isShellCompletion() {
		return ctx, nil
	}

	if c.NArg() < 1 {
		return ctx, fmt.Errorf("need at least 1 layout")
	}
	return ctx, nil
}

// analyseAction coordinates the loading of corpus and target data, executes a
// detailed ergonomic analysis for one or more layouts, and renders the results.
func analyseAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	input, err := buildAnalyseInput(c)
	if err != nil {
		return err
	}

	result, err := kc.AnalyseLayouts(input)
	if err != nil {
		return err
	}

	displayOpts := kc.AnalyseDisplayOptions{
		MaxRows:         c.Int("rows"),
		CompactTrigrams: c.Bool("compact-trigrams"),
		TrigramRows:     c.Int("trigram-rows"),
	}

	return tui.RenderAnalyse(result, displayOpts)
}

// buildAnalyseInput gathers all input parameters for layout analysis.
func buildAnalyseInput(c *cli.Command) (kc.AnalyseInput, error) {
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.AnalyseInput{}, err
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return kc.AnalyseInput{}, err
	}

	return kc.AnalyseInput{
		LayoutFiles: getLayoutArgs(c),
		Corpus:      corpus,
		TargetLoads: targets,
	}, nil
}
