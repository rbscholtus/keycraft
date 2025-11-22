package keycraft

import (
	"io"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

// BLSParams holds all configuration parameters for the Breakout Local Search algorithm.
type BLSParams struct {
	// Core BLS parameters

	L0      int     // Initial jump magnitude (number of perturbation swaps)
	LMax    int     // Maximum jump magnitude for strong diversification
	T       int     // Stagnation threshold: consecutive non-improving local optima before strong perturbation
	TabuMin int     // Minimum tabu tenure (iterations a swap is forbidden)
	TabuMax int     // Maximum tabu tenure
	P0      float64 // Minimum probability for directed perturbation

	// Perturbation distribution (probabilities for non-directed perturbations, should sum to 1.0)

	PatternWeight float64 // Weight for pattern-guided perturbation (targets bad patterns)
	ColumnWeight  float64 // Weight for column swaps (structural moves)
	RandomWeight  float64 // Weight for random perturbation (strong diversification)
	RecencyWeight float64 // Weight for recency-based perturbation (history-based)

	// Pattern analysis parameters

	TopKProblematic int // Number of worst keys to consider for pattern-guided perturbation

	// Optimization control

	MaxIterations  int           // Maximum number of iterations (local optima visited)
	MaxTime        time.Duration // Maximum wall-clock time for optimization
	Seed           int64         // Random seed for reproducibility
	ReportInterval int           // Report progress every N iterations (0 = no reporting)

	// Parallelism control

	UseParallel     bool // Enable parallel evaluation in steepest descent
	ParallelWorkers int  // Number of parallel workers (0 = use runtime.NumCPU())
}

// DefaultBLSParams returns recommended BLS parameters for keyboard layout optimization.
// Parameters are automatically scaled based on the number of free (non-pinned) keys.
func DefaultBLSParams(numFreeKeys int) BLSParams {
	return BLSParams{
		// Core parameters (scaled to number of free keys)
		L0:      int(0.1 * float64(numFreeKeys)), // was: 0.15
		LMax:    int(0.5 * float64(numFreeKeys)),
		T:       500, // was: 2500
		TabuMin: int(0.9 * float64(numFreeKeys)),
		TabuMax: int(1.1 * float64(numFreeKeys)),
		P0:      0.75,

		// Perturbation distribution
		PatternWeight: 0.25,
		ColumnWeight:  0.15,
		RandomWeight:  0.40,
		RecencyWeight: 0.20,

		// Pattern analysis
		TopKProblematic: 8,

		// Optimization control
		MaxIterations:  2000,
		MaxTime:        15 * time.Minute,
		Seed:           time.Now().UnixNano(),
		ReportInterval: 100,

		// Parallelism control
		UseParallel:     true, // Disabled by default
		ParallelWorkers: 4,    // Use runtime.NumCPU() when enabled
	}
}

// PerturbationType identifies the type of perturbation to apply.
type PerturbationType int

const (
	DirectedPerturb      PerturbationType = iota // Quality-guided swap selection (tabu search style)
	PatternGuidedPerturb                         // Targets keys involved in problematic patterns
	ColumnPerturb                                // Swaps entire columns
	RecencyPerturb                               // Selects least recently used swaps
	RandomPerturb                                // Completely random swap selection
)

// String returns the name of the perturbation type.
func (pt PerturbationType) String() string {
	switch pt {
	case DirectedPerturb:
		return "Directed"
	case PatternGuidedPerturb:
		return "Pattern-Guided"
	case ColumnPerturb:
		return "Column"
	case RecencyPerturb:
		return "Recency"
	case RandomPerturb:
		return "Random"
	default:
		return "Unknown"
	}
}

// BLSState tracks the current state of the search algorithm.
type BLSState struct {
	omega       int          // Consecutive non-improving local optima visited
	L           int          // Current jump magnitude (number of perturbation swaps)
	lastOptCost float64      // Cost of the previous local optimum
	tabuMatrix  [][]int      // tabuMatrix[i][j] = iteration when swap(i,j) was last performed
	iteration   int          // Global iteration counter
	bestCost    float64      // Best cost found so far
	bestLayout  *SplitLayout // Best layout found
	startTime   time.Time    // Start time of optimization
}

// BLS implements the Breakout Local Search algorithm for keyboard layout optimization.
type BLS struct {
	params     BLSParams
	state      BLSState
	scorer     *Scorer
	corpus     *Corpus
	pinned     *PinnedKeys // Flags indicating which keys are pinned (cannot be swapped)
	rng        *rand.Rand
	numFree    int        // Number of free (non-pinned) keys
	validPairs [][2]uint8 // Pre-calculated valid key pairs (excludes pinned keys)

	// Pre-filtered bigrams for pattern analysis (computed per layout in Optimize())
	relevantBigrams []BigramCount // Only bigrams with both chars on layout, sorted by frequency
}

// BigramCount holds a bigram and its frequency for pre-filtering.
type BigramCount struct {
	Bigram Bigram
	Count  uint64
}

// NewBLS creates a new BLS optimizer instance.
func NewBLS(params BLSParams, scorer *Scorer, corpus *Corpus, pinned *PinnedKeys) *BLS {
	// Count number of free keys
	numFree := 0
	for _, isPinned := range pinned {
		if !isPinned {
			numFree++
		}
	}

	// Pre-calculate all valid key pairs (excludes pinned keys)
	// This avoids O(n²) nested loops in hot paths during optimization
	validPairs := make([][2]uint8, 0, (numFree*(numFree-1))/2)
	for i := range uint8(42) {
		if pinned[i] {
			continue
		}
		for j := i + 1; j < 42; j++ {
			if pinned[j] {
				continue
			}
			validPairs = append(validPairs, [2]uint8{i, j})
		}
	}

	return &BLS{
		params:     params,
		scorer:     scorer,
		corpus:     corpus,
		pinned:     pinned,
		rng:        rand.New(rand.NewSource(params.Seed)),
		numFree:    numFree,
		validPairs: validPairs,
	}
}

// prefilterBigrams filters and sorts corpus bigrams that are relevant to the given layout.
// This must be done per layout since different layouts may have different character sets.
// Only bigrams where both characters are present on the layout are kept.
func (bls *BLS) prefilterBigrams(layout *SplitLayout) {
	bls.relevantBigrams = make([]BigramCount, 0, len(bls.corpus.Bigrams))
	for bi, cnt := range bls.corpus.Bigrams {
		_, ok1 := layout.GetKeyInfo(bi[0])
		_, ok2 := layout.GetKeyInfo(bi[1])
		if ok1 && ok2 {
			bls.relevantBigrams = append(bls.relevantBigrams, BigramCount{
				Bigram: bi,
				Count:  cnt,
			})
		}
	}

	// Sort by frequency descending (most common bigrams first)
	sort.Slice(bls.relevantBigrams, func(i, j int) bool {
		return bls.relevantBigrams[i].Count > bls.relevantBigrams[j].Count
	})
}

// Optimize runs the BLS algorithm on the given layout and returns the best layout found.
// Progress can optionally be reported to the provided writer (use nil to disable).
func (bls *BLS) Optimize(layout *SplitLayout, progressWriter io.Writer) *SplitLayout {
	// Pre-filter and sort bigrams for pattern analysis
	bls.prefilterBigrams(layout)

	// Initialize search state
	bls.state = BLSState{
		L:          bls.params.L0,
		omega:      0,
		iteration:  0,
		tabuMatrix: make([][]int, 42),
		startTime:  time.Now(),
	}

	for i := range bls.state.tabuMatrix {
		bls.state.tabuMatrix[i] = make([]int, 42)
	}

	// Make a working copy of the layout
	current := layout.Clone()

	// Compute initial cost
	bls.state.bestCost = bls.scorer.Score(current)
	bls.state.bestLayout = current.Clone()
	bls.state.lastOptCost = bls.state.bestCost

	if progressWriter != nil {
		MustFprintf(progressWriter, "Starting BLS optimization\n")
		MustFprintf(progressWriter, "Initial cost: %.4f\n", bls.state.bestCost)
		MustFprintf(progressWriter, "Free keys: %d/%d\n\n", bls.numFree, 42)
		MustFprintln(progressWriter, bls.state.bestLayout)
	}

	// Main optimization loop
	for bls.state.iteration < bls.params.MaxIterations {
		// Check time limit
		elapsed := time.Since(bls.state.startTime)
		if elapsed >= bls.params.MaxTime {
			if progressWriter != nil {
				MustFprintf(progressWriter, "\nTime limit reached: %v\n", elapsed)
			}
			break
		}

		// Phase 1: Steepest Descent to local optimum
		bls.steepestDescent(current)

		// Compute cost of local optimum
		currentCost := bls.scorer.Score(current)
		bls.state.iteration++

		// Check if we improved
		if currentCost < bls.state.bestCost {
			bls.state.bestCost = currentCost
			bls.state.bestLayout = current.Clone()
			bls.state.omega = 0

			if progressWriter != nil {
				MustFprintf(progressWriter, "Iter %d: New best cost: %.4f (elapsed: %v)\n",
					bls.state.iteration, bls.state.bestCost, time.Since(bls.state.startTime).Round(time.Second))
				MustFprintln(progressWriter, current)
			}
		} else if math.Abs(currentCost-bls.state.lastOptCost) > 1e-9 {
			// Escaped to a different local optimum (but not better)
			bls.state.omega++
		}

		// Determine jump magnitude L
		if bls.state.omega > bls.params.T {
			// Strong diversification: search is stagnating
			bls.state.L = bls.params.LMax
			bls.state.omega = 0

			if progressWriter != nil && bls.params.ReportInterval > 0 {
				MustFprintf(progressWriter, "Iter %d: Strong perturbation triggered (L=%d)\n",
					bls.state.iteration, bls.state.L)
			}
		} else if math.Abs(currentCost-bls.state.lastOptCost) < 1e-9 {
			// Returned to same local optimum: increase jump magnitude
			bls.state.L++
		} else {
			// Escaped to different local optimum: reset jump magnitude
			bls.state.L = bls.params.L0
		}

		// Phase 2: Perturbation
		bls.state.lastOptCost = currentCost
		bls.perturb(current, bls.state.L)

		// Progress reporting
		if progressWriter != nil && bls.params.ReportInterval > 0 &&
			bls.state.iteration%bls.params.ReportInterval == 0 {
			MustFprintf(progressWriter, "Iter %d: Current: %.4f, Best: %.4f, L=%d, ω=%d\n",
				bls.state.iteration, currentCost, bls.state.bestCost, bls.state.L, bls.state.omega)
		}
	}

	if progressWriter != nil {
		elapsed := time.Since(bls.state.startTime)
		MustFprintf(progressWriter, "\nOptimization complete\n")
		MustFprintf(progressWriter, "Final best cost: %.4f\n", bls.state.bestCost)
		MustFprintf(progressWriter, "Total iterations: %d\n", bls.state.iteration)
		MustFprintf(progressWriter, "Total time: %v\n", elapsed.Round(time.Second))
	}

	bls.state.bestLayout.Name += "-best"
	return bls.state.bestLayout
}

// steepestDescent performs local search until a local optimum is reached.
// Uses best-improvement strategy: evaluates all valid swaps and applies the best one.
// Dispatches to parallel or sequential implementation based on params.
func (bls *BLS) steepestDescent(layout *SplitLayout) {
	if bls.params.UseParallel {
		bls.steepestDescentParallel(layout)
	} else {
		bls.steepestDescentSequential(layout)
	}
}

// steepestDescentSequential is the sequential implementation.
func (bls *BLS) steepestDescentSequential(layout *SplitLayout) {
	improved := true

	for improved {
		improved = false
		bestDelta := 0.0
		var bestI, bestJ uint8

		// Evaluate all possible swaps using pre-calculated pairs
		costBefore := bls.scorer.Score(layout)
		for _, pair := range bls.validPairs {
			i, j := pair[0], pair[1]

			// Compute delta by scoring after swap
			layout.Swap(i, j)
			costAfter := bls.scorer.Score(layout)
			layout.Swap(i, j) // Swap back

			delta := costAfter - costBefore

			if delta < bestDelta {
				bestDelta = delta
				bestI = i
				bestJ = j
				improved = true
			}
		}

		if improved {
			// Apply best swap
			layout.Swap(bestI, bestJ)

			// Update tabu matrix
			bls.state.tabuMatrix[bestI][bestJ] = bls.state.iteration
			bls.state.tabuMatrix[bestJ][bestI] = bls.state.iteration
			bls.state.iteration++
		}
	}
}

// swapResult holds the result of evaluating a single swap.
type swapResult struct {
	i     uint8
	j     uint8
	delta float64
}

// steepestDescentParallel is the parallel implementation using worker goroutines.
func (bls *BLS) steepestDescentParallel(layout *SplitLayout) {
	improved := true

	// Determine number of workers
	numWorkers := bls.params.ParallelWorkers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	for improved {
		improved = false
		bestDelta := 0.0
		var bestI, bestJ uint8

		costBefore := bls.scorer.Score(layout)

		// Distribute work into chunks
		numPairs := len(bls.validPairs)
		chunkSize := (numPairs + numWorkers - 1) / numWorkers

		results := make(chan swapResult, numWorkers)
		var wg sync.WaitGroup

		// Spawn workers
		for w := 0; w < numWorkers; w++ {
			start := w * chunkSize
			if start >= numPairs {
				break
			}
			end := min(start+chunkSize, numPairs)

			wg.Add(1)
			go func(pairs [][2]uint8) {
				defer wg.Done()

				// Each worker needs its own layout clone
				localLayout := layout.Clone()
				localBestDelta := 0.0
				var localBestI, localBestJ uint8
				found := false

				for _, pair := range pairs {
					i, j := pair[0], pair[1]

					// Compute delta by scoring after swap
					localLayout.Swap(i, j)
					costAfter := bls.scorer.Score(localLayout)
					localLayout.Swap(i, j) // Swap back

					delta := costAfter - costBefore

					if delta < localBestDelta {
						localBestDelta = delta
						localBestI = i
						localBestJ = j
						found = true
					}
				}

				if found {
					results <- swapResult{localBestI, localBestJ, localBestDelta}
				}
			}(bls.validPairs[start:end])
		}

		// Close results channel when all workers complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results from workers
		for result := range results {
			if result.delta < bestDelta {
				bestDelta = result.delta
				bestI = result.i
				bestJ = result.j
				improved = true
			}
		}

		if improved {
			// Apply best swap
			layout.Swap(bestI, bestJ)

			// Update tabu matrix
			bls.state.tabuMatrix[bestI][bestJ] = bls.state.iteration
			bls.state.tabuMatrix[bestJ][bestI] = bls.state.iteration
			bls.state.iteration++
		}
	}
}

// perturb applies L perturbation moves to escape the current local optimum.
func (bls *BLS) perturb(layout *SplitLayout, L int) {
	for range L {
		pertType := bls.selectPerturbationType()

		var swapI, swapJ uint8
		var valid bool

		switch pertType {
		case DirectedPerturb:
			swapI, swapJ, valid = bls.selectDirectedSwap(layout)
		case PatternGuidedPerturb:
			swapI, swapJ, valid = bls.selectPatternGuidedSwap(layout)
		case ColumnPerturb:
			bls.applyColumnSwap(layout)
			continue // Column swap applies multiple swaps
		case RecencyPerturb:
			swapI, swapJ, valid = bls.selectRecencySwap(layout)
		case RandomPerturb:
			swapI, swapJ, valid = bls.selectRandomSwap(layout)
		}

		if valid {
			layout.Swap(swapI, swapJ)

			// Update tabu matrix
			bls.state.tabuMatrix[swapI][swapJ] = bls.state.iteration
			bls.state.tabuMatrix[swapJ][swapI] = bls.state.iteration
		}

		bls.state.iteration++

		// Check if perturbation led to a new best
		cost := bls.scorer.Score(layout)
		if cost < bls.state.bestCost {
			bls.state.bestCost = cost
			bls.state.bestLayout = layout.Clone()
			bls.state.omega = 0
		}
	}
}

// selectPerturbationType chooses which perturbation type to use based on search state.
// Uses adaptive probability: directed perturbation is more likely early on,
// stronger diversification becomes more likely as search stagnates.
func (bls *BLS) selectPerturbationType() PerturbationType {
	// Calculate probability of directed perturbation
	P := math.Exp(-float64(bls.state.omega) / float64(bls.params.T))
	if P < bls.params.P0 {
		P = bls.params.P0
	}

	r := bls.rng.Float64()

	if r < P {
		return DirectedPerturb
	}

	// Distribute remaining probability among other perturbations
	remaining := 1.0 - P
	r = (r - P) / remaining // Normalize to [0, 1]

	cumulative := 0.0

	cumulative += bls.params.PatternWeight
	if r < cumulative {
		return PatternGuidedPerturb
	}

	cumulative += bls.params.ColumnWeight
	if r < cumulative {
		return ColumnPerturb
	}

	cumulative += bls.params.RandomWeight
	if r < cumulative {
		return RandomPerturb
	}

	return RecencyPerturb
}

// selectDirectedSwap selects a swap that minimizes cost degradation (tabu search style).
// Returns indices and validity flag.
func (bls *BLS) selectDirectedSwap(layout *SplitLayout) (uint8, uint8, bool) {
	bestDelta := math.Inf(1)
	var bestI, bestJ uint8
	found := false

	tabuTenure := bls.params.TabuMin + bls.rng.Intn(bls.params.TabuMax-bls.params.TabuMin+1)
	costBefore := bls.scorer.Score(layout)

	for _, pair := range bls.validPairs {
		i, j := pair[0], pair[1]

		// Check if move is tabu
		isTabu := (bls.state.iteration - bls.state.tabuMatrix[i][j]) < tabuTenure

		// Aspiration criterion: accept if leads to best solution
		layout.Swap(i, j)
		costAfter := bls.scorer.Score(layout)
		layout.Swap(i, j) // Swap back

		delta := costAfter - costBefore
		aspirationMet := costAfter < bls.state.bestCost

		if (!isTabu || aspirationMet) && delta < bestDelta {
			bestDelta = delta
			bestI = i
			bestJ = j
			found = true
		}
	}

	if !found {
		// Fallback to random
		return bls.selectRandomSwap(layout)
	}

	return bestI, bestJ, true
}

// selectPatternGuidedSwap targets keys involved in problematic patterns (SFB, LSB, etc.).
func (bls *BLS) selectPatternGuidedSwap(layout *SplitLayout) (uint8, uint8, bool) {
	problematic := bls.identifyProblematicKeys2(layout)

	if len(problematic) == 0 {
		return bls.selectRandomSwap(layout)
	}

	// Select a problematic key
	badIdx := problematic[bls.rng.Intn(len(problematic))]

	// Find a better position for this key
	goodIdx := bls.findBetterPosition(layout, badIdx)

	if goodIdx == 255 || bls.pinned[badIdx] || bls.pinned[goodIdx] {
		return bls.selectRandomSwap(layout)
	}

	return badIdx, goodIdx, true
}

// identifyProblematicKeys finds keys that contribute most to bad patterns.
func (bls *BLS) identifyProblematicKeys(layout *SplitLayout) []uint8 {
	// Create an analyzer to get pattern information
	// an := NewAnalyser(layout, bls.corpus, nil, nil)

	keyBadness := make(map[uint8]float64, 42)

	// Check SFB contribution
	for bi, biCnt := range bls.corpus.Bigrams {
		key1, ok1 := layout.GetKeyInfo(bi[0])
		key2, ok2 := layout.GetKeyInfo(bi[1])
		if !ok1 || !ok2 {
			continue
		}

		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			weight := float64(biCnt) / float64(bls.corpus.TotalBigramsCount)
			keyBadness[key1.Index] += weight * 100.0
			keyBadness[key2.Index] += weight * 100.0
		}
	}

	// Check LSB contribution
	for _, lsb := range layout.LSBs {
		bi := Bigram{layout.Runes[lsb.KeyIdx1], layout.Runes[lsb.KeyIdx2]}
		if cnt, ok := bls.corpus.Bigrams[bi]; ok {
			weight := float64(cnt) / float64(bls.corpus.TotalBigramsCount)
			keyBadness[lsb.KeyIdx1] += weight * 80.0
			keyBadness[lsb.KeyIdx2] += weight * 80.0
		}
	}

	// Sort by badness
	type keyScore struct {
		key   uint8
		score float64
	}

	scores := make([]keyScore, 0, len(keyBadness))
	for key, score := range keyBadness {
		if !bls.pinned[key] {
			scores = append(scores, keyScore{key, score})
		}
	}

	if len(scores) == 0 {
		return []uint8{}
	}

	// Sort descending by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top K
	k := min(bls.params.TopKProblematic, len(scores))

	result := make([]uint8, k)
	for i := range k {
		result[i] = scores[i].key
	}

	return result
}

