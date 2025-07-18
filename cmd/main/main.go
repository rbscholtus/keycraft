package main

import (
	"fmt"

	layout "github.com/rbscholtus/kb/internal/layout"
)

func main() {
	f, err := ParseFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load the corpus from the specified file
	corp, err := layout.NewCorpusFromFile(f.Corpus, "data/corpus/"+f.Corpus)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load the layout from the specified file
	layout, err := layout.NewLayoutFromFile("data/layouts/" + f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !f.Optimize {
		fmt.Println(layout)
		doHandUsage(layout, corp)
		doSfb(layout, corp)
		doSfs(layout, corp)
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

		err := best.SaveToFile("best.kb")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doHandUsage(lay *layout.SplitLayout, corp *layout.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	fmt.Println(handInfo)
	// godump.Dump(handInfo)
}

func doSfb(lay *layout.SplitLayout, corp *layout.Corpus) {
	sfbInfo := lay.AnalyzeSfbs(corp)
	fmt.Println(sfbInfo)
}

func doSfs(lay *layout.SplitLayout, corp *layout.Corpus) {
	sfsInfo := lay.AnalyzeSfss(corp)
	fmt.Println(sfsInfo)
}
