// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
)

type WXX struct {
	buffer *bytes.Buffer

	// range of grids is AA to ZZ.
	// grids are stored in row-column format.
	grids                     [26][26]*Grid
	totalGrids                int
	minGridRow, minGridColumn int
	maxGridRow, maxGridColumn int
}

// GetGrid grids are stored in row-column format
func (w *WXX) GetGrid(row, column int) *Grid {
	if !(0 <= row && row < 26) {
		panic("out of bounds")
	} else if !(0 <= column && column < 26) {
		panic("out of bounds")
	}
	g := w.grids[row][column]
	if g == nil {
		panic("grid not defined")
	}
	return g
}
