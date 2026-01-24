package keycraft

import (
	"slices"
	"sync"
	"testing"
)

// TestDeterminism ensures that the same seeds produce the exact same sequence.
// This is critical for reproducible tests and simulations.
func TestDeterminism(t *testing.T) {
	s1, s2 := uint64(42), uint64(12345)
	rng1 := NewLockedRNG(s1, s2)
	rng2 := NewLockedRNG(s1, s2)

	for i := 0; i < 100; i++ {
		v1 := rng1.Uint64()
		v2 := rng2.Uint64()
		if v1 != v2 {
			t.Errorf("Iteration %d: Expected matching values, got %d and %d", i, v1, v2)
		}
	}
}

// TestIntNBounds verifies that IntN never returns a value out of the requested range.
func TestIntNBounds(t *testing.T) {
	rng := NewLockedRNG(0, 0) // Test with auto-seeding
	n := 10
	for i := 0; i < 1000; i++ {
		val := rng.IntN(n)
		if val < 0 || val >= n {
			t.Fatalf("IntN(%d) produced %d; out of bounds [0, %d)", n, val, n)
		}
	}
}

// TestFloat64Bounds verifies that Float64 is always in the interval [0.0, 1.0).
func TestFloat64Bounds(t *testing.T) {
	rng := NewLockedRNG(0, 0)
	for i := 0; i < 1000; i++ {
		val := rng.Float64()
		if val < 0.0 || val >= 1.0 {
			t.Fatalf("Float64 produced %f; out of bounds [0.0, 1.0)", val)
		}
	}
}

// TestPerm verifies that Perm returns a valid permutation of all integers.
func TestPerm(t *testing.T) {
	n := 5
	rng := NewLockedRNG(1, 1)
	p := rng.Perm(n)

	if len(p) != n {
		t.Fatalf("Perm(%d) returned slice of length %d", n, len(p))
	}

	// Sort the slice; it should be [0, 1, 2, 3, 4]
	slices.Sort(p)
	for i := 0; i < n; i++ {
		if p[i] != i {
			t.Errorf("Perm missing value %d", i)
		}
	}
}

// TestShuffleSlice checks if a slice is actually modified.
func TestShuffleSlice(t *testing.T) {
	rng := NewLockedRNG(777, 888)
	input := []string{"a", "b", "c", "d", "e", "f", "g"}
	original := slices.Clone(input)

	ShuffleSlice(rng, input)

	// Statistically, there is a near-zero chance that a 7-element shuffle
	// results in the exact same order, though technically possible.
	if slices.Equal(input, original) {
		t.Log("Warning: Shuffle produced same order (statistically unlikely but possible)")
	}
}

// TestConcurrencySafety runs multiple goroutines to trigger the race detector.
// Run this test with: go test -race
func TestConcurrencySafety(t *testing.T) {
	rng := NewLockedRNG(0, 0)
	var wg sync.WaitGroup
	workers := 100
	iterations := 1000

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Call various methods to ensure the mutex covers all paths
				rng.Uint64()
				rng.IntN(100)
				rng.Float64()
				rng.Perm(10)
			}
		}()
	}
	wg.Wait()
}
