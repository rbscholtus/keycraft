// Package layout provides functionality for analyzing keyboard layouts.
package layout

// HandUsageAnalysis holds statistics about hand, finger, column, and row usage.
type HandUsageAnalysis struct {
	// HandUsage stores the percentage of usage for each hand.
	HandUsage [2]float64
	// FingerUsage stores the percentage of usage for each finger.
	FingerUsage [10]float64
	// ColumnUsage stores the percentage of usage for each column.
	ColumnUsage [12]float64
	// RowUsage stores the percentage of usage for each row.
	RowUsage [4]float64
}

// Analyser holds references to a keyboard layout and a corpus, and provides methods for analysis.
type Analyser struct {
	// Reference to the analysed Layout.
	Layout *SplitLayout
	// Reference to the Corpus used to analyse the layout.
	Corpus *Corpus
	// HandUsage holds statistics about hand usage.
	HandUsage HandUsageAnalysis
	// Metrics holds basic metrics about the layout.
	Metrics map[string]float64
}

// NewAnalyser creates a new Analyser instance and performs initial analysis.
func NewAnalyser(layout *SplitLayout, corpus *Corpus, style string) *Analyser {
	a := &Analyser{
		Layout:  layout,
		Corpus:  corpus,
		Metrics: make(map[string]float64),
	}
	a.quickHandAnalysis()
	switch style {
	case "keysolve":
		a.keysolvewebMetricAnalysis()
	default:
		a.quickMetricAnalysis()
	}
	return a
}

// quickHandAnalysis calculates hand, finger, column, and row usage statistics.
func (an *Analyser) quickHandAnalysis() {
	// Initialize counters for total unigrams, hand usage, finger usage, column usage, and row usage.
	var totalUnigramCount uint64
	var handCount [2]uint64
	var fingerCount [10]uint64
	var columnCount [12]uint64
	var rowCount [4]uint64

	// Iterate over unigrams in the corpus and calculate usage statistics.
	for uniGr, uniCnt := range an.Corpus.Unigrams {
		key, ok := an.Layout.RuneInfo[rune(uniGr)]
		if !ok {
			// Skip unigrams that are not present in the layout.
			continue
		}

		totalUnigramCount += uniCnt
		if key.Hand == "left" {
			handCount[0] += uniCnt
		} else {
			handCount[1] += uniCnt
		}
		fingerCount[key.Finger] += uniCnt
		columnCount[key.Column] += uniCnt
		rowCount[key.Row] += uniCnt
	}

	// Calculate the percentages.
	factor := 100 / float64(totalUnigramCount)
	for i, c := range handCount {
		an.HandUsage.HandUsage[i] = float64(c) * factor
	}
	for i, c := range fingerCount {
		an.HandUsage.FingerUsage[i] = float64(c) * factor
	}
	for i, c := range columnCount {
		an.HandUsage.ColumnUsage[i] = float64(c) * factor
	}
	for i, c := range rowCount {
		an.HandUsage.RowUsage[i] = float64(c) * factor
	}
}

