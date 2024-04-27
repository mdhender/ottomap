// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
)

// splitSections splits the input. It returns the sections along
// with the section separator. We trim leading and trailing new-lines
// from each section and then force the section to end with a new-line.
//
// We check for a few types of separators and use the
// first one that we find. If we can't find a separator,
// we return the entire input as the first value and nil
// for the separator.
//
// NB: The first turn report (the "setup" turn) might have just
// one section, so we wouldn't find a section separator. The
// instructions should tell the user to manually add one. Or the
// caller should have logic to handle.
func splitSections(input []byte) ([][]byte, []byte) {
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

	// split the input
	var sections [][]byte
	if separator == nil {
		sections = [][]byte{input}
	} else {
		bytes.Split(input, separator)
	}

	// our parsers expect sections to not start or end with blank lines.
	// they also require that the last line end with a new-line.
	for i, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
		sections[i] = section
	}

	return sections, separator
}
