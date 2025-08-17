package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

const (
	layoutDir = "data/layouts/"
	corpusDir = "data/corpus/"
	pinsDir   = "data/pins/"
)

var corpusFlag = &cli.StringFlag{
	Name:    "corpus",
	Aliases: []string{"c"},
	Usage:   "specify the corpus file",
	Value:   "default.txt",
}

//	&cli.StringFlag{
//		Name:    "style",
//		Aliases: []string{"s"},
//		Usage:   "Specify the style (layoutsdoc or keysolve)",
//		Value:   "layoutsdoc",
//	},
var weightsFlag = &cli.StringFlag{
	Name:    "weights",
	Aliases: []string{"w"},
	Usage:   "specify weights for metrics, e.g. sfb=-3.0,lsb=-2.0",
	Value:   "",
}
var weightsFileFlag = &cli.StringFlag{
	Name:    "weights-file",
	Aliases: []string{"wf"},
	Usage:   "load weights from a text file; weights flag overrides these values",
	Value:   "",
}
var deltasFlag = &cli.StringFlag{
	Name:    "deltas",
	Aliases: []string{"d"},
	Usage:   "show delta rows: none, rows, median, or <some-layout.klf>",
	Value:   "none",
}
var metricsFlag = &cli.StringFlag{
	Name:    "metrics",
	Aliases: []string{"m"},
	Usage:   "choose metrics set: basic, extended, or fingers",
	Value:   "basic",
}

var pinsFileFlag = &cli.StringFlag{
	Name:     "pins-file",
	Aliases:  []string{"pf"},
	Usage:    "specify the pins file",
	Required: false,
}
var pinsFlag = &cli.StringFlag{
	Name:     "pins",
	Aliases:  []string{"p"},
	Usage:    "specify pins",
	Required: false,
}
var gensFlag = &cli.UintFlag{
	Name:     "generations",
	Aliases:  []string{"g"},
	Usage:    "specify the number of generations",
	Required: false,
	Value:    250,
}
var acceptFlag = &cli.StringFlag{
	Name:     "accept",
	Aliases:  []string{"a"},
	Usage:    "specify the accept function",
	Required: false,
	Value:    "drop-slow",
}

func main() {
	app := &cli.App{
		Name:  "layout-cli",
		Usage: "A CLI tool for various layout operations",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			viewCommand,
			rankCommand,
			analyseCommand,
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

func validateFlags(c *cli.Context) error {
	// style := c.String("style")
	// if style != "" && style != "layoutsdoc" && style != "keysolve" {
	// 	return cli.Exit("Invalid style. Supported styles are 'layoutsdoc' and 'keysolve'.", 1)
	// }
	return nil
}

func loadCorpus(corpusFile string) (*layout.Corpus, error) {
	if corpusFile == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusPath := filepath.Join(corpusDir, corpusFile)
	if _, err := os.Stat(corpusPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("corpus file %s does not exist", corpusPath)
	}
	return layout.NewCorpusFromFile(corpusFile, corpusPath)
}

func loadLayout(layoutFile string) (*layout.SplitLayout, error) {
	if layoutFile == "" {
		return nil, fmt.Errorf("layout file is required")
	}
	layoutPath := filepath.Join(layoutDir, layoutFile)
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("layout file %s does not exist", layoutPath)
	}
	return layout.NewLayoutFromFile(layoutFile, layoutPath)
}
