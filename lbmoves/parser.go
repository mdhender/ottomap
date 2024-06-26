// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lbmoves

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
	"unicode"
	"unicode/utf8"
)

//go:generate pigeon -o grammar.go grammar.peg

var (
	rxScoutLine  *regexp.Regexp
	rxStatusLine *regexp.Regexp
)

// ParseMoveResults parses the results of a Land Based Movement.
//
// The line should be the text as extracted directly from the turn report.
// Handles Tribe Follows, Tribe Movement, and Scout lines.
//
// Returns the steps and the first error encountered.
func ParseMoveResults(turnId, unitId string, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Step, error) {
	if rxScoutLine == nil {
		rxScoutLine = regexp.MustCompile(`^Scout [12345678]:Scout `)
		rxStatusLine = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])? Status: `)
	}
	if bytes.HasPrefix(line, []byte("Tribe Follows")) {
		return parseTribeFollows(turnId, unitId, line)
	} else if bytes.HasPrefix(line, []byte("Tribe Movement: Move ")) {
		return parseSteps(turnId, unitId, lineNo, line, bytes.TrimPrefix(line, []byte("Tribe Movement: Move ")), debugSteps, debugNodes)
	}
	return nil, cerrs.ErrNotMovementResults
}

// ParseScoutLine parses the results of a Land Based Movement scout line.
//
// The line should be the text as extracted directly from the turn report.
//
// Returns the steps and the first error encountered.
func ParseScoutLine(turnId, unitId string, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Step, error) {
	if rxScoutLine == nil {
		rxScoutLine = regexp.MustCompile(`^Scout [12345678]:Scout `)
		rxStatusLine = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])? Status: `)
	}
	if rxScoutLine.Match(line) {
		return parseSteps(turnId, unitId, lineNo, line, line[len("Scout ?:Scout "):], debugSteps, debugNodes)
	}
	return nil, cerrs.ErrNotMovementResults
}

