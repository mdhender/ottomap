// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lbmoves

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
	"regexp"
	"unicode"
	"unicode/utf8"
)

//go:generate pigeon -o grammar.go grammar.peg

var (
	rxScoutLine *regexp.Regexp
)

// ParseMoveResults parses the results of a Land Based Movement.
//
// The line should be the text as extracted directly from the turn report.
// Handles Tribe Follows, Tribe Movement, and Scout lines.
//
// Returns the steps and the first error encountered.
func ParseMoveResults(line []byte) ([]*Step, error) {
	if rxScoutLine == nil {
		rxScoutLine = regexp.MustCompile(`^Scout [12345678]:Scout `)
	}
	if bytes.HasPrefix(line, []byte("Tribe Follows")) {
		return parseTribeFollows(line)
	} else if bytes.HasPrefix(line, []byte("Tribe Movement: Move ")) {
		return parseSteps(line, bytes.TrimPrefix(line, []byte("Tribe Movement: Move ")))
	} else if rxScoutLine.Match(line) {
		return parseSteps(line, line[len("Scout ?:Scout "):])
	}
	return nil, cerrs.ErrNotMovementResults
}

func parseTribeFollows(line []byte) ([]*Step, error) {
	fields := bytes.Split(bytes.TrimSpace(line), []byte{' '})
	if len(fields) != 3 {
		return nil, cerrs.ErrMissingFollowsUnit
	}
	return []*Step{{
		Result:  Follows,
		Follows: string(fields[2]),
	}}, nil
}

// parseSteps parses all the steps from the results of a Land Based Movement.
func parseSteps(line, steps []byte) (results []*Step, err error) {
	// split the steps into single steps, which are backslash-separated, and
	// parse each step individually after trimming spaces and trailing commas.
	for _, step := range bytes.Split(steps, []byte{'\\'}) {
		if result, err := parseStep(bytes.TrimSpace(step)); err != nil {
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
func parseStep(step []byte) (result *Step, err error) {
	// split the step into its components, which are comma-separated, and
	// parse each component individually.
	var activeNeighbor *Neighbor
	var patrolling bool
	for _, subStep := range bytes.Split(step, []byte{','}) {
		subStep = bytes.TrimSpace(subStep)
		if len(subStep) == 0 {
			continue
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
				Attempted: v.Direction,
				Result:    Blocked,
				BlockedBy: v,
			}
		case DidNotReturn:
			if result != nil { // only allowed at the beginning of the step
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("multiple direction-terrain forbidden")
			}
			result = &Step{
				Result: Vanished,
			}
		case directions.Direction:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("terrain forbidden at beginning of step")
			}
			if activeNeighbor != nil {
				result.Neighbors = append(result.Neighbors, &Neighbor{Direction: v, Terrain: activeNeighbor.Terrain})
			} else {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("unexpected terrain")
			}
		case DirectionTerrain:
			if result != nil { // only allowed at the beginning of the step
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("multiple direction-terrain forbidden")
			}
			result = &Step{
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
			if activeNeighbor != nil {
				activeNeighbor = nil
			}
		case *Exhausted:
			if result != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("exhaustion must start step")
			}
			result = &Step{
				Attempted: v.Direction,
				Result:    ExhaustedMovementPoints,
				Terrain:   v.Terrain,
				Exhausted: v,
			}
		case FoundNothing:
			// ignore?
		case FoundUnit:
			patrolling = true
		case *Neighbor:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("neighbors forbidden at beginning of step")
			}
			result.Neighbors = append(result.Neighbors, v)
			activeNeighbor = v
		case *ProhibitedFrom:
			if result != nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("prohibition must start step")
			}
			result = &Step{
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
			result.Resources = append(result.Resources, v)
		case *Settlement:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("settlement forbidden at beginning of step")
			}
			result.Settlement = v
		case UnitID:
			if result == nil {
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("units forbidden at beginning of step")
			}
			if !patrolling {
				result.Units = append(result.Units, string(v))
			}
		default:
			log.Printf("parser:  sub: %q\n", subStep)
			panic(fmt.Sprintf("unexpected %T", v))
		}
	}
	return result, nil
}
