package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// helper: write a minimal layout file for testing
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

// go test -timeout 30s -v -run ^TestParseFingerLoad$ github.com/rbscholtus/keycraft/cmd/keycraft
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

// BenchmarkParseFingerLoad runs parseFingerLoad repeatedly to measure performance.
// go test ./cmd/keycraft -bench . -benchmem
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

func TestScaleFingerLoad(t *testing.T) {
	// Test normal scaling
	vals := &[10]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	scaled, err := scaleFingerLoad(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sum float64
	for _, v := range scaled {
		sum += v
	}
	if sum < 99.99 || sum > 100.01 {
		t.Errorf("scaled sum not 100, got %v", sum)
	}

	// Test zero-sum scaling
	zeroVals := &[10]float64{}
	_, err = scaleFingerLoad(zeroVals)
	if err == nil {
		t.Error("expected error for zero sum, got nil")
	}
}

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
