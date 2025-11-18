package keycraft

import (
	"math"
	"strconv"
)

// MetricsMap groups named metric sets used for different ranking views.
// Each metric set defines which columns appear in the ranking table output.
var MetricsMap = map[string][]string{
	"basic": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", // "LSS", "FSS", "HSS",
		"ALT", "2RL", "3RL", "RED", "RED-WEAK",
		"IN:OUT", "RBL", "FBL", "POH", "FLW",
	},
	"extended": {
		"SFB", "LSB", "FSB", "HSB",
		"SFS", "LSS", "FSS", "HSS",
		"ALT", "ALT-NML", "ALT-SFS",
		"2RL", "2RL-IN", "2RL-OUT", "2RL-SFB",
		"3RL", "3RL-IN", "3RL-OUT", "3RL-SFB",
		"RED", "RED-NML", "RED-WEAK", "RED-SFS",
		"IN:OUT", "RBL", "FBL", "POH", "FLW",
	},
	"fingers": {
		"H0", "H1",
		"R0", "R1", "R2", "R3",
		"F0", "F1", "F2", "F3", "F4",
		"F5", "F6", "F7", "F8", "F9",
	},
}

// DefaultIdealRowLoad
func DefaultIdealRowLoad() *[3]float64 {
	return &[3]float64{
		18.5, // top
		73.0, // home
		8.5,  // bottom
	}
}

// DefaultIdealFingerLoad returns the default target finger load distribution (as percentages).
// Fingers 0-9: left pinky to right pinky. Thumbs (F4, F5) are set to 0 (not counted in main row usage).
// Right-hand loads mirror the left for symmetry.
func DefaultIdealFingerLoad() *[10]float64 {
	return &[10]float64{
		7.5,  // F0
		11.0, // F1
		16.0, // F2
		15.5, // F3
		0.0,  // F4
		0.0,  // F5
		15.5, // F6 (mirror of F3)
		16.0, // F7 (mirror of F2)
		11.0, // F8 (mirror of F1)
		7.5,  // F9 (mirror of F0)
	}
}

// MetricDetails contains detailed analysis results for a single metric.
// Includes per-n-gram counts, distances, and custom attributes (e.g., hand, finger, direction).
type MetricDetails struct {
	Corpus       *Corpus                   // Reference to the corpus being analyzed
	CorpusNGramC uint64                    // Total n-grams of this type in the corpus
	Metric       string                    // Metric name (e.g., "SFB", "LSB", "ALT")
	NGramCount   map[string]uint64         // Frequency of each relevant n-gram
	NGramDist    map[string]float64        // Distance metric for each n-gram
	TotalNGrams  uint64                    // Total count of relevant n-grams
	TotalDist    float64                   // Sum of weighted distances
	Custom       map[string]map[string]any // Additional per-n-gram attributes
}

// TrigramInfo holds a trigram with its frequency for performance.
// KeyInfo is looked up fresh from the layout during analysis to ensure correctness.
type TrigramInfo struct {
	Count uint64  // Frequency of this trigram in the corpus
	Runes [3]rune // The 3 runes in this trigram
}

// Analyser computes ergonomic metrics for a keyboard layout using corpus n-gram frequencies.
// Metrics are stored as percentages or ratios in the Metrics map.
type Analyser struct {
	Layout       *SplitLayout       // The keyboard layout being analyzed
	Corpus       *Corpus            // Text corpus for n-gram frequencies
	IdealRowLoad *[3]float64        // Target row load distribution (percentages for top, home, bottom)
	IdealfgrLoad *[10]float64       // Target finger load distribution (percentages for F0-F9)
	Metrics      map[string]float64 // Computed metrics (e.g., "SFB", "ALT", "FBL")

	// Pre-filtered n-grams (injected by Scorer to avoid redundant filtering)
	relevantTrigrams []TrigramInfo // Only trigrams with all 3 runes on layout
}

