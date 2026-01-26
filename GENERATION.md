Keyboard Layout Generation
==========================

Purpose
-------
This document defines the scope and design of the generation feature built into Keycraft.

Two iterations of a generator have been built previously:
- A generator that generates keyboards with some of the 13 most frequent letters placed at the home keys and thumbs. However, there are too many possible combinations of 13 letters to generate and optimize them all efficiently.
- A generator that fixes the letters IHEA on the right hand and iterates through 6 other high-frequency letters on the left hand and thumb. This approach is not flexible enough for exploring different layout strategies.

The new generator completely replaces the existing generation code and tests. This flexible layout generator will:
- Allow users to specify their layout preferences in a `.gen` (generation config) file
- Generate layouts according to user-defined rules for fixed characters, character groups, and random positions
- Optimize each generated layout (if specified on the CLI) with automatic pinning of structural characters during the optimization process

Generation Config File Format
------------------------------

The config file for the generation process uses the `.gen` extension and has this structure:

```
<<LAYOUT_TYPE>>
~ 0 0 0 0 0   0 0 o u 0 ~
~ 2 2 2 2 0   0 0 a e i ~
~ 0 0 0 0 0   0 0 0 0 0 ~
      ~ ~ 1   _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz,./;'_
set1=rn
set2=tnshrd
```

**File structure:**
1. **Layout type** (first line): One of the 4 supported keyboard geometries: `rowstag`, `anglemod`, `ortho`, or `colstag`
2. **42-position template** (next 4 lines): Uses the same visual format as `.klf` layout files, showing 3 rows of 12 keys plus 1 row of 6 thumb keys
3. **Character definitions** (remaining lines): Define the character sets for allocation

**Position markers** in the 42-position template:
- `~` = Unused position (no character allocated here)
- `_` = Fixed space character
- Any other character (letter, punctuation, special character) = Fixed character at this position
- `0` = Random allocation from remaining charset (after sets 1-9 are allocated)
- `1` through `9` = Group number (allocate from corresponding `set1`, `set2`, etc.)

**Character set definitions:**
- `charset=<characters>` = All characters available for allocation across the entire layout. This includes 26 letters, punctuation, and space (use `_` to represent space). A standard charset has 32 characters. Users may include more characters if there are enough positions available.
- `set1=<characters>`, `set2=<characters>`, etc. = Character pools for groups 1-9. Sets are processed in ascending order (set1 first, then set2, etc.). After all sets are allocated, remaining charset characters are assigned to `0` positions.

**Comments and blank lines:**
- Lines starting with `#` are treated as comments and ignored
- Blank lines are ignored
- Comment handling follows the same rules as `.klf` files

CLI Design
----------

### Command Structure

```bash
keycraft generate <config-file.gen> [flags]
```

The command is `generate` (alias: `g`). Config file is a required positional argument with `.gen` extension.

**Config file resolution:**
- If the specified path exists, it's used directly
- Otherwise, the file is searched in `data/config/` directory
- This allows simple references like `example.gen` instead of `data/config/example.gen`

**Note:** This completely replaces the old `generate` command. The previous generator code and tests are archived as `*.bak` files.

### Flags

**Generation Flags:**
- `--max-layouts`, `-m` (int, default=1500): Maximum number of permutations to generate. Set to 0 to generate all permutations.
- `--seed`, `-s` (uint64, default=0): Random seed for random position allocation (0=timestamp). Seed is incremented for each permutation to vary random fills.

**Optimization Flags:**
- `--optimize`, `-o` (bool): Run optimization after generation
- `--generations`, `-g` (uint, default=1000): Optimization iterations
- `--pins`, `-p` (string): Characters to pin during optimization (e.g., 'aeiouy'). Overrides default pins. Unused positions and space are always pinned.
- `--keep-unoptimized`, `-k` (bool): Keep unoptimized layouts when using --optimize. By default, unoptimized layouts are deleted after optimization.

**Common Flags** (when --optimize used):
- All existing corpus, targets, weights flags from optimize command

### Flag Behavior

