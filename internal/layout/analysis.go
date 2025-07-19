// Package layout provides functionality for analyzing keyboard layouts.
package layout

import (
	"fmt"
	"sort"
	"strings"
)

// Usage represents the usage statistics for a hand, row, column, or finger.
type Usage struct {
	// Count is the total count of key presses.
	Count uint64
	// Percentage is the percentage of key presses.
	Percentage float64
}

// HandAnalysis represents the hand usage analysis for a layout.
type HandAnalysis struct {
	// LayoutName is the name of the layout.
	LayoutName string
	// CorpusName is the name of the corpus.
	CorpusName string
	// TotalUnigramCount is the total count of unigrams in the corpus.
	TotalUnigramCount uint64
	// HandUsage is the usage statistics for each hand.
	HandUsage [2]Usage
	// RowUsage is the usage statistics for each row.
	RowUsage [4]Usage
	// ColumnUsage is the usage statistics for each column.
	ColumnUsage [12]Usage
	// FingerUsage is the usage statistics for each finger.
	FingerUsage [10]Usage
	// RunesUnavailable is a map of runes that are not available in the layout.
	RunesUnavailable map[rune]uint64
}

// String returns a string representation of the hand analysis.
func (ha HandAnalysis) String() string {
	return fmt.Sprintf("%s (%s):\nHands: %s\nRows: %s\nColumns: %s\nFingers: %s\nUnavailable runes: %s",
		ha.LayoutName,
		ha.CorpusName,
		formatUsage(ha.HandUsage[:]),
		formatUsage(ha.RowUsage[:]),
		formatUsage(ha.ColumnUsage[:]),
		formatUsage(ha.FingerUsage[:]),
		formatUnavailable(ha.RunesUnavailable),
	)
}

// formatUsage formats the usage statistics for a slice of Usage.
func formatUsage(usage []Usage) string {
	var parts []string
	for _, u := range usage {
		parts = append(parts, fmt.Sprintf("%.1f%%", u.Percentage))
	}
	return strings.Join(parts, ", ")
}

// formatUnavailable formats the unavailable runes map.
func formatUnavailable(na map[rune]uint64) string {
	var pairs []CountPair[rune]
	for r, c := range na {
		pairs = append(pairs, CountPair[rune]{Key: r, Count: c})
	}

	// Sort pairs based on the count in descending orders
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	var parts []string
	for _, pair := range pairs {
		parts = append(parts, fmt.Sprintf("%c (%s)", pair.Key, Comma(pair.Count)))
	}

	return strings.Join(parts, ", ")
}

// AnalyzeHandUsage analyzes the hand usage for a layout and corpus.
func (sl *SplitLayout) AnalyzeHandUsage(corp *Corpus) HandAnalysis {
	var handUsage [2]Usage
	var rowUsage [4]Usage
	var columnUsage [12]Usage
	var fingerUsage [10]Usage
	var totalUnigramCount uint64
	var runesUnavailable = make(map[rune]uint64, 0)

	// Iterate over unigrams in the corpus and calculate usage statistics.
	for r, count := range corp.Unigrams {
		info, ok := sl.RuneInfo[rune(r)]
		if !ok {
			runesUnavailable[rune(r)] += count
			continue
		}

		totalUnigramCount += count
		hand := IfThen(info.Hand == "left", 0, 1)
		handUsage[hand].Count += count
		rowUsage[info.Row].Count += count
		columnUsage[info.Column].Count += count
		fingerUsage[info.Finger].Count += count
	}

	// Calculate the percentages.
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
		RunesUnavailable:  runesUnavailable,
	}
}

// Sfb represents a same-finger bigrams (SFB) with its count and percentage.
type Sfb struct {
	// Bigram is the SFB bigram.
	Bigram Bigram
	// The distance from one key to the next
	Distance float32
	// Count is the count of the SFB.
	Count uint64
	// Percentage is the percentage of the SFB.
	Percentage float32
}

// SfbAnalysis represents the SFB analysis for a layout and corpus.
type SfbAnalysis struct {
	// Reference to the analysed layout.
	SplitLayout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	Corpus *Corpus
	// Sfbs is a slice of SFBs.
	Sfbs []Sfb
	// TotalSfbCount is the total count of SFBs.
	TotalSfbCount uint64
	// TotalSfbPerc is the total percentage of SFBs.
	TotalSfbPerc float32
	// Bigrams not supported by the layout due to missing characters.
	Unsupported []BigramCount
}

// SimpleSfbs returns the SFB percentage for a layout and corpus.
func (sl *SplitLayout) SimpleSfbs(corp *Corpus) float64 {
	var totalCount uint64

	for bi, cnt := range corp.Bigrams {
		if bi[0] == bi[1] {
			continue
		}
		info0, ok0 := sl.RuneInfo[bi[0]]
		info1, ok1 := sl.RuneInfo[bi[1]]
		if ok0 && ok1 && info0.Finger == info1.Finger {
			totalCount += cnt
		}
	}

	return float64(totalCount) / float64(corp.TotalBigramsNoSpace)
}

