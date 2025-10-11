package main

import (
	"fmt"
	"maps"
	"slices"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

var validMetricSets []string = slices.Collect(maps.Keys(kc.MetricsMap))

// Centralized map of CLI flags used across commands.
// Keeps flag definitions in one place so commands can select only the flags they need.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "corpus file to calculate keyboard metrics",
		Value:   "default.txt",
	},
	"coverage-threshold": &cli.Float64Flag{
		Name:    "coverage-threshold",
		Aliases: []string{"ct"},
		Usage:   "word coverage threshold percentage for corpus analysis. Low frequency words beyond the threshold are discarded. Applies to the word list only. Specifying this forces rebuilding the cache",
		Value:   98.0,
		Action: func(c *cli.Context, value float64) error {
			if value < 0.1 || value > 100.0 {
				return fmt.Errorf("--coverage-threshold must be 0.1-100 (got %f)", value)
			}
			return nil
		},
	},
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Usage:   "ideal finger load: 4 or 8 floats for F0..F3[,F6..F9]. 4 values are mirrored to F9..F6. F4/F5 are 0.0. Values are scaled to add up to 100%",
		Value:   "7.5,11,16,15.5", // default 4-values mirrored
	},
	"rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "number of rows to show in data tables",
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
		Usage:   "text file containing weights for scoring layouts; weights flag overrides these values",
		Value:   "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage:   "weights for scoring layouts, eg: sfb=-3.0,lsb=-2.0",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage:   fmt.Sprintf("metrics to show: %v", validMetricSets),
		Value:   "basic",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage:   "deltas to show: none, rows, median, or <layout>",
		Value:   "none",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage:   "text file containing keys to pin to their current position; if no file is specified, ~ keys and _ are pinned",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage:   "additional characters to pin, eg: aeiouy",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage:   "characters free to be moved (all others pinned), eg: zqjx",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "number of generations",
		Value:   250,
	},
	"accept-worse": &cli.StringFlag{
		Name:    "accept-worse",
		Aliases: []string{"aw"},
		Usage:   fmt.Sprintf("accept-worse function (how likely is it a worse layout is accepted): %v", validAcceptFuncs),
		Value:   "drop-slow",
	},
}

// flagsSlice returns a slice of cli.Flag corresponding to the given keys from appFlagsMap.
// Useful for commands that only need a subset of globally defined CLI flags.
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

// getLayoutArgs returns the list of layout arguments passed to the command,
// normalizing each layout name with ensureKlf.
func getLayoutArgs(c *cli.Context) []string {
	layouts := c.Args().Slice()
	for i := range layouts {
		layouts[i] = ensureKlf(layouts[i])
	}
	return layouts
}
