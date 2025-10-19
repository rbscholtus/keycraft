package keycraft

import (
	"fmt"
	"maps"
	"math"
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

// getAcceptFunc returns an acceptance function for simulated annealing.
// The function determines the probability of accepting a worse solution based on the policy:
//   - "always": always accept worse solutions (pure random search)
//   - "never": never accept worse solutions (hill climbing)
//   - "drop-slow": cosine-based decay (smooth, gradual cooling)
//   - "linear": linear decay
//   - "drop-fast": exponential decay (rapid cooling)
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

// Optimise optimizes a keyboard layout using simulated annealing to minimize the weighted score.
// The optimization swaps unpinned keys iteratively, evaluating each candidate using the Evaluate method.
// Returns a new optimized layout (the original is not modified).
func (sl *SplitLayout) Optimise(corp *Corpus, idealRowLoad *[3]float64, idealfgrLoad *[10]float64, weights *Weights, generations uint, acceptWorse string) *SplitLayout {
	// Store optimization parameters
	sl.optCorpus = corp
	sl.optWeights = weights
	sl.optIdealRowLoad = idealRowLoad
	sl.optIdealfgrLoad = idealfgrLoad

	// Load reference layouts to compute normalization statistics
	analysers := Must(LoadAnalysers("data/layouts/", corp, idealRowLoad, idealfgrLoad))
	sl.optMedians, sl.optIqrs = computeMediansAndIQR(analysers)

	// Configure simulated annealing
	cfg := eaopt.NewDefaultGAConfig()
	cfg.NGenerations = generations
	cfg.Model = eaopt.ModSimulatedAnnealing{
		Accept: getAcceptFunc(acceptWorse),
	}

	// Track and display progress
	minFit := math.MaxFloat64
	cfg.Callback = func(ga *eaopt.GA) {
		hof0 := ga.HallOfFame[0]
		fit := hof0.Fitness
		if fit == minFit {
			return // Only output when fitness improves
		}
		fmt.Printf("Best fitness at generation %d: %.3f\n", ga.Generations, fit)
		fmt.Println(ga.HallOfFame[0])
		minFit = fit
	}

	// Run optimization
	ga := Must(cfg.NewGA())

	newGenome := func(rng *rand.Rand) eaopt.Genome {
		return sl // Start from the provided layout
	}
	Must0(ga.Minimize(newGenome))

	// Extract best solution
	hof0 := ga.HallOfFame[0]
	best := hof0.Genome.(*SplitLayout)
	best.Name = best.Name + "-opt"

	return best
}

// Evaluate computes the fitness of the layout (lower is better).
// Uses the stored corpus, weights, and normalization stats to compute a weighted score.
// Returns the negative score so that minimization finds better layouts.
func (sl *SplitLayout) Evaluate() (float64, error) {
	analyser := NewAnalyser(sl, sl.optCorpus, sl.optIdealRowLoad, sl.optIdealfgrLoad)

	score := 0.0
	for metric, value := range analyser.Metrics {
		// Skip metrics with no variation
		if sl.optIqrs[metric] == 0 {
			continue
		}
		weight := sl.optWeights.Get(metric)
		if weight == 0 {
			continue
		}
		// Apply robust normalization and weighting
		scaledValue := (value - sl.optMedians[metric]) / sl.optIqrs[metric]
		score += weight * scaledValue
	}
	return -score, nil // Negate for minimization
}

// Mutate randomly swaps two unpinned keys in the layout.
// This is the mutation operator for the genetic algorithm.
func (sl *SplitLayout) Mutate(rng *rand.Rand) {
	// Collect all unpinned, non-empty keys
	pairs := make([]Pair[int, rune], 0, len(sl.Runes))
	for i, r := range sl.Runes {
		if r != 0 && !sl.optPinned[i] {
			pairs = append(pairs, Pair[int, rune]{Key: i, Value: r})
		}
	}

	if len(pairs) < 2 {
		panic(fmt.Sprintf("Not enough unpinned keys on this layout to make a swap: %d", len(pairs)))
	}

	// Select two distinct random keys
	i := rng.Intn(len(pairs))
	j := rng.Intn(len(pairs))
	for j == i {
		j = rng.Intn(len(pairs))
	}

	// Perform the swap in both Runes array and RuneInfo map
	sl.Runes[pairs[i].Key], sl.Runes[pairs[j].Key] = sl.Runes[pairs[j].Key], sl.Runes[pairs[i].Key]
	sl.RuneInfo[pairs[i].Value], sl.RuneInfo[pairs[j].Value] = sl.RuneInfo[pairs[j].Value], sl.RuneInfo[pairs[i].Value]
}

// Crossover is not used in simulated annealing but required by the eaopt.Genome interface.
func (sl *SplitLayout) Crossover(_ eaopt.Genome, _ *rand.Rand) {}

// Clone creates a deep copy of the layout for the genetic algorithm.
func (sl *SplitLayout) Clone() eaopt.Genome {
	cc := SplitLayout{
		Name:             sl.Name,
		LayoutType:       sl.LayoutType,
		Runes:            sl.Runes,
		RuneInfo:         make(map[rune]KeyInfo),
		KeyPairDistances: sl.KeyPairDistances,
		LSBs:             sl.LSBs,
		FScissors:        sl.FScissors,
		HScissors:        sl.HScissors,
		optPinned:        sl.optPinned,
		optCorpus:        sl.optCorpus,
		optIdealfgrLoad:  sl.optIdealfgrLoad,
		optWeights:       sl.optWeights,
		optMedians:       sl.optMedians,
		optIqrs:          sl.optIqrs,
	}

	maps.Copy(cc.RuneInfo, sl.RuneInfo)

	return &cc
}
