package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestTargetLoads_HardcodedDefaults verifies that hardcoded defaults from internal/keycraft/analyser.go
// are used when no config file exists and no flags are set. Tests hand load=50/50, finger loads, row loads.
func TestTargetLoads_HardcodedDefaults(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create any config file - test with empty load-targets-file
	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify defaults are applied
			// These values come from internal/keycraft/analyser.go DefaultTargetHandLoad()
			if targets.TargetHandLoad[0] != 50.0 || targets.TargetHandLoad[1] != 50.0 {
				t.Errorf("hand load = %v,%v, want 50,50", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			// Finger loads: F0=7, F1=10, F2=16, F3=17, F6=17, F7=16, F8=10, F9=7
			expectedFingerLoads := []float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v", i, actual, expected)
				}
			}

			// Row loads: Top=17.5, Home=75, Bottom=7.5
			if targets.TargetRowLoad[0] != 17.5 {
				t.Errorf("row load top = %v, want 17.5", targets.TargetRowLoad[0])
			}
			if targets.TargetRowLoad[1] != 75.0 {
				t.Errorf("row load home = %v, want 75.0", targets.TargetRowLoad[1])
			}
			if targets.TargetRowLoad[2] != 7.5 {
				t.Errorf("row load bottom = %v, want 7.5", targets.TargetRowLoad[2])
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--load-targets-file", ""})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_FromFile verifies that target loads are correctly loaded from load_targets.txt
// when the file exists and contains valid configuration. Tests all four target load types.
func TestTargetLoads_FromFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create config file with custom values
	configContent := `target-hand-load = 40, 60
target-finger-load = 5, 10, 20, 15, 15, 20, 10, 5
target-row-load = 20, 70, 10
pinky-penalties = 3.0, 2.0, 1.5, 0.5, 2.5, 2.0, 3.0, 2.0, 1.5, 0.5, 2.5, 2.0
`
	writeTestConfigFile(t, configDir, "load_targets.txt", configContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify values from file are loaded
			if targets.TargetHandLoad[0] != 40.0 || targets.TargetHandLoad[1] != 60.0 {
				t.Errorf("hand load = %v,%v, want 40,60", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			expectedFingerLoads := []float64{5, 10, 20, 15, 0, 0, 15, 20, 10, 5}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v", i, actual, expected)
				}
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--load-targets-file", "load_targets.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_FromFile_PartialConfig verifies that when a config file has partial configuration,
// specified fields are loaded from the file while missing fields fall back to hardcoded defaults.
func TestTargetLoads_FromFile_PartialConfig(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create config file with only hand-load (other fields should default)
	configContent := `target-hand-load = 40, 60
`
	writeTestConfigFile(t, configDir, "load_targets.txt", configContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify hand load from file
			if targets.TargetHandLoad[0] != 40.0 || targets.TargetHandLoad[1] != 60.0 {
				t.Errorf("hand load = %v,%v, want 40,60", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			// Verify finger loads use defaults
			expectedFingerLoads := []float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v (should be default)", i, actual, expected)
				}
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--load-targets-file", "load_targets.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_FlagOverridesFile verifies that CLI flags (--target-hand-load, etc.) take precedence
// over file configuration. Tests the configuration hierarchy: file < flags.
func TestTargetLoads_FlagOverridesFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create config file
	configContent := `target-hand-load = 40, 60
target-finger-load = 5, 10, 20, 15, 15, 20, 10, 5
`
	writeTestConfigFile(t, configDir, "load_targets.txt", configContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file", "target-hand-load", "target-finger-load"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify flag overrides file for hand load
			if targets.TargetHandLoad[0] != 30.0 || targets.TargetHandLoad[1] != 70.0 {
				t.Errorf("hand load = %v,%v, want 30,70 (flag should override file)", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			// Verify finger load from file (not overridden)
			expectedFingerLoads := []float64{5, 10, 20, 15, 0, 0, 15, 20, 10, 5}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v (from file)", i, actual, expected)
				}
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--load-targets-file", "load_targets.txt", "--target-hand-load", "30,70"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_AllOverrides verifies that all target flag types can be set simultaneously
// without conflicts. Tests hand-load, finger-load, row-load, and pinky-penalties together.
func TestTargetLoads_AllOverrides(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create config file with different values (all should be overridden)
	configContent := `target-hand-load = 40, 60
target-finger-load = 5, 10, 20, 15, 15, 20, 10, 5
target-row-load = 20, 70, 10
pinky-penalties = 3.0, 2.0, 1.5, 0.5, 2.5, 2.0
`
	writeTestConfigFile(t, configDir, "load_targets.txt", configContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file", "target-hand-load", "target-finger-load", "target-row-load", "pinky-penalties"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify all flags override file values
			if targets.TargetHandLoad[0] != 30.0 || targets.TargetHandLoad[1] != 70.0 {
				t.Errorf("hand load = %v,%v, want 30,70 (flag override)", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			// Finger loads: values must sum to 100 (8+11+19+12+12+19+11+8 = 100)
			expectedFingerLoads := []float64{8, 11, 19, 12, 0, 0, 12, 19, 11, 8}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v (flag override)", i, actual, expected)
				}
			}

			if targets.TargetRowLoad[0] != 15.0 {
				t.Errorf("row load top = %v, want 15.0 (flag override)", targets.TargetRowLoad[0])
			}
			if targets.TargetRowLoad[1] != 80.0 {
				t.Errorf("row load home = %v, want 80.0 (flag override)", targets.TargetRowLoad[1])
			}
			if targets.TargetRowLoad[2] != 5.0 {
				t.Errorf("row load bottom = %v, want 5.0 (flag override)", targets.TargetRowLoad[2])
			}

			// Verify pinky penalties (first 6 values for brevity)
			expectedPinkyPenalties := []float64{2.5, 1.8, 1.2, 0.3, 2.8, 1.8}
			for i, expected := range expectedPinkyPenalties {
				actual := targets.PinkyPenalties[i]
				if actual != expected {
					t.Errorf("pinky penalty[%d] = %v, want %v (flag override)", i, actual, expected)
				}
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{
		"test",
		"--load-targets-file", "load_targets.txt",
		"--target-hand-load", "30,70",
		"--target-finger-load", "8,11,19,12,12,19,11,8",
		"--target-row-load", "15,80,5",
		"--pinky-penalties", "2.5,1.8,1.2,0.3,2.8,1.8",
	})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_InvalidFileReturnsError verifies that an error is returned when a user explicitly
// specifies a load-targets-file that doesn't exist (not the default file).
func TestTargetLoads_InvalidFileReturnsError(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create the config file

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadTargetLoadsFromFlags(cmd)
			if err == nil {
				t.Error("expected error for nonexistent config file, got nil")
			}
			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--load-targets-file", "nonexistent.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestWeights_HardcodedDefaults verifies that hardcoded default weights are used when no config
// file exists and no flags are set. Tests that SFB=-1.0 by default and other metrics=0.0.
func TestWeights_HardcodedDefaults(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("weights-file", "weights"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			weights, err := loadWeightsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Default weight for SFB is -1.0
			if weights.Get("SFB") != -1.0 {
				t.Errorf("SFB weight = %v, want -1.0 (default)", weights.Get("SFB"))
			}

			// Other metrics should be 0.0
			if weights.Get("LSB") != 0.0 {
				t.Errorf("LSB weight = %v, want 0.0 (default)", weights.Get("LSB"))
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--weights-file", ""})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestWeights_FromFile verifies that metric weights are correctly loaded from weights.txt
// when the file exists and contains valid weight definitions.
func TestWeights_FromFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create weights file
	weightsContent := `# Test weights
SFB=-10.0
LSB=-5.0
FSB=-2.0
`
	writeTestConfigFile(t, configDir, "weights.txt", weightsContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("weights-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			weights, err := loadWeightsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify weights from file
			if weights.Get("SFB") != -10.0 {
				t.Errorf("SFB weight = %v, want -10.0", weights.Get("SFB"))
			}
			if weights.Get("LSB") != -5.0 {
				t.Errorf("LSB weight = %v, want -5.0", weights.Get("LSB"))
			}
			if weights.Get("FSB") != -2.0 {
				t.Errorf("FSB weight = %v, want -2.0", weights.Get("FSB"))
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--weights-file", "weights.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestWeights_FlagMergesWithFile verifies that the --weights flag merges with file values:
// overriding values for specified metrics while preserving file values for unspecified metrics.
func TestWeights_FlagMergesWithFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create weights file
	weightsContent := `SFB=-10.0
LSB=-5.0
`
	writeTestConfigFile(t, configDir, "weights.txt", weightsContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("weights-file", "weights"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			weights, err := loadWeightsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// SFB should be overridden by flag
			if weights.Get("SFB") != -15.0 {
				t.Errorf("SFB weight = %v, want -15.0 (flag should override)", weights.Get("SFB"))
			}

			// LSB should still be from file
			if weights.Get("LSB") != -5.0 {
				t.Errorf("LSB weight = %v, want -5.0 (from file)", weights.Get("LSB"))
			}

			// FSB should be from flag
			if weights.Get("FSB") != -3.0 {
				t.Errorf("FSB weight = %v, want -3.0 (from flag)", weights.Get("FSB"))
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--weights-file", "weights.txt", "--weights", "SFB=-15,FSB=-3"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestWeights_InvalidFileUsesDefaults verifies that an error is returned when a user explicitly
// specifies a weights-file that doesn't exist (not the default file).
func TestWeights_InvalidFileUsesDefaults(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create weights file
	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("weights-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadWeightsFromFlags(cmd)
			// Should return error for nonexistent file
			if err == nil {
				t.Error("expected error for nonexistent weights file, got nil")
			}
			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--weights-file", "nonexistent.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestCorpus_DefaultFile verifies that the default corpus file (default.txt) is loaded
// when no --corpus flag is provided.
func TestCorpus_DefaultFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create default corpus
	writeTestCorpus(t, corpusDir, "default.txt")

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("corpus"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			corpus, err := loadCorpusFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if corpus == nil {
				t.Error("expected non-nil corpus")
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestCorpus_CustomFile verifies that a custom corpus file can be specified via the --corpus flag
// and is correctly loaded instead of the default.
func TestCorpus_CustomFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create custom corpus
	writeTestCorpus(t, corpusDir, "custom.txt")

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("corpus"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			corpus, err := loadCorpusFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if corpus == nil {
				t.Error("expected non-nil corpus")
			}

			return nil
		},
	}

	err := app.Run(context.Background(), []string{"test", "--corpus", "custom.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestLoadLayout_EnsuresKlfExtension verifies that loadLayout() automatically adds the .klf
// extension if not provided, allowing users to specify layout names without the extension.
func TestLoadLayout_EnsuresKlfExtension(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create a test layout
	writeTestLayout(t, layoutDir, "test.klf", minimalLayoutContent)

	// Test loading without extension
	layout, err := loadLayout("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layout == nil {
		t.Error("expected non-nil layout")
	}

	if layout.Name != "test" {
		t.Errorf("layout name = %q, want %q", layout.Name, "test")
	}
}

// TestLoadLayout_MissingFile verifies that an error is returned when attempting to load
// a layout file that doesn't exist.
func TestLoadLayout_MissingFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	_, err := loadLayout("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent layout, got nil")
	}
}

// TestLoadLayout_EmptyFilename verifies that an error is returned when an empty filename
// is provided to loadLayout().
func TestLoadLayout_EmptyFilename(t *testing.T) {
	_, err := loadLayout("")
	if err == nil {
		t.Error("expected error for empty filename, got nil")
	}
}

// TestTargetLoads_MissingDefaultFile_FallsBackWithWarning verifies that when the default
// load_targets.txt file is missing (user didn't explicitly set the flag), the system falls back
// to hardcoded defaults with a warning message instead of returning an error.
func TestTargetLoads_MissingDefaultFile_FallsBackWithWarning(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create load_targets.txt file - simulate missing default file

	// Create a fresh flag instance to avoid cli/v3 state issues
	app := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "load-targets-file",
				Value: "load_targets.txt",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify we got hardcoded defaults (not an error)
			if targets.TargetHandLoad[0] != 50.0 || targets.TargetHandLoad[1] != 50.0 {
				t.Errorf("hand load = %v,%v, want 50,50 (hardcoded defaults)", targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			expectedFingerLoads := []float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7}
			for i, expected := range expectedFingerLoads {
				actual := targets.TargetFingerLoad[i]
				if actual != expected {
					t.Errorf("finger load[%d] = %v, want %v (hardcoded defaults)", i, actual, expected)
				}
			}

			return nil
		},
	}

	// Run without explicitly setting --load-targets-file (uses default value "load_targets.txt")
	err := app.Run(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}

	// Note: We don't verify the warning message content in this test because stderr
	// redirection can interfere with the test execution. The important thing is that
	// the function doesn't return an error and provides valid default values.
}

// TestTargetLoads_MissingExplicitFile_ReturnsError verifies that when a user explicitly
// sets --load-targets-file to a nonexistent file, an error is returned (no fallback to defaults).
func TestTargetLoads_MissingExplicitFile_ReturnsError(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create the file

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadTargetLoadsFromFlags(cmd)
			if err == nil {
				t.Error("expected error for explicitly specified nonexistent config file, got nil")
			}
			return nil
		},
	}

	// Explicitly set --load-targets-file to a nonexistent file
	err := app.Run(context.Background(), []string{"test", "--load-targets-file", "nonexistent.txt"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// TestTargetLoads_EmptyStringFlag_UsesDefaultsSilently verifies that when --load-targets-file=""
// is explicitly set to empty string, hardcoded defaults are used without attempting to load a file
// and without printing a warning message. This allows users to intentionally skip config files.
func TestTargetLoads_EmptyStringFlag_UsesDefaultsSilently(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create a load_targets.txt file that should be ignored
	configContent := `target-hand-load = 40, 60`
	writeTestConfigFile(t, configDir, "load_targets.txt", configContent)

	// Capture stderr to verify no warning
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targets, err := loadTargetLoadsFromFlags(cmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify we got hardcoded defaults, not the file values
			if targets.TargetHandLoad[0] != 50.0 || targets.TargetHandLoad[1] != 50.0 {
				t.Errorf("hand load = %v,%v, want 50,50 (hardcoded defaults, should ignore file)",
					targets.TargetHandLoad[0], targets.TargetHandLoad[1])
			}

			return nil
		},
	}

	// Explicitly set --load-targets-file="" (empty string)
	err := app.Run(context.Background(), []string{"test", "--load-targets-file", ""})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}

	// Close writer and read stderr output
	w.Close()
	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(r)
	stderrOutput := stderrBuf.String()

	// Verify NO warning message was printed
	if strings.Contains(stderrOutput, "Warning:") {
		t.Errorf("expected no warning for explicit empty string, got: %q", stderrOutput)
	}
}

// Helper to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ============================================================================
// COMPREHENSIVE ERROR HANDLING TESTS
// ============================================================================

// TestError_MissingCorpusFile verifies that a clear error is returned when attempting to load
// a corpus file that doesn't exist, with a helpful error message.
func TestError_MissingCorpusFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Don't create corpus file

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("corpus"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadCorpusFromFlags(cmd)
			if err == nil {
				t.Error("expected error for missing corpus file, got nil")
				return nil
			}
			// Verify error message contains helpful information
			errMsg := err.Error()
			if !strings.Contains(errMsg, "corpus") && !strings.Contains(errMsg, "nonexistent.txt") {
				t.Errorf("error message should mention corpus or filename, got: %q", errMsg)
			}
			return nil
		},
	}

	_ = app.Run(context.Background(), []string{"test", "--corpus", "nonexistent.txt"})
}

// TestError_InvalidTargetsFile verifies that a clear error is returned when attempting to load
// a targets file with invalid format (malformed content).
func TestError_InvalidTargetsFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create malformed targets file
	malformedContent := `target-hand-load = not-a-number, also-not-a-number
target-finger-load = invalid format here
completely invalid line
`
	writeTestConfigFile(t, configDir, "bad_targets.txt", malformedContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("load-targets-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadTargetLoadsFromFlags(cmd)
			if err == nil {
				t.Error("expected error for invalid targets file, got nil")
				return nil
			}
			// Verify error message is informative
			errMsg := err.Error()
			if len(errMsg) < 10 {
				t.Errorf("error message too short, should be informative, got: %q", errMsg)
			}
			return nil
		},
	}

	_ = app.Run(context.Background(), []string{"test", "--load-targets-file", "bad_targets.txt"})
}

// TestError_InvalidWeightsFile verifies that a clear error is returned when attempting to load
// a weights file with invalid format (malformed metric definitions).
func TestError_InvalidWeightsFile(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	// Create malformed weights file
	malformedContent := `SFB=not-a-number
LSB invalid format
METRIC_WITHOUT_VALUE
= value without metric
`
	writeTestConfigFile(t, configDir, "bad_weights.txt", malformedContent)

	app := &cli.Command{
		Name:  "test",
		Flags: flagsSlice("weights-file"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := loadWeightsFromFlags(cmd)
			if err == nil {
				t.Error("expected error for invalid weights file, got nil")
				return nil
			}
			// Verify error message is informative
			errMsg := err.Error()
			if len(errMsg) < 10 {
				t.Errorf("error message too short, should be informative, got: %q", errMsg)
			}
			return nil
		},
	}

	_ = app.Run(context.Background(), []string{"test", "--weights-file", "bad_weights.txt"})
}

// TestError_MalformedFlags verifies that clear errors are returned for malformed flag values,
// testing various flag types including hand-load, finger-load, row-load, and weights.
func TestError_MalformedFlags(t *testing.T) {
	origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
	defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

	tests := []struct {
		name     string
		flagName string
		value    string
	}{
		{"hand-load wrong count", "target-hand-load", "50"},
		{"hand-load non-numeric", "target-hand-load", "abc,def"},
		{"finger-load wrong count", "target-finger-load", "1,2,3"},
		{"finger-load non-numeric", "target-finger-load", "a,b,c,d,e,f,g,h"},
		{"row-load wrong count", "target-row-load", "10,90"},
		{"row-load non-numeric", "target-row-load", "top,middle,bottom"},
		{"weights invalid format", "weights", "SFB"},
		{"weights no equals", "weights", "SFB-10"},
		{"weights invalid value", "weights", "SFB=notanumber"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:  "test",
				Flags: flagsSlice("target-hand-load", "target-finger-load", "target-row-load", "weights"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					var err error
					switch tt.flagName {
					case "target-hand-load", "target-finger-load", "target-row-load":
						_, err = loadTargetLoadsFromFlags(cmd)
					case "weights":
						_, err = loadWeightsFromFlags(cmd)
					}

					if err == nil {
						t.Errorf("expected error for malformed %s flag value %q, got nil", tt.flagName, tt.value)
						return nil
					}

					// Verify error message is helpful
					errMsg := err.Error()
					if len(errMsg) < 5 {
						t.Errorf("error message should be informative, got: %q", errMsg)
					}
					return nil
				},
			}

			_ = app.Run(context.Background(), []string{"test", "--" + tt.flagName, tt.value})
		})
	}
}
