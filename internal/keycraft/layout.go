// Package keycraft provides ergonomic analysis tools for keyboard layouts.
//
// This package supports various keyboard geometries (row-staggered, ortholinear,
// column-staggered, angle-mod) and computes ergonomic metrics based on n-gram
// frequencies from text corpora. Metrics include same-finger patterns, lateral stretches,
// scissors, alternations, rolls, and redirections. See README.md for full metric descriptions.
package keycraft

import (
	"bufio"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
)

// Finger constants representing fingers 0-9.
// LP = left pinky, LR = left ring, LM = left middle, LI = left index, LT = left thumb.
// RT = right thumb, RI = right index, RM = right middle, RR = right ring, RP = right pinky.
const (
	LP uint8 = iota // 0: left pinky
	LR              // 1: left ring
	LM              // 2: left middle
	LI              // 3: left index
	LT              // 4: left thumb

	RT // 5: right thumb
	RI // 6: right index
	RM // 7: right middle
	RR // 8: right ring
	RP // 9: right pinky
)

// keyToFinger maps each of the 42 key positions to the finger that types it.
// Layout: 3 rows of 12 keys (6 left, 6 right), plus 1 row of 6 thumb keys (3 left, 3 right).
var keyToFinger = [...]uint8{
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP, // Row 0
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP, // Row 1
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP, // Row 2
	LT, LT, LT, RT, RT, RT, // Row 3 (thumbs)
}

// angleModKeyToFinger maps key positions to fingers for angle-mod layouts.
// In angle-mod, the bottom-left key shifts to accommodate the hand's natural angle.
var angleModKeyToFinger = [...]uint8{
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP, // Row 0
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP, // Row 1
	LP, LR, LM, LI, LI, LI, RI, RI, RM, RR, RP, RP, // Row 2 (angle-mod difference here)
	LT, LT, LT, RT, RT, RT, // Row 3 (thumbs)
}

// LayoutType represents the physical geometry of a keyboard.
type LayoutType uint8

const (
	ROWSTAG  LayoutType = iota // Row-staggered (traditional)
	ANGLEMOD                   // Angle-mod (row-staggered with bottom-left adjustment)
	ORTHO                      // Ortholinear (grid layout)
	COLSTAG                    // Column-staggered (ergonomic)
)

// layoutTypeStrings maps LayoutType constants to their string representations.
var layoutTypeStrings = map[LayoutType]string{
	ROWSTAG:  "rowstag",
	ANGLEMOD: "anglemod",
	ORTHO:    "ortho",
	COLSTAG:  "colstag",
}

// SplitLayout represents a split keyboard layout with 42 keys (30 alphas + 6 thumbs per hand).
// Contains rune-to-key mappings, precomputed distance metrics, and identified ergonomic patterns
// (lateral stretches, scissors). Also includes optional optimization state (pinned keys, corpus,
// weights) used during layout generation.
type SplitLayout struct {
	Name             string                       // layout identifier (filename or user-provided)
	LayoutType       LayoutType                   // geometry type (ROWSTAG, ORTHO, COLSTAG)
	Runes            [42]rune                     // runes mapped to physical key positions (42 positions)
	RuneInfo         map[rune]KeyInfo             // map from rune to KeyInfo for quick lookup
	KeyInfos         [95]KeyInfo                  // fast lookup for ASCII runes (32-126, indexed by rune-32)
	KeyInfoValid     [95]bool                     // validity bitmap for KeyInfos array
	KeyPairDistances *map[KeyPair]KeyPairDistance // cache of distances between key index pairs
	SFBs             []SFBInfo                    // same-finger bigram key-pairs (pre-computed for performance)
	LSBs             []LSBInfo                    // notable lateral-stretch bigram key-pairs
	FScissors        []ScissorInfo                // notable full scissor key-pairs
	HScissors        []ScissorInfo                // notable half scissor key-pairs
}

