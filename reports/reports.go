// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/lbmoves"
	"log"
	"regexp"
	"time"

	xscts "github.com/mdhender/ottomap/pkg/sections"

	pfollows "github.com/mdhender/ottomap/parsers/follows"
	ploc "github.com/mdhender/ottomap/parsers/locations"
	pmoves "github.com/mdhender/ottomap/parsers/movements"
	pscouts "github.com/mdhender/ottomap/parsers/scouts"
	pstatus "github.com/mdhender/ottomap/parsers/status"
	pturn "github.com/mdhender/ottomap/parsers/turns"
)

var (
	rxScoutLine *regexp.Regexp
	rxWinds     = regexp.MustCompile(`^(CALM|MILD|STRONG|GALE)\s+(NE|SE|SW|NW|N|S)\s`)
)

func init() {
	rxScoutLine = regexp.MustCompile(`^Scout [12345678]:Scout `)
}

type Reports []*Report

// Len implements the sort.Interface interface.
func (r Reports) Len() int {
	return len(r)
}

// Less implements the sort.Interface interface.
func (r Reports) Less(i, j int) bool {
	return r[i].Id < r[j].Id
}

// Swap implements the sort.Interface interface.
func (r Reports) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Reports) Contains(id string) bool {
	for _, rpt := range r {
		if rpt.Id == id {
			return true
		}
	}
	return false
}

// Report is a single turn report file that we want to load.
type Report struct {
	Id          string     `json:"id,omitempty"`          // unique identifier for the report file
	Path        string     `json:"path,omitempty"`        // path to the report file
	TurnId      string     `json:"turnId,omitempty"`      // turn ID of the report`
	Year        int        `json:"year,omitempty"`        // year of the report
	Month       int        `json:"month,omitempty"`       // month of the report
	Clan        string     `json:"clan,omitempty"`        // identity of clan from the report
	Ignore      bool       `json:"ignore,omitempty"`      // ignore this report file
	Parsed      string     `json:"parsed,omitempty"`      // path to the parsed report data
	Loaded      *time.Time `json:"loaded,omitempty"`      // time the report was loaded
	Fingerprint string     `json:"fingerprint,omitempty"` // hash of the report file to detect changes
	Sections    []*Section `json:"-"`                     // sections of the report file
}

func (r *Report) Parse() ([]*lbmoves.MovementResults, error) {
	panic("!obsolete!")
	//var mrls []*lbmoves.MovementResults
	//for _, section := range r.Sections {
	//	ums, err := r.parseSection(section)
	//	if err != nil {
	//		return nil, err
	//	}
	//	moves = append(moves, ums...)
	//}
	//return mrls, nil
}

