package keycraft

import (
	"fmt"
	"testing"
)

// Test helper: check if slice has unique elements
func isUnique(slice []rune) bool {
	seen := make(map[rune]bool)
	for _, r := range slice {
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	return true
}

// TestIsVowel verifies the basic logic for identifying vowels.
func TestIsVowel(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'A', true}, {'E', true}, {'I', true}, {'O', true}, {'U', true},
		{'a', true}, {'e', true}, {'i', true}, {'o', true}, {'u', true},
		{'B', false}, {'Z', false}, {'T', false}, {'N', false},
		{'x', false}, {'q', false}, {'y', false},
		{' ', false}, {'!', false}, {'1', false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Char_%c", tt.char), func(t *testing.T) {
			got := IsVowel(tt.char)
			if got != tt.expected {
				t.Errorf("IsVowel(%c): expected %v, got %v", tt.char, tt.expected, got)
			}
		})
	}
}

// TestConstraints verifies correct output lengths based on parameters.
func TestConstraints(t *testing.T) {
	tests := []struct {
		vRight bool
		tAlpha bool
		want   int
	}{
		{true, true, 9},
		{false, true, 9},
		{true, false, 8},
		{false, false, 8},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("vRight=%v,tAlpha=%v", tt.vRight, tt.tAlpha)
		t.Run(name, func(t *testing.T) {
			result, _ := processCharacters(tt.vRight, tt.tAlpha, nil)
			if len(result) != tt.want {
				t.Errorf("Expected length %d, got %d", tt.want, len(result))
			}
			if !isUnique(result) {
				t.Error("Result contains duplicate characters")
			}
		})
	}
}

// Test1000RandomizedRuns performs 1000 iterations per configuration.
func Test1000RandomizedRuns(t *testing.T) {
	configs := []struct {
		vRight bool
		tAlpha bool
		length int
	}{
		{true, true, 9},
		{false, true, 9},
		{true, false, 8},
		{false, false, 8},
	}

	for _, cfg := range configs {
		name := fmt.Sprintf("vRight:%v/tAlpha:%v", cfg.vRight, cfg.tAlpha)
		t.Run(name, func(t *testing.T) {
			for i := range 1000 {
				result, _ := processCharacters(cfg.vRight, cfg.tAlpha, nil)

				if len(result) != cfg.length {
					t.Fatalf("Run %d: Expected length %d, got %d", i, cfg.length, len(result))
				}

				if !isUnique(result) {
					t.Fatalf("Run %d: Duplicates found in %s", i, string(result))
				}

				if cfg.vRight {
					for j := range 4 {
						if IsVowel(result[j]) {
							t.Fatalf("Run %d: Found vowel %c in left-hand side", i, result[j])
						}
					}
				}
			}
		})
	}
}

// calculateChiSquare helper for distribution testing
func calculateChiSquare(freqs map[rune]int, expected float64) float64 {
	chiSquare := 0.0
	for _, observed := range freqs {
		diff := float64(observed) - expected
		chiSquare += (diff * diff) / expected
	}
	return chiSquare
}

// TestCharacterDistribution validates uniform distribution with Chi-Square test.
func TestCharacterDistribution(t *testing.T) {
	const iterations = 20000
	configs := []struct {
		vRight bool
		tAlpha bool
		pool   int
	}{
		{true, true, 13},
		{false, true, 13},
		{true, false, 12},
		{false, false, 13},
	}

	criticalValues := map[int]float64{
		11: 19.68,
		12: 21.03,
	}

	for _, cfg := range configs {
		name := fmt.Sprintf("Stats-vRight:%v-tAlpha:%v", cfg.vRight, cfg.tAlpha)
		t.Run(name, func(t *testing.T) {
			freq := make(map[rune]int)
			length := 8
			if cfg.tAlpha {
				length = 9
			}
			// Create a single RNG and reuse it to avoid time-based seed collision issues
			rng := getRNG(0)

			for range iterations {
				res, _ := processCharacters(cfg.vRight, cfg.tAlpha, rng)
				for _, r := range res {
					freq[r]++
				}
			}

			if cfg.vRight && !cfg.tAlpha {
				if freq['u'] > 0 {
					t.Errorf("Constraint Violation: 'u' appeared in excluded-vowel mode")
				}
			}

			expected := float64(iterations*length) / float64(cfg.pool)
			chiStat := calculateChiSquare(freq, expected)
			df := cfg.pool - 1
			limit := criticalValues[df]

			if chiStat > limit {
				t.Errorf("%s failed Chi-Square test: stat: %.2f > limit: %.2f", name, chiStat, limit)
			}
		})
	}
}

