package main

import (

	// ly "github.com/rbscholtus/kb/internal/layout"

	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "layout-cli",
		Usage: "A CLI tool for various layout operations",
		Commands: []*cli.Command{
			viewCommand,
			analyseCommand,
			optimiseCommand,
			compareCommand,
			rankCommand,
			experimentCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// func main() {
// 	f, err := ParseFlags()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// Load the corpus from the specified file
// 	corp, err := ly.NewCorpusFromFile(f.Corpus, "data/corpus/"+f.Corpus)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// Load the layout from the specified file
// 	layout, err := ly.NewLayoutFromFile(f.Layout, "data/layouts/"+f.Layout)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	an := ly.NewAnalyser(layout, corp)
// 	// godump.Dump(an.HandUsage)
// 	// godump.Dump(an.Metrics)
// 	fmt.Println(an.HandUsageString())
// 	fmt.Println(an.MetricsString())
// }

// func doHandUsage(lay *ly.SplitLayout, corp *ly.Corpus) {
// 	handInfo := lay.AnalyzeHandUsage(corp)
// 	fmt.Println(handInfo)
// }

// func I2col[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
// 	return uint8(i % 12)
// }

// func I2row[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
// 	return uint8(i / 12)
// }

// func Experiment() {
// 	// a := ly.SortedMap(corp.Unigrams)
// 	// for _, v := range a[:10] {
// 	// 	fmt.Println(v.Key.String(), " ", v.Count)
// 	// }

// 	// var fav *ly.KeyInfo

// 	// keys := [42]ly.KeyInfo{}
// 	// for i := range keys {
// 	// 	keys[i].Column = I2col(i)
// 	// 	keys[i].Row = I2row(i)
// 	// 	if i == 17 {
// 	// 		fav = &keys[i]
// 	// 	}
// 	// }
// 	// for _, k := range keys {
// 	// 	fmt.Println(k)
// 	// }
// 	// fmt.Println(*fav)
// 	// fmt.Println(fav)
// 	// fmt.Println(fav.Row)

// 	// for i := range 42 {
// 	// 	ik := ly.NewKeyInfo(uint8(i/12), uint8(i%12))
// 	// 	for j := range 42 {
// 	// 		if i == j {
// 	// 			continue
// 	// 		}
// 	// 		jk := ly.NewKeyInfo(uint8(j/12), uint8(j%12))
// 	// 		if ik.Finger == jk.Finger {
// 	// 			// fmt.Printf("SFB: %d,%d: %v %v\n", i, j, ik, jk)
// 	// 		}
// 	// 	}
// 	// }
// }
