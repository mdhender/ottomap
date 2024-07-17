// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/resources"
	"github.com/mdhender/ottomap/internal/terrain"
)

// Hex is a hex on the Tribenet map.
type Hex struct {
	Location   coords.Map // coordinates from the turn report
	Offset     Offset     // coordinates in a grid hex are one-based
	Terrain    terrain.Terrain_e
	WasScouted bool
	WasVisited bool
	Features   Features
}

func (h *Hex) Grid() string {
	return h.Location.GridId()
}

// Tile is a hex on the Worldographer map.
type Tile struct {
	created    string     // turn id when the tile was created
	updated    string     // turn id when the tile was updated
	Location   coords.Map // original grid coordinates
	Terrain    terrain.Terrain_e
	Elevation  int
	IsIcy      bool
	IsGMOnly   bool
	Resources  Resources
	WasScouted bool
	WasVisited bool
	Features   Features
}

// Features are things to display on the map
type Features struct {
	Edges struct {
		Ford      []direction.Direction_e
		Pass      []direction.Direction_e
		River     []direction.Direction_e
		StoneRoad []direction.Direction_e
	}

	// set label for either Coords or Numbers, not both
	CoordsLabel  string
	NumbersLabel string

	IsOrigin    bool // true for the clan's origin hex
	Label       *Label
	Resources   []resources.Resource_e
	Settlements []*parser.Settlement_t // name of settlement
}

type Resources struct {
	Animal int
	Brick  int
	Crops  int
	Gems   int
	Lumber int
	Metals int
	Rock   int
}

// Offset captures the layout.
// Are these one-based or zero-based?
type Offset struct {
	Column int
	Row    int
}