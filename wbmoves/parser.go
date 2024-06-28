// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wbmoves

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/hexes"
	"github.com/mdhender/ottomap/internal/winds"
	"log"
	"regexp"
)

//go:generate pigeon -o grammar.go grammar.peg

func ParseUnitLocation(line *Line_t) error {
	return fmt.Errorf("!implemented")
}
func ParseTurnReport(lines []*Line_t, debugSteps, debugNodes bool) (*Results_t, error) {
	if len(lines) == 0 {
		return nil, cerrs.ErrNotATurnReport
	}

	// first line must be the clan location
	return nil, fmt.Errorf("!implemented")
}

// ParseFleetMovement parses a fleet movement line.
//
// The input starts with the Winds, the phrase "Fleet Movement:" and then the Steps.
// The first Step starts with "Move"; the remainder start with a backslash.
// Steps are complicated but can be thought of as Direction "-" Terrain LandObservations "-(" Direction DeckObservations ")(" CrowsNestObservations ")"
func ParseFleetMovement(unitId string, previousHex hexes.Hex_t, lineNo int, line []byte) (*Results_t, error) {
	fm := &Results_t{
		Id:          unitId,
		LineNo:      lineNo,
		PreviousHex: previousHex,
		Text:        bdup(line),
	}
	if !IsFleetMovementLine(line) {
		fm.Error = cerrs.ErrNotAFleetMovementLine
		return fm, fm.Error
	}

	// first word is the wind strength
	word, rest, ok := bytes.Cut(line, []byte{' '})
	if !ok {
		panic("assert(cut == ok)")
	}
	fm.Winds.Strength, ok = winds.StringToEnum[string(word)]
	if !ok {
		panic(fmt.Sprintf("assert(%q is valid)", word))
	}
	rest = bytes.TrimSpace(rest)

	// next word is the direction
	word, rest, ok = bytes.Cut(rest, []byte{' '})
	fm.Winds.From, ok = direction.StringToEnum[(string(word))]
	if !ok {
		panic(fmt.Sprintf("assert(%q is valid)", word))
	}
	rest = bytes.TrimSpace(rest)

	// next word is the phrase "Fleet Movement:"
	if !bytes.HasPrefix(rest, []byte("Fleet Movement:")) {
		panic("assert(starts with \"Fleet Movement:\")")
	}
	rest = bytes.TrimSpace(rest)

	// the first step must start with "Move"
	if !bytes.HasPrefix(rest, []byte("Move")) {
		fm.Error = fmt.Errorf("expected 'Move' to start first step")
		return fm, fm.Error
	}
	rest = bytes.TrimPrefix(rest, []byte("Move"))

	// the rest of the line is the steps. steps are separated by backslashes, so split them up.
	for _, text := range bytes.Split(rest, []byte{'\\'}) {
		fm.Steps = append(fm.Steps, &Step_t{
			LineNo: lineNo,
			No:     len(fm.Steps) + 1,
			Text:   bdup(bytes.TrimSpace(text)),
		})
	}
	log.Printf("hey, we found %d steps in this line!\n", len(fm.Steps))

	// successful moves start with a direction, dash, and terrain.

	// fleet movements can contain an extra set of observations.

	return fm, fmt.Errorf("not implemented")
}

func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

//// splitSteps splits the line into individual steps.
//// the first step must start with "Move" or an error is returned.
//// the remaining steps are separated by backslashes.
//func splitSteps(lineNo int, line []byte) ([]*Step_t, error) {
//	// the first step must start with "Move"
//	if !bytes.HasPrefix(line, []byte("Move")) {
//		return nil, fmt.Errorf("line %d: step %d: expected 'Move'\n", lineNo, 1)
//	}
//	line = bytes.TrimPrefix(line, []byte("Move"))
//
//	// steps are separated by backslashes, so split them up.
//	var steps []*Step_t
//	for _, text := range bytes.Split(line, []byte{'\\'}) {
//		steps = append(steps, &Step_t{LineNo: lineNo, No: len(steps)+1, Text: bdup(bytes.TrimSpace(text))}
//	}
//	log.Printf("hey, we found %d steps in this line!\n", len(steps))
//
//	//for stepNo := 1; len(line) != 0; stepNo++ {
//	//	var step Step_t
//	//	step.LineNo = lineNo
//	//	step.No = stepNo
//	//	step.Text = bdup(line)
//	//
//	//	// step must start with "Move" or "\\"
//	//	if stepNo == 1 {
//	//		if !bytes.HasPrefix(line, []byte("Move")) {
//	//			step.Error = fmt.Errorf("line %d: step %d: expected 'Move'\n", lineNo, stepNo)
//	//			break
//	//		}
//	//		line = bytes.TrimPrefix(line, []byte("Move"))
//	//	} else {
//	//		if !bytes.HasPrefix(line, []byte{'\\'}) {
//	//			step.Error = fmt.Errorf("line %d: step %d: expected '\\'\n", lineNo, stepNo)
//	//			break
//	//		}
//	//		line = bytes.TrimPrefix(line, []byte{'\\'})
//	//	}
//	//
//	//	// steps can have two types of movement: land or water. They both start with DIRECTION DASH TERRAIN,
//	//	// that's followed by observations and encounters and terminated with a backslash. so let's split
//	//	// the steps out.
//	//	stepText := bytes.Split(line, []byte{'\\'})
//	//	log.Printf("hey, we found %d steps in this line!\n", len(stepText))
//	//	// but water is followed immediately by the "-(...)(...)" stanza. Land can have more interesting observations.
//	//
//	//	// we must split the step into three sections. the first is terminated by "-(", the second by ")(", and the third by a backslash or end of line.
//	//	dashParenIndex := bytes.Index(line, []byte{'-', '('})
//	//	if dashParenIndex == -1 {
//	//		log.Printf("fleet movement line %d: step %d: does not contain \"-(\"\n", lineNo, stepNo)
//	//		return nil, cerrs.ErrNotAFleetMovementLine
//	//	}
//	//	closeOpenParenIndex := bytes.Index(line, []byte{')', '('})
//	//	if closeOpenParenIndex == -1 {
//	//		log.Printf("fleet movement line %d: step %d: does not contain \")(\"\n", lineNo, stepNo)
//	//		return nil, cerrs.ErrNotAFleetMovementLine
//	//	}
//	//
//	//	// the step is terminated by a backslash or end of line
//	//	line = bytes.TrimSpace(line)
//	//}
//
//	return steps, nil
//}

