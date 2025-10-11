// Package keycraft provides ergonomic analysis tools for keyboard layouts.
//
// Metrics are described in README.md.
// Metrics are calculated from unigrams, bigrams, skipgrams, and trigrams in the corpus.
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

const (
	LP uint8 = iota
	LR
	LM
	LI
	LT

	RT
	RI
	RM
	RR
	RP
)

var keyToFinger = [...]uint8{
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LT, LT, LT, RT, RT, RT,
}

var angleModKeyToFinger = [...]uint8{
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LR, LM, LI, LI, LI, RI, RI, RM, RR, RP, RP,
	LT, LT, LT, RT, RT, RT,
}

type LayoutType uint8

const (
	ROWSTAG LayoutType = iota
	ANGLEMOD
	ORTHO
	COLSTAG
)

// Map from LayoutType to string
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
// The combinations of row/column distance functions are chosen as follows:
//  1. ROWSTAG: AbsRowDist, AbsColDistAdj (accounts for row-staggered columns)
//  2. ANGLEMOD: AbsRowDist, AbsColDistAdj (angle-modified layouts use similar logic)
//  3. ORTHO: AbsRowDist, AbsColDist (ortholinear layouts use simple absolute distances; this is intentional)
//  4. COLSTAG: AbsRowDistAdj, AbsColDist (column-staggered layouts adjust row distance only)
var keyDistances = []map[KeyPair]KeyPairDistance{
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &keyToFinger),         // ROWSTAG
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &angleModKeyToFinger), // ANGLEMOD
	calcKeyDistances(AbsRowDist, AbsColDist, &keyToFinger),            // ORTHO
	calcKeyDistances(AbsRowDistAdj, AbsColDist, &keyToFinger),         // COLSTAG
}

// Standard row-staggered keyboard column offsets
var rowStagOffsets = [4]float64{
	0, 0.25, 0.75, 0,
}

// Corne-style row offsets
var colStagOffsets = [12]float64{
	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
}

const (
	LEFT  uint8 = 0
	RIGHT uint8 = 1
)

// KeyInfo represents a key's position on a keyboard
type KeyInfo struct {
	Index  uint8 // 0-41
	Hand   uint8 // LEFT or RIGHT
	Row    uint8 // 0-3
	Column uint8 // 0-11 for Row=0-2, 0-5 for Row=3
	Finger uint8 // 0-9
}

// NewKeyInfo returns a new KeyInfo struct with some fields derived from row and col.
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

// SplitLayout represents a split keyboard layout and associated analysis metadata.
// It contains rune placement, per-rune key information, precomputed pairwise distances,
// notable lateral-stretch and scissor pairs, pinned-key flags, and optional fields used
// during optimisation.
type SplitLayout struct {
	Name             string                       // layout identifier (filename or user-provided)
	LayoutType       LayoutType                   // geometry type (ROWSTAG, ORTHO, COLSTAG)
	Runes            [42]rune                     // runes mapped to physical key positions (42 positions)
	RuneInfo         map[rune]KeyInfo             // map from rune to KeyInfo for quick lookup
	KeyPairDistances *map[KeyPair]KeyPairDistance // cache of distances between key index pairs
	LSBs             []LSBInfo                    // notable lateral-stretch bigram key-pairs
	FScissors        []ScissorInfo                // notable full scissor key-pairs
	HScissors        []ScissorInfo                // notable half scissor key-pairs
	optPinned        [42]bool                     // optimisation: flags indicating keys that must not be moved
	optCorpus        *Corpus                      // optimisation: corpus used during layout optimisation (optional)
	optIdealfgrLoad  *[10]float64                 // optimisation:
	optWeights       *Weights                     // optimisation: metric weights used (optional)
	optMedians       map[string]float64           // optimisation: median values per metric (optional)
	optIqrs          map[string]float64           // optimisation: IQR values per metric (optional)
}

// NewSplitLayout creates a new split layout
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

// NewLayoutFromFile loads a SplitLayout from the named file.
// The file must contain:
//   - a first non-empty, non-comment line indicating layout type:
//     "rowstag", "ortho", or "colstag" (prefix matching allowed).
//   - three subsequent rows of 12 keys each (6 left, 6 right).
//   - one final row of 6 thumb keys (3 left, 3 right).
//
// Special tokens in the file:
//
//	"~"   -> empty key
//	"_"   -> space character
//	"~~"  -> literal '~'
//	"__"  -> literal '_'
//	"##"  -> literal '#'
//
// Lines starting with '#' and empty lines are ignored. Each key entry must be
// either one character or one of the special tokens above. Characters must not
// be repeated. Returns a parsed *SplitLayout or an error on malformed input.
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

	// Read layout type
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

			// Check for duplicate runes (skip rune(0) as it represents empty/null positions)
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

// SaveToFile saves a layout layout to a text file
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

