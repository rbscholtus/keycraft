package keycraft

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// layoutCacheKey generates a unique cache key based on LayoutType and Runes.
// This ensures cache hits for layouts with the same configuration regardless of their name.
func layoutCacheKey(layout *SplitLayout) string {
	var b strings.Builder
	b.Grow(1 + 42) // layoutType byte + 42 runes
	b.WriteByte(byte(layout.LayoutType))
	b.WriteString(string(layout.Runes[:]))
	return b.String()
}

// Scorer evaluates keyboard layouts by comparing their metrics against reference statistics.
// It uses robust normalization (median and IQR) to score layouts based on weighted metrics.
// Scores are cached to avoid redundant calculations for the same layout.
// Statistics are tracked to monitor cache effectiveness.
// Thread-safe for concurrent scoring operations.
type Scorer struct {
	corpus            *Corpus            // Text corpus used for analysis
	targets           *TargetLoads       // Target load distributions and penalty weights
	medians           map[string]float64 // Median values for each metric (filtered)
	iqrs              map[string]float64 // Interquartile ranges for each metric (filtered)
	weights           map[string]float64 // Importance weights for each metric (filtered)
	scoreCache        map[string]float64 // Cache of computed scores by layout identifier
	cacheMu           sync.RWMutex       // Protects scoreCache for concurrent access
	DisableScoreCache bool               // If true, skip score cache lookup/storage

	// Pre-filtered n-gram caches (computed lazily on first Score() call)
	trigramCache      []TrigramInfo // Pre-filtered trigrams with KeyInfo lookups
	trigramCacheOnce  sync.Once     // Ensures trigram cache is initialized exactly once
	DisableNGramCache bool          // If true, don't inject n-gram caches into Analyser

	// Statistics tracking (atomic for thread safety)
	cacheHits   atomic.Int64 // Number of cache hits
	cacheMisses atomic.Int64 // Number of cache misses
}

// NewScorer creates a new Scorer by analyzing reference layouts from the given directory.
// It computes median and IQR statistics from the reference layouts and filters out metrics
// with insignificant variance or weight to ensure robust scoring.
func NewScorer(layoutsDir string, corpus *Corpus, targets *TargetLoads, weights *Weights) (*Scorer, error) {
	analysers, err := LoadAnalysers(layoutsDir, corpus, targets)
	if err != nil {
		return nil, fmt.Errorf("could not load analysers: %w", err)
	}
	medians, iqrs := computeMediansAndIQR(analysers)

	// Filter out metrics with insignificant IQR values or weights
	numMetrics := len(medians)
	filteredMedians := make(map[string]float64, numMetrics)
	filteredIQRs := make(map[string]float64, numMetrics)
	filteredWeights := make(map[string]float64, numMetrics)
	for metric, median := range medians {
		const epsilon = 1e-9
		iqr, iqrExists := iqrs[metric]
		if !iqrExists || iqr <= epsilon {
			continue
		}
		weight := weights.Get(metric)
		if math.Abs(weight) <= 0.01 {
			// Often, tiny weights are assigned to have them in the Weights struct, but not
			// to actually count towards anything. So, ignore tiny weights.
			continue
		}
		filteredMedians[metric] = median
		filteredIQRs[metric] = iqr
		filteredWeights[metric] = weight
	}

	sc := &Scorer{
		corpus:     corpus,
		targets:    targets,
		medians:    filteredMedians,
		iqrs:       filteredIQRs,
		weights:    filteredWeights,
		scoreCache: make(map[string]float64, 1000),
	}

	return sc, nil
}

