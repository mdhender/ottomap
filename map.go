// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/domain"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var argsMap struct {
	input  string // parsed report file to process
	output string // path to create map in
	debug  struct {
		units bool
	}
}

var cmdMap = &cobra.Command{
	Use:   "map",
	Short: "Create a map from a report",
	Long:  `Load a parsed report and create a map.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var reports []*domain.Report
		if data, err := os.ReadFile(argsMap.input); err != nil {
			log.Fatalf("map: failed to read input file: %v", err)
		} else if err = json.Unmarshal(data, &reports); err != nil {
			log.Fatalf("map: failed to unmarshal input file: %v", err)
		}

		if strings.TrimSpace(argsMap.output) != argsMap.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsMap.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsMap.output = path
		}

		log.Printf("map: input  %s\n", argsMap.input)
		log.Printf("map: input  %d records\n", len(reports))

		log.Printf("map: output %s\n", argsMap.output)

		return nil
	},
}
