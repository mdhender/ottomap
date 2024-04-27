// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var argsIndexReports struct {
	input  string // path to read input files from
	output string // path to create index file in
}

var cmdIndexReports = &cobra.Command{
	Use:   "reports",
	Short: "Add all reports in the input path to the index file",
	Long:  `Find all TribeNet turn reports in the input and add them to the index file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(argsIndexReports.input) != argsIndexReports.input {
			log.Fatalf("index: reports: input: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsIndexReports.input); err != nil {
			log.Fatalf("index: reports: input: %v\n", err)
		} else {
			argsIndexReports.input = path
		}

		if strings.TrimSpace(argsIndexReports.output) != argsIndexReports.output {
			log.Fatalf("index: reports: output: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsIndexReports.output); err != nil {
			log.Fatalf("index: reports: output: %v\n", err)
		} else {
			argsIndexReports.output = path
		}

		// find all turn reports in the input path and add them to our index.
		// the files have names that match the pattern YEAR-MONTH.CLAN_ID.input.txt.
		index := domain.Index{
			ReportFiles: map[string]*domain.ReportFile{},
		}
		rxTurnReportFile, err := regexp.Compile(`^(\d{3})-(\d{2})\.(0\d{3})\.input\.txt$`)
		if err != nil {
			log.Fatal(err)
		}
		entries, err := os.ReadDir(argsIndexReports.input)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				fileName := entry.Name()
				if matches := rxTurnReportFile.FindStringSubmatch(fileName); len(matches) == 4 {
					log.Printf("index: reports: %s\n", filepath.Join(argsIndexReports.input, fileName))
					year, _ := strconv.Atoi(matches[1])
					month, _ := strconv.Atoi(matches[2])
					clan, _ := strconv.Atoi(matches[3])
					id := fmt.Sprintf("%03d-%02d.%04d", year, month, clan)
					index.ReportFiles[id] = &domain.ReportFile{
						Id:    id,
						Path:  filepath.Join(argsIndexReports.input, fileName),
						Year:  year,
						Month: month,
						Clan:  clan,
					}
				}
			}
		}

		// save the index to a JSON file in the output path
		indexFile := filepath.Join(argsIndexReports.output, "index.json")
		if data, err := json.MarshalIndent(index, "", "  "); err != nil {
			log.Fatalf("index: reports: marshal index: %v\n", err)
		} else if err := os.WriteFile(indexFile, data, 0644); err != nil {
			log.Fatalf("index: reports: create index: %v\n", err)
		}
		log.Printf("index: reports: created %s\n", indexFile)
	},
}
