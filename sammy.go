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
	turnId string // maximum turn id to use
	debug  struct {
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

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		started := time.Now()
		log.Printf("data:   %s\n", argsSammy.paths.data)
		log.Printf("input:  %s\n", argsSammy.paths.input)
		log.Printf("output: %s\n", argsSammy.paths.output)

		inputs, err := turns.CollectInputs(argsSammy.paths.input)
		if err != nil {
			log.Fatalf("error: inputs: %v\n", err)
		}
		log.Printf("inputs: found %d turn reports\n", len(inputs))

		var allMoves []*parser.Movement_t
		for _, i := range inputs {
			started := time.Now()
			data, err := os.ReadFile(i.Path)
			if err != nil {
				log.Fatalf("error: read: %v\n", err)
			}
			mt, err := parser.ParseInput(i.Id, data, argsSammy.debug.parser, argsSammy.debug.sections, argsSammy.debug.steps, argsSammy.debug.nodes)
			if err != nil {
				log.Fatal(err)
			}
			allMoves = append(allMoves, mt...)
			log.Printf("%q: parsed %6d moves in %v\n", i.Id, len(mt), time.Since(started))
		}
		log.Printf("parsed %d inputs in %v\n", len(inputs), time.Since(started))

		// map all the sections
		err = turns.Map(allMoves, argsSammy.debug.maps)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("elapsed: %v\n", time.Since(started))
	},
}
