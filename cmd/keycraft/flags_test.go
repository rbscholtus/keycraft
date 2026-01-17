package main

import (
	"context"
	"slices"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestAllSharedFlagsExist verifies that all expected shared flags exist in appFlagsMap.
// Ensures no required flags are missing from the central flag definition map.
func TestAllSharedFlagsExist(t *testing.T) {
	expectedFlags := []string{
		"corpus",
		"load-targets-file",
		"target-hand-load",
		"target-finger-load",
		"target-row-load",
		"pinky-penalties",
		"weights-file",
		"weights",
	}

	for _, flagName := range expectedFlags {
		if _, ok := appFlagsMap[flagName]; !ok {
			t.Errorf("expected flag %q not found in appFlagsMap", flagName)
		}
	}
}

// TestNoExtraSharedFlags verifies that appFlagsMap contains only the 8 expected shared flags
// and no unexpected flags have been added. Prevents flag definition drift.
func TestNoExtraSharedFlags(t *testing.T) {
	expectedFlags := map[string]bool{
		"corpus":             true,
		"load-targets-file":  true,
		"target-hand-load":   true,
		"target-finger-load": true,
		"target-row-load":    true,
		"pinky-penalties":    true,
		"weights-file":       true,
		"weights":            true,
	}

	for flagName := range appFlagsMap {
		if !expectedFlags[flagName] {
			t.Errorf("unexpected flag %q found in appFlagsMap", flagName)
		}
	}

	// Verify count matches
	if len(appFlagsMap) != len(expectedFlags) {
		t.Errorf("appFlagsMap has %d flags, expected %d", len(appFlagsMap), len(expectedFlags))
	}
}

// TestSharedFlagDefaults verifies that each shared flag has the correct default value.
// Tests corpus="default.txt", load-targets-file="load_targets.txt", weights-file="weights.txt",
// and ensures override flags (target-*, weights) default to empty strings.
func TestSharedFlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		expectedType string
		expectedVal  any
	}{
		{"corpus", "string", "default.txt"},
		{"load-targets-file", "string", "load_targets.txt"},
		{"target-hand-load", "string", ""},
		{"target-finger-load", "string", ""},
		{"target-row-load", "string", ""},
		{"pinky-penalties", "string", ""},
		{"weights-file", "string", "weights.txt"},
		{"weights", "string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, ok := appFlagsMap[tt.name]
			if !ok {
				t.Fatalf("flag %q not found", tt.name)
			}

			switch f := flag.(type) {
			case *cli.StringFlag:
				if tt.expectedType != "string" {
					t.Errorf("flag %q: expected type %s, got string", tt.name, tt.expectedType)
				}
				if f.Value != tt.expectedVal {
					t.Errorf("flag %q: expected default %v, got %v", tt.name, tt.expectedVal, f.Value)
				}
			default:
				t.Errorf("flag %q: unexpected flag type %T", tt.name, flag)
			}
		})
	}
}

// TestSharedFlagAliases verifies that each shared flag has the correct short alias.
// Tests aliases like -c for --corpus, -ltf for --load-targets-file, -w for --weights, etc.
func TestSharedFlagAliases(t *testing.T) {
	tests := []struct {
		name            string
		expectedAliases []string
	}{
		{"corpus", []string{"c"}},
		{"load-targets-file", []string{"ltf"}},
		{"target-hand-load", []string{"thl"}},
		{"target-finger-load", []string{"tfl"}},
		{"target-row-load", []string{"trl"}},
		{"pinky-penalties", []string{"pp"}},
		{"weights-file", []string{"wf"}},
		{"weights", []string{"w"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, ok := appFlagsMap[tt.name]
			if !ok {
				t.Fatalf("flag %q not found", tt.name)
			}

			// In cli/v3, check the flag struct directly
			switch f := flag.(type) {
			case *cli.StringFlag:
				if len(f.Aliases) != len(tt.expectedAliases) {
					t.Errorf("flag %q: expected %d aliases, got %d", tt.name, len(tt.expectedAliases), len(f.Aliases))
					return
				}
				for i, expected := range tt.expectedAliases {
					if i >= len(f.Aliases) || f.Aliases[i] != expected {
						t.Errorf("flag %q: expected alias %q, got %q", tt.name, expected, f.Aliases[i])
					}
				}
			default:
				t.Errorf("flag %q: unexpected flag type %T", tt.name, flag)
			}
		})
	}
}

