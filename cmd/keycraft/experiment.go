package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
	"github.com/urfave/cli/v2"
)

var experimentCommand = &cli.Command{
	Name:      "experiment",
	Aliases:   []string{"x"},
	Usage:     "Run experiments (for the developer)",
	ArgsUsage: "<layout.klf>",
	Flags:     flagsSlice("corpus", "weights-file", "weights", "free", "generations", "accept-worse"),
	Action:    DoExperiment3,
}

type KeyInfo struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Finger string `json:"finger"`
}

type Layout struct {
	Name   string `json:"name"`
	User   any    `json:"user"`
	Author string
	Board  string             `json:"board"`
	Keys   map[string]KeyInfo `json:"keys"`
	Likes  uint
}

func DoExperiment3(c *cli.Context) (err error) {
	// Load authors.json
	authorsFile := "./authors.json"
	authorsData, err := os.ReadFile(authorsFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", authorsFile, err)
	}
	var authorsMap map[string]any
	if err := json.Unmarshal(authorsData, &authorsMap); err != nil {
		return fmt.Errorf("error parsing %s: %v", authorsFile, err)
	}
	// Reverse lookup: from int64 to string
	reverseAuthors := make(map[any]string)
	for name, id := range authorsMap {
		reverseAuthors[id] = name
	}

	// Load likes.json
	likesFile := "./likes.json"
	likesData, err := os.ReadFile(likesFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", likesFile, err)
	}
	var likeEntries map[string][]int64
	if err := json.Unmarshal(likesData, &likeEntries); err != nil {
		return fmt.Errorf("error parsing %s: %v", likesFile, err)
	}
	likesMap := make(map[string]uint)
	for name, numbers := range likeEntries {
		likesMap[name] = uint(len(numbers))
	}

	// load layouts
	dir := "./cmini"

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	layouts := make([]Layout, 0, 3000)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(dir, f.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("error reading %s: %v\n", f.Name(), err)
			continue
		}

		// skip files with "keys": {}
		if strings.Contains(string(data), `"keys": {}`) {
			continue
		}

		var layout Layout
		if err := json.Unmarshal(data, &layout); err != nil {
			fmt.Printf("error parsing %s: %v\n", f.Name(), err)
			continue
		}

		// Check if all letters a-z are present in layout.Keys
		if len(layout.Keys) < 26 {
			continue
		}
		missingLetter := false
		for r := 'a'; r <= 'z'; r++ {
			if _, ok := layout.Keys[string(r)]; !ok {
				missingLetter = true
				break
			}
		}
		if missingLetter {
			continue
		}

		// Assign likes from likesMap
		layout.Likes = likesMap[layout.Name]

		// Assign author from reverseAuthors map if possible
		if authorName, ok := reverseAuthors[layout.User]; ok {
			layout.Author = authorName
		}

		// Skip layouts with 0 likes
		if layout.Likes == 0 {
			continue
		}
		layouts = append(layouts, layout)
	}

	// godump.Dump(layouts[0].Keys)

	// Sort layouts by Likes descending
	slices.SortFunc(layouts, func(a, b Layout) int {
		if a.Likes > b.Likes {
			return -1
		}
		if a.Likes < b.Likes {
			return 1
		}
		return 0
	})

	for _, l := range layouts {
		fmt.Printf("Name: %s, Author: %s, Board: %s, Keys: %d, Likes: %d\n",
			l.Name, l.Author, l.Board, len(l.Keys), l.Likes)
	}

	// Count how many layouts each author has authored
	authorCounts := make(map[string]int)
	for _, layout := range layouts {
		authorCounts[layout.Author]++
	}

	// Print authors and counts, sorted by count descending
	fmt.Println("\nAuthors:")
	type authorCount struct {
		Author string
		Count  int
	}
	var authorCountSlice []authorCount
	for author, count := range authorCounts {
		authorCountSlice = append(authorCountSlice, authorCount{Author: author, Count: count})
	}
	slices.SortFunc(authorCountSlice, func(a, b authorCount) int {
		if a.Count > b.Count {
			return -1
		}
		if a.Count < b.Count {
			return 1
		}
		return 0
	})
	for _, ac := range authorCountSlice {
		fmt.Printf("%s (%d layouts)\n", ac.Author, ac.Count)
	}

	return
}

