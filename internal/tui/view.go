package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// ASCII templates for rendering keyboard layouts in the terminal.
const (
	rowstagTempl = `╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮   
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│   
╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮  
 │%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│  
 ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮
   │%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
   ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯
               │%3s│%3s│%3s│  │%3s│%3s│%3s│            
               ╰───┴───┴───╯  ╰───┴───┴───╯            `
	anglemodTempl = `╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮   
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│   
╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮  
 │%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│  
 ╰┬┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮
  ╰┤%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
   ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯
               │%3s│%3s│%3s│  │%3s│%3s│%3s│            
               ╰───┴───┴───╯  ╰───┴───┴───╯            `
	orthoTempl = `╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
├───┼───┼───┼───┼───┼───┤  ├───┼───┼───┼───┼───┼───┤
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
├───┼───┼───┼───┼───┼───┤  ├───┼───┼───┼───┼───┼───┤
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯
            │%3s│%3s│%3s│  │%3s│%3s│%3s│            
            ╰───┴───┴───╯  ╰───┴───┴───╯            `
	colstagTempl = `        ╭───┬───┬───╮          ╭───┬───┬───╮        
╭───┬───┤%3s│%3s│%3s├───╮  ╭───┤%3s│%3s│%3s├───┬───╮
│%3s│%3s├───┼───┼───┤%3s│  │%3s├───┼───┼───┤%3s│%3s│
├───┼───┤%3s│%3s│%3s├───┤  ├───┤%3s│%3s│%3s├───┼───┤
│%3s│%3s├───┼───┼───┤%3s│  │%3s├───┼───┼───┤%3s│%3s│
├───┼───┤%3s│%3s│%3s├───┤  ├───┤%3s│%3s│%3s├───┼───┤
│%3s│%3s├───┼───┼───┤%3s│  │%3s├───┼───┼───┤%3s│%3s│
╰───┴───╯   │%3s│%3s├───┤  ├───┤%3s│%3s│   ╰───┴───╯
            ╰───┴───┤%3s│  │%3s├───┴───╯            
                    ╰───╯  ╰───╯                    `
)

// RenderView renders the view results to stdout.
// Displays board, hand balance, row balance, and stats overview for each layout.
func RenderView(result *kc.ViewResult) error {
	if len(result.Analysers) < 1 {
		return fmt.Errorf("need at least 1 layout")
	}

	// Make a table with a column for each layout
	twOuter := table.NewWriter()
	twOuter.SetStyle(EmptyStyle())
	twOuter.Style().Options.SeparateRows = true
	colConfigs := make([]table.ColumnConfig, 0, len(result.Analysers))
	for i := range len(result.Analysers) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i + 2,
			AlignHeader: text.AlignCenter, Align: text.AlignCenter})
	}
	twOuter.SetColumnConfigs(colConfigs)

	// Add header
	h := table.Row{""}
	for _, an := range result.Analysers {
		h = append(h, an.Layout.Name)
	}
	twOuter.AppendHeader(h)

	// Layout picture
	h = table.Row{"Board"}
	for _, an := range result.Analysers {
		h = append(h, SplitLayoutString(an.Layout))
	}
	twOuter.AppendRow(h)

	// Hand balance
	h = table.Row{"Hand"}
	for _, an := range result.Analysers {
		h = append(h, HandUsageString(an))
	}
	twOuter.AppendRow(h)

	// Row balance
	h = table.Row{"Row"}
	for _, an := range result.Analysers {
		h = append(h, RowUsageString(an))
	}
	twOuter.AppendRow(h)

	// Metrics overview
	h = table.Row{"Stats"}
	for _, an := range result.Analysers {
		h = append(h, MetricsString(an))
	}
	twOuter.AppendRow(h)

	// Print layout(s) in the table
	fmt.Println(twOuter.Render())
	return nil
}

// SplitLayoutString returns a formatted ASCII representation of a keyboard layout.
func SplitLayoutString(sl *kc.SplitLayout) string {
	switch sl.LayoutType {
	case kc.ANGLEMOD:
		return genLayoutStringFor(sl, anglemodTempl, nil)
	case kc.ORTHO:
		return genLayoutStringFor(sl, orthoTempl, nil)
	case kc.COLSTAG:
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
		return genLayoutStringFor(sl, colstagTempl, mapper[:])
	default: // kc.ROWSTAG
		return genLayoutStringFor(sl, rowstagTempl, nil)
	}
}

// genLayoutStringFor applies runes to an ASCII template, optionally reordering via mapper.
func genLayoutStringFor(sl *kc.SplitLayout, template string, mapper []int) string {
	args := make([]any, len(sl.Runes))
	for i, r := range sl.Runes {
		var m rune
		if mapper != nil {
			m = sl.Runes[mapper[i]]
		} else {
			m = r
		}
		switch m {
		case 0:
			args[i] = " "
		case ' ':
			args[i] = " _ "
		default:
			args[i] = string(m) + " "
		}
	}
	return fmt.Sprintf(strings.ReplaceAll(template, " ", "\u00A0"), args...)
}

