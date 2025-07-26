package layout

import (
	"bufio"
	"fmt"
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

type LSBKeyPair struct {
	runeIdx1 int
	runeIdx2 int
	distance float32
}

// SplitLayout represents a split layout layout
type SplitLayout struct {
	Filename            string
	Name                string
	Runes               [42]rune
	RuneInfo            map[rune]KeyInfo
	LayoutType          LayoutType
	KeyPairHorDistances []LSBKeyPair
	distances           *KeyDistance
	Pinned              [42]bool
	optCorpus           *Corpus
}

// NewSplitLayout creates a new split layout
func NewSplitLayout(filename, name string, runes [42]rune, runeInfo map[rune]KeyInfo, layoutType LayoutType) *SplitLayout {
	return &SplitLayout{
		Filename:            filename,
		Name:                name,
		Runes:               runes,
		RuneInfo:            runeInfo,
		LayoutType:          layoutType,
		KeyPairHorDistances: calcKeyPairDistances(runes, runeInfo, layoutType),
		distances:           NewKeyDistance(layoutType),
	}
}

func calcKeyPairDistances(runes [42]rune, runeInfo map[rune]KeyInfo, layoutType LayoutType) []LSBKeyPair {
	// Get the x "coordinate", which is adjusted for row-staggered keyboards
	getAdjustedColumn := func(row uint8, column uint8) float32 {
		if layoutType != RowStagLayout {
			return float32(column)
		}
		switch row {
		case 1:
			return float32(column) + 0.25
		case 2:
			return float32(column) + 0.75
		default:
			return float32(column)
		}
	}

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

	keyPairHorDistances := []LSBKeyPair{}

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
			dx := Abs(getAdjustedColumn(ri1.Row, ri1.Column) - getAdjustedColumn(ri2.Row, ri2.Column))
			if dx >= minHorDistance {
				keyPairHorDistances = append(keyPairHorDistances, LSBKeyPair{i1, i2, dx})
			}
		}
	}

	// As per Keyboard Layout Doc, section 7.4.2
	// Add a few more notable LSBs on row-staggered
	if layoutType == RowStagLayout {
		keyPairHorDistances = append(keyPairHorDistances, LSBKeyPair{1, 26, float32(1.75)})
		keyPairHorDistances = append(keyPairHorDistances, LSBKeyPair{2, 27, float32(1.75)})
		keyPairHorDistances = append(keyPairHorDistances, LSBKeyPair{3, 28, float32(1.75)})
	}

	return keyPairHorDistances
}

// String returns a string representation of the layout layout
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
