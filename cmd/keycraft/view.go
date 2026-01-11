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
	Before:        validateViewFlags,
	Action:        viewAction,
	ShellComplete: layoutShellComplete,
}

// validateViewFlags validates CLI flags before running the view command.
func validateViewFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
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

// viewAction loads data and performs layout analysis.
func viewAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if isShellCompletion() {
		return nil
	}

	// 1. Load inputs
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return err
	}

	layouts := getLayoutArgs(c)

	// 2. Perform computation (pure, no display concerns)
	result, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layouts,
		Corpus:      corpus,
		Targets:     targets,
	})
	if err != nil {
		return err
	}

	// 3. Render results
	return tui.RenderView(result)
}
