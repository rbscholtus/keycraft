package tui

import (
	"testing"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// loadBenchmarkCorpus loads a corpus for benchmarking.
// Uses a cached corpus to avoid repeated loading overhead.
var benchCorpus *kc.Corpus

func getBenchmarkCorpus(b *testing.B) *kc.Corpus {
	if benchCorpus == nil {
		var err error
		// Load from JSON cache directly for faster benchmarks
		benchCorpus, err = kc.LoadJSON("../../data/corpus/default.txt.json")
		if err != nil {
			b.Fatalf("Failed to load corpus: %v", err)
		}
	}
	return benchCorpus
}

func BenchmarkCorpusWordLenDistStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)

	for b.Loop() {
		_ = corpusWordLenDistStr(corpus)
	}
}

func BenchmarkCorpusUnigramsStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusUnigramsStr(corpus, bm.nrows)
			}
		})
	}
}

func BenchmarkCorpusBigramsStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusBigramsStr(corpus, bm.nrows)
			}
		})
	}
}

func BenchmarkCorpusBigramConsonStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusBigramConsonStr(corpus, bm.nrows)
			}
		})
	}
}

func BenchmarkCorpusTrigramsStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusTrigramsStr(corpus, bm.nrows)
			}
		})
	}
}

func BenchmarkCorpusSkipgramsStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusSkipgramsStr(corpus, bm.nrows)
			}
		})
	}
}

func BenchmarkCorpusWordsStr(b *testing.B) {
	corpus := getBenchmarkCorpus(b)
	benchmarks := []struct {
		name  string
		nrows int
	}{
		{"10rows", 10},
		{"50rows", 50},
		{"100rows", 100},
		{"500rows", 500},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_ = corpusWordsStr(corpus, bm.nrows)
			}
		})
	}
}
