// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"github.com/mdhender/ottomap/actions"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	argsParse struct {
		paths struct {
			db string
		}
		reports struct {
			inDB     bool
			inFolder bool
		}
	}
)

var cmdParse = &cobra.Command{
	Use:   "parse",
	Short: "Parse reports",
	Long:  `Parse new reports in the database.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsParse.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsParse.paths.db != strings.TrimSpace(argsParse.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsParse.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsParse.paths.db = path
		}
		log.Printf("parse: db   %q\n", argsParse.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// todo: database open for read should be in a function
		dbName := filepath.Join(argsParse.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err != nil {
			log.Printf("parse: db %s: does not exist\n", dbName)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("parse: db %s: is not a regular file\n", dbName)
		}
		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("parse: db: open %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		log.Printf("parse: parsing pending imports\n")

		// this could be a race condition
		pendingRows, err := db.Queries.ReadPendingInputMetadata(db.Ctx)
		if err != nil {
			log.Fatalf("parsePending: %v\n", err)
		}

		parsed := 0
		for n, pendingRow := range pendingRows {
			log.Printf("parsePending: name %s: id %d: row %d of %d\n", pendingRow.Name, pendingRow.ID, n+1, len(pendingRows))
			if err = actions.ParsePendingInput(db, pendingRow); err != nil {
				log.Fatalf("parsePending: name %s: id %d: %v\n", pendingRow.Name, pendingRow.ID, err)
			}
			parsed++
		}

		log.Printf("parse: parsed %d imports\n", parsed)
	},
}
