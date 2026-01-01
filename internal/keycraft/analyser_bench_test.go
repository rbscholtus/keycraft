package keycraft

import (
	"testing"
)

// BenchmarkAnalyser benchmarks the full Analyser creation and metric computation.
// This helps identify performance bottlenecks in layout analysis.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyser -benchmem -cpuprofile=cpu.prof ./internal/keycraft
//	go tool pprof -http=:8080 cpu.prof
func BenchmarkAnalyser(b *testing.B) {
	// Load corpus once (not part of benchmark)
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	// Load layout once (not part of benchmark)
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	// Use default ideal loads
	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	// Benchmark the analyser creation and metric computation
	for b.Loop() {
		_ = NewAnalyser(layout, corpus, idealRowLoad, idealFingerLoad, pinkyWeights)
	}
}

// BenchmarkAnalyserBigrams benchmarks only bigram analysis.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserBigrams -benchmem -cpuprofile=cpu.prof ./internal/keycraft
func BenchmarkAnalyserBigrams(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	for b.Loop() {
		a := &Analyser{
			Layout:  layout,
			Corpus:  corpus,
			Metrics: make(map[string]float64),
		}
		a.analyseBigrams()
	}
}

// BenchmarkAnalyserSkipgrams benchmarks only skipgram analysis.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserSkipgrams -benchmem -cpuprofile=cpu.prof ./internal/keycraft
func BenchmarkAnalyserSkipgrams(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	for b.Loop() {
		a := &Analyser{
			Layout:  layout,
			Corpus:  corpus,
			Metrics: make(map[string]float64),
		}
		a.analyseSkipgrams()
	}
}

// BenchmarkAnalyserTrigrams benchmarks only trigram analysis.
// This is typically the most expensive operation.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserTrigrams -benchmem -cpuprofile=cpu.prof ./internal/keycraft
func BenchmarkAnalyserTrigrams(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	for b.Loop() {
		a := &Analyser{
			Layout:  layout,
			Corpus:  corpus,
			Metrics: make(map[string]float64),
		}
		a.analyseTrigrams()
	}
}

// BenchmarkAnalyserWithScorer benchmarks Analyser performance when using Scorer (real-world scenario).
// This shows the benefit of pre-filtered trigrams cached at the Scorer level.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserWithScorer -benchmem ./internal/keycraft
func BenchmarkAnalyserWithScorer(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	// Load proper weights from default file
	weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	if err != nil {
		b.Fatalf("Failed to load weights: %v", err)
	}

	// Create Scorer (pre-filters trigrams once)
	scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, pinkyWeights, weights)
	if err != nil {
		b.Fatalf("Failed to create scorer: %v", err)
	}

	// Disable score cache to benchmark raw analysis performance
	scorer.DisableScoreCache = true

	for b.Loop() {
		// Score the layout (score cache disabled, trigram cache enabled)
		_ = scorer.Score(layout)
	}
}

// BenchmarkAnalyserHand benchmarks hand/finger load analysis.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserHand -benchmem ./internal/keycraft
func BenchmarkAnalyserHand(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	for b.Loop() {
		a := &Analyser{
			Layout:       layout,
			Corpus:       corpus,
			IdealRowLoad: idealRowLoad,
			IdealfgrLoad: idealFingerLoad,
			PinkyWeights: pinkyWeights,
			Metrics:      make(map[string]float64),
		}
		a.analyseHand()
	}
}

// BenchmarkAnalyserBigramsWithCache benchmarks bigram analysis with pre-filtered cache.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserBigramsWithCache -benchmem ./internal/keycraft
func BenchmarkAnalyserBigramsWithCache(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	// idealRowLoad := DefaultIdealRowLoad()
	// idealFingerLoad := DefaultIdealFingerLoad()

	// weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	// if err != nil {
	// 	b.Fatalf("Failed to load weights: %v", err)
	// }

	// scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, weights)
	// if err != nil {
	// 	b.Fatalf("Failed to create scorer: %v", err)
	// }

	// Initialize the bigram cache
	// scorer.prepareBigramCache(layout)

	for b.Loop() {
		a := &Analyser{
			Layout:  layout,
			Corpus:  corpus,
			Metrics: make(map[string]float64),
			// relevantBigrams: scorer.bigramCache,
		}
		a.analyseBigrams()
	}
}

