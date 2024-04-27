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
)

var argsParseReports struct {
	debug struct {
		clanShowSlugs      bool
		clanCaptureRawText bool
	}
}

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
		for _, rpf := range index.ReportFiles {
			rss, parseErr := clans.Parse(rpf, argsParseReports.debug.clanShowSlugs, argsParseReports.debug.clanCaptureRawText)
			if parseErr != nil {
				log.Printf("parse: reports: %s: error: %v\n", rpf.Id, parseErr)
				err = cerrs.ErrParseFailed
				continue
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

		if err != nil {
			log.Printf("parse: reports: error parsing input: %v\n", err)
		}

		log.Printf("parse: reports: todo: push section data into the domain model structs instead of files\n")
		log.Printf("parse: reports: todo: ignore the temptation to push section data into a database\n")
	},
}
