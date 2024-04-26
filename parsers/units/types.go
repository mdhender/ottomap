// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package units

import "fmt"

type InputFile struct {
	Year  string // three digits
	Month string // two digits
	Clan  string // four digits
	Unit  string // four digits, maybe followed by letter and single digit
	File  string // path/YEAR-MONTH.CLAN.UNIT.input.txt
}

type Clan struct {
	Id    string
	Units []*Unit
}

type Unit struct {
	Id        string     // four digits plus optional code plus optional number
	Started   *Hex       // hex the unit started the turn in, nil if that is missing
	Finished  *Hex       // hex the unit completed the turn in
	Status    string     // status text
	Movements *Movements // unit's movements
	Scouts    []*Scout   // scouting results
}

// Hex is a hex on the grid map.
type Hex struct {
	Grid       string // may be ## or NN, depending on if the GM is publishing the grid or not
	Col        int
	Row        int
	Settlement string
	Terrain    string
	Edges      [6]string
	Contains   []string
	Found      []string
}

func (h *Hex) String() string {
	if h == nil {
		return "N/A"
	}
	return fmt.Sprintf("%s %04d", h.Grid, h.Col*100+h.Row)
}

type Movements struct {
	Moves  []*Movement
	Failed struct {
		Direction string
		Edge      string
		Terrain   string
		Text      string
	}
	Found []string
}

type Movement struct {
	Direction string
	Result    string
	Found     []string
	Raw       string
}

// Scout is
type Scout struct {
	Text string // scouting text
}
