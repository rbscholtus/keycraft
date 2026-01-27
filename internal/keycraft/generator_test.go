package keycraft

import (
	"strings"
	"testing"
)

// Embedded test configurations
const (
	testConfigSimple = `# Simple test config: fixed + random positions only
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 e a i o ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
`

	testConfigWithGroups = `# Test config with groups
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 2 2 2 2 0  0 e a h i ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 1  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
set1=s
set2=tnrd
`

	testConfigMultiGroups = `# Test config with multiple groups (for permutation testing)
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 2 2 0 0 0  0 e a i o ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 1  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
set1=tn
set2=sr
`

	testConfigMinimal = `# Minimal valid config - all fixed positions except one random
colstag
~ q w f p b  j l u y ; ~
~ a r s t g  m n e i o ~
~ z x c d v  k h , . ' ~
      ~ ~ /  _ ~ ~

charset=qwfpbjluy;arstgmneioxcdvkh,.'/_
`

	testConfigOverlapping = `# Test config with overlapping groups (same chars in multiple sets)
# This should either fail validation or produce layouts without duplicates
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 2 2 0 0 0  0 e a i o ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 1  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
set1=tn
set2=nr
`

	testConfigLargeGroups = `# Large test case for benchmarking permutation generation
# set1: 2 chars for 1 position = P(2,1) = 2
# set2: 5 chars for 4 positions = P(5,4) = 120
# set3: 4 chars for 2 positions = P(4,2) = 12
# Total: 2 * 120 * 12 = 2880 permutations
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 2 2 2 2 0  0 e a i o ~
~ 3 3 0 0 0  0 0 0 0 0 ~
      ~ ~ 1  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
set1=tn
set2=shrdz
set3=lcuq
`

	testConfigInvalidType = `# Invalid layout type
invalid
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
`

	testConfigInvalidPositions = `# Wrong position count (missing positions)
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _

charset=etaoinshrdlcumwfgypbvkjxqz,./;'
`

	testConfigInvalidCharset = `# Duplicate character in charset
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'ee
`

	testConfigInvalidGroups = `# Group used but setN not defined
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 2 2 2 2 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _ ~ ~

charset=etaoinshrdlcumwfgypbvkjxqz ,./;'
`

	testConfigInvalidCounts = `# Character count mismatch (too few chars)
colstag
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
~ 0 0 0 0 0  0 0 0 0 0 ~
      ~ ~ 0  _ ~ ~

charset=etaoinshrd
`
)

// ============================================================================
// Parser Tests
// ============================================================================

func TestParseConfigString_ValidSimple(t *testing.T) {
	config, err := ParseConfigString(testConfigSimple)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	if config.LayoutType != COLSTAG {
		t.Errorf("expected layout type COLSTAG, got %v", config.LayoutType)
	}

	if len(config.Charset) != 32 {
		t.Errorf("expected charset length 32, got %d", len(config.Charset))
	}

	// Check that fixed characters are parsed correctly
	fixedCount := 0
	for _, spec := range config.Template {
		if spec.Type == PositionFixed {
			fixedCount++
		}
	}
	// e, a, i, o, space = 5 fixed positions
	if fixedCount != 5 {
		t.Errorf("expected 5 fixed positions, got %d", fixedCount)
	}
}

func TestParseConfigString_ValidWithGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Check groups are parsed
	if len(config.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(config.Groups))
	}

	if _, ok := config.Groups[1]; !ok {
		t.Error("expected group 1 to be defined")
	}

	if _, ok := config.Groups[2]; !ok {
		t.Error("expected group 2 to be defined")
	}

	// Check group contents
	if len(config.Groups[1]) != 1 || config.Groups[1][0] != 's' {
		t.Errorf("expected group 1 to contain 's', got %v", config.Groups[1])
	}

	if len(config.Groups[2]) != 4 {
		t.Errorf("expected group 2 to have 4 chars, got %d", len(config.Groups[2]))
	}
}

func TestParseConfigString_InvalidType(t *testing.T) {
	_, err := ParseConfigString(testConfigInvalidType)
	if err == nil {
		t.Fatal("expected error for invalid layout type")
	}
	if !strings.Contains(err.Error(), "invalid layout type") {
		t.Errorf("expected 'invalid layout type' error, got: %v", err)
	}
}