func (an *Analyser) quickMetricAnalysis() {
	// Initialize counters and factor for percentage calculation.
	var factor float64
	var count1, count2, count3, count4 uint64

	// Calculate basic bigram statistics.
	for bi, biCnt := range an.Corpus.Bigrams {
		key1, ok1 := an.Layout.RuneInfo[bi[0]]
		key2, ok2 := an.Layout.RuneInfo[bi[1]]
		if !ok1 || !ok2 {
			// Skip bigrams that are not present in the layout.
			continue
		}

		// All bigram stats are on 1 hand.
		if key1.Hand != key2.Hand {
			continue
		}

		// Calculate SFB (Same Finger Bigram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += biCnt
			continue
		}
	}

	// Calculate LSB (Lateral Stretch Bigram) count.
	for _, lsb := range an.Layout.LSBs {
		bi := Bigram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count2 += cnt
		}
	}

	// Calculate Scissors counts.
	for _, sci := range an.Layout.Scirrors {
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			if sci.rowDist > 1.5 {
				count3 += cnt
			} else {
				count4 += cnt
			}
		}
	}

	// Calculate percentages for bigram statistics.
	factor = 100 / float64(an.Corpus.TotalBigramsCount)
	an.Metrics["SFB"] = float64(count1) * factor
	an.Metrics["LSB"] = float64(count2) * factor
	an.Metrics["FSB"] = float64(count3) * factor
	an.Metrics["HSB"] = float64(count4) * factor

	// Calculate skipgram statistics (SFS, LSS, FSS, HSS).
	count1, count2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.Corpus.Trigrams {
		key1, ok1 := an.Layout.RuneInfo[tri[0]]
		key2, ok2 := an.Layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 {
			// Skip trigrams that are not present in the layout.
			continue
		}

		// First and third character are on 1 hand.
		if key1.Hand != key2.Hand {
			continue
		}

		// Calculate SFS (Same Finger Skipgram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += triCnt
			continue
		}

		// Calculate LSS (Lateral Stretch Skipgram) count.
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			count2 += triCnt
		}

		// Function to check if a finger is on the bottom row.
		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// Calculate FSS (Full Scissor Skipgram) count.
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += triCnt
		}

		// Calculate HSS (Half Scissor Skipgram) count.
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += triCnt
		}
	}

	// Calculate percentages for skipgram statistics.
	factor = 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["SFS"] = float64(count1) * factor
	an.Metrics["LSS"] = float64(count2) * factor
	an.Metrics["FSS"] = float64(count3) * factor
	an.Metrics["HSS"] = float64(count4) * factor

	// Calculate trigram statistics (ALT, ROL, ONE, RED).
	count1, count2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.Corpus.Trigrams {
		key1, ok1 := an.Layout.RuneInfo[tri[0]]
		key2, ok2 := an.Layout.RuneInfo[tri[1]]
		key3, ok3 := an.Layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 || !ok3 {
			// Skip trigrams that are not present in the layout.
			continue
		}

		// Calculate ALT (Alternation) count.
		if key1.Hand == key3.Hand && key1.Hand != key2.Hand {
			count1 += triCnt
		}

		// Calculate ROL (Roll) count.
		if key1.Hand != key3.Hand &&
			key1.Finger != key2.Finger && key2.Finger != key3.Finger {
			count2 += triCnt
		}

		// Check if all fingers are on the same hand and in order.
		allSameHand := key1.Hand == key2.Hand && key1.Hand == key3.Hand
		inOrder := (key1.Finger < key2.Finger && key2.Finger < key3.Finger) ||
			(key1.Finger > key2.Finger && key2.Finger > key3.Finger)
		allDiffFingers := key1.Finger != key2.Finger &&
			key1.Finger != key3.Finger &&
			key2.Finger != key3.Finger

		// Calculate ONE (One hand, fingers in order) count.
		if allSameHand && inOrder {
			count3 += triCnt
		}

		// Calculate RED (Redirection) count.
		if allSameHand && allDiffFingers && !inOrder {
			count4 += triCnt
		}
	}

	// Calculate percentages for trigram statistics.
	factor = 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["ALT"] = float64(count1) * factor
	an.Metrics["ROL"] = float64(count2) * factor
	an.Metrics["ONE"] = float64(count3) * factor
	an.Metrics["RED"] = float64(count4) * factor
}

