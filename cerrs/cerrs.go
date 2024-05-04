// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package cerrs implements constant errors.
package cerrs

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrInvalidGridCoordinates = Error("invalid grid coordinates")
	ErrInvalidIndexFile       = Error("invalid index file")
	ErrInvalidInputPath       = Error("invalid input path")
	ErrInvalidOutputPath      = Error("invalid output path")
	ErrInvalidPath            = Error("invalid path")
	ErrMissingIndexFile       = Error("missing index file")
	ErrMissingMovementResults = Error("missing movement results")
	ErrNoSeparator            = Error("no separator")
	ErrNotATurnReport         = Error("not a turn report")
	ErrNotDirectory           = Error("not a directory")
	ErrNotImplemented         = Error("not implemented")
	ErrParseFailed            = Error("parse failed")
)
