// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"github.com/mdhender/ottomap/domain"
)

// CreateGrid creates the tiles for a single grid on the larger Tribenet map.
// The caller is responsible for stitching the grids together in the final map.
func (w *WXX) CreateGrid(hexes []*Hex, showGridCoords, showGridNumbers bool) ([][]Tile, error) {
	// one grid on the Worldographer map is 30 columns wide by 21 rows high.
	const columns, rows = 30, 21

	// create a new grid with blank tiles.
	grid := make([][]Tile, columns)
	for column := 0; column < columns; column++ {
		grid[column] = make([]Tile, rows)
		for row := 0; row < rows; row++ {
			grid[column][row] = Tile{}
			tile := &grid[column][row]
			tile.Elevation = 1
		}
	}

	// convert the grid hexes to tiles
	for _, hex := range hexes {
		tile := &grid[hex.Coords.Column-1][hex.Coords.Row-1]
		tile.Terrain = hex.Terrain
		// todo: add the missing terrain types here.
		switch tile.Terrain {
		case domain.TLake:
			tile.Elevation = -1
		case domain.TOcean:
			tile.Elevation = -3
		case domain.TPrairie:
			tile.Elevation = 1_000
		}
		tile.Features = hex.Features
		if showGridCoords {
			tile.Features.Coords = fmt.Sprintf("%s %02d%02d", hex.Grid, hex.Coords.Column, hex.Coords.Row)
		} else if showGridNumbers {
			tile.Features.Coords = fmt.Sprintf("%02d%02d", hex.Coords.Column, hex.Coords.Row)
		}
	}

	return grid, nil
}
