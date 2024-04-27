// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/clans"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var cmdParseReports = &cobra.Command{
	Use:   "reports",
	Short: "Parse all reports in the index file",
	Long:  `Create unit movement files for all TribeNet turn reports listed in the index file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("parse: reports: index  %s\n", argsParse.index)
		log.Printf("parse: reports: output %s\n", argsParse.output)

		var index *domain.Index
		if data, err := os.ReadFile(argsParse.index); err != nil {
			log.Fatalf("parse: reports: failed to read index file: %v", err)
		} else if err = json.Unmarshal(data, &index); err != nil {
			log.Fatalf("parse: reports: failed to parse index file: %v", err)
		}
		log.Printf("parse: reports: loaded index file\n")

		var err error
		for _, reportFile := range index.ReportFiles {
			clan, parseErr := clans.Parse(reportFile)
			if parseErr != nil {
				log.Printf("parse: reports: %s: error: %v\n", reportFile.Name, parseErr)
				err = cerrs.ErrParseFailed
				continue
			}
			log.Printf("parse: reports: %s: clan %p\n", reportFile.Name, clan)

			log.Printf("parse: reports: %s: units %3d: transfers %6d: settlements %6d\n", reportFile.Name, len(clan.Units), len(clan.Transfers), len(clan.Settlements))

			log.Printf("parse: reports: todo: year, month, and clan must be added to the ReportFile struct\n")
			log.Printf("parse: reports: todo: push section data into the domain model structs instead of files\n")
			log.Printf("parse: reports: todo: ignore the temptation to push section data into a database\n")

			//for _, unit := range clan.Units {
			//	path := filepath.Join(argsParse.output, fmt.Sprintf("%s-%s.%s.%s.input.txt", inputFile.Year, inputFile.Month, inputFile.Clan, unit.Id))
			//	log.Printf("parse: reports: %s: %-8s => %s\n", inputFile.File, unit.Id, path)
			//	if err := os.WriteFile(path, unit.Text, 0644); err != nil {
			//		log.Fatalf("parse: reports: %s: %v\n", inputFile.File, err)
			//	}
			//}
		}

		if err != nil {
			log.Printf("parse: reports: error parsing input: %v\n", err)
		}
	},
}
