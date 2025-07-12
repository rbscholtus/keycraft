package layout

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	corpus "github.com/rbscholtus/kb/internal/corpus"
)

// SplitLayout represents a split layout layout
type SplitLayout struct {
	Filename    string
	Left        [3][6]rune
	Right       [3][6]rune
	LeftThumbs  [3]rune
	RightThumbs [3]rune
	RuneInfo    map[rune]KeyInfo
}

type KeyInfo struct {
	// Char   rune
	Hand   string // "left" or "right"
	Row    int
	Column int
	Finger int
}

// NewSplitLayout creates a new split layout layout
func NewSplitLayout(filename string, leftKeys, rightKeys [3][6]rune, leftThumbs, rightThumbs [3]rune, runeInfo map[rune]KeyInfo) *SplitLayout {
	return &SplitLayout{
		Filename:    filename,
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
	keyInfoMap := make(map[rune]KeyInfo, 42)

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
					Hand:   "left",
					Row:    row,
					Column: col,
					Finger: colToFinger(col),
				}
			}

			rightKeys[row][col] = rune(keys[col+6][0])
			if keys[col+6][0] != '~' {
				keyInfoMap[rightKeys[row][col]] = KeyInfo{
					Hand:   "right",
					Row:    row,
					Column: col + 6,
					Finger: colToFinger(col + 6),
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
				Column: -1,
			}
		}

		rightThumbs[col] = rune(keys[col+3][0])
		if keys[col+3][0] != '~' {
			keyInfoMap[rightThumbs[col]] = KeyInfo{
				Finger: 5,
				Hand:   "right",
				Row:    4,
				Column: -1,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return NewSplitLayout(filename, leftKeys, rightKeys, leftThumbs, rightThumbs, keyInfoMap), nil
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
	} else if col > 9 {
		return 9
	}
	return -1
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

type Usage struct {
	Count      int
	Percentage float64
}

type HandAnalysis struct {
	LayoutName        string
	CorpusName        string
	TotalUnigramCount int
	HandUsage         [2]Usage
	RowUsage          [4]Usage
	ColumnUsage       [12]Usage
	FingerUsage       [10]Usage
}

func (lay *SplitLayout) AnalyzeHandUsage(corp *corpus.Corpus) HandAnalysis {
	var handUsage [2]Usage
	var rowUsage [4]Usage
	var columnUsage [12]Usage
	var fingerUsage [10]Usage
	totalUnigramCount := 0

	for r, count := range corp.Unigrams {
		info, ok := lay.RuneInfo[r]
		if ok {
			totalUnigramCount += count
			if info.Hand == "left" {
				handUsage[0].Count += count
			} else {
				handUsage[1].Count += count
			}
			rowUsage[info.Row].Count += count
			if info.Column >= 0 {
				columnUsage[info.Column].Count += count
			}
			fingerUsage[info.Finger].Count += count
		}
	}

	for i := range handUsage {
		handUsage[i].Percentage = 100 * float64(handUsage[i].Count) / float64(totalUnigramCount)
	}
	for i := range rowUsage {
		rowUsage[i].Percentage = 100 * float64(rowUsage[i].Count) / float64(totalUnigramCount)
	}
	for i := range columnUsage {
		columnUsage[i].Percentage = 100 * float64(columnUsage[i].Count) / float64(totalUnigramCount)
	}
	for i := range fingerUsage {
		fingerUsage[i].Percentage = 100 * float64(fingerUsage[i].Count) / float64(totalUnigramCount)
	}

	return HandAnalysis{
		LayoutName:        lay.Filename,
		CorpusName:        corp.Name,
		TotalUnigramCount: totalUnigramCount,
		HandUsage:         handUsage,
		RowUsage:          rowUsage,
		ColumnUsage:       columnUsage,
		FingerUsage:       fingerUsage,
	}
}

type SfbAnalysis struct {
	LayoutName    string
	CorpusName    string
	TotalBigrams  int
	Sfbs          []Sfb
	TotalSfbCount int
	TotalSfbPerc  float64
}

type Sfb struct {
	Bigram     corpus.Bigram
	Count      int
	Percentage float64
}

func (lay *SplitLayout) AnalyzeSfbs(corp *corpus.Corpus) SfbAnalysis {
	// get the SFBs in the corpus that occur in this layout, sorted by counts
	sfbs, totalCount := lay.extractSfbs(corp)
	sort.Slice(sfbs, func(i, j int) bool {
		return sfbs[i].Count > sfbs[j].Count
	})

	// add the percentage of counts over total corpus' bigrams
	for i := range sfbs {
		sfbs[i].Percentage = float64(sfbs[i].Count) / float64(corp.TotalBigramsCount)
	}

	// return
	return SfbAnalysis{
		LayoutName:    lay.Filename,
		CorpusName:    corp.Name,
		TotalBigrams:  corp.TotalBigramsCount,
		Sfbs:          sfbs,
		TotalSfbCount: totalCount,
		TotalSfbPerc:  float64(totalCount) / float64(corp.TotalBigramsCount),
	}
}

func (lay *SplitLayout) extractSfbs(corp *corpus.Corpus) ([]Sfb, int) {
	var sfbs []Sfb
	var totalCount int
	for bi, cnt := range corp.Bigrams {
		if bi[0] != bi[1] && lay.isSameFinger(bi) {
			sfbs = append(sfbs, Sfb{Bigram: bi, Count: cnt})
			totalCount += cnt
		}
	}
	return sfbs, totalCount
}

func (lay *SplitLayout) isSameFinger(bi corpus.Bigram) bool {
	info0, ok0 := lay.RuneInfo[bi[0]]
	info1, ok1 := lay.RuneInfo[bi[1]]
	return ok0 && ok1 && info0.Finger == info1.Finger
}
