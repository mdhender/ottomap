// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/direction"
)

// Map represents coordinates (column and row) on the map.
// They start at 0,0 and increase to the right and down.
type Map struct {
	Column int
	Row    int
}

func (m Map) GridId() string {
	return m.ToGrid().String()[:2]
}

func (m Map) GridColumnRow() (int, int) {
	return m.Column%30 + 1, m.Row%21 + 1
}

func (m Map) GridString() string {
	return m.ToGrid().String()
}

func (m Map) String() string {
	return fmt.Sprintf("(%d, %d)", m.Column, m.Row)
}

func (m Map) Add(d direction.Direction_e) Map {
	var vec [2]int
	if m.Column%2 == 0 { // even column
		vec = EvenColumnVectors[d]
	} else { // odd column
		vec = OddColumnVectors[d]
	}
	return Map{
		Column: m.Column + vec[0],
		Row:    m.Row + vec[1],
	}
}

func (m Map) ToGrid() Grid {
	return Grid{
		BigMapRow:    m.Row / 21,
		BigMapColumn: m.Column / 30,
		GridColumn:   m.Column%30 + 1,
		GridRow:      m.Row%21 + 1,
	}
}

func (m Map) ToHex() string {
	bigMapRow := m.Row / 21
	bigMapColumn := m.Column / 30
	littleMapColumn := m.Column%30 + 1
	littleMapRow := m.Row%21 + 1

	return fmt.Sprintf("%c%c %02d%02d", bigMapRow+'A', bigMapColumn+'A', littleMapColumn, littleMapRow)
}
