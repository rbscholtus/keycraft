package layout

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

// KeyInfo represents a key's position on a keyboard
type KeyInfo struct {
	// Char   rune
	Hand   string // "left" or "right"
	Row    uint8  // 0-3
	Column uint8  // 0-11 for Row=0-2, 0-5 for Row=3
	Finger uint8  // 0-9
}

var colToFingerMap = [...]uint8{
	0, 0, 1, 2, 3, 3,
	6, 6, 7, 8, 9, 9,
}

// NewKeyInfo returns a new KeyInfo struct with some fields derived from row and col.
func NewKeyInfo(row, col uint8) KeyInfo {
	if col >= uint8(len(colToFingerMap)) {
		panic(fmt.Sprintf("col exceeds max value: %d", col))
	}
	if row > 3 {
		panic(fmt.Sprintf("row exceeds max value: %d", row))
	}

	hand := "right"
	if row < 3 && col < 6 {
		hand = "left"
	} else if row == 3 && col < 3 {
		hand = "left"
	}

	var finger uint8
	if row < 3 {
		finger = colToFingerMap[col]
	} else if col < 3 {
		finger = 4
	} else {
		finger = 5
	}

	return KeyInfo{
		Hand:   hand,
		Row:    row,
		Column: col,
		Finger: finger,
	}
}

type LayoutType string

const (
	RowStagLayout LayoutType = "rowstag"
	OrthoLayout   LayoutType = "ortho"
	ColStagLayout LayoutType = "colstag"
)

type LSBInfo struct {
	runeIdx1 int
	runeIdx2 int
	distance float32
}

type ScissorInfo struct {
	runeIdx1   int
	runeIdx2   int
	bigram     string // for debugging, should remove at some point
	fingerDist uint8
	rowDist    uint8
	angle      float32
}

// SplitLayout represents a split layout
type SplitLayout struct {
	Filename    string
	Name        string
	Runes       [42]rune
	RuneInfo    map[rune]KeyInfo
	LayoutType  LayoutType
	LSBInfo     []LSBInfo
	ScissorInfo []ScissorInfo
	distances   *KeyDistance
	Pinned      [42]bool
	optCorpus   *Corpus
}

// NewSplitLayout creates a new split layout
func NewSplitLayout(filename, name string, runes [42]rune, runeInfo map[rune]KeyInfo, layoutType LayoutType) *SplitLayout {
	return &SplitLayout{
		Filename:    filename,
		Name:        name,
		Runes:       runes,
		RuneInfo:    runeInfo,
		LayoutType:  layoutType,
		LSBInfo:     calcLSBKeyPairs(runes, runeInfo, layoutType),
		ScissorInfo: calcFSBKeyPairs(runes, runeInfo, layoutType, FsbPairs),
		distances:   NewKeyDistance(layoutType),
	}
}

// Get the x "coordinate", which is adjusted for row-staggered keyboards
func getAdjustedColumnStaggered(row uint8, column uint8) float32 {
	switch row {
	case 1:
		return float32(column) + 0.25
	case 2:
		return float32(column) + 0.75
	default:
		return float32(column)
	}
}

// Get the x "coordinate", which is just the column for ortho and colstag
func getAdjustedColumn(row uint8, column uint8) float32 {
	return float32(column)
}

// Get the y "coordinate", which is adjusted for col-staggered keyboards
func getAdjustedRowStaggered(row uint8, column uint8) float32 {
	return float32(row) + colStagOffsets[column]
}

// Get the y "coordinate", which is just the row for ortho and rowstag
func getAdjustedRow(row uint8, column uint8) float32 {
	return float32(row)
}

