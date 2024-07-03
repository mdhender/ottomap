// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/actions"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/results"
	"github.com/mdhender/ottomap/internal/terrain"
	"github.com/mdhender/ottomap/internal/turns"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var argsSammy struct {
	paths struct {
		data   string // path to data folder
		input  string // path to input folder
		output string // path to output folder
	}
	originGrid          string
	noWarnOnInvalidGrid bool
	quitOnInvalidGrid   bool
	warnOnInvalidGrid   bool
	turnId              string // maximum turn id to use
	debug               struct {
		dumpAll  bool
		maps     bool
		merge    bool
		nodes    bool
		parser   bool
		sections bool
		steps    bool
	}
}

var cmdSammy = &cobra.Command{
	Use:   "sammy",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if argsSammy.paths.data == "" {
			return fmt.Errorf("path to data folder is required")
		}

		// do the abs path check for data
		if strings.TrimSpace(argsSammy.paths.data) != argsSammy.paths.data {
			log.Fatalf("error: data: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsSammy.paths.data); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsSammy.paths.data = path
		}

		argsSammy.paths.input = filepath.Join(argsSammy.paths.data, "input")
		if path, err := abspath(argsSammy.paths.input); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsSammy.paths.input = path
		}

		argsSammy.paths.output = filepath.Join(argsSammy.paths.data, "output")
		if path, err := abspath(argsSammy.paths.output); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsSammy.paths.output = path
		}

		if len(argsSammy.originGrid) == 0 {
			// terminate on ## in location
			argsSammy.quitOnInvalidGrid = true
		} else if len(argsSammy.originGrid) != 2 {
			log.Fatalf("error: originGrid %q: must be two upper-case letters\n", argsSammy.originGrid)
		} else if strings.Trim(argsSammy.originGrid, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			log.Fatalf("error: originGrid %q: must be two upper-case letters\n", argsSammy.originGrid)
		} else {
			// don't quit when we replace ## with the location
			argsSammy.quitOnInvalidGrid = false
		}
		argsSammy.warnOnInvalidGrid = !argsSammy.noWarnOnInvalidGrid

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		argsSammy.originGrid = "RR"
		argsSammy.quitOnInvalidGrid = false
		argsSammy.warnOnInvalidGrid = true
		started := time.Now()
		log.Printf("data:   %s\n", argsSammy.paths.data)
		log.Printf("input:  %s\n", argsSammy.paths.input)
		log.Printf("output: %s\n", argsSammy.paths.output)

		inputs, err := turns.CollectInputs(argsSammy.paths.input)
		if err != nil {
			log.Fatalf("error: inputs: %v\n", err)
		}
		log.Printf("inputs: found %d turn reports\n", len(inputs))

		// allTurns holds the turn and move data and allows multiple clans to be loaded.
		allTurns := map[string][]*parser.Turn_t{}
		totalUnitMoves := 0
		for _, i := range inputs {
			started := time.Now()
			data, err := os.ReadFile(i.Path)
			if err != nil {
				log.Fatalf("error: read: %v\n", err)
			}
			turnId := fmt.Sprintf("%04d-%02d", i.Turn.Year, i.Turn.Month)
			turn, err := parser.ParseInput(i.Id, turnId, data, argsSammy.debug.parser, argsSammy.debug.sections, argsSammy.debug.steps, argsSammy.debug.nodes)
			if err != nil {
				log.Fatal(err)
			} else if turnId != fmt.Sprintf("%04d-%02d", turn.Year, turn.Month) {
				log.Fatalf("error: expected turn %q: got turn %q\n", turnId, fmt.Sprintf("%04d-%02d", turn.Year, turn.Month))
			}
			allTurns[turnId] = append(allTurns[turnId], turn)
			totalUnitMoves += len(turn.UnitMoves)
			log.Printf("%q: parsed %6d units in %v\n", i.Id, len(turn.UnitMoves), time.Since(started))
		}
		log.Printf("parsed %d inputs in to %d turns and %d units %v\n", len(inputs), len(allTurns), totalUnitMoves, time.Since(started))

		// consolidate the turns, then sort by year and month
		var consolidatedTurns []*parser.Turn_t
		foundDuplicates := false
		for _, unitTurns := range allTurns {
			if len(unitTurns) == 0 {
				// we shouldn't have any empty turns, but be safe
				continue
			}
			// create a new turn to hold the consolidated unit moves for the turn
			turn := &parser.Turn_t{
				Id:        fmt.Sprintf("%04d-%02d", unitTurns[0].Year, unitTurns[0].Month),
				Year:      unitTurns[0].Year,
				Month:     unitTurns[0].Month,
				UnitMoves: map[parser.UnitId_t]*parser.Moves_t{},
			}
			consolidatedTurns = append(consolidatedTurns, turn)

			// copy all the unit moves into this new turn, calling out duplicates
			for _, unitTurn := range unitTurns {
				for id, unitMoves := range unitTurn.UnitMoves {
					if turn.UnitMoves[id] != nil {
						foundDuplicates = true
						log.Printf("error: %s: %-6s: duplicate unit\n", turn.Id, id)
					}
					turn.UnitMoves[id] = unitMoves
					turn.SortedMoves = append(turn.SortedMoves, unitMoves)
				}
			}
		}
		if foundDuplicates {
			log.Fatalf("error: please fix the duplicate units and restart\n")
		}
		sort.Slice(consolidatedTurns, func(i, j int) bool {
			a, b := consolidatedTurns[i], consolidatedTurns[j]
			if a.Year < b.Year {
				return true
			} else if a.Year == b.Year {
				return a.Month < b.Month
			}
			return false
		})
		for _, turn := range consolidatedTurns {
			log.Printf("%s: %8d units\n", turn.Id, len(turn.UnitMoves))
			sort.Slice(turn.SortedMoves, func(i, j int) bool {
				return turn.SortedMoves[i].Id < turn.SortedMoves[j].Id
			})
		}

		// link prev and next turns
		for n, turn := range consolidatedTurns {
			if n > 0 {
				turn.Prev = consolidatedTurns[n-1]
			}
			if n+1 < len(consolidatedTurns) {
				turn.Next = consolidatedTurns[n+1]
			}
		}

		// check for N/A values in locations and quit if we find any
		naLocationCount := 0
		for _, turn := range consolidatedTurns {
			for _, unitMoves := range turn.UnitMoves {
				if unitMoves.FromHex == "N/A" {
					naLocationCount++
					log.Printf("%s: %-6s: location %q: invalid location\n", unitMoves.TurnId, unitMoves.Id, unitMoves.FromHex)
				}
			}
		}
		if naLocationCount != 0 {
			log.Fatalf("please update the invalid locations and restart\n")
		}

		// sanity check on the current and prior locations.
		badLinks, goodLinks := 0, 0
		for _, turn := range consolidatedTurns {
			if turn.Next == nil { // nothing to update
				continue
			}
			for _, unitMoves := range turn.UnitMoves {
				nextUnitMoves := turn.Next.UnitMoves[unitMoves.Id]
				if nextUnitMoves == nil {
					continue
				}
				if unitMoves.ToHex[2:] != nextUnitMoves.FromHex[2:] {
					badLinks++
					log.Printf("error: %s: %-6s: to   %q\n", turn.Id, unitMoves.Id, unitMoves.ToHex)
					log.Printf("     : %s: %-6s: from %q\n", turn.Next.Id, nextUnitMoves.Id, nextUnitMoves.FromHex)
				} else {
					goodLinks++
				}
				nextUnitMoves.FromHex = unitMoves.ToHex
			}
		}
		log.Printf("links: %d good, %d bad\n", goodLinks, badLinks)
		if badLinks != 0 {
			// this should never happen. if it does then something is wrong with the report generator.
			log.Printf("sorry: the previous and current hexes don't align in some reports\n")
			log.Fatalf("please report this error")
		}

		// proactively patch some of the obscured locations.
		// turn reports initially gave obscured locations for from and to hexes.
		// around 0902-02, the current location stopped being obscured,
		// but the previous location is still obscured.
		// NB: links between the locations must be validated before patching them!
		updatedCurrentLinks, updatedPreviousLinks := 0, 0
		for _, turn := range consolidatedTurns {
			for _, unitMoves := range turn.UnitMoves {
				var prevTurnMoves *parser.Moves_t
				if turn.Prev != nil {
					prevTurnMoves = turn.Prev.UnitMoves[unitMoves.Id]
				}
				var nextTurnMoves *parser.Moves_t
				if turn.Next != nil {
					nextTurnMoves = turn.Next.UnitMoves[unitMoves.Id]
				}
				//if unitMoves.Id == "0138" {
				//	log.Printf("this: %s: %-6s: this prior %q current %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.FromHex, unitMoves.ToHex)
				//	if prevTurnMoves != nil {
				//		log.Printf("      %s: %-6s: prev prior %q current %q\n", prevTurnMoves.TurnId, prevTurnMoves.Id, prevTurnMoves.FromHex, prevTurnMoves.ToHex)
				//	}
				//	if nextTurnMoves != nil {
				//		log.Printf("      %s: %-6s: next prior %q current %q\n", nextTurnMoves.TurnId, nextTurnMoves.Id, nextTurnMoves.FromHex, nextTurnMoves.ToHex)
				//	}
				//}

				// link prior.ToHex and this.FromHex if this.FromHex is not obscured
				if !strings.HasPrefix(unitMoves.FromHex, "##") && prevTurnMoves != nil {
					if prevTurnMoves.ToHex != unitMoves.FromHex {
						updatedPreviousLinks++
						prevTurnMoves.ToHex = unitMoves.FromHex
					}
				}

				// link this.ToHex and next.FromHex if this.ToHex is not obscured
				if !strings.HasPrefix(unitMoves.ToHex, "##") && nextTurnMoves != nil {
					if unitMoves.ToHex != nextTurnMoves.FromHex {
						updatedCurrentLinks++
						nextTurnMoves.FromHex = unitMoves.ToHex
					}
				}
			}
		}
		log.Printf("updated %8d obscured 'Previous Hex' locations\n", updatedPreviousLinks)
		log.Printf("updated %8d obscured 'Current Hex'  locations\n", updatedCurrentLinks)

		// walk the data
		err = turns.Walk(consolidatedTurns, argsSammy.originGrid, argsSammy.quitOnInvalidGrid, argsSammy.warnOnInvalidGrid, argsSammy.debug.maps)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		if argsSammy.debug.dumpAll {
			for _, turn := range consolidatedTurns {
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

		// merge the data
		reports, err := turns.MergeMoves(consolidatedTurns, argsSammy.debug.merge)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("merge returned %8d reports in %v\n", len(reports), time.Since(started))

		// map the data
		var cfg actions.MapConfig
		cfg.Clan = "0138"
		cfg.Show.Grid.Coords = true
		cfg.Show.Grid.Numbers = true
		err = actions.MapWorld(reports, cfg)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("map: %8d nodes: elapsed %v\n", len(reports), time.Since(started))

		log.Printf("elapsed: %v\n", time.Since(started))
	},
}
