// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	argsList struct {
		paths struct {
			db string
		}
		reports struct {
			inDB     bool
			inFolder bool
		}
	}
)

var cmdList = &cobra.Command{
	Use:   "list",
	Short: "Display a list of reports in the database or folder",
	Long:  `Display a list of reports in the database or folder.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsList.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsList.paths.db != strings.TrimSpace(argsList.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsList.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsList.paths.db = path
		}
		log.Printf("list: db   %q\n", argsList.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// todo: database open for read should be in a function
		dbName := filepath.Join(argsList.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err != nil {
			log.Printf("list: db %s: does not exist\n", dbName)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("list: db %s: is not a regular file\n", dbName)
		}
		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("list: db: open %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		input, _, err := db.ReadInputOutputPaths()
		if err != nil {
			log.Fatalf("list: db: read paths: %v\n", err)
		}
		log.Printf("list: input %s\n", input)
		if sb, err := os.Stat(input); err != nil {
			log.Fatalf("list: input %s: does not exist\n", input)
		} else if !sb.IsDir() {
			log.Fatalf("list: input %s: is not a folder\n", input)
		}

		// inputReports are all the reports in the input path on the file system
		inputReports, err := allTheInputReports(input)
		if err != nil {
			log.Fatal(err)
		}

		// todo: write to a tabwriter
		for _, report := range inputReports {
			fmt.Printf("input: %s %s\n", report.name, report.cksum)
		}
	},
}

type reportFileMetadata struct {
	id     string
	path   string
	name   string
	year   int
	month  int
	clanId string
	cksum  string
}

func allTheInputReports(path string) (reports []reportFileMetadata, err error) {
	// find all turn reports in the input path and add them to our configuration.
	// the files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
	rxTurnReportFile, err := regexp.Compile(`^(\d{3,4})-(\d{2})\.(0\d{3})\.report\.txt$`)
	if err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(argsIndexReports.input)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		matches := rxTurnReportFile.FindStringSubmatch(name)
		if len(matches) != 4 {
			continue
		}

		// fetch the metadata for the report
		lrf, err := reportFileMetadataFromPathName(path, name)
		if err != nil {
			return nil, err
		}

		reports = append(reports, lrf)
	}

	return reports, nil
}

func reportFileMetadataFromPathName(path, name string) (reportFileMetadata, error) {
	// check for pattern of YEAR-MONTH.CLAN_ID.report.txt.
	rxTurnReportFile, err := regexp.Compile(`^(\d{3,4})-(\d{2})\.(0\d{3})\.report\.txt$`)
	if err != nil {
		panic(err)
	}

	// check that the path is absolute
	if ap, err := filepath.Abs(path); err != nil {
		return reportFileMetadata{}, err
	} else if ap != path {
		return reportFileMetadata{}, fmt.Errorf("path %q is not absolute", path)
	}

	matches := rxTurnReportFile.FindStringSubmatch(name)
	if len(matches) != 4 {
		return reportFileMetadata{}, cerrs.ErrNotATurnReport
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	clanId := matches[3]

	// load the file and generate the sha256 checksum
	cksum := ""
	if data, err := os.ReadFile(filepath.Join(path, name)); err != nil {
		log.Fatal(err)
	} else {
		cksum = fmt.Sprintf("%x", sha256.Sum256(data))
	}

	return reportFileMetadata{
		path:   path,
		name:   name,
		year:   year,
		month:  month,
		clanId: clanId,
		cksum:  cksum,
	}, nil
}
