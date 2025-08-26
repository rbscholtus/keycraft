// Package main provides the CLI entrypoint and helper functions for the
// keycraft command-line tool.
//
// view.go implements the "view" command for the keycraft CLI; it loads a corpus
// and analyzes one or more keyboard layout files for display.
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
	"log"
	"os"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// Data directories used by the CLI (relative to repository root).
const (
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"
)

// Centralized map of CLI flags used across commands.
// Keeps flag definitions in one place so commands can select only the flags they need.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "corpus file to calculate keyboard metrics",
		Value:   "default.txt",
	},
	"weights-file": &cli.StringFlag{
		Name:    "weights-file",
		Aliases: []string{"wf"},
		Usage:   "text file containing weights for scoring layouts; weights flag overrides these values",
		Value:   "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage:   "weights for for scoring layouts, eg: sfb=-3.0,lsb=-2.0",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage:   "metrics to show: basic, extended, or fingers",
		Value:   "basic",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage:   "deltas to show: none, rows, median, or <layout.klf>",
		Value:   "none",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage:   "text file containing keys to pin to their current position; if no file is specified, ~ keys and _ are pinned",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage:   "additional characters to pin, eg: aeiouy",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage:   "characters free to be moved (all others pinned), eg: zqjx",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "number of generations",
		Value:   250,
	},
	"accept-worse": &cli.StringFlag{
		Name:    "accept-worse",
		Aliases: []string{"aw"},
		Usage:   fmt.Sprintf("accept-worse function (how likely is it a worse layout is accepted): %v", validAcceptFuncs),
		Value:   "drop-slow",
	},
}

// Helper: convert selected flag keys to a slice.
// Used by commands to pick a subset of flags from appFlagsMap.
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

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
			experimentCommand,
		},
		Before: func(c *cli.Context) error {
			return validateFlags(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// validateFlags is called by urfave/cli Before hook.
// Keep it small: perform early validation of flags here (no-op for now).
func validateFlags(_ *cli.Context) error {
	return nil
}

// loadCorpus loads a corpus by filename from the corpusDir and returns a Corpus.
// It trims the extension to produce a corpus name and delegates loading to kc.NewCorpusFromFile.
func loadCorpus(filename string) (*kc.Corpus, error) {
	if filename == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	return kc.NewCorpusFromFile(corpusName, path)
}

// loadLayout loads a layout by filename from layoutDir and returns a SplitLayout.
// It validates presence of the file and uses kc.NewLayoutFromFile to create the layout.
func loadLayout(filename string) (*kc.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout file is required")
	}
	layoutName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(layoutDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("layout file %s does not exist", path)
	}
	return kc.NewLayoutFromFile(layoutName, path)
}
