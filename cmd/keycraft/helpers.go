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
// considering the --coverage-threshold flag if set.
func getCorpusFromFlags(c *cli.Context) (*kc.Corpus, error) {
	return loadCorpus(c.String("corpus"), c.IsSet("coverage-threshold"), c.Float64("coverage-threshold"))
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
// forceReload bypasses cache, and coverageThreshold filters low-frequency words.
func loadCorpus(filename string, forceReload bool, coverageThreshold float64) (*kc.Corpus, error) {
	if filename == "" {
		return nil, fmt.Errorf("corpus file is required")
	}
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	return kc.NewCorpusFromFile(corpusName, path, forceReload, coverageThreshold)
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
