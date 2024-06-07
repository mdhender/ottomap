// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
)

// one grid on the Worldographer map is 30 columns wide by 21 rows high.
// one grid on the consolidated map is 30 columns wide by 21 rows high.
const columnsPerGrid, rowsPerGrid = 30, 21

type Grid struct {
	id    string // AA ... ZZ
	tiles [columnsPerGrid][rowsPerGrid]Tile
}

func (w *WXX) newGrid(id string) *Grid {
	if len(id) != 2 {
		panic(fmt.Sprintf("assert(len(id) == 2)"))
	}

	// create a new grid with blank tiles and a default elevation of 1.
	grid := &Grid{id: id}
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			grid.tiles[column][row].Elevation = 1
		}
	}

	return grid
}

//// createGrid creates the tiles for a single grid on the larger Tribenet map.
//// The caller is responsible for stitching the grids together in the final map.
//func (w *WXX) createGrid(id string, hexes []*Hex, showGridCoords, showGridNumbers bool) (*Grid, error) {
//	// create a new grid with blank tiles and a default elevation of 1.
//	grid := &Grid{id: id}
//	for column := 0; column < columnsPerGrid; column++ {
//		for row := 0; row < rowsPerGrid; row++ {
//			grid.tiles[column][row].Elevation = 1
//		}
//	}
//
//	// convert the grid hexes to tiles
//	for _, hex := range hexes {
//		col, row := hex.Coords.Column-1, hex.Coords.Row-1
//		grid.tiles[col][row].Terrain = hex.Terrain
//		// todo: add the missing terrain types here.
//		switch grid.tiles[col][row].Terrain {
//		case domain.TLake:
//			grid.tiles[col][row].Elevation = -1
//		case domain.TOcean:
//			grid.tiles[col][row].Elevation = -3
//		case domain.TPrairie:
//			grid.tiles[col][row].Elevation = 1_000
//		}
//		grid.tiles[col][row].Features = hex.Features
//		if showGridCoords {
//			grid.tiles[col][row].Features.Coords = fmt.Sprintf("%s %02d%02d", hex.Grid, hex.Coords.Column, hex.Coords.Row)
//		} else if showGridNumbers {
//			grid.tiles[col][row].Features.Coords = fmt.Sprintf("%02d%02d", hex.Coords.Column, hex.Coords.Row)
//		}
//	}
//
//	return grid, nil
//}

func (g *Grid) addCoords() {
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			if g.tiles[column][row].created != "" {
				g.tiles[column][row].Features.CoordsLabel = fmt.Sprintf("%s %02d%02d", g.id, column+1, row+1)
			}
		}
	}
}

func (g *Grid) addNumbers() {
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			if g.tiles[column][row].created != "" {
				g.tiles[column][row].Features.NumbersLabel = fmt.Sprintf("%02d%02d", column+1, row+1)
			}
		}
	}
}

func (g *Grid) addTile(turnId string, hex *Hex) error {
	column, row := hex.Offset.Column-1, hex.Offset.Row-1

	tile := &g.tiles[column][row]
	tile.updated = turnId

	// does tile already exist in the grid?
	alreadyExists := tile.created != ""

	// if it does, verify that the terrain has not changed
	if alreadyExists && tile.Terrain != hex.Terrain {
		log.Printf("error: turn %q: tile %q\n", turnId, tile.GridCoords)
		log.Printf("error: turn %q: hex  %q\n", turnId, hex.GridCoords)
		panic("assert(tile.Terrain == hex.Terrain)")
	}

	// if it doesn't, set up the terrain
	if !alreadyExists {
		tile.created = turnId
		tile.Terrain = hex.Terrain
		tile.Elevation = 1
		switch tile.Terrain {
		case domain.TAridTundra,
			domain.TBrush,
			domain.TBrushHills,
			domain.TConiferHills,
			domain.TDeciduousForest,
			domain.TDeciduousHills,
			domain.TDesert,
			domain.TGrassyHills,
			domain.TGrassyHillsPlateau,
			domain.THighSnowyMountains,
			domain.TJungle,
			domain.TJungleHills,
			domain.TLowAridMountains,
			domain.TLowConiferMountains,
			domain.TLowJungleMountains,
			domain.TLowSnowyMountains,
			domain.TLowVolcanicMountains,
			domain.TPrairie,
			domain.TPrairiePlateau,
			domain.TRockyHills,
			domain.TTundra:
			tile.Elevation = 1_250
		case domain.TLake:
			tile.Elevation = -1
		case domain.TOcean:
			tile.Elevation = -3
		case domain.TPolarIce:
			tile.Elevation = 10
		case domain.TSwamp:
			tile.Elevation = 1
		default:
			log.Printf("grid: addTile: unknown terrain type %d %q", hex.Terrain, hex.Terrain.String())
			panic(fmt.Sprintf("assert(hex.Terrain != %d)", hex.Terrain))
		}
	}

	tile.Features = hex.Features

	return nil
}

func gridIdToRowColumn(id string) (row, column int) {
	if len(id) != 2 {
		log.Printf("error: invalid grid id %q\n", id)
		panic(fmt.Sprintf("assert(len(id) == 3)"))
	}
	row, column = int(id[0]-'A'), int(id[1]-'A')
	if row < 0 || row > 25 {
		log.Printf("error: invalid grid row %q\n", id)
		panic(fmt.Sprintf("assert(0 < row <= 25)"))
	} else if column < 0 || column > 25 {
		log.Printf("error: invalid grid column %q\n", id)
		panic(fmt.Sprintf("assert(0 < column <= 25)"))
	}
	return row, column
}
