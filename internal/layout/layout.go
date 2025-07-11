package layout

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SplitLayout represents a split layout layout
type SplitLayout struct {
	Left        [3][6]rune
	Right       [3][6]rune
	LeftThumbs  [3]rune
	RightThumbs [3]rune
	RuneInfo    map[rune]KeyInfo
}

type KeyInfo struct {
	// Char   rune
	Finger int
	Hand   string // "left" or "right"
	Row    int
}

// NewSplitLayout creates a new split layout layout
func NewSplitLayout(leftKeys, rightKeys [3][6]rune, leftThumbs, rightThumbs [3]rune, runeInfo map[rune]KeyInfo) *SplitLayout {
	return &SplitLayout{
		Left:        leftKeys,
		Right:       rightKeys,
		LeftThumbs:  leftThumbs,
		RightThumbs: rightThumbs,
		RuneInfo:    runeInfo,
	}
}

// String returns a string representation of the layout layout
func (kb *SplitLayout) String() string {
	var sb strings.Builder
	for ri, row := range kb.Left {
		for i, key := range row {
			if i > 0 {
				sb.WriteRune(' ')
			}
			sb.WriteRune(key)
		}

		// add a separator
		sb.WriteString("   ")

		for i, key := range kb.Right[ri] {
			if i > 0 {
				sb.WriteRune(' ')
			}
			sb.WriteRune(key)
		}

		sb.WriteRune('\n')
	}

	sb.WriteString("      ")

	for i, key := range kb.LeftThumbs {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(key)
	}

	// add a separator
	sb.WriteString("   ")

	for i, key := range kb.RightThumbs {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(key)
	}

	// Print runes on the layout
	// sb.WriteRune('\n')
	// sb.WriteString(kb.StringRunes())

	return sb.String()
}

func (kb *SplitLayout) StringRunes() string {
	var sb strings.Builder
	for k, v := range kb.RuneInfo {
		sb.WriteString(fmt.Sprintf("Key: %c, Hand: %s, Row: %d, Finger: %d\n", k, v.Hand, v.Row, v.Finger))
	}
	return sb.String()
}

// LoadFromFile loads a layout from a text file
// The file format is:
//
//	3 lines of 12 keys each (6 left, 6 right)
//	1 line of 6 thumb keys (3 left, 3 right)
func LoadFromFile(filename string) (*SplitLayout, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var leftKeys [3][6]rune
	var rightKeys [3][6]rune
	var leftThumbs [3]rune
	var rightThumbs [3]rune
	keyInfoMap := make(map[rune]KeyInfo)

	scanner := bufio.NewScanner(file)
	for row := range 3 {
		if !scanner.Scan() {
			return nil, fmt.Errorf("invalid file format: not enough rows")
		}
		line := scanner.Text()
		keys := strings.Fields(line)
		if len(keys) != 12 {
			return nil, fmt.Errorf("invalid file format: row %d has %d keys, expected 12", row+1, len(keys))
		}
		for col := range 6 {
			leftKeys[row][col] = rune(keys[col][0])
			if keys[col][0] != '~' {
				keyInfoMap[leftKeys[row][col]] = KeyInfo{
					Finger: colToFinger(col),
					Hand:   "left",
					Row:    row,
				}
			}

			rightKeys[row][col] = rune(keys[col+6][0])
			if keys[col+6][0] != '~' {
				keyInfoMap[rightKeys[row][col]] = KeyInfo{
					Finger: colToFinger(col + 6),
					Hand:   "right",
					Row:    row,
				}
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
	for col := range 3 {
		leftThumbs[col] = rune(keys[col][0])
		if keys[col][0] != '~' {
			keyInfoMap[leftThumbs[col]] = KeyInfo{
				Finger: 4,
				Hand:   "left",
				Row:    4,
			}
		}

		rightThumbs[col] = rune(keys[col+3][0])
		if keys[col+3][0] != '~' {
			keyInfoMap[rightThumbs[col]] = KeyInfo{
				Finger: 5,
				Hand:   "right",
				Row:    4,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(leftKeys, rightKeys, leftThumbs, rightThumbs, keyInfoMap), nil
}

func colToFinger(col int) int {
	if col < 2 {
		return 0
	} else if col == 2 {
		return 1
	} else if col == 3 {
		return 2
	} else if col == 4 || col == 5 {
		return 3
	} else if col == 6 || col == 7 {
		return 6
	} else if col == 8 {
		return 7
	} else if col == 9 {
		return 8
	} else {
		return 9
	}
}

// SaveToFile saves a layout layout to a text file
func (kb *SplitLayout) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for ri, row := range kb.Left {
		for i, key := range row {
			if i > 0 {
				fmt.Fprintf(writer, " ")
			}
			fmt.Fprintf(writer, "%c", key)
		}

		fmt.Fprint(writer, "   ")

		for i, key := range kb.Right[ri] {
			if i > 0 {
				fmt.Fprintf(writer, " ")
			}
			fmt.Fprintf(writer, "%c", key)
		}

		fmt.Fprintln(writer)
	}

	fmt.Fprint(writer, "      ")

	for i, thumb := range kb.LeftThumbs {
		if i > 0 {
			fmt.Fprintf(writer, " ")
		}
		fmt.Fprintf(writer, "%c", thumb)
	}

	fmt.Fprint(writer, "   ")

	for i, thumb := range kb.RightThumbs {
		if i > 0 {
			fmt.Fprintf(writer, " ")
		}
		fmt.Fprintf(writer, "%c", thumb)
	}

	return nil
}