// NewAnalyser creates an Analyser and computes all metrics for the given layout.
// If idealRowLoad or idealfgrLoad are nil, uses default target loads.
func NewAnalyser(layout *SplitLayout, corpus *Corpus, idealRowLoad *[3]float64, idealfgrLoad *[10]float64) *Analyser {
	if idealRowLoad == nil {
		idealRowLoad = DefaultIdealRowLoad()
	}
	if idealfgrLoad == nil {
		idealfgrLoad = DefaultIdealFingerLoad()
	}
	an := &Analyser{
		Layout:       layout,
		Corpus:       corpus,
		IdealRowLoad: idealRowLoad,
		IdealfgrLoad: idealfgrLoad,
		Metrics:      make(map[string]float64, 60),
	}
	an.analyseHand()
	an.analyseBigrams()
	an.analyseSkipgrams()
	an.analyseTrigrams()
	return an
}

// analyseHand computes usage metrics for hands, fingers, columns, and rows from unigrams.
// Also calculates finger balance (FBL) as the sum of absolute deviations from ideal loads,
// with special handling for pinkies (only positive deviations count).
func (an *Analyser) analyseHand() {
	var totalUnigramCount uint64
	var pinkyOffHomeCount uint64
	var handCount [2]uint64
	var fingerCount [10]uint64
	var columnCount [12]uint64
	var rowCount [4]uint64

	for uniGr, uniCnt := range an.Corpus.Unigrams {
		key, ok := an.Layout.GetKeyInfo(rune(uniGr))
		if !ok {
			continue
		}

		// Count main row keys (rows 0-2) for balance calculations
		if key.Row < 3 {
			totalUnigramCount += uniCnt
			if (key.Finger == LP || key.Finger == RP) && key.Row != 1 {
				pinkyOffHomeCount += uniCnt
			}
			handCount[key.Hand] += uniCnt
			fingerCount[key.Finger] += uniCnt
			columnCount[key.Column] += uniCnt
		}

		rowCount[key.Row] += uniCnt
	}

	// Convert counts to percentages
	var totFactor float64
	if totalUnigramCount > 0 {
		totFactor = 100 / float64(totalUnigramCount)
	}

	an.Metrics["POH"] = float64(pinkyOffHomeCount) * totFactor
	for i, c := range handCount {
		an.Metrics["H"+strconv.Itoa(i)] = float64(c) * totFactor
	}

	const (
		topRow    = 0
		homeRow   = 1
		bottomRow = 2
		mainRows  = 3
	)

	for i, c := range rowCount {
		ri := "R" + strconv.Itoa(i)
		an.Metrics[ri] = float64(c) * totFactor

		// Calculate row balance (RBL) for main rows (top, home, bottom)
		if i < mainRows {
			diff := an.Metrics[ri] - an.IdealRowLoad[i]
			// Home row: penalize below ideal usage
			// Top/Bottom rows: penalize above ideal usage
			if i == homeRow {
				an.Metrics["RBL"] -= diff
			} else {
				an.Metrics["RBL"] += diff
			}
		}
	}
	// fmt.Printf("%.1f %.1f %.1f - %.1f\n", an.Metrics["R0"], an.Metrics["R1"], an.Metrics["R2"], an.Metrics["RBL"])

	for i, c := range fingerCount {
		fi := "F" + strconv.Itoa(i)
		an.Metrics[fi] = float64(c) * totFactor
		// For pinkies (LP and RP), only add positive deviations to FBL
		if i == int(LP) || i == int(RP) {
			diff := an.Metrics[fi] - an.IdealfgrLoad[i]
			if diff > 0 {
				an.Metrics["FBL"] += diff
			}
		} else {
			an.Metrics["FBL"] += math.Abs(an.Metrics[fi] - an.IdealfgrLoad[i])
		}
	}

	for i, c := range columnCount {
		an.Metrics["C"+strconv.Itoa(i)] = float64(c) * totFactor
	}
}

