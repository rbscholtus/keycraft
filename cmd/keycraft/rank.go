package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// rankFlags defines flags specific to the rank command.
var rankFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage: fmt.Sprintf("Metrics to display. Options: %v, or \"weighted\" "+
			"(metrics with |weight|>=0.01), or comma-separated list.",
			slices.Sorted(maps.Keys(kc.MetricsMap))),
		Value:    "weighted",
		Category: "Display",
	},
	&cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage: "Delta display mode: \"none\", \"rows\" (row-by-row), " +
			"\"median\" (vs median), or \"<layout>\" name to compare against.",
		Value:    "none",
		Category: "Display",
	},
	&cli.StringFlag{
		Name:     "output",
		Aliases:  []string{"o"},
		Usage:    "Output format: \"table\", \"html\", or \"csv\".",
		Value:    "table",
		Category: "Display",
	},
}

// rankFlagsSlice returns all flags for the rank command.
func rankFlagsSlice() []cli.Flag {
	commonFlags := flagsSlice("corpus", "load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties", "weights-file", "weights")
	return append(commonFlags, rankFlags...)
}

// rankCommand defines the "rank" CLI command for comparing and ranking layouts.
// It supports filtering layouts, displaying metric deltas, and applying custom weights.
var rankCommand = &cli.Command{
	Name:          "rank",
	Aliases:       []string{"r"},
	Usage:         "Rank keyboard layouts and view detailed metrics and deltas",
	Flags:         rankFlagsSlice(),
	ArgsUsage:     "<layout1> <layout2> ...",
	Action:        rankAction,
	ShellComplete: layoutShellComplete,
}

// rankAction handles the rank command, loading data and displaying layout rankings.
func rankAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if isShellCompletion() {
		return nil
	}

	// 1. Build display options (includes loading weights)
	displayOpts, err := buildDisplayOptions(c)
	if err != nil {
		return fmt.Errorf("could not parse display options: %w", err)
	}

	// 2. Parse all CLI flags and build input (using weights from displayOpts)
	input, err := buildRankingInput(c, displayOpts.Weights)
	if err != nil {
		return fmt.Errorf("could not parse user input for rankings: %w", err)
	}

	// Set corpus name for display (used in table title when deltas are not shown)
	displayOpts.CorpusName = input.Corpus.Name

	// 3. Compute rankings (business logic)
	rankings, err := kc.ComputeRankings(input)
	if err != nil {
		return fmt.Errorf("could not compute rankings: %w", err)
	}

	// 4. Render results (presentation layer)
	return tui.RenderRankingTable(rankings, displayOpts)
}

// buildRankingInput gathers all input parameters.
func buildRankingInput(c *cli.Command, weights *kc.Weights) (kc.RankingInput, error) {
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return kc.RankingInput{}, fmt.Errorf("could not load corpus: %w", err)
	}

	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return kc.RankingInput{}, fmt.Errorf("could not load target loads: %w", err)
	}

	// Check if deltas references a specific layout (not "none", "rows", or "median")
	deltasValue := c.String("deltas")
	deltasValueLower := strings.ToLower(deltasValue)
	var baseLayout string
	if deltasValueLower != "none" && deltasValueLower != "rows" && deltasValueLower != "median" {
		baseLayout = deltasValue
	}

	layouts, err := getLayoutsFromArgs(c, baseLayout)
	if err != nil {
		return kc.RankingInput{}, fmt.Errorf("could not get layouts from args: %w", err)
	}

	return kc.RankingInput{
		LayoutsDir:  layoutDir,
		LayoutFiles: layouts,
		Corpus:      corpus,
		Targets:     targets,
		Weights:     weights,
	}, nil
}

