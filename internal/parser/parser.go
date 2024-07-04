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
	"github.com/mdhender/ottomap/internal/terrain"
	"github.com/mdhender/ottomap/internal/unit_movement"
	"github.com/mdhender/ottomap/internal/winds"
	"log"
	"regexp"
	"sort"
	"strings"
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

const (
	LastTurnCurrentLocationObscured = "0902-01"
)

type ParseConfig struct {
	Ignore struct {
		Scouts bool
		Logged struct {
			Scouts bool
		}
	}
}

func ParseInput(fid, tid string, input []byte, debugParser, debugSections, debugSteps, debugNodes bool, cfg ParseConfig) (*Turn_t, error) {
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
	debugp("%s: parser: %8d bytes\n", fid, len(input))

	t := &Turn_t{
		UnitMoves: map[UnitId_t]*Moves_t{},
	}
	var unitId UnitId_t // current unit being parsed
	var moves *Moves_t  // current move being parsed

	var statusLinePrefix []byte
	for n, line := range bytes.Split(input, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		lineNo := n + 1

		if rxCourierSection.Match(line) {
			unitId = UnitId_t(line[8:14])
			debugs("%s: %d: found %q\n", fid, lineNo, unitId)
			location, err := ParseLocationLine(fid, tid, unitId, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 14))
				return t, nil
			} else if _, ok := t.UnitMoves[unitId]; ok {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 14))
				return t, fmt.Errorf("duplicate unit in turn")
			} else if t.Id > LastTurnCurrentLocationObscured && strings.HasPrefix(location.CurrentHex, "##") {
				log.Printf("info: last turn current location is obscured is %s\n", LastTurnCurrentLocationObscured)
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, location.CurrentHex)
				return t, fmt.Errorf("current location is obscured")
			}
			moves = &Moves_t{TurnId: t.Id, Id: unitId, FromHex: location.PreviousHex, ToHex: location.CurrentHex}
			t.UnitMoves[moves.Id] = moves
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", unitId))
		} else if rxElementSection.Match(line) {
			unitId = UnitId_t(line[8:14])
			debugs("%s: %d: found %q\n", fid, lineNo, unitId)
			location, err := ParseLocationLine(fid, tid, unitId, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 14))
				return t, nil
			} else if _, ok := t.UnitMoves[unitId]; ok {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 14))
				return t, fmt.Errorf("duplicate unit in turn")
			}
			moves = &Moves_t{TurnId: t.Id, Id: unitId, FromHex: location.PreviousHex, ToHex: location.CurrentHex}
			t.UnitMoves[moves.Id] = moves
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", unitId))
		} else if rxFleetSection.Match(line) {
			unitId = UnitId_t(line[6:12])
			debugs("%s: %d: found %q\n", fid, lineNo, unitId)
			location, err := ParseLocationLine(fid, tid, unitId, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 12))
				return t, nil
			} else if _, ok := t.UnitMoves[unitId]; ok {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 12))
				return t, fmt.Errorf("duplicate unit in turn")
			}
			moves = &Moves_t{TurnId: t.Id, Id: unitId, FromHex: location.PreviousHex, ToHex: location.CurrentHex}
			t.UnitMoves[moves.Id] = moves
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", unitId))
		} else if rxGarrisonSection.Match(line) {
			unitId = UnitId_t(line[9:15])
			debugs("%s: %d: found %q\n", fid, lineNo, unitId)
			location, err := ParseLocationLine(fid, tid, unitId, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 15))
				return t, nil
			} else if _, ok := t.UnitMoves[unitId]; ok {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 15))
				return t, fmt.Errorf("duplicate unit in turn")
			}
			moves = &Moves_t{TurnId: t.Id, Id: unitId, FromHex: location.PreviousHex, ToHex: location.CurrentHex}
			t.UnitMoves[moves.Id] = moves
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", unitId))
		} else if rxTribeSection.Match(line) {
			unitId = UnitId_t(line[6:10])
			debugs("%s: %d: found %q\n", fid, lineNo, unitId)
			location, err := ParseLocationLine(fid, tid, unitId, lineNo, line, debugParser)
			if err != nil {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 10))
				return t, nil
			} else if _, ok := t.UnitMoves[unitId]; ok {
				log.Printf("%s: %s: %d: location %q\n", fid, unitId, lineNo, slug(line, 10))
				return t, fmt.Errorf("duplicate unit in turn")
			}
			moves = &Moves_t{TurnId: t.Id, Id: unitId, FromHex: location.PreviousHex, ToHex: location.CurrentHex}
			t.UnitMoves[moves.Id] = moves
			statusLinePrefix = []byte(fmt.Sprintf("%s Status: ", unitId))
		} else if moves == nil {
			log.Printf("%s: %s: %d: found line outside of section: %q\n", fid, unitId, lineNo, slug(line, 20))
		} else if bytes.HasPrefix(line, []byte("Current Turn ")) {
			debugs("%s: %d: found %q\n", fid, lineNo, slug(line, 19))
			if va, err := Parse(fid, line, Entrypoint("TurnInfo")); err != nil {
				log.Printf("%s: %s: %d: error parsing turn info", fid, unitId, lineNo)
				return t, err
			} else if turnInfo, ok := va.(TurnInfo_t); !ok {
				log.Printf("%s: %s: %d: error parsing turn info", fid, unitId, lineNo)
				log.Printf("error: parser.TurnInfo_t, got %T\n", va)
				log.Printf("please report this error\n")
				panic(fmt.Sprintf("unexpected type %T", va))
			} else {
				if t.Id == "" {
					t.Year, t.Month = turnInfo.CurrentTurn.Year, turnInfo.CurrentTurn.Month
					t.Id = fmt.Sprintf("%04d-%02d", t.Year, t.Month)
				}
				if turnInfo.CurrentTurn.Year != t.Year || turnInfo.CurrentTurn.Month != t.Month {
					log.Printf("%s: %s: %d: current turn: %04d-%02d", fid, unitId, lineNo, t.Year, t.Month)
					log.Printf("%s: %s: %d:    unit turn: %04d-%02d", fid, unitId, lineNo, turnInfo.CurrentTurn.Year, turnInfo.CurrentTurn.Month)
					return t, fmt.Errorf("turn mismatch in report")
				}
			}
		} else if rxFleetMovement.Match(line) {
			pfx, _, ok := bytes.Cut(line, []byte{':'})
			if !ok {
				pfx = []byte(slug(line, 23))
			}
			debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, pfx)
			unitMoves, err := ParseFleetMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return t, err
			}
			if len(unitMoves) > 0 {
				moves.Moves = append(moves.Moves, unitMoves...)
			}
		} else if bytes.HasPrefix(line, []byte("Tribe Follows ")) {
			debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, slug(line, 13))
			if moves.Follows != "" {
				log.Printf("error: %s: %s: %d: found multiple follows\n", fid, unitId, lineNo)
				return t, fmt.Errorf("multiple follows")
			}
			followMove, err := ParseTribeFollowsLine(fid, tid, unitId, lineNo, line, false)
			if err != nil {
				return t, err
			}
			moves.Follows = followMove.Follows
			moves.Moves = append(moves.Moves, followMove)
		} else if bytes.HasPrefix(line, []byte("Tribe Goes to ")) {
			debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, slug(line, 14))
			if moves.GoesTo != "" {
				log.Printf("error: %s: %s: %d: found multiple goes to\n", fid, unitId, lineNo)
				return t, fmt.Errorf("multiple goes to")
			}
			goesToMove, err := ParseTribeGoesToLine(fid, tid, unitId, lineNo, line, false)
			if err != nil {
				return t, err
			}
			moves.GoesTo = goesToMove.GoesTo
			moves.Moves = append(moves.Moves, goesToMove)
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, slug(line, 14))
			unitMoves, err := ParseTribeMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return t, err
			}
			if len(unitMoves) > 0 {
				moves.Moves = append(moves.Moves, unitMoves...)
			}
		} else if rxScoutLine.Match(line) {
			if cfg.Ignore.Scouts {
				if !cfg.Ignore.Logged.Scouts {
					log.Printf("%s: %s: %d: ignoring scouts\n", fid, unitId, lineNo)
					cfg.Ignore.Logged.Scouts = true
				}
			} else {
				debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, slug(line, 14))
				scoutMoves, err := ParseScoutMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
				if err != nil {
					return t, err
				}
				moves.Scouts = append(moves.Scouts, scoutMoves)
			}
		} else if bytes.HasPrefix(line, statusLinePrefix) {
			debugs("%s: %s: %d: found %q\n", fid, unitId, lineNo, statusLinePrefix)
			statusMoves, err := ParseStatusLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
			if err != nil {
				return t, err
			}
			if len(statusMoves) > 0 {
				moves.Moves = append(moves.Moves, statusMoves...)
			}
		}
	}

	// stuff the turn id into all the moves so that sammy can sort them later
	turnId := fmt.Sprintf("%04d-%02d", t.Year, t.Month)
	for _, v := range t.UnitMoves {
		v.TurnId = turnId
		for _, move := range v.Moves {
			move.TurnId = turnId
		}
	}

	return t, nil
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
func ParseFleetMovementLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Move_t, error) {
	if va, err := Parse(fid, line, Entrypoint("FleetMovement")); err != nil {
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		line = mt.Text
	}
	if debugSteps {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, slug(line, 44))
	}

	// remove the prefix and trim the line
	if !bytes.HasPrefix(line, []byte{'M', 'o', 'v', 'e'}) {
		return nil, fmt.Errorf("expected 'Move', found '%s'", slug(line, 12))
	}
	line = bytes.TrimPrefix(line, []byte{'M', 'o', 'v', 'e'})

	return parseMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
}

func ParseLocationLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debug bool) (Location_t, error) {
	if va, err := Parse(fid, line, Entrypoint("Location")); err != nil {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, slug(line, 14))
		return Location_t{}, err
	} else if location, ok := va.(Location_t); !ok {
		log.Printf("%s: %s: %d: location: %q\n", fid, unitId, lineNo, slug(line, 15))
		log.Printf("error: invalid type\n")
		log.Printf("please report this error")
		panic(fmt.Errorf("want Location_t, got %T", va))
	} else {
		return location, nil
	}
}

func ParseScoutMovementLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) (*Scout_t, error) {
	scout := &Scout_t{
		TurnId: tid,
		LineNo: lineNo,
		Line:   bdup(line),
	}

	if va, err := Parse(fid, line, Entrypoint("ScoutMovement")); err != nil {
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		scout.No = mt.ScoutNo
		line = mt.Text
	}
	if debugSteps {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
	}

	// remove the prefix and trim the line
	if !bytes.HasPrefix(line, []byte{'S', 'c', 'o', 'u', 't'}) {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		return nil, fmt.Errorf("expected 'Scout', found '%s'", slug(line, 8))
	}
	line = bytes.TrimPrefix(line, []byte{'S', 'c', 'o', 'u', 't'})

	// parse the moves and then update each with the turn we did the scouting in
	moves, err := parseMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
	if err != nil {
		return nil, err
	}
	for _, move := range moves {
		move.Report.TurnId = tid
		move.Report.ScoutedTurnId = tid
	}
	scout.Moves = moves

	return scout, nil
}

func ParseStatusLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Move_t, error) {
	if va, err := Parse(fid, line, Entrypoint("StatusLine")); err != nil {
		log.Printf("status %v\n", err)
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		line = mt.Text
	}
	if debugSteps {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
	}

	return parseMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
}

func ParseTribeFollowsLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debug bool) (*Move_t, error) {
	var follows UnitId_t
	if va, err := Parse(fid, line, Entrypoint("TribeFollows")); err != nil {
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		follows = mt.Follows
	}
	if debug {
		log.Printf("parser: %s: %s: %d: follows %q\n", fid, unitId, lineNo, follows)
	}

	return &Move_t{
		Follows: follows,
		Report:  &Report_t{TurnId: tid},
		LineNo:  lineNo,
		StepNo:  1,
		Line:    bdup(line),
		TurnId:  tid,
	}, nil
}

func ParseTribeGoesToLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debug bool) (*Move_t, error) {
	var goesTo string
	if va, err := Parse(fid, line, Entrypoint("TribeGoesTo")); err != nil {
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		goesTo = mt.GoesTo
	}
	if debug {
		log.Printf("%s: %s: %d: goes to %q\n", fid, unitId, lineNo, goesTo)
	}

	return &Move_t{
		GoesTo: goesTo,
		Report: &Report_t{TurnId: tid},
		LineNo: lineNo,
		StepNo: 1,
		Line:   bdup(line),
		TurnId: tid,
	}, nil
}

func ParseTribeMovementLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Move_t, error) {
	if va, err := Parse(fid, line, Entrypoint("TribeMovement")); err != nil {
		return nil, err
	} else if mt, ok := va.(Movement_t); !ok {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
		log.Printf("error: want Movement_t, got %T\n", va)
		log.Printf("please report this error\n")
		panic(fmt.Errorf("unexpected type %T\n", va))
	} else {
		line = mt.Text
	}
	if debugSteps {
		log.Printf("%s: %s: %d: %q\n", fid, unitId, lineNo, line)
	}

	// remove the prefix
	if !bytes.HasPrefix(line, []byte{'M', 'o', 'v', 'e'}) {
		return nil, fmt.Errorf("expected 'Move', found '%s'", slug(line, 8))
	}
	line = bytes.TrimPrefix(line, []byte{'M', 'o', 'v', 'e'})

	moves, err := parseMovementLine(fid, tid, unitId, lineNo, line, debugSteps, debugNodes)
	if err != nil {
		return nil, err
	}
	for _, move := range moves {
		move.Report.TurnId = tid
	}

	return moves, nil
}

// parseMovementLine parses all the moves on a single line.
// it returns a slice containing the results for each move or an error.
func parseMovementLine(fid, tid string, unitId UnitId_t, lineNo int, line []byte, debugSteps, debugNodes bool) ([]*Move_t, error) {
	var moves []*Move_t

	line = bytes.TrimSpace(line)

	// we've done this over and over. movement results look like step (\ step)*.
	if bytes.Equal(line, []byte{'\\'}) {
		// "Move \" should be treated as a stay in place
		return []*Move_t{
			{LineNo: lineNo, StepNo: 1, Line: []byte{},
				Still: true, Result: results.Succeeded, Report: &Report_t{TurnId: tid}},
		}, nil
	}

	for _, move := range splitMoves(fid, tid, lineNo, line) {
		// move is the current step in the line

		if debugSteps {
			log.Printf("%s: %s: %d: step %d: %q\n", fid, unitId, lineNo, move.StepNo, move.Line)
		}

		// steps mostly look the same. they are the move attempt and any observations of the immediate terrain (the hex the unit is in).
		// if the movement line is a fleet movement, it may contain additional observations for the adjacent hexes and those one hex away.
		// our first task is to split the steps into sections for this hex, the inner ring of hexes and the outer ring.
		var thisHex, innerRing, outerRing []byte
		var ok bool

		// does this hex contain observations of the inner ring?
		thisHex, innerRing, ok = bytes.Cut(move.Line, []byte{'-', '('})
		if ok {
			// it does, so there must be observations of the outer ring, too
			innerRing, outerRing, ok = bytes.Cut(innerRing, []byte{')', '('})
			if !ok {
				log.Printf("%s: %s: %d: step %d: iring %q\n", fid, unitId, lineNo, move.StepNo, innerRing)
				return nil, fmt.Errorf("inner ring contains '-(' but not ')(")
			}
			// outer ring must end with a closing parentheses
			if bytes.IndexByte(outerRing, ')') == -1 {
				log.Printf("%s: %s: %d: step %d: oring %q\n", fid, unitId, lineNo, move.StepNo, outerRing)
				return nil, fmt.Errorf("outer ring missing ')'")
			}
			// outer ring must end with a closing parentheses
			if outerRing[len(outerRing)-1] != ')' {
				log.Printf("%s: %s: %d: step %d: oring %q\n", fid, unitId, lineNo, move.StepNo, outerRing)
				return nil, fmt.Errorf("outer ring contains text after ')'")
			}
			// remove that parentheses to make later processing simpler
			outerRing = outerRing[:len(outerRing)-1]
		}

		// clean up all the observations. note that after we do, any or all of these may be empty.
		thisHex = bytes.TrimSpace(bytes.TrimRight(thisHex, ", \t"))
		innerRing = bytes.TrimSpace(bytes.TrimRight(innerRing, ", \t"))
		outerRing = bytes.TrimSpace(bytes.TrimRight(outerRing, ", \t"))

		// thisHexMove could contain an actual move command and observations, so parse it.
		// this is a hack - if the parse succeeds, we update the move from the loop
		// because that is the move that we're returning.
		if len(thisHex) != 0 {
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: dirt %q\n", fid, unitId, lineNo, move.StepNo, slug(thisHex, 44))
			}

			mt, err := parseMove(fid, tid, unitId, move.LineNo, move.StepNo, thisHex, debugSteps, debugNodes)
			if err != nil {
				return nil, err
			}
			move.Advance, move.Still, move.Result, move.Report = mt.Advance, mt.Still, mt.Result, mt.Report
		}

		// if the inner ring is present, parse it. this ring contains observations of the surrounding
		// hexes, so each observation will update the border for this move.
		if len(innerRing) != 0 {
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: deck %q\n", fid, unitId, lineNo, move.StepNo, slug(innerRing, 44))
			}

			for no, obs := range bytes.Split(innerRing, []byte{','}) {
				obs = bytes.TrimSpace(obs)
				if len(obs) == 0 {
					continue
				}
				if va, err := Parse(fid, obs, Entrypoint("DeckObservation")); err != nil {
					log.Printf("%s: %s: %d: step %d: deck %q\n", fid, unitId, lineNo, move.StepNo, slug(innerRing, 44))
					log.Printf("%s: %s: %d: step %d: deck %d: obs %q\n", fid, unitId, lineNo, move.StepNo, no+1, obs)
					return nil, err
				} else if deckObservation, ok := va.(NearHorizon_t); !ok {
					log.Printf("%s: %s: %d: step %d: deck %q\n", fid, unitId, lineNo, move.StepNo, slug(innerRing, 44))
					log.Printf("%s: %s: %d: step %d: deck %d: obs %q\n", fid, unitId, lineNo, move.StepNo, no+1, obs)
					log.Printf("error: want NearHorizon_t, got %T\n", va)
					log.Printf("please report this error\n")
					panic(fmt.Sprintf("unexpected type %T", va))
				} else {
					move.Report.MergeBorders(&Border_t{
						Direction: deckObservation.Point,
						Terrain:   deckObservation.Terrain,
					})
				}
			}
		}

		// if the outer ring is present, parse it.
		// this ring contains observations of the twelve hexes that are one-hex away from the current hex.
		// these should only be "unknown land" and "unknown water" values.
		if len(outerRing) != 0 {
			if debugSteps {
				log.Printf("%s: %s: %d: step %d: crow %q\n", fid, unitId, lineNo, move.StepNo, slug(outerRing, 44))
			}

			for nn, orStep := range bytes.Split(outerRing, []byte{','}) {
				orStep = bytes.TrimSpace(orStep)
				if len(orStep) == 0 {
					continue
				}
				crowNo := nn + 1
				if va, err := Parse(fid, orStep, Entrypoint("CrowsNestObservation")); err != nil {
					log.Printf("%s: %s: %d: step %d: crow %d: %q\n", fid, unitId, lineNo, move.StepNo, crowNo, orStep)
					return nil, err
				} else if fh, ok := va.(FarHorizon_t); !ok {
					log.Printf("%s: %s: %d: step %d: crow %d: %q\n", fid, unitId, lineNo, move.StepNo, crowNo, orStep)
					log.Printf("error: want FarHorizon_t, got %T", va)
					log.Printf("please report this error")
					panic(fmt.Errorf("unexpected type %T", va))
				} else {
					move.Report.mergeFarHorizons(fh)
				}
			}
		}

		if len(move.Report.Borders) != 0 {
			sort.Slice(move.Report.Borders, func(i, j int) bool {
				a, b := move.Report.Borders[i], move.Report.Borders[j]
				if a.Direction < b.Direction {
					return true
				} else if a.Direction == b.Direction {
					if a.Edge < b.Edge {
						return true
					} else if a.Edge == b.Edge {
						return a.Terrain < b.Terrain
					}
				}
				return false
			})
		}
		if len(move.Report.Encounters) != 0 {
			sort.Slice(move.Report.Encounters, func(i, j int) bool {
				a, b := move.Report.Encounters[i], move.Report.Encounters[j]
				if a.TurnId < b.TurnId {
					return a.UnitId < b.UnitId
				}
				return false
			})
		}
		if len(move.Report.FarHorizons) != 0 {
			sort.Slice(move.Report.FarHorizons, func(i, j int) bool {
				a, b := move.Report.FarHorizons[i], move.Report.FarHorizons[j]
				if a.Point < b.Point {
					return true
				} else if a.Point == b.Point {
					return a.Terrain < b.Terrain
				}
				return false
			})
		}
		if len(move.Report.Resources) != 0 {
			sort.Slice(move.Report.Resources, func(i, j int) bool {
				return move.Report.Resources[i] < move.Report.Resources[j]
			})
		}

		moves = append(moves, move)
	}

	return moves, nil
}

