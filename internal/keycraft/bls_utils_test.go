package keycraft

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// createTestLayout creates a simple test layout for testing pin functions.
func createTestLayout() *SplitLayout {
	// Create a simple layout with lowercase letters and some special chars
	runes := [42]rune{
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', '[', ']',
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', '\'', '\\',
		'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', 0, 0,
		' ', 0, 0, ' ', 0, 0,
	}

	runeInfo := make(map[rune]KeyInfo)
	for i, r := range runes {
		if r != 0 {
			// Calculate row and column from index
			var row, col uint8
			if i < 12 {
				row = 0
				col = uint8(i)
			} else if i < 24 {
				row = 1
				col = uint8(i - 12)
			} else if i < 36 {
				row = 2
				col = uint8(i - 24)
			} else {
				row = 3
				col = uint8(i - 36)
			}
			runeInfo[r] = NewKeyInfo(row, col, ROWSTAG)
		}
	}

	return NewSplitLayout("test", ROWSTAG, runes, runeInfo)
}

// TestLoadPins_ValidFile tests loading a valid pins file.
func TestLoadPins_ValidFile(t *testing.T) {
	// Create temporary pins file
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * * -  - - * * - *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	// Load pins
	pinned, err := LoadPins(pinsFile)
	if err != nil {
		t.Fatalf("LoadPins failed: %v", err)
	}

	if pinned == nil {
		t.Fatal("LoadPins returned nil")
	}

	// Verify specific positions
	expected := []struct {
		index  int
		pinned bool
	}{
		{0, true},   // *
		{1, false},  // -
		{2, false},  // -
		{3, true},   // *
		{12, true},  // * (row 2, col 1)
		{13, true},  // *
		{14, true},  // *
		{24, true},  // * (row 3, col 1)
		{25, false}, // -
		{38, true},  // * (thumb row)
	}

	for _, exp := range expected {
		if pinned[exp.index] != exp.pinned {
			t.Errorf("Expected pinned[%d] = %v, got %v", exp.index, exp.pinned, pinned[exp.index])
		}
	}
}

// TestLoadPins_WithComments tests loading pins file with comments and empty lines.
func TestLoadPins_WithComments(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `# This is a comment
* - - * * -  - - * * - *

# Another comment
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	pinned, err := LoadPins(pinsFile)
	if err != nil {
		t.Fatalf("LoadPins failed: %v", err)
	}

	if pinned == nil {
		t.Fatal("LoadPins returned nil")
	}

	// Just verify we got 42 keys
	count := 0
	for range pinned {
		count++
	}
	if count != 42 {
		t.Errorf("Expected 42 keys, got %d", count)
	}
}

// TestLoadPins_DifferentSymbols tests all valid pin/unpin symbols.
func TestLoadPins_DifferentSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	// Use different valid symbols for pinned and unpinned
	content := `X x * X x *  X x * X x *
_ - . _ - .  _ - . _ - .
. . . . . .  . . . . . .
      * x X  - _ .`

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	pinned, err := LoadPins(pinsFile)
	if err != nil {
		t.Fatalf("LoadPins failed: %v", err)
	}

	// Row 1: X x * X x * X x * X x *
	// X=pinned, x=pinned, *=pinned (all capitals and * are pinned)
	if !pinned[0] {
		t.Error("Row 1 position 0 (X) should be pinned")
	}
	if !pinned[1] {
		t.Error("Row 1 position 1 (x) should be pinned")
	}
	if !pinned[2] {
		t.Error("Row 1 position 2 (*) should be pinned")
	}

	// Row 2: _ - . _ - . _ - . _ - . (all unpinned)
	if pinned[12] || pinned[13] || pinned[14] {
		t.Errorf("Row 2 positions 12-14 should all be unpinned, got: [%v, %v, %v]", pinned[12], pinned[13], pinned[14])
	}

	// Row 3: . . . . . . . . . . . . (all unpinned)
	for i := 24; i < 36; i++ {
		if pinned[i] {
			t.Errorf("Position %d should be unpinned", i)
		}
	}

	// Row 4: * x X - _ . (mixed)
	// *=pinned(36), x=pinned(37), X=pinned(38), -=unpinned(39), _=unpinned(40), .=unpinned(41)
	if !pinned[36] {
		t.Error("Position 36 (*) should be pinned")
	}
	if !pinned[37] {
		t.Error("Position 37 (x) should be pinned")
	}
	if !pinned[38] {
		t.Error("Position 38 (X) should be pinned")
	}
	if pinned[39] {
		t.Error("Position 39 (-) should be unpinned")
	}
	if pinned[40] {
		t.Error("Position 40 (_) should be unpinned")
	}
	if pinned[41] {
		t.Error("Position 41 (.) should be unpinned")
	}
}

// TestLoadPins_FileNotFound tests error when file doesn't exist.
func TestLoadPins_FileNotFound(t *testing.T) {
	_, err := LoadPins("/nonexistent/path/pins.pin")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

// TestLoadPins_InvalidFormat_TooFewRows tests error when file has too few rows.
func TestLoadPins_InvalidFormat_TooFewRows(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * * -  - - * * - *
* * * * * -  - - * * * -`
	// Missing 2 rows

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	_, err := LoadPins(pinsFile)
	if err == nil {
		t.Fatal("Expected error for too few rows")
	}
	if !contains(err.Error(), "not enough rows") {
		t.Errorf("Expected 'not enough rows' error, got: %v", err)
	}
}

