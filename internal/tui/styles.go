package tui

import "github.com/jedib0t/go-pretty/v6/table"

// EmptyStyle returns a table style with no borders on the sides.
func EmptyStyle() table.Style {
	s := table.StyleDefault
	s.Box = table.StyleBoxRounded
	s.Box = table.BoxStyle{
		BottomLeft:       s.Box.BottomLeft,
		BottomRight:      s.Box.BottomRight,
		BottomSeparator:  s.Box.BottomSeparator,
		Left:             " ",
		LeftSeparator:    s.Box.LeftSeparator,
		MiddleHorizontal: " ",
		MiddleSeparator:  s.Box.MiddleSeparator,
		MiddleVertical:   " ",
		Right:            " ",
		RightSeparator:   s.Box.RightSeparator,
		TopLeft:          s.Box.TopLeft,
		TopRight:         s.Box.TopRight,
		TopSeparator:     s.Box.TopSeparator,
		UnfinishedRow:    " ",
	}
	return s
}
