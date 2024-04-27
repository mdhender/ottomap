// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package headers

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"log"
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
