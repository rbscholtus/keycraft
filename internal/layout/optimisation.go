// Package layout provides functionality for optimizing keyboard layouts.
package layout

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"sort"

	"github.com/MaxHalford/eaopt"
	corpus "github.com/rbscholtus/kb/internal/corpus"
)

// Optimise optimizes a keyboard layout using simulated annealing.
func (sl *SplitLayout) Optimise(corp *corpus.Corpus, acceptWorse string, generations int) *SplitLayout {
	sl.optCorpus = corp

	// Configure the simulated annealing algorithm.
	cfg := eaopt.NewDefaultGAConfig()
	cfg.Model = eaopt.ModSimulatedAnnealing{
		// Define the acceptance function for simulated annealing.
		Accept: func(g, ng uint, e0, e1 float64) float64 {
			switch acceptWorse {
			case "always":
				return 1.0
			case "never":
				return 0.0
			case "drop-slow":
				t := 1.0 - float64(g)/float64(ng)
				return (math.Cos(t*math.Pi) + 1.0) / 2.0
			case "temp":
				t := 1.0 - float64(g)/float64(ng)
				return t
			case "cold":
				t := 1.0 - float64(g)/float64(ng)
				return 0.5 * t
			case "drop-fast":
				t := 1.0 - float64(g)/float64(ng)
				return math.Exp(-3.0 * (1 - t))
			default:
				panic("unknown accept worse function")
			}
		},
	}
	cfg.NGenerations = uint(generations)

	// Add a custom callback function to track progress.
	minFit := math.MaxFloat64
	cfg.Callback = func(ga *eaopt.GA) {
		hof0 := ga.HallOfFame[0]
		fit := hof0.Fitness
		if fit == minFit {
			// Output only when we make an improvement.
			return
		}
		// best := hof0.Genome.(*SplitLayout)
		fmt.Printf("Best fitness at generation %3d: %.3f%%\n", ga.Generations, 100*fit)
		minFit = fit
	}

	// Run the simulated-annealing algorithm.
	ga, err := cfg.NewGA()
	if err != nil {
		panic(err)
	}
	err = ga.Minimize(func(rng *rand.Rand) eaopt.Genome {
		return sl
	})
	if err != nil {
		panic(err)
	}

	// Return the best encountered solution.
	hof0 := ga.HallOfFame[0]
	best := hof0.Genome.(*SplitLayout)

	return best
}

// Evaluate evaluates the fitness of the current layout.
func (sl *SplitLayout) Evaluate() (float64, error) {
	return sl.SimpleSfbs(sl.optCorpus), nil
}

// randomBigram selects a random bigram based on the count of SFBs.
func randomBigram(rng *rand.Rand, sfbs []Sfb) corpus.Bigram {
	// Calculate the total count of bigrams.
	var total uint64
	for _, sfb := range sfbs {
		total += sfb.Count
	}

	// Generate a random number between 0 and the total.
	randNum := rng.Int63n(int64(total))

	// Select the bigram based on the random number.
	var cumulative int64
	for _, sfb := range sfbs {
		cumulative += int64(sfb.Count)
		if randNum <= cumulative {
			return sfb.Bigram
		}
	}

	// If the random number exceeds the total percentage, return the last bigram.
	return sfbs[len(sfbs)-1].Bigram
}

// Mutate mutates the layout by swapping two keys.
func (sl *SplitLayout) Mutate(rng *rand.Rand) {
	// Get a list of keys from the RuneInfo map.
	keys := make([]rune, 0, len(sl.RuneInfo))
	for k := range sl.RuneInfo {
		keys = append(keys, k)
	}

	// Get a list of SFBs and sort them by count.
	sfbs, _ := sl.extractSfbs(sl.optCorpus)
	sort.Slice(sfbs, func(i, j int) bool {
		return sfbs[i].Count > sfbs[j].Count
	})

	// Select a random bigram from the SFBs.
	bi := randomBigram(rng, sfbs)
	key1 := bi[rng.Intn(2)]
	for sl.RuneInfo[key1].Row == 1 { // Pin row 1.
		bi = randomBigram(rng, sfbs)
		key1 = bi[rng.Intn(2)]
	}

	// Select a random key to swap with.
	j := rng.Intn(len(keys))
	key2 := keys[j]
	for key1 == key2 || sl.RuneInfo[key2].Row == 1 { // No swap with self or pin row 1.
		j = rng.Intn(len(keys))
		key2 = keys[j]
	}

	// Calculate indices.
	index1 := sl.getIndex(key1)
	index2 := sl.getIndex(key2)

	// Swap the values associated with the two keys.
	sl.Runes[index1], sl.Runes[index2] = sl.Runes[index2], sl.Runes[index1]
	sl.RuneInfo[key1], sl.RuneInfo[key2] = sl.RuneInfo[key2], sl.RuneInfo[key1]
}

// getIndex calculates the index of a key in the Runes array.
func (sl *SplitLayout) getIndex(key rune) int {
	info := sl.RuneInfo[key]
	return int(info.Row*12 + info.Column)
}

// Crossover does nothing. It is defined only so *SplitLayout implements the eaopt.Genome interface.
func (sl *SplitLayout) Crossover(other eaopt.Genome, rng *rand.Rand) {}

// Clone returns a copy of the layout.
func (sl *SplitLayout) Clone() eaopt.Genome {
	cc := SplitLayout{
		Filename:  sl.Filename,
		Runes:     sl.Runes,
		RuneInfo:  make(map[rune]KeyInfo),
		optCorpus: sl.optCorpus,
	}

	maps.Copy(cc.RuneInfo, sl.RuneInfo)

	return &cc
}
