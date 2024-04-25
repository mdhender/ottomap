// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
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

		// turn reports have names like YEAR-MONTH.CLAN.input.txt
		pattern := `^(\d{3})-(\d{2})\.(0\d{3})\.input\.txt$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatal(err)
		}

		// find all turn reports in the input path
		var inputFiles []clans.InputFile
		entries, err := os.ReadDir(argsParse.input)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fileName := entry.Name()
			matches := re.FindStringSubmatch(fileName)
			//log.Printf("matches %2d %v\n", len(matches), matches)
			if len(matches) != 4 {
				continue
			}
			inputFiles = append(inputFiles, clans.InputFile{
				Year:  matches[1],
				Month: matches[2],
				Clan:  matches[3],
				File:  filepath.Join(argsParse.input, fileName),
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