func TestParseConfigString_InvalidPositions(t *testing.T) {
	_, err := ParseConfigString(testConfigInvalidPositions)
	if err == nil {
		t.Fatal("expected error for invalid position count")
	}
	if !strings.Contains(err.Error(), "expected") && !strings.Contains(err.Error(), "positions") {
		t.Errorf("expected position count error, got: %v", err)
	}
}

func TestParseLayoutType(t *testing.T) {
	tests := []struct {
		input    string
		expected LayoutType
		ok       bool
	}{
		{"rowstag", ROWSTAG, true},
		{"ROWSTAG", ROWSTAG, true},
		{"anglemod", ANGLEMOD, true},
		{"ortho", ORTHO, true},
		{"colstag", COLSTAG, true},
		{"invalid", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		lt, ok := parseLayoutType(tt.input)
		if ok != tt.ok {
			t.Errorf("parseLayoutType(%q): expected ok=%v, got ok=%v", tt.input, tt.ok, ok)
		}
		if ok && lt != tt.expected {
			t.Errorf("parseLayoutType(%q): expected %v, got %v", tt.input, tt.expected, lt)
		}
	}
}

func TestParsePositionToken(t *testing.T) {
	tests := []struct {
		token    string
		expected PositionSpec
		hasError bool
	}{
		{"~", PositionSpec{Type: PositionUnused}, false},
		{"_", PositionSpec{Type: PositionFixed, FixedChar: ' '}, false},
		{"0", PositionSpec{Type: PositionRandom}, false},
		{"1", PositionSpec{Type: PositionGroup, GroupNum: 1}, false},
		{"9", PositionSpec{Type: PositionGroup, GroupNum: 9}, false},
		{"a", PositionSpec{Type: PositionFixed, FixedChar: 'a'}, false},
		{"e", PositionSpec{Type: PositionFixed, FixedChar: 'e'}, false},
		{",", PositionSpec{Type: PositionFixed, FixedChar: ','}, false},
		{".", PositionSpec{Type: PositionFixed, FixedChar: '.'}, false},
		{"'", PositionSpec{Type: PositionFixed, FixedChar: '\''}, false},
		{";", PositionSpec{Type: PositionFixed, FixedChar: ';'}, false},
		{"~~", PositionSpec{Type: PositionFixed, FixedChar: '~'}, false},
		{"__", PositionSpec{Type: PositionFixed, FixedChar: '_'}, false},
		{"##", PositionSpec{Type: PositionFixed, FixedChar: '#'}, false},
		{"", PositionSpec{}, true},
		{"abc", PositionSpec{}, true},
	}

	for _, tt := range tests {
		spec, err := parsePositionToken(tt.token)
		if tt.hasError {
			if err == nil {
				t.Errorf("parsePositionToken(%q): expected error, got nil", tt.token)
			}
		} else {
			if err != nil {
				t.Errorf("parsePositionToken(%q): unexpected error: %v", tt.token, err)
			}
			if spec != tt.expected {
				t.Errorf("parsePositionToken(%q): expected %+v, got %+v", tt.token, tt.expected, spec)
			}
		}
	}
}

func TestParseCharsetValue(t *testing.T) {
	// Test that _ is converted to space
	result := parseCharsetValue("abc_def")
	expected := []rune{'a', 'b', 'c', ' ', 'd', 'e', 'f'}

	if len(result) != len(expected) {
		t.Errorf("expected length %d, got %d", len(expected), len(result))
	}

	for i, r := range expected {
		if result[i] != r {
			t.Errorf("index %d: expected %q, got %q", i, r, result[i])
		}
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateConfig_Valid(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	err = ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig failed on valid config: %v", err)
	}
}

func TestValidateConfig_DuplicateCharset(t *testing.T) {
	config, err := ParseConfigString(testConfigInvalidCharset)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	err = ValidateConfig(config)
	if err == nil {
		t.Error("expected validation error for duplicate charset")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %v", err)
	}
}

func TestValidateConfig_MissingGroup(t *testing.T) {
	config, err := ParseConfigString(testConfigInvalidGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	err = ValidateConfig(config)
	if err == nil {
		t.Error("expected validation error for missing group definition")
	}
	if !strings.Contains(err.Error(), "not defined") {
		t.Errorf("expected 'not defined' in error, got: %v", err)
	}
}

func TestValidateConfig_CharCountMismatch(t *testing.T) {
	config, err := ParseConfigString(testConfigInvalidCounts)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	err = ValidateConfig(config)
	if err == nil {
		t.Error("expected validation error for character count mismatch")
	}
	if !strings.Contains(err.Error(), "charset has") {
		t.Errorf("expected charset count error, got: %v", err)
	}
}

// ============================================================================
// Permutation Tests
// ============================================================================

func TestGenerateOrderedSelections(t *testing.T) {
	tests := []struct {
		chars    []rune
		k        int
		expected int // P(n,k)
	}{
		{[]rune{'a', 'b', 'c'}, 2, 6},            // P(3,2) = 6
		{[]rune{'a', 'b', 'c', 'd'}, 4, 24},      // P(4,4) = 4! = 24
		{[]rune{'a', 'b', 'c', 'd'}, 2, 12},      // P(4,2) = 12
		{[]rune{'a', 'b', 'c', 'd', 'e'}, 3, 60}, // P(5,3) = 60
		{[]rune{'a'}, 1, 1},                      // P(1,1) = 1
		{[]rune{'a', 'b'}, 0, 1},                 // P(n,0) = 1 (empty selection)
	}

	for _, tt := range tests {
		result := generateOrderedSelections(tt.chars, tt.k)
		if len(result) != tt.expected {
			t.Errorf("generateOrderedSelections(%v, %d): expected %d permutations, got %d",
				string(tt.chars), tt.k, tt.expected, len(result))
		}

		// Verify all permutations are unique
		seen := make(map[string]bool)
		for _, perm := range result {
			key := string(perm)
			if seen[key] {
				t.Errorf("duplicate permutation: %s", key)
			}
			seen[key] = true
		}
	}
}

func TestGeneratePermutations_NoGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigSimple)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, warnings := GeneratePermutations(config, 0)

	if total != 1 {
		t.Errorf("expected 1 total permutation (no groups), got %d", total)
	}

	if len(perms) != 1 {
		t.Errorf("expected 1 permutation, got %d", len(perms))
	}

	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestGeneratePermutations_WithGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, _ := GeneratePermutations(config, 0)

	// set1=s (1 char for 1 pos) = P(1,1) = 1
	// set2=tnrd (4 chars for 4 pos) = P(4,4) = 24
	// Total = 1 * 24 = 24
	expectedTotal := 24
	if total != expectedTotal {
		t.Errorf("expected %d total permutations, got %d", expectedTotal, total)
	}

	if len(perms) != expectedTotal {
		t.Errorf("expected %d permutations, got %d", expectedTotal, len(perms))
	}
}

func TestGeneratePermutations_MultiGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigMultiGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, _ := GeneratePermutations(config, 0)

	// set1=tn (2 chars for 1 pos) = P(2,1) = 2
	// set2=sr (2 chars for 2 pos) = P(2,2) = 2
	// Total = 2 * 2 = 4
	expectedTotal := 4
	if total != expectedTotal {
		t.Errorf("expected %d total permutations, got %d", expectedTotal, total)
	}

	if len(perms) != expectedTotal {
		t.Errorf("expected %d permutations, got %d", expectedTotal, len(perms))
	}
}