func ExperimentAction(c *cli.Context) error {
	_, err := loadCorpus(c.String("corpus"))
	layout, err2 := loadLayout(c.Args().First())
	if err != nil || err2 != nil {
		return fmt.Errorf("sorry / %v / %v", err, err2)
	}

	tw := table.NewWriter()
	tw.SetAutoIndex(true)
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "eucl.dist", Transformer: Fraction},
	})
	tw.AppendHeader(table.Row{"Idx1", "Idx2", "bi", "coldist", "rowdist", "fingerdist", "eucl.dist"})

	for i := range 42 {
		r := layout.Runes[i]
		ki := layout.RuneInfo[r]
		for j := range 42 {
			r2 := layout.Runes[j]
			ki2 := layout.RuneInfo[r2]
			if (ki.Finger == kc.LM && ki2.Finger == kc.LI) ||
				(ki.Finger == kc.RM && ki2.Finger == kc.RI) {
				bi := string(r) + string(r2)
				dist := layout.KeyPairDistances[kc.KeyPair{ki.Index, ki2.Index}]
				if dist.Distance >= 2.0 {
					tw.AppendRow(table.Row{ki.Index, ki2.Index, bi, dist.ColDist, dist.RowDist, dist.FingerDist, dist.Distance})
				}
			}
		}
	}

	// fmt.Println(tw.Render())

	fmt.Println(layout)
	for _, lsb := range layout.LSBs {
		bi := string(layout.Runes[lsb.KeyIdx1]) + string(layout.Runes[lsb.KeyIdx2])
		fmt.Println(bi, " ", lsb)
	}

	return nil
}

func ExperimentAction2(c *cli.Context) error {
	fmt.Println("Running experiment...")

	// Load the corpus used for analysing layouts.
	corpus, err := loadCorpus(c.String("corpus"))
	if err != nil {
		return err
	}

	weightsPath := c.String("weights-file")
	if weightsPath != "" {
		weightsPath = filepath.Join(weightsDir, weightsPath)
	}
	weights, err := kc.NewWeightsFromParams(weightsPath, c.String("weights"))
	if err != nil {
		return err
	}

	acceptFunction := c.String("accept-worse")
	if !slices.Contains(validAcceptFuncs, acceptFunction) {
		return fmt.Errorf("invalid accept function: %s. Must be one of: %v", acceptFunction, validAcceptFuncs)
	}

	numGenerations := c.Uint("generations")
	if numGenerations <= 0 {
		return fmt.Errorf("number of generations must be above 0. Got: %d", numGenerations)
	}

	if c.Args().Len() != 1 {
		if err := compareAllQwertyLayouts(); err != nil {
			return err
		}

		return fmt.Errorf("expected exactly 1 layout file, got %d", c.Args().Len())
	}
	layoutFile := c.Args().First()
	layout, err := loadLayout(layoutFile)
	if err != nil {
		return err
	}

	fixed := "etfjio"
	variable := "dknl;uyrs"
	variations := generateVariations(variable, 5)
	for _, v := range variations {
		v = fixed + v
		fmt.Println(v)
		if err := layout.LoadPinsFromParams("", "", v); err != nil {
			return err
		}

		best := layout.Optimise(corpus, weights, numGenerations, acceptFunction)

		// Save best layout to file
		name := filepath.Base(layout.Name)
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".klf" {
			name = name[:len(name)-len(ext)]
		}
		bestFilename := fmt.Sprintf("%s-%s.klf", name, v)
		bestPath := filepath.Join(layoutDir, bestFilename)
		if err := best.SaveToFile(bestPath); err != nil {
			return fmt.Errorf("failed to save best layout to %s: %v", bestPath, err)
		}
	}

	return nil
}

