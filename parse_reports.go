// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/parsers/clan_turn"
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

		// find all turn report files in the input directory.
		// they will be all files with a name like YEAR-MONTH.TRIBE.input.txt
		var inputFiles []clan_turn.InputFile
		entries, err := os.ReadDir(argsParse.input)
		if err != nil {
			log.Fatal(err)
		} else {
			// tribe input files have the pattern of YYY-MM.CLAN.input.txt.
			pattern := `^\d{3}-\d{2}\.0\d{3}\.input\.txt$`
			re, err := regexp.Compile(pattern)
			if err != nil {
				log.Fatal(err)
			}
			for _, entry := range entries {
				// If the entry is a file and the name matches our pattern...
				if !entry.IsDir() && re.MatchString(entry.Name()) {
					fileName := entry.Name()
					inputFiles = append(inputFiles, clan_turn.InputFile{
						Year:  fileName[0:3],
						Month: fileName[4:6],
						Clan:  fileName[7:11],
						File:  filepath.Join(argsParse.input, fileName),
					})
				}
			}
		}

		for _, inputFile := range inputFiles {
			clan, parseErr := clan_turn.Parse(inputFile)
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
