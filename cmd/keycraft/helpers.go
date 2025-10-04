package main

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Usage:   "ideal finger load: 4 or 8 floats for F0..F3[,F6..F9]. 4 values are mirrored to F9..F6. F4/F5 are 0.0. Values are scaled to add up to 100%",
		Value:   "8,11,16,15", // default 4-values mirrored
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

// getCorpusFromFlag loads the corpus specified by --corpus
func getCorpusFromFlag(c *cli.Context) (*kc.Corpus, error) {
	return loadCorpus(c.String("corpus"))
}

// getFingerLoadFromFlag parses the --finger-load string into a [10]float64
func getFingerLoadFromFlag(c *cli.Context) (*[10]float64, error) {
	fbStr := c.String("finger-load")
	vals, err := parseFingerLoad(fbStr)
	if err != nil {
		return nil, err
	}
	scaled, err := scaleFingerLoad(vals)
	if err != nil {
		return nil, err
	}
	return scaled, nil
}

// loadWeightsFromFlags loads weights from --weights-file and --weights
func loadWeightsFromFlags(c *cli.Context) (*kc.Weights, error) {
	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
	}
	return kc.NewWeightsFromParams(weightsPath, c.String("weights"))
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

// loadLayout resolves a layout filename relative to the layoutDir directory,
// appends the ".klf" extension if it is missing (case-insensitive), and
// then delegates parsing to kc.NewLayoutFromFile to construct and return a SplitLayout.
func loadLayout(filename string) (*kc.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout is required")
	}

	var layoutName string
	ext := filepath.Ext(filename)
	if strings.ToLower(ext) != ".klf" {
		layoutName = filename
		filename += ".klf"
	} else {
		layoutName = strings.TrimSuffix(filename, ext)
	}
	path := filepath.Join(layoutDir, filename)

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

	// convert values to float64
	var fingerVals [10]float64
	for i, p := range parts {
		var v float64
		if p == "" {
			return nil, fmt.Errorf("empty value in finger-load")
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in finger-load: %v", err)
		}
		fingerVals[i] = v
	}

	return &fingerVals, nil
}

// scaleFingerLoad scales the finger load values so their sum is 100.0.
// Returns an error if the sum is zero.
func scaleFingerLoad(vals *[10]float64) (*[10]float64, error) {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	if sum == 0.0 {
		return nil, fmt.Errorf("cannot scale finger load: sum is zero")
	}
	scale := 100.0 / sum
	var scaled [10]float64
	for i, v := range vals {
		scaled[i] = v * scale
	}
	return &scaled, nil
}

// ensureKlf appends the ".klf" extension to the given name if it does not already have it.
// The check is case-insensitive to handle filenames with uppercase or mixed-case extensions,
// ensuring consistent file naming regardless of how the extension was originally cased.
func ensureKlf(name string) string {
	if !strings.HasSuffix(strings.ToLower(name), ".klf") {
		return name + ".klf"
	}
	return name
}

// ensureNoKlf removes the ".klf" extension from the given name if it has it.
// The case-insensitive check allows the function to correctly strip the extension
// regardless of the case used in the filename, supporting flexible user input.
func ensureNoKlf(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".klf") {
		return name[:len(name)-4]
	}
	return name
}
