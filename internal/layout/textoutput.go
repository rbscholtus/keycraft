package layout

import (
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

// String returns a string representation of the SFB analysis.
func (sa SfbAnalysis) String() string {
	var sb strings.Builder

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{{Name: "Count", Mode: table.DscNumeric}})
	t.SetStyle(table.StyleColoredCyanWhiteOnBlack)
	t.Style().Title.Align = text.AlignCenter

	t.SetTitle("%s (%s)\nSame Finger Bigrams", sa.SplitLayout.Name, sa.SplitLayout.LayoutType)
	t.AppendHeader(table.Row{"SFB", "Distance", "Count", "%"})
	for _, sfb := range sa.Sfbs {
		t.AppendRow([]any{sfb.Bigram, Frac(sfb.Distance), sfb.Count, Perc(sfb.Percentage)})
	}
	t.AppendFooter(table.Row{"", "", Comma(sa.TotalSfbCount), Perc(sa.TotalSfbPerc)})
	t.SetCaption("Corpus: %v (%s bigrams)", sa.Corpus.Name, sa.Corpus.TotalBigramsNoSpace)
	t.Render()

	return sb.String()
}

// String returns a string representation of the SFB analysis.
func (sa SfsAnalysis) String() string {
	var sb strings.Builder

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{{Name: "Count", Mode: table.DscNumeric}})
	t.SetStyle(table.StyleColoredBlueWhiteOnBlack)
	t.Style().Title.Align = text.AlignCenter

	t.SetTitle("%s (%s)\nSame Finger Skipgrams (>=1.2U)", sa.SplitLayout.Name, sa.SplitLayout.LayoutType)
	t.AppendHeader(table.Row{"SFS", "Distance", "Count", "%"})
	for _, sfs := range sa.Sfss {
		if sfs.Distance > 1.2 {
			t.AppendRow([]any{sfs.Trigram, Frac(sfs.Distance), sfs.Count, Perc(sfs.Percentage)})
		}
	}
	t.AppendFooter(table.Row{"", "", Comma(sa.TotalSfsCount), Perc(sa.TotalSfsPerc)})
	t.SetCaption("Corpus: %v (%s trigrams)", sa.Corpus.Name, sa.Corpus.TotalTrigramsCount)
	t.Render()

	return sb.String()
}

// // String returns a string representation of the SFS analysis.
// func (sa SfsAnalysis) String() string {
// 	return sa.SFSString(0, 0, 10, 0)
// }

// // SFSString returns a string representation of an SfsAnalysis
// func (sa SfsAnalysis) SFSString(sfsGt1UCount, sfs1UCount, mergedSfsGt1UCount, mergedSfs1UCount int) string {
// 	var sb strings.Builder
// 	var printed int

// 	sb.WriteString(fmt.Sprintf("Corpus: %s (%s trigrams)\n", sa.CorpusName, Comma(sa.TotalTrigrams)))
// 	sb.WriteString(fmt.Sprintf("Total SFSs: %s (%.2f%% of corpus) in %v\n",
// 		Comma(sa.TotalSfsCount), 100*sa.TotalSfsPerc, sa.LayoutName))

// 	if sfsGt1UCount > 0 {
// 		printed = 0
// 		sb.WriteString(fmt.Sprintf("Top-%d >1U SFSes:\n", sfsGt1UCount))
// 		for _, sfs := range sa.Sfss {
// 			if sfs.Distance > 1.2 {
// 				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
// 					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
// 				printed++
// 				if printed >= sfsGt1UCount {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	if sfs1UCount > 0 {
// 		sb.WriteString(fmt.Sprintf("Top-%d 1U SFSes:\n", sfs1UCount))
// 		printed = 0
// 		for _, sfs := range sa.Sfss {
// 			if sfs.Distance <= 1.2 {
// 				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
// 					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
// 				printed++
// 				if printed >= sfs1UCount {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	if mergedSfsGt1UCount > 0 {
// 		printed = 0
// 		sb.WriteString(fmt.Sprintf("Top-%d >1U Merged SFSes:\n", mergedSfsGt1UCount))
// 		for _, sfs := range sa.MergedSfss {
// 			if sfs.Distance > 1.2 {
// 				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
// 					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
// 				printed++
// 				if printed >= mergedSfsGt1UCount {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	if mergedSfs1UCount > 0 {
// 		sb.WriteString(fmt.Sprintf("Top-%d 1U Merged SFSes:\n", mergedSfs1UCount))
// 		printed = 0
// 		for _, sfs := range sa.MergedSfss {
// 			if sfs.Distance <= 1.2 {
// 				sb.WriteString(fmt.Sprintf("%2d. %v (%.2fU, %s, %.3f%%)\n", printed+1,
// 					sfs.Trigram, sfs.Distance, Comma(sfs.Count), 100*sfs.Percentage))
// 				printed++
// 				if printed >= mergedSfs1UCount {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	return sb.String()
// }

func (la *LsbAnalysis) String() string {
	var sb strings.Builder

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{{Name: "Count", Mode: table.DscNumeric}})
	t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
	t.Style().Title.Align = text.AlignCenter

	t.SetTitle("%s (%s)\nLateral Stretch Bigrams", la.SplitLayout.Name, la.SplitLayout.LayoutType)
	t.AppendHeader(table.Row{"LSB", "Distance", "Count", "%"})
	for _, lsb := range la.Lsbs {
		t.AppendRow([]any{lsb.Bigram, Frac(lsb.Distance), lsb.Count, Perc(lsb.Percentage)})
	}
	t.AppendFooter(table.Row{"", "", Comma(la.TotalLsbCount), Perc(la.TotalLsbPerc)})
	t.SetCaption("Corpus: %v", la.Corpus.Name)
	t.Render()

	return sb.String()
}
