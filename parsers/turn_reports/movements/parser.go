// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

var (
	DebugBuffer   = &bytes.Buffer{}
	rxRiverEdge   *regexp.Regexp
	rxTerrainCost *regexp.Regexp
	rxWaterEdge   *regexp.Regexp
)

// ParseMovements parses the unit's movements.
// Accepts either "Tribe Follows ..." or "Tribe Movements: ..."
func ParseMovements(id string, input []byte) (*domain.Movement, error) {
	if rxRiverEdge == nil || rxTerrainCost == nil || rxWaterEdge == nil {
		// No Ford on River to NW of HEX
		if rx, err := regexp.Compile(`^No Ford on River to ([A-Z]+) of HEX$`); err != nil {
			return nil, fmt.Errorf("regex: river: %w", err)
		} else {
			rxRiverEdge = rx
		}
		// Not enough M.P's to move to NE into ROCKY HILLS
		if rx, err := regexp.Compile(`^Not enough M\.P's to move to ([A-Z]+) into (.+)$`); err != nil {
			return nil, fmt.Errorf("regex: terrain: %w", err)
		} else {
			rxTerrainCost = rx
		}
		// Can't Move on Lake to S of HEX
		// Can't Move on Ocean to NW of HEX
		if rx, err := regexp.Compile(`^Can't Move on ([LO][a-z]+) to ([A-Z]+) of HEX$`); err != nil {
			return nil, fmt.Errorf("regex: water: %w", err)
		} else {
			rxWaterEdge = rx
		}
	}

	// aggressively ignore leading and trailing spaces
	input = bytes.TrimSpace(input)

	// do nothing if the unit has no movement to report on
	if len(input) == 0 {
		return nil, nil
	}

	// if this is a follows movement, just return the unit we're following
	if bytes.HasPrefix(input, []byte("Tribe Follows ")) {
		// expect "Tribe Follows UNIT"
		if fields := bytes.Split(input, []byte{' '}); len(fields) == 3 {
			return &domain.Movement{
				Follows: string(fields[2]),
			}, nil
		}
		return nil, fmt.Errorf("invalid follows input")
	}

	// we're expecting the input to look like "Tribe Movement: STEPS? BACKSLASH RESULTS?".
	// return an error if that's not the case.
	if !bytes.HasPrefix(input, []byte("Tribe Movement: ")) {
		return nil, fmt.Errorf("invalid movement input")
	}
	// skip the prefix when splitting the input into steps and results
	rawSteps, rawResults, ok := stepsSlashResults(input[21:])
	if !ok {
		return nil, fmt.Errorf("invalid movement input")
	}
	DebugBuffer.WriteString(fmt.Sprintf("%-16s --------------------------------------------\n", id))
	DebugBuffer.WriteString(fmt.Sprintf("input_ `%s`\n", string(input)))
	DebugBuffer.WriteString(fmt.Sprintf(" steps `%s`\n", string(rawSteps)))
	DebugBuffer.WriteString(fmt.Sprintf(" rslts `%s`\n", string(rawResults)))
	DebugBuffer.WriteString(fmt.Sprintf("    ok  %v\n", ok))
	log.Printf("steps %q: results %q\n", string(rawSteps), string(rawResults))

	// steps should look like STEP (BACKSLASH STEP)*
	var steps [][]byte
	for _, step := range bytes.Split(rawSteps, []byte{'\\'}) {
		// again, aggressively trim spaces from the input
		steps = append(steps, bytes.TrimSpace(step))
	}
	for n, step := range steps {
		DebugBuffer.WriteString(fmt.Sprintf("  step %2d `%s`\n", n+1, string(step)))
	}

	return nil, nil
}

// the input looks something like STUFF BACKSLASH STATUS.
// we'd like to use bytes.Cut, but STUFF can contain back-slashes.
// so we need to find the last back-slash and split there.
func stepsSlashResults(input []byte) ([]byte, []byte, bool) {
	if pos := bytes.LastIndexByte(input, '\\'); pos >= 0 {
		return input[:pos], input[pos+1:], true
	}
	return input, nil, false
}
