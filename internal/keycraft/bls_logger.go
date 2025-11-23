package keycraft

import (
	"encoding/json"
	"io"
	"time"
)

// BLSLogger provides dual-format logging for BLS optimization.
// Console output is human-readable, file output is JSONL for analysis.
type BLSLogger struct {
	console   io.Writer // Human-readable output (can be nil)
	file      io.Writer // JSONL structured output (can be nil)
	startTime time.Time
}

// NewBLSLogger creates a new logger with separate console and file outputs.
// Either writer can be nil to disable that output channel.
func NewBLSLogger(console, file io.Writer) *BLSLogger {
	return &BLSLogger{
		console:   console,
		file:      file,
		startTime: time.Now(),
	}
}

// LogEvent represents a single log entry in JSONL format.
type LogEvent struct {
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
	ElapsedMs int64     `json:"elapsed_ms"`

	// Optimization state (present in most events)
	Iteration *int     `json:"iteration,omitempty"`
	Cost      *float64 `json:"cost,omitempty"`
	BestCost  *float64 `json:"best_cost,omitempty"`
	Delta     *float64 `json:"delta,omitempty"`

	// BLS-specific state
	JumpMagnitude *int `json:"jump_magnitude,omitempty"` // L value
	Omega         *int `json:"omega,omitempty"`          // Stagnation counter

	// Perturbation info
	PerturbationType string `json:"perturbation_type,omitempty"`

	// Layout info (for start/end events)
	LayoutName string   `json:"layout_name,omitempty"`
	FreeKeys   *int     `json:"free_keys,omitempty"`
	TotalKeys  *int     `json:"total_keys,omitempty"`
	Layout     []string `json:"layout,omitempty"` // Layout rows as strings

	// Parameters (for start event)
	Params *BLSLogParams `json:"params,omitempty"`

	// Cache statistics (for end event)
	CacheStats *CacheStatsLog `json:"cache_stats,omitempty"`

	// Message for generic events
	Message string `json:"message,omitempty"`

	// Descent info (for descent events)
	SwapCount *int     `json:"swap_count,omitempty"`
	StartCost *float64 `json:"start_cost,omitempty"`
	EndCost   *float64 `json:"end_cost,omitempty"`

	// Perturbation info (for perturb events)
	PerturbStrategies map[string]int `json:"perturb_strategies,omitempty"` // Count of each strategy used
	PerturbSwaps      *int           `json:"perturb_swaps,omitempty"`      // Total swaps applied
}

// BLSLogParams captures BLS parameters for the start event.
type BLSLogParams struct {
	L0            int     `json:"l0"`
	LMax          int     `json:"l_max"`
	T             int     `json:"t"`
	TabuMin       int     `json:"tabu_min"`
	TabuMax       int     `json:"tabu_max"`
	MaxIterations int     `json:"max_iterations"`
	MaxTimeMs     int64   `json:"max_time_ms"`
	Seed          int64   `json:"seed"`
	UseParallel   bool    `json:"use_parallel"`
	Workers       int     `json:"workers"`
	PatternWeight float64 `json:"pattern_weight"`
	ColumnWeight  float64 `json:"column_weight"`
	RandomWeight  float64 `json:"random_weight"`
	RecencyWeight float64 `json:"recency_weight"`
}

