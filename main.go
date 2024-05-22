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
	version = semver.Version{Major: 0, Minor: 1, Patch: 0}
)

func main() {
	log.Printf("todo: detect when a unit is created as an after-move action\n")

	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.AddCommand(cmdIndex, cmdMap, cmdParse, cmdSetup, cmdVersion)

	cmdIndex.AddCommand(cmdIndexReports)
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.input, "input", "i", ".", "path to read input from")
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.output, "output", "o", ".", "path to write output to")
	if err := cmdIndexReports.MarkFlagRequired("input"); err != nil {
		log.Fatalf("input: parse: input: mark required: %v\n", err)
	} else if err = cmdIndexReports.MarkFlagRequired("output"); err != nil {
		log.Fatalf("input: parse: output: mark required: %v\n", err)
	}

	cmdMap.Flags().BoolVar(&argsMap.debug.units, "debug-units", false, "enable unit debugging")
	cmdMap.Flags().BoolVar(&argsMap.show.gridNumbers, "show-numbers", false, "show grid numbers")
	cmdMap.Flags().StringVar(&argsMap.clanId, "clan", "", "clan id to process")
	if err := cmdMap.MarkFlagRequired("clan"); err != nil {
		log.Fatalf("map: clan: mark required: %v\n", err)
	}
	cmdMap.Flags().StringVar(&argsMap.config, "config", "config.json", "configuration file to use")
	if err := cmdMap.MarkFlagRequired("config"); err != nil {
		log.Fatalf("map: config: mark required: %v\n", err)
	}
	cmdMap.Flags().StringVar(&argsMap.turnId, "turn", "", "turn to process (yyyy-mm format)")

	cmdParse.PersistentFlags().BoolVar(&argsParse.debug.units, "debug-units", false, "enable unit debugging")
	cmdParse.PersistentFlags().StringVarP(&argsParse.index, "index", "i", ".", "index file to process")
	cmdParse.PersistentFlags().StringVarP(&argsParse.output, "output", "o", ".", "path to write output to")
	cmdParseReports.Flags().BoolVar(&argsParseReports.debug.captureRawText, "capture-raw-text", false, "capture raw text")
	cmdParseReports.Flags().StringVarP(&argsParseReports.gridOrigin, "grid-origin", "g", "OO", "initial grid value for '##'")
	cmdParse.AddCommand(cmdParseReports, cmdParseUnits)

	cmdSetup.Flags().BoolVar(&argsSetup.debug.showSlugs, "debug-slugs", false, "show slugs")
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
