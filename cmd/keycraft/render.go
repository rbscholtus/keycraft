package main

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

// MetricDetailsString renders metric details as a paginated table.
func MetricDetailsString(ma *kc.MetricDetails, nrows int) string {
	t := createSimpleTable()

	// Collect unique custom keys
	customKeys := []string{}
	customKeySet := make(map[string]bool)
	for _, fields := range ma.Custom {
		for k := range fields {
			if !customKeySet[k] {
				customKeySet[k] = true
				customKeys = append(customKeys, k)
			}
		}
		break // we don't need to visit all rows - they all have the same custom fields
	}

	// Header
	header := table.Row{"orderby", ma.Metric, "Count", "%", "Dist"}
	for _, ck := range customKeys {
		header = append(header, ck)
	}
	t.AppendHeader(header)

	// Rows
	for ngram := range ma.NGramCount {
		row := []any{
			ma.NGramCount[ngram],
			ngram,
			ma.NGramCount[ngram],
			float64(ma.NGramCount[ngram]) / float64(ma.CorpusNGramC),
			ma.NGramDist[ngram],
		}
		for _, ck := range customKeys {
			if fields, ok := ma.Custom[ngram]; ok {
				row = append(row, fields[ck])
			} else {
				row = append(row, "")
			}
		}
		t.AppendRow(row)
	}

	footer := table.Row{"", "", ma.TotalNGrams, float64(ma.TotalNGrams) / float64(ma.CorpusNGramC)}
	for range customKeys {
		footer = append(footer, "")
	}
	t.AppendFooter(footer)

	return t.Pager(table.PageSize(nrows)).Render()
}

// createSimpleTable returns a configured table writer with rounded style and common settings.
func createSimpleTable() table.Writer {
	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.SetStyle(table.StyleRounded)
	tw.Style().Title.Align = text.AlignCenter
	tw.Style().Box.PaddingLeft = ""
	tw.Style().Box.PaddingRight = ""
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Distance", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Dist", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Row", Transformer: Fraction},
		{Name: "Count", Transformer: Thousands, TransformerFooter: Thousands},
		{Name: "%", Transformer: Percentage, TransformerFooter: Percentage},
		{Name: "Cum%", Transformer: Percentage, TransformerFooter: Percentage},
		{Name: "Δrow", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Δcol", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Angle", Transformer: Angle, TransformerFooter: Angle},
		{Name: "Length", Align: text.AlignRight},
	})
	tw.SortBy([]table.SortBy{{Name: "orderby", Mode: table.DscNumeric}})
	return tw
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

// EmptyStyle returns a table style with minimal separators for compact layout display.
func EmptyStyle() table.Style {
	s := table.StyleDefault
	s.Box = table.StyleBoxRounded
	s.Box = table.BoxStyle{
		BottomLeft:       s.Box.BottomLeft,
		BottomRight:      s.Box.BottomRight,
		BottomSeparator:  s.Box.BottomSeparator,
		Left:             " ",
		LeftSeparator:    s.Box.LeftSeparator,
		MiddleHorizontal: " ",
		MiddleSeparator:  s.Box.MiddleSeparator,
		MiddleVertical:   " ",
		Right:            " ",
		RightSeparator:   s.Box.RightSeparator,
		TopLeft:          s.Box.TopLeft,
		TopRight:         s.Box.TopRight,
		TopSeparator:     s.Box.TopSeparator,
		UnfinishedRow:    " ",
	}
	return s
}

// Comma formats a numeric value with thousand separators.
// Accepts integer types (int, int32, int64, uint, uint32, uint64) via generics.
func Comma[T ~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64](v T) string {
	// Convert to uint64 for processing
	val := uint64(v)

	// Calculate the number of digits and commas needed.
	var count byte
	for n := val; n != 0; n = n / 10 {
		count++
	}
	count += (count - 1) / 3

	// Create an output slice to hold the formatted number.
	output := make([]byte, count)
	j := len(output) - 1

	// Populate the output slice with digits and commas.
	var counter byte
	for val > 9 {
		output[j] = byte(val%10) + '0'
		val = val / 10
		j--
		if counter == 2 {
			counter = 0
			output[j] = ','
			j--
		} else {
			counter++
		}
	}

	output[j] = byte(val) + '0'

	return string(output)
}

// Thousands formats a uint64 count using comma separators, or falls back to %v for other types.
func Thousands(val any) string {
	if number, ok := val.(uint64); ok {
		return Comma(number)
	}
	return fmt.Sprintf("%v", val)
}

// Fraction formats a float64 with two decimals, or falls back to %v for other types.
func Fraction(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f", number)
	}
	return fmt.Sprintf("%v", val)
}

// Angle formats numeric values as degree strings.
func Angle(val any) string {
	switch v := val.(type) {
	case float32, float64:
		return fmt.Sprintf("%.1f°", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d°", v)
	default:
		return fmt.Sprint(val)
	}
}

// Percentage formats a fractional value (0..1) as a percentage with two decimals.
func Percentage(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.2f%%", 100*number)
	}
	return fmt.Sprintf("%v", val)
}

// Percentage3 formats a fractional value (0..1) as a percentage with three decimals.
func Percentage3(val any) string {
	if number, ok := val.(float64); ok {
		return fmt.Sprintf("%.3f%%", 100*number)
	}
	return fmt.Sprintf("%v", val)
}
