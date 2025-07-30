package main

import (
	"fmt"
	"path/filepath"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var experimentCommand = &cli.Command{
	Name:  "experiment",
	Usage: "Run experiments",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
	Action: experimentAction,
}

func experimentAction(c *cli.Context) error {
	fmt.Println("Running experiment...")

	corpusFile := c.String("corpus")

	if corpusFile == "" {
		return fmt.Errorf("corpus file is required")
	}

	corpusPath := filepath.Join(corpusDir, corpusFile)
	corp, err := layout.NewCorpusFromFile(corpusFile, corpusPath)
	if err != nil {
		return fmt.Errorf("failed to load corpus from %s: %v", corpusPath, err)
	}

	doExperiment(corp)
	return nil
}

func doExperiment(corp *layout.Corpus) {
	a := layout.SortedMap(corp.Unigrams)
	for _, v := range a[:10] {
		fmt.Println(v.Key.String(), " ", v.Count)
	}

	var fav *layout.KeyInfo

	keys := [42]layout.KeyInfo{}
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
		ik := layout.NewKeyInfo(uint8(i/12), uint8(i%12))
		for j := range 42 {
			if i == j {
				continue
			}
			jk := layout.NewKeyInfo(uint8(j/12), uint8(j%12))
			if ik.Finger == jk.Finger {
				// fmt.Printf("SFB: %d,%d: %v %v\n", i, j, ik, jk)
			}
		}
	}

}

func I2col[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i % 12)
}

func I2row[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i / 12)
}