// analyseBigrams computes bigram-based metrics from corpus frequencies:
//   - SFB: Same Finger Bigrams
//   - LSB: Lateral Stretch Bigrams
//   - FSB: Full Scissor Bigrams
//   - HSB: Half Scissor Bigrams
func (an *Analyser) analyseBigrams() {
	var count1, count2, count3, count4 uint64

	// SFB calculation using pre-computed cache
	for _, sfb := range an.Layout.SFBs {
		bi := Bigram{an.Layout.Runes[sfb.KeyIdx1], an.Layout.Runes[sfb.KeyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count1 += cnt
		}
	}

	for _, lsb := range an.Layout.LSBs {
		bi := Bigram{an.Layout.Runes[lsb.KeyIdx1], an.Layout.Runes[lsb.KeyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count2 += cnt
		}
	}

	for _, sci := range an.Layout.FScissors {
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count3 += cnt
		}
	}

	for _, sci := range an.Layout.HScissors {
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count4 += cnt
		}
	}

	factor := 100 / float64(an.Corpus.TotalBigramsCount)
	an.Metrics["SFB"] = float64(count1) * factor
	an.Metrics["LSB"] = float64(count2) * factor
	an.Metrics["FSB"] = float64(count3) * factor
	an.Metrics["HSB"] = float64(count4) * factor
}

// analyseSkipgrams computes skipgram-based metrics (same patterns as bigrams, but for skipgrams):
//   - SFS: Same Finger Skipgrams
//   - LSS: Lateral Stretch Skipgrams
//   - FSS: Full Scissor Skipgrams
//   - HSS: Half Scissor Skipgrams
func (an *Analyser) analyseSkipgrams() {
	var count1, count2, count3, count4 uint64

	// SFS calculation using pre-computed cache
	for _, sfb := range an.Layout.SFBs {
		skp := Skipgram{an.Layout.Runes[sfb.KeyIdx1], an.Layout.Runes[sfb.KeyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count1 += cnt
		}
	}

	for _, lsb := range an.Layout.LSBs {
		skp := Skipgram{an.Layout.Runes[lsb.KeyIdx1], an.Layout.Runes[lsb.KeyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count2 += cnt
		}
	}

	for _, sci := range an.Layout.FScissors {
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count3 += cnt
		}
	}

	for _, sci := range an.Layout.HScissors {
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count4 += cnt
		}
	}

	factor := 100 / float64(an.Corpus.TotalSkipgramsCount)
	an.Metrics["SFS"] = float64(count1) * factor
	an.Metrics["LSS"] = float64(count2) * factor
	an.Metrics["FSS"] = float64(count3) * factor
	an.Metrics["HSS"] = float64(count4) * factor
}

