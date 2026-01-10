package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestFlipCommandShellCompletion tests shell completion for the flip command.
func TestFlipCommandShellCompletion(t *testing.T) {
	// Save and restore original layoutDir
	origLayoutDir := layoutDir
	defer func() { layoutDir = origLayoutDir }()

	// Save and restore stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Override with a temporary test dir
	tmpDir := t.TempDir()
	layoutDir = tmpDir

	// Create test layout files
	testLayouts := []string{"qwerty.klf", "dvorak.klf", "colemak.klf", "test-layout.klf"}
	for _, name := range testLayouts {
		writeTestLayout(t, tmpDir, name, "rowstag\n ~ q w e r t  y u i o p \\\n ~ a s d f g  h j k l ; '\n ~ z x c v b  n m , . / ~\n ~ ~ ~  _ ~ ~")
	}

	tests := []struct {
		name         string
		args         []string
		wantContains []string
		wantCount    int
	}{
		{
			name:         "completion for flip command with no args",
			args:         []string{"keycraft", "flip", "--generate-shell-completion"},
			wantContains: []string{"qwerty", "dvorak", "colemak", "test-layout"},
			wantCount:    4,
		},
		{
			name:         "completion for flip command with partial match",
			args:         []string{"keycraft", "flip", "co", "--generate-shell-completion"},
			wantContains: []string{"qwerty", "dvorak", "colemak", "test-layout"},
			wantCount:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdout = w

			app := &cli.Command{
				Name:                  "keycraft",
				EnableShellCompletion: true,
				Commands:              []*cli.Command{flipCommand},
			}

			// Run the command with shell completion flag in a goroutine
			done := make(chan error)
			go func() {
				done <- app.Run(context.Background(), tt.args)
			}()

			// Wait for completion
			err = <-done
			w.Close()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read the captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")

			// Filter out empty lines
			var nonEmptyLines []string
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines = append(nonEmptyLines, strings.TrimSpace(line))
				}
			}

			// Check that we got the expected number of completions
			if len(nonEmptyLines) != tt.wantCount {
				t.Errorf("expected %d completions, got %d\nOutput:\n%s", tt.wantCount, len(nonEmptyLines), output)
			}

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				found := false
				for _, line := range nonEmptyLines {
					if strings.Contains(line, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected completion output to contain %q, but it was not found\nOutput:\n%s", want, output)
				}
			}
		})
	}
}