func (r *Report) parseSection(section *Section) ([]*Move, error) {
	// parse the location so that we can get the unit id.
	// that id is needed to extract the status line.
	var ul *ploc.Location
	var ok bool
	if v, err := ploc.Parse("location", section.Location.Text); err != nil {
		log.Printf("parse: report %s: parsing error\n", r.Id)
		log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
		log.Printf("parse: report %s: section %s: line %d: parsing error\n", r.Id, section.Id, section.Location.No)
		log.Printf("parse: input: %q\n", string(section.Location.Text))
		log.Fatalf("parse: error: %v\n", err)
	} else if ul, ok = v.(*ploc.Location); !ok {
		panic(fmt.Sprintf("expected *locations.Location, got %T", v))
	}
	log.Printf("parse: report %s: unit %s (%s %s)\n", r.Id, ul.UnitId, ul.PrevCoords, ul.CurrCoords)

	var ti *pturn.TurnInfo
	if v, err := pturn.Parse("turnInfo", section.TurnInfo.Text); err != nil {
		log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
		log.Printf("parse: report %s: section %s: line %d: parsing error\n", r.Id, section.Id, section.TurnInfo.No)
		log.Printf("parse: input: %q\n", string(section.TurnInfo.Text))
		log.Fatalf("parse: error: %v\n", err)
	} else if ti, ok = v.(*pturn.TurnInfo); !ok {
		panic(fmt.Sprintf("expected *turns.TurnInfo, got %T", v))
	}
	log.Printf("parse: report %s: unit %s: turn %04d-%02d\n", r.Id, ul.UnitId, ti.TurnDate.Year, ti.TurnDate.Month)

	var moves []*Move

	// parse the unit's movement
	if section.Follows == nil && section.Moves == nil { // unit didn't move this turn
		log.Printf("parse: report %s: unit %s: stayed in place\n", r.Id, ul.UnitId)
	}
	if section.Follows != nil { // unit followed another unit this turn
		var ums []*pfollows.Move
		log.Printf("parse: report %s: section %s: line %d: unit %q: todo: parse follows\n", r.Id, section.Id, section.Follows.No, ul.UnitId)
		if v, err := pfollows.Parse("follows", section.Follows.Text); err != nil {
			log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
			log.Printf("parse: report %s: section %s: line %d: unit %q: parsing error\n", r.Id, section.Id, section.Follows.No, ul.UnitId)
			log.Printf("parse: input: %q\n", string(section.Follows.Text))
			log.Fatalf("parse: error: %v\n", err)
		} else if ums, ok = v.([]*pfollows.Move); !ok {
			panic(fmt.Sprintf("expected []*follows.Move, got %T", v))
		} else if len(ums) != 1 {
			log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
			log.Printf("parse: report %s: section %s: line %d: unit %q: parsing error\n", r.Id, section.Id, section.Follows.No, ul.UnitId)
			log.Printf("parse: input: %q\n", string(section.Follows.Text))
			log.Fatalf("parse: error: %v\n", cerrs.ErrUnexpectedNumberOfMoves)
		}
		log.Printf("parse: report %s: unit %s: followed %q\n", r.Id, ul.UnitId, ums[0].Follows)
		moves = append(moves, &Move{
			TurnId:  r.TurnId,
			UnitId:  ul.UnitId,
			Follows: ums[0].Follows,
		})
	}
	if len(section.Moves) != 0 { // unit moved this turn
		for n, mm := range section.Moves {
			log.Printf("parse: report %s: unit %s: move %d: %q\n", r.Id, ul.UnitId, n+1, mm.Text)
		}
		for _, mm := range section.Moves {
			log.Printf("parse: todo: parse movement %q\n", mm.Text)
			v, err := pmoves.Parse("movement", mm.Text)
			if err != nil {
				log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
				log.Printf("parse: input: %q\n", mm.Text)
				log.Fatalf("parse: error: %v\n", err)
			}
			switch t := v.(type) {
			case pmoves.Step:
				mv := &Move{
					TurnId: r.TurnId,
					UnitId: ul.UnitId,
					Step: Step{
						Direction: t.Direction,
						Hex: Hex{
							Terrain: t.Hex.Terrain,
						},
					},
				}
				switch t.Result {
				case pmoves.StayedInPlace:
					mv.Step.Result = StayedInPlace
				case pmoves.Blocked:
					mv.Step.Result = Blocked
				case pmoves.ExhaustedMovementPoints:
					mv.Step.Result = ExhaustedMovementPoints
				case pmoves.Succeeded:
					mv.Step.Result = Succeeded
				}
				for _, e := range t.Hex.Edges {
					mv.Step.Hex.Edges = append(mv.Step.Hex.Edges, &Edge{
						Direction: e.Direction,
						Edge:      e.Edge,
					})
				}
				for _, n := range t.Hex.Neighbors {
					mv.Step.Hex.Neighbors = append(mv.Step.Hex.Neighbors, &Neighbor{
						Direction: n.Direction,
						Terrain:   n.Terrain,
					})
				}
				moves = append(moves, mv)
			default:
				panic(fmt.Sprintf("unexpected %T", v))
			}
		}
	}

	// parse the unit's scouts
	for _, scouts := range section.Scout {
		log.Printf("parse: report %s: section %s: line %d: scout %q\n", r.Id, section.Id, scouts.Line.No, scouts.Line.Text)
		for _, scout := range scouts.Moves {
			log.Printf("parse: report %s: section %s: line %d: scout move %q\n", r.Id, section.Id, scout.No, scout.Text)
			v, err := pscouts.Parse("scout", scout.Text)
			if err != nil {
				log.Printf("parse: report %s: section %s: line %d: unit %s: parsing error\n", r.Id, section.Id, scouts.Line.No, ul.UnitId)
				log.Printf("parse: report %s: section %s: line %d: scout %d: move %q\n", r.Id, section.Id, scouts.Line.No, scouts.ScoutId, scout.Text)
				log.Printf("parse: input: %q\n", scout.Text)
				log.Fatalf("parse: error: %v\n", err)
			}
			log.Printf("parse: scout: returned %T", v)

			switch t := v.(type) {
			case pscouts.Step:
				mv := &Move{
					TurnId: r.TurnId,
					UnitId: ul.UnitId,
					Step: Step{
						Direction: t.Direction,
						Hex: Hex{
							Terrain: t.Hex.Terrain,
						},
					},
				}
				switch t.Result {
				case pscouts.StayedInPlace:
					mv.Step.Result = StayedInPlace
				case pscouts.Blocked:
					mv.Step.Result = Blocked
				case pscouts.ExhaustedMovementPoints:
					mv.Step.Result = ExhaustedMovementPoints
				case pscouts.Succeeded:
					mv.Step.Result = Succeeded
				}
				for _, e := range t.Hex.Edges {
					mv.Step.Hex.Edges = append(mv.Step.Hex.Edges, &Edge{
						Direction: e.Direction,
						Edge:      e.Edge,
					})
				}
				for _, n := range t.Hex.Neighbors {
					mv.Step.Hex.Neighbors = append(mv.Step.Hex.Neighbors, &Neighbor{
						Direction: n.Direction,
						Terrain:   n.Terrain,
					})
				}
				moves = append(moves, mv)
			default:
				panic(fmt.Sprintf("unexpected %T", v))
			}

		}
	}

	// parse the unit's status line into the final tile
	var sh *pstatus.Hex
	log.Printf("parse: report %s: section %s: line %d: status %q\n", r.Id, section.Id, section.Status.No, section.Status.Text)
	if v, err := pstatus.Parse("status", section.Status.Text); err != nil {
		log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
		log.Printf("parse: report %s: section %s: line %d: unit %q: parsing error\n", r.Id, section.Id, section.Status.No, ul.UnitId)
		log.Printf("parse: input: %q\n", string(section.Status.Text))
		log.Fatalf("parse: error: %v\n", err)
	} else if sh, ok = v.(*pstatus.Hex); !ok {
		panic(fmt.Sprintf("expected *status.Hex, got %T", v))
	}
	log.Printf("parse: section: status: terrain %s\n", sh.Terrain)
	if sh.Resource != domain.RNone {
		log.Printf("parse: section: status: resource %s\n", sh.Resource)
	}
	if len(sh.Settlements) == 1 {
		log.Printf("parse: section: status: settlement %q\n", sh.Settlements[0].Name)
	} else if len(sh.Settlements) > 1 {
		log.Printf("parse: report %s: section %s: parsing error\n", r.Id, section.Id)
		log.Printf("parse: report %s: section %s: line %d: unit %q: parsing error\n", r.Id, section.Id, section.Status.No, ul.UnitId)
		log.Printf("parse: input: %q\n", string(section.Status.Text))
		for _, s := range sh.Settlements {
			log.Printf("parse: input: settlement %q\n", s.Name)
		}
		log.Fatalf("parse: error: %v\n", fmt.Errorf("multiple settlements"))
	}
	for _, f := range sh.Found {
		if f.Edge != nil {
			log.Printf("parse: section: status: edge %+v\n", *f.Edge)
		}
		if f.UnitId != "" {
			log.Printf("parse: section: status: unit %s\n", f.UnitId)
		}
	}

	if len(moves) == 0 {
		if sh == nil {
			panic("assert(sh != nil")
		}

		// the unit didn't move this turn so use the unit's current location
		moves = append(moves, &Move{
			TurnId: r.TurnId,
			UnitId: ul.UnitId,
			Step: Step{
				Direction: directions.DUnknown,
				Result:    StayedInPlace,
				Hex: Hex{
					Terrain: sh.Terrain,
				},
			},
		})
	}

	//// extract the found items
	//if unit.Status == nil {
	//	unit.Status = &Found{}
	//}
	//unit.Status.Terrain = st.Terrain
	//for _, f := range st.Found {
	//	log.Printf("parse: section: status: found %+v\n", *f)
	//	if f.UnitId != "" {
	//		if f.UnitId != unit.Id { // only add if it isn't the unit we are parsing
	//			unit.Status.Units = append(unit.Status.Units, f.UnitId)
	//		}
	//	}
	//}

	return moves, nil
}