**Permutation Generation:**
- Generates all possible permutations of group characters (set1, set2, etc.)
- For N characters in N positions: generates N! layouts
- Multiple groups: generates cartesian product of all group permutations
- Example: set1 with 3 chars (3!=6) and set2 with 4 chars (4!=24) → 6×24=144 layouts
- Warning: Prints warning if > 1,000 permutations

**Max-Layouts Limiting:**
- `--max-layouts 0`: Generate all permutations (no limit)
- `--max-layouts N` (default 1500): Generate only first N permutations

**Seed Handling:**
- `--seed 0` (default): timestamp-based for random positions
- `--seed N`: reproducible random position allocation
- Each permutation uses seed+i for its random positions (i = permutation index)
- Group positions are always deterministic (permutation-based, no randomness)

**Optimization Cleanup:**
- Without `--keep-unoptimized`: Deletes original layouts after optimization, keeps only `-best` versions
- With `--keep-unoptimized`: Keeps both original and optimized versions
- Reduces clutter when generating many permutations

**Default Pinning (when --optimize used without --pins):**
- Pins all non-0 positions (fixed chars + group chars)
- Only 0 positions are free to move during optimization
- This preserves the structure defined in config file

**Pin Override (--pins flag):**
- When specified, overrides all default pins
- Only the specified characters are pinned (plus unused positions and space)
- Format: string of characters (e.g., 'aeiouy')
- Example: `--pins "eaio"` allows all other characters to move, but keeps e, a, i, o, space, and unused positions pinned
- This allows fine-grained control over which characters can move during optimization

### Example Usage

```bash
# Generate up to 1000 permutations (default limit)
keycraft generate example.gen

# Or using alias
keycraft g example.gen

# Generate all permutations (no limit)
keycraft generate example.gen --max-layouts 0

# Generate only first 10 permutations
keycraft generate example.gen --max-layouts 10

# Reproducible generation (same seed for random positions across permutations)
keycraft generate example.gen --seed 42

# Generate and optimize (pins all non-0 chars by default, deletes unoptimized)
keycraft generate example.gen --optimize

# Generate and optimize, keeping both original and optimized versions
keycraft generate example.gen --optimize --keep-unoptimized

# Generate and optimize with custom pins (override default, only pin specified chars + space)
keycraft generate example.gen --optimize --pins "eaio"

# Generate first 5 permutations with more optimization iterations
keycraft generate example.gen --max-layouts 5 --optimize --generations 2000
```

generateAction Flow
-------------------

The `generateAction()` function in `cmd/keycraft/generate.go` orchestrates the full generation workflow:

### Step 1: buildGenerateInput()

The `buildGenerateInput()` function handles all generation-specific input:
- Validates exactly 1 argument (config file path)
- Validates `.gen` extension (case-insensitive)
- Resolves config file path (checks as-is, then `data/config/`)
- Captures generation flags

```go
type GenerateInput struct {
    ConfigPath      string // resolved .gen file path
    MaxLayouts      int    // from --max-layouts (default 1500, 0=all)
    Seed            uint64 // from --seed (0=timestamp)
    Optimize        bool   // from --optimize flag
    KeepUnoptimized bool   // from --keep-unoptimized flag
}
```

### Step 2: If --optimize, buildOptimizeInput()

Reuses the existing `buildOptimizeInput()` from `optimize.go`, adapted with a `skipLayoutLoad` parameter:
- Loads corpus from flags (--corpus)
- Loads targets from flags (--load-targets-file, --target-* flags)
- Loads weights from flags (--weights-file, --weights)
- Captures optimization flags (generations, maxtime, seed)

```go
// Signature: buildOptimizeInput(c *cli.Command, layout *kc.SplitLayout, skipLayoutLoad bool)
// - optimize command calls: buildOptimizeInput(c, nil, false) → loads layout from args
// - generate command calls: buildOptimizeInput(c, nil, true)  → skips layout loading
```

The existing `kc.OptimizeInput` struct is reused. For generate, the `Layout` and `Pinned` fields are set later per generated layout.

