// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package directions defines the directions on the map.
package directions

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

// MarshalText implements the encoding.TextMarshaler interface.
// This is needed for marshalling the enum as map keys.
//
// Note that this is called by the json package, unlike the UnmarshalText function.
func (d Direction) MarshalText() (text []byte, err error) {
	return []byte(directionEnumToString[d]), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Direction) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *d, ok = DirectionStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Direction %q", s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// This is needed for unmarshalling the enum as map keys.
//
// Note that this is never called; it just changes the code path in UnmarshalJSON.
func (d Direction) UnmarshalText(text []byte) error {
	panic("!")
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
	// DirectionStringToEnum is a helper map for unmarshalling the enum
	DirectionStringToEnum = map[string]Direction{
		"?":  DUnknown,
		"N":  DNorth,
		"NE": DNorthEast,
		"SE": DSouthEast,
		"S":  DSouth,
		"SW": DSouthWest,
		"NW": DNorthWest,
	}
)
