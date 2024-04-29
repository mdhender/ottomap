// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package hexes_test

import (
	"github.com/mdhender/ottomap/hexes"
	"testing"
)

func TestStringToGridCoordinates(t *testing.T) {
	tests := []struct {
		id         int
		input      string
		wantResult hexes.GridCoords
		wantOK     bool
	}{
		// Add as many test cases as you want
		{1001, "AA 0101", hexes.GridCoords{Row: 0, Column: 0, GridColumn: 0, GridRow: 0}, true},
		{1002, "AZ 3001", hexes.GridCoords{Row: 0, Column: 25, GridColumn: 29, GridRow: 0}, true},
		{1003, "ZA 0121", hexes.GridCoords{Row: 25, Column: 0, GridColumn: 0, GridRow: 20}, true},
		{1004, "ZZ 3021", hexes.GridCoords{Row: 25, Column: 25, GridColumn: 29, GridRow: 20}, true},
		{2101, "AA0101", hexes.GridCoords{}, false},
		{2102, "AA-0101", hexes.GridCoords{}, false},
		{2201, "aB 1230", hexes.GridCoords{}, false},
		{2301, "Ab 1230", hexes.GridCoords{}, false},
		{2401, "AA 0001", hexes.GridCoords{}, false},
		{2402, "AA 3100", hexes.GridCoords{}, false},
		{2501, "AA 0100", hexes.GridCoords{}, false},
		{2502, "AA 0122", hexes.GridCoords{}, false},
	}

	for _, tc := range tests {
		gotResult, gotOk := hexes.GridCoordsFromString(tc.input)
		if gotOk != tc.wantOK {
			t.Errorf("%d: ok        : got %v, want %v", tc.id, gotOk, tc.wantOK)
		} else if tc.wantOK {
			checkString := true
			if gotResult.Row != tc.wantResult.Row {
				checkString = false
				t.Errorf("%d: %q: row       : got %6d, want %6d", tc.id, tc.input, gotResult.Row, tc.wantResult.Row)
			}
			if gotResult.Column != tc.wantResult.Column {
				checkString = false
				t.Errorf("%d: %q: column    : got %6d, want %6d", tc.id, tc.input, gotResult.Column, tc.wantResult.Column)
			}
			if gotResult.GridColumn != tc.wantResult.GridColumn {
				checkString = false
				t.Errorf("%d: %q: gridColumn: got %6d, want %6d", tc.id, tc.input, gotResult.GridColumn, tc.wantResult.GridColumn)
			}
			if gotResult.GridRow != tc.wantResult.GridRow {
				checkString = false
				t.Errorf("%d: %q: gridRow   : got %6d, want %6d", tc.id, tc.input, gotResult.GridRow, tc.wantResult.GridRow)
			}
			if checkString && gotResult.String() != tc.input {
				t.Errorf("%d: %q: got %q, want %q", tc.id, tc.input, gotResult.String(), tc.input)
			}
		}
	}
}
