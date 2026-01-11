package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

// Note: Parse and scale tests have been moved to internal/keycraft/targets_test.go
// These functions are now part of the centralized targets package.

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

func TestFloatConv(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "simple string",
			input:   "1.23",
			want:    1.23,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    1.23,
			wantErr: true,
		},
		{
			name:    "space string",
			input:   " ",
			want:    1.23,
			wantErr: true,
		},
		{
			name:    "simple string with l space",
			input:   " 1.23",
			want:    1.23,
			wantErr: true,
		},
		{
			name:    "simple string with t space",
			input:   "1.23 ",
			want:    1.23,
			wantErr: true,
		},
		{
			name:    "simple string with l/t space",
			input:   " 1.23 ",
			want:    1.23,
			wantErr: true,
		},
		{
			name:    "simple string with text",
			input:   " hello ",
			want:    1.23,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strconv.ParseFloat(tt.input, 64)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkFloatConv benchmarks strconv.ParseFloat performance with various inputs.
func BenchmarkFloatConv(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"simple-float", "1.23"},
		{"integer", "42"},
		{"large-float", "123456.789012"},
		{"scientific", "1.23e-10"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = strconv.ParseFloat(bm.input, 64)
			}
		})
	}
}
