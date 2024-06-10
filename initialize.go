// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	_ "embed"
	"github.com/mdhender/ottomap/pkg/sqlc"
	_ "modernc.org/sqlite"

	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	argsInitialize struct {
		paths struct {
			db        string
			input     string
			output    string
			public    string
			templates string
		}
		admin struct {
			username string
			email    string
			secret   string
		}
	}
)

var cmdInitialize = &cobra.Command{
	Use:   "initialize",
	Short: "Create and initialize the database for the server",
	Long:  `Create a new database for the server with an admin user.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsInitialize.paths.db) == 0 {
			return fmt.Errorf("missing database path")
		} else if argsInitialize.paths.db != strings.TrimSpace(argsInitialize.paths.db) {
			return fmt.Errorf("database path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsInitialize.paths.db); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("database path is not a directory")
		} else {
			argsInitialize.paths.db = path
		}
		if len(argsInitialize.paths.input) == 0 {
			return fmt.Errorf("missing input path")
		} else if argsInitialize.paths.input != strings.TrimSpace(argsInitialize.paths.input) {
			return fmt.Errorf("input path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsInitialize.paths.input); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("input path is not a directory")
		} else {
			argsInitialize.paths.input = path
		}
		if len(argsInitialize.paths.output) == 0 {
			return fmt.Errorf("missing output path")
		} else if argsInitialize.paths.output != strings.TrimSpace(argsInitialize.paths.output) {
			return fmt.Errorf("output path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsInitialize.paths.public); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("output path is not a directory")
		} else {
			argsInitialize.paths.output = path
		}
		if len(argsInitialize.paths.public) == 0 {
			return fmt.Errorf("missing public path")
		} else if argsInitialize.paths.public != strings.TrimSpace(argsInitialize.paths.public) {
			return fmt.Errorf("public path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsInitialize.paths.public); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("public path is not a directory")
		} else {
			argsInitialize.paths.public = path
		}
		if len(argsInitialize.paths.templates) == 0 {
			return fmt.Errorf("missing templates path")
		} else if argsInitialize.paths.templates != strings.TrimSpace(argsInitialize.paths.templates) {
			return fmt.Errorf("templates path must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsInitialize.paths.templates); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("templates path is not a directory")
		} else {
			argsInitialize.paths.templates = path
		}
		if len(argsInitialize.admin.username) == 0 {
			return fmt.Errorf("missing admin username")
		} else if argsInitialize.admin.username != strings.TrimSpace(argsInitialize.admin.username) {
			return fmt.Errorf("admin username must not contain leading or trailing spaces")
		} else if argsInitialize.admin.username != strings.ToLower(argsInitialize.admin.username) {
			return fmt.Errorf("admin username must be lowercase")
		}
		if len(argsInitialize.admin.email) == 0 {
			return fmt.Errorf("missing admin email")
		} else if argsInitialize.admin.email != strings.TrimSpace(argsInitialize.admin.email) {
			return fmt.Errorf("admin email must not contain leading or trailing spaces")
		} else if argsInitialize.admin.email != strings.ToLower(argsInitialize.admin.email) {
			return fmt.Errorf("admin email must be lowercase")
		}
		if len(argsInitialize.admin.secret) == 0 {
			return fmt.Errorf("missing admin secret")
		} else if argsInitialize.admin.secret != strings.TrimSpace(argsInitialize.admin.secret) {
			return fmt.Errorf("admin secret must not contain leading or trailing spaces")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbName := filepath.Join(argsInitialize.paths.db, "ottomap.sqlite")
		if sb, err := os.Stat(dbName); err == nil {
			log.Printf("initialize: %s: exists\n", dbName)
			if !sb.Mode().IsRegular() {
				log.Fatalf("initialize: %s: is not a regular file\n", dbName)
			} else if err = os.Remove(dbName); err != nil {
				log.Fatalf("initialize: %s: %v\n", dbName, err)
			}
			log.Printf("initialize: %s: removed\n", dbName)
		}

		// verify that database does not exist.
		if _, err := os.Stat(dbName); !os.IsNotExist(err) {
			log.Fatalf("error: %s: exists\n", dbName)
		}

		if err := sqlc.CreateDatabase(dbName); err != nil {
			log.Fatalf("error: %v\n", err)
		}

		db, err := sqlc.OpenDatabase(dbName, context.Background())
		if err != nil {
			log.Fatalf("initialize: db.open: %v\n", err)
		}
		defer func() {
			db.CloseDatabase()
		}()

		// update the metadata for the input, output, public, and template paths
		if err := db.UpdateInputOutputPaths(argsInitialize.paths.input, argsInitialize.paths.output); err != nil {
			log.Fatalf("initialize: iopaths: %v\n", err)
		} else if err := db.UpdateMetadataPublicPath(argsInitialize.paths.public); err != nil {
			log.Fatalf("initialize: public %q: %v\n", argsInitialize.paths.public, err)
		} else if err = db.UpdateMetadataTemplatesPath(argsInitialize.paths.templates); err != nil {
			log.Fatalf("initialize: templates %q: %v\n", argsInitialize.paths.templates, err)
		}

		// insert the default set of roles
		for _, rlid := range []string{"anonymous", "administrator", "operator", "service", "user", "authenticated"} {
			if err := db.CreateRole(rlid); err != nil {
				log.Fatalf("initialize: role %q: %v\n", rlid, err)
			}
		}

		// create the administrator account
		log.Printf("initialize: admin handle %q\n", argsInitialize.admin.username)
		log.Printf("initialize: admin secret %q\n", argsInitialize.admin.secret)
		if uid, err := db.CreateUser(argsInitialize.admin.username, "ottomap@example.com", argsInitialize.admin.secret); err != nil {
			log.Fatalf("initialize: admin %q: %v\n", argsInitialize.admin.username, err)
		} else {
			for _, rlid := range []string{"administrator", "operator", "service"} {
				if err := db.CreateUserRole(uid, rlid); err != nil {
					log.Fatalf("initialize: admin %q: role %q: %v\n", argsInitialize.admin.username, rlid, err)
				}
			}
		}

		// create the user account
		log.Printf("initialize: user  handle %q\n", "clan0138")
		log.Printf("initialize: user  secret %q\n", "password")
		if uid, err := db.CreateUser("clan0138", "clan0138@example.com", "password"); err != nil {
			log.Fatalf("initialize: user  %q: %v\n", "clan0138", err)
		} else {
			for _, rlid := range []string{"user"} {
				if err := db.CreateUserRole(uid, rlid); err != nil {
					log.Fatalf("initialize: user  %q: role %q: %v\n", "cla0138", rlid, err)
				}
			}
			if err = db.CreateClan(uid, "0138"); err != nil {
				log.Fatalf("initialize: user  %q: clan %q: %v\n", "clan0138", "0138", err)
			}
		}

		log.Printf("initialize: todo: encrypt the admin password\n")

		log.Printf("initialize: database created\n")
	},
}
