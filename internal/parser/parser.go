// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/resources"
	"github.com/mdhender/ottomap/internal/results"
	"github.com/mdhender/ottomap/internal/unit_movement"
	"github.com/mdhender/ottomap/internal/winds"
	"log"
	"regexp"
	"unicode"
	"unicode/utf8"
)

//go:generate pigeon -o grammar.go grammar.peg

var (
	rxCourierSection  = regexp.MustCompile(`^Courier \d{4}c\d, ,`)
	rxElementSection  = regexp.MustCompile(`^Element \d{4}e\d, ,`)
	rxFleetSection    = regexp.MustCompile(`^Fleet \d{4}f\d, ,`)
	rxFleetMovement   = regexp.MustCompile(`^(CALM|MILD|STRONG|GALE)\s(NE|SE|SW|NW|N|S)\sFleet\sMovement:\sMove\s`)
	rxGarrisonSection = regexp.MustCompile(`^Garrison \d{4}g\d, ,`)
	rxScoutLine       = regexp.MustCompile(`^Scout \d:Scout `)
	rxTribeSection    = regexp.MustCompile(`^Tribe \d{4}, ,`)
)

func ParseInput(id string, input []byte, debugParser, debugSections, debugSteps, debugNodes bool) ([]*Movement_t, error) {
	log.Printf("parser: %q\n", id)
	debugp := func(format string, args ...any) {
		if debugParser {
			log.Printf(format, args...)
		}
	}
	debugs := func(format string, args ...any) {
		if debugSections {
			log.Printf(format, args...)
		}
	}
	debugp("%s: parser: %8d bytes\n", id, len(input))

	var ms []*Movement_t
	var m *Movement_t
	var currentTurn, nextTurn string
	var statusLinePrefix []byte
	for n, line := range bytes.Split(input, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		lineNo := n + 1

		if rxCourierSection.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 14))
			mt, err := ParseLocationLine(id, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", id, m.UnitId, lineNo, slug(line, 14))
				return ms, nil
			}
			ms, m = append(ms, &mt), &mt
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", m.UnitId))
		} else if rxElementSection.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 14))
			mt, err := ParseLocationLine(id, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", id, m.UnitId, lineNo, slug(line, 14))
				return ms, nil
			}
			ms, m = append(ms, &mt), &mt
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", m.UnitId))
		} else if rxFleetSection.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 12))
			mt, err := ParseLocationLine(id, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", id, m.UnitId, lineNo, slug(line, 12))
				return ms, nil
			}
			ms, m = append(ms, &mt), &mt
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", m.UnitId))
		} else if rxGarrisonSection.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 15))
			mt, err := ParseLocationLine(id, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", id, m.UnitId, lineNo, slug(line, 15))
				return ms, nil
			}
			ms, m = append(ms, &mt), &mt
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", m.UnitId))
		} else if rxTribeSection.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 10))
			mt, err := ParseLocationLine(id, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", id, m.UnitId, lineNo, slug(line, 10))
				return ms, err
			}
			ms, m = append(ms, &mt), &mt
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", m.UnitId))
		} else if m == nil {
			log.Printf("%s: %s: %d: found line outside of section: %q\n", id, m.UnitId, lineNo, slug(line, 20))
		} else if bytes.HasPrefix(line, []byte("Current Turn ")) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 19))
			if va, err := Parse(id, line, Entrypoint("TurnInfo")); err != nil {
				log.Printf("%s: %s: %d: error parsing turn info", id, m.UnitId, lineNo)
				return ms, err
			} else if turnInfo, ok := va.(TurnInfo_t); !ok {
				log.Printf("%s: %s: %d: error parsing turn info", id, m.UnitId, lineNo)
				log.Printf("error: parser.TurnInfo_t, got %T\n", va)
				log.Printf("please report this error\n")
				panic(fmt.Sprintf("unexpected type %T", va))
			} else {
				m.CurrentTurn = fmt.Sprintf("%04d-%02d", turnInfo.CurrentTurn.Year, turnInfo.CurrentTurn.Month)
				if lineNo == 2 {
					m.NextTurn = fmt.Sprintf("%04d-%02d", turnInfo.NextTurn.Year, turnInfo.NextTurn.Month)
					currentTurn, nextTurn = m.CurrentTurn, m.NextTurn
				} else {
					m.NextTurn = nextTurn
				}
				if m.CurrentTurn != currentTurn {
					log.Printf("%s: %s: %d: currTurn %q", id, m.UnitId, lineNo, m.CurrentTurn)
					log.Printf("error: expected %q, got %q", currentTurn, m.CurrentTurn)
					return ms, fmt.Errorf("invalid current turn")
				}
			}
		} else if rxFleetMovement.Match(line) {
			pfx, _, ok := bytes.Cut(line, []byte{':'})
			if !ok {
				pfx = []byte(slug(line, 23))
			}
			debugs("%s: %d: found %q\n", id, lineNo, pfx)
			mt, err := ParseFleetMovementLine(id, m.UnitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			m.Winds.Strength = mt.Winds.Strength
			m.Winds.From = mt.Winds.From
			for _, step := range mt.Steps {
				step.Movement = m
				step.TurnId = m.CurrentTurn
				step.UnitId = m.UnitId
				step.No = len(m.Steps) + 1
				m.Steps = append(m.Steps, step)
			}
		} else if bytes.HasPrefix(line, []byte("Tribe Follows ")) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 13))
			mt, err := ParseTribeFollowsLine(id, m.UnitId, lineNo, line, false)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			m.Follows = mt.Follows
		} else if bytes.HasPrefix(line, []byte("Tribe Goes to ")) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 14))
			mt, err := ParseTribeGoesToLine(id, m.UnitId, lineNo, line, false)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			m.GoesTo = mt.GoesTo
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 14))
			mt, err := ParseTribeMovementLine(id, m.UnitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			for _, step := range mt.Steps {
				step.Movement = m
				step.TurnId = m.CurrentTurn
				step.UnitId = m.UnitId
				step.No = len(m.Steps) + 1
				m.Steps = append(m.Steps, step)
			}
		} else if rxScoutLine.Match(line) {
			debugs("%s: %d: found %q\n", id, lineNo, slug(line, 14))
			mt, err := ParseScoutMovementLine(id, m.UnitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			for _, step := range mt.Steps {
				step.Movement = m
				step.TurnId = m.CurrentTurn
				step.UnitId = UnitId_t(fmt.Sprintf("%ss%d", m.UnitId, mt.ScoutNo))
				step.No = len(m.Steps) + 1
				m.Steps = append(m.Steps, step)
			}
		} else if bytes.HasPrefix(line, statusLinePrefix) {
			debugs("%s: %d: found %q\n", id, lineNo, statusLinePrefix)
			mt, err := ParseStatusLine(id, m.UnitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return ms, err
			}
			m.Type = mt.Type
			for _, step := range mt.Steps {
				step.Movement = m
				step.TurnId = m.CurrentTurn
				step.UnitId = m.UnitId
				step.No = len(m.Steps) + 1
				m.Steps = append(m.Steps, step)
			}
		}
	}

	return ms, nil
}