// Section makes parsing easier by splitting the report into the lines
// that make up each section that we want to parse.
type Section struct {
	Id string

	UnitId     string
	PrevCoords string // grid coordinates before the unit moved
	CurrCoords string // grid coordinates after the unit moved

	Location      *Line
	TurnInfo      *Line
	Follows       *Line
	FollowsLine   *Line
	Moves         []*Line
	MovementLine  *Line
	FleetMovement *FleetMovement
	Scout         []*ScoutLine
	ScoutLines    []*Line
	Status        *Line
	StatusLine    *Line
	Error         *Error
}

type Error struct {
	Line  *Line
	Error error
}

type Line struct {
	No   int
	Text []byte
}

type FleetMovement struct {
	LineNo    int
	Winds     domain.WindStrength_e
	Dir       directions.Direction
	MovesText []byte
}

type ScoutLine struct {
	ScoutId int
	Line    *Line
	Moves   []*Line
}

func Sections(id string, input []byte, showSkippedSections bool) ([]*Section, *Error) {
	// check for bom and remove it if present.
	for _, bom := range [][]byte{
		// see https://en.wikipedia.org/wiki/Byte_order_mark for BOM values
		[]byte{0xEF, 0xBB, 0xBF},
	} {
		if bytes.HasPrefix(input, bom) {
			log.Printf("report %s: skipped %8d BOM bytes\n", id, len(bom))
			input = input[len(bom):]
			break
		}
	}

	inputSections, ok := xscts.SplitRegEx(id, input, showSkippedSections)
	if !ok {
		return nil, &Error{
			Line: &Line{
				No:   0,
				Text: []byte(""),
			},
			Error: fmt.Errorf("regex: no sections found"),
		}
	}
	if showSkippedSections {
		log.Printf("report %s: %d sections\n", id, len(inputSections))
	}

	var sections []*Section
	for n, chunk := range inputSections {
		// ignore non-unit sections
		if chunk.Type == domain.UTUnknown {
			if showSkippedSections {
				log.Printf("report %s: section %5d: skipping %q\n", id, n+1, chunk.Slug())
			}
		}

		sections = append(sections, &Section{Id: fmt.Sprintf("%d", n+1)})
		section := sections[len(sections)-1]

		if len(chunk.Lines) == 0 {
			section.Error = &Error{
				Line: &Line{
					No:   0,
					Text: []byte(""),
				},
				Error: cerrs.ErrNotATurnReport,
			}
			continue
		} else if len(chunk.Lines) < 2 || chunk.Id == "" {
			section.Error = &Error{
				Line: &Line{
					No:   chunk.Lines[0].No,
					Text: bdup(chunk.Lines[0].Text),
				},
				Error: cerrs.ErrNotATurnReport,
			}
			continue
		}

		section.Location = &Line{
			No:   chunk.Lines[0].No,
			Text: bdup(chunk.Lines[0].Text),
		}

		// parse the location so that we can get the unit id.
		// that id is needed to extract the status line.
		var ul *ploc.Location
		var ok bool
		if v, err := ploc.Parse("location", section.Location.Text); err != nil {
			log.Printf("report %s: section %s: line %d: parse error\n\t%v\n", id, section.Id, section.Location.No, err)
			section.Error = &Error{
				Line: &Line{
					No:   section.Location.No,
					Text: bdup(section.Location.Text),
				},
				Error: err,
			}
			continue
		} else if ul, ok = v.(*ploc.Location); !ok {
			log.Printf("report %s: section %s: parse error\n", id, section.Id)
			log.Printf("report %s: section %s: line %d: parse error\n", id, section.Id, section.Location.No)
			log.Printf("report %s: section %s: line %d: parse error\n\t%v\n", id, section.Id, section.Location.No, err)
			panic(fmt.Sprintf("expected *locations.Location, got %T", v))
		} else if ul.UnitId != chunk.Id {
			log.Printf("report %s: section %s: line %d: element id %q\n", id, section.Id, section.Location.No, chunk.Id)
			log.Printf("report %s: section %s: line %d:    unit id %q\n", id, section.Id, section.Location.No, ul.UnitId)
			panic("assert(ul.UnitId == chunk.Id)")
		}
		section.UnitId = ul.UnitId
		section.PrevCoords = ul.PrevCoords
		section.CurrCoords = ul.CurrCoords

		// now that we know the unit id, we can extract the remaining lines that we are interested in
		tribeFollowsLine := []byte("Tribe Follows ")
		tribeMovesLine := []byte("Tribe Movement: ")
		var scoutLines [8][]byte
		for sid := 0; sid < 8; sid++ {
			scoutLines[sid] = []byte(fmt.Sprintf("Scout %d:Scout ", sid+1))
		}
		statusLine := []byte(fmt.Sprintf("%s Status: ", ul.UnitId))
		for n, ctext := range chunk.Lines {
			if n == 0 {
				if !ctext.IsLocation {
					log.Printf("report %s: section %s: line %d: input %q\n", id, section.Id, ctext.No, ctext.Slug(20))
					panic("expected location line")
				}
			} else if n == 1 {
				if !ctext.IsTurnInfo {
					log.Printf("report %s: section %s: line %d: input %q\n", id, section.Id, ctext.No, ctext.Slug(20))
					panic("expected turn info line")
				}
				section.TurnInfo = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
			} else if ctext.IsStatus {
				if section.Status != nil {
					section.Error = &Error{
						Line: &Line{
							No:   ctext.No,
							Text: bdup(ctext.Text),
						},
						Error: cerrs.ErrMultipleStatusLines,
					}
					break
				}
				section.Status = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
				section.StatusLine = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
			} else if ctext.MovementType == domain.UMFleet {
				//log.Printf("%s: %d: %d: found fleet movement\n\t%s\n\t\t%s\n", id, chunk.No, ctext.No, chunk.Slug(), ctext.Slug(35))
				if section.FleetMovement != nil {
					section.Error = &Error{
						Line: &Line{
							No:   ctext.No,
							Text: bdup(ctext.Text),
						},
						Error: cerrs.ErrMultipleFleetMovementLines,
					}
					break
				}
				section.FleetMovement = scrubFleetMoves(&Line{
					No:   ctext.No,
					Text: ctext.Text,
				})
			} else if ctext.MovementType == domain.UMFollows {
				if section.Follows != nil {
					section.Error = &Error{
						Line: &Line{
							No:   ctext.No,
							Text: bdup(ctext.Text),
						},
						Error: cerrs.ErrMultipleFollowsLines,
					}
					break
				}
				section.Follows = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
				section.FollowsLine = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
			} else if ctext.MovementType == domain.UMScouts {
				if rxScoutLine.Match(ctext.Text) {
					section.ScoutLines = append(section.ScoutLines, &Line{
						No:   ctext.No,
						Text: bdup(ctext.Text),
					})
				}
				for sid := 0; sid < 8; sid++ {
					if bytes.HasPrefix(ctext.Text, scoutLines[sid]) {
						scoutLine := &ScoutLine{ScoutId: sid + 1, Line: &Line{No: ctext.No, Text: bdup(ctext.Text)}}
						for _, jo := range scrubScouts(scoutLine.Line) {
							if len(jo.Text) != 0 {
								scoutLine.Moves = append(scoutLine.Moves, jo)
							}
						}
						section.Scout = append(section.Scout, scoutLine)
						break
					}
				}
			} else if ctext.MovementType == domain.UMTribe {
				section.MovementLine = &Line{
					No:   ctext.No,
					Text: bdup(ctext.Text),
				}
				// remove the prefix and trim the line
				text := bytes.TrimSpace(bytes.TrimPrefix(ctext.Text, tribeMovesLine))
				if bytes.HasPrefix(text, []byte{'M', 'o', 'v', 'e'}) {
					text = bytes.TrimSpace(bytes.TrimPrefix(ctext.Text, []byte{'M', 'o', 'v', 'e'}))
				}
				section.Moves = scrubLandMoves(&Line{
					No:   ctext.No,
					Text: text,
				})
			} else if bytes.HasPrefix(ctext.Text, tribeFollowsLine) {
				panic("should not find follows line here")
			} else if bytes.HasPrefix(ctext.Text, tribeMovesLine) {
				panic("should not find tribe movement line here")
			} else if bytes.HasPrefix(ctext.Text, []byte{'S', 'c', 'o', 'u', 't'}) {
				panic("should not find scout line here")
			} else if bytes.HasPrefix(ctext.Text, statusLine) {
				panic("should not find status line here")
			} else {
				log.Printf("reports: sections: unexpected line: %q\n", chunk.Slug())
				log.Printf("reports: sections: unexpected line: %q\n", ctext.Slug(35))
				panic("unhandled line?")
			}
		}
		if section.Error != nil {
			continue
		}

		// consistency checks
		if section.Follows != nil && section.Moves != nil {
			section.Error = &Error{
				Line: &Line{
					No:   section.Follows.No,
					Text: bdup(section.Follows.Text),
				},
				Error: cerrs.ErrUnitMovesAndFollows,
			}
			continue
		} else if len(section.Scout) > 8 {
			section.Error = &Error{
				Line: &Line{
					No:   section.Scout[0].Line.No,
					Text: bdup(section.Scout[0].Line.Text),
				},
				Error: cerrs.ErrTooManyScoutLines,
			}
			continue
		} else if section.Status == nil {
			section.Error = &Error{
				Line: &Line{
					No:   section.Location.No,
					Text: bdup(section.Location.Text),
				},
				Error: cerrs.ErrMissingStatusLine,
			}
			continue
		}
	}

	// return the first error we find, if any
	for _, section := range sections {
		if section.Error != nil {
			return sections, section.Error
		}
	}

	return sections, nil
}