// LoadPins loads a pins file and populates the Pinned array.
func (sl *SplitLayout) LoadPins(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("pins file %s does not exist", path)
	}

	// Open the file for reading.
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)
	index := 0
	expectedKeys := []int{12, 12, 12, 6}

	// Read the pins from the file.
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
				// Unpinned keys.
				sl.optPinned[index] = false
			case '*', 'x', 'X':
				// Pinned keys.
				sl.optPinned[index] = true
			default:
				return fmt.Errorf("invalid character in %s '%c' at position %d in row %d", path, key[0], col+1, row+1)
			}
			index++
		}
	}

	// Check for any scanner errors.
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// LoadPinsFromParams loads pin information into the SplitLayout from a file, pins string,
// or a free string (specifying which runes are free, all others pinned).
//
// Parameters:
//   - path: path to a pins file (optional). If empty, no file-based pins are loaded.
//   - pins: a string of characters to pin individually in the layout.
//   - free: a string of characters that are free to move (all others are pinned).
//
// If path or pins are provided, free must be empty.
// If free is provided, all runes except those in free are pinned.
func (sl *SplitLayout) LoadPinsFromParams(path, pins, free string) error {
	// If pins-file or pins are specified, free must be empty.
	if (path != "" || pins != "") && free != "" {
		return fmt.Errorf("cannot use both --free and --pins/--pins-file options together")
	}

	if free != "" {
		// Pin all runes except those in free string
		// First, mark all as pinned
		for i := range sl.optPinned {
			sl.optPinned[i] = true
		}
		// Unpin the runes in free, if they exist in layout
		for _, r := range free {
			key, ok := sl.RuneInfo[r]
			if !ok {
				return fmt.Errorf("cannot free unavailable character: %c", r)
			}
			sl.optPinned[key.Index] = false
		}
		return nil
	}

	// Pin keys as specified in the pinfile
	if path != "" {
		if err := sl.LoadPins(path); err != nil {
			return err
		}
	} else {
		// Otherwise, pin keys that are not used for an actual rune and Space
		for i, r := range sl.Runes {
			if r == 0 || unicode.IsSpace(r) {
				sl.optPinned[i] = true
			}
		}
	}

	// Additionally, pin keys in the pins parameter
	for _, r := range pins {
		key, ok := sl.RuneInfo[r]
		if !ok {
			return fmt.Errorf("cannot pin unavailable character: %c", r)
		}
		sl.optPinned[key.Index] = true
	}

	return nil
}

// readLine reads a line, ignoring empty lines and lines that start with #
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

// There is a minor error in the calcs for the thumb keys!
func calcKeyDistances(
	rowDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	colDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	keyToFinger *[42]uint8,
) map[KeyPair]KeyPairDistance {
	keyDistances := make(map[KeyPair]KeyPairDistance, 624)

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

			// skip if the keys are on different hands
			if ((row1 < 3 && col1 < 6) || (row1 >= 3 && col1 < 3)) !=
				((row2 < 3 && col2 < 6) || (row2 >= 3 && col2 < 3)) {
				continue
			}

			// skip if exactly one of the keys is on the thumb cluster
			if (row1 < 3) != (row2 < 3) {
				continue
			}

			// calculate distances
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

func AbsRowDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(row1) - float64(row2))
}

func AbsRowDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(row1) + colStagOffsets[col1] -
		(float64(row2) + colStagOffsets[col2])))
}

func AbsColDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(col1) - float64(col2))
}

func AbsColDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(col1) + rowStagOffsets[row1] -
		(float64(col2) + rowStagOffsets[row2])))
}

// LSBInfo holds information about a lateral-stretch bigram candidate on the layout.
type LSBInfo struct {
	KeyIdx1     uint8
	KeyIdx2     uint8
	ColDistance float64
}

// Initialize LSB key-pairs
func (sl *SplitLayout) initLSBs() {
	// Which two fingers (nrs 0..9) may form pairs,
	// and what it the minimum distance (2.0 or 3.5) to note them
	// Each pair is noted in both directions
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

			// find a pair of runes on the layout typed by a predefined finger pair
			fingerPair := [2]uint8{ri1.Finger, ri2.Finger}
			minHorDistance, ok := fingerPairsToTrack[fingerPair]
			if !ok {
				continue
			}

			// Get horizontal distance and add
			dx := sl.Distance(uint8(key1), uint8(key2)).ColDist
			if dx >= minHorDistance {
				sl.LSBs = append(sl.LSBs, LSBInfo{uint8(key1), uint8(key2), dx})
			}
		}
	}

	// As per Keyboard Layouts Doc, section 7.4.2
	// Add a few more notable LSBs on row-staggered
	switch sl.LayoutType {
	case ROWSTAG:
		sl.LSBs = append(sl.LSBs, LSBInfo{1, 26, 1.75})
		sl.LSBs = append(sl.LSBs, LSBInfo{2, 27, 1.75})
		sl.LSBs = append(sl.LSBs, LSBInfo{3, 28, 1.75})
	case ANGLEMOD:
		// only the middle - index situation is a stretch with anglemod
		sl.LSBs = append(sl.LSBs, LSBInfo{3, 28, 1.75})
	}
}

// ScissorInfo describes a scissor key-pair (full or half) including finger distance, row distance and angle.
type ScissorInfo struct {
	keyIdx1    uint8
	keyIdx2    uint8
	fingerDist uint8
	rowDist    float64
	colDist    float64
	angle      float64
}

// Helper to make map from slice of pairs
func makePairs(pairs [][2]uint8) map[[2]uint8]bool {
	m := make(map[[2]uint8]bool, len(pairs))
	for _, p := range pairs {
		m[p] = true
	}
	return m
}

// A config holds index ranges and the valid finger pair map
type scissorConfig struct {
	i1Start, i1End uint8
	i2Start, i2End uint8
	fingerPairs    map[[2]uint8]bool
}

// Helper to initialize scissor pairs
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

// Initializes full scissors
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

// Initializes half scissors
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
