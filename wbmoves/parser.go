// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wbmoves

import (
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
)

//go:generate pigeon -o grammar.go grammar.peg

type FleetMovement struct {
	FleetId string
	Winds   struct {
		Strength domain.WindStrength_e
		From     directions.Direction
	}
}
