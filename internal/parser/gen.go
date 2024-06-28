// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/internal/compass"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/unit_movement"
	"github.com/mdhender/ottomap/internal/winds"
	"log"
)

//go:generate pigeon -o grammar.go grammar.peg

type Movement_t struct {
	TurnReportId string
	LineNo       int
	Type         unit_movement.Type_e
	Winds        struct {
		Strength winds.Strength_e
		From     direction.Direction_e
	}
	Steps struct {
		Steps []Step_t
	}
	CrowsNestTerrain [][]string // indexed by step, then compass.Point_e
	Text             []byte
}

// ParseFleetMovementLine parses a fleet movement line.
// It returns the generic struct that covers all the known movement steps and cases.
func ParseFleetMovementLine(id string, lineNo int, line []byte, debug bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
	}

	if va, err := Parse(id, line, Entrypoint("FleetMovement")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("id %q: type: want Movement_t, got %T\n", id, va))
	} else {
		m.Type = mt.Type
		m.Winds.Strength = mt.Winds.Strength
		m.Winds.From = mt.Winds.From
		m.Text = mt.Text
	}
	log.Printf("%s: results %q\n", id, m.Text)

	// remove the prefix and trim the line
	if !bytes.HasPrefix(m.Text, []byte{'M', 'o', 'v', 'e'}) {
		if len(m.Text) < 8 {
			return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text))
		}
		return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text[:8]))
	}
	m.Steps.Steps = scrubSteps(bytes.TrimSpace(bytes.TrimPrefix(m.Text, []byte{'M', 'o', 'v', 'e'})))

	// we've done this over and over. movement results look like step (\ step)*.
	err := parseMovementLine(&m, debug)
	if err != nil {
		return m, err
	}

	return m, nil
}

func ParseTribeMovementLine(id string, lineNo int, line []byte, debug bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
	}

	if va, err := Parse(id, line, Entrypoint("TribeMovement")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("id %q: type: want Movement_t, got %T\n", id, va))
	} else {
		m.Type = mt.Type
		m.Text = mt.Text
	}
	log.Printf("%s: results %q\n", id, m.Text)

	// remove the prefix and trim the line
	if !bytes.HasPrefix(m.Text, []byte{'M', 'o', 'v', 'e'}) {
		if len(m.Text) < 8 {
			return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text))
		}
		return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text[:8]))
	}
	m.Steps.Steps = scrubSteps(bytes.TrimSpace(bytes.TrimPrefix(m.Text, []byte{'M', 'o', 'v', 'e'})))

	err := parseMovementLine(&m, debug)
	if err != nil {
		return m, err
	}

	return m, nil
}

func parseMovementLine(m *Movement_t, debug bool) error {
	// we've done this over and over. movement results look like step (\ step)*.
	for n, step := range m.Steps.Steps {
		log.Printf("%s: %d: text %q\n", m.TurnReportId, n+1, step.Text)

		// ugh. not happy having to constantly add the crows nest, but it's a generic structure
		m.CrowsNestTerrain = append(m.CrowsNestTerrain, []string{})

		// steps mostly look the same. they are the observations of the immediate terrain (the hex the unit is in).
		// if the movement line is a fleet movement, it may contain additional observations for the adjacent hexes and those one hex away.
		// our first task is to split the steps into sections for this hex, the inner ring of hexes and the outer ring.
		var thisHex, innerRing, outerRing []byte
		var ok bool

		thisHex = step.Text

		// does this hex contain observations of the inner ring?
		thisHex, innerRing, ok = bytes.Cut(thisHex, []byte{'-', '('})
		if ok {
			// it does, so there must be observations of the outer ring, too
			innerRing, outerRing, ok = bytes.Cut(innerRing, []byte{')', '('})
			if !ok {
				return fmt.Errorf("step %d: contains '-(' but not ')(", n+1)
			}
			// outer ring must end with a closing parentheses
			if outerRing[len(outerRing)-1] != ')' {
				return fmt.Errorf("step %d: contains text after ')'", n+1)
			}
			// remove that parentheses to make later processing simpler
			outerRing = outerRing[:len(outerRing)-1]
		}

		// clean up all the observations. note that after we do, any or all of these may be empty.
		thisHex = bytes.TrimSpace(bytes.TrimRight(thisHex, ", \t"))
		innerRing = bytes.TrimSpace(bytes.TrimRight(innerRing, ", \t"))
		outerRing = bytes.TrimSpace(bytes.TrimRight(outerRing, ", \t"))

		// parse this hex
		if len(thisHex) != 0 {
			log.Printf("%s: %d: dirt %q\n", m.TurnReportId, n+1, thisHex)
		}

		// parse the inner ring
		if len(innerRing) != 0 {
			log.Printf("%s: %d: deck %q\n", m.TurnReportId, n+1, innerRing)
		}

		// parse the outer ring
		if len(outerRing) != 0 {
			log.Printf("%s: %d: crow %q\n", m.TurnReportId, n+1, outerRing)
		}
		for nn, step := range bytes.Split(outerRing, []byte{','}) {
			step = bytes.TrimSpace(step)
			if len(step) == 0 {
				continue
			}
			// log.Printf("%s: %d: crow %d: %q\n", id, n+1, nn+1, cnoStep)
			if va, err := Parse(m.TurnReportId, step, Entrypoint("CrowsNestObservation")); err != nil {
				log.Printf("%s: %d: crow %2d: %q\n", m.TurnReportId, n+1, nn+1, step)
				log.Printf("error: %s: %d: crow %2d: %v\n", m.TurnReportId, n+1, nn+1, err)
				return err
			} else if cno, ok := va.(CrowsNestObservation_t); !ok {
				panic(fmt.Errorf("id %q: type: want CrowsNestObservation_t, got %T\n", m.TurnReportId, va))
			} else {
				if len(m.CrowsNestTerrain[n]) == 0 {
					m.CrowsNestTerrain[n] = make([]string, 13)
				}
				m.CrowsNestTerrain[n][cno.Point] = cno.Terrain
			}
		}
	}

	if debug {
		for n := range m.CrowsNestTerrain {
			if len(m.CrowsNestTerrain[n]) == 0 {
				continue
			}
			for p := compass.North; p <= compass.NorthNorthWest; p++ {
				if m.CrowsNestTerrain[n][p] == "" {
					continue
				}
				log.Printf("%s: %d: crow %-14s %s\n", m.TurnReportId, n+1, p, m.CrowsNestTerrain[n][p])
			}
		}
	}

	return nil
}

type Step_t struct {
	Text  []byte
	Text1 []byte
	Text2 []byte
}

// scrubSteps splits the line into individual steps. steps are separated by backslashes.
// leading and trailing spaces and any trailing commas are from each step.
// empty steps are ignored. maybe they shouldn't be.
func scrubSteps(line []byte) (steps []Step_t) {
	for _, step := range bytes.Split(line, []byte{'\\'}) {
		step = bytes.TrimSpace(bytes.TrimRight(step, ", \t"))
		if len(step) != 0 {
			steps = append(steps, Step_t{Text: step})
		}
	}
	return steps
}
