// Package layout provides ergonomic analysis tools for keyboard layouts.
//
// Glossary of abbreviations used throughout this package:
//
//	SFB  - Same Finger Bigram
//	LSB  - Lateral Stretch Bigram
//	FSB  - Full Scissor Bigram
//	HSB  - Half Scissor Bigram
//	SFS  - Same Finger Skipgram
//	LSS  - Lateral Stretch Skipgram
//	FSS  - Full Scissor Skipgram
//	HSS  - Half Scissor Skipgram
//	ALT  - Alternation between hands (trigrams)
//	2RL  - Two-key rolls (inward/outward)
//	3RL  - Three-key rolls (inward/outward)
//	RED  - Redirections (direction changes on one hand)
//	FBL  - Finger Balance Load (deviation from ideal finger load)
//
// Metrics are calculated from unigrams, bigrams, skipgrams, and trigrams in the corpus.
package keycraft

import (
	"math"
	"strconv"
)

// idealFingerLoad specifies the target distribution of load (in percentages) across fingers for ergonomic balance.
var idealFingerLoad = map[string]float64{
	"F0": 8.0,
	"F1": 11.0,
	"F2": 16.0,
	"F3": 15.0,
	"F4": 0.0,
	"F5": 0.0,
	"F6": 15.0,
	"F7": 16.0,
	"F8": 11.0,
	"F9": 8.0,
}

// Analyser performs ergonomic analysis on a given layout using a corpus. It computes both quick percentage-based metrics and detailed breakdowns of relevant n-grams and distances.
type Analyser struct {
	// Reference to the analysed Layout.
	Layout *SplitLayout
	// Reference to the Corpus used to analyse the layout.
	Corpus *Corpus
	// Metrics holds all metrics about the layout.
	Metrics map[string]float64
}

// NewAnalyser constructs an Analyser for the given layout and corpus, runs a quick analysis, and initializes core metrics.
func NewAnalyser(layout *SplitLayout, corpus *Corpus) *Analyser {
	an := &Analyser{
		Layout:  layout,
		Corpus:  corpus,
		Metrics: make(map[string]float64),
	}
	an.quickHandAnalysis()
	an.quickMetricAnalysis()
	return an
}

// quickHandAnalysis computes hand, finger, column, and row usage metrics from unigrams. Same-finger balance (FBL) is also calculated as the cumulative deviation from the idealFingerLoad distribution.
func (an *Analyser) quickHandAnalysis() {
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

		totalUnigramCount += uniCnt
		if (key.Finger == LP || key.Finger == RP) && key.Row != 1 {
			pinkyOffHomeCount += uniCnt
		}
		handCount[key.Hand] += uniCnt
		fingerCount[key.Finger] += uniCnt
		columnCount[key.Column] += uniCnt
		rowCount[key.Row] += uniCnt
	}

	var factor float64
	if totalUnigramCount > 0 {
		factor = 100 / float64(totalUnigramCount)
	}
	an.Metrics["POH"] = float64(pinkyOffHomeCount) * factor
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

// quickMetricAnalysis computes a core set of ergonomic motion metrics, grouped as bigram, skipgram, and trigram features.
func (an *Analyser) quickMetricAnalysis() {
	an.analyzeBigrams()
	an.analyzeSkipgrams()
	an.analyzeTrigrams()
}

// analyzeBigrams computes bigram-based metrics: SFB, LSB, FSB, and HSB.
func (an *Analyser) analyzeBigrams() {
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
		bi := Bigram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if cnt, ok := an.Corpus.Bigrams[bi]; ok {
			count2 += cnt
		}
	}
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
	factor := 100 / float64(an.Corpus.TotalBigramsCount)
	an.Metrics["SFB"] = float64(count1) * factor
	an.Metrics["LSB"] = float64(count2) * factor
	an.Metrics["FSB"] = float64(count3) * factor
	an.Metrics["HSB"] = float64(count4) * factor
}

