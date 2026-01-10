package main

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"unicode/utf8"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v3"
)

// appFlagsMap is a centralized map of CLI flags used across various commands.
// It keeps flag definitions in one place, allowing commands to select only the
// flags they need, promoting reusability and consistency.
var appFlagsMap = map[string]cli.Flag{
	"corpus": &cli.StringFlag{
		Name:    "corpus",
		Aliases: []string{"c"},
		Usage:   "Corpus file for calculating metrics (from data/corpus directory).",
		Value:   "default.txt",
	},
	"corpus-rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "Maximum number of rows to display in corpus data tables.",
		Value:   100,
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if value < 1 {
				return fmt.Errorf("--rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	"coverage": &cli.Float64Flag{
		Name: "coverage",
		Usage: "Corpus word coverage percentage (0.1-100.0). Filters " +
			"low-frequency words. Forces cache rebuild.",
		Value: 98.0,
		Action: func(ctx context.Context, c *cli.Command, value float64) error {
			if value < 0.1 || value > 100.0 {
				return fmt.Errorf("--coverage must be 0.1-100 (got %f)", value)
			}
			return nil
		},
	},
	"row-load": &cli.StringFlag{
		Name:    "row-load",
		Aliases: []string{"rl"},
		Usage: "Ideal row load percentages: 3 comma-separated values for top, " +
			"home, bottom rows (auto-scaled to 100%).",
		Value: "18.5,73,8.5", // default: top, home, bottom
	},
	"finger-load": &cli.StringFlag{
		Name:    "finger-load",
		Aliases: []string{"fl"},
		Usage: "Ideal finger load percentages: 4 values (left 4 fingers, " +
			"mirrored to right) or 8 values. Thumbs always 0. Auto-scaled to 100%.",
		Value: "7.5,11,16,15.5", // default 4-values mirrored
	},
	"pinky-penalties": &cli.StringFlag{
		Name:    "pinky-penalties",
		Aliases: []string{"pp"},
		Usage: "Pinky off-home penalties: 6 values (left top-outer, top-inner, " +
			"home-outer, home-inner, bottom-outer, bottom-inner; mirrored) or " +
			"12 values (left, then right). Higher = more penalty.",
		Value: "1,1,1,0,1,1", // default 6-values mirrored
	},
	"rows": &cli.IntFlag{
		Name:    "rows",
		Aliases: []string{"r"},
		Usage:   "Maximum number of rows to display in data tables.",
		Value:   10,
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if value < 1 {
				return fmt.Errorf("--rows must be at least 1 (got %d)", value)
			}
			return nil
		},
	},
	"weights-file": &cli.StringFlag{
		Name:    "weights-file",
		Aliases: []string{"wf"},
		Usage: "Weights file for scoring layouts (from data/config directory). " +
			"Overridden by --weights flag.",
		Value: "default.txt",
	},
	"weights": &cli.StringFlag{
		Name:    "weights",
		Aliases: []string{"w"},
		Usage: "Custom metric weights as comma-separated pairs " +
			"(e.g., \"SFB=-10,LSB=-5\"). Overrides weights file.",
	},
	"metrics": &cli.StringFlag{
		Name:    "metrics",
		Aliases: []string{"m"},
		Usage: fmt.Sprintf("Metrics to display. Options: %v, or \"weighted\" "+
			"(metrics with |weight|>=0.01), or comma-separated list.",
			slices.Sorted(maps.Keys(kc.MetricsMap))),
		Value: "weighted",
	},
	"deltas": &cli.StringFlag{
		Name:    "deltas",
		Aliases: []string{"d"},
		Usage: "Delta display mode: \"none\", \"rows\" (row-by-row), " +
			"\"median\" (vs median), or \"<layout>\" name to compare against.",
		Value: "none",
	},
	"output": &cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "Output format: \"table\", \"html\", or \"csv\".",
		Value:   "table",
	},
	"pins-file": &cli.StringFlag{
		Name:    "pins-file",
		Aliases: []string{"pf"},
		Usage: "File specifying keys to pin during optimization. " +
			"Defaults to pinning '~' and '_'.",
	},
	"pins": &cli.StringFlag{
		Name:    "pins",
		Aliases: []string{"p"},
		Usage: "Additional characters to pin (e.g., 'aeiouy'). " +
			"Combined with pins-file.",
	},
	"free": &cli.StringFlag{
		Name:    "free",
		Aliases: []string{"f"},
		Usage: "Characters free to move during optimization. " +
			"All others are pinned.",
	},
	"generations": &cli.UintFlag{
		Name:    "generations",
		Aliases: []string{"gens", "g"},
		Usage:   "Number of optimization iterations to run.",
		Value:   1000,
	},
	"maxtime": &cli.UintFlag{
		Name:    "maxtime",
		Aliases: []string{"mt"},
		Usage:   "Maximum optimization time in minutes.",
		Value:   5,
	},
	"seed": &cli.Int64Flag{
		Name:    "seed",
		Aliases: []string{"s"},
		Usage:   "Random seed for reproducible results. Uses current timestamp if 0.",
		Value:   0,
	},
	"log-file": &cli.StringFlag{
		Name:    "log-file",
		Aliases: []string{"lf"},
		Usage:   "JSONL log file path for detailed optimization metrics.",
	},
	"compact-trigrams": &cli.BoolFlag{
		Name:  "compact-trigrams",
		Usage: "Omit common trigram categories (ALT-NML, 2RL-IN, 2RL-OUT, 3RL-IN, 3RL-OUT) from trigram table.",
		Value: false,
	},
	"trigram-rows": &cli.IntFlag{
		Name:  "trigram-rows",
		Usage: "Maximum number of trigrams to display in trigram table.",
		Value: 50,
		Action: func(ctx context.Context, c *cli.Command, value int) error {
			if value < 1 {
				return fmt.Errorf("--trigram-rows must be at least 1 (got %d)", value)
			}
			return nil
		},
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
// with the given extension, stripped of that extension for cleaner completion.
func listFilesForCompletion(dir, ext string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ext) {
			// Strip the extension for completion
			files = append(files, strings.TrimSuffix(name, filepath.Ext(name)))
		}
	}
	return files
}