// identifyProblematicKeys2 is an optimized version that uses pre-filtered bigrams.
// This avoids iterating over all corpus bigrams and filtering out irrelevant ones.
func (bls *BLS) identifyProblematicKeys2(layout *SplitLayout) []uint8 {
	keyBadness := make(map[uint8]float64, 42)

	// Check SFB contribution using pre-filtered bigrams
	// Only iterate over bigrams that are actually on the layout
	for _, bc := range bls.relevantBigrams {
		key1, _ := layout.GetKeyInfo(bc.Bigram[0])
		key2, _ := layout.GetKeyInfo(bc.Bigram[1])

		// Check if this is a same-finger bigram
		if key1.Finger == key2.Finger && key1.Index != key2.Index {
			weight := float64(bc.Count) / float64(bls.corpus.TotalBigramsCount)
			keyBadness[key1.Index] += weight * 100.0
			keyBadness[key2.Index] += weight * 100.0
		}
	}

	// Check LSB contribution
	// LSBs are already calculated and stored in the layout
	// We can use a map lookup instead of iterating all bigrams
	lsbSet := make(map[[2]uint8]bool, len(layout.LSBs))
	for _, lsb := range layout.LSBs {
		lsbSet[[2]uint8{lsb.KeyIdx1, lsb.KeyIdx2}] = true
	}

	// Only check relevant bigrams to see if they're LSBs
	for _, bc := range bls.relevantBigrams {
		key1, _ := layout.GetKeyInfo(bc.Bigram[0])
		key2, _ := layout.GetKeyInfo(bc.Bigram[1])

		// Check if this key pair is an LSB
		if lsbSet[[2]uint8{key1.Index, key2.Index}] {
			weight := float64(bc.Count) / float64(bls.corpus.TotalBigramsCount)
			keyBadness[key1.Index] += weight * 80.0
			keyBadness[key2.Index] += weight * 80.0
		}
	}

	// Sort by badness
	type keyScore struct {
		key   uint8
		score float64
	}

	scores := make([]keyScore, 0, len(keyBadness))
	for key, score := range keyBadness {
		if !bls.pinned[key] {
			scores = append(scores, keyScore{key, score})
		}
	}

	if len(scores) == 0 {
		return []uint8{}
	}

	// Sort descending by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top K
	k := min(bls.params.TopKProblematic, len(scores))

	result := make([]uint8, k)
	for i := range k {
		result[i] = scores[i].key
	}

	return result
}

