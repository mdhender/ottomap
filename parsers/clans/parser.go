// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"os"
)

// Parse splits the input into individual sections.
func Parse(rpf *domain.ReportFile) (*Turn, error) {
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

	// process all the sections, adding each to the turn.
	turn := &Turn{
		Clan: fmt.Sprintf("%04d", rpf.Clan),
	}
	for n, section := range sections {
		var slug []byte
		if len(section) > 40 {
			slug = section[:40]
		} else {
			slug = section
		}
		log.Printf("%3d: %6d: %q\n", n+1, len(section), string(slug))
		if bytes.HasPrefix(section, []byte("Courier ")) {
			id := string(section[8:14])
			unit := &Unit{Id: id, Text: sniffMovement(id, section)}
			//log.Printf("courier   unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Element ")) {
			id := string(section[8:14])
			unit := &Unit{Id: id, Text: sniffMovement(id, section)}
			//log.Printf("element   unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Garrison ")) {
			id := string(section[9:15])
			unit := &Unit{Id: id, Text: sniffMovement(id, section)}
			//log.Printf("garrison  unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Tribe ")) {
			id := string(section[6:10])
			unit := &Unit{Id: id, Text: sniffMovement(id, section)}
			//log.Printf("tribe     unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Transfers\n")) {
			turn.Transfers = string(section)
		} else if bytes.HasPrefix(section, []byte("Settlements\n")) {
			turn.Settlements = string(section)
		} else {
			log.Fatalf("%3d: %6d: error: unknown section\n", n+1, len(section))
		}
	}

	return turn, nil
}
