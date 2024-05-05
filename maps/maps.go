// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package maps implements a conversion to Worldographer files.
package maps

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/coords"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
	"sort"
	"strings"
)

func New(reports []*domain.Report) (*Map, error) {
	m := &Map{
		Turns:   make(map[string]*Turn),
		Units:   make(map[string]*Unit),
		Origins: make(map[string]*coords.Grid),
	}

	// note: the input must be sorted or the logic for determining the
	// unit's origin will produce incorrect results.

	// add all the report turns to our map
	for _, rpt := range reports {
		_ = m.AddTurn(rpt.Year, rpt.Month)
	}
	// sort the turns
	sort.Slice(m.Sorted.Turns, func(i, j int) bool {
		return m.Sorted.Turns[i].Id < m.Sorted.Turns[j].Id
	})

	// add all the units to the map
	for _, rpt := range reports {
		for _, u := range rpt.Units {
			_ = m.AddUnit(u.Id)
		}
	}
	// sort the units
	sort.Slice(m.Sorted.Units, func(i, j int) bool {
		return m.Sorted.Units[i].Id < m.Sorted.Units[j].Id
	})

	// link units to their parents
	for _, unit := range m.Sorted.Units {
		if isclan(unit.Id) {
			continue
		}
		var pid string
		switch len(unit.Id) {
		case 4: // tribes report to the clan
			pid = "0" + unit.Id[1:4]
		case 6: // elements roll up to the tribe
			pid = unit.Id[:4]
		case 8: // scouts report to the unit
			panic("scouts not implemented")
		default:
			panic("assert(len(unit.Id) in (4,6)")
		}
		parent, ok := m.FetchUnit(pid)
		if !ok {
			panic(fmt.Sprintf("assert(parent.Id == %q)", pid))
		}
		unit.Parent = parent
	}

	// convert the movement data to map data
	for _, rpt := range reports {
		turn, ok := m.FetchTurn(rpt.Year, rpt.Month)
		if !ok {
			panic("assert(turn is ok)")
		}
		log.Printf("map: input: report %s: turn %s\n", rpt.Id, turn.Id)

		for _, u := range rpt.Units {
			if u.Movement == nil {
				continue
			}
			unit, ok := m.FetchUnit(u.Id)
			if !ok {
				panic("assert(unit is ok)")
			}

			// track first known hex (the starting hex) for all units.
			// key is unit.Id, value is prevHex (or currHex if prevHex is N/A)
			if m.Origins[unit.Id] == nil {
				if u.PrevHex != nil {
					m.Origins[unit.Id] = u.PrevHex
				} else if u.CurrHex != nil {
					m.Origins[unit.Id] = u.CurrHex
				} else {
					panic("assert(!(u.PrevHex == nil && u.CurrHex == nil))")
				}
			}

			mv := &Move{
				Turn: turn,
				Unit: unit,
			}
			m.Sorted.Moves = append(m.Sorted.Moves, mv)
			if u.Movement.Follows != "" {
				// do something with this
			}
			for _, ums := range u.Movement.Steps {
				m.Sorted.Steps = append(m.Sorted.Steps, mv.AddStep(ums.Direction, ums.Status))
			}
		}
	}

	// sort the moves
	sort.Slice(m.Sorted.Moves, func(i, j int) bool {
		return m.Sorted.Moves[i].Less(m.Sorted.Moves[j])
	})
	// sort the steps
	sort.Slice(m.Sorted.Steps, func(i, j int) bool {
		return m.Sorted.Steps[i].Less(m.Sorted.Steps[j])
	})

	// now that moves are sorted, add them to their units
	for _, mv := range m.Sorted.Moves {
		mv.Unit.Moves = append(mv.Unit.Moves, mv)
		for _, step := range mv.Steps {
			mv.Unit.Steps = append(mv.Unit.Steps, step)
		}
	}

	// stuff the origin hex into the clan's first move
	clans, ok := m.FetchClans()
	if !ok {
		return nil, fmt.Errorf("map: input: no clans found")
	}
	if len(clans) != 1 {
		log.Printf("map: warning: grid origins broken for multiple clans\n")
	}
	for _, clan := range clans {
		gridOrigin := m.Origins[clan.Id]
		if gridOrigin == nil {
			panic("assert(gridOrigin != nil")
		}
		mc, err := gridOrigin.ToMapCoords()
		if err != nil {
			log.Fatalf("map: input: gridOrigin %q: mc %v\n", gridOrigin, err)
		}

		hex := &Hex{Coords: mc}
		m.Sorted.Hexes = append(m.Sorted.Hexes, hex)
		clan.StartingHex = hex
	}

	return m, nil
}

func (m *Map) AddTurn(year, month int) *Turn {
	id := fmt.Sprintf("%03d-%02d", year, month)
	if t, ok := m.Turns[id]; ok {
		return t
	}
	t := &Turn{Id: id, Year: year, Month: month}
	m.Turns[t.Id] = t
	m.Sorted.Turns = append(m.Sorted.Turns, t)
	return t
}

func (m *Map) FetchTurn(year, month int) (*Turn, bool) {
	id := fmt.Sprintf("%03d-%02d", year, month)
	t, ok := m.Turns[id]
	return t, ok
}

func (m *Map) AddUnit(id string) *Unit {
	if u, ok := m.Units[id]; ok {
		return u
	}
	u := &Unit{Id: id}
	m.Units[u.Id] = u
	m.Sorted.Units = append(m.Sorted.Units, u)
	return u
}

