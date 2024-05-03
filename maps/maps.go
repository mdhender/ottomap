// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package maps implements a conversion to Worldographer files.
package maps

import (
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"sort"
	"strconv"
	"strings"
)

func New(reports []*domain.Report) (*Map, error) {
	m := &Map{
		Turns: make(map[string]*Turn),
		Units: make(map[string]*Unit),
	}

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
	gridOrigin := ""
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
			if isclan(u.Id) && gridOrigin == "" {
				gridOrigin = u.PrevHex
				if gridOrigin == "N/A" {
					gridOrigin = u.CurrHex
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
	if gridOrigin == "" {
		panic("assert(gridOrigin != \"\"")
	} else if gridOrigin == "N/A" {
		panic("assert(gridOrigin != \"N/A\"")
	}
	clan, ok := m.FetchClan()
	if !ok {
		panic("assert(clan != nil)")
	}
	log.Printf("map: origin hex: clan %q: origin %s\n", clan.Id, gridOrigin)
	column, err := strconv.Atoi(gridOrigin[3:5])
	if err != nil {
		log.Fatalf("map: input: gridOrigin %q: column %v\n", gridOrigin, err)
	}
	row, err := strconv.Atoi(gridOrigin[5:])
	if err != nil {
		log.Fatalf("map: input: gridOrigin %q: row %v\n", gridOrigin, err)
	}
	log.Printf("map: origin hex: clan %q: origin %2d %2d\n", clan.Id, column, row)
	hex := &Hex{Column: column, Row: row}
	m.Sorted.Hexes = append(m.Sorted.Hexes, hex)
	clan.StartingHex = hex

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

// CreateOriginHex creates the origin hex.
// That is the hex that the Clan unit first appears in.
// todo: assumes only one clan!
func (m *Map) CreateOriginHex() error {
	// find the clan
	clan, ok := m.FetchClan()
	if !ok {
		return fmt.Errorf("unable to locate clan")
	}
	// find the clan's first hex
	originHex := clan.StartingHex
	if originHex == nil {
		return fmt.Errorf("clan's starting hex is missing")
	}
	log.Printf("map: clan %q: origin hex (%d, %d)\n", clan.Id, originHex.Column, originHex.Row)
	return nil
}

func (m *Map) FetchClan() (*Unit, bool) {
	for id, unit := range m.Units {
		if isclan(id) {
			return unit, true
		}
	}
	return nil, false
}

func (m *Map) FetchUnit(id string) (*Unit, bool) {
	u, ok := m.Units[id]
	return u, ok
}

func (m *Move) AddStep(d domain.Direction, status domain.MoveStatus) *Step {
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