type Move struct {
	TurnId  string // turn id this move belongs to
	UnitId  string // unit id this move belongs to
	Follows string // unit id this unit follows
	Step    Step
}

type Step struct {
	// direction will be Unknown when unit doesn't try to move
	Direction directions.Direction
	Result    Result
	// Hex is the hex where the unit ended up. It could be the same
	// as where it started if the step failed
	Hex Hex
}

type Result int

const (
	StayedInPlace Result = iota
	Succeeded
	Blocked
	ExhaustedMovementPoints
	Followed
)

type Hex struct {
	Terrain     domain.Terrain
	Resource    domain.Resource
	Edges       []*Edge
	Neighbors   []*Neighbor
	Resources   []domain.Resource
	Settlements []*Settlement
	Occupants   []string
}

type Edge struct {
	Direction directions.Direction
	Edge      domain.Edge
}

type BlockedBy struct {
	Direction directions.Direction
	Terrain   domain.Terrain
}

type Exhausted struct {
	Direction directions.Direction
	Terrain   domain.Terrain
}

type Neighbor struct {
	Direction directions.Direction
	Terrain   domain.Terrain
}

type Settlement struct {
	Name string
}

func bdup(b []byte) []byte {
	return append([]byte{}, b...)
}

func isUnitSection(chunk []byte) bool {
	for _, tag := range []string{"Courier ", "Element ", "Fleet ", "Garrison ", "Tribe "} {
		if bytes.HasPrefix(chunk, []byte(tag)) {
			return true
		}
	}
	return false
}

