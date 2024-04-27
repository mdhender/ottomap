// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turn_reports

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/turn_reports/headers"
	"github.com/mdhender/ottomap/parsers/turn_reports/locations"
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
			log.Printf("turn_reports: %s: location %q: parse %v\n", rpf.Id, string(location), err)
		} else if hi == nil {
			log.Printf("turn_reports: %s: location %q: parse => nil!\n", rpf.Id, string(location))
		} else if hexes, ok := hi.([2]*domain.GridHex); ok {
			unit.PrevHex = hexes[0]
			unit.CurrHex = hexes[1]
		}
		log.Printf("turn_reports: %s: location %q: ==> %q %q\n", rpf.Id, string(location), unit.PrevHex, unit.CurrHex)

		movement := sections.ParseMovementLine(rs.Id, lines)
		if unit.Raw != nil {
			unit.Raw.Movement = string(movement)
		}
		unit.Movement = string(movement)

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
