package keycraft

import (
	"math"
	"strconv"
)

// DefaultIdealFingerLoad returns the default ideal loads for F0..F9.
// F4 and F5 are 0.0; F6..F9 are mirrored from F3..F0.
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

// MetricDetails contains detailed results for a single metric, including counts of relevant n-grams, unsupported n-grams, weighted distance values (row, column, or Euclidean), and totals.
type MetricDetails struct {
	Corpus       *Corpus
	CorpusNGramC uint64
	Metric       string
	// Unsupported  map[string]uint64
	NGramCount  map[string]uint64
	NGramDist   map[string]float64
	TotalNGrams uint64
	TotalDist   float64
	Custom      map[string]map[string]any
}

// Analyser performs ergonomic analysis on a given layout using a corpus. It computes both quick percentage-based metrics and detailed breakdowns of relevant n-grams and distances.
type Analyser struct {
	// Reference to the analysed Layout.
	Layout *SplitLayout
	// Reference to the Corpus used to analyse the layout.
	Corpus *Corpus
	// IdealfgrLoad holds the ideal percentages for fingers 0..9.
	// Use [10]float64 with keys 0..9 (F0..F9). F4/F5 normally 0.
	IdealfgrLoad *[10]float64
	// Metrics holds all metrics about the layout.
	Metrics map[string]float64
}

// NewAnalyser constructs an Analyser for the given layout and corpus.
// Pass an ideal map[uint8]float64 (keys 0..9). If ideal is nil, defaults are used.
func NewAnalyser(layout *SplitLayout, corpus *Corpus, idealfgrLoad *[10]float64) *Analyser {
	if idealfgrLoad == nil {
		idealfgrLoad = DefaultIdealFingerLoad()
	}
	an := &Analyser{
		Layout:       layout,
		Corpus:       corpus,
		IdealfgrLoad: idealfgrLoad,
		Metrics:      make(map[string]float64),
	}
	an.analyseHand()
	an.analyseBigrams()
	an.analyseSkipgrams()
	an.analyseTrigrams()
	return an
}

// analyseHand computes hand, finger, column, and row usage metrics from unigrams.
// Finger balance (FBL) is calculated as the cumulative deviation from the idealFingerLoad distribution.
func (an *Analyser) analyseHand() {
	var totalUnigramCount uint64
	var pinkyOffHomeCount uint64
	var handCount [2]uint64
	var fingerCount [10]uint64
	var columnCount [12]uint64
	var rowCount [4]uint64

	for uniGr, uniCnt := range an.Corpus.Unigrams {
		key, ok := an.Layout.RuneInfo[rune(uniGr)]
		if !ok {
			continue
		}

		// only include non-thumb keys
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

	// total-based factor for metrics that relate to whole-corpus percentages
	var totFactor float64
	if totalUnigramCount > 0 {
		totFactor = 100 / float64(totalUnigramCount)
	}

	an.Metrics["POH"] = float64(pinkyOffHomeCount) * totFactor
	for i, c := range handCount {
		an.Metrics["H"+strconv.Itoa(i)] = float64(c) * totFactor
	}
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
	for i, c := range rowCount {
		an.Metrics["R"+strconv.Itoa(i)] = float64(c) * totFactor
	}
}

// analyseBigrams computes bigram-based metrics: SFB, LSB, FSB, and HSB.
func (an *Analyser) analyseBigrams() {
	var count1, count2, count3, count4 uint64
	for bi, biCnt := range an.Corpus.Bigrams {
		key1, ok1 := an.Layout.RuneInfo[bi[0]]
		key2, ok2 := an.Layout.RuneInfo[bi[1]]
		if !ok1 || !ok2 {
			continue
		}
		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			count1 += biCnt
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

// analyseSkipgrams computes skipgram-based metrics: SFS, LSS, FSS, and HSS.
func (an *Analyser) analyseSkipgrams() {
	var count1, count2, count3, count4 uint64
	for skp, skpCnt := range an.Corpus.Skipgrams {
		key1, ok1 := an.Layout.RuneInfo[skp[0]]
		key2, ok2 := an.Layout.RuneInfo[skp[1]]
		if !ok1 || !ok2 {
			continue
		}
		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			count1 += skpCnt
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

// analyseTrigrams computes trigram-based metrics: ALT (alternations), 2RL (two-key rolls), 3RL (three-key rolls), and RED (redirections).
func (an *Analyser) analyseTrigrams() {
	var rl2SFB, rl2In, rl2Out, altSFS, altNml, rl3SFB, rl3In, rl3Out, redWeak, redSFS, redNml uint64

	for tri, cnt := range an.Corpus.Trigrams {
		var r0, r1, r2 KeyInfo
		var ok bool
		if r0, ok = an.Layout.RuneInfo[tri[0]]; !ok {
			continue
		}
		if r1, ok = an.Layout.RuneInfo[tri[1]]; !ok {
			continue
		}
		if r2, ok = an.Layout.RuneInfo[tri[2]]; !ok {
			continue
		}

		// pre-access
		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger

		add2Roll := func(fA, fB uint8) {
			switch {
			case fA == fB: // the same index is also a SFB??
				rl2SFB += cnt
			case (fA < fB) == (h1 == LEFT):
				rl2In += cnt
			default:
				rl2Out += cnt
			}
		}

		switch h0 {
		case h2:
			if h0 != h1 { // it's an Alternation
				if f0 == f2 && r0.Index != r2.Index {
					altSFS += cnt
				} else {
					altNml += cnt
				}
			} else { // it's a One-Hand pattern
				switch {
				case f0 == f1 || f1 == f2: // the same index is also a SFS??
					rl3SFB += cnt
				case (f0 < f1) == (f1 < f2):
					if (f0 < f1) == (h0 == LEFT) {
						rl3In += cnt
					} else {
						rl3Out += cnt
					}
				default:
					if f0 != LI && f0 != RI &&
						f1 != LI && f1 != RI &&
						f2 != LI && f2 != RI {
						redWeak += cnt
					} else if f0 == f2 && r0.Index != r2.Index {
						redSFS += cnt
					} else {
						redNml += cnt
					}
				}
			}
		case h1: // 2-roll with h0 == h1
			add2Roll(f0, f1)
		default: // 2-roll with h1 == h2
			add2Roll(f1, f2)
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

// AllMetricsDetails runs detailed analyses for multiple metrics, returning results for bigrams, skipgrams, and trigrams. Includes: SFB, LSB, FSB, HSB, SFS, LSS, FSS, HSS, ALT, 2RL, 3RL, and RED.
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

// SFBiDetails performs a detailed Same Finger Bigram (SFB) analysis.
// It scans all bigrams in the corpus and identifies those typed with the same finger (but not the same key).
// Unsupported bigrams (not present in the layout) are also tracked, and the Euclidean distance between keys is included.
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
		key1, ok1 := an.Layout.RuneInfo[bi[0]]
		key2, ok2 := an.Layout.RuneInfo[bi[1]]

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
			ma.Custom[biStr]["Δrow"] = kpDist.RowDist
		}
	}

	return ma
}

// LSBiDetails performs a detailed Lateral Stretch Bigram (LSB) analysis.
// Evaluates all layout-defined lateral stretch bigrams, returning their corpus frequency and column distance distribution.
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
		bi := Bigram{an.Layout.Runes[lsb.KeyIdx1], an.Layout.Runes[lsb.KeyIdx2]}
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
			ma.Custom[biStr]["Δcol"] = kpDist.ColDist
		}
	}

	return ma
}

// ScissBiDetails performs a detailed analysis of scissor bigrams, splitting into FSB (Full Scissor Bigrams, large vertical movement) and HSB (Half Scissor Bigrams, smaller vertical movement).
// Returns both analyses, with row distance included for each.
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
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
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
			ma.Custom[biStr]["Δrow"] = sci.rowDist
			ma.Custom[biStr]["Δcol"] = sci.colDist
			ma.Custom[biStr]["Angle"] = sci.angle
		}
	}

	for _, sci := range an.Layout.HScissors {
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
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
			ma2.Custom[biStr]["Δrow"] = sci.rowDist
			ma2.Custom[biStr]["Δcol"] = sci.colDist
			ma2.Custom[biStr]["Angle"] = sci.angle
		}
	}

	return ma, ma2
}