// CacheStatsLog captures cache statistics for the end event.
type CacheStatsLog struct {
	Hits        uint64  `json:"hits"`
	Misses      uint64  `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	UniqueKeys  int     `json:"unique_keys"`
	MemoryBytes int64   `json:"memory_bytes"`
}

// writeJSON writes a log event to the file output as JSONL.
func (l *BLSLogger) writeJSON(event LogEvent) {
	if l.file == nil {
		return
	}

	event.Timestamp = time.Now()
	event.ElapsedMs = time.Since(l.startTime).Milliseconds()

	data, err := json.Marshal(event)
	if err != nil {
		return // Silently ignore JSON errors
	}

	data = append(data, '\n')
	_, _ = l.file.Write(data)
}

// LogStart logs the start of optimization.
func (l *BLSLogger) LogStart(params BLSParams, layout *SplitLayout, numFree int) {
	if l.console != nil {
		MustFprintf(l.console, "Starting BLS optimization\n")
		MustFprintf(l.console, "Initial cost: calculating...\n")
		MustFprintf(l.console, "Free keys: %d/%d\n\n", numFree, 42)
		MustFprintln(l.console, layout)
	}

	totalKeys := 42
	l.writeJSON(LogEvent{
		Event:      "start",
		LayoutName: layout.Name,
		FreeKeys:   &numFree,
		TotalKeys:  &totalKeys,
		Layout:     layoutToStrings(layout),
		Params: &BLSLogParams{
			L0:            params.L0,
			LMax:          params.LMax,
			T:             params.T,
			TabuMin:       params.TabuMin,
			TabuMax:       params.TabuMax,
			MaxIterations: params.MaxIterations,
			MaxTimeMs:     params.MaxTime.Milliseconds(),
			Seed:          params.Seed,
			UseParallel:   params.UseParallel,
			Workers:       params.ParallelWorkers,
			PatternWeight: params.PatternWeight,
			ColumnWeight:  params.ColumnWeight,
			RandomWeight:  params.RandomWeight,
			RecencyWeight: params.RecencyWeight,
		},
	})
}

// LogInitialCost logs the initial cost after it's calculated.
func (l *BLSLogger) LogInitialCost(cost float64) {
	if l.console != nil {
		MustFprintf(l.console, "Initial cost: %.4f\n", cost)
	}

	l.writeJSON(LogEvent{
		Event: "initial_cost",
		Cost:  &cost,
	})
}

// LogImprovement logs when a new best solution is found.
func (l *BLSLogger) LogImprovement(iteration int, newCost, prevBest float64, layout *SplitLayout, elapsed time.Duration) {
	delta := newCost - prevBest

	if l.console != nil {
		MustFprintf(l.console, "Iter %d: New best cost: %.4f (elapsed: %v)\n",
			iteration, newCost, elapsed.Round(time.Second))
		MustFprintln(l.console, layout)
	}

	l.writeJSON(LogEvent{
		Event:      "improvement",
		Iteration:  &iteration,
		Cost:       &newCost,
		BestCost:   &newCost,
		Delta:      &delta,
		LayoutName: layout.Name,
		Layout:     layoutToStrings(layout),
	})
}

// LogStrongPerturbation logs when strong diversification is triggered.
func (l *BLSLogger) LogStrongPerturbation(iteration, jumpMagnitude int) {
	if l.console != nil {
		MustFprintf(l.console, "Iter %d: Strong perturbation triggered (L=%d)\n",
			iteration, jumpMagnitude)
	}

	l.writeJSON(LogEvent{
		Event:         "strong_perturbation",
		Iteration:     &iteration,
		JumpMagnitude: &jumpMagnitude,
	})
}

// LogProgress logs periodic progress updates.
func (l *BLSLogger) LogProgress(iteration int, currentCost, bestCost float64, jumpMagnitude, omega int) {
	if l.console != nil {
		MustFprintf(l.console, "Iter %d: Current: %.4f, Best: %.4f, L=%d, omega=%d\n",
			iteration, currentCost, bestCost, jumpMagnitude, omega)
	}

	l.writeJSON(LogEvent{
		Event:         "progress",
		Iteration:     &iteration,
		Cost:          &currentCost,
		BestCost:      &bestCost,
		JumpMagnitude: &jumpMagnitude,
		Omega:         &omega,
	})
}

// LogTimeLimit logs when the time limit is reached.
func (l *BLSLogger) LogTimeLimit(elapsed time.Duration) {
	if l.console != nil {
		MustFprintf(l.console, "\nTime limit reached: %v\n", elapsed)
	}

	l.writeJSON(LogEvent{
		Event:   "time_limit",
		Message: elapsed.String(),
	})
}

// LogDescent logs the completion of a steepest descent phase.
func (l *BLSLogger) LogDescent(iteration int, swapCount int, startCost, endCost float64) {
	// Only log to file (console would be too verbose)
	l.writeJSON(LogEvent{
		Event:     "descent",
		Iteration: &iteration,
		SwapCount: &swapCount,
		StartCost: &startCost,
		EndCost:   &endCost,
	})
}

// LogPerturb logs the completion of a perturbation phase.
func (l *BLSLogger) LogPerturb(iteration int, strategies map[string]int, totalSwaps int, startCost, endCost float64) {
	// Only log to file (console would be too verbose)
	l.writeJSON(LogEvent{
		Event:             "perturb",
		Iteration:         &iteration,
		PerturbStrategies: strategies,
		PerturbSwaps:      &totalSwaps,
		StartCost:         &startCost,
		EndCost:           &endCost,
	})
}

// LogEnd logs the end of optimization.
func (l *BLSLogger) LogEnd(bestCost float64, totalIterations int, elapsed time.Duration, layout *SplitLayout) {
	if l.console != nil {
		MustFprintf(l.console, "\nOptimization complete\n")
		MustFprintf(l.console, "Final best cost: %.4f\n", bestCost)
		MustFprintf(l.console, "Total iterations: %d\n", totalIterations)
		MustFprintf(l.console, "Total time: %v\n", elapsed.Round(time.Second))
	}

	l.writeJSON(LogEvent{
		Event:      "end",
		Iteration:  &totalIterations,
		BestCost:   &bestCost,
		LayoutName: layout.Name,
		Layout:     layoutToStrings(layout),
	})
}

// LogCacheStats logs cache statistics (typically at end of optimization).
func (l *BLSLogger) LogCacheStats(hits, misses uint64, uniqueKeys int, memoryBytes int64) {
	hitRate := 0.0
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses)
	}

	// Console output is handled by Scorer.LogStats, so only write JSON here
	l.writeJSON(LogEvent{
		Event: "cache_stats",
		CacheStats: &CacheStatsLog{
			Hits:        hits,
			Misses:      misses,
			HitRate:     hitRate,
			UniqueKeys:  uniqueKeys,
			MemoryBytes: memoryBytes,
		},
	})
}

// HasConsole returns true if console output is enabled.
func (l *BLSLogger) HasConsole() bool {
	return l.console != nil
}

// HasFile returns true if file output is enabled.
func (l *BLSLogger) HasFile() bool {
	return l.file != nil
}

// Console returns the console writer (for backward compatibility).
func (l *BLSLogger) Console() io.Writer {
	return l.console
}

// layoutToStrings converts a layout to a slice of row strings for JSON output.
func layoutToStrings(layout *SplitLayout) []string {
	if layout == nil {
		return nil
	}

	rows := make([]string, 4)

	// Row 0: keys 0-11
	row0 := make([]rune, 12)
	for i := 0; i < 12; i++ {
		r := layout.Runes[i]
		if r == 0 {
			r = ' '
		}
		row0[i] = r
	}
	rows[0] = string(row0)

	// Row 1: keys 12-23
	row1 := make([]rune, 12)
	for i := 0; i < 12; i++ {
		r := layout.Runes[12+i]
		if r == 0 {
			r = ' '
		}
		row1[i] = r
	}
	rows[1] = string(row1)

	// Row 2: keys 24-35
	row2 := make([]rune, 12)
	for i := 0; i < 12; i++ {
		r := layout.Runes[24+i]
		if r == 0 {
			r = ' '
		}
		row2[i] = r
	}
	rows[2] = string(row2)

	// Row 3: keys 36-41 (thumb row)
	row3 := make([]rune, 6)
	for i := 0; i < 6; i++ {
		r := layout.Runes[36+i]
		if r == 0 {
			r = ' '
		}
		row3[i] = r
	}
	rows[3] = string(row3)

	return rows
}