func slug(b []byte, n int) string {
	if len(b) < n {
		return string(b)
	}
	return string(b[:n])
}

type Movement_t struct {
	TurnReportId string
	LineNo       int

	UnitId  UnitId_t
	ScoutNo int
	Type    unit_movement.Type_e

	PreviousHex string
	CurrentHex  string

	CurrentTurn string
	NextTurn    string

	Winds struct {
		Strength winds.Strength_e
		From     direction.Direction_e
	}

	// movement results
	Follows UnitId_t
	GoesTo  string
	Steps   []*Step_t

	Text []byte
}

type Step_t struct {
	Movement *Movement_t
	TurnId   string
	UnitId   UnitId_t
	No       int // original step number, indexed from 1

	// Attempted direction is the direction the unit tried to move.
	// It will be Unknown if the unit stays in place.
	// When the unit fails to move, this will be derived from the failed results.
	Attempted direction.Direction_e

	// Result is the result of the step.
	// The attempt may succeed or fail; this captures the reasons.
	Result results.Result_e

	// properties below may be set even if the step failed.
	// that means they may be for the hex where the unit started.

	GridHex string
	Terrain domain.Terrain

	BlockedBy        *BlockedByEdge_t
	Edges            []*Edge_t
	Exhausted        *Exhausted_t
	Neighbors        []*Neighbor_t
	ProhibitedFrom   *ProhibitedFrom_t
	Resources        resources.Resource_e
	Settlement       *Settlement_t
	Units            []UnitId_t
	CrowsNestTerrain []string // indexed by step, then compass.Point_e
	Text             []byte
}

