package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// viewCommand defines the CLI command for viewing keyboard layout analysis.
var viewCommand = &cli.Command{
	Name:      "view",
	Aliases:   []string{"v"},
	Usage:     "Analyse and display one or more keyboard layouts",
	Flags:     flagsSlice("corpus", "finger-load"),
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
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	fingerLoad, err := getFingerLoadFromFlag(c)
	if err != nil {
		return err
	}

	layouts := getLayoutArgs(c)

	// Analyse all provided layouts using given corpus.
	// The 'false' parameter indicates not to include detailed metrics.
	if err := DoAnalysis(layouts, corpus, fingerLoad, false, 0); err != nil {
		return err
	}

	return nil
}
