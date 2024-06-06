// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"github.com/mdhender/ottomap/htmx"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/mdhender/ottomap/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var argsServeHTMX struct {
	paths struct {
		db string // path to database
	}
}

var cmdServeHTMX = &cobra.Command{
	Use:   "htmx",
	Short: "Serve HTMX client files",
	Long:  `Start a web server to serve HTMX client files.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsServeHTMX.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsServeHTMX.paths.db != strings.TrimSpace(argsServeHTMX.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsServeHTMX.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsServeHTMX.paths.db = path
		}
		log.Printf("serve: htmx: db %s\n", argsServeHTMX.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbName := filepath.Join(argsServeHTMX.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err != nil {
			log.Printf("serve: htmx: db %s: does not exist\n", dbName)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("serve: htmx: db %s: is not a regular file\n", dbName)
		}

		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("serve: htmx: db: open %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		a, err := htmx.New(db)
		if err != nil {
			log.Fatal(err)
		}

		s, err := server.New(
			server.WithHost("localhost"),
			server.WithPort("3030"),
			server.WithHTMX(a),
		)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("serve: htmx: listening on %s\n", s.BaseURL())
		if err := http.ListenAndServe(s.Addr, s.Router()); err != nil {
			log.Fatal(err)
		}
	},
}
