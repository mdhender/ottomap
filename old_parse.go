// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var argsOldParse struct {
	index  string // index file to process
	output string // path to create output files in
	debug  struct {
		units bool
	}
}

var cmdOldParse = &cobra.Command{
	Use: "parse",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("old-parse: index  %s\n", argsOldParse.index)
		if sb, err := os.Stat(argsOldParse.index); err != nil && os.IsNotExist(err) {
			return cerrs.ErrInvalidIndexFile
		} else if os.IsNotExist(err) {
			return cerrs.ErrMissingIndexFile
		} else if sb.IsDir() {
			return cerrs.ErrInvalidIndexFile
		}

		log.Printf("old-parse: output %s\n", argsOldParse.output)
		if strings.TrimSpace(argsOldParse.output) != argsOldParse.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsOldParse.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsOldParse.output = path
		}

		log.Printf("old-parse: index  %s\n", argsOldParse.index)
		log.Printf("old-parse: output %s\n", argsOldParse.output)

		return nil
	},
}