// ParseFleetMovementLine parses a fleet movement line.
// It returns the generic struct that covers all the known movement steps and cases.
func ParseFleetMovementLine(id string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("FleetMovement")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("%s: %d: type: want Movement_t, got %T\n", m.TurnReportId, m.LineNo, va))
	} else {
		m.Type = mt.Type
		m.Winds.Strength = mt.Winds.Strength
		m.Winds.From = mt.Winds.From
		m.Text = mt.Text
	}
	if debugSteps {
		log.Printf("parser: %s: %s: %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.Text)
	}

	// remove the prefix and trim the line
	if !bytes.HasPrefix(m.Text, []byte{'M', 'o', 'v', 'e'}) {
		if len(m.Text) < 8 {
			return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text))
		}
		return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text[:8]))
	}
	line = bytes.TrimPrefix(m.Text, []byte{'M', 'o', 'v', 'e'})

	// we've done this over and over. movement results look like step (\ step)*.
	err := parseMovementLine(&m, line, debugSteps, debugNodes)
	if err != nil {
		return m, err
	}

	return m, nil
}

func ParseLocationLine(id string, lineNo int, line []byte, debug bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
	}

	if va, err := Parse(id, line, Entrypoint("Location")); err != nil {
		log.Printf("%s: %d: courier %q\n", id, lineNo, slug(line, 14))
		return m, err
	} else if location, ok := va.(Location_t); !ok {
		log.Printf("%s: %d: location: %q\n", id, lineNo, slug(line, 15))
		log.Printf("error: invalid type\n")
		log.Printf("please report this error")
		panic(fmt.Errorf("want Location_t, got %T", va))
	} else {
		m.UnitId = location.UnitId
		m.PreviousHex = location.PreviousHex
		m.CurrentHex = location.CurrentHex
	}

	return m, nil
}

func ParseScoutMovementLine(id string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("ScoutMovement")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("%s: %d: type: want Movement_t, got %T\n", m.TurnReportId, m.LineNo, va))
	} else {
		m.Type = mt.Type
		m.ScoutNo = mt.ScoutNo
		m.Text = mt.Text
	}
	if debugSteps {
		log.Printf("parser: %s: %s: %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.Text)
	}

	// remove the prefix and trim the line
	if !bytes.HasPrefix(m.Text, []byte{'S', 'c', 'o', 'u', 't'}) {
		if len(m.Text) < 8 {
			return m, fmt.Errorf("expected 'Scout', found '%s'", string(m.Text))
		}
		return m, fmt.Errorf("expected 'Scout', found '%s'", string(m.Text[:8]))
	}
	line = bytes.TrimPrefix(m.Text, []byte{'S', 'c', 'o', 'u', 't'})

	err := parseMovementLine(&m, line, debugSteps, debugNodes)
	if err != nil {
		return m, err
	}

	return m, nil
}

func ParseStatusLine(id string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("StatusLine")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("%s: %d: type: want Movement_t, got %T\n", m.TurnReportId, m.LineNo, va))
	} else {
		m.Type = mt.Type
		m.UnitId = mt.UnitId
		m.Text = mt.Text
	}
	if debugSteps {
		log.Printf("parser: %s: %s: %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.Text)
	}

	// remove the prefix and trim the line
	_, steps, ok := bytes.Cut(line, []byte{':'})
	if !ok {
		return m, fmt.Errorf("expected 'Status:', found '%s'", slug(m.Text, 8))
	}

	err := parseMovementLine(&m, steps, debugSteps, debugNodes)
	if err != nil {
		return m, err
	}

	return m, nil
}

