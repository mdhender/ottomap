// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/turn_reports"
	"github.com/mdhender/ottomap/parsers/turn_reports/movements"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var argsParseReports struct {
	gridOrigin string // grid value to replace ## with
	debug      struct {
		captureRawText bool
		clanShowSlugs  bool
	}
}

var cmdParseReports = &cobra.Command{
	Use:   "reports",
	Short: "Parse all reports in the index file",
	Long:  `Create unit movement files for all TribeNet turn reports listed in the index file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if argsParseReports.gridOrigin == "" {
			return fmt.Errorf("missing grid origin")
		} else if strings.TrimSpace(argsParseReports.gridOrigin) != argsParseReports.gridOrigin {
			return fmt.Errorf("grid origin can not contain spaces")
		} else if len(argsParseReports.gridOrigin) != 2 {
			return fmt.Errorf("grid orgin must be two upper-case letters")
		} else if strings.Trim(argsParseReports.gridOrigin, "ABCDEFGHIJKLMNOPQRSTUVWYZ") != "" {
			return fmt.Errorf("grid orgin must be two upper-case letters")
		}
		log.Printf("parse: reports: index  %s\n", argsParse.index)
		log.Printf("parse: reports: output %s\n", argsParse.output)
		log.Printf("parse: reports: origin %s\n", argsParseReports.gridOrigin)
		if argsParse.debug.units {
			log.Printf("parse: reports: debug: units %v\n", argsParse.debug.units)
		}

		var index *domain.Index
		if data, err := os.ReadFile(argsParse.index); err != nil {
			log.Fatalf("parse: reports: failed to read index file: %v", err)
		} else if err = json.Unmarshal(data, &index); err != nil {
			log.Fatalf("parse: reports: failed to parse index file: %v", err)
		}
		log.Printf("parse: reports: loaded index file\n")

		// for consistency in reporting, sort the indexes
		var ids []string
		for id := range index.ReportFiles {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		log.Printf("parse: reports: sorted index file\n")

		// enable debug buffers
		movements.EnableDebugBuffer()

		log.Printf("parse: reports: todo: fail on first parsing error\n")

		errCount := 0
		for _, id := range ids {
			rpf := index.ReportFiles[id]

			// skip if we're debugging units and this report doesn't have
			// a debug section or any units in that section.
			if argsParse.debug.units {
				if rpf.Debug == nil {
					log.Printf("parse: reports: %s: debug: units: skipping (unset)\n", rpf.Id)
					continue
				} else if rpf.Debug.Units == nil {
					log.Printf("parse: reports: %s: debug: units: skipping (no map)\n", rpf.Id)
					continue
				} else if len(rpf.Debug.Units) == 0 {
					log.Printf("parse: reports: %s: debug: units: skipping (empty map)\n", rpf.Id)
					continue
				}
				log.Printf("parse: reports: %s: debug: units: %d entries\n", rpf.Id, len(rpf.Debug.Units))
				for k, v := range rpf.Debug.Units {
					log.Printf("parse: reports: %s: debug: unit %-6s: %v\n", rpf.Id, k, v)
				}
			}

			rss, err := turn_reports.Parse(rpf, argsParseReports.debug.clanShowSlugs, argsParseReports.debug.captureRawText)
			if err != nil {
				log.Printf("parse: reports: %s: error: %v\n", rpf.Id, err)
				errCount++
				break
			}
			//log.Printf("parse: reports: %s: sections %3d\n", rpf.Id, len(rss))

			for _, rs := range rss {
				path := filepath.Join(argsParse.output, fmt.Sprintf("%s.%s.json", rpf.Id, rs.Id))
				data, err := json.MarshalIndent(rs, "", "  ")
				if err != nil {
					log.Fatalf("parse: reports: %s: %v\n", rpf.Id, err)
				}
				err = os.WriteFile(path, data, 0644)
				if err != nil {
					log.Fatalf("parse: reports: %s: %s: %v\n", rpf.Id, rs.Id, err)
				}
				log.Printf("parse: reports: %s: %-8s ==> %s\n", rpf.Id, rs.Id, path)
			}
		}

		// write out our debug logs
		if err := os.WriteFile(filepath.Join(argsParse.output, "debug_turn_report_movements.txt"), movements.GetDebugBuffer(), 0644); err != nil {
			log.Fatal(err)
		}

		if errCount != 0 {
			log.Fatalf("parse: reports: halting due to %d errors above\n", errCount)
		}

		log.Printf("parse: reports: todo: find that one scouting step that i had to modify back in the day\n")
		log.Printf("parse: reports: todo: ignore the temptation to push section data into a database\n")

		return nil
	},
}