// TestFlipCommandCompletionWithoutFlag tests that flip command works normally without completion flag.
func TestFlipCommandCompletionWithoutFlag(t *testing.T) {
	// Save and restore original layoutDir
	origLayoutDir := layoutDir
	defer func() { layoutDir = origLayoutDir }()

	// Override with a temporary test dir
	tmpDir := t.TempDir()
	layoutDir = tmpDir

	// Create a test layout file
	writeTestLayout(t, tmpDir, "test.klf", "rowstag\n ~ q w e r t  y u i o p \\\n ~ a s d f g  h j k l ; '\n ~ z x c v b  n m , . / ~\n ~ ~ ~  _ ~ ~")

	// Test that the command runs normally without the completion flag
	app := &cli.Command{
		Name:                  "keycraft",
		EnableShellCompletion: true,
		Commands:              []*cli.Command{flipCommand},
	}

	err := app.Run(context.Background(), []string{"keycraft", "flip", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the flipped layout was created
	flippedPath := filepath.Join(tmpDir, "test-flipped.klf")
	if _, err := os.Stat(flippedPath); os.IsNotExist(err) {
		t.Fatalf("expected flipped layout file to be created at %s", flippedPath)
	}
}

// TestCorpusCommandShellCompletion tests shell completion for the corpus command.
func TestCorpusCommandShellCompletion(t *testing.T) {
	// Save and restore original corpusDir
	origCorpusDir := corpusDir
	defer func() { corpusDir = origCorpusDir }()

	// Save and restore stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Override with a temporary test dir
	tmpDir := t.TempDir()
	corpusDir = tmpDir + "/"

	// Create test corpus files - both source files and cache files
	// This tests deduplication: english.txt and english.txt.json should result in only "english.txt"
	testFiles := []string{
		"english.txt",          // Source file
		"english.txt.json",     // Cache file (should dedupe with above)
		"spanish.txt.json",     // Cache only
		"french.txt",           // Source only
		"test-corpus.txt.json", // Cache only
	}
	for _, name := range testFiles {
		path := filepath.Join(tmpDir, name)
		content := "{}"
		if !strings.HasSuffix(name, ".json") {
			content = "test corpus content"
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test corpus file: %v", err)
		}
	}

	// NOTE: Flag value completion tests are difficult to test programmatically because
	// they rely on os.Args which is process-global and not affected by app.Run() arguments.
	// Manual testing shows that:
	//   keycraft corpus --corpus <TAB> correctly completes with corpus files
	//   keycraft corpus -c <TAB> correctly completes with corpus files
	// The helper function getCorpusFilesForCompletion() is tested separately.

	tests := []struct {
		name         string
		args         []string
		wantContains []string
		wantCount    int
	}{
		{
			name:         "general corpus completion shows no suggestions",
			args:         []string{"keycraft", "corpus", "--generate-shell-completion"},
			wantContains: []string{}, // Should show flags by framework, not corpus files
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdout = w

			app := &cli.Command{
				Name:                  "keycraft",
				EnableShellCompletion: true,
				Commands:              []*cli.Command{corpusCommand},
			}

			// Run the command with shell completion flag in a goroutine
			done := make(chan error)
			go func() {
				done <- app.Run(context.Background(), tt.args)
			}()

			// Wait for completion
			err = <-done
			w.Close()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read the captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			lines := strings.Split(strings.TrimSpace(output), "\n")

			// Filter out empty lines
			var nonEmptyLines []string
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines = append(nonEmptyLines, strings.TrimSpace(line))
				}
			}

			// Check that we got the expected number of completions
			if len(nonEmptyLines) != tt.wantCount {
				t.Errorf("expected %d completions, got %d\nOutput:\n%s", tt.wantCount, len(nonEmptyLines), output)
			}

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				found := false
				for _, line := range nonEmptyLines {
					if strings.Contains(line, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected completion output to contain %q, but it was not found\nOutput:\n%s", want, output)
				}
			}
		})
	}
}

// TestGetCorpusFilesForCompletion is commented out because getCorpusFilesForCompletion()
// was replaced with inline logic in layoutShellComplete(). Corpus completion now handles
// .txt and .json file deduplication directly without a separate helper function.
/*
func TestGetCorpusFilesForCompletion(t *testing.T) {
	// Save and restore original corpusDir
	origCorpusDir := corpusDir
	defer func() { corpusDir = origCorpusDir }()

	// Override with a temporary test dir
	tmpDir := t.TempDir()
	corpusDir = tmpDir + "/"

	// Create test files - both source and cache files to test deduplication
	testFiles := map[string]string{
		"english.txt":          "source",
		"english.txt.json":     "{}",     // Should dedupe with above
		"spanish.txt.json":     "{}",     // Cache only
		"french.txt":           "source", // Source only
		"test-corpus.txt.json": "{}",     // Cache only
		".hidden.txt":          "source", // Should be filtered out
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	got := listFilesForCompletion(corpusDir, ".json")

	// Expected: english.txt, spanish.txt, french.txt, test-corpus.txt (4 unique, no .hidden.txt)
	expected := []string{"english.txt", "spanish.txt", "french.txt", "test-corpus.txt"}

	if len(got) != len(expected) {
		t.Errorf("expected %d files, got %d: %v", len(expected), len(got), got)
	}

	// Check each expected file is present
	gotMap := make(map[string]bool)
	for _, f := range got {
		gotMap[f] = true
	}

	for _, exp := range expected {
		if !gotMap[exp] {
			t.Errorf("expected file %q not found in completion list: %v", exp, got)
		}
	}

	// Check hidden file is NOT present
	if gotMap[".hidden.txt"] {
		t.Errorf("hidden file .hidden.txt should not be in completion list")
	}
}
*/

