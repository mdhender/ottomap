// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"log"
	"strings"
	"time"
)

// Map creates a new map or returns an error
func Map(input []*ParseResults_t, debug bool) error {
	started := time.Now()
	log.Printf("map: input: %8d nodes\n", len(input))

	// last seen is a map containing a pointer to the last node a unit was seen in
	lastSeen := map[string]*ParseResults_t{}

	for _, node := range input {
		if node.PrevCoords == "N/A" {
			// do something
		}
		if strings.HasPrefix(node.PrevCoords, "##") {
			pc, ok := lastSeen[node.Id]
			if !ok {
				log.Fatalf("map: %s: %s: %d: prev coords %q\n", node.TurnReportId, node.Id, node.LineNo, node.PrevCoords)
			}
			_ = pc
		}
		lastSeen[node.Id] = node
		log.Printf("map: %s: %s: %d: prev coords %q\n", node.TurnReportId, node.Id, node.LineNo, node.PrevCoords)
	}

	log.Printf("map: input: %8d nodes: elapsed %v\n", len(input), time.Since(started))
	return nil
}
