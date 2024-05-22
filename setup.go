// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/mdhender/ottomap/parsers/report"
	"github.com/mdhender/ottomap/parsers/turn_reports"
	"github.com/spf13/cobra"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var argsSetup struct {
	report        string // turn report to process
	output        string // path to create output files in
	setup         string // path to create setup file
	originTerrain string // origin terrain
	debug         struct {
		captureRawText bool
		showSlugs      bool
	}
}

var cmdSetup = &cobra.Command{
	Use:   "setup",
	Short: "Setup a new mapping environment",
	Long:  `Create a setup.json from a new report file.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// todo: maybe we shouldn't be forcing output to be an absolute path

		if strings.TrimSpace(argsSetup.output) != argsSetup.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsSetup.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsSetup.output = path
		}

		if sb, err := os.Stat(argsSetup.report); err != nil && os.IsNotExist(err) {
			return err
		} else if os.IsNotExist(err) {
			return cerrs.ErrMissingReportFile
		} else if sb.IsDir() {
			return cerrs.ErrInvalidReportFile
		}

		argsSetup.setup = filepath.Join(argsSetup.output, "setup.json")
		if _, err := os.Stat(argsSetup.setup); err == nil {
			log.Printf("setup: setup file %s exists\n", argsSetup.setup)
			//return cerrs.ErrSetupExists
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ot, ok := domain.StringToTerrain(argsSetup.originTerrain)
		if !ok || ot == domain.TBlank {
			log.Fatalf("setup: origin terrain %q is not valid\n", argsSetup.originTerrain)
		}

		log.Printf("setup: report %s\n", argsSetup.report)
		log.Printf("setup: output %s\n", argsSetup.output)
		log.Printf("setup: setup  %s\n", argsSetup.setup)
		log.Printf("setup: otrn   %s\n", ot.String())

		data, err := os.ReadFile(argsSetup.report)
		if err != nil {
			log.Fatalf("setup: read report: %v", err)
		}
		log.Printf("setup: report: read %d bytes\n", len(data))

		sections, separator := turn_reports.Split(data)
		log.Printf("setup: sections: %d sections: separator %q\n", len(sections), separator)
		if len(sections) == 0 {
			log.Fatalf("setup: sections: error: report is empty\n")
		}

		var units []*report.Unit
		for _, section := range sections {
			lines := bytes.Split(section, []byte{'\n'})

			var sectionType domain.ReportSectionType
			if len(lines) != 0 {
				if bytes.HasPrefix(lines[0], []byte("Tribe ")) {
					sectionType = domain.RSUnit
				} else if bytes.HasPrefix(lines[0], []byte("Courier ")) {
					sectionType = domain.RSUnit
				} else if bytes.HasPrefix(lines[0], []byte("Element ")) {
					sectionType = domain.RSUnit
				} else if bytes.HasPrefix(lines[0], []byte("Fleet ")) {
					sectionType = domain.RSUnit
				} else if bytes.HasPrefix(lines[0], []byte("Garrison ")) {
					sectionType = domain.RSUnit
				} else if bytes.HasPrefix(lines[0], []byte("Settlements")) {
					sectionType = domain.RSSettlements
				} else if bytes.HasPrefix(lines[0], []byte("Transfers")) {
					sectionType = domain.RSTransfers
				} else {
					var slug []byte
					for n, line := range lines {
						if n >= 4 {
							break
						}
						if len(line) < 73 {
							slug = append(slug, line...)
						} else {
							slug = append(slug, line[:73]...)
							slug = append(slug, '.', '.', '.')
						}
						slug = append(slug, '\n')
					}
					log.Printf("parse: section: lines %8d\n%s\n", len(lines), string(slug))
					log.Fatalf("setup: section: error: unknown section type\n")
				}
			}
			switch sectionType {
			case domain.RSUnit:
				// these we process
			case domain.RSSettlements:
				continue
			case domain.RSTransfers:
				continue
			case domain.RSUnknown:
				panic("assert(sectionType!= domain.RSUnknown)")
			}

			// use a temporary id for the header section
			id := "setup.report"

			kind, unitId, year, month, hex, err := turn_reports.ParseHeaders(id, lines)
			if err != nil {
				log.Fatalf("setup: sections: error: %v\n", err)
			}
			log.Printf("setup: sections: %q: unit %q: year %04d: month %02d: hex %q\n", kind, unitId, year, month, hex)
			if !(len(unitId) == 4 && strings.HasPrefix(unitId, "0")) {
				log.Fatalf("setup: sections: error: not a setup report: invalid unit id %q\n", unitId)
			} else if hex == "" {
				log.Fatalf("setup: sections: error: not a setup report: missing current hex\n")
			}

			// we can create the "permanent" id for the report now that we know the clan and date
			id = fmt.Sprintf("%04d-%02d.%s", year, month, unitId)

			unit, err := report.ParseSection(lines, argsSetup.debug.showSlugs)
			if err != nil {
				log.Fatalf("setup: sections: error: %v\n", err)
			}
			units = append(units, unit)
		}

		if len(units) == 0 {
			log.Fatalf("setup: sections: error: no units in report file\n")
		}
		log.Printf("setup: sections: units %d\n", len(units))

		// sort the units by id
		sort.Slice(units, func(i, j int) bool {
			return units[i].Id < units[j].Id
		})

		// verify that the clan is the first unit
		if !units[0].IsClan() {
			log.Fatalf("setup: sections: error: could not find clan in report\n")
		}
		clan := units[0]

		// having a map of units is useful for linking parents
		allUnits := map[string]*report.Unit{}
		for _, unit := range units {
			allUnits[unit.Id] = unit
		}

		// fake having information for unit's prior turns.
		// this is not needed for the setup report, but it is needed for future turn reports.
		priorTurns := map[string]*report.Unit{}
		priorTurns[clan.Id] = &report.Unit{
			Id:    clan.Id,
			Start: clan.Start,
		}

		// link parents
		for _, unit := range units {
			if unit.IsClan() {
				continue
			}
			if unit.IsTribe() {
				unit.Parent, unit.ParentId = clan, clan.Id
			} else {
				parent, ok := allUnits[unit.Id[0:4]]
				if !ok {
					log.Fatalf("setup: sections: error: could not find parent for %q\n", unit.Id)
				}
				unit.Parent, unit.ParentId = parent, parent.Id
			}
		}

		// inefficient, but clear out the start and end hexes for units before processing the movements
		for _, unit := range units {
			if unit.IsClan() {
				if strings.HasPrefix(unit.Start, "##") {
					log.Printf("setup: sections: warning: starting hex uses hidden grid %q\n", unit.Start)
					log.Printf("setup: sections: warning: substituting \"OO\" for grid\n")
					unit.Start = "OO" + unit.Start[2:]
					log.Printf("setup: sections: warning: starting hex now %q\n", unit.Start)
				}
				unit.End = ""
			} else {
				// we are going to calculate the position of the unit in the grid,
				// so clear out the ending hex for all moves
				unit.Start, unit.End = "", ""
			}
		}

		// walk through every unit's movements and calculate the position in the grid
		for _, unit := range units {
			// save for debugging
			if b, err := json.MarshalIndent(unit, "", "  "); err != nil {
				log.Printf("setup: unit %q: error: %v\n", unit.Id, err)
			} else {
				log.Printf("setup: unit %q: results\n%s\n", unit.Id, string(b))
			}

			// starting position always depends on the parent's starting position.
			// this is true except for when the unit was created as an after-move action.
			// there's no way to know the starting position of the unit in that case.
			if _, ok := priorTurns[unit.Id]; !ok {
				unit.Start = unit.Parent.Start
			}

			// step through all the moves and calculate the position of the unit
			unit.Walk()
		}

		// save for debugging
		unitsMovesFile := filepath.Join(argsSetup.output, fmt.Sprintf("%04d-%02d.%s.moves.json", clan.Turn.Year, clan.Turn.Month, clan.Id))
		if data, err := json.MarshalIndent(units, "", "  "); err != nil {
			log.Printf("setup: units: error: %v\n", err)
		} else if err = os.WriteFile(unitsMovesFile, data, 0644); err != nil {
			log.Fatalf("setup: write units: %v", err)
		} else {
			log.Printf("setup: created %s\n", unitsMovesFile)
		}

		config := &domain.Config{
			Self:       argsSetup.setup,
			OutputPath: argsSetup.output,
			Reports: []*domain.ConfigReport{&domain.ConfigReport{
				Id:          fmt.Sprintf("%04d-%02d.%s", clan.Turn.Year, clan.Turn.Month, clan.Id),
				Input:       argsSetup.report,
				Year:        clan.Turn.Year,
				Month:       clan.Turn.Month,
				Clan:        clan.Id,
				Loaded:      time.Now().UTC(),
				Parsed:      unitsMovesFile,
				Fingerprint: hashData(data),
			}},
		}
		if data, err := json.MarshalIndent(config, "", "\t"); err != nil {
			log.Fatalf("setup: marshal config: %v", err)
		} else if err = os.WriteFile(argsSetup.setup, data, 0644); err != nil {
			log.Fatalf("setup: write config: %v", err)
		}
		log.Printf("setup: created %s\n", argsSetup.setup)

		return nil
	},
}

func hashData(data []byte) string {
	hash := fnv.New64a()
	_, _ = hash.Write(data)
	return fmt.Sprintf("%016x", hash.Sum64())
}