Generate-specific flags (`--pins`, `--keep-unoptimized`) are captured separately and applied per layout.

### Step 3: buildRankingInput()

Reuses `buildRankingInput()` from `rank.go` with `skipLayoutsFromArgs=true`:
- Loads corpus, targets from flags
- Reuses weights from OptimizeInput if available, otherwise loads from flags
- Layout files are set after generation (not loaded from CLI args)

```go
// Signature: buildRankingInput(c *cli.Command, weights *kc.Weights, skipLayoutsFromArgs bool)
// - rank command calls: buildRankingInput(c, weights, false) → loads layouts from args
// - generate command calls: buildRankingInput(c, weights, true) → skips layout loading
```

### Step 4: Execute generation (stub for now)

Later implementation will:
- Call `ParseConfigFile()` to parse the .gen file
- Call `ValidateConfig()` to validate
- Call generation functions from internal/keycraft/
- Save generated layouts

### Step 5: Print rankings

Display rankings for generated layouts using `RankingInput`:
- Set `LayoutFiles` to the generated layout paths
- Call `kc.ComputeRankings()` and `tui.RenderRankingTable()`
- If --optimize was used: optionally delete unoptimized layouts (unless --keep-unoptimized)

**Current stub implementation:**
- Prints `GenerateInput`, `OptimizeInput`, and `RankingInput` structs for verification
- Returns without actual generation

CLI Testing
-----------

The CLI layer (`cmd/keycraft/generate.go`) is tested separately from the generation logic. Tests verify that flags are correctly parsed and validation works properly.

### Test Cases (commands_test.go)

**Argument Validation:**
1. No arguments → error: "expected exactly 1 config file argument, got 0"
2. Multiple arguments → error: "expected exactly 1 config file argument, got 2"
3. File without .gen extension → error: "config file must have .gen extension"
4. Valid .gen file → success

**Flag Defaults:**
- `--max-layouts` defaults to 1500
- `--seed` defaults to 0
- `--optimize` defaults to false
- `--generations` defaults to 1000
- `--pins` defaults to empty string
- `--keep-unoptimized` defaults to false

**Flag Parsing:**
1. `--max-layouts 10` → maxLayouts=10
2. `--max-layouts 0` → maxLayouts=0 (generate all)
3. `--seed 42` → seed=42
4. `--optimize` → optimize=true
5. `--optimize --generations 2000` → generations=2000
6. `--optimize --pins "eaio"` → pins="eaio"
7. `--optimize --keep-unoptimized` → keepUnoptimized=true
8. Short flags: `-m 10 -s 42 -o -g 2000 -p "eaio" -k`

**Config File Resolution:**
1. Absolute path exists → use as-is
2. Relative path exists → use as-is
3. Name only, exists in `data/config/` → use `data/config/<name>`
4. Not found anywhere → error with helpful message

**Error Messages:**
- All errors should be clear and actionable
- Include actual vs expected values where applicable
- File-not-found errors should mention both locations tried

Algorithm Design
----------------

### Data Structures

The implementation will use these primary structures:

```go
// GenerationConfig represents a parsed .klg file
type GenerationConfig struct {
    LayoutType   LayoutType        // ROWSTAG, ANGLEMOD, ORTHO, or COLSTAG
    Template     [42]PositionSpec  // Specification for each position
    Charset      []rune            // All characters to allocate
    Groups       map[int][]rune    // Group number -> character set
}

// PositionSpec defines what should go in a position
type PositionSpec struct {
    Type      TemplateType
    FixedChar rune  // used when Type == TemplateFixed
    GroupNum  int   // used when Type == TemplateGroup
}

type TemplateType uint8
const (
    TemplateExcluded   TemplateType = iota // ~ (not allocated)
    TemplateFixed                           // specific char or _
    TemplateRandom                          // ? (random from charset)
    TemplateGroup                           // 1-9 (from group set)
)
```

### Generation Algorithm

The generation process follows these steps:

1. **Parse .gen file**
   - Read layout type (first line)
   - Read 42-position template (4 lines in .klf format)
   - Parse charset= and setN= definitions
   - Validate all inputs immediately

