// Package layout provides common structs and utility functions.
package layout

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func (an *Analyser) HandUsageString() string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = true
	tw.Style().Title.Align = text.AlignCenter
	tw.SetTitle(fmt.Sprintf("Hand Usage Analysis (%s)", an.Layout.Name))

	// Define column headers
	fingerAbbreviations := table.Row{"LP", "LP", "LR", "LM", "LI", "LI", "RI", "RI", "RM", "RR", "RP", "RP"}
	tw.AppendHeader(fingerAbbreviations, table.RowConfig{AutoMerge: true})

	// Append row with ColumnUsage data
	columnUsageRow := make(table.Row, 12)
	for i := range columnUsageRow {
		columnUsageRow[i] = fmt.Sprintf("%.2f%%", an.HandUsage.ColumnUsage[i])
	}
	tw.AppendRow(columnUsageRow)

	// Append row with FingerUsage data and merged cells
	fingerUsageRow := make(table.Row, 12)
	fi := [12]int{0, 0, 1, 2, 3, 3, 6, 6, 7, 8, 9, 9}
	for i := range fingerUsageRow {
		fingerUsageRow[i] = fmt.Sprintf("%.2f%%", an.HandUsage.FingerUsage[fi[i]])
	}
	tw.AppendRow(fingerUsageRow, table.RowConfig{AutoMerge: true})

	// Append row with HandUsage data in merged cells
	handUsageRow := make(table.Row, 12)
	hi := [12]int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}
	for i := range handUsageRow {
		handUsageRow[i] = fmt.Sprintf("%.2f%%", an.HandUsage.HandUsage[hi[i]])
	}
	tw.AppendRow(handUsageRow, table.RowConfig{AutoMerge: true})

	return tw.Render()
}

func (an *Analyser) RowUsageString() string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = true
	tw.Style().Title.Align = text.AlignCenter
	tw.SetTitle("Row Usage")

	tw.AppendRow(table.Row{"Top", fmt.Sprintf("%.2f%%", an.HandUsage.RowUsage[0])})
	tw.AppendRow(table.Row{"Home", fmt.Sprintf("%.2f%%", an.HandUsage.RowUsage[1])})
	tw.AppendRow(table.Row{"Bottom", fmt.Sprintf("%.2f%%", an.HandUsage.RowUsage[2])})

	return tw.Render()
}

func (an *Analyser) MetricsString() string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = true
	tw.Style().Title.Align = text.AlignCenter
	tw.SetTitle("Metrics")

	data := []table.Row{
		{"Bigrams",
			fmt.Sprintf("SFB: %.2f%%", an.Metrics["SFB"]),
			fmt.Sprintf("LSB: %.2f%%", an.Metrics["LSB"]),
			fmt.Sprintf("FSB: %.2f%%", an.Metrics["FSB"]),
			fmt.Sprintf("HSB: %.2f%%", an.Metrics["HSB"]),
		},
		{"Skipgrams",
			fmt.Sprintf("SFS: %.2f%%", an.Metrics["SFS"]),
			fmt.Sprintf("LSS: %.2f%%", an.Metrics["LSS"]),
			fmt.Sprintf("FSS: %.2f%%", an.Metrics["FSS"]),
			fmt.Sprintf("HSS: %.2f%%", an.Metrics["HSS"]),
		},
		{"Trigrams",
			fmt.Sprintf("ALT: %.2f%%", an.Metrics["ALT"]),
			fmt.Sprintf("ROL: %.2f%%", an.Metrics["ROL"]),
			fmt.Sprintf("ONE: %.2f%%", an.Metrics["ONE"]),
			fmt.Sprintf("RED: %.2f%%", an.Metrics["RED"]),
		},
	}
	tw.AppendRows(data)

	return tw.Render()
}

func createTable(title string, style table.Style) table.Writer {
	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.SetStyle(style)
	tw.Style().Title.Align = text.AlignCenter
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Distance", Transformer: Fraction},
		{Name: "Angle", Transformer: Fraction},
		{Name: "Count", Transformer: Thousands, TransformerFooter: Thousands},
		{Name: "%", Transformer: Percentage, TransformerFooter: Percentage},
	})
	tw.SortBy([]table.SortBy{{Name: "orderby", Mode: table.DscNumeric}})
	tw.SetTitle(title)
	return tw
}

// String returns a string representation of the SFB analysis.
func (sa SfbAnalysis) String() string {
	t := createTable("Same Finger Bigrams", table.StyleColoredBlackOnCyanWhite)
	t.AppendHeader(table.Row{"orderby", "SFB", "Distance", "Count", "%", "   "})
	for _, sfb := range sa.Sfbs {
		t.AppendRow([]any{sfb.Count, sfb.Bigram, sfb.Distance, sfb.Count, sfb.Percentage})
	}
	t.AppendFooter(table.Row{"", "", "", sa.TotalSfbCount, sa.TotalSfbPerc})
	return t.Pager(table.PageSize(sa.NumRowsInOutput)).Render()
}

// String returns a string representation of the SFS analysis.
func (sa SfsAnalysis) String() string {
	title := fmt.Sprintf("Same Finger Skipgrams (>=%.1fU)", sa.MinDistanceInOutput)
	t := createTable(title, table.StyleColoredCyanWhiteOnBlack)
	t.AppendHeader(table.Row{"orderby", "SFS", "Distance", "Count", "%", "   "})
	for _, sfs := range sa.MergedSfss {
		if sfs.Distance >= sa.MinDistanceInOutput {
			t.AppendRow([]any{sfs.Count, sfs.Trigram, sfs.Distance, sfs.Count, sfs.Percentage})
		}
	}
	t.AppendFooter(table.Row{"", "", "", sa.TotalSfsCount, sa.TotalSfsPerc})
	return t.Pager(table.PageSize(sa.NumRowsInOutput)).Render()
}

// String returns a string representation of the LSB analysis.
func (la *LsbAnalysis) String() string {
	t := createTable("Lateral Stretch Bigrams", table.StyleColoredBlackOnBlueWhite)
	t.AppendHeader(table.Row{"orderby", "LSB", "Distance", "Count", "%", "   "})
	for _, lsb := range la.Lsbs {
		t.AppendRow([]any{lsb.Count, lsb.Bigram, lsb.HorDistance, lsb.Count, lsb.Percentage})
	}
	t.AppendFooter(table.Row{"", "", "", la.TotalLsbCount, la.TotalLsbPerc})
	return t.Pager(table.PageSize(la.NumRowsInOutput)).Render()
}

// String returns a string representation of the Scissors analysis.
func (la *ScissorAnalysis) String() string {
	t := createTable("Full Scissor Bigrams", table.StyleColoredBlackOnMagentaWhite)
	t.AppendHeader(table.Row{"orderby", "Bigr", "Adj.", "Row", "Angle", "Count", "%"})
	for _, sci := range la.Scissors {
		t.AppendRow([]any{sci.Count, sci.Bigram, sci.FingerDistance, sci.RowDistance, sci.Angle, sci.Count, sci.Percentage})
	}
	t.AppendFooter(table.Row{"", "", "", "", "", la.TotalScissorCount, la.TotalScissorPerc})
	return t.Pager(table.PageSize(la.NumRowsInOutput)).Render()
}
