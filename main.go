// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottomap application
package main

import (
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/semver"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	version = semver.Version{Major: 0, Minor: 8, Patch: 2}
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	log.Printf("todo: detect when a unit is created as an after-move action\n")

	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.AddCommand(cmdIndex, cmdInitialize, cmdMap, cmdParse, cmdServe, cmdSetup, cmdVersion)

	cmdIndex.AddCommand(cmdIndexReports)
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.config, "config", "c", "data", "path to create configuration file in")
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.input, "input", "i", "data/input", "path to read input from")
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.output, "output", "o", "data/output", "path to write output to")

	cmdInitialize.Flags().StringVar(&argsInitialize.admin.email, "email", "", "email for administrator")
	if err := cmdInitialize.MarkFlagRequired("email"); err != nil {
		log.Fatalf("initialize: email: mark required: %v\n", err)
	}
	cmdInitialize.Flags().StringVar(&argsInitialize.admin.secret, "secret", "", "secret (passphrase) for administrator")
	if err := cmdInitialize.MarkFlagRequired("secret"); err != nil {
		log.Fatalf("initialize: secret: mark required: %v\n", err)
	}
	cmdInitialize.Flags().StringVar(&argsInitialize.admin.username, "handle", "", "handle (user name) for administrator")
	if err := cmdInitialize.MarkFlagRequired("secret"); err != nil {
		log.Fatalf("initialize: handle: mark handle: %v\n", err)
	}
	cmdInitialize.Flags().StringVar(&argsInitialize.paths.db, "db", "", "path to create server database in")
	if err := cmdInitialize.MarkFlagRequired("db"); err != nil {
		log.Fatalf("initialize: db: mark required: %v\n", err)
	}
	cmdInitialize.Flags().StringVar(&argsInitialize.paths.public, "public", "", "path to public files")
	if err := cmdInitialize.MarkFlagRequired("public"); err != nil {
		log.Fatalf("initialize: public: mark required: %v\n", err)
	}
	cmdInitialize.Flags().StringVar(&argsInitialize.paths.templates, "templates", "", "path to template files")
	if err := cmdInitialize.MarkFlagRequired("templates"); err != nil {
		log.Fatalf("initialize: templates: mark required: %v\n", err)
	}

	cmdMap.Flags().BoolVar(&argsMap.debug.sectionMaps, "debug-section-maps", false, "save section maps for debugging")
	cmdMap.Flags().BoolVar(&argsMap.debug.units, "debug-units", false, "enable unit debugging")
	cmdMap.Flags().BoolVar(&argsMap.show.gridCenters, "show-grid-centers", false, "show grid centers")
	cmdMap.Flags().BoolVar(&argsMap.show.gridCoords, "show-grid-id-coords", false, "show grid id and coordinates")
	cmdMap.Flags().BoolVar(&argsMap.show.gridNumbers, "show-grid-coords", false, "show grid coordinates")
	cmdMap.Flags().StringVar(&argsMap.clanId, "clan", "", "clan id to process")
	if err := cmdMap.MarkFlagRequired("clan"); err != nil {
		log.Fatalf("map: clan: mark required: %v\n", err)
	}
	cmdMap.Flags().StringVar(&argsMap.config, "config", "data/config.json", "configuration file to use")
	cmdMap.Flags().StringVar(&argsMap.turnId, "turn", "", "turn to process (yyyy-mm format)")

	cmdParse.PersistentFlags().BoolVar(&argsParse.debug.units, "debug-units", false, "enable unit debugging")
	cmdParse.PersistentFlags().StringVarP(&argsParse.index, "index", "i", ".", "index file to process")
	cmdParse.PersistentFlags().StringVarP(&argsParse.output, "output", "o", ".", "path to write output to")
	cmdParseReports.Flags().BoolVar(&argsParseReports.debug.captureRawText, "capture-raw-text", false, "capture raw text")
	cmdParseReports.Flags().StringVarP(&argsParseReports.gridOrigin, "grid-origin", "g", "OO", "initial grid value for '##'")
	cmdParse.AddCommand(cmdParseReports, cmdParseUnits)

	cmdServe.PersistentFlags().StringVar(&argsServe.host, "host", "", "host to serve on")
	cmdServe.PersistentFlags().StringVar(&argsServe.port, "port", "8080", "port to serve on")
	cmdServe.PersistentFlags().StringVar(&argsServe.paths.db, "db", "", "path to server database")
	if err := cmdServe.MarkPersistentFlagRequired("db"); err != nil {
		log.Fatalf("serve: db: mark required: %v\n", err)
	}

	cmdServe.AddCommand(cmdServeCatalyst)
	cmdServeCatalyst.Flags().StringVar(&argsServeCatalyst.paths.public, "build", "", "path to build folder")
	if err := cmdServeCatalyst.MarkFlagRequired("build"); err != nil {
		log.Fatalf("serve: catalyst: build: mark required: %v\n", err)
	}

	cmdServe.AddCommand(cmdServeEmberJS)
	cmdServeEmberJS.Flags().StringVar(&argsServeEmberJS.paths.public, "dist", "", "path distribution folder")
	if err := cmdServeEmberJS.MarkFlagRequired("dist"); err != nil {
		log.Fatalf("serve: emberjs: dist: mark required: %v\n", err)
	}

	cmdServe.AddCommand(cmdServeHTMX)

	cmdSetup.Flags().StringVar(&argsSetup.originTerrain, "origin-terrain", "PR", "origin terrain")
	if err := cmdSetup.MarkFlagRequired("origin-terrain"); err != nil {
		log.Fatalf("setup: origin-terrain: mark required: %v\n", err)
	}
	cmdSetup.Flags().StringVarP(&argsSetup.output, "output", "o", ".", "path to write map to")
	if err := cmdSetup.MarkFlagRequired("output"); err != nil {
		log.Fatalf("setup: output: mark required: %v\n", err)
	}
	cmdSetup.Flags().StringVarP(&argsSetup.report, "report", "r", "", "report file to process")
	if err := cmdSetup.MarkFlagRequired("report"); err != nil {
		log.Fatalf("setup: report: mark required: %v\n", err)
	}

	return cmdRoot.Execute()
}

var cmdRoot = &cobra.Command{
	Use:   "ottomap",
	Short: "Root command for our application",
	Long:  `Create maps from TribeNet turn reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Hello from root command\n")
	},
}

func abspath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	} else if sb, err := os.Stat(absPath); err != nil {
		return "", err
	} else if !sb.IsDir() {
		return "", cerrs.ErrNotDirectory
	}
	return absPath, nil
}

func isdir(path string) (bool, error) {
	sb, err := os.Stat(path)
	if err != nil {
		return false, err
	} else if !sb.IsDir() {
		return false, nil
	}
	return true, nil
}