func TestGeneratePermutations_OverlappingGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigOverlapping)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, warnings := GeneratePermutations(config, 0)

	// set1=tn (2 chars for 1 pos) = P(2,1) = 2 choices: [t], [n]
	// set2=nr (2 chars for 2 pos) = P(2,2) = 2 choices: [n,r], [r,n]
	// But 'n' is shared, so:
	// - [t] + [n,r] = valid (t,n,r)
	// - [t] + [r,n] = valid (t,r,n)
	// - [n] + [n,r] = INVALID (n appears twice)
	// - [n] + [r,n] = INVALID (n appears twice)
	// After removing 'n' from set2 when set1 uses 'n':
	// - [n] + only 'r' available, but need 2 positions = INVALID (not enough chars)
	// So only 2 valid permutations
	expectedTotal := 2
	if total != expectedTotal {
		t.Errorf("expected %d total permutations, got %d", expectedTotal, total)
	}

	if len(perms) != expectedTotal {
		t.Errorf("expected %d permutations, got %d", expectedTotal, len(perms))
	}

	// Should have warning about overlapping groups
	hasOverlapWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "share characters") {
			hasOverlapWarning = true
			break
		}
	}
	if !hasOverlapWarning {
		t.Error("expected warning about overlapping groups")
	}

	// Verify the actual permutations are correct
	// Both should have 't' in group 1
	for i, perm := range perms {
		if len(perm[1]) != 1 || perm[1][0] != 't' {
			t.Errorf("permutation %d: expected group 1 to be [t], got %v", i, string(perm[1]))
		}
		// Group 2 should have 'n' and 'r' in some order
		g2 := string(perm[2])
		if g2 != "nr" && g2 != "rn" {
			t.Errorf("permutation %d: expected group 2 to be [n,r] or [r,n], got %v", i, g2)
		}
	}
}

