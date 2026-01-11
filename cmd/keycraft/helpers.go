package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v3"
)

// getCorpusFromFlags loads the corpus specified by the --corpus flag,
// considering the --coverage flag if set.
func getCorpusFromFlags(c *cli.Command) (*kc.Corpus, error) {
	return loadCorpus(c.String("corpus"), c.IsSet("coverage"), c.Float64("coverage"))
}

// getTargetHandLoadFromFlag parses and scales the --target-hand-load flag into percentages.
// Accepts 2 values for left hand and right hand.
// Values are validated and scaled to sum to 100.
func getTargetHandLoadFromFlag(c *cli.Command) (*[2]float64, error) {
	hbStr := c.String("target-hand-load")
	vals, err := parseTargetHandLoad(hbStr)
	if err != nil {
		return nil, err
	}
	if err := scaleTargetHandLoad(vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// getTargetFingerLoadFromFlag parses and scales the --target-target-finger-load flag into percentages.
// Accepts 4 values (mirrored for both hands) or 8 values (F0-F3, F6-F9).
// Thumbs (F4, F5) are always set to 0. Values are validated and scaled to sum to 100.
func getTargetFingerLoadFromFlag(c *cli.Command) (*[10]float64, error) {
	fbStr := c.String("target-target-finger-load")
	vals, err := parseFingerLoad(fbStr)
	if err != nil {
		return nil, err
	}
	if err := scaleFingerLoad(vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// getTargetRowLoadFromFlag parses and scales the --target-target-row-load flag into percentages.
// Accepts 3 values for top row, home row, and bottom row.
// Values are validated and scaled to sum to 100.0.
func getTargetRowLoadFromFlag(c *cli.Command) (*[3]float64, error) {
	rlStr := c.String("target-target-row-load")
	vals, err := parseRowLoad(rlStr)
	if err != nil {
		return nil, err
	}
	if err := scaleRowLoad(vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// getPinkyPenaltiesFromFlag parses the --pinky-penalties flag into penalty weights.
// Accepts 6 values (mirrored for both hands) or 12 values (left then right).
// Values are not scaled (used as-is for penalty calculations).
func getPinkyPenaltiesFromFlag(c *cli.Command) (*[12]float64, error) {
	ppStr := c.String("pinky-penalties")
	return parsePinkyPenalties(ppStr)
}

// loadWeightsFromFlags loads weights from the --weights-file and --weights flags.
// Weights specified via --weights take precedence over file-based weights.
func loadWeightsFromFlags(c *cli.Command) (*kc.Weights, error) {
	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(configDir, weightsPath)
	}
	return kc.NewWeightsFromParams(weightsPath, c.String("weights"))
}

// loadCorpus loads a corpus from corpusDir.
// forceReload bypasses cache, and coverage filters low-frequency words.
func loadCorpus(filename string, forceReload bool, coverage float64) (*kc.Corpus, error) {
	if filename == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	return kc.NewCorpusFromFile(corpusName, path, forceReload, coverage)
}

// loadLayout loads a layout from layoutDir, automatically appending .klf if needed.
func loadLayout(filename string) (*kc.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout is required")
	}

	var layoutName string
	ext := filepath.Ext(filename)
	if strings.ToLower(ext) != ".klf" {
		layoutName = filename
		filename += ".klf"
	} else {
		layoutName = strings.TrimSuffix(filename, ext)
	}
	path := filepath.Join(layoutDir, filename)

	return kc.NewLayoutFromFile(layoutName, path)
}

// parseTargetHandLoad parses hand load values from a comma-separated string.
// Expects exactly 2 values for left hand and right hand.
func parseTargetHandLoad(s string) (*[2]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("target-hand-load must have exactly 2 comma-separated values (got %d)", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var handVals [2]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in target-hand-load at position %d", i)
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in target-hand-load at position %d: %v", i, err)
		}
		handVals[i] = v
	}

	return &handVals, nil
}

// scaleTargetHandLoad scales hand load values in-place so their sum equals 100.0.
// Validates all values are non-negative and sum is above epsilon threshold.
// Returns an error if any value is negative or if the sum is too small to scale safely.
func scaleTargetHandLoad(vals *[2]float64) error {
	var sum float64
	for _, v := range vals {
		if v < 0.0 {
			return fmt.Errorf("cannot scale hand load: negative value %f", v)
		}
		sum += v
	}
	const epsilon = 1e-9
	if sum < epsilon {
		return fmt.Errorf("cannot scale hand load: sum is zero or too small")
	}
	scale := 100.0 / sum
	for i := range vals {
		vals[i] *= scale
	}
	return nil
}

// parseFingerLoad parses finger load values from a comma-separated string.
// Accepts 4 values (mirrored to 8) or 8 values directly for F0-F3,F6-F9.
// Thumbs (F4, F5) are set to 0.0.
func parseFingerLoad(s string) (*[10]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 && len(parts) != 8 {
		return nil, fmt.Errorf("target-finger-load must have 4 or 8 comma-separated values")
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// If the user provided 4 values, mirror them to create the 8-value representation.
	// We append reversed order so after inserting thumb zeros the indices map to F0..F9.
	if len(parts) == 4 {
		parts = append(parts, parts[3], parts[2], parts[1], parts[0])
	}
	parts = slices.Insert(parts, 4, "0.0", "0.0")

	// convert values to float64
	var fingerVals [10]float64
	for i, p := range parts {
		var v float64
		if p == "" {
			return nil, fmt.Errorf("empty value in target-finger-load")
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in target-finger-load: %v", err)
		}
		fingerVals[i] = v
	}

	return &fingerVals, nil
}

// scaleFingerLoad scales finger load values in-place so their sum equals 100.0.
// Validates all values are non-negative and sum is above epsilon threshold.
// Returns an error if any value is negative or if the sum is too small to scale safely.
func scaleFingerLoad(vals *[10]float64) error {
	var sum float64
	for _, v := range vals {
		if v < 0.0 {
			return fmt.Errorf("cannot scale finger load: negative value %f", v)
		}
		sum += v
	}
	const epsilon = 1e-9
	if sum < epsilon {
		return fmt.Errorf("cannot scale finger load: sum is zero or too small")
	}
	scale := 100.0 / sum
	for i := range vals {
		vals[i] *= scale
	}
	return nil
}

// parseRowLoad parses row load values from a comma-separated string.
// Expects exactly 3 values for top row, home row, and bottom row.
func parseRowLoad(s string) (*[3]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("target-row-load must have exactly 3 comma-separated values (got %d)", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var rowVals [3]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in target-row-load at position %d", i)
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in target-row-load at position %d: %v", i, err)
		}
		rowVals[i] = v
	}

	return &rowVals, nil
}

// scaleRowLoad scales row load values in-place so their sum equals 100.0.
// Validates all values are non-negative and sum is above epsilon threshold.
// Returns an error if any value is negative or if the sum is too small to scale safely.
func scaleRowLoad(vals *[3]float64) error {
	var sum float64
	for _, v := range vals {
		if v < 0.0 {
			return fmt.Errorf("cannot scale row load: negative value %f", v)
		}
		sum += v
	}
	const epsilon = 1e-9
	if sum < epsilon {
		return fmt.Errorf("cannot scale row load: sum is zero or too small")
	}
	scale := 100.0 / sum
	for i := range vals {
		vals[i] *= scale
	}
	return nil
}

// parsePinkyPenalties parses pinky penalty values from a comma-separated string.
// Accepts 6 values (mirrored to 12) or 12 values directly.
// Order per hand: top-outer, top-inner, home-outer, home-inner, bottom-outer, bottom-inner.
func parsePinkyPenalties(s string) (*[12]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 6 && len(parts) != 12 {
		return nil, fmt.Errorf("pinky-penalties must have 6 or 12 comma-separated values (got %d)", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// If the user provided 6 values, mirror them to create the 12-value representation.
	if len(parts) == 6 {
		parts = append(parts, parts...)
	}

	// convert values to float64
	var pinkyVals [12]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in pinky-penalties at position %d", i)
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float in pinky-penalties at position %d: %v", i, err)
		}
		pinkyVals[i] = v
	}

	return &pinkyVals, nil
}

// ensureKlf appends .klf extension if not present (case-insensitive check).
func ensureKlf(name string) string {
	if !strings.HasSuffix(strings.ToLower(name), ".klf") {
		return name + ".klf"
	}
	return name
}

// ensureNoKlf removes .klf extension if present (case-insensitive check).
func ensureNoKlf(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".klf") {
		return name[:len(name)-4]
	}
	return name
}

// loadTargetLoadsFromFlags loads TargetLoads from flags and config file.
// Command-line flags override config file values.
func loadTargetLoadsFromFlags(c *cli.Command) (*kc.TargetLoads, error) {
	// Try to load from config file first
	// Use --load-dist-targets flag if provided, otherwise default to load_targets.txt
	configFile := c.String("load-targets-file")
	if configFile == "" {
		configFile = "load_targets.txt"
	}
	configPath := filepath.Join(configDir, configFile)
	targets, err := loadTargetLoadsFromFile(configPath)
	if err != nil {
		// If config file doesn't exist, use hardcoded defaults
		targets = &kc.TargetLoads{}
	}

	if c.IsSet("target-hand-load") {
		handLoad, err := getTargetHandLoadFromFlag(c)
		if err != nil {
			return nil, err
		}
		targets.TargetHandLoad = handLoad
	} else if targets.TargetHandLoad == nil {
		// Use hardcoded default if not in config and not in flag
		handLoad, _ := parseTargetHandLoad("50,50")
		_ = scaleTargetHandLoad(handLoad)
		targets.TargetHandLoad = handLoad
	}

	if c.IsSet("target-finger-load") {
		fingerLoad, err := getTargetFingerLoadFromFlag(c)
		if err != nil {
			return nil, err
		}
		targets.TargetFingerLoad = fingerLoad
	} else if targets.TargetFingerLoad == nil {
		// Use hardcoded default if not in config and not in flag
		fingerLoad, _ := parseFingerLoad("7,10,16,17")
		_ = scaleFingerLoad(fingerLoad)
		targets.TargetFingerLoad = fingerLoad
	}

	// Override with flags if they are set
	if c.IsSet("target-row-load") {
		rowLoad, err := getTargetRowLoadFromFlag(c)
		if err != nil {
			return nil, err
		}
		targets.TargetRowLoad = rowLoad
	} else if targets.TargetRowLoad == nil {
		// Use hardcoded default if not in config and not in flag
		rowLoad, _ := parseRowLoad("17.5,75.0,7.5")
		_ = scaleRowLoad(rowLoad)
		targets.TargetRowLoad = rowLoad
	}

	if c.IsSet("pinky-penalties") {
		pinkyPenalties, err := getPinkyPenaltiesFromFlag(c)
		if err != nil {
			return nil, err
		}
		targets.PinkyPenalties = pinkyPenalties
	} else if targets.PinkyPenalties == nil {
		// Use hardcoded default if not in config and not in flag
		pinkyPenalties, _ := parsePinkyPenalties("2.0,1.5,1.0,0.0,2.0,1.5")
		targets.PinkyPenalties = pinkyPenalties
	}

	return targets, nil
}

// loadTargetLoadsFromFile loads defaults from config/load_targets.txt.
// Returns empty TargetLoads if file doesn't exist or has errors.
func loadTargetLoadsFromFile(filePath string) (*kc.TargetLoads, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	targets := &kc.TargetLoads{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key: value pairs
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "target-hand-load":
			handLoad, err := parseTargetHandLoad(value)
			if err != nil {
				return nil, fmt.Errorf("invalid target-hand-load in config file: %v", err)
			}
			if err := scaleTargetHandLoad(handLoad); err != nil {
				return nil, fmt.Errorf("failed to scale target-hand-load in config file: %v", err)
			}
			targets.TargetHandLoad = handLoad
		case "target-finger-load":
			fingerLoad, err := parseFingerLoad(value)
			if err != nil {
				return nil, fmt.Errorf("invalid target-finger-load in config file: %v", err)
			}
			if err := scaleFingerLoad(fingerLoad); err != nil {
				return nil, fmt.Errorf("failed to scale target-finger-load in config file: %v", err)
			}
			targets.TargetFingerLoad = fingerLoad
		case "target-row-load":
			rowLoad, err := parseRowLoad(value)
			if err != nil {
				return nil, fmt.Errorf("invalid target-row-load in config file: %v", err)
			}
			if err := scaleRowLoad(rowLoad); err != nil {
				return nil, fmt.Errorf("failed to scale target-row-load in config file: %v", err)
			}
			targets.TargetRowLoad = rowLoad
		case "pinky-penalties":
			pinkyPenalties, err := parsePinkyPenalties(value)
			if err != nil {
				return nil, fmt.Errorf("invalid pinky-penalties in config file: %v", err)
			}
			targets.PinkyPenalties = pinkyPenalties
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	return targets, nil
}