// analyzeSkipgrams computes skipgram-based metrics: SFS, LSS, FSS, and HSS.
func (an *Analyser) analyzeSkipgrams() {
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
		skp := Skipgram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if cnt, ok := an.Corpus.Skipgrams[skp]; ok {
			count2 += cnt
		}
	}
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
	factor := 100 / float64(an.Corpus.TotalSkipgramsCount)
	an.Metrics["SFS"] = float64(count1) * factor
	an.Metrics["LSS"] = float64(count2) * factor
	an.Metrics["FSS"] = float64(count3) * factor
	an.Metrics["HSS"] = float64(count4) * factor
}

// analyzeTrigrams computes trigram-based metrics: ALT (alternations), 2RL (two-key rolls), 3RL (three-key rolls), and RED (redirections).
// ALT captures alternation between hands (including ALT-SFS for same-finger alternations).
// 2RL and 3RL distinguish inward and outward rolling motions between adjacent fingers, with 2RL-SF and 3RL-SF for same-finger cases.
// RED includes redirections—direction changes on one hand—split into RED-OTH (general), RED-SFS (same-finger skipgram), and RED-WEAK (redirections without index involvement).
// Derived totals (ALT, 2RL, 3RL, RED, IN:OUT) are computed as sums or ratios of the above for overall ergonomic summary.
func (an *Analyser) analyzeTrigrams() {
	var rl2SF, rl2In, rl2Out, altSFS uint64
	var altOth, rl3SF, rl3In, rl3Out, redWeak, redSFS, redOth uint64

	for tri, cnt := range an.Corpus.Trigrams {
		r0, ok0 := an.Layout.RuneInfo[tri[0]]
		r1, ok1 := an.Layout.RuneInfo[tri[1]]
		r2, ok2 := an.Layout.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			continue
		}
		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger
		diffIdx02 := r0.Index != r2.Index

		add2Roll := func(fA, fB uint8) {
			switch {
			case fA == fB:
				rl2SF += cnt
			case (fA < fB) == (h1 == LEFT):
				rl2In += cnt
			default:
				rl2Out += cnt
			}
		}

		if h0 == h2 {
			if h0 != h1 {
				if f0 == f2 && diffIdx02 {
					altSFS += cnt
				} else {
					altOth += cnt
				}
			} else {
				switch {
				case f0 == f1 || f1 == f2:
					rl3SF += cnt
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
					} else if f0 == f2 && diffIdx02 {
						redSFS += cnt
					} else {
						redOth += cnt
					}
				}
			}
		} else if h0 == h1 {
			add2Roll(f0, f1)
		} else {
			add2Roll(f1, f2)
		}
	}
	factor := 100 / float64(an.Corpus.TotalTrigramsCount)
	an.Metrics["ALT-SFS"] = float64(altSFS) * factor
	an.Metrics["ALT-OTH"] = float64(altOth) * factor
	an.Metrics["ALT"] = an.Metrics["ALT-OTH"] + an.Metrics["ALT-SFS"]

	an.Metrics["2RL-SF"] = float64(rl2SF) * factor
	an.Metrics["2RL-IN"] = float64(rl2In) * factor
	an.Metrics["2RL-OUT"] = float64(rl2Out) * factor
	an.Metrics["2RL"] = an.Metrics["2RL-IN"] + an.Metrics["2RL-OUT"]

	an.Metrics["3RL-SF"] = float64(rl3SF) * factor
	an.Metrics["3RL-IN"] = float64(rl3In) * factor
	an.Metrics["3RL-OUT"] = float64(rl3Out) * factor
	an.Metrics["3RL"] = an.Metrics["3RL-IN"] + an.Metrics["3RL-OUT"]

	an.Metrics["RED-WEAK"] = float64(redWeak) * factor
	an.Metrics["RED-SFS"] = float64(redSFS) * factor
	an.Metrics["RED-OTH"] = float64(redOth) * factor
	an.Metrics["RED"] = an.Metrics["RED-OTH"] + an.Metrics["RED-SFS"] + an.Metrics["RED-WEAK"]

	an.Metrics["IN:OUT"] = (an.Metrics["2RL-IN"] + an.Metrics["3RL-IN"]) / (an.Metrics["2RL-OUT"] + an.Metrics["3RL-OUT"])
}

