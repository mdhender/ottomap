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
	paths struct {
		data   string // path to data folder
		config string // path to configuration file to use
		input  string // path to read input files from
		output string // path to create index file in
	}
}

var cmdIndexReports = &cobra.Command{
	Use:   "reports",
	Short: "Add all reports in the input path to the configuration file",
	Long:  `Find all TribeNet turn reports in the input and add them to the configuration file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// if paths.data is set, then it is an absolute path and the other values must be blank since they will be set by the absolute path
		if argsIndexReports.paths.data != "" {
			// strip the default values if all of them are set
			if argsIndexReports.paths.config == "data" && argsIndexReports.paths.input == "data/input" && argsIndexReports.paths.output == "data/output" {
				argsIndexReports.paths.config, argsIndexReports.paths.input, argsIndexReports.paths.output = "", "", ""
			}
			// now check that they are not set
			if argsIndexReports.paths.config != "" {
				log.Fatalf("index: reports: config: cannot be set when data is set")
			} else if argsIndexReports.paths.input != "" {
				log.Fatalf("index: reports: input: cannot be set when data is set")
			} else if argsIndexReports.paths.output != "" {
				log.Fatalf("index: reports: output: cannot be set when data is set")
			}
			// do the abs path check for data
			if strings.TrimSpace(argsIndexReports.paths.data) != argsIndexReports.paths.data {
				log.Fatalf("index: reports: data: leading or trailing spaces are not allowed\n")
			} else if path, err := abspath(argsIndexReports.paths.data); err != nil {
				log.Fatalf("index: reports: data: %v\n", err)
			} else if sb, err := os.Stat(path); err != nil {
				log.Fatalf("index: reports: data: %v\n", err)
			} else if !sb.IsDir() {
				log.Fatalf("index: reports: data: %v is not a directory\n", path)
			} else {
				argsIndexReports.paths.data = path
			}
			// finally, update the other paths
			argsIndexReports.paths.config = argsIndexReports.paths.data
			argsIndexReports.paths.input = filepath.Join(argsIndexReports.paths.data, "input")
			argsIndexReports.paths.output = filepath.Join(argsIndexReports.paths.data, "output")
		}

		if strings.TrimSpace(argsIndexReports.paths.config) != argsIndexReports.paths.config {
			log.Fatalf("index: reports: config: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsIndexReports.paths.config); err != nil {
			log.Fatalf("index: reports: config: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("index: reports: config: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("index: reports: config: %v is not a directory\n", path)
		} else {
			argsIndexReports.paths.config = path
		}

		if strings.TrimSpace(argsIndexReports.paths.input) != argsIndexReports.paths.input {
			log.Fatalf("index: reports: input: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsIndexReports.paths.input); err != nil {
			log.Fatalf("index: reports: input: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("index: reports: input: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("index: reports: input: %v is not a directory\n", path)
		} else {
			argsIndexReports.paths.input = path
		}

		if strings.TrimSpace(argsIndexReports.paths.output) != argsIndexReports.paths.output {
			log.Fatalf("index: reports: output: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsIndexReports.paths.output); err != nil {
			log.Fatalf("index: reports: output: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("index: reports: output: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("index: reports: output: %v is not a directory\n", path)
		} else {
			argsIndexReports.paths.output = path
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// read any existing configuration file
		cfg := &config.Config{
			Path:       filepath.Join(argsIndexReports.paths.config, "config.json"),
			OutputPath: argsIndexReports.paths.output,
		}
		cfg.Read()
		log.Printf("index: todo: update to cache report data\n")
		cfg.Reports = nil

		// find all turn reports in the input path and add them to our configuration.
		// the files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
		rxTurnReportFile, err := regexp.Compile(`^(\d{3,4})-(\d{2})\.(0\d{3})\.report\.txt$`)
		if err != nil {
			log.Fatal(err)
		}
		entries, err := os.ReadDir(argsIndexReports.paths.input)
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
				path := filepath.Join(argsIndexReports.paths.input, fileName)
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
