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

	scanner := bufio.NewScanner(file)
	index := 0
	for row := range uint8(3) {
		if !scanner.Scan() {
			return nil, fmt.Errorf("invalid file format: not enough rows")
		}
		line := scanner.Text()
		keys := strings.Fields(line)
		if len(keys) != 12 {
			return nil, fmt.Errorf("invalid file format: row %d has %d keys, expected 12", row+1, len(keys))
		}
		for col := range uint8(12) {
			r := rune(keys[col][0])
			runeArray[index] = r
			index++
			if r != '~' {
				runeInfoMap[r] = NewKeyInfo(row, col)
			}
		}
	}

	if !scanner.Scan() {
		return nil, fmt.Errorf("invalid file format: not enough rows for thumbs")
	}
	line := scanner.Text()
	keys := strings.Fields(line)
	if len(keys) != 6 {
		return nil, fmt.Errorf("invalid file format: thumbs row has %d keys, expected 6", len(keys))
	}
	for col := range uint8(6) {
		r := rune(keys[col][0])
		runeArray[index] = r
		index++
		if r != '~' {
			runeInfoMap[r] = NewKeyInfo(3, col)
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