// MetricDetails contains detailed results for a single metric, including counts of relevant n-grams, unsupported n-grams, weighted distance values (row, column, or Euclidean), and totals.
type MetricDetails struct {
	Corpus       *Corpus
	CorpusNGramC uint64
	Metric       string
	Unsupported  map[string]uint64
	NGramCount   map[string]uint64
	NGramDist    map[string]float64 // todo: change to a map for any ngram info
	TotalNGrams  uint64
	TotalDist    float64 // todo: do we need it?
	Custom       map[string]map[string]any
}

// AllMetricsDetails runs detailed analyses for multiple metrics, returning results for bigrams, skipgrams, and trigrams. Includes: SFB, LSB, FSB, HSB, SFS, LSS, FSS, HSS, ALT, 2RL, 3RL, and RED.
func (an *Analyser) AllMetricsDetails() []*MetricDetails {
	all := make([]*MetricDetails, 0, 30)

	all = append(all, an.SFBDetails())
	all = append(all, an.LSBDetails())
	ma, ma2 := an.SBDetails()
	all = append(all, ma, ma2)
	all = append(all, an.SFSDetails())
	all = append(all, an.LSSDetails())
	ma3, ma4 := an.SSDetails()
	all = append(all, ma3, ma4)
	ta1, ta2, ta3, ta4 := an.TrigramDetails()
	all = append(all, ta1, ta2, ta3, ta4)

	return all
}

// SFBDetails performs a detailed Same Finger Bigram (SFB) analysis.
// It scans all bigrams in the corpus and identifies those typed with the same finger (but not the same key).
// Unsupported bigrams (not present in the layout) are also tracked, and the Euclidean distance between keys is included.
func (an *Analyser) SFBDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "SFB",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}

	for bi, biCnt := range an.Corpus.Bigrams {
		biStr := bi.String()
		key1, ok1 := an.Layout.RuneInfo[bi[0]]
		key2, ok2 := an.Layout.RuneInfo[bi[1]]

		if !ok1 || !ok2 {
			ma.Unsupported[biStr] += biCnt
			continue
		}

		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			ma.NGramCount[biStr] = biCnt
			ma.TotalNGrams += biCnt
			kp := KeyPair{key1.Index, key2.Index}
			dist := an.Layout.KeyPairDistances[kp].Distance
			ma.NGramDist[biStr] = dist
			ma.TotalDist += dist * float64(biCnt)
		}
	}

	return ma
}

// LSBDetails performs a detailed Lateral Stretch Bigram (LSB) analysis.
// Evaluates all layout-defined lateral stretch bigrams, returning their corpus frequency and column distance distribution.
func (an *Analyser) LSBDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "LSB",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}

	for _, lsb := range an.Layout.LSBs {
		bi := Bigram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if biCnt, ok := an.Corpus.Bigrams[bi]; ok {
			biStr := bi.String()

			ma.NGramCount[biStr] = biCnt
			ma.TotalNGrams += biCnt
			kp := KeyPair{uint8(lsb.keyIdx1), uint8(lsb.keyIdx2)}
			dist := an.Layout.KeyPairDistances[kp].ColDist
			ma.NGramDist[biStr] = dist
			ma.TotalDist += dist * float64(biCnt)
		}
	}

	return ma
}

