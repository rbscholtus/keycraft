package keycraft

import (
	"fmt"
	"maps"
	"math"
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

// getAcceptFunc returns an acceptance function for simulated annealing based on the chosen policy.
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
	case "linear":
		return func(g, ng uint, e0, e1 float64) float64 {
			t := 1.0 - float64(g)/float64(ng)
			return t
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
func (sl *SplitLayout) Optimise(corp *Corpus, weights *Weights, generations uint, acceptWorse string) *SplitLayout {
	sl.optCorpus = corp
	sl.optWeights = weights
	analysers := Must(LoadAnalysers("data/layouts/", corp))
	sl.optMedians, sl.optIqrs = computeMediansAndIQR(analysers)

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
		fmt.Printf("Best fitness at generation %d: %.3f\n", ga.Generations, fit)
		fmt.Println(ga.HallOfFame[0])
		minFit = fit
	}

	// Run the simulated-annealing algorithm.
	ga := Must(cfg.NewGA())

	newGenome := func(rng *rand.Rand) eaopt.Genome {
		return sl
	}
	Must0(ga.Minimize(newGenome))

	// Return the best encountered solution.
	hof0 := ga.HallOfFame[0]
	best := hof0.Genome.(*SplitLayout)

	return best
}

// Evaluate evaluates the fitness of the current layout.
func (sl *SplitLayout) Evaluate() (float64, error) {
	analyser := NewAnalyser(sl, sl.optCorpus)
	// return 5*sl.SimpleSfbs(sl.optCorpus) + sl.SimpleLsbs(sl.optCorpus), nil

	score := 0.0
	for metric, value := range analyser.Metrics {
		if sl.optIqrs[metric] == 0 {
			continue
		}
		weight := sl.optWeights.Get(metric)
		if weight == 0 {
			continue
		}
		scaledValue := (value - sl.optMedians[metric]) / sl.optIqrs[metric]
		score += weight * scaledValue
	}
	return -score, nil
}

// Mutate randomly swaps two keys in the layout.
func (sl *SplitLayout) Mutate(rng *rand.Rand) {
	// Create a slice to store pairs of indexes and keys that are not pinned.
	pairs := make([]Pair[int, rune], 0, len(sl.Runes))
	for i, r := range sl.Runes {
		if r != 0 && !sl.optPinned[i] {
			// Add the pair to the slice and increment the count.
			pairs = append(pairs, Pair[int, rune]{Key: i, Value: r})
		}
	}

	// Check if there are at least two pairs to swap.
	if len(pairs) < 2 {
		panic(fmt.Sprintf("Not enough unpinned keys on this layout to make a swap: %d", len(pairs)))
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
func (sl *SplitLayout) Crossover(_ eaopt.Genome, _ *rand.Rand) {}

// Clone returns a copy of the layout.
func (sl *SplitLayout) Clone() eaopt.Genome {
	cc := SplitLayout{
		Name:             sl.Name,
		LayoutType:       sl.LayoutType,
		Runes:            sl.Runes,
		RuneInfo:         make(map[rune]KeyInfo),
		GetRowDist:       sl.GetRowDist,
		GetColDist:       sl.GetColDist,
		KeyPairDistances: sl.KeyPairDistances,
		LSBs:             sl.LSBs,
		Scissors:         sl.Scissors,
		optPinned:        sl.optPinned,
		optCorpus:        sl.optCorpus,
		optWeights:       sl.optWeights,
		optMedians:       sl.optMedians,
		optIqrs:          sl.optIqrs,
	}

	maps.Copy(cc.RuneInfo, sl.RuneInfo)

	return &cc
}
