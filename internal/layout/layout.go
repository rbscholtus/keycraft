package layout

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	corpus "github.com/rbscholtus/kb/internal/corpus"
)

// SplitLayout represents a split layout layout
type SplitLayout struct {
	Filename  string
	Runes     [42]rune
	RuneInfo  map[rune]KeyInfo
	Pinned    [42]bool
	optCorpus *corpus.Corpus
}

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
	} else {
		finger = ifThen(col < 3, uint8(4), uint8(5))
	}

	return KeyInfo{
		Hand:   hand,
		Row:    row,
		Column: col,
		Finger: finger,
	}
}

// NewSplitLayout creates a new split layout layout
func NewSplitLayout(filename string, runes [42]rune, runeInfo map[rune]KeyInfo) *SplitLayout {
	return &SplitLayout{
		Filename: filename,
		Runes:    runes,
		RuneInfo: runeInfo,
	}
}

// String returns a string representation of the layout layout
func (sl *SplitLayout) String() string {
	var sb strings.Builder
	for row := range 3 {
		for col := range 12 {
			if col == 6 {
				sb.WriteRune(' ')
			}
			sb.WriteRune(sl.Runes[row*12+col])
			if col < 11 {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune('\n')
	}

	sb.WriteString("      ")

	for col := range 6 {
		if col == 3 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(sl.Runes[36+col])
		if col < 5 {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

func (sl *SplitLayout) StringRunes() string {
	var sb strings.Builder
	for k, v := range sl.RuneInfo {
		sb.WriteString(fmt.Sprintf("Key: %c, Hand: %s, Row: %d, Column: %d, Finger: %d\n",
			k, v.Hand, v.Row, v.Column, v.Finger))
	}
	return sb.String()
}

// NewFromFile loads a layout from a text file
// The file format is:
//
//	3 lines of 12 keys each (6 left, 6 right)
//	1 line of 6 thumb keys (3 left, 3 right)
func NewFromFile(filename string) (*SplitLayout, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var runeArray [42]rune
	runeInfoMap := make(map[rune]KeyInfo, 42)
	expectedKeys := []int{12, 12, 12, 6}

	scanner := bufio.NewScanner(file)
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
			if len(key) != 1 {
				return nil, fmt.Errorf("invalid file format: key '%s' in row %d must have 1 character only", key, row+1)
			}
			r := rune(key[0])
			runeArray[index] = r
			index++
			if r != '~' {
				runeInfoMap[r] = NewKeyInfo(uint8(row), uint8(col))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(filename, runeArray, runeInfoMap), nil
}

// SaveToFile saves a layout layout to a text file
func (sl *SplitLayout) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write main keys
	for row := range 3 {
		for col := range 12 {
			if col == 6 {
				fmt.Fprint(writer, " ")
			}
			fmt.Fprintf(writer, "%c", sl.Runes[row*12+col])
			if col < 11 {
				fmt.Fprint(writer, " ")
			}
		}
		fmt.Fprintln(writer)
	}

	// Write thumbs
	fmt.Fprint(writer, "      ")
	for col := range 6 {
		if col == 3 {
			fmt.Fprint(writer, " ")
		}
		fmt.Fprintf(writer, "%c", sl.Runes[36+col])
		if col < 5 {
			fmt.Fprint(writer, " ")
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
	defer file.Close()

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
