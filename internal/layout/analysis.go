// Package layout provides functionality for analyzing keyboard layouts.
package layout

import (
	"fmt"
	"sort"
)

// Usage represents the usage statistics for a hand, row, column, or finger.
type Usage struct {
	// Count is the total count of key presses.
	Count uint64
	// Percentage is the percentage of key presses.
	Percentage float64
}

// Sfb represents a same-finger bigrams (SFB) with its count and percentage.
type Sfb struct {
	// Bigram is the SFB bigram.
	Bigram Bigram
	// The distance from one key to the next
	Distance float64
	// Count is the count of the SFB.
	Count uint64
	// Percentage is the percentage of the SFB.
	Percentage float64
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
	TotalSfbPerc float64
	// Bigrams not supported by the layout due to missing characters.
	Unsupported []BigramCount
	//
	NumRowsInOutput int
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

	return float64(totalCount) / float64(corp.TotalBigramsCount)
}

// AnalyzeSfbs analyzes the SFBs for a layout and corpus.
func (sl *SplitLayout) AnalyzeSfbs(corpus *Corpus) SfbAnalysis {
	an := SfbAnalysis{
		SplitLayout:     sl,
		Corpus:          corpus,
		Sfbs:            make([]Sfb, 0),
		Unsupported:     make([]BigramCount, 0),
		NumRowsInOutput: 10,
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
			dist := sl.KeyPairDistances[KeyPair{rune0.Index, rune1.Index}].Distance
			biPerc := float64(biCount) / float64(corpus.TotalBigramsCount)
			sfb := Sfb{
				Bigram:     bi,
				Distance:   dist,
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
	Distance float64
	// The number of times the SFS occurs in a corpus.
	Count uint64
	// Percentage of trigrams in a corpus with this SFS.
	Percentage float64
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
	TotalSfsPerc float64
	// SFSs with the same first and last character merged.
	MergedSfss []Sfs
	// Trigrams not supported by the layout due to missing characters.
	Unsupported []TrigramCount
	//
	NumRowsInOutput int
	//
	MinDistanceInOutput float64
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
		SplitLayout:         sl,
		Corpus:              corpus,
		Sfss:                make([]Sfs, 0),
		MergedSfss:          make([]Sfs, 0),
		Unsupported:         make([]TrigramCount, 0),
		NumRowsInOutput:     10,
		MinDistanceInOutput: 1,
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
			triDist := sl.KeyPairDistances[KeyPair{rune0.Index, rune2.Index}].Distance
			triPerc := float64(triCount) / float64(corpus.TotalTrigramsCount)
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
	// The horizontal distance from one key to the next
	HorDistance float64
	// Count is the number of occurrences of the LSB in a corpus.
	Count uint64
	// Percentage is the percentage of the corpus with this LSB.
	Percentage float64
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
	TotalLsbPerc float64
	//
	NumRowsInOutput int
}

func (sl *SplitLayout) SimpleLsbs(corpus *Corpus) float64 {
	var totalLsbCount uint64

	for _, pair := range sl.LSBs {
		r0, r1 := sl.Runes[pair.keyIdx1], sl.Runes[pair.keyIdx2]
		if r0 == 0 || r1 == 0 {
			// position on layout has no character
			panic(fmt.Errorf("%c or %c not found on layout, which should be impossible", r0, r1))
		}

		// look up the bigram in the corpus
		lsbBi := Bigram{r0, r1}
		biCount, ok := corpus.Bigrams[lsbBi]
		if !ok {
			// Corpus doesn't have this LSB, yay!
			continue
		}

		totalLsbCount += biCount
	}

	// This doesn't take into account the distance of each LSB!

	return float64(totalLsbCount) / float64(corpus.TotalBigramsCount)
}

func (sl *SplitLayout) AnalyzeLsbs(corpus *Corpus) *LsbAnalysis {
	an := &LsbAnalysis{
		SplitLayout:     sl,
		Corpus:          corpus,
		Lsbs:            make([]Lsb, 0),
		NumRowsInOutput: 10,
	}

	for _, pair := range sl.LSBs {
		r0, r1 := sl.Runes[pair.keyIdx1], sl.Runes[pair.keyIdx2]
		if r0 == 0 || r1 == 0 {
			// position on layout has no character
			panic(fmt.Errorf("%c or %c not found on layout, which should be impossible", r0, r1))
		}

		// look up the bigram in the corpus
		lsbBi := Bigram{r0, r1}
		biCount, ok := corpus.Bigrams[lsbBi]
		if !ok {
			// Corpus doesn't have this LSB, yay!
			continue
		}

		biPerc := float64(biCount) / float64(corpus.TotalBigramsCount)

		// Add new LSB, update totals
		lsb := Lsb{
			Bigram:      lsbBi,
			HorDistance: pair.colDistance,
			Count:       biCount,
			Percentage:  biPerc,
		}
		an.Lsbs = append(an.Lsbs, lsb)
		an.TotalLsbCount += biCount
		an.TotalLsbPerc += biPerc
	}

	return an
}

// Scissor represents a Full or Half Scissor Bigram
type Scissor struct {
	// Bigram is the Scissor bigram.
	Bigram Bigram
	//
	FingerDistance uint8
	//
	RowDistance float64
	//
	Angle float64
	// Count is the number of occurrences of the Scissor in a corpus.
	Count uint64
	// Percentage is the percentage of the corpus with this Scissor.
	Percentage float64
}

type ScissorAnalysis struct {
	// Reference to the analysed layout.
	SplitLayout *SplitLayout
	// Reference to the corpus used to analyse the layout.
	Corpus *Corpus
	// LSBs occurring in the corpus.
	Scissors []Scissor
	// TotalLsbCount is the total number of LSB occurences in the corpus.
	TotalScissorCount uint64
	// TotalLsbPerc is the percentage of bigrams in the corpus that are LSB.
	TotalScissorPerc float64
	//
	NumRowsInOutput int
}

func (sl *SplitLayout) SimpleScissors(corpus *Corpus) float64 {
	var totalScissorCount uint64

	for _, pair := range sl.Scirrors {
		r0, r1 := sl.Runes[pair.keyIdx1], sl.Runes[pair.keyIdx2]
		if r0 == 0 || r1 == 0 {
			// position on layout has no character
			panic(fmt.Errorf("%c or %c not found on layout, which should be impossible", r0, r1))
		}

		// look up the bigram in the corpus
		lsbBi := Bigram{r0, r1}
		biCount, ok := corpus.Bigrams[lsbBi]
		if !ok {
			// Corpus doesn't have this Scissor, yay!
			continue
		}

		totalScissorCount += biCount
	}

	return float64(totalScissorCount) / float64(corpus.TotalBigramsCount)
}

func (sl *SplitLayout) AnalyzeScissors(corpus *Corpus) *ScissorAnalysis {
	an := &ScissorAnalysis{
		SplitLayout:     sl,
		Corpus:          corpus,
		Scissors:        make([]Scissor, 0),
		NumRowsInOutput: 10,
	}

	for _, pair := range sl.Scirrors {
		r0, r1 := sl.Runes[pair.keyIdx1], sl.Runes[pair.keyIdx2]
		if r0 == 0 || r1 == 0 {
			// position on layout has no character
			panic(fmt.Errorf("%c or %c not found on layout, which should be impossible", r0, r1))
		}

		// look up the bigram in the corpus
		sciBi := Bigram{r0, r1}
		biCount, ok := corpus.Bigrams[sciBi]
		if !ok {
			// Corpus doesn't have this FSB, yay!
			continue
		}

		biPerc := float64(biCount) / float64(corpus.TotalBigramsCount)

		// Add new FSB, update totals
		scissor := Scissor{
			Bigram:         sciBi,
			FingerDistance: pair.fingerDist,
			RowDistance:    pair.rowDist,
			Angle:          pair.angle,
			Count:          biCount,
			Percentage:     biPerc,
		}
		an.Scissors = append(an.Scissors, scissor)
		an.TotalScissorCount += biCount
		an.TotalScissorPerc += biPerc
	}

	return an
}
