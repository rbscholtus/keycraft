package keycraft

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWeights(t *testing.T) {
	weights := NewWeights()
	if weights == nil {
		t.Fatal("NewWeights() returned nil")
	}
	if weights.weights == nil {
		t.Error("weights map should be initialized")
	}
}

func TestNewWeightsFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		checks  map[string]float64
	}{
		{
			name:    "single metric",
			input:   "SFB=-8.0",
			wantErr: false,
			checks:  map[string]float64{"SFB": -8.0},
		},
		{
			name:    "multiple metrics",
			input:   "SFB=-8.0,LSB=-4.0,FSB=-2.666",
			wantErr: false,
			checks: map[string]float64{
				"SFB": -8.0,
				"LSB": -4.0,
				"FSB": -2.666,
			},
		},
		{
			name:    "metrics with spaces",
			input:   " SFB = -8.0 , LSB = -4.0 ",
			wantErr: false,
			checks: map[string]float64{
				"SFB": -8.0,
				"LSB": -4.0,
			},
		},
		{
			name:    "lowercase converted to uppercase",
			input:   "sfb=-8.0,lsb=-4.0",
			wantErr: false,
			checks: map[string]float64{
				"SFB": -8.0,
				"LSB": -4.0,
			},
		},
		{
			name:    "positive values allowed",
			input:   "ALT-NML=4.0,2RL-IN=2.0",
			wantErr: false,
			checks: map[string]float64{
				"ALT-NML": 4.0,
				"2RL-IN":  2.0,
			},
		},
		{
			name:    "empty string is valid",
			input:   "",
			wantErr: false,
			checks:  map[string]float64{},
		},
		{
			name:    "invalid format - missing equals",
			input:   "SFB-8.0",
			wantErr: true,
		},
		{
			name:    "invalid format - multiple equals",
			input:   "SFB=-8.0=extra",
			wantErr: true,
		},
		{
			name:    "invalid metric name",
			input:   "INVALID_METRIC=-8.0",
			wantErr: true,
		},
		{
			name:    "invalid weight value - not a number",
			input:   "SFB=abc",
			wantErr: true,
		},
		{
			name:    "invalid weight value - empty",
			input:   "SFB=",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights, err := NewWeightsFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWeightsFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for metric, expectedValue := range tt.checks {
				gotValue := weights.Get(metric)
				if gotValue != expectedValue {
					t.Errorf("metric %s: got %f, want %f", metric, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestAddWeightsFromString(t *testing.T) {
	weights := NewWeights()

	// Add first set of weights
	err := weights.AddWeightsFromString("SFB=-8.0,LSB=-4.0")
	if err != nil {
		t.Fatalf("AddWeightsFromString() first call error = %v", err)
	}

	if weights.Get("SFB") != -8.0 {
		t.Errorf("SFB: got %f, want -8.0", weights.Get("SFB"))
	}
	if weights.Get("LSB") != -4.0 {
		t.Errorf("LSB: got %f, want -4.0", weights.Get("LSB"))
	}

	// Add more weights (should not overwrite existing unless explicitly set)
	err = weights.AddWeightsFromString("FSB=-2.666,HSB=-1.333")
	if err != nil {
		t.Fatalf("AddWeightsFromString() second call error = %v", err)
	}

	// Check all weights are present
	if weights.Get("SFB") != -8.0 {
		t.Errorf("SFB after second add: got %f, want -8.0", weights.Get("SFB"))
	}
	if weights.Get("FSB") != -2.666 {
		t.Errorf("FSB: got %f, want -2.666", weights.Get("FSB"))
	}

	// Override existing weight
	err = weights.AddWeightsFromString("SFB=-10.0")
	if err != nil {
		t.Fatalf("AddWeightsFromString() override error = %v", err)
	}

	if weights.Get("SFB") != -10.0 {
		t.Errorf("SFB after override: got %f, want -10.0", weights.Get("SFB"))
	}
}

func TestWeightsGet(t *testing.T) {
	weights := NewWeights()
	_ = weights.AddWeightsFromString("SFB=-8.0,LSB=-4.0")

	tests := []struct {
		metric   string
		expected float64
	}{
		{"SFB", -8.0},
		{"LSB", -4.0},
		{"NONEXISTENT", 0.0}, // Should return 0 for missing metrics
		{"", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			got := weights.Get(tt.metric)
			if got != tt.expected {
				t.Errorf("Get(%q) = %f, want %f", tt.metric, got, tt.expected)
			}
		})
	}
}

func TestNewWeightsFromParams(t *testing.T) {
	// Create a temporary weights file
	tmpDir := t.TempDir()
	weightsPath := filepath.Join(tmpDir, "test_weights.txt")

	content := `# Test weights file
SFB = -8.0
LSB = -4.0
FSB = -2.666

# Comment line
HSB = -1.333
`

	err := os.WriteFile(weightsPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test weights file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		weightsStr  string
		wantErr     bool
		checkMetric string
		checkValue  float64
	}{
		{
			name:        "load from file only",
			filePath:    weightsPath,
			weightsStr:  "",
			wantErr:     false,
			checkMetric: "SFB",
			checkValue:  -8.0,
		},
		{
			name:        "load from string only",
			filePath:    "",
			weightsStr:  "SFB=-10.0,LSB=-5.0",
			wantErr:     false,
			checkMetric: "SFB",
			checkValue:  -10.0,
		},
		{
			name:        "string overrides file",
			filePath:    weightsPath,
			weightsStr:  "SFB=-100.0",
			wantErr:     false,
			checkMetric: "SFB",
			checkValue:  -100.0,
		},
		{
			name:        "missing file with valid string",
			filePath:    "/nonexistent/path.txt",
			weightsStr:  "SFB=-8.0",
			wantErr:     true, // File error should propagate
			checkMetric: "SFB",
			checkValue:  0,
		},
		{
			name:        "both empty",
			filePath:    "",
			weightsStr:  "",
			wantErr:     false,
			checkMetric: "SFB",
			checkValue:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights, err := NewWeightsFromParams(tt.filePath, tt.weightsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWeightsFromParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			got := weights.Get(tt.checkMetric)
			if got != tt.checkValue {
				t.Errorf("metric %s: got %f, want %f", tt.checkMetric, got, tt.checkValue)
			}
		})
	}
}

func TestAddWeightsFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		checks   map[string]float64
	}{
		{
			name: "valid file with comments",
			content: `# Comment line
SFB = -8.0
LSB = -4.0

# Another comment
FSB = -2.666
`,
			wantErr: false,
			checks: map[string]float64{
				"SFB": -8.0,
				"LSB": -4.0,
				"FSB": -2.666,
			},
		},
		{
			name: "file with blank lines",
			content: `
SFB = -8.0

LSB = -4.0

`,
			wantErr: false,
			checks: map[string]float64{
				"SFB": -8.0,
				"LSB": -4.0,
			},
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
			checks:  map[string]float64{},
		},
		{
			name:    "only comments",
			content: "# Just comments\n# Nothing else\n",
			wantErr: false,
			checks:  map[string]float64{},
		},
		{
			name:    "invalid format in file",
			content: "SFB-8.0\n",
			wantErr: true,
		},
		{
			name:    "invalid metric in file",
			content: "INVALID_METRIC=-8.0\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			filePath := filepath.Join(tmpDir, "test_"+tt.name+".txt")
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			weights := NewWeights()
			err = weights.AddWeightsFromFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddWeightsFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for metric, expectedValue := range tt.checks {
				gotValue := weights.Get(metric)
				if gotValue != expectedValue {
					t.Errorf("metric %s: got %f, want %f", metric, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestAddWeightsFromFile_MissingFile(t *testing.T) {
	weights := NewWeights()
	err := weights.AddWeightsFromFile("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestWeightsCaseInsensitive(t *testing.T) {
	weights := NewWeights()

	// Add with lowercase
	err := weights.AddWeightsFromString("sfb=-8.0")
	if err != nil {
		t.Fatalf("AddWeightsFromString() error = %v", err)
	}

	// Should be stored as uppercase
	if weights.Get("SFB") != -8.0 {
		t.Errorf("Get(\"SFB\") = %f, want -8.0", weights.Get("SFB"))
	}
	if weights.Get("sfb") != 0.0 {
		t.Errorf("Get(\"sfb\") should return 0.0, got %f", weights.Get("sfb"))
	}
}

func TestWeightsAllMetrics(t *testing.T) {
	// Verify all metrics in MetricsMap["all"] are valid for weights
	allMetrics := MetricsMap["all"]

	weights := NewWeights()

	// Try to add each metric with a test value
	for _, metric := range allMetrics {
		input := metric + "=-1.0"
		err := weights.AddWeightsFromString(input)
		if err != nil {
			t.Errorf("Metric %s should be valid but got error: %v", metric, err)
		}

		// Verify it was set
		got := weights.Get(metric)
		if got != -1.0 {
			t.Errorf("Metric %s: expected -1.0, got %f", metric, got)
		}
	}
}

func TestWeightsZeroValue(t *testing.T) {
	weights := NewWeights()
	err := weights.AddWeightsFromString("SFB=0.0,LSB=0,FSB=-0.0")
	if err != nil {
		t.Fatalf("AddWeightsFromString() error = %v", err)
	}

	tests := []struct {
		metric   string
		expected float64
	}{
		{"SFB", 0.0},
		{"LSB", 0.0},
		{"FSB", 0.0},
	}

	for _, tt := range tests {
		got := weights.Get(tt.metric)
		if got != tt.expected {
			t.Errorf("Get(%q) = %f, want %f", tt.metric, got, tt.expected)
		}
	}
}

func BenchmarkNewWeightsFromString(b *testing.B) {
	input := "SFB=-8.0,LSB=-4.0,FSB=-2.666,HSB=-1.333,SFS=-3.0,LSS=-1.0"
	for b.Loop() {
		_, err := NewWeightsFromString(input)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWeightsGet(b *testing.B) {
	weights := NewWeights()
	_ = weights.AddWeightsFromString("SFB=-8.0,LSB=-4.0,FSB=-2.666")

	b.ResetTimer()
	for b.Loop() {
		_ = weights.Get("SFB")
	}
}

func BenchmarkAddWeightsFromString(b *testing.B) {
	input := "SFB=-8.0,LSB=-4.0,FSB=-2.666"
	for b.Loop() {
		weights := NewWeights()
		_ = weights.AddWeightsFromString(input)
	}
}
