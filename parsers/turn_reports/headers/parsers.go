// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package headers

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

var (
	rxScout *regexp.Regexp
)

// ParseHeader returns
func ParseHeader(id string, input []byte) (*Header, error) {
	// the header is the first two lines of the report and must start with "Tribe: "
	input = theFirstTwoLines(input)
	if !bytes.HasPrefix(input, []byte("Tribe ")) {
		return nil, cerrs.ErrNotATurnReport
	}
	if input[len(input)-1] != '\n' {
		panic("hey, bad function")
	}

	// parse the header
	hi, err := Parse(id, input)
	if err != nil {
		log.Printf("clans: header: %v\n", err)
		return nil, cerrs.ErrNotATurnReport
	}
	header, ok := hi.(*Header)
	if !ok {
		log.Fatalf("clans: %s: internal error: want *Header, got %T\n", id, hi)
	}
	return header, nil
}

// ParseSectionType returns the section's identifier and type.
func ParseSectionType(lines [][]byte) (string, domain.ReportSectionType) {
	if len(lines) == 0 {
		return "", domain.RSUnknown
	}

	// use the first few bytes of the line to determine the unit
	line := lines[0]
	if id, ut := ParseUnitType(line); ut != domain.UTUnknown {
		return id, domain.RSUnit
	} else if bytes.HasPrefix(line, []byte("Settlements\n")) {
		return "", domain.RSSettlements
	} else if bytes.HasPrefix(line, []byte("Transfers\n")) {
		return "", domain.RSTransfers
	}
	return "", domain.RSUnknown
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

// ParseUnitType returns the unit's type and identifier.
func ParseUnitType(line []byte) (string, domain.UnitType) {
	// use the first few bytes of the line to determine the unit
	if bytes.HasPrefix(line, []byte("Courier ")) {
		return string(line[8:14]), domain.UTCourier
	} else if bytes.HasPrefix(line, []byte("Element ")) {
		return string(line[8:14]), domain.UTElement
	} else if bytes.HasPrefix(line, []byte("Garrison ")) {
		return string(line[9:15]), domain.UTGarrison
	} else if bytes.HasPrefix(line, []byte("Tribe ")) {
		return string(line[6:10]), domain.UTTribe
	}
	return "", domain.UTUnknown
}

// theFirstTwoLines returns the first two lines of the input as a single slice.
// returns nil if there are not at least two lines in the input.
func theFirstTwoLines(input []byte) []byte {
	for pos, nlCount := 0, 0; nlCount < 2 && pos < len(input); pos++ {
		if input[pos] == '\n' {
			nlCount++
			if nlCount == 2 {
				return input[:pos+1]
			}
		}
	}
	return nil
}