// SBDetails performs a detailed analysis of scissor bigrams, splitting into FSB (Full Scissor Bigrams, large vertical movement) and HSB (Half Scissor Bigrams, smaller vertical movement).
// Returns both analyses, with row distance included for each.
func (an *Analyser) SBDetails() (*MetricDetails, *MetricDetails) {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "FSB",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}
	ma2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalBigramsCount,
		Metric:       "HSB",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}
	for _, sci := range an.Layout.Scissors {
		bi := Bigram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if biCnt, ok := an.Corpus.Bigrams[bi]; ok {
			biStr := bi.String()
			kp := KeyPair{uint8(sci.keyIdx1), uint8(sci.keyIdx2)}
			dist := an.Layout.KeyPairDistances[kp].RowDist
			if dist > 1.5 {
				ma.NGramCount[biStr] = biCnt
				ma.TotalNGrams += biCnt
				ma.NGramDist[biStr] = dist
				ma.TotalDist += dist * float64(biCnt)
			} else {
				ma2.NGramCount[biStr] = biCnt
				ma2.TotalNGrams += biCnt
				ma2.NGramDist[biStr] = dist
				ma2.TotalDist += dist * float64(biCnt)
			}
		}
	}
	return ma, ma2
}

// SFSDetails performs a detailed Same Finger Skipgram (SFS) analysis.
// All skipgrams typed with the same finger (but not the same key) are included, with unsupported skipgrams and Euclidean distances tracked.
func (an *Analyser) SFSDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "SFS",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}

	for skp, skpCnt := range an.Corpus.Skipgrams {
		skpStr := skp.String()
		key1, ok1 := an.Layout.RuneInfo[skp[0]]
		key2, ok2 := an.Layout.RuneInfo[skp[1]]

		if !ok1 || !ok2 {
			ma.Unsupported[skpStr] += skpCnt
			continue
		}

		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			ma.NGramCount[skpStr] = skpCnt
			ma.TotalNGrams += skpCnt
			kp := KeyPair{key1.Index, key2.Index}
			dist := an.Layout.KeyPairDistances[kp].Distance
			ma.NGramDist[skpStr] = dist
			ma.TotalDist += dist * float64(skpCnt)
		}
	}

	return ma
}

// LSSDetails performs a detailed Lateral Stretch Skipgram (LSS) analysis.
// Evaluates all layout-defined lateral stretch skipgrams, reporting their frequency and column distance.
func (an *Analyser) LSSDetails() *MetricDetails {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "LSS",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}

	for _, lsb := range an.Layout.LSBs {
		skp := Skipgram{an.Layout.Runes[lsb.keyIdx1], an.Layout.Runes[lsb.keyIdx2]}
		if skpCnt, ok := an.Corpus.Skipgrams[skp]; ok {
			skpStr := skp.String()

			ma.NGramCount[skpStr] = skpCnt
			ma.TotalNGrams += skpCnt
			kp := KeyPair{uint8(lsb.keyIdx1), uint8(lsb.keyIdx2)}
			dist := an.Layout.KeyPairDistances[kp].ColDist
			ma.NGramDist[skpStr] = dist
			ma.TotalDist += dist * float64(skpCnt)
		}
	}

	return ma
}

// SSDetails performs a detailed analysis of scissor skipgrams, splitting into FSS (Full Scissor Skipgrams, large vertical movement) and HSS (Half Scissor Skipgrams, smaller vertical movement).
// Returns both analyses, including row distance for each skipgram.
func (an *Analyser) SSDetails() (*MetricDetails, *MetricDetails) {
	ma := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "FSS",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}
	ma2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalSkipgramsCount,
		Metric:       "HSS",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
	}

	for _, sci := range an.Layout.Scissors {
		skp := Skipgram{an.Layout.Runes[sci.keyIdx1], an.Layout.Runes[sci.keyIdx2]}
		if skpCnt, ok := an.Corpus.Skipgrams[skp]; ok {
			skpStr := skp.String()

			kp := KeyPair{uint8(sci.keyIdx1), uint8(sci.keyIdx2)}
			dist := an.Layout.KeyPairDistances[kp].RowDist
			if dist > 1.5 {
				ma.NGramCount[skpStr] = skpCnt
				ma.TotalNGrams += skpCnt
				ma.NGramDist[skpStr] = dist
				ma.TotalDist += dist * float64(skpCnt)
			} else {
				ma2.NGramCount[skpStr] = skpCnt
				ma2.TotalNGrams += skpCnt
				ma2.NGramDist[skpStr] = dist
				ma2.TotalDist += dist * float64(skpCnt)
			}
		}
	}

	return ma, ma2
}