func TestGeneratePermutations_MaxLayoutsLimit(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	maxLayouts := 5
	perms, total, warnings := GeneratePermutations(config, maxLayouts)

	// Total should still be 24
	if total != 24 {
		t.Errorf("expected 24 total permutations, got %d", total)
	}

	// But only 5 should be generated
	if len(perms) != maxLayouts {
		t.Errorf("expected %d permutations with limit, got %d", maxLayouts, len(perms))
	}

	// Should have warning about limiting
	hasLimitWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "limiting") {
			hasLimitWarning = true
			break
		}
	}
	if !hasLimitWarning {
		t.Error("expected warning about limiting layouts")
	}
}

// ============================================================================
// Layout Generation Tests
// ============================================================================

// assertNoDuplicateCharacters checks that a layout has no duplicate characters
// in used positions (non-zero runes).
func assertNoDuplicateCharacters(t *testing.T, layout *SplitLayout) {
	t.Helper()
	seen := make(map[rune]int) // rune -> first position
	for i, r := range layout.Runes {
		if r == 0 { // Skip unused positions
			continue
		}
		if firstPos, exists := seen[r]; exists {
			t.Errorf("duplicate character %q at positions %d and %d", r, firstPos, i)
		}
		seen[r] = i
	}
}

func TestGenerateLayout_Simple(t *testing.T) {
	config, err := ParseConfigString(testConfigSimple)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Generate with empty group permutation (no groups)
	layout := GenerateLayout(config, map[int][]rune{}, 42, 0)

	if layout == nil {
		t.Fatal("GenerateLayout returned nil")
	}

	// Check no duplicate characters
	assertNoDuplicateCharacters(t, layout)

	// Check fixed positions
	// Position 19 should be 'e', 20 should be 'a', 21 should be 'i', 22 should be 'o'
	// Position 39 should be space
	if layout.Runes[19] != 'e' {
		t.Errorf("expected 'e' at position 19, got %q", layout.Runes[19])
	}
	if layout.Runes[20] != 'a' {
		t.Errorf("expected 'a' at position 20, got %q", layout.Runes[20])
	}
	if layout.Runes[21] != 'i' {
		t.Errorf("expected 'i' at position 21, got %q", layout.Runes[21])
	}
	if layout.Runes[22] != 'o' {
		t.Errorf("expected 'o' at position 22, got %q", layout.Runes[22])
	}
	if layout.Runes[39] != ' ' {
		t.Errorf("expected space at position 39, got %q", layout.Runes[39])
	}

	// Check unused positions
	if layout.Runes[0] != 0 {
		t.Errorf("expected unused (0) at position 0, got %q", layout.Runes[0])
	}
}

func TestGenerateLayout_WithGroups(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Create a specific group permutation
	groupPerm := map[int][]rune{
		1: {'s'},
		2: {'t', 'n', 'r', 'd'},
	}

	layout := GenerateLayout(config, groupPerm, 42, 0)

	if layout == nil {
		t.Fatal("GenerateLayout returned nil")
	}

	// Check no duplicate characters
	assertNoDuplicateCharacters(t, layout)

	// Check group 1 position (thumb position 38)
	if layout.Runes[38] != 's' {
		t.Errorf("expected 's' at group 1 position, got %q", layout.Runes[38])
	}

	// Check group 2 positions (13, 14, 15, 16)
	if layout.Runes[13] != 't' {
		t.Errorf("expected 't' at position 13, got %q", layout.Runes[13])
	}
	if layout.Runes[14] != 'n' {
		t.Errorf("expected 'n' at position 14, got %q", layout.Runes[14])
	}
	if layout.Runes[15] != 'r' {
		t.Errorf("expected 'r' at position 15, got %q", layout.Runes[15])
	}
	if layout.Runes[16] != 'd' {
		t.Errorf("expected 'd' at position 16, got %q", layout.Runes[16])
	}
}

