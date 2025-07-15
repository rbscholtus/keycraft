// Package layout provides functionality for optimizing keyboard layouts.
package layout

import (
	"fmt"
	"maps"
	"math"
	"math/rand"

	"github.com/MaxHalford/eaopt"
	corpus "github.com/rbscholtus/kb/internal/corpus"
)

func getAcceptFunc(acceptWorse string) func(g, ng uint, e0, e1 float64) float64 {
	switch acceptWorse {
	case "always":
		return func(g, ng uint, e0, e1 float64) float64 { return 1.0 }
	case "never":
		return func(g, ng uint, e0, e1 float64) float64 { return 0.0 }
	case "drop-slow":
		return func(g, ng uint, e0, e1 float64) float64 {
			t := 1.0 - float64(g)/float64(ng)
			return (math.Cos(t*math.Pi) + 1.0) / 2.0
		}
	case "temp":
		return func(g, ng uint, e0, e1 float64) float64 {
			t := 1.0 - float64(g)/float64(ng)
			return t
		}
	case "cold":
		return func(g, ng uint, e0, e1 float64) float64 {
			t := 1.0 - float64(g)/float64(ng)
			return 0.5 * t
		}
	case "drop-fast":
		return func(g, ng uint, e0, e1 float64) float64 {
			t := 1.0 - float64(g)/float64(ng)
			return math.Exp(-3.0 * (1 - t))
		}
	default:
		panic("unknown accept worse function")
	}
}

// Optimise optimizes a keyboard layout using simulated annealing.
func (sl *SplitLayout) Optimise(corp *corpus.Corpus, generations uint, acceptWorse string) *SplitLayout {
	sl.optCorpus = corp

	// Configure the simulated annealing algorithm.
	cfg := eaopt.NewDefaultGAConfig()
	cfg.NGenerations = generations
	cfg.Model = eaopt.ModSimulatedAnnealing{
		// Determine the acceptance function based on the acceptWorse parameter.
		Accept: getAcceptFunc(acceptWorse),
	}

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

	best.Filename = "best.kb"

	return best
}

// Evaluate evaluates the fitness of the current layout.
func (sl *SplitLayout) Evaluate() (float64, error) {
	return sl.SimpleSfbs(sl.optCorpus), nil
}

// Mutate randomly swaps two keys in the layout.
func (sl *SplitLayout) Mutate(rng *rand.Rand) {
	// Create a slice to store pairs of indexes and keys that are not pinned.
	pairs := make([]Pair[int, rune], 0, len(sl.Runes))
	for i, r := range sl.Runes {
		if r != '~' && !sl.Pinned[i] {
			// Add the pair to the slice and increment the count.
			pairs = append(pairs, Pair[int, rune]{Key: i, Value: r})
		}
	}

	// Check if there are at least two pairs to swap.
	if len(pairs) < 2 {
		panic(fmt.Sprintf("Not enough keys on this layout to make a swap: %d", len(pairs)))
	}

	// Generate two random indexes for the pairs to swap.
	i := rng.Intn(len(pairs))
	j := rng.Intn(len(pairs))
	for j == i {
		j = rng.Intn(len(pairs))
	}

	// Swap the values associated with the two keys.
	sl.Runes[pairs[i].Key], sl.Runes[pairs[j].Key] = sl.Runes[pairs[j].Key], sl.Runes[pairs[i].Key]
	sl.RuneInfo[pairs[i].Value], sl.RuneInfo[pairs[j].Value] = sl.RuneInfo[pairs[j].Value], sl.RuneInfo[pairs[i].Value]
}

// Crossover does nothing. It is defined only so *SplitLayout implements the eaopt.Genome interface.
func (sl *SplitLayout) Crossover(other eaopt.Genome, rng *rand.Rand) {}

// Clone returns a copy of the layout.
func (sl *SplitLayout) Clone() eaopt.Genome {
	cc := SplitLayout{
		Filename:  sl.Filename,
		Runes:     sl.Runes,
		RuneInfo:  make(map[rune]KeyInfo),
		Pinned:    sl.Pinned,
		optCorpus: sl.optCorpus,
	}

	maps.Copy(cc.RuneInfo, sl.RuneInfo)

	return &cc
}
