// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

var (
	rxRiverEdge   *regexp.Regexp
	rxTerrainCost *regexp.Regexp
	rxWaterEdge   *regexp.Regexp
)

// Movement is a single line of a unit's movement or follows report.
type Movement struct {
	StartingHex domain.GridHex // the hex the unit is starting from
	Follows     string         // the unit this unit is following
	Steps       []*Step        // the steps in the movement
	EndingHex   domain.GridHex // the hex the unit is ending at
}

// Step is a single step in a unit's movement or follows report.
//
// NB: the ending hex will be the same as the starting hex only
// if the movement is blocked by terrain or stopped by MP exhaustion.
type Step struct {
	StartingHex domain.GridHex   // the hex the unit is starting from
	Direction   domain.Direction // direction the unit is moving in
	Blocked     bool             // true if the step is blocked by terrain
	Exhausted   bool             // true if the step is stopped by MP exhaustion
	EndingHex   domain.GridHex   // the hex the unit is ending at
	Found       *Found           // things found in the ending hex
}

// Found is the set of things found in a hex
type Found struct {
	// terrain in the hex
	Terrain domain.Terrain
	// edges that allow or prevent movement
	Edges map[domain.Direction]domain.Edge
	// neighboring terrain that can be seen. usually water, but sometimes lava.
	NeighboringTerrain map[domain.Direction]domain.Terrain
	// settlement in the hex
	Settlement string
}

type ParsedMovement struct {
	Follows string
	Moves   []*ParsedMove
	Results string
}
type ParsedMove struct {
	Direction domain.Direction
	Terrain   domain.Terrain
	Blocked   bool // set only when step is blocked
	Results   []string
}

// Blocked is created when a step is blocked by terrain in the destination hex.
type Blocked struct {
	By domain.Terrain
}

