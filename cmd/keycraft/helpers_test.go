package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeTestLayout creates a layout file with the given content for testing.
// Returns the full path. Calls t.Fatalf on write error.
func writeTestLayout(t *testing.T, dir, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test layout: %v", err)
	}
	return path
}

func TestLoadLayout(t *testing.T) {
	// Save and restore original layoutDir
	origLayoutDir := layoutDir
	defer func() { layoutDir = origLayoutDir }()

	// Override with a temporary test dir
	tmpDir := t.TempDir()
	layoutDir = tmpDir

	tests := []struct {
		name     string
		filename string
		setup    bool
		wantErr  bool
	}{
		{
			name:     "empty filename",
			filename: "",
			wantErr:  true,
		},
		{
			name:     "valid layout file with .klf",
			filename: "testlayout.klf",
			setup:    true,
			wantErr:  false,
		},
		{
			name:     "filename without extension adds .klf",
			filename: "plainname",
			setup:    true, // will create plainname.klf
			wantErr:  false,
		},
		{
			name:     "filename with uppercase extension",
			filename: "upper.KLF",
			setup:    true,
			wantErr:  false,
		},
		{
			name:     "missing layout file",
			filename: "missing.klf",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup {
				// normalize extension for setup: add .klf if not present (case-insensitive)
				fname := tt.filename
				ext := filepath.Ext(fname)
				if strings.ToLower(ext) != ".klf" {
					fname += ".klf"
				}
				writeTestLayout(t, tmpDir, fname, "rowstag\n ~ q w e r t  y u i o p \\\n ~ a s d f g  h j k l ; '\n ~ z x c v b  n m , . / ~\n ~ ~ ~  _ ~ ~")
			}

			layout, err := loadLayout(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}

			if !tt.wantErr && layout == nil {
				t.Fatal("expected non-nil layout, got nil")
			}
		})
	}
}

// TestParseFingerLoad verifies parsing of 4-value (mirrored) and 8-value formats,
// plus error handling for invalid inputs (wrong count, empty values, non-numeric, negative).
func TestParseFingerLoad(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *[10]float64
		wantErr bool
	}{
		{
			name:  "valid 4 values #1",
			input: "1.0,2.0,3.0,4.0",
			want:  &[10]float64{1, 2, 3, 4, 0, 0, 4, 3, 2, 1},
		},
		{
			name:  "valid 4 values #2",
			input: " 1.0, 2.0, 3, 4 ",
			want:  &[10]float64{1, 2, 3, 4, 0, 0, 4, 3, 2, 1},
		},
		{
			name:  "valid 8 values",
			input: "1,2,3,4,5,6,7,8",
			want:  &[10]float64{1, 2, 3, 4, 0, 0, 5, 6, 7, 8},
		},
		{
			name:    "invalid length #1",
			input:   "1,2,3",
			wantErr: true,
		},
		{
			name:    "invalid length #2",
			input:   "1,2,3,4,5",
			wantErr: true,
		},
		{
			name:    "invalid length #3",
			input:   "1,2,3,4,5,6,7,8,9",
			wantErr: true,
		},
		{
			name:    "invalid length #4",
			input:   "1,2,3,4,5,6,7,8,9,10",
			wantErr: true,
		},
		{
			name:    "empty value",
			input:   "1,,3,4",
			wantErr: true,
		},
		{
			name:    "invalid float #1",
			input:   "1,2,3,q",
			wantErr: true,
		},
		{
			name:    "invalid float #2",
			input:   "a,b,c,d",
			wantErr: true,
		},
		{
			name:    "negative value",
			input:   "1.0,-2.0,3.0,4.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFingerLoad(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkParseFingerLoad benchmarks parsing performance for 4-value and 8-value inputs.
func BenchmarkParseFingerLoad(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"4-values", "1.0,2.0,3.0,4.0"},
		{"8-values", "1.0,2.0,3.0,4.0,5.0,6.0,7.0,8.0"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_, err := parseFingerLoad(bm.input)
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestScaleFingerLoad verifies in-place scaling to sum 100.0, non-negative validation,
// and edge cases (zero sum, negative values, epsilon threshold).
func TestScaleFingerLoad(t *testing.T) {
	tests := []struct {
		name    string
		input   [10]float64
		wantErr bool
		checkFn func(*testing.T, *[10]float64)
	}{
		{
			name:    "normal scaling",
			input:   [10]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			wantErr: false,
			checkFn: func(t *testing.T, vals *[10]float64) {
				var sum float64
				for _, v := range vals {
					sum += v
				}
				if sum < 99.99 || sum > 100.01 {
					t.Errorf("scaled sum not 100, got %v", sum)
				}
			},
		},
		{
			name:    "all equal values",
			input:   [10]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			wantErr: false,
			checkFn: func(t *testing.T, vals *[10]float64) {
				expected := 10.0
				for i, v := range vals {
					if v < expected-0.01 || v > expected+0.01 {
						t.Errorf("vals[%d] = %v, want ~%v", i, v, expected)
					}
				}
			},
		},
		{
			name:    "zero sum error",
			input:   [10]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr: true,
		},
		{
			name:    "negative value error",
			input:   [10]float64{1, 2, -3, 4, 5, 6, 7, 8, 9, 10},
			wantErr: true,
		},
		{
			name:    "all negative values error",
			input:   [10]float64{-1, -2, -3, -4, -5, -6, -7, -8, -9, -10},
			wantErr: true,
		},
		{
			name:    "very small sum error",
			input:   [10]float64{1e-10, 1e-10, 1e-10, 1e-10, 1e-10, 1e-10, 1e-10, 1e-10, 1e-10, 1e-10},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := tt.input
			err := scaleFingerLoad(&vals)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, &vals)
			}
		})
	}
}

// TestEnsureKlf verifies .klf extension handling (case-insensitive check, append if missing).
func TestEnsureKlf(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"foo", "foo.klf"},
		{"foo.klf", "foo.klf"},
		{"foo.KLF", "foo.KLF"},
		{"foo.bar", "foo.bar.klf"},
		{"foo.klF", "foo.klF"}, // Only .klf (lowercase) triggers match
		{"foo.KlF", "foo.KlF"},
		{"foo.", "foo..klf"},
		{"", ".klf"},
	}
	for _, tt := range tests {
		got := ensureKlf(tt.in)
		if got != tt.out {
			t.Errorf("ensureKlf(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

// TestEnsureNoKlf verifies .klf extension removal (case-insensitive check).
func TestEnsureNoKlf(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"foo.klf", "foo"},
		{"foo.KLF", "foo"},
		{"foo.klF", "foo"},
		{"foo.KlF", "foo"},
		{"foo.bar", "foo.bar"},
		{"foo", "foo"},
		{".klf", ""},
		{".KLF", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := ensureNoKlf(tt.in)
		if got != tt.out {
			t.Errorf("ensureNoKlf(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}
