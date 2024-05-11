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
	"strconv"
	"strings"
)

var argsMap struct {
	config string // path to configuration file
	clanId string // clan id to use
	turnId string // turn id to use
	debug  struct {
		units bool
	}
}

var cmdMap = &cobra.Command{
	Use:   "map",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("maps: todo: detect when a unit is created as an after-move action\n")

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

		var allSteps []*lbmoves.Step

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

			// parse the report, stopping if there's an error
			for _, section := range rpt.Sections {
				if section.FollowsLine != nil {
					log.Printf("map: report %s: section %2s: follows %q\n", rpt.Id, section.Id, section.FollowsLine)
					if steps, err := lbmoves.ParseMoveResults(section.FollowsLine); err != nil {
						log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
					} else {
						for _, step := range steps {
							allSteps = append(allSteps, step)
						}
					}
				}
				if section.MovementLine != nil {
					log.Printf("map: report %s: section %2s: moves   %q\n", rpt.Id, section.Id, section.MovementLine)
					if steps, err := lbmoves.ParseMoveResults(section.MovementLine); err != nil {
						log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
					} else {
						for _, step := range steps {
							allSteps = append(allSteps, step)
						}
					}
				}
				for _, scoutLine := range section.ScoutLines {
					if scoutLine != nil {
						log.Printf("map: report %s: section %2s: scouts  %q\n", rpt.Id, section.Id, scoutLine)
						if steps, err := lbmoves.ParseMoveResults(scoutLine); err != nil {
							log.Fatalf("map: report %s: section %2s: %v\n", rpt.Id, section.Id, err)
						} else {
							for _, step := range steps {
								allSteps = append(allSteps, step)
							}
						}
					}
				}
			}
		}

		log.Printf("map: todo: maybe status line can be mapped like a step\n")
		log.Printf("map: todo: named hexes that are only in the status line are missed\n")

		if cfg.Inputs.ShowSteps {
			for _, us := range allSteps {
				boo, err := json.MarshalIndent(us, "", "\t")
				if err != nil {
					log.Fatalf("map: step: %v\n", err)
				}
				fmt.Printf("step: %s\n", string(boo))
			}
		}

		//// sort unit moves by turn then unit
		//sort.Slice(unitMoves, func(i, j int) bool {
		//	return unitMoves[i].SortKey() < unitMoves[j].SortKey()
		//})

		// unitNode is a unit that will be added to the map.
		// it will contain all the unit's moves. the parent
		// link is included to help with linking moves together.
		type unitNode struct {
			Id     string
			Parent *unitNode
			Moves  []*report.Unit
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
		//// walk the consolidate unit moves, creating chain of hexes at each step
		//for _, u := range unitMoves {
		//	start := u.Start
		//	if start == "" {
		//		log.Fatalf("map: unit %-8q: starting hex is missing\n", u.Id)
		//	}
		//	log.Printf("map: unit %-8q: origin %s\n", u.Id, start)
		//}
		//
		//// resolve "follows" links
		//for _, u := range unitMoves {
		//	if u.Follows == nil {
		//		continue
		//	}
		//	u.End = u.Follows.End
		//}
		//
		//// final sanity check for ending positions
		//for _, u := range unitMoves {
		//	if u.End == "" {
		//		log.Fatalf("map: unit %-8q: turn %04d-%02d: ending hex is missing\n", u.Id, u.Turn.Year, u.Turn.Month)
		//	}
		//}

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