// ParseMovements parses the unit's movements or follows lines.
// Accepts either "Tribe Follows ..." or  lines.
//
// Returns nil, nil if the input is empty. It's up to the caller to
// decide if this is an error.
//
// Returns an error on any unexpected input. Previous versions of this
// function attempted to clean up the input, but that caused issues down
// the line.
//
// For "Tribe Follows ..." lines, the unit being followed is returned
// in the ParseMovement struct.
//
// For "Tribe Movements: ..." lines, the input will be split into individual
// "steps," with a backslash used to separate the steps.
//
// An empty movements line looks like "Tribe Movements: \" and is returned
// as the zero value for ParsedMovement.
//
// Updated to loudly fail on unexpected input. It's annoying, but likely
// better than returning invalid data. The user is expected to fix the
// input and restart.
func ParseMovements(id string, input []byte) (*ParsedMovement, error) {
	log.Printf("parsers: turn_reports: movements: todo: update this to use Movement structs\n")
	log.Printf("parsers: turn_reports: movements: todo: update this to use Step     structs\n")
	log.Printf("parsers: turn_reports: movements: todo: update this to use Found    structs\n")
	// AndExpr = "A" &"B" // matches "A" if followed by a "B" (does not consume "B")
	// NotExpr = "A" !"B" // matches "A" if not followed by a "B" (does not consume "B")

	// initialize the regex machines. should probably be moved to an init function.
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
	line := bytes.TrimSpace(input)

	// do nothing if the unit has no movement to process
	if len(line) == 0 {
		return nil, nil
	}

	// if this is a follows line, return the unit we're following
	if bytes.HasPrefix(line, []byte("Tribe Follows ")) {
		// expect "Tribe Follows UNIT"
		if fields := bytes.Split(line, []byte{' '}); len(fields) == 3 {
			return &ParsedMovement{Follows: string(fields[2])}, nil
		}
		log.Printf("parsers: turn_reports: %q: movements: error parsing Tribe Follows\n", id)
		log.Printf("The input was: %s\n", string(input))
		return nil, fmt.Errorf("invalid follows input")
	}

	// if this is not a movements line, return an error
	if !bytes.HasPrefix(input, []byte("Tribe Movement: ")) {
		log.Printf("parsers: turn_reports: %q: movements: internal error parsing Tribe Movement\n", id)
		log.Printf("The input was: %s\n", string(input))
		log.Printf("Please report this issue on the Discord server.\n")
		return nil, fmt.Errorf("internal error: unexpected input")
	}

	// start a new section in the debug log
	debugf(fmt.Sprintf("%-16s --------------------------------------------\n", id))
	debugf(fmt.Sprintf("input_ `%s`\n", string(input)))

	// trim the prefix and split on backslashes
	input = bytes.TrimPrefix(input, []byte("Tribe Movement: "))

	// split the movement into steps. the report uses backslashes to separate them.
	inputMoves := bytes.Split(input, []byte{'\\'})
	for n, inputMove := range inputMoves {
		debugf(fmt.Sprintf(" moves %2d 《%s》\n", n+1, string(inputMove)))
	}

	// if there are no steps, return an empty ParsedMovement
	if len(inputMoves) == 0 {
		return &ParsedMovement{}, nil
	}

	// otherwise, parse each step, returning immediately on error
	pm := &ParsedMovement{}
	for n, inputStep := range inputMoves {
		step := bytes.TrimSpace(inputStep)
		if n == 0 {
			if bytes.Equal(step, []byte{'M', 'o', 'v', 'e'}) {
				// "Move /" is a special case for the first step of a movement.
				continue
			} else if !bytes.HasPrefix(step, []byte{'M', 'o', 'v', 'e', ' '}) {
				log.Printf("parsers: turn_reports: %q: movements: error parsing Tribe Movement\n", id)
				log.Printf("The move  is: %s\n", string(input))
				log.Printf("The index is: %d\n", n+1)
				log.Printf("The step  is: %s\n", string(inputStep))
				log.Printf("The error is: expected step to start with \"Move \".\n")
				return nil, fmt.Errorf("missing move prefix")
			}
			step = step[5:]
		}
		if len(step) == 0 {
			// this is an error on the first step only
			if n == 0 {
				log.Printf("parsers: turn_reports: %q: movements: error parsing Tribe Movement\n", id)
				log.Printf("The move  is: %s\n", string(input))
				log.Printf("The index is: %d\n", n+1)
				log.Printf("The step  is: %s\n", string(inputStep))
				log.Printf("The error is: expected move results after \"Move\" was found\n")
				return nil, fmt.Errorf("missing move results")
			}
			// ignore empty steps after the first
			continue
		}

		v, err := Parse(id, step)
		if err != nil {
			log.Printf("parsers: turn_reports: %q: movements: error parsing Tribe Movement\n", id)
			log.Printf("The move  is: %s\n", string(input))
			log.Printf("The index is: %d\n", n+1)
			log.Printf("The step  is: %s\n", string(inputStep))
			log.Printf("The error is: %v\n", err)
			log.Printf("movements: error parsing step %d: %v\n", n+1, err)
			return nil, errors.Join(fmt.Errorf("tribe movement"), err)
		}
		if ss, ok := v.(*stepSucceeded); ok {
			//log.Printf("ms ss   %+v\n", *ss)
			pm.Moves = append(pm.Moves, &ParsedMove{
				Direction: ss.Direction,
				Terrain:   ss.Terrain,
			})
		} else if sb, ok := v.(*stepBlocked); ok {
			//log.Printf("ms sb   %+v\n", *sb)
			pm.Moves = append(pm.Moves, &ParsedMove{
				Direction: sb.Direction,
				Terrain:   sb.BlockedBy,
				Blocked:   true,
			})
		} else if semp, ok := v.(*stepExhaustedMP); ok {
			//log.Printf("ms semp %+v\n", *semp)
			pm.Moves = append(pm.Moves, &ParsedMove{
				Direction: semp.Direction,
				Terrain:   semp.Terrain,
				Blocked:   true,
			})
		} else {
			panic(fmt.Sprintf("assert(type != %T)", v))
		}
	}

	//// the report uses backslashes to separate each step in the set of steps,
	//// so we'll use that character to split them up.
	//var steps [][]byte
	//for n, step := range bytes.Split(inputSteps, []byte{'\\'}) {
	//	// trim spaces from the start of the step
	//	// (this seems to only happen when there's a typo in the input)
	//	step = bytes.TrimSpace(step)
	//	// trim spaces and commas from the end of the step
	//	step = bytes.TrimRight(step, ", ")
	//	debugf("  step %2d 《%s》\n", n+1, string(step))
	//	steps = append(steps, step)
	//	// just to see what the parser will see, split up the step
	//	for nn, ss := range bytes.Split(step, []byte{','}) {
	//		debugf("       %2d %2d 《%s》\n", n+1, nn+1, string(ss))
	//	}
	//}
	//log.Printf("movements: call the parser to parse every step\n")
	//log.Printf("movements: todo: split by commas\n")
	//
	//// suss out the nightmare of DIRECTION COMMA ONE-OR-MORE-SPACES DIRECTION
	//pm := &ParsedMovement{Results: string(inputResults)}
	//for n, step := range steps {
	//	debugf("  step %2d `%s`\n", n+1, string(step))
	//	for x, ch := range step {
	//		if ch == ',' && validDirFollows(step[x:]) {
	//			step[x] = ' '
	//		}
	//	}
	//	//// spaces are important (maybe?) so don't trim them
	//	//for nn, boo := range bytes.Split(step, []byte{','}) {
	//	//	debugf("       %2d %2d `%s`\n", n+1, nn+1, string(boo))
	//	//}
	//	// just to see what it does to the parser, trim those spaces
	//	var move *ParsedMove
	//	for nn, boo := range bytes.Split(step, []byte{','}) {
	//		debugf("       %2d %2d `%s`\n", n+1, nn+1, string(boo))
	//		if nn == 0 {
	//			move = &ParsedMove{Step: string(boo)}
	//			pm.Moves = append(pm.Moves, move)
	//			continue
	//		}
	//		result := string(bytes.TrimSpace(boo))
	//		if result != "" {
	//			move.Results = append(move.Results, result)
	//		}
	//	}
	//}
	//
	//// these are the results for the entire movement.
	//// that's not the same as the results for a single step.
	//debugf(" rslts 《%s》\n", string(inputResults))

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

var (
	debug struct {
		buf *bytes.Buffer
	}
)

func EnableDebugBuffer() {
	debug.buf = &bytes.Buffer{}
}

func GetDebugBuffer() []byte {
	if debug.buf == nil {
		return nil
	}
	buf := append([]byte{}, debug.buf.Bytes()...)
	debug.buf = &bytes.Buffer{}
	return buf
}

func debugf(format string, args ...any) {
	if debug.buf == nil {
		return
	}
	debug.buf.WriteString(fmt.Sprintf(format, args...))
}
