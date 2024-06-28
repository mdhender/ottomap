// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wbmoves

import (
	"encoding/json"
	"fmt"
)

// Edge_e is an enum for the edge of a terrain.
type Edge_e int

const (
	ENone Edge_e = iota
	EFord
	EPass
	ERiver
	EStoneRoad
)

// MarshalJSON implements the json.Marshaler interface.
func (e Edge_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(edgeEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Edge_e) UnmarshalJSON(data []byte) error {
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
func (e Edge_e) String() string {
	if str, ok := edgeEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Edge(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	edgeEnumToString = map[Edge_e]string{
		ENone:      "",
		EFord:      "Ford",
		EPass:      "Pass",
		ERiver:     "River",
		EStoneRoad: "Stone Road",
	}
	// helper map for unmarshalling the enum
	edgeStringToEnum = map[string]Edge_e{
		"":           ENone,
		"Ford":       EFord,
		"Pass":       EPass,
		"River":      ERiver,
		"Stone Road": EStoneRoad,
	}
)

// Resource_e is an enum for resources
type Resource_e int

const (
	RNone Resource_e = iota
	RCoal
	RCopperOre
	RDiamond
	RFrankincense
	RGold
	RIronOre
	RJade
	RKaolin
	RLeadOre
	RLimestone
	RNickelOre
	RPearls
	RPyrite
	RRubies
	RSalt
	RSilver
	RSulphur
	RTinOre
	RVanadiumOre
	RZincOre
)

var (
	// helper map for marshalling the enum
	resourceEnumToString = map[Resource_e]string{
		RNone:         "",
		RCoal:         "Coal",
		RCopperOre:    "Copper Ore",
		RDiamond:      "Diamond",
		RFrankincense: "Frankincense",
		RGold:         "Gold",
		RIronOre:      "Iron Ore",
		RJade:         "Jade",
		RKaolin:       "Kaolin",
		RLeadOre:      "Lead Ore",
		RLimestone:    "Limestone",
		RNickelOre:    "Nickel Ore",
		RPearls:       "Pearls",
		RPyrite:       "Pyrite",
		RRubies:       "Rubies",
		RSalt:         "Salt",
		RSilver:       "Silver",
		RSulphur:      "Sulphur",
		RTinOre:       "Tin Ore",
		RVanadiumOre:  "Vanadium Ore",
		RZincOre:      "Zinc Ore",
	}
	// helper map for unmarshalling the enum
	resourceStringToEnum = map[string]Resource_e{
		"":             RNone,
		"Coal":         RCoal,
		"Copper Ore":   RCopperOre,
		"Diamond":      RDiamond,
		"Frankincense": RFrankincense,
		"Gold":         RGold,
		"Iron Ore":     RIronOre,
		"Jade":         RJade,
		"Kaolin":       RKaolin,
		"Lead Ore":     RLeadOre,
		"Limestone":    RLimestone,
		"Nickel Ore":   RNickelOre,
		"Pearls":       RPearls,
		"Pyrite":       RPyrite,
		"Rubies":       RRubies,
		"Salt":         RSalt,
		"Silver":       RSilver,
		"Sulphur":      RSulphur,
		"Tin Ore":      RTinOre,
		"Vanadium Ore": RVanadiumOre,
		"Zinc Ore":     RZincOre,
	}
)

// MarshalJSON implements the json.Marshaler interface.
func (r Resource_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(resourceEnumToString[r])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Resource_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *r, ok = resourceStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Resource_e %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (r Resource_e) String() string {
	if str, ok := resourceEnumToString[r]; ok {
		return str
	}
	return fmt.Sprintf("Resource_e(%d)", int(r))
}

type Sighted_e int

const (
	Land Sighted_e = iota
	Water
)

type UnitMovement_e int

const (
	UMUnknown UnitMovement_e = iota
	UMFleet
	UMFollows
	UMGoto
	UMScouts
	UMStill
	UMTribe
)

// MarshalJSON implements the json.Marshaler interface.
func (e UnitMovement_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(unitMoveEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *UnitMovement_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = unitMoveStringToEnum[s]; !ok {
		return fmt.Errorf("invalid UnitMovement %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e UnitMovement_e) String() string {
	if str, ok := unitMoveEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("UnitMovement(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	unitMoveEnumToString = map[UnitMovement_e]string{
		UMUnknown: "N/A",
		UMFleet:   "Fleet",
		UMFollows: "Follows",
		UMGoto:    "Goto",
		UMScouts:  "Scout",
		UMStill:   "Still",
		UMTribe:   "Tribe",
	}
	// helper map for unmarshalling the enum
	unitMoveStringToEnum = map[string]UnitMovement_e{
		"N/A":     UMUnknown,
		"Fleet":   UMFleet,
		"Follows": UMFollows,
		"Goto":    UMGoto,
		"Scout":   UMScouts,
		"Still":   UMStill,
		"Tribe":   UMTribe,
	}
)
