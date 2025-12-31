package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

// getCorpusFromFlags loads the corpus specified by the --corpus flag,
// considering the --coverage flag if set.
func getCorpusFromFlags(c *cli.Context) (*kc.Corpus, error) {
	return loadCorpus(c.String("corpus"), c.IsSet("coverage"), c.Float64("coverage"))
}

// getFingerLoadFromFlag parses and scales the --finger-load flag into percentages.
// Accepts 4 values (mirrored for both hands) or 8 values (F0-F3, F6-F9).
// Thumbs (F4, F5) are always set to 0. Values are validated and scaled to sum to 100.
func getFingerLoadFromFlag(c *cli.Context) (*[10]float64, error) {
	fbStr := c.String("finger-load")
	vals, err := parseFingerLoad(fbStr)
	if err != nil {
		return nil, err
	}
	if err := scaleFingerLoad(vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// getRowLoadFromFlag parses and scales the --row-load flag into percentages.
// Accepts 3 values for top row, home row, and bottom row.
// Values are validated and scaled to sum to 100.0.
func getRowLoadFromFlag(c *cli.Context) (*[3]float64, error) {
	rlStr := c.String("row-load")
	vals, err := parseRowLoad(rlStr)
	if err != nil {
		return nil, err
	}
	if err := scaleRowLoad(vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// getPinkyWeightsFromFlag parses the --pinky-weights flag into penalty weights.
// Accepts 6 values (mirrored for both hands) or 12 values (left then right).
// Values are not scaled (used as-is for penalty calculations).
func getPinkyWeightsFromFlag(c *cli.Context) (*[12]float64, error) {
	pwStr := c.String("pinky-weights")
	return parsePinkyWeights(pwStr)
}

// loadWeightsFromFlags loads weights from the --weights-file and --weights flags.
// Weights specified via --weights take precedence over file-based weights.
func loadWeightsFromFlags(c *cli.Context) (*kc.Weights, error) {
	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
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

// parseRowLoad parses row load values from a comma-separated string.
// Expects exactly 3 values for top row, home row, and bottom row.
func parseRowLoad(s string) (*[3]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("row-load must have exactly 3 comma-separated values (got %d)", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	var rowVals [3]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in row-load at position %d", i)
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in row-load at position %d: %v", i, err)
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

// parseFingerLoad parses finger load values from a comma-separated string.
// Accepts 4 values (mirrored to 8) or 8 values directly for F0-F3,F6-F9.
// Thumbs (F4, F5) are set to 0.0.
func parseFingerLoad(s string) (*[10]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 && len(parts) != 8 {
		return nil, fmt.Errorf("finger-load must have 4 or 8 comma-separated values")
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// If the user provided 4 values, mirror them to create the 8-value representation.
	// We append reversed order so after inserting thumb zeros the indices map to F0..F9.
	if len(parts) == 4 {
		for i := len(parts) - 1; i >= 0; i-- {
			parts = append(parts, parts[i])
		}
	}
	parts = slices.Insert(parts, 4, "0.0", "0.0")

	// convert values to float64
	var fingerVals [10]float64
	for i, p := range parts {
		var v float64
		if p == "" {
			return nil, fmt.Errorf("empty value in finger-load")
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in finger-load: %v", err)
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

// parsePinkyWeights parses pinky weight values from a comma-separated string.
// Accepts 6 values (mirrored to 12) or 12 values directly.
// Order per hand: top-outer, top-inner, home-outer, home-inner, bottom-outer, bottom-inner.
func parsePinkyWeights(s string) (*[12]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 6 && len(parts) != 12 {
		return nil, fmt.Errorf("pinky-weights must have 6 or 12 comma-separated values (got %d)", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// If the user provided 6 values, mirror them to create the 12-value representation.
	if len(parts) == 6 {
		for i := 0; i < 6; i++ {
			parts = append(parts, parts[i])
		}
	}

	// convert values to float64
	var pinkyVals [12]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in pinky-weights at position %d", i)
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float in pinky-weights at position %d: %v", i, err)
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