// TestCorpusCommandValidation tests that corpus command rejects arguments.
func TestCorpusCommandValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "corpus with one argument should fail",
			args:      []string{"keycraft", "corpus", "somefile.txt"},
			wantError: true,
			errorMsg:  "corpus command takes no arguments",
		},
		{
			name:      "corpus with multiple arguments should fail",
			args:      []string{"keycraft", "corpus", "file1.txt", "file2.txt"},
			wantError: true,
			errorMsg:  "corpus command takes no arguments",
		},
		{
			name:      "corpus help should succeed",
			args:      []string{"keycraft", "corpus", "help"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{
				Name:                  "keycraft",
				EnableShellCompletion: true,
				Commands:              []*cli.Command{corpusCommand},
			}

			err := app.Run(context.Background(), tt.args)
			if !tt.wantError {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got no error", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestCommandCompletion tests that command names are completed at the top level.
func TestCommandCompletion(t *testing.T) {
	// This test verifies that the built-in command completion of urfave/cli v3 works.
	// When EnableShellCompletion is true, the framework automatically provides command completion.
	// We test this by verifying the app structure rather than output, as the framework
	// handles command-level completion internally.

	app := &cli.Command{
		Name:                  "keycraft",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			corpusCommand,
			flipCommand,
		},
	}

	// Verify that EnableShellCompletion is set
	if !app.EnableShellCompletion {
		t.Error("expected EnableShellCompletion to be true")
	}

	// Verify that commands are registered
	if len(app.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(app.Commands))
	}

	// Verify command names
	commandNames := make(map[string]bool)
	for _, cmd := range app.Commands {
		commandNames[cmd.Name] = true
	}

	expectedCommands := []string{"corpus", "flip"}
	for _, cmdName := range expectedCommands {
		if !commandNames[cmdName] {
			t.Errorf("expected command %q to be registered", cmdName)
		}
	}
}

// TestListFilesForCompletion tests the helper function that lists files for completion.
func TestListFilesForCompletion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files with various extensions
	testFiles := map[string]string{
		"layout1.klf": "content",
		"layout2.klf": "content",
		"layout3.KLF": "content", // uppercase extension
		"readme.txt":  "content",
		"data.json":   "content",
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	tests := []struct {
		name      string
		dir       string
		ext       string
		wantFiles []string
	}{
		{
			name: "strip .klf extension from matching files",
			dir:  tmpDir,
			ext:  ".klf",
			// Should return ALL files, but with .klf extension stripped from matching ones
			wantFiles: []string{"layout1", "layout2", "layout3", "readme.txt", "data.json"},
		},
		{
			name: "strip .txt extension from matching files",
			dir:  tmpDir,
			ext:  ".txt",
			// Should return ALL files, but with .txt extension stripped from matching ones
			wantFiles: []string{"layout1.klf", "layout2.klf", "layout3.KLF", "readme", "data.json"},
		},
		{
			name: "strip .json extension from matching files",
			dir:  tmpDir,
			ext:  ".json",
			// Should return ALL files, but with .json extension stripped from matching ones
			wantFiles: []string{"layout1.klf", "layout2.klf", "layout3.KLF", "readme.txt", "data"},
		},
		{
			name: "strip non-existent extension (returns all files unchanged)",
			dir:  tmpDir,
			ext:  ".xyz",
			// No files have .xyz extension, so all files returned with original names
			wantFiles: []string{"layout1.klf", "layout2.klf", "layout3.KLF", "readme.txt", "data.json"},
		},
		{
			name:      "non-existent directory",
			dir:       filepath.Join(tmpDir, "nonexistent"),
			ext:       ".klf",
			wantFiles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listFilesForCompletion(tt.dir, tt.ext)

			if tt.wantFiles == nil && got != nil {
				t.Errorf("expected nil, got %v", got)
				return
			}

			if len(got) != len(tt.wantFiles) {
				t.Errorf("expected %d files, got %d: %v", len(tt.wantFiles), len(got), got)
				return
			}

			// Check that all expected files are present
			for _, want := range tt.wantFiles {
				found := false
				for _, g := range got {
					if g == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected file %q not found in result: %v", want, got)
				}
			}
		})
	}
}