func TestGenerateLayout_SeedReproducibility(t *testing.T) {
	config, err := ParseConfigString(testConfigSimple)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Generate two layouts with the same seed
	layout1 := GenerateLayout(config, map[int][]rune{}, 42, 0)
	layout2 := GenerateLayout(config, map[int][]rune{}, 42, 0)

	// Check no duplicate characters
	assertNoDuplicateCharacters(t, layout1)
	assertNoDuplicateCharacters(t, layout2)

	// They should be identical
	for i := range layout1.Runes {
		if layout1.Runes[i] != layout2.Runes[i] {
			t.Errorf("layouts differ at position %d: %q vs %q", i, layout1.Runes[i], layout2.Runes[i])
		}
	}
}

func TestGenerateLayout_DifferentSeeds(t *testing.T) {
	config, err := ParseConfigString(testConfigSimple)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Generate layouts with different seeds
	layout1 := GenerateLayout(config, map[int][]rune{}, 42, 0)
	layout2 := GenerateLayout(config, map[int][]rune{}, 43, 0)

	// Check no duplicate characters
	assertNoDuplicateCharacters(t, layout1)
	assertNoDuplicateCharacters(t, layout2)

	// Random positions should differ (at least some)
	// Fixed positions should be the same
	if layout1.Runes[19] != layout2.Runes[19] { // 'e' is fixed
		t.Error("fixed positions should be the same")
	}

	// Count differences in random positions
	diffCount := 0
	for i := range layout1.Runes {
		if config.Template[i].Type == PositionRandom && layout1.Runes[i] != layout2.Runes[i] {
			diffCount++
		}
	}

	if diffCount == 0 {
		t.Error("expected some differences in random positions with different seeds")
	}
}

func TestGenerateLayout_AllPermutationsNoDuplicates(t *testing.T) {
	// Test with-groups.gen: all permutations should have no duplicates
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, _ := GeneratePermutations(config, 0)
	if total == 0 {
		t.Fatal("expected some permutations")
	}

	for i, perm := range perms {
		layout := GenerateLayout(config, perm, 42, i)
		if layout == nil {
			t.Fatalf("GenerateLayout returned nil for permutation %d", i)
		}
		assertNoDuplicateCharacters(t, layout)
	}
}

