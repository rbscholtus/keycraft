package main

import (
	"fmt"
	"math"

	corpus "github.com/rbscholtus/kb/internal/corpus"
	layout "github.com/rbscholtus/kb/internal/layout"
	"github.com/yassinebenaid/godump"
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
	lay, err := layout.LoadFromFile("data/layouts/" + f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(lay)

	doHandUsage(lay, corp)
	doSfbs(lay, corp)
}

func doHandUsage(lay *layout.SplitLayout, corp *corpus.Corpus) {
	handInfo := lay.AnalyzeHandUsage(corp)
	godump.Dump(handInfo)
}

func doSfbs(lay *layout.SplitLayout, corp *corpus.Corpus) {
	sfbInfo := lay.AnalyzeSfbs(corp)
	fmt.Printf("Corpus: %s (%s bigrams) \n", sfbInfo.CorpusName, Comma(sfbInfo.TotalBigrams))
	fmt.Printf("Total SFBs in %v: %s (%.3f%% of corpus)\n",
		sfbInfo.LayoutName, Comma(sfbInfo.TotalSfbCount), 100*sfbInfo.TotalSfbPerc)
	printCount := min(10, len(sfbInfo.Sfbs))
	fmt.Printf("Top-%d SFBs:\n", printCount)
	for i := range printCount {
		sfb := sfbInfo.Sfbs[i]
		fmt.Printf("%2d. %v (%s, %.3f%%)\n", i+1, sfb.Bigram, Comma(sfb.Count), 100*sfb.Percentage)
	}
}

func Comma(vi int) string {
	v := int64(vi)

	// Shortcut for [0, 7]
	if v&^0b111 == 0 {
		return string([]byte{byte(v) + 48})
	}

	// Min int64 can't be negated to a usable value, so it has to be special cased.
	if v == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}
	// Counting the number of digits.
	var count byte = 0
	for n := v; n != 0; n = n / 10 {
		count++
	}

	count += (count - 1) / 3
	if v < 0 {
		v = 0 - v
		count++
	}
	output := make([]byte, count)
	j := len(output) - 1

	var counter byte = 0
	for v > 9 {
		output[j] = byte(v%10) + 48
		v = v / 10
		j--
		if counter == 2 {
			counter = 0
			output[j] = ','
			j--
		} else {
			counter++
		}
	}

	output[j] = byte(v) + 48
	if j == 1 {
		output[0] = '-'
	}
	return string(output)
}
