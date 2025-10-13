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
	"unicode"
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

// KeyPair represents an ordered pair of key indices.
type KeyPair [2]uint8

// KeyPairDistance contains precomputed distance metrics between two key indices.
type KeyPairDistance struct {
	RowDist    float64 // vertical (row) distance in layout units
	ColDist    float64 // horizontal (column) distance in layout units
	FingerDist uint8   // absolute difference between the two keys' finger indices
	Distance   float64 // Euclidean distance (sqrt(RowDist^2 + ColDist^2))
}

// keyDistances contains precomputed key pair distances for each LayoutType.
// Distance functions are selected based on keyboard geometry:
//   - ROWSTAG: AbsRowDist, AbsColDistAdj (accounts for row stagger)
//   - ANGLEMOD: AbsRowDist, AbsColDistAdj (similar to row-staggered)
//   - ORTHO: AbsRowDist, AbsColDist (simple grid distances)
//   - COLSTAG: AbsRowDistAdj, AbsColDist (accounts for column stagger)
var keyDistances = []map[KeyPair]KeyPairDistance{
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &keyToFinger),         // ROWSTAG
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &angleModKeyToFinger), // ANGLEMOD
	calcKeyDistances(AbsRowDist, AbsColDist, &keyToFinger),            // ORTHO
	calcKeyDistances(AbsRowDistAdj, AbsColDist, &keyToFinger),         // COLSTAG
}

// rowStagOffsets defines the horizontal offset for each row in row-staggered layouts.
// Traditional keyboards have rows offset by different amounts (in key units).
var rowStagOffsets = [4]float64{
	0, 0.25, 0.75, 0, // Top, home, bottom, thumb rows
}

// colStagOffsets defines the vertical offset for each column in column-staggered layouts.
// Ergonomic keyboards (e.g., Corne) stagger columns to match natural finger lengths.
var colStagOffsets = [12]float64{
	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
}

const (
	LEFT  uint8 = 0 // Left hand
	RIGHT uint8 = 1 // Right hand
)

// KeyInfo represents a key's physical position and typing finger on a keyboard.
type KeyInfo struct {
	Index  uint8 // 0-41
	Hand   uint8 // LEFT or RIGHT
	Row    uint8 // 0-3
	Column uint8 // 0-11 for Row=0-2, 0-5 for Row=3
	Finger uint8 // 0-9
}

