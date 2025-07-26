package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	ly "github.com/rbscholtus/kb/internal/layout"
)

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

	// Load the layout from the specified file
	layout, err := ly.NewLayoutFromFile(f.Layout, "data/layouts/"+f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !f.Optimize {
		fmt.Println(layout)
		doHandUsage(layout, corp)
		sfb := layout.AnalyzeSfbs(corp)
		sfs := layout.AnalyzeSfss(corp)
		lsb := layout.AnalyzeLsbs(corp)

		twOuter := table.NewWriter()
		twOuter.AppendRow(table.Row{sfb, sfs})
		twOuter.AppendRow(table.Row{lsb, ""})
		twOuter.SetStyle(table.StyleLight)
		twOuter.Style().Options.SeparateRows = true
		fmt.Println(twOuter.Render())
	} else {
		fmt.Println(layout)

		if f.Pins != "" {
			err := layout.LoadPins("data/pins/" + f.Pins)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		best := layout.Optimise(corp, f.Generations, f.AcceptWorse)
		fmt.Println(best)
		doHandUsage(best, corp)
		doSfb(best, corp)
		doSfs(best, corp)
		doLsb(best, corp)

		err := best.SaveToFile("best.kb")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doHandUsage(lay *ly.SplitLayout, corp *ly.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	fmt.Println(handInfo)
}

func doSfb(lay *ly.SplitLayout, corp *ly.Corpus) {
	sfbInfo := lay.AnalyzeSfbs(corp)
	fmt.Println(sfbInfo)
}

func doSfs(lay *ly.SplitLayout, corp *ly.Corpus) {
	sfsInfo := lay.AnalyzeSfss(corp)
	fmt.Println(sfsInfo)
}

func doLsb(layout *ly.SplitLayout, corp *ly.Corpus) {
	lsb := layout.AnalyzeLsbs(corp)
	fmt.Println(lsb)
}
