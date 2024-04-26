// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// this file defines the details of what we're reporting on for each player

// Clan defines the units in a single hierarchy (all the units that share a common prefix).
// One of those units will be a tribe.
// I'm using Clan here instead of Tribe to make the parsing easier for me to understand.
type Clan struct {
	Id    string           `json:"id,omitempty"` // 1138
	Units map[string]*Unit `json:"units,omitempty"`
}

// Unit is a unit, something other than a Settlement which reports back up to the Clan.
type Unit struct {
	Id       string     `json:"id,omitempty"` // 1138c3
	Kind     KindOfUnit `json:"kind,omitempty"`
	Location *GridHex   `json:"location,omitempty"`
	Status   string     `json:"status,omitempty"` // will every unit have a status?
}

// Settlement is a settlement.
type Settlement struct {
	Id       string   `json:"id,omitempty"` // maybe name?
	Name     string   `json:"name,omitempty"`
	Location *GridHex `json:"location,omitempty"`
}