// parseMove parses a single step of a move, returning the results or an error
func parseMove(fid, tid string, unitId UnitId_t, lineNo, stepNo int, line []byte, debugSteps, debugNodes bool) (*Move_t, error) {
	line = bytes.TrimSpace(bytes.TrimRight(line, ","))
	if debugSteps {
		log.Printf("%s: %s: %d: step %d: %q\n", fid, unitId, lineNo, stepNo, line)
	}

	m := &Move_t{LineNo: lineNo, StepNo: stepNo, Line: line, Report: &Report_t{TurnId: tid}}

	// each move should find at most one settlement
	var settlement *Settlement_t

	root := hexReportToNodes(line, debugNodes)
	steps, err := nodesToSteps(root)
	if err != nil {
		log.Printf("parser: %s: %s: %d: step %d: %q\n", fid, unitId, lineNo, stepNo, line)
		return nil, err
	}

	// parse and report on each step of this move separately.
	for n, subStep := range steps {
		subStepNo := n + 1
		if debugSteps {
			log.Printf("parser: %s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
		}

		var obj any
		if obj, err = Parse("step", subStep, Entrypoint("Step")); err != nil {
			// hack - an unrecognized step might be a settlement name
			if settlement == nil {
				// if it is the first thing after the direction-terrain code
				if m.Result != results.Unknown {
					if r, _ := utf8.DecodeRune(subStep); unicode.IsUpper(r) {
						obj, err = &Settlement_t{Name: string(subStep)}, nil
					}
				}
			}
			if err != nil {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				log.Printf("error: %v\n", err)
				return nil, fmt.Errorf("error parsing step")
			}
		}
		switch v := obj.(type) {
		case *BlockedByEdge_t:
			if m.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("blocked by must start sub-step")
			}
			m.Advance = v.Direction
			m.Result = results.Failed
			m.Report.MergeBorders(&Border_t{
				Direction: v.Direction,
				Edge:      v.Edge,
			})
		case DirectionTerrain_t:
			if m.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("multiple direction-terrain forbidden")
			}
			m.Advance = v.Direction
			m.Result = results.Succeeded
			m.Report.Terrain = v.Terrain
		case []*Edge_t:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("edges forbidden at beginning of step")
			}
			for _, edge := range v {
				m.Report.MergeBorders(&Border_t{
					Direction: edge.Direction,
					Edge:      edge.Edge,
				})
			}
		case *Exhausted_t:
			if m.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("exhaustion must start step")
			}
			m.Advance = v.Direction
			m.Result = results.Failed
			m.Report.MergeBorders(&Border_t{
				Direction: v.Direction,
				Terrain:   v.Terrain,
			})
		case FoundNothing_t: // ignore
		case FoundUnit_t:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return m, fmt.Errorf("units forbidden at beginning of step")
			}
			m.Report.MergeEncounters(&Encounter_t{TurnId: tid, UnitId: v.Id})
		case []FoundUnit_t:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("units forbidden at beginning of step")
			}
			for _, unit := range v {
				m.Report.MergeEncounters(&Encounter_t{TurnId: tid, UnitId: unit.Id})
			}
		case Longhouse_t: // ignore
		case MissingEdge_t:
			m.Result, m.Still, m.Advance = results.Failed, true, v.Direction
		case []*Neighbor_t:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("neighbors forbidden at beginning of step")
			}
			for _, neighbor := range v {
				m.Report.MergeBorders(&Border_t{
					Direction: neighbor.Direction,
					Terrain:   neighbor.Terrain,
				})
			}
		case *ProhibitedFrom_t:
			if m.Result != results.Unknown { // only allowed at the beginning of the step
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("prohibition must start step")
			}
			m.Advance = v.Direction
			m.Result = results.Failed
			m.Report.MergeBorders(&Border_t{
				Direction: v.Direction,
				Terrain:   v.Terrain,
			})
		case resources.Resource_e:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("resources forbidden at beginning of step")
			}
			m.Report.MergeResources(v)
		case *Settlement_t:
			if m.Result == results.Unknown {
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("settlement forbidden at beginning of step")
			}
			m.Report.MergeSettlements(v)
		case terrain.Terrain_e:
			if m.Result != results.Unknown { // valid only at the beginning of the step for status line
				log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
				return nil, fmt.Errorf("terrain must start status")
			}
			m.Result, m.Still = results.Succeeded, true
			m.Report.Terrain = v
		default:
			log.Printf("%s: %s: %d: step %d: sub %d: %q\n", fid, unitId, lineNo, stepNo, subStepNo, subStep)
			log.Printf("error: unexpected type %T\n", v)
			log.Printf("please report this error\n")
			panic(fmt.Sprintf("unexpected %T", v))
		}
	}

	return m, nil
}

// splitMoves splits the line into individual moves. moves are separated by backslashes.
// leading and trailing spaces and any trailing commas are from each move.
func splitMoves(fid, tid string, lineNo int, line []byte) (moves []*Move_t) {
	line = bytes.TrimSpace(bytes.TrimRight(line, " \t\\,"))
	if len(line) == 0 {
		return nil
	}
	for n, text := range bytes.Split(line, []byte{'\\'}) {
		text = bytes.TrimSpace(bytes.TrimRight(text, ", \t"))
		moves = append(moves, &Move_t{LineNo: lineNo, StepNo: n + 1, Line: bdup(text), Report: &Report_t{TurnId: tid}})
	}
	return moves
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

type NoGroupsFound_t struct{}
