// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/actions"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	argsImport struct {
		debug bool
		paths struct {
			db string
		}
		reports struct {
			inDB     bool
			inFolder bool
		}
	}
)

var cmdImport = &cobra.Command{
	Use:   "import",
	Short: "Import reports to the database",
	Long:  `Import individual reports or an entire folder to the database.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsImport.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsImport.paths.db != strings.TrimSpace(argsImport.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsImport.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsImport.paths.db = path
		}
		log.Printf("import: db   %q\n", argsImport.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// todo: database open for read should be in a function
		dbName := filepath.Join(argsImport.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err != nil {
			log.Printf("import: db %s: does not exist\n", dbName)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("import: db %s: is not a regular file\n", dbName)
		}
		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("import: db: open %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		input, _, err := db.ReadInputOutputPaths()
		if err != nil {
			log.Fatalf("import: db: read paths: %v\n", err)
		}
		log.Printf("import: %s\n", input)
		if sb, err := os.Stat(input); err != nil {
			log.Fatalf("import: %s: does not exist\n", input)
		} else if !sb.IsDir() {
			log.Fatalf("import: %s: is not a folder\n", input)
		}

		var inputReports []reportFileMetadata

		// did the user specify a report or are we defaulting to loading all the reports in the input path?
		importAllFiles := len(args) == 0
		if importAllFiles {
			inputReports, err = allTheInputReports(input)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			for _, arg := range args {
				// verify that the file exists
				if sb, err := os.Stat(arg); err != nil {
					log.Fatalf("import: %s: does not exist\n", arg)
				} else if !sb.Mode().IsRegular() {
					log.Fatalf("import: %s: is not a regular file\n", arg)
				}

				path, name := filepath.Split(arg)
				if path == "" {
					path = "."
				}
				if ap, err := filepath.Abs(path); err != nil {
					log.Fatalf("import: %s: %v\n", arg, err)
				} else {
					path = ap
				}

				lrf, err := reportFileMetadataFromPathName(path, name)
				if err != nil {
					log.Fatalf("import: %s: %v\n", arg, err)
				}

				inputReports = append(inputReports, lrf)
			}
		}

		// todo: write to a tabwriter
		for _, report := range inputReports {
			_, err := actions.ImportReport(db, filepath.Join(report.path, report.name), argsImport.debug)
			if err != nil {
				if !errors.Is(err, cerrs.ErrDuplicateChecksum) {
					log.Fatalf("import: %s: %v\n", report.name, err)
				} else if !importAllFiles {
					// user specified this report, and it already exists in the database
					log.Fatalf("import: %s: %v\n", report.name, err)
				}
				log.Printf("import: %s: duplicate file, ignored\n", report.name)
				continue
			}
			log.Printf("import: %s: imported\n", report.name)
		}

		log.Printf("import: imported %d reports\n", len(inputReports))
	},
}
