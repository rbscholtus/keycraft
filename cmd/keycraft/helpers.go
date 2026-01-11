package main

import (
	"fmt"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v3"
)

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

// loadCorpusFromFlags loads the corpus specified by the --corpus flag,
// considering the --coverage flag if set.
func loadCorpusFromFlags(c *cli.Command) (*kc.Corpus, error) {
	return loadCorpus(c.String("corpus"), c.IsSet("coverage"), c.Float64("coverage"))
}

// loadLayout loads a layout from layoutDir, automatically appending .klf if needed.
func loadLayout(filename string) (*kc.SplitLayout, error) {
	if filename == "" {
		return nil, fmt.Errorf("layout is required")
	}

	layoutName := ensureNoKlf(filename)
	filename = ensureKlf(filename)
	path := filepath.Join(layoutDir, filename)

	return kc.NewLayoutFromFile(layoutName, path)
}

// ensureKlf appends .klf extension if not present (case-insensitive check).
func ensureKlf(name string) string {
	if strings.ToLower(filepath.Ext(name)) != ".klf" {
		return name + ".klf"
	}
	return name
}

// ensureNoKlf removes .klf extension if present (case-insensitive check).
func ensureNoKlf(name string) string {
	if strings.ToLower(filepath.Ext(name)) == ".klf" {
		return name[:len(name)-4]
	}
	return name
}

// loadTargetLoadsFromFlags loads TargetLoads from flags and config file.
// Command-line flags override config file values.
func loadTargetLoadsFromFlags(c *cli.Command) (*kc.TargetLoads, error) {
	// Try to load from config file first
	configFile := c.String("load-targets-file")
	if configFile == "" {
		configFile = "load_targets.txt"
	}
	configPath := filepath.Join(configDir, configFile)

	targets, err := kc.NewTargetLoadsFromFile(configPath)
	if err != nil {
		// If config file doesn't exist, use defaults
		targets = kc.NewTargetLoads()
	}

	// Apply CLI flag overrides
	if c.IsSet("target-hand-load") {
		if err := targets.SetHandLoad(c.String("target-hand-load")); err != nil {
			return nil, err
		}
	}

	if c.IsSet("target-finger-load") {
		if err := targets.SetFingerLoad(c.String("target-finger-load")); err != nil {
			return nil, err
		}
	}

	if c.IsSet("target-row-load") {
		if err := targets.SetRowLoad(c.String("target-row-load")); err != nil {
			return nil, err
		}
	}

	if c.IsSet("pinky-penalties") {
		if err := targets.SetPinkyPenalties(c.String("pinky-penalties")); err != nil {
			return nil, err
		}
	}

	return targets, nil
}