func TestGenerateLayout_MultiGroupsAllPermutationsNoDuplicates(t *testing.T) {
	// Test multi-groups.gen: all permutations should have no duplicates
	config, err := ParseConfigString(testConfigMultiGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, _ := GeneratePermutations(config, 0)
	if total == 0 {
		t.Fatal("expected some permutations")
	}

	for i, perm := range perms {
		layout := GenerateLayout(config, perm, 42, i)
		if layout == nil {
			t.Fatalf("GenerateLayout returned nil for permutation %d", i)
		}
		assertNoDuplicateCharacters(t, layout)
	}
}

func TestGenerateLayout_OverlappingGroupsNoDuplicates(t *testing.T) {
	// Test overlapping-groups.gen: groups share characters (n is in both set1 and set2)
	// All permutations should still have no duplicates
	config, err := ParseConfigString(testConfigOverlapping)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	perms, total, _ := GeneratePermutations(config, 0)
	if total == 0 {
		t.Fatal("expected some permutations")
	}

	// This test will FAIL until the cross-group duplicate issue is fixed
	// When set1=tn and set2=nr, the permutation (n, nr) would have 'n' twice
	for i, perm := range perms {
		layout := GenerateLayout(config, perm, 42, i)
		if layout == nil {
			t.Fatalf("GenerateLayout returned nil for permutation %d", i)
		}
		assertNoDuplicateCharacters(t, layout)
	}
}

// ============================================================================
// Name Generation Tests
// ============================================================================

func TestGenerateLayoutNameFromRunes(t *testing.T) {
	// Create a runes array with known values at home row and thumb positions
	var runes [42]rune
	// Home row: positions 13-16, 19-22
	runes[13] = 't'
	runes[14] = 'n'
	runes[15] = 'r'
	runes[16] = 'd'
	runes[19] = 'e'
	runes[20] = 'a'
	runes[21] = 'i'
	runes[22] = 'o'
	// Thumbs: positions 36-41
	runes[36] = 'x'
	runes[37] = 'y'
	runes[38] = 's'
	runes[39] = ' ' // space, not a-z, should be skipped
	runes[40] = 'z'
	runes[41] = 'q'

	name := generateLayoutNameFromRunes(runes, 0x002f)

	// Name should start with _
	if !strings.HasPrefix(name, "_") {
		t.Errorf("expected name to start with _, got %s", name)
	}

	// Should contain home keys and thumb keys (excluding space)
	if !strings.Contains(name, "tnrdeaio") {
		t.Errorf("expected home keys in name, got %s", name)
	}

	// Should end with hex suffix
	if !strings.HasSuffix(name, "-002f") {
		t.Errorf("expected name to end with -002f, got %s", name)
	}
}

func TestGenerateLayoutNameFromRunes_HexSuffix(t *testing.T) {
	var runes [42]rune
	for i := range runes {
		runes[i] = 'a' + rune(i%26)
	}

	tests := []struct {
		permIndex int
		expected  string
	}{
		{0, "-0000"},
		{1, "-0001"},
		{15, "-000f"},
		{16, "-0010"},
		{255, "-00ff"},
		{4095, "-0fff"},
		{65535, "-ffff"},
	}

	for _, tt := range tests {
		name := generateLayoutNameFromRunes(runes, tt.permIndex)
		if !strings.HasSuffix(name, tt.expected) {
			t.Errorf("permIndex %d: expected suffix %s, got name %s", tt.permIndex, tt.expected, name)
		}
	}
}

// ============================================================================
// Default Pins Tests
// ============================================================================

func TestComputeDefaultPins(t *testing.T) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		t.Fatalf("ParseConfigString failed: %v", err)
	}

	// Generate a layout
	groupPerm := map[int][]rune{
		1: {'s'},
		2: {'t', 'n', 'r', 'd'},
	}
	layout := GenerateLayout(config, groupPerm, 42, 0)

	pins := ComputeDefaultPins(config, layout)

	// Check that unused positions are pinned
	if !pins[0] { // Position 0 is unused (~)
		t.Error("expected unused position 0 to be pinned")
	}

	// Check that fixed positions are pinned
	if !pins[19] { // 'e' is fixed
		t.Error("expected fixed position 19 (e) to be pinned")
	}
	if !pins[39] { // space is fixed
		t.Error("expected fixed position 39 (space) to be pinned")
	}

	// Check that group positions are pinned
	if !pins[13] { // Group 2 position
		t.Error("expected group position 13 to be pinned")
	}
	if !pins[38] { // Group 1 position
		t.Error("expected group position 38 to be pinned")
	}

	// Check that random positions are NOT pinned
	randomPosFound := false
	for i, spec := range config.Template {
		if spec.Type == PositionRandom {
			if pins[i] {
				t.Errorf("expected random position %d to NOT be pinned", i)
			}
			randomPosFound = true
		}
	}
	if !randomPosFound {
		t.Error("no random positions found in config")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGeneratePermutations_SmallGroups(b *testing.B) {
	config, err := ParseConfigString(testConfigWithGroups)
	if err != nil {
		b.Fatalf("ParseConfigString failed: %v", err)
	}

	for b.Loop() {
		GeneratePermutations(config, 0)
	}
}

func BenchmarkGeneratePermutations_LargeGroups(b *testing.B) {
	// 3 groups: P(2,1) * P(5,4) * P(4,2) = 2 * 120 * 12 = 2880 permutations
	config, err := ParseConfigString(testConfigLargeGroups)
	if err != nil {
		b.Fatalf("ParseConfigString failed: %v", err)
	}

	for b.Loop() {
		GeneratePermutations(config, 0)
	}
}

func BenchmarkGeneratePermutations_Overlapping(b *testing.B) {
	config, err := ParseConfigString(testConfigOverlapping)
	if err != nil {
		b.Fatalf("ParseConfigString failed: %v", err)
	}

	for b.Loop() {
		GeneratePermutations(config, 0)
	}
}

func BenchmarkCountConstrainedPerms_LargeGroups(b *testing.B) {
	config, err := ParseConfigString(testConfigLargeGroups)
	if err != nil {
		b.Fatalf("ParseConfigString failed: %v", err)
	}

	// Setup like GeneratePermutations does
	groupNums := []int{1, 2, 3}
	groupPositions := map[int]int{1: 1, 2: 4, 3: 2}

	for b.Loop() {
		countConstrainedPerms(config, groupNums, groupPositions, 0, "")
	}
}
