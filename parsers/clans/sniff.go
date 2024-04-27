// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/parsers/clans/headers"
	"log"
	"regexp"
)

// sniffHeader extracts the clan id, year, and month from the input.
func sniffHeader(name string, input []byte) (headers.Header, error) {
	if !bytes.HasPrefix(input, []byte("Tribe ")) {
		return headers.Header{}, cerrs.ErrNotATurnReport
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
		return headers.Header{}, cerrs.ErrNotATurnReport
	}
	input = input[:length]

	// parse the header
	hi, err := headers.Parse(name, input)
	if err != nil {
		return headers.Header{}, cerrs.ErrNotATurnReport
	}
	header, ok := hi.(headers.Header)
	if !ok {
		log.Fatalf("clans: %s: internal error: want headers.Header, got %T\n", name, hi)
	}

	return header, nil
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
