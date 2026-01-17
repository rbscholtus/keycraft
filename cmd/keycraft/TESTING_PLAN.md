# Testing Plan for cmd/keycraft Package

## Overview

This document outlines a comprehensive testing strategy for the cmd/keycraft CLI package. The tests focus on the CLI interface layer, validating:
- Command parsing and flag handling
- Configuration loading precedence (hardcoded defaults → file config → flags)
- Input validation
- Integration between CLI flags and configuration files
- Error handling at the CLI boundary

**Out of scope**: Internal processing logic (internal/keycraft) and rendering (internal/tui).

## Architecture Summary

### Configuration Loading Hierarchy

```
Priority Level 1 (Lowest):  Hardcoded Defaults (internal/keycraft/analyser.go)
Priority Level 2:           Config Files (data/config/*.txt)
Priority Level 3 (Highest): CLI Flags (--target-*, --weights, etc.)
```

### Commands Overview

| Command | Aliases | Purpose | Key Flags |
|---------|---------|---------|-----------|
| `corpus` | `c` | Display corpus statistics | `--corpus`, `--corpus-rows`, `--coverage` |
| `view` | `v` | High-level layout analysis | `--corpus`, `--load-targets-file`, `--target-*` |
| `analyse` | `a` | Detailed layout analysis | `--corpus`, `--load-targets-file`, `--target-*`, `--rows`, `--compact-trigrams`, `--trigram-rows` |
| `rank` | `r` | Compare and rank layouts | `--corpus`, `--load-targets-file`, `--target-*`, `--weights-file`, `--weights`, `--metrics`, `--deltas`, `--output` |
| `flip` | `f` | Flip layout horizontally | (none) |
| `optimise` | `o` | Optimize layout with BLS | `--corpus`, `--load-targets-file`, `--target-*`, `--weights-file`, `--weights`, `--pins-file`, `--pins`, `--free`, `--generations`, `--maxtime`, `--seed`, `--log-file` |
| `generate` | `g` | Generate random layout | `--layout-type`, `--vowels-right`, `--alpha-thumb`, `--seed`, `--optimize`, `--generations`, plus all corpus/targets/weights flags when `--optimize` is used |

### Key Files to Test

- [main.go](main.go:1) - App entry point and directory configuration
- [flags.go](flags.go:1) - Shared flag definitions
- [helpers.go](helpers.go:1) - Configuration loading functions
- [corpus.go](corpus.go:1) - Corpus command
- [view.go](view.go:1) - View command
- [analyse.go](analyse.go:1) - Analyse command
- [rank.go](rank.go:1) - Rank command
- [flip.go](flip.go:1) - Flip command
- [optimise.go](optimise.go:1) - Optimise command
- [generate.go](generate.go:1) - Generate command

## Testing Strategy

### 1. Test File Structure

**Note**: The actual implementation consolidates tests into 4 files for better maintainability:

```
cmd/keycraft/
├── testing_utils.go          # Shared test infrastructure and utilities (NEW)
├── flags_test.go             # Flag parsing and validation tests (NEW)
│                             # - 6 test functions covering all 29 flags
├── config_loading_test.go    # Configuration loading hierarchy tests (NEW)
│                             # - 11 test functions for weights, targets, corpus
├── commands_test.go          # All command validation tests (NEW)
│                             # - 15 test functions covering all 7 commands
│                             # - Corpus, View, Analyse, Rank, Flip, Optimise, Generate
└── helpers_test.go           # Helper function tests (EXISTS)
                              # - 16 existing test functions
```

**Rationale for Consolidation:**
- Single `commands_test.go` instead of 7 separate files reduces duplication
- All commands share similar test patterns (validation, build input, flags)
- Easier to maintain and ensures consistent test patterns
- All test utilities centralized in `testing_utils.go`

### 2. Test Categories

#### A. Flag Validation Tests (`flags_test.go`)

Test that all flags defined in code are accessible and have correct types/defaults.

**Test Cases:**
- `TestAllFlagsExist` - Verify all flags in `appFlagsMap` exist
- `TestFlagDefaults` - Verify default values for all flags
- `TestFlagAliases` - Verify short aliases work correctly
- `TestFlagCategories` - Verify flags are categorized correctly

**Coverage:**
- All flags in `appFlagsMap` (corpus, load-targets-file, target-hand-load, target-finger-load, target-row-load, pinky-penalties, weights-file, weights)
- Command-specific flags (rows, compact-trigrams, trigram-rows, metrics, deltas, output, pins-file, pins, free, generations, maxtime, seed, log-file, corpus-rows, coverage)

