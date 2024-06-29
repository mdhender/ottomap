// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/items"
	"github.com/mdhender/ottomap/internal/resources"
	"github.com/mdhender/ottomap/internal/results"
	"github.com/mdhender/ottomap/internal/terrain"
	"strings"
)

// These are the types returned from the parser and parsing functions.

// Turn_t represents a single turn identified by year and month.
type Turn_t struct {
	Id    string // YYYY-MM
	Year  int
	Month int

	// UnitMoves holds the units that moved in this turn
	UnitMoves map[string]*Moves_t
}

// Moves_t represents the results for a unit that moves and reports in a turn.
// There will be one instance of this struct for each turn the unit moves in.
type Moves_t struct {
	Id int // unit that is moving

	// all the moves made this turn
	Moves []*Move_t

	// Scouts are optional and move at the end of the turn
	Scouts []*Scout_t

	// FromHex is the hex the unit starts the move in.
	// This could be "N/A" if the unit was created this turn.
	// In that case, we will populate it when we know where the unit started.
	FromHex string

	// ToHex is the hex is unit ends the movement in.
	// This should always be set from the turn report.
	// It might be the same as the FromHex if the unit stays in place or fails to move.
	ToHex string
}

// Move_t represents a single move by a unit.
// The move can be follows, goes to, stay in place, or attempt to advance a direction.
// The move will fail, succeed, or the unit can simply vanish without a trace.
type Move_t struct {
	// the types of movement that a unit can make.
	Advance direction.Direction_e // set only if the unit is advancing
	Follows string                // id of the unit being followed
	GoesTo  string                // hex teleporting to
	Still   bool                  // true if the unit is not moving (garrison) or a status entry

	// Result should be failed, succeeded, or vanished
	Result results.Result_e

	Report *Report_t // all observations made by the unit at the end of this move
}

// Report_t represents the observations made by a unit.
// All reports are relative to the hex that the unit is reporting from.
type Report_t struct {
	// permanent items in this hex
	Terrain terrain.Terrain_e
	Borders []*Border_t

	// transient items in this hex
	Encounters  []string // other units in the hex
	Items       []*FoundItem_t
	Resources   []resources.Resource_e
	Settlements []*Settlement_t
}

// Border_t represents details about the hex border.
type Border_t struct {
	Direction direction.Direction_e
	// Edge is set if there is an edge feature like a river or pass
	Edge edges.Edge_e
	// Terrain is set if the neighbor is observable from this hex
	Terrain terrain.Terrain_e
}

// FoundItem_t represents items discovered by Scouts as they pass through a hex.
type FoundItem_t struct {
	Quantity int
	Item     items.Item_e
}

func (f *FoundItem_t) String() string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("found(%d-%s)", f.Quantity, f.Item)
}

// Scout_t represents a scout sent out by a unit.
type Scout_t struct {
	No    int // usually from 1..8
	Moves []*Move_t
}

// Settlement_t is a settlement that the unit sees in the current hex.
type Settlement_t struct {
	Name string
}

func (s *Settlement_t) String() string {
	if s == nil {
		return ""
	}
	return s.Name
}

// helper functions

// MergeBorders adds a new border to the list if it's not already in the list
func MergeBorders(list []*Border_t, b *Border_t) []*Border_t {
	if b == nil {
		return list
	} else if list == nil {
		return []*Border_t{b}
	}
	for _, l := range list {
		if l.Direction != b.Direction {
			continue
		} else if l.Edge == b.Edge && l.Terrain == b.Terrain {
			return list
		}
	}
	return append(list, b)
}

// MergeEncounters adds a new encounter to the list if it's not already in the list
func MergeEncounters(list []string, e string) []string {
	if e == "" {
		return list
	} else if list == nil {
		return []string{e}
	}
	for _, l := range list {
		if l == e {
			return list
		}
	}
	return append(list, e)
}

// MergeItems adds an item to the list. If it is already in the list, the quantity is updated.
func MergeItems(list []*FoundItem_t, f *FoundItem_t) []*FoundItem_t {
	if f == nil {
		return list
	} else if list == nil {
		return []*FoundItem_t{f}
	}
	for _, l := range list {
		if l.Item != f.Item {
			l.Quantity += f.Quantity
			return list
		}
	}
	return append(list, f)
}

// MergeResources adds a new resource to the list if it's not already in the list
func MergeResources(list []resources.Resource_e, r resources.Resource_e) []resources.Resource_e {
	if r == resources.None {
		return list
	} else if list == nil {
		return []resources.Resource_e{r}
	}
	for _, l := range list {
		if l == r {
			return list
		}
	}
	return append(list, r)
}

// MergeSettlements adds a new settlement to the list if it's not already in the list
func MergeSettlements(list []*Settlement_t, s *Settlement_t) []*Settlement_t {
	if s == nil {
		return list
	} else if list == nil {
		return []*Settlement_t{s}
	}
	for _, l := range list {
		if strings.ToLower(l.Name) == strings.ToLower(s.Name) {
			return list
		}
	}
	return append(list, s)
}
