// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottomap application
package main

import (
	"github.com/mdhender/semver"
	"github.com/spf13/cobra"
	"log"
	"os"
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
	cmdRoot.AddCommand(cmdVersion)
	cmdRoot.AddCommand(cmdSplitInput)
	cmdSplitInput.Flags().StringVar(&argsSplitInput.input, "input", ".", "path to read input from")
	cmdSplitInput.Flags().StringVar(&argsSplitInput.output, "output", ".", "path to write output to")
	if err := cmdSplitInput.MarkFlagRequired("input"); err != nil {
		log.Fatalf("split-input: marking input flag as required: %v", err)
	} else if err := cmdSplitInput.MarkFlagRequired("output"); err != nil {
		log.Fatalf("split-input: marking output flag as required: %v", err)
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

func isdir(path string) (bool, error) {
	sb, err := os.Stat(path)
	if err != nil {
		return false, err
	} else if !sb.IsDir() {
		return false, nil
	}
	return true, nil
}