#### B. Configuration Loading Tests (`config_loading_test.go`)

Test the precedence hierarchy: hardcoded defaults → config files → CLI flags.

**Test Cases:**

1. **Target Loads Configuration**
   - `TestTargetLoads_HardcodedDefaults` - When no config file and no flags, uses hardcoded defaults
   - `TestTargetLoads_FromFile` - When config file exists, loads from file
   - `TestTargetLoads_FromFile_PartialConfig` - When config file has some fields, fills missing with defaults
   - `TestTargetLoads_FlagOverridesFile` - When flag is set, overrides file value
   - `TestTargetLoads_FlagOverridesDefaults` - When flag is set but no file, overrides defaults
   - `TestTargetLoads_AllOverrides` - All target flags override file simultaneously
   - `TestTargetLoads_InvalidFileReturnsDefaults` - Invalid/missing config file falls back to defaults

2. **Weights Configuration**
   - `TestWeights_HardcodedDefaults` - When no config file and no flags, uses hardcoded default (SFB=-1.0)
   - `TestWeights_FromFile` - When config file exists, loads from file
   - `TestWeights_FlagOverridesFile` - When --weights flag is set, overrides file values
   - `TestWeights_FlagAddsToFile` - When --weights adds new metrics, merges with file
   - `TestWeights_InvalidFileReturnsDefaults` - Invalid/missing config file falls back to defaults

3. **Corpus Configuration**
   - `TestCorpus_DefaultFile` - Uses default corpus file (default.txt)
   - `TestCorpus_CustomFile` - Uses custom corpus file via --corpus flag
   - `TestCorpus_Coverage` - Coverage flag is applied correctly

**Implementation Approach:**
- Create temporary config files with known values
- Override `configDir` variable for tests
- Parse CLI args programmatically
- Use `loadTargetLoadsFromFlags()` and `loadWeightsFromFlags()` directly
- Assert configuration values match expected precedence

#### C. Command Tests

##### C1. Corpus Command Tests (`corpus_test.go`)

**Test Cases:**
- `TestCorpusCommand_NoArgs` - Corpus command accepts no arguments
- `TestCorpusCommand_WithArgs_ReturnsError` - Corpus command rejects arguments
- `TestCorpusCommand_DefaultCorpus` - Uses default corpus file
- `TestCorpusCommand_CustomCorpus` - Uses custom corpus via --corpus flag
- `TestCorpusCommand_CorpusRows` - Validates --corpus-rows flag
- `TestCorpusCommand_CorpusRowsInvalid` - Rejects invalid --corpus-rows (< 1)
- `TestCorpusCommand_Coverage` - Validates --coverage flag
- `TestCorpusCommand_CoverageInvalid` - Rejects invalid --coverage (out of 0.1-100 range)

##### C2. View Command Tests (`view_test.go`)

**Test Cases:**
- `TestViewCommand_NoArgs_ReturnsError` - Requires at least 1 layout
- `TestViewCommand_SingleLayout` - Accepts single layout argument
- `TestViewCommand_MultipleLayouts` - Accepts multiple layout arguments
- `TestViewCommand_WithTargetFlags` - Target flags are applied correctly
- `TestViewCommand_WithTargetsFile` - Loads targets from file
- `TestViewCommand_FlagsOverrideFile` - Target flags override file config
- `TestViewCommand_WithCorpus` - Custom corpus flag works

##### C3. Analyse Command Tests (`analyse_test.go`)

**Test Cases:**
- `TestAnalyseCommand_NoArgs_ReturnsError` - Requires at least 1 layout
- `TestAnalyseCommand_SingleLayout` - Accepts single layout argument
- `TestAnalyseCommand_MultipleLayouts` - Accepts multiple layout arguments
- `TestAnalyseCommand_WithTargetFlags` - Target flags are applied correctly
- `TestAnalyseCommand_RowsFlag` - Validates --rows flag
- `TestAnalyseCommand_RowsInvalid` - Rejects invalid --rows (< 1)
- `TestAnalyseCommand_CompactTrigrams` - Validates --compact-trigrams flag
- `TestAnalyseCommand_TrigramRows` - Validates --trigram-rows flag
- `TestAnalyseCommand_TrigramRowsInvalid` - Rejects invalid --trigram-rows (< 1)

##### C4. Rank Command Tests (`rank_test.go`)

