// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/results"
	"log"
	"strings"
	"time"
)

// Walk visits every unit move and updates the location.
// You will get errors if the input is not sorted by turn.
func Walk(input []*parser.Turn_t, originGrid string, quitOnInvalidGrid, warnOnInvalidGrid, debug bool) error {
	started := time.Now()
	log.Printf("walk: input: %8d turns\n", len(input))

	// first and last seen are map containing a pointer to the first or last move a unit was seen in
	firstSeen, lastSeen := map[parser.UnitId_t]*parser.Moves_t{}, map[parser.UnitId_t]*parser.Moves_t{}

	// loop through the turns in order, walking each unit move and updating locations
	finalDestinationMismatchCount := 0
	for _, turn := range input {
		followsQueue := map[parser.UnitId_t][]*parser.Moves_t{}

		// walk the moves for all the units in this turn
		for _, unitMoves := range turn.SortedMoves {
			id := unitMoves.Id
			if _, ok := firstSeen[id]; !ok {
				firstSeen[id] = unitMoves
			}
			var priorMoves *parser.Moves_t
			if turn.Prev != nil {
				priorMoves = turn.Prev.UnitMoves[id]
			}
			var nextMoves *parser.Moves_t
			if turn.Next != nil {
				nextMoves = turn.Next.UnitMoves[id]
			}
			_ = nextMoves

			// fetch the parent in case we have to derive a starting location
			parentMoves := turn.UnitMoves[id.Parent()]
			if parentMoves == nil {
				log.Printf("%s: %-6s: parent %q: missing\n", unitMoves.TurnId, unitMoves.Id, unitMoves.Id.Parent())
				log.Fatalf("error: expected unit to have parent\n")
			}

			// if the starting location is obscured, try to derive it.
			if strings.HasPrefix(unitMoves.FromHex, "##") {
				//log.Printf("%s: %-6s: prev %q: curr %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.FromHex, unitMoves.ToHex)

				// it should be the same as the ending location on the prior turn.
				if priorMoves != nil {
					log.Printf("%s: %-6s: prev %q: curr %q (from priorMoves)\n", priorMoves.TurnId, priorMoves.Id, priorMoves.FromHex, priorMoves.ToHex)
					if priorMoves.ToHex[2:] != unitMoves.FromHex[2:] {
						// the digits of the location don't match; this is a problem with the report
						panic("bad location link")
					}
				}

				// failing that, it could be parent's starting or ending location this turn.
				// (there's no way to know which it is, so we play the odds and prefer the starting location).
				// if we can't, either warn the user or quit (there's a command line flag to choose)
				priorValue, ok := unitMoves.FromHex, false
				unitMoves.FromHex, ok = deriveGrid(unitMoves.Id, unitMoves.FromHex, priorMoves, parentMoves, originGrid)
				if !ok {
					if quitOnInvalidGrid {
						log.Fatalf("%s: %-6s: error: invalid grid %q\n", unitMoves.TurnId, unitMoves.Id, priorValue)
					} else if warnOnInvalidGrid {
						log.Printf("%s: %-6s: warning: updated %q to %q\n", unitMoves.TurnId, unitMoves.Id, priorValue, unitMoves.FromHex)
					}
				}
			}

			// sanity check: the unit's starting location should not be obscured
			if strings.HasPrefix(unitMoves.FromHex, "##") {
				panic("previous hex is obscured")
			}

			// walk all the moves, updating the current hex with each step
			currentHex := unitMoves.FromHex

			// sanity check: the current location should not be obscured
			if strings.HasPrefix(currentHex, "##") {
				panic("current hex is obscured")
			}

			//if unitMoves.Id == "2138" && turn.Id == "0900-02" {
			//	log.Printf("%s: %-6s: follows %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.Follows)
			//}
			if unitMoves.Follows != "" {
				// unit is following another unit, so update the location to the leader's location
				if unitMoves.Follows == unitMoves.Id {
					log.Printf("error: %s: %-6s: follows %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.Follows)
					panic("assert(move.Follows != unitMoves.Id)")
				}
				resolvedMovement := false
				// if the to hex isn't obscured, just use it.
				if !strings.HasPrefix(unitMoves.ToHex, "##") {
					currentHex = unitMoves.ToHex
					for _, move := range unitMoves.Moves {
						move.CurrentHex = currentHex
					}
					resolvedMovement = true
				} else {
					// jump through several hoops to try and figure out where we're headed
					leaderId := unitMoves.Follows
					leaderMoves, ok := turn.UnitMoves[leaderId]
					if !ok {
						log.Printf("%s: %-6s: follows %q\n", unitMoves.TurnId, unitMoves.Id, leaderId)
						log.Printf("%s: %-6s: missing from turn\n", unitMoves.TurnId, leaderId)
						log.Fatalf("error: unable to find leader this turn\n")
					}
					// if the leader's location isn't obscured, just use it.
					if !strings.HasPrefix(leaderMoves.ToHex, "##") {
						// move the follower to the leader's location
						currentHex = leaderMoves.ToHex
						for _, move := range unitMoves.Moves {
							move.CurrentHex = currentHex
						}
						resolvedMovement = true
					} else if leaderMoves.GoesTo != "" && !strings.HasPrefix(leaderMoves.GoesTo, "##") {
						// we caught a break - the leader has an un-obscured goes to line,
						// so we can move the follower to the location the leader is going to.
						currentHex = leaderMoves.GoesTo
						for _, move := range unitMoves.Moves {
							move.CurrentHex = currentHex
						}
						resolvedMovement = true
					} else {
						// the leader's location is obscured.
						// if we're lucky, it's because it hasn't moved yet.
						// if we're not lucky, it's because it was created in the current turn.
						// either way, queue the request until it does move.
						if leaderMoves.Follows != "" {
							log.Printf("%s: %-6s: follows %q: follows %q\n", unitMoves.TurnId, unitMoves.Id, leaderId, leaderMoves.Follows)
						}
						followsQueue[leaderId] = append(followsQueue[leaderId], unitMoves)
						// movement is still pending
						resolvedMovement = false
					}
				}
				if resolvedMovement {
					// todo: is it safe to assume that we visited this location?
					for _, move := range unitMoves.Moves {
						move.Report.WasVisited = true
					}
				}
			} else if unitMoves.GoesTo != "" {
				// unit is going to a specific location, so update the location to that location
				// sanity check, these should always be the same value
				if unitMoves.GoesTo != unitMoves.ToHex {
					log.Printf("turn %s: unit %-6s: current hex is %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.ToHex)
					log.Printf("turn %s: unit %-6s: goes to hex is %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.GoesTo)
					log.Fatalf("error: current hex != goes to hex\n")
				}
				currentHex = unitMoves.GoesTo
				for _, move := range unitMoves.Moves {
					move.CurrentHex = currentHex
					// todo: is it safe to assume that we visited this location?
					move.Report.WasVisited = true
				}
			} else {
				// unit is moving by itself
				// todo: the hex the unit starts in should always be flagged as visited
				for _, move := range unitMoves.Moves {
					var nextHex string
					// update the current hex if the unit successfully moved to another hex
					if move.Still {
						// stays in the current hex, so nothing to update
						nextHex = currentHex
						move.Report.WasVisited = true
					} else if move.GoesTo != "" {
						// took care of this above
						nextHex = currentHex
					} else if move.Follows != "" {
						// took care of this above
						nextHex = currentHex
					} else if move.Result == results.Succeeded {
						// update current hex based on the direction
						nextHex = coords.Move(currentHex, move.Advance)
						//log.Printf("curr %s + %-2s == %q\n", currentHex, move.Advance, nextHex)
					} else if move.Result == results.Failed {
						// nothing changes
						nextHex = currentHex
					} else {
						log.Printf("%s: %-6s: %d: step %d: result %q\n", unitMoves.TurnId, unitMoves.Id, move.LineNo, move.StepNo, move.Result)
						panic(fmt.Sprintf("assert(result != %q)", move.Result))
					}
					//if unitMoves.Id == "0138" {
					//	log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", unitMoves.TurnId, unitMoves.Id, move.LineNo, move.StepNo, move.Result, nextHex)
					//}
					currentHex, move.CurrentHex = nextHex, nextHex
					// todo: how do we update the last location in the move?
				}
			}

			// sanity check that the calculated hex matches the Current Hex from the report
			var finalDestinationMatches bool
			if strings.HasPrefix(unitMoves.ToHex, "##") {
				finalDestinationMatches = unitMoves.ToHex[2:] == currentHex[2:]
			} else {
				finalDestinationMatches = unitMoves.ToHex == currentHex
			}
			if !finalDestinationMatches {
				finalDestinationMismatchCount++
				log.Printf("%s: %-6s: toHex %q: currentHex %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.ToHex, currentHex)
			}

			// do not forget to update the move's final hex
			unitMoves.ToHex = currentHex

			// if the final destination matches, let's update the starting position for the next turn
			if finalDestinationMatches && nextMoves != nil {
				if unitMoves.ToHex[2:] != nextMoves.FromHex[2:] {
					// we don't match
					log.Printf("%s: %-6s: ending at     %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.ToHex)
					log.Printf("%s: %-6s: starting from %q\n", nextMoves.TurnId, nextMoves.Id, nextMoves.FromHex)
					log.Fatalf("error: unable to align location between turns\n")
				}
				nextMoves.FromHex = unitMoves.ToHex
			}

			// walk all the scout moves
			for _, scout := range unitMoves.Scouts {
				// every scout starts in the hex their parent end up in
				currentHex = unitMoves.ToHex
				for _, move := range scout.Moves {
					// update the current hex if the unit successfully moved to another hex
					if move.Still {
						// nothing changes
						move.CurrentHex = currentHex
						move.Report.WasVisited, move.Report.WasScouted = true, true
					} else if move.GoesTo != "" {
						panic("scouts are not allowed to teleport")
					} else if move.Follows != "" {
						panic("scouts are not allowed to follow")
					} else if move.Result == results.Succeeded {
						// update current hex based on the direction
						move.CurrentHex = coords.Move(currentHex, move.Advance)
						move.Report.WasVisited, move.Report.WasScouted = true, true
					} else if move.Result == results.Failed {
						// nothing changes
						move.CurrentHex = currentHex
						// todo: can we update the visited and scouted flags?
					} else {
						log.Printf("%s: %-6s: %d: step %d: result %q\n", unitMoves.TurnId, unitMoves.Id, move.LineNo, move.StepNo, string(move.Line))
						log.Printf("%s: %-6s: %d: step %d: result %q\n", unitMoves.TurnId, unitMoves.Id, move.LineNo, move.StepNo, move.Result)
						panic(fmt.Sprintf("assert(result != %q)", move.Result))
					}
					currentHex = move.CurrentHex
					// todo: how do we update the last location in the move?
				}
			}

			// update when the unit was last seen
			lastSeen[id] = unitMoves
		}

		for maxFollowLevels := 9; len(followsQueue) != 0 && maxFollowLevels != 0; maxFollowLevels-- {
			for leaderId, followers := range followsQueue {
				leaderMoves, ok := turn.UnitMoves[leaderId]
				if !ok {
					panic("assert(leader is ok)")
				} else if strings.HasPrefix(leaderMoves.ToHex, "##") {
					// leader's location is still obscured, so move on.
					// we'll try again in the next loop.
					continue
				}
				for _, unitMoves := range followers {
					log.Printf("%s: leader %-6s: %s: follower %q\n", leaderMoves.TurnId, leaderMoves.Id, leaderMoves.ToHex, unitMoves.Id)
					// move the follower to the leader's location
					unitMoves.ToHex = leaderMoves.ToHex
					for _, move := range unitMoves.Moves {
						move.CurrentHex = leaderMoves.ToHex
					}
				}
				delete(followsQueue, leaderId)
			}
		}
		if len(followsQueue) != 0 {
			for _, followers := range followsQueue {
				for _, moves := range followers {
					log.Printf("%s: %q follows %q\n", turn.Id, moves.Id, moves.Follows)
				}
			}
			log.Fatalf("error: %s: did not clear followers queue\n", turn.Id)
		}

		// sanity check: no locations should be obscured
		obscuredLocations := 0
		for _, unitMoves := range turn.SortedMoves {
			if strings.HasPrefix(unitMoves.FromHex, "##") || strings.HasPrefix(unitMoves.ToHex, "##") {
				obscuredLocations++
				log.Printf("%s: %-6s: fromHex  %q toHex %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.FromHex, unitMoves.ToHex)
			}
		}
		if obscuredLocations != 0 {
			panic("location contains obscured hexes")
		}

	}

	if finalDestinationMismatchCount != 0 {
		log.Printf("error: there were %d times that we moved a unit to the wrong place\n", finalDestinationMismatchCount)
		log.Fatalf("please report this error")
	}

	log.Printf("walk: %8d nodes: elapsed %v\n", len(input), time.Since(started))
	return nil
}
