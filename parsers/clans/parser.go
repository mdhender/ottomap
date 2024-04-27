// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/clans/headers"
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

	// scan the input to find the section separator
	var separator []byte
	for _, pattern := range [][]byte{
		[]byte{0xE2, 0x80, 0x83},                         // MS Word section break
		[]byte{0x0a, 0x2f, 0x2f, 0x2d, 0x2d, 0x2d, 0x2d}, // \n//----
		[]byte{'\f'}, // simple form feed
	} {
		if bytes.Index(input, pattern) == -1 {
			continue
		}
		separator = pattern
		break
	}
	if separator == nil {
		log.Printf("clans: %s: missing separator\n", rpf.Name)
		return nil, cerrs.ErrNoSeparator
	}
	log.Printf("clans: %s: separator %q\n", rpf.Name, separator)

	//
	//turn := &Turn{
	//	Clan: input.Clan,
	//}
	//
	//sections := bytes.Split(data, separator)
	//for n, section := range sections {
	//	section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
	//	section = append(section, '\n')
	//	var slug []byte
	//	if len(section) > 40 {
	//		slug = section[:40]
	//	} else {
	//		slug = section
	//	}
	//	log.Printf("%3d: %6d: %q\n", n+1, len(section), string(slug))
	//	if bytes.HasPrefix(section, []byte("Courier ")) {
	//		id := string(section[8:14])
	//		unit := &Unit{Id: id, Text: sniffMovement(id, section)}
	//		//log.Printf("courier   unit id %q\n", unit.Id)
	//		turn.Units = append(turn.Units, unit)
	//	} else if bytes.HasPrefix(section, []byte("Element ")) {
	//		id := string(section[8:14])
	//		unit := &Unit{Id: id, Text: sniffMovement(id, section)}
	//		//log.Printf("element   unit id %q\n", unit.Id)
	//		turn.Units = append(turn.Units, unit)
	//	} else if bytes.HasPrefix(section, []byte("Garrison ")) {
	//		id := string(section[9:15])
	//		unit := &Unit{Id: id, Text: sniffMovement(id, section)}
	//		//log.Printf("garrison  unit id %q\n", unit.Id)
	//		turn.Units = append(turn.Units, unit)
	//	} else if bytes.HasPrefix(section, []byte("Tribe ")) {
	//		id := string(section[6:10])
	//		unit := &Unit{Id: id, Text: sniffMovement(id, section)}
	//		//log.Printf("tribe     unit id %q\n", unit.Id)
	//		turn.Units = append(turn.Units, unit)
	//	} else if bytes.HasPrefix(section, []byte("Transfers\n")) {
	//		turn.Transfers = string(section)
	//	} else if bytes.HasPrefix(section, []byte("Settlements\n")) {
	//		turn.Settlements = string(section)
	//	} else {
	//		log.Printf("%3d: %6d: error: unknown section\n", n+1, len(section))
	//	}
	//}
	//
	//return turn, nil
	return nil, cerrs.ErrNotImplemented
}

func sniffHeader(name string, input []byte) (headers.Header, error) {
	if !bytes.HasPrefix(input, []byte("Tribe ")) {
		return headers.Header{}, cerrs.ErrNotATurnReport
	}

	// the header will be the first two lines of the input
	nlCount, length := 0, 0
	for pos := 0; nlCount < 2 && pos < len(input); pos++ {
		if input[pos] == '\n' {
			nlCount++
		}
		length++
	}
	if nlCount != 2 {
		return headers.Header{}, cerrs.ErrNotATurnReport
	}
	input = input[:length]

	// parse the header
	hi, err := headers.Parse(name, input)
	if err != nil {
		return headers.Header{}, cerrs.ErrNotATurnReport
	}
	header, ok := hi.(headers.Header)
	if !ok {
		log.Fatalf("clans: %s: internal error: want headers.Header, got %T\n", name, hi)
	}

	return header, nil
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
			unitMovement = line
		} else if reScout.Match(line) {
			scoutMovements = append(scoutMovements, line)
		} else if reStatus.Match(line) {
			unitStatus = line
		}
	}

	var results []byte
	results = append(results, location...)
	results = append(results, '\n')
	if unitMovement != nil {
		results = append(results, unitMovement...)
		results = append(results, '\n')
	}
	for _, scoutMovement := range scoutMovements {
		results = append(results, scoutMovement...)
		results = append(results, '\n')
	}
	if unitStatus != nil {
		results = append(results, unitStatus...)
		results = append(results, '\n')
	}
	return results
}
