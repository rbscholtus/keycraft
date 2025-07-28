// Package layout provides functionality for analyzing keyboard layouts.
package layout

type HandUsageAnalysis struct {
	HandUsage   [2]float32
	FingerUsage [10]float32
	ColumnUsage [12]float32
	RowUsage    [4]float32
}

type Analyser struct {
	// Reference to the analysed layout.
	layout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	corpus *Corpus
	//
	HandUsage HandUsageAnalysis
	//
	Metrics map[string]float32
}

func NewAnalyser(layout *SplitLayout, corpus *Corpus) *Analyser {
	a := &Analyser{
		layout:  layout,
		corpus:  corpus,
		Metrics: make(map[string]float32),
	}
	a.quickHandAnalysis()
	a.quickMetricAnalysis()
	return a
}

func (an *Analyser) quickHandAnalysis() {
	var totalUnigramCount uint64
	var handCount [2]uint64
	var fingerCount [10]uint64
	var columnCount [12]uint64
	var rowCount [4]uint64

	// Iterate over unigrams in the corpus and calculate usage statistics
	for uniGr, uniCnt := range an.corpus.Unigrams {
		key, ok := an.layout.RuneInfo[rune(uniGr)]
		if !ok {
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

	// Calculate the percentages
	factor := 100 / float32(totalUnigramCount)
	for i, c := range handCount {
		an.HandUsage.HandUsage[i] = float32(c) * factor
	}
	for i, c := range fingerCount {
		an.HandUsage.FingerUsage[i] = float32(c) * factor
	}
	for i, c := range columnCount {
		an.HandUsage.ColumnUsage[i] = float32(c) * factor
	}
	for i, c := range rowCount {
		an.HandUsage.RowUsage[i] = float32(c) * factor
	}
}

func (an *Analyser) quickMetricAnalysis() {
	var factor float32
	var count1, cound2, count3, count4 uint64

	// Calculate basic bigrams first
	for bi, biCnt := range an.corpus.Bigrams {
		key1, ok1 := an.layout.RuneInfo[bi[0]]
		key2, ok2 := an.layout.RuneInfo[bi[1]]
		if !ok1 || !ok2 {
			continue
		}

		// All bigram stats are on 1 hand
		if key1.Hand != key2.Hand {
			continue
		}

		// SFB
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += biCnt
			continue
		}

		// LSB
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			cound2 += biCnt
		}

		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// FSB
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += biCnt
		}

		// HSB
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += biCnt
		}
	}

	// percentages
	factor = 100 / float32(an.corpus.TotalBigramsNoSpace)
	an.Metrics["SFB"] = float32(count1) * factor
	an.Metrics["LSB"] = float32(cound2) * factor
	an.Metrics["FSB"] = float32(count3) * factor
	an.Metrics["HSB"] = float32(count4) * factor

	// Calculate the Skipgram versions
	count1, cound2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.corpus.Trigrams {
		key1, ok1 := an.layout.RuneInfo[tri[0]]
		key2, ok2 := an.layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 {
			continue
		}

		// All trigram stats are on 1 hand
		if key1.Hand != key2.Hand {
			continue
		}

		// SFS
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += triCnt
			continue
		}

		// LSS
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			cound2 += triCnt
		}

		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// FSS
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += triCnt
		}

		// HSS
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += triCnt
		}
	}

	// percentages
	factor = 100 / float32(an.corpus.TotalTrigramsCount)
	an.Metrics["SFS"] = float32(count1) * factor
	an.Metrics["LSS"] = float32(cound2) * factor
	an.Metrics["FSS"] = float32(count3) * factor
	an.Metrics["HSS"] = float32(count4) * factor

	// Calculate trigram stats
	count1, cound2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range an.corpus.Trigrams {
		key1, ok1 := an.layout.RuneInfo[tri[0]]
		key2, ok2 := an.layout.RuneInfo[tri[1]]
		key3, ok3 := an.layout.RuneInfo[tri[2]]
		if !ok1 || !ok2 || !ok3 {
			continue
		}

		// ALT
		if key1.Hand == key3.Hand && key1.Hand != key2.Hand {
			count1 += triCnt
		}

		// ROL
		if key1.Hand != key3.Hand &&
			key1.Finger != key2.Finger && key2.Finger != key3.Finger {
			cound2 += triCnt
		}

		allSameHand := key1.Hand == key2.Hand && key1.Hand == key3.Hand
		inOrder := (key1.Finger < key2.Finger && key2.Finger < key3.Finger) ||
			(key1.Finger > key2.Finger && key2.Finger > key3.Finger)
		allDiffFingers := key1.Finger != key2.Finger &&
			key1.Finger != key3.Finger &&
			key2.Finger != key3.Finger

		// ONE
		if allSameHand && inOrder {
			count3 += triCnt
		}

		// RED
		if allSameHand && allDiffFingers && !inOrder {
			count4 += triCnt
		}
	}

	// percentages
	factor = 100 / float32(an.corpus.TotalTrigramsCount)
	an.Metrics["ALT"] = float32(count1) * factor
	an.Metrics["ROL"] = float32(cound2) * factor
	an.Metrics["ONE"] = float32(count3) * factor
	an.Metrics["RED"] = float32(count4) * factor
}
