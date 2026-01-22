package keycraft

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NewTargetLoads creates a TargetLoads instance with default values.
// All fields are initialized using the documented default distributions.
func NewTargetLoads() *TargetLoads {
	return &TargetLoads{
		TargetHandLoad:   DefaultTargetHandLoad(),
		TargetFingerLoad: DefaultTargetFingerLoad(),
		TargetRowLoad:    DefaultTargetRowLoad(),
		PinkyPenalties:   DefaultPinkyPenalties(),
	}
}

// NewTargetLoadsFromFile loads target loads configuration from a file.
// Returns defaults for any fields not present in the file.
func NewTargetLoadsFromFile(filePath string) (*TargetLoads, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	targets := &TargetLoads{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key: value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "target-hand-load":
			if err := targets.SetHandLoad(value); err != nil {
				return nil, fmt.Errorf("invalid target-hand-load in config file: %w", err)
			}
		case "target-finger-load":
			if err := targets.SetFingerLoad(value); err != nil {
				return nil, fmt.Errorf("invalid target-finger-load in config file: %w", err)
			}
		case "target-row-load":
			if err := targets.SetRowLoad(value); err != nil {
				return nil, fmt.Errorf("invalid target-row-load in config file: %w", err)
			}
		case "pinky-penalties":
			if err := targets.SetPinkyPenalties(value); err != nil {
				return nil, fmt.Errorf("invalid pinky-penalties in config file: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Fill in any missing fields with defaults
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

	return targets, nil
}

// SetHandLoad parses and sets the hand load distribution from a string.
// Expects exactly 2 comma-separated values for left and right hands.
// Values are automatically scaled to sum to 100%.
func (tl *TargetLoads) SetHandLoad(s string) error {
	handLoad, err := parseTargetHandLoad(s)
	if err != nil {
		return fmt.Errorf("could not parse target hand load: %w", err)
	}
	if err := scaleTargetHandLoad(handLoad); err != nil {
		return fmt.Errorf("could not scale target hand load: %w", err)
	}
	tl.TargetHandLoad = handLoad
	return nil
}

// SetFingerLoad parses and sets the finger load distribution from a string.
// Accepts 4 values (mirrored for both hands) or 8 values (F0-F3, F6-F9).
// Thumbs (F4, F5) are always set to 0. Values are automatically scaled to sum to 100%.
func (tl *TargetLoads) SetFingerLoad(s string) error {
	fingerLoad, err := parseFingerLoad(s)
	if err != nil {
		return fmt.Errorf("could not parse finger load: %w", err)
	}
	if err := scaleFingerLoad(fingerLoad); err != nil {
		return fmt.Errorf("could not scale finger load: %w", err)
	}
	tl.TargetFingerLoad = fingerLoad
	return nil
}

// SetRowLoad parses and sets the row load distribution from a string.
// Expects exactly 3 comma-separated values for top, home, and bottom rows.
// Values are automatically scaled to sum to 100%.
func (tl *TargetLoads) SetRowLoad(s string) error {
	rowLoad, err := parseRowLoad(s)
	if err != nil {
		return fmt.Errorf("could not parse row load: %w", err)
	}
	if err := scaleRowLoad(rowLoad); err != nil {
		return fmt.Errorf("could not scale row load: %w", err)
	}
	tl.TargetRowLoad = rowLoad
	return nil
}

// SetPinkyPenalties parses and sets the pinky penalty weights from a string.
// Accepts 6 values (mirrored for both hands) or 12 values (left then right).
// Order per hand: top-outer, top-inner, home-outer, home-inner, bottom-outer, bottom-inner.
// Values are NOT scaled (used as-is for penalty calculations).
func (tl *TargetLoads) SetPinkyPenalties(s string) error {
	pinkyPenalties, err := parsePinkyPenalties(s)
	if err != nil {
		return fmt.Errorf("could not parse pinky penalties: %w", err)
	}
	tl.PinkyPenalties = pinkyPenalties
	return nil
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
			return nil, fmt.Errorf("invalid float in target-hand-load at position %d: %w", i, err)
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

	// Insert thumb zeros at positions 4 and 5
	newParts := make([]string, 0, 10)
	newParts = append(newParts, parts[:4]...)
	newParts = append(newParts, "0.0", "0.0")
	newParts = append(newParts, parts[4:]...)
	parts = newParts

	// convert values to float64
	var fingerVals [10]float64
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("empty value in target-finger-load")
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil || v < 0.0 {
			return nil, fmt.Errorf("invalid float in target-finger-load: %w", err)
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
			return nil, fmt.Errorf("invalid float in target-row-load at position %d: %w", i, err)
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
			return nil, fmt.Errorf("invalid float in pinky-penalties at position %d: %w", i, err)
		}
		pinkyVals[i] = v
	}

	return &pinkyVals, nil
}
