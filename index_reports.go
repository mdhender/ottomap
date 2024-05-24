// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/config"
	"github.com/mdhender/ottomap/reports"
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
	Short: "Add all reports in the input path to the configuration file",
	Long:  `Find all TribeNet turn reports in the input and add them to the configuration file.`,
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

		// read any existing configuration file
		cfg := &config.Config{
			Path:       filepath.Join(argsIndexReports.output, "config.json"),
			OutputPath: argsIndexReports.output,
		}
		cfg.Read()
		log.Printf("index: todo: update to cache report data\n")
		cfg.Reports = nil

		// find all turn reports in the input path and add them to our configuration.
		// the files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
		rxTurnReportFile, err := regexp.Compile(`^(\d{3})-(\d{2})\.(0\d{3})\.report\.txt$`)
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
				matches := rxTurnReportFile.FindStringSubmatch(fileName)
				if len(matches) != 4 {
					continue
				}
				year, _ := strconv.Atoi(matches[1])
				month, _ := strconv.Atoi(matches[2])
				clanId := matches[3]
				id := fmt.Sprintf("%04d-%02d.%s", year, month, clanId)
				path := filepath.Join(argsIndexReports.input, fileName)
				// log.Printf("index: %s\n", path)

				if !cfg.Reports.Contains(id) {
					cfg.AddReport(&reports.Report{
						Id:     id,
						Path:   path,
						TurnId: fmt.Sprintf("%04d-%02d", year, month),
						Year:   year,
						Month:  month,
						Clan:   clanId,
					})
				}
			}
		}

		cfg.Save()
		log.Printf("index: created %s\n", cfg.Path)
	},
}