func generateVariations(s string, k int) []string {
	var results []string
	n := len(s)

	var comb func(start, depth int, indices []int)
	comb = func(start, depth int, indices []int) {
		if depth == k {
			// create a map for quick lookup of indices to remove
			remove := make(map[int]bool)
			for _, idx := range indices {
				remove[idx] = true
			}
			var b []byte
			for i := range n {
				if !remove[i] {
					b = append(b, s[i])
				}
			}
			results = append(results, string(b))
			return
		}
		for i := start; i < n; i++ {
			comb(i+1, depth+1, append(indices, i))
		}
	}
	comb(0, 0, []int{})
	return results
}

// countRuneDifferences compares two layouts by their Runes slices
// and returns the number of positions where the rune differs.
func countRuneDifferences(base, other *kc.SplitLayout) int {
	if len(base.Runes) != len(other.Runes) {
		// not directly comparable
		return -1
	}
	changed := 0
	for i := range base.Runes {
		if base.Runes[i] != other.Runes[i] {
			changed++
		}
	}
	return changed
}

func compareAllQwertyLayouts() error {
	baseLayoutPath := filepath.Join(layoutDir, "qwerty.klf")
	base, err := loadLayout("qwerty.klf")
	if err != nil {
		return fmt.Errorf("failed to load base layout %s: %v", baseLayoutPath, err)
	}

	pattern := filepath.Join(layoutDir, "qwerty*.klf")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob pattern %s: %v", pattern, err)
	}

	for _, file := range files {
		if filepath.Base(file) == "qwerty.klf" {
			continue
		}
		other, err := loadLayout(filepath.Base(file))
		if err != nil {
			return fmt.Errorf("failed to load layout %s: %v", filepath.Base(file), err)
		}
		diff := countRuneDifferences(base, other)
		fmt.Printf("%s: %d differing runes\n", filepath.Base(file), diff)
	}
	return nil
}

func DoExperiment2(corp *kc.Corpus, lay *kc.SplitLayout) {
	stats := make(map[string]uint64)

	for tri, cnt := range corp.Trigrams {
		// Cross-hand trigrams
		add2Roll := func(h, fA, fB uint8) {
			switch {
			case fA == fB:
				stats["2RL-SFB"] += cnt
			case (fA < fB) == (h == kc.LEFT):
				stats["2RL-I"] += cnt
			default:
				stats["2RL-O"] += cnt
			}
		}

		r0, ok0 := lay.RuneInfo[tri[0]]
		r1, ok1 := lay.RuneInfo[tri[1]]
		r2, ok2 := lay.RuneInfo[tri[2]]
		if !ok0 || !ok1 || !ok2 {
			stats["TRI-SKP"] += cnt
			continue
		}

		h0, h1, h2 := r0.Hand, r1.Hand, r2.Hand
		f0, f1, f2 := r0.Finger, r1.Finger, r2.Finger
		diffIdx02 := r0.Index != r2.Index

		if h0 == h2 {
			if h0 != h1 {
				// ALT or ALT-SFS
				if f0 == f2 && diffIdx02 {
					stats["ALT-SFS"] += cnt
				} else {
					stats["ALT"] += cnt
				}
			} else {
				// One-hand trigrams
				switch {
				case f0 == f1 || f1 == f2:
					stats["3RL-SFS"] += cnt
				case (f0 < f1) == (f1 < f2):
					if (f0 < f1) == (h0 == kc.LEFT) {
						stats["3RL-I"] += cnt
					} else {
						stats["3RL-O"] += cnt
					}
				default:
					if f0 != 3 && f0 != 6 &&
						f1 != 3 && f1 != 6 &&
						f2 != 3 && f2 != 6 {
						stats["RED-WEAK"] += cnt
					} else if f0 == f2 && diffIdx02 {
						stats["RED-SFS"] += cnt
					} else {
						stats["RED"] += cnt
					}
				}
			}
		} else if h0 == h1 {
			add2Roll(h0, f0, f1)
		} else { // h1 == h2
			add2Roll(h1, f1, f2)
		}
	}

	// Print stats
	tot := 0.0
	for k, v := range stats {
		tot += float64(v)
		fmt.Println(k, float64(v)/float64(corp.TotalTrigramsCount)*100)
	}
	fmt.Println(float64(tot) / float64(corp.TotalTrigramsCount) * 100)
}

func DoExperiment1(corp *kc.Corpus) {
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
