package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// viewCommand defines the CLI command for viewing and analysing keyboard layouts.
// It supports analysing one or more layouts using a specified corpus of text.
var viewCommand = &cli.Command{
	Name:      "view",
	Aliases:   []string{"v"},
	Usage:     "Analyse and display one or more keyboard layouts",
	ArgsUsage: "<layout1.klf> <layout2.klf> ...",
	Action:    viewAction,
	Flags:     flagsSlice("corpus"),
}

// viewAction implements the view command's functionality: loading corpus,
// validating layouts, and performing analysis.
func viewAction(c *cli.Context) error {
	corp, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// Analyse all provided layouts using the corpus.
	// The 'false' parameter indicates not to include detailed metrics.
	if err := DoAnalysis(corp, c.Args().Slice(), false, 0); err != nil {
		return err
	}

	return nil
}
