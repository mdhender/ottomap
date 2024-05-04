// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
)

var (
	rxRiverEdge   *regexp.Regexp
	rxTerrainCost *regexp.Regexp
	rxWaterEdge   *regexp.Regexp
)

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
func ParseMovements(id string, input []byte) (*domain.Movement, error) {
	// todo: update this to use Domain structs

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
			return &domain.Movement{Follows: string(fields[2])}, nil
		}
		log.Printf("parsers: turn_reports: %q: movements: error parsing Tribe Follows\n", id)
		log.Printf("The input is: %s\n", string(input))
		return nil, fmt.Errorf("invalid follows input")
	}

	// if this is not a movements line, return an error
	if !bytes.HasPrefix(input, []byte("Tribe Movement: ")) {
		log.Printf("parsers: turn_reports: %q: movements: internal error parsing Tribe Movement\n", id)
		log.Printf("The input is: %s\n", string(input))
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
		return &domain.Movement{}, nil
	}

	// otherwise, parse each step, returning immediately on error
	dm := &domain.Movement{}
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
			st := &domain.Step{
				Direction: ss.Direction,
				Status:    domain.MSSucceeded,
				Found: domain.Found{
					Terrain:    ss.Found.Terrain,
					Settlement: ss.Found.Settlement,
				},
			}

			// add river and ford edges
			if len(ss.Found.Edges.Ford) != 0 || len(ss.Found.Edges.River) != 0 {
				st.Found.Edges = map[directions.Direction]domain.Edge{}
			}
			for _, dir := range ss.Found.Edges.Ford {
				st.Found.Edges[dir] = domain.EFord
			}
			for _, dir := range ss.Found.Edges.River {
				st.Found.Edges[dir] = domain.ERiver
			}

			// add lake and ocean neighbors
			if len(ss.Found.Edges.Lake) != 0 || len(ss.Found.Edges.Ocean) != 0 {
				st.Found.Seen = map[directions.Direction]domain.Terrain{}
			}
			for _, dir := range ss.Found.Edges.Lake {
				st.Found.Seen[dir] = domain.TLake
			}
			for _, dir := range ss.Found.Edges.Ocean {
				st.Found.Seen[dir] = domain.TOcean
			}

			dm.Steps = append(dm.Steps, st)
		} else if sb, ok := v.(*stepBlocked); ok {
			//log.Printf("ms sb   %+v\n", *sb)
			dm.Steps = append(dm.Steps, &domain.Step{
				Direction: sb.Direction,
				Status:    domain.MSBlocked,
				Found:     domain.Found{Terrain: sb.BlockedBy},
			})
		} else if semp, ok := v.(*stepExhaustedMP); ok {
			//log.Printf("ms semp %+v\n", *semp)
			dm.Steps = append(dm.Steps, &domain.Step{
				Direction: semp.Direction,
				Status:    domain.MSExhausted,
				Found:     domain.Found{Terrain: semp.Terrain},
			})
		} else {
			panic(fmt.Sprintf("assert(type != %T)", v))
		}
	}

	return dm, nil
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