// SFSkpDetails performs a detailed Same Finger Skipgram (SFS) analysis.
// All skipgrams typed with the same finger (but not the same key) are included, with unsupported skipgrams and Euclidean distances tracked.
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
		key1, ok1 := an.Layout.RuneInfo[skp[0]]
		key2, ok2 := an.Layout.RuneInfo[skp[1]]

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
			ma.Custom[skpStr]["Δrow"] = kpDist.RowDist
		}
	}

	return ma
}

// LSSkpDetails performs a detailed Lateral Stretch Skipgram (LSS) analysis.
// Evaluates all layout-defined lateral stretch skipgrams, reporting their frequency and column distance.
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
		skp := Skipgram{an.Layout.Runes[lsb.KeyIdx1], an.Layout.Runes[lsb.KeyIdx2]}
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
			ma.Custom[skpStr]["Δcol"] = kpDist.ColDist
		}
	}

	return ma
}

// ScissSkpDetails performs a detailed analysis of scissor skipgrams, splitting into FSS (Full Scissor Skipgrams, large vertical movement) and HSS (Half Scissor Skipgrams, smaller vertical movement).
// Returns both analyses, including row distance for each skipgram.
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
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
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
			ma.Custom[skpStr]["Δrow"] = sci.rowDist
			ma.Custom[skpStr]["Δcol"] = sci.colDist
			ma.Custom[skpStr]["Angle"] = sci.angle
		}
	}

	for _, sci := range an.Layout.HScissors {
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
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
			ma2.Custom[skpStr]["Δrow"] = sci.rowDist
			ma2.Custom[skpStr]["Δcol"] = sci.colDist
			ma2.Custom[skpStr]["Angle"] = sci.angle
		}
	}

	return ma, ma2
}

// TrigramDetails performs a detailed analysis of trigram-based metrics:
//   - ALT: Alternations between hands (including ALT-SFS for same-finger alternations)
//   - 2RL: Two-key rolls (inward/outward) between adjacent fingers on one hand
//   - 3RL: Three-key rolls (inward/outward) on one hand
//   - RED: Redirections—direction changes on one hand, split into RED-NML (general), RED-SFS (same-finger skipgram), and RED-WEAK (without index involvement)
//
// Each returned MetricAnalysis includes frequency counts and can be used to compute derived totals and ratios.
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
		r0, ok0 := an.Layout.RuneInfo[tri[0]]
		r1, ok1 := an.Layout.RuneInfo[tri[1]]
		r2, ok2 := an.Layout.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			// alt.Unsupported[triStr] += cnt
			continue
		}

		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger

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

		switch h0 {
		case h2:
			if h0 != h1 { // it's an Alternation
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
			} else { // it's a One-Hand pattern
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
