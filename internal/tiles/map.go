// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tiles

import (
	"github.com/mdhender/ottomap/internal/coords"
	"sort"
)

// Map_t represents a map of tiles.
type Map_t struct {
	// key is the grid location of the tile
	Tiles map[coords.Map]*Tile_t
}

// NewMap creates a new map.
func NewMap() *Map_t {
	return &Map_t{
		Tiles: map[coords.Map]*Tile_t{},
	}
}

func (m *Map_t) Bounds() (upperLeft, lowerRight coords.Map) {
	if m.Length() == 0 {
		return coords.Map{}, coords.Map{}
	}

	for _, tile := range m.Tiles {
		if upperLeft.Column == 0 {
			// assume that we're on the first tile
			upperLeft.Column, upperLeft.Row = tile.Location.Column, tile.Location.Row
			lowerRight.Column, lowerRight.Row = tile.Location.Column, tile.Location.Row
		}
		if tile.Location.Column < upperLeft.Column {
			upperLeft.Column = tile.Location.Column
		}
		if tile.Location.Row < upperLeft.Row {
			upperLeft.Row = tile.Location.Row
		}
		if lowerRight.Column < tile.Location.Column {
			lowerRight.Column = tile.Location.Column
		}
		if lowerRight.Row < tile.Location.Row {
			lowerRight.Row = tile.Location.Row
		}
	}

	return upperLeft, lowerRight
}

func (m *Map_t) Dump() {
	var sortedTiles []*Tile_t
	for _, tile := range m.Tiles {
		sortedTiles = append(sortedTiles, tile)
	}
	sort.Slice(sortedTiles, func(i, j int) bool {
		return sortedTiles[i].Location.GridString() < sortedTiles[j].Location.GridString()
	})
	for _, tile := range sortedTiles {
		tile.Dump()
	}
}

func (m *Map_t) Length() int {
	if m == nil {
		return 0
	}
	return len(m.Tiles)
}

// FetchTile returns the tile at the given location.
// If the tile does not exist, it is created.
func (m *Map_t) FetchTile(location coords.Map) *Tile_t {
	if tile, ok := m.Tiles[location]; ok {
		return tile
	}

	// create a new tile to add to the map
	tile := &Tile_t{Location: location}

	//// visit all the neighbors in turn and ensure that they link back to the tile that we just created.
	//// todo: why do we need to do this? why do we care about the neighbors now?
	//if neighbor, ok := m.Tiles[location.Add(direction.North)]; ok {
	//	neighbor.Neighbors[direction.South] = tile
	//}
	//if neighbor, ok := m.Tiles[location.Add(direction.NorthEast)]; ok {
	//	neighbor.Neighbors[direction.SouthWest] = tile
	//}
	//if neighbor, ok := m.Tiles[location.Add(direction.SouthEast)]; ok {
	//	neighbor.Neighbors[direction.NorthWest] = tile
	//}
	//if neighbor, ok := m.Tiles[location.Add(direction.South)]; ok {
	//	neighbor.Neighbors[direction.North] = tile
	//}
	//if neighbor, ok := m.Tiles[location.Add(direction.SouthWest)]; ok {
	//	neighbor.Neighbors[direction.NorthEast] = tile
	//}
	//if neighbor, ok := m.Tiles[location.Add(direction.NorthWest)]; ok {
	//	neighbor.Neighbors[direction.SouthEast] = tile
	//}

	// add the tile to the map
	m.Tiles[tile.Location] = tile
	// and return it
	return tile
}
