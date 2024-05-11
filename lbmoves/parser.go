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
func ParseMoveResults(line []byte) ([]*Step, error) {
	if rxScoutLine == nil {
		rxScoutLine = regexp.MustCompile(`^Scout [12345678]:Scout `)
		rxStatusLine = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])? Status: `)
	}
	if bytes.HasPrefix(line, []byte("Tribe Follows")) {
		return parseTribeFollows(line)
	} else if bytes.HasPrefix(line, []byte("Tribe Movement: Move ")) {
		return parseSteps(line, bytes.TrimPrefix(line, []byte("Tribe Movement: Move ")))
	} else if rxScoutLine.Match(line) {
		return parseSteps(line, line[len("Scout ?:Scout "):])
	} else if rxStatusLine.Match(line) {
		_, b, _ := bytes.Cut(line, []byte{':'})
		return parseSteps(line, b)
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
		if result, err := parseStep(step); err != nil {
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
	//log.Printf("parser: step: %q\n", step)
	step = bytes.TrimSpace(step)
	//log.Printf("parser: step: %q\n", step)

	// split the step into its components, which are comma-separated, and remove the empty components.
	type node struct {
		text []byte
		next *node
	}
	var root, tail *node
	for _, component := range bytes.Split(step, []byte{','}) {
		if component = bytes.TrimSpace(component); len(component) != 0 {
			if root == nil {
				root = &node{text: component}
				tail = root
			} else {
				tail.next = &node{text: component}
				tail = tail.next
			}
		}
	}

	// that step broke some things. there are components that use commas as separators internally.
	// we need to find them and splice them back together. brute force it.
	for tmp := root; tmp != nil && tmp.next != nil; {
		if acceptFordEdge(tmp.text) && acceptDirection(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else if acceptLakeEdge(tmp.text) && acceptDirection(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else if acceptOceanEdge(tmp.text) && acceptDirection(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else if acceptPassEdge(tmp.text) && acceptDirection(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else if acceptPatrolledAndFound(tmp.text) && acceptUnitId(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else if acceptRiverEdge(tmp.text) && acceptDirection(tmp.next.text) {
			tmp.text = append(tmp.text, ' ')
			tmp.text = append(tmp.text, tmp.next.text...)
			tmp.next = tmp.next.next
		} else {
			tmp = tmp.next
		}
	}

	// convert our linked list into a slice for parsing.
	var steps [][]byte
	for tmp := root; tmp != nil; tmp = tmp.next {
		text := bytes.TrimSpace(tmp.text)
		if len(text) != 0 {
			steps = append(steps, tmp.text)
		}
	}
	root = nil
	//for n, subStep := range steps {
	//	log.Printf("parser:  sub: %2d: %q\n", n+1, subStep)
	//}

	// parse each sub-step separately.
	for _, subStep := range steps {
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
				log.Printf("parser:  sub: %q\n", subStep)
				return nil, fmt.Errorf("multiple neighbors forbidden")
			}
			result.Neighbors = v
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
			result.Resources = v
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
			result.Units = append(result.Units, string(v))
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
				Result:  Status,
				Terrain: v,
			}
		default:
			log.Printf("parser:  sub: %q\n", subStep)
			panic(fmt.Sprintf("unexpected %T", v))
		}
	}

	//if result != nil && result.Result == Status {
	//	if boo, err := json.MarshalIndent(result, "", "\t"); err == nil {
	//		log.Printf("step: %s\n", string(boo))
	//	}
	//}

	return result, nil
}
