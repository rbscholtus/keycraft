package main

import (
	"fmt"

	"github.com/goforj/godump"
	"github.com/jedib0t/go-pretty/v6/table"
	ly "github.com/rbscholtus/kb/internal/layout"
)

func I2col[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i % 12)
}

func I2row[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i / 12)
}

func main() {
	f, err := ParseFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load the corpus from the specified file
	corp, err := ly.NewCorpusFromFile(f.Corpus, "data/corpus/"+f.Corpus)
	if err != nil {
		fmt.Println(err)
		return
	}

	// a := ly.SortedMap(corp.Unigrams)
	// for _, v := range a[:10] {
	// 	fmt.Println(v.Key.String(), " ", v.Count)
	// }

	var fav *ly.KeyInfo

	keys := [42]ly.KeyInfo{}
	for i := range keys {
		keys[i].Column = I2col(i)
		keys[i].Row = I2row(i)
		if i == 17 {
			fav = &keys[i]
		}
	}
	for _, k := range keys {
		fmt.Println(k)
	}
	fmt.Println(*fav)
	fmt.Println(fav)
	fmt.Println(fav.Row)

	for i := range 42 {
		ik := ly.NewKeyInfo(uint8(i/12), uint8(i%12))
		for j := range 42 {
			if i == j {
				continue
			}
			jk := ly.NewKeyInfo(uint8(j/12), uint8(j%12))
			if ik.Finger == jk.Finger {
				// fmt.Printf("SFB: %d,%d: %v %v\n", i, j, ik, jk)
			}
		}
	}

	// Load the layout from the specified file
	layout, err := ly.NewLayoutFromFile(f.Layout, "data/layouts/"+f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}

	an := ly.NewAnalyser(layout, corp)
	godump.Dump(an.Metrics)
	fmt.Println(an.Metrics)

	if f.Optimize {
		fmt.Println(layout)
		fmt.Println(layout.AnalyzeScissors(corp))
		if f.Pins != "" {
			err := layout.LoadPins("data/pins/" + f.Pins)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		layout = layout.Optimise(corp, f.Generations, f.AcceptWorse)
	}

	fmt.Println(layout)
	doHandUsage(layout, corp)
	sfb := layout.AnalyzeSfbs(corp)
	sfs := layout.AnalyzeSfss(corp)
	lsb := layout.AnalyzeLsbs(corp)
	fsb := layout.AnalyzeScissors(corp)

	twOuter := table.NewWriter()
	twOuter.AppendRow(table.Row{sfb, sfs})
	twOuter.AppendRow(table.Row{lsb, "LSS"})
	twOuter.AppendRow(table.Row{fsb, "FSS"})
	twOuter.SetStyle(table.StyleLight)
	twOuter.Style().Options.SeparateRows = true
	// fmt.Println(twOuter.Render())

	if f.Optimize {
		err := layout.SaveToFile("best.kb")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doHandUsage(lay *ly.SplitLayout, corp *ly.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	fmt.Println(handInfo)
}
