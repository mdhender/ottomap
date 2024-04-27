// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/clans/parsers"
	"log"
	"os"
)

// Parse splits the input into individual sections.
func Parse(rpf *domain.ReportFile, debugSlugs, captureRawText bool) ([]*domain.ReportSection, error) {
	// read the entire input file into memory
	input, err := os.ReadFile(rpf.Path)
	if err != nil {
		return nil, err
	}

	// extract the header from the input so that we can verify the settings
	header, err := sniffHeader(rpf.Id, input)
	if err != nil {
		return nil, err
	}
	if header.ClanId != fmt.Sprintf("%04d", rpf.Clan) {
		log.Fatalf("clans: %s: mismatched clan: id %q: want %q\n", rpf.Id, header.ClanId, fmt.Sprintf("%04d", rpf.Clan))
	} else if header.Game.Year != fmt.Sprintf("%03d", rpf.Year) {
		log.Fatalf("clans: %s: mismatched clan: year %q: want %q\n", rpf.Id, header.Game.Year, fmt.Sprintf("%03d", rpf.Year))
	} else if header.Game.Month != fmt.Sprintf("%02d", rpf.Month) {
		log.Fatalf("clans: %s: mismatched clan: month %q, want %q\n", rpf.Id, header.Game.Month, fmt.Sprintf("%02d", rpf.Month))
	}

	sections, separator := splitSections(input)
	if separator == nil {
		log.Printf("clans: %s: missing separator\n", rpf.Id)
		return nil, err
	}
	// log.Printf("clans: %s: sections %3d: separator %q\n", rpf.Id, len(sections), separator)

	// capture only the unit sections
	var rss []*domain.ReportSection
	for n, section := range sections {
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
		rs.Type, rs.Id = parsers.ParseSectionType(lines)
		if rs.Type != domain.RSUnit {
			continue
		}

		rs.Location = string(parsers.ParseLocationLine(rs.Id, lines))
		rs.Movement = string(parsers.ParseMovementLine(rs.Id, lines))
		for _, line := range parsers.ParseScoutLines(rs.Id, lines) {
			rs.ScoutLines = append(rs.ScoutLines, string(line))
		}
		rs.Status = string(parsers.ParseStatusLine(rs.Id, lines))
		if captureRawText {
			rs.RawText = string(section)
		}

		rss = append(rss, rs)
	}

	return rss, nil
}