**Test Cases:**
- `TestRankCommand_NoArgs` - Uses all layouts in directory when no args
- `TestRankCommand_SingleLayout` - Accepts single layout argument
- `TestRankCommand_MultipleLayouts` - Accepts multiple layout arguments
- `TestRankCommand_WithWeightsFile` - Loads weights from file
- `TestRankCommand_WithWeightsFlag` - Applies --weights flag
- `TestRankCommand_WeightsFlagOverridesFile` - Weights flag overrides file
- `TestRankCommand_MetricsWeighted` - Default metrics="weighted" works
- `TestRankCommand_MetricsAll` - metrics="all" displays all metrics
- `TestRankCommand_MetricsCustom` - Custom comma-separated metrics list
- `TestRankCommand_MetricsInvalid` - Rejects invalid metric names
- `TestRankCommand_DeltasNone` - deltas="none" (default)
- `TestRankCommand_DeltasRows` - deltas="rows" shows row-by-row deltas
- `TestRankCommand_DeltasMedian` - deltas="median" shows vs median
- `TestRankCommand_DeltasCustomLayout` - deltas="<layout>" compares against specific layout
- `TestRankCommand_OutputTable` - output="table" (default)
- `TestRankCommand_OutputHTML` - output="html" format
- `TestRankCommand_OutputCSV` - output="csv" format
- `TestRankCommand_OutputInvalid` - Rejects invalid output format

##### C5. Flip Command Tests (`flip_test.go`)

**Test Cases:**
- `TestFlipCommand_NoArgs_ReturnsError` - Requires exactly 1 layout
- `TestFlipCommand_MultipleArgs_ReturnsError` - Rejects multiple arguments
- `TestFlipCommand_ValidLayout` - Flips layout and saves with "-flipped" suffix
- `TestFlipCommand_OutputPath` - Verifies output path is correct

##### C6. Optimise Command Tests (`optimise_test.go`)

**Test Cases:**
- `TestOptimiseCommand_NoArgs_ReturnsError` - Requires exactly 1 layout
- `TestOptimiseCommand_MultipleArgs_ReturnsError` - Rejects multiple arguments
- `TestOptimiseCommand_ValidLayout` - Accepts single layout
- `TestOptimiseCommand_WithPinsFile` - Loads pins from file
- `TestOptimiseCommand_WithPinsFlag` - Applies --pins flag
- `TestOptimiseCommand_WithFreeFlag` - Applies --free flag (overrides pins)
- `TestOptimiseCommand_Generations` - Validates --generations flag
- `TestOptimiseCommand_GenerationsZero` - Rejects --generations=0
- `TestOptimiseCommand_MaxTime` - Validates --maxtime flag
- `TestOptimiseCommand_MaxTimeZero` - Rejects --maxtime=0
- `TestOptimiseCommand_Seed` - Validates --seed flag (allows 0)
- `TestOptimiseCommand_LogFile` - Validates --log-file flag
- `TestOptimiseCommand_WithWeights` - Loads weights correctly
- `TestOptimiseCommand_WithTargets` - Loads targets correctly

##### C7. Generate Command Tests (`generate_test.go`)

**Test Cases:**
- `TestGenerateCommand_NoArgs` - Generate command accepts no arguments (auto-generates name)
- `TestGenerateCommand_WithArgs_ReturnsError` - Rejects positional arguments
- `TestGenerateCommand_LayoutTypeRowstag` - Validates --layout-type=rowstag
- `TestGenerateCommand_LayoutTypeAnglemod` - Validates --layout-type=anglemod
- `TestGenerateCommand_LayoutTypeOrtho` - Validates --layout-type=ortho
- `TestGenerateCommand_LayoutTypeColstag` - Validates --layout-type=colstag (default)
- `TestGenerateCommand_LayoutTypeInvalid` - Rejects invalid layout type
- `TestGenerateCommand_VowelsRight` - Validates --vowels-right flag
- `TestGenerateCommand_AlphaThumb` - Validates --alpha-thumb flag
- `TestGenerateCommand_Seed` - Validates --seed flag (allows 0)
- `TestGenerateCommand_OptimizeFlag` - Validates --optimize flag
- `TestGenerateCommand_GenerationsWithOptimize` - Validates --generations with --optimize
- `TestGenerateCommand_BuildInput` - Tests buildGeneratorInput() function

#### D. Integration Tests

##### D1. End-to-End Configuration Tests

