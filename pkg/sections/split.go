// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sections

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

type Section struct {
	No    int
	Id    string // unit id
	Type  domain.UnitType
	Lines []*Line
}

func (s *Section) Slug() string {
	if len(s.Lines) == 0 {
		return fmt.Sprintf("%d: %s: no lines", s.No, s.Type)
	} else if len(s.Lines[0].Text) < 37 {
		return fmt.Sprintf("%d: %s: %q", s.No, s.Type, s.Lines[0].Text)
	}
	return fmt.Sprintf("%d: %s: %q", s.No, s.Type, s.Lines[0].Text[:37])
}

type Line struct {
	No           int
	IsLocation   bool
	IsStatus     bool
	IsTurnInfo   bool
	MovementType domain.UnitMovement_e
	Text         []byte
}

func (l *Line) Slug(n int) string {
	if len(l.Text) < n {
		return string(l.Text)
	}
	return string(l.Text[:n])
}

var (
	rxCourierSection  *regexp.Regexp
	rxElementSection  *regexp.Regexp
	rxFleetMove       *regexp.Regexp
	rxFleetSection    *regexp.Regexp
	rxGarrisonSection *regexp.Regexp
	rxScoutLine       *regexp.Regexp
	rxTribeSection    *regexp.Regexp
)

func SplitRegEx(id string, input []byte, showSections bool) ([]*Section, bool) {
	if len(input) == 0 {
		return nil, false
	}

	debugf := func(format string, args ...any) {
		if showSections {
			log.Printf(format, args...)
		}
	}

	if rxCourierSection == nil {
		var err error
		if rxCourierSection, err = regexp.Compile(`^Courier \d{4}c\d, ,`); err != nil {
			panic(err)
		} else if rxElementSection, err = regexp.Compile(`^Element \d{4}e\d, ,`); err != nil {
			panic(err)
		} else if rxFleetMove, err = regexp.Compile(`^(CALM|MILD|STRONG|GALE)\s+(NE|SE|SW|NW|N|S)\s+Fleet Movement: Move\s+`); err != nil {
			panic(err)
		} else if rxFleetSection, err = regexp.Compile(`^Fleet \d{4}f\d, ,`); err != nil {
			panic(err)
		} else if rxGarrisonSection, err = regexp.Compile(`^Garrison \d{4}g\d, ,`); err != nil {
			panic(err)
		} else if rxScoutLine, err = regexp.Compile(`^Scout \d:Scout `); err != nil {
			panic(err)
		} else if rxTribeSection, err = regexp.Compile(`^Tribe \d{4}, ,`); err != nil {
			panic(err)
		}
	}

	// report must start with a tribe section
	if !rxTribeSection.Match(input) {
		return nil, false
	}

	var scts []*Section
	var sct *Section
	var elementStatusPrefix []byte

	lines := bytes.Split(input, []byte{'\n'})
	for no, line := range lines {
		lineNo := no + 1
		var nextLine []byte
		if no < len(lines)-1 {
			nextLine = lines[no+1]
		}

		if rxCourierSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{
				No:   len(scts) + 1,
				Id:   string(line[8 : 8+6]),
				Type: domain.UTCourier,
				Lines: []*Line{
					{No: lineNo, IsLocation: true, Text: line},
					{No: lineNo, IsTurnInfo: true, Text: nextLine},
				},
			}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", sct.Id))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:14], sct.Id)
		} else if rxElementSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{
				No:   len(scts) + 1,
				Id:   string(line[8 : 8+6]),
				Type: domain.UTElement,
				Lines: []*Line{
					{No: lineNo, IsLocation: true, Text: line},
					{No: lineNo, IsTurnInfo: true, Text: nextLine},
				},
			}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", sct.Id))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:14], sct.Id)
		} else if rxFleetSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{
				No:   len(scts) + 1,
				Id:   string(line[6 : 6+6]),
				Type: domain.UTFleet,
				Lines: []*Line{
					{No: lineNo, IsLocation: true, Text: line},
					{No: lineNo, IsTurnInfo: true, Text: nextLine},
				},
			}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", sct.Id))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:12], sct.Id)
		} else if rxGarrisonSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{
				No:   len(scts) + 1,
				Id:   string(line[9 : 9+6]),
				Type: domain.UTGarrison,
				Lines: []*Line{
					{No: lineNo, IsLocation: true, Text: line},
					{No: lineNo, IsTurnInfo: true, Text: nextLine},
				},
			}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", sct.Id))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:15], sct.Id)
		} else if rxTribeSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{
				No:   len(scts) + 1,
				Id:   string(line[6 : 6+4]),
				Type: domain.UTTribe,
				Lines: []*Line{
					{No: lineNo, IsLocation: true, Text: line},
					{No: lineNo, IsTurnInfo: true, Text: nextLine},
				},
			}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", sct.Id))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:10], sct.Id)
		} else if sct == nil {
			// ignore all lines that are not in a section
		} else if bytes.HasPrefix(line, []byte("Tribe Follows: ")) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, MovementType: domain.UMFollows, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:13])
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, MovementType: domain.UMTribe, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:14])
		} else if rxScoutLine.Match(line) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, MovementType: domain.UMScouts, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:14])
		} else if bytes.HasPrefix(line, elementStatusPrefix) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, IsStatus: true, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:len(elementStatusPrefix)])
		} else if rxFleetMove.Match(line) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, MovementType: domain.UMFleet, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:13])
		} else {
			// ignore all other lines
		}
	}
	if sct != nil {
		scts = append(scts, sct)
	}
	debugf("report %s: found %d sections\n", id, len(scts))

	return scts, true
}
