// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package sections implements parsers for lines in a report section
package sections

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"regexp"
)

var (
	rxScout *regexp.Regexp
)

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
	pfxMoves, pfxFollows := []byte("Tribe Movement: "), []byte("Tribe Follows ")
	for _, line := range lines {
		if bytes.HasPrefix(line, pfxMoves) || bytes.HasPrefix(line, pfxFollows) {
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
