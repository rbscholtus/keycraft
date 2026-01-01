package keycraft

import (
	"runtime"
	"sort"
	"sync"
	"testing"
)

// Parallel version of SortedMap using goroutines for conversion
func SortedMapParallel[K comparable](m map[K]uint64) []CountPair[K] {
	if m == nil {
		return []CountPair[K]{}
	}

	size := len(m)
	pairs := make([]CountPair[K], size)

	// Determine chunk size and number of workers
	numWorkers := runtime.GOMAXPROCS(0)
	if size < 1000 {
		// For small maps, parallel overhead isn't worth it
		numWorkers = 1
	}

	chunkSize := (size + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	idx := 0

	// Convert map to slice in parallel
	for chunk := 0; chunk < numWorkers && idx < size; chunk++ {
		start := chunk * chunkSize
		end := start + chunkSize
		if end > size {
			end = size
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			i := start
			for k, v := range m {
				if i >= end {
					break
				}
				pairs[i] = CountPair[K]{k, v}
				i++
			}
		}(start, end)

		idx = end
	}

	wg.Wait()

	// Sort is already pretty optimized, parallelizing it is complex
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	return pairs
}

// Alternative: Use buffered channel for map conversion
func SortedMapChannel[K comparable](m map[K]uint64) []CountPair[K] {
	if m == nil {
		return []CountPair[K]{}
	}

	size := len(m)
	pairs := make([]CountPair[K], 0, size)
	ch := make(chan CountPair[K], 100)

	go func() {
		for k, v := range m {
			ch <- CountPair[K]{k, v}
		}
		close(ch)
	}()

	for pair := range ch {
		pairs = append(pairs, pair)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	return pairs
}

// Benchmark the original SortedMap
func BenchmarkSortedMap(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100", 100},
		{"1k", 1000},
		{"10k", 10000},
		{"100k", 100000},
	}

	for _, s := range sizes {
		m := make(map[Bigram]uint64, s.size)
		for i := 0; i < s.size; i++ {
			m[Bigram{rune(i % 256), rune((i / 256) % 256)}] = uint64(i)
		}

		b.Run(s.name, func(b *testing.B) {
			for b.Loop() {
				_ = SortedMap(m)
			}
		})
	}
}

// Benchmark parallel version
func BenchmarkSortedMapParallel(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100", 100},
		{"1k", 1000},
		{"10k", 10000},
		{"100k", 100000},
	}

	for _, s := range sizes {
		m := make(map[Bigram]uint64, s.size)
		for i := 0; i < s.size; i++ {
			m[Bigram{rune(i % 256), rune((i / 256) % 256)}] = uint64(i)
		}

		b.Run(s.name, func(b *testing.B) {
			for b.Loop() {
				_ = SortedMapParallel(m)
			}
		})
	}
}

// Benchmark channel version
func BenchmarkSortedMapChannel(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100", 100},
		{"1k", 1000},
		{"10k", 10000},
		{"100k", 100000},
	}

	for _, s := range sizes {
		m := make(map[Bigram]uint64, s.size)
		for i := 0; i < s.size; i++ {
			m[Bigram{rune(i % 256), rune((i / 256) % 256)}] = uint64(i)
		}

		b.Run(s.name, func(b *testing.B) {
			for b.Loop() {
				_ = SortedMapChannel(m)
			}
		})
	}
}
