package keycraft

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTargetLoads(t *testing.T) {
	targets := NewTargetLoads()

	if targets.TargetHandLoad == nil {
		t.Error("TargetHandLoad should not be nil")
	}
	if targets.TargetFingerLoad == nil {
		t.Error("TargetFingerLoad should not be nil")
	}
	if targets.TargetRowLoad == nil {
		t.Error("TargetRowLoad should not be nil")
	}
	if targets.PinkyPenalties == nil {
		t.Error("PinkyPenalties should not be nil")
	}

	// Verify default values match documented defaults
	if targets.TargetHandLoad[0] != 50.0 || targets.TargetHandLoad[1] != 50.0 {
		t.Errorf("Expected hand load [50, 50], got [%f, %f]",
			targets.TargetHandLoad[0], targets.TargetHandLoad[1])
	}

	expectedFingers := [10]float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7}
	for i, expected := range expectedFingers {
		if targets.TargetFingerLoad[i] != expected {
			t.Errorf("Finger %d: expected %f, got %f", i, expected, targets.TargetFingerLoad[i])
		}
	}

	if targets.TargetRowLoad[0] != 17.5 || targets.TargetRowLoad[1] != 75.0 || targets.TargetRowLoad[2] != 7.5 {
		t.Errorf("Expected row load [17.5, 75.0, 7.5], got [%f, %f, %f]",
			targets.TargetRowLoad[0], targets.TargetRowLoad[1], targets.TargetRowLoad[2])
	}

	expectedPinky := [12]float64{2, 1.5, 1, 0, 2, 1.5, 2, 1.5, 1, 0, 2, 1.5}
	for i, expected := range expectedPinky {
		if targets.PinkyPenalties[i] != expected {
			t.Errorf("Pinky penalty %d: expected %f, got %f", i, expected, targets.PinkyPenalties[i])
		}
	}
}

func TestParseTargetHandLoad(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		expected  *[2]float64
		afterScale *[2]float64
	}{
		{
			name:       "valid equal split",
			input:      "50, 50",
			wantErr:    false,
			expected:   &[2]float64{50, 50},
			afterScale: &[2]float64{50, 50},
		},
		{
			name:       "valid unequal split",
			input:      "60, 40",
			wantErr:    false,
			expected:   &[2]float64{60, 40},
			afterScale: &[2]float64{60, 40},
		},
		{
			name:       "needs scaling",
			input:      "1, 1",
			wantErr:    false,
			expected:   &[2]float64{1, 1},
			afterScale: &[2]float64{50, 50},
		},
		{
			name:    "too few values",
			input:   "50",
			wantErr: true,
		},
		{
			name:    "too many values",
			input:   "50, 30, 20",
			wantErr: true,
		},
		{
			name:    "invalid float",
			input:   "abc, 50",
			wantErr: true,
		},
		{
			name:    "negative value",
			input:   "-10, 50",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTargetHandLoad(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTargetHandLoad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if result[0] != tt.expected[0] || result[1] != tt.expected[1] {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// Test scaling
			err = scaleTargetHandLoad(result)
			if err != nil {
				t.Errorf("scaleTargetHandLoad() error = %v", err)
				return
			}

			if result[0] != tt.afterScale[0] || result[1] != tt.afterScale[1] {
				t.Errorf("After scaling: expected %v, got %v", tt.afterScale, result)
			}
		})
	}
}