**Test Cases:**
- `TestE2E_WeightsHierarchy` - Full hierarchy: defaults → file → flags
- `TestE2E_TargetsHierarchy` - Full hierarchy: defaults → file → flags
- `TestE2E_CompleteOverride` - All flags set, all override file config
- `TestE2E_NoConfigFiles` - No config files present, uses all defaults
- `TestE2E_EmptyConfigFiles` - Empty config files, uses all defaults

##### D2. Validation Tests

**Test Cases:**
- `TestValidation_TargetHandLoad_Format` - Validates format: 2 comma-separated values
- `TestValidation_TargetFingerLoad_Format` - Validates format: 4 or 8 values
- `TestValidation_TargetRowLoad_Format` - Validates format: 3 values
- `TestValidation_PinkyPenalties_Format` - Validates format: 6 or 12 values
- `TestValidation_Weights_Format` - Validates format: metric=value pairs
- `TestValidation_LayoutFileExtension` - Tests .klf extension handling

##### D3. Error Handling Tests

**Test Cases:**
- `TestError_MissingLayoutFile` - Returns clear error for missing layout
- `TestError_MissingCorpusFile` - Returns clear error for missing corpus
- `TestError_InvalidTargetsFile` - Returns clear error for invalid targets file
- `TestError_InvalidWeightsFile` - Returns clear error for invalid weights file
- `TestError_MalformedFlags` - Returns clear error for malformed flag values

## Implementation Guidelines

### Test Utilities

Create shared test utilities in a new file `testing_utils.go`:

```go
// setupTestDirs creates temporary directories for testing
func setupTestDirs(t *testing.T) (layoutDir, corpusDir, configDir string)

// writeTestConfigFile writes a config file for testing
func writeTestConfigFile(t *testing.T, dir, name, content string) string

// writeTestCorpus writes a minimal test corpus
func writeTestCorpus(t *testing.T, dir, name string) string

// parseTestFlags simulates CLI flag parsing
func parseTestFlags(args []string) (*cli.Context, error)

// getTestApp returns a test CLI app instance
func getTestApp() *cli.Command
```

### Mocking Strategy

**Don't mock:**
- Flag parsing (use real urfave/cli/v3)
- Configuration loading functions (test the real implementations)

**Stub/mock at boundaries:**
- File I/O for layouts/corpus (use temp directories)
- Calls to `internal/keycraft` functions (mock or use test doubles)
- Calls to `internal/tui` rendering (mock or capture output)

### Test Data

Create minimal test fixtures:

```
testdata/
├── layouts/
│   ├── test-layout.klf           # Minimal valid layout
│   └── test-layout-2.klf         # Second layout for comparison
├── corpus/
│   └── test-corpus.txt           # Minimal corpus (few words)
└── config/
    ├── weights.txt               # Test weights config
    ├── load_targets.txt          # Test targets config
    ├── weights_partial.txt       # Partial weights (for testing merge)
    └── load_targets_partial.txt  # Partial targets (for testing defaults)
```

### Assertion Patterns

For configuration loading tests:

```go
// Test hardcoded defaults
func TestTargetLoads_HardcodedDefaults(t *testing.T) {
    // Setup: no config file, no flags
    app := getTestApp()
    c := parseTestFlags([]string{"view", "layout.klf"})

    // Load targets
    targets, err := loadTargetLoadsFromFlags(c)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Assert: verify hardcoded defaults
    expected := kc.DefaultTargetHandLoad()
    if !reflect.DeepEqual(targets.HandLoad, expected) {
        t.Errorf("hand load = %v, want %v", targets.HandLoad, expected)
    }
}
```

For flag override tests:

```go
// Test flags override file
func TestTargetLoads_FlagOverridesFile(t *testing.T) {
    // Setup: create config file with known values
    tmpDir := t.TempDir()
    configDir = tmpDir
    configFile := writeTestConfigFile(t, tmpDir, "load_targets.txt",
        "target-hand-load: 40, 60\n")

    // Parse with override flag
    c := parseTestFlags([]string{
        "view", "layout.klf",
        "--target-hand-load", "30,70",
    })

    // Load targets
    targets, err := loadTargetLoadsFromFlags(c)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Assert: flag value overrides file
    if targets.HandLoad.Left != 30.0 || targets.HandLoad.Right != 70.0 {
        t.Errorf("hand load = %v,%v, want 30,70",
            targets.HandLoad.Left, targets.HandLoad.Right)
    }
}
```

