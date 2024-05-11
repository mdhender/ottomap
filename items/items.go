// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package items

import (
	"encoding/json"
	"fmt"
)

type Item int

const (
	None Item = iota
	Diamond
	Horses
)

// MarshalJSON implements the json.Marshaler interface.
func (e Item) MarshalJSON() ([]byte, error) {
	return json.Marshal(itemEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Item) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = itemStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Item %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Item) String() string {
	if str, ok := itemEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Item(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	itemEnumToString = map[Item]string{
		None:    "N/A",
		Diamond: "Diamond",
		Horses:  "HORSES",
	}
	// helper map for unmarshalling the enum
	itemStringToEnum = map[string]Item{
		"N/A":     None,
		"Diamond": Diamond,
		"HORSES":  Horses,
	}
)