func ParseTribeFollowsLine(id string, unitId UnitId_t, lineNo int, line []byte, debug bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("TribeFollows")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("%s: %d: type: want Movement_t, got %T\n", m.TurnReportId, m.LineNo, va))
	} else {
		m.Type = mt.Type
		m.Follows = mt.Follows
	}
	if debug {
		log.Printf("parser: %s: %s: %d: follows %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.Follows)
	}

	return m, nil
}

func ParseTribeGoesToLine(id string, unitId UnitId_t, lineNo int, line []byte, debug bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("TribeGoesTo")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("%s: %d: type: want Movement_t, got %T\n", m.TurnReportId, m.LineNo, va))
	} else {
		m.Type = mt.Type
		m.GoesTo = mt.GoesTo
	}
	if debug {
		log.Printf("parser: %s: %s: %d: goes to %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.GoesTo)
	}

	return m, nil
}

func ParseTribeMovementLine(id string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) (Movement_t, error) {
	m := Movement_t{
		TurnReportId: id,
		LineNo:       lineNo,
		UnitId:       unitId,
	}

	if va, err := Parse(id, line, Entrypoint("TribeMovement")); err != nil {
		return m, err
	} else if mt, ok := va.(Movement_t); !ok {
		panic(fmt.Errorf("id %q: type: want Movement_t, got %T\n", id, va))
	} else {
		m.Type = mt.Type
		m.Text = mt.Text
	}
	if debugSteps {
		log.Printf("parser: %s: %s: %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, m.Text)
	}

	// remove the prefix
	if !bytes.HasPrefix(m.Text, []byte{'M', 'o', 'v', 'e'}) {
		if len(m.Text) < 8 {
			return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text))
		}
		return m, fmt.Errorf("expected 'Move', found '%s'", string(m.Text[:8]))
	}
	line = bytes.TrimPrefix(m.Text, []byte{'M', 'o', 'v', 'e'})

	err := parseMovementLine(&m, line, debugSteps, debugNodes)
	if err != nil {
		return m, err
	}

	return m, nil
}

func parseMovementLine(m *Movement_t, line []byte, debugSteps, debugNodes bool) error {
	// split the line into single steps
	m.Steps = splitSteps(line)

	// we've done this over and over. movement results look like step (\ step)*.
	for _, step := range m.Steps {
		if debugSteps {
			log.Printf("%s: %s: %d: step %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, step.Text)
		}

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
				log.Printf("%s: %s: %d: step %d: iring %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, innerRing)
				return fmt.Errorf("inner ring contains '-(' but not ')(")
			}
			// outer ring must end with a closing parentheses
			if bytes.IndexByte(outerRing, ')') == -1 {
				log.Printf("%s: %s: %d: step %d: oring %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, outerRing)
				return fmt.Errorf("outer ring missing ')'")
			}
			// outer ring must end with a closing parentheses
			if outerRing[len(outerRing)-1] != ')' {
				log.Printf("%s: %s: %d: step %d: oring %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, outerRing)
				return fmt.Errorf("outer ring contains text after ')'")
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
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: dirt %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, slug(thisHex, 44))
			}

			err := step.parse(m, "?", thisHex, debugSteps, debugNodes)
			if err != nil {
				return err
			}
		}

		// parse the inner ring
		if len(innerRing) != 0 {
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: deck %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, slug(innerRing, 44))
			}
		}

		// parse the outer ring
		if len(outerRing) != 0 {
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: crow %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, slug(outerRing, 44))
			}

			step.CrowsNestTerrain = make([]string, 13)

			for nn, orStep := range bytes.Split(outerRing, []byte{','}) {
				orStep = bytes.TrimSpace(orStep)
				if len(orStep) == 0 {
					continue
				}
				crowNo := nn + 1
				if va, err := Parse(m.TurnReportId, orStep, Entrypoint("CrowsNestObservation")); err != nil {
					log.Printf("%s: %s: %d: step %d: crow %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, crowNo, orStep)
					log.Printf("error: %s: %d: crow %2d: %v\n", m.TurnReportId, step.No, nn+1, err)
					return err
				} else if cno, ok := va.(CrowsNestObservation_t); !ok {
					log.Printf("%s: %s: %d: step %d: crow %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, step.No, crowNo, orStep)
					log.Printf("error: want CrowsNestObservation_t, got %T", va)
					log.Printf("please report this error")
					panic(fmt.Errorf("want CrowsNestObservation_t, got %T\n", va))
				} else {
					step.CrowsNestTerrain[cno.Point] = cno.Terrain
				}
			}
		}
	}

	return nil
}