// analyseTrigrams computes trigram-based flow metrics by categorizing each trigram:
//   - ALT: Alternations (hand switching)
//   - 2RL: Two-key rolls (two fingers on same hand)
//   - 3RL: Three-key rolls (all on same hand, monotonic finger order)
//   - RED: Redirections (all on same hand, non-monotonic)
//
// Each category includes subcategories (e.g., ALT-SFS, 2RL-IN, RED-WEAK).
func (an *Analyser) analyseTrigrams() {
	var rl2SFB, rl2In, rl2Out, altSFS, altNml, rl3SFB, rl3In, rl3Out, redWeak, redSFS, redNml uint64

	// Use pre-filtered trigrams if available (injected by Scorer), otherwise filter on-the-fly
	trigrams := an.relevantTrigrams
	if trigrams == nil {
		// Fallback: pre-filter trigrams now (for non-Scorer callers)
		trigrams = make([]TrigramInfo, 0, len(an.Corpus.Trigrams)/10)
		for tri, cnt := range an.Corpus.Trigrams {
			_, ok0 := an.Layout.GetKeyInfo(tri[0])
			_, ok1 := an.Layout.GetKeyInfo(tri[1])
			_, ok2 := an.Layout.GetKeyInfo(tri[2])
			if ok0 && ok1 && ok2 {
				trigrams = append(trigrams, TrigramInfo{
					Count: cnt,
					Runes: [3]rune{tri[0], tri[1], tri[2]},
				})
			}
		}
	}

	for _, ti := range trigrams {
		cnt := ti.Count
		// Look up fresh KeyInfo from current layout
		k0, _ := an.Layout.GetKeyInfo(ti.Runes[0])
		k1, _ := an.Layout.GetKeyInfo(ti.Runes[1])
		k2, _ := an.Layout.GetKeyInfo(ti.Runes[2])

		// Extract key properties for classification
		h0, h1, h2 := k0.Hand, k1.Hand, k2.Hand
		f0, f1, f2 := k0.Finger, k1.Finger, k2.Finger

		// Classify trigram by hand pattern
		switch h0 {
		case h2:
			if h0 != h1 { // Alternation (two hands alternate)
				if f0 == f2 && k0.Index != k2.Index {
					altSFS += cnt
				} else {
					altNml += cnt
				}
			} else { // One-hand trigram
				switch {
				case f0 == f1 || f1 == f2: // Contains same-finger (SFB/SFS)
					rl3SFB += cnt
				case (f0 < f1) == (f1 < f2): // Monotonic finger sequence (roll)
					if (f0 < f1) == (h0 == LEFT) {
						rl3In += cnt
					} else {
						rl3Out += cnt
					}
				default: // Non-monotonic (redirection)
					if f0 != LI && f0 != RI &&
						f1 != LI && f1 != RI &&
						f2 != LI && f2 != RI {
						redWeak += cnt
					} else if f0 == f2 && k0.Index != k2.Index {
						redSFS += cnt
					} else {
						redNml += cnt
					}
				}
			}
		case h1: // 2-roll with h0 == h1 (inlined for performance)
			switch {
			case f0 == f1: // Same finger
				rl2SFB += cnt
			case (f0 < f1) == (h1 == LEFT):
				rl2In += cnt
			default:
				rl2Out += cnt
			}
		default: // 2-roll with h1 == h2 (inlined for performance)
			switch {
			case f1 == f2: // Same finger
				rl2SFB += cnt
			case (f1 < f2) == (h2 == LEFT):
				rl2In += cnt
			default:
				rl2Out += cnt
			}
		}
	}

	factor := 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["ALT-SFS"] = float64(altSFS) * factor
	an.Metrics["ALT-NML"] = float64(altNml) * factor
	an.Metrics["ALT"] = an.Metrics["ALT-NML"] + an.Metrics["ALT-SFS"]

	an.Metrics["2RL-SFB"] = float64(rl2SFB) * factor
	an.Metrics["2RL-IN"] = float64(rl2In) * factor
	an.Metrics["2RL-OUT"] = float64(rl2Out) * factor
	an.Metrics["2RL"] = an.Metrics["2RL-SFB"] + an.Metrics["2RL-IN"] + an.Metrics["2RL-OUT"]

	an.Metrics["3RL-SFB"] = float64(rl3SFB) * factor
	an.Metrics["3RL-IN"] = float64(rl3In) * factor
	an.Metrics["3RL-OUT"] = float64(rl3Out) * factor
	an.Metrics["3RL"] = an.Metrics["3RL-SFB"] + an.Metrics["3RL-IN"] + an.Metrics["3RL-OUT"]

	an.Metrics["RED-WEAK"] = float64(redWeak) * factor
	an.Metrics["RED-SFS"] = float64(redSFS) * factor
	an.Metrics["RED-NML"] = float64(redNml) * factor
	an.Metrics["RED"] = an.Metrics["RED-NML"] + an.Metrics["RED-SFS"] + an.Metrics["RED-WEAK"]

	an.Metrics["IN:OUT"] = (an.Metrics["2RL-IN"] + an.Metrics["3RL-IN"]) / (an.Metrics["2RL-OUT"] + an.Metrics["3RL-OUT"])
	an.Metrics["FLW"] = an.Metrics["2RL-IN"] + an.Metrics["2RL-OUT"] + an.Metrics["3RL-IN"] + an.Metrics["3RL-OUT"] + an.Metrics["ALT-NML"]
}

// AllMetricsDetails computes detailed analysis for all major metrics.
// Returns a slice of MetricDetails, one for each metric (SFB, LSB, FSB, HSB, SFS, LSS, FSS, HSS, ALT, 2RL, 3RL, RED).
func (an *Analyser) AllMetricsDetails() []*MetricDetails {
	all := make([]*MetricDetails, 0, 30)

	all = append(all, an.SFBiDetails())
	all = append(all, an.LSBiDetails())
	ma, ma2 := an.ScissBiDetails()
	all = append(all, ma, ma2)
	all = append(all, an.SFSkpDetails())
	all = append(all, an.LSSkpDetails())
	ma3, ma4 := an.ScissSkpDetails()
	all = append(all, ma3, ma4)
	ta1, ta2, ta3, ta4 := an.TrigramDetails()
	all = append(all, ta1, ta2, ta3, ta4)

	return all
}