// quickMetricAnalysis calculates basic metrics about the layout, based on Keysolve-web.
// All metrics exclude n-grams with space characters.
// SFB - Same Finger Bigram: percentage of bigrams typed with the same finger.
// LSB - Lateral Stretch Bigram: percentage of bigrams with lateral stretch (between index and pinky fingers).
// FSB - Full Scissor Bigram: percentage of bigrams typed with full scissor motion.
// HSB - Half Scissor Bigram: percentage of bigrams typed with half scissor motion.
// SFS - Same Finger Skipgram: percentage of skipgrams (trigrams with the first and last key typed by the same finger) typed with the same finger.
// LSS - Lateral Stretch Skipgram: percentage of skipgrams with lateral stretch.
// FSS - Full Scissor Skipgram: percentage of skipgrams typed with full scissor motion.
// HSS - Half Scissor Skipgram: percentage of skipgrams typed with half scissor motion.
// ALT - Alternation: percentage of trigrams where the first and last key are typed by the same hand and the middle key is typed by the other hand.
// ROL - Roll: percentage of trigrams where each key is typed by a different finger and the hands alternate.
// ONE - One hand, fingers in order: percentage of trigrams typed with one hand and fingers in order.
// RED - Redirection: percentage of trigrams typed with one hand, all fingers different, and fingers not in order.
func (an *Analyser) keysolvewebMetricAnalysis() {
	// Initialize counters and factor for percentage calculation.
	var factor float64
	var count1, count2, count3, count4 uint64

	// Calculate basic bigram statistics.
	for bi, biCnt := range an.Corpus.Bigrams {
		key1, ok1 := an.Layout.RuneInfo[bi[0]]
		key2, ok2 := an.Layout.RuneInfo[bi[1]]
		if !ok1 || !ok2 {
			// Skip bigrams that are not present in the layout.
			continue
		}

		// All bigram stats are on 1 hand.
		if key1.Hand != key2.Hand {
			continue
		}

		// Calculate SFB (Same Finger Bigram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += biCnt
			continue
		}

		// Calculate LSB (Lateral Stretch Bigram) count.
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			count2 += biCnt
		}

		// Function to check if a finger is on the bottom row.
		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// Calculate FSB (Forward Stretch Bigram) count.
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += biCnt

		}

		// Calculate HSB (Home row Stretch Bigram) count.
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += biCnt
		}
	}

	// Calculate percentages for bigram statistics.
	factor = 100 / float64(an.Corpus.TotalBigramsCount)
	an.Metrics["SFB"] = float64(count1) * factor
	an.Metrics["LSB"] = float64(count2) * factor
	an.Metrics["FSB"] = float64(count3) * factor
	an.Metrics["HSB"] = float64(count4) * factor

	// Calculate skipgram statistics (SFS, LSS, FSS, HSS).
	count1, count2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.Corpus.Trigrams {
		key1, ok1 := an.Layout.RuneInfo[tri[0]]
		key2, ok2 := an.Layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 {
			// Skip trigrams that are not present in the layout.
			continue
		}

		// First and third character are on 1 hand.
		if key1.Hand != key2.Hand {
			continue
		}

		// Calculate SFS (Same Finger Skipgram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += triCnt
			continue
		}

		// Calculate LSS (Lateral Stretch Skipgram) count.
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			count2 += triCnt
		}

		// Function to check if a finger is on the bottom row.
		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// Calculate FSS (Full Scissor Skipgram) count.
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += triCnt
		}

		// Calculate HSS (Half Scissor Skipgram) count.
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += triCnt
		}
	}

	// Calculate percentages for skipgram statistics.
	factor = 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["SFS"] = float64(count1) * factor
	an.Metrics["LSS"] = float64(count2) * factor
	an.Metrics["FSS"] = float64(count3) * factor
	an.Metrics["HSS"] = float64(count4) * factor

	// Calculate trigram statistics (ALT, ROL, ONE, RED).
	count1, count2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.Corpus.Trigrams {
		key1, ok1 := an.Layout.RuneInfo[tri[0]]
		key2, ok2 := an.Layout.RuneInfo[tri[1]]
		key3, ok3 := an.Layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 || !ok3 {
			// Skip trigrams that are not present in the layout.
			continue
		}

		// Calculate ALT (Alternation) count.
		if key1.Hand == key3.Hand && key1.Hand != key2.Hand {
			count1 += triCnt
		}

		// Calculate ROL (Roll) count.
		if key1.Hand != key3.Hand &&
			key1.Finger != key2.Finger && key2.Finger != key3.Finger {
			count2 += triCnt
		}

		// Check if all fingers are on the same hand and in order.
		allSameHand := key1.Hand == key2.Hand && key1.Hand == key3.Hand
		inOrder := (key1.Finger < key2.Finger && key2.Finger < key3.Finger) ||
			(key1.Finger > key2.Finger && key2.Finger > key3.Finger)
		allDiffFingers := key1.Finger != key2.Finger &&
			key1.Finger != key3.Finger &&
			key2.Finger != key3.Finger

		// Calculate ONE (One hand, fingers in order) count.
		if allSameHand && inOrder {
			count3 += triCnt
		}

		// Calculate RED (Redirection) count.
		if allSameHand && allDiffFingers && !inOrder {
			count4 += triCnt
		}
	}

	// Calculate percentages for trigram statistics.
	factor = 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["ALT"] = float64(count1) * factor
	an.Metrics["ROL"] = float64(count2) * factor
	an.Metrics["ONE"] = float64(count3) * factor
	an.Metrics["RED"] = float64(count4) * factor
}
