package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

const (
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"
)

// Centralized flag map
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "specify the corpus file to calculate keyboard metrics",
		Value:   "default.txt",
	},
	"weights-file": &cli.StringFlag{
		Name:    "weights-file",
		Aliases: []string{"wf"},
		Usage:   "load weights from a text file; weights flag overrides these values",
		Value:   "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage:   "specify weights for metrics, e.g. sfb=-3.0,lsb=-2.0",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage:   "show metrics (viewing only): basic, extended, or fingers",
		Value:   "basic",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage:   "show delta rows: none, rows, median, or <some-layout.klf>",
		Value:   "none",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage:   "load pins from file; if no file specified, ~ and _ keys are pinned",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage:   "specify additional pins",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage:   "specify characters that are free to move (all others pinned)",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "specify the number of generations",
		Value:   250,
	},
	"accept-worse": &cli.StringFlag{
		Name:    "accept-worse",
		Aliases: []string{"aw"},
		Usage:   "specify the accept-worse function",
		Value:   "drop-slow",
	},
}

// Helper: convert selected flag keys to a slice
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

func main() {
	app := &cli.App{
		Name:  "layout-cli",
		Usage: "A CLI tool for various layout operations",
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

func validateFlags(_ *cli.Context) error {
	return nil
}

func loadCorpus(filename string) (*layout.Corpus, error) {
	if filename == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	return layout.NewCorpusFromFile(corpusName, path)
}

func loadLayout(filename string) (*layout.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout file is required")
	}
	layoutName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(layoutDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("layout file %s does not exist", path)
	}
	return layout.NewLayoutFromFile(layoutName, path)
}
