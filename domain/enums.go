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

// UnitType is an enum for the type of unit.
// Having Tribe as a unit type makes the unit code easier to understand.
type UnitType int

const (
	TRIBE UnitType = iota
	ELEMENT
	COURIER
	GARRISON
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
		TRIBE:    "TRIBE",
		ELEMENT:  "ELEMENT",
		COURIER:  "COURIER",
		GARRISON: "GARRISON",
	}
	// helper map for unmarshalling the enum
	unitTypeStringToEnum = map[string]UnitType{
		"TRIBE":    TRIBE,
		"ELEMENT":  ELEMENT,
		"COURIER":  COURIER,
		"GARRISON": GARRISON,
	}
)
