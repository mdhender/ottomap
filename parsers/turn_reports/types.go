// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package tribe_turn implements a parser for tribe turn files.
package turn_reports

type InputFile struct {
	Year  string // three digits
	Month string // two digits
	Clan  string // four digits
	File  string // path/Year-Month.Clan.input.txt
}

type Turn struct {
	Clan        string  // four digits
	Units       []*Unit // text from all the unit sections
	Transfers   string  // text from the Transfers section
	Settlements string  // text from the Settlements section
	Tail        string  // text following breaking parsing error
}

type Unit struct {
	Id   string // four digits plus optional code plus optional number
	Text []byte
}

// Hex is a hex on the grid map
type Hex struct {
	Grid string // may be ## or NN, depending on if the GM is publishing the grid or not
	X    int
	Y    int
}
