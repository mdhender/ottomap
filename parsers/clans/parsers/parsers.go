// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parsers

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

var (
	rxScout *regexp.Regexp
)

// ParseSectionType returns the section's type.
func ParseSectionType(lines [][]byte) (domain.ReportSectionType, string) {
	if len(lines) == 0 {
		return domain.RSUnknown, ""
	}

	// use the first few bytes of the line to determine the unit
	line := lines[0]
	if bytes.HasPrefix(line, []byte("Courier ")) {
		return domain.RSUnit, string(lines[0][8:14])
	} else if bytes.HasPrefix(line, []byte("Element ")) {
		return domain.RSUnit, string(line[8:14])
	} else if bytes.HasPrefix(line, []byte("Garrison ")) {
		return domain.RSUnit, string(line[9:15])
	} else if bytes.HasPrefix(line, []byte("Tribe ")) {
		return domain.RSUnit, string(line[6:10])
	} else if bytes.HasPrefix(line, []byte("Settlements\n")) {
		return domain.RSSettlements, ""
	} else if bytes.HasPrefix(line, []byte("Transfers\n")) {
		return domain.RSTransfers, ""
	}
	return domain.RSUnknown, ""
}

// ParseLocationLine returns the unit's location line.
func ParseLocationLine(id string, lines [][]byte) []byte {
	if len(lines) != 0 {
		return lines[0]
	}
	return nil
}

// ParseMovementLine returns the unit's movement line.
func ParseMovementLine(id string, lines [][]byte) []byte {
	pfx := []byte("Tribe Movement:")
	for _, line := range lines {
		if bytes.HasPrefix(line, pfx) {
			return line
		}
	}
	return nil
}

// ParseScoutLines return's the unit's scout lines.
func ParseScoutLines(id string, lines [][]byte) [][]byte {
	if rxScout == nil {
		var err error
		rxScout, err = regexp.Compile(`^Scout \d:Scout`)
		if err != nil {
			panic(err)
		}
	}

	var scoutLines [][]byte
	for _, line := range lines {
		if rxScout.Match(line) {
			scoutLines = append(scoutLines, line)
		}
	}

	return scoutLines
}

// ParseStatusLine returns the unit's status line.
func ParseStatusLine(id string, lines [][]byte) []byte {
	pfx := []byte(fmt.Sprintf("%s Status: ", id))
	for _, line := range lines {
		if bytes.HasPrefix(line, pfx) {
			return line
		}
	}
	return nil
}

// sniffMovement extracts only the movement lines from the input.
// these include tribe movement and scout results.
//
// that is a lie. it looks like we also grab the unit's location
// and final status.
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
