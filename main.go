// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottomap application
package main

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/parsers/clan_turn"
	"github.com/mdhender/semver"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

var argsSplitInput struct {
	input  string // path to read input files from
	output string // path to create output files in
}

var cmdSplitInput = &cobra.Command{
	Use:   "split-input",
	Short: "Split input files into multiple files",
	Long: `Parse all input files and create one output file per tribe per unit per turn.
These files will be used as input to future steps.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("split-input: input  %s\n", argsSplitInput.input)
		log.Printf("split-input: output %s\n", argsSplitInput.output)
		if ok, _ := isdir(argsSplitInput.input); !ok {
			log.Fatalf("split-input: input %s is not a directory", argsSplitInput.input)
		} else if ok, _ = isdir(argsSplitInput.output); !ok {
			log.Fatalf("split-input: output %s is not a directory", argsSplitInput.output)
		}

		// find all turn report files in the input directory.
		// they will be all files with a name like YEAR-MONTH.TRIBE.input.txt
		var inputFiles []clan_turn.InputFile
		entries, err := os.ReadDir(argsSplitInput.input)
		if err != nil {
			log.Fatal(err)
		} else {
			// tribe input files have the pattern of YYY-MM.TTTT.input.txt.
			pattern := `^\d{3}-\d{2}\.0\d{3}\.input\.txt$`
			re, err := regexp.Compile(pattern)
			if err != nil {
				log.Fatal(err)
			}
			for _, entry := range entries {
				// If the entry is a file and the name matches our pattern...
				if !entry.IsDir() && re.MatchString(entry.Name()) {
					fileName := entry.Name()
					inputFiles = append(inputFiles, clan_turn.InputFile{
						Year:  fileName[0:3],
						Month: fileName[4:6],
						Clan:  fileName[7:11],
						File:  filepath.Join(argsSplitInput.input, fileName),
					})
				}
			}
		}

		for _, inputFile := range inputFiles {
			clan, parseErr := clan_turn.Parse(inputFile)
			if parseErr != nil {
				log.Printf("split-input: %s: error: %v\n", inputFile.File, parseErr)
				err = cerrs.ErrParseFailed
				continue
			}
			log.Printf("split-input: %s: units %3d: transfers %6d: settlements %6d\n", inputFile.File, len(clan.Units), len(clan.Transfers), len(clan.Settlements))
			for _, unit := range clan.Units {
				path := filepath.Join(argsSplitInput.output, fmt.Sprintf("%s-%s.%s.%s.input.txt", inputFile.Year, inputFile.Month, inputFile.Clan, unit.Id))
				log.Printf("split-input: %s: %-8s => %s\n", inputFile.File, unit.Id, path)
				if err := os.WriteFile(path, unit.Text, 0644); err != nil {
					log.Fatalf("split-input: %s: %v\n", inputFile.File, err)
				}
			}
		}

		if err != nil {
			log.Printf("split-input: error parsing input: %v\n", err)
		}
	},
}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of this application",
	Long:  `All software has versions. This is our application's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version.String())
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
