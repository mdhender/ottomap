// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clan_turn

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"log"
	"os"
	"regexp"
)

// Parse splits the input into individual sections.
func Parse(input InputFile) (*Turn, error) {
	data, err := os.ReadFile(input.File)
	if err != nil {
		return nil, err
	}

	// try to find the separator
	var separator []byte
	for _, pattern := range [][]byte{
		[]byte{0xE2, 0x80, 0x83},                         // MS Word section break
		[]byte{0x0a, 0x2f, 0x2f, 0x2d, 0x2d, 0x2d, 0x2d}, // \n//----
		[]byte{'\f'}, // simple form feed
	} {
		if bytes.Index(data, pattern) == -1 {
			continue
		}
		separator = pattern
		break
	}
	if separator == nil {
		log.Printf("clan_turn: %s: missing separator\n", input.File)
		return nil, cerrs.ErrNoSeparator
	}
	log.Printf("clan_turn: %s: separator %q\n", input.File, separator)

	turn := &Turn{
		Clan: input.Clan,
	}

	sections := bytes.Split(data, separator)
	for n, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
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
			log.Printf("%3d: %6d: error: unknown section\n", n+1, len(section))
		}
	}

	return turn, nil
}

func sniffMovement(id string, input []byte) []byte {
	var location []byte
	var unitMovement []byte
	var unitStatus []byte
	var scoutMovements [][]byte

	reScout, err := regexp.Compile(`^Scout \d{1}:Scout`)
	if err != nil {
		log.Fatal(err)
	}
	reStatus, err := regexp.Compile("^" + id + " Status: ")
	if err != nil {
		log.Fatal(err)
	}

	for n, line := range bytes.Split(input, []byte{'\n'}) {
		if n == 0 {
			location = line
		} else if bytes.HasPrefix(line, []byte("Tribe Movement:")) {
			// unit movement should skip the word "Tribe" and the space
			unitMovement = line[6:]
		} else if reScout.Match(line) {
			// the scout needs to skip the scout id and colon
			scoutMovements = append(scoutMovements, line[8:])
			//scoutMovements = append(scoutMovements, line[8:])
		} else if reStatus.Match(line) {
			// the status needs to skip the unit id and the space
			unitStatus = line[len(id)+1:]
		}
	}

	var results []byte
	results = append(results, location...)
	results = append(results, '\n')
	if unitMovement == nil {
		results = append(results, []byte("Tribe Movement: Still\\")...)
	} else {
		results = append(results, unitMovement...)
	}
	results = append(results, '\n')
	for _, scoutMovement := range scoutMovements {
		results = append(results, scoutMovement...)
		results = append(results, '\n')
	}
	if unitStatus == nil {
		results = append(results, []byte("Status: UNKNOWN, "+id)...)
	} else {
		results = append(results, unitStatus...)
	}
	results = append(results, '\n')
	return results
}
