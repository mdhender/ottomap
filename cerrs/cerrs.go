// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package cerrs implements constant errors.
package cerrs

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrNoSeparator    = Error("no separator")
	ErrNotImplemented = Error("not implemented")
	ErrParseFailed    = Error("parse failed")
)
