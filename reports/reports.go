// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
	"time"

	pfollows "github.com/mdhender/ottomap/parsers/follows"
	ploc "github.com/mdhender/ottomap/parsers/locations"
	pmoves "github.com/mdhender/ottomap/parsers/movements"
	pscouts "github.com/mdhender/ottomap/parsers/scouts"
	pstatus "github.com/mdhender/ottomap/parsers/status"
	pturn "github.com/mdhender/ottomap/parsers/turns"
)

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

func (r *Report) Parse() ([]*Move, error) {
	var moves []*Move
	for _, section := range r.Sections {
		ums, err := r.parseSection(section)
		if err != nil {
			return nil, err
		}
		moves = append(moves, ums...)
	}
	return moves, nil
}

func (r *Report) parseSection(section *Section) ([]*Move, error) {
	// parse the location so that we can get the unit id.
	// that id is needed to extract the status line.
	var ul *ploc.Location
	var ok bool
	if v, err := ploc.Parse("location", section.Location); err != nil {
		log.Printf("parse: report %s: parsing error\n", r.Id)
		log.Printf("parse: input: %q\n", string(section.Location))
		log.Fatalf("parse: error: %v\n", err)
	} else if ul, ok = v.(*ploc.Location); !ok {
		panic(fmt.Sprintf("expected *locations.Location, got %T", v))
	}
	log.Printf("parse: report %s: unit %s (%s %s)\n", r.Id, ul.UnitId, ul.PrevCoords, ul.CurrCoords)

	var ti *pturn.TurnInfo
	if v, err := pturn.Parse("turnInfo", section.TurnInfo); err != nil {
		log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
		log.Printf("parse: input: %q\n", string(section.TurnInfo))
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
		log.Printf("parse: todo: parse follows  %q\n", string(section.Follows))
		if v, err := pfollows.Parse("follows", section.Follows); err != nil {
			log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
			log.Printf("parse: input: %q\n", string(section.Follows))
			log.Fatalf("parse: error: %v\n", err)
		} else if ums, ok = v.([]*pfollows.Move); !ok {
			panic(fmt.Sprintf("expected []*follows.Move, got %T", v))
		} else if len(ums) != 1 {
			log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
			log.Printf("parse: input: %q\n", string(section.Follows))
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
			log.Printf("parse: report %s: unit %s: move %d: %q\n", r.Id, ul.UnitId, n+1, string(mm))
		}
		for _, mm := range section.Moves {
			log.Printf("parse: todo: parse movement %q\n", string(mm))
			v, err := pmoves.Parse("movement", mm)
			if err != nil {
				log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
				log.Printf("parse: input: %q\n", string(mm))
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
	for _, scoutSteps := range section.Scout {
		for _, scout := range scoutSteps {
			log.Printf("parse: section: scout %q\n", string(scout))
			if v, err := pscouts.Parse("scout", scout); err != nil {
				log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
				log.Printf("parse: input: %q\n", string(scout))
				log.Fatalf("parse: error: %v\n", err)
			} else {
				log.Printf("parse: scout: returned %T", v)
			}
		}
	}

	// parse the unit's status line into the final tile
	var sh *pstatus.Hex
	log.Printf("parse: section: status %q\n", string(section.Status))
	if v, err := pstatus.Parse("status", section.Status); err != nil {
		log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
		log.Printf("parse: input: %q\n", string(section.Status))
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
		log.Printf("parse: report %s: unit %s: parsing error\n", r.Id, ul.UnitId)
		log.Printf("parse: input: %q\n", string(section.Status))
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
	Id       string
	Location []byte
	TurnInfo []byte
	Follows  []byte
	Moves    [][]byte
	Scout    [][][]byte
	Status   []byte
	Error    error
}

func Sections(input []byte, showSkippedSections bool) ([]*Section, error) {
	var sections []*Section
	chunks, _ := split(input)

	for n, chunk := range chunks {
		// ignore non-unit sections
		if !isUnitSection(chunk) {
			if showSkippedSections {
				log.Printf("reports: sections: skipping %q\n", slug(chunk))
			}
			continue
		}

		sections = append(sections, &Section{Id: fmt.Sprintf("%d", n+1)})
		section := sections[len(sections)-1]

		lines := bytes.Split(chunk, []byte("\n"))

		if len(lines) < 2 {
			section.Error = cerrs.ErrNotATurnReport
			continue
		}

		section.Location = bdup(lines[0])

		// parse the location so that we can get the unit id.
		// that id is needed to extract the status line.
		var ul *ploc.Location
		var ok bool
		if v, err := ploc.Parse("location", section.Location); err != nil {
			log.Printf("reports: sections: location: %q\n", string(section.Location))
			log.Printf("reports: sections: location: %v\n", err)
			section.Error = err
			continue
		} else if ul, ok = v.(*ploc.Location); !ok {
			log.Printf("reports: sections: location: %q\n", string(section.Location))
			panic(fmt.Sprintf("expected *locations.Location, got %T", v))
		}

		// now that we know the unit id, we can extract the remaining lines that we are interested in
		followsLine := []byte("Tribe Follows ")
		movesLine := []byte("Tribe Movement: ")
		var scoutLines [8][]byte
		for sid := 0; sid < 8; sid++ {
			scoutLines[sid] = []byte(fmt.Sprintf("Scout %d:Scout ", sid+1))
		}
		statusLine := []byte(fmt.Sprintf("%s Status: ", ul.UnitId))
		for n, line := range lines {
			if n == 1 {
				section.TurnInfo = bdup(line)
			} else if bytes.HasPrefix(line, followsLine) {
				if section.Follows != nil {
					section.Error = cerrs.ErrMultipleFollowsLines
					break
				}
				section.Follows = line
			} else if bytes.HasPrefix(line, movesLine) {
				// remove the prefix and trim the line
				line = bytes.TrimSpace(bytes.TrimPrefix(line, movesLine))
				if bytes.HasPrefix(line, []byte{'M', 'o', 'v', 'e'}) {
					line = bytes.TrimSpace(bytes.TrimPrefix(line, []byte{'M', 'o', 'v', 'e'}))
				}
				section.Moves = scrubMoves(line)
			} else if bytes.HasPrefix(line, []byte{'S', 'c', 'o', 'u', 't'}) {
				for sid := 0; sid < 8; sid++ {
					var scoutSteps [][]byte
					if bytes.HasPrefix(line, scoutLines[sid]) {
						for _, jo := range scrubScouts(line) {
							if len(jo) != 0 {
								scoutSteps = append(scoutSteps, jo)
							}
						}
						if len(scoutSteps) > 0 {
							section.Scout = append(section.Scout, scoutSteps)
						}
						break
					}
				}
			} else if bytes.HasPrefix(line, statusLine) {
				if section.Status != nil {
					section.Error = cerrs.ErrMultipleStatusLines
					break
				}
				section.Status = bdup(line)
			}
		}
		if section.Error != nil {
			continue
		}

		// consistency checks
		if section.Follows != nil && section.Moves != nil {
			section.Error = cerrs.ErrUnitMovesAndFollows
			continue
		} else if len(section.Scout) > 8 {
			section.Error = cerrs.ErrTooManyScoutLines
			continue
		} else if section.Status == nil {
			section.Error = cerrs.ErrMissingStatusLine
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

// scrubMoves splits the line into individual moves and then removes
// leading and trailing spaces and any trailing commas from the move.
func scrubMoves(line []byte) [][]byte {
	var moves [][]byte
	for _, move := range bytes.Split(line, []byte{'\\'}) {
		move = bytes.TrimSpace(bytes.TrimRight(move, ", \t"))
		if len(move) != 0 {
			moves = append(moves, move)
		}
	}
	return moves
}

func scrubScouts(line []byte) [][]byte {
	// trim the scout number and then split the rest of the line
	line = line[len("Scout 8:Scout "):]
	return scrubMoves(line)
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
func split(input []byte) ([][]byte, []byte) {
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
	}

	return sections, separator
}
