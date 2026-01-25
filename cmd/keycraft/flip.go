package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

// flipCommand defines the CLI command for flipping a layout horizontally.
var flipCommand = &cli.Command{
	Name:          "flip",
	Aliases:       []string{"f"},
	Usage:         "Flip a keyboard layout horizontally and save as new layout",
	ArgsUsage:     "<layout>",
	Action:        flipAction,
	ShellComplete: layoutShellComplete,
}

// flipAction loads a keyboard layout, performs a horizontal mirror transformation,
// and saves the resulting layout to a new file with a "-flipped" suffix.
func flipAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	if c.NArg() != 1 {
		return fmt.Errorf("expected exactly 1 layout, got %d", c.Args().Len())
	}

	layoutArg := c.Args().First()

	// Load the layout using helper function
	layout, err := loadLayout(layoutArg)
	if err != nil {
		return fmt.Errorf("could not load layout: %w", err)
	}

	// Flip the layout horizontally
	layout.FlipHorizontal()

	// Update the name with "-flipped" suffix
	layout.Name = layout.Name + "-flipped"

	// Save to new file
	outputPath := filepath.Join(layoutDir, layout.Name+".klf")

	if err := layout.SaveToFile(outputPath); err != nil {
		return fmt.Errorf("could not save flipped layout: %w", err)
	}

	fmt.Printf("Flipped layout and saved to: %s.klf\n", layout.Name)
	return nil
}
