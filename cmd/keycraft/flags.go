package main

import (
	"fmt"
	"maps"
	"slices"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// appFlagsMap is a centralized map of CLI flags used across various commands.
// It keeps flag definitions in one place, allowing commands to select only the
// flags they need, promoting reusability and consistency.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "Corpus file for calculating metrics (from data/corpus directory).",
		Value:   "default.txt",
	},
	"corpus-rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "Maximum number of rows to display in corpus data tables.",
		Value:   100,
		Action: func(c *cli.Context, value int) error {
			if value < 1 {
				return fmt.Errorf("--rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	"coverage": &cli.Float64Flag{
		Name: "coverage",
		Usage: "Corpus word coverage percentage (0.1-100.0). Filters " +
			"low-frequency words. Forces cache rebuild.",
		Value: 98.0,
		Action: func(c *cli.Context, value float64) error {
			if value < 0.1 || value > 100.0 {
				return fmt.Errorf("--coverage must be 0.1-100 (got %f)", value)
			}
			return nil
		},
	},
	"row-load": &cli.StringFlag{
		Name:    "row-load",
		Aliases: []string{"rl"},
		Usage: "Ideal row load percentages: 3 comma-separated values for top, " +
			"home, bottom rows (auto-scaled to 100%).",
		Value: "18.5,73,8.5", // default: top, home, bottom
	},
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Usage: "Ideal finger load percentages: 4 values (left 4 fingers, " +
			"mirrored to right) or 8 values. Thumbs always 0. Auto-scaled to 100%.",
		Value: "7.5,11,16,15.5", // default 4-values mirrored
	},
	"pinky-penalties": &cli.StringFlag{
		Name:    "pinky-penalties",
		Aliases: []string{"pp"},
		Usage: "Pinky off-home penalties: 6 values (left top-outer, top-inner, " +
			"home-outer, home-inner, bottom-outer, bottom-inner; mirrored) or " +
			"12 values (left, then right). Higher = more penalty.",
		Value: "1,1,1,0,1,1", // default 6-values mirrored
	},
	"rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "Maximum number of rows to display in data tables.",
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
		Usage: "Weights file for scoring layouts (from data/weights directory). " +
			"Overridden by --weights flag.",
		Value: "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage: "Custom metric weights as comma-separated pairs " +
			"(e.g., \"SFB=-10,LSB=-5\"). Overrides weights file.",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage: fmt.Sprintf("Metrics to display. Options: %v, or \"weighted\" "+
			"(metrics with |weight|>=0.01), or comma-separated list.",
			slices.Sorted(maps.Keys(kc.MetricsMap))),
		Value: "weighted",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage: "Delta display mode: \"none\", \"rows\" (row-by-row), " +
			"\"median\" (vs median), or \"<layout>\" name to compare against.",
		Value: "none",
	},
	"output": &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "Output format: \"table\", \"html\", or \"csv\".",
		Value:   "table",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage: "File specifying keys to pin during optimization. " +
			"Defaults to pinning '~' and '_'.",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage: "Additional characters to pin (e.g., 'aeiouy'). " +
			"Combined with pins-file.",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage: "Characters free to move during optimization. " +
			"All others are pinned.",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "Number of optimization iterations to run.",
		Value:   1000,
	},
	"maxtime": &cli.UintFlag{
		Name:    "maxtime",
		Aliases: []string{"mt"},
		Usage:   "Maximum optimization time in minutes.",
		Value:   5,
	},
	"seed": &cli.Int64Flag{
		Name:    "seed",
		Aliases: []string{"s"},
		Usage:   "Random seed for reproducible results. Uses current timestamp if 0.",
		Value:   0,
	},
	"log-file": &cli.StringFlag{
		Name:    "log-file",
		Aliases: []string{"lf"},
		Usage:   "JSONL log file path for detailed optimization metrics.",
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
