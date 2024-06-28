// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package items

import (
	"encoding/json"
	"fmt"
)

type Item_e int

const (
	None Item_e = iota
	Diamond
	Horses
)

// MarshalJSON implements the json.Marshaler interface.
func (e Item_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Item_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid Item %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Item_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Item(%d)", int(e))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Item_e]string{
		None:    "N/A",
		Diamond: "Diamond",
		Horses:  "HORSES",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Item_e{
		"N/A":     None,
		"Diamond": Diamond,
		"HORSES":  Horses,
	}
)