// AnalyzeSfbs analyzes the SFBs for a layout and corpus.
func (sl *SplitLayout) AnalyzeSfbs(corpus *Corpus) SfbAnalysis {
	an := SfbAnalysis{
		SplitLayout: sl,
		Corpus:      corpus,
		Sfbs:        make([]Sfb, 0),
		Unsupported: make([]BigramCount, 0),
	}

	for bi, biCount := range corpus.Bigrams {
		if bi[0] == bi[1] {
			// ignore 0U Bigrams
			continue
		}

		rune0, ok0 := sl.RuneInfo[bi[0]]
		rune1, ok1 := sl.RuneInfo[bi[1]]
		if !ok0 || !ok1 {
			// detected a bigram that has a rune not on the layout
			an.Unsupported = append(an.Unsupported, BigramCount{bi, biCount})
		} else if rune0.Finger == rune1.Finger {
			biPerc := float32(biCount) / float32(corpus.TotalBigramsNoSpace)
			sfb := Sfb{
				Bigram:     bi,
				Distance:   sl.distances.GetDistance(rune0, rune1),
				Count:      biCount,
				Percentage: biPerc,
			}
			an.Sfbs = append(an.Sfbs, sfb)
			an.TotalSfbCount += biCount
			an.TotalSfbPerc += biPerc
		}
	}

	// sort SFSs by the number of times they occur in the corpus
	sort.Slice(an.Sfbs, func(i, j int) bool {
		return an.Sfbs[i].Count > an.Sfbs[j].Count
	})

	return an
}

// Sfs represents a same-finger skipgrams (SFS) with its count and percentage.
type Sfs struct {
	// Trigram is the SFS trigram.
	Trigram Trigram
	// The distance from one key to the next
	Distance float32
	// The number of times the SFS occurs in a corpus.
	Count uint64
	// Percentage of trigrams in a corpus with this SFS.
	Percentage float32
}

// SfsAnalysis represents the SFS analysis for a layout and corpus.
type SfsAnalysis struct {
	// Reference to the analysed layout.
	SplitLayout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	Corpus *Corpus
	// SFSs occurring in the corpus.
	Sfss []Sfs
	// TotalSfbCount is the total count of SFSs.
	TotalSfsCount uint64
	// TotalSfbPerc is the total percentage of SFSs.
	TotalSfsPerc float32
	// SFSs with the same first and last character merged.
	MergedSfss []Sfs
	// Trigrams not supported by the layout due to missing characters.
	Unsupported []TrigramCount
}

// SimpleSfss analyzes the SFSs for this layout and some corpus.
func (sl *SplitLayout) SimpleSfss(corp *Corpus) float64 {
	var TotalSfsCount uint64

	for tri, cnt := range corp.Trigrams {
		if tri[0] == tri[2] {
			// ignore 0U SFS
			continue
		}

		rune0, ok0 := sl.RuneInfo[tri[0]]
		rune1, ok1 := sl.RuneInfo[tri[1]]
		rune2, ok2 := sl.RuneInfo[tri[2]]
		if ok0 && ok1 && ok2 && rune0.Finger == rune2.Finger && rune0.Finger != rune1.Finger {
			TotalSfsCount += cnt
		}
	}

	// calculate the percentage in a single div op
	return float64(TotalSfsCount) / float64(corp.TotalTrigramsCount)
}

// AnalyzeSfss analyzes the SFSs for this layout and some corpus.
func (sl *SplitLayout) AnalyzeSfss(corpus *Corpus) SfsAnalysis {
	an := SfsAnalysis{
		SplitLayout: sl,
		Corpus:      corpus,
		Sfss:        make([]Sfs, 0),
		MergedSfss:  make([]Sfs, 0),
		Unsupported: make([]TrigramCount, 0),
	}

	merged := make(map[Trigram]Sfs)

	for tri, triCount := range corpus.Trigrams {
		if tri[0] == tri[2] {
			// ignore 0U SFS
			continue
		}

		rune0, ok0 := sl.RuneInfo[tri[0]]
		rune1, ok1 := sl.RuneInfo[tri[1]]
		rune2, ok2 := sl.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			an.Unsupported = append(an.Unsupported, TrigramCount{tri, triCount})
		} else if rune0.Finger == rune2.Finger && rune0.Finger != rune1.Finger {
			triDist := sl.distances.GetDistance(rune0, rune2)
			triPerc := float32(triCount) / float32(corpus.TotalTrigramsCount)
			sfs := Sfs{
				Trigram:    tri,
				Distance:   triDist,
				Count:      triCount,
				Percentage: triPerc}
			an.Sfss = append(an.Sfss, sfs)
			an.TotalSfsCount += triCount
			an.TotalSfsPerc += triPerc

			// keep track of skipgrams with no middle character
			var mergedTrigram Trigram
			if tri[0] < tri[2] {
				mergedTrigram = Trigram{tri[0], '_', tri[2]}
			} else {
				mergedTrigram = Trigram{tri[2], '_', tri[0]}
			}
			if existingSfs, ok := merged[mergedTrigram]; ok {
				existingSfs.Count += triCount
				existingSfs.Percentage += triPerc
				merged[mergedTrigram] = existingSfs
			} else {
				merged[mergedTrigram] = Sfs{
					Trigram:    mergedTrigram,
					Distance:   triDist,
					Count:      triCount,
					Percentage: triPerc,
				}
			}
		}
	}

	// sort SFSs by the number of times they occur in the corpus
	sort.Slice(an.Sfss, func(i, j int) bool {
		return an.Sfss[i].Count > an.Sfss[j].Count
	})

	// flatten and sort merged SFSes
	an.MergedSfss = make([]Sfs, 0, len(merged))
	for _, sfs := range merged {
		an.MergedSfss = append(an.MergedSfss, sfs)
	}
	sort.Slice(an.MergedSfss, func(i, j int) bool {
		return an.MergedSfss[i].Count > an.MergedSfss[j].Count
	})

	return an
}

