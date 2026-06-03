package hex

// Axial is a hex-grid coordinate.
//
// The pair (Q, R) addresses a unique hex on an infinite grid. The implicit
// third axial value S = -Q - R is never stored.
//
// Axial is a comparable value type, safe to use as a map key.
type Axial struct {
	Q, R int
}

// At is a convenience constructor: hex.At(1, 2) == hex.Axial{Q: 1, R: 2}.
func At(q, r int) Axial { return Axial{Q: q, R: r} }

// S returns the implicit third axial component (S = -Q - R).
func (a Axial) S() int { return -a.Q - a.R }

// Add returns a + b.
func (a Axial) Add(b Axial) Axial { return Axial{Q: a.Q + b.Q, R: a.R + b.R} }

// Sub returns a - b.
func (a Axial) Sub(b Axial) Axial { return Axial{Q: a.Q - b.Q, R: a.R - b.R} }

// Distance returns the number of hex steps between a and b.
func (a Axial) Distance(b Axial) int {
	d := a.Sub(b)
	return (abs(d.Q) + abs(d.R) + abs(d.S())) / 2
}

// axialDirections are the six unit vectors in axial space. The orientation of
// the hex (flat-top vs pointy-top) does not change these neighbors; only the
// pixel projection differs.
var axialDirections = [6]Axial{
	{+1, 0}, {+1, -1}, {0, -1},
	{-1, 0}, {-1, +1}, {0, +1},
}

// Neighbors returns the six axial neighbors of a, in a fixed order.
func (a Axial) Neighbors() [6]Axial {
	var out [6]Axial
	for i, d := range axialDirections {
		out[i] = a.Add(d)
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
