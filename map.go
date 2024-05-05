// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/maps"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"strings"
)

var argsMap struct {
	input  string // parsed report file to process
	output string // path to create map in
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

		var reports domain.Reports
		if data, err := os.ReadFile(argsMap.input); err != nil {
			log.Fatalf("map: failed to read input file: %v", err)
		} else if err = json.Unmarshal(data, &reports); err != nil {
			log.Fatalf("map: failed to unmarshal input file: %v", err)
		}

		if strings.TrimSpace(argsMap.output) != argsMap.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsMap.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsMap.output = path
		}

		log.Printf("map: input: file %s\n", argsMap.input)

		m, err := maps.New(reports)
		if err != nil {
			log.Fatalf("map: failed to create map: %v", err)
		}
		log.Printf("map: input: imported %6d reports\n", len(reports))
		log.Printf("map: input: imported %6d turns\n", len(m.Turns))
		log.Printf("map: input: imported %6d units\n", len(m.Units))
		log.Printf("map: input: imported %6d moves\n", len(m.Sorted.Moves))
		log.Printf("map: input: imported %6d steps\n", len(m.Sorted.Steps))

		// maybe log origins for debugging
		var sortedOrigins []string
		for id := range m.Origins {
			sortedOrigins = append(sortedOrigins, id)
		}
		sort.Strings(sortedOrigins)
		for _, id := range sortedOrigins {
			log.Printf("map: origin hex: unit %-8q: origin %q\n", id, m.Origins[id])
		}

		// create chain of hexes that track movement for each unit
		for _, unit := range m.Sorted.Units {
			//originHex := unit.StartingHex
			//if originHex == nil {
			//	log.Fatalf("map: unit %-8q: starting hex is missing\n", unit.Id)
			//}
			//log.Printf("map: unit %-8q: origin hex (%d, %d)\n", unit.Id, originHex.Column, originHex.Row)

			err = m.TrackUnit(unit)
			if err != nil {
				if !errors.Is(err, cerrs.ErrTrackingGarrison) {
					log.Fatalf("map: unit %-8q: failed to track unit: %v\n", unit.Id, err)
				}
				log.Printf("map: unit %-8q: failed to track unit: %v\n", unit.Id, err)
			}
		}

		log.Printf("map: hexes: %6d\n", len(m.Sorted.Hexes))
		log.Printf("map: hexes: %+v\n", m.Sorted.Hexes[0])

		log.Printf("map: output %s\n", argsMap.output)

		return nil
	},
}
