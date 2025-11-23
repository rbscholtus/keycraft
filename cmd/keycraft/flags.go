package main

import (
	"fmt"
	"maps"
	"slices"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

var validMetricSets []string = slices.Collect(maps.Keys(kc.MetricsMap))

// appFlagsMap is a centralized map of CLI flags used across various commands.
// It keeps flag definitions in one place, allowing commands to select only the flags they need,
// promoting reusability and consistency.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "The corpus file used for calculating keyboard metrics.",
		Value:   "default.txt",
	},
	"coverage-threshold": &cli.Float64Flag{
		Name:    "coverage-threshold",
		Aliases: []string{"ct"},
		Usage:   "Percentage threshold (0.1-100.0) for corpus word coverage. Used to filter out low frequency words. The words kept in memory cover this %% of the corpus, and low frequency words beyond the threshold are discarded. This applies only to the word list, and forces a cache rebuild if specified.",
		Value:   98.0,
		Action: func(c *cli.Context, value float64) error {
			if value < 0.1 || value > 100.0 {
				return fmt.Errorf("--coverage-threshold must be 0.1-100 (got %f)", value)
			}
			return nil
		},
	},
	"row-load": &cli.StringFlag{
		Name:    "row-load",
		Aliases: []string{"rl"},
		Usage:   "Define ideal row load percentages. Provide 3 comma-separated floats for top row, home row, and bottom row. Values are scaled to sum to 100%.",
		Value:   "18.5,73,8.5", // default: top, home, bottom
	},
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Usage:   "Define ideal finger load percentages. Provide 4 comma-separated floats (F0-F3, mirrored to F9-F6) or 8 floats (F0-F3, F6-F9). Thumbs (F4/F5) are always 0.0. Values are scaled to sum to 100%.",
		Value:   "7.5,11,16,15.5", // default 4-values mirrored
	},
	"pinky-weights": &cli.StringFlag{
		Name:    "pinky-weights",
		Aliases: []string{"pw"},
		Usage:   "Define pinky off-home penalty weights. Provide 6 comma-separated floats (left hand: top-outer, top-inner, home-outer, home-inner, bottom-outer, bottom-inner; mirrored to right hand) or 12 floats (left then right). Higher values penalize more.",
		Value:   "3,2,1,0,2,1", // default 6-values mirrored
	},
	"rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "Specify the maximum number of rows to display in data tables (must be at least 1).",
		Value:   10,
		Action: func(c *cli.Context, value int) error {
			if value < 1 {
				return fmt.Errorf("--rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	"weights-file": &cli.StringFlag{
		Name:    "weights-file",
		Aliases: []string{"wf"},
		Usage:   "The text file containing custom weights for scoring layouts. Values provided via the '--weights' flag will override these file-based weights.",
		Value:   "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage:   "Custom weights for scoring layouts, provided as a comma-separated string (e.g., 'sfb=-3.0,lsb=-2.0'). These values override any weights specified in a weights file.",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage:   fmt.Sprintf("Specify which set of metrics to display. Available options: %v.", validMetricSets),
		Value:   "basic",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage:   "Control how metric deltas are displayed: 'none' (no deltas), 'rows' (row-by-row differences), 'median' (difference relative to the median), or provide a '<layout>' name to compare against a specific base layout.",
		Value:   "none",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage:   "The text file specifying keys to pin (keep in their current position) during optimization. If no file is provided, default keys ('~' and '_') are pinned.",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage:   "Additional characters to pin to their current positions during optimization (e.g., 'aeiouy'). These are added to any keys specified in a pins file.",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage:   "Specify characters that are free to be moved during optimization. All other characters not explicitly listed here will be pinned.",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "The number of generations (iterations) to run the optimization.",
		Value:   1000,
	},
	"maxtime": &cli.UintFlag{
		Name:    "maxtime",
		Aliases: []string{"mt"},
		Usage:   "Maximum time in minutes to spend optimizing the layout.",
		Value:   5,
	},
	"seed": &cli.Int64Flag{
		Name:    "seed",
		Aliases: []string{"s"},
		Usage:   "Random seed for reproducible optimization results. If not specified, uses current Unix timestamp.",
		Value:   0,
	},
	"log-file": &cli.StringFlag{
		Name:    "log-file",
		Aliases: []string{"lf"},
		Usage:   "Path to write JSONL log file with detailed optimization metrics for analysis.",
	},
}

// flagsSlice returns a slice of cli.Flag pointers for the given keys from appFlagsMap.
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

// getLayoutArgs retrieves the list of layout arguments passed to the CLI command.
// Each layout name is normalized by ensuring it has the ".klf" extension.
func getLayoutArgs(c *cli.Context) []string {
	layouts := c.Args().Slice()
	for i := range layouts {
		layouts[i] = ensureKlf(layouts[i])
	}
	return layouts
}
