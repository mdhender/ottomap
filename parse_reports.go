// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/clans"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var cmdParseReports = &cobra.Command{
	Use:   "reports",
	Short: "Parse all reports in the input path",
	Long:  `Read all TribeNet turn reports in the input and split them into unit movement files.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("parse: reports: input  %s\n", argsParse.input)
		log.Printf("parse: reports: output %s\n", argsParse.output)

		// find all turn reports in the input path. the files have
		// names that match the pattern YEAR-MONTH.CLAN_ID.input.txt
		index := domain.Index{
			ReportFiles: map[string]*domain.ReportFile{},
		}
		rxTurnReportFile, err := regexp.Compile(`^(\d{3})-(\d{2})\.(0\d{3})\.input\.txt$`)
		if err != nil {
			log.Fatal(err)
		}
		entries, err := os.ReadDir(argsParse.input)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				fileName := entry.Name()
				if matches := rxTurnReportFile.FindStringSubmatch(fileName); len(matches) == 4 {
					index.ReportFiles[fileName] = &domain.ReportFile{
						Path: argsParse.input,
						Name: fileName,
					}
				}
			}
		}
		for _, file := range index.ReportFiles {
			log.Printf("parse: reports: %s\n", file.Path)
		}
		if data, err := json.MarshalIndent(index, "", "  "); err != nil {
			log.Fatalf("parse: reports: marshal index: %v\n", err)
		} else if err := os.WriteFile(filepath.Join(argsParse.output, "index.json"), data, 0644); err != nil {
			log.Fatalf("parse: reports: create index: %v\n", err)
		} else {
			log.Printf("parse: reports: created %s\n", filepath.Join(argsParse.output, "index.json"))
		}

		// todo: remove this section and use the index directly
		var inputFiles []clans.InputFile
		for k, v := range index.ReportFiles {
			matches := rxTurnReportFile.FindStringSubmatch(k)
			if len(matches) != 4 {
				log.Printf("parse: reports: %s: matches %d, want 4!\n", k, len(matches))
				panic("assert(matches(index.ReportFiles.Name) == 4)")
			}
			inputFiles = append(inputFiles, clans.InputFile{
				Year:  matches[1],
				Month: matches[2],
				Clan:  matches[3],
				File:  filepath.Join(v.Path, v.Name),
			})
		}

		for _, inputFile := range inputFiles {
			clan, parseErr := clans.Parse(inputFile)
			if parseErr != nil {
				log.Printf("parse: reports: %s: error: %v\n", inputFile.File, parseErr)
				err = cerrs.ErrParseFailed
				continue
			}
			log.Printf("parse: reports: %s: units %3d: transfers %6d: settlements %6d\n", inputFile.File, len(clan.Units), len(clan.Transfers), len(clan.Settlements))
			for _, unit := range clan.Units {
				path := filepath.Join(argsParse.output, fmt.Sprintf("%s-%s.%s.%s.input.txt", inputFile.Year, inputFile.Month, inputFile.Clan, unit.Id))
				log.Printf("parse: reports: %s: %-8s => %s\n", inputFile.File, unit.Id, path)
				if err := os.WriteFile(path, unit.Text, 0644); err != nil {
					log.Fatalf("parse: reports: %s: %v\n", inputFile.File, err)
				}
			}
		}

		if err != nil {
			log.Printf("parse: reports: error parsing input: %v\n", err)
		}

	},
}
