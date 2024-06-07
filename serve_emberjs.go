// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/server"
	"github.com/mdhender/ottomap/spa"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var argsServeEmberJS struct {
	paths struct {
		public string // path to database
	}
}

var cmdServeEmberJS = &cobra.Command{
	Use:   "emberjs",
	Short: "Serve EmberJS client files",
	Long:  `Start a web server to serve EmberJS client files.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsServeEmberJS.paths.public) == 0 {
			return fmt.Errorf("missing path to public files")
		} else if argsServeEmberJS.paths.public != strings.TrimSpace(argsServeEmberJS.paths.public) {
			return fmt.Errorf("path to public files must not contain leading or trailing spaces")
		} else if path, err := filepath.Abs(argsServeEmberJS.paths.public); err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("path to public files is not a directory")
		} else {
			argsServeEmberJS.paths.public = path
		}
		log.Printf("serve: host   %q\n", argsServe.host)
		log.Printf("serve: port   %q\n", argsServe.port)
		log.Printf("serve: public %q\n", argsServeEmberJS.paths.public)
		log.Printf("serve: db     %q\n", argsServe.paths.db)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		a, err := spa.New(argsServeEmberJS.paths.public)
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

		log.Printf("serve: emberjs: listening on %s\n", s.BaseURL())
		if err := http.ListenAndServe(s.Addr, s.Router()); err != nil {
			log.Fatal(err)
		}
	},
}
