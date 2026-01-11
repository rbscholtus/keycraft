// keycraft is a CLI tool for analyzing and optimizing keyboard layouts.
//
// It provides commands for viewing layout metrics, comparing layouts,
// analyzing text corpora, and running optimization.
package main

import (
	"context"
	"fmt"
	"net/mail"
	"os"

	"github.com/urfave/cli/v3"
)

// Data directories relative to repository root.
var (
	layoutDir = "data/layouts/"
	corpusDir = "data/corpus/"
	configDir = "data/config/"
)

func main() {
	cmd := &cli.Command{
		Name:                  "keycraft",
		Version:               "v0.4.0",
		Usage:                 "A CLI tool for crafting better keyboard layouts",
		EnableShellCompletion: true,
		Suggest:               true,
		CommandNotFound:       customCommandNotFound,
		Description: "Keycraft is a CLI tool for analyzing, comparing, and " +
			"optimizing keyboard layouts. It evaluates layouts using a wide " +
			"range of metrics including same-finger bigrams (SFB), lateral " +
			"stretch bigrams (LSB), finger and row load distribution, and " +
			"trigram patterns. Keycraft can analyze text corpora to display " +
			"typing patterns, rank layouts against customizable weighted " +
			"metrics, display layout statistics with detailed tables, and " +
			"optimize layouts using the Breakout Local Search (BLS) algorithm " +
			"with configurable constraints like pinned keys and target " +
			"hand/finger/row loads. Keycraft can export analysis results in " +
			"multiple formats (table, HTML, CSV).",
		Authors: []any{
			&mail.Address{Name: "Barend Scholtus", Address: "barend.scholtus@gmail.com"},
		},
		Commands: []*cli.Command{
			corpusCommand,
			viewCommand,
			analyseCommand,
			rankCommand,
			flipCommand,
			optimiseCommand,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %[1]v\n", err)
		os.Exit(1)
	}
}

// customCommandNotFound provides a friendly error message with suggestions
// when a command is not found.
func customCommandNotFound(ctx context.Context, cmd *cli.Command, command string) {
	suggestion := cli.SuggestCommand(cmd.Commands, command)

	fmt.Fprintf(cmd.Root().Writer, "Command '%s' is not a thing.", command)

	if suggestion != "" {
		fmt.Fprintf(cmd.Root().Writer, " Did you mean: %s?\n", suggestion)
	} else {
		fmt.Fprintf(cmd.Root().Writer, " Run '%s help' to see available commands.\n", cmd.Name)
	}
}
