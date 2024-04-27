// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"github.com/mdhender/ottomap/domain"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// Parse splits the input into individual sections.
func Parse(rpf *domain.ReportFile) (*Turn, error) {
	// extract clan, year, and month from the file name
	rxTurnReportFile, err := regexp.Compile(`^(\d{3})-(\d{2})\.(0\d{3})\.input\.txt$`)
	if err != nil {
		log.Fatal(err)
	}
	var year, month, clan string
	if matches := rxTurnReportFile.FindStringSubmatch(rpf.Name); len(matches) != 4 {
		log.Fatalf("clans: %s: internal error: regex did not match clan/year/month\n", rpf.Name)
	} else {
		year, month, clan = matches[1], matches[2], matches[3]
	}

	// read the entire input file into memory
	input, err := os.ReadFile(filepath.Join(rpf.Path, rpf.Name))
	if err != nil {
		return nil, err
	}

	// extract the header from the input so that we can verify the settings
	header, err := sniffHeader(rpf.Name, input)
	if err != nil {
		return nil, err
	}
	if header.ClanId != clan {
		log.Fatalf("clans: %s: mismatched clan: id %q: want %q\n", rpf.Name, header.ClanId, clan)
	} else if header.Game.Year != year {
		log.Fatalf("clans: %s: mismatched clan: year %q: want %q\n", rpf.Name, header.Game.Year, year)
	} else if header.Game.Month != month {
		log.Fatalf("clans: %s: mismatched clan: month %q, want %q\n", rpf.Name, header.Game.Month, month)
	}

	sections, separator := splitSections(input)
	if separator == nil {
		log.Printf("clans: %s: missing separator\n", rpf.Name)
		return nil, err
	}
	log.Printf("clans: %s: separator %q\n", rpf.Name, separator)

	// process all the sections, adding each to the turn.
	turn := &Turn{
		Clan: clan,
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
