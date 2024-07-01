// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/results"
	"log"
	"sort"
	"strings"
	"time"
)

// Map creates a new map or returns an error
func Map(input map[string][]*parser.Turn_t, originGrid string, quitOnInvalidGrid, warnOnInvalidGrid, debug bool) error {
	started := time.Now()
	log.Printf("map: input: %8d turns\n", len(input))

	// in some other world we might aggregate by turn, sort, and append
	var allTurnIds []string
	for turnId := range input {
		allTurnIds = append(allTurnIds, turnId)
	}
	sort.Strings(allTurnIds)
	log.Printf("map: input: %8d turns\n", len(allTurnIds))

	// sort the input by turn and then unit id
	var allUnitMoves []*parser.Moves_t
	for _, turns := range input {
		for _, turn := range turns {
			for _, moves := range turn.UnitMoves {
				allUnitMoves = append(allUnitMoves, moves)
			}
		}
	}
	sort.Slice(allUnitMoves, func(i, j int) bool {
		a, b := allUnitMoves[i], allUnitMoves[j]
		if a.TurnId < b.TurnId {
			return true
		} else if a.TurnId == b.TurnId {
			if a.Id < b.Id {
				return true
			} else if a.Id == b.Id {
				panic("duplicate unit in allUnitMoves")
			}
		}
		return false
	})
	log.Printf("map: sorted %8d steps in %v\n", len(allUnitMoves), time.Since(started))

	// first and last seen are map containing a pointer to the first or last node a unit was seen in
	firstSeen, lastSeen := map[parser.UnitId_t]*parser.Moves_t{}, map[parser.UnitId_t]*parser.Moves_t{}
	for _, moves := range allUnitMoves {
		if _, ok := firstSeen[moves.Id]; !ok {
			firstSeen[moves.Id] = moves
		}
		if _, ok := lastSeen[moves.Id]; !ok {
			lastSeen[moves.Id] = nil
		}
	}
	var allUnitIds []parser.UnitId_t
	for k := range firstSeen {
		allUnitIds = append(allUnitIds, k)
	}
	sort.Slice(allUnitIds, func(i, j int) bool {
		return allUnitIds[i] < allUnitIds[j]
	})
	log.Printf("map: hashed %8d units in %v\n", len(lastSeen), time.Since(started))

	// aggressive but annoying check for N/A values in locations
	naNodes := 0
	for _, node := range allUnitMoves {
		if node.FromHex != "N/A" {
			continue
		}
		if warnOnInvalidGrid {
			log.Printf("warning: turn %s: unit %-6s: previous hex is 'N/A'\n", node.TurnId, node.Id)
			continue
		}
		log.Printf("error: turn %s: unit %-6s: previous hex is 'N/A'\n", node.TurnId, node.Id)
		naNodes++
	}
	if naNodes == 1 {
		log.Fatalf("please update the value to the correct starting hex and restart\n")
	} else if naNodes > 1 {
		log.Fatalf("please update the values to the correct starting hex and restart\n")
	}

	followsQueue := map[parser.UnitId_t][]*parser.Moves_t{}

	// walk all the nodes in order, setting from and to hex to grid coordinates when possible.
	// note that moves must be sorted by turn then unit for the N/A and ## updates to work.
	finalDestinationMismatchCount := 0
	for _, node := range allUnitMoves {
		ls := lastSeen[node.Id]
		if node.FromHex == "N/A" {
			priorValue := node.FromHex
			pls, ok := lastSeen[node.Id.Parent()]
			if !ok {
				log.Fatalf("%s: %-6s: error: missing parent %q\n", node.TurnId, node.Id, node.Id.Parent())
			} else if pls == nil {
				log.Fatalf("%s: %-6s: error: parent %q: not yet seen\n", node.TurnId, node.Id, node.Id.Parent())
			}
			node.FromHex = pls.FromHex
			log.Printf("%s: %-6s: warning: updated %q to %q\n", node.TurnId, node.Id, priorValue, node.FromHex)
		}
		if strings.HasPrefix(node.FromHex, "##") {
			priorValue, ok := node.FromHex, false
			node.FromHex, ok = deriveGrid(node.Id, node.FromHex, ls, lastSeen[node.Id.Parent()], originGrid)
			if !ok {
				if quitOnInvalidGrid {
					log.Fatalf("%s: %-6s: error: invalid grid %q\n", node.TurnId, node.Id, priorValue)
				} else if warnOnInvalidGrid {
					log.Printf("%s: %-6s: warning: updated %q to %q\n", node.TurnId, node.Id, priorValue, node.FromHex)
				}
			}
		}

		// update the last seen turn
		lastSeen[node.Id] = node

		// walk all the moves, updating the current hex with each step
		currentHex, moved := node.FromHex, false

		if node.Follows != "" {
			if node.Follows == node.Id {
				log.Printf("error: %s: %-6s: follows %q\n", node.TurnId, node.Id, node.Follows)
				panic("assert(move.Follows != node.Id)")
			}
			ffs, ok := firstSeen[node.Follows]
			if !ok {
				log.Printf("warning: %s: %-6s: follows %q\n", node.TurnId, node.Id, node.Follows)
				log.Fatalf("error: that unit has never been seen\n")
			}
			fls, ok := lastSeen[node.Follows]
			if !ok || fls == nil {
				// this can happen if the unit is following a unit that was just created?
				log.Printf("warning: %s: %-6s: following %q: created this turn?\n", node.TurnId, node.Id, node.Follows)
				currentHex, moved = ffs.ToHex, true
			} else if fls.TurnId != node.TurnId {
				// unit we're following hasn't been seen this turn
				// this can cause us to move the unit to the wrong hex
				log.Printf("warning: %s: %-6s: following %q: hasn't been seen this turn\n", node.TurnId, node.Id, node.Follows)
				currentHex, moved = fls.ToHex, true
			} else {
				// unit we're following has been seen this turn, so follow it
				currentHex, moved = fls.ToHex, true
			}
			for _, move := range node.Moves {
				move.CurrentHex = currentHex
			}
		} else if node.GoesTo != "" {
			// sanity check, these should always be the same value
			if node.GoesTo != node.ToHex {
				log.Printf("turn %s: unit %-6s: current hex is %q\n", node.TurnId, node.Id, node.ToHex)
				log.Printf("turn %s: unit %-6s: goes to hex is %q\n", node.TurnId, node.Id, node.GoesTo)
				log.Fatalf("error: current hex != goes to hex\n")
			}
			currentHex, moved = node.GoesTo, true
			for _, move := range node.Moves {
				move.CurrentHex = currentHex
			}
		} else {
			for _, move := range node.Moves {
				var nextHex string
				// update the current hex if the unit successfully moved to another hex
				if move.Still {
					// stays in the current hex, so nothing to update
					moved, nextHex = true, currentHex
				} else if move.GoesTo != "" {
					// took care of this above
					moved, nextHex = true, currentHex
				} else if move.Follows != "" {
					// took care of this above
					moved, nextHex = true, currentHex
				} else if move.Result == results.Succeeded {
					// update current hex based on the direction
					moved, nextHex = true, coords.Move(currentHex, move.Advance)
					//log.Printf("curr %s + %-2s == %q\n", currentHex, move.Advance, nextHex)
				} else if move.Result == results.Failed {
					// nothing changes
					moved, nextHex = true, currentHex
				} else {
					log.Printf("turn %s: unit %-6s: line %d: step %d: result %q\n", node.TurnId, node.Id, move.LineNo, move.StepNo, move.Result)
					panic(fmt.Sprintf("assert(result != %q)", move.Result))
				}
				currentHex, move.CurrentHex = nextHex, nextHex
			}
		}
		// sanity check that the calculated hex matches the Current Hex from the report
		var finalDestinationMatches bool
		if strings.HasPrefix(node.ToHex, "##") {
			finalDestinationMatches = node.ToHex[2:] == currentHex[2:]
		} else {
			finalDestinationMatches = node.ToHex == currentHex
		}
		if !finalDestinationMatches {
			finalDestinationMismatchCount++
			log.Printf("turn %s: unit %-6s: toHex %q: currentHex %q\n", node.TurnId, node.Id, node.ToHex, currentHex)
		}

		// if we moved AND we have followers we must update their locations
		if moved {
			var updateFollowers func(id parser.UnitId_t)
			updateFollowers = func(id parser.UnitId_t) {
				if followers, ok := followsQueue[node.Id]; ok {
					delete(followsQueue, id) // delete this node from the queue
					// update this node's followers
					for _, follower := range followers {
						follower.ToHex = node.ToHex
						// recursively handle follower of a follower
						updateFollowers(follower.Id)
					}
				}
			}
			updateFollowers(node.Id)
		}

		if strings.HasPrefix(node.ToHex, "##") {
			node.ToHex = currentHex[:2] + node.ToHex[2:]
		}
	}

	if finalDestinationMismatchCount != 0 {
		log.Printf("error: there were %d times that we moved a unit to the wrong place\n", finalDestinationMismatchCount)
		log.Fatalf("please report this error")
	}

	//for _, node := range allUnitMoves {
	//	if !node.Id.IsFleet() {
	//		continue
	//	}
	//	for _, move := range node.Moves {
	//		if move.Report == nil {
	//			log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, node.Id, move.LineNo, move.StepNo, move.CurrentHex, "missing report!")
	//		} else if move.Report.Terrain == terrain.Blank {
	//			if move.Result == results.Failed {
	//				log.Printf("%s: %-6s: %s: failed\n", move.TurnId, node.Id, move.CurrentHex)
	//			} else if move.Still {
	//				log.Printf("%s: %-6s: %s: stayed in place\n", move.TurnId, node.Id, move.CurrentHex)
	//			} else if move.Follows != "" {
	//				log.Printf("%s: %-6s: %s: follows %s\n", move.TurnId, node.Id, move.CurrentHex, move.Follows)
	//			} else if move.GoesTo != "" {
	//				log.Printf("%s: %-6s: %s: goes to %s\n", move.TurnId, node.Id, move.CurrentHex, move.GoesTo)
	//			} else {
	//				log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, node.Id, move.LineNo, move.StepNo, move.CurrentHex, "missing terrain")
	//			}
	//		} else {
	//			log.Printf("%s: %-6s: %s: terrain %s\n", move.TurnId, node.Id, move.CurrentHex, move.Report.Terrain)
	//		}
	//	}
	//}

	log.Printf("map: input: %8d nodes: elapsed %v\n", len(input), time.Since(started))
	return nil
}

func deriveGrid(id parser.UnitId_t, hex string, priorTurn, parent *parser.Moves_t, originGrid string) (string, bool) {
	//debug := id == "0138e1"
	//if debug {
	//	log.Printf("hex %q priorTurn %v parent %v, origin %q\n", hex, priorTurn, parent, originGrid)
	//}
	if !strings.HasPrefix(hex, "##") {
		return hex, true
	}
	// do we have a prior turn to reference?
	if priorTurn != nil {
		return priorTurn.ToHex, true
	}
	// do we have a parent in this turn that we can reference?
	if parent != nil {
		for _, parentHex := range []string{parent.ToHex, parent.FromHex} {
			if strings.HasPrefix(parentHex, "##") { // grid is not assigned
				continue
			} else if parentHex[2:] != hex[2:] { // not the same digits for the hex
				continue
			}
			return parentHex[:2] + hex[2:], true
		}
	}
	return originGrid + hex[2:], false
}