func (s *Step_t) parse(m *Movement_t, unitId string, line []byte, debugSteps, debugNodes bool) error {
	line = bytes.TrimSpace(bytes.TrimRight(line, ","))
	if debugSteps {
		log.Printf("parser: %s: %s: %d: step %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, line)
	}

	root := hexReportToNodes(line, debugNodes)
	steps, err := nodesToSteps(root)
	if err != nil {
		log.Printf("parser: %s: %s: %d: step %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, line)
		return err
	}

	// parse each sub-step separately.
	for n, subStep := range steps {
		subStepNo := n + 1
		if debugSteps {
			log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
		}

		var obj any
		if obj, err = Parse("step", subStep, Entrypoint("Step")); err != nil {
			// hack - an unrecognized step might be a settlement name
			if s.Settlement == nil {
				// if it is the first thing after the direction-terrain code
				if s.Result != results.Unknown {
					if r, _ := utf8.DecodeRune(subStep); unicode.IsUpper(r) {
						obj, err = &Settlement_t{Name: string(subStep)}, nil
					}
				}
			}
			if err != nil {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				log.Printf("error: %v\n", err)
				return fmt.Errorf("error parsing step")
			}
		}
		switch v := obj.(type) {
		case *BlockedByEdge_t:
			if s.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("blocked by must start sub-step")
			}
			s.Attempted = v.Direction
			s.Result = results.Blocked
			s.BlockedBy = v
		case DirectionTerrain_t:
			if s.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("multiple direction-terrain forbidden")
			}
			s.Attempted = v.Direction
			s.Result = results.Succeeded
			s.Terrain = v.Terrain
		case []*Edge_t:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("edges forbidden at beginning of step")
			}
			for _, edge := range v { // todo: de-dup edges
				s.Edges = append(s.Edges, edge)
			}
		case *Exhausted_t:
			if s.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("exhaustion must start step")
			}
			s.Attempted = v.Direction
			s.Result = results.ExhaustedMovementPoints
			s.Terrain = v.Terrain
			s.Exhausted = v
		case FoundNothing_t: // ignore
		case FoundUnit_t:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("units forbidden at beginning of step")
			}
			s.Units = append(s.Units, v.Id)
		case []FoundUnit_t:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("units forbidden at beginning of step")
			}
			for _, unit := range v { // todo: de-duplicate units
				s.Units = append(s.Units, unit.Id)
			}
		case Longhouse_t: // ignore
		case MissingEdge_t: // ignore
		case []*Neighbor_t:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("neighbors forbidden at beginning of step")
			} else if s.Neighbors != nil {
				// cross compare neighbors, returning an error if either list contains the same edge
				for _, nn := range s.Neighbors {
					for _, nv := range v {
						if nn.Direction == nv.Direction {
							log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
							return fmt.Errorf("duplicate neighbor direction %s", nn.Direction)
						}
					}
				}
				for _, nv := range v {
					for _, nn := range s.Neighbors {
						if nv.Direction == nn.Direction {
							log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
							return fmt.Errorf("duplicate neighbor direction %s", nv.Direction)
						}
					}
				}
				s.Neighbors = append(s.Neighbors, v...)
			} else {
				s.Neighbors = v
			}
		case *ProhibitedFrom_t:
			if s.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("prohibition must start step")
			}
			s.Attempted = v.Direction
			s.Result = results.Prohibited
			s.Terrain = v.Terrain
			s.ProhibitedFrom = v
		case resources.Resource_e:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("resources forbidden at beginning of step")
			}
			s.Resources = v
		case *Settlement_t:
			if s.Result == results.Unknown {
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("settlement forbidden at beginning of step")
			}
			s.Settlement = v
		case domain.Terrain:
			if s.Result != results.Unknown { // valid only at the beginning of the step for status line
				log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
				return fmt.Errorf("terrain must start status")
			}
			s.Result = results.StatusLine
			s.Terrain = v
		default:
			log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", m.TurnReportId, m.UnitId, m.LineNo, s.No, subStepNo, subStep)
			log.Printf("error: unexpected type %T\n", v)
			log.Printf("please report this error\n")
			panic(fmt.Sprintf("unexpected %T", v))
		}
	}

	return nil
}