// buildDisplayOptions gathers display configuration.
func buildDisplayOptions(c *cli.Command) (tui.RankingDisplayOptions, error) {
	// Load weights for display and delta coloring
	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return tui.RankingDisplayOptions{}, fmt.Errorf("could not load weights: %w", err)
	}

	// Parse output format
	outputFmt := tui.OutputTable
	if c.IsSet("output") {
		switch strings.ToLower(c.String("output")) {
		case "table":
			outputFmt = tui.OutputTable
		case "html":
			outputFmt = tui.OutputHTML
		case "csv":
			outputFmt = tui.OutputCSV
		default:
			return tui.RankingDisplayOptions{}, fmt.Errorf("invalid output format; must be one of: table, html, csv")
		}
	}

	metricsValue := strings.ToLower(c.String("metrics"))

	var metricsOpt tui.MetricsOption
	var customMetrics []string

	// Check if it's "weighted" (special case - computed dynamically)
	if metricsValue == "weighted" {
		metricsOpt = tui.MetricsWeighted
	} else if _, ok := kc.MetricsMap[metricsValue]; ok {
		// Check if it's a predefined metrics set
		metricsOpt = tui.MetricsOption(metricsValue)
	} else {
		// Treat as custom comma-separated list
		metricsOpt = tui.MetricsCustom
		customMetrics = strings.Split(metricsValue, ",")
		for i := range customMetrics {
			customMetrics[i] = strings.TrimSpace(customMetrics[i])
			customMetrics[i] = strings.ToUpper(customMetrics[i])
		}

		// Validate that all custom metrics exist
		if err := validateMetrics(customMetrics); err != nil {
			return tui.RankingDisplayOptions{}, fmt.Errorf("could not validate metrics: %w", err)
		}
	}

	deltasValue := c.String("deltas")
	deltasValueLower := strings.ToLower(deltasValue)
	var deltasOpt tui.DeltasOption
	var baseLayoutName string

	switch deltasValueLower {
	case "none":
		deltasOpt = tui.DeltasNone
	case "rows":
		deltasOpt = tui.DeltasRows
	case "median":
		deltasOpt = tui.DeltasMedian
	default:
		deltasOpt = tui.DeltasCustom
		baseLayoutName = ensureNoKlf(deltasValue)
	}

	return tui.RankingDisplayOptions{
		OutputFormat:   outputFmt,
		MetricsOption:  metricsOpt,
		CustomMetrics:  customMetrics,
		ShowWeights:    true,
		Weights:        weights,
		DeltasOption:   deltasOpt,
		BaseLayoutName: baseLayoutName,
	}, nil
}

// validateMetrics checks that all provided metrics exist in the "all" metrics set.
func validateMetrics(metrics []string) error {
	allMetrics := kc.MetricsMap["all"]
	validMetrics := make(map[string]bool, len(allMetrics))
	for _, m := range allMetrics {
		validMetrics[m] = true
	}

	var invalid []string
	for _, metric := range metrics {
		if !validMetrics[metric] {
			invalid = append(invalid, metric)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid metric(s): %v; run with --metrics=all to see all available metrics", invalid)
	}

	return nil
}

// getLayoutsFromArgs returns layouts from CLI args, or all .klf files if no args provided.
func getLayoutsFromArgs(c *cli.Command, baseLayout string) ([]string, error) {
	var layouts []string
	if c.Args().Len() == 0 {
		var err error
		layouts, err = allLayoutFiles()
		if err != nil {
			return nil, fmt.Errorf("could not get all layout files: %w", err)
		}
	} else {
		layouts = c.Args().Slice()
		for i := range layouts {
			arg := layouts[i]
			// // Check if it's an existing file
			// if _, err := os.Stat(arg); err == nil {
			// 	absPath, err := filepath.Abs(arg)
			// 	if err == nil {
			// 		layouts[i] = absPath
			// 		continue
			// 	}
			// }

			// Otherwise assume it's a name in layoutDir
			layouts[i] = filepath.Join(layoutDir, ensureKlf(arg))
		}
	}

	if baseLayout != "" {
		// // Handle baseLayout similarly
		// var baseLayoutPath string
		// if _, err := os.Stat(baseLayout); err == nil {
		// 	baseLayoutPath, _ = filepath.Abs(baseLayout)
		// } else {
		baseLayoutPath := filepath.Join(layoutDir, ensureKlf(baseLayout))
		// }

		if !slices.Contains(layouts, baseLayoutPath) {
			layouts = append(layouts, baseLayoutPath)
		}
	}

	return layouts, nil
}

// allLayoutFiles returns all .klf files in layoutDir.
func allLayoutFiles() ([]string, error) {
	var layoutsToCmp []string

	entries, err := os.ReadDir(layoutDir)
	if err != nil {
		return nil, fmt.Errorf("could not read layout directory %s: %w", layoutDir, err)
	}
	for _, entry := range entries {
		entryName := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entryName), ".klf") {
			layoutsToCmp = append(layoutsToCmp, filepath.Join(layoutDir, entryName))
		}
	}

	return layoutsToCmp, nil
}
