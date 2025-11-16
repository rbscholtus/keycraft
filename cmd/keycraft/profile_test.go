package main

import (
	"os"
	"testing"

	"github.com/urfave/cli/v2"
)

// TestProfileOptimize is a test wrapper for profiling the optimize command.
// Run from repo root with: go test -run=TestProfileOptimize -cpuprofile=cpu.prof -memprofile=mem.prof -timeout 10m ./cmd/keycraft
func TestProfileOptimize(t *testing.T) {
	// Ensure we're in the repo root (where data/ directory exists)
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("Failed to change to repo root: %v", err)
	}

	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up command line arguments as if running:
	// keycraft o -g 5 --mt 1 --seed 42 qwerty
	os.Args = []string{
		"keycraft",
		"o",
		"-g", "100", // Change to 100 generations
		"--mt", "1", // Change to 1 minutes
		"--seed", "42",
		"qwerty",
	}

	// Create and run the app
	app := &cli.App{
		Name:  "keycraft",
		Usage: "A CLI tool for crafting better keyboard layouts",
		Commands: []*cli.Command{
			optimiseCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		t.Fatalf("Command failed: %v", err)
	}
}

// BenchmarkOptimize is a benchmark wrapper for profiling the optimize command.
// Run from repo root with: go test -bench=BenchmarkOptimize -benchtime=1x -cpuprofile=cpu.prof -memprofile=mem.prof ./cmd/keycraft
func BenchmarkOptimize(b *testing.B) {
	// Ensure we're in the repo root (where data/ directory exists)
	if err := os.Chdir("../.."); err != nil {
		b.Fatalf("Failed to change to repo root: %v", err)
	}

	// Save original args and restore after benchmark
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up command line arguments
	os.Args = []string{
		"keycraft",
		"o",
		"-g", "100",
		"--mt", "1",
		"--seed", "42",
		"qwerty",
	}

	for b.Loop() {
		app := &cli.App{
			Name:  "keycraft",
			Usage: "A CLI tool for crafting better keyboard layouts",
			Commands: []*cli.Command{
				optimiseCommand,
			},
		}

		if err := app.Run(os.Args); err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}