// scrubLandMoves splits the line into individual moves and then removes
// leading and trailing spaces and any trailing commas from the move.
func scrubLandMoves(line *Line) []*Line {
	var moves []*Line
	for _, move := range bytes.Split(line.Text, []byte{'\\'}) {
		move = bytes.TrimSpace(bytes.TrimRight(move, ", \t"))
		if len(move) != 0 {
			moves = append(moves, &Line{
				No:   line.No,
				Text: bdup(move),
			})
		}
	}
	return moves
}

func scrubScouts(line *Line) []*Line {
	// trim the scout number and then split the rest of the line
	line = &Line{
		No:   line.No,
		Text: line.Text[len("Scout 8:Scout "):],
	}
	return scrubLandMoves(line)
}

func scrubFleetMoves(line *Line) *FleetMovement {
	matches := rxWinds.FindStringSubmatch(string(line.Text))
	if len(matches) != 3 {
		panic(fmt.Sprintf("expected 3 matches, got %d", len(matches)))
	}
	// strip the winds and direction from the text
	moves := bytes.TrimSpace(line.Text[len(matches[0]):])
	// todo: split and scrub the moves
	return &FleetMovement{
		LineNo:    line.No,
		Winds:     domain.WindStrengthStringToEnum[matches[1]],
		Dir:       directions.DirectionStringToEnum[matches[2]],
		MovesText: bdup(moves),
	}
}

