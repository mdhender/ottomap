// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turn_reports

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/coords"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/turn_reports/headers"
	"github.com/mdhender/ottomap/parsers/turn_reports/locations"
	"github.com/mdhender/ottomap/parsers/turn_reports/movements"
	"github.com/mdhender/ottomap/parsers/turn_reports/sections"
	"log"
	"os"
	"strconv"
)

// Parse splits the input into individual sections.
func Parse(rpf *domain.ReportFile, debugSlugs, captureRawText bool) ([]*domain.ReportSection, error) {
	// read the entire input file into memory
	input, err := os.ReadFile(rpf.Path)
	if err != nil {
		//log.Printf("clans: %s: %8d: %v\n", rpf.Id, len(input), err)
		return nil, err
	}

	// extract the header from the input so that we can verify the settings
	header, err := headers.ParseHeader(rpf.Id, bytes.Split(input, []byte{'\n'}))
	if err != nil {
		return nil, err
	} else if header.ClanId != fmt.Sprintf("%04d", rpf.Clan) {
		log.Fatalf("clans: %s: mismatched clan: id %q: want %q\n", rpf.Id, header.ClanId, fmt.Sprintf("%04d", rpf.Clan))
	} else if header.Game.Year != fmt.Sprintf("%03d", rpf.Year) {
		log.Fatalf("clans: %s: mismatched clan: year %q: want %q\n", rpf.Id, header.Game.Year, fmt.Sprintf("%03d", rpf.Year))
	} else if header.Game.Month != fmt.Sprintf("%02d", rpf.Month) {
		log.Fatalf("clans: %s: mismatched clan: month %q, want %q\n", rpf.Id, header.Game.Month, fmt.Sprintf("%02d", rpf.Month))
	}

	ss, separator := Split(input)
	//log.Printf("clans: %s: sections %d\n", rpf.Id, len(ss))
	if separator == nil {
		log.Printf("clans: %s: missing separator\n", rpf.Id)
		return nil, err
	}
	// log.Printf("clans: %s: sections %3d: separator %q\n", rpf.Id, len(ss), separator)

	// capture only the unit sections
	var rss []*domain.ReportSection
	for n, section := range ss {
		//log.Printf("clans: %s: section %2d/%2d: %8d\n", rpf.Id, n+1, len(ss), len(section))
		if debugSlugs {
			var slug []byte
			if len(section) > 40 {
				slug = section[:40]
			} else {
				slug = section
			}
			log.Printf("%3d: %6d: %q\n", n+1, len(section), string(slug))
		}

		lines := bytes.Split(section, []byte{'\n'})
		if len(lines) == 0 { // skip empty sections
			continue
		}

		////log.Printf("turn_reports: parse: parsing %s\n", string(lines[0]))
		//for n, line := range lines {
		//	log.Printf("section: line %3d: %s\n", n+1, string(line))
		//}

		rs := &domain.ReportSection{}
		rs.Id, rs.Type = sections.ParseSectionType(lines)
		//log.Printf("clans: %s: rs %q %q\n", rpf.Id, rs.Id, rs.Type)
		if rs.Type != domain.RSUnit {
			continue
		}

		unit := &domain.ReportUnit{}
		if captureRawText {
			unit.Raw = &domain.ReportUnitRaw{}
		}
		rs.Unit = unit
		unit.Id, unit.Type = sections.ParseUnitType(lines[0])

		location := sections.ParseLocationLine(rs.Id, lines)
		if unit.Raw != nil {
			unit.Raw.Location = string(location)
		}
		if hi, err := locations.Parse(rpf.Id, location); err != nil {
			log.Printf("turn_reports: %s: %s: location %q: parse %v\n", rpf.Id, unit.Id, string(location), err)
		} else if hi == nil {
			log.Printf("turn_reports: %s: %s: location %q: parse => nil!\n", rpf.Id, unit.Id, string(location))
		} else if hexes, ok := hi.([2]*coords.Grid); !ok {
			panic(fmt.Sprintf("assert(type != %T)", hi))
		} else {
			unit.PrevHex, unit.CurrHex = hexes[0], hexes[1]
		}
		//log.Printf("turn_reports: %s: location %q: ==> %q %q\n", rpf.Id, string(location), unit.PrevHex, unit.CurrHex)

		//log.Printf("turn_reports: %s: %s: movements\n", rpf.Id, unit.Id)
		movement := sections.ParseMovementLine(rs.Id, lines)
		if unit.Raw != nil {
			unit.Raw.Movement = string(movement)
		}
		//log.Printf("turn_reports: %s: %s: movements: input <== %q\n", rpf.Id, unit.Id, string(movement))
		//log.Printf("turn_reports: %s: %s: movements: start %q end %q\n", rpf.Id, unit.Id, unit.PrevHex, unit.CurrHex)
		um, err := movements.ParseMovements(fmt.Sprintf("%-6s %s", rpf.Id, unit.Id), movement)
		if err != nil {
			log.Fatalf("turn_reports: %s: %s: movements: parse: error %v\n", rpf.Id, unit.Id, err)
		}
		if um == nil {
			// no movement so nothing to do
			// log.Printf("turn_reports: %s: %s: movements: parse: no movements\n", rpf.Id, unit.Id)
			um = &domain.Movement{}
		} else if um.Follows != "" {
			// capture the unit this unit is following
			//log.Printf("turn_reports: %s: %s: movements: parse: follows %q\n", rpf.Id, unit.Id, m.Follows)
		} else if um.Steps != nil {
			// capture the movement, including all of its steps
			//log.Printf("turn_reports: %s: %s: movements: parse: steps %d\n", rpf.Id, unit.Id, len(m.Moves))
		}
		unit.Movement = um
		unit.Movement.Turn = fmt.Sprintf("%03d-%02d", rpf.Year, rpf.Month)

		for _, line := range sections.ParseScoutLines(rs.Id, lines) {
			unit.ScoutLines = append(unit.ScoutLines, string(line))
		}
		unit.Status = string(sections.ParseStatusLine(rs.Id, lines))

		if captureRawText {
			rs.RawText = string(section)
		}

		rss = append(rss, rs)
	}

	return rss, nil
}

