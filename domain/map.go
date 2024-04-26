// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// this file defines the features of a map in the game

type GridHex struct {
	Grid string `json:"grid,omitempty"` // "##", NN, or "N/A"
	Col  int    `json:"col,omitempty"`  // 00..21
	Row  int    `json:"row,omitempty"`  // 00..30
	Hex  *Hex   `json:"hex,omitempty"`
}

type Hex struct {
	Col        int         `json:"col,omitempty"` // coordinates on the big map
	Row        int         `json:"row,omitempty"` // coordinates on the big map
	Terrain    string      `json:"terrain,omitempty"`
	Edges      [6]string   `json:"edges,omitempty"`
	Settlement *Settlement `json:"settlement,omitempty"`
}

// Step is a unit's attempt to move in a given direction.
// It captures the results of that attempt.
type Step struct {
	Direction string `json:"direction,omitempty"`
	Results   string `json:"results,omitempty"`
	RawText   string `json:"rawText,omitempty"`
}
