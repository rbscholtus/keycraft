package keycraft

import (
	"testing"
)

// loadBenchmarkCorpus loads a corpus for benchmarking.
// Uses a cached corpus to avoid repeated loading overhead.
var benchCorpus *Corpus

func getBenchmarkCorpus(b *testing.B) *Corpus {
	if benchCorpus == nil {
		var err error
		benchCorpus, err = LoadJSON("../../data/corpus/default.txt.json")
		if err != nil {
			b.Fatalf("Failed to load corpus: %v", err)
		}
	}
	return benchCorpus
}

func BenchmarkTopUnigrams(b *testing.B) {
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
				_ = corpus.TopUnigrams(bm.nrows)
			}
		})
	}
}

func BenchmarkTopBigrams(b *testing.B) {
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
				_ = corpus.TopBigrams(bm.nrows)
			}
		})
	}
}

func BenchmarkTopConsonantBigrams(b *testing.B) {
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
				_, _ = corpus.TopConsonantBigrams(bm.nrows)
			}
		})
	}
}

func BenchmarkTopTrigrams(b *testing.B) {
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
				_ = corpus.TopTrigrams(bm.nrows)
			}
		})
	}
}

func BenchmarkTopSkipgrams(b *testing.B) {
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
				_ = corpus.TopSkipgrams(bm.nrows)
			}
		})
	}
}

func BenchmarkTopWords(b *testing.B) {
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
				_ = corpus.TopWords(bm.nrows)
			}
		})
	}
}
