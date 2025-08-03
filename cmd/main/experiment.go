package main

import (
	"fmt"
	"path/filepath"

	ly "github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var experimentCommand = &cli.Command{
	Name:   "experiment",
	Usage:  "Run experiments",
	Action: experimentAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
}

func experimentAction(c *cli.Context) error {
	fmt.Println("Running experiment...")

	corpusFile := c.String("corpus")
	if corpusFile == "" {
		return fmt.Errorf("corpus file is required")
	}

	corpusPath := filepath.Join(corpusDir, corpusFile)
	corp, err := ly.NewCorpusFromFile(corpusFile, corpusPath)
	if err != nil {
		return fmt.Errorf("failed to load corpus from %s: %v", corpusPath, err)
	}

	doExperiment1(corp)
	return nil
}

func doExperiment1(corp *ly.Corpus) {
	// a := ly.SortedMap(corp.Unigrams)
	// for _, v := range a[:10] {
	// 	fmt.Println(v.Key.String(), " ", v.Count)
	// }
	fmt.Println(corp)
	// var fav *ly.KeyInfo

	// keys := [42]ly.KeyInfo{}
	// for i := range keys {
	// 	keys[i].Column = I2col(i)
	// 	keys[i].Row = I2row(i)
	// 	if i == 17 {
	// 		fav = &keys[i]
	// 	}
	// }
	// for _, k := range keys {
	// 	fmt.Println(k)
	// }
	// fmt.Println(*fav)
	// fmt.Println(fav)
	// fmt.Println(fav.Row)

	// for i := range 42 {
	// 	ik := ly.NewKeyInfo(uint8(i/12), uint8(i%12))
	// 	for j := range 42 {
	// 		if i == j {
	// 			continue
	// 		}
	// 		jk := ly.NewKeyInfo(uint8(j/12), uint8(j%12))
	// 		if ik.Finger == jk.Finger {
	// 			// fmt.Printf("SFB: %d,%d: %v %v\n", i, j, ik, jk)
	// 		}
	// 	}
	// }

}

// // Standard keyboard offsets
// var rowStagOffsets = [4]float64{
// 	0, 0.25, 0.75, 0,
// }

// // Corne-style offsets
// var colStagOffsets = [12]float64{
// 	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
// }

// // Get the x "coordinate", which is just the column for ortho and colstag
// func getAdjustedColumn(row uint8, column uint8) float64 {
// 	return float64(column)
// }

// // Get the y "coordinate", which is adjusted for col-staggered keyboards
// func getAdjustedRowStaggered(row uint8, column uint8) float64 {
// 	return float64(row) + colStagOffsets[column]
// }

// // Get the y "coordinate", which is just the row for ortho and rowstag
// func getAdjustedRow(row uint8, column uint8) float64 {
// 	return float64(row)
// }

// type KeyPairDistance struct {
// 	rowDist  float64
// 	colDist  float64
// 	distance float64
// }

// func doExperiment(corp *ly.Corpus) {
// 	ltype := ly.ORTHO

// 	// How to calculate adjusted row and col
// 	var getAdjRowDist, getAdjColDist func(uint8, uint8, uint8, uint8) float64
// 	switch ltype {
// 	case ly.ROWSTAG:
// 		getAdjRowDist, getAdjColDist = AbsRowDist, AbsColDistAdj
// 	case ly.COLSTAG:
// 		getAdjRowDist, getAdjColDist = AbsRowDistAdj, AbsColDist
// 	default:
// 		getAdjRowDist, getAdjColDist = AbsRowDist, AbsColDist
// 	}

// 	var k1, k2 uint8
// 	for k1 = range 42 {
// 		row1, col1 := k1/12, k1%12
// 		for k2 = range 42 {
// 			if k1 == k2 {
// 				continue
// 			}
// 			row2, col2 := k2/12, k2%12

// 			// skip if we on different hands
// 			if ((row1 < 3 && col1 < 6) || (row1 >= 3 && col1 < 3)) !=
// 				((row2 < 3 && col2 < 6) || (row2 >= 3 && col2 < 3)) {
// 				continue
// 			}

// 			// calculate distances
// 			dx := getAdjRowDist(row1, col1, row2, col2)
// 			dy := getAdjColDist(row1, col1, row2, col2)
// 			dist := math.Sqrt(dx*dx + dy*dy)
// 			pair := KeyPairDistance{
// 				rowDist:  dy,
// 				colDist:  dx,
// 				distance: dist,
// 			}
// 			fmt.Println(pair)
// 		}
// 	}
// }

// func AbsRowDist(row1, col1, row2, col2 uint8) float64 {
// 	return math.Abs(float64(row1) - float64(row2))
// }

// func AbsRowDistAdj(row1, col1, row2, col2 uint8) float64 {
// 	return math.Abs((float64(row1) + colStagOffsets[col1] -
// 		(float64(row2) + colStagOffsets[col2])))
// }

// func AbsColDist(row1, col1, row2, col2 uint8) float64 {
// 	return math.Abs(float64(col1) - float64(col2))
// }

// func AbsColDistAdj(row1, col1, row2, col2 uint8) float64 {
// 	return math.Abs((float64(col1) + rowStagOffsets[row1] -
// 		(float64(col2) + rowStagOffsets[row2])))
// }

func I2col[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i % 12)
}

func I2row[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i / 12)
}