// TestInternalOrderRandomness verifies sufficient entropy in ordering.
func TestInternalOrderRandomness(t *testing.T) {
	orders := make(map[string]int)
	iterations := 1000
	// Create a single RNG and reuse it to avoid time-based seed collision issues
	rng := getRNG(0)

	for range iterations {
		result, _ := processCharacters(true, true, rng)
		firstFour := string(result[:4])
		orders[firstFour]++
	}

	if len(orders) < (iterations / 2) {
		t.Errorf("Low entropy: only saw %d unique orderings", len(orders))
	}
}

// TestRemainderIntegrity validates no overlap between selected and remainder.
func TestRemainderIntegrity(t *testing.T) {
	selected, remainder := processCharacters(true, true, nil)

	totalLen := len(selected) + len(remainder)
	if totalLen != len(DefaultChars) {
		t.Errorf("Total length mismatch: expected %d, got %d", len(DefaultChars), totalLen)
	}

	seen := make(map[rune]bool)
	for _, r := range selected {
		seen[r] = true
	}

	for _, r := range remainder {
		if seen[r] {
			t.Errorf("Overlap found: character %c", r)
		}
	}

	if !isUnique(remainder) {
		t.Error("Remainder contains duplicates")
	}
}

// TestFullSetCoverage ensures all characters appear exactly once.
func TestFullSetCoverage(t *testing.T) {
	scenarios := []struct {
		vRight bool
		tAlpha bool
	}{
		{true, true},
		{false, true},
		{true, false},
		{false, false},
	}

	for _, s := range scenarios {
		name := fmt.Sprintf("vRight:%v,tAlpha:%v", s.vRight, s.tAlpha)
		t.Run(name, func(t *testing.T) {
			selected, remainder := processCharacters(s.vRight, s.tAlpha, nil)

			counts := make(map[rune]int)
			for _, r := range selected {
				counts[r]++
			}
			for _, r := range remainder {
				counts[r]++
			}

			for _, expectedChar := range DefaultChars {
				actualCount, exists := counts[expectedChar]
				if !exists {
					t.Errorf("%s: Character '%c' missing from both Result and Remainder", name, expectedChar)
				}
				if actualCount > 1 {
					t.Errorf("%s: Character '%c' appeared %d times (should be unique)", name, expectedChar, actualCount)
				}
			}

			expectedLen := len(DefaultChars)
			actualLen := len(selected) + len(remainder)
			if actualLen != expectedLen {
				t.Errorf("%s: Combined length mismatch. Expected %d, got %d", name, expectedLen, actualLen)
			}

			if len(counts) != expectedLen {
				t.Errorf("%s: Found %d unique characters, but DefaultChars only has %d", name, len(counts), expectedLen)
			}
		})
	}
}

