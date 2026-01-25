package main

import (
	"context"
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// viewCommand defines the CLI command for viewing keyboard layout analysis.
var viewCommand = &cli.Command{
	Name:          "view",
	Aliases:       []string{"v"},
	Usage:         "Analyse and display one or more keyboard layouts",
	Flags:         flagsSlice("corpus", "load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties"),
	ArgsUsage:     "<layout1> <layout2> ...",
	Action:        viewAction,
	ShellComplete: layoutShellComplete,
}

// viewAction gathers the necessary corpus and target load parameters, performs
// a high-level ergonomic analysis of the specified layouts, and renders a summary view.
func viewAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	input, err := buildViewInput(c)
	if err != nil {
		return fmt.Errorf("could not parse user input: %w", err)
	}

	result, err := kc.ViewLayouts(input)
	if err != nil {
		return fmt.Errorf("could not view layouts: %w", err)
	}

	return tui.RenderView(result)
}

// buildViewInput gathers all input parameters for layout viewing.
func buildViewInput(c *cli.Command) (kc.ViewInput, error) {
	if c.NArg() < 1 {
		return kc.ViewInput{}, fmt.Errorf("need at least 1 layout")
	}

	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.ViewInput{}, fmt.Errorf("could not load corpus: %w", err)
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return kc.ViewInput{}, fmt.Errorf("could not load target loads: %w", err)
	}

	return kc.ViewInput{
		LayoutFiles: getLayoutArgs(c),
		Corpus:      corpus,
		Targets:     targets,
	}, nil
}
