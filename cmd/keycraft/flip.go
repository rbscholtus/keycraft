package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
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
	if !strings.EqualFold(filepath.Ext(layoutArg), ".klf") {
		layoutArg = layoutArg + ".klf"
	}
	layoutPath := filepath.Join(layoutDir, layoutArg)

	res, err := kc.FlipLayout(kc.FlipInput{LayoutFile: layoutPath})
	if err != nil {
		return fmt.Errorf("could not flip layout: %w", err)
	}

	outputPath := filepath.Join(layoutDir, res.Layout.Name+".klf")
	if err := res.Layout.SaveToFile(outputPath); err != nil {
		return fmt.Errorf("could not save flipped layout: %w", err)
	}

	fmt.Printf("Flipped layout and saved to: %s.klf\n", res.Layout.Name)
	return nil
}
