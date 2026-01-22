package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// generateFlags are flags specific to the generate command
var generateFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "layout-type",
		Aliases:  []string{"lt"},
		Usage:    "Layout geometry: rowstag, anglemod, ortho, colstag",
		Value:    "colstag",
		Category: "Generation",
	},
	&cli.BoolFlag{
		Name:     "vowels-right",
		Aliases:  []string{"vr"},
		Usage:    "Ensure vowels are placed on right hand",
		Value:    false,
		Category: "Generation",
	},
	&cli.BoolFlag{
		Name:     "alpha-thumb",
		Aliases:  []string{"at"},
		Usage:    "Place alpha character on thumb keys (9 chars instead of 8)",
		Value:    false,
		Category: "Generation",
	},
	&cli.Uint64Flag{
		Name:     "seed",
		Aliases:  []string{"s"},
		Usage:    "Random seed for reproducible results (0 = timestamp)",
		Value:    0,
		Category: "Generation",
	},
	&cli.BoolFlag{
		Name:     "optimize",
		Aliases:  []string{"opt"},
		Usage:    "Run optimization after generation",
		Value:    false,
		Category: "Optimization",
	},
	&cli.UintFlag{
		Name:     "generations",
		Aliases:  []string{"gens", "g"},
		Usage:    "Number of optimization iterations to run.",
		Value:    1000,
		Category: "Optimization",
	},
}

// generateFlagsSlice returns all flags for the generate command
func generateFlagsSlice() []cli.Flag {
	// When --optimize is used, we need corpus, targets, weights flags
	commonFlags := flagsSlice("corpus", "load-targets-file", "target-hand-load",
		"target-finger-load", "target-row-load", "pinky-penalties",
		"weights-file", "weights")
	return append(generateFlags, commonFlags...)
}

// generateCommand defines the "generate" CLI command for creating random keyboard layouts.
var generateCommand = &cli.Command{
	Name:    "generate",
	Aliases: []string{"g"},
	Usage:   "Generate a new random keyboard layout",
	Flags:   generateFlagsSlice(),
	Before:  validateGenerateFlags,
	Action:  generateAction,
}

// validateGenerateFlags validates CLI flags before running the generate command.
func validateGenerateFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
	// Skip validation during shell completion
	if isShellCompletion() {
		return ctx, nil
	}

	// No positional arguments required (name is auto-generated)
	if c.Args().Len() > 0 {
		return ctx, fmt.Errorf("generate command does not accept arguments (name is auto-generated)")
	}

	return ctx, nil
}

// generateAction manages the full generation workflow.
func generateAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	// Build GeneratorInput from flags
	input, err := buildGeneratorInput(c)
	if err != nil {
		return fmt.Errorf("could not parse user input: %w", err)
	}

	// Generate layout (name is set inside NewRandomLayout)
	layout, err := kc.NewRandomLayout(input)
	if err != nil {
		return fmt.Errorf("could not generate random layout: %w", err)
	}

	// Save generated layout
	layoutPath := filepath.Join(layoutDir, layout.Name+".klf")
	if err := layout.SaveToFile(layoutPath); err != nil {
		return fmt.Errorf("could not save layout: %w", err)
	}
	fmt.Printf("Generated layout saved to: %s\n", layoutPath)

	// If --optimize flag is set, run optimization
	if c.Bool("optimize") {
		return runOptimizationAfterGeneration(c, layout)
	}

	// Otherwise just display the generated layout
	return displayGeneratedLayout(c, layout)
}

// buildGeneratorInput gathers all input parameters for layout generation.
func buildGeneratorInput(c *cli.Command) (kc.GeneratorInput, error) {
	// Parse layout type
	layoutTypeStr := c.String("layout-type")
	var layoutType kc.LayoutType
	switch layoutTypeStr {
	case "rowstag":
		layoutType = kc.ROWSTAG
	case "anglemod":
		layoutType = kc.ANGLEMOD
	case "ortho":
		layoutType = kc.ORTHO
	case "colstag":
		layoutType = kc.COLSTAG
	default:
		return kc.GeneratorInput{}, fmt.Errorf("invalid layout type: %s (must be rowstag, anglemod, ortho, or colstag)", layoutTypeStr)
	}

	return kc.GeneratorInput{
		LayoutType:  layoutType,
		AlphaThumb:  c.Bool("alpha-thumb"),
		VowelsRight: c.Bool("vowels-right"),
		Seed:        c.Uint64("seed"),
	}, nil
}