// splitSteps splits the line into individual steps. steps are separated by backslashes.
// leading and trailing spaces and any trailing commas are from each step.
// empty steps are ignored. maybe they shouldn't be.
func splitSteps(line []byte) (steps []*Step_t) {
	for n, step := range bytes.Split(line, []byte{'\\'}) {
		step = bytes.TrimSpace(bytes.TrimRight(step, ", \t"))
		if len(step) != 0 {
			steps = append(steps, &Step_t{No: n + 1, Text: step})
		}
	}
	return steps
}

// BlockedByEdge_t is returned when a step fails because the unit was blocked by an edge feature.
type BlockedByEdge_t struct {
	Direction direction.Direction_e
	Edge      edges.Edge_e
}

func (b *BlockedByEdge_t) String() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("b(%s-%s)", b.Direction, b.Edge)
}

type DidNotReturn_t struct{}

func (d *DidNotReturn_t) String() string {
	if d == nil {
		return ""
	}
	return "did not return"
}

// DirectionTerrain_t is the first component returned from a successful step.
type DirectionTerrain_t struct {
	Direction direction.Direction_e
	Terrain   domain.Terrain
}

func (d DirectionTerrain_t) String() string {
	return fmt.Sprintf("%s-%s", d.Direction, d.Terrain)
}

// Edge_t is an edge feature that the unit sees in the current hex.
type Edge_t struct {
	Direction direction.Direction_e
	Edge      edges.Edge_e
}

func (e *Edge_t) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", e.Direction, e.Edge)
}

// Exhausted_t is returned when a step fails because the unit was exhausted.
type Exhausted_t struct {
	Direction direction.Direction_e
	Terrain   domain.Terrain
}

func (e *Exhausted_t) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("x(%s-%s)", e.Direction, e.Terrain)
}

type FoundNothing_t struct{}

func (f *FoundNothing_t) String() string {
	if f == nil {
		return ""
	}
	return "nothing of interest found"
}

type FoundUnit_t struct {
	Id UnitId_t
}

type Location_t struct {
	UnitId      UnitId_t
	Message     string
	CurrentHex  string
	PreviousHex string
}

type Longhouse_t struct {
	Id       string
	Capacity int
}

// MissingEdge_t is returned for "No River Adjacent to Hex"
type MissingEdge_t struct{}

// Neighbor_t is the terrain in a neighboring hex that the unit from the current hex.
type Neighbor_t struct {
	Direction direction.Direction_e
	Terrain   domain.Terrain
}

func (n *Neighbor_t) String() string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", n.Direction, n.Terrain)
}

type NoGroupsFound_t struct{}

// ProhibitedFrom_t is returned when a step fails because the unit is not allowed to enter the terrain.
type ProhibitedFrom_t struct {
	Direction direction.Direction_e
	Terrain   domain.Terrain
}

func (p *ProhibitedFrom_t) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("p(%s-%s)", p.Direction, p.Terrain)
}

type UnitId_t string

func (u UnitId_t) String() string {
	return string(u)
}