// NewSplitLayout creates a new split layout and initializes precomputed ergonomic patterns
// (lateral stretches and scissors) based on the layout geometry.
func NewSplitLayout(name string, layoutType LayoutType, runes [42]rune, runeInfo map[rune]KeyInfo) *SplitLayout {
	sl := &SplitLayout{
		Name:             name,
		LayoutType:       layoutType,
		Runes:            runes,
		RuneInfo:         runeInfo,
		KeyPairDistances: &keyDistances[layoutType],
	}

	// Populate KeyInfos array for ASCII printable runes (32-126)
	// KeyInfoValid is zero-initialized (all false)
	for r, ki := range runeInfo {
		if r >= 32 && r < 127 {
			idx := r - 32
			sl.KeyInfos[idx] = ki
			sl.KeyInfoValid[idx] = true
		}
	}

	sl.initSFBs()
	sl.initLSBs()
	sl.initFScissors()
	sl.initHScissors()
	return sl
}

// Clone creates a deep copy of the SplitLayout.
// The cloned layout has the same configuration but is independent of the original.
// This is useful for optimization algorithms that need to modify layouts without affecting the original.
func (sl *SplitLayout) Clone() *SplitLayout {
	// Copy the RuneInfo map
	runeInfoCopy := make(map[rune]KeyInfo, len(sl.RuneInfo))
	maps.Copy(runeInfoCopy, sl.RuneInfo)

	// Create new layout with copied data
	// Note: Runes, KeyInfos, and KeyInfoValid are fixed-size arrays, copied by value
	// LSBs, FScissors, HScissors, and SFBs are shared (derived data, not modified after init)
	clone := &SplitLayout{
		Name:             sl.Name,
		LayoutType:       sl.LayoutType,
		Runes:            sl.Runes,            // Array is copied by value
		RuneInfo:         runeInfoCopy,        // Deep copied map
		KeyInfos:         sl.KeyInfos,         // Array is copied by value
		KeyInfoValid:     sl.KeyInfoValid,     // Array is copied by value
		KeyPairDistances: sl.KeyPairDistances, // Shared reference to immutable data
		SFBs:             sl.SFBs,             // Shared - derived data, not modified
		LSBs:             sl.LSBs,             // Shared - derived data, not modified
		FScissors:        sl.FScissors,        // Shared - derived data, not modified
		HScissors:        sl.HScissors,        // Shared - derived data, not modified
	}

	return clone
}

// GetKeyInfo returns the KeyInfo for a given rune and a boolean indicating whether the rune exists in the layout.
// For ASCII printable runes (32-126), it uses direct array indexing with validity bitmap for O(1) lookup.
// For non-ASCII runes or control characters, it falls back to the RuneInfo map.
func (sl *SplitLayout) GetKeyInfo(r rune) (KeyInfo, bool) {
	if r >= 32 && r < 127 {
		idx := r - 32
		if sl.KeyInfoValid[idx] {
			return sl.KeyInfos[idx], true
		}
		return KeyInfo{}, false
	}
	ki, ok := sl.RuneInfo[r]
	return ki, ok
}

// Swap exchanges the runes at two key positions and updates the RuneInfo map and KeyInfos array accordingly.
// This is the fundamental operation for layout optimization algorithms.
func (sl *SplitLayout) Swap(idx1, idx2 uint8) {
	if idx1 >= 42 || idx2 >= 42 {
		panic(fmt.Sprintf("swap indices out of bounds: %d, %d", idx1, idx2))
	}
	if idx1 == idx2 {
		return
	}

	// Swap runes in the array
	r1, r2 := sl.Runes[idx1], sl.Runes[idx2]
	if r1 == 0 || r2 == 0 {
		panic(fmt.Sprintf("can't swap unused key at index %d or %d", idx1, idx2))
	}
	sl.Runes[idx1], sl.Runes[idx2] = r2, r1

	// Update RuneInfo map
	sl.RuneInfo[r1], sl.RuneInfo[r2] = sl.RuneInfo[r2], sl.RuneInfo[r1]

	// Update KeyInfos array for ASCII printable runes (32-126)
	if r1 >= 32 && r1 < 127 {
		idx := r1 - 32
		sl.KeyInfos[idx] = sl.RuneInfo[r1]
		sl.KeyInfoValid[idx] = true
	}
	if r2 >= 32 && r2 < 127 {
		idx := r2 - 32
		sl.KeyInfos[idx] = sl.RuneInfo[r2]
		sl.KeyInfoValid[idx] = true
	}
}

