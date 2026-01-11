package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

// appFlagsMap is a centralized map of CLI flags shared across multiple commands.
// Command-specific flags are defined in their respective command files.
// Flags are categorized for better help output organization.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:     "corpus",
		Aliases:  []string{"c"},
		Usage:    "Corpus file for calculating metrics (from data/corpus directory).",
		Value:    "default.txt",
		Category: "", // General/uncategorized
	},
	"load-targets-file": &cli.StringFlag{
		Name:    "load-targets-file",
		Aliases: []string{"ldt"},
		Usage: "Configuration file for target load distributions (row/finger/hand loads, pinky penalties). " +
			"Overridden by individual flags. (from data/config directory)",
		Value:    "load_targets.txt",
		Category: "Targets and Weights",
	},
	"target-hand-load": &cli.StringFlag{
		Name:    "target-hand-load",
		Aliases: []string{"thl"},
		Usage: "Target hand load percentages: 2 comma-separated values for left, " +
			"right hands (auto-scaled to 100%). Overrides load_targets file.",
		Category: "Targets and Weights",
	},
	"target-finger-load": &cli.StringFlag{
		Name:    "target-finger-load",
		Aliases: []string{"tfl"},
		Usage: "Target finger load percentages: 4 values (left 4 fingers, " +
			"mirrored to right) or 8 values. Thumbs always 0. Auto-scaled to 100%." +
			" Overrides load_targets file.",
		Category: "Targets and Weights",
	},
	"target-row-load": &cli.StringFlag{
		Name:    "target-row-load",
		Aliases: []string{"trl"},
		Usage: "Target row load percentages: 3 comma-separated values for top, " +
			"home, bottom rows (auto-scaled to 100%). Overrides load_targets file.",
		Category: "Targets and Weights",
	},
	"pinky-penalties": &cli.StringFlag{
		Name:    "pinky-penalties",
		Aliases: []string{"pp"},
		Usage: "Pinky off-home penalties: 6 values (left top-outer, top-inner, " +
			"home-outer, home-inner, bottom-outer, bottom-inner; mirrored) or " +
			"12 values (left, then right). Higher = more penalty. Overrides load_targets file.",
		Category: "Targets and Weights",
	},
	"weights-file": &cli.StringFlag{
		Name:    "weights-file",
		Aliases: []string{"wf"},
		Usage: "Weights file for scoring layouts (from data/config directory). " +
			"Overridden by --weights flag.",
		Value:    "weights.txt",
		Category: "Targets and Weights",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage: "Custom metric weights as comma-separated pairs " +
			"(e.g., \"SFB=-10,LSB=-5\"). Overrides weights file.",
		Category: "Targets and Weights",
	},
}

// flagsSlice returns a slice of cli.Flag pointers for the given keys from appFlagsMap.
func flagsSlice(keys ...string) []cli.Flag {
	flags := make([]cli.Flag, 0, len(keys))
	for _, k := range keys {
		if f, ok := appFlagsMap[k]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

// isShellCompletion returns true if the current invocation is for shell completion.
// This is used to skip validation during completion to avoid error messages.
func isShellCompletion() bool {
	return slices.Contains(os.Args, "--generate-shell-completion")
}

// getLayoutArgs retrieves the list of layout arguments passed to the CLI command.
// Each layout name is normalized by ensuring it has the ".klf" extension.
func getLayoutArgs(c *cli.Command) []string {
	layouts := c.Args().Slice()
	for i := range layouts {
		layouts[i] = ensureKlf(layouts[i])
	}
	return layouts
}

// listFilesForCompletion returns a list of files from the specified directory
// that have the given extension, with the extension stripped for cleaner completion.
// It also skips hidden/system files and deduplicates basenames.
func listFilesForCompletion(dir, stripExt string) []string {
	// Normalize ext to lowercase for reliable case-insensitive matching
	if stripExt != "" {
		stripExt = strings.ToLower(stripExt)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool, len(entries))
	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Skip system/hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		// If an extension is specified, strip it from matching files
		// (but don't filter out non-matching files)
		if stripExt != "" && strings.HasSuffix(strings.ToLower(name), stripExt) {
			// Strip the extension from the name
			name = name[:len(name)-len(stripExt)]
		}

		// Deduplicate basenames (preserves first-seen order)
		if !seen[name] {
			seen[name] = true
			files = append(files, name)
		}
	}

	return files
}

// layoutShellComplete provides shell completion for layout file arguments and flags.
// It suggests .klf layout files from the data/layouts directory, corpus files for --corpus flag,
// or shows flags when appropriate.
//
// Note: When the user types "command --<TAB>", zsh sends "command -- --generate-shell-completion".
// The urfave/cli framework treats "--" as an argument terminator and won't call this function.
// As a result, "command --<TAB>" produces no completions (shell beeps), which is acceptable
// behavior since there are limited scenarios where this would be useful.
func layoutShellComplete(ctx context.Context, c *cli.Command) {
	// Find the position of --generate-shell-completion in os.Args
	completionPos := slices.Index(os.Args, "--generate-shell-completion")

	// If we found it and there's an arg before it, check what it is
	if completionPos > 0 {
		prevArg := os.Args[completionPos-1]

		// Check if it's a flag that needs value completion
		if prevArg == "--corpus" || prevArg == "-c" {
			// For corpus files, we need to list both .txt and .txt.json files
			// and deduplicate them. We pass "" to list all files, then filter manually.
			entries, _ := os.ReadDir(corpusDir)
			seen := make(map[string]bool)
			var files []string
			for _, entry := range entries {
				if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				name := entry.Name()
				// Strip .json suffix from cached files
				name = strings.TrimSuffix(name, ".json")
				if !seen[name] {
					seen[name] = true
					files = append(files, name)
				}
			}
			for _, f := range files {
				fmt.Println(f)
			}
			return
		}

		// Check if previous arg starts with "-" (e.g., "--cor", "--ro")
		// This means user is completing a partial flag name, so show matching flags
		if strings.HasPrefix(prevArg, "-") {
			// Print all flags with descriptions
			for _, flag := range c.Flags {
				// Get flag names
				names := flag.Names()
				if len(names) == 0 {
					continue
				}

				// Use the longest name (usually the --long-form)
				flagName := "--" + names[0]
				for _, name := range names {
					if len(name) > len(names[0]) {
						flagName = "--" + name
					}
				}

				// Get description
				usage := ""
				if u, ok := flag.(interface{ GetUsage() string }); ok {
					usage = u.GetUsage()
				}

				// Print in format: --flag:description
				if usage != "" {
					fmt.Printf("%s:%s\n", flagName, usage)
				} else {
					fmt.Println(flagName)
				}
			}
			return
		}
	}

	if c.Name != "corpus" {
		// For regular arguments (layout files), suggest .klf files
		files := listFilesForCompletion(layoutDir, ".klf")
		for _, f := range files {
			fmt.Println(f)
		}
	}
}
