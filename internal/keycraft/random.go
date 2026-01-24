package keycraft

import (
	"math/rand/v2"
	"sync"
	"time"
)

// LockedSource wraps a PCG source with a mutex to make it thread-safe.
// This is essential for Go 1.22+ when sharing a specific seeded generator
// across multiple goroutines.
type LockedSource struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewLockedRNG creates a new thread-safe RNG instance.
// If both seeds are 0, it initializes with a time-based seed.
func NewLockedRNG(seed1, seed2 uint64) *LockedSource {
	if seed1 == 0 && seed2 == 0 {
		seed1 = uint64(time.Now().UnixNano())
		seed2 = seed1 ^ 0x9e3779b97f4a7c15
	}
	return &LockedSource{
		rng: rand.New(rand.NewPCG(seed1, seed2)),
	}
}

// Uint64 returns a pseudo-random 64-bit unsigned integer.
// Common Use Case: Generating unique identifiers, raw bitwise operations,
// or as a base for custom distribution logic.
func (ls *LockedSource) Uint64() uint64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.rng.Uint64()
}

// IntN returns a pseudo-random integer in the interval [0, n).
// It panics if n <= 0.
// Common Use Case: Selecting a random index from a slice or simulating
// discrete events like a dice roll.
func (ls *LockedSource) IntN(n int) int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.rng.IntN(n)
}

// Float64 returns a pseudo-random number in the interval [0.0, 1.0).
// Common Use Case: Implementing probability-based logic, such as a
// "25% chance to trigger an effect" (if rng.Float64() < 0.25).
func (ls *LockedSource) Float64() float64 {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.rng.Float64()
}

// Perm returns a slice of n integers containing a random permutation
// of the integers [0, n).
// Common Use Case: Determining a unique, non-repeating order for a
// set of tasks, or selecting a random sequence of unique IDs.
func (ls *LockedSource) Perm(n int) []int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.rng.Perm(n)
}

// Shuffle pseudo-randomizes the order of elements using an in-place swap.
// n is the number of elements, and swap is the function to swap them.
// Common Use Case: Reordering an existing collection, such as shuffling
// a deck of cards or a playlist of songs.
func (ls *LockedSource) Shuffle(n int, swap func(i, j int)) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.rng.Shuffle(n, swap)
}

// ShuffleSlice is a generic helper that shuffles any slice using the locked RNG.
func ShuffleSlice[T any](ls *LockedSource, slice []T) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.rng.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}
