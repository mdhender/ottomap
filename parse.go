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

var argsParse struct {
	index  string // index file to process
	output string // path to create output files in
}

var cmdParse = &cobra.Command{
	Use: "parse",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("parse: index  %s\n", argsParse.index)
		if sb, err := os.Stat(argsParse.index); err != nil && os.IsNotExist(err) {
			return cerrs.ErrInvalidIndexFile
		} else if os.IsNotExist(err) {
			return cerrs.ErrMissingIndexFile
		} else if sb.IsDir() {
			return cerrs.ErrInvalidIndexFile
		}

		log.Printf("parse: output %s\n", argsParse.output)
		if strings.TrimSpace(argsParse.output) != argsParse.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsParse.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsParse.output = path
		}

		log.Printf("parse: index  %s\n", argsParse.index)
		log.Printf("parse: output %s\n", argsParse.output)

		return nil
	},
}