// getCorpusFilesForCompletion returns a deduplicated list of corpus files
// from the corpus directory, handling both source and cache files.
func getCorpusFilesForCompletion() []string {
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		return nil
	}

	// Use a map to deduplicate corpus names
	seen := make(map[string]bool)
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

		// For cached files (*.json), strip the .json extension
		name = strings.TrimSuffix(name, ".json")

		// Add to list if not already seen
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
func layoutShellComplete(ctx context.Context, c *cli.Command) {
	// Find the position of --generate-shell-completion in os.Args
	completionPos := slices.Index(os.Args, "--generate-shell-completion")

	// If we found it and there's an arg before it, check what it is
	if completionPos > 0 {
		prevArg := os.Args[completionPos-1]

		// Check if it's a flag that needs value completion
		if prevArg == "--corpus" || prevArg == "-c" {
			files := getCorpusFilesForCompletion()
			for _, f := range files {
				fmt.Println(f)
			}
			return
		}

		// Check if previous arg starts with "-" (including "-" and "--")
		// This means user is completing a flag name, so show all flags
		// Note: "--" by itself is the argument terminator and should also show flags
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

// ////////
func tracef(format string, a ...any) {
	if os.Getenv("URFAVE_CLI_TRACING") != "on" {
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}

	pc, _, _, _ := runtime.Caller(1)
	cf := runtime.FuncForPC(pc)

	fmt.Fprintf(
		os.Stderr,
		strings.Join([]string{
			"## TRACE ",
			//file,
			//":",
			//fmt.Sprintf("%v", line),
			//" ",
			fmt.Sprintf("(%s)", cf.Name()),
			" ",
			format,
		}, ""),
		a...,
	)
}

func printCommandSuggestions(commands []*cli.Command, writer io.Writer) {
	for _, command := range commands {
		if command.Hidden {
			continue
		}
		if strings.HasSuffix(os.Getenv("SHELL"), "zsh") {
			_, _ = fmt.Fprintf(writer, "%s:%s\n", command.Name, command.Usage)
		} else {
			_, _ = fmt.Fprintf(writer, "%s\n", command.Name)
		}
	}
}

func printFlagSuggestions(lastArg string, flags []cli.Flag, writer io.Writer) {
	// Trim to handle both "-short" and "--long" flags.
	cur := strings.TrimLeft(lastArg, "-")
	for _, flag := range flags {
		if bflag, ok := flag.(*cli.BoolFlag); ok && bflag.Hidden {
			continue
		}

		usage := ""
		if docFlag, ok := flag.(cli.DocGenerationFlag); ok {
			usage = docFlag.GetUsage()
		}

		name := strings.TrimSpace(flag.Names()[0])
		// this will get total count utf8 letters in flag name
		count := utf8.RuneCountInString(name)
		if count > 2 {
			count = 2 // reuse this count to generate single - or -- in flag completion
		}
		// if flag name has more than one utf8 letter and last argument in cli has -- prefix then
		// skip flag completion for short flags example -v or -x
		if strings.HasPrefix(lastArg, "--") && count == 1 {
			continue
		}
		// match if last argument matches this flag and it is not repeated
		if strings.HasPrefix(name, cur) && cur != name /* && !cliArgContains(name, os.Args)*/ {
			flagCompletion := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
			if usage != "" && strings.HasSuffix(os.Getenv("SHELL"), "zsh") {
				flagCompletion = fmt.Sprintf("%s:%s", flagCompletion, usage)
			}
			fmt.Fprintln(writer, flagCompletion)
		}
	}
}

func MyCompleteWithFlags(ctx context.Context, cmd *cli.Command) {
	args := os.Args
	//if cmd != nil && cmd.parent != nil {
	//	args = cmd.Args().Slice()
	//	tracef("running default complete with flags[%v] on command %[2]q", args, cmd.Name)
	//} else {
	//	tracef("running default complete with os.Args flags[%v]", args)
	//}
	argsLen := len(args)
	lastArg := ""
	// parent command will have --generate-shell-completion so we need
	// to account for that
	if argsLen > 1 {
		lastArg = args[argsLen-2]
	} else if argsLen > 0 {
		lastArg = args[argsLen-1]
	}

	if lastArg == "--" {
		tracef("No completions due to termination")
		return
	}

	if lastArg == "--generate-shell-completion" {
		lastArg = ""
	}

	if strings.HasPrefix(lastArg, "-") {
		tracef("printing flag suggestion for flag[%v] on command %[1]q", lastArg, cmd.Name)
		printFlagSuggestions(lastArg, cmd.Flags, cmd.Root().Writer)
		return
	}

	if cmd != nil {
		tracef("printing command suggestions on command %[1]q", cmd.Name)
		printCommandSuggestions(cmd.Commands, cmd.Root().Writer)
		return
	}
}
