// Package layout provides functionality for analyzing keyboard layouts.
package layout

import (
	"fmt"
	"sort"
	"strings"

	corpus "github.com/rbscholtus/kb/internal/corpus"
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
	var pairs []Pair[rune, uint64]
	for r, c := range na {
		pairs = append(pairs, Pair[rune, uint64]{r, c})
	}

	// Sort pairs based on the count in descending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	var parts []string
	for _, pair := range pairs {
		parts = append(parts, fmt.Sprintf("%c (%s)", pair.Key, Comma(pair.Value)))
	}

	return strings.Join(parts, ", ")
}

// AnalyzeHandUsage analyzes the hand usage for a layout and corpus.
func (sl *SplitLayout) AnalyzeHandUsage(corp *corpus.Corpus) HandAnalysis {
	var handUsage [2]Usage
	var rowUsage [4]Usage
	var columnUsage [12]Usage
	var fingerUsage [10]Usage
	var totalUnigramCount uint64 = 0
	var runesUnavailable = make(map[rune]uint64, 0)

	// Iterate over unigrams in the corpus and calculate usage statistics.
	for r, count := range corp.Unigrams {
		info, ok := sl.RuneInfo[r]
		if !ok {
			runesUnavailable[r] += count
			continue
		}

		totalUnigramCount += count
		hand := ifThen(info.Hand == "left", 0, 1)
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
	Bigram corpus.Bigram
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
func (sl *SplitLayout) SimpleSfbs(corp *corpus.Corpus) float64 {
	var totalCount uint64
	for bi, cnt := range corp.Bigrams {
		if bi[0] != bi[1] && sl.isSameFinger(bi) {
			totalCount += cnt
		}
	}
	return float64(totalCount) / float64(corp.TotalBigramsCount)
}

// AnalyzeSfbs analyzes the SFBs for a layout and corpus.
func (sl *SplitLayout) AnalyzeSfbs(corp *corpus.Corpus) SfbAnalysis {
	// Get the SFBs in the corpus that occur in this layout, sorted by counts.
	sfbs, totalCount := sl.extractSfbs(corp)
	sort.Slice(sfbs, func(i, j int) bool {
		return sfbs[i].Count > sfbs[j].Count
	})

	// Calculate the percentages.
	for i := range sfbs {
		sfbs[i].Percentage = float64(sfbs[i].Count) / float64(corp.TotalBigramsCount)
	}

	return SfbAnalysis{
		LayoutName:    sl.Filename,
		CorpusName:    corp.Name,
		TotalBigrams:  corp.TotalBigramsCount,
		Sfbs:          sfbs,
		TotalSfbCount: totalCount,
		TotalSfbPerc:  float64(totalCount) / float64(corp.TotalBigramsCount),
	}
}

// extractSfbs extracts the SFBs from the corpus for a layout.
func (sl *SplitLayout) extractSfbs(corp *corpus.Corpus) ([]Sfb, uint64) {
	var sfbs []Sfb
	var totalCount uint64
	for bi, cnt := range corp.Bigrams {
		if bi[0] != bi[1] && sl.isSameFinger(bi) {
			sfbs = append(sfbs, Sfb{Bigram: bi, Count: cnt})
			totalCount += cnt
		}
	}
	return sfbs, totalCount
}

// isSameFinger checks if two runes are pressed by the same finger.
func (sl *SplitLayout) isSameFinger(bi corpus.Bigram) bool {
	info0, ok0 := sl.RuneInfo[bi[0]]
	info1, ok1 := sl.RuneInfo[bi[1]]
	return ok0 && ok1 && info0.Finger == info1.Finger
}