// TestSharedFlagCategories verifies that flags are assigned to the correct help categories.
// Tests that target and weight flags are in "Targets and Weights" category, corpus is uncategorized.
func TestSharedFlagCategories(t *testing.T) {
	tests := []struct {
		name             string
		expectedCategory string
	}{
		{"corpus", ""},
		{"load-targets-file", "Targets and Weights"},
		{"target-hand-load", "Targets and Weights"},
		{"target-finger-load", "Targets and Weights"},
		{"target-row-load", "Targets and Weights"},
		{"pinky-penalties", "Targets and Weights"},
		{"weights-file", "Targets and Weights"},
		{"weights", "Targets and Weights"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, ok := appFlagsMap[tt.name]
			if !ok {
				t.Fatalf("flag %q not found", tt.name)
			}

			var category string
			if f, ok := flag.(interface{ GetCategory() string }); ok {
				category = f.GetCategory()
			}

			if category != tt.expectedCategory {
				t.Errorf("flag %q: expected category %q, got %q", tt.name, tt.expectedCategory, category)
			}
		})
	}
}

// TestCommandSpecificFlagsComplete verifies that each command-specific flag collection
// contains exactly the expected flags - no more, no less. This bidirectional test ensures
// no flags are missing and no unexpected flags exist, using a single source of truth.
// Covers corpusFlags, analyseFlags, rankFlags, optimiseFlags, and generateFlags.
func TestCommandSpecificFlagsComplete(t *testing.T) {
	// Note: view command has no command-specific flags, only uses shared flags
	tests := []struct {
		name          string
		flags         *[]cli.Flag
		expectedFlags []string // Using slice to maintain order and make test output clearer
	}{
		{
			name:          "corpusFlags",
			flags:         &corpusFlags,
			expectedFlags: []string{"corpus-rows", "coverage"},
		},
		{
			name:          "analyseFlags",
			flags:         &analyseFlags,
			expectedFlags: []string{"rows", "compact-trigrams", "trigram-rows"},
		},
		{
			name:          "rankFlags",
			flags:         &rankFlags,
			expectedFlags: []string{"metrics", "deltas", "output"},
		},
		{
			name:          "optimiseFlags",
			flags:         &optimiseFlags,
			expectedFlags: []string{"pins-file", "pins", "free", "generations", "maxtime", "seed", "log-file"},
		},
		{
			name:          "generateFlags",
			flags:         &generateFlags,
			expectedFlags: []string{"layout-type", "vowels-right", "alpha-thumb", "seed", "optimize", "generations"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a map of expected flags for easy lookup
			expectedMap := make(map[string]bool)
			for _, flagName := range tt.expectedFlags {
				expectedMap[flagName] = true
			}

			// Collect all actual flag names (only the primary/longest name, ignoring aliases)
			foundFlags := make(map[string]bool)
			for _, flag := range *tt.flags {
				names := flag.Names()
				if len(names) == 0 {
					continue
				}

				// Find the longest name (primary name, not alias)
				primaryName := names[0]
				for _, name := range names {
					if len(name) > len(primaryName) {
						primaryName = name
					}
				}
				foundFlags[primaryName] = true
			}

			// Check for unexpected flags (in actual but not in expected)
			for flagName := range foundFlags {
				if !expectedMap[flagName] {
					t.Errorf("%s: unexpected flag %q found", tt.name, flagName)
				}
			}

			// Check for missing flags (in expected but not in actual)
			for _, expectedFlag := range tt.expectedFlags {
				if !foundFlags[expectedFlag] {
					t.Errorf("%s: expected flag %q not found", tt.name, expectedFlag)
				}
			}

			// Verify counts match
			if len(foundFlags) != len(tt.expectedFlags) {
				t.Errorf("%s: found %d flags, expected %d", tt.name, len(foundFlags), len(tt.expectedFlags))
			}
		})
	}
}

