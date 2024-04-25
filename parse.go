// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/spf13/cobra"
	"strings"
)

var argsParse struct {
	input  string // path to read input files from
	output string // path to create output files in
}

var cmdParse = &cobra.Command{
	Use: "parse",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		//log.Printf("parse: input  %s\n", argsParse.input)
		if strings.TrimSpace(argsParse.input) != argsParse.input {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidInputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsParse.input); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidInputPath, err)
		} else {
			argsParse.input = path
		}

		//log.Printf("parse: output %s\n", argsParse.output)
		if strings.TrimSpace(argsParse.output) != argsParse.output {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
		} else if path, err := abspath(argsParse.output); err != nil {
			return errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
		} else {
			argsParse.output = path
		}

		//log.Printf("parse: input  %s\n", argsParse.input)
		//log.Printf("parse: output %s\n", argsParse.output)

		return nil
	},
}