2. **Build permutations** (NEW)
   - For each group (set1, set2, etc.), compute all possible permutations of its characters
   - Calculate total permutation count (product of factorials)
   - Validate permutation count doesn't exceed 10,000 (hard limit)
   - If multiple groups: compute cartesian product of all group permutations

3. **For each permutation combination:**

   a. **Allocate fixed characters**
      - For each position with TemplateFixed, place the character
      - Remove character from working sets to prevent duplication

   b. **Allocate group permutation**
      - For each group (1, 2, 3...):
        - Use the specific permutation for this group (deterministic, not random)
        - Assign permuted characters to group positions
        - Remove from working sets to prevent duplication

   c. **Allocate random positions**
      - Initialize RNG using `NewLockedRNG(seed+i, 0)` where i is permutation index
      - Shuffle remaining characters using `ShuffleSlice()`
      - Assign to all `?` positions

   d. **Create layout**
      - Call `NewSplitLayout(name, layoutType, runes)`
      - Generate layout name with permutation index
      - Name format: `_` + homekeys + thumbkeys + `-` + 4-digit index

**Example**: Config with `set2=tnrd` (4 characters for 4 positions) generates 4! = 24 layouts systematically.

**Permutation vs Random**:
- Group positions: All permutations generated deterministically
- Random positions (`?`): Allocated randomly per permutation (using seed+i)
- Result: Complete exploration of group character arrangements

### Validation Checks

All validation happens before generation:

- Layout type must be one of: `rowstag`, `anglemod`, `ortho`, or `colstag`
- Exactly 42 positions in template (3 rows × 12 keys + 1 row × 6 thumb keys)
- Position markers must be one of: `~`, `_`, `?`, `1`-`9`, or any valid character (letters, punctuation, special characters)
- `charset=` must be defined and non-empty
- No duplicate characters in `charset`
- Each group number used in the template (1-9) must have a corresponding `setN=` definition
- No duplicate characters within any `setN`
- Each `setN` must have enough unique characters to fill all positions assigned to that group
- Total `charset` size must equal the number of non-excluded positions (all positions except those marked with `~`)
- No fixed character appears multiple times in the template

Validation errors report line numbers and specific issues for easy debugging.

Test Plan
---------

### Unit Tests (generator_test.go)

