// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

import (
	"encoding/json"
	"fmt"
)

type KindOfUnit int

const (
	TRIBE KindOfUnit = iota
	ELEMENT
	COURIER
	GARRISON
)

// MarshalJSON implements the json.Marshaler interface.
func (k KindOfUnit) MarshalJSON() ([]byte, error) {
	return json.Marshal(kindOfUnitString[k])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (k *KindOfUnit) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *k, ok = kindOfUnitId[s]; !ok {
		return fmt.Errorf("invalid KindOfUnit %q", s)
	}
	return nil
}

var (
	kindOfUnitString = map[KindOfUnit]string{
		TRIBE:    "TRIBE",
		ELEMENT:  "ELEMENT",
		COURIER:  "COURIER",
		GARRISON: "GARRISON",
	}
	kindOfUnitId = map[string]KindOfUnit{
		"TRIBE":    TRIBE,
		"ELEMENT":  ELEMENT,
		"COURIER":  COURIER,
		"GARRISON": GARRISON,
	}
)

// String implements the Stringer interface.
func (k KindOfUnit) String() string {
	if str, ok := kindOfUnitString[k]; ok {
		return str
	}
	return fmt.Sprintf("KindOfUnit(%d)", int(k))
}
