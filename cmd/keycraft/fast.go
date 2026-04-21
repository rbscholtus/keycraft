package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/rbscholtus/keycraft/internal/tui"
	"github.com/urfave/cli/v3"
)

// fastOptimizeCommand runs a single BLS steepest-descent pass and stops —
// same surface as `optimize` but without perturbation and without the outer
// BLS iteration loop. Finishes in seconds on a small corpus.
var fastOptimizeCommand = &cli.Command{
	Name:          "fast-optimize",
	Aliases:       []string{"fo"},
	Usage:         "Fast optimize: run BLS local search only (no perturbation)",
	Flags:         append(commonFlags(), optFlags("pins-file", "pins", "free", "seed", "log-file")...),
	ArgsUsage:     "<layout>",
	Action:        fastOptimizeAction,
	ShellComplete: layoutShellComplete,
}

func fastOptimizeAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	input, err := buildOptimizeInput(c, nil, false)
	if err != nil {
		return fmt.Errorf("could not parse user input: %w", err)
	}
	input.LocalSearchOnly = true

	// Open log file if requested
	if logFilePath := c.String("log-file"); logFilePath != "" {
		f, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("could not create log file %s: %w", logFilePath, err)
		}
		defer kc.CloseFile(f)
		input.LogFile = f
	}

	optResult, err := kc.OptimizeLayout(ctx, input, os.Stdout)
	if err != nil {
		return fmt.Errorf("could not fast-optimize layout: %w", err)
	}

	origPath := filepath.Join(layoutDir, optResult.OriginalLayout.Name+".klf")
	bestPath := filepath.Join(layoutDir, optResult.BestLayout.Name+".klf")
	if err := optResult.BestLayout.SaveToFile(bestPath); err != nil {
		return fmt.Errorf("could not save best layout to %s: %w", bestPath, err)
	}

	// Display original vs optimised side-by-side
	viewResult, err := kc.ViewLayouts(kc.ViewInput{
		LayoutFiles: []string{origPath, bestPath},
		Corpus:      input.Corpus,
		Targets:     input.Targets,
	})
	if err != nil {
		return fmt.Errorf("could not perform layout analysis: %w", err)
	}
	return tui.RenderView(viewResult)
}

// fastGenerateCommand is the `generate` pipeline without the post-optimise
// pass. It purely emits layouts from a .gen config.
var fastGenerateCommand = &cli.Command{
	Name:      "fast-generate",
	Aliases:   []string{"fg"},
	Usage:     "Fast generate: produce layouts from a config without the optimisation pass",
	ArgsUsage: "<config.gen>",
	Flags:     append(commonFlags(), generationFlags("max-layouts", "seed")...),
	Action:    fastGenerateAction,
}

func fastGenerateAction(ctx context.Context, c *cli.Command) error {
	if isShellCompletion() {
		return nil
	}

	genInput, err := buildGenerateInput(c)
	if err != nil {
		return err
	}
	genInput.Optimize = false
	genInput.KeepUnoptimized = false

	config, err := kc.ParseConfigFile(genInput.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not parse config file: %w", err)
	}
	if err := kc.ValidateConfig(config); err != nil {
		return err
	}

	result, err := kc.GenerateFromConfig(ctx, config, genInput, layoutDir)
	if err != nil {
		return fmt.Errorf("could not generate layouts: %w", err)
	}

	for _, warning := range result.Warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
	fmt.Printf("Generated %d layouts (total permutations: %d)\n", result.Generated, result.TotalPerms)
	return nil
}
