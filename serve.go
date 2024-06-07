// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"github.com/mdhender/ottomap/app"
	"github.com/mdhender/ottomap/pkg/simba"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/mdhender/ottomap/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var argsServe struct {
	host  string
	port  string
	paths struct {
		db string
	}
}

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Long:  `Run a web server.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsServe.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsServe.paths.db != strings.TrimSpace(argsServe.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsServe.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsServe.paths.db = path
		}
		if len(argsServe.host) == 0 {
			argsServe.host = "localhost"
		} else if argsServe.host != strings.TrimSpace(argsServe.host) {
			return fmt.Errorf("host must not contain leading or trailing spaces")
		}
		if len(argsServe.port) == 0 {
			argsServe.port = "8080"
		} else if argsServe.port != strings.TrimSpace(argsServe.port) {
			return fmt.Errorf("port must not contain leading or trailing spaces")
		}
		log.Printf("serve: host %q\n", argsServe.host)
		log.Printf("serve: port %q\n", argsServe.port)
		log.Printf("serve: db   %q\n", argsServe.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbName := filepath.Join(argsServe.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err != nil {
			log.Printf("serve: db %s: does not exist\n", dbName)
		} else if !sb.Mode().IsRegular() {
			log.Fatalf("serve: db %s: is not a regular file\n", dbName)
		}

		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("serve: db: open %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		agent, err := simba.NewAgent(db, context.Background())
		if err != nil {
			log.Fatal(err)
		}

		a, err := app.New(
			app.WithVersion(version),
			app.WithStore(db),
			app.WithPolicyAgent(agent),
		)
		if err != nil {
			log.Fatal(err)
		}

		s, err := server.New(
			server.WithHost("localhost"),
			server.WithPort("3030"),
			server.WithApp(a),
		)
		if err != nil {
			log.Fatal(err)
		}

		s.ShowMeSomeRoutes()

		log.Printf("serve: listening on %s\n", s.BaseURL())
		if err := http.ListenAndServe(s.Addr, s.Router()); err != nil {
			log.Fatal(err)
		}
	},
}
