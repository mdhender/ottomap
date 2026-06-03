package hex

import "fmt"

// Layout identifies one of the four offset-coordinate conventions used by
// renderers and on-disk formats.
//
// The two pairs are:
//
//   - OddQ / EvenQ:  flat-top hexes; columns are offset.
//   - OddR / EvenR:  pointy-top hexes; rows are offset.
//
// "Odd" means the odd-indexed column (or row) is shifted, "even" means the
// even-indexed column (or row) is shifted.
type Layout int

const (
	OddQ Layout = iota
	EvenQ
	OddR
	EvenR
)

// String implements the fmt.Stringer interface.
func (l Layout) String() string {
	switch l {
	case OddQ:
		return "odd-q"
	case EvenQ:
		return "even-q"
	case OddR:
		return "odd-r"
	case EvenR:
		return "even-r"
	}
	return fmt.Sprintf("Layout(%d)", int(l))
}

// OffsetCoord is a (column, row) coordinate pair as used by most map files.
type OffsetCoord struct {
	Col, Row int
}

// ToOffset converts an axial coordinate to the given offset layout.
func (a Axial) ToOffset(l Layout) OffsetCoord {
	switch l {
	case OddQ:
		return OffsetCoord{Col: a.Q, Row: a.R + (a.Q-(a.Q&1))/2}
	case EvenQ:
		return OffsetCoord{Col: a.Q, Row: a.R + (a.Q+(a.Q&1))/2}
	case OddR:
		return OffsetCoord{Col: a.Q + (a.R-(a.R&1))/2, Row: a.R}
	case EvenR:
		return OffsetCoord{Col: a.Q + (a.R+(a.R&1))/2, Row: a.R}
	}
	panic("hex: unknown Layout")
}

// FromOffset converts an offset coordinate to axial under the given layout.
func FromOffset(c OffsetCoord, l Layout) Axial {
	switch l {
	case OddQ:
		return Axial{Q: c.Col, R: c.Row - (c.Col-(c.Col&1))/2}
	case EvenQ:
		return Axial{Q: c.Col, R: c.Row - (c.Col+(c.Col&1))/2}
	case OddR:
		return Axial{Q: c.Col - (c.Row-(c.Row&1))/2, R: c.Row}
	case EvenR:
		return Axial{Q: c.Col - (c.Row+(c.Row&1))/2, R: c.Row}
	}
	panic("hex: unknown Layout")
}
