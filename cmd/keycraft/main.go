// Package main provides the CLI entrypoint and helper functions for the
// keycraft command-line tool.
//
// view.go implements the "view" command for the keycraft CLI; it loads a corpus
// and analyses one or more keyboard layout files for display.
//
// analyse.go contains functions to analyse keyboard layouts and render
// human-friendly tables summarising hand/row usage and other metrics.
//
// rank.go provides the implementation for the "rank" command in the kb CLI tool.
// It allows users to compare keyboard layouts based on various metrics and user-defined weights.
// The command can compare specific layouts or all layouts in a directory, and supports ordering
// of results and custom metric weighting.
//
// optimise.go implements the optimise command which runs simulated
// optimisation on a layout using corpus, pins and weight configuration.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// Data directories used by the CLI (relative to repository root).
const (
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"
)

// Centralized map of CLI flags used across commands.
// Keeps flag definitions in one place so commands can select only the flags they need.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "corpus file to calculate keyboard metrics",
		Value:   "default.txt",
	},
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Value:   "8.0,11.0,16.0,15.0", // default 4-values mirrored
		Usage:   "ideal finger load: 4 or 8 floats for F0..F3[,F6..F9]. 4 values are mirrored to F9..F6. F4/F5 are 0.0",
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
		Usage:   "weights for for scoring layouts, eg: sfb=-3.0,lsb=-2.0",
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
		Usage:   "deltas to show: none, rows, median, or <layout.klf>",
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

// Helper: convert selected flag keys to a slice.
// Used by commands to pick a subset of flags from appFlagsMap.
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

// main sets up the CLI application and registers commands.
// Validation hooks run before command execution (validateFlags).
func main() {
	app := &cli.App{
		Name:  "keycraft",
		Usage: "A CLI tool for crafting better keyboard layouts",
		Commands: []*cli.Command{
			viewCommand,
			analyseCommand,
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
		fmt.Println(err)
	}
}

// validateFlags is called by urfave/cli Before hook.
// Keep it small: perform early validation of flags here (no-op for now).
func validateFlags(_ *cli.Context) error {
	return nil
}

// loadCorpus loads a corpus by filename from the corpusDir and returns a Corpus.
// It trims the extension to produce a corpus name and delegates loading to kc.NewCorpusFromFile.
func loadCorpus(filename string) (*kc.Corpus, error) {
	if filename == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	return kc.NewCorpusFromFile(corpusName, path)
}

// loadLayout loads a layout by filename from layoutDir and returns a SplitLayout.
// It validates presence of the file and uses kc.NewLayoutFromFile to create the layout.
func loadLayout(filename string) (*kc.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout file is required")
	}
	layoutName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(layoutDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("layout file %s does not exist", path)
	}
	return kc.NewLayoutFromFile(layoutName, path)
}

// parseFingerLoad parses a compact CLI representation of "ideal" finger loads.
//
// Accepted forms:
//   - 4 comma-separated floats: interpreted as F0,F1,F2,F3 and mirrored to F9..F6
//   - 8 comma-separated floats: interpreted as F0,F1,F2,F3,F6,F7,F8,F9
//
// The function injects zeros for the thumb indices F4 and F5 and returns a pointer
// to a fixed-size [10]float64 array mapping directly to F0..F9. Returning a pointer
// avoids copying the array value on return (arrays are value types in Go).
func parseFingerLoad(s string) (*[10]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 && len(parts) != 8 {
		return nil, fmt.Errorf("finger-load must have 4 or 8 comma-separated values")
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// If user provided 4 values, mirror them to create the 8-value representation.
	// We append reversed order so after inserting thumb zeros the indices map to F0..F9.
	if len(parts) == 4 {
		for i := len(parts) - 1; i >= 0; i-- {
			parts = append(parts, parts[i])
		}
	}
	parts = slices.Insert(parts, 4, "0.0", "0.0")

	var vals [10]float64
	for i, p := range parts {
		var v float64
		if p == "" {
			return nil, fmt.Errorf("empty value in finger-load")
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in finger-load: %v", err)
		}
		vals[i] = v
	}

	return &vals, nil
}
