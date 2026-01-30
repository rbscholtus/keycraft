package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	progress "github.com/jedib0t/go-pretty/v6/progress"
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

// optResult holds the result of a single layout optimisation.
type optResult struct {
	bestLayout    *kc.SplitLayout
	optimizedPath string
	originalPath  string
	err           error
}

func optimiseLayout(result *kc.GenerationResult, config *kc.GenerationConfig, c *cli.Command, optInput kc.OptimizeInput, genInput kc.GenerateInput) error {
	numLayouts := len(result.Layouts)
	fmt.Printf("Optimizing %d layouts...\n", numLayouts)

	// Compute shared reference stats once (avoids loading ~1256 layouts per goroutine)
	medians, iqrs, filteredWeights, err := kc.ComputeReferenceStats(layoutDir, optInput.Corpus, optInput.Targets, optInput.Weights)
	if err != nil {
		return fmt.Errorf("could not compute reference stats: %w", err)
	}

	optInput.LayoutsDir = layoutDir
	optInput.Medians = medians
	optInput.IQRs = iqrs
	optInput.FilteredWeights = filteredWeights
	optInput.UseParallel = false // Disable BLS internal parallelism

	// Set up progress writer
	s := progress.StyleDefault
	s.Colors = progress.StyleColorsExample
	s.Options.TimeInProgressPrecision = time.Second

	pw := progress.NewWriter()
	pw.SetStyle(s)
	pw.SetTrackerLength(40)
	pw.SetNumTrackersExpected(1)
	pw.SetAutoStop(true)
	pw.SetTrackerPosition(progress.PositionRight)
	go pw.Render()

	tracker := &progress.Tracker{
		Message:    "optimising",
		Total:      int64(numLayouts),
		DeferStart: true,
	}
	pw.AppendTracker(tracker)

	// Create all trackers and work items upfront so go-pretty knows the full
	// set of trackers from the start, ensuring stable cursor positioning.
	type workItem struct {
		index  int
		layout *kc.SplitLayout
		pinned kc.PinnedKeys
	}

	work := make(chan workItem, numLayouts)
	results := make([]optResult, numLayouts)

	pinsStr := c.String("pins")
	for i, layout := range result.Layouts {
		var pinned kc.PinnedKeys
		if pinsStr == "" {
			pinned = kc.ComputeDefaultPins(config, layout)
		} else {
			pinned = computeCustomPins(layout, pinsStr, config)
		}

		work <- workItem{index: i, layout: layout, pinned: pinned}
	}
	close(work)

	// Launch fixed worker goroutines that consume work items.
	numWorkers := min(runtime.NumCPU(), numLayouts)
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Go(func() {
			for item := range work {
				// Clone optInput with per-layout fields
				localInput := optInput
				localInput.Layout = item.layout
				localInput.Pinned = &item.pinned

				// Run optimization (nil writer = no console output)
				optimizeResult, err := kc.OptimizeLayout(localInput, nil)
				tracker.Increment(1)
				if err != nil {
					results[item.index] = optResult{err: fmt.Errorf("optimization failed for %s: %w", item.layout.Name, err)}
					continue
				}

				bestLayout := optimizeResult.BestLayout
				optimizedPath := filepath.Join(layoutDir, bestLayout.Name+".klf")
				if err := bestLayout.SaveToFile(optimizedPath); err != nil {
					results[item.index] = optResult{err: fmt.Errorf("failed to save optimized layout %s: %w", bestLayout.Name, err)}
					continue
				}

				results[item.index] = optResult{
					bestLayout:    bestLayout,
					optimizedPath: optimizedPath,
					originalPath:  filepath.Join(layoutDir, item.layout.Name+".klf"),
				}
			}
		})
	}

	wg.Wait()
	tracker.MarkAsDone()

	// Wait for renderer to flush
	for pw.IsRenderInProgress() {
		time.Sleep(10 * time.Millisecond)
	}

	// Apply results and check for errors
	for i, res := range results {
		if res.err != nil {
			return res.err
		}
		result.LayoutPaths[i] = res.optimizedPath
		result.Layouts[i] = res.bestLayout

		// Delete original if not keeping unoptimized
		if !genInput.KeepUnoptimized {
			if err := os.Remove(res.originalPath); err != nil {
				fmt.Printf("Warning: could not delete original layout %s: %v\n", res.originalPath, err)
			}
		}
	}

	// fmt.Printf("Optimization complete\n")
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
