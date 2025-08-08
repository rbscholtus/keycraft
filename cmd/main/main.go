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
)

func main() {
	app := &cli.App{
		Name:  "layout-cli",
		Usage: "A CLI tool for various layout operations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "corpus",
				Aliases:  []string{"c"},
				Usage:    "specify the corpus file",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "style",
				Aliases: []string{"s"},
				Usage:   "Specify the style (layoutsdoc or keysolve)",
				Value:   "layoutsdoc",
			},
		},
		Commands: []*cli.Command{
			viewCommand,
			analyseCommand,
			compareCommand,
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

func validateFlags(c *cli.Context) error {
	style := c.String("style")
	if style != "" && style != "layoutsdoc" && style != "keysolve" {
		return cli.Exit("Invalid style. Supported styles are 'layoutsdoc' and 'keysolve'.", 1)
	}
	return nil
}

func loadCorpus(c *cli.Context) (*layout.Corpus, error) {
	corpusFile := c.String("corpus")
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
