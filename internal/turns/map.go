// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"github.com/mdhender/ottomap/actions"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/results"
	"github.com/mdhender/ottomap/internal/terrain"
	"log"
	"time"
)

// Map creates a new map or returns an error.
// You will get errors if the input is not sorted by turn.
func Map(input []*parser.Turn_t, cfg actions.MapConfig) error {
	started := time.Now()
	log.Printf("map: %8d turns\n", len(input))

	if cfg.Dump.All {
		for _, turn := range input {
			for _, unit := range turn.SortedMoves {
				for _, move := range unit.Moves {
					if move.Report == nil {
						log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, unit.Id, move.LineNo, move.StepNo, move.CurrentHex, "missing report!")
					} else if move.Report.Terrain == terrain.Blank {
						if move.Result == results.Failed {
							log.Printf("%s: %-6s: %s: failed\n", move.TurnId, unit.Id, move.CurrentHex)
						} else if move.Still {
							log.Printf("%s: %-6s: %s: stayed in place\n", move.TurnId, unit.Id, move.CurrentHex)
						} else if move.Follows != "" {
							log.Printf("%s: %-6s: %s: follows %s\n", move.TurnId, unit.Id, move.CurrentHex, move.Follows)
						} else if move.GoesTo != "" {
							log.Printf("%s: %-6s: %s: goes to %s\n", move.TurnId, unit.Id, move.CurrentHex, move.GoesTo)
						} else {
							log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, unit.Id, move.LineNo, move.StepNo, move.CurrentHex, "missing terrain")
						}
					} else {
						log.Printf("%s: %-6s: %s: terrain %s\n", move.TurnId, unit.Id, move.CurrentHex, move.Report.Terrain)
					}
					for _, border := range move.Report.Borders {
						if border.Edge != edges.None {
							log.Printf("%s: %-6s: %s: border  %-14s %q\n", move.TurnId, unit.Id, move.CurrentHex, border.Direction, border.Edge)
						}
						if border.Terrain != terrain.Blank {
							log.Printf("%s: %-6s: %s: border  %-14s %q\n", move.TurnId, unit.Id, move.CurrentHex, border.Direction, border.Terrain)
						}
					}
					for _, point := range move.Report.FarHorizons {
						log.Printf("%s: %-6s: %s: compass %-14s sighted %q\n", move.TurnId, unit.Id, move.CurrentHex, point.Point, point.Terrain)
					}
					for _, settlement := range move.Report.Settlements {
						log.Printf("%s: %-6s: %s: village %q\n", move.TurnId, unit.Id, move.CurrentHex, settlement.Name)
					}
				}
			}
		}
	}

	log.Printf("map: %8d nodes: elapsed %v\n", len(input), time.Since(started))
	return nil
}
