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
func (k UnitType) MarshalJSON() ([]byte, error) {
	return json.Marshal(unitTypeEnumToString[k])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (k *UnitType) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *k, ok = unitTypeStringToEnum[s]; !ok {
		return fmt.Errorf("invalid UnitType %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (k UnitType) String() string {
	if str, ok := unitTypeEnumToString[k]; ok {
		return str
	}
	return fmt.Sprintf("UnitType(%d)", int(k))
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
