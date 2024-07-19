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
	version = semver.Version{Major: 0, Minor: 13, Patch: 0}
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// todo: detect when a unit is created as an after-move action

	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.AddCommand(cmdSammy, cmdVersion)

	cmdSammy.Flags().BoolVar(&argsSammy.debug.dumpAllTiles, "debug-dump-all-tiles", false, "dump all tiles")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.dumpAllTurns, "debug-dump-all-turns", false, "dump all turns")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.maps, "debug-maps", false, "enable maps debugging")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.nodes, "debug-nodes", false, "enable node debugging")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.parser, "debug-parser", false, "enable parser debugging")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.sections, "debug-sections", false, "enable sections debugging")
	cmdSammy.Flags().BoolVar(&argsSammy.debug.steps, "debug-steps", false, "enable step debugging")
	cmdSammy.Flags().BoolVar(&argsSammy.mapper.Dump.BorderCounts, "dump-border-counts", false, "dump border counts")
	cmdSammy.Flags().BoolVar(&argsSammy.parser.Ignore.Scouts, "ignore-scouts", false, "ignore scout reports")
	cmdSammy.Flags().BoolVar(&argsSammy.noWarnOnInvalidGrid, "no-warn-on-invalid-grid", false, "disable grid id warnings")
	cmdSammy.Flags().BoolVar(&argsSammy.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdSammy.Flags().BoolVar(&argsSammy.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdSammy.Flags().BoolVar(&argsSammy.show.origin, "show-origin", false, "show origin hex")
	cmdSammy.Flags().StringVar(&argsSammy.clanId, "clan-id", "", "clan for output file names")
	if err := cmdSammy.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdSammy.Flags().StringVar(&argsSammy.paths.data, "data", "data", "path to root of data files")
	cmdSammy.Flags().StringVar(&argsSammy.originGrid, "origin-grid", "", "grid id to substitute for ##")
	cmdSammy.Flags().StringVar(&argsSammy.maxTurn.id, "max-turn", "", "last turn to map (yyyy-mm format)")

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