// NewKeyInfo constructs a KeyInfo from row, column, and layout type.
// Automatically determines hand and finger assignments based on position and geometry.
func NewKeyInfo(row, col uint8, layoutType LayoutType) KeyInfo {
	if col >= uint8(len(keyToFinger)) {
		panic(fmt.Sprintf("col exceeds max value: %d", col))
	}
	if row > 3 {
		panic(fmt.Sprintf("row exceeds max value: %d", row))
	}

	index := 12*row + col
	if index >= 42 {
		panic(fmt.Sprintf("index exceeds max value: %d", index))
	}

	hand := RIGHT
	if row < 3 && col < 6 {
		hand = LEFT
	} else if row == 3 && col < 3 {
		hand = LEFT
	}

	var finger uint8
	if layoutType == ANGLEMOD {
		finger = angleModKeyToFinger[index]
	} else {
		finger = keyToFinger[index]
	}

	return KeyInfo{
		Index:  index,
		Hand:   hand,
		Row:    row,
		Column: col,
		Finger: finger,
	}
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
	KeyPairDistances *map[KeyPair]KeyPairDistance // cache of distances between key index pairs
	LSBs             []LSBInfo                    // notable lateral-stretch bigram key-pairs
	FScissors        []ScissorInfo                // notable full scissor key-pairs
	HScissors        []ScissorInfo                // notable half scissor key-pairs
	optPinned        [42]bool                     // Optimization: flags indicating keys that must not be swapped
	optCorpus        *Corpus                      // Optimization: corpus for evaluating layout quality
	optIdealfgrLoad  *[10]float64                 // Optimization: ideal finger load distribution
	optWeights       *Weights                     // Optimization: metric weights for scoring
	optMedians       map[string]float64           // Optimization: median values for normalization
	optIqrs          map[string]float64           // Optimization: IQR values for normalization
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
	sl.initLSBs()
	sl.initFScissors()
	sl.initHScissors()
	return sl
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

	sb.WriteRune('\n')
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

// LoadPins loads a pins file specifying which keys should be fixed during optimization.
// The file format mirrors a layout file, but uses symbols to indicate pin status:
//   - '.', '_', '-' : unpinned (key can be moved)
//   - '*', 'x', 'X' : pinned (key is fixed)
func (sl *SplitLayout) LoadPins(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pins file %s does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)
	index := 0
	expectedKeys := []int{12, 12, 12, 6}

	// Parse pins from file
	for row, expectedKeyCount := range expectedKeys {
		if !scanner.Scan() {
			return fmt.Errorf("invalid file format in %s: not enough rows", path)
		}
		keys := strings.Fields(scanner.Text())
		if len(keys) != expectedKeyCount {
			return fmt.Errorf("invalid file format in %s: row %d has %d keys, expected %d", path, row+1, len(keys), expectedKeyCount)
		}
		for col, key := range keys {
			if len(key) != 1 {
				return fmt.Errorf("invalid file format in %s: key '%s' in row %d must have 1 character only", path, key, row+1)
			}
			switch rune(key[0]) {
			case '.', '_', '-':
				sl.optPinned[index] = false
			case '*', 'x', 'X':
				sl.optPinned[index] = true
			default:
				return fmt.Errorf("invalid character in %s '%c' at position %d in row %d", path, key[0], col+1, row+1)
			}
			index++
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// LoadPinsFromParams configures which keys are pinned during optimization.
// Three modes:
//  1. Load from pins file (path) - uses pin file format
//  2. Pin specific characters (pins) - comma-separated characters to fix
//  3. Free specific characters (free) - all others are pinned
//
// Modes 1 and 2 can be combined, but mode 3 (free) is mutually exclusive.
// If no options are provided, only empty keys and spaces are pinned by default.
func (sl *SplitLayout) LoadPinsFromParams(path, pins, free string) error {
	if (path != "" || pins != "") && free != "" {
		return fmt.Errorf("cannot use both --free and --pins/--pins-file options together")
	}

	if free != "" {
		// Pin everything except specified characters
		for i := range sl.optPinned {
			sl.optPinned[i] = true
		}
		// Unpin characters in free string
		for _, r := range free {
			key, ok := sl.RuneInfo[r]
			if !ok {
				return fmt.Errorf("cannot free unavailable character: %c", r)
			}
			sl.optPinned[key.Index] = false
		}
		return nil
	}

	// Load pins from file if specified
	if path != "" {
		if err := sl.LoadPins(path); err != nil {
			return err
		}
	} else {
		// By default, pin empty keys and spaces
		for i, r := range sl.Runes {
			if r == 0 || unicode.IsSpace(r) {
				sl.optPinned[i] = true
			}
		}
	}

	// Pin additional characters from pins string
	for _, r := range pins {
		key, ok := sl.RuneInfo[r]
		if !ok {
			return fmt.Errorf("cannot pin unavailable character: %c", r)
		}
		sl.optPinned[key.Index] = true
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

// Distance returns the precomputed distance between two key indices.
// If the key pair is not found, it returns nil.
func (sl *SplitLayout) Distance(k1, k2 uint8) *KeyPairDistance {
	kpd, ok := (*sl.KeyPairDistances)[KeyPair{k1, k2}]
	if !ok {
		return nil
	}
	return &kpd
}

// calcKeyDistances precomputes all pairwise distances between keys on the same hand.
// Uses the provided distance functions to account for layout-specific geometry.
// Note: thumb key distance calculations have a known minor inaccuracy.
func calcKeyDistances(
	rowDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	colDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	keyToFinger *[42]uint8,
) map[KeyPair]KeyPairDistance {
	keyDistances := make(map[KeyPair]KeyPairDistance, 624)

	// Optimized square root for common cases
	sqrt := func(mul float64) float64 {
		switch mul {
		case 1:
			return 1 // no calc necessary
		case 2:
			return math.Sqrt2 // pre-calculated
		default:
			return math.Sqrt(mul)
		}
	}

	absDist := func(x, y uint8) uint8 {
		if x > y {
			return x - y
		} else {
			return y - x
		}
	}

	var k1, k2 uint8
	for k1 = range 42 {
		row1, col1 := k1/12, k1%12
		for k2 = range 42 {
			if k1 == k2 {
				continue
			}
			row2, col2 := k2/12, k2%12

			// Skip pairs on different hands
			if ((row1 < 3 && col1 < 6) || (row1 >= 3 && col1 < 3)) !=
				((row2 < 3 && col2 < 6) || (row2 >= 3 && col2 < 3)) {
				continue
			}

			// Skip pairs between main rows and thumb row
			if (row1 < 3) != (row2 < 3) {
				continue
			}

			// Compute distance metrics
			dx := colDistFunc(row1, col1, row2, col2)
			dy := rowDistFunc(row1, col1, row2, col2)
			dist := sqrt(dx*dx + dy*dy)
			keyDistances[KeyPair{k1, k2}] = KeyPairDistance{
				RowDist:    dy,
				ColDist:    dx,
				FingerDist: absDist(keyToFinger[k1], keyToFinger[k2]),
				Distance:   dist,
			}
		}
	}

	return keyDistances
}

// AbsRowDist computes the absolute vertical distance between two keys (simple).
func AbsRowDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(row1) - float64(row2))
}

// AbsRowDistAdj computes vertical distance accounting for column-stagger offsets.
func AbsRowDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(row1) + colStagOffsets[col1] -
		(float64(row2) + colStagOffsets[col2])))
}

// AbsColDist computes the absolute horizontal distance between two keys (simple).
func AbsColDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(col1) - float64(col2))
}

// AbsColDistAdj computes horizontal distance accounting for row-stagger offsets.
func AbsColDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(col1) + rowStagOffsets[row1] -
		(float64(col2) + rowStagOffsets[row2])))
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
		ri1, ok1 := sl.RuneInfo[rune1]
		if !ok1 {
			continue
		}

		for key2, rune2 := range sl.Runes {
			if rune2 == 0 || key1 == key2 {
				continue
			}
			ri2, ok2 := sl.RuneInfo[rune2]
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
			ki1 := sl.RuneInfo[r1]

			for i2 = cfg.i2Start; i2 <= cfg.i2End; i2++ {
				r2 := sl.Runes[i2]
				if r2 == 0 {
					continue
				}
				ki2 := sl.RuneInfo[r2]

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
