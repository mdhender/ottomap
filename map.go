// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/actions"
	"github.com/mdhender/ottomap/config"
	"github.com/mdhender/ottomap/reports"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var argsMap struct {
	paths struct {
		data   string
		config string // path to configuration file
	}
	clanId string // clan id to use
	turnId string // turn id to use
	debug  struct {
		nodes       bool
		sectionMaps bool
		steps       bool
		units       bool
	}
	features struct{}
	parse    struct{}
	show     struct {
		gridCenters     bool
		gridCoords      bool
		gridNumbers     bool
		ignoredSections bool
		sectionData     bool
		skippedSections bool
		steps           bool
	}
}

var cmdMap = &cobra.Command{
	Use:   "map",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// if paths.data is set, then it is an absolute path and the other values must be blank since they will be set by the absolute path
		if argsMap.paths.data != "" {
			// strip the default values if all of them are set
			if argsMap.paths.config == "data/config.json" {
				argsMap.paths.config = ""
			}
			// now check that they are not set
			if argsMap.paths.config != "" {
				log.Fatalf("error: config: cannot be set when data is set")
			}
			// do the abs path check for data
			if strings.TrimSpace(argsMap.paths.data) != argsMap.paths.data {
				log.Fatalf("error: data: leading or trailing spaces are not allowed\n")
			} else if path, err := abspath(argsMap.paths.data); err != nil {
				log.Fatalf("error: data: %v\n", err)
			} else if sb, err := os.Stat(path); err != nil {
				log.Fatalf("error: data: %v\n", err)
			} else if !sb.IsDir() {
				log.Fatalf("error: data: %v is not a directory\n", path)
			} else {
				argsMap.paths.data = path
			}
			// finally, update the other paths
			argsMap.paths.config = filepath.Join(argsMap.paths.data, "config.json")
		}

		if strings.TrimSpace(argsMap.paths.config) != argsMap.paths.config {
			log.Fatalf("error: config: leading or trailing spaces are not allowed\n")
		} else if path, err := filepath.Abs(argsMap.paths.config); err != nil {
			log.Printf("config: output: %q\n", argsMap.paths.config)
			log.Printf("config: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: config: %v\n", err)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("error: config: %v is not a file\n", path)
		} else {
			argsMap.paths.config = path
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("maps: todo: detect when a unit is created as an after-move action\n")

		log.Printf("config: file %s\n", argsMap.paths.config)
		cfg, err := config.Load(argsMap.paths.config)
		if err != nil {
			log.Fatalf("error: config: %v\n", err)
		}
		if len(cfg.Reports) == 0 {
			log.Fatalf("error: config: no reports\n")
		}

		log.Printf("config: path   %s\n", cfg.Path)
		log.Printf("config: output %s\n", cfg.OutputPath)

		cfg.Inputs.ClanId = argsMap.clanId
		log.Printf("config: clan %q\n", cfg.Inputs.ClanId)

		if argsMap.show.gridCoords && argsMap.show.gridNumbers {
			argsMap.show.gridNumbers = false
		}
		if argsMap.debug.sectionMaps {
			panic("this needs to be fixed")
		}

		// if turn id is not on the command line, use the current turn from the configuration.
		if argsMap.turnId == "" {
			// assumes that the configuration's reports are sorted by turn id.
			rptCurr := cfg.Reports[len(cfg.Reports)-1]
			cfg.Inputs.TurnId = rptCurr.TurnId
			cfg.Inputs.Year = rptCurr.Year
			cfg.Inputs.Month = rptCurr.Month
		} else {
			// convert command line's yyyy-mm to year, month
			if yyyy, mm, ok := strings.Cut(argsMap.turnId, "-"); !ok {
				log.Fatalf("error: invalid turn %q\n", argsMap.turnId)
			} else if yyyy = strings.TrimSpace(yyyy); yyyy == "" {
				log.Fatalf("error: invalid turn %q\n", argsMap.turnId)
			} else if cfg.Inputs.Year, err = strconv.Atoi(yyyy); err != nil {
				log.Fatalf("error: invalid turn %q: year %v\n", argsMap.turnId, err)
			} else if mm = strings.TrimSpace(mm); mm == "" {
				log.Fatalf("error: invalid turn %q\n", argsMap.turnId)
			} else if cfg.Inputs.Month, err = strconv.Atoi(mm); err != nil {
				log.Fatalf("error: invalid turn %q: month %v\n", argsMap.turnId, err)
			} else {
				cfg.Inputs.TurnId = fmt.Sprintf("%04d-%02d", cfg.Inputs.Year, cfg.Inputs.Month)
			}
		}
		log.Printf("config: turn year  %4d\n", cfg.Inputs.Year)
		log.Printf("config: turn month %4d\n", cfg.Inputs.Month)

		// update the ignore flag based on the turn from the configuration
		for _, rpt := range cfg.Reports {
			if rpt.Clan == argsMap.clanId {
				rptTurnId := fmt.Sprintf("%04d-%02d", rpt.Year, rpt.Month)
				if rptTurnId > cfg.Inputs.TurnId {
					log.Printf("config: %s: forcing ignore\n", rptTurnId)
					rpt.Ignore = true
				}
			}
		}

		// collect the reports that we're going to process
		var allReports []*reports.Report
		for _, rpt := range cfg.Reports {
			if rpt.Ignore {
				continue
			}
			allReports = append(allReports, rpt)
		}
		if len(allReports) == 0 {
			log.Fatalf("error: no files matched constraints\n")
		}
		//log.Printf("reports %d\n", len(allReports))
		sort.Slice(allReports, func(i, j int) bool {
			return allReports[i].TurnId < allReports[j].TurnId
		})

		//log.Printf("todo: followers are not updated after movement\n")
		//log.Printf("todo: hexes are not assigned for each step in the results\n")
		//log.Printf("todo: named hexes that are only in the status line are missed\n")
		//log.Printf("todo: walk the hex reports and update grid as well as ending coordinates\n")

		// users are required to provide starting grid coordinates if they're not already in the report
		log.Printf("warning: otto now requires starting grid coordinates\n")

		if argsMap.show.skippedSections {
			log.Printf("warning: show sections skipped is enabled!\n")
		}
		if argsMap.show.steps {
			log.Printf("warning: show steps is enabled!\n")
		}
		if argsMap.debug.sectionMaps {
			panic("this needs to be fixed")
		}

		err = actions.MapReports(allReports,
			argsMap.clanId,
			cfg.Inputs.GridOriginId,
			cfg.OutputPath,
			argsMap.show.gridCenters,
			argsMap.show.gridCoords,
			argsMap.show.gridNumbers,
			argsMap.show.ignoredSections,
			argsMap.show.sectionData,
			argsMap.show.skippedSections,
			argsMap.show.steps,
			argsMap.debug.steps,
			argsMap.debug.nodes)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		return nil
	},
}
