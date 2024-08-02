// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tw

type Layout_t struct {
	Site    Site_t
	Content any
}

type Site_t struct {
	Title string
}

type Clans_t struct {
	Id    string   // id of the player's clan
	Clans []string // clans that the player has uploaded reports for
}

type ClanDetail_t struct {
	Id    string
	Maps  []string
	Turns []string
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
