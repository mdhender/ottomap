// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"github.com/mdhender/ottomap/internal/parser"
	"log"
	"sort"
	"time"
)

// Map creates a new map or returns an error
func Map(input []*parser.Movement_t, debug bool) error {
	started := time.Now()
	log.Printf("map: input: %8d moves\n", len(input))

	var allSteps []*parser.Step_t
	for _, move := range input {
		if len(move.Steps) != 0 {
			allSteps = append(allSteps, move.Steps...)
		}
	}
	log.Printf("map: input: %8d steps\n", len(allSteps))

	// sort the input by turn and then unit id
	sort.Slice(allSteps, func(i, j int) bool {
		a, b := allSteps[i], allSteps[j]
		if a.TurnId < b.TurnId {
			return true
		} else if a.TurnId == b.TurnId {
			if a.UnitId < b.UnitId {
				return true
			} else if a.UnitId == b.UnitId {
				return a.No < b.No
			}
		}
		return false
	})
	log.Printf("map: sorted %8d steps in %v\n", len(allSteps), time.Since(started))

	// last seen is a map containing a pointer to the last node a unit was seen in
	lastSeen := map[parser.UnitId_t]*parser.Step_t{}
	for _, step := range allSteps {
		if _, ok := lastSeen[step.UnitId]; !ok {
			lastSeen[step.UnitId] = nil
		}
	}
	log.Printf("map: hashed %8d units in %v\n", len(lastSeen), time.Since(started))


	//for _, node := range input {
	//	if node.PrevCoords == "N/A" {
	//		// do something
	//	}
	//	if strings.HasPrefix(node.PrevCoords, "##") {
	//		pc, ok := lastSeen[node.Id]
	//		if !ok {
	//			log.Fatalf("map: %s: %s: %d: prev coords %q\n", node.TurnReportId, node.Id, node.LineNo, node.PrevCoords)
	//		}
	//		_ = pc
	//	}
	//	lastSeen[node.Id] = node
	//	log.Printf("map: %s: %s: %d: prev coords %q\n", node.TurnReportId, node.Id, node.LineNo, node.PrevCoords)
	//}

	log.Printf("map: input: %8d nodes: elapsed %v\n", len(input), time.Since(started))
	return nil
}
