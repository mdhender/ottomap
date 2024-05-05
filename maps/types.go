// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package maps

import (
	"github.com/mdhender/ottomap/coords"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
)

type Map struct {
	Turns   map[string]*Turn
	Units   map[string]*Unit
	Origins map[string]*coords.Grid
	Sorted  struct {
		Turns []*Turn // sorted by Turn.Id
		Units []*Unit // sorted by Unit.Id
		Moves []*Move // sorted by Move.Turn then Move.Unit
		Steps []*Step
		Hexes []*Hex
	}
}

type Turn struct {
	Id    string // year-month
	Year  int
	Month int
}

type Unit struct {
	Id          string
	Parent      *Unit
	StartingHex *Hex
	EndingHex   *Hex
	Moves       []*Move
	Steps       []*Step
}

func (u *Unit) IsGarrison() bool {
	return len(u.Id) == 6 && u.Id[4] == 'g'
}

type Move struct {
	Turn        *Turn
	Unit        *Unit
	StartingHex *Hex
	EndingHex   *Hex
	Steps       []*Step // should be sorted by SeqNo
}

type Step struct {
	Move        *Move
	SeqNo       int
	StartingHex *Hex
	Direction   directions.Direction
	Status      domain.MoveStatus
	EndingHex   *Hex
}

type Hex struct {
	Coords    coords.Map
	Terrain   domain.Terrain
	Neighbors [7]*Hex    // indexed by domain.Direction
	Edges     [7]*Edge   // indexed by domain.Direction
	Contents  []*Content // doesn't include any history
}

func (h *Hex) String() string {
	return h.Coords.String()
}

type Edge struct {
	Feature string
}

type Content struct {
	Kind string
	What string
}