func calcLSBKeyPairs(runes [42]rune, runeInfo map[rune]KeyInfo, layoutType LayoutType) []LSBInfo {
	// Which two fingers (nrs 0..9) may form pairs,
	// and what it the minimum distance (2.0 or 3.5) to note them
	// Each pair is noted in both directions
	validFingerPairs := map[[2]uint8]float32{
		{2, 3}: 2, {3, 2}: 2,
		{7, 6}: 2, {6, 7}: 2,
		{1, 3}: 3.5, {3, 1}: 3.5,
		{8, 6}: 3.5, {6, 8}: 3.5,
		{0, 1}: 2, {1, 0}: 2,
		{9, 8}: 2, {8, 9}: 2,
	}

	var getColumn func(uint8, uint8) float32
	if layoutType == RowStagLayout {
		getColumn = getAdjustedColumnStaggered
	} else {
		getColumn = getAdjustedColumn
	}

	keyPairHorDistances := []LSBInfo{}

	for i1, rune1 := range runes {
		if rune1 == 0 {
			continue
		}
		ri1, ok1 := runeInfo[rune1]
		if !ok1 {
			continue
		}

		for i2, rune2 := range runes {
			if rune2 == 0 || i1 == i2 {
				continue
			}
			ri2, ok2 := runeInfo[rune2]
			if !ok2 {
				continue
			}

			// find a pair of runes on the layout typed by a predefined finger pair
			fingerPair := [2]uint8{ri1.Finger, ri2.Finger}
			minHorDistance, ok := validFingerPairs[fingerPair]
			if !ok {
				continue
			}

			// Get horizontal distance and add
			dx := Abs(getColumn(ri1.Row, ri1.Column) - getColumn(ri2.Row, ri2.Column))
			if dx >= minHorDistance {
				keyPairHorDistances = append(keyPairHorDistances, LSBInfo{i1, i2, dx})
			}
		}
	}

	// As per Keyboard Layout Doc, section 7.4.2
	// Add a few more notable LSBs on row-staggered
	if layoutType == RowStagLayout {
		keyPairHorDistances = append(keyPairHorDistances, LSBInfo{1, 26, float32(1.75)})
		keyPairHorDistances = append(keyPairHorDistances, LSBInfo{2, 27, float32(1.75)})
		keyPairHorDistances = append(keyPairHorDistances, LSBInfo{3, 28, float32(1.75)})
	}

	return keyPairHorDistances
}

