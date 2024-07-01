// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/turns"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
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
		maps     bool
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

		mapTurns := map[string][]*parser.Turn_t{}
		totalUnitMoves := 0
		for _, i := range inputs {
			started := time.Now()
			data, err := os.ReadFile(i.Path)
			if err != nil {
				log.Fatalf("error: read: %v\n", err)
			}
			turn, err := parser.ParseInput(i.Id, data, argsSammy.debug.parser, argsSammy.debug.sections, argsSammy.debug.steps, argsSammy.debug.nodes)
			if err != nil {
				log.Fatal(err)
			}
			turnId := fmt.Sprintf("%04d-%02d", turn.Year, turn.Month)
			mapTurns[turnId] = append(mapTurns[turnId], turn)
			totalUnitMoves += len(turn.UnitMoves)
			log.Printf("%q: parsed %6d units in %v\n", i.Id, len(turn.UnitMoves), time.Since(started))
		}
		log.Printf("parsed %d inputs in to %d turns and %d units %v\n", len(inputs), len(mapTurns), totalUnitMoves, time.Since(started))

		// map all the sections
		err = turns.Map(mapTurns, argsSammy.originGrid, argsSammy.quitOnInvalidGrid, argsSammy.warnOnInvalidGrid, argsSammy.debug.maps)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("elapsed: %v\n", time.Since(started))
	},
}
