// Package layout provides functionality for analyzing keyboard layouts.
package layout

import (
	"math"
	"strconv"
)

var idealFingerLoad = map[string]float64{
	"F0": 8.5,
	"F1": 10.5,
	"F2": 15.5,
	"F3": 15.5,
	"F4": 0.0,
	"F5": 0.0,
	"F6": 15.5,
	"F7": 15.5,
	"F8": 10.5,
	"F9": 8.5,
	// "H0": 50.0,
	// "H1": 50.0,
}

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
	// Metrics holds all metrics about the layout.
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
// All metrics exclude n-grams with space characters.
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
		handCount[key.Hand] += uniCnt
		fingerCount[key.Finger] += uniCnt
		columnCount[key.Column] += uniCnt
		rowCount[key.Row] += uniCnt
	}

	// Calculate the percentages as well as cumulative finger usage deviation off the ideal
	factor := 100 / float64(totalUnigramCount)
	for i, c := range handCount {
		an.Metrics["H"+strconv.Itoa(i)] = float64(c) * factor
	}
	for i, c := range fingerCount {
		fi := "F" + strconv.Itoa(i)
		an.Metrics[fi] = float64(c) * factor
		an.Metrics["FBL"] += math.Abs(an.Metrics[fi] - idealFingerLoad[fi])
	}
	for i, c := range columnCount {
		an.Metrics["C"+strconv.Itoa(i)] = float64(c) * factor
	}
	for i, c := range rowCount {
		an.Metrics["R"+strconv.Itoa(i)] = float64(c) * factor
	}
}

// quickMetricAnalysis computes a core set of ergonomic and motion pattern metrics for the layout.
// It analyzes bigrams, skipgrams, and trigrams from the corpus using the current layout mapping.
// The metrics are grouped as follows:
//
// Bigram metrics:
// - SFB (Same Finger Bigram): % of bigrams typed with the same finger.
// - LSB (Lateral Stretch Bigram): % of bigrams involving a lateral stretch (index ↔ pinky).
// - FSB (Full Scissor Bigram): % requiring large vertical movement.
// - HSB (Half Scissor Bigram): % requiring smaller vertical movement.
//
// Skipgram metrics (compare first and last keys in a trigram):
// - SFS (Same Finger Skipgram): % of skipgrams typed with the same finger.
// - LSS (Lateral Stretch Skipgram): % involving a lateral stretch.
// - FSS (Full Scissor Skipgram): % requiring large vertical movement.
// - HSS (Half Scissor Skipgram): % requiring smaller vertical movement.
//
// Trigram / sequence metrics:
// - ALT: Alternation between hands.
// - ALT-SFS: Alternation patterns that are also same-finger skipgrams.
// - 2RL-IN / 2RL-OUT: Two-key rolls inward/outward between adjacent fingers.
// - 2RL-SF: Two-key rolls using the same finger.
// - 3RL-IN / 3RL-OUT: Three-key inward/outward rolls on one hand.
// - 3RL-SF: Three-key rolls with a same-finger occurrence.
// - RED: Redirections (direction changes on one hand).
// - RED-SFS: Redirections that are also same-finger skipgrams.
// - RED-BAD: Less ergonomic redirections (no pinky involvement).
// - ALT: Sum of ALT-OTH and ALT-SFS, total alternation percentage.
// - 2RL: Sum of 2RL-IN and 2RL-OUT, total two-key rolls.
// - 3RL: Sum of 3RL-IN and 3RL-OUT, total three-key rolls.
// - RED: Sum of RED-OTH, RED-SFS, and RED-BAD, total redirections.
// - IN:OUT: Ratio of inward to outward rolls (2-key and 3-key combined).
//
// All percentages are stored in an.Metrics for later ranking and comparison.
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

		// Calculate SFB (Same Finger Bigram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += biCnt
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
	for _, sci := range an.Layout.Scissors {
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

	for skp, skpCnt := range an.Corpus.Skipgrams {
		key1, ok1 := an.Layout.RuneInfo[skp[0]]
		key2, ok2 := an.Layout.RuneInfo[skp[1]]
		if !ok1 || !ok2 {
			// Skip skipgrams that are not present in the layout.
			continue
		}

		// Calculate SFS (Same Finger Skipgram) count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += skpCnt
		}
	}

	// Calculate LSS count.
	for _, lsb := range an.Layout.LSBs {
		skp := Skipgram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count2 += cnt
		}
	}

	// Calculate Scissor skipgrams counts.
	for _, sci := range an.Layout.Scissors {
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			if sci.rowDist > 1.5 {
				count3 += cnt
			} else {
				count4 += cnt
			}
		}
	}

	// Calculate percentages for skipgram statistics.
	factor = 100 / float64(an.Corpus.TotalSkipgramsCount)
	an.Metrics["SFS"] = float64(count1) * factor
	an.Metrics["LSS"] = float64(count2) * factor
	an.Metrics["FSS"] = float64(count3) * factor
	an.Metrics["HSS"] = float64(count4) * factor

	// Calculate trigram statistics (ALT, ROL, ONE, RED).
	// var countSkipped uint64
	count1, count2, count3, count4 = 0, 0, 0, 0
	var count5, count6, count7, count8, count9, count10, count11 uint64

	for tri, cnt := range an.Corpus.Trigrams {
		// Cross-hand trigrams
		add2Roll := func(h, fA, fB uint8) {
			switch {
			case fA == fB:
				count1 += cnt
			case (fA < fB) == (h == LEFT):
				count2 += cnt
			default:
				count3 += cnt
			}
		}

		r0, ok0 := an.Layout.RuneInfo[tri[0]]
		r1, ok1 := an.Layout.RuneInfo[tri[1]]
		r2, ok2 := an.Layout.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			//countSkipped += cnt
			continue
		}

		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger
		diffIdx02 := r0.Index != r2.Index

		//lint:ignore QF1003 use if for clarity
		if h0 == h2 {
			if h0 != h1 {
				// ALT or ALT-SFS
				if f0 == f2 && diffIdx02 {
					count4 += cnt
				} else {
					count5 += cnt
				}
			} else {
				// One-hand trigrams
				switch {
				case f0 == f1 || f1 == f2:
					count6 += cnt // 3-roll with same finger
				case (f0 < f1) == (f1 < f2):
					if (f0 < f1) == (h0 == LEFT) {
						count7 += cnt // 3-roll in
					} else {
						count8 += cnt // 3-roll out
					}
				default:
					if f0 != LI && f0 != RI &&
						f1 != LI && f1 != RI &&
						f2 != LI && f2 != RI {
						count9 += cnt // bad/weak redirect (without pinky)
					} else if f0 == f2 && diffIdx02 {
						count10 += cnt // redirect skipgram
					} else {
						count11 += cnt // redirect
					}
				}
			}
		} else if h0 == h1 {
			add2Roll(h0, f0, f1)
		} else { // h1 == h2
			add2Roll(h1, f1, f2)
		}

		factor = 100 / float64(an.Corpus.TotalTrigramsCount)
		an.Metrics["2RL-SF"] = float64(count1) * factor
		an.Metrics["2RL-IN"] = float64(count2) * factor
		an.Metrics["2RL-OUT"] = float64(count3) * factor
		an.Metrics["ALT-SFS"] = float64(count4) * factor
		an.Metrics["ALT-OTH"] = float64(count5) * factor
		an.Metrics["3RL-SF"] = float64(count6) * factor
		an.Metrics["3RL-IN"] = float64(count7) * factor
		an.Metrics["3RL-OUT"] = float64(count8) * factor
		an.Metrics["RED-BAD"] = float64(count9) * factor
		an.Metrics["RED-SFS"] = float64(count10) * factor
		an.Metrics["RED-OTH"] = float64(count11) * factor

		an.Metrics["ALT"] = an.Metrics["ALT-OTH"] + an.Metrics["ALT-SFS"]
		an.Metrics["2RL"] = an.Metrics["2RL-IN"] + an.Metrics["2RL-OUT"]
		an.Metrics["3RL"] = an.Metrics["3RL-IN"] + an.Metrics["3RL-OUT"]
		an.Metrics["RED"] = an.Metrics["RED-OTH"] + an.Metrics["RED-SFS"] + an.Metrics["RED-BAD"]
		an.Metrics["IN:OUT"] = (an.Metrics["2RL-IN"] + an.Metrics["3RL-IN"]) / (an.Metrics["2RL-OUT"] + an.Metrics["3RL-OUT"])
	}
}

