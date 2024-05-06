// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var argsMap struct {
	config string // path to configuration file
	clanId string // clan id to use
	turnId string // turn id to use
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

		log.Printf("map: config: file %s\n", argsMap.config)
		var config domain.Config
		if data, err := os.ReadFile(argsMap.config); err != nil {
			log.Fatalf("map: failed to read config file: %v", err)
		} else if err = json.Unmarshal(data, &config); err != nil {
			log.Fatalf("map: failed to unmarshal config file: %v", err)
		}

		log.Printf("map: config: clan %q\n", argsMap.clanId)
		// convert yyyy-mm to year, month
		var year, month int
		var err error
		if yyyy, mm, ok := strings.Cut(argsMap.turnId, "-"); !ok {
			log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
		} else if yyyy = strings.TrimSpace(yyyy); yyyy == "" {
			log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
		} else if year, err = strconv.Atoi(yyyy); err != nil {
			log.Fatalf("map: invalid turn %q: year %v\n", argsMap.turnId, err)
		} else if mm = strings.TrimSpace(mm); mm == "" {
			log.Fatalf("map: invalid turn %q\n", argsMap.turnId)
		} else if month, err = strconv.Atoi(mm); err != nil {
			log.Fatalf("map: invalid turn %q: month %v\n", argsMap.turnId, err)
		}
		log.Printf("map: config: turn %q: year %4d month %2d\n", argsMap.turnId, year, month)

		if strings.TrimSpace(argsMap.output) != argsMap.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsMap.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsMap.output = path
		}

		// filter turns from the configuration
		var unitMoveFiles []string
		for _, rpt := range config.Reports {
			if rpt.Clan == argsMap.clanId {
				if rpt.Year < year {
					unitMoveFiles = append(unitMoveFiles, rpt.Parsed)
				} else if rpt.Year == year && rpt.Month <= month {
					unitMoveFiles = append(unitMoveFiles, rpt.Parsed)
				}
			}
		}
		log.Printf("map: files %v\n", unitMoveFiles)
		if len(unitMoveFiles) == 0 {
			log.Fatalf("map: files: no files matched constraints\n")
		}

		// load all the unit movement files
		var unitMoves []*report.Unit
		for _, path := range unitMoveFiles {
			var u []*report.Unit
			if data, err := os.ReadFile(path); err != nil {
				log.Fatalf("map: unit file %s: %v", path, err)
			} else if err = json.Unmarshal(data, &u); err != nil {
				log.Fatalf("map: unit: file %s: %v", path, err)
			}
			unitMoves = append(unitMoves, u...)
			log.Printf("map: unit: loaded %d units\n", len(u))
		}
		log.Printf("map: unit: loaded %d unit files\n", len(unitMoves))

		// sort unit moves by turn then unit
		sort.Slice(unitMoves, func(i, j int) bool {
			return unitMoves[i].SortKey() < unitMoves[j].SortKey()
		})

		// save for debugging
		for _, u := range unitMoves {
			if b, err := json.MarshalIndent(u, "", "  "); err != nil {
				log.Printf("map: unit %q: error: %v\n", u.Id, err)
			} else {
				log.Printf("map: unit %q: results\n%s\n", u.Id, string(b))
			}
		}

		// walk the consolidate unit moves, creating chain of hexes at each step
		for _, u := range unitMoves {
			start := u.Start
			if start == "" {
				log.Fatalf("map: unit %-8q: starting hex is missing\n", u.Id)
			}
			log.Printf("map: unit %-8q: origin %s\n", u.Id, start)
		}

		//m, err := maps.New(config)
		//if err != nil {
		//	log.Fatalf("map: failed to create map: %v", err)
		//}
		//log.Printf("map: input: imported %6d reports\n", len(reports))
		//log.Printf("map: input: imported %6d turns\n", len(m.Turns))
		//log.Printf("map: input: imported %6d units\n", len(m.Units))
		//log.Printf("map: input: imported %6d moves\n", len(m.Sorted.Moves))
		//log.Printf("map: input: imported %6d steps\n", len(m.Sorted.Steps))
		//
		//// maybe log origins for debugging
		//var sortedOrigins []string
		//for id := range m.Origins {
		//	sortedOrigins = append(sortedOrigins, id)
		//}
		//sort.Strings(sortedOrigins)
		//for _, id := range sortedOrigins {
		//	log.Printf("map: origin hex: unit %-8q: origin %q\n", id, m.Origins[id])
		//}
		//
		//// create chain of hexes that track movement for each unit
		//for _, unit := range m.Sorted.Units {
		//	//originHex := unit.StartingHex
		//	//if originHex == nil {
		//	//	log.Fatalf("map: unit %-8q: starting hex is missing\n", unit.Id)
		//	//}
		//	//log.Printf("map: unit %-8q: origin hex (%d, %d)\n", unit.Id, originHex.Column, originHex.Row)
		//
		//	err = m.TrackUnit(unit)
		//	if err != nil {
		//		if !errors.Is(err, cerrs.ErrTrackingGarrison) {
		//			log.Fatalf("map: unit %-8q: failed to track unit: %v\n", unit.Id, err)
		//		}
		//		log.Printf("map: unit %-8q: failed to track unit: %v\n", unit.Id, err)
		//	}
		//}
		//
		//log.Printf("map: hexes: %6d\n", len(m.Sorted.Hexes))
		//log.Printf("map: hexes: %+v\n", m.Sorted.Hexes[0])

		log.Printf("map: output %s\n", argsMap.output)

		return nil
	},
}
