package keycraft

import (
	"path/filepath"
	"strings"
	"testing"
)

const (
	layoutDir = "../data/layouts/"
	corpusDir = "../data/corpus/"
	// weightsDir = "../data/weights/"
)

func Benchmark_Rankings(b *testing.B) {
	// get corpus
	filename := "default.txt"
	corpusName := strings.TrimSuffix(filename, filepath.Ext(filename))
	path := filepath.Join(corpusDir, filename)
	corpus, err := NewCorpusFromFile(corpusName, path, false, 98)
	if err != nil {
		panic(err)
	}

	// weightsPath := filepath.Join(weightsDir, "default.txt")
	// weights, err := kc.NewWeightsFromParams(weightsPath, "")
	// if err != nil {
	// 	panic(err)
	// }

	// Load all analysers for layouts in the directory
	for b.Loop() {
		analysers, err := LoadAnalysers(layoutDir, corpus, DefaultIdealRowLoad(), DefaultIdealFingerLoad(), DefaultPinkyWeights())
		if err != nil {
			panic(err)
		}
		if len(analysers) <= 0 {
			panic("no analysers returned")
		}

		// // Compute median and IQR for each metric for normalization
		// medians, iqrs := kc.ComputeMediansAndIQR(analysers)

		// // Compute weighted scores for filtered layouts
		// layoutScores := kc.ComputeScores(analysers, medians, iqrs, weights)

		// // Sort layouts by rank
		// sort.Slice(layoutScores, func(i, j int) bool {
		// 	return layoutScores[i].Score > layoutScores[j].Score
		// })
	}
}
