// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/hexes"
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

		log.Printf("map: input  %s\n", argsMap.input)
		log.Printf("map: input  %d records\n", len(reports))

		// force reports to be sorted by date then clan
		sort.Sort(reports)
		log.Printf("map: input  sorted %d records\n", len(reports))
		for _, rpt := range reports {
			log.Printf("map: input: report %s\n", rpt.Id)
			for _, u := range rpt.Units {
				gc, ok := hexes.GridCoordsFromString(u.PrevHex)
				if !ok && u.PrevHex == "N/A" {
					gc, ok = hexes.GridCoordsFromString(u.CurrHex)
				}
				if !ok {
					log.Fatalf("map: input: report %s: unit %-8s: prev hex %s: not okay\n", rpt.Id, u.Id, u.PrevHex)
				}
				ph, err := gc.ToMapCoords()
				if err != nil {
					log.Fatalf("map: input: report %s: unit %-8s: prev hex %s: %v\n", rpt.Id, u.Id, u.PrevHex, err)
				}
				log.Printf("map: input: report %s: unit %-8s: prev hex %s: (%4d %4d)\n", rpt.Id, u.Id, u.PrevHex, ph.Row, ph.Column)
			}
		}

		log.Printf("map: output %s\n", argsMap.output)

		return nil
	},
}
