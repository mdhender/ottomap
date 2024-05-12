// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/ottomap/config"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/lbmoves"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/mdhender/ottomap/reports"
	"github.com/mdhender/ottomap/wxx"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var argsMap struct {
	config       string // path to configuration file
	clanId       string // clan id to use
	turnId       string // turn id to use
	gridOriginId string // grid id to use for origin
	debug        struct {
		units bool
	}
}

var cmdMap = &cobra.Command{
	Use:   "map",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// if grid origin id is specified, it must be AA ... ZZ
		if argsMap.gridOriginId != "" {
			if strings.TrimSpace(argsMap.gridOriginId) != argsMap.gridOriginId {
				return fmt.Errorf("grid-origin: can not contain spaces")
			} else if len(argsMap.gridOriginId) != 2 {
				return fmt.Errorf("grid orgin: must be two upper-case letters")
			} else if strings.Trim(argsMap.gridOriginId, "ABCDEFGHIJKLMNOPQRSTUVWYZ") != "" {
				return fmt.Errorf("grid orgin: must be two upper-case letters")
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("maps: todo: detect when a unit is created as an after-move action\n")

		log.Printf("map: config: file %s\n", argsMap.config)
		cfg, err := config.Load(argsMap.config)
		if err != nil {
			log.Fatalf("map: config: %v\n", err)
		}
		if len(cfg.Reports) == 0 {
			log.Fatalf("map: config: no reports\n")
		}

		log.Printf("map: config: path   %s\n", cfg.Path)
		log.Printf("map: config: output %s\n", cfg.OutputPath)

		cfg.Inputs.ClanId = argsMap.clanId
		log.Printf("map: config: clan %q\n", cfg.Inputs.ClanId)

		cfg.Inputs.GridOriginId = argsMap.gridOriginId
		if cfg.Inputs.GridOriginId != "" {
			log.Printf("map: config: originGridId %q\n", cfg.Inputs.GridOriginId)
		}

		// if turn id is not on the command line, use the current turn from the configuration.
		if argsMap.turnId == "" {
			// assumes that the configuration's reports are sorted by turn id.
			rptCurr := cfg.Reports[len(cfg.Reports)-1]
			cfg.Inputs.TurnId = rptCurr.TurnId
			cfg.Inputs.Year = rptCurr.Year
			cfg.Inputs.Month = rptCurr.Month
		} else {
			// convert command line's yyyy-mm to year, month
			if yyyy, mm, ok := strings.Cut(argsMap.turnId, "-"); !ok {
				log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
			} else if yyyy = strings.TrimSpace(yyyy); yyyy == "" {
				log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
			} else if cfg.Inputs.Year, err = strconv.Atoi(yyyy); err != nil {
				log.Fatalf("map: invalid turn %q: year %v\n", argsMap.turnId, err)
			} else if mm = strings.TrimSpace(mm); mm == "" {
				log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
			} else if cfg.Inputs.Month, err = strconv.Atoi(mm); err != nil {
				log.Fatalf("map: invalid turn %q: month %v\n", argsMap.turnId, err)
			} else {
				cfg.Inputs.TurnId = fmt.Sprintf("%04d-%02d", cfg.Inputs.Year, cfg.Inputs.Month)
			}
		}
		log.Printf("map: config: turn year  %4d\n", cfg.Inputs.Year)
		log.Printf("map: config: turn month %4d\n", cfg.Inputs.Month)

		// update the ignore flag based on the turn from the configuration
		for _, rpt := range cfg.Reports {
			if rpt.Clan == argsMap.clanId {
				rptTurnId := fmt.Sprintf("%04d-%02d", rpt.Year, rpt.Month)
				if rptTurnId > cfg.Inputs.TurnId {
					rpt.Ignore = true
				}
			}
		}

		// collect the reports that we're going to process
		var allReports []*reports.Report
		for _, rpt := range cfg.Reports {
			if rpt.Ignore {
				continue
			}
			allReports = append(allReports, rpt)
		}
		if len(allReports) == 0 {
			log.Fatalf("map: files: no files matched constraints\n")
		}
		log.Printf("map: reports %d\n", len(allReports))

		// create a map for every movement result we have
		var allMovementResults []*lbmoves.MovementResults

		// parse the report files into a single map
		for _, rpt := range cfg.Reports {
			if rpt.Ignore {
				if cfg.Inputs.ShowIgnoredReports {
					log.Printf("map: report %s: ignored report\n", rpt.Id)
				}
				continue
			}

			// load the report file
			data, err := os.ReadFile(rpt.Path)
			if err != nil {
				log.Fatalf("map: report %s: %v", rpt.Path, err)
			}
			log.Printf("map: report %s: loaded %8d bytes\n", rpt.Id, len(data))

			// split the report into sections before parsing it
			rpt.Sections, err = reports.Sections(data, cfg.Inputs.ShowSkippedSections)
			log.Printf("map: report %s: loaded %8d sections\n", rpt.Id, len(rpt.Sections))
			if err != nil {
				for _, section := range rpt.Sections {
					if section.Error != nil {
						log.Printf("map: report %s: section %s: %v\n", rpt.Id, section.Id, section.Error)
					}
				}
				log.Fatalf("map: report %s: please fix errors listed above, then restart\n", rpt.Id)
			}

			// each section may contain multiple movement results, so we need to collect them all

			// parse the report, stopping if there's an error
			for _, section := range rpt.Sections {
				log.Printf("map: report %s: section %2s: need to extract grid hex data\n", rpt.Id, section.Id)

				turnId, unitId, prevGridCoords := rpt.TurnId, section.UnitId, section.PrevCoords
				if prevGridCoords == "" {
					prevGridCoords = section.CurrCoords
					if prevGridCoords == "" {
						log.Fatalf("map: report %s: section %2s: no starting grid coordinates\n", rpt.Id, section.Id)
					}
				}

				mrl := &lbmoves.MovementResults{
					TurnId:                  turnId,
					UnitId:                  unitId,
					StartingGridCoordinates: prevGridCoords,
				}
				allMovementResults = append(allMovementResults, mrl)

				if section.StatusLine == nil {
					log.Fatalf("map: report %s: section %2s: no status line\n", rpt.Id, section.Id)
				} else {
					steps, err := lbmoves.ParseMoveResults(turnId, unitId, section.StatusLine, cfg.Inputs.ShowSteps)
					if err != nil {
						log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
					} else if len(steps) != 1 {
						log.Fatalf("map: report %s: section %2s: want 1 step, got %d\n", rpt.Id, section.Id, len(steps))
					}
					mrl.StatusLine = steps[0]
				}
				if section.FollowsLine != nil {
					//log.Printf("map: report %s: section %2s: follows %q\n", rpt.Id, section.Id, section.FollowsLine)
					steps, err := lbmoves.ParseMoveResults(turnId, unitId, section.FollowsLine, cfg.Inputs.ShowSteps)
					if err != nil {
						log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
					}
					mrl.Follows = steps[0].Follows
				}
				if section.MovementLine != nil {
					//log.Printf("map: report %s: section %2s: moves   %q\n", rpt.Id, section.Id, section.MovementLine)
					steps, err := lbmoves.ParseMoveResults(turnId, unitId, section.MovementLine, cfg.Inputs.ShowSteps)
					if err != nil {
						log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
					}
					mrl.MovementReports = append(mrl.MovementReports, steps...)
				}
				for _, scoutLine := range section.ScoutLines {
					if scoutLine != nil {
						//log.Printf("map: report %s: section %2s: scouts  %q\n", rpt.Id, section.Id, scoutLine)
						steps, err := lbmoves.ParseMoveResults(turnId, unitId, scoutLine, cfg.Inputs.ShowSteps)
						if err != nil {
							log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
						}
						mrl.ScoutReports = append(mrl.ScoutReports, steps)
					}
				}
			}
		}

		if len(allMovementResults) == 0 {
			log.Fatalf("map: no movement results found\n")
		}

		// users are required to provide starting grid coordinates if they're not already in the report
		log.Printf("map: starting grid coordinates: %q\n", allMovementResults[0].StartingGridCoordinates)
		if strings.HasPrefix(allMovementResults[0].StartingGridCoordinates, "##") {
			log.Printf("map: warning: hidden grid origin: %q\n", allMovementResults[0].StartingGridCoordinates)
			if cfg.Inputs.GridOriginId == "" {
				log.Fatalf("map: starting grid coordinates must be specified\n")
			}
			allMovementResults[0].StartingGridCoordinates = cfg.Inputs.GridOriginId + strings.TrimPrefix(allMovementResults[0].StartingGridCoordinates, "##")
			log.Printf("map: warning: grid origin set to %q\n", allMovementResults[0].StartingGridCoordinates)
		}

		//// sort unit moves by turn then unit
		//sort.Slice(unitMoves, func(i, j int) bool {
		//	return unitMoves[i].SortKey() < unitMoves[j].SortKey()
		//})

		movementResultsMap := map[string]*lbmoves.MovementResults{}
		for _, mrl := range allMovementResults {
			movementResultsMap[fmt.Sprintf("%s.%s", mrl.TurnId, mrl.UnitId)] = mrl
		}

		log.Printf("map: mrl summary commented out\n")
		//for _, uss := range allMovementResults {
		//	var firstGridCoords, lastGridCoords string
		//	firstGridCoords, lastGridCoords = uss.StartingGridCoordinates, "?"
		//	log.Printf("map: mrl: %-24s %-16s %-12s %3d %3d %-10q %-10q\n", uss.Id(), uss.TurnId, uss.UnitId, len(uss.MovementReports), len(uss.ScoutReports), firstGridCoords, lastGridCoords)
		//}

		// assume that unit moves are in order and create unit follower links
		for _, mrl := range allMovementResults {
			if mrl.Follows == "" {
				continue
			}
			turnUnitStepId := fmt.Sprintf("%s.%s", mrl.TurnId, mrl.Follows)
			theOtherUnit, ok := movementResultsMap[turnUnitStepId]
			if !ok {
				log.Fatalf("map: turn %s: unit %-8s: follows: %-8s: turn %s not found\n", mrl.TurnId, mrl.UnitId, mrl.Follows, turnUnitStepId)
			}
			theOtherUnit.Followers = append(theOtherUnit.Followers, mrl)
		}

		log.Printf("map: todo: followers are not updated after movement\n")

		log.Printf("map: todo: hexes are not assigned for each step in the results\n")

		log.Printf("map: todo: named hexes that are only in the status line are missed\n")

		if cfg.Inputs.ShowSteps {
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

		hexes := []wxx.Hex{
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 8}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 7}, Terrain: domain.TOcean},
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 6}, Terrain: domain.TSwamp},
			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 7}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 6}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 7}, Terrain: domain.TSwamp},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 6}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 5}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 4}, Terrain: domain.TPrairie},

			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 8}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 7}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 8}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 7, Row: 9}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 7, Row: 8}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 7, Row: 7}, Terrain: domain.TLowAridMountains},

			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 6}, Terrain: domain.TSwamp},

			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 9}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 10}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 11}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 11}, Terrain: domain.TRockyHills},

			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 8}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 9}, Terrain: domain.TBrushHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 12, Row: 9}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 13, Row: 9}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 14, Row: 9}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 14, Row: 8}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 13, Row: 8}, Terrain: domain.TRockyHills},
		}

		//// create a map of all units to help with linking moves together.
		//allUnits := map[string]*unitNode{}
		//for _, u := range unitMoves {
		//	un, ok := allUnits[u.Id]
		//	if !ok {
		//		un = &unitNode{Id: u.Id}
		//		if !u.IsClan() {
		//			un.Parent, ok = allUnits[u.ParentId]
		//			if !ok {
		//				log.Fatalf("map: unit %q: parent %q not found\n", u.Id, u.ParentId)
		//			}
		//		}
		//		allUnits[u.Id] = un
		//	}
		//	un.Moves = append(un.Moves, u)
		//}
		//
		//// create a sorted list of all units for dealing with parent/child relationships.
		//var sortedNodes []*unitNode
		//for _, un := range allUnits {
		//	sortedNodes = append(sortedNodes, un)
		//}
		//sort.Slice(sortedNodes, func(i, j int) bool {
		//	return sortedNodes[i].Id < sortedNodes[j].Id
		//})

		// This code is responsible for establishing the link between a unit's move and
		// the move of the unit it is following. By setting the Follows field of the
		// current move (cm) to the corresponding move of the followed unit (fm), it allows
		// the program to access the movement data of the followed unit when processing the
		// current unit's movement.
		//
		// This linking process is necessary because the parser may encounter a "follows"
		// command before it has parsed the movement data of the unit being followed. By
		// linking the moves, the program can correctly apply the movement of the followed
		// unit to the following unit when processing the movement data.
		//
		// Since we link to a specific unit's turn and use that turn's Moves field, we don't
		// need to worry about circular references.
		//
		// Iterate over all units.
		//for _, u := range allUnits {
		//	// Iterate over each unit's moves.
		//	for _, cm := range u.Moves {
		//		// If the current move doesn't follow another unit, skip it
		//		if cm.FollowsId == "" {
		//			continue
		//		}
		//
		//		// Look up the unit being followed using the FollowsId.
		//		// If the unit being followed is not found, log a fatal error
		//		fu, ok := allUnits[cm.FollowsId]
		//		if !ok {
		//			log.Fatalf("map: unit %q: follower %q not found\n", u.Id, cm.FollowsId)
		//		}
		//
		//		// Iterate over the moves of the unit being followed to find the move matching the current turn.
		//		for _, fm := range fu.Moves {
		//			// Check if the move of the followed unit matches the current turn
		//			if fm.Turn.Year == cm.Turn.Year && fm.Turn.Month == cm.Turn.Month {
		//				// If a matching move is found, link the current move to the followed unit's move
		//				cm.Follows = fm
		//				break
		//			}
		//		}
		//
		//		// If no matching move is found for the followed unit in the current turn, log a fatal error
		//		if cm.Follows == nil {
		//			log.Fatalf("map: unit %q: follower %q: turn %04d-%-2d not found\n", u.Id, cm.FollowsId, cm.Turn.Year, cm.Turn.Month)
		//		}
		//	}
		//}
		//
		//// save for debugging
		//for _, u := range unitMoves {
		//	if b, err := json.MarshalIndent(u, "", "  "); err != nil {
		//		log.Printf("map: unit %q: error: %v\n", u.Id, err)
		//	} else {
		//		log.Printf("map: unit %q: results\n%s\n", u.Id, string(b))
		//	}
		//}
		//

		// get a list of all the turn ids for later use
		var allTurnIds []string
		turnIdCounter := map[string]int{}
		for _, mrl := range allMovementResults {
			turnIdCounter[mrl.TurnId] = turnIdCounter[mrl.TurnId] + 1
		}
		for k := range turnIdCounter {
			allTurnIds = append(allTurnIds, k)
		}
		sort.Strings(allTurnIds)
		//for _, id := range allTurnIds {
		//	log.Printf("map: %s: %5d\n", id, turnIdCounter[id])
		//}

		type gridHexes struct {
			Grid  string              // the grid in the big map
			Hexes map[string]*wxx.Hex // key is hex coordinates
		}

		worldMap := map[string]*gridHexes{}

		// unitNode is a unit that will be added to the map.
		// it will contain all the unit's moves. the parent
		// link is included to help with linking moves together.
		type unitNode struct {
			Id     string
			Parent *unitNode
			Moves  []*report.Unit
		}

		// process the movement results one turn at a time
		for _, turnId := range allTurnIds {
			log.Printf("map: todo: walk the hex reports and update grid as well as ending coordinates\n")
			for _, mrl := range allMovementResults {
				if mrl.TurnId != turnId {
					continue
				}

				statusLine := mrl.StatusLine
				if statusLine == nil {
					log.Fatalf("map: %s: %-8q: status line is missing\n", mrl.TurnId, mrl.UnitId)
				}

				start := mrl.StartingGridCoordinates
				if start == "" {
					log.Fatalf("map: %s: %-8q: starting hex is missing\n", mrl.TurnId, mrl.UnitId)
				}
				// log.Printf("map: %s: %-8q: origin %s\n", mrl.TurnId, mrl.UnitId, start)

				if len(mrl.MovementReports) == 0 {
					//log.Printf("map: %s: %-8q: steps %2d\n", mrl.TurnId, mrl.UnitId, len(mrl.HexReports))
					mrl.EndingGridCoordinates = start
				} else {
					//log.Printf("map: %s: %-8q: steps %2d\n", mrl.TurnId, mrl.UnitId, len(mrl.MovementReports))
					// mrl.EndingGridCoordinates = "AA 0101"
				}

				// use the ending grid coordinates as the starting grid coordinates for the next turn's movement
			}

			// resolve "follows" links
			for _, mrl := range allMovementResults {
				if mrl.TurnId != turnId || mrl.Followers == nil {
					continue
				}
				for _, follower := range mrl.Followers {
					log.Printf("map: %s: %-8q: follower %-8q %q -> %q\n", mrl.TurnId, mrl.UnitId, follower.UnitId, follower.EndingGridCoordinates, mrl.EndingGridCoordinates)
					follower.EndingGridCoordinates = mrl.EndingGridCoordinates
				}
			}

			//// final sanity check for ending positions
			//for _, u := range unitMoves {
			//	if u.End == "" {
			//		log.Fatalf("map: unit %-8q: turn %04d-%02d: ending hex is missing\n", u.Id, u.Turn.Year, u.Turn.Month)
			//	}
			//}
		}

		log.Printf("map: world: %d\n", len(worldMap))

		// now we can create the Worldographer map!
		log.Printf("map: creating WXX map\n")
		w := &wxx.WXX{}
		if err := w.Create("working/testmap.wxx", hexes, false); err != nil {
			log.Fatal(err)
		}
		log.Printf("map: created %s\n", "working/testmap.wxx")

		return nil
	},
}
