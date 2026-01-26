package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
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
	ArgsUsage: "<layout>",
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

	// Print GenerateInput for verification
	fmt.Printf("GenerateInput{\n")
	fmt.Printf("  ConfigPath: %q\n", genInput.ConfigPath)
	fmt.Printf("  MaxLayouts: %d\n", genInput.MaxLayouts)
	fmt.Printf("  Seed: %d\n", genInput.Seed)
	fmt.Printf("  Optimize: %v\n", genInput.Optimize)
	fmt.Printf("  KeepUnoptimized: %v\n", genInput.KeepUnoptimized)
	fmt.Printf("}\n")

	// Step 2: If --optimize, build OptimizeInput (without layout/pins for now)
	var optInput kc.OptimizeInput
	if genInput.Optimize {
		// skipLayoutLoad=true: don't load layout from args, it will be set per generated layout
		optInput, err = buildOptimizeInput(c, nil, true)
		if err != nil {
			return fmt.Errorf("could not build optimize input: %w", err)
		}

		fmt.Printf("OptimizeInput{\n")
		fmt.Printf("  Layout: (set per generated layout): %q\n", optInput.Layout)
		fmt.Printf("  LayoutsDir: (set per generated layout): %q\n", optInput.LayoutsDir)
		fmt.Printf("  Corpus: %q\n", optInput.Corpus.Name)
		fmt.Printf("  Targets: (loaded): %v\n", optInput.Targets)
		fmt.Printf("  Weights: (loaded): %v\n", optInput.Weights)
		fmt.Printf("  Pinned: (computed per generated layout): %v\n", optInput.Pinned)
		fmt.Printf("  NumGenerations: %d\n", optInput.NumGenerations)
		fmt.Printf("  MaxTime: %d\n", optInput.MaxTime)
		fmt.Printf("  Seed: %d\n", optInput.Seed)
		fmt.Printf("  Logfile: %v\n", optInput.LogFile)
		fmt.Printf("}\n")
	}

	// Step 3: Build RankingInput (for displaying results)
	// skipLayoutsFromArgs=true: layouts come from generation, not CLI args
	rankingInput, err := buildRankingInput(c, optInput.Weights, true)
	if err != nil {
		return fmt.Errorf("could not build ranking input: %w", err)
	}

	fmt.Printf("RankingInput{\n")
	fmt.Printf("  LayoutsDir: %q\n", rankingInput.LayoutsDir)
	fmt.Printf("  LayoutFiles: (set after generation): %v\n", rankingInput.LayoutFiles)
	fmt.Printf("  Corpus: %q\n", rankingInput.Corpus.Name)
	fmt.Printf("  Targets: (loaded): %v\n", rankingInput.Targets)
	fmt.Printf("  Weights: (loaded): %v\n", rankingInput.Weights)
	fmt.Printf("}\n")

	// Step 4: Stub - no actual generation yet
	fmt.Printf("\n[STUB] Generation not implemented yet\n")

	return nil
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