// AllCorpusDetails analyzes the top N bigrams from the corpus, categorizing each by type
// (ALT, SFB, LSB, FSB, HSB, or regular). Useful for understanding corpus characteristics.
func (an *Analyser) AllCorpusDetails(nRows int) []*MetricDetails {
	lsbLookup := make(map[[2]uint8]bool, len(an.Layout.LSBs))
	for _, lsb := range an.Layout.LSBs {
		lsbLookup[[2]uint8{lsb.KeyIdx1, lsb.KeyIdx2}] = true
	}

	fsbLookup := make(map[[2]uint8]bool, len(an.Layout.FScissors))
	for _, fsb := range an.Layout.FScissors {
		fsbLookup[[2]uint8{fsb.keyIdx1, fsb.keyIdx2}] = true
	}

	hsbLookup := make(map[[2]uint8]bool, len(an.Layout.HScissors))
	for _, hsb := range an.Layout.HScissors {
		hsbLookup[[2]uint8{hsb.keyIdx1, hsb.keyIdx2}] = true
	}

	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "CORPUS BIGRAMS",
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
		Custom:       make(map[string]map[string]any),
	}

	topBigrams := an.Corpus.TopBigrams(nRows)
	for _, cb := range topBigrams {
		bi := cb.Key
		biStr := bi.String()
		key1, ok1 := an.Layout.GetKeyInfo(bi[0])
		key2, ok2 := an.Layout.GetKeyInfo(bi[1])

		ma.NGramCount[biStr] += cb.Count
		ma.TotalNGrams += cb.Count
		if _, ok := ma.Custom[biStr]; !ok {
			ma.Custom[biStr] = make(map[string]any)
		}

		biIdx := [2]uint8{key1.Index, key2.Index}
		if !ok1 || !ok2 {
			ma.Custom[biStr]["Type"] = "Not found"
		} else if key1.Hand != key2.Hand {
			ma.Custom[biStr]["Type"] = "ALT"
			// tt
		} else if key1.Finger == key2.Finger {
			if key1.Index == key2.Index {
				ma.Custom[biStr]["Type"] = "Repeat"
			} else {
				ma.Custom[biStr]["Type"] = "SFB"
			}
		} else if lsbLookup[biIdx] {
			ma.Custom[biStr]["Type"] = "LSB"
		} else if fsbLookup[biIdx] {
			ma.Custom[biStr]["Type"] = "FSB"
		} else if hsbLookup[biIdx] {
			ma.Custom[biStr]["Type"] = "HSB"
		} else {
			ma.Custom[biStr]["Type"] = "-"
		}
	}

	return []*MetricDetails{ma}
}

// SFBiDetails performs detailed Same Finger Bigram (SFB) analysis.
// Identifies bigrams typed with the same finger on different keys, reporting frequency,
// distance, hand, finger, and row distance for each.
func (an *Analyser) SFBiDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "SFB",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for bi, biCnt := range an.Corpus.Bigrams {
		biStr := bi.String()
		key1, ok1 := an.Layout.GetKeyInfo(bi[0])
		key2, ok2 := an.Layout.GetKeyInfo(bi[1])

		if !ok1 || !ok2 {
			// ma.Unsupported[biStr] += biCnt
			continue
		}

		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			ma.NGramCount[biStr] = biCnt
			ma.TotalNGrams += biCnt
			kpDist := an.Layout.Distance(key1.Index, key2.Index)
			ma.NGramDist[biStr] = kpDist.Distance
			ma.TotalDist += kpDist.Distance * float64(biCnt)

			if _, ok := ma.Custom[biStr]; !ok {
				ma.Custom[biStr] = make(map[string]any)
			}
			ma.Custom[biStr]["Hd"] = key1.Hand + 1
			ma.Custom[biStr]["Fgr"] = key1.Finger + 1
			ma.Custom[biStr]["Δrow"] = kpDist.RowDist
		}
	}

	return ma
}

