package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// generateFlagsMap contains flags specific to the generate command,
// keyed by their primary name.
var generateFlagsMap = map[string]cli.Flag{
	"max-layouts": &cli.IntFlag{
		Name:     "max-layouts",
		Aliases:  []string{"m"},
		Usage:    "Maximum number of permutations to generate (0 = all)",
		Value:    1500,
		Category: "Generation",
	},
	"seed": &cli.Uint64Flag{
		Name:     "seed",
		Aliases:  []string{"s"},
		Usage:    "Random seed for random position allocation (0 = timestamp)",
		Value:    0,
		Category: "Generation",
	},
	"optimize": &cli.BoolFlag{
		Name:     "optimize",
		Aliases:  []string{"o"},
		Usage:    "Run optimization after generation",
		Value:    false,
		Category: "Optimization",
	},
	// "pins": &cli.StringFlag{
	// 	Name:     "pins",
	// 	Aliases:  []string{"p"},
	// 	Usage:    "Characters to pin during optimization (e.g., 'aeiouy'). Overrides default pins.",
	// 	Category: "Optimization",
	// },
	// "generations": &cli.UintFlag{
	// 	Name:     "generations",
	// 	Aliases:  []string{"g"},
	// 	Usage:    "Number of optimization iterations to run",
	// 	Value:    1000,
	// 	Category: "Optimization",
	// },
	// "maxtime": &cli.UintFlag{
	// 	Name:     "maxtime",
	// 	Aliases:  []string{"mt"},
	// 	Usage:    "Maximum optimization time in minutes.",
	// 	Value:    5,
	// 	Category: "Optimization",
	// },
	"keep-unoptimized": &cli.BoolFlag{
		Name:     "keep-unoptimized",
		Aliases:  []string{"k"},
		Usage:    "Keep unoptimized layouts when --optimize is used",
		Value:    false,
		Category: "Optimization",
	},
}

// generationFlags returns a slice of cli.Flag pointers for the specified keys from generateFlagsMap,
// or all Flags if no keys are specified
func generationFlags(keys ...string) []cli.Flag {
	return flags(generateFlagsMap, keys...)
}

// generateCmdFlags returns all flags for the generate command
func generateCmdFlags() []cli.Flag {
	optF := optFlags("pins", "generations", "maxtime")
	return append(append(commonFlags(), optF...), generationFlags()...)
}

// generateCommand defines the "generate" CLI command for creating layouts from config files
var generateCommand = &cli.Command{
	Name:      "generate",
	Aliases:   []string{"g"},
	Usage:     "Generate layouts from config file",
	ArgsUsage: "<config.gen>",
	Flags:     generateCmdFlags(),
	Action:    generateAction,
}

// generateAction manages the full generation workflow
func generateAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	// Step 1: Build GenerateInput (validates args, resolves path, captures flags)
	genInput, err := buildGenerateInput(c)
	if err != nil {
		return err
	}

	// Step 2: Parse and validate config file
	config, err := kc.ParseConfigFile(genInput.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not parse config file: %w", err)
	}

	if err := kc.ValidateConfig(config); err != nil {
		return err
	}

	// Step 3: If --optimize, build OptimizeInput (without layout/pins for now)
	// TODO: check if this can go inside if genInput.Optimize. If so, move code and update GENERATION.md
	var optInput kc.OptimizeInput
	if genInput.Optimize {
		// skipLayoutLoad=true: don't load layout from args, it will be set per generated layout
		optInput, err = buildOptimizeInput(c, nil, true)
		if err != nil {
			return fmt.Errorf("could not build optimize input: %w", err)
		}
	}

	// Step 4: Generate layouts
	result, err := kc.GenerateFromConfig(config, genInput, layoutDir)
	if err != nil {
		return fmt.Errorf("could not generate layouts: %w", err)
	}

	// Print warnings
	for _, warning := range result.Warnings {
		fmt.Printf("Warning: %s\n", warning)
	}

	fmt.Printf("Generated %d layouts (total permutations: %d)\n", result.Generated, result.TotalPerms)

	// Step 5: Optimize if requested
	if genInput.Optimize {
		err := optimiseLayout(result, config, c, optInput, genInput)
		if err != nil {
			return fmt.Errorf("could not optimise generated layout: %w", err)
		}
	}

	// Step 6: Build RankingInput and display rankings
	rankingInput, err := buildRankingInput(c, optInput.Weights, true)
	if err != nil {
		return fmt.Errorf("could not build ranking input: %w", err)
	}

	// Set layout files from generation result
	rankingInput.LayoutFiles = result.LayoutPaths

	// Compute and display rankings
	rankings, err := kc.ComputeRankings(rankingInput)
	if err != nil {
		return fmt.Errorf("could not compute rankings: %w", err)
	}

	// Build display options for rendering
	displayOpts := tui.RankingDisplayOptions{
		OutputFormat:  tui.OutputTable,
		MetricsOption: tui.MetricsWeighted,
		Weights:       rankingInput.Weights,
		DeltasOption:  tui.DeltasNone,
	}

	if err := tui.RenderRankingTable(rankings, displayOpts); err != nil {
		return fmt.Errorf("could not render rankings: %w", err)
	}

	return nil
}

