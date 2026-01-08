package main

import (
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v2"
)

// viewCommand defines the CLI command for viewing keyboard layout analysis.
var viewCommand = &cli.Command{
	Name:      "view",
	Aliases:   []string{"v"},
	Usage:     "Analyse and display one or more keyboard layouts",
	Flags:     flagsSlice("corpus", "row-load", "finger-load", "pinky-penalties"),
	ArgsUsage: "<layout1> <layout2> ...",
	Before:    validateViewFlags,
	Action:    viewAction,
}

// validateViewFlags validates CLI flags before running the view command.
func validateViewFlags(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}
	return nil
}

// viewAction loads data and performs layout analysis.
func viewAction(c *cli.Context) error {
	// 1. Load inputs
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	prefs, err := loadPreferredLoadsFromFlags(c)
	if err != nil {
		return err
	}

	layouts := getLayoutArgs(c)

	// 2. Perform computation (pure, no display concerns)
	result, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layouts,
		Corpus:      corpus,
		Prefs:       prefs,
	})
	if err != nil {
		return err
	}

	// 3. Render results
	return tui.RenderView(result)
}
