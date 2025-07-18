// Package layout provides functionality for analyzing keyboard layouts.
package layout

import (
	"fmt"
	"math"
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
	Percentage float64
}

// SfbAnalysis represents the SFB analysis for a layout and corpus.
type SfbAnalysis struct {
	// LayoutName is the name of the layout.
	LayoutName string
	// CorpusName is the name of the corpus.
	CorpusName string
	// TotalBigrams is the total count of bigrams in the corpus.
	TotalBigrams uint64
	// Bigrams not supported by the layout due to missing characters.
	Unsupported []BigramCount
	// Sfbs is a slice of SFBs.
	Sfbs []Sfb
	// TotalSfbCount is the total count of SFBs.
	TotalSfbCount uint64
	// TotalSfbPerc is the total percentage of SFBs.
	TotalSfbPerc float64
}

// String returns a string representation of the SFB analysis.
func (sa SfbAnalysis) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Corpus: %s (%s bigrams)\n", sa.CorpusName, Comma(sa.TotalBigrams)))
	sb.WriteString(fmt.Sprintf("Total SFBs: %s (%.2f%% of corpus) in %v\n",
		Comma(sa.TotalSfbCount), 100*sa.TotalSfbPerc, sa.LayoutName))
	printCount := min(10, len(sa.Sfbs))
	sb.WriteString(fmt.Sprintf("Top-%d SFBs:\n", printCount))
	for i := range printCount {
		sfb := sa.Sfbs[i]
		sb.WriteString(fmt.Sprintf("%2d. %v (%s, %.3f%%)\n", i+1, sfb.Bigram, Comma(sfb.Count), 100*sfb.Percentage))
	}

	return sb.String()
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
func (sl *SplitLayout) AnalyzeSfbs(corp *Corpus) SfbAnalysis {
	an := SfbAnalysis{
		LayoutName:   sl.Filename,
		CorpusName:   corp.Name,
		TotalBigrams: corp.TotalBigramsNoSpace,
	}

	for bi, cnt := range corp.Bigrams {
		if bi[0] == bi[1] {
			// ignore 0U Bigrams
			continue
		}

		info0, ok0 := sl.RuneInfo[bi[0]]
		info1, ok1 := sl.RuneInfo[bi[1]]
		if !ok0 || !ok1 {
			// detected a bigram that has a rune not on the layout
			an.Unsupported = append(an.Unsupported, BigramCount{bi, cnt})
		} else if info0.Finger == info1.Finger {
			perc := float64(cnt) / float64(corp.TotalBigramsCount)
			an.Sfbs = append(an.Sfbs, Sfb{Bigram: bi, Count: cnt, Percentage: perc})
			an.TotalSfbCount += cnt
		}
	}

	// calculate the percentage in a single div op
	an.TotalSfbPerc = float64(an.TotalSfbCount) / float64(an.TotalBigrams)

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
	Percentage float64
}

// SfsAnalysis represents the SFS analysis for a layout and corpus.
type SfsAnalysis struct {
	// LayoutName is the name of the layout.
	LayoutName string
	// CorpusName is the name of the corpus.
	CorpusName string
	// TotalTrigrams is the total number of trigrams in the corpus.
	TotalTrigrams uint64
	// Trigrams not supported by the layout due to missing characters.
	Unsupported []TrigramCount
	// SFSs occurring in the corpus.
	Sfss []Sfs
	// TotalSfbCount is the total count of SFSs.
	TotalSfsCount uint64
	// TotalSfbPerc is the total percentage of SFSs.
	TotalSfsPerc float64
	// SFSs with the same first and last character merged.
	MergedSfss []Sfs
}

// String returns a string representation of the SFB analysis.
func (sa SfsAnalysis) String() string {
	return sa.SFSString(0, 0, 10, 0)
}

