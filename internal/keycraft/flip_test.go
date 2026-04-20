package keycraft

import (
	"path/filepath"
	"testing"
)

func TestFlipLayout(t *testing.T) {
	path := filepath.Join("..", "..", "data", "layouts", "qwerty.klf")

	orig, err := NewLayoutFromFile("qwerty", path)
	if err != nil {
		t.Fatalf("load original: %v", err)
	}

	res, err := FlipLayout(FlipInput{LayoutFile: path})
	if err != nil {
		t.Fatalf("FlipLayout: %v", err)
	}
	if res == nil || res.Layout == nil {
		t.Fatal("nil result")
	}
	if res.Layout.Name != "qwerty-flipped" {
		t.Errorf("name: got %q, want %q", res.Layout.Name, "qwerty-flipped")
	}

	// Position 0 after flip must equal position 11 before flip (same row).
	if res.Layout.Runes[0] != orig.Runes[11] {
		t.Errorf("flip[0]=%q, orig[11]=%q", res.Layout.Runes[0], orig.Runes[11])
	}
	// Flipping twice must restore the original arrangement.
	res2, err := FlipLayout(FlipInput{LayoutFile: path, NewName: "tmp"})
	if err != nil {
		t.Fatalf("FlipLayout 2nd: %v", err)
	}
	res2.Layout.FlipHorizontal()
	for i := range orig.Runes {
		if res2.Layout.Runes[i] != orig.Runes[i] {
			t.Errorf("double-flip mismatch at %d: got %q, want %q", i, res2.Layout.Runes[i], orig.Runes[i])
		}
	}
}

func TestFlipLayoutEmptyPath(t *testing.T) {
	if _, err := FlipLayout(FlipInput{}); err == nil {
		t.Error("expected error for empty LayoutFile")
	}
}

func TestFlipLayoutCustomName(t *testing.T) {
	path := filepath.Join("..", "..", "data", "layouts", "qwerty.klf")
	res, err := FlipLayout(FlipInput{LayoutFile: path, NewName: "custom-name"})
	if err != nil {
		t.Fatalf("FlipLayout: %v", err)
	}
	if res.Layout.Name != "custom-name" {
		t.Errorf("name: got %q, want %q", res.Layout.Name, "custom-name")
	}
}
