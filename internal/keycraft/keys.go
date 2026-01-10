package keycraft

import (
	"fmt"
	"math"
)

// KeyPair represents an ordered pair of key indices.
type KeyPair [2]uint8

// KeyPairDistance contains precomputed distance metrics between two key indices.
type KeyPairDistance struct {
	RowDist    float64 // vertical (row) distance in layout units
	ColDist    float64 // horizontal (column) distance in layout units
	FingerDist uint8   // absolute difference between the two keys' finger indices
	Distance   float64 // Euclidean distance (sqrt(RowDist^2 + ColDist^2))
}

// keyDistances contains precomputed key pair distances for each LayoutType.
// Distance functions are selected based on keyboard geometry:
//   - ROWSTAG: AbsRowDist, AbsColDistAdj (accounts for row stagger)
//   - ANGLEMOD: AbsRowDist, AbsColDistAdj (similar to row-staggered)
//   - ORTHO: AbsRowDist, AbsColDist (simple grid distances)
//   - COLSTAG: AbsRowDistAdj, AbsColDist (accounts for column stagger)
var keyDistances = []map[KeyPair]KeyPairDistance{
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &keyToFinger),         // ROWSTAG
	calcKeyDistances(AbsRowDist, AbsColDistAdj, &angleModKeyToFinger), // ANGLEMOD
	calcKeyDistances(AbsRowDist, AbsColDist, &keyToFinger),            // ORTHO
	calcKeyDistances(AbsRowDistAdj, AbsColDist, &keyToFinger),         // COLSTAG
}

// rowStagOffsets defines the horizontal offset for each row in row-staggered layouts.
// Traditional keyboards have rows offset by different amounts (in key units).
var rowStagOffsets = [4]float64{
	0, 0.25, 0.75, 0, // Top, home, bottom, thumb rows
}

// colStagOffsets defines the vertical offset for each column in column-staggered layouts.
// Ergonomic keyboards (e.g., Corne) stagger columns to match natural finger lengths.
var colStagOffsets = [12]float64{
	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
}

const (
	LEFT  uint8 = 0 // Left hand
	RIGHT uint8 = 1 // Right hand
)

// KeyInfo represents a key's physical position and typing finger on a keyboard.
type KeyInfo struct {
	Index  uint8 // 0-41
	Hand   uint8 // LEFT or RIGHT
	Row    uint8 // 0-3
	Column uint8 // 0-11 for Row=0-2, 0-5 for Row=3
	Finger uint8 // 0-9
}

// NewKeyInfo constructs a KeyInfo from row, column, and layout type.
// Automatically determines hand and finger assignments based on position and geometry.
func NewKeyInfo(row, col uint8, layoutType LayoutType) KeyInfo {
	if col >= uint8(len(keyToFinger)) {
		panic(fmt.Sprintf("col exceeds max value: %d", col))
	}
	if row > 3 {
		panic(fmt.Sprintf("row exceeds max value: %d", row))
	}

	index := 12*row + col
	if index >= 42 {
		panic(fmt.Sprintf("index exceeds max value: %d", index))
	}

	hand := RIGHT
	if row < 3 && col < 6 {
		hand = LEFT
	} else if row == 3 && col < 3 {
		hand = LEFT
	}

	var finger uint8
	if layoutType == ANGLEMOD {
		finger = angleModKeyToFinger[index]
	} else {
		finger = keyToFinger[index]
	}

	return KeyInfo{
		Index:  index,
		Hand:   hand,
		Row:    row,
		Column: col,
		Finger: finger,
	}
}

// Distance returns the precomputed distance between two key indices.
// If the key pair is not found, it returns nil.
func (sl *SplitLayout) Distance(k1, k2 uint8) *KeyPairDistance {
	kpd, ok := (*sl.KeyPairDistances)[KeyPair{k1, k2}]
	if !ok {
		return nil
	}
	return &kpd
}

// calcKeyDistances precomputes all pairwise distances between keys on the same hand.
// Uses the provided distance functions to account for layout-specific geometry.
// Note: thumb key distance calculations have a known minor inaccuracy.
func calcKeyDistances(
	rowDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	colDistFunc func(row1 uint8, col1 uint8, row2 uint8, col2 uint8) float64,
	keyToFinger *[42]uint8,
) map[KeyPair]KeyPairDistance {
	keyDistances := make(map[KeyPair]KeyPairDistance, 624)

	// Optimized square root for common cases
	sqrt := func(mul float64) float64 {
		switch mul {
		case 1:
			return 1 // no calc necessary
		case 2:
			return math.Sqrt2 // pre-calculated
		default:
			return math.Sqrt(mul)
		}
	}

	absDist := func(x, y uint8) uint8 {
		if x > y {
			return x - y
		} else {
			return y - x
		}
	}

	var k1, k2 uint8
	for k1 = range 42 {
		row1, col1 := k1/12, k1%12
		for k2 = range 42 {
			if k1 == k2 {
				continue
			}
			row2, col2 := k2/12, k2%12

			// Skip pairs on different hands
			if ((row1 < 3 && col1 < 6) || (row1 >= 3 && col1 < 3)) !=
				((row2 < 3 && col2 < 6) || (row2 >= 3 && col2 < 3)) {
				continue
			}

			// Skip pairs between main rows and thumb row
			if (row1 < 3) != (row2 < 3) {
				continue
			}

			// Compute distance metrics
			dx := colDistFunc(row1, col1, row2, col2)
			dy := rowDistFunc(row1, col1, row2, col2)
			dist := sqrt(dx*dx + dy*dy)
			keyDistances[KeyPair{k1, k2}] = KeyPairDistance{
				RowDist:    dy,
				ColDist:    dx,
				FingerDist: absDist(keyToFinger[k1], keyToFinger[k2]),
				Distance:   dist,
			}
		}
	}

	return keyDistances
}

// AbsRowDist computes the absolute vertical distance between two keys (simple).
func AbsRowDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(row1) - float64(row2))
}

// AbsRowDistAdj computes vertical distance accounting for column-stagger offsets.
func AbsRowDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(row1) + colStagOffsets[col1] -
		(float64(row2) + colStagOffsets[col2])))
}

// AbsColDist computes the absolute horizontal distance between two keys (simple).
func AbsColDist(row1, col1, row2, col2 uint8) float64 {
	return math.Abs(float64(col1) - float64(col2))
}

// AbsColDistAdj computes horizontal distance accounting for row-stagger offsets.
func AbsColDistAdj(row1, col1, row2, col2 uint8) float64 {
	return math.Abs((float64(col1) + rowStagOffsets[row1] -
		(float64(col2) + rowStagOffsets[row2])))
}
