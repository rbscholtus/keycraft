package main

import (
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v2"
)

// corpusCommand defines the CLI command for displaying corpus statistics.
var corpusCommand = &cli.Command{
	Name:    "corpus",
	Aliases: []string{"c"},
	Usage:   "Display statistics for a text corpus",
	Flags:   flagsSlice("corpus", "corpus-rows", "coverage"),
	Action:  corpusAction,
}

// corpusAction loads the specified corpus and displays its statistics.
func corpusAction(c *cli.Context) error {
	// 1. Build input from CLI flags
	input, err := buildCorpusInput(c)
	if err != nil {
		return err
	}

	// 2. Process (business logic in internal/keycraft/)
	result, err := kc.DisplayCorpus(input)
	if err != nil {
		return err
	}

	// 3. Render (presentation layer in tui package)
	return tui.RenderCorpus(result)
}

// buildCorpusInput gathers all input parameters for corpus display.
func buildCorpusInput(c *cli.Context) (kc.CorpusInput, error) {
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return kc.CorpusInput{}, err
	}

	nrows := c.Int("rows")

	return kc.CorpusInput{
		Corpus: corpus,
		NRows:  nrows,
	}, nil
}
