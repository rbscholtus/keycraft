package main

import (
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v2"
)

// analyseCommand defines the "analyse" CLI command.
// It prints detailed analysis for one or more layouts,
// optionally including data tables.
var analyseCommand = &cli.Command{
	Name:      "analyse",
	Aliases:   []string{"a"},
	Usage:     "Analyse one or more keyboard layouts in detail",
	Flags:     flagsSlice("rows", "corpus", "row-load", "finger-load", "pinky-penalties", "compact-trigrams", "trigram-rows"),
	ArgsUsage: "<layout1> <layout2> ...",
	Before:    validateAnalyseFlags,
	Action:    analyseAction,
}

// validateAnalyseFlags validates CLI flags before running the analyse command.
func validateAnalyseFlags(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}
	return nil
}

// analyseAction loads the specified corpus, load preferences, and layouts,
// then executes the analysis process.
// It returns an error if loading or analysis fails.
func analyseAction(c *cli.Context) error {
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
	result, err := kc.AnalyseLayouts(kc.AnalyseInput{
		LayoutFiles: layouts,
		Corpus:      corpus,
		Prefs:       prefs,
	})
	if err != nil {
		return err
	}

	// 3. Render results with display options
	displayOpts := kc.AnalyseDisplayOptions{
		MaxRows:         c.Int("rows"),
		CompactTrigrams: c.Bool("compact-trigrams"),
		TrigramRows:     c.Int("trigram-rows"),
	}

	return tui.RenderAnalyse(result, displayOpts)
}

