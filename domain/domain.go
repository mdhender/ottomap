// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package domain defines the data types used to parse the turn reports
// and (in an unknowable, far away future) generate maps.
package domain

// Index is list of turn report files that will be sent to the parser.
// It limits the number of files, which is helpful for development.
//
// NB: The key for the ReportFiles map is the name of the file without the path.
type Index struct {
	ReportFiles map[string]*ReportFile `json:"reportFiles"`
}

// ReportFile is a single turn report file.
//
// The application handles parsing in stages. The first stage opens the report file
// (which must be plain text) and splits it into sections, one section per unit.
// Future stages will translate the raw text in the sections to usable turn data.
type ReportFile struct {
	Path     string           `json:"path"`             // path to the input file
	Name     string           `json:"name"`             // name of the input file
	Player   *Player          `json:"player,omitempty"` // optional link to information on player that owns the data in the file
	Sections []*ReportSection `json:"sections,omitempty"`
}

// ReportSection captures the text from a single section of the turn report.
//
// NB: These should be []byte, but string is easier to debug.
type ReportSection struct {
	LocationLine string   `json:"locationLine,omitempty"` // location line from the section
	MovementLine string   `json:"movementLine,omitempty"` // movement line from the section
	ScoutLines   []string `json:"scoutLines,omitempty"`   // scout lines from the section
	StatusLine   string   `json:"statusLine,omitempty"`   // status line from the section
	RawText      string   `json:"rawText"`                // this is the un-parsed text of the entire section
}

// Turn defines the data extracted from the turn report.
//
// NB: Defined separately so we may include reports from multiple players in the future.
type Turn struct {
	Year  int   `json:"year"`  // 3 digit year (e.g. 901)
	Month int   `json:"month"` // 2 digit month (e.g. 05)
	Clan  *Clan `json:"clan,omitempty"`
}

// Clan defines the units in a single hierarchy.
//
// Tribes are the highest level; they are identified by a 4-digit number.
// All tribes in a clan share the same last three digits (0138, 1138,
// 2138, etc.). The tribe that starts with a zero is special; it is the main
// tribe for the clan.
//
// NB: I'm using Clan here instead of Tribe to make the parsing easier
// for me to understand.
type Clan struct {
	Id    string           `json:"id"` // 4 digit string (e.g. 1138)
	Units map[string]*Unit `json:"units"`
}

// Unit is a unit which reports back up to the Clan.
//
// NB: All units that belong to a tribe share a common prefix which is
// just the 4-digit Id for the tribe.
type Unit struct {
	Id       string   `json:"id"` // 4 or 6 char string (e.g. 0138 or 1138c3)
	Type     UnitType `json:"type"`
	Location *GridHex `json:"location,omitempty"`
	Status   string   `json:"status,omitempty"` // will every unit have a status?
}

// Settlement is a settlement.
type Settlement struct {
	Id       string   `json:"id"` // maybe name?
	Name     string   `json:"name"`
	Location *GridHex `json:"location"`
}

// GridHex is a hex on a single grid of the map.
// It has three components: Grid, Column, and Row.
//
// Grid identifies the location of the map on the "big map."
// The big map contains 26 columns by 26 rows of smaller maps
// called "grids." Each grid is identified by a two-letter code
// representing its column and row. The grid at the top left of
// the big map has an Id of "AA" and the one at the bottom right
// is "ZZ."
//
// NB: The GM will sometimes hide the actual location of the grid,
// usually when a player is just starting out. In that case, the
// grid will show as "##" in the turn reports. Also, we tend to
// use "##" in examples where the actual location is not important.
//
// The Column and Row values for the hex are relative to the grid.
// In the reports, they are shown as a single four-digit number,
// with the column displayed first. (If it helps, you can think of
// it as "column * 100 + row" with leading zeroes.)
//
// Column starts at 00 on the left side of the map. There are 30
// columns on each grid, so the value ranges from 00 to 29.
//
// Row starts at 00 on the top of the map. There are 21 rows on
// each grid, so the value ranges from 00 to 20.
//
// The hex in the top left corner of a grid is "## 0000" and the
// hex in the bottom right corner is "## 2920." (The reports
// always put a space between the grid and the column/row numbers.)
//
// NB: Sometimes the location isn't known or available. When that
// happens, the location is shown as "N/A" in the reports. We implement
// that by setting the Grid, Column, and Row to "", 0, and 0. That just
// happens to be the zero-value for the struct in Go, so we're happy.
//
// NB: See https://tribenet.wiki/mapping/grid for actual details on this system.
type GridHex struct {
	Grid   string `json:"grid,omitempty"` // "##", NN, or "N/A"
	Column int    `json:"col,omitempty"`  // 00..20
	Row    int    `json:"row,omitempty"`  // 00..29
	Hex    *Hex   `json:"hex,omitempty"`  // optional details for the hex
}

// Hex captures the details needed to map out the hex.
//
// Be aware that Column and Row are the coordinates on an imaginary map,
// not on the grid. That imaginary map is NOT the big map. It's a magical
// thing centered on the clan's first hex. There's a lot of angry code
// that needs to be written to allow multiple clans to exist in this
// magical coordinate system. That code probably will never happen.
type Hex struct {
	Column     int         `json:"col,omitempty"` // coordinates on the big map
	Row        int         `json:"row,omitempty"` // coordinates on the big map
	Terrain    string      `json:"terrain,omitempty"`
	Edges      [6]string   `json:"edges,omitempty"`
	Settlement *Settlement `json:"settlement,omitempty"`
}

// Step captures data from a unit's attempt to move from one hex to another.
//
// Results include terrain and edge features, units encountered,
// settlements, and other things of interest. Note that even a move
// that fails because of M.P.'s can reveal what terrain is in that
// destination hex.
//
// NB: The From and To hexes are helpful when plotting moves. We have to
// take some care to avoid duplicates. Imagine a unit moves N S N S N.
// The naive implementation creates a chain of 5 hexes. There should be
// only two.
type Step struct {
	From      *GridHex `json:"from,omitempty"`
	To        *GridHex `json:"to,omitempty"`
	Direction string   `json:"direction,omitempty"`
	Results   string   `json:"results,omitempty"`
	RawText   string   `json:"rawText,omitempty"`
}
