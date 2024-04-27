// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package locations

// this file defines all the nodes returned by the parser

type Location struct {
	CurrentHex  Hex
	PreviousHex Hex
}

type Hex struct {
	NA         bool
	Grid       string
	Col        string
	Row        string
	Settlement string
	Terrain    string
	Edges      [6]string
	Contains   string
	Found      string
}
