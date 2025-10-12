package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// rankCommand defines the "rank" CLI command for comparing and ranking layouts.
// It supports filtering layouts, displaying metric deltas, and applying custom weights.
var rankCommand = &cli.Command{
	Name:      "rank",
	Aliases:   []string{"r"},
	Usage:     "Rank keyboard layouts and optionally view deltas",
	Flags:     flagsSlice("metrics", "deltas", "corpus", "finger-load", "weights-file", "weights"),
	ArgsUsage: "<layout1> <layout2> ...",
	Action:    rankAction,
}

// rankAction handles the rank command, loading data and displaying layout rankings.
func rankAction(c *cli.Context) error {
	metrics, err := getMetricsFromFlag(c)
	if err != nil {
		return err
	}

	deltas, baseLayout, err := getDeltasFromFlag(c)
	if err != nil {
		return err
	}

	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return err
	}

	fingerBal, err := getFingerLoadFromFlag(c)
	if err != nil {
		return err
	}

	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return err
	}

	layouts, err := getLayoutsFromArgs(c, baseLayout)
	if err != nil {
		return err
	}

	// Perform the layout comparison and display results
	return kc.DoLayoutRankings(layoutDir, layouts, corpus, fingerBal, weights, metrics, deltas)
}

// getMetricsFromFlag validates the --metrics flag and returns the metric set name.
func getMetricsFromFlag(c *cli.Context) (string, error) {
	m := strings.ToLower(c.String("metrics"))

	if _, ok := kc.MetricsMap[m]; !ok {
		opts := slices.Collect(maps.Keys(kc.MetricsMap))
		return "", fmt.Errorf("invalid metrics mode %q; must be one of %v", m, opts)
	}

	return m, nil
}

// getDeltasFromFlag parses the --deltas flag.
// Returns "none", "rows", "median", or a layout name to use as baseline.
func getDeltasFromFlag(c *cli.Context) (deltas string, baseLayout string, err error) {
	val := c.String("deltas")
	lower := strings.ToLower(val)

	switch lower {
	case "none", "rows", "median":
		deltas = lower
	default:
		deltas = ensureNoKlf(val)
		baseLayout = deltas
	}

	return
}

// getLayoutsFromArgs returns layouts from CLI args, or all .klf files if no args provided.
func getLayoutsFromArgs(c *cli.Context, baseLayout string) ([]string, error) {
	var layouts []string
	if c.Args().Len() == 0 {
		var err error
		layouts, err = allLayoutFiles()
		if err != nil {
			return nil, err
		}
	} else {
		layouts = c.Args().Slice()
		for i := range layouts {
			layouts[i] = ensureKlf(layouts[i])
		}
	}

	if baseLayout != "" {
		baseLayout = ensureKlf(baseLayout)
		if !slices.Contains(layouts, baseLayout) {
			layouts = append(layouts, baseLayout)
		}
	}

	return layouts, nil
}

// allLayoutFiles returns all .klf files in layoutDir.
func allLayoutFiles() ([]string, error) {
	var layoutsToCmp []string

	entries, err := os.ReadDir(layoutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read layout directory %s: %v", layoutDir, err)
	}
	for _, entry := range entries {
		entryName := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entryName), ".klf") {
			layoutsToCmp = append(layoutsToCmp, entryName)
		}
	}

	return layoutsToCmp, nil
}
