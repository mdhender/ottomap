// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package report

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
)

type Unit struct {
	Id       string `json:"id,omitempty"`
	Parent   *Unit  `json:"-"`
	ParentId string `json:"parentId,omitempty"`
	Turn     struct {
		Year  int `json:"year,omitempty"`
		Month int `json:"month,omitempty"`
	} `json:"turn,omitempty"`
	Start   string  `json:"start,omitempty"`
	Follows string  `json:"follows,omitempty"`
	Moves   []*Move `json:"moves,omitempty"`
	End     string  `json:"end,omitempty"`
	Status  *Found  `json:"status,omitempty"`
	sortKey string
}

func (u *Unit) IsClan() bool {
	return len(u.Id) == 4 && u.Id[0] == '0'
}
func (u *Unit) IsCourier() bool {
	return len(u.Id) == 6 && u.Id[4] == 'c'
}
func (u *Unit) IsElement() bool {
	return len(u.Id) == 6 && u.Id[4] == 'e'
}
func (u *Unit) IsFleet() bool {
	return len(u.Id) == 6 && u.Id[4] == 'f'
}
func (u *Unit) IsGarrison() bool {
	return len(u.Id) == 6 && u.Id[4] == 'g'
}
func (u *Unit) IsTribe() bool {
	return len(u.Id) == 4
}

func (u *Unit) SortKey() string {
	if u.sortKey == "" {
		u.sortKey = fmt.Sprintf("%04d.%02d.%s", u.Turn.Year, u.Turn.Month, u.Id)
	}
	return u.sortKey
}

type Move struct {
	Seq       int    `json:"seq"`
	Direction string `json:"direction,omitempty"`
	Blocked   bool   `json:"blocked,omitempty"`
	Exhausted bool   `json:"exhausted,omitempty"`
	Still     bool   `json:"still,omitempty"`
	Found     *Found `json:"found,omitempty"`
}

type Found struct {
	Terrain    domain.Terrain `json:"terrain,omitempty"`
	Edges      []*Edge        `json:"edges,omitempty"`
	Units      []string       `json:"units,omitempty"`
	Settlement string         `json:"settlement,omitempty"`
}

type Edge struct {
	Direction directions.Direction `json:"direction,omitempty"`
	Edge      domain.Edge          `json:"edge,omitempty"`
	Terrain   domain.Terrain       `json:"terrain,omitempty"`
}

