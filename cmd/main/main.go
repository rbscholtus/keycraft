package main

import (
	"fmt"

	corpus "github.com/rbscholtus/kb/internal/corpus"
	layout "github.com/rbscholtus/kb/internal/layout"
)

func main() {
	f, err := ParseFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load the corpus from the specified file
	corp, err := corpus.NewFromFile(f.Corpus, "data/corpus/"+f.Corpus)
	if err != nil {
		fmt.Println(err)
		return
	}
	// for i, v := range corp.Unigrams {
	// 	fmt.Printf("%c %v\n", i, v)

	// }
	// godump.Dump(corp.Unigrams)

	// Load the layout from the specified file
	layout, err := layout.LoadFromFile("data/layouts/" + f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(layout)
	doHandUsage(layout, corp)
	doSfbs(layout, corp)

	best := layout.Optimise(corp, f.AcceptWorse, f.Generations)
	fmt.Println(best)
	doHandUsage(best, corp)
	doSfbs(best, corp)
}

func doHandUsage(lay *layout.SplitLayout, corp *corpus.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	fmt.Println(handInfo)
	// godump.Dump(handInfo)
}

func doSfbs(lay *layout.SplitLayout, corp *corpus.Corpus) {
	sfbInfo := lay.AnalyzeSfbs(corp)
	fmt.Println(sfbInfo)
}