func (sl *SplitLayout) String() string {
	var sb strings.Builder

	writeRune := func(r rune) {
		switch r {
		case 0:
			sb.WriteRune(' ')
		case ' ':
			sb.WriteRune('_')
		default:
			sb.WriteRune(r)
		}
		sb.WriteRune(' ')
	}

	//sb.WriteRune('\n')
	for row := range 3 {
		if sl.LayoutType == ANGLEMOD && row == 2 {
			sb.WriteRune(' ')
		}
		for col := range 12 {
			idx := row*12 + col
			writeRune(sl.Runes[idx])
			if col == 5 {
				sb.WriteRune(' ')
				if sl.LayoutType != ANGLEMOD || row != 2 {
					sb.WriteRune(' ')
				}
			}
		}
		sb.WriteRune('\n')
	}

	sb.WriteString("      ")
	for col := range 6 {
		idx := 36 + col
		writeRune(sl.Runes[idx])
		if col == 2 {
			sb.WriteRune(' ')
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

// NewLayoutFromFile loads a SplitLayout from a .klf file.
//
// File format:
//   - First non-comment line: layout type ("rowstag", "anglemod", "ortho", or "colstag")
//   - Next 3 lines: 12 keys each (6 left, 6 right) for main rows
//   - Last line: 6 thumb keys (3 left, 3 right)
//   - Lines starting with '#' are comments
//   - Empty lines are ignored
//
// Special tokens:
//   - "~"  : empty key (no character assigned)
//   - "_"  : space character
//   - "~~" : literal tilde character
//   - "__" : literal underscore character
//   - "##" : literal hash character
//
// Each character can appear only once in the layout.
// Returns an error if the file format is invalid or contains duplicate characters.
func NewLayoutFromFile(name, path string) (*SplitLayout, error) {
	keyMap := map[string]rune{
		"~":  rune(0),
		"_":  rune(' '),
		"~~": rune('~'),
		"__": rune('_'),
		"##": rune('#'),
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)

	// Parse layout type from first line
	layoutTypeStr, err := readLine(scanner)
	if err != nil {
		return nil, fmt.Errorf("invalid file format in %s: missing layout type", path)
	}

	var layoutType LayoutType
	layoutTypeStr = strings.ToLower(layoutTypeStr)
	switch { // must include all of layoutTypeStrings
	case strings.HasPrefix(layoutTypeStr, "rowstag"):
		layoutType = ROWSTAG
	case strings.HasPrefix(layoutTypeStr, "anglemod"):
		layoutType = ANGLEMOD
	case strings.HasPrefix(layoutTypeStr, "ortho"):
		layoutType = ORTHO
	case strings.HasPrefix(layoutTypeStr, "colstag"):
		layoutType = COLSTAG
	default:
		types := slices.Collect(maps.Values(layoutTypeStrings))
		return nil, fmt.Errorf("invalid layout type in %s: %s. Must start with one of: %v",
			path, layoutTypeStr, types)
	}

	var runeArray [42]rune
	runeInfoMap := make(map[rune]KeyInfo, 42)
	seenRunes := make(map[rune]struct{})
	expectedKeys := []int{12, 12, 12, 6}

	index := 0
	for row, expectedKeyCount := range expectedKeys {
		line, err := readLine(scanner)
		if err != nil {
			return nil, fmt.Errorf("invalid file format in %s: not enough rows", path)
		}
		keys := strings.Fields(line)
		if len(keys) != expectedKeyCount {
			return nil, fmt.Errorf("invalid file format in %s: row %d has %d keys, expected %d",
				path, row+1, len(keys), expectedKeyCount)
		}

		for col, key := range keys {
			r, ok := keyMap[strings.ToLower(key)]
			if !ok {
				if len(key) != 1 {
					return nil, fmt.Errorf("invalid file format in %s: key '%s' in row %d must have 1 character or be '__' (for _) or '~~' (for ~) or '##' (for #)", path, key, row+1)
				}
				r = rune(key[0])
			}

			// Check for duplicate runes (empty keys are allowed to repeat)
			if r != rune(0) {
				if _, exists := seenRunes[r]; exists {
					return nil, fmt.Errorf("invalid file format in %s: duplicate rune '%c' found at row %d, col %d",
						path, r, row+1, col+1)
				}
				seenRunes[r] = struct{}{}
				runeInfoMap[r] = NewKeyInfo(uint8(row), uint8(col), layoutType)
			}

			runeArray[index] = r
			index++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(name, layoutType, runeArray, runeInfoMap), nil
}

// SaveToFile saves the layout to a .klf file in the standard format.
func (sl *SplitLayout) SaveToFile(path string) error {
	inverseKeyMap := map[rune]string{
		rune(0): "~",
		' ':     "_",
		'~':     "~~",
		'_':     "__",
		'#':     "##",
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	writer := bufio.NewWriter(file)
	defer FlushWriter(writer)

	writeRune := func(r rune) {
		if str, ok := inverseKeyMap[r]; ok {
			_, _ = fmt.Fprint(writer, str)
		} else {
			_, _ = fmt.Fprintf(writer, "%c", r)
		}
	}

	// Write layout type
	_, _ = fmt.Fprintln(writer, layoutTypeStrings[sl.LayoutType])

	// Write main keys
	for row := range 3 {
		if sl.LayoutType == ANGLEMOD && row == 2 {
			_, _ = fmt.Fprint(writer, " ")
		}
		for col := range 12 {
			if col == 6 {
				_, _ = fmt.Fprint(writer, " ")
				if sl.LayoutType != ANGLEMOD || row != 2 {
					_, _ = fmt.Fprint(writer, " ")
				}
			}
			writeRune(sl.Runes[row*12+col])
			if col < 11 {
				_, _ = fmt.Fprint(writer, " ")
			}
		}
		_, _ = fmt.Fprintln(writer)
	}

	// Write thumbs
	_, _ = fmt.Fprint(writer, "      ")
	for col := range 6 {
		if col == 3 {
			_, _ = fmt.Fprint(writer, "  ")
		}
		writeRune(sl.Runes[36+col])
		if col < 5 {
			_, _ = fmt.Fprint(writer, " ")
		}
	}

	return nil
}

// readLine reads the next non-empty, non-comment line from the scanner.
// Returns an error if EOF is reached without finding a valid line.
func readLine(scanner *bufio.Scanner) (string, error) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && line != "" {
			return line, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("unexpected end of file")
}

// SFBInfo represents a same-finger bigram: two different keys typed by the same finger.
// This cache enables fast lookup of all potential SFBs based on layout geometry.
type SFBInfo struct {
	KeyIdx1 uint8
	KeyIdx2 uint8
}

// initSFBs identifies all same-finger bigram key pairs in the layout.
// Same finger bigrams occur when two different keys are typed by the same finger.
// This pre-computation enables fast SFB/SFS lookup during analysis.
func (sl *SplitLayout) initSFBs() {
	sl.SFBs = make([]SFBInfo, 0, 144)

	for key1 := range uint8(41) {
		rune1 := sl.Runes[key1]
		if rune1 == 0 {
			continue // Skip empty key positions
		}
		ki1, ok1 := sl.GetKeyInfo(rune1)
		if !ok1 {
			continue
		}

		// Only check key2 > key1 to avoid duplicate pairs
		for key2 := key1 + 1; key2 < 42; key2++ {
			rune2 := sl.Runes[key2]
			if rune2 == 0 {
				continue // Skip empty key positions
			}
			ki2, ok2 := sl.GetKeyInfo(rune2)
			if !ok2 {
				continue
			}

			// Check if same finger (different keys guaranteed by key2 > key1)
			if ki1.Finger == ki2.Finger {
				// Store both directions for consistency with LSBs/Scissors pattern
				sl.SFBs = append(sl.SFBs,
					SFBInfo{key1, key2},
					SFBInfo{key2, key1},
				)
			}
		}
	}
}

// LSBInfo represents a lateral-stretch bigram: two keys typed by non-adjacent fingers
// on the same hand that are uncomfortably far apart horizontally.
type LSBInfo struct {
	KeyIdx1     uint8
	KeyIdx2     uint8
	ColDistance float64
}

// initLSBs identifies all lateral-stretch bigram key pairs in the layout.
// Stretches occur between specific finger combinations when keys exceed a minimum horizontal distance.
func (sl *SplitLayout) initLSBs() {
	// Define finger combinations that can produce lateral stretches,
	// with minimum horizontal distance thresholds (in key units)
	fingerPairsToTrack := map[KeyPair]float64{
		{LM, LI}: 2.0, {LI, LM}: 2.0,
		{LR, LI}: 3.5, {LI, LR}: 3.5,
		{LP, LR}: 2.0, {LR, LP}: 2.0,
		{RM, RI}: 2.0, {RI, RM}: 2.0,
		{RR, RI}: 3.5, {RI, RR}: 3.5,
		{RP, RR}: 2.0, {RR, RP}: 2.0,
	}

	sl.LSBs = make([]LSBInfo, 0, 72)

	for key1, rune1 := range sl.Runes {
		if rune1 == 0 {
			continue
		}
		ri1, ok1 := sl.GetKeyInfo(rune1)
		if !ok1 {
			continue
		}

		for key2, rune2 := range sl.Runes {
			if rune2 == 0 || key1 == key2 {
				continue
			}
			ri2, ok2 := sl.GetKeyInfo(rune2)
			if !ok2 {
				continue
			}

			// Check if this finger combination is tracked
			fingerPair := [2]uint8{ri1.Finger, ri2.Finger}
			minHorDistance, ok := fingerPairsToTrack[fingerPair]
			if !ok {
				continue
			}

			// Check if distance exceeds threshold
			dx := sl.Distance(uint8(key1), uint8(key2)).ColDist
			if dx >= minHorDistance {
				sl.LSBs = append(sl.LSBs, LSBInfo{uint8(key1), uint8(key2), dx})
			}
		}
	}

	// Add geometry-specific edge cases for row-staggered layouts
	switch sl.LayoutType {
	case ROWSTAG:
		sl.LSBs = append(sl.LSBs, LSBInfo{1, 26, 1.75})
		sl.LSBs = append(sl.LSBs, LSBInfo{2, 27, 1.75})
		sl.LSBs = append(sl.LSBs, LSBInfo{3, 28, 1.75})
	case ANGLEMOD:
		// Angle-mod only stretches middle-index in this configuration
		sl.LSBs = append(sl.LSBs, LSBInfo{3, 28, 1.75})
	}
}

// ScissorInfo represents a scissor motion: two keys on the same hand typed in
// quick succession with uncomfortable vertical displacement between adjacent or close fingers.
type ScissorInfo struct {
	keyIdx1    uint8
	keyIdx2    uint8
	fingerDist uint8
	rowDist    float64
	colDist    float64
	angle      float64
}

// makePairs converts a slice of finger pairs into a lookup map.
func makePairs(pairs [][2]uint8) map[[2]uint8]bool {
	m := make(map[[2]uint8]bool, len(pairs))
	for _, p := range pairs {
		m[p] = true
	}
	return m
}

// scissorConfig defines key index ranges and valid finger pairs for finding scissors.
type scissorConfig struct {
	i1Start, i1End uint8
	i2Start, i2End uint8
	fingerPairs    map[[2]uint8]bool
}

// initScissorPairs finds all scissor key pairs matching the given configurations.
func (sl *SplitLayout) initScissorPairs(configs []scissorConfig, out *[]ScissorInfo) {
	var i1, i2 uint8
	for _, cfg := range configs {
		for i1 = cfg.i1Start; i1 <= cfg.i1End; i1++ {
			r1 := sl.Runes[i1]
			if r1 == 0 {
				continue
			}
			ki1, ok1 := sl.GetKeyInfo(r1)
			if !ok1 {
				continue
			}

			for i2 = cfg.i2Start; i2 <= cfg.i2End; i2++ {
				r2 := sl.Runes[i2]
				if r2 == 0 {
					continue
				}
				ki2, ok2 := sl.GetKeyInfo(r2)
				if !ok2 {
					continue
				}

				if cfg.fingerPairs[[2]uint8{ki1.Finger, ki2.Finger}] {
					kp := sl.Distance(i1, i2)
					angle := math.Atan2(kp.RowDist, kp.ColDist) * 180 / math.Pi

					*out = append(*out,
						ScissorInfo{i1, i2, kp.FingerDist, kp.RowDist, kp.ColDist, angle},
						ScissorInfo{i2, i1, kp.FingerDist, kp.RowDist, kp.ColDist, angle},
					)
				}
			}
		}
	}
}

// initFScissors identifies full scissor patterns (large vertical displacement, 2 rows).
func (sl *SplitLayout) initFScissors() {
	configs := []scissorConfig{
		{
			i1Start: 24, i1End: 29, // left-hand indices
			i2Start: 0, i2End: 5,
			fingerPairs: makePairs([][2]uint8{
				{LM, LP}, {LM, LR}, {LM, LI}, // LM with anything else
				{LR, LP}, {LR, LI}, // LR with LP and LI
				{LP, LR}, {LR, LM}, // LP is lower than LR, LR is lower than LM
			}),
		},
		{
			i1Start: 30, i1End: 35, // right-hand indices
			i2Start: 6, i2End: 11,
			fingerPairs: makePairs([][2]uint8{
				{RM, RI}, {RM, RR}, {RM, RP},
				{RR, RP}, {RR, RI},
				{RP, RR}, {RR, RM},
			}),
		},
	}
	sl.FScissors = make([]ScissorInfo, 0, 48)
	sl.initScissorPairs(configs, &sl.FScissors)
}

// initHScissors identifies half scissor patterns (moderate vertical displacement, 1 row).
func (sl *SplitLayout) initHScissors() {
	configs := []scissorConfig{
		{
			i1Start: 12, i1End: 17, // left-hand indices
			i2Start: 0, i2End: 5,
			fingerPairs: makePairs([][2]uint8{
				{LM, LP}, {LM, LR}, {LM, LI}, // LM with anything else
				{LR, LP}, {LR, LI}, // LR with LP and LI
				// {LP, LR}, {LR, LM}, // LP is lower than LR, LR is lower than LM
			}),
		},
		{
			i1Start: 24, i1End: 29,
			i2Start: 12, i2End: 17,
			fingerPairs: makePairs([][2]uint8{
				{LM, LP}, {LM, LR}, {LM, LI},
				{LR, LP}, {LR, LI},
				// {LP, LR}, {LR, LM},
			}),
		},
		{
			i1Start: 18, i1End: 23,
			i2Start: 6, i2End: 11,
			fingerPairs: makePairs([][2]uint8{
				{RM, RI}, {RM, RR}, {RM, RP},
				{RR, RP}, {RR, RI},
				// {RP, RR}, {RR, RM},
			}),
		},
		{
			i1Start: 30, i1End: 35,
			i2Start: 18, i2End: 23,
			fingerPairs: makePairs([][2]uint8{
				{RM, RI}, {RM, RR}, {RM, RP},
				{RR, RP}, {RR, RI},
				// {RP, RR}, {RR, RM},
			}),
		},
	}
	sl.HScissors = make([]ScissorInfo, 0, 72)
	sl.initScissorPairs(configs, &sl.HScissors)
}
