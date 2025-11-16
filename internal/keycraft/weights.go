package keycraft

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Weights holds metric weights used for scoring layouts.
// Metrics not explicitly set in the input string default to predefined values.
// ALT, ROL, and ONE have default negative weights since they represent positive aspects.
type Weights struct {
	weights map[string]float64
}

// DefaultMetrics contains built-in metric weights used as defaults when no custom weight is provided.
var DefaultMetrics = map[string]float64{
	"SFB": -1.0,
}

// NewWeights creates an empty Weights structure ready to be populated.
func NewWeights() *Weights {
	weights := make(map[string]float64)
	// maps.Copy(weights, DefaultMetrics)
	return &Weights{weights}
}

// NewWeightsFromString parses a comma-separated `metric=weight` string into a Weights instance.
// Returns an error if the format is invalid or weights cannot be parsed.
func NewWeightsFromString(weightsStr string) (*Weights, error) {
	w := Weights{}
	err := w.AddWeightsFromString(weightsStr)
	return &w, err
}

// NewWeightsFromParams constructs weights from an optional file and CLI string.
func NewWeightsFromParams(path, weightsStr string) (*Weights, error) {
	weights := NewWeights()

	// Load weights from a file if specified.
	if path != "" {
		if err := weights.AddWeightsFromFile(path); err != nil {
			return nil, err
		}
	}

	// Override or add weights from the --weights string flag.
	if err := weights.AddWeightsFromString(weightsStr); err != nil {
		return nil, fmt.Errorf("failed to parse weights: %v", err)
	}

	return weights, nil
}

// AddWeightsFromFile reads weights from a file (ignoring comments/blanks) and applies them to the receiver.
func (w *Weights) AddWeightsFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read weights file %q: %v", path, err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") && line != "" {
			if err := w.AddWeightsFromString(line); err != nil {
				return fmt.Errorf("failed to parse weights from file %q: %v", path, err)
			}
		}
	}
	return nil
}

// AddWeightsFromString parses and applies a comma-separated `metric=weight` string.
// If weightsStr is empty, returns the existing Weights unchanged.
func (w *Weights) AddWeightsFromString(weightsStr string) error {
	if weightsStr == "" {
		return nil
	}

	weightsStr = strings.ToUpper(strings.TrimSpace(weightsStr))
	for pair := range strings.SplitSeq(weightsStr, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid weights format: %s", pair)
		}
		metric := strings.TrimSpace(parts[0])
		weight, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return fmt.Errorf("invalid weight value for metric %s", metric)
		}
		w.weights[metric] = weight
	}

	return nil
}

// Get returns the weight for a metric or 0 if not present.
func (w *Weights) Get(metric string) float64 {
	if val, ok := w.weights[metric]; ok {
		return val
	}
	return 0.0
}