// ParseHeaders extracts the header from the input so that we can verify the settings
func ParseHeaders(id string, lines [][]byte) (kind domain.ReportSectionType, unitId string, year, month int, currHex string, err error) {
	header, err := headers.ParseHeader(id, lines)
	if err != nil {
		return kind, unitId, year, month, currHex, err
	} else if year, err = strconv.Atoi(header.Game.Year); err != nil {
		return kind, unitId, year, month, currHex, err
	} else if month, err = strconv.Atoi(header.Game.Month); err != nil {
		return kind, unitId, year, month, currHex, err
	}
	currHex = header.CurrHex

	unitId, kind = sections.ParseSectionType(lines)

	return kind, unitId, year, month, currHex, err
}

type PSection struct {
	Self  string                   `json:"self"`
	Kind  domain.ReportSectionType `json:"-"`
	Units []*PUnit                 `json:"units,omitempty"`
}

type PUnit struct {
	Id      string   `json:"id"`
	Start   string   `json:"start,omitempty"`
	Follows string   `json:"follows,omitempty"`
	Moves   []*PMove `json:"moves,omitempty"`
	End     string   `json:"end,omitempty"`
}

type PMove struct {
	Seq       int     `json:"seq"`
	Direction string  `json:"direction"`
	Blocked   bool    `json:"blocked,omitempty"`
	Exhausted bool    `json:"exhausted,omitempty"`
	Found     *PFound `json:"found,omitempty"`
}

type PFound struct {
	Terrain    string   `json:"terrain"`
	Edges      []*PEdge `json:"edges,omitempty"`
	Units      []string `json:"units,omitempty"`
	Settlement string   `json:"settlement,omitempty"`
}

type PEdge struct {
	Direction string `json:"direction"`
	Edge      string `json:"edge,omitempty"`
	Terrain   string `json:"terrain,omitempty"`
}

func ParseSection(sectionId string, section []byte, showSlugs, captureRawText bool) (*PSection, error) {
	lines := bytes.Split(section, []byte{'\n'})
	log.Printf("parse: section %s: lines %8d\n", sectionId, len(lines))

	if showSlugs {
		var slug []byte
		if len(section) > 40 {
			slug = section[:40]
		} else {
			slug = section
		}
		log.Printf("parse: section %s: lines %8d: %q\n", sectionId, len(lines), string(slug))
	}

	// funk up some header values
	followsLine := []byte("Tribe Follows ")
	movesLine := []byte("Tribe Movement: ")
	var scoutLines [8][]byte
	for sid := 0; sid < 8; sid++ {
		scoutLines[sid] = []byte(fmt.Sprintf("Scout %d:Scout  ", sid+1))
	}

	// chunk up the input into the various chunks
	var chunks struct {
		Header  [][]byte
		Follows []byte
		Moves   []byte
		Scout   [][]byte
		Status  []byte
	}
	for n, line := range lines {
		if n == 0 || n == 1 {
			chunks.Header = append(chunks.Header, line)
		} else if bytes.HasPrefix(line, followsLine) {
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
		}
	}
	if chunks.Follows == nil && chunks.Moves == nil {
		log.Printf("parse: section %s: warning: missing follows and movement lines\n", sectionId)
	}
	if len(chunks.Scout) > 8 {
		return nil, cerrs.ErrTooManyScoutLines
	}

	kind, unitId, year, month, hex, err := ParseHeaders(sectionId, chunks.Header)
	if err != nil {
		log.Fatalf("parse: sections: error: %v\n", err)
	}
	log.Printf("parse: section %s: %q: unit %q: year %04d: month %02d: hex %q\n", sectionId, kind, unitId, year, month, hex)
	if kind != domain.RSUnit {
		// not a turn report, so return an error
		return nil, cerrs.ErrNotATurnReport
	}

	pu := &PUnit{
		Id:    unitId,
		Start: hex,
	}

	// now that we know the unit id, we can find the status line, too
	statusLine := []byte(fmt.Sprintf("%s Status: ", unitId))
	for _, line := range lines {
		if bytes.HasPrefix(line, statusLine) {
			if chunks.Status != nil {
				return nil, cerrs.ErrMultipleStatusLines
			}
			chunks.Status = line
		}
	}
	if chunks.Status == nil {
		return nil, cerrs.ErrMissingStatusLine
	}
	log.Printf("parse: section %s: %q: unit %q: status %q\n", sectionId, kind, unitId, string(chunks.Status))

	var um *domain.Movement
	if chunks.Follows != nil {
		um, err = movements.ParseMovements(sectionId, chunks.Follows)
	} else if chunks.Moves != nil {
		um, err = movements.ParseMovements(sectionId, chunks.Moves)
	}
	if err != nil {
		log.Fatalf("parse: section %s: unit %s: movements: error %v\n", sectionId, unitId, err)
	} else if um == nil {
		// unit did not move, must use the status line for the end hex
		log.Printf("parse: section %s: unit %s: movements nil: replace with status\n", sectionId, unitId)
	} else {
		log.Printf("parse: section %s: unit %s: movements %v\n", sectionId, unitId, *um)
	}

	ps := &PSection{
		Kind:  kind,
		Units: []*PUnit{pu},
	}

	return ps, nil
}
