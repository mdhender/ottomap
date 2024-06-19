// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/config"
	"github.com/mdhender/ottomap/coords"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/lbmoves"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/mdhender/ottomap/reports"
	"github.com/mdhender/ottomap/wxx"

	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var argsMap struct {
	paths struct {
		data   string
		config string // path to configuration file
	}
	clanId string // clan id to use
	turnId string // turn id to use
	debug  struct {
		units               bool
		sectionMaps         bool
		showIgnoredSections bool
		showSectionData     bool
	}
	show struct {
		gridCenters bool
		gridCoords  bool
		gridNumbers bool
	}
}

var cmdMap = &cobra.Command{
	Use:   "map",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// if paths.data is set, then it is an absolute path and the other values must be blank since they will be set by the absolute path
		if argsMap.paths.data != "" {
			// strip the default values if all of them are set
			if argsMap.paths.config == "data/config.json" {
				argsMap.paths.config = ""
			}
			// now check that they are not set
			if argsMap.paths.config != "" {
				log.Fatalf("map: config: cannot be set when data is set")
			}
			// do the abs path check for data
			if strings.TrimSpace(argsMap.paths.data) != argsMap.paths.data {
				log.Fatalf("map: data: leading or trailing spaces are not allowed\n")
			} else if path, err := abspath(argsMap.paths.data); err != nil {
				log.Fatalf("map: data: %v\n", err)
			} else if sb, err := os.Stat(path); err != nil {
				log.Fatalf("map: data: %v\n", err)
			} else if !sb.IsDir() {
				log.Fatalf("map: data: %v is not a directory\n", path)
			} else {
				argsMap.paths.data = path
			}
			// finally, update the other paths
			argsMap.paths.config = filepath.Join(argsMap.paths.data, "config.json")
		}

		if strings.TrimSpace(argsMap.paths.config) != argsMap.paths.config {
			log.Fatalf("map: config: leading or trailing spaces are not allowed\n")
		} else if path, err := filepath.Abs(argsMap.paths.config); err != nil {
			log.Printf("map: config: output: %q\n", argsMap.paths.config)
			log.Printf("map: config: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("map: config: %v\n", err)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("map: config: %v is not a file\n", path)
		} else {
			argsMap.paths.config = path
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("maps: todo: detect when a unit is created as an after-move action\n")

		log.Printf("map: config: file %s\n", argsMap.paths.config)
		cfg, err := config.Load(argsMap.paths.config)
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

		if argsMap.show.gridCoords && argsMap.show.gridNumbers {
			argsMap.show.gridNumbers = false
		}
		if argsMap.debug.sectionMaps {
			panic("this needs to be fixed")
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
					log.Printf("map: config: %s: forcing ignore\n", rptTurnId)
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
		//log.Printf("map: reports %d\n", len(allReports))

		//log.Printf("map: todo: followers are not updated after movement\n")
		//log.Printf("map: todo: hexes are not assigned for each step in the results\n")
		//log.Printf("map: todo: named hexes that are only in the status line are missed\n")
		//log.Printf("map: todo: walk the hex reports and update grid as well as ending coordinates\n")

		// create a map for every movement result we have
		var allMovementResults []*lbmoves.MovementResults

		// this hex is the clan's starting hex from the first turn report
		var originHex *wxx.Hex

		// parse the report files into a single map
		for _, rpt := range cfg.Reports {
			if rpt.Ignore {
				if argsMap.debug.showIgnoredSections {
					log.Printf("map: report %s: ignored report\n", rpt.Id)
				}
				continue
			}

			if argsMap.debug.showSectionData {
				log.Printf("map: report %s: parsing\n", rpt.Path)
			}

			// load the report file
			data, err := os.ReadFile(rpt.Path)
			if err != nil {
				log.Fatalf("map: report %s: %v", rpt.Path, err)
			}
			if argsMap.debug.showSectionData {
				log.Printf("map: report %s: loaded %8d bytes\n", rpt.Id, len(data))
			}

			// split the report into sections before parsing it
			rpt.Sections, err = reports.Sections(data, cfg.Inputs.ShowSkippedSections)
			if argsMap.debug.showSectionData {
				log.Printf("map: report %s: loaded %8d sections\n", rpt.Id, len(rpt.Sections))
			}
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
				//log.Printf("map: report %s: section %2s: need to extract grid hex data\n", rpt.Id, section.Id)

				turnId, unitId, prevGridCoords := rpt.TurnId, section.UnitId, section.PrevCoords
				missingPrevGridCoords := prevGridCoords == ""
				if missingPrevGridCoords {
					// this can happen when a unit is created as an after-move action.
					// it can also happen when a unit is created as a before-move action and doesn't move.
					log.Printf("map: report %s: section %2s: turn %s: unit %-6s: warning: missing coordinates\n", rpt.Id, section.Id, turnId, unitId)
				} else if strings.HasPrefix(prevGridCoords, "##") {
					// remove these and let the walk logic populate them
					//log.Printf("map: report %s: section %2s: turn %s: unit %-6s: stripping %q\n", rpt.Id, section.Id, turnId, unitId, prevGridCoords)
					prevGridCoords = ""
				} else {
					//log.Printf("map: report %s: section %2s: turn %s: unit %-6s: keeping %q\n", rpt.Id, section.Id, turnId, unitId, prevGridCoords)
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
				if missingPrevGridCoords {
					// todo: report on if we were able to find the previous coordinates or not
				}
			}

			if argsMap.debug.showSectionData {
				log.Printf("map: report %s: len(moves) now %8d\n", rpt.Path, len(allMovementResults))
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
				log.Fatalf("map: turn %s: unit %-8s: follows: %-8s: turn %s not found\n", mrl.TurnId, mrl.UnitId, mrl.Follows, turnUnitStepId)
			}
			theOtherUnit.Followers = append(theOtherUnit.Followers, mrl)
		}

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
			log.Fatalf("map: error: no movement results\n")
		}
		originTurnId := ""
		for _, mrl := range allMovementResults[:1] { // just the first result of the first turn
			log.Printf("map: %s: %-6s: casting about for origin\n", mrl.TurnId, mrl.UnitId)

			originTurnId = mrl.TurnId
			statusLine := mrl.StatusLine

			start := mrl.StartingGridCoordinates
			if start == "" {
				log.Fatalf("map: %s: %-6s: origin is not defined for first unit's first turn\n", mrl.TurnId, mrl.UnitId)
			}

			originGridCoords, err := coords.StringToGridCoords(start)
			if err != nil {
				log.Fatalf("map: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
			}
			log.Printf("map: %s: %-6s: status line: grid %s\n", mrl.TurnId, mrl.UnitId, originGridCoords)
			originMapCoords, err := originGridCoords.ToMapCoords()
			if err != nil {
				log.Fatalf("map: %s: %-6s: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
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
			log.Printf("map: %s: %-6s: status line: %s\n", mrl.TurnId, mrl.UnitId, statusLine.Terrain)

			// stuff the origin into our consolidated map because it may never show up again
			allHexes[originMapCoords.GridString()] = originHex
		}
		if originHex == nil {
			log.Printf("map: %s: origin is not defined\n", originTurnId)
			panic("assert(originHex != nil")
		}

		// ensure that status line is present for all turns
		for _, mrl := range allMovementResults {
			if mrl.StatusLine == nil {
				log.Fatalf("map: %s: %-6s: status line is missing\n", mrl.TurnId, mrl.UnitId)
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
					//log.Printf("map: %s: %-6s: start %s: parent %-6s: turnId %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurnUnitStepId)
					parentTurn, ok := movementResultsMap[parentTurnUnitStepId]
					if !ok {
						log.Fatalf("map: %s: %-6s: start %s: parent %-6s: turnId %q: missing\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId)
					} else if parentTurn.StartingGridCoordinates == "" {
						log.Printf("map: %s: %-6s: start %s: parent %-6s: turnId %q: start %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId, parentTurn.StartingGridCoordinates)
						panic(`assert(parentTurn.StartingGridCoordinates != "")`)
					}
					start, mrl.StartingGridCoordinates = parentTurn.StartingGridCoordinates, parentTurn.StartingGridCoordinates
					//log.Printf("map: %s: %-6s: start %s: parent %-6s: turnId %q: start %q\n", mrl.TurnId, mrl.UnitId, start, lbmoves.ParentId(mrl.UnitId), parentTurn.TurnId, parentTurn.StartingGridCoordinates)
				}
				//log.Printf("map: %s: %-6s: origin %s\n", mrl.TurnId, mrl.UnitId, start)

				currGridCoords, err := coords.StringToGridCoords(start)
				if err != nil {
					log.Fatalf("map: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
				}
				//log.Printf("map: %s: %-6s: step %2d: grid %s\n", mrl.TurnId, mrl.UnitId, n, currGridCoords)
				currMapCoords, err := currGridCoords.ToMapCoords()
				if err != nil {
					log.Fatalf("map: %s: %-6s: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
				}

				debugStep := false
				for stepNo, step := range mrl.MovementReports {
					//log.Printf("map: %s: %-6s: step %2d: mapc %s\n", mrl.TurnId, mrl.UnitId, stepNo+1, currMapCoords)
					nextCoords := walk(turnId, mrl.UnitId, allHexes, stepNo+1, step, currMapCoords, debugStep)
					//log.Printf("map: %s: %-6s: step %2d: mapc %s next %s\n", mrl.TurnId, mrl.UnitId, stepNo+1, currMapCoords, nextCoords)
					currMapCoords = nextCoords
				}

				//log.Printf("map: %s: %-6s: steps %2d start %s: curr %s\n", mrl.TurnId, mrl.UnitId, len(mrl.MovementReports), start, currMapCoords.GridString())
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
					//log.Printf("map: %s: %-6s: follower %-8q %q -> %q\n", mrl.TurnId, mrl.UnitId, follower.UnitId, follower.EndingGridCoordinates, mrl.EndingGridCoordinates)
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
						log.Fatalf("map: %s: %-6s: toGridCoords: error %v\n", mrl.TurnId, mrl.UnitId, err)
					}

					//log.Printf("map: %s: %-6s: scout %d: grid %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, startGridCoords)
					currMapCoords, err := startGridCoords.ToMapCoords()
					if err != nil {
						log.Fatalf("map: %s: %-6s: scout %d: toMapCoords: error %v\n", mrl.TurnId, mrl.UnitId, scoutNo+1, err)
					}

					debugStep := false
					movementReports := scout
					for stepNo, step := range movementReports {
						//log.Printf("map: %s: %-6s: scout %d: step %2d: mapc %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, stepNo+1, currMapCoords)
						nextCoords := walk(turnId, mrl.UnitId, allHexes, stepNo+1, step, currMapCoords, debugStep)
						//log.Printf("map: %s: %-6s: scout %d: step %2d: mapc %s next %s\n", mrl.TurnId, mrl.UnitId, scoutNo+1, stepNo+1, currMapCoords, nextCoords)
						currMapCoords = nextCoords
					}

					log.Printf("turn %-8s: unit %-8s: origin %-8s: ended %-8s\n", turnId, fmt.Sprintf("%ss%d", mrl.UnitId, scoutNo+1), startGridCoords.String(), currMapCoords.GridString())
				}
			}

			// if there are more turns to process, then the ending position this turn becomes the starting position for the next turn
			if n < len(allTurnIds)-1 {
				missingEnds, nextTurnId := 0, allTurnIds[n+1]
				//log.Printf("map: turnId %s: next turnId %s\n", turnId, nextTurnId)
				for _, mrl := range allMovementResults {
					if mrl.TurnId != turnId {
						continue
					}

					//log.Printf("map: %s: %-6s: start %s: end: %s\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates, mrl.EndingGridCoordinates)
					if mrl.EndingGridCoordinates == "" {
						log.Printf("map: %s: %-6s: start %s: end: missing\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates)
						missingEnds++
					}

					// does the unit have a move next turn? if they don't, we're hosed since we can't look forward multiple turns yet.
					nextTurnUnitStepId := fmt.Sprintf("%s.%s", nextTurnId, mrl.UnitId)
					nextTurn, ok := movementResultsMap[nextTurnUnitStepId]
					if !ok {
						log.Fatalf("map: %s: %-6s: start %s: end: missing next turn\n", mrl.TurnId, mrl.UnitId, mrl.StartingGridCoordinates)
					}
					nextTurn.StartingGridCoordinates = mrl.EndingGridCoordinates
					//log.Printf("map: %s: %-6s: start %s: end: %s <-- next turn\n", nextTurn.TurnId, nextTurn.UnitId, nextTurn.StartingGridCoordinates, nextTurn.EndingGridCoordinates)
				}
				if missingEnds != 0 {
					log.Fatalf("map: %s: error: walk did not populate %d units\n", turnId, missingEnds)
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

			if argsMap.show.gridCoords {
				consolidatedMap.AddGridCoords()
			} else if argsMap.show.gridNumbers {
				consolidatedMap.AddGridNumbering()
			}

			if argsMap.debug.sectionMaps {
				panic("this needs to be fixed")
			}

			// now we can create the Worldographer map!
			//log.Printf("map: world %6d: consolidated %6d\n", len(worldMap), len(consolidatedMap))
			mapName := filepath.Join(cfg.OutputPath, fmt.Sprintf("%s.%s.wxx", turnId, argsMap.clanId))
			if err := consolidatedMap.Create(mapName, argsMap.show.gridCenters); err != nil {
				log.Printf("map: creating %s\n", mapName)
				log.Fatal(err)
			}
			log.Printf("map: created  %s\n", mapName)
		}

		if argsMap.show.gridCoords {
			consolidatedMap.AddGridCoords()
		} else if argsMap.show.gridNumbers {
			consolidatedMap.AddGridNumbering()
		}

		// now we can create the Worldographer map!
		mapName := filepath.Join(cfg.OutputPath, fmt.Sprintf("%s.wxx", argsMap.clanId))
		if err := consolidatedMap.Create(mapName, argsMap.show.gridCenters); err != nil {
			log.Printf("map: creating %s\n", mapName)
			log.Fatal(err)
		}
		log.Printf("map: created  %s\n", mapName)

		return nil
	},
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
	//log.Printf("map: %s: %-6s: step %2d: mapc %s\n", turnId, unitId, stepNo, start)

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
	//log.Printf("map: %s: %-6s: step %2d: mapc %s: grid id %q\n", turnId, unitId, stepNo, curr, curr.GridId())
	daHex.Features.Updated = turnId

	if step.Result == lbmoves.Succeeded {
		daHex.Features.Visited = turnId

		if step.Settlement != nil {
			//log.Printf(">>> %+v\n", *step.Settlement)
			daHex.Features.Settlement = &wxx.Settlement{
				UUID: uuid.New().String(),
				Name: step.Settlement.Name,
			}
		}
	}

	if step.Result == lbmoves.Blocked {
		// log.Printf("step: blocked: attempted %s\n", step.Attempted)
		if step.BlockedBy == nil {
			panic("assert(blockedBy != nil")
		}
		// log.Printf("step: blocked: edge: %+v\n", *step.BlockedBy)
	}

	if step.Result == lbmoves.Prohibited {
		if step.ProhibitedFrom == nil {
			panic("assert(step.ProhibitedFrom != nil)")
		}
		if step.ProhibitedFrom.Direction == directions.DUnknown {
			panic(fmt.Sprintf("assert(prohibitedFrom.Direction != %d)", int(step.ProhibitedFrom.Direction)))
		}
		neighborMapCoords := curr.Add(step.ProhibitedFrom.Direction)
		//log.Printf("map: %s: %-6s: step %2d: curr %s: pro %s\n", turnId, unitId, stepNo, curr, neighborMapCoords)

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
		//log.Printf("map: %s: %-6s: step %2d: curr %s: nbr %s\n", turnId, unitId, stepNo, curr, neighborMapCoords)

		daNeighborHex := fetchHex(neighborMapCoords, neighbor.Terrain)
		if daNeighborHex.Terrain != neighbor.Terrain {
			log.Printf("map: %s: %-6s: step %2d: nbr: want %s, got %s\n", turnId, unitId, stepNo, neighbor.Terrain, daNeighborHex.Terrain)
			panic(fmt.Sprintf("assert(daNeighborHex.Terrain  == %d)", int(neighbor.Terrain)))
		}
	}

	// update the hex edges
	for _, edge := range step.Edges {
		switch edge.Edge {
		case domain.ENone:
			// ignore, shouldn't ever happen
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
	case lbmoves.Status:
	case lbmoves.Succeeded:
	case lbmoves.StayedInPlace:
	case lbmoves.Vanished:
	default:
		panic(fmt.Sprintf("assert(step.Result != %d)", int(step.Result)))
	}

	return curr
}
