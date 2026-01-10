package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/urfave/cli/v3"
)

// flipCommand defines the CLI command for flipping a layout horizontally.
var flipCommand = &cli.Command{
	Name:          "flip",
	Aliases:       []string{"f"},
	Usage:         "Flip a keyboard layout horizontally and save as new layout",
	ArgsUsage:     "<layout>",
	Before:        validateFlipFlags,
	Action:        flipAction,
	ShellComplete: layoutShellComplete,
}

// validateFlipFlags validates CLI flags before running the flip command.
func validateFlipFlags(ctx context.Context, c *cli.Command) (context.Context, error) {
	// Skip validation during shell completion
	// Check os.Args directly since -- prevents flag parsing
	if slices.Contains(os.Args, "--generate-shell-completion") {
		return ctx, nil
	}

	if c.NArg() != 1 {
		return ctx, fmt.Errorf("expected exactly 1 layout, got %d", c.Args().Len())
	}
	return ctx, nil
}

// flipAction loads a layout, flips it horizontally, and saves it with "-flipped" suffix.
func flipAction(ctx context.Context, c *cli.Command) error {
	// During shell completion, action should not run
	if slices.Contains(os.Args, "--generate-shell-completion") {
		return nil
	}

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