// ParseStatusLine parses the status line of a Land Based Movement.
//
// The line should be the text as extracted directly from the turn report.
//
// Returns the steps and the first error encountered.
func ParseStatusLine(turnId, unitId string, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Step, error) {
	if rxStatusLine == nil {
		rxStatusLine = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])? Status: `)
	}
	if rxStatusLine.Match(line) {
		_, b, _ := bytes.Cut(line, []byte{':'})
		b = bytes.TrimSpace(b)
		return parseSteps(turnId, unitId, lineNo, line, b, debugSteps, debugNodes)
	}
	return nil, cerrs.ErrNotMovementResults
}

func parseTribeFollows(turnId, unitId string, line []byte) ([]*Step, error) {
	fields := bytes.Split(bytes.TrimSpace(line), []byte{' '})
	if len(fields) != 3 {
		return nil, cerrs.ErrMissingFollowsUnit
	}
	return []*Step{{
		TurnId:  turnId,
		UnitId:  unitId,
		Result:  Follows,
		Follows: string(fields[2]),
	}}, nil
}

// parseSteps parses all the steps from the results of a Land Based Movement.
func parseSteps(turnId, unitId string, lineNo int, line, steps []byte, debugSteps, debugNodes bool) (results []*Step, err error) {
	// split the steps into single steps, which are backslash-separated, and
	// parse each step individually after trimming spaces and trailing commas.
	for _, step := range bytes.Split(steps, []byte{'\\'}) {
		if result, err := parseStep(turnId, unitId, lineNo, step, debugSteps, debugNodes); err != nil {
			log.Printf("parser: step: %q\n", step)
			log.Printf("parser: line: %q\n", line)
			return nil, err
		} else if result != nil {
			results = append(results, result)
		}
	}
	return results, nil
}

// parseStep parses a single step from the results of a Land Based Movement.
func parseStep(turnId, unitId string, lineNo int, step []byte, debugSteps, debugNodes bool) (result *Step, err error) {
	if debugSteps {
		log.Printf("parser: step: %q\n", step)
	}
	step = bytes.TrimSpace(step)
	//log.Printf("parser: step: %q\n", step)

	root := hexReportToNodes(step, debugNodes)
	steps, err := nodesToSteps(root)
	if err != nil {
		log.Printf("parser: step: %q\n", step)
		return nil, err
	}

	// parse each sub-step separately.
	for n, subStep := range steps {
		if debugSteps {
			log.Printf("parser: step %d: sub %q\n", n+1, subStep)
		}
		var obj any
		if obj, err = Parse("step", subStep); err != nil {
			// hack - an unrecognized step might be a settlement name
			if result != nil && result.Settlement == nil {
				if r, _ := utf8.DecodeRune(subStep); unicode.IsUpper(r) {
					obj, err = &Settlement{Name: string(subStep)}, nil
				}
			}
			if err != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, err
			}
		}
		switch v := obj.(type) {
		case *BlockedByEdge:
			if result != nil { // only allowed at the beginning of the step
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("blocked by must start step")
			}
			result = &Step{
				TurnId:    turnId,
				UnitId:    unitId,
				Attempted: v.Direction,
				Result:    Blocked,
				BlockedBy: v,
			}
		//case DidNotReturn:
		//	if result != nil { // only allowed at the beginning of the step
		//		log.Printf("parser:  sub: %q\n", subStep)
		//		return nil, fmt.Errorf("multiple direction-terrain forbidden")
		//	}
		//	result = &Step{
		//		Result: Vanished,
		//	}
		case DirectionTerrain:
			if result != nil { // only allowed at the beginning of the step
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("multiple direction-terrain forbidden")
			}
			result = &Step{
				TurnId:    turnId,
				UnitId:    unitId,
				Attempted: v.Direction,
				Result:    Succeeded,
				Terrain:   v.Terrain,
			}
		case []*Edge:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("edges forbidden at beginning of step")
			}
			for _, edge := range v {
				result.Edges = append(result.Edges, edge)
			}
		case *Exhausted:
			if result != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("exhaustion must start step")
			}
			result = &Step{
				TurnId:    turnId,
				UnitId:    unitId,
				Attempted: v.Direction,
				Result:    ExhaustedMovementPoints,
				Terrain:   v.Terrain,
				Exhausted: v,
			}
		case FoundNothing:
			// ignore
		case FoundUnit:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("units forbidden at beginning of step")
			}
			result.Units = append(result.Units, string(v.Id))
		case []*Neighbor:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("neighbors forbidden at beginning of step")
			} else if result.Neighbors != nil {
				// cross compare neighbors, returning an error if either list contains the same edge
				for _, nn := range result.Neighbors {
					for _, nv := range v {
						if nn.Direction == nv.Direction {
							log.Printf("parser:  sub: %q\n", subStep)
							return nil, fmt.Errorf("duplicate neighbor direction %s", nn.Direction)
						}
					}
				}
				for _, nv := range v {
					for _, nn := range result.Neighbors {
						if nv.Direction == nn.Direction {
							log.Printf("parser:  sub: %q\n", subStep)
							return nil, fmt.Errorf("duplicate neighbor direction %s", nv.Direction)
						}
					}
				}
				result.Neighbors = append(result.Neighbors, v...)
			} else {
				result.Neighbors = v
			}
		case *ProhibitedFrom:
			if result != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("prohibition must start step")
			}
			result = &Step{
				TurnId:         turnId,
				UnitId:         unitId,
				Attempted:      v.Direction,
				Result:         Prohibited,
				Terrain:        v.Terrain,
				ProhibitedFrom: v,
			}
		case domain.Resource:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("resources forbidden at beginning of step")
			}
			result.Resources = v
		case *Settlement:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("settlement forbidden at beginning of step")
			}
			result.Settlement = v
		case []UnitID:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("units forbidden at beginning of step")
			}
			for _, u := range v {
				result.Units = append(result.Units, string(u))
			}
		case domain.Terrain:
			// valid only at the beginning of the step for status line
			if result != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("terrain must start status")
			}
			result = &Step{
				TurnId:  turnId,
				UnitId:  unitId,
				Result:  StatusLine,
				Terrain: v,
			}
		default:
			log.Printf("parser: turn %s: unit %s: line %d: sub: %q\n", turnId, unitId, lineNo, subStep)
			panic(fmt.Sprintf("unexpected %T\nplease report this", v))
		}
	}

	//if showDebug {
	//	if result != nil && (result.Resources != domain.RNone || result.Settlement != nil) {
	//		log.Printf("parser: root: showDebug: %s\n", printNodes(root))
	//		if boo, err := json.MarshalIndent(result, "", "\t"); err == nil {
	//			log.Printf("step: %s\n", string(boo))
	//		}
	//	}
	//}

	return result, nil
}
