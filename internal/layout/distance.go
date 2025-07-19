package layout

import (
	"math"
)

type KeyPair struct {
	row1, col1, row2, col2 uint8
}

func NewKeyPair(a, b KeyInfo) KeyPair {
	// Ensure that the KeyPair is symmetric (the order of the keys doesn't matter)
	if a.Row < b.Row || (a.Row == b.Row && a.Column < b.Column) {
		return KeyPair{
			row1: a.Row,
			col1: a.Column,
			row2: b.Row,
			col2: b.Column,
		}
	}
	return KeyPair{
		row1: b.Row,
		col1: b.Column,
		row2: a.Row,
		col2: a.Column,
	}
}

type KeyDistance struct {
	distances  map[KeyPair]float32
	layoutType LayoutType
}

func NewKeyDistance(layoutType LayoutType) *KeyDistance {
	return &KeyDistance{
		distances:  make(map[KeyPair]float32),
		layoutType: layoutType,
	}
}

func (kd *KeyDistance) GetDistance(key1, key2 KeyInfo) float32 {
	// Only distance on the same hand is supported
	if key1.Hand != key2.Hand {
		return 0
	}

	// Only distance between 2 thumb keys or 2 finger keys is supported
	if (key1.Row == 3 && key2.Row != 3) || (key1.Row != 3 && key2.Row == 3) {
		return 0
	}

	kp := NewKeyPair(key1, key2)

	if distance, ok := kd.distances[kp]; ok {
		return distance
	}

	distance := kd.calculateDistance(kp)
	kd.distances[kp] = distance
	return distance
}

func (kd *KeyDistance) calculateDistance(kp KeyPair) float32 {
	// We assume keys are always on the same finger, and KeyPair is symmetric.
	switch kd.layoutType {
	case OrthoLayout:
		return calcDistOrtho(kp)
	case RowStagLayout:
		return calcDistRowStag(kp)
	case ColStagLayout:
		return calcDistColStag(kp)
	default:
		panic("unsupported layout type")
	}
}

func calcDistRowStag(kp KeyPair) float32 {
	// thumbs row uses columns
	if kp.row1 == 3 {
		return float32(kp.col2 - kp.col1)
	}

	// calculate column offset based on row
	var colOffset1, colOffset2 float32
	switch kp.row1 {
	case 1:
		colOffset1 = 0.25
	case 2:
		colOffset1 = 0.75
	}
	switch kp.row2 {
	case 1:
		colOffset2 = 0.25
	case 2:
		colOffset2 = 0.75
	}

	// calculate effective column positions
	col1 := float32(kp.col1) + colOffset1
	col2 := float32(kp.col2) + colOffset2

	// calculate distance
	dx := col2 - col1
	dy := float32(kp.row2 - kp.row1)
	mul := dx*dx + dy*dy
	switch mul {
	case 1:
		return 1
	case 2:
		return math.Sqrt2
	default:
		return float32(math.Sqrt(float64(mul)))
	}
}

func calcDistOrtho(kp KeyPair) float32 {
	// thumbs row uses columns
	if kp.row1 == 3 {
		return float32(kp.col2 - kp.col1)
	}

	// fingers, same column, diff is the rows
	if kp.col1 == kp.col2 {
		return float32(kp.row2 - kp.row1)
	}

	// cases of some finger, diff column (index and pinky)
	var dx uint8
	if kp.col1 < kp.col2 {
		dx = kp.col2 - kp.col1
	} else {
		dx = kp.col1 - kp.col2
	}
	dy := kp.row2 - kp.row1
	mul := dx*dx + dy*dy
	switch mul {
	case 1:
		return 1 // no calc necessary
	case 2:
		return math.Sqrt2 // pre-calculated
	default:
		return float32(math.Sqrt(float64(mul)))
	}
}

// Corne-style offsets
var colStagOffsets = [12]float32{
	0.35, 0.35, 0.1, 0, 0.1, 0.2, 0.2, 0.1, 0, 0.1, 0.35, 0.35,
}

func calcDistColStag(kp KeyPair) float32 {
	// thumbs row uses columns
	if kp.row1 == 3 {
		return float32(kp.col2 - kp.col1)
	}

	// calculate effective row positions
	row1 := float32(kp.row1) + colStagOffsets[kp.col1]
	row2 := float32(kp.row2) + colStagOffsets[kp.col2]

	// calculate distance
	dx := float32(kp.col2) - float32(kp.col1)
	dy := row2 - row1
	mul := dx*dx + dy*dy
	switch mul {
	case 1:
		return 1
	case 2:
		return math.Sqrt2
	default:
		return float32(math.Sqrt(float64(mul)))
	}
}