func ParseSection(section [][]byte, showSlugs bool) (*Unit, error) {
	lines := section
	log.Printf("parse: section: lines %8d\n", len(lines))

	if showSlugs {
		var slug []byte
		for n, line := range lines {
			if n >= 4 {
				break
			}
			if len(line) < 73 {
				slug = append(slug, line...)
			} else {
				slug = append(slug, line[:73]...)
				slug = append(slug, '.', '.', '.')
			}
			slug = append(slug, '\n')
		}
		log.Printf("parse: section: lines %8d\n%s\n", len(lines), string(slug))
	}

	// things are easier to think about if we split the input into the various chunks
	var chunks struct {
		Header  [][]byte
		Follows []byte
		Moves   []byte
		Scout   [][]byte
		Status  []byte
	}

	unit := &Unit{}

	var prevHex, currHex string

	// start with the header lines
	if len(lines) < 2 {
		log.Printf("parse: section: missing header\n")
		return nil, cerrs.ErrNotATurnReport
	}
	chunks.Header = [][]byte{lines[0], lines[1]}
	if v, err := Parse("header", lines[0]); err != nil {
		log.Printf("parse: section: header: %q\n", string(lines[0]))
		log.Fatalf("parse: section: header: %v\n", err)
	} else if tl, ok := v.(*TribeLocation); !ok {
		log.Fatalf("parse: section: header: expected *TribeLocation, got %T\n", v)
	} else {
		log.Printf("parse: section: header: %T\n", tl)
		unit.Id = tl.UnitId
		if tl.Prev == "" && tl.Curr == "" {
			log.Printf("parse: section: header: missing prev and curr locations\n")
			return nil, cerrs.ErrNotATurnReport
		}
		unit.Start, unit.End = tl.Prev, tl.Curr
		if unit.Start == "" {
			log.Printf("parse: section: warning: substituting prevHex with currHex\n")
			unit.Start = unit.End
		} else if unit.End == "" {
			log.Printf("parse: section: warning: substituting currHex with prevHex\n")
			unit.End = unit.Start
		}
	}
	log.Printf("parse: section: header: found unit %q\n", unit.Id)
	log.Printf("parse: section: header: found prev %q\n", prevHex)
	log.Printf("parse: section: header: found curr %q\n", currHex)

	if v, err := Parse("header", lines[1]); err != nil {
		log.Printf("parse: section: header: %q\n", string(lines[1]))
		log.Fatalf("parse: section: header: %v\n", err)
	} else if dt, ok := v.(*Date); !ok {
		log.Fatalf("parse: section: header: expected *Date, got %T\n", v)
	} else {
		log.Printf("parse: section: header: %T\n", dt)
		unit.Turn.Year = dt.Year
		unit.Turn.Month = dt.Month
	}
	log.Printf("parse: section: header: found year  %4d\n", unit.Turn.Year)
	log.Printf("parse: section: header: found month %4d\n", unit.Turn.Month)

	// now that we know the unit id, we can extract the remaining lines that we are interested in
	followsLine := []byte("Tribe Follows ")
	movesLine := []byte("Tribe Movement: ")
	var scoutLines [8][]byte
	for sid := 0; sid < 8; sid++ {
		scoutLines[sid] = []byte(fmt.Sprintf("Scout %d:Scout  ", sid+1))
	}
	statusLine := []byte(fmt.Sprintf("%s Status: ", unit.Id))
	for _, line := range lines {
		if bytes.HasPrefix(line, followsLine) {
			if chunks.Follows != nil {
				return nil, cerrs.ErrMultipleFollowsLines
			}
			chunks.Follows = line
		} else if bytes.HasPrefix(line, movesLine) {
			if chunks.Moves != nil {
				return nil, cerrs.ErrMultipleMovementLines
			}
			chunks.Moves = line
		} else if bytes.HasPrefix(line, []byte{'S', 'c', 'o', 'u', 't'}) {
			for sid := 0; sid < 8; sid++ {
				if bytes.HasPrefix(line, scoutLines[sid]) {
					chunks.Scout = append(chunks.Scout, line)
					break
				}
			}
		} else if bytes.HasPrefix(line, statusLine) {
			if chunks.Status != nil {
				return nil, cerrs.ErrMultipleStatusLines
			}
			chunks.Status = line
		}
	}
	if chunks.Follows == nil && chunks.Moves == nil {
		log.Printf("parse: section: warning: missing follows and movement lines\n")
	}
	if len(chunks.Scout) > 8 {
		return nil, cerrs.ErrTooManyScoutLines
	}
	if chunks.Status == nil {
		return nil, cerrs.ErrMissingStatusLine
	}

	// parse the status line first. we'll need the information from it if the unit doesn't move this turn
	log.Printf("parse: section: status %q\n", string(chunks.Status))
	if v, err := Parse("status", chunks.Status); err != nil {
		log.Printf("parse: section: status: %q\n", string(chunks.Status))
		log.Fatalf("parse: section: status: %v\n", err)
	} else if st, ok := v.(*Status); !ok {
		log.Fatalf("parse: section: status: expected *Status, got %T\n", v)
	} else {
		log.Printf("parse: section: status: %T\n", st)
		log.Printf("parse: section: status: %+v\n", *st)
		// extract the found items
		if unit.Status == nil {
			unit.Status = &Found{}
		}
		unit.Status.Terrain = st.Terrain
		for _, f := range st.Found {
			log.Printf("parse: section: status: found %+v\n", *f)
			if f.UnitId != "" {
				if f.UnitId != unit.Id { // only add if it isn't the unit we are parsing
					unit.Status.Units = append(unit.Status.Units, f.UnitId)
				}
			}
		}
	}

	// parse the unit's movement. this will be either a follows or a movement line.
	// we tested for the presence of both above, so we know we don't have both.
	if chunks.Follows != nil { // unit followed another unit this turn
		panic("follows is not implemented")
	} else if chunks.Moves != nil { // unit moved this turn
		panic("movement is not implemented")
	} else { // unit didn't move this turn
		// leave moves empty, and fill in the unit's movement later
	}

	return unit, nil
}

func (u *Unit) Walk() {
	// if the unit didn't move this turn, use the information from the
	// status line to fill in the unit's movement.
	if len(u.Moves) == 0 {
		u.Moves = []*Move{&Move{
			Seq:   1,
			Still: true,
			Found: u.Status,
		}}
		u.End = u.Start
		return
	}

	// unit moved this turn
	for n, m := range u.Moves { // TODO: implement
		log.Printf("unit %s: walk: %d: %+v\n", u.Id, n+1, *m)
	}
}