func (m *Map) FetchClans() ([]*Unit, bool) {
	var clans []*Unit
	for id, unit := range m.Units {
		if isclan(id) {
			clans = append(clans, unit)
		}
	}
	return clans, clans != nil
}

func (m *Map) FetchUnit(id string) (*Unit, bool) {
	u, ok := m.Units[id]
	return u, ok
}

func (m *Map) TrackUnit(unit *Unit) error {
	// may need to derive the unit's starting hex from the parent's hex
	// on the turn that the unit was first seen.
	if unit.StartingHex == nil {
		if len(unit.Moves) == 0 {
			log.Printf("map: unit %-8q: track: moves %d\n", unit.Id, len(unit.Moves))
			log.Printf("map: unit %-8q: todo: tracking logic is broken for units with no moves\n", unit.Id)
			log.Printf("map: unit %-8q: %+v\n", unit.Id, *unit)
			if unit.IsGarrison() {
				return cerrs.ErrTrackingGarrison
			}
			return cerrs.ErrUnableToFindStartingHex
		}
		parent, turn := unit.Parent, unit.Moves[0].Turn
		log.Printf("map: unit %-8q: track: parent %-8q: turn %s\n", unit.Id, parent.Id, turn.Id)
		for _, pmv := range parent.Moves {
			if pmv.Turn.Id == turn.Id {
				unit.StartingHex = pmv.StartingHex
				break
			}
		}
		if unit.StartingHex == nil {
			return cerrs.ErrUnableToFindStartingHex
		}
	}
	log.Printf("map: unit %-8q: track: origin %s\n", unit.Id, unit.StartingHex.Coords.GridString())

	prev, curr := unit.StartingHex, unit.StartingHex
	for n, mv := range unit.Moves {
		log.Printf("map: unit %-8q: track: %s %2d\n", unit.Id, mv.Turn.Id, n+1)
		log.Printf("map: unit %-8q: track: %s %2d %+v\n", unit.Id, mv.Turn.Id, n+1, *mv)
		if mv.StartingHex == nil {
			log.Printf("map: unit %-8q: track: %s %2d starting now %+v\n", unit.Id, mv.Turn.Id, n+1, *curr)
			mv.StartingHex = curr
			log.Printf("map: unit %-8q: track: %s %2d %+v\n", unit.Id, mv.Turn.Id, n+1, *mv)
		}
		for _, step := range mv.Steps {
			log.Printf("map: unit %-8q: track: %s %2d %2d %-2s\n", unit.Id, mv.Turn.Id, n+1, step.SeqNo+1, step.Direction)
			step.StartingHex = curr
			neighbor := curr.Neighbors[step.Direction]
			if neighbor == nil {
				// need to create a new hex for the neighbor
				neighbor = &Hex{Coords: curr.Coords.Add(step.Direction)}
				// and link it to the current hex
				switch step.Direction {
				case directions.DNorth:
					neighbor.Neighbors[directions.DSouth] = curr
				case directions.DNorthEast:
					neighbor.Neighbors[directions.DSouthWest] = curr
				case directions.DSouthEast:
					neighbor.Neighbors[directions.DNorthWest] = curr
				case directions.DSouth:
					neighbor.Neighbors[directions.DNorth] = curr
				case directions.DSouthWest:
					neighbor.Neighbors[directions.DNorthEast] = curr
				case directions.DNorthWest:
					neighbor.Neighbors[directions.DSouthEast] = curr
				}
				curr.Neighbors[step.Direction] = neighbor
			}
			log.Printf("map: unit %-8q: track: %s %2d %2d %-2s from %v to %v\n", unit.Id, mv.Turn.Id, n+1, step.SeqNo+1, step.Direction, curr.Coords.GridString(), neighbor.Coords.GridString())
			log.Printf("map: unit %-8q: track: %s %2d %2d %-2s from %v to %v step %v\n", unit.Id, mv.Turn.Id, n+1, step.SeqNo+1, step.Direction, curr.Coords.GridString(), neighbor.Coords.GridString(), *step)
			prev, curr = curr, neighbor
			step.EndingHex = curr
		}
	}
	log.Printf("map: unit %-8q: track: prev %v: curr %v\n", unit.Id, prev, curr)

	log.Printf("map: todo: must carry origin to children\n")

	return nil
}

func (m *Move) AddStep(d directions.Direction, status domain.MoveStatus) *Step {
	step := &Step{
		Move:      m,
		SeqNo:     len(m.Steps),
		Direction: d,
		Status:    status,
	}
	m.Steps = append(m.Steps, step)
	return step
}

func (m *Move) Less(mm *Move) bool {
	if m.Turn.Id < mm.Turn.Id {
		return true
	} else if m.Turn.Id > mm.Turn.Id {
		return false
	}
	return m.Unit.Id < mm.Unit.Id
}

func (s *Step) Less(ss *Step) bool {
	if s.Move.Turn.Id < ss.Move.Turn.Id {
		return true
	} else if s.Move.Turn.Id > ss.Move.Turn.Id {
		return false
	}
	return s.SeqNo < ss.SeqNo
}

func (u *Unit) FirstMove() (*Move, bool) {
	if len(u.Moves) == 0 {
		return nil, false
	}
	return u.Moves[0], true
}

func (u *Unit) FirstStep() (*Step, bool) {
	if len(u.Steps) == 0 {
		return nil, false
	}
	return u.Steps[0], true
}

func isclan(id string) bool {
	return len(id) == 4 && strings.HasPrefix(id, "0")
}