// Lsb represents a Lateral Stretch Bigram (LSB) with its count and percentage.
type Lsb struct {
	// Bigram is the LSB bigram.
	Bigram Bigram
	// The distance from one key to the next
	Distance float32
	// Count is the number of occurrences of the LSB in a corpus.
	Count uint64
	// Percentage is the percentage of the corpus with this LSB.
	Percentage float32
}

type LsbAnalysis struct {
	// Reference to the analysed layout.
	SplitLayout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	Corpus *Corpus
	// LSBs occurring in the corpus.
	Lsbs []Lsb
	// TotalLsbCount is the total number of LSB occurences in the corpus.
	TotalLsbCount uint64
	// TotalLsbPerc is the percentage of bigrams in the corpus that are LSB.
	TotalLsbPerc float32
}

func (sl *SplitLayout) AnalyzeLsbs(corpus *Corpus) *LsbAnalysis {
	an := &LsbAnalysis{
		SplitLayout: sl,
		Corpus:      corpus,
		Lsbs:        make([]Lsb, 0),
	}

	sl.analyzeLsbs(corpus, an, 2.0, keyPairs20)
	sl.analyzeLsbs(corpus, an, 3.5, keyPairs35)

	return an
}

func (sl *SplitLayout) analyzeLsbs(corpus *Corpus, an *LsbAnalysis, minDistance float32, keyPairs [][2]int) {
	for _, pair := range keyPairs {
		r0, r1 := sl.Runes[pair[0]], sl.Runes[pair[1]]
		if r0 == 0 || r1 == 0 {
			// position on layout has no character
			continue
		}

		// look up the bigram in the corpus
		lsbBi := Bigram{r0, r1}
		biCount, ok := corpus.Bigrams[lsbBi]
		if !ok {
			// Corpus doesn't have this LSB
			continue
		}

		// Look up rune details to get distance
		biDist, err := sl.GetDistance(r0, r1)
		if err != nil {
			panic(fmt.Errorf("internal error finding runes %c or %c: ", r0, r1))
		}
		if biDist < minDistance {
			// ignore bigrams with a short distance
			continue
		}
		biPerc := float32(biCount) / float32(corpus.TotalBigramsCount)

		// Add new LSB, update totals
		lsb := Lsb{
			Bigram:     lsbBi,
			Distance:   biDist,
			Count:      biCount,
			Percentage: biPerc,
		}
		an.Lsbs = append(an.Lsbs, lsb)
		an.TotalLsbCount += biCount
		an.TotalLsbPerc += biPerc
	}
}

// LSBs with >=2U
var keyPairs20 = [][2]int{
	{0, 2},
	{0, 14},
	{0, 26},
	{12, 2},
	{12, 14},
	{12, 26},
	{24, 2},
	{24, 14},
	{24, 26},
	{3, 5},
	{3, 17},
	{3, 29},
	{15, 5},
	{15, 17},
	{15, 29},
	{27, 5},
	{27, 17},
	{27, 29},
	{6, 8},
	{6, 20},
	{6, 32},
	{18, 8},
	{18, 20},
	{18, 32},
	{30, 8},
	{30, 20},
	{30, 32},
	{9, 11},
	{9, 23},
	{9, 35},
	{21, 11},
	{21, 23},
	{21, 35},
	{33, 11},
	{33, 23},
	{33, 35},
}

// LSBs with >=3.5U
var keyPairs35 = [][2]int{
	{2, 16},
	{2, 17},
	{2, 28},
	{2, 29},
	{14, 28},
	{14, 29},
	{26, 4},
	{26, 5},
	{26, 16},
	{26, 17},
	{9, 18},
	{9, 19},
	{9, 30},
	{9, 31},
	{21, 6},
	{21, 7},
	{33, 6},
	{33, 7},
	{33, 18},
	{33, 19},
}
