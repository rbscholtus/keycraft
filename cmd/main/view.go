package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var viewCommand = &cli.Command{
	Name:      "view",
	Aliases:   []string{"v"},
	Usage:     "View a layout file with a corpus file",
	ArgsUsage: "<layout file>",
	Action:    viewAction,
	Flags:     flagsSlice("corpus"),
}

func viewAction(c *cli.Context) error {
	corp, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	if c.NArg() < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	if err := DoAnalysis(corp, c.Args().Slice(), false); err != nil {
		return err
	}

	return nil
}
