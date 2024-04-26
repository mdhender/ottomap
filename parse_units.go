// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/parsers/units"
	"github.com/mdhender/ottomap/parsers/units/movements"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var cmdParseUnits = &cobra.Command{
	Use:   "units",
	Short: "Parse all unit files in the input path",
	Long:  `Read all unit input files, parse them for locations, status, and movement data. Write the combined results into the output path.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("parse: units: input  %s\n", argsParse.input)
		log.Printf("parse: units: output %s\n", argsParse.output)

		// unit files have names like YEAR-MONTH.CLAN.UNIT.input.txt
		pattern := `^(\d{3})-(\d{2})\.(0\d{3})\.(\d{4}([cefg][1-9])?)\.input\.txt$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatal(err)
		}

		// find all unit input files in the input directory.
		var inputFiles []units.InputFile
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
			// check for 6 because there are actually 6 captures in the pattern
			if len(matches) != 6 {
				continue
			}
			inputFiles = append(inputFiles, units.InputFile{
				Year:  matches[1],
				Month: matches[2],
				Clan:  matches[3],
				Unit:  matches[4],
				File:  filepath.Join(argsParse.input, fileName),
			})
		}

		clans := map[string]*units.Clan{}

		for _, inputFile := range inputFiles {
			clan, ok := clans[inputFile.Clan]
			if !ok {
				clan = &units.Clan{Id: inputFile.Clan}
				clans[clan.Id] = clan
			}
			//log.Printf("parse: units: input  %+v\n", inputFile)
			unit, parseErr := units.Parse(inputFile)
			if parseErr != nil {
				log.Printf("unit %s: error: %v\n", inputFile.File, parseErr)
				err = cerrs.ErrParseFailed
				continue
			}
			clan.Units = append(clan.Units, unit)
			log.Printf("clan %s: unit %6s: %-7s => %-7s\n", clan.Id, unit.Id, unit.Started, unit.Finished)
			if unit.Movement == nil {
				log.Printf("clan %s: unit %s: missing movement\n", clan.Id, unit.Id)
			} else {
				var raw, buf string
				for n, step := range unit.Movement.Steps {
					if n != 0 {
						raw += ", "
						buf += ", "
					}
					raw += fmt.Sprintf("{%s}", step.RawText)
					buf += fmt.Sprintf("%v", *step)
				}
				log.Printf("clan %s: unit %s: steps %d => %s\n", clan.Id, unit.Id, len(unit.Movement.Steps), raw)
				log.Printf("clan %s: unit %s: steps %d => %s\n", clan.Id, unit.Id, len(unit.Movement.Steps), buf)
			}
			if unit.Movement.Failed.Direction != "" {
				log.Printf("clan %s: unit %s: failed %+v\n", clan.Id, unit.Id, unit.Movement.Failed)
			}
		}

		if err != nil {
			log.Printf("error parsing input: %v\n", err)
		}

		// write out our debug log
		if b := movements.DebugBuffer.Bytes(); len(b) > 0 {
			if err := os.WriteFile(filepath.Join(argsParse.output, "debug_movements.txt"), b, 0644); err != nil {
				log.Fatal(err)
			}
		}
	},
}
