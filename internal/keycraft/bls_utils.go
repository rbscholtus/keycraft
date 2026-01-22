package keycraft

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"
)

// PinnedKeys represents which of the 42 keys are pinned (fixed) during optimization.
// A value of true means the key cannot be moved, false means it's free to optimize.
type PinnedKeys [42]bool

// OptimizeLayoutBLS is the main entry point for BLS optimization.
// It handles all setup and runs the optimization algorithm.
//
// Parameters:
//   - layout: Initial layout to optimize (can be random or a seed layout)
//   - layoutsDir: Directory containing reference layouts for scorer initialization
//   - corpus: Text corpus for evaluation
//   - weights: Metric weights for scoring
//   - rowBal: Target row load distribution (use nil for defaults)
//   - fingerBal: Target finger load distribution (use nil for defaults)
//   - pinned: Which keys are pinned (fixed) during optimization
//   - maxIterations: Maximum number of iterations (0 = use default)
//   - maxTimeMinutes: Maximum time in minutes (0 = use default)
//   - seed: Random seed for reproducibility (0 = use current time)
//   - consoleWriter: Where to write human-readable progress (use os.Stdout or nil)
//   - logFileWriter: Where to write JSONL structured logs (use nil to disable)
//
// Returns the optimized layout.
func OptimizeLayoutBLS(
	layout *SplitLayout,
	layoutsDir string,
	corpus *Corpus,
	weights *Weights,
	targets *TargetLoads,
	pinned *PinnedKeys,
	maxIterations int,
	maxTimeMinutes int,
	seed int64,
	consoleWriter io.Writer,
	logFileWriter io.Writer,
) (*SplitLayout, error) {
	// Count free keys
	numFree := 0
	for _, isPinned := range pinned {
		if !isPinned {
			numFree++
		}
	}

	if numFree == 0 {
		return nil, fmt.Errorf("no free keys to optimize")
	}

	// Create parameters with defaults, then override from arguments
	params := DefaultBLSParams(numFree)
	if maxIterations > 0 {
		params.MaxIterations = maxIterations
	}
	if maxTimeMinutes > 0 {
		params.MaxTime = time.Duration(maxTimeMinutes) * time.Minute
	}
	if seed != 0 {
		params.Seed = seed
	} else {
		params.Seed = time.Now().UnixNano()
	}

	// Create scorer - use provided targets or defaults
	if targets == nil {
		targets = &TargetLoads{
			TargetHandLoad:   DefaultTargetHandLoad(),
			TargetFingerLoad: DefaultTargetFingerLoad(),
			TargetRowLoad:    DefaultTargetRowLoad(),
			PinkyPenalties:   DefaultPinkyPenalties(),
		}
	}
	if targets.TargetHandLoad == nil {
		targets.TargetHandLoad = DefaultTargetHandLoad()
	}
	if targets.TargetFingerLoad == nil {
		targets.TargetFingerLoad = DefaultTargetFingerLoad()
	}
	if targets.TargetRowLoad == nil {
		targets.TargetRowLoad = DefaultTargetRowLoad()
	}
	if targets.PinkyPenalties == nil {
		targets.PinkyPenalties = DefaultPinkyPenalties()
	}

	scorer, err := NewScorer(layoutsDir, corpus, targets, weights)
	if err != nil {
		return nil, fmt.Errorf("could not create scorer: %w", err)
	}

	// Create BLS optimizer
	bls := NewBLS(params, scorer, corpus, pinned)

	// Create logger with dual output
	logger := NewBLSLogger(consoleWriter, logFileWriter)

	// Run optimization
	bestLayout := bls.Optimize(layout, logger)

	// Log scorer statistics if console writer provided
	if consoleWriter != nil {
		scorer.LogStats(consoleWriter)
	}

	// Log cache stats to JSONL if file writer provided
	if logFileWriter != nil {
		hits := uint64(scorer.cacheHits.Load())
		misses := uint64(scorer.cacheMisses.Load())
		scorer.cacheMu.RLock()
		uniqueKeys := len(scorer.scoreCache)
		scorer.cacheMu.RUnlock()
		memoryBytes := int64(uniqueKeys) * 8 // Rough estimate: 8 bytes per float64
		logger.LogCacheStats(hits, misses, uniqueKeys, memoryBytes)
	}

	return bestLayout, nil
}

