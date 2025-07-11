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
}

// NewSplitLayout creates a new split layout layout
func NewSplitLayout(leftKeys, rightKeys [3][6]rune, leftThumbs, rightThumbs [3]rune) *SplitLayout {
	return &SplitLayout{
		Left:        leftKeys,
		Right:       rightKeys,
		LeftThumbs:  leftThumbs,
		RightThumbs: rightThumbs,
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

	for i, thumb := range kb.LeftThumbs {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(thumb)
	}

	// add a separator
	sb.WriteString("   ")

	for i, thumb := range kb.RightThumbs {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(thumb)
	}

	return sb.String()
}

// LoadFromFile loads a layout from a text file
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
			rightKeys[row][col] = rune(keys[col+6][0])
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
	for i := range 3 {
		leftThumbs[i] = rune(keys[i][0])
		rightThumbs[i] = rune(keys[i+3][0])
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(leftKeys, rightKeys, leftThumbs, rightThumbs), nil
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