**Parser Tests:**
1. Valid config with all features (groups, fixed, random, excluded)
2. Valid minimal config (only fixed + random)
3. Invalid layout type → error with line number
4. Wrong position count (not 42) → error
5. Missing charset → error
6. Missing setN for used group → error
7. Malformed key=value line → error with line number
8. Comments (# lines) and empty lines ignored correctly
9. Special characters in fixed positions parsed correctly (e.g., `'`, `,`, `.`, `;`)

**Validation Tests:**
1. Empty charset → error
2. Duplicate character in charset → error
3. Group used but setN not defined → error
4. setN defined but group not used → warning
5. Duplicate character in setN → error
6. setN has too few characters for positions → error
7. Total characters mismatch (too few) → error
8. Total characters mismatch (too many) → error
9. Fixed character appears multiple times → error

**Generation Tests:**
1. Simple config (fixed + random) generates valid layout
2. Config with single group allocates correctly
3. Config with multiple groups allocates without overlap
4. Seed reproducibility (same seed → same layout)
5. Different seeds → different layouts
6. All positions fixed → minimal randomness
7. All positions random → maximum randomness
8. Group characters removed from other groups after allocation
9. Charset completely consumed (no leftover characters)

**Edge Cases:**
1. Single group with exact character count
2. Multiple groups initially sharing characters
3. Large group (20+ positions)
4. Config with all 31 DefaultChars
5. Config with punctuation in groups
6. All vowels in one group, consonants in another
7. Empty group sets (all chars already allocated)

**Name Generation:**
1. Layout name format: `_<homekeys><thumbkeys>-<4hex>`
2. Name uses only a-z from positions 13-16, 19-22, 36-41
3. Hex suffix is exactly 4 characters (0-padded)
4. Different timestamps → different hex suffixes
5. `generateLayoutName()` method added to `SplitLayout` with `_` prefix

### Integration Tests (commands_test.go)

**CLI Argument Tests:**
1. No arguments → error
2. Multiple arguments → error
3. File without .klg extension → error (or warning/acceptance?)
4. Valid .klg file → success

**CLI Flag Tests:**
1. --count flag creates N layouts
2. --seed produces reproducible results
3. --seed 0 produces different results each run
4. --optimize runs optimization
5. --pins overrides default pinning
6. --generations affects optimization iterations
7. All corpus/weights/targets flags work with --optimize

**File I/O Tests:**
1. Generated layout saves to data/layouts/<name>.klf
2. Layout name has correct format
3. Optimized layout saves with different name
4. Multiple layouts (--count N) all save successfully
5. Saved layout can be reloaded with existing parser

**Error Handling Tests:**
1. Invalid .klg file shows clear error with line number
2. Missing .klg file shows file-not-found error
3. Validation error shows all issues at once
4. Generation error shows specific problem

### Test Data Files

Create testdata/ directory with sample .klg files:

**Valid configs:**
- `testdata/simple.klg` - basic fixed + random
- `testdata/example.klg` - the IHEA example from spec
- `testdata/all-groups.klg` - using groups 1-9
- `testdata/minimal.klg` - smallest valid config
- `testdata/maximal.klg` - all features used

**Invalid configs:**
- `testdata/invalid-type.klg` - bad layout type
- `testdata/invalid-positions.klg` - wrong position count
- `testdata/invalid-charset.klg` - duplicate chars
- `testdata/invalid-groups.klg` - missing setN
- `testdata/invalid-counts.klg` - char count mismatch

File Organization
-----------------

### New Files

**internal/keycraft/generator.go** (~350 lines)
- `GenerationConfig`, `PositionSpec`, `TemplateType` types
- `ParseConfigFile(path)` - parses `.klg` files with comment support
- `ValidateConfig(config)` - validates all constraints
- `GenerateFromConfig(config, seed)` - generates layout
- Helper functions for parsing and allocation
- Uses `NewLockedRNG()` from random.go for thread-safe random number generation

**internal/keycraft/generator_test.go** (~600 lines)
- Parser tests (valid and invalid configs)
- Validation tests (all error conditions)
- Generation tests (correctness and reproducibility)
- Edge case tests (special characters, overlapping sets, etc.)
- Benchmark tests

**cmd/keycraft/generate.go** (~300 lines)
- CLI command definition (replaces old generate command)
- Command name: `generate`, alias: `g`
- Flag definitions
- `generateAction()` orchestrates workflow
- `buildConfigInput()` parses flags
- Integration with optimization (reuses patterns from archived generate.go)
- Helper for default pin computation

### Modified Files

**internal/keycraft/layout.go**
- Add `generateLayoutName()` method to `SplitLayout` (reused from archived generator.go, with `_` prefix added)

**cmd/keycraft/commands_test.go**
- Replace old generate tests with new generate command tests
- Test argument validation
- Test flag combinations
- Test file I/O

### Files to Archive

The following files are renamed to `*.bak` (no longer used):
- `cmd/keycraft/generate.go` → `cmd/keycraft/generate.go.bak`
- `internal/keycraft/generator.go` → `internal/keycraft/generator.go.bak`
- `internal/keycraft/generator_test.go` → `internal/keycraft/generator_test.go.bak`

**Patterns to extract before archiving:**
- `getRNG(seed)` - random number generator initialization
- `generateLayoutName()` - layout naming (add to `SplitLayout` with `_` prefix)
- `HomeThumbChars()` - already exists as `SplitLayout` method
- Flag handling patterns from old generate.go
- Optimization integration workflow

Implementation Files Involved
-----------------------------

### Critical Files for Reference

**Patterns to extract from archived files before archiving:**
- [generator.go](internal/keycraft/generator.go) → `generateLayoutName()` (add to SplitLayout with `_` prefix)
- [generate.go](cmd/keycraft/generate.go) → flag handling patterns, optimization integration workflow

**Existing patterns to reuse:**
- [random.go](internal/keycraft/random.go) - `NewLockedRNG()`, `ShuffleSlice()` for thread-safe random generation using rand/v2
- [layout.go](internal/keycraft/layout.go) - `NewSplitLayout()` (now builds RuneInfo internally), `SaveToFile()`, `HomeThumbChars()`
- [bls_utils.go](internal/keycraft/bls_utils.go) - `LoadPinsFromParams()`
- [optimize.go](internal/keycraft/optimize.go) - `OptimizeLayout()`

**Layout structure:**
- 42 positions: rows 0-2 (12 each), row 3 (6 thumbs)
- Positions 13-16, 19-22: home row (8 keys)
- Positions 36-41: thumbs (6 keys)
- Position 39: typically space

User Manual
-----------

### Creating a Generation Config File

Generation config files use the `.klg` extension (Keyboard Layout Generation).

#### File Format

```
<layout-type>
<row 0: 12 positions>
<row 1: 12 positions>
<row 2: 12 positions>
<row 3: 6 positions (thumbs)>

charset=<all characters to allocate>
set1=<characters for group 1>
set2=<characters for group 2>
...
```

#### Position Markers

- `~` = Excluded (don't allocate any character)
- `_` = Space character (fixed)
- Any character (letters, punctuation, special chars) = Fixed character at this position
- `?` = Random allocation from remaining charset
- `1`-`9` = Allocate from corresponding set (set1, set2, etc.)

**Examples of fixed characters:**
- Letters: `e`, `a`, `t`, `n`, etc.
- Punctuation: `'`, `,`, `.`, `;`, etc.
- Special: Any valid character from the charset

#### Example: Vowels on Right

```
colstag
~ ? ? ? ? ?   ? ? ? ? ? ~
~ 2 2 2 2 ?   ? e a i o ~
~ ? ? ? ? ?   ? ? ? ? ? ~
      ~ ~ 1   _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz,./;'
set1=rl
set2=tnshrdlc
```

This config:
- Uses column-staggered layout type
- Fixes vowels e, a, i, o on right home row
- Allocates r or l to one thumb position (group 1)
- Allocates 4 characters from set2 to left home row (group 2)
- Randomly distributes remaining characters to ? positions
- Excludes some positions (~) from allocation

#### Rules

1. Exactly 42 positions (3 rows of 12, 1 row of 6)
2. charset must include all characters you want to allocate
3. Each group (1-9) must have a corresponding setN definition
4. No character can appear multiple times (duplicates error)
5. Total positions to fill must equal charset length
6. Each setN must have enough characters for its group

### Generating Layouts

#### Basic Generation

```bash
# Generate all permutations of group characters
keycraft generate my-config.klg

# Using alias
keycraft g my-config.klg

# Generate only first 10 permutations
keycraft generate my-config.klg --max-layouts 10

# Reproducible generation (same seed for random positions)
keycraft generate my-config.klg --seed 42
```

Generated layouts are saved to `data/layouts/` with auto-generated names (using permutation index).

#### With Optimization

```bash
# Generate all permutations and optimize (default: pins all non-? positions, deletes unoptimized)
keycraft generate my-config.klg --optimize

# Keep both original and optimized layouts
keycraft generate my-config.klg --optimize --keep-unoptimized

# Customize optimization iterations
keycraft generate my-config.klg --optimize --generations 2000

# Override pinning (advanced)
keycraft generate my-config.klg --optimize --pins "eaio"

# Generate subset with optimization
keycraft generate my-config.klg --max-layouts 5 --optimize
```

**Default pinning behavior:**
When `--optimize` is used, all fixed characters and group characters are pinned. Only positions marked with `?` are free to move during optimization. This preserves your config structure while optimizing the random positions.

Use `--pins` to completely override this behavior if needed.

**Cleanup behavior:**
By default, unoptimized layouts are deleted after optimization to avoid clutter when generating many permutations. Use `--keep-unoptimized` to preserve both versions.