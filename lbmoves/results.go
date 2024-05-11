// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lbmoves

import (
	"encoding/json"
	"fmt"
)

type Result int

const (
	StayedInPlace Result = iota
	Blocked
	ExhaustedMovementPoints
	Follows
	Prohibited
	Status
	Succeeded
	Vanished
)

var (
	// helper map for marshalling the enum
	resultEnumToString = map[Result]string{
		StayedInPlace:           "N/A",
		Blocked:                 "Blocked",
		ExhaustedMovementPoints: "Exhausted MPs",
		Follows:                 "Follows",
		Prohibited:              "Prohibited",
		Status:                  "Status",
		Succeeded:               "Succeeded",
		Vanished:                "Vanished",
	}
	// helper map for unmarshalling the enum
	resultStringToEnum = map[string]Result{
		"N/A":           StayedInPlace,
		"Blocked":       Blocked,
		"Exhausted MPs": ExhaustedMovementPoints,
		"Follows":       Follows,
		"Prohibited":    Prohibited,
		"Status":        Status,
		"Succeeded":     Succeeded,
		"Vanished":      Vanished,
	}
)

// MarshalJSON implements the json.Marshaler interface.
func (r Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(resultEnumToString[r])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Result) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *r, ok = resultStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Result %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (r Result) String() string {
	if str, ok := resultEnumToString[r]; ok {
		return str
	}
	return fmt.Sprintf("Result(%d)", int(r))
}