// TestLoadPins_InvalidFormat_WrongKeyCount tests error when row has wrong number of keys.
func TestLoadPins_InvalidFormat_WrongKeyCount(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * * -  - - * * - * * *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`
	// First row has 14 keys instead of 12

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	_, err := LoadPins(pinsFile)
	if err == nil {
		t.Fatal("Expected error for wrong key count")
	}
	if !contains(err.Error(), "has 14 keys, expected 12") {
		t.Errorf("Expected key count error, got: %v", err)
	}
}

// TestLoadPins_InvalidCharacter tests error for invalid characters.
func TestLoadPins_InvalidCharacter(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * * -  - - * * - *
* * a * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`
	// 'a' is not a valid pin symbol

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	_, err := LoadPins(pinsFile)
	if err == nil {
		t.Fatal("Expected error for invalid character")
	}
	if !contains(err.Error(), "invalid character") {
		t.Errorf("Expected 'invalid character' error, got: %v", err)
	}
}

// TestLoadPins_MultiCharacterKey tests error when key has multiple characters.
func TestLoadPins_MultiCharacterKey(t *testing.T) {
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * ** -  - - * * - *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`
	// '**' should be two separate keys but is parsed as one

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	_, err := LoadPins(pinsFile)
	if err == nil {
		t.Fatal("Expected error for multi-character key")
	}
}

// TestLoadPinsFromParams_NilLayout tests error when layout is nil.
func TestLoadPinsFromParams_NilLayout(t *testing.T) {
	_, err := LoadPinsFromParams("", "", "", nil)
	if err == nil {
		t.Fatal("Expected error for nil layout")
	}
	if !contains(err.Error(), "layout cannot be nil") {
		t.Errorf("Expected 'layout cannot be nil' error, got: %v", err)
	}
}

// TestLoadPinsFromParams_MutualExclusion tests that free and pins cannot be used together.
func TestLoadPinsFromParams_MutualExclusion(t *testing.T) {
	layout := createTestLayout()

	// Test free + pins
	_, err := LoadPinsFromParams("", "abc", "xyz", layout)
	if err == nil {
		t.Fatal("Expected error when using both --free and --pins")
	}
	if !contains(err.Error(), "cannot use both") {
		t.Errorf("Expected mutual exclusion error, got: %v", err)
	}

	// Test free + path
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")
	content := `* - - * * -  - - * * - *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`
	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	_, err = LoadPinsFromParams(pinsFile, "", "xyz", layout)
	if err == nil {
		t.Fatal("Expected error when using both --free and --pins-file")
	}
}

// TestLoadPinsFromParams_DefaultBehavior tests default pinning (empty keys and spaces).
func TestLoadPinsFromParams_DefaultBehavior(t *testing.T) {
	layout := createTestLayout()

	pinned, err := LoadPinsFromParams("", "", "", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// Check that spaces are pinned (indices 36, 39)
	if !pinned[36] || !pinned[39] {
		t.Error("Spaces should be pinned by default")
	}

	// Check that empty keys are pinned (indices 34, 35, 37, 38, 40, 41)
	emptyIndices := []int{34, 35, 37, 38, 40, 41}
	for _, idx := range emptyIndices {
		if !pinned[idx] {
			t.Errorf("Empty key at index %d should be pinned by default", idx)
		}
	}

	// Check that regular letters are not pinned
	if pinned[0] || pinned[1] || pinned[12] {
		t.Error("Regular letters should not be pinned by default")
	}
}

// TestLoadPinsFromParams_PinSpecificChars tests pinning specific characters.
func TestLoadPinsFromParams_PinSpecificChars(t *testing.T) {
	layout := createTestLayout()

	// Pin 'q', 'w', 'a'
	pinned, err := LoadPinsFromParams("", "qwa", "", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// q=0, w=1, a=12
	if !pinned[0] || !pinned[1] || !pinned[12] {
		t.Error("Specified characters should be pinned")
	}

	// Other letters should not be pinned
	if pinned[2] || pinned[3] || pinned[13] {
		t.Error("Unspecified letters should not be pinned")
	}

	// Spaces should still be pinned (default behavior)
	if !pinned[36] {
		t.Error("Spaces should be pinned by default")
	}
}

// TestLoadPinsFromParams_PinWithSeparators tests that commas and spaces are ignored.
func TestLoadPinsFromParams_PinWithSeparators(t *testing.T) {
	layout := createTestLayout()

	// Use comma-separated and space-separated
	pinned, err := LoadPinsFromParams("", "q, w, a", "", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// q=0, w=1, a=12
	if !pinned[0] || !pinned[1] || !pinned[12] {
		t.Error("Specified characters should be pinned even with separators")
	}
}

// TestLoadPinsFromParams_PinUnavailableChar tests error when pinning unavailable character.
func TestLoadPinsFromParams_PinUnavailableChar(t *testing.T) {
	layout := createTestLayout()

	_, err := LoadPinsFromParams("", "qwZ", "", layout)
	if err == nil {
		t.Fatal("Expected error when pinning unavailable character")
	}
	if !contains(err.Error(), "cannot pin unavailable character") {
		t.Errorf("Expected unavailable character error, got: %v", err)
	}
}

// TestLoadPinsFromParams_FreeMode tests free mode (all others pinned).
func TestLoadPinsFromParams_FreeMode(t *testing.T) {
	layout := createTestLayout()

	// Free only 'q', 'w', 'e'
	pinned, err := LoadPinsFromParams("", "", "qwe", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// q=0, w=1, e=2 should be free (not pinned)
	if pinned[0] || pinned[1] || pinned[2] {
		t.Error("Free characters should not be pinned")
	}

	// All others should be pinned
	if !pinned[3] || !pinned[12] || !pinned[24] {
		t.Error("Non-free characters should be pinned in free mode")
	}
}

// TestLoadPinsFromParams_FreeWithSeparators tests free mode with separators.
func TestLoadPinsFromParams_FreeWithSeparators(t *testing.T) {
	layout := createTestLayout()

	pinned, err := LoadPinsFromParams("", "", "q, w, e", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// q=0, w=1, e=2 should be free
	if pinned[0] || pinned[1] || pinned[2] {
		t.Error("Free characters should not be pinned even with separators")
	}
}

// TestLoadPinsFromParams_FreeUnavailableChar tests error when freeing unavailable character.
func TestLoadPinsFromParams_FreeUnavailableChar(t *testing.T) {
	layout := createTestLayout()

	_, err := LoadPinsFromParams("", "", "qwZ", layout)
	if err == nil {
		t.Fatal("Expected error when freeing unavailable character")
	}
	if !contains(err.Error(), "cannot free unavailable character") {
		t.Errorf("Expected unavailable character error, got: %v", err)
	}
}

// TestLoadPinsFromParams_LoadFromFile tests loading pins from file.
func TestLoadPinsFromParams_LoadFromFile(t *testing.T) {
	layout := createTestLayout()
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	content := `* - - * * -  - - * * - *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	pinned, err := LoadPinsFromParams(pinsFile, "", "", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// Verify some positions from file
	if !pinned[0] || pinned[1] || pinned[2] || !pinned[3] {
		t.Error("Pins not loaded correctly from file")
	}
}

// TestLoadPinsFromParams_FileAndPins tests combining file and pins string.
func TestLoadPinsFromParams_FileAndPins(t *testing.T) {
	layout := createTestLayout()
	tmpDir := t.TempDir()
	pinsFile := filepath.Join(tmpDir, "test.pin")

	// File has position 1 (w) unpinned
	content := `* - - * * -  - - * * - *
* * * * * -  - - * * * -
* - - - * -  - - - - - *
      * * *  * * *`

	if err := os.WriteFile(pinsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test pins file: %v", err)
	}

	// Now also pin 'w' via pins string
	pinned, err := LoadPinsFromParams(pinsFile, "w", "", layout)
	if err != nil {
		t.Fatalf("LoadPinsFromParams failed: %v", err)
	}

	// w=1 should now be pinned (overridden by pins string)
	if !pinned[1] {
		t.Error("Pins string should override file settings")
	}
}

// TestLoadPinsFromParams_FileNotFound tests error when pins file doesn't exist.
func TestLoadPinsFromParams_FileNotFound(t *testing.T) {
	layout := createTestLayout()

	_, err := LoadPinsFromParams("/nonexistent/pins.pin", "", "", layout)
	if err == nil {
		t.Fatal("Expected error for nonexistent pins file")
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to create a test BLS instance for benchmarking
func createBenchBLS(b *testing.B) (*BLS, *SplitLayout) {
	b.Helper()

	// Load corpus
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	// Load layout
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	// Create pinned keys (default: pin spaces and empty keys)
	pinned := &PinnedKeys{}
	for i, r := range layout.Runes {
		if r == 0 || r == ' ' {
			pinned[i] = true
		}
	}

	// Count free keys
	numFree := 0
	for _, isPinned := range pinned {
		if !isPinned {
			numFree++
		}
	}

	// Create parameters
	params := DefaultBLSParams(numFree)

	// Create scorer
	scorer, err := NewScorer("../../data/layouts", corpus, &TargetLoads{TargetRowLoad: DefaultTargetRowLoad(), TargetFingerLoad: DefaultTargetFingerLoad(), TargetHandLoad: DefaultTargetHandLoad(), PinkyPenalties: DefaultPinkyPenalties()}, NewWeights())
	if err != nil {
		b.Fatalf("Failed to create scorer: %v", err)
	}

	// Create BLS instance
	bls := NewBLS(params, scorer, corpus, pinned)

	return bls, layout
}

// BenchmarkIdentifyProblematicKeys benchmarks the original implementation
func BenchmarkIdentifyProblematicKeys(b *testing.B) {
	bls, layout := createBenchBLS(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bls.identifyProblematicKeys(layout)
	}
}

// BenchmarkIdentifyProblematicKeys2 benchmarks the optimized implementation
func BenchmarkIdentifyProblematicKeys2(b *testing.B) {
	bls, layout := createBenchBLS(b)

	// Pre-filter bigrams for the layout (normally done in Optimize())
	bls.relevantBigrams = make([]BigramCount, 0, len(bls.corpus.Bigrams))
	for bi, cnt := range bls.corpus.Bigrams {
		_, ok1 := layout.RuneInfo[bi[0]]
		_, ok2 := layout.RuneInfo[bi[1]]
		if ok1 && ok2 {
			bls.relevantBigrams = append(bls.relevantBigrams, BigramCount{
				Bigram: bi,
				Count:  cnt,
			})
		}
	}
	sort.Slice(bls.relevantBigrams, func(i, j int) bool {
		return bls.relevantBigrams[i].Count > bls.relevantBigrams[j].Count
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bls.identifyProblematicKeys2(layout)
	}
}

// TestIdentifyProblematicKeysEquivalence verifies both implementations produce same results
func TestIdentifyProblematicKeysEquivalence(t *testing.T) {
	// Load corpus
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 0)
	if err != nil {
		t.Skipf("Skipping test - corpus not available: %v", err)
	}

	// Load layout
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		t.Skipf("Skipping test - layout not available: %v", err)
	}

	// Create pinned keys
	pinned := &PinnedKeys{}
	for i, r := range layout.Runes {
		if r == 0 || r == ' ' {
			pinned[i] = true
		}
	}

	// Count free keys
	numFree := 0
	for _, isPinned := range pinned {
		if !isPinned {
			numFree++
		}
	}

	// Create parameters
	params := DefaultBLSParams(numFree)

	// Create scorer
	scorer, err := NewScorer("../../data/layouts", corpus, &TargetLoads{TargetRowLoad: DefaultTargetRowLoad(), TargetFingerLoad: DefaultTargetFingerLoad(), TargetHandLoad: DefaultTargetHandLoad(), PinkyPenalties: DefaultPinkyPenalties()}, NewWeights())
	if err != nil {
		t.Skipf("Skipping test - layouts not available: %v", err)
	}

	// Create BLS instance
	bls := NewBLS(params, scorer, corpus, pinned)

	// Pre-filter bigrams for the layout (normally done in Optimize())
	bls.relevantBigrams = make([]BigramCount, 0, len(corpus.Bigrams))
	for bi, cnt := range corpus.Bigrams {
		_, ok1 := layout.RuneInfo[bi[0]]
		_, ok2 := layout.RuneInfo[bi[1]]
		if ok1 && ok2 {
			bls.relevantBigrams = append(bls.relevantBigrams, BigramCount{
				Bigram: bi,
				Count:  cnt,
			})
		}
	}
	sort.Slice(bls.relevantBigrams, func(i, j int) bool {
		return bls.relevantBigrams[i].Count > bls.relevantBigrams[j].Count
	})

	// Get results from both implementations
	result1 := bls.identifyProblematicKeys(layout)
	result2 := bls.identifyProblematicKeys2(layout)

	// They should have the same length
	if len(result1) != len(result2) {
		t.Errorf("Results have different lengths: original=%d, optimized=%d", len(result1), len(result2))
	}

	// They should contain the same keys (order might differ due to floating point)
	// Convert to sets for comparison
	set1 := make(map[uint8]bool)
	for _, k := range result1 {
		set1[k] = true
	}

	set2 := make(map[uint8]bool)
	for _, k := range result2 {
		set2[k] = true
	}

	// Check if sets are equal
	for k := range set1 {
		if !set2[k] {
			t.Errorf("Key %d in original but not in optimized", k)
		}
	}

	for k := range set2 {
		if !set1[k] {
			t.Errorf("Key %d in optimized but not in original", k)
		}
	}
}
