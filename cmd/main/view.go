package main

import (
	"fmt"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var viewCommand = &cli.Command{
	Name:      "view",
	Usage:     "View a layout file with a corpus file",
	ArgsUsage: "<layout file>",
	Action:    viewAction,
}

func viewAction(c *cli.Context) error {
	corp, err := loadCorpus(c)
	if err != nil {
		return err
	}

	lay, err := loadLayout(c.Args().First())
	if err != nil {
		return err
	}

	style := c.String("style")
	doViewLayout(lay, corp, style)
	return nil
}

func doViewLayout(lay *layout.SplitLayout, corp *layout.Corpus, style string) {
	fmt.Println(lay)
	an := layout.NewAnalyser(lay, corp, style)
	// fmt.Println(an.HandUsageString())
	// fmt.Println(an.RowUsageString())
	fmt.Println(an.MetricsString())
}
