package keycraft

import (
	"runtime"
	"testing"
)

// BenchmarkSteepestDescentSequential benchmarks the sequential steepest descent implementation.
func BenchmarkSteepestDescentSequential(b *testing.B) {
	bls, layout := createBenchBLS(b)

	// Disable parallelism
	bls.params.UseParallel = false

	// Pre-filter bigrams (normally done in Optimize())
	bls.prefilterBigrams(layout)

	// Initialize state
	bls.state = BLSState{
		L:          bls.params.L0,
		omega:      0,
		iteration:  0,
		tabuMatrix: make([][]int, 42),
	}
	for i := range bls.state.tabuMatrix {
		bls.state.tabuMatrix[i] = make([]int, 42)
	}

	b.ResetTimer()
	for b.Loop() {
		testLayout := layout.Clone()
		bls.steepestDescentSequential(testLayout)
	}
}

// BenchmarkSteepestDescentParallelWorkers benchmarks parallel descent with different worker counts.
func BenchmarkSteepestDescentParallelWorkers(b *testing.B) {
	benchmarks := []struct {
		name    string
		workers int // 0 means use runtime.NumCPU()
	}{
		{"1worker", 1},
		{"2workers", 2},
		{"4workers", 4},
		{"6workers", 6},
		{"8workers", 8},
		{"auto", 0}, // Use runtime.NumCPU()
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bls, layout := createBenchBLS(b)

			bls.params.UseParallel = true
			bls.params.ParallelWorkers = bm.workers

			bls.prefilterBigrams(layout)

			bls.state = BLSState{
				L:          bls.params.L0,
				omega:      0,
				iteration:  0,
				tabuMatrix: make([][]int, 42),
			}
			for i := range bls.state.tabuMatrix {
				bls.state.tabuMatrix[i] = make([]int, 42)
			}

			b.ResetTimer()
			for b.Loop() {
				testLayout := layout.Clone()
				bls.steepestDescentParallel(testLayout)
			}
		})
	}
}

// BenchmarkSteepestDescentAuto benchmarks using the automatic dispatcher.
func BenchmarkSteepestDescentAuto(b *testing.B) {
	bls, layout := createBenchBLS(b)

	// Use automatic parallelism (should choose parallel for this workload)
	bls.params.UseParallel = true
	bls.params.ParallelWorkers = 0

	bls.prefilterBigrams(layout)

	bls.state = BLSState{
		L:          bls.params.L0,
		omega:      0,
		iteration:  0,
		tabuMatrix: make([][]int, 42),
	}
	for i := range bls.state.tabuMatrix {
		bls.state.tabuMatrix[i] = make([]int, 42)
	}

	b.ResetTimer()
	for b.Loop() {
		testLayout := layout.Clone()
		bls.steepestDescent(testLayout)
	}
}

// TestSteepestDescentEquivalence verifies parallel and sequential produce same results.
func TestSteepestDescentEquivalence(t *testing.T) {
	// Load corpus
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 0)
	if err != nil {
		t.Skipf("Skipping test - corpus not available: %v", err)
	}

	// Load layout
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		t.Skipf("Skipping test - layout not available: %v", err)
	}

	// Create pinned keys
	pinned := &PinnedKeys{}
	for i, r := range layout.Runes {
		if r == 0 || r == ' ' {
			pinned[i] = true
		}
	}

	// Count free keys
	numFree := 0
	for _, isPinned := range pinned {
		if !isPinned {
			numFree++
		}
	}

	// Create parameters
	params := DefaultBLSParams(numFree)
	params.Seed = 42 // Fixed seed for reproducibility

	// Create scorer
	scorer, err := NewScorer("../../data/layouts", corpus, DefaultIdealRowLoad(), DefaultIdealFingerLoad(), DefaultPinkyPenalties(), NewWeights())
	if err != nil {
		t.Skipf("Skipping test - layouts not available: %v", err)
	}

	// Test sequential version
	blsSeq := NewBLS(params, scorer, corpus, pinned)
	blsSeq.params.UseParallel = false
	blsSeq.prefilterBigrams(layout)
	blsSeq.state = BLSState{
		L:          blsSeq.params.L0,
		omega:      0,
		iteration:  0,
		tabuMatrix: make([][]int, 42),
	}
	for i := range blsSeq.state.tabuMatrix {
		blsSeq.state.tabuMatrix[i] = make([]int, 42)
	}

	layoutSeq := layout.Clone()
	blsSeq.steepestDescentSequential(layoutSeq)
	costSeq := scorer.Score(layoutSeq)

	// Test parallel version
	blsPar := NewBLS(params, scorer, corpus, pinned)
	blsPar.params.UseParallel = true
	blsPar.params.ParallelWorkers = runtime.NumCPU()
	blsPar.prefilterBigrams(layout)
	blsPar.state = BLSState{
		L:          blsPar.params.L0,
		omega:      0,
		iteration:  0,
		tabuMatrix: make([][]int, 42),
	}
	for i := range blsPar.state.tabuMatrix {
		blsPar.state.tabuMatrix[i] = make([]int, 42)
	}

	layoutPar := layout.Clone()
	blsPar.steepestDescentParallel(layoutPar)
	costPar := scorer.Score(layoutPar)

	// Costs should be identical (both find same local optimum)
	const epsilon = 1e-9
	if costSeq != costPar {
		// Allow small floating point differences
		diff := costSeq - costPar
		if diff < 0 {
			diff = -diff
		}
		if diff > epsilon {
			t.Errorf("Sequential and parallel produced different costs: seq=%.10f, par=%.10f, diff=%.10f",
				costSeq, costPar, diff)
		}
	}

	// Layouts should be identical
	for i := range layoutSeq.Runes {
		if layoutSeq.Runes[i] != layoutPar.Runes[i] {
			t.Errorf("Layouts differ at position %d: seq=%c, par=%c",
				i, layoutSeq.Runes[i], layoutPar.Runes[i])
		}
	}
}
