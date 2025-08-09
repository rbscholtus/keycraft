package main

import (
	"fmt"

	ly "github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var experimentCommand = &cli.Command{
	Name:   "experiment",
	Usage:  "Run experiments",
	Action: experimentAction,
}

func experimentAction(c *cli.Context) error {
	fmt.Println("Running experiment...")

	corp, err := loadCorpus(c)
	if err != nil {
		return err
	}

	lay, err := loadLayout(c.Args().First())
	if err != nil {
		return err
	}

	// style := c.String("style")

	doExperiment2(corp, lay)
	return nil
}

func doExperiment2(corp *ly.Corpus, lay *ly.SplitLayout) {
	stats := make(map[string]uint64)
	for tri, cnt := range corp.Trigrams {
		r0, ok0 := lay.RuneInfo[tri[0]]
		r1, ok1 := lay.RuneInfo[tri[1]]
		r2, ok2 := lay.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			stats["SKP"] += cnt
			continue
		}
		if r0.Hand == r2.Hand {
			if r0.Hand != r1.Hand {
				if r0.Finger == r2.Finger && r0.Index != r2.Index {
					stats["ALT-SFS"] += cnt
				} else {
					stats["ALT"] += cnt
				}
			} else {
				// One hand trigrams here
				if r0.Finger == r1.Finger || r1.Finger == r2.Finger {
					// Same finger in a row
					stats["OTH"] += cnt
				} else if (r0.Finger < r1.Finger) == (r1.Finger < r2.Finger) {
					// 3-roll in or out
					if (r0.Finger < r1.Finger) == (r0.Hand == ly.LEFT) {
						stats["3RL-I"] += cnt
					} else {
						stats["3RL-O"] += cnt
					}
				} else {
					// redirects
					if r0.Finger != 3 && r0.Finger != 6 && r1.Finger != 3 && r1.Finger != 6 && r2.Finger != 3 && r2.Finger != 6 {
						stats["RED-BAD"] += cnt
					} else if r0.Finger == r2.Finger && r0.Index != r2.Index {
						stats["RED-SFS"] += cnt
					} else {
						stats["RED"] += cnt
					}
				}
			}

			continue
		}

		// At this point r0.Hand != r2.Hand

		// Helper to add 2-roll stats
		add2Roll := func(h uint8, f0, f1 uint8) {
			if f0 == f1 {
				stats["OTH"] += cnt
			} else if (f0 < f1) == (h == ly.LEFT) {
				stats["2RL-I"] += cnt
			} else {
				stats["2RL-O"] += cnt
			}
		}

		// Check pairs r0,r1 or r1,r2 for same hand
		if r0.Hand == r1.Hand {
			add2Roll(r0.Hand, r0.Finger, r1.Finger)
		} else {
			add2Roll(r1.Hand, r1.Finger, r2.Finger)
		}
	}

	tot := 0.0
	for k, v := range stats {
		tot += float64(v)
		fmt.Println(k, float64(v)/float64(corp.TotalTrigramsCount)*100)
	}
	fmt.Println(float64(tot) / float64(corp.TotalTrigramsCount) * 100)
	// godump.Dump(stats)
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