var FsbPairs = [][]int{
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

func calcFSBKeyPairs(runes [42]rune, runeInfo map[rune]KeyInfo, layoutType LayoutType, pairs [][]int) []ScissorInfo {
	var getColumn, getRow func(uint8, uint8) float32
	if layoutType == RowStagLayout {
		getColumn = getAdjustedColumnStaggered
	} else {
		getColumn = getAdjustedColumn
	}
	if layoutType == ColStagLayout {
		getRow = getAdjustedRowStaggered
	} else {
		getRow = getAdjustedRow
	}

	var keyPairs []ScissorInfo
	for _, pair := range pairs {
		r0, r1 := runes[pair[0]], runes[pair[1]]
		if r0 == 0 || r1 == 0 {
			// key on layout has no character
			continue
		}
		ri1, ok := runeInfo[r0]
		if !ok {
			continue
		}
		ri2, ok := runeInfo[r1]
		if !ok {
			continue
		}

		// Calculate finger and row distance
		fingerDist := IfThen(ri1.Finger > ri2.Finger, ri1.Finger-ri2.Finger, ri2.Finger-ri1.Finger)
		rowDist := IfThen(ri1.Row > ri2.Row, ri1.Row-ri2.Row, ri2.Row-ri1.Row)

		// Calculate angle
		dx := Abs(getColumn(ri1.Row, ri1.Column) - getColumn(ri2.Row, ri2.Column))
		dy := Abs(getRow(ri1.Row, ri1.Column) - getRow(ri2.Row, ri2.Column))
		angle := float32(math.Atan2(float64(dy), float64(dx))) * 180 / math.Pi

		// Add the new pair (bi-directional)
		keyPairs = append(keyPairs, ScissorInfo{
			runeIdx1:   pair[0],
			runeIdx2:   pair[1],
			bigram:     string([]rune{r0, r1}),
			fingerDist: fingerDist,
			rowDist:    rowDist,
			angle:      angle,
		}, ScissorInfo{
			runeIdx1:   pair[1],
			runeIdx2:   pair[0],
			bigram:     string([]rune{r1, r0}),
			fingerDist: fingerDist,
			rowDist:    rowDist,
			angle:      angle,
		})
	}

	return keyPairs
}

// String returns a string representation of the layout
func (sl *SplitLayout) String() string {
	var sb strings.Builder
	writeRune := func(r rune) {
		switch r {
		case 0:
			sb.WriteString("no")
		case ' ':
			sb.WriteString("spc")
		default:
			sb.WriteRune(r)
		}
	}

	sb.WriteString(strings.ToLower(string(sl.LayoutType)))
	sb.WriteRune('\n')

	for row := range 3 {
		for col := range 12 {
			if col == 6 {
				sb.WriteRune(' ')
			}
			writeRune(sl.Runes[row*12+col])
			if col < 11 {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune('\n')
	}

	sb.WriteString("    ")

	for col := range 6 {
		if col == 3 {
			sb.WriteRune(' ')
		}
		writeRune(sl.Runes[36+col])
		if col < 5 {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
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
// The file format is:
//
//	3 lines of 12 keys each (6 left, 6 right)
//	1 line of 6 thumb keys (3 left, 3 right)
func NewLayoutFromFile(name, filename string) (*SplitLayout, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)

	// Read layout type
	if !scanner.Scan() {
		return nil, fmt.Errorf("invalid file format: missing layout type")
	}
	layoutTypeStr := strings.TrimSpace(scanner.Text())
	var layoutType LayoutType
	switch strings.ToLower(layoutTypeStr) {
	case "rowstag":
		layoutType = RowStagLayout
	case "ortho":
		layoutType = OrthoLayout
	case "colstag":
		layoutType = ColStagLayout
	default:
		types := []LayoutType{RowStagLayout, OrthoLayout, ColStagLayout}
		return nil, fmt.Errorf("invalid layout type: %s. Must be one of: %v", layoutTypeStr, types)
	}

	var runeArray [42]rune
	runeInfoMap := make(map[rune]KeyInfo, 42)
	expectedKeys := []int{12, 12, 12, 6}

	index := 0
	for row, expectedKeyCount := range expectedKeys {
		if !scanner.Scan() {
			return nil, fmt.Errorf("invalid file format: not enough rows")
		}
		keys := strings.Fields(scanner.Text())
		if len(keys) != expectedKeyCount {
			return nil, fmt.Errorf("invalid file format: row %d has %d keys, expected %d", row+1, len(keys), expectedKeyCount)
		}
		for col, key := range keys {
			switch strings.ToLower(key) {
			case "no":
				runeArray[index] = rune(0)
				index++
			case "spc":
				r := rune(' ')
				runeArray[index] = r
				index++
				runeInfoMap[r] = NewKeyInfo(uint8(row), uint8(col))
			default:
				if len(key) != 1 {
					return nil, fmt.Errorf("invalid file format: key '%s' in row %d must have 1 character or be 'no' or 'spc'", key, row+1)
				}
				r := rune(key[0])
				runeArray[index] = r
				index++
				runeInfoMap[r] = NewKeyInfo(uint8(row), uint8(col))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(filename, name, runeArray, runeInfoMap, layoutType), nil
}

// SaveToFile saves a layout layout to a text file
func (sl *SplitLayout) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	writer := bufio.NewWriter(file)
	defer FlushWriter(writer)

	writeRune := func(r rune) {
		switch r {
		case 0:
			_, _ = fmt.Fprint(writer, "no")
		case ' ':
			_, _ = fmt.Fprint(writer, "spc")
		default:
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

// GetDistance finds and returns the distance between two runes measured in U units.
// Returns an error if one or both keys do not occur on the layout.
func (sl *SplitLayout) GetDistance(r1, r2 rune) (float32, error) {
	key1, ok1 := sl.RuneInfo[r1]
	if !ok1 {
		return 0, fmt.Errorf("unsupported character in this layout: %c", r1)
	}
	key2, ok2 := sl.RuneInfo[r2]
	if !ok2 {
		return 0, fmt.Errorf("unsupported character in this layout: %c", r2)
	}
	return sl.distances.GetDistance(key1, key2), nil
}
