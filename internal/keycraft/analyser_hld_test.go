package keycraft

import (
	"math"
	"testing"
)

// TestHLD_Calculation tests the HLD calculation logic directly.
func TestHLD_Calculation(t *testing.T) {
	tests := []struct {
		name        string
		h0          float64 // left hand %
		h1          float64 // right hand %
		targetH0    float64
		targetH1    float64
		expectedHLD float64
	}{
		{
			name:        "Perfect 50/50 balance",
			h0:          50.0,
			h1:          50.0,
			targetH0:    50.0,
			targetH1:    50.0,
			expectedHLD: 0.0,
		},
		{
			name:        "70/30 split vs 50/50 target",
			h0:          70.0,
			h1:          30.0,
			targetH0:    50.0,
			targetH1:    50.0,
			expectedHLD: 40.0, // |70-50| + |30-50| = 20 + 20
		},
		{
			name:        "60/40 split matches 60/40 target",
			h0:          60.0,
			h1:          40.0,
			targetH0:    60.0,
			targetH1:    40.0,
			expectedHLD: 0.0,
		},
		{
			name:        "Extreme 100/0 vs 50/50 target",
			h0:          100.0,
			h1:          0.0,
			targetH0:    50.0,
			targetH1:    50.0,
			expectedHLD: 100.0, // |100-50| + |0-50| = 50 + 50
		},
		{
			name:        "55/45 vs 50/50 target",
			h0:          55.0,
			h1:          45.0,
			targetH0:    50.0,
			targetH1:    50.0,
			expectedHLD: 10.0, // |55-50| + |45-50| = 5 + 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate HLD using the formula
			hld := math.Abs(tt.h0-tt.targetH0) + math.Abs(tt.h1-tt.targetH1)

			if math.Abs(hld-tt.expectedHLD) > 0.01 {
				t.Errorf("Expected HLD=%.2f, got %.2f", tt.expectedHLD, hld)
			}
		})
	}
}

// TestHLD_WithRealLayout tests HLD with a real layout and corpus.
func TestHLD_WithRealLayout(t *testing.T) {
	// Load a real corpus
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		t.Skipf("Skipping test: could not load corpus: %v", err)
	}

	// Load a real layout
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		t.Skipf("Skipping test: could not load layout: %v", err)
	}

	// Test with default 50/50 targets
	targets := &TargetLoads{
		TargetHandLoad:   DefaultTargetHandLoad(), // 50/50
		TargetFingerLoad: DefaultTargetFingerLoad(),
		TargetRowLoad:    DefaultTargetRowLoad(),
		PinkyPenalties:   DefaultPinkyPenalties(),
	}

	analyser := NewAnalyser(layout, corpus, targets)

	// Verify HLD is calculated
	hld, exists := analyser.Metrics["HLD"]
	if !exists {
		t.Fatal("HLD metric not found")
	}

	// Verify HLD formula: |H0 - targetH0| + |H1 - targetH1|
	h0 := analyser.Metrics["H0"]
	h1 := analyser.Metrics["H1"]
	expectedHLD := math.Abs(h0-targets.TargetHandLoad[0]) + math.Abs(h1-targets.TargetHandLoad[1])

	if math.Abs(hld-expectedHLD) > 0.01 {
		t.Errorf("HLD calculation mismatch: got %.2f, expected %.2f (H0=%.2f, H1=%.2f)",
			hld, expectedHLD, h0, h1)
	}

	// HLD should be non-negative
	if hld < 0 {
		t.Errorf("HLD should be non-negative, got %.2f", hld)
	}

	// HLD should be <= 100 (max possible imbalance)
	if hld > 100 {
		t.Errorf("HLD should be <= 100, got %.2f", hld)
	}
}

// TestHLD_CustomPreference tests HLD with a custom hand load deviation target.
func TestHLD_CustomPreference(t *testing.T) {
	// Load a real corpus
	corpus, err := NewCorpusFromFile("default", "../../data/corpus/default.txt", false, 98.0)
	if err != nil {
		t.Skipf("Skipping test: could not load corpus: %v", err)
	}

	// Load a real layout
	layout, err := NewLayoutFromFile("qwerty", "../../data/layouts/qwerty.klf")
	if err != nil {
		t.Skipf("Skipping test: could not load layout: %v", err)
	}

	// Test with custom 60/40 target
	target := &TargetLoads{
		TargetRowLoad:    DefaultTargetRowLoad(),
		TargetFingerLoad: DefaultTargetFingerLoad(),
		TargetHandLoad:   &[2]float64{60.0, 40.0},
		PinkyPenalties:   DefaultPinkyPenalties(),
	}

	analyser := NewAnalyser(layout, corpus, target)

	// Verify HLD formula with custom preference
	h0 := analyser.Metrics["H0"]
	h1 := analyser.Metrics["H1"]
	hld := analyser.Metrics["HLD"]
	expectedHLD := math.Abs(h0-60.0) + math.Abs(h1-40.0)

	if math.Abs(hld-expectedHLD) > 0.01 {
		t.Errorf("HLD calculation mismatch with custom target: got %.2f, expected %.2f",
			hld, expectedHLD)
	}
}
