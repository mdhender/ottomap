// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"github.com/mdhender/ottomap/htmx"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/mdhender/ottomap/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var argsServeHTMX struct{}

var cmdServeHTMX = &cobra.Command{
	Use:   "htmx",
	Short: "Serve HTMX client files",
	Long:  `Start a web server to serve HTMX client files.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("serve: host   %q\n", argsServe.host)
		log.Printf("serve: port   %q\n", argsServe.port)
		log.Printf("serve: db     %q\n", argsServe.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbName := filepath.Join(argsServe.paths.db, "ottomap.sqlite")
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
			server.WithHost(argsServe.host),
			server.WithPort(argsServe.port),
			server.WithApp(a),
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
