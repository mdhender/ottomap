// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/mdhender/ottomap/wxx"
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

		//if data, err := os.ReadFile(filepath.Join("input", "899-12.0138.xml")); err != nil {
		//	log.Fatal(err)
		//} else if len(data)%2 != 0 { // verify the length
		//	log.Fatalf("UTF-16 data must contain an even number of bytes")
		//} else if !bytes.HasPrefix(data, []byte{0xfe, 0xff}) { // verify the BOM
		//	log.Fatalf("UTF-16 data must start with a BOM")
		//} else {
		//	// consume the bom
		//	data = data[2:]
		//	// convert the slice of byte to a slice of uint16
		//	chars := make([]uint16, len(data)/2)
		//	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, &chars); err != nil {
		//		log.Fatal(err)
		//	}
		//	// create a buffer for the results
		//	dst := bytes.Buffer{}
		//	// convert the UTF-16 to runes, then to UTF-8 bytes
		//	var utfBuffer [utf8.UTFMax]byte
		//	for _, r := range utf16.Decode(chars) {
		//		utf8Size := utf8.EncodeRune(utfBuffer[:], r)
		//		dst.Write(utfBuffer[:utf8Size])
		//	}
		//	log.Printf("maps: todo: verify that the map is valid XML %d %d\n", len(data), len(dst.Bytes()))
		//	out := &bytes.Buffer{}
		//	for _, line := range bytes.Split(dst.Bytes(), []byte{'\n'}) {
		//		out.WriteString(fmt.Sprintf("\tw.println(`%s`)\n", string(line)))
		//	}
		//	if err := os.WriteFile("working/println.go.txt", out.Bytes(), 0644); err != nil {
		//		log.Fatal(err)
		//	}
		//}

		hexes := []wxx.Hex{
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 8}, Terrain: domain.TPrairie},
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 7}, Terrain: domain.TOcean},
			{Grid: "OO", Coords: wxx.Offset{Column: 11, Row: 6}, Terrain: domain.TSwamp},
			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 7}, Terrain: domain.TRockyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 10, Row: 6}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 6}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 9, Row: 5}, Terrain: domain.TGrassyHills},
			{Grid: "OO", Coords: wxx.Offset{Column: 8, Row: 4}, Terrain: domain.TPrairie},
		}

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

		// unitNode is a unit that will be added to the map.
		// it will contain all the unit's moves. the parent
		// link is included to help with linking moves together.
		type unitNode struct {
			Id     string
			Parent *unitNode
			Moves  []*report.Unit
		}

		// create a map of all units to help with linking moves together.
		allUnits := map[string]*unitNode{}
		for _, u := range unitMoves {
			un, ok := allUnits[u.Id]
			if !ok {
				un = &unitNode{Id: u.Id}
				if !u.IsClan() {
					un.Parent, ok = allUnits[u.ParentId]
					if !ok {
						log.Fatalf("map: unit %q: parent %q not found\n", u.Id, u.ParentId)
					}
				}
				allUnits[u.Id] = un
			}
			un.Moves = append(un.Moves, u)
		}

		// create a sorted list of all units for dealing with parent/child relationships.
		var sortedNodes []*unitNode
		for _, un := range allUnits {
			sortedNodes = append(sortedNodes, un)
		}
		sort.Slice(sortedNodes, func(i, j int) bool {
			return sortedNodes[i].Id < sortedNodes[j].Id
		})

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
		for _, u := range allUnits {
			// Iterate over each unit's moves.
			for _, cm := range u.Moves {
				// If the current move doesn't follow another unit, skip it
				if cm.FollowsId == "" {
					continue
				}

				// Look up the unit being followed using the FollowsId.
				// If the unit being followed is not found, log a fatal error
				fu, ok := allUnits[cm.FollowsId]
				if !ok {
					log.Fatalf("map: unit %q: follower %q not found\n", u.Id, cm.FollowsId)
				}

				// Iterate over the moves of the unit being followed to find the move matching the current turn.
				for _, fm := range fu.Moves {
					// Check if the move of the followed unit matches the current turn
					if fm.Turn.Year == cm.Turn.Year && fm.Turn.Month == cm.Turn.Month {
						// If a matching move is found, link the current move to the followed unit's move
						cm.Follows = fm
						break
					}
				}

				// If no matching move is found for the followed unit in the current turn, log a fatal error
				if cm.Follows == nil {
					log.Fatalf("map: unit %q: follower %q: turn %04d-%-2d not found\n", u.Id, cm.FollowsId, cm.Turn.Year, cm.Turn.Month)
				}
			}
		}

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

		// resolve "follows" links
		for _, u := range unitMoves {
			if u.Follows == nil {
				continue
			}
			u.End = u.Follows.End
		}

		// final sanity check for ending positions
		for _, u := range unitMoves {
			if u.End == "" {
				log.Fatalf("map: unit %-8q: turn %04d-%02d: ending hex is missing\n", u.Id, u.Turn.Year, u.Turn.Month)
			}
		}

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
