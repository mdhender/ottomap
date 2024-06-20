// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sections

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
)

type Section struct {
	No    int
	Lines []*Line
}

type Line struct {
	No   int
	Text []byte
}

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
	var elementId []byte
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
			sct = &Section{No: len(scts) + 1, Lines: []*Line{{No: lineNo, Text: line}, {No: lineNo, Text: nextLine}}}
			elementId = line[8 : 8+6]
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:14], elementId)
		} else if rxElementSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: []*Line{{No: lineNo, Text: line}, {No: lineNo, Text: nextLine}}}
			elementId = line[8 : 8+6]
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:14], elementId)
		} else if rxFleetSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: []*Line{{No: lineNo, Text: line}, {No: lineNo, Text: nextLine}}}
			elementId = line[8 : 8+6]
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:12], elementId)
		} else if rxGarrisonSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: []*Line{{No: lineNo, Text: line}, {No: lineNo, Text: nextLine}}}
			elementId = line[9 : 9+6]
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:15], elementId)
		} else if rxTribeSection.Match(line) && bytes.HasPrefix(nextLine, []byte("Current Turn ")) {
			if sct != nil {
				scts = append(scts, sct)
			}
			sct = &Section{No: len(scts) + 1, Lines: []*Line{{No: lineNo, Text: line}, {No: lineNo, Text: nextLine}}}
			elementId = line[6 : 6+4]
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
			debugf("report %s: section %5d: line %5d: found %q %q\n", id, sct.No, lineNo, line[:10], elementId)
		} else if sct == nil {
			// ignore all lines that are not in a section
			//} else if len(scts) == 1 && bytes.HasPrefix(line, []byte("Current Turn ")) {
			//	sct.Lines = append(sct.Lines, &Line{No: lineNo, Text: line})
			//	//debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:12])
		} else if bytes.HasPrefix(line, []byte("Tribe Follows: ")) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:13])
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:14])
		} else if rxScoutLine.Match(line) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:14])
		} else if bytes.HasPrefix(line, elementStatusPrefix) {
			sct.Lines = append(sct.Lines, &Line{No: lineNo, Text: line})
			debugf("report %s: section %5d: line %5d: found %q\n", id, sct.No, lineNo, line[:len(elementStatusPrefix)])
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
