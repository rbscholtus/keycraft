package keycraft

import (
	"fmt"
	"strconv"
	"testing"
)

func LayoutCacheKey2(layout *SplitLayout) string {
	return fmt.Sprintf("%d%s", layout.LayoutType, string(layout.Runes[:]))
}

func LayoutCacheKey3(layout *SplitLayout) string {
	return strconv.Itoa(int(layout.LayoutType)) + string(layout.Runes[:])
}

func LayoutCacheKey4(layout *SplitLayout) string {
	key := make([]byte, 1+42*4) // 1 byte for type + up to 4 bytes per rune (UTF-8)
	key[0] = byte(layout.LayoutType)
	n := 1
	for _, r := range layout.Runes {
		n += copy(key[n:], string(r))
	}
	return string(key[:n])
}

func LayoutCacheKey5(layout *SplitLayout) string {
	// Pre-allocate exact size: 1 byte for type + 42 runes (each up to 4 bytes in UTF-8)
	buf := make([]byte, 0, 1+42*4)
	buf = append(buf, byte(layout.LayoutType))
	for _, r := range layout.Runes {
		buf = append(buf, []byte(string(r))...)
	}
	return string(buf)
}

// TestLayoutCacheKey tests that all cache key implementations generate unique keys
// for different layouts and identical keys for the same layout configuration.
func TestLayoutCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		layout1  *SplitLayout
		layout2  *SplitLayout
		shouldEq bool
	}{
		{
			name: "identical layouts should produce same key",
			layout1: &SplitLayout{
				Name:       "layout1",
				LayoutType: ROWSTAG,
				Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			layout2: &SplitLayout{
				Name:       "layout2", // Different name
				LayoutType: ROWSTAG,
				Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			shouldEq: true,
		},
		{
			name: "different layout types should produce different keys",
			layout1: &SplitLayout{
				Name:       "layout1",
				LayoutType: ROWSTAG,
				Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			layout2: &SplitLayout{
				Name:       "layout1",
				LayoutType: ORTHO,
				Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			shouldEq: false,
		},
		{
			name: "different runes should produce different keys",
			layout1: &SplitLayout{
				Name:       "layout1",
				LayoutType: ROWSTAG,
				Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
			},
			layout2: &SplitLayout{
				Name:       "layout1",
				LayoutType: ROWSTAG,
				Runes:      [42]rune{'a', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}, // First rune different
			},
			shouldEq: false,
		},
	}

	// Test all implementations
	implementations := []struct {
		name string
		fn   func(*SplitLayout) string
	}{
		{"LayoutCacheKey (strings.Builder original)", layoutCacheKey},
		{"LayoutCacheKey2 (fmt.Sprintf)", LayoutCacheKey2},
		{"LayoutCacheKey3 (strconv.Itoa)", LayoutCacheKey3},
		{"LayoutCacheKey4 (byte array copy)", LayoutCacheKey4},
		{"LayoutCacheKey5 (byte array append)", LayoutCacheKey5},
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					key1 := impl.fn(tt.layout1)
					key2 := impl.fn(tt.layout2)

					if tt.shouldEq {
						if key1 != key2 {
							t.Errorf("Expected identical keys, got %q and %q", key1, key2)
						}
					} else {
						if key1 == key2 {
							t.Errorf("Expected different keys, but both were %q", key1)
						}
					}
				})
			}
		})
	}
}