// prepareTrigramCache pre-filters corpus trigrams using a template layout and applies
// 99% coverage filtering to keep only high-frequency trigrams.
// This eliminates redundant filtering across all future Score() calls and reduces cache size
// by discarding low-frequency trigrams that contribute minimally to the analysis.
// The cache contains only trigrams where all 3 runes exist on the template layout.
// KeyInfo is looked up fresh during analysis to ensure correctness when layouts change.
func (sc *Scorer) prepareTrigramCache(templateLayout *SplitLayout) {
	// Step 1: Filter by layout (keep only trigrams where all runes exist on layout)
	layoutFiltered := make([]TrigramInfo, 0, len(sc.corpus.Trigrams)/10)
	var totalCount uint64

	for tri, cnt := range sc.corpus.Trigrams {
		_, ok0 := templateLayout.GetKeyInfo(tri[0])
		_, ok1 := templateLayout.GetKeyInfo(tri[1])
		_, ok2 := templateLayout.GetKeyInfo(tri[2])

		if ok0 && ok1 && ok2 {
			layoutFiltered = append(layoutFiltered, TrigramInfo{
				Count: cnt,
				Runes: [3]rune{tri[0], tri[1], tri[2]},
			})
			totalCount += cnt
		}
	}

	// Step 2: Sort by frequency (descending) to prioritize high-frequency trigrams
	sort.Slice(layoutFiltered, func(i, j int) bool {
		return layoutFiltered[i].Count > layoutFiltered[j].Count
	})

	// Step 3: Apply 99% coverage threshold - keep only trigrams accounting for 99% of occurrences
	targetCount := uint64(float64(totalCount) * 0.98)
	var cumulative uint64
	cutoffIndex := 0

	for i, ti := range layoutFiltered {
		cumulative += ti.Count
		if cumulative >= targetCount {
			cutoffIndex = i + 1
			break
		}
	}

	// Step 4: Keep only trigrams that meet 99% coverage threshold
	sc.trigramCache = layoutFiltered[:cutoffIndex]
}

// Score evaluates a layout by computing a weighted sum of normalized metrics.
// Each metric is normalized using robust scaling: (value - median) / IQR.
// Only metrics with non-zero weights and sufficient variance are scored.
// The weighted sum is subtracted to produce a cost score where lower is better.
// Results are cached by layout configuration to avoid redundant calculations (unless DisableScoreCache is true).
// Thread-safe for concurrent access.
func (sc *Scorer) Score(layout *SplitLayout) float64 {
	// Initialize n-gram caches lazily on first call (unless disabled)
	if !sc.DisableNGramCache {
		sc.trigramCacheOnce.Do(func() {
			sc.prepareTrigramCache(layout)
		})
	}

	// Check score cache first (unless disabled)
	var cacheKey string
	if !sc.DisableScoreCache {
		cacheKey = layoutCacheKey(layout)

		// Read lock for cache check
		sc.cacheMu.RLock()
		cachedScore, exists := sc.scoreCache[cacheKey]
		sc.cacheMu.RUnlock()

		if exists {
			sc.cacheHits.Add(1)
			return cachedScore
		}
		sc.cacheMisses.Add(1)
	}

	// Calculate score
	an := &Analyser{
		Layout:           layout,
		Corpus:           sc.corpus,
		Targets:          sc.targets,
		Metrics:          make(map[string]float64, 60),
		relevantTrigrams: sc.trigramCache, // Inject pre-filtered trigrams for performance optimization
	}

	an.analyseHand()
	an.analyseBigrams()
	an.analyseSkipgrams()
	an.analyseTrigrams()

	score := 0.0
	for metric, iqr := range sc.iqrs {
		if value, exists := an.Metrics[metric]; exists {
			scaledValue := (value - sc.medians[metric]) / iqr
			score -= sc.weights[metric] * scaledValue
		}
	}

	// Update cache (unless disabled)
	if !sc.DisableScoreCache {
		sc.cacheMu.Lock()
		sc.scoreCache[cacheKey] = score
		sc.cacheMu.Unlock()
	}

	return score
}