// TestNewRandomLayout verifies the full layout generation process.
func TestNewRandomLayout(t *testing.T) {
	configs := []struct {
		name        string
		alphaThumb  bool
		vowelsRight bool
		layoutType  LayoutType
	}{
		{"rowstag-no-thumb", false, false, ROWSTAG},
		{"rowstag-thumb", true, false, ROWSTAG},
		{"rowstag-vowels-right", false, true, ROWSTAG},
		{"rowstag-thumb-vowels", true, true, ROWSTAG},
		{"ortho-no-thumb", false, false, ORTHO},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			input := GeneratorInput{
				LayoutType:  cfg.layoutType,
				AlphaThumb:  cfg.alphaThumb,
				VowelsRight: cfg.vowelsRight,
				Seed:        42, // Fixed seed for reproducibility
			}

			layout, err := NewRandomLayout(input)
			if err != nil {
				t.Fatalf("NewRandomLayout failed: %v", err)
			}

			// Verify layout name is set
			if layout.Name == "" {
				t.Error("Layout name is empty")
			}

			// Verify layout type matches
			if layout.LayoutType != cfg.layoutType {
				t.Errorf("Layout type mismatch: expected %v, got %v", cfg.layoutType, layout.LayoutType)
			}

			// Count non-zero positions
			nonZeroCount := 0
			for _, r := range layout.Runes {
				if r != 0 {
					nonZeroCount++
				}
			}

			// Should have exactly 32 non-zero positions (31 from DefaultChars + 1 space)
			if nonZeroCount != 32 {
				t.Errorf("Expected 32 non-zero positions, got %d", nonZeroCount)
			}

			// Verify space is at position 38 or 39
			hasSpace := layout.Runes[38] == ' ' || layout.Runes[39] == ' '
			if !hasSpace {
				t.Error("Space should be at position 38 or 39")
			}

			// Verify home row has letters
			for i := 13; i <= 16; i++ {
				r := layout.Runes[i]
				if r < 'a' || r > 'z' {
					t.Errorf("Position %d should have lowercase letter, got %c", i, r)
				}
			}
			for i := 19; i <= 22; i++ {
				r := layout.Runes[i]
				if r < 'a' || r > 'z' {
					t.Errorf("Position %d should have lowercase letter, got %c", i, r)
				}
			}

			// If alphaThumb, verify position 38 has a character
			if cfg.alphaThumb {
				r := layout.Runes[38]
				if r == 0 {
					t.Error("Position 38 should have a character when alphaThumb=true")
				}
				// Verify no vowel at position 38 (should have been swapped to 39)
				if IsVowel(r) {
					t.Errorf("Position 38 should not have vowel when alphaThumb=true, got %c", r)
				}
			}

			// Verify vowels constraint if vowelsRight
			if cfg.vowelsRight {
				// Positions 13-16 should be consonants
				for i := 13; i <= 16; i++ {
					r := layout.Runes[i]
					if IsVowel(r) {
						t.Errorf("Position %d should be consonant when vowelsRight=true, got %c", i, r)
					}
				}
			}

			// Verify RuneInfo map is populated
			if len(layout.RuneInfo) < 10 {
				t.Errorf("RuneInfo map seems incomplete: only %d entries", len(layout.RuneInfo))
			}

			// Verify caches are initialized
			if layout.KeyPairDistances == nil {
				t.Error("KeyPairDistances not initialized")
			}
			if len(layout.SFBs) == 0 {
				t.Error("SFBs cache not initialized")
			}
		})
	}
}

// Benchmarks
func BenchmarkProcessCharacters_VRightTrue_TAlphaTrue(b *testing.B) {
	for b.Loop() {
		_, _ = processCharacters(true, true, nil)
	}
}

func BenchmarkProcessCharacters_VRightTrue_TAlphaFalse(b *testing.B) {
	for b.Loop() {
		_, _ = processCharacters(true, false, nil)
	}
}

func BenchmarkProcessCharacters_VRightFalse_TAlphaTrue(b *testing.B) {
	for b.Loop() {
		_, _ = processCharacters(false, true, nil)
	}
}

func BenchmarkProcessCharacters_VRightFalse_TAlphaFalse(b *testing.B) {
	for b.Loop() {
		_, _ = processCharacters(false, false, nil)
	}
}

func BenchmarkNewRandomLayout(b *testing.B) {
	input := GeneratorInput{
		LayoutType:  ROWSTAG,
		AlphaThumb:  false,
		VowelsRight: false,
		Seed:        42,
	}

	for b.Loop() {
		_, _ = NewRandomLayout(input)
	}
}
