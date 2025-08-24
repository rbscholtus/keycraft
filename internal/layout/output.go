// Package layout provides common structs and utility functions.
package layout

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const ( //\u00A0
	rowstagTempl = `╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮   
│%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│   
╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮  
 │%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│  
 ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮
   │%3s│%3s│%3s│%3s│%3s│%3s│  │%3s│%3s│%3s│%3s│%3s│%3s│
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

// String returns a string representation of the layout
func (sl *SplitLayout) String() string {
	switch sl.LayoutType {
	case ORTHO:
		return sl.genLayoutString(orthoTempl, nil)
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
		return sl.genLayoutString(colstagTempl, mapper[:])
	default:
		return sl.genLayoutString(rowstagTempl, nil)
	}
}

func (sl *SplitLayout) genLayoutString(template string, mapper []int) string {
	// make array for filling in template
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

	// fill in template
	return fmt.Sprintf(strings.ReplaceAll(template, " ", "\u00A0"), args...)
}

func createSimpleTable() table.Writer {
	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.Style().Title.Align = text.AlignCenter
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "orderby", Hidden: true},
		{Name: "Distance", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Dist", Transformer: Fraction, TransformerFooter: Fraction},
		{Name: "Row", Transformer: Fraction},
		{Name: "Angle", Transformer: Fraction},
		{Name: "Count", Transformer: Thousands, TransformerFooter: Thousands},
		{Name: "%", Transformer: Percentage, TransformerFooter: Percentage},
	})
	tw.SortBy([]table.SortBy{{Name: "orderby", Mode: table.DscNumeric}})
	return tw
}

func (ma *MetricDetails) String() string {
	t := createSimpleTable()
	t.SetStyle(table.StyleRounded)

	// Collect all unique custom keys
	customKeys := []string{}
	customKeySet := make(map[string]bool)
	for _, fields := range ma.Custom {
		for k := range fields {
			if !customKeySet[k] {
				customKeySet[k] = true
				customKeys = append(customKeys, k)
			}
		}
	}

	// Build header row
	header := table.Row{"orderby", ma.Metric, "Dist", "Count", "%"}
	for _, ck := range customKeys {
		header = append(header, ck)
	}
	t.AppendHeader(header)

	// Build rows
	for ngram := range ma.NGramCount {
		row := []any{
			ma.NGramCount[ngram],
			ngram,
			ma.NGramDist[ngram],
			ma.NGramCount[ngram],
			float64(ma.NGramCount[ngram]) / float64(ma.CorpusNGramC),
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

	// Build footer
	footer := table.Row{"", "", "", ma.TotalNGrams, float64(ma.TotalNGrams) / float64(ma.CorpusNGramC)}
	for range customKeys {
		footer = append(footer, "")
	}
	t.AppendFooter(footer)

	return t.Pager(table.PageSize(15)).Render()
}
