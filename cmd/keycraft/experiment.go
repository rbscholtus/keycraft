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

// experimentCommand defines the "experiment" CLI command.
// This command is intended for developer use to run various experiments.
var experimentCommand = &cli.Command{
	Name:      "experiment",
	Aliases:   []string{"x"},
	Usage:     "Run experiments (for developer use)",
	ArgsUsage: "<layout.klf>",
	Flags:     flagsSlice("corpus", "finger-load", "weights-file", "weights", "free", "generations", "accept-worse"),
	Action:    DoExperiment3,
}

// KeyInfo represents the position and finger assignment of a key in a layout.
type KeyInfo struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Finger string `json:"finger"`
}

// Layout represents a keyboard layout with metadata and key mappings.
type Layout struct {
	Name   string `json:"name"`
	User   any    `json:"user"`
	Author string
	Board  string             `json:"board"`
	Keys   map[string]KeyInfo `json:"keys"`
	Likes  uint
}

// DoExperiment3 processes and analyzes layout data from external sources.
// It loads author and like data, filters layouts, and prints statistics.
func DoExperiment3(c *cli.Context) (err error) {
	authorsFile := "./authors.json"
	authorsData, err := os.ReadFile(authorsFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", authorsFile, err)
	}
	var authorsMap map[string]any
	if err := json.Unmarshal(authorsData, &authorsMap); err != nil {
		return fmt.Errorf("error parsing %s: %v", authorsFile, err)
	}
	reverseAuthors := make(map[any]string)
	for name, id := range authorsMap {
		reverseAuthors[id] = name
	}

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

		layout.Likes = likesMap[layout.Name]

		if authorName, ok := reverseAuthors[layout.User]; ok {
			layout.Author = authorName
		}

		if layout.Likes == 0 {
			continue
		}
		layouts = append(layouts, layout)
	}

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

	authorCounts := make(map[string]int)
	for _, layout := range layouts {
		authorCounts[layout.Author]++
	}

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

// ExperimentAction analyzes distance metrics between key pairs on a layout.
func ExperimentAction(c *cli.Context) error {
	_, err := loadCorpus(c.String("corpus"), false, 98)
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
				dist := layout.Distance(ki.Index, ki2.Index)
				if dist.Distance >= 2.0 {
					tw.AppendRow(table.Row{ki.Index, ki2.Index, bi, dist.ColDist, dist.RowDist, dist.FingerDist, dist.Distance})
				}
			}
		}
	}

	fmt.Println(layout)
	for _, lsb := range layout.LSBs {
		bi := string(layout.Runes[lsb.KeyIdx1]) + string(layout.Runes[lsb.KeyIdx2])
		fmt.Println(bi, " ", lsb)
	}

	return nil
}

// ExperimentAction2 runs layout optimization experiments with pin variations.
func ExperimentAction2(c *cli.Context) error {
	fmt.Println("Running experiment...")

	corpus, err := loadCorpus(c.String("corpus"), false, 98)
	if err != nil {
		return err
	}

	fbStr := c.String("finger-load")
	fingerBal, err := parseFingerLoad(fbStr)
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

		best := layout.Optimise(corpus, fingerBal, weights, numGenerations, acceptFunction)

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

// generateVariations generates all combinations created by removing k characters from s.
func generateVariations(s string, k int) []string {
	var results []string
	n := len(s)

	var comb func(start, depth int, indices []int)
	comb = func(start, depth int, indices []int) {
		if depth == k {
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

// countRuneDifferences returns the number of positions where two layouts differ.
// Returns -1 if the layouts have different lengths.
func countRuneDifferences(base, other *kc.SplitLayout) int {
	if len(base.Runes) != len(other.Runes) {
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

// compareAllQwertyLayouts compares qwerty.klf against all qwerty*.klf variants.
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

// DoExperiment2 analyzes and categorizes trigram statistics for a layout.
func DoExperiment2(corp *kc.Corpus, lay *kc.SplitLayout) {
	stats := make(map[string]uint64)

	for tri, cnt := range corp.Trigrams {
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

		switch h0 {
		case h2:
			if h0 != h1 {
				if f0 == f2 && diffIdx02 {
					stats["ALT-SFS"] += cnt
				} else {
					stats["ALT"] += cnt
				}
			} else {
				switch {
				case f0 == f1 || f1 == f2:
					stats["3RL-SFB"] += cnt
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
		case h1:
			add2Roll(h0, f0, f1)
		default:
			add2Roll(h1, f1, f2)
		}
	}

	tot := 0.0
	for k, v := range stats {
		tot += float64(v)
		fmt.Println(k, float64(v)/float64(corp.TotalTrigramsCount)*100)
	}
	fmt.Println(float64(tot) / float64(corp.TotalTrigramsCount) * 100)
}

// DoExperiment1 prints basic corpus information.
func DoExperiment1(corp *kc.Corpus) {
	fmt.Println(corp)
}

// I2col converts a linear key index 'i' to its corresponding column index (0-11).
func I2col[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i % 12)
}

// I2row converts a linear key index 'i' to its corresponding row index (0-3).
func I2row[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](i T) uint8 {
	return uint8(i / 12)
}
