// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clan_turn

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"log"
	"os"
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
			unit := &Unit{Id: string(section[8:14]), Text: section}
			//log.Printf("courier   unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Element ")) {
			unit := &Unit{Id: string(section[8:14]), Text: section}
			//log.Printf("element   unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Garrison ")) {
			unit := &Unit{Id: string(section[9:15]), Text: section}
			//log.Printf("garrison  unit id %q\n", unit.Id)
			turn.Units = append(turn.Units, unit)
		} else if bytes.HasPrefix(section, []byte("Tribe ")) {
			unit := &Unit{Id: string(section[6:10]), Text: section}
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