// quickMetricAnalysis calculates basic metrics about the layout, based on Keysolve-web.
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

	for skp, skpCnt := range an.Corpus.Skipgrams {
		key1, ok1 := an.Layout.RuneInfo[skp[0]]
		key2, ok2 := an.Layout.RuneInfo[skp[1]]
		if !ok1 || !ok2 {
			// Skip skipgrams that are not present in the layout.
			continue
		}

		// First and third character are on 1 hand.
		if key1.Hand != key2.Hand {
			continue
		}

		// Calculate SFS count.
		if key1.Finger == key2.Finger && key1 != key2 {
			count1 += skpCnt
			continue
		}

		// Calculate LSS count.
		if (key1.Column == 5 || key1.Column == 6 || key2.Column == 5 || key2.Column == 6) &&
			(key1.Column == 3 || key1.Column == 8 || key2.Column == 3 || key2.Column == 8) {
			count2 += skpCnt
		}

		// Function to check if a finger is on the bottom row.
		bRow := func(fgr uint8) bool {
			return fgr == 1 || fgr == 2 || fgr == 7 || fgr == 8
		}

		// Calculate FSS count.
		if (key2.Row-key1.Row == 2 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 2 && bRow(key1.Finger)) {
			count3 += skpCnt
		}

		// Calculate HSS count.
		if (key2.Row-key1.Row == 1 && bRow(key2.Finger)) ||
			(key1.Row-key2.Row == 1 && bRow(key1.Finger)) {
			count4 += skpCnt
		}
	}

	// Calculate percentages for skipgram statistics.
	factor = 100 / float64(an.Corpus.TotalSkipgramsCount)
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