// ScorerStats holds statistics about Scorer performance.
type ScorerStats struct {
	TotalCalls     int64   // Total number of Score() calls
	CacheHits      int64   // Number of cache hits
	CacheMisses    int64   // Number of cache misses
	HitRate        float64 // Cache hit rate as percentage (0-100)
	UniqueLayouts  int     // Number of unique layouts cached
	CacheSizeBytes int     // Estimated cache memory usage in bytes
}

// GetStats returns current statistics about the Scorer's performance.
// Thread-safe for concurrent access.
func (sc *Scorer) GetStats() ScorerStats {
	hits := sc.cacheHits.Load()
	misses := sc.cacheMisses.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = (float64(hits) / float64(total)) * 100.0
	}

	// Read lock to get cache size
	sc.cacheMu.RLock()
	cacheLen := len(sc.scoreCache)
	sc.cacheMu.RUnlock()

	// Estimate cache memory usage
	// Each entry: ~44 bytes for key (layout type + 42 runes) + 8 bytes for float64 value
	// Plus map overhead (~48 bytes per entry in Go)
	avgEntrySize := 100 // Conservative estimate
	cacheSize := cacheLen * avgEntrySize

	return ScorerStats{
		TotalCalls:     total,
		CacheHits:      hits,
		CacheMisses:    misses,
		HitRate:        hitRate,
		UniqueLayouts:  cacheLen,
		CacheSizeBytes: cacheSize,
	}
}

// LogStats writes Scorer statistics to the provided writer.
// Typically called at the end of an optimization run to understand cache effectiveness.
func (sc *Scorer) LogStats(w io.Writer) {
	stats := sc.GetStats()

	MustFprintf(w, "\n")
	MustFprintf(w, "Scorer Statistics:\n")
	MustFprintf(w, "==================\n")
	MustFprintf(w, "Total Score() calls:     %s\n", formatInt(stats.TotalCalls))
	MustFprintf(w, "Cache hits:              %s (%.1f%%)\n", formatInt(stats.CacheHits), stats.HitRate)
	MustFprintf(w, "Cache misses:            %s (%.1f%%)\n", formatInt(stats.CacheMisses), 100.0-stats.HitRate)
	MustFprintf(w, "Unique layouts cached:   %s\n", formatInt(int64(stats.UniqueLayouts)))
	MustFprintf(w, "Cache memory usage:      ~%s\n", formatBytes(stats.CacheSizeBytes))
	MustFprintf(w, "\n")
}

// formatInt formats an integer with thousand separators for readability.
func formatInt(n int64) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return formatInt(n/1000) + "," + fmt.Sprintf("%03d", n%1000)
}

// formatBytes formats bytes into human-readable format (B, KB, MB).
func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	kb := float64(bytes) / 1024.0
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024.0
	return fmt.Sprintf("%.2f MB", mb)
}

// LayoutScore represents a layout name, its computed score and the analyser providing metrics.
type LayoutScore struct {
	Name     string    // Layout identifier or filename.
	Score    float64   // Weighted score for ranking.
	Analyser *Analyser // Analyser with detailed metric values.
}

// LoadAnalysers loads and analyses all .klf layout files from a directory in parallel.
// Only loads reference layouts, excluding files that start with "_" or contain "-flipped", "-best", or "-opt".
// Uses bounded concurrency based on GOMAXPROCS to avoid overloading the system.
func LoadAnalysers(layoutsDir string, corpus *Corpus, targets *TargetLoads) ([]*Analyser, error) {
	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading layout files from %v: %w", layoutsDir, err)
	}

	var (
		analysers = make([]*Analyser, 0, len(layoutFiles)) // Pre-allocate
		mu        sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, runtime.GOMAXPROCS(0)) // Semaphore to limit concurrent goroutines
		errs      = make(chan error, len(layoutFiles))
	)

	for _, file := range layoutFiles {
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".klf") {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			layoutName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			layoutPath := filepath.Join(layoutsDir, f.Name())
			layout, err := NewLayoutFromFile(layoutName, layoutPath)
			if err != nil {
				errs <- fmt.Errorf("could not load layout from file %s: %w", layoutPath, err)
				return
			}
			analyser := NewAnalyser(layout, corpus, targets)

			mu.Lock()
			analysers = append(analysers, analyser)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	close(errs)

	var allErrors []error
	for err := range errs {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return nil, fmt.Errorf("could not load all analysers: %v", allErrors)
	}

	return analysers, nil
}