// IsFleetMovementLine returns true if the line is a fleet movement line.
func IsFleetMovementLine(line []byte) bool {
	return rxFleetMove.Match(line)
}

type Line_t struct {
	LineNo int    // original line number in the input
	Text   []byte // copy of the original line
}

// Results_t is the result of parsing the entire fleet movement line.
//
// Individual steps are in the Steps slice, where each step represents a single hex of movement.
//
// NB: when "GOTO" orders are processed, the step may represent a "teleportation" across multiple hexes.
type Results_t struct {
	Id          string      // unit id
	LineNo      int         // line number in the input
	PreviousHex hexes.Hex_t // hex where the unit is at the start of this line
	CurrentHex  hexes.Hex_t // hex where the unit is at the end of this line
	Winds       *Winds_t    // optional winds
	Steps       []*Step_t   // optional steps, one for each hex of movement in the line
	Text        []byte      // copy of the original line
	Warning     string
	Error       error
}

// Step_t is a single step in a movement result. Generally, a step represents a single hex of movement.
// However, a step may represent a "teleportation" across multiple hexes (for example, the "GOTO" command).
//
// Observations are for terrain, edges, neighbors; anything that is "permanent."
// We should report a warning if a new observation conflicts with an existing observation,
// but we leave that to the tile generator.
//
// Encounters are for units, settlements, resources, and "random" encounters.
// These are "temporary" (units can move, settlements can be captured, etc.).
type Step_t struct {
	No           int // step number in this result
	LineNo       int // line number in the input
	Movement     *Movement_t
	Observations []*Observation_t
	Encounters   []*Encounter_t
	Text         []byte // copy of the original step
	Warning      string
	Error        error
}

// Movement_t is the attempted movement of a unit.
//
// Moves can fail, in which case the unit stays where it is.
// Failed moves are not reported as warnings or errors.
type Movement_t struct {
	LineNo     int // line number in the input
	Type       UnitMovement_e
	Direction  direction.Direction_e
	CurrentHex hexes.Hex_t // hex where the unit ends up
	Text       []byte      // copy of the original movement
	Warning    string
	Error      error
}

type Observation_t struct {
	No         int         // index number in this step
	CurrentHex hexes.Hex_t // hex where the observation is taking place
	Direction  []direction.Direction_e
	Edge       *Edge_t
	Neighbor   *Neighbor_t
	Text       []byte // copy of the original observation
	Warning    string
	Error      error
}

type Edge_t struct {
	Direction direction.Direction_e
	Edge      Edge_e
	Text      []byte // copy of the original edge
	Warning   string
	Error     error
}

type Neighbor_t struct {
	Hex       hexes.Hex_t // hex where the neighbor is
	Direction direction.Direction_e
	Terrain   domain.Terrain
	Text      []byte // copy of the original neighbor
	Warning   string
	Error     error
}

type Encounter_t struct {
	No         int // index number in this step
	Element    *Element_t
	Item       *Item_t
	Resource   *Resource_t
	Settlement *Settlement_t
	Text       []byte // copy of the original encounter
	Warning    string
	Error      error
}

type Element_t struct {
	Id      string // unit id
	Text    []byte // copy of the original element
	Warning string
	Error   error
}

type Item_t struct {
	Quantity int
	Item     string
	Text     []byte // copy of the original item
	Warning  string
	Error    error
}

type Resource_t struct {
	Resource domain.Resource
	Text     []byte // copy of the original resource
	Warning  string
	Error    error
}

type Settlement_t struct {
	Name    string
	Text    []byte // copy of the original settlement
	Warning string
	Error   error
}

// Winds_t is the winds that are present in the movement.
// They are optional since they are only on fleet movement lines.
type Winds_t struct {
	Strength winds.Strength_e
	From     direction.Direction_e
	Text     []byte // copy of the original winds
	Warning  string
	Error    error
}

type DeckObservation_t struct {
	No        int // index number in this step's deck observations
	Direction direction.Direction_e
	Terrain   domain.Terrain
	Text      []byte // copy of the original deck observation
	Warning   string
	Error     error
}

type CrowsNestObservation_t struct {
	No        int // index number in this step's crows nest observations
	Sighted   Sighted_e
	Direction []direction.Direction_e
	Text      []byte // copy of the original crows nest observation
	Warning   string
	Error     error
}

var (
	rxFleetMove = regexp.MustCompile(`^(CALM|MILD|STRONG|GALE)\s+(NE|SE|SW|NW|N|S)\s+Fleet Movement: Move\s+`)
)
