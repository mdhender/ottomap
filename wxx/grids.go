// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
)

func (w *WXX) CreateGrid(hexes []*Hex, addGridCoords bool) ([][]Tile, error) {
	// one grid on the Worldographer map is 30 columns wide by 21 rows high.
	const columns, rows = 30, 21

	// create a new grid.
	grid := make([][]Tile, columns)

	for column := 0; column < columns; column++ {
		grid[column] = make([]Tile, rows)
		for row := 0; row < rows; row++ {
			grid[column][row] = Tile{}
			tile := &grid[column][row]
			tile.Elevation = 1
			if addGridCoords {
				tile.Label = &Label{Text: fmt.Sprintf("%02d%02d", column+1, row+1)}
			}
		}
	}

	return grid, nil
}