### Testing Commands Without Rendering

To test commands without triggering rendering:

1. **Extract input building logic**: Already done (e.g., `buildViewInput()`)
2. **Test input building only**: Call `buildViewInput()` directly
3. **Mock internal/keycraft functions**: Use interfaces or test doubles
4. **Test validation functions**: Call `validateViewFlags()` directly

Example:

```go
func TestViewCommand_BuildInput(t *testing.T) {
    // Setup test environment
    tmpDir := setupTestDirs(t)
    defer cleanupTestDirs(tmpDir)

    // Create test layout and corpus
    writeTestLayout(t, tmpDir+"/layouts", "test.klf", validLayoutContent)
    writeTestCorpus(t, tmpDir+"/corpus", "default.txt")

    // Parse flags
    c := parseTestFlags([]string{
        "view", "test.klf",
        "--target-hand-load", "40,60",
    })

    // Test: build input (doesn't call ViewLayouts or render)
    input, err := buildViewInput(c)
    if err != nil {
        t.Fatalf("buildViewInput failed: %v", err)
    }

    // Assert: verify input was built correctly
    if len(input.LayoutFiles) != 1 {
        t.Errorf("got %d layouts, want 1", len(input.LayoutFiles))
    }
    if input.Targets.HandLoad.Left != 40.0 {
        t.Errorf("left hand load = %v, want 40", input.Targets.HandLoad.Left)
    }
}
```

## Test Coverage Goals

| Category | Target Coverage |
|----------|-----------------|
| Flag parsing | 100% of flags |
| Config loading (helpers.go) | 100% of functions |
| Command validation (Before hooks) | 100% of validation functions |
| Input building (build*Input functions) | 100% of functions |
| Error paths | 80%+ of error cases |

## Test Execution

### Running Tests

