// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// this file defines the data extracted from a turn report

// Index is list of reports that will be sent to the parser.
// It limits the number of files, which is helpful for development.
type Index struct {
	Reports map[string]*Report `json:"reports,omitempty"` // key is path to report
}

// Turn is a single turn.
// Optimistically assumes we may include reports from multiple players.
type Turn struct {
	Year    int                `json:"year,omitempty"`  // 3 digits in reports
	Month   int                `json:"month,omitempty"` // 2 digits in reports
	Players map[string]*Player `json:"players,omitempty"`
	Reports []*Report          `json:"reports,omitempty"`
}

type Report struct {
	Path     string    `json:"path"`             // path to the input file
	Player   *Player   `json:"player,omitempty"` // player that owns the data in the file
	Sections []*Report `json:"sections,omitempty"`
}

// ReportSection captures the text from a single section of the turn report.
// Implementation note: These should be []byte, but string is easier to debug.
type ReportSection struct {
	LocationLine string   `json:"locationLine,omitempty"` // location line from the section
	MovementLine string   `json:"movementLine,omitempty"` // movement line from the section
	ScoutLines   []string `json:"scoutLines,omitempty"`   // scout lines from the section
	StatusLine   string   `json:"statusLine,omitempty"`   // status line from the section
	RawText      string   `json:"rawText,omitempty"`      // this is the un-parsed text of the entire section
}
