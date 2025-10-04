// Package main provides the CLI entrypoint and helper functions for the
// keycraft command-line tool.
//
// view.go implements the "view" command for the keycraft CLI; it loads a corpus
// and analyses one or more keyboard layout files for display.
//
// analyse.go contains functions to analyse keyboard layouts and render
// human-friendly tables summarising hand/row usage and other metrics.
//
// rank.go provides the implementation for the "rank" command in the kb CLI tool.
// It allows users to compare keyboard layouts based on various metrics and user-defined weights.
// The command can compare specific layouts or all layouts in a directory, and supports ordering
// of results and custom metric weighting.
//
// optimise.go implements the optimise command which runs simulated
// optimisation on a layout using corpus, pins and weight configuration.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Data directories used by the CLI (relative to repository root).
var (
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"
)

// main sets up the CLI application and registers commands.
// Validation hooks run before command execution (validateFlags).
func main() {
	app := &cli.App{
		Name:  "keycraft",
		Usage: "A CLI tool for crafting better keyboard layouts",
		Commands: []*cli.Command{
			viewCommand,
			analyseCommand,
			rankCommand,
			optimiseCommand,
			// experimentCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
