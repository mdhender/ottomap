// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/coords"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/lbmoves"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/mdhender/ottomap/reports"
	"github.com/mdhender/ottomap/wxx"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func MapReports(allReports []*reports.Report, clanId, originGridId, outputPath string, showGridCenters, showGridCoords, showGridNumbers, showIgnoredSections, showSectionData, showSkippedSections, showSteps, debugSteps, debugNodes bool) error {
	// create a map for every movement result we have
	var allMovementResults []*lbmoves.MovementResults

	// this hex is the clan's starting hex from the first turn report
	var originHex *wxx.Hex

	// start collecting turn and data so that we can track units across turns
	type TurnData_t struct {
		Id       string // turn id
		NewUnits map[string]bool
		AllUnits []string
	}
	type UnitData_t struct {
		Id            string // unit id
		FirstSeenTurn string // turn id unit was first seen
		LastSeenTurn  string // turn id unit was last seen (not counting this turn)
		prevHex       *wxx.Hex
		currHex       *wxx.Hex
	}
	turnData, unitData := map[string]*TurnData_t{}, map[string]*UnitData_t{}
	var allTurnIds, allUnitIds []string

	// parse the report files into a single map
	for _, rpt := range allReports {
		if rpt.Ignore {
			if showIgnoredSections {
				log.Printf("report %s: ignored report\n", rpt.Id)
			}
			continue
		}

		// load the report file
		if showSectionData {
			log.Printf("report %s: reading file %s\n", rpt.Id, rpt.Path)
		}
		data, err := os.ReadFile(rpt.Path)
		if err != nil {
			log.Fatalf("error: report %s: %v", rpt.Path, err)
		}
		if showSectionData {
			log.Printf("report %s: loaded %8d bytes\n", rpt.Id, len(data))
		}

		// check for bom and remove it if present
		if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
			log.Printf("report %s: skipped %8d BOM bytes\n", rpt.Id, 3)
			data = data[3:]
		}

		// split the report into sections before parsing it.
		// note that the Sections() function does some processing to set fields in the Section struct.
		var secError *reports.Error
		rpt.Sections, secError = reports.Sections(rpt.Id, data, showSkippedSections)
		if showSectionData {
			log.Printf("report %s: loaded %8d sections\n", rpt.Id, len(rpt.Sections))
		}
		if secError != nil {
			for _, section := range rpt.Sections {
				if section.Error != nil {
					log.Printf("error: report %s: section %s: line %d: %v\n", rpt.Id, section.Id, section.Error.Line.No, section.Error.Error)
				}
			}
			log.Fatalf("error: report %s: please fix errors listed above, then restart\n", rpt.Id)
		}

		// each section may contain multiple movement results, so we need to parse all of them and consolidate the results
		for _, section := range rpt.Sections {
			if showSectionData {
				log.Printf("report %s: section %s: parsing\n", rpt.Id, section.Id)
			}
			if section.StatusLine == nil {
				log.Fatalf("error: report %s: section %s: no status line\n", rpt.Id, section.Id)
			}

			turnId, unitId, prevGridCoords := rpt.TurnId, section.UnitId, section.PrevCoords
			log.Printf("report %s: section %s: turn %s: unit %s: prev %s\n", rpt.Id, section.Id, turnId, unitId, prevGridCoords)
			if _, ok := turnData[turnId]; !ok {
				turnData[turnId] = &TurnData_t{Id: turnId, NewUnits: map[string]bool{}}
				allTurnIds = append(allTurnIds, turnId)
			}
			td := turnData[turnId]
			if _, ok := unitData[unitId]; !ok {
				td.NewUnits[unitId] = true
				unitData[unitId] = &UnitData_t{Id: unitId, FirstSeenTurn: turnId}
				allUnitIds = append(allUnitIds, unitId)
			}
			ud := unitData[unitId]
			ud.LastSeenTurn = turnId
			td.AllUnits = append(td.AllUnits, unitId)

			missingPrevGridCoords := prevGridCoords == ""
			if missingPrevGridCoords {
				// this can happen when a unit is created as an after-move action.
				// it can also happen when a unit is created as a before-move action and doesn't move.
				log.Printf("report %s: section %2s: turn %s: unit %-6s: warning: missing coordinates\n", rpt.Id, section.Id, turnId, unitId)
			} else if strings.HasPrefix(prevGridCoords, "##") {
				// remove these and let the walk logic populate them
				//log.Printf("report %s: section %2s: turn %s: unit %-6s: stripping %q\n", rpt.Id, section.Id, turnId, unitId, prevGridCoords)
				prevGridCoords = ""
			} else {
				//log.Printf("report %s: section %2s: turn %s: unit %-6s: keeping %q\n", rpt.Id, section.Id, turnId, unitId, prevGridCoords)
			}

			mrl := &lbmoves.MovementResults{
				TurnId:                  turnId,
				UnitId:                  unitId,
				StartingGridCoordinates: prevGridCoords,
			}
			allMovementResults = append(allMovementResults, mrl)

			if section.FleetMovement != nil {
				slug := string(section.FleetMovement.Text)
				if len(slug) > 55 {
					slug = slug[:55] + "..."
				}
				log.Printf("report %s: section %2s: line %d: found fleet movement\n\t%s\n", rpt.Id, section.Id, section.FleetMovement.LineNo, slug)
			}

			if section.FollowsLine != nil {
				//log.Printf("report %s: section %2s: follows %q\n", rpt.Id, section.Id, section.FollowsLine)
				steps, err := lbmoves.ParseMoveResults(turnId, unitId, section.FollowsLine.No, section.FollowsLine.Text, debugSteps, debugNodes)
				if err != nil {
					log.Fatalf("error: report %s: section %2s: line %d: %v\n", rpt.Id, section.Id, section.Follows.No, err)
				}
				mrl.Follows = steps[0].Follows
			}

			if section.MovementLine != nil {
				//log.Printf("report %s: section %2s: moves   %q\n", rpt.Id, section.Id, section.MovementLine)
				steps, err := lbmoves.ParseMoveResults(turnId, unitId, section.MovementLine.No, section.MovementLine.Text, debugSteps, debugNodes)
				if err != nil {
					log.Fatalf("error: report %s: section %2s: line %d: %v\n", rpt.Id, section.Id, section.MovementLine.No, err)
				}
				mrl.MovementReports = append(mrl.MovementReports, steps...)
			}

			// status line should count as a "stand still" movement so that we can update the hex with the unit's position
			steps, err := lbmoves.ParseStatusLine(turnId, unitId, section.StatusLine.No, section.StatusLine.Text, debugSteps, debugNodes)
			if err != nil {
				log.Fatalf("error: report %s: section %2s: line %d: %v\n", rpt.Id, section.Id, section.StatusLine.No, err)
			} else if len(steps) != 1 {
				log.Fatalf("error: report %s: section %2s: want 1 step, got %d\n", rpt.Id, section.Id, len(steps))
			}
			// todo: use the status line to update the hex with the unit's position
			// doing it here breaks the "after move" logic because the status line is not a movement line
			//mrl.MovementReports = append(mrl.MovementReports, steps...)
			mrl.StatusLine = steps[0] // todo: remove this and use the steps directly
			// daHex.Features.Visited = turnId
			//for n, step := range mrl.MovementReports {
			//	log.Printf("report %s: section %s: status %d: terrain %+v\n", rpt.Id, section.Id, n+1, step.Terrain)
			//	log.Printf("report %s: section %s: status %d: result  %d %q\n", rpt.Id, section.Id, n+1, step.Result, step.Result)
			//}

			for _, scoutLine := range section.ScoutLines {
				if scoutLine != nil {
					//log.Printf("report %s: section %2s: scouts  %q\n", rpt.Id, section.Id, scoutLine)
					steps, err := lbmoves.ParseScoutLine(turnId, unitId, scoutLine.No, scoutLine.Text, debugSteps, debugNodes)
					if err != nil {
						log.Fatalf("error: report %s: section %2s: line %d: %v\n", rpt.Id, section.Id, scoutLine.No, err)
					}
					mrl.ScoutReports = append(mrl.ScoutReports, steps)
				}
			}

			if missingPrevGridCoords {
				// todo: report on if we were able to find the previous coordinates or not
			}
		}

		if showSectionData {
			log.Printf("report %s: len(moves) now %8d\n", rpt.Path, len(allMovementResults))
		}
	}

	if len(allMovementResults) == 0 {
		log.Fatalf("error: no movement results found\n")
	}

	sort.Strings(allTurnIds)
	for _, turnId := range allTurnIds {
		td := turnData[turnId]
		log.Printf("turn %s: %+v %+v\n", td.Id, td.NewUnits, td.AllUnits)
	}

	sort.Strings(allUnitIds)
	for _, unitId := range allUnitIds {
		log.Printf("unit %-8s: first %s: last %s\n", unitData[unitId].Id, unitData[unitId].FirstSeenTurn, unitData[unitId].LastSeenTurn)
	}

	// users are required to provide starting grid coordinates if they're not already in the report
	log.Printf("starting grid coordinates: %q\n", allMovementResults[0].StartingGridCoordinates)
	if strings.HasPrefix(allMovementResults[0].StartingGridCoordinates, "##") {
		log.Printf("warning: hidden grid origin: %q\n", allMovementResults[0].StartingGridCoordinates)
		if originGridId == "" {
			log.Fatalf("error: starting grid coordinates must be specified\n")
		}
		allMovementResults[0].StartingGridCoordinates = originGridId + strings.TrimPrefix(allMovementResults[0].StartingGridCoordinates, "##")
		log.Printf("warning: grid origin set to %q\n", allMovementResults[0].StartingGridCoordinates)
	}

	movementResultsMap := map[string]*lbmoves.MovementResults{}
	for _, mrl := range allMovementResults {
		movementResultsMap[fmt.Sprintf("%s.%s", mrl.TurnId, mrl.UnitId)] = mrl
	}

	// assume that unit moves are in order and create unit follower links
	for _, mrl := range allMovementResults {
		if mrl.Follows == "" {
			continue
		}
		turnUnitStepId := fmt.Sprintf("%s.%s", mrl.TurnId, mrl.Follows)
		theOtherUnit, ok := movementResultsMap[turnUnitStepId]
		if !ok {
			log.Fatalf("error: turn %s: unit %-8s: follows: %-8s: turn %s not found\n", mrl.TurnId, mrl.UnitId, mrl.Follows, turnUnitStepId)
		}
		theOtherUnit.Followers = append(theOtherUnit.Followers, mrl)
	}

	if showSteps {
		for _, mrl := range allMovementResults {
			if boo, err := json.MarshalIndent(mrl.StatusLine, "", "\t"); err == nil {
				fmt.Printf("status: %s\n", string(boo))
			}
			for _, mr := range mrl.MovementReports {
				if boo, err := json.MarshalIndent(mr, "", "\t"); err == nil {
					fmt.Printf("movement report: %s\n", string(boo))
				}
			}
			for _, scout := range mrl.ScoutReports {
				for _, mr := range scout {
					if boo, err := json.MarshalIndent(mr, "", "\t"); err == nil {
						fmt.Printf("scout report: %s\n", string(boo))
					}
				}
			}
		}
	}

	allHexes := map[string]*wxx.Hex{}
	consolidatedMap := &wxx.WXX{}

	// unitNode is a unit that will be added to the map.
	// it will contain all the unit's moves. the parent
	// link is included to help with linking moves together.
	type unitNode struct {
		Id     string
		Parent *unitNode
		Moves  []*report.Unit
	}

	// initialize the world hex map from the first unit's status line
	if len(allMovementResults) == 0 {
		log.Fatalf("error: error: no movement results\n")
	}
	originTurnId := ""
	for _, mrl := range allMovementResults[:1] { // just the first result of the first turn
		log.Printf("%s: %-6s: casting about for origin\n", mrl.TurnId, mrl.UnitId)

		originTurnId = mrl.TurnId
		statusLine := mrl.StatusLine

		start := mrl.StartingGridCoordinates
		if start == "" {
			log.Fatalf("error: %s: %-6s: origin is not defined for first unit's first turn\n", mrl.TurnId, mrl.UnitId)
		}

		originGridCoords, err := coords.StringToGridCoords(start)
		if err != nil {
			log.Fatalf("error: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
		}
		log.Printf("%s: %-6s: status line: grid %s\n", mrl.TurnId, mrl.UnitId, originGridCoords)
		originMapCoords, err := originGridCoords.ToMapCoords()
		if err != nil {
			log.Fatalf("error: %s: %-6s: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
		}
		gridColumn, gridRow := originMapCoords.GridColumnRow()
		originHex = &wxx.Hex{
			GridId:     originMapCoords.GridId(),
			GridCoords: originMapCoords.GridString(),
			Offset:     wxx.Offset{Column: gridColumn, Row: gridRow},
			Terrain:    statusLine.Terrain,
			Features: wxx.Features{
				IsOrigin: true,
				Created:  originTurnId,
				Updated:  originTurnId,
				Visited:  originTurnId,
			},
		}
		log.Printf("%s: %-6s: status line: %s\n", mrl.TurnId, mrl.UnitId, statusLine.Terrain)

		// stuff the origin into our consolidated map because it may never show up again
		log.Printf("warning: the origin hex needs to be re-worked\n")
		allHexes[originMapCoords.GridString()] = originHex
	}
	if originHex == nil {
		log.Printf("%s: origin is not defined\n", originTurnId)
		panic("assert(originHex != nil")
	}

	// ensure that status line is present for all turns
	for _, mrl := range allMovementResults {
		if mrl.StatusLine == nil {
			log.Fatalf("error: %s: %-6s: status line is missing\n", mrl.TurnId, mrl.UnitId)
		}
	}

	// process the movement results one turn at a time
	for n, turnId := range allTurnIds {
		// process the movement results for this turn
		unitId, unitOrigin, unitEnding := "", "", ""
		for _, mrl := range allMovementResults {
			if mrl.TurnId != turnId {
				continue
			}

			if unitId == "" {
				unitId, unitOrigin = mrl.UnitId, mrl.StartingGridCoordinates
			} else if unitId != mrl.UnitId {
				log.Printf("turn %-8s: unit %-8s: origin %-8s: ended %-8s\n", turnId, unitId, unitOrigin, unitEnding)
				unitId = mrl.UnitId
			}

			start := mrl.StartingGridCoordinates
			if start == "" {
				// attempt to use the coordinates from the parent's move
				parentTurnUnitStepId := fmt.Sprintf("%s.%s", turnId, lbmoves.ParentId(mrl.UnitId))
				//log.Printf("%s: %-6s: start %s: parent %-6s: turnId %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurnUnitStepId)
				parentTurn, ok := movementResultsMap[parentTurnUnitStepId]
				if !ok {
					log.Fatalf("error: %s: %-6s: start %s: parent %-6s: turnId %q: missing\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId)
				} else if parentTurn.StartingGridCoordinates == "" {
					log.Printf("%s: %-6s: start %s: parent %-6s: turnId %q: start %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId, parentTurn.StartingGridCoordinates)
					panic(`assert(parentTurn.StartingGridCoordinates != "")`)
				}
				start, mrl.StartingGridCoordinates = parentTurn.StartingGridCoordinates, parentTurn.StartingGridCoordinates
				//log.Printf("%s: %-6s: start %s: parent %-6s: turnId %q: start %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId, parentTurn.StartingGridCoordinates)
			}
			//log.Printf("%s: %-6s: origin %s\n", mrl.TurnId, mrl.UnitId, start)

			currGridCoords, err := coords.StringToGridCoords(start)
			if err != nil {
				log.Fatalf("error: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
			}
			//log.Printf("%s: %-6s: step %2d: grid %s\n", mrl.TurnId, mrl.UnitId, n, currGridCoords)
			currMapCoords, err := currGridCoords.ToMapCoords()
			if err != nil {
				log.Fatalf("error: %s: %-6s: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
			}

			debugStep := false
			for stepNo, step := range mrl.MovementReports {
				//log.Printf("%s: %-6s: step %2d: mapc %s\n", mrl.TurnId, mrl.UnitId, stepNo+1, currMapCoords)
				nextCoords := walk(turnId, mrl.UnitId, allHexes, stepNo+1, step, currMapCoords, debugStep)
				//log.Printf("%s: %-6s: step %2d: mapc %s next %s\n", mrl.TurnId, mrl.UnitId, stepNo+1, currMapCoords, nextCoords)
				currMapCoords = nextCoords
			}

			//log.Printf("%s: %-6s: steps %2d start %s: curr %s\n", mrl.TurnId, mrl.UnitId, len(mrl.MovementReports), start, currMapCoords.GridString())
			mrl.EndingGridCoordinates = currMapCoords.GridString()
			unitEnding = mrl.EndingGridCoordinates
		}
		if unitId != "" {
			log.Printf("turn %-8s: unit %-8s: origin %-8s: ended %-8s\n", turnId, unitId, unitOrigin, unitEnding)
		}

		// resolve "follows" links for this turn
		for _, mrl := range allMovementResults {
			if mrl.TurnId != turnId || mrl.Followers == nil {
				continue
			}

			for _, follower := range mrl.Followers {
				//log.Printf("%s: %-6s: follower %-8q %q -> %q\n", mrl.TurnId, mrl.UnitId, follower.UnitId, follower.EndingGridCoordinates, mrl.EndingGridCoordinates)
				follower.EndingGridCoordinates = mrl.EndingGridCoordinates
			}
		}

		// process the scout results for this turn
		for _, mrl := range allMovementResults {
			if mrl.TurnId != turnId {
				continue
			}

			for scoutNo, scout := range mrl.ScoutReports {
				// each scout starts at the unit's current location
				startGridCoords, err := coords.StringToGridCoords(mrl.EndingGridCoordinates)
				if err != nil {
					log.Fatalf("error: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
				}

				//log.Printf("%s: %-6s: scout %d: grid %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, startGridCoords)
				currMapCoords, err := startGridCoords.ToMapCoords()
				if err != nil {
					log.Fatalf("error: %s: %-6s: scout %d: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, scoutNo+1, err)
				}

				debugStep := false
				movementReports := scout
				for stepNo, step := range movementReports {
					//log.Printf("%s: %-6s: scout %d: step %2d: mapc %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, stepNo+1, currMapCoords)
					nextCoords := walk(turnId, mrl.UnitId, allHexes, stepNo+1, step, currMapCoords, debugStep)
					//log.Printf("%s: %-6s: scout %d: step %2d: mapc %s next %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, stepNo+1, currMapCoords, nextCoords)
					currMapCoords = nextCoords
				}

				log.Printf("turn %-8s: unit %-8s: origin %-8s: ended %-8s\n", turnId, fmt.Sprintf("%ss%d", mrl.UnitId, scoutNo+1), startGridCoords.String(), currMapCoords.GridString())
			}
		}

		// if there are more turns to process, then the ending position this turn becomes the starting position for the next turn
		if n < len(allTurnIds)-1 {
			missingEnds, nextTurnId := 0, allTurnIds[n+1]
			//log.Printf("turnId %s: next turnId %s\n", turnId, nextTurnId)
			for _, mrl := range allMovementResults {
				if mrl.TurnId != turnId {
					continue
				}

				//log.Printf("%s: %-6s: start %s: end: %s\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates, mrl.EndingGridCoordinates)
				if mrl.EndingGridCoordinates == "" {
					log.Printf("%s: %-6s: start %s: end: missing\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates)
					missingEnds++
				}

				// does the unit have a move next turn? if they don't, we're hosed since we can't look forward multiple turns yet.
				nextTurnUnitStepId := fmt.Sprintf("%s.%s", nextTurnId, mrl.UnitId)
				nextTurn, ok := movementResultsMap[nextTurnUnitStepId]
				//if !ok {
				//	log.Fatalf("error: %s: %-6s: start %s: end: missing next turn\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates)
				//}
				//nextTurn.StartingGridCoordinates = mrl.EndingGridCoordinates
				if !ok {
					log.Printf("%s: %-6s: start %s: unit is missing next turn\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates)
				} else {
					nextTurn.StartingGridCoordinates = mrl.EndingGridCoordinates
				}
				//log.Printf("%s: %-6s: start %s: end: %s <-- next turn\n", nextTurn.TurnId, nextTurn.UnitId, nextTurn.StartingGridCoordinates, nextTurn.EndingGridCoordinates)
			}
			if missingEnds != 0 {
				log.Fatalf("error: %s: error: walk did not populate %d units\n", turnId, missingEnds)
			}
		}

		var newHexes, updatedHexes []*wxx.Hex
		for _, hex := range allHexes {
			if hex.Features.Created == turnId {
				newHexes = append(newHexes, hex)
			} else if hex.Features.Updated == turnId {
				updatedHexes = append(updatedHexes, hex)
			}

			if hex.GridCoords == "" {
				panic("!")
			}
		}
		if err := consolidatedMap.MergeHexes(turnId, newHexes); err != nil {
			log.Printf("error: wxx: mergeHexes: newHexes: %v\n", err)
			return err
		} else if err = consolidatedMap.MergeHexes(turnId, updatedHexes); err != nil {
			log.Printf("error: wxx: mergeHexes: updatedHexes: %v\n", err)
			return err
		}

		if showGridCoords {
			consolidatedMap.AddGridCoords()
		} else if showGridNumbers {
			consolidatedMap.AddGridNumbering()
		}

		//if sectionMaps {
		//	panic("this needs to be fixed")
		//}

		// now we can create the Worldographer map!
		//log.Printf("world %6d: consolidated %6d\n", len(worldMap), len(consolidatedMap))
		mapName := filepath.Join(outputPath, fmt.Sprintf("%s.%s.wxx", turnId, clanId))
		if err := consolidatedMap.Create(mapName, showGridCenters); err != nil {
			log.Printf("creating %s\n", mapName)
			log.Fatal(err)
		}
		log.Printf("created  %s\n", mapName)
	}

	if showGridCoords {
		consolidatedMap.AddGridCoords()
	} else if showGridNumbers {
		consolidatedMap.AddGridNumbering()
	}

	// now we can create the Worldographer map!
	mapName := filepath.Join(outputPath, fmt.Sprintf("%s.wxx", clanId))
	if err := consolidatedMap.Create(mapName, showGridCenters); err != nil {
		log.Printf("creating %s\n", mapName)
		log.Fatalf("error: %v\n", err)
	}
	log.Printf("created  %s\n", mapName)

	return nil
}

type gridHexes struct {
	Grid  string              // the grid in the big map
	Hexes map[string]*wxx.Hex // key is hex coordinates
}

func walk(turnId, unitId string, worldHexMap map[string]*wxx.Hex, stepNo int, step *lbmoves.Step, start coords.Map, debugStep bool) coords.Map {
	if debugStep {
		if buf, err := json.MarshalIndent(step, "", "\t"); err == nil {
			log.Printf("debugStep: step %s\n", string(buf))
		}
	}
	//log.Printf("%s: %-6s: step %2d: mapc %s\n", turnId, unitId, stepNo, start)

	// advance a hex if the move succeeded
	curr := start
	if step.Result == lbmoves.Succeeded {
		if step.Attempted == directions.DUnknown {
			panic(fmt.Sprintf("assert(step.Attempted != %d)", int(step.Attempted)))
		}
		curr = start.Add(step.Attempted)
	}

	// fetch hex based on the current map coordinates (create as needed)
	fetchHex := func(mc coords.Map, defaultTerrain domain.Terrain) *wxx.Hex {
		daHex, ok := worldHexMap[mc.GridString()]
		if !ok { // create a new hex with normalized map coordinates (to the grid's origin)
			gridColumn, gridRow := mc.GridColumnRow()
			daHex = &wxx.Hex{
				GridId:     mc.GridId(),
				GridCoords: mc.GridString(),
				Offset:     wxx.Offset{Column: gridColumn, Row: gridRow},
				Terrain:    defaultTerrain,
				Features: wxx.Features{
					Created: turnId,
				},
			}
			worldHexMap[mc.GridString()] = daHex
		}
		return daHex
	}

	daHex := fetchHex(curr, step.Terrain)
	//log.Printf("%s: %-6s: step %2d: mapc %s: grid id %q\n", turnId, unitId, stepNo, curr, curr.GridId())
	daHex.Features.Updated = turnId

	if step.Result == lbmoves.StatusLine {
		daHex.Features.Visited = turnId
		if step.Settlement != nil {
			daHex.Features.Settlement = &wxx.Settlement{
				UUID: uuid.New().String(),
				Name: step.Settlement.Name,
			}
		}
	} else if step.Result == lbmoves.Succeeded {
		daHex.Features.Visited = turnId
		if step.Settlement != nil {
			daHex.Features.Settlement = &wxx.Settlement{
				UUID: uuid.New().String(),
				Name: step.Settlement.Name,
			}
		}
	} else if step.Result == lbmoves.Blocked {
		// log.Printf("step: blocked: attempted %s\n", step.Attempted)
		if step.BlockedBy == nil {
			panic("assert(blockedBy != nil")
		}
		// log.Printf("step: blocked: edge: %+v\n", *step.BlockedBy)
	} else if step.Result == lbmoves.Prohibited {
		if step.ProhibitedFrom == nil {
			panic("assert(step.ProhibitedFrom != nil)")
		}
		if step.ProhibitedFrom.Direction == directions.DUnknown {
			panic(fmt.Sprintf("assert(prohibitedFrom.Direction != %d)", int(step.ProhibitedFrom.Direction)))
		}
		neighborMapCoords := curr.Add(step.ProhibitedFrom.Direction)
		//log.Printf("%s: %-6s: step %2d: curr %s: pro %s\n", turnId, unitId, stepNo, curr, neighborMapCoords)

		daNeighborHex := fetchHex(neighborMapCoords, step.ProhibitedFrom.Terrain)
		if daNeighborHex.Terrain != step.ProhibitedFrom.Terrain {
			panic(fmt.Sprintf("assert(daNeighborHex.Terrain  == %d)", int(step.ProhibitedFrom.Terrain)))
		}
	}

	// update the hex's neighbors
	for _, neighbor := range step.Neighbors {
		if neighbor.Direction == directions.DUnknown {
			panic(fmt.Sprintf("assert(neighbor.Direction != %d)", int(neighbor.Direction)))
		}
		neighborMapCoords := curr.Add(neighbor.Direction)

		daNeighborHex := fetchHex(neighborMapCoords, neighbor.Terrain)
		if daNeighborHex.Terrain != neighbor.Terrain {
			log.Printf("%s: %-6s: step %2d: curr %s\n", turnId, unitId, stepNo, curr)
			log.Printf("%s: %-6s: step %2d:  nbr %s\n", turnId, unitId, stepNo, neighborMapCoords)
			log.Printf("%s: %-6s: step %2d:  nbr terrain: %s\n", turnId, unitId, stepNo, neighbor.Terrain)
			log.Printf("%s: %-6s: step %2d:  nbr    want: %s\n", turnId, unitId, stepNo, daNeighborHex.Terrain)
			panic(fmt.Sprintf("assert(hex.Terrain == %q)", neighbor.Terrain))
		}
	}

	// update the hex edges
	for _, edge := range step.Edges {
		switch edge.Edge {
		case domain.ENone:
			// ignore, shouldn't ever happen. maybe we should panic?
		case domain.EFord:
			daHex.Features.Edges.Ford = append(daHex.Features.Edges.Ford, edge.Direction)
		case domain.EPass:
			daHex.Features.Edges.Pass = append(daHex.Features.Edges.Pass, edge.Direction)
		case domain.ERiver:
			daHex.Features.Edges.River = append(daHex.Features.Edges.River, edge.Direction)
		case domain.EStoneRoad:
			daHex.Features.Edges.StoneRoad = append(daHex.Features.Edges.StoneRoad, edge.Direction)
		default:
			panic(fmt.Sprintf("assert(edge != %d)", edge.Edge))
		}
	}

	// update the hex resources
	if step.Resources != domain.RNone {
		if daHex.Features.Resources == domain.RNone {
			daHex.Features.Resources = step.Resources
		} else if daHex.Features.Resources != step.Resources {
			log.Printf("why? changing %q to %q\n", daHex.Features.Resources, step.Resources)
			panic("!")
		}
	}

	switch step.Result {
	case lbmoves.Blocked:
	case lbmoves.ExhaustedMovementPoints:
	case lbmoves.Follows:
	case lbmoves.Prohibited:
	case lbmoves.StatusLine:
	case lbmoves.StayedInPlace:
	case lbmoves.Succeeded:
	case lbmoves.Vanished:
	default:
		panic(fmt.Sprintf("assert(step.Result != %d)", int(step.Result)))
	}

	return curr
}
