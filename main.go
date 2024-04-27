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
	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.AddCommand(cmdIndex, cmdParse, cmdVersion)

	cmdIndex.AddCommand(cmdIndexReports)
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.input, "input", "i", ".", "path to read input from")
	cmdIndexReports.Flags().StringVarP(&argsIndexReports.output, "output", "o", ".", "path to write output to")
	if err := cmdIndexReports.MarkFlagRequired("input"); err != nil {
		log.Fatalf("input: parse: input: mark required: %v\n", err)
	} else if err = cmdIndexReports.MarkFlagRequired("output"); err != nil {
		log.Fatalf("input: parse: output: mark required: %v\n", err)
	}

	cmdParse.PersistentFlags().StringVarP(&argsParse.input, "input", "i", ".", "path to read input from")
	cmdParse.PersistentFlags().StringVarP(&argsParse.output, "output", "o", ".", "path to write output to")
	cmdParse.AddCommand(cmdParseReports, cmdParseUnits)

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