// HandUsageString renders column, finger, and hand usage percentages as a table.
func HandUsageString(an *kc.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""

	// Define column headers
	fingerAbbreviations := table.Row{"LP", "LP", "LR", "LM", "LI", "LI", "RI", "RI", "RM", "RR", "RP", "RP"}
	colConfigs := make([]table.ColumnConfig, 0, len(fingerAbbreviations))
	for i := range len(fingerAbbreviations) {
		colConfigs = append(colConfigs, table.ColumnConfig{Number: i,
			AlignHeader: text.AlignCenter, Align: text.AlignCenter})
	}
	tw.SetColumnConfigs(colConfigs)
	tw.AppendHeader(fingerAbbreviations, table.RowConfig{AutoMerge: true})

	// Append row with ColumnUsage data
	columnUsageRow := make(table.Row, 12)
	for i := range columnUsageRow {
		key := "C" + strconv.Itoa(i)
		columnUsageRow[i] = fmt.Sprintf("%.1f", an.Metrics[key])
	}
	tw.AppendRow(columnUsageRow)

	// Append row with FingerUsage data and merged cells
	fingerUsageRow := make(table.Row, 12)
	fi := [12]int{0, 0, 1, 2, 3, 3, 6, 6, 7, 8, 9, 9}
	for i, v := range fi {
		key := "F" + strconv.Itoa(v)
		fingerUsageRow[i] = fmt.Sprintf("%.1f", an.Metrics[key])
	}
	tw.AppendRow(fingerUsageRow, table.RowConfig{AutoMerge: true})

	// Append row with HandUsage data in merged cells
	handUsageRow := make(table.Row, 12)
	hi := [12]int{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1}
	for i, v := range hi {
		key := "H" + strconv.Itoa(v)
		handUsageRow[i] = fmt.Sprintf("%.1f%%", an.Metrics[key])
	}
	tw.AppendRow(handUsageRow, table.RowConfig{AutoMerge: true})

	return tw.Render()
}

// RowUsageString renders row usage percentages (Top, Home, Bottom, Thumb) as a table.
func RowUsageString(an *kc.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 2, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 3, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
		{Number: 4, AlignHeader: text.AlignCenter, Align: text.AlignCenter},
	})

	tw.AppendRow(table.Row{"Top", "Home", "Bottom", "Thumb"})
	tw.AppendRow(table.Row{
		fmt.Sprintf("%.1f%%", an.Metrics["R0"]), fmt.Sprintf("%.1f%%", an.Metrics["R1"]),
		fmt.Sprintf("%.1f%%", an.Metrics["R2"]), fmt.Sprintf("%.1f%%", an.Metrics["R3"]),
	})

	return tw.Render()
}

// MetricsString renders a summary of key layout metrics (SFB, LSB, rolls, etc.) as a table.
func MetricsString(an *kc.Analyser) string {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = true
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignJustify},
		{Number: 2, Align: text.AlignJustify},
		{Number: 3, Align: text.AlignJustify},
		{Number: 4, Align: text.AlignJustify},
	})

	data := []table.Row{
		{
			fmt.Sprintf("SFB: %.2f%%", an.Metrics["SFB"]),
			fmt.Sprintf("LSB: %.2f%%", an.Metrics["LSB"]),
			fmt.Sprintf("FSB: %.2f%%", an.Metrics["FSB"]),
			fmt.Sprintf("HSB: %.2f%%", an.Metrics["HSB"]),
		},
		{
			fmt.Sprintf("SFS: %.2f%%", an.Metrics["SFS"]),
			fmt.Sprintf("LSS: %.2f%%", an.Metrics["LSS"]),
			fmt.Sprintf("FSS: %.2f%%", an.Metrics["FSS"]),
			fmt.Sprintf("HSS: %.2f%%", an.Metrics["HSS"]),
		},
		{
			fmt.Sprintf("ALT: %.2f%%", an.Metrics["ALT"]),
			fmt.Sprintf(".NML: %.2f%%", an.Metrics["ALT-NML"]),
			fmt.Sprintf(".SFS: %.2f%%", an.Metrics["ALT-SFS"]),
			"",
		},
		{
			fmt.Sprintf("2RL: %.2f%%", an.Metrics["2RL"]),
			fmt.Sprintf(".IN: %.2f%%", an.Metrics["2RL-IN"]),
			fmt.Sprintf(".OUT: %.2f%%", an.Metrics["2RL-OUT"]),
			fmt.Sprintf(".SFB: %.2f%%", an.Metrics["2RL-SFB"]),
		},
		{
			fmt.Sprintf("3RL: %.2f%%", an.Metrics["3RL"]),
			fmt.Sprintf(".IN: %.2f%%", an.Metrics["3RL-IN"]),
			fmt.Sprintf(".OUT: %.2f%%", an.Metrics["3RL-OUT"]),
			fmt.Sprintf(".SFB: %.2f%%", an.Metrics["3RL-SFB"]),
		},
		{
			fmt.Sprintf("RED: %.2f%%", an.Metrics["RED"]),
			fmt.Sprintf(".NML: %.2f%%", an.Metrics["RED-NML"]),
			fmt.Sprintf(".WEAK: %.2f%%", an.Metrics["RED-WEAK"]),
			fmt.Sprintf(".SFS: %.2f%%", an.Metrics["RED-SFS"]),
		},
		{
			fmt.Sprintf("I:O: %.2f", an.Metrics["IN:OUT"]),
			fmt.Sprintf("FLW: %.2f%%", an.Metrics["FLW"]),
			"",
			"",
		},
		{
			fmt.Sprintf("RBL: %.2f", an.Metrics["RBL"]),
			fmt.Sprintf("FBL: %.2f%%", an.Metrics["FBL"]),
			fmt.Sprintf("POH: %.2f%%", an.Metrics["POH"]),
			"",
		},
	}
	tw.AppendRows(data)

	return tw.Render()
}
