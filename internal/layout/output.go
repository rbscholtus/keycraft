// Package layout provides common structs and utility functions.
package layout

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	rowstagTempl = `%s
╭─────┬─────┬─────┬─────┬─────┬─────╮   ╭─────┬─────┬─────┬─────┬─────┬─────╮
│ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │   │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
╰┬────┴┬────┴┬────┴┬────┴┬────┴┬────┴╮  ╰┬────┴┬────┴┬────┴┬────┴┬────┴┬────┴╮
 │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │   │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
 ╰──┬──┴──┬──┴──┬──┴──┬──┴──┬──┴──┬──┴──╮╰──┬──┴──┬──┴──┬──┴──┬──┴──┬──┴──┬──┴──╮
    │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │   │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
    ╰─────┴─────┴─────┼─────┼─────┼─────┤   ├─────┼─────┼─────┼─────┴─────┴─────╯
                      │ %3s │ %3s │ %3s │   │ %3s │ %3s │ %3s │                 
                      ╰─────┴─────┴─────╯   ╰─────┴─────┴─────╯`
	orthoTempl = `%s
╭─────┬─────┬─────┬─────┬─────┬─────╮         ╭─────┬─────┬─────┬─────┬─────┬─────╮
│ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │         │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
├─────┼─────┼─────┼─────┼─────┼─────┤         ├─────┼─────┼─────┼─────┼─────┼─────┤
│ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │         │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
├─────┼─────┼─────┼─────┼─────┼─────┤         ├─────┼─────┼─────┼─────┼─────┼─────┤
│ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │         │ %3s │ %3s │ %3s │ %3s │ %3s │ %3s │
╰─────┴─────┴─────┼─────┼─────┼─────┤         ├─────┼─────┼─────┼─────┴─────┴─────╯
                  │ %3s │ %3s │ %3s │         │ %3s │ %3s │ %3s │                 
                  ╰─────┴─────┴─────╯         ╰─────┴─────┴─────╯`
	colstagTempl = `%s
            ╭─────┬─────┬─────╮                     ╭─────┬─────┬─────╮
╭─────┬─────┤ %3s │ %3s │ %3s ├─────╮         ╭─────┤ %3s │ %3s │ %3s ├─────┬─────╮
│ %3s │ %3s ├─────┼─────┼─────┤ %3s │         │ %3s ├─────┼─────┼─────┤ %3s │ %3s │
├─────┼─────┤ %3s │ %3s │ %3s ├─────┤         ├─────┤ %3s │ %3s │ %3s ├─────┼─────┤
│ %3s │ %3s ├─────┼─────┼─────┤ %3s │         │ %3s ├─────┼─────┼─────┤ %3s │ %3s │
├─────┼─────┤ %3s │ %3s │ %3s ├─────┤         ├─────┤ %3s │ %3s │ %3s ├─────┼─────┤
│ %3s │ %3s ├─────┼─────┼─────┤ %3s │         │ %3s ├─────┼─────┼─────┤ %3s │ %3s │
╰─────┴─────╯     │ %3s │ %3s ├─────┤         ├─────┤ %3s │ %3s │     ╰─────┴─────╯           
                  ╰─────┴─────┤ %3s │         │ %3s ├─────┴─────╯
                              ╰─────╯         ╰─────╯`
)

// String returns a string representation of the layout
func (sl *SplitLayout) String() string {
	switch sl.LayoutType {
	case ORTHO:
		return sl.genLayoutString(orthoTempl, 84, nil)
	case COLSTAG:
		mapper := [42]int{
			2, 3, 4, 7, 8, 9,
			0, 1, 5, 6, 10, 11,
			14, 15, 16, 19, 20, 21,
			12, 13, 17, 18, 22, 23,
			26, 27, 28, 31, 32, 33,
			24, 25, 29, 30, 34, 35,
			36, 37, 40, 41,
			38, 39,
		}
		return sl.genLayoutString(colstagTempl, 84, mapper[:])
	default:
		return sl.genLayoutString(rowstagTempl, 77, nil)
	}
}

func (sl *SplitLayout) genLayoutString(template string, width int, mapper []int) string {
	// spaces for center alignment
	spaces := strings.Repeat(" ", (width-len(sl.Name))/2)

	// make array for filling in template
	args := make([]any, len(sl.Runes)+1)
	args[0] = spaces + sl.Name
	for i, r := range sl.Runes {
		var m rune
		if mapper != nil {
			m = sl.Runes[mapper[i]]
		} else {
			m = r
		}
		switch m {
		case 0:
			args[i+1] = " "
		case ' ':
			args[i+1] = "spc"
		default:
			args[i+1] = string(m) + " "
		}
	}

	// fill in template
	return fmt.Sprintf(template, args...)
}

func (an *Analyser) HandUsageString() string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleLight)
	tw.Style().Options.SeparateRows = true
	tw.Style().Title.Align = text.AlignCenter
	tw.SetTitle("Hand Usage Analysis")

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
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, Align: text.AlignJustify},
		{Number: 3, Align: text.AlignJustify},
		{Number: 4, Align: text.AlignJustify},
		{Number: 5, Align: text.AlignJustify},
	})
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
			fmt.Sprintf("2RL: %.2f%%", an.Metrics["2RL"]),
			fmt.Sprintf("3RL: %.2f%%", an.Metrics["3RL"]),
			fmt.Sprintf("RED: %.2f%%", an.Metrics["RED"]),
		},
		{"Trigrams",
			fmt.Sprintf("ALT-SFS: %.2f%%", an.Metrics["ALT-SFS"]),
			fmt.Sprintf("2RL-IN: %.2f%%", an.Metrics["2RL-IN"]),
			fmt.Sprintf("3RL-IN: %.2f%%", an.Metrics["3RL-IN"]),
			fmt.Sprintf("RED-BAD: %.2f%%", an.Metrics["RED-BAD"]),
		},
		{"Trigrams",
			fmt.Sprintf("ALT-OTH: %.2f%%", an.Metrics["ALT-OTH"]),
			fmt.Sprintf("2RL-OUT: %.2f%%", an.Metrics["2RL-OUT"]),
			fmt.Sprintf("3RL-OUT: %.2f%%", an.Metrics["3RL-OUT"]),
			fmt.Sprintf("RED-SFS: %.2f%%", an.Metrics["RED-SFS"]),
		},
		{"Trigrams",
			"",
			fmt.Sprintf("2RL-SF: %.2f%%", an.Metrics["2RL-SF"]),
			fmt.Sprintf("3RL-SF: %.2f%%", an.Metrics["3RL-SF"]),
			fmt.Sprintf("RED-OTH: %.2f%%", an.Metrics["RED-OTH"]),
		},
	}
	tw.AppendRows(data)

	tw.AppendRow(table.Row{"Trigrams",
		"",
		fmt.Sprintf("IN:OUT: %.2f", an.Metrics["IN:OUT"]),
		fmt.Sprintf("IN:OUT: %.2f", an.Metrics["IN:OUT"]),
		"",
	}, table.RowConfig{AutoMerge: true})

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
		{Name: "Row", Transformer: Fraction},
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
