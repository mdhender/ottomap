// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package cerrs implements constant errors.
package cerrs

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrEmptyReport             = Error("empty report")
	ErrInvalidGridCoordinates  = Error("invalid grid coordinates")
	ErrInvalidIndexFile        = Error("invalid index file")
	ErrInvalidInputPath        = Error("invalid input path")
	ErrInvalidOutputPath       = Error("invalid output path")
	ErrInvalidPath             = Error("invalid path")
	ErrInvalidReportFile       = Error("invalid report file")
	ErrMissingFollowsUnit      = Error("missing follows unit")
	ErrMissingIndexFile        = Error("missing index file")
	ErrMissingMovementResults  = Error("missing movement results")
	ErrMissingReportFile       = Error("missing report file")
	ErrMissingStatusLine       = Error("missing status line")
	ErrMultipleFollowsLines    = Error("multiple follows lines")
	ErrMultipleMovementLines   = Error("multiple movement lines")
	ErrMultipleStatusLines     = Error("multiple status lines")
	ErrNoSeparator             = Error("no separator")
	ErrNotATurnReport          = Error("not a turn report")
	ErrNotDirectory            = Error("not a directory")
	ErrNotImplemented          = Error("not implemented")
	ErrNotMovementResults      = Error("not movement results")
	ErrParseFailed             = Error("parse failed")
	ErrSetupExists             = Error("setup.json exists")
	ErrTooManyScoutLines       = Error("too many scout lines")
	ErrTrackingGarrison        = Error("tracking garrison")
	ErrUnableToFindStartingHex = Error("unable to find starting hex")
	ErrUnexpectedNumberOfMoves = Error("unexpected number of moves")
	ErrUnitMovesAndFollows     = Error("unit moves and follows")
)
