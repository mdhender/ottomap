// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tw

type Layout_t struct {
	Content any
}

type TurnList_t struct {
	Turns []string
}

type TurnDetails_t struct {
	Id    string
	Clans []string
}

type TurnReportDetails_t struct {
	Id    string
	Clan  string
	Map   string // set when there is a map file
	Units []UnitDetails_t
}

type UnitDetails_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}
