// keycraft is a CLI tool for analyzing and optimizing keyboard layouts.
//
// It provides commands for viewing layout metrics, comparing layouts,
// analyzing text corpora, and running optimization.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Data directories relative to repository root.
var (
	layoutDir = "data/layouts/"
	corpusDir = "data/corpus/"
	configDir = "data/config/"
	pinsDir   = "data/pins/"
)

func main() {
	app := &cli.App{
		Name:  "keycraft",
		Usage: "A CLI tool for crafting better keyboard layouts",
		Commands: []*cli.Command{
			corpusCommand,
			viewCommand,
			analyseCommand,
			rankCommand,
			optimiseCommand,
			flipCommand,
			// experimentCommand (under development)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