// BenchmarkAnalyserSkipgramsWithCache benchmarks skipgram analysis with pre-filtered cache.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserSkipgramsWithCache -benchmem ./internal/keycraft
func BenchmarkAnalyserSkipgramsWithCache(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	// idealRowLoad := DefaultIdealRowLoad()
	// idealFingerLoad := DefaultIdealFingerLoad()

	// weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	// if err != nil {
	// 	b.Fatalf("Failed to load weights: %v", err)
	// }

	// scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, weights)
	// if err != nil {
	// 	b.Fatalf("Failed to create scorer: %v", err)
	// }

	// Initialize the skipgram cache
	// scorer.prepareSkipgramCache(layout)

	for b.Loop() {
		a := &Analyser{
			Layout:  layout,
			Corpus:  corpus,
			Metrics: make(map[string]float64),
			// relevantSkipgrams: scorer.skipgramCache,
		}
		a.analyseSkipgrams()
	}
}

// BenchmarkAnalyserTrigramsWithCache benchmarks trigram analysis with pre-filtered cache.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserTrigramsWithCache -benchmem ./internal/keycraft
func BenchmarkAnalyserTrigramsWithCache(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	if err != nil {
		b.Fatalf("Failed to load weights: %v", err)
	}

	scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, pinkyWeights, weights)
	if err != nil {
		b.Fatalf("Failed to create scorer: %v", err)
	}

	// Initialize the trigram cache
	scorer.prepareTrigramCache(layout)

	for b.Loop() {
		a := &Analyser{
			Layout:           layout,
			Corpus:           corpus,
			Metrics:          make(map[string]float64),
			relevantTrigrams: scorer.trigramCache,
		}
		a.analyseTrigrams()
	}
}

// BenchmarkAnalyserWithScorerAllCached benchmarks full analysis with all n-gram caches enabled.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserWithScorerAllCached -benchmem ./internal/keycraft
func BenchmarkAnalyserWithScorerAllCached(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	if err != nil {
		b.Fatalf("Failed to load weights: %v", err)
	}

	scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, pinkyWeights, weights)
	if err != nil {
		b.Fatalf("Failed to create scorer: %v", err)
	}

	// Disable score cache to benchmark raw analysis performance
	scorer.DisableScoreCache = true
	// N-gram caches enabled by default

	for b.Loop() {
		_ = scorer.Score(layout)
	}
}

// BenchmarkAnalyserWithScorerNoCaches benchmarks full analysis with all n-gram caches disabled.
//
// Run with:
//
//	go test -bench=BenchmarkAnalyserWithScorerNoCaches -benchmem ./internal/keycraft
func BenchmarkAnalyserWithScorerNoCaches(b *testing.B) {
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		b.Fatalf("Failed to load corpus: %v", err)
	}

	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		b.Fatalf("Failed to load layout: %v", err)
	}

	idealRowLoad := DefaultIdealRowLoad()
	idealFingerLoad := DefaultIdealFingerLoad()
	pinkyWeights := DefaultPinkyWeights()

	weights, err := NewWeightsFromParams("../../data/weights/default.txt", "")
	if err != nil {
		b.Fatalf("Failed to load weights: %v", err)
	}

	scorer, err := NewScorer("../../data/layouts", corpus, idealRowLoad, idealFingerLoad, pinkyWeights, weights)
	if err != nil {
		b.Fatalf("Failed to create scorer: %v", err)
	}

	// Disable both score cache and n-gram caches
	scorer.DisableScoreCache = true
	scorer.DisableNGramCache = true

	for b.Loop() {
		_ = scorer.Score(layout)
	}
}