func optimiseLayout(result *kc.GenerationResult, config *kc.GenerationConfig, c *cli.Command, optInput kc.OptimizeInput, genInput kc.GenerateInput) error {
	fmt.Printf("Optimizing %d layouts...\n", len(result.Layouts))

	for i, layout := range result.Layouts {
		// Compute default pins for this layout
		pinned := kc.ComputeDefaultPins(config, layout)

		// Check if custom pins were specified via --pins flag
		pinsStr := c.String("pins")
		if pinsStr != "" {
			// Override with custom pins
			pinned = computeCustomPins(layout, pinsStr, config)
		}

		// Set up optimization input for this layout
		optInput.Layout = layout
		optInput.Pinned = &pinned
		optInput.LayoutsDir = layoutDir

		// Run optimization (nil writer = no console output during optimization)
		optimizeResult, err := kc.OptimizeLayout(optInput, nil)
		if err != nil {
			return fmt.Errorf("optimization failed for %s: %w", layout.Name, err)
		}

		// Save optimized layout with -opt suffix
		bestLayout := optimizeResult.BestLayout
		optimizedPath := filepath.Join(layoutDir, bestLayout.Name+".klf")
		if err := bestLayout.SaveToFile(optimizedPath); err != nil {
			return fmt.Errorf("failed to save optimized layout %s: %w", bestLayout.Name, err)
		}

		// Update the result with optimized layout path
		result.LayoutPaths[i] = optimizedPath
		result.Layouts[i] = bestLayout

		// Delete original if not keeping unoptimized
		if !genInput.KeepUnoptimized {
			originalPath := filepath.Join(layoutDir, layout.Name+".klf")
			if err := os.Remove(originalPath); err != nil {
				// Just warn, don't fail
				fmt.Printf("Warning: could not delete original layout %s: %v\n", originalPath, err)
			}
		}
	}

	fmt.Printf("Optimization complete\n")
	return nil
}

// computeCustomPins computes pinned keys from a custom pins string.
// Only the specified characters are pinned (plus unused positions and space).
func computeCustomPins(layout *kc.SplitLayout, pinsStr string, config *kc.GenerationConfig) kc.PinnedKeys {
	var pinned kc.PinnedKeys

	// Create set of characters to pin
	pinChars := make(map[rune]bool)
	for _, r := range pinsStr {
		pinChars[r] = true
	}
	pinChars[' '] = true // Always pin space

	// Pin positions based on config template (unused always pinned)
	// and whether the character at that position is in pinChars
	for i, spec := range config.Template {
		if spec.Type == kc.PositionUnused {
			pinned[i] = true // Always pin unused
		} else {
			// Pin if character is in pinChars set
			pinned[i] = pinChars[layout.Runes[i]]
		}
	}

	return pinned
}

// buildGenerateInput validates arguments, resolves config path, and captures generation flags.
func buildGenerateInput(c *cli.Command) (kc.GenerateInput, error) {
	// Validate exactly one argument (config file)
	if c.Args().Len() != 1 {
		return kc.GenerateInput{}, fmt.Errorf("expected exactly 1 config file argument, got %d", c.Args().Len())
	}

	// Check .gen extension (case-insensitive)
	configPath := c.Args().Get(0)
	if !strings.HasSuffix(strings.ToLower(configPath), ".gen") {
		return kc.GenerateInput{}, fmt.Errorf("config file must have .gen extension, got: %s", configPath)
	}

	// Resolve config file path
	resolvedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return kc.GenerateInput{}, err
	}

	return kc.GenerateInput{
		ConfigPath:      resolvedPath,
		MaxLayouts:      c.Int("max-layouts"),
		Seed:            c.Uint64("seed"),
		Optimize:        c.Bool("optimize"),
		KeepUnoptimized: c.Bool("keep-unoptimized"),
	}, nil
}

// resolveConfigPath resolves a config file path, searching in data/config if needed
func resolveConfigPath(configPath string) (string, error) {
	// If the path exists as-is, use it
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Try in data/config directory
	fullPath := filepath.Join(configDir, configPath)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// Not found in either location
	return "", fmt.Errorf("config file not found: %s (tried current directory and %s)", configPath, configDir)
}