// TestLayoutCacheKeyConsistency verifies that each implementation is self-consistent
func TestLayoutCacheKeyConsistency(t *testing.T) {
	layout := &SplitLayout{
		Name:       "test",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	// Each implementation should be self-consistent
	key1a := layoutCacheKey(layout)
	key1b := layoutCacheKey(layout)
	if key1a != key1b {
		t.Errorf("LayoutCacheKey is not self-consistent: %q vs %q", key1a, key1b)
	}

	key2a := LayoutCacheKey2(layout)
	key2b := LayoutCacheKey2(layout)
	if key2a != key2b {
		t.Errorf("LayoutCacheKey2 is not self-consistent: %q vs %q", key2a, key2b)
	}

	key3a := LayoutCacheKey3(layout)
	key3b := LayoutCacheKey3(layout)
	if key3a != key3b {
		t.Errorf("LayoutCacheKey3 is not self-consistent: %q vs %q", key3a, key3b)
	}

	key4a := LayoutCacheKey4(layout)
	key4b := LayoutCacheKey4(layout)
	if key4a != key4b {
		t.Errorf("LayoutCacheKey4 is not self-consistent: %q vs %q", key4a, key4b)
	}

	key5a := LayoutCacheKey5(layout)
	key5b := LayoutCacheKey5(layout)
	if key5a != key5b {
		t.Errorf("LayoutCacheKey5 is not self-consistent: %q vs %q", key5a, key5b)
	}
}

// Benchmark helpers
var benchLayout = &SplitLayout{
	Name:       "benchmark",
	LayoutType: ROWSTAG,
	Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
}

var result string

// BenchmarkLayoutCacheKey benchmarks the fmt.Sprintf implementation
func BenchmarkLayoutCacheKey(b *testing.B) {
	var r string
	for b.Loop() {
		r = layoutCacheKey(benchLayout)
	}
	result = r
}

// BenchmarkLayoutCacheKey2 benchmarks the strings.Builder implementation
func BenchmarkLayoutCacheKey2(b *testing.B) {
	var r string
	for b.Loop() {
		r = LayoutCacheKey2(benchLayout)
	}
	result = r
}

// BenchmarkLayoutCacheKey3 benchmarks the strconv.Itoa implementation
func BenchmarkLayoutCacheKey3(b *testing.B) {
	var r string
	for b.Loop() {
		r = LayoutCacheKey3(benchLayout)
	}
	result = r
}

// BenchmarkLayoutCacheKey4 benchmarks the byte array copy implementation
func BenchmarkLayoutCacheKey4(b *testing.B) {
	var r string
	for b.Loop() {
		r = LayoutCacheKey4(benchLayout)
	}
	result = r
}

// BenchmarkLayoutCacheKey5 benchmarks the byte array append implementation
func BenchmarkLayoutCacheKey5(b *testing.B) {
	var r string
	for b.Loop() {
		r = LayoutCacheKey5(benchLayout)
	}
	result = r
}

// createTestCorpus creates a minimal corpus for testing
func createTestCorpus() *Corpus {
	return &Corpus{
		Name: "test",
		Unigrams: map[Unigram]uint64{
			'a': 100, 'e': 120, 't': 90, 'n': 80, 'i': 110,
			'o': 85, 's': 75, 'r': 70, 'h': 65, 'l': 60,
		},
		TotalUnigramsCount: 1000,
		Bigrams: map[Bigram]uint64{
			{'t', 'h'}: 50, {'h', 'e'}: 45, {'a', 'n'}: 40,
			{'i', 'n'}: 35, {'e', 'r'}: 30, {'o', 'n'}: 25,
		},
		TotalBigramsCount: 225,
		Trigrams: map[Trigram]uint64{
			{'t', 'h', 'e'}: 30, {'a', 'n', 'd'}: 25,
			{'i', 'n', 'g'}: 20, {'t', 'i', 'o'}: 15,
		},
		TotalTrigramsCount: 90,
	}
}

// createTestScorer creates a minimal Scorer for testing purposes
func createTestScorer() *Scorer {
	// Create a simple scorer with known weights, medians, and IQRs
	corpus, err := NewCorpusFromFile("default", "data/corpus/default.txt", false, 0)
	if err != nil {
		corpus = createTestCorpus()
	}
	return &Scorer{
		corpus:              corpus,
		targetRowBalance:    &[3]float64{0.3, 0.4, 0.3},
		targetFingerBalance: &[10]float64{0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1},
		medians: map[string]float64{
			"SFB": 1.5,
			"LSB": 2.0,
		},
		iqrs: map[string]float64{
			"SFB": 0.5,
			"LSB": 0.8,
		},
		weights: map[string]float64{
			"SFB": -1.0,
			"LSB": -0.5,
		},
		scoreCache: make(map[string]float64, 100),
	}
}

// TestScoreCache verifies that Score() correctly caches results
func TestScoreCache(t *testing.T) {
	scorer := createTestScorer()

	layout := &SplitLayout{
		Name:       "test",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	// First call should calculate and cache
	score1 := scorer.Score(layout)

	// Verify it was cached
	cacheKey := layoutCacheKey(layout)
	cachedScore, exists := scorer.scoreCache[cacheKey]
	if !exists {
		t.Error("Score was not cached after first call")
	}
	if cachedScore != score1 {
		t.Errorf("Cached score %f doesn't match returned score %f", cachedScore, score1)
	}

	// Second call should return cached value
	score2 := scorer.Score(layout)
	if score1 != score2 {
		t.Errorf("Cached score changed: first=%f, second=%f", score1, score2)
	}

	// Third call should also return same value
	score3 := scorer.Score(layout)
	if score1 != score3 {
		t.Errorf("Cached score changed: first=%f, third=%f", score1, score3)
	}
}

// TestScoreCacheUniqueness verifies different layouts get different cache entries
func TestScoreCacheUniqueness(t *testing.T) {
	scorer := createTestScorer()

	layout1 := &SplitLayout{
		Name:       "layout1",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	layout2 := &SplitLayout{
		Name:       "layout2",
		LayoutType: ORTHO, // Different layout type
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	layout3 := &SplitLayout{
		Name:       "layout3",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'a', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}, // Different runes
	}

	score1 := scorer.Score(layout1)
	score2 := scorer.Score(layout2)
	score3 := scorer.Score(layout3)

	// Verify all three are cached
	if len(scorer.scoreCache) != 3 {
		t.Errorf("Expected 3 cache entries, got %d", len(scorer.scoreCache))
	}

	// Verify they produce different scores (since they have different configurations)
	// Note: We can't guarantee they're different without a full Analyser, so we just check caching works
	key1 := layoutCacheKey(layout1)
	key2 := layoutCacheKey(layout2)
	key3 := layoutCacheKey(layout3)

	if scorer.scoreCache[key1] != score1 {
		t.Error("Layout 1 score not properly cached")
	}
	if scorer.scoreCache[key2] != score2 {
		t.Error("Layout 2 score not properly cached")
	}
	if scorer.scoreCache[key3] != score3 {
		t.Error("Layout 3 score not properly cached")
	}
}

// TestScoreCacheIgnoresName verifies cache uses layout configuration, not name
func TestScoreCacheIgnoresName(t *testing.T) {
	scorer := createTestScorer()

	layout1 := &SplitLayout{
		Name:       "name1",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	layout2 := &SplitLayout{
		Name:       "name2", // Different name
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	score1 := scorer.Score(layout1)
	score2 := scorer.Score(layout2)

	// Should use cache for second call since configuration is identical
	if score1 != score2 {
		t.Errorf("Same configuration with different names produced different scores: %f vs %f", score1, score2)
	}

	// Should only have one cache entry
	if len(scorer.scoreCache) != 1 {
		t.Errorf("Expected 1 cache entry for identical configurations, got %d", len(scorer.scoreCache))
	}
}

// BenchmarkScoreFirstCall benchmarks the first Score() call (uncached)
func BenchmarkScoreFirstCall(b *testing.B) {
	layout := &SplitLayout{
		Name:       "bench",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		scorer := createTestScorer() // Fresh scorer each time to avoid cache
		b.StartTimer()
		_ = scorer.Score(layout)
	}
}

// BenchmarkScoreCachedCall benchmarks repeated Score() calls (cached)
func BenchmarkScoreCachedCall(b *testing.B) {
	scorer := createTestScorer()
	layout := &SplitLayout{
		Name:       "bench",
		LayoutType: ROWSTAG,
		Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	}

	// Prime the cache
	_ = scorer.Score(layout)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(layout)
	}
}

// BenchmarkScoreCacheVsUncached compares cache hit vs cache miss
func BenchmarkScoreCacheVsUncached(b *testing.B) {
	b.Run("Uncached", func(b *testing.B) {
		layout := &SplitLayout{
			Name:       "bench",
			LayoutType: ROWSTAG,
			Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			scorer := createTestScorer()
			b.StartTimer()
			_ = scorer.Score(layout)
		}
	})

	b.Run("Cached", func(b *testing.B) {
		scorer := createTestScorer()
		layout := &SplitLayout{
			Name:       "bench",
			LayoutType: ROWSTAG,
			Runes:      [42]rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', ';', 'z', 'x', 'c', 'v', 'b', 'n', 'm', ',', '.', '/', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
		}
		_ = scorer.Score(layout)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = scorer.Score(layout)
		}
	})
}

/*
====================================================================================
HOW TO RUN TESTS AND BENCHMARKS
====================================================================================

RUNNING TESTS:
-------------
1. Run all tests:
   go test -v ./internal/keycraft

2. Run specific test pattern:
   go test -v ./internal/keycraft -run TestLayoutCacheKey
   go test -v ./internal/keycraft -run TestScore

3. Run cache-related tests only:
   go test -v ./internal/keycraft -run Cache


RUNNING BENCHMARKS:
------------------
1. Run all benchmarks:
   go test -bench=. ./internal/keycraft

2. Run benchmarks with memory allocations (RECOMMENDED):
   go test -bench=. -benchmem ./internal/keycraft

3. Run specific benchmark pattern:
   go test -bench=BenchmarkLayoutCacheKey -benchmem ./internal/keycraft
   go test -bench=BenchmarkScore -benchmem ./internal/keycraft

4. Compare cached vs uncached Score() performance:
   go test -bench=BenchmarkScore -benchmem ./internal/keycraft

5. Run benchmarks multiple times for statistical accuracy:
   go test -bench=. -benchmem -count=10 ./internal/keycraft

6. Save benchmark results to file:
   go test -bench=. -benchmem ./internal/keycraft > bench_results.txt

7. Compare benchmarks over time using benchstat:
   go test -bench=. -benchmem -count=10 ./internal/keycraft > old.txt
   # (make changes)
   go test -bench=. -benchmem -count=10 ./internal/keycraft > new.txt
   benchstat old.txt new.txt


BENCHMARK OUTPUT EXPLANATION:
-----------------------------
With -benchmem flag, you'll see:
   - ns/op: nanoseconds per operation (lower is better)
   - B/op: bytes allocated per operation (lower is better)
   - allocs/op: number of allocations per operation (lower is better)


====================================================================================
BENCHMARK RESULTS (Apple M3, darwin/arm64)
====================================================================================

CACHE KEY GENERATION BENCHMARKS:
--------------------------------
Implementation                      ns/op      B/op    allocs/op   Ranking
--------------------------------------------------------------------------------
LayoutCacheKey (strings.Builder)    137.4     96        2          🥇 WINNER
LayoutCacheKey3 (strconv.Itoa)      142.6     96        2          🥈 2nd
LayoutCacheKey5 (byte append)       162.4     48        1          🥉 3rd
LayoutCacheKey4 (byte copy)         172.8     48        1          4th
LayoutCacheKey2 (fmt.Sprintf)       193.0    112        3          ❌ Slowest

Detailed Analysis (average of 5 runs):
--------------------------------------
🥇 LayoutCacheKey (strings.Builder with Grow) - CURRENT IMPLEMENTATION
  - Time:   137.4 ns/op
  - Memory: 96 B/op
  - Allocs: 2 allocs/op
  - Why it wins: strings.Builder with pre-allocation (Grow) is highly optimized
  - Minimal allocations, excellent cache locality

🥈 LayoutCacheKey3 (strconv.Itoa + string concat)
  - Time:   142.6 ns/op (+3.8% slower)
  - Memory: 96 B/op (same)
  - Allocs: 2 allocs/op (same)
  - Very close to the winner, string concatenation is well-optimized

🥉 LayoutCacheKey5 (byte array with append)
  - Time:   162.4 ns/op (+18.2% slower)
  - Memory: 48 B/op (50% LESS memory)
  - Allocs: 1 allocs/op (50% fewer allocations)
  - Trade-off: Uses less memory but slower due to rune->string conversions

LayoutCacheKey4 (byte array with copy)
  - Time:   172.8 ns/op (+25.8% slower)
  - Memory: 48 B/op (50% less memory)
  - Allocs: 1 allocs/op (50% fewer allocations)
  - Slower due to copy overhead and rune->string conversions

❌ LayoutCacheKey2 (fmt.Sprintf)
  - Time:   193.0 ns/op (+40.5% SLOWER)
  - Memory: 112 B/op (16.7% MORE memory)
  - Allocs: 3 allocs/op (50% more allocations)
  - fmt.Sprintf has significant overhead for reflection and formatting

Key Insights:
-------------
1. The CURRENT implementation (strings.Builder) is ALREADY OPTIMAL
2. Byte array approaches use less memory (48B vs 96B) but are 18-26% slower
3. The slowdown is due to rune→string conversions in the loop
4. For this hot path (called on every Score()), speed > memory savings
5. strings.Builder strikes the perfect balance

Recommendation:
--------------
✅ KEEP the current strings.Builder implementation (layoutCacheKey)
   - Fastest execution time
   - Reasonable memory usage
   - Minimal allocations
   - Already production-ready


SCORE CACHING BENCHMARKS:
------------------------
Run: go test -bench=BenchmarkScore -benchmem ./internal/keycraft

Actual Results (Apple M3, darwin/arm64):
BenchmarkScoreFirstCall-8                436828     2597 ns/op    3752 B/op     35 allocs/op
BenchmarkScoreCachedCall-8              8098814      148.8 ns/op    96 B/op      2 allocs/op
BenchmarkScoreCacheVsUncached/Uncached  469886     2646 ns/op    3752 B/op     35 allocs/op
BenchmarkScoreCacheVsUncached/Cached    8139982      148.6 ns/op    96 B/op      2 allocs/op

Cache Performance Analysis:
---------------------------
The cached version is dramatically faster:
  ✓ 94% faster (148.8 ns vs 2597 ns) - 17.5x speedup!
  ✓ 97% less memory (96 B vs 3752 B)
  ✓ 94% fewer allocations (2 vs 35)

Why the cache is so effective:
  - Uncached: Creates full Analyser, computes all metrics, normalizes, weights
  - Cached: Simple map lookup + cache key generation (strings.Builder)
  - Cache key generation (96 B, 2 allocs) is the only overhead on cache hit
  - This demonstrates why caching is critical for iterative layout optimization


TESTS INCLUDED:
--------------
1. TestLayoutCacheKey - Verifies cache key generation for all implementations
2. TestLayoutCacheKeyConsistency - Ensures deterministic cache key generation
3. TestScoreCache - Validates Score() caches results correctly
4. TestScoreCacheUniqueness - Ensures different layouts get different cache entries
5. TestScoreCacheIgnoresName - Verifies cache uses configuration, not layout name
*/
