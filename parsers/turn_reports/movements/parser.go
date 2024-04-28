// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
)

var (
	DebugBuffer   = &bytes.Buffer{}
	rxRiverEdge   *regexp.Regexp
	rxTerrainCost *regexp.Regexp
	rxWaterEdge   *regexp.Regexp
)

type ParsedMovement struct {
	Follows string
	Moves   []*ParsedMove
	Results string
}
type ParsedMove struct {
	Step    string
	Results []string
}

// ParseMovements parses the unit's movements.
// Accepts either "Tribe Follows ..." or "Tribe Movements: ..."
func ParseMovements(id string, input []byte) (*ParsedMovement, error) {
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
			return &ParsedMovement{Follows: string(fields[2])}, nil
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
	log.Printf("movements: todo: split by commas\n")

	// suss out the nightmare of DIRECTION COMMA ONE-OR-MORE-SPACES DIRECTION
	pm := &ParsedMovement{Results: string(rawResults)}
	for n, step := range steps {
		DebugBuffer.WriteString(fmt.Sprintf("  step %2d `%s`\n", n+1, string(step)))
		for x, ch := range step {
			if ch == ',' && validDirFollows(step[x:]) {
				step[x] = ' '
			}
		}
		//// spaces are important (maybe?) so don't trim them
		//for nn, boo := range bytes.Split(step, []byte{','}) {
		//	DebugBuffer.WriteString(fmt.Sprintf("       %2d %2d `%s`\n", n+1, nn+1, string(boo)))
		//}
		// just to see what it does to the parser, trim those spaces
		var move *ParsedMove
		for nn, boo := range bytes.Split(step, []byte{','}) {
			DebugBuffer.WriteString(fmt.Sprintf("       %2d %2d `%s`\n", n+1, nn+1, string(boo)))
			if nn == 0 {
				move = &ParsedMove{Step: string(boo)}
				pm.Moves = append(pm.Moves, move)
				continue
			}
			result := string(bytes.TrimSpace(boo))
			if result != "" {
				move.Results = append(move.Results, result)
			}
		}
	}

	return pm, nil
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

var (
	rxDirFollows *regexp.Regexp
)

// validDirFollows returns true if the input starts with a comma followed by
// one or two spaces followed by a direction followed by a terminator
// (either a comma or end of input).
func validDirFollows(input []byte) bool {
	if rxDirFollows == nil {
		rxDirFollows = regexp.MustCompile(`^,[ ]{1,2}(NW|NE|SW|SE|N|S)(,|$)`)
	}
	return rxDirFollows.Match(input)
}
