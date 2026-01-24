package keycraft

import (
	"math/rand/v2"
	"strings"
	"time"
)

// GeneratorInput contains parameters for layout generation.
type GeneratorInput struct {
	LayoutType  LayoutType // ROWSTAG, ANGLEMOD, ORTHO, or COLSTAG
	AlphaThumb  bool       // If true, place alpha character on thumbs (9 chars), else 8
	VowelsRight bool       // If true, ensure vowels are placed on right hand
	Seed        uint64     // Random seed for reproducibility (0 = use timestamp)
}

// DefaultChars contains lowercase English characters sorted by frequency.
// First 26 are letters (frequency-sorted), followed by common punctuation.
// Total: 31 characters (26 letters + 5 punctuation)
const DefaultChars = "etaoinshrdlcumwfgypbvkjxqz,./;'"

// processCharacters selects high-frequency characters from DefaultChars.
// Returns (selected characters, remainder characters).
// No changes needed from the original algorithm - works as-is.
func processCharacters(vowelsRight bool, thumbAlpha bool, rand *rand.Rand) ([]rune, []rune) {
	if rand == nil {
		rand = getRNG(0)
	}

	// 1. Setup Parameters
	poolSize := 13
	resultLength := 8
	if thumbAlpha {
		resultLength = 9
	} else if vowelsRight {
		poolSize = 12
	}

	// 2. Select Primary Keys
	// We only need to look at the first 'poolSize' of DefaultChars
	indices := rand.Perm(poolSize)
	selected := make([]rune, resultLength)

	// A fixed-size array is allocated on the stack, not the heap.
	// We use the rune's value as an index to track what we've picked.
	var isSelected [256]bool

	for i := 0; i < resultLength; i++ {
		char := rune(DefaultChars[indices[i]])
		selected[i] = char
		isSelected[char] = true
	}

	// 3. Partition and Shuffle (Vowels to Right)
	if vowelsRight {
		// In-place partition (Two-pointer)
		left, right := 0, len(selected)-1
		for left <= right {
			if IsVowel(selected[left]) {
				selected[left], selected[right] = selected[right], selected[left]
				right--
			} else {
				left++
			}
		}

		// Shuffle the left segment (consonants)
		lLen := left // number of consonants on the left
		rand.Shuffle(lLen, func(i, j int) {
			selected[i], selected[j] = selected[j], selected[i]
		})

		// Shuffle the right segment (vowels)
		rStart := left
		rLen := len(selected) - left
		rand.Shuffle(rLen, func(i, j int) {
			selected[rStart+i], selected[rStart+j] = selected[rStart+j], selected[rStart+i]
		})
	}

	// 4. Calculate Remainder (Unselected)
	// Pre-allocate with exact capacity to avoid multiple heap reallocations
	remainder := make([]rune, 0, len(DefaultChars)-resultLength)
	for _, r := range DefaultChars {
		if !isSelected[r] {
			remainder = append(remainder, r)
		}
	}

	// Shuffle the leftover pool
	rand.Shuffle(len(remainder), func(i, j int) {
		remainder[i], remainder[j] = remainder[j], remainder[i]
	})

	return selected, remainder
}

// NewRandomLayout generates a new random keyboard layout based on the input parameters.
func NewRandomLayout(input GeneratorInput) (*SplitLayout, error) {
	rand := getRNG(input.Seed)

	// Get selected characters and remainder
	selected, remainder := processCharacters(input.VowelsRight, input.AlphaThumb, rand)

	// Initialize rune array with all zeros
	var runes [42]rune

	// Place home row characters
	// Left hand home row: positions 13-16 (4 chars)
	runes[13] = selected[0]
	runes[14] = selected[1]
	runes[15] = selected[2]
	runes[16] = selected[3]

	// Right hand home row: positions 19-22 (4 chars)
	runes[19] = selected[4]
	runes[20] = selected[5]
	runes[21] = selected[6]
	runes[22] = selected[7]

	// Fixed assignment: position 39 = space
	runes[39] = ' '

	// Left thumb: position 38 (1 char) - only if thumbAlpha
	if input.AlphaThumb {
		char := selected[8]

		// Special vowel handling: if the 9th character is a vowel, put it at position 39
		// and put the space at position 38
		if IsVowel(char) {
			runes[39] = char // vowel goes to position 39
			runes[38] = ' '  // space goes to position 38
		} else {
			runes[38] = char // non-vowel goes to position 38, space stays at 39
		}
	}

	// Fill remaining positions with remainder characters
	// Positions to fill in order: 1-10, 17-18, 25-34, 23 (if needed)
	fillPositions := []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, // top row center (10)
		17, 18, // home row middle (2)
		25, 26, 27, 28, 29, 30, 31, 32, 33, 34, // bottom row center (10)
		23, // only if 23 remainder chars
	}

	remainderIdx := 0
	for _, pos := range fillPositions {
		if remainderIdx >= len(remainder) {
			break
		}
		runes[pos] = remainder[remainderIdx]
		remainderIdx++
	}

	// Create the layout using NewSplitLayout (which precalculates all data structures)
	layout := NewSplitLayout("", input.LayoutType, runes)

	// Generate and set the layout name
	layout.Name = layout.generateLayoutName()

	return layout, nil
}

// homeThumbChars creates an auto-generated string: <chars>
// Extracts lowercase a-z characters from positions 13-16, 19-22, 36-41.
func (sl *SplitLayout) HomeThumbChars() string {
	var b strings.Builder
	b.Grow(14)

	for _, i := range [14]int{13, 14, 15, 16, 19, 20, 21, 22, 36, 37, 38, 39, 40, 41} {
		if r := sl.Runes[i]; 'a' <= r && r <= 'z' {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// generateLayoutName creates an auto-generated name: <chars>-<random>
// Extracts lowercase a-z characters from positions 13-16, 19-22, 36-41.
// Generates a random hexadecimal suffix based on UnixMillis
func (sl *SplitLayout) generateLayoutName() string {
	var b strings.Builder
	b.Grow(19)

	for _, i := range [14]int{13, 14, 15, 16, 19, 20, 21, 22, 36, 37, 38, 39, 40, 41} {
		if r := sl.Runes[i]; 'a' <= r && r <= 'z' {
			b.WriteRune(r)
		}
	}

	b.WriteByte('-')

	now := time.Now().UnixMilli()
	suffix := uint64(now & 0xFFFF)

	// Write exactly 4 hex digits (fast path – no string alloc for padding)
	const hex = "0123456789abcdef"
	for i := 3; i >= 0; i-- {
		b.WriteByte(hex[(suffix>>(i*4))&0xF])
	}

	return b.String()
}