// displayGeneratedLayout shows the generated layout to the user.
func displayGeneratedLayout(c *cli.Command, layout *kc.SplitLayout) error {
	// Load corpus for analysis (if available)
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return fmt.Errorf("can't load corpus to display generated layout: %w", err)
	}

	// Load targets for analysis
	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return fmt.Errorf("can't load targets to display generated layout: %w", err)
	}

	// View the generated layout
	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: []string{filepath.Join(layoutDir, layout.Name+".klf")},
		Corpus:      corpus,
		Targets:     targets,
	})
	if err != nil {
		return fmt.Errorf("can't analyse generated layout: %w", err)
	}

	if err := tui.RenderView(viewResult); err != nil {
		return fmt.Errorf("can't render generated layout: %w", err)
	}

	return nil
}

// runOptimizationAfterGeneration chains optimization after generation.
func runOptimizationAfterGeneration(c *cli.Command, generatedLayout *kc.SplitLayout) error {
	fmt.Printf("\nRunning optimization on generated layout %s...\n", generatedLayout.Name)

	// Load corpus
	corpus, err := loadCorpusFromFlags(c)
	if err != nil {
		return fmt.Errorf("could not load corpus for optimization: %w", err)
	}

	// Load targets
	targets, err := loadTargetLoadsFromFlags(c)
	if err != nil {
		return fmt.Errorf("could not load target loads for optimization: %w", err)
	}

	// Load weights
	weights, err := loadWeightsFromFlags(c)
	if err != nil {
		return fmt.Errorf("could not load weights for optimization: %w", err)
	}

	// Build optimization input
	homeThumb := generatedLayout.HomeThumbChars()
	pinned, _ := kc.LoadPinsFromParams("", homeThumb, "", generatedLayout)

	optInput := kc.OptimiseInput{
		Layout:         generatedLayout,
		LayoutsDir:     layoutDir,
		Corpus:         corpus,
		Targets:        targets,
		Weights:        weights,
		Pinned:         pinned,
		NumGenerations: int(c.Uint("generations")),
		MaxTime:        int(c.Uint("maxtime")),
		Seed:           c.Int64("seed"),
	}

	// Run optimization
	optResult, err := kc.OptimizeLayout(optInput, os.Stdout)
	if err != nil {
		return fmt.Errorf("optimization failed: %w", err)
	}

	// Save optimized layout
	bestPath := filepath.Join(layoutDir, optResult.BestLayout.Name+".klf")
	if err := optResult.BestLayout.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("could not save optimized layout: %w", err)
	}

	// Show comparison
	origPath := filepath.Join(layoutDir, generatedLayout.Name+".klf")
	layoutsToCompare := []string{origPath, bestPath}

	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: layoutsToCompare,
		Corpus:      corpus,
		Targets:     targets,
	})
	if err != nil {
		return fmt.Errorf("could not analyse optimised layout: %w", err)
	}

	if err := tui.RenderView(viewResult); err != nil {
		return fmt.Errorf("could not render optimised layout: %w", err)
	}

	// Show ranking comparison
	rankingInput := kc.RankingInput{
		LayoutsDir:  layoutDir,
		LayoutFiles: layoutsToCompare,
		Corpus:      corpus,
		Targets:     targets,
		Weights:     weights,
	}

	rankingResult, err := kc.ComputeRankings(rankingInput)
	if err != nil {
		return fmt.Errorf("could not compute layout rankings: %w", err)
	}

	displayOpts := tui.RankingDisplayOptions{
		OutputFormat:   tui.OutputTable,
		MetricsOption:  tui.MetricsWeighted,
		ShowWeights:    true,
		Weights:        weights,
		DeltasOption:   tui.DeltasCustom,
		BaseLayoutName: generatedLayout.Name,
	}

	if err := tui.RenderRankingTable(rankingResult, displayOpts); err != nil {
		return fmt.Errorf("could not render layout rankings: %w", err)
	}

	return nil
}