// computeMediansAndIQR computes median and interquartile range (IQR) for each metric
// across all analysers. These values are used for robust normalization of layout scores.
// Only uses reference layouts for normalization, excluding layouts that start with "_" or
// contain "-flipped", "-best", or "-opt" in their name.
func computeMediansAndIQR(analysers []*Analyser) (map[string]float64, map[string]float64) {
	metrics := make(map[string][]float64)
	for _, analyser := range analysers {
		layoutName := analyser.Layout.Name

		// Skip non-reference layouts for normalization statistics
		if strings.HasPrefix(layoutName, "_") ||
			strings.Contains(layoutName, "-flipped") ||
			strings.Contains(layoutName, "-best") ||
			strings.Contains(layoutName, "-opt") {
			continue
		}

		for metric, value := range analyser.Metrics {
			metrics[metric] = append(metrics[metric], value)
		}
	}

	medians := make(map[string]float64)
	iqr := make(map[string]float64)
	for metric, values := range metrics {
		sort.Float64s(values)
		medians[metric] = Median(values)
		q1, q3 := Quartiles(values)
		iqr[metric] = q3 - q1
	}

	return medians, iqr
}

// computeScores normalizes metrics using median and IQR, then computes weighted layout scores.
// Only metrics with non-zero weights are included in the final score calculation.
func computeScores(analysers []*Analyser, medians, iqr map[string]float64, weights *Weights) []LayoutScore {
	var layoutScores []LayoutScore

	for _, analyser := range analysers {
		score := 0.0
		for metric, value := range analyser.Metrics {
			// Skip metrics with zero IQR (all values identical)
			if iqr[metric] == 0 {
				continue
			}
			weight := weights.Get(metric)
			if weight == 0 {
				continue
			}
			// Apply robust normalization and weight
			scaledValue := (value - medians[metric]) / iqr[metric]
			score += weight * scaledValue
		}
		layoutScores = append(layoutScores, LayoutScore{
			Name:     analyser.Layout.Name,
			Score:    score,
			Analyser: analyser,
		})
	}

	return layoutScores
}

// Median calculates the median of a sorted slice.
// The slice must already be sorted in ascending order.
func Median(sortedData []float64) float64 {
	n := len(sortedData)
	mid := n / 2
	if n%2 == 0 {
		return (sortedData[mid-1] + sortedData[mid]) / 2.0
	} else {
		return sortedData[mid]
	}
}

// Quartiles calculates the first and third quartiles (Q1 and Q3) of a sorted slice.
// The slice must already be sorted in ascending order.
func Quartiles(sortedData []float64) (float64, float64) {
	n := len(sortedData)
	q1 := Median(sortedData[:n/2])
	q3 := Median(sortedData[(n+1)/2:])
	return q1, q3
}

// RobustScale applies robust scaling to the data using median and interquartile range (IQR).
// This scaling method is less sensitive to outliers than standard normalization.
// Each value is transformed to: (value - median) / IQR
func RobustScale(data []float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	// Create a sorted copy for computing statistics
	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	medianValue := Median(sortedData)
	q1, q3 := Quartiles(sortedData)
	iqr := q3 - q1

	// If all values are identical, return zeros
	if iqr == 0 {
		return make([]float64, len(data))
	}

	// Apply robust scaling transformation
	scaledData := make([]float64, len(data))
	for i, x := range data {
		scaledData[i] = (x - medianValue) / iqr
	}

	return scaledData
}