// TestFlagDefaults_CommandSpecific verifies that each command-specific flag has the correct
// default value. Tests corpus-rows=100, coverage=98.0, rows=10, generations=1000, etc.
// Ensures commands have sensible defaults when flags are not explicitly set.
func TestFlagDefaults_CommandSpecific(t *testing.T) {
	tests := []struct {
		name        string
		flags       *[]cli.Flag
		flagName    string
		expectedVal any
	}{
		{"corpus-rows", &corpusFlags, "corpus-rows", int64(100)},
		{"coverage", &corpusFlags, "coverage", 98.0},
		{"rows", &analyseFlags, "rows", int64(10)},
		{"compact-trigrams", &analyseFlags, "compact-trigrams", false},
		{"trigram-rows", &analyseFlags, "trigram-rows", int64(50)},
		{"metrics", &rankFlags, "metrics", "weighted"},
		{"deltas", &rankFlags, "deltas", "none"},
		{"output", &rankFlags, "output", "table"},
		{"generations_optimise", &optimiseFlags, "generations", uint64(1000)},
		{"maxtime", &optimiseFlags, "maxtime", uint64(5)},
		{"seed_optimise", &optimiseFlags, "seed", int64(0)},
		{"layout-type", &generateFlags, "layout-type", "colstag"},
		{"vowels-right", &generateFlags, "vowels-right", false},
		{"alpha-thumb", &generateFlags, "alpha-thumb", false},
		{"optimize", &generateFlags, "optimize", false},
		{"seed_generate", &generateFlags, "seed", uint64(0)},
		{"generations_generate", &generateFlags, "generations", uint64(1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flag cli.Flag
			for _, f := range *tt.flags {
				names := f.Names()
				if slices.Contains(names, tt.flagName) {
					flag = f
				}
				if flag != nil {
					break
				}
			}

			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			// Compare values with proper type handling
			switch f := flag.(type) {
			case *cli.IntFlag:
				expected, ok := tt.expectedVal.(int64)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want int64", tt.expectedVal)
					return
				}
				if int64(f.Value) != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.Float64Flag:
				expected, ok := tt.expectedVal.(float64)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want float64", tt.expectedVal)
					return
				}
				if f.Value != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.StringFlag:
				expected, ok := tt.expectedVal.(string)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want string", tt.expectedVal)
					return
				}
				if f.Value != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.BoolFlag:
				expected, ok := tt.expectedVal.(bool)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want bool", tt.expectedVal)
					return
				}
				if f.Value != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.UintFlag:
				expected, ok := tt.expectedVal.(uint64)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want uint64", tt.expectedVal)
					return
				}
				if uint64(f.Value) != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.Uint64Flag:
				expected, ok := tt.expectedVal.(uint64)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want uint64", tt.expectedVal)
					return
				}
				if f.Value != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			case *cli.Int64Flag:
				expected, ok := tt.expectedVal.(int64)
				if !ok {
					t.Errorf("expected value type mismatch: got %T, want int64", tt.expectedVal)
					return
				}
				if f.Value != expected {
					t.Errorf("expected %v, got %v", expected, f.Value)
				}
			default:
				t.Errorf("unexpected flag type %T", flag)
			}
		})
	}
}

// ============================================================================
// FLAG VALIDATION TESTS
// ============================================================================
// These tests verify that validation functions correctly accept/reject inputs

// TestCorpusFlagValidation verifies that the corpus command accepts no positional arguments
// and correctly rejects arguments when provided. Tests validateCorpusFlags() function.
func TestCorpusFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args accepted", []string{"corpus"}, false},
		{"rejects positional args", []string{"corpus", "arg1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "corpus",
				Before: validateCorpusFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCorpusFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestViewFlagValidation verifies that the view command requires at least one layout argument
// and accepts multiple layouts. Tests validateViewFlags() function.
func TestViewFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args rejected", []string{"view"}, true},
		{"single layout accepted", []string{"view", "layout1"}, false},
		{"multiple layouts accepted", []string{"view", "layout1", "layout2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "view",
				Before: validateViewFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateViewFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAnalyseFlagValidation verifies that the analyse command requires at least one layout argument
// and accepts multiple layouts. Tests validateAnalyseFlags() function.
func TestAnalyseFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args rejected", []string{"analyse"}, true},
		{"single layout accepted", []string{"analyse", "layout1"}, false},
		{"multiple layouts accepted", []string{"analyse", "layout1", "layout2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "analyse",
				Before: validateAnalyseFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAnalyseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFlipFlagValidation verifies that the flip command requires exactly one layout argument.
// Tests that it rejects no arguments and rejects multiple arguments. Tests validateFlipFlags().
func TestFlipFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args rejected", []string{"flip"}, true},
		{"single layout accepted", []string{"flip", "layout1"}, false},
		{"multiple layouts rejected", []string{"flip", "layout1", "layout2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "flip",
				Before: validateFlipFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFlipFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOptimiseFlagValidation verifies that the optimise command requires exactly one layout argument.
// Tests that it rejects no arguments and rejects multiple arguments. Tests validateOptFlags().
func TestOptimiseFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args rejected", []string{"optimise"}, true},
		{"single layout accepted", []string{"optimise", "layout1"}, false},
		{"multiple layouts rejected", []string{"optimise", "layout1", "layout2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "optimise",
				Before: validateOptFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateFlagValidation verifies that the generate command accepts no positional arguments
// and correctly rejects arguments when provided. Tests validateGenerateFlags() function.
func TestGenerateFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args accepted", []string{"generate"}, false},
		{"rejects positional args", []string{"generate", "arg1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:   "generate",
				Before: validateGenerateFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			}

			err := app.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGenerateFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
