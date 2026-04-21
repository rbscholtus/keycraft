package keycraft

import (
	"context"
	"errors"
	"testing"
	"time"
)

// newTestBLS builds a minimal BLS configured for fast tests. Corpus and
// layouts come from the repo's data/ directory.
func newTestBLS(t *testing.T) (*BLS, *SplitLayout) {
	t.Helper()

	corpus, err := NewCorpusFromFile("monkeyracer", "../../data/corpus/monkeyracer.txt", false, 98.0)
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		t.Fatalf("load layout: %v", err)
	}

	targets := &TargetLoads{
		TargetHandLoad:   DefaultTargetHandLoad(),
		TargetFingerLoad: DefaultTargetFingerLoad(),
		TargetRowLoad:    DefaultTargetRowLoad(),
		PinkyPenalties:   DefaultPinkyPenalties(),
	}
	scorer, err := NewScorer("../../data/layouts", corpus, targets, NewWeights())
	if err != nil {
		t.Fatalf("new scorer: %v", err)
	}

	pinned := &PinnedKeys{}
	for i, r := range layout.Runes {
		if r == 0 || r == ' ' {
			pinned[i] = true
		}
	}
	numFree := 0
	for _, p := range pinned {
		if !p {
			numFree++
		}
	}

	params := DefaultBLSParams(numFree)
	params.Seed = 42
	params.UseParallel = false
	params.MaxIterations = 50
	params.MaxTime = time.Minute

	return NewBLS(params, scorer, corpus, pinned), layout
}

// TestOptimizeLocalSearchOnly — fast mode runs one descent and stops. The
// result must be at least as good as the starting layout.
func TestOptimizeLocalSearchOnly(t *testing.T) {
	bls, layout := newTestBLS(t)
	bls.params.LocalSearchOnly = true
	bls.params.MaxIterations = 10000 // should not matter; loop is skipped

	startCost := bls.scorer.Score(layout)

	best, err := bls.Optimize(context.Background(), layout, nil)
	if err != nil {
		t.Fatalf("Optimize: %v", err)
	}
	if best == nil {
		t.Fatal("nil best layout")
	}

	endCost := bls.scorer.Score(best)
	if endCost > startCost {
		t.Errorf("cost regressed: start=%.6f, end=%.6f", startCost, endCost)
	}
}

// TestOptimizeContextCancel — cancelling the ctx mid-run exits promptly and
// returns ctx.Err plus the best-so-far layout.
func TestOptimizeContextCancel(t *testing.T) {
	bls, layout := newTestBLS(t)
	bls.params.MaxIterations = 100000 // will not complete naturally
	bls.params.MaxTime = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately so the first ctx check trips.
	cancel()

	start := time.Now()
	best, err := bls.Optimize(ctx, layout, nil)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if best == nil {
		t.Error("expected best-so-far layout even on cancel")
	}
	if elapsed > 2*time.Second {
		t.Errorf("cancel not prompt: took %v", elapsed)
	}
}

// TestOptimizeContextCancelMidRun — cancel after a short delay (search is
// actually underway) still exits promptly.
func TestOptimizeContextCancelMidRun(t *testing.T) {
	bls, layout := newTestBLS(t)
	bls.params.MaxIterations = 100000
	bls.params.MaxTime = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := bls.Optimize(ctx, layout, nil)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// Allow up to a couple of seconds to finish the current descent's iteration.
	if elapsed > 5*time.Second {
		t.Errorf("cancel too slow: %v", elapsed)
	}
}

// TestSteepestDescentContextCancel — cancelling ctx interrupts the descent's
// inner loops.
func TestSteepestDescentContextCancel(t *testing.T) {
	bls, layout := newTestBLS(t)
	bls.prefilterBigrams(layout)
	bls.state = BLSState{tabuMatrix: make([][]int, 42)}
	for i := range bls.state.tabuMatrix {
		bls.state.tabuMatrix[i] = make([]int, 42)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	current := layout.Clone()
	err := bls.steepestDescentSequential(ctx, current)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
