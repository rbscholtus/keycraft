package main

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// flipCommand defines the CLI command for flipping a layout horizontally.
var flipCommand = &cli.Command{
	Name:      "flip",
	Aliases:   []string{"f"},
	Usage:     "Flip a keyboard layout horizontally and save as new layout",
	ArgsUsage: "<layout>",
	Before:    validateFlipFlags,
	Action:    flipAction,
}

// validateFlipFlags validates CLI flags before running the flip command.
func validateFlipFlags(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("exactly 1 layout required")
	}
	return nil
}

// flipAction loads a layout, flips it horizontally, and saves it with "-flipped" suffix.
func flipAction(c *cli.Context) error {
	layoutArg := c.Args().First()

	// Load the layout using helper function
	layout, err := loadLayout(layoutArg)
	if err != nil {
		return fmt.Errorf("failed to load layout: %w", err)
	}

	// Flip the layout horizontally
	layout.FlipHorizontal()

	// Update the name with "-flipped" suffix
	layout.Name = layout.Name + "-flipped"

	// Save to new file
	outputPath := filepath.Join(layoutDir, layout.Name+".klf")

	if err := layout.SaveToFile(outputPath); err != nil {
		return fmt.Errorf("failed to save flipped layout: %w", err)
	}

	fmt.Printf("Flipped layout and saved to: %s.klf\n", layout.Name)
	return nil
}
