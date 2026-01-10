package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// rankCommand defines the "rank" CLI command for comparing and ranking layouts.
// It supports filtering layouts, displaying metric deltas, and applying custom weights.
var rankCommand = &cli.Command{
	Name:          "rank",
	Aliases:       []string{"r"},
	Usage:         "Rank keyboard layouts and view detailed metrics and deltas",
	Flags:         flagsSlice("metrics", "deltas", "output", "corpus", "row-load", "finger-load", "pinky-penalties", "weights-file", "weights"),
	ArgsUsage:     "<layout1> <layout2> ...",
	Action:        rankAction,
	ShellComplete: layoutShellComplete,
}

// rankAction handles the rank command, loading data and displaying layout rankings.
func rankAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if slices.Contains(os.Args, "--generate-shell-completion") {
		return nil
	}

	// 1. Build display options (includes loading weights)
	displayOpts, err := buildDisplayOptions(c)
	if err != nil {
		return err
	}

	// 2. Parse all CLI flags and build input (using weights from displayOpts)
	input, err := buildRankingInput(c, displayOpts.Weights)
	if err != nil {
		return err
	}

	// 3. Compute rankings (business logic)
	rankings, err := kc.ComputeRankings(input)
	if err != nil {
		return err
	}

	// 4. Render results (presentation layer)
	return tui.RenderRankingTable(rankings, displayOpts)
}

// buildRankingInput gathers all input parameters.
func buildRankingInput(c *cli.Command, weights *kc.Weights) (kc.RankingInput, error) {
	corpus, err := getCorpusFromFlags(c)
	if err != nil {
		return kc.RankingInput{}, err
	}

	prefs, err := loadPreferredLoadsFromFlags(c)
	if err != nil {
		return kc.RankingInput{}, err
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
		return kc.RankingInput{}, err
	}

	return kc.RankingInput{
		LayoutsDir:  layoutDir,
		LayoutFiles: layouts,
		Corpus:      corpus,
		Prefs:       prefs,
		Weights:     weights,
	}, nil
}

// buildDisplayOptions gathers display configuration.
func buildDisplayOptions(c *cli.Command) (tui.RankingDisplayOptions, error) {
	// Load weights for display and delta coloring
	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return tui.RankingDisplayOptions{}, err
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
			return tui.RankingDisplayOptions{}, err
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