// LSBiDetails performs detailed Lateral Stretch Bigram (LSB) analysis.
// Evaluates preidentified lateral stretch pairs, reporting frequency, distance, hand, and column distance.
func (an *Analyser) LSBiDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "LSB",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for _, lsb := range an.Layout.LSBs {
		rune1 := an.Layout.Runes[lsb.KeyIdx1]
		bi := Bigram{rune1, an.Layout.Runes[lsb.KeyIdx2]}
		if biCnt, ok := an.Corpus.Bigrams[bi]; ok {
			biStr := bi.String()

			ma.NGramCount[biStr] = biCnt
			ma.TotalNGrams += biCnt
			kpDist := an.Layout.Distance(lsb.KeyIdx1, lsb.KeyIdx2)
			ma.NGramDist[biStr] = kpDist.Distance
			ma.TotalDist += kpDist.Distance * float64(biCnt)

			if _, ok := ma.Custom[biStr]; !ok {
				ma.Custom[biStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma.Custom[biStr]["Hd"] = key1.Hand + 1
			ma.Custom[biStr]["Δcol"] = kpDist.ColDist
		}
	}

	return ma
}

// ScissBiDetails analyzes full and half scissor bigrams separately.
// Returns two MetricDetails: FSB (full scissors, 2-row jumps) and HSB (half scissors, 1-row jumps).
// Includes frequency, distance, hand, row/column distance, and angle for each.
func (an *Analyser) ScissBiDetails() (*MetricDetails, *MetricDetails) {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "FSB",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}
	ma2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "HSB",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for _, sci := range an.Layout.FScissors {
		rune1 := an.Layout.Runes[sci.keyIdx1]
		bi := Bigram{rune1, an.Layout.Runes[sci.keyIdx2]}
		if biCnt, ok := an.Corpus.Bigrams[bi]; ok {
			biStr := bi.String()
			dist := an.Layout.Distance(sci.keyIdx1, sci.keyIdx2).Distance
			ma.NGramCount[biStr] = biCnt
			ma.TotalNGrams += biCnt
			ma.NGramDist[biStr] = dist
			ma.TotalDist += dist * float64(biCnt)

			if _, ok := ma.Custom[biStr]; !ok {
				ma.Custom[biStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma.Custom[biStr]["Hd"] = key1.Hand + 1
			ma.Custom[biStr]["Δrow"] = sci.rowDist
			ma.Custom[biStr]["Δcol"] = sci.colDist
			ma.Custom[biStr]["Angle"] = sci.angle
		}
	}

	for _, sci := range an.Layout.HScissors {
		rune1 := an.Layout.Runes[sci.keyIdx1]
		bi := Bigram{rune1, an.Layout.Runes[sci.keyIdx2]}
		if biCnt, ok := an.Corpus.Bigrams[bi]; ok {
			biStr := bi.String()
			dist := an.Layout.Distance(sci.keyIdx1, sci.keyIdx2).Distance
			ma2.NGramCount[biStr] = biCnt
			ma2.TotalNGrams += biCnt
			ma2.NGramDist[biStr] = dist
			ma2.TotalDist += dist * float64(biCnt)

			if _, ok := ma.Custom[biStr]; !ok {
				ma2.Custom[biStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma2.Custom[biStr]["Hd"] = key1.Hand + 1
			ma2.Custom[biStr]["Δrow"] = sci.rowDist
			ma2.Custom[biStr]["Δcol"] = sci.colDist
			ma2.Custom[biStr]["Angle"] = sci.angle
		}
	}

	return ma, ma2
}

// SFSkpDetails performs detailed Same Finger Skipgram (SFS) analysis.
// Similar to SFBiDetails but for skipgrams (1st and 3rd characters of trigrams).
func (an *Analyser) SFSkpDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "SFS",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for skp, skpCnt := range an.Corpus.Skipgrams {
		skpStr := skp.String()
		key1, ok1 := an.Layout.GetKeyInfo(skp[0])
		key2, ok2 := an.Layout.GetKeyInfo(skp[1])

		if !ok1 || !ok2 {
			// ma.Unsupported[skpStr] += skpCnt
			continue
		}

		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			ma.NGramCount[skpStr] = skpCnt
			ma.TotalNGrams += skpCnt
			kpDist := an.Layout.Distance(key1.Index, key2.Index)
			ma.NGramDist[skpStr] = kpDist.Distance
			ma.TotalDist += kpDist.Distance * float64(skpCnt)

			if _, ok := ma.Custom[skpStr]; !ok {
				ma.Custom[skpStr] = make(map[string]any)
			}
			ma.Custom[skpStr]["Hd"] = key1.Hand + 1
			ma.Custom[skpStr]["Fgr"] = key1.Finger + 1
			ma.Custom[skpStr]["Δrow"] = kpDist.RowDist
		}
	}

	return ma
}

// LSSkpDetails performs detailed Lateral Stretch Skipgram (LSS) analysis.
// Similar to LSBiDetails but for skipgrams.
func (an *Analyser) LSSkpDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "LSS",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for _, lsb := range an.Layout.LSBs {
		rune1 := an.Layout.Runes[lsb.KeyIdx1]
		skp := Skipgram{rune1, an.Layout.Runes[lsb.KeyIdx2]}
		if skpCnt, ok := an.Corpus.Skipgrams[skp]; ok {
			skpStr := skp.String()

			ma.NGramCount[skpStr] = skpCnt
			ma.TotalNGrams += skpCnt
			kpDist := an.Layout.Distance(uint8(lsb.KeyIdx1), uint8(lsb.KeyIdx2))
			ma.NGramDist[skpStr] = kpDist.Distance
			ma.TotalDist += kpDist.Distance * float64(skpCnt)

			if _, ok := ma.Custom[skpStr]; !ok {
				ma.Custom[skpStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma.Custom[skpStr]["Hd"] = key1.Hand + 1
			ma.Custom[skpStr]["Δcol"] = kpDist.ColDist
		}
	}

	return ma
}

// ScissSkpDetails analyzes full and half scissor skipgrams separately.
// Similar to ScissBiDetails but for skipgrams.
func (an *Analyser) ScissSkpDetails() (*MetricDetails, *MetricDetails) {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "FSS",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}
	ma2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "HSS",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for _, sci := range an.Layout.FScissors {
		rune1 := an.Layout.Runes[sci.keyIdx1]
		skp := Skipgram{rune1, an.Layout.Runes[sci.keyIdx2]}
		if skpCnt, ok := an.Corpus.Skipgrams[skp]; ok {
			skpStr := skp.String()

			dist := an.Layout.Distance(sci.keyIdx1, sci.keyIdx2).Distance
			ma.NGramCount[skpStr] = skpCnt
			ma.TotalNGrams += skpCnt
			ma.NGramDist[skpStr] = dist
			ma.TotalDist += dist * float64(skpCnt)

			if _, ok := ma.Custom[skpStr]; !ok {
				ma.Custom[skpStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma.Custom[skpStr]["Hd"] = key1.Hand + 1
			ma.Custom[skpStr]["Δrow"] = sci.rowDist
			ma.Custom[skpStr]["Δcol"] = sci.colDist
			ma.Custom[skpStr]["Angle"] = sci.angle
		}
	}

	for _, sci := range an.Layout.HScissors {
		rune1 := an.Layout.Runes[sci.keyIdx1]
		skp := Skipgram{rune1, an.Layout.Runes[sci.keyIdx2]}
		if skpCnt, ok := an.Corpus.Skipgrams[skp]; ok {
			skpStr := skp.String()

			dist := an.Layout.Distance(sci.keyIdx1, sci.keyIdx2).Distance
			ma2.NGramCount[skpStr] = skpCnt
			ma2.TotalNGrams += skpCnt
			ma2.NGramDist[skpStr] = dist
			ma2.TotalDist += dist * float64(skpCnt)

			if _, ok := ma.Custom[skpStr]; !ok {
				ma2.Custom[skpStr] = make(map[string]any)
			}
			key1, _ := an.Layout.GetKeyInfo(rune1)
			ma2.Custom[skpStr]["Hd"] = key1.Hand + 1
			ma2.Custom[skpStr]["Δrow"] = sci.rowDist
			ma2.Custom[skpStr]["Δcol"] = sci.colDist
			ma2.Custom[skpStr]["Angle"] = sci.angle
		}
	}

	return ma, ma2
}

// TrigramDetails categorizes all trigrams into flow patterns:
//   - ALT: Alternations (hand switches), with subcategories ALT-NML and ALT-SFS
//   - 2RL: Two-key rolls, with directions (IN, OUT, SFB)
//   - 3RL: Three-key rolls, with directions (IN, OUT, SFB)
//   - RED: Redirections, with subcategories (NML, SFS, WEAK)
//
// Returns four MetricDetails, one for each category.
func (an *Analyser) TrigramDetails() (*MetricDetails, *MetricDetails, *MetricDetails, *MetricDetails) {
	alt := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "ALT",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}
	rl2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "2RL",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}
	rl3 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "3RL",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}
	red := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "RED",
		// Unsupported:  make(map[string]uint64),
		NGramCount: make(map[string]uint64),
		NGramDist:  make(map[string]float64),
		Custom:     make(map[string]map[string]any),
	}

	for tri, cnt := range an.Corpus.Trigrams {
		triStr := tri.String()
		r0, ok0 := an.Layout.GetKeyInfo(tri[0])
		r1, ok1 := an.Layout.GetKeyInfo(tri[1])
		r2, ok2 := an.Layout.GetKeyInfo(tri[2])
		if !ok0 || !ok1 || !ok2 {
			// alt.Unsupported[triStr] += cnt
			continue
		}

		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger

		// Helper to classify and record 2-key rolls
		add2Roll := func(fA, fB uint8) {
			rl2.NGramCount[triStr] = cnt
			rl2.TotalNGrams += cnt
			switch {
			case fA == fB:
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Dir"] = "SFB"
			case (fA < fB) == (h1 == LEFT):
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Dir"] = "IN"
			default:
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Dir"] = "OUT"
			}
		}

		// Classify and record trigram
		switch h0 {
		case h2:
			if h0 != h1 { // Alternation
				alt.NGramCount[triStr] = cnt
				alt.TotalNGrams += cnt
				if _, ok := alt.Custom[triStr]; !ok {
					alt.Custom[triStr] = make(map[string]any)
				}
				if f0 == f2 && r0.Index != r2.Index {
					alt.Custom[triStr]["Dir"] = "SFS"
				} else {
					alt.Custom[triStr]["Dir"] = "NML"
				}
			} else { // One-hand pattern
				if f0 == f1 || f1 == f2 {
					rl3.NGramCount[triStr] = cnt
					rl3.TotalNGrams += cnt
					if _, ok := rl3.Custom[triStr]; !ok {
						rl3.Custom[triStr] = make(map[string]any)
					}
					rl3.Custom[triStr]["Dir"] = "SFB"
				} else if (f0 < f1) == (f1 < f2) {
					rl3.NGramCount[triStr] = cnt
					rl3.TotalNGrams += cnt
					if _, ok := rl3.Custom[triStr]; !ok {
						rl3.Custom[triStr] = make(map[string]any)
					}
					if (f0 < f1) == (h0 == LEFT) {
						rl3.Custom[triStr]["Dir"] = "IN"
					} else {
						rl3.Custom[triStr]["Dir"] = "OUT"
					}
				} else {
					red.NGramCount[triStr] = cnt
					red.TotalNGrams += cnt
					if _, ok := red.Custom[triStr]; !ok {
						red.Custom[triStr] = make(map[string]any)
					}
					if f0 != LI && f0 != RI &&
						f1 != LI && f1 != RI &&
						f2 != LI && f2 != RI {
						red.Custom[triStr]["Dir"] = "WEAK"
					} else if f0 == f2 && r0.Index != r2.Index {
						red.Custom[triStr]["Dir"] = "SFS"
					} else {
						red.Custom[triStr]["Dir"] = "NML"
					}
				}
			}
		case h1:
			add2Roll(f0, f1)
		default: // h1 == h2
			add2Roll(f1, f2)
		}
	}

	return alt, rl2, rl3, red
}