func TestParseFingerLoad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected *[10]float64
	}{
		{
			name:     "4 values (mirrored)",
			input:    "7, 10, 16, 17",
			wantErr:  false,
			expected: &[10]float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7},
		},
		{
			name:     "8 values",
			input:    "7, 10, 16, 17, 17, 16, 10, 7",
			wantErr:  false,
			expected: &[10]float64{7, 10, 16, 17, 0, 0, 17, 16, 10, 7},
		},
		{
			name:    "wrong count",
			input:   "7, 10, 16",
			wantErr: true,
		},
		{
			name:    "invalid float",
			input:   "7, abc, 16, 17",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFingerLoad(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFingerLoad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for i := 0; i < 10; i++ {
				if result[i] != tt.expected[i] {
					t.Errorf("Position %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}

			// Verify thumbs are always 0
			if result[4] != 0 || result[5] != 0 {
				t.Errorf("Thumbs should be 0, got F4=%f, F5=%f", result[4], result[5])
			}
		})
	}
}

func TestParseRowLoad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected *[3]float64
	}{
		{
			name:     "valid values",
			input:    "17.5, 75.0, 7.5",
			wantErr:  false,
			expected: &[3]float64{17.5, 75.0, 7.5},
		},
		{
			name:    "too few values",
			input:   "17.5, 75.0",
			wantErr: true,
		},
		{
			name:    "too many values",
			input:   "17.5, 75.0, 7.5, 10",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRowLoad(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRowLoad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for i := 0; i < 3; i++ {
				if result[i] != tt.expected[i] {
					t.Errorf("Position %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestParsePinkyPenalties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected *[12]float64
	}{
		{
			name:     "6 values (mirrored)",
			input:    "2.0, 1.5, 1.0, 0.0, 2.0, 1.5",
			wantErr:  false,
			expected: &[12]float64{2, 1.5, 1, 0, 2, 1.5, 2, 1.5, 1, 0, 2, 1.5},
		},
		{
			name:     "12 values",
			input:    "2, 1.5, 1, 0, 2, 1.5, 2, 1.5, 1, 0, 2, 1.5",
			wantErr:  false,
			expected: &[12]float64{2, 1.5, 1, 0, 2, 1.5, 2, 1.5, 1, 0, 2, 1.5},
		},
		{
			name:    "wrong count",
			input:   "2, 1.5, 1, 0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePinkyPenalties(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePinkyPenalties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for i := 0; i < 12; i++ {
				if result[i] != tt.expected[i] {
					t.Errorf("Position %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestSetHandLoad(t *testing.T) {
	targets := NewTargetLoads()

	err := targets.SetHandLoad("60, 40")
	if err != nil {
		t.Errorf("SetHandLoad() error = %v", err)
	}

	if targets.TargetHandLoad[0] != 60.0 || targets.TargetHandLoad[1] != 40.0 {
		t.Errorf("Expected [60, 40], got [%f, %f]",
			targets.TargetHandLoad[0], targets.TargetHandLoad[1])
	}
}

func TestSetFingerLoad(t *testing.T) {
	targets := NewTargetLoads()

	err := targets.SetFingerLoad("8, 11, 15, 16")
	if err != nil {
		t.Errorf("SetFingerLoad() error = %v", err)
	}

	expected := [10]float64{8, 11, 15, 16, 0, 0, 16, 15, 11, 8}
	for i, exp := range expected {
		if targets.TargetFingerLoad[i] != exp {
			t.Errorf("Position %d: expected %f, got %f", i, exp, targets.TargetFingerLoad[i])
		}
	}
}

func TestSetRowLoad(t *testing.T) {
	targets := NewTargetLoads()

	err := targets.SetRowLoad("20, 70, 10")
	if err != nil {
		t.Errorf("SetRowLoad() error = %v", err)
	}

	if targets.TargetRowLoad[0] != 20.0 || targets.TargetRowLoad[1] != 70.0 || targets.TargetRowLoad[2] != 10.0 {
		t.Errorf("Expected [20, 70, 10], got [%f, %f, %f]",
			targets.TargetRowLoad[0], targets.TargetRowLoad[1], targets.TargetRowLoad[2])
	}
}

func TestSetPinkyPenalties(t *testing.T) {
	targets := NewTargetLoads()

	err := targets.SetPinkyPenalties("1, 1, 1, 0, 1, 1")
	if err != nil {
		t.Errorf("SetPinkyPenalties() error = %v", err)
	}

	expected := [12]float64{1, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 1}
	for i, exp := range expected {
		if targets.PinkyPenalties[i] != exp {
			t.Errorf("Position %d: expected %f, got %f", i, exp, targets.PinkyPenalties[i])
		}
	}
}

func TestNewTargetLoadsFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_targets.txt")

	content := `# Test config
target-hand-load: 55, 45
target-finger-load: 8, 11, 15, 16
target-row-load: 20, 70, 10
pinky-penalties: 1.5, 1.0, 0.5, 0.0, 1.5, 1.0
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	targets, err := NewTargetLoadsFromFile(configPath)
	if err != nil {
		t.Fatalf("NewTargetLoadsFromFile() error = %v", err)
	}

	// Verify hand load (should be scaled)
	if targets.TargetHandLoad[0] != 55.0 || targets.TargetHandLoad[1] != 45.0 {
		t.Errorf("Hand load: expected [55, 45], got [%f, %f]",
			targets.TargetHandLoad[0], targets.TargetHandLoad[1])
	}

	// Verify finger load (4 values mirrored, with thumbs at 0)
	expectedFingers := [10]float64{8, 11, 15, 16, 0, 0, 16, 15, 11, 8}
	for i, exp := range expectedFingers {
		if targets.TargetFingerLoad[i] != exp {
			t.Errorf("Finger %d: expected %f, got %f", i, exp, targets.TargetFingerLoad[i])
		}
	}

	// Verify row load
	if targets.TargetRowLoad[0] != 20.0 || targets.TargetRowLoad[1] != 70.0 || targets.TargetRowLoad[2] != 10.0 {
		t.Errorf("Row load: expected [20, 70, 10], got [%f, %f, %f]",
			targets.TargetRowLoad[0], targets.TargetRowLoad[1], targets.TargetRowLoad[2])
	}

	// Verify pinky penalties (6 values mirrored)
	expectedPinky := [12]float64{1.5, 1.0, 0.5, 0.0, 1.5, 1.0, 1.5, 1.0, 0.5, 0.0, 1.5, 1.0}
	for i, exp := range expectedPinky {
		if targets.PinkyPenalties[i] != exp {
			t.Errorf("Pinky penalty %d: expected %f, got %f", i, exp, targets.PinkyPenalties[i])
		}
	}
}

func TestNewTargetLoadsFromFile_MissingFile(t *testing.T) {
	_, err := NewTargetLoadsFromFile("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestNewTargetLoadsFromFile_PartialConfig(t *testing.T) {
	// Create a temporary config file with only some fields
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial_targets.txt")

	content := `# Partial config
target-hand-load: 60, 40
# Other fields are missing
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	targets, err := NewTargetLoadsFromFile(configPath)
	if err != nil {
		t.Fatalf("NewTargetLoadsFromFile() error = %v", err)
	}

	// Verify specified field
	if targets.TargetHandLoad[0] != 60.0 || targets.TargetHandLoad[1] != 40.0 {
		t.Errorf("Hand load: expected [60, 40], got [%f, %f]",
			targets.TargetHandLoad[0], targets.TargetHandLoad[1])
	}

	// Verify missing fields got defaults
	if targets.TargetFingerLoad == nil {
		t.Error("TargetFingerLoad should have defaults")
	}
	if targets.TargetRowLoad == nil {
		t.Error("TargetRowLoad should have defaults")
	}
	if targets.PinkyPenalties == nil {
		t.Error("PinkyPenalties should have defaults")
	}
}
