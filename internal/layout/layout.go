package layout

import (
	"bufio"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"

	"github.com/MaxHalford/eaopt"
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
	optCorpus   *corpus.Corpus
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
func (sl *SplitLayout) String() string {
	var sb strings.Builder
	for ri, row := range sl.Left {
		for i, key := range row {
			if i > 0 {
				sb.WriteRune(' ')
			}
			sb.WriteRune(key)
		}

		// add a separator
		sb.WriteString("   ")

		for i, key := range sl.Right[ri] {
			if i > 0 {
				sb.WriteRune(' ')
			}
			sb.WriteRune(key)
		}

		sb.WriteRune('\n')
	}

	sb.WriteString("      ")

	for i, key := range sl.LeftThumbs {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteRune(key)
	}

	// add a separator
	sb.WriteString("   ")

	for i, key := range sl.RightThumbs {
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

func (sl *SplitLayout) StringRunes() string {
	var sb strings.Builder
	for k, v := range sl.RuneInfo {
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
func (sl *SplitLayout) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for ri, row := range sl.Left {
		for i, key := range row {
			if i > 0 {
				fmt.Fprintf(writer, " ")
			}
			fmt.Fprintf(writer, "%c", key)
		}

		fmt.Fprint(writer, "   ")

		for i, key := range sl.Right[ri] {
			if i > 0 {
				fmt.Fprintf(writer, " ")
			}
			fmt.Fprintf(writer, "%c", key)
		}

		fmt.Fprintln(writer)
	}

	fmt.Fprint(writer, "      ")

	for i, thumb := range sl.LeftThumbs {
		if i > 0 {
			fmt.Fprintf(writer, " ")
		}
		fmt.Fprintf(writer, "%c", thumb)
	}

	fmt.Fprint(writer, "   ")

	for i, thumb := range sl.RightThumbs {
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

func (ha HandAnalysis) String() string {
	return fmt.Sprintf("%s (%s):\nHands=%s\nRows=%s\nColumns=%s\nFingers=%s",
		ha.LayoutName,
		ha.CorpusName,
		formatUsage(ha.HandUsage[:]),
		formatUsage(ha.RowUsage[:]),
		formatUsage(ha.ColumnUsage[:]),
		formatUsage(ha.FingerUsage[:]),
	)
}

func formatUsage(usage []Usage) string {
	var parts []string
	for _, u := range usage {
		parts = append(parts, fmt.Sprintf("%.1f%%", u.Percentage))
	}
	return strings.Join(parts, ", ")
}

func (sl *SplitLayout) AnalyzeHandUsage(corp *corpus.Corpus) HandAnalysis {
	var handUsage [2]Usage
	var rowUsage [4]Usage
	var columnUsage [12]Usage
	var fingerUsage [10]Usage
	totalUnigramCount := 0

	for r, count := range corp.Unigrams {
		info, ok := sl.RuneInfo[r]
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
		LayoutName:        sl.Filename,
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

func (sa SfbAnalysis) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Corpus: %s (%s bigrams)\n", sa.CorpusName, comma(sa.TotalBigrams)))
	sb.WriteString(fmt.Sprintf("Total SFBs in %v: %s (%.3f%% of corpus)\n",
		sa.LayoutName, comma(sa.TotalSfbCount), 100*sa.TotalSfbPerc))
	printCount := min(10, len(sa.Sfbs))
	sb.WriteString(fmt.Sprintf("Top-%d SFBs:\n", printCount))
	for i := range printCount {
		sfb := sa.Sfbs[i]
		sb.WriteString(fmt.Sprintf("%2d. %v (%s, %.3f%%)\n", i+1, sfb.Bigram, comma(sfb.Count), 100*sfb.Percentage))
	}

	return sb.String()
}

// For if we just want to know the SFB perc for a layout and corpus
func (sl *SplitLayout) SimpleSfbs(corp *corpus.Corpus) float64 {
	var totalCount int
	for bi, cnt := range corp.Bigrams {
		if bi[0] != bi[1] && sl.isSameFinger(bi) {
			totalCount += cnt
		}
	}
	return float64(totalCount) / float64(corp.TotalBigramsCount)
}

func (sl *SplitLayout) AnalyzeSfbs(corp *corpus.Corpus) SfbAnalysis {
	// get the SFBs in the corpus that occur in this layout, sorted by counts
	sfbs, totalCount := sl.extractSfbs(corp)
	sort.Slice(sfbs, func(i, j int) bool {
		return sfbs[i].Count > sfbs[j].Count
	})

	// add the percentage of counts over total corpus' bigrams
	for i := range sfbs {
		sfbs[i].Percentage = float64(sfbs[i].Count) / float64(corp.TotalBigramsCount)
	}

	// return
	return SfbAnalysis{
		LayoutName:    sl.Filename,
		CorpusName:    corp.Name,
		TotalBigrams:  corp.TotalBigramsCount,
		Sfbs:          sfbs,
		TotalSfbCount: totalCount,
		TotalSfbPerc:  float64(totalCount) / float64(corp.TotalBigramsCount),
	}
}

func (sl *SplitLayout) extractSfbs(corp *corpus.Corpus) ([]Sfb, int) {
	var sfbs []Sfb
	var totalCount int
	for bi, cnt := range corp.Bigrams {
		if bi[0] != bi[1] && sl.isSameFinger(bi) {
			sfbs = append(sfbs, Sfb{Bigram: bi, Count: cnt})
			totalCount += cnt
		}
	}
	return sfbs, totalCount
}

func (sl *SplitLayout) isSameFinger(bi corpus.Bigram) bool {
	info0, ok0 := sl.RuneInfo[bi[0]]
	info1, ok1 := sl.RuneInfo[bi[1]]
	return ok0 && ok1 && info0.Finger == info1.Finger
}

func (sl *SplitLayout) Optimise(corp *corpus.Corpus, acceptWorse string, generations int) *SplitLayout {
	sl.optCorpus = corp

	// Simulated annealing is implemented as a GA using the ModSimulatedAnnealing model.
	cfg := eaopt.NewDefaultGAConfig()
	cfg.Model = eaopt.ModSimulatedAnnealing{
		Accept: func(g, ng uint, e0, e1 float64) float64 {
			switch acceptWorse {
			case "always":
				return 1.0
			case "never":
				return 0.0
			case "drop-slow":
				t := 1.0 - float64(g)/float64(ng)
				return (math.Cos(t*math.Pi) + 1.0) / 2.0
			case "temp":
				t := 1.0 - float64(g)/float64(ng)
				return t
			case "cold":
				t := 1.0 - float64(g)/float64(ng)
				return 0.5 * t
			case "drop-fast":
				t := 1.0 - float64(g)/float64(ng)
				return math.Exp(-3.0 * (1 - t))
			default:
				panic("unknown accept worse function")
			}
		},
	}
	cfg.NGenerations = uint(generations)

	// Add a custom callback function to track progress.
	minFit := math.MaxFloat64
	cfg.Callback = func(ga *eaopt.GA) {
		hof0 := ga.HallOfFame[0]
		fit := hof0.Fitness
		if fit == minFit {
			// Output only when we make an improvement.
			return
		}
		// best := hof0.Genome.(*SplitLayout)
		fmt.Printf("Best fitness at generation %3d: %.3f%%\n", ga.Generations, 100*fit)
		minFit = fit
	}

	// Run the simulated-annealing algorithm.
	ga, err := cfg.NewGA()
	if err != nil {
		panic(err)
	}
	err = ga.Minimize(func(rng *rand.Rand) eaopt.Genome {
		return sl
	})
	if err != nil {
		panic(err)
	}

	// Return the best encountered solution
	hof0 := ga.HallOfFame[0]
	best := hof0.Genome.(*SplitLayout)

	return best
}

// Evaluate evaluates the Holder-table function at the current coordinates.
func (sl *SplitLayout) Evaluate() (float64, error) {
	return sl.SimpleSfbs(sl.optCorpus), nil
	// z := -math.Abs(math.Sin(sl.X) * math.Cos(sl.Y) *
	// 	math.Exp(math.Abs(1.0-math.Sqrt(sl.X*sl.X+sl.Y*sl.Y)/math.Pi)))
	// return z, nil
}

func randomBigram(rng *rand.Rand, sfbs []Sfb) corpus.Bigram {
	// Calculate the total percentage
	var total int
	for _, sfb := range sfbs {
		total += sfb.Count
	}

	// Generate a random number between 0 and the total percentage
	randNum := rng.Intn(total)

	// Select the bigram based on the random number
	var cumulative int
	for _, sfb := range sfbs {
		cumulative += sfb.Count
		if randNum <= cumulative {
			return sfb.Bigram
		}
	}

	// If the random number exceeds the total percentage, return the last bigram
	return sfbs[len(sfbs)-1].Bigram
}

// Mutate replaces one of the current coordinates with a random value in [-10, 10).
func (sl *SplitLayout) Mutate(rng *rand.Rand) {
	// Get a list of keys from the RuneInfo map
	keys := make([]rune, 0, len(sl.RuneInfo))
	for k := range sl.RuneInfo {
		keys = append(keys, k)
	}

	sfbs, _ := sl.extractSfbs(sl.optCorpus)
	sort.Slice(sfbs, func(i, j int) bool {
		return sfbs[i].Count > sfbs[j].Count
	})

	bi := randomBigram(rng, sfbs)
	key1 := bi[rng.Intn(2)]
	for sl.RuneInfo[key1].Row == 1 {
		bi = randomBigram(rng, sfbs)
		key1 = bi[rng.Intn(2)]
	}

	j := rng.Intn(len(keys))
	key2 := keys[j]
	for key1 == key2 || sl.RuneInfo[key2].Row == 1 {
		j = rng.Intn(len(keys))
		key2 = keys[j]
	}

	// Swap the values associated with the two keys
	if sl.RuneInfo[key2].Hand == "left" {
		r, c := sl.RuneInfo[key2].Row, sl.RuneInfo[key2].Column
		sl.Left[r][c%6] = key1
	} else {
		r, c := sl.RuneInfo[key2].Row, sl.RuneInfo[key2].Column
		sl.Right[r][c%6] = key1
	}
	if sl.RuneInfo[key1].Hand == "left" {
		r, c := sl.RuneInfo[key1].Row, sl.RuneInfo[key1].Column
		sl.Left[r][c%6] = key2
	} else {
		r, c := sl.RuneInfo[key1].Row, sl.RuneInfo[key1].Column
		sl.Right[r][c%6] = key2
	}
	sl.RuneInfo[key1], sl.RuneInfo[key2] = sl.RuneInfo[key2], sl.RuneInfo[key1]

	//fmt.Printf("%v->%c%c ", bi, key1, key2)
}

// Crossover does nothing. It is defined only so *SplitLayout implements the eaopt.Genome interface.
func (sl *SplitLayout) Crossover(other eaopt.Genome, rng *rand.Rand) {}

// Clone returns a copy of a *Coord2D.
func (sl *SplitLayout) Clone() eaopt.Genome {
	cc := SplitLayout{
		Filename:    sl.Filename,
		Left:        sl.Left,
		Right:       sl.Right,
		LeftThumbs:  sl.LeftThumbs,
		RightThumbs: sl.RightThumbs,
		RuneInfo:    make(map[rune]KeyInfo),
		optCorpus:   sl.optCorpus,
	}

	maps.Copy(cc.RuneInfo, sl.RuneInfo)

	return &cc
}

func comma(vi int) string {
	v := int64(vi)

	// Counting the number of digits.
	var count byte = 0
	for n := v; n != 0; n = n / 10 {
		count++
	}

	count += (count - 1) / 3
	output := make([]byte, count)
	j := len(output) - 1

	var counter byte = 0
	for v > 9 {
		output[j] = byte(v%10) + '0'
		v = v / 10
		j--
		if counter == 2 {
			counter = 0
			output[j] = ','
			j--
		} else {
			counter++
		}
	}

	output[j] = byte(v) + '0'

	return string(output)
}
