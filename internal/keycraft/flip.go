package keycraft

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FlipInput specifies a layout to flip horizontally.
type FlipInput struct {
	// LayoutFile is the full filepath to a .klf layout.
	LayoutFile string
	// NewName overrides the flipped layout's name. When empty, the original
	// name with a "-flipped" suffix is used.
	NewName string
}

// FlipResult contains the flipped layout.
type FlipResult struct {
	Layout *SplitLayout
}

// FlipLayout loads a layout from disk, flips it horizontally, and returns the
// result in memory. The caller decides whether to persist it.
func FlipLayout(input FlipInput) (*FlipResult, error) {
	if input.LayoutFile == "" {
		return nil, fmt.Errorf("layout file is required")
	}

	name := strings.TrimSuffix(filepath.Base(input.LayoutFile), filepath.Ext(input.LayoutFile))
	layout, err := NewLayoutFromFile(name, input.LayoutFile)
	if err != nil {
		return nil, fmt.Errorf("could not load layout: %w", err)
	}

	layout.FlipHorizontal()

	if input.NewName != "" {
		layout.Name = input.NewName
	} else {
		layout.Name = layout.Name + "-flipped"
	}

	return &FlipResult{Layout: layout}, nil
}
