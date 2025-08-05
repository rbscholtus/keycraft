package layout

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
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

var colToFingerMap = [...]uint8{
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LP, LP, LR, LM, LI, LI, RI, RI, RM, RR, RP, RP,
	LT, LT, LT, RT, RT, RT,
}

type LayoutType string

const (
	ROWSTAG LayoutType = "rowstag"
	ORTHO   LayoutType = "ortho"
	COLSTAG LayoutType = "colstag"
)

// Standard keyboard offsets
var rowStagOffsets = [4]float64{
	0, 0.25, 0.75, 0,
}

// Corne-style offsets
var colStagOffsets = [12]float64{
	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
}

// KeyInfo represents a key's position on a keyboard
type KeyInfo struct {
	// Char   rune
	Index  uint8
	Hand   string // "left" or "right"
	Row    uint8  // 0-3
	Column uint8  // 0-11 for Row=0-2, 0-5 for Row=3
	Finger uint8  // 0-9
}

// NewKeyInfo returns a new KeyInfo struct with some fields derived from row and col.
func NewKeyInfo(row, col uint8) KeyInfo {
	if col >= uint8(len(colToFingerMap)) {
		panic(fmt.Sprintf("col exceeds max value: %d", col))
	}
	if row > 3 {
		panic(fmt.Sprintf("row exceeds max value: %d", row))
	}

	index := 12*row + col
	if index >= 42 {
		panic(fmt.Sprintf("index exceeds max value: %d", index))
	}

	hand := "right"
	if row < 3 && col < 6 {
		hand = "left"
	} else if row == 3 && col < 3 {
		hand = "left"
	}

	finger := colToFingerMap[index]

	return KeyInfo{
		Index:  index,
		Hand:   hand,
		Row:    row,
		Column: col,
		Finger: finger,
	}
}

type KeyPair [2]uint8

type KeyPairDistance struct {
	RowDist    float64
	ColDist    float64
	FingerDist uint8
	Distance   float64
}

// SplitLayout represents a split layout
type SplitLayout struct {
	Name             string
	LayoutType       LayoutType
	Runes            [42]rune
	RuneInfo         map[rune]KeyInfo
	GetRowDist       func(uint8, uint8, uint8, uint8) float64
	GetColDist       func(uint8, uint8, uint8, uint8) float64
	KeyPairDistances map[KeyPair]KeyPairDistance
	LSBs             []LSBInfo
	Scissors         []ScissorInfo
	Pinned           [42]bool
	optCorpus        *Corpus
}

// NewSplitLayout creates a new split layout
func NewSplitLayout(name string, layoutType LayoutType, runes [42]rune, runeInfo map[rune]KeyInfo) *SplitLayout {
	rowDistFunc := IfThen(layoutType == COLSTAG, AbsRowDistAdj, AbsRowDist)
	colDistFunc := IfThen(layoutType == ROWSTAG, AbsColDistAdj, AbsColDist)
	keyDistances := getKeyDistances(rowDistFunc, colDistFunc)
	lsbs := calcLSBKeyPairs(runes, runeInfo, keyDistances, layoutType)
	scissors := calcScissorKeyPairs(runes, keyDistances)

	return &SplitLayout{
		Name:             name,
		LayoutType:       layoutType,
		Runes:            runes,
		RuneInfo:         runeInfo,
		GetRowDist:       rowDistFunc,
		GetColDist:       colDistFunc,
		KeyPairDistances: keyDistances,
		LSBs:             lsbs,
		Scissors:         scissors,
		// distances:        NewKeyDistance(layoutType),
	}
}

// StringRunes returns a string that represents	the characters on a layout.
func (sl *SplitLayout) StringRunes() string {
	var sb strings.Builder
	for k, v := range sl.RuneInfo {
		sb.WriteString(fmt.Sprintf("Key: %c, Hand: %s, Row: %d, Column: %d, Finger: %d\n",
			k, v.Hand, v.Row, v.Column, v.Finger))
	}
	return sb.String()
}

