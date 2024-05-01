// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// this file implements enums for the domain.
// code shouldn't be in the domain package, but we really want
// stringers and json support for our enums, so it's stuffed in here.
// a different person would recognize that the enums and their
// implementation should be moved to separate packages.

import (
	"encoding/json"
	"fmt"
)

// Direction is an enum for the direction
type Direction int

const (
	DUnknown Direction = iota
	DNorth
	DNorthEast
	DSouthEast
	DSouth
	DSouthWest
	DNorthWest
)

// MarshalJSON implements the json.Marshaler interface.
func (d Direction) MarshalJSON() ([]byte, error) {
	return json.Marshal(directionEnumToString[d])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Direction) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *d, ok = directionStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Direction %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (d Direction) String() string {
	if str, ok := directionEnumToString[d]; ok {
		return str
	}
	return fmt.Sprintf("Direction(%d)", int(d))
}

var (
	// helper map for marshalling the enum
	directionEnumToString = map[Direction]string{
		DUnknown:   "?",
		DNorth:     "N",
		DNorthEast: "NE",
		DSouthEast: "SE",
		DSouth:     "S",
		DSouthWest: "SW",
		DNorthWest: "NW",
	}
	// helper map for unmarshalling the enum
	directionStringToEnum = map[string]Direction{
		"?":  DUnknown,
		"N":  DNorth,
		"NE": DNorthEast,
		"SE": DSouthEast,
		"S":  DSouth,
		"SW": DSouthWest,
		"NW": DNorthWest,
	}
)

// Edge is an enum for the edge of a terrain.
type Edge int

const (
	ENone Edge = iota
	EFord
	EPass
	ERiver
)

// MarshalJSON implements the json.Marshaler interface.
func (e Edge) MarshalJSON() ([]byte, error) {
	return json.Marshal(edgeEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Edge) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = edgeStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Edge %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Edge) String() string {
	if str, ok := edgeEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Edge(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	edgeEnumToString = map[Edge]string{
		ENone:  "",
		EFord:  "Ford",
		EPass:  "Pass",
		ERiver: "River",
	}
	// helper map for unmarshalling the enum
	edgeStringToEnum = map[string]Edge{
		"":      ENone,
		"Ford":  EFord,
		"Pass":  EPass,
		"River": ERiver,
	}
)

// ReportSectionType is an enum for the type of report section.
type ReportSectionType int

const (
	RSUnknown ReportSectionType = iota
	RSUnit
	RSSettlements
	RSTransfers
)

// MarshalJSON implements the json.Marshaler interface.
func (t ReportSectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(reportSectionTypeEnumToString[t])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *ReportSectionType) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *t, ok = reportSectionTypeStringToEnum[s]; !ok {
		return fmt.Errorf("invalid ReportSectionType %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (t ReportSectionType) String() string {
	if str, ok := reportSectionTypeEnumToString[t]; ok {
		return str
	}
	return fmt.Sprintf("ReportSectionType(%d)", int(t))
}

var (
	// helper map for marshalling the enum
	reportSectionTypeEnumToString = map[ReportSectionType]string{
		RSUnknown:     "Unknown",
		RSUnit:        "Unit",
		RSSettlements: "Settlements",
		RSTransfers:   "Transfers",
	}
	// helper map for unmarshalling the enum
	reportSectionTypeStringToEnum = map[string]ReportSectionType{
		"Unknown":     RSUnknown,
		"Unit":        RSUnit,
		"Settlements": RSSettlements,
		"Transfers":   RSTransfers,
	}
)

// Terrain is an enum for the terrain
type Terrain int

const (
	TUnknown Terrain = iota
	TConiferHills
	TGrassyHills
	TOcean
	TPrairie
	TRockyHills
	TSwamp
)

// MarshalJSON implements the json.Marshaler interface.
func (d Terrain) MarshalJSON() ([]byte, error) {
	return json.Marshal(terrainEnumToString[d])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Terrain) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *d, ok = terrainStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Terrain %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (d Terrain) String() string {
	if str, ok := terrainEnumToString[d]; ok {
		return str
	}
	return fmt.Sprintf("Terrain(%d)", int(d))
}

var (
	// helper map for marshalling the enum
	terrainEnumToString = map[Terrain]string{
		TUnknown:      "?",
		TConiferHills: "CH",
		TGrassyHills:  "GH",
		TOcean:        "O",
		TPrairie:      "PR",
		TRockyHills:   "RH",
		TSwamp:        "SW",
	}
	// helper map for unmarshalling the enum
	terrainStringToEnum = map[string]Terrain{
		"?":  TUnknown,
		"CH": TConiferHills,
		"GH": TGrassyHills,
		"O":  TOcean,
		"PR": TPrairie,
		"RH": TRockyHills,
		"SW": TSwamp,
	}
)

// UnitType is an enum for the type of unit.
// Having Tribe as a unit type makes the unit code easier to understand.
type UnitType int

const (
	UTUnknown UnitType = iota
	UTTribe
	UTCourier
	UTElement
	UTGarrison
)

// MarshalJSON implements the json.Marshaler interface.
func (t UnitType) MarshalJSON() ([]byte, error) {
	return json.Marshal(unitTypeEnumToString[t])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *UnitType) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *t, ok = unitTypeStringToEnum[s]; !ok {
		return fmt.Errorf("invalid UnitType %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (t UnitType) String() string {
	if str, ok := unitTypeEnumToString[t]; ok {
		return str
	}
	return fmt.Sprintf("UnitType(%d)", int(t))
}

var (
	// helper map for marshalling the enum
	unitTypeEnumToString = map[UnitType]string{
		UTUnknown:  "Unknown",
		UTTribe:    "Tribe",
		UTCourier:  "Courier",
		UTElement:  "Element",
		UTGarrison: "Garrison",
	}
	// helper map for unmarshalling the enum
	unitTypeStringToEnum = map[string]UnitType{
		"Unknown":  UTUnknown,
		"Tribe":    UTTribe,
		"Courier":  UTCourier,
		"Element":  UTElement,
		"Garrison": UTGarrison,
	}
)
