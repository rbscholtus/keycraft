package main

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDirs creates temporary directories for testing and returns their paths.
// It also sets the global directory variables to point to these temp directories.
// The caller should defer restoreTestDirs() to restore the original values.
func setupTestDirs(t *testing.T) (origLayoutDir, origCorpusDir, origConfigDir string) {
	t.Helper()

	// Save original directories
	origLayoutDir = layoutDir
	origCorpusDir = corpusDir
	origConfigDir = configDir

	// Create temporary directories
	tmpDir := t.TempDir()
	layoutDir = filepath.Join(tmpDir, "layouts")
	corpusDir = filepath.Join(tmpDir, "corpus")
	configDir = filepath.Join(tmpDir, "config")

	// Create the directories
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		t.Fatalf("failed to create temp layout dir: %v", err)
	}
	if err := os.MkdirAll(corpusDir, 0755); err != nil {
		t.Fatalf("failed to create temp corpus dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create temp config dir: %v", err)
	}

	// Create default load_targets.txt to avoid warnings in tests
	writeDefaultLoadTargetsFile(t, configDir)

	return origLayoutDir, origCorpusDir, origConfigDir
}

// restoreTestDirs restores the global directory variables to their original values.
func restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir string) {
	layoutDir = origLayoutDir
	corpusDir = origCorpusDir
	configDir = origConfigDir
}

// writeTestConfigFile writes a config file for testing.
func writeTestConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config file %s: %v", name, err)
	}
	return path
}

// writeTestCorpus writes a minimal test corpus file.
func writeTestCorpus(t *testing.T, dir, name string) string {
	t.Helper()
	// Minimal corpus with enough data to avoid errors
	content := `the quick brown fox jumps over the lazy dog
the five boxing wizards jump quickly
pack my box with five dozen liquor jugs
how vexingly quick daft zebras jump
`
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test corpus %s: %v", name, err)
	}
	return path
}

// writeDefaultLoadTargetsFile writes a default load_targets.txt config file.
func writeDefaultLoadTargetsFile(t *testing.T, dir string) string {
	t.Helper()
	// Default target loads configuration
	content := `# Default target loads for testing
target-hand-load = 50, 50
target-finger-load = 7, 10, 16, 17
target-row-load = 17.5, 75.0, 7.5
pinky-penalties = 2, 1.5, 1, 0, 2, 1.5
`
	return writeTestConfigFile(t, dir, "load_targets.txt", content)
}

// Minimal valid layout content for testing
const minimalLayoutContent = `rowstag
 ~ q w e r t  y u i o p \
 ~ a s d f g  h j k l ; '
 ~ z x c v b  n m , . / ~
 ~ ~ ~  _ ~ ~
`

// Alternative layout content for testing
const alternativeLayoutContent = `colstag
 ~ q w f p g  j l u y ; \
 ~ a r s t d  h n e i o '
 ~ z x c v b  k m , . / ~
 ~ ~ ~  _ ~ ~
`
