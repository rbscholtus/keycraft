// rank.go provides the implementation for the "rank" command in the kb CLI tool.
// It allows users to compare keyboard layouts based on various metrics and user-defined weights.
// The command can compare specific layouts or all layouts in a directory, and supports ordering
// of results and custom metric weighting.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	kc "github.com/rbscholtus/kb/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// rankCommand defines the "rank" CLI command for the kb tool.
// It lists keyboard layouts from a specified directory, optionally filtered by filename.
// Users can:
//   - Filter which layouts are listed (by CLI args or all .klf files in the data/layouts directory)
//   - Apply metric weights directly or from a file
//   - Display delta rows showing metric differences between layouts
//   - Show extended metrics for deeper analysis
//
// By default, layouts are shown in CLI-specified order unless `--show-deltas` is enabled,
// in which case the default order switches to 'rank'.
var rankCommand = &cli.Command{
	Name:    "rank",
	Aliases: []string{"r"},
	Usage:   "Rank keyboard layouts with optional delta rows",
	Flags:   flagsSlice("corpus", "weights-file", "weights", "metrics", "deltas"),
	Action:  rankAction,
}

// rankAction is the CLI action handler for the "rank" command.
// It processes user inputs, loads layouts and weights, validates flags,
// and executes the layout ranking display.
func rankAction(c *cli.Context) error {
	// Load the corpus used for analyzing layouts.
	corpus, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
	}
	weights, err := kc.NewWeightsFromParams(weightsPath, c.String("weights"))
	if err != nil {
		return err
	}

	// Validate the --deltas flag; must be 'none', 'rows', 'median', or a KLF filename.
	deltas := c.String("deltas")
	deltasLow := strings.ToLower(deltas)
	baseFile := ""
	if strings.HasSuffix(deltasLow, ".klf") {
		baseFile = deltas
	}

	valid := deltasLow == "none" || deltasLow == "rows" || deltasLow == "median" || baseFile != ""
	if !valid {
		return fmt.Errorf("invalid deltas %q; must be none, rows, median, or a .klf filename", deltas)
	}

	// Determine which layouts to compare.
	// If no CLI args, gather all layouts plus baseFile if specified.
	// Otherwise, use provided layout filenames and ensure baseFile is included.
	var layoutsToCmp []string
	if c.Args().Len() == 0 {
		layoutsToCmp, err = allLayoutsAnd(baseFile)
		if err != nil {
			return err
		}
	} else {
		layoutsToCmp = c.Args().Slice()
		if baseFile != "" && !slices.Contains(layoutsToCmp, baseFile) {
			layoutsToCmp = append(layoutsToCmp, baseFile)
		}
		if err := filesExist(layoutsToCmp); err != nil {
			return err
		}
	}

	// Perform the layout comparison and display results,
	return kc.DoLayoutRankings(corpus, layoutDir, layoutsToCmp, weights, c.String("metrics"), deltas)
}

// filesExist checks that all specified layout files exist in the layoutDir.
// This ensures that user-specified layouts are valid before processing.
func filesExist(layoutsToCmp []string) error {
	for _, layoutName := range layoutsToCmp {
		path := filepath.Join(layoutDir, layoutName)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("layout file %s does not exist in %s", layoutName, layoutDir)
			}
			return fmt.Errorf("error checking layout file %s: %v", layoutName, err)
		}
	}
	return nil
}

// allLayoutsAnd gathers all .klf layout files from layoutDir,
// optionally including a specified base layout file at the start.
func allLayoutsAnd(baseFile string) ([]string, error) {
	var layoutsToCmp []string

	// Include baseFile first if specified and validate it exists.
	if baseFile != "" {
		if err := filesExist([]string{baseFile}); err != nil {
			return nil, err
		}
	}

	// Append all .klf files found in layoutDir.
	entries, err := os.ReadDir(layoutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read layout directory %s: %v", layoutDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".klf") {
			layoutsToCmp = append(layoutsToCmp, entry.Name())
		}
	}

	if baseFile != "" && !slices.Contains(layoutsToCmp, baseFile) {
		layoutsToCmp = append(layoutsToCmp, baseFile)
	}

	return layoutsToCmp, nil
}