func slug(lines []byte) string {
	for _, line := range bytes.Split(lines, []byte{'\n'}) {
		if len(line) < 73 {
			return string(line)
		}
		return string(line[:73])
	}
	return ""
}

// split splits the input into sections. It returns the sections along
// with the section separator. We trim leading and trailing new-lines
// from each section and then force the section to end with a new-line.
//
// We check for a few types of separators and use the
// first one that we find. If we can't find a separator,
// we return the entire input as the first value and nil
// for the separator.
//
// NB: The first turn report (the "setup" turn) might have just
// one section, so we wouldn't find a section separator. The
// instructions should tell the user to manually add one. Or the
// caller should have logic to handle.
func split(id string, input []byte) ([][]byte, []byte) {
	// scan the input to find the section separator
	var separator []byte
	for _, pattern := range [][]byte{
		[]byte{0xE2, 0x80, 0x83},                         // MS Word section break
		[]byte{0x0a, 0x2f, 0x2f, 0x2d, 0x2d, 0x2d, 0x2d}, // \n//----
		[]byte{'\f'}, // simple form feed
	} {
		if bytes.Index(input, pattern) == -1 {
			continue
		}
		separator = pattern
		break
	}

	// split the input
	var sections [][]byte
	if separator == nil {
		sections = [][]byte{input}
	} else {
		sections = bytes.Split(input, separator)
	}

	// our parsers expect sections to not start or end with blank lines.
	// they also require that the last line end with a new-line.
	for i, section := range sections {
		section = bytes.TrimRight(bytes.TrimLeft(section, "\n"), "\n")
		section = append(section, '\n')
		sections[i] = section

		if len(section) < 35 {
			log.Printf("report %s: section %3d: %q\n", id, i+1, section)
		} else {
			log.Printf("report %s: section %3d: %q\n", id, i+1, section[:35])
		}
	}

	return sections, separator
}