// findBetterPosition finds a good position to swap the given key to.
func (bls *BLS) findBetterPosition(layout *SplitLayout, keyIdx uint8) uint8 {
	keyRune := layout.Runes[keyIdx]
	keyInfo, _ := layout.GetKeyInfo(keyRune)

	candidates := []uint8{}

	for pos := range uint8(42) {
		if pos == keyIdx || bls.pinned[pos] {
			continue
		}

		posRune := layout.Runes[pos]
		posInfo, _ := layout.GetKeyInfo(posRune)

		// Prefer different finger or opposite hand
		if posInfo.Hand != keyInfo.Hand || posInfo.Finger != keyInfo.Finger {
			candidates = append(candidates, pos)
		}
	}

	if len(candidates) == 0 {
		// Fallback: any valid position
		for pos := range uint8(42) {
			if pos != keyIdx && !bls.pinned[pos] {
				candidates = append(candidates, pos)
			}
		}
	}

	if len(candidates) == 0 {
		return 255 // Invalid
	}

	return candidates[bls.rng.Intn(len(candidates))]
}

// applyColumnSwap swaps two entire columns of keys.
func (bls *BLS) applyColumnSwap(layout *SplitLayout) {
	// Identify columns (0-11)
	colA := bls.rng.Intn(12)
	colB := bls.rng.Intn(12)

	for colA == colB {
		colB = bls.rng.Intn(12)
	}

	// Swap keys in corresponding positions of each column
	for row := range uint8(3) {
		posA := row*12 + uint8(colA)
		posB := row*12 + uint8(colB)

		if !bls.pinned[posA] && !bls.pinned[posB] {
			layout.Swap(posA, posB)
		}
	}
}

// selectRecencySwap selects the least recently performed swap.
func (bls *BLS) selectRecencySwap(layout *SplitLayout) (uint8, uint8, bool) {
	oldestTime := bls.state.iteration
	var oldestI, oldestJ uint8
	found := false

	for _, pair := range bls.validPairs {
		i, j := pair[0], pair[1]

		lastTime := bls.state.tabuMatrix[i][j]
		if lastTime < oldestTime {
			oldestTime = lastTime
			oldestI = i
			oldestJ = j
			found = true
		}
	}

	if !found {
		return bls.selectRandomSwap(layout)
	}

	return oldestI, oldestJ, true
}

// selectRandomSwap selects a completely random valid swap.
func (bls *BLS) selectRandomSwap(_ *SplitLayout) (uint8, uint8, bool) {
	if len(bls.validPairs) == 0 {
		return 0, 0, false
	}

	swap := bls.validPairs[bls.rng.Intn(len(bls.validPairs))]
	return swap[0], swap[1], true
}
