// compare.go provides the implementation for the "compare" command in the kb CLI tool.
// It allows users to compare keyboard layouts based on various metrics and user-defined weights.
// The command can compare specific layouts or all layouts in a directory, and supports ordering
// of results and custom metric weighting.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

// listCommand defines the "list" CLI command.
// This command lists keyboard layouts from the specified directory.
// It supports filtering by layout filenames, ordering the output,
// applying metric weights, and optionally showing delta rows between layouts.
//
// Flags:
//
//	--weights / -w    : Specify metric weights, e.g., sfb=3.0,lsb=2.0
//	--order   / -o    : Specify layout ordering: 'rank' (by penalty score) or 'cli' (as listed)
//	--show-deltas / -d: Display delta rows comparing metrics between layouts (default: true)
var listCommand = &cli.Command{
	Name:   "list",
	Usage:  "List keyboard layouts with optional delta rows",
	Action: listAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "weights",
			Aliases: []string{"w"},
			Usage:   "specify weights for metrics, e.g. sfb=3.0,lsb=2.0",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "order",
			Aliases: []string{"o"},
			Usage:   "Order of layouts: 'rank' or 'cli' (as listed)",
			Value:   "rank",
		},
		&cli.BoolFlag{
			Name:    "show-deltas",
			Aliases: []string{"d"},
			Usage:   "show delta rows in the output",
			Value:   false,
		},
	},
}

func listAction(c *cli.Context) error {
	// Step 1: Load the corpus used for analyzing layouts.
	corpus, err := loadCorpus(c)
	if err != nil {
		return err
	}

	// Step 2: Parse the user-provided weights for layout metrics.
	weights, err := layout.NewWeightsFromString(c.String("weights"))
	if err != nil {
		return err
	}

	// Step 3: Determine which layouts to list.
	// If no layout filenames are provided as CLI arguments,
	// scan the layout directory for all .klf files and use them.
	// Otherwise, use the provided layout filenames and verify their existence.
	var layoutsToCompare []string
	if c.Args().Len() == 0 {
		entries, err := os.ReadDir(layoutDir)
		if err != nil {
			return fmt.Errorf("failed to read layout directory %s: %v", layoutDir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".klf") {
				layoutsToCompare = append(layoutsToCompare, entry.Name())
			}
		}
		sort.Strings(layoutsToCompare)
	} else {
		layoutsToCompare = c.Args().Slice()
		for _, layoutName := range layoutsToCompare {
			path := filepath.Join(layoutDir, layoutName)
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("layout file %s does not exist in %s", layoutName, layoutDir)
				}
				return fmt.Errorf("error checking layout file %s: %v", layoutName, err)
			}
		}
	}

	// Step 4: Validate the --order flag; it must be 'cli' or 'rank'.
	orderOption := c.String("order")
	if orderOption != "cli" && orderOption != "rank" {
		return fmt.Errorf("invalid order option %q; must be 'cli' or 'rank'", orderOption)
	}

	// Step 5: Perform the layout comparison and display results,
	// passing all relevant parameters including the showDeltas flag.
	return layout.DoLayoutList(corpus, layoutDir, weights, c.String("style"), layoutsToCompare, orderOption, c.Bool("show-deltas"))
}