```bash
# Run all cmd/keycraft tests
go test ./cmd/keycraft/

# Run specific test file
go test ./cmd/keycraft/ -run TestConfigLoading

# Run with coverage
go test ./cmd/keycraft/ -cover

# Run with verbose output
go test ./cmd/keycraft/ -v

# Generate coverage report
go test ./cmd/keycraft/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Organization

Use table-driven tests where possible:

```go
func TestTargetHandLoad_Format(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        wantL   float64
        wantR   float64
    }{
        {
            name:    "valid 2 values",
            input:   "40,60",
            wantErr: false,
            wantL:   40.0,
            wantR:   60.0,
        },
        {
            name:    "invalid 1 value",
            input:   "50",
            wantErr: true,
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic here
        })
    }
}
```

## Automation

### CI Integration

Add to CI pipeline (e.g., GitHub Actions):

```yaml
- name: Run CLI tests
  run: |
    go test ./cmd/keycraft/ -v -cover
    go test ./cmd/keycraft/ -coverprofile=coverage.out
    go tool cover -func=coverage.out
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
go test ./cmd/keycraft/ || exit 1
```

## Success Criteria

Tests are considered complete when:

1. ✅ All flags in code are covered by tests
2. ✅ All three configuration precedence levels are tested
3. ✅ All commands have validation tests
4. ✅ All `build*Input()` functions have unit tests
5. ✅ All error paths have tests
6. ✅ Test coverage > 80% for cmd/keycraft package
7. ✅ Tests run automatically in CI
8. ✅ Tests are documented and maintainable

## Next Steps

### Phase 1: Foundation (Priority 1)
1. Create `testing_utils.go` with shared test utilities
2. Create `testdata/` directory with minimal fixtures
3. Implement `flags_test.go` - validate all flags exist and have correct defaults

### Phase 2: Core Configuration (Priority 1)
4. Implement `config_loading_test.go` - test precedence hierarchy
5. Test `loadTargetLoadsFromFlags()` thoroughly
6. Test `loadWeightsFromFlags()` thoroughly

### Phase 3: Command Tests (Priority 2)
7. Implement validation tests for each command (Before hooks)
8. Implement input building tests for each command (build*Input functions)
9. Test error cases for each command

### Phase 4: Integration (Priority 2)
10. Implement end-to-end configuration tests
11. Implement comprehensive validation tests
12. Implement error handling tests

### Phase 5: Automation (Priority 3)
13. Add tests to CI pipeline
14. Set up coverage reporting
15. Document test patterns for future contributors

## Appendix A: Flag Reference

### Shared Flags (appFlagsMap)

| Flag | Aliases | Type | Default | Category |
|------|---------|------|---------|----------|
| `--corpus` | `-c` | string | `default.txt` | General |
| `--load-targets-file` | `-ldt` | string | `load_targets.txt` | Targets and Weights |
| `--target-hand-load` | `-thl` | string | (none) | Targets and Weights |
| `--target-finger-load` | `-tfl` | string | (none) | Targets and Weights |
| `--target-row-load` | `-trl` | string | (none) | Targets and Weights |
| `--pinky-penalties` | `-pp` | string | (none) | Targets and Weights |
| `--weights-file` | `-wf` | string | `weights.txt` | Targets and Weights |
| `--weights` | `-w` | string | (none) | Targets and Weights |

### Command-Specific Flags

#### Corpus Command
| Flag | Aliases | Type | Default | Validation |
|------|---------|------|---------|------------|
| `--corpus-rows` | `-cr` | int | 100 | ≥ 1 |
| `--coverage` | (none) | float64 | 98.0 | 0.1-100.0 |

#### Analyse Command
| Flag | Aliases | Type | Default | Validation |
|------|---------|------|---------|------------|
| `--rows` | `-r` | int | 10 | ≥ 1 |
| `--compact-trigrams` | (none) | bool | false | N/A |
| `--trigram-rows` | (none) | int | 50 | ≥ 1 |

#### Rank Command
| Flag | Aliases | Type | Default | Validation |
|------|---------|------|---------|------------|
| `--metrics` | `-m` | string | `weighted` | Valid metrics set or comma-separated list |
| `--deltas` | `-d` | string | `none` | "none", "rows", "median", or layout name |
| `--output` | `-o` | string | `table` | "table", "html", or "csv" |

#### Optimise Command
| Flag | Aliases | Type | Default | Validation |
|------|---------|------|---------|------------|
| `--pins-file` | `-pf` | string | (none) | Valid file path |
| `--pins` | `-p` | string | (none) | Valid characters |
| `--free` | `-f` | string | (none) | Valid characters |
| `--generations` | `-gens`, `-g` | uint | 1000 | > 0 |
| `--maxtime` | `-mt` | uint | 5 | > 0 |
| `--seed` | `-s` | int64 | 0 | Any |
| `--log-file` | `-lf` | string | (none) | Valid file path |

#### Generate Command
| Flag | Aliases | Type | Default | Validation |
|------|---------|------|---------|------------|
| `--layout-type` | `-lt` | string | `colstag` | "rowstag", "anglemod", "ortho", or "colstag" |
| `--vowels-right` | `-vr` | bool | false | N/A |
| `--alpha-thumb` | `-at` | bool | false | N/A |
| `--seed` | `-s` | uint64 | 0 | Any |
| `--optimize` | `-opt` | bool | false | N/A |
| `--generations` | `-gens`, `-g` | uint | 1000 | > 0 (when --optimize is used) |

## Appendix B: Configuration File Formats

### load_targets.txt Format

```
# Comments start with #
target-hand-load: 50, 50
target-finger-load: 7, 10, 16, 17, 17, 16, 10, 7
target-row-load: 17.5, 75.0, 7.5
pinky-penalties: 2.0, 1.5, 1.0, 0.0, 2.0, 1.5, 2.0, 1.5, 1.0, 0.0, 2.0, 1.5
```

### weights.txt Format

```
# Comments start with #
SFB=-10.0
LSB=-5.0
Scissors=-2.0
# Can be comma or space separated
SFS=-3.0, FSS=-1.5
```

## Appendix C: Hardcoded Defaults Reference

### Hand Load Defaults
```go
Left: 50.0%
Right: 50.0%
```

### Finger Load Defaults
```go
F0 (L-pinky):  7.0%      F6 (R-index): 17.0%
F1 (L-ring):   10.0%     F7 (R-middle): 16.0%
F2 (L-middle): 16.0%     F8 (R-ring):   10.0%
F3 (L-index):  17.0%     F9 (R-pinky):  7.0%
F4, F5 (thumbs): 0.0%
```

### Row Load Defaults
```go
Top row:    17.5%
Home row:   75.0%
Bottom row: 7.5%
```

### Pinky Penalties Defaults
```go
Per hand (left then right):
Top-outer:      2.0
Top-inner:      1.5
Home-outer:     1.0
Home-inner:     0.0
Bottom-outer:   2.0
Bottom-inner:   1.5
```

### Weight Defaults
```go
SFB: -1.0
(all others default to 0.0)
```
