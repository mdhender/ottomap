// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package headers

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"log"
)

// ParseHeader returns
func ParseHeader(id string, lines [][]byte) (*Header, error) {
	// the header is the first two lines of the report and must start with "Tribe: "
	if len(lines) < 2 {
		return nil, cerrs.ErrNotATurnReport
	}
	if !bytes.HasPrefix(lines[0], []byte("Tribe ")) {
		return nil, cerrs.ErrNotATurnReport
	}

	// parse the header
	hi, err := Parse(id, append(bytes.Join(lines, []byte{'\n'}), '\n'))
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
