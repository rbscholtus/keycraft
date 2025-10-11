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

// rankCommand defines the "rank" CLI command for the kb tool.
// It lists keyboard layouts from a specified directory, optionally filtered by filename.
// Users can:
//   - Filter which layouts are listed (by CLI args or all .klf files in the data/layouts directory)
//   - Display delta rows showing metric differences between layouts
//   - Show extended metrics for deeper analysis
//   - Apply metric weights directly or from a file
//
// By default, layouts are shown in CLI-specified order unless `--show-deltas` is enabled,
// in which case the default order switches to 'rank'.
var rankCommand = &cli.Command{
	Name:      "rank",
	Aliases:   []string{"r"},
	Usage:     "Rank keyboard layouts and optionally view deltas",
	Flags:     flagsSlice("metrics", "deltas", "corpus", "finger-load", "weights-file", "weights"),
	ArgsUsage: "<layout1> <layout2> ...",
	Action:    rankAction,
}

// rankAction is the CLI action handler for the "rank" command.
// It processes user inputs, loads layouts and weights, validates flags,
// and executes the layout ranking display.
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

// getMetricsFromFlag validates the --metrics flag against allowed sets
// Returns the validated metric mode in lowercase or an error if invalid.
func getMetricsFromFlag(c *cli.Context) (string, error) {
	m := strings.ToLower(c.String("metrics"))

	if _, ok := kc.MetricsMap[m]; !ok {
		opts := slices.Collect(maps.Keys(kc.MetricsMap))
		return "", fmt.Errorf("invalid metrics mode %q; must be one of %v", m, opts)
	}

	return m, nil
}

// getDeltasFromFlag validates the --deltas flag.
// It must be one of "none", "rows", "median", or else is treated as a layout name (without .klf).
func getDeltasFromFlag(c *cli.Context) (deltas string, baseLayout string, err error) {
	val := c.String("deltas")
	lower := strings.ToLower(val)

	switch lower {
	case "none", "rows", "median":
		deltas = lower
	default:
		// treat as layout name (case preserved for filesystem lookup)
		deltas = ensureNoKlf(val)
		baseLayout = deltas
	}

	return
}

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

// filesExist checks that all specified layout files exist in the layoutDir.
// This ensures that user-specified layouts are valid before processing.
// func filesExist(layouts []string) error {
// 	for _, layoutFile := range layouts {
// 		path := filepath.Join(layoutDir, layoutFile)
// 		if _, err := os.Stat(path); err != nil {
// 			if os.IsNotExist(err) {
// 				return fmt.Errorf("layout file %s does not exist in %s", layoutFile, layoutDir)
// 			}
// 			return fmt.Errorf("error checking layout file %s: %v", layoutFile, err)
// 		}
// 	}
// 	return nil
// }

// allLayoutFiles returns all `.klf` files in layoutDir.
func allLayoutFiles() ([]string, error) {
	var layoutsToCmp []string

	// Append all .klf files found in layoutDir.
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