// NewLayoutFromFile loads a layout from a text file
// Lines that begin with # are ignored
// The file format is:
// - 3 lines of 12 keys each (6 left, 6 right)
// - 1 line of 6 thumb keys (3 left, 3 right)
// - _ means no key
// - spc means the Space character
// - characters cannot be repeated
func NewLayoutFromFile(name, filename string) (*SplitLayout, error) {
	keyMap := map[string]rune{
		"~":  rune(0),
		"_":  rune(' '),
		"~~": rune('~'),
		"__": rune('_'),
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)

	// Read layout type
	layoutTypeStr, err := readLine(scanner)
	if err != nil {
		return nil, fmt.Errorf("invalid file format: missing layout type")
	}

	var layoutType LayoutType
	layoutTypeStr = strings.ToLower(layoutTypeStr)
	switch {
	case strings.HasPrefix(layoutTypeStr, "rowstag"):
		layoutType = ROWSTAG
	case strings.HasPrefix(layoutTypeStr, "ortho"):
		layoutType = ORTHO
	case strings.HasPrefix(layoutTypeStr, "colstag"):
		layoutType = COLSTAG
	default:
		types := []LayoutType{ROWSTAG, ORTHO, COLSTAG}
		return nil, fmt.Errorf("invalid layout type: %s. Must start with one of: %v", layoutTypeStr, types)
	}

	var runeArray [42]rune
	runeInfoMap := make(map[rune]KeyInfo, 42)
	expectedKeys := []int{12, 12, 12, 6}

	index := 0
	for row, expectedKeyCount := range expectedKeys {
		line, err := readLine(scanner)
		if err != nil {
			return nil, fmt.Errorf("invalid file format: not enough rows")
		}
		keys := strings.Fields(line)
		if len(keys) != expectedKeyCount {
			return nil, fmt.Errorf("invalid file format: row %d has %d keys, expected %d", row+1, len(keys), expectedKeyCount)
		}

		for col, key := range keys {
			r, ok := keyMap[strings.ToLower(key)]
			if !ok {
				if len(key) != 1 {
					return nil, fmt.Errorf("invalid file format: key '%s' in row %d must have 1 character or be '__' (for _) or '~~' (for ~)", key, row+1)
				}
				r = rune(key[0])
			}

			runeArray[index] = r
			index++
			if r != rune(0) {
				runeInfoMap[r] = NewKeyInfo(uint8(row), uint8(col))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(name, layoutType, runeArray, runeInfoMap), nil
}

// SaveToFile saves a layout layout to a text file
func (sl *SplitLayout) SaveToFile(filename string) error {
	inverseKeyMap := map[rune]string{
		rune(0): "~",
		' ':     "_",
		'~':     "~~",
		'_':     "__",
	}

	file, err := os.Create(filename)
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
	_, _ = fmt.Fprintln(writer, strings.ToLower(string(sl.LayoutType)))

	// Write main keys
	for row := range 3 {
		for col := range 12 {
			if col == 6 {
				_, _ = fmt.Fprint(writer, " ")
			}
			writeRune(sl.Runes[row*12+col])
			if col < 11 {
				_, _ = fmt.Fprint(writer, " ")
			}
		}
		_, _ = fmt.Fprintln(writer)
	}

	// Write thumbs
	_, _ = fmt.Fprint(writer, "    ")
	for col := range 6 {
		if col == 3 {
			_, _ = fmt.Fprint(writer, " ")
		}
		writeRune(sl.Runes[36+col])
		if col < 5 {
			_, _ = fmt.Fprint(writer, " ")
		}
	}

	return nil
}

// LoadPins loads a pins file and populates the Pinned array.
func (sl *SplitLayout) LoadPins(filename string) error {
	// Open the file for reading.
	file, err := os.Open(filename)
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
			return fmt.Errorf("invalid file format: not enough rows")
		}
		keys := strings.Fields(scanner.Text())
		if len(keys) != expectedKeyCount {
			return fmt.Errorf("invalid file format: row %d has %d keys, expected %d", row+1, len(keys), expectedKeyCount)
		}
		for col, key := range keys {
			if len(key) != 1 {
				return fmt.Errorf("invalid file format: key '%s' in row %d must have 1 character only", key, row+1)
			}
			switch rune(key[0]) {
			case '.', '_', '-':
				// Unpinned keys.
				sl.Pinned[index] = false
			case '*', 'x', 'X':
				// Pinned keys.
				sl.Pinned[index] = true
			default:
				return fmt.Errorf("invalid character '%c' at position %d in row %d", key[0], col+1, row+1)
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

// There is a minor error in the calcs for the thumb keys!
func getKeyDistances(rowDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64, colDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64) map[KeyPair]KeyPairDistance {
	keyDistances := make(map[KeyPair]KeyPairDistance)

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
				FingerDist: absDist(colToFingerMap[k1], colToFingerMap[k2]),
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

type LSBInfo struct {
	keyIdx1     int
	keyIdx2     int
	colDistance float64
}

func calcLSBKeyPairs(runes [42]rune, runeInfo map[rune]KeyInfo, keyPairDists map[KeyPair]KeyPairDistance, layoutType LayoutType) []LSBInfo {
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

	// LSBs we're going to find and track
	LSBs := []LSBInfo{}

	for key1, rune1 := range runes {
		if rune1 == 0 {
			continue
		}
		ri1, ok1 := runeInfo[rune1]
		if !ok1 {
			continue
		}

		for key2, rune2 := range runes {
			if rune2 == 0 || key1 == key2 {
				continue
			}
			ri2, ok2 := runeInfo[rune2]
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
			dx := keyPairDists[KeyPair{uint8(key1), uint8(key2)}].ColDist
			if dx >= minHorDistance {
				LSBs = append(LSBs, LSBInfo{key1, key2, dx})
			}
		}
	}

	// As per Keyboard Layout Doc, section 7.4.2
	// Add a few more notable LSBs on row-staggered
	if layoutType == ROWSTAG {
		LSBs = append(LSBs, LSBInfo{1, 26, 1.75})
		LSBs = append(LSBs, LSBInfo{2, 27, 1.75})
		LSBs = append(LSBs, LSBInfo{3, 28, 1.75})
	}

	return LSBs
}

type ScissorInfo struct {
	keyIdx1    uint8
	keyIdx2    uint8
	fingerDist uint8
	rowDist    float64
	angle      float64
}

func calcScissorKeyPairs(runes [42]rune, keyPairDists map[KeyPair]KeyPairDistance) []ScissorInfo {
	var indexPairs = []KeyPair{
		// Full Scissors
		// left-hand side
		{26, 1},
		{27, 2},
		{27, 4},
		{27, 5},
		{26, 4},
		{26, 5},
		{27, 1},
		// right-hand side
		{33, 10},
		{32, 9},
		{32, 7},
		{32, 6},
		{33, 7},
		{33, 6},
		{32, 10},
		// pinky is lower than the ring
		{25, 2},
		{26, 3},
		{34, 9},
		{33, 8},
		// Half Scissors
		// left-hand side, row 0/1
		{14, 1},
		{15, 2},
		{15, 4},
		{15, 5},
		{14, 4},
		{14, 5},
		{15, 1},
		// left-hand, row 1/2
		{26, 13},
		{27, 14},
		{27, 16},
		{27, 17},
		{26, 16},
		{26, 17},
		{27, 13},
		// right-hand side, row 0/1
		{21, 10},
		{20, 9},
		{20, 7},
		{20, 6},
		{21, 7},
		{21, 6},
		{20, 10},
		// right-hand, row 1/2
		{33, 22},
		{32, 21},
		{32, 19},
		{32, 18},
		{33, 19},
		{33, 18},
		{32, 22},
		// // pinky is lower than the ring, row 0/1
		// {13, 2},
		// {14, 3},
		// {22, 9},
		// {21, 8},
		// // pinky is lower than the ring, row 1/2
		// {25, 14},
		// {26, 15},
		// {34, 21},
		// {33, 20},
	}

	// Scissors we're going to find and track
	var scissors []ScissorInfo
	for _, idxPair := range indexPairs {
		r0, r1 := runes[idxPair[0]], runes[idxPair[1]]
		if r0 == 0 || r1 == 0 {
			// key on layout has no character
			continue
		}

		kp := keyPairDists[idxPair]
		dx := kp.ColDist
		dy := kp.RowDist
		angle := math.Atan2(dy, dx) * 180 / math.Pi

		// Add the new pair (bi-directional)
		scissors = append(scissors, ScissorInfo{
			keyIdx1:    idxPair[0],
			keyIdx2:    idxPair[1],
			fingerDist: kp.FingerDist,
			rowDist:    dy,
			angle:      angle,
		}, ScissorInfo{
			keyIdx1:    idxPair[1],
			keyIdx2:    idxPair[0],
			fingerDist: kp.FingerDist,
			rowDist:    dy,
			angle:      angle,
		})
	}

	// godump.Dump(scissors)

	return scissors
}