// TrigramDetails performs a detailed analysis of trigram-based metrics:
//   - ALT: Alternations between hands (including ALT-SFS for same-finger alternations)
//   - 2RL: Two-key rolls (inward/outward) between adjacent fingers on one hand
//   - 3RL: Three-key rolls (inward/outward) on one hand
//   - RED: Redirections—direction changes on one hand, split into RED-OTH (general), RED-SFS (same-finger skipgram), and RED-WEAK (without index involvement)
//
// Each returned MetricAnalysis includes frequency counts and can be used to compute derived totals and ratios.
func (an *Analyser) TrigramDetails() (*MetricDetails, *MetricDetails, *MetricDetails, *MetricDetails) {
	alt := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "ALT",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
		Custom:       make(map[string]map[string]any),
	}
	rl2 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "2RL",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
		Custom:       make(map[string]map[string]any),
	}
	rl3 := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "3RL",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
		Custom:       make(map[string]map[string]any),
	}
	red := &MetricDetails{
		Corpus:       an.Corpus,
		CorpusNGramC: an.Corpus.TotalTrigramsCount,
		Metric:       "RED",
		Unsupported:  make(map[string]uint64),
		NGramCount:   make(map[string]uint64),
		NGramDist:    make(map[string]float64),
		Custom:       make(map[string]map[string]any),
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
		diffIdx02 := r0.Index != r2.Index

		add2Roll := func(fA, fB uint8) {
			rl2.NGramCount[triStr] = cnt
			rl2.TotalNGrams += cnt
			switch {
			case fA == fB:
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Kind"] = "SF"
			case (fA < fB) == (h1 == LEFT):
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Kind"] = "IN"
			default:
				if _, ok := rl2.Custom[triStr]; !ok {
					rl2.Custom[triStr] = make(map[string]any)
				}
				rl2.Custom[triStr]["Kind"] = "OUT"
			}
		}

		if h0 == h2 {
			if h0 != h1 {
				alt.NGramCount[triStr] = cnt
				alt.TotalNGrams += cnt
				if _, ok := alt.Custom[triStr]; !ok {
					alt.Custom[triStr] = make(map[string]any)
				}
				if f0 == f2 && diffIdx02 {
					alt.Custom[triStr]["Kind"] = "SFS"
				} else {
					alt.Custom[triStr]["Kind"] = "OTH"
				}
			} else {
				if f0 == f1 || f1 == f2 {
					rl3.NGramCount[triStr] = cnt
					rl3.TotalNGrams += cnt
					if _, ok := rl3.Custom[triStr]; !ok {
						rl3.Custom[triStr] = make(map[string]any)
					}
					rl3.Custom[triStr]["Kind"] = "SF"
				} else if (f0 < f1) == (f1 < f2) {
					rl3.NGramCount[triStr] = cnt
					rl3.TotalNGrams += cnt
					if _, ok := rl3.Custom[triStr]; !ok {
						rl3.Custom[triStr] = make(map[string]any)
					}
					if (f0 < f1) == (h0 == LEFT) {
						rl3.Custom[triStr]["Kind"] = "IN"
					} else {
						rl3.Custom[triStr]["Kind"] = "OUT"
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
						red.Custom[triStr]["Kind"] = "WEAK"
					} else if f0 == f2 && diffIdx02 {
						red.Custom[triStr]["Kind"] = "SFS"
					} else {
						red.Custom[triStr]["Kind"] = "OTH"
					}
				}
			}
		} else if h0 == h1 {
			add2Roll(f0, f1)
		} else { // h1 == h2
			add2Roll(f1, f2)
		}
	}

	return alt, rl2, rl3, red
}
