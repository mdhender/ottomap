// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// Player captures the state of a Player at the end of a single turn.
type Player struct {
	Id          string                 `json:"id,omitempty"` // 0138
	Clans       map[string]*Clan       `json:"clans,omitempty"`
	Settlements map[string]*Settlement `json:"settlements,omitempty"`
	Units       map[string]*Unit       `json:"units,omitempty"`
}
