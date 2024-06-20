// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sections

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
)

// SplitComment splits the input assuming comment-style section breaks
func SplitComment(input []byte) ([][]byte, bool) {
	// scan the input to find the comment separator
	separator := []byte{0x0a, 0x2f, 0x2f, 0x2d, 0x2d, 0x2d, 0x2d} // \n//----
	if bytes.Index(input, separator) == -1 {
		return nil, false
	}

	// split the input
	sections := bytes.Split(input, separator)

	// our parsers expect sections to not start or end with blank lines.
	// they also require that the last line end with a new-line.
	for i, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
		sections[i] = section
	}

	return sections, true
}

// SplitMSWord splits the input assuming MS Word section breaks.
func SplitMSWord(input []byte) ([][]byte, bool) {
	// scan the input to find the section separator
	separator := []byte{0xE2, 0x80, 0x83} // MS Word section break
	if bytes.Index(input, separator) == -1 {
		return nil, false
	}

	// split the input
	sections := bytes.Split(input, separator)

	// our parsers expect sections to not start or end with blank lines.
	// they also require that the last line end with a new-line.
	for i, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
		sections[i] = section
	}

	return sections, true
}

var (
	rxCourierSection  *regexp.Regexp
	rxElementSection  *regexp.Regexp
	rxFleetSection    *regexp.Regexp
	rxGarrisonSection *regexp.Regexp
	rxScoutLine       *regexp.Regexp
	rxTribeSection    *regexp.Regexp
)

func SplitRegEx(id string, input []byte, showSections bool) ([][]byte, bool) {
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

	type Section struct {
		No    int
		Lines [][]byte
	}

	var scts []*Section
	var sct *Section
	var elementId []byte
	var elementStatusPrefix []byte

	for lineNo, line := range bytes.Split(input, []byte{'\n'}) {
		if rxCourierSection.Match(line) {
			elementId = line[8 : 8+6]
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: [][]byte{line}}
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo+1, line[:14], elementId)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxElementSection.Match(line) {
			elementId = line[8 : 8+6]
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: [][]byte{line}}
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo+1, line[:14], elementId)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxFleetSection.Match(line) {
			elementId = line[6 : 6+6]
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: [][]byte{line}}
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo+1, line[:12], elementId)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxGarrisonSection.Match(line) {
			elementId = line[9 : 9+6]
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: [][]byte{line}}
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo+1, line[:15], elementId)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxTribeSection.Match(line) {
			elementId = line[6 : 6+4]
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: [][]byte{line}}
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo+1, line[:10], elementId)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if sct == nil {
			if len(line) < 35 {
				log.Fatalf("report %s: section %5d: line %5d: found line outside of section: %q\n", id, len(scts)+1, lineNo+1, line)
			} else {
				log.Fatalf("report %s: section %5d: line %5d: found line outside of section: %q\n", id, len(scts)+1, lineNo+1, line[:35])
			}
		} else if len(scts) == 1 && bytes.HasPrefix(line, []byte("Current Turn ")) {
			sct.Lines = append(sct.Lines, line)
			//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo+1, line[:12])
		} else if bytes.HasPrefix(line, []byte("Tribe Follows: ")) {
			sct.Lines = append(sct.Lines, line)
			//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo+1, line[:13])
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			sct.Lines = append(sct.Lines, line)
			//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo+1, line[:14])
		} else if rxScoutLine.Match(line) {
			sct.Lines = append(sct.Lines, line)
			//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo+1, line[:14])
		} else if bytes.HasPrefix(line, elementStatusPrefix) {
			sct.Lines = append(sct.Lines, line)
			//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo+1, line[:len(elementStatusPrefix)])
		} else {
			sct.Lines = append(sct.Lines, line)
		}
	}

	// convert those Section slices into byte slices
	var sections [][]byte
	for _, sct := range scts {
		// our parsers expect sections to not start or end with blank lines.
		// they also require that the last line end with a new-line.
		lines := bytes.Join(sct.Lines, []byte{'\n'})
		lines = bytes.TrimRight(bytes.TrimLeft(lines, "\n"), "\n")
		lines = append(lines, '\n')
		sections = append(sections, lines)
	}

	return sections, true
}

// SplitSimpleFormFeed splits the input assuming simple form feeds.
func SplitSimpleFormFeed(input []byte) ([][]byte, bool) {
	// scan the input to find the section separator
	separator := []byte{'\f'} // simple form feed
	if bytes.Index(input, separator) == -1 {
		return nil, false
	}

	// split the input
	sections := bytes.Split(input, separator)

	// our parsers expect sections to not start or end with blank lines.
	// they also require that the last line end with a new-line.
	for i, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
		sections[i] = section
	}

	return sections, true
}
