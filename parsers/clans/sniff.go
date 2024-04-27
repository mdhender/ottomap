// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/parsers/clans/parsers"
	"log"
)

// sniffHeader extracts the clan id, year, and month from the input.
func sniffHeader(name string, input []byte) (parsers.Header, error) {
	if !bytes.HasPrefix(input, []byte("Tribe ")) {
		return parsers.Header{}, cerrs.ErrNotATurnReport
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
		return parsers.Header{}, cerrs.ErrNotATurnReport
	}
	input = input[:length]

	// parse the header
	hi, err := parsers.Parse(name, input)
	if err != nil {
		return parsers.Header{}, cerrs.ErrNotATurnReport
	}
	header, ok := hi.(parsers.Header)
	if !ok {
		log.Fatalf("clans: %s: internal error: want headers.Header, got %T\n", name, hi)
	}

	return header, nil
}