// SFSString returns a string representation of an SfsAnalysis
func (sa SfsAnalysis) SFSString(sfsGt1UCount, sfs1UCount, mergedSfsGt1UCount, mergedSfs1UCount int) string {
	var sb strings.Builder
	var printed int

	sb.WriteString(fmt.Sprintf("Corpus: %s (%s trigrams)\n", sa.CorpusName, Comma(sa.TotalTrigrams)))
	sb.WriteString(fmt.Sprintf("Total SFSs: %s (%.2f%% of corpus) in %v\n",
		Comma(sa.TotalSfsCount), 100*sa.TotalSfsPerc, sa.LayoutName))

	if sfsGt1UCount > 0 {
		printed = 0
		sb.WriteString(fmt.Sprintf("Top-%d >1U SFSes:\n", sfsGt1UCount))
		for _, sfs := range sa.Sfss {
			if sfs.Distance > 1 {
				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
				printed++
				if printed >= sfsGt1UCount {
					break
				}
			}
		}
	}

	if sfs1UCount > 0 {
		sb.WriteString(fmt.Sprintf("Top-%d 1U SFSes:\n", sfs1UCount))
		printed = 0
		for _, sfs := range sa.Sfss {
			if sfs.Distance == 1 {
				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
				printed++
				if printed >= sfs1UCount {
					break
				}
			}
		}
	}

	if mergedSfsGt1UCount > 0 {
		printed = 0
		sb.WriteString(fmt.Sprintf("Top-%d >1U Merged SFSes:\n", mergedSfsGt1UCount))
		for _, sfs := range sa.MergedSfss {
			if sfs.Distance > 1 {
				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
				printed++
				if printed >= mergedSfsGt1UCount {
					break
				}
			}
		}
	}

	if mergedSfs1UCount > 0 {
		sb.WriteString(fmt.Sprintf("Top-%d 1U Merged SFSes:\n", mergedSfs1UCount))
		printed = 0
		for _, sfs := range sa.MergedSfss {
			if sfs.Distance == 1 {
				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
				printed++
				if printed >= mergedSfs1UCount {
					break
				}
			}
		}
	}

	return sb.String()
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
func (sl *SplitLayout) AnalyzeSfss(corp *Corpus) SfsAnalysis {
	an := SfsAnalysis{
		LayoutName:    sl.Filename,
		CorpusName:    corp.Name,
		TotalTrigrams: corp.TotalTrigramsCount,
	}

	merged := make(map[Trigram]Sfs)

	for tri, cnt := range corp.Trigrams {
		if tri[0] == tri[2] {
			// ignore 0U SFS
			continue
		}

		rune0, ok0 := sl.RuneInfo[tri[0]]
		rune1, ok1 := sl.RuneInfo[tri[1]]
		rune2, ok2 := sl.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			an.Unsupported = append(an.Unsupported, TrigramCount{tri, cnt})
		} else if rune0.Finger == rune2.Finger && rune0.Finger != rune1.Finger {
			dist := calcDistance(rune0, rune2)
			perc := float64(cnt) / float64(corp.TotalTrigramsCount)
			an.Sfss = append(an.Sfss, Sfs{Trigram: tri, Distance: dist, Count: cnt, Percentage: perc})
			an.TotalSfsCount += cnt

			// keep track of skipgrams with no middle character
			var mergedTrigram Trigram
			if tri[0] < tri[2] {
				mergedTrigram = Trigram{tri[0], '_', tri[2]}
			} else {
				mergedTrigram = Trigram{tri[2], '_', tri[0]}
			}
			if existingSfs, ok := merged[mergedTrigram]; ok {
				existingSfs.Count += cnt
				existingSfs.Percentage += perc
				merged[mergedTrigram] = existingSfs
			} else {
				merged[mergedTrigram] = Sfs{
					Trigram:    mergedTrigram,
					Distance:   dist,
					Count:      cnt,
					Percentage: perc,
				}
			}
		}
	}

	// calculate the percentage in a single div op
	an.TotalSfsPerc = float64(an.TotalSfsCount) / float64(an.TotalTrigrams)

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

func calcDistance(rune0 KeyInfo, rune2 KeyInfo) float32 {
	// thumbs row uses columns
	if rune0.Row == 3 {
		if rune0.Column > rune2.Column {
			return float32(rune0.Column - rune2.Column)
		}
		return float32(rune2.Column - rune0.Column)
	}

	// other fingers, same column diff
	if rune0.Column == rune2.Column {
		if rune0.Row > rune2.Row {
			return float32(rune0.Row - rune2.Row)
		}
		return float32(rune2.Row - rune0.Row)
	}

	// cases of same finger, diff column (index and pinky)
	dx := IfThen(rune0.Row > rune2.Row, rune0.Row-rune2.Row, rune2.Row-rune0.Row)
	dy := IfThen(rune0.Column > rune2.Column, rune0.Column-rune2.Column, rune2.Column-rune0.Column)
	mul := dx*dx + dy*dy
	if mul == 1 {
		return 1
	}
	if mul == 2 {
		return math.Sqrt2
	}
	return float32(math.Sqrt(float64(mul)))
}
