// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turn_reports

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/turn_reports/headers"
	"github.com/mdhender/ottomap/parsers/turn_reports/locations"
	"github.com/mdhender/ottomap/parsers/turn_reports/movements"
	"github.com/mdhender/ottomap/parsers/turn_reports/sections"
	"log"
	"os"
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
	header, err := headers.ParseHeader(rpf.Id, input)
	if err != nil {
		return nil, err
	} else if header.ClanId != fmt.Sprintf("%04d", rpf.Clan) {
		log.Fatalf("clans: %s: mismatched clan: id %q: want %q\n", rpf.Id, header.ClanId, fmt.Sprintf("%04d", rpf.Clan))
	} else if header.Game.Year != fmt.Sprintf("%03d", rpf.Year) {
		log.Fatalf("clans: %s: mismatched clan: year %q: want %q\n", rpf.Id, header.Game.Year, fmt.Sprintf("%03d", rpf.Year))
	} else if header.Game.Month != fmt.Sprintf("%02d", rpf.Month) {
		log.Fatalf("clans: %s: mismatched clan: month %q, want %q\n", rpf.Id, header.Game.Month, fmt.Sprintf("%02d", rpf.Month))
	}

	// debug logic to limit testing to just one interesting turn
	if !(rpf.Year == 900 && rpf.Month == 2) {
		log.Printf("turn_reports: parse: skipping %03d-%02d.%04d\n", rpf.Year, rpf.Month, rpf.Clan)
		return nil, cerrs.ErrNotImplemented
	}

	ss, separator := splitSections(input)
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

		// debug logic to test one tribe that has units with follows and movement in the same turn
		//if !(bytes.HasPrefix(lines[0], []byte("Tribe 2138")) || bytes.HasPrefix(lines[0], []byte("Element 2138e1"))) {
		//	log.Printf("turn_reports: parse: skipping %s\n", string(lines[0]))
		//	continue
		//}

		log.Printf("turn_reports: parse: parsing %s\n", string(lines[0]))
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
		} else if hexes, ok := hi.([2]*domain.GridHex); ok {
			unit.PrevHex = hexes[0]
			unit.CurrHex = hexes[1]
		}
		//log.Printf("turn_reports: %s: location %q: ==> %q %q\n", rpf.Id, string(location), unit.PrevHex, unit.CurrHex)

		movement := sections.ParseMovementLine(rs.Id, lines)
		if unit.Raw != nil {
			unit.Raw.Movement = string(movement)
		}
		log.Printf("turn_reports: %s: %s: movements: input <== %q\n", rpf.Id, unit.Id, string(movement))
		log.Printf("turn_reports: %s: movements: start %q end %q\n", rpf.Id, unit.PrevHex, unit.CurrHex)
		m, err := movements.ParseMovements(fmt.Sprintf("%-6s %s", rpf.Id, unit.Id), movement)
		if err != nil {
			log.Printf("turn_reports: %s: %s: movements: parse: error %v\n", rpf.Id, unit.Id, err)
		} else if m == nil {
			log.Printf("turn_reports: %s: %s: movements: parse: no movements\n", rpf.Id, unit.Id)
		} else if m.Follows != "" {
			unit.Follows = m.Follows
			log.Printf("turn_reports: %s: %s: movements: parse: follows %q\n", rpf.Id, unit.Id, m.Follows)
		} else if m.Moves != nil {
			log.Printf("turn_reports: %s: %s: movements: parse: steps %d\n", rpf.Id, unit.Id, len(m.Moves))
			unit.Movement = &domain.Movement{}
			for _, pm := range m.Moves {
				sr := &domain.StepResults{Step: pm.Step}
				unit.Movement.Steps = append(unit.Movement.Steps, sr)
				for _, ms := range pm.Results {
					sr.Results = append(sr.Results, ms)
				}
			}
		}
		log.Printf("turn_reports: %s: %s: movements: todo: split by commas\n", rpf.Id, unit.Id)

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
