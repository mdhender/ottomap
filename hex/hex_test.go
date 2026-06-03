package hex

import "testing"

func TestAxialBasics(t *testing.T) {
	a := At(1, 2)
	if a.Q != 1 || a.R != 2 || a.S() != -3 {
		t.Fatalf("Axial fields wrong: %+v S=%d", a, a.S())
	}
	if got := a.Add(At(3, -1)); got != (Axial{4, 1}) {
		t.Errorf("Add: %+v", got)
	}
	if got := a.Sub(At(3, -1)); got != (Axial{-2, 3}) {
		t.Errorf("Sub: %+v", got)
	}
}

func TestDistance(t *testing.T) {
	cases := []struct {
		a, b Axial
		want int
	}{
		{At(0, 0), At(0, 0), 0},
		{At(0, 0), At(1, 0), 1},
		{At(0, 0), At(0, 1), 1},
		{At(0, 0), At(1, -1), 1},
		{At(0, 0), At(3, 0), 3},
		{At(0, 0), At(2, -1), 2},
		{At(-1, 1), At(2, -2), 3},
	}
	for _, c := range cases {
		if got := c.a.Distance(c.b); got != c.want {
			t.Errorf("Distance(%v,%v)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNeighborsAreUnitDistance(t *testing.T) {
	a := At(5, -3)
	for i, n := range a.Neighbors() {
		if d := a.Distance(n); d != 1 {
			t.Errorf("neighbor %d at %v has distance %d", i, n, d)
		}
	}
}

func TestOffsetRoundTrip(t *testing.T) {
	layouts := []Layout{OddQ, EvenQ, OddR, EvenR}
	for _, l := range layouts {
		for q := -5; q <= 5; q++ {
			for r := -5; r <= 5; r++ {
				a := At(q, r)
				oc := a.ToOffset(l)
				back := FromOffset(oc, l)
				if back != a {
					t.Errorf("layout %d: %v -> %v -> %v", l, a, oc, back)
				}
			}
		}
	}
}

func TestOffsetKnownValues(t *testing.T) {
	// Spot-check against Red Blob Games' canonical examples.
	cases := []struct {
		a    Axial
		l    Layout
		want OffsetCoord
	}{
		{At(0, 0), OddQ, OffsetCoord{0, 0}},
		{At(1, 0), OddQ, OffsetCoord{1, 0}},   // odd col, shifted down by 0
		{At(1, 1), OddQ, OffsetCoord{1, 1}},
		{At(0, 0), OddR, OffsetCoord{0, 0}},
		{At(0, 1), OddR, OffsetCoord{0, 1}},
	}
	for _, c := range cases {
		if got := c.a.ToOffset(c.l); got != c.want {
			t.Errorf("ToOffset(%v, %d) = %v, want %v", c.a, c.l, got, c.want)
		}
	}
}
