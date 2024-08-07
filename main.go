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
	version = semver.Version{Major: 0, Minor: 14, Patch: 4}
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// todo: detect when a unit is created as an after-move action

	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.AddCommand(cmdRender, cmdServe, cmdVersion)

	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTiles, "debug-dump-all-tiles", false, "dump all tiles")
	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTurns, "debug-dump-all-turns", false, "dump all turns")
	cmdRender.Flags().BoolVar(&argsRender.debug.maps, "debug-maps", false, "enable maps debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.nodes, "debug-nodes", false, "enable node debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.parser, "debug-parser", false, "enable parser debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.sections, "debug-sections", false, "enable sections debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.steps, "debug-steps", false, "enable step debugging")
	cmdRender.Flags().BoolVar(&argsRender.experimental.stripCR, "debug-strip-cr", false, "experimental: enable conversion of DOS EOL")
	cmdRender.Flags().BoolVar(&argsRender.experimental.splitTrailingUnits, "x-split-units", false, "experimental: split trailing units")
	cmdRender.Flags().BoolVar(&argsRender.mapper.Dump.BorderCounts, "dump-border-counts", false, "dump border counts")
	cmdRender.Flags().BoolVar(&argsRender.parser.Ignore.Scouts, "ignore-scouts", false, "ignore scout reports")
	cmdRender.Flags().BoolVar(&argsRender.noWarnOnInvalidGrid, "no-warn-on-invalid-grid", false, "disable grid id warnings")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.saveWithTurnId, "save-with-turn-id", false, "add turn id to file name")
	cmdRender.Flags().BoolVar(&argsRender.show.origin, "show-origin", false, "show origin hex")
	cmdRender.Flags().BoolVar(&argsRender.show.shiftMap, "shift-map", false, "shift map up and left")
	cmdRender.Flags().StringVar(&argsRender.clanId, "clan-id", "", "clan for output file names")
	if err := cmdRender.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdRender.Flags().StringVar(&argsRender.paths.data, "data", "data", "path to root of data files")
	cmdRender.Flags().StringVar(&argsRender.originGrid, "origin-grid", "", "grid id to substitute for ##")
	cmdRender.Flags().StringVar(&argsRender.maxTurn.id, "max-turn", "", "last turn to map (yyyy-mm format)")

	cmdServe.Flags().StringVar(&argsServe.paths.assets, "assets", "assets", "path to public assets")
	cmdServe.Flags().StringVar(&argsServe.paths.data, "data", "userdata", "path to root of user data files")
	cmdServe.Flags().StringVar(&argsServe.paths.templates, "templates", "templates", "path to template files")
	cmdServe.Flags().StringVar(&argsServe.server.host, "host", "localhost", "host to serve on")
	cmdServe.Flags().StringVar(&argsServe.server.port, "port", "29631", "port to bind to")

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
