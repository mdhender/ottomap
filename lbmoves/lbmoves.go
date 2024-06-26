// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package lbmoves implements Land Based Movement parsing and map generation.
package lbmoves

import (
	"fmt"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/items"
	"strconv"
	"strings"
)

// Land Based Movement is a series of steps.
// Each step is an attempt to move in a certain direction.
// The attempt may succeed or fail. If it fails, it may be
// because the unit was blocked by an edge feature (a River),
// or the unit was exhausted (it didn't have enough MPs),
// or the unit is not allowed to enter the terrain.

// MovementResults is the set of hex reports from a single movement results line.
type MovementResults struct {
	TurnId                  string
	UnitId                  string
	StartingGridCoordinates string // grid coordinates before the unit moves
	EndingGridCoordinates   string // grid coordinates after the unit moves

	Follows   string             `json:"follows,omitempty"` // unit id this unit follows
	Followers []*MovementResults `json:"-"`                 // link to following units' movement

	// ParentMR is set only if this unit did not move and doesn't have a starting grid coordinate.
	ParentMR *MovementResults `json:"-"`

	// hex reports
	StatusLine      *Step     `json:"statusLine,omitempty"`
	MovementReports []*Step   `json:"movementReports,omitempty"`
	ScoutReports    [][]*Step `json:"scoutReports,omitempty"`
}

func (m *MovementResults) Id() string {
	return fmt.Sprintf("%s.%s", m.TurnId, m.UnitId)
}

// Step is one step of a Land Based Movement.
type Step struct {
	TurnId       string `json:"turnId,omitempty"`
	UnitId       string `json:"unitId,omitempty"`
	StartGridHex string `json:"startGridHex,omitempty"`

	// Attempted direction is the direction the unit tried to move.
	// It will be Unknown if the unit stays in place.
	// When the unit fails to move, this will be derived from the failed results.
	Attempted directions.Direction `json:"attempted,omitempty"`

	// Result is the result of the step.
	// The attempt may succeed or fail; this captures the reasons.
	Result Result `json:"result,omitempty"`

	// properties below are set even if the step failed.
	// that means they may be for the hex where the unit started.

	GridHex        string          `json:"gridHex,omitempty"`
	Terrain        domain.Terrain  `json:"terrain,omitempty"`
	BlockedBy      *BlockedByEdge  `json:"blockedBy,omitempty"`
	Edges          []*Edge         `json:"edges,omitempty"`
	Exhausted      *Exhausted      `json:"exhausted,omitempty"`
	Follows        string          `json:"follows,omitempty"` // unit id this unit follows
	Neighbors      []*Neighbor     `json:"neighbors,omitempty"`
	ProhibitedFrom *ProhibitedFrom `json:"prohibitedFrom,omitempty"`
	Resources      domain.Resource `json:"resources,omitempty"`
	Settlement     *Settlement     `json:"settlement,omitempty"`
	Units          []string        `json:"units,omitempty"` // unit ids
}

// BlockedByEdge is returned when a step fails because the unit was blocked by an edge feature.
type BlockedByEdge struct {
	Direction directions.Direction
	Edge      domain.Edge
}

func (b *BlockedByEdge) String() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("b(%s-%s)", b.Direction, b.Edge)
}

type DidNotReturn struct{}

// DirectionTerrain is the first component returned from a successful step.
type DirectionTerrain struct {
	Direction directions.Direction
	Terrain   domain.Terrain
}

func (d DirectionTerrain) String() string {
	return fmt.Sprintf("%s-%s", d.Direction, d.Terrain)
}

func (d *DidNotReturn) String() string {
	if d == nil {
		return ""
	}
	return "did not return"
}

// Edge is an edge feature that the unit sees in the current hex.
type Edge struct {
	Direction directions.Direction
	Edge      domain.Edge
}

func (e *Edge) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", e.Direction, e.Edge)
}

// Exhausted is returned when a step fails because the unit was exhausted.
type Exhausted struct {
	Direction directions.Direction
	Terrain   domain.Terrain
}

func (e *Exhausted) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("x(%s-%s)", e.Direction, e.Terrain)
}

type FoundNothing struct{}

func (f *FoundNothing) String() string {
	if f == nil {
		return ""
	}
	return "nothing of interest found"
}

type FoundUnit struct {
	Id UnitID
}

// Neighbor is the terrain in a neighboring hex that the unit from the current hex.
type Neighbor struct {
	Direction directions.Direction `json:"direction,omitempty"`
	Terrain   domain.Terrain       `json:"terrain,omitempty"`
}

func (n *Neighbor) String() string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", n.Direction, n.Terrain)
}

type NoGroupsFound struct{}

// ProhibitedFrom is returned when a step fails because the unit is not allowed to enter the terrain.
type ProhibitedFrom struct {
	Direction directions.Direction `json:"direction,omitempty"`
	Terrain   domain.Terrain       `json:"terrain,omitempty"`
}

func (p *ProhibitedFrom) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("p(%s-%s)", p.Direction, p.Terrain)
}

type RandomEncounter struct {
	Quantity int
	Item     items.Item
}

func (r *RandomEncounter) String() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("r(%d-%s)", r.Quantity, r.Item)
}

// Settlement is a settlement that the unit sees in the current hex.
type Settlement struct {
	Name string
}

func (s *Settlement) String() string {
	if s == nil {
		return ""
	}
	return s.Name
}

type UnitID string

func (u UnitID) ParentId() string {
	return ParentId(string(u))
}

func ParentId(unitId string) string {
	switch len(unitId) {
	case 4:
		if strings.HasPrefix(unitId, "0") {
			return ""
		}
		return "0" + unitId[1:]
	case 6:
		return unitId[:4]
	}
	panic(fmt.Sprintf("assert(len(unitId) != %d)", len(unitId)))
}

func NextTurnId(turnId string) string {
	if len(turnId) != 7 {
		panic(fmt.Sprintf("assert(len(turnId) != %d)", len(turnId)))
	}
	yyyy, mm, ok := strings.Cut(turnId, "-")
	if !ok {
		panic("assert(ok)")
	}
	year, err := strconv.Atoi(yyyy)
	if err != nil {
		panic(err)
	}
	month, err := strconv.Atoi(mm)
	if err != nil {
		panic(err)
	}
	month++
	if month == 13 {
		year, month = year+1, 1
	}
	return fmt.Sprintf("%04d-%02d", year, month)
}
