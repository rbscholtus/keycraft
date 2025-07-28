// Package layout provides functionality for analyzing keyboard layouts.
package layout

type Analyser struct {
	// Reference to the analysed layout.
	layout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	corpus *Corpus
	//
	Metrics map[string]float32
}

func NewAnalyser(layout *SplitLayout, corpus *Corpus) *Analyser {
	a := &Analyser{
		layout:  layout,
		corpus:  corpus,
		Metrics: make(map[string]float32),
	}
	// godump.DumpJSON(layout.RuneInfo)
	a.quickAnalysis()
	return a
}

func (a *Analyser) quickAnalysis() {
	var count1, cound2, count3, count4 uint64
	var factor float32

	// Calculate basic bigrams first
	for bi, biCnt := range a.corpus.Bigrams {
		key1, ok1 := a.layout.RuneInfo[bi[0]]
		key2, ok2 := a.layout.RuneInfo[bi[1]]
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
	factor = 100 / float32(a.corpus.TotalBigramsNoSpace)
	a.Metrics["SFB"] = float32(count1) * factor
	a.Metrics["LSB"] = float32(cound2) * factor
	a.Metrics["FSB"] = float32(count3) * factor
	a.Metrics["HSB"] = float32(count4) * factor

	// Calculate the Skipgram versions
	count1, cound2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range a.corpus.Trigrams {
		key1, ok1 := a.layout.RuneInfo[tri[0]]
		key2, ok2 := a.layout.RuneInfo[tri[2]]
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
	factor = 100 / float32(a.corpus.TotalTrigramsCount)
	a.Metrics["SFS"] = float32(count1) * factor
	a.Metrics["LSS"] = float32(cound2) * factor
	a.Metrics["FSS"] = float32(count3) * factor
	a.Metrics["HSS"] = float32(count4) * factor

	// Calculate trigram stats
	count1, cound2, count3, count4 = 0, 0, 0, 0

	for tri, triCnt := range a.corpus.Trigrams {
		key1, ok1 := a.layout.RuneInfo[tri[0]]
		key2, ok2 := a.layout.RuneInfo[tri[1]]
		key3, ok3 := a.layout.RuneInfo[tri[2]]
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
	factor = 100 / float32(a.corpus.TotalTrigramsCount)
	a.Metrics["ALT"] = float32(count1) * factor
	a.Metrics["ROL"] = float32(cound2) * factor
	a.Metrics["ONE"] = float32(count3) * factor
	a.Metrics["RED"] = float32(count4) * factor
}