// LoadPins loads a pins file specifying which keys should be fixed during optimization.
// The file format mirrors a layout file, but uses symbols to indicate pin status:
//   - '.', '_', '-' : unpinned (key can be moved)
//   - '*', 'x', 'X' : pinned (key is fixed)
//
// The file must contain exactly 4 rows with 12, 12, 12, and 6 keys respectively
// (matching the 42-key split layout structure). Empty lines and lines starting with
// '#' are ignored.
//
// Returns an error if the file cannot be opened, has invalid format, or contains
// invalid characters.
func LoadPins(path string) (*PinnedKeys, error) {
	file, err := os.Open(path)
	if err != nil {
		// This handles both "does not exist" and other errors
		return nil, fmt.Errorf("could not open pins file %s: %w", path, err)
	}
	defer CloseFile(file)

	pinned := PinnedKeys{}
	scanner := bufio.NewScanner(file)
	index := 0
	expectedKeys := []int{12, 12, 12, 6}

	// Parse pins from file
	for rowIdx, expectedKeyCount := range expectedKeys {
		// Skip empty lines and comments
		var line string
		for scanner.Scan() {
			line = strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				break
			}
		}
		if line == "" {
			return nil, fmt.Errorf("invalid file format in %s: not enough rows (expected 4, got %d)", path, rowIdx)
		}

		keys := strings.Fields(line)
		if len(keys) != expectedKeyCount {
			return nil, fmt.Errorf("invalid file format in %s: row %d has %d keys, expected %d",
				path, rowIdx+1, len(keys), expectedKeyCount)
		}

		for colIdx, key := range keys {
			if len(key) != 1 {
				return nil, fmt.Errorf("invalid file format in %s: key '%s' at row %d col %d must be exactly 1 character",
					path, key, rowIdx+1, colIdx+1)
			}
			switch rune(key[0]) {
			case '.', '_', '-':
				pinned[index] = false
			case '*', 'x', 'X':
				pinned[index] = true
			default:
				return nil, fmt.Errorf("invalid character '%c' in %s at row %d col %d (use . _ - for unpinned, * x X for pinned)",
					key[0], path, rowIdx+1, colIdx+1)
			}
			index++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading pins file %s: %w", path, err)
	}

	// Validation: ensure we processed exactly 42 keys
	if index != 42 {
		return nil, fmt.Errorf("invalid file format in %s: expected 42 keys total, got %d", path, index)
	}

	return &pinned, nil
}

// LoadPinsFromParams configures which keys are pinned during optimization.
// Three modes:
//  1. Load from pins file (path) - uses pin file format
//  2. Pin specific characters (pins) - comma-separated characters to fix
//  3. Free specific characters (free) - all others are pinned
//
// Modes 1 and 2 can be combined, but mode 3 (free) is mutually exclusive.
// If no options are provided, only empty keys and spaces are pinned by default.
//
// Returns an error if:
//   - Both free and (path or pins) are specified
//   - A character to pin/free is not in the layout
//   - The pins file cannot be loaded
//   - The layout is nil
func LoadPinsFromParams(path, pins, free string, sl *SplitLayout) (*PinnedKeys, error) {
	if sl == nil {
		return nil, fmt.Errorf("layout cannot be nil")
	}

	if (path != "" || pins != "") && free != "" {
		return nil, fmt.Errorf("cannot use both --free and --pins/--pins-file options together")
	}

	// Mode 3: Free specific characters (all others pinned)
	if free != "" {
		pinned := &PinnedKeys{}

		// Pin everything by default
		for i := range pinned {
			pinned[i] = true
		}

		// Unpin characters in free string
		for _, r := range free {
			key, ok := sl.RuneInfo[r]
			if !ok {
				return nil, fmt.Errorf("cannot free unavailable character: %q (%U)", r, r)
			}
			pinned[key.Index] = false
		}
		return pinned, nil
	}

	var pinned *PinnedKeys

	// Mode 1: Load from file
	if path != "" {
		loadedPins, err := LoadPins(path)
		if err != nil {
			return nil, fmt.Errorf("could not load pins from file: %w", err)
		}
		pinned = loadedPins
	} else {
		// Default: pin empty keys and spaces
		pinned = &PinnedKeys{}
		for i, r := range sl.Runes {
			if r == 0 || unicode.IsSpace(r) {
				pinned[i] = true
			}
		}
	}

	// Mode 2: Pin additional characters from pins string
	if pins != "" {
		for _, r := range pins {
			key, ok := sl.RuneInfo[r]
			if !ok {
				return nil, fmt.Errorf("cannot pin unavailable character: %q (%U)", r, r)
			}
			pinned[key.Index] = true
		}
	}

	return pinned, nil
}
