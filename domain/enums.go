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

// Edge is an enum for the edge of a terrain.
type Edge int

const (
	ENone Edge = iota
	EFord
	EPass
	ERiver
	EStoneRoad
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
		ENone:      "",
		EFord:      "Ford",
		EPass:      "Pass",
		ERiver:     "River",
		EStoneRoad: "Stone Road",
	}
	// helper map for unmarshalling the enum
	edgeStringToEnum = map[string]Edge{
		"":           ENone,
		"Ford":       EFord,
		"Pass":       EPass,
		"River":      ERiver,
		"Stone Road": EStoneRoad,
	}
)

// MoveStatus is an enum for the outcome of a movement step.
type MoveStatus int

const (
	MSSucceeded MoveStatus = iota
	MSBlocked
	MSExhausted
)

// MarshalJSON implements the json.Marshaler interface.
func (e MoveStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(moveStatusEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *MoveStatus) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = moveStatusStringToEnum[s]; !ok {
		return fmt.Errorf("invalid MoveStatus %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e MoveStatus) String() string {
	if str, ok := moveStatusEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("MoveStatus(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	moveStatusEnumToString = map[MoveStatus]string{
		MSBlocked:   "Blocked",
		MSExhausted: "Exhausted",
		MSSucceeded: "Succeeded",
	}
	// helper map for unmarshalling the enum
	moveStatusStringToEnum = map[string]MoveStatus{
		"Blocked":   MSBlocked,
		"Exhausted": MSExhausted,
		"Succeeded": MSSucceeded,
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

// Resource is an enum for resources
type Resource int

const (
	RNone Resource = iota
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
	resourceEnumToString = map[Resource]string{
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
	resourceStringToEnum = map[string]Resource{
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
func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(resourceEnumToString[r])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Resource) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *r, ok = resourceStringToEnum[s]; !ok {
		return fmt.Errorf("invalid Resource %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (r Resource) String() string {
	if str, ok := resourceEnumToString[r]; ok {
		return str
	}
	return fmt.Sprintf("Resource(%d)", int(r))
}

// Terrain is an enum for the terrain
type Terrain int

const (
	// TBlank must be the first enum value or the map will not render
	TBlank Terrain = iota
	TAlps
	TAridHills
	TAridTundra
	TBrushFlat
	TBrushHills
	TConiferHills
	TDeciduous
	TDeciduousHills
	TDesert
	TGrassyHills
	TGrassyHillsPlateau
	THighSnowyMountains
	TJungle
	TJungleHills
	TLake
	TLowAridMountains
	TLowConiferMountains
	TLowJungleMountains
	TLowSnowyMountains
	TLowVolcanicMountains
	TOcean
	TPolarIce
	TPrairie
	TPrairiePlateau
	TRockyHills
	TSnowyHills
	TSwamp
	TTundra
)

// NumberOfTerrainTypes must be updated if we add new terrain types
const NumberOfTerrainTypes = int(TTundra + 1)

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

func StringToTerrain(s string) (Terrain, bool) {
	if d, ok := terrainStringToEnum[s]; ok {
		return d, ok
	}
	return TBlank, false
}

var (
	// helper map for marshalling the enum
	terrainEnumToString = map[Terrain]string{
		TBlank:                "",
		TAlps:                 "ALPS",
		TAridHills:            "AH",
		TAridTundra:           "AR",
		TBrushFlat:            "BF",
		TBrushHills:           "BH",
		TConiferHills:         "CH",
		TDeciduous:            "D",
		TDesert:               "DE",
		TDeciduousHills:       "DH",
		TGrassyHills:          "GH",
		TGrassyHillsPlateau:   "GHP",
		THighSnowyMountains:   "HSM",
		TJungle:               "JG",
		TJungleHills:          "JH",
		TLake:                 "L",
		TLowAridMountains:     "LAM",
		TLowConiferMountains:  "LCM",
		TLowJungleMountains:   "LJM",
		TLowSnowyMountains:    "LSM",
		TLowVolcanicMountains: "LVM",
		TOcean:                "O",
		TPolarIce:             "PI",
		TPrairie:              "PR",
		TPrairiePlateau:       "PPR",
		TRockyHills:           "RH",
		TSnowyHills:           "SH",
		TSwamp:                "SW",
		TTundra:               "TU",
	}
	// helper map for unmarshalling the enum
	terrainStringToEnum = map[string]Terrain{
		"":     TBlank,
		"ALPS": TAlps,
		"AH":   TAridHills,
		"AR":   TAridTundra,
		"BF":   TBrushFlat,
		"BH":   TBrushHills,
		"CH":   TConiferHills,
		"D":    TDeciduous,
		"DH":   TDeciduousHills,
		"DE":   TDesert,
		"GH":   TGrassyHills,
		"GHP":  TGrassyHillsPlateau,
		"HSM":  THighSnowyMountains,
		"JG":   TJungle,
		"JH":   TJungleHills,
		"L":    TLake,
		"LAM":  TLowAridMountains,
		"LCM":  TLowConiferMountains,
		"LJM":  TLowJungleMountains,
		"LSM":  TLowSnowyMountains,
		"LVM":  TLowVolcanicMountains,
		"O":    TOcean,
		"PI":   TPolarIce,
		"PPR":  TPrairiePlateau,
		"PR":   TPrairie,
		"RH":   TRockyHills,
		"SH":   TSnowyHills,
		"SW":   TSwamp,
		"TU":   TTundra,
	}
	// TileTerrainNames is the map for tile terrain name matching. the text values
	// are extracted from the Worldographer tileset. they must match exactly.
	// if you're adding to this list, the values are found by hovering over the
	// terrain in the GUI.
	TileTerrainNames = map[Terrain]string{
		TBlank:                "Blank",
		TAlps:                 "Mountains",
		TAridHills:            "Hills",
		TAridTundra:           "Flat Moss",
		TBrushFlat:            "Flat Shrubland",
		TBrushHills:           "Hills Shrubland",
		TConiferHills:         "Hills Forest Evergreen",
		TDeciduous:            "Flat Forest Deciduous Heavy",
		TDeciduousHills:       "Hills Deciduous Forest",
		TDesert:               "Flat Desert Sandy",
		TGrassyHills:          "Hills Grassland",
		TGrassyHillsPlateau:   "Hills Grassy",
		THighSnowyMountains:   "Mountain Snowcapped",
		TJungle:               "Flat Forest Jungle Heavy",
		TJungleHills:          "Hills Forest Jungle",
		TLake:                 "Water Shoals",
		TLowAridMountains:     "Mountains Dead Forest",
		TLowConiferMountains:  "Mountains Forest Evergreen",
		TLowJungleMountains:   "Mountain Forest Jungle",
		TLowSnowyMountains:    "Mountains Snowcapped",
		TLowVolcanicMountains: "Mountain Volcano Dormant",
		TOcean:                "Water Sea",
		TPolarIce:             "Mountains Glacier",
		TPrairie:              "Flat Grazing Land",
		TPrairiePlateau:       "Flat Grassland",
		TRockyHills:           "Underdark Broken Lands",
		TSnowyHills:           "Flat Snowfields",
		TSwamp:                "Flat Swamp",
		TTundra:               "Flat Steppe",
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
	UTFleet
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
		UTFleet:    "Fleet",
		UTGarrison: "Garrison",
	}
	// helper map for unmarshalling the enum
	unitTypeStringToEnum = map[string]UnitType{
		"Unknown":  UTUnknown,
		"Tribe":    UTTribe,
		"Courier":  UTCourier,
		"Element":  UTElement,
		"Fleet":    UTFleet,
		"Garrison": UTGarrison,
	}
)

type WindStrength_e int

const (
	WSUnknown WindStrength_e = iota
	WSCalm
	WSMild
	WSStrong
	WSGale
)

// MarshalJSON implements the json.Marshaler interface.
func (e WindStrength_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(windStrengthEnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *WindStrength_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = WindStrengthStringToEnum[s]; !ok {
		return fmt.Errorf("invalid WindStrength %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e WindStrength_e) String() string {
	if str, ok := windStrengthEnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("WindStrength(%d)", int(e))
}

var (
	// helper map for marshalling the enum
	windStrengthEnumToString = map[WindStrength_e]string{
		WSUnknown: "N/A",
		WSCalm:    "CALM",
		WSMild:    "MILD",
		WSStrong:  "STRONG",
		WSGale:    "GALE",
	}
	// WindStrengthStringToEnum is a helper map for unmarshalling the enum
	WindStrengthStringToEnum = map[string]WindStrength_e{
		"N/A":    WSUnknown,
		"CALM":   WSCalm,
		"MILD":   WSMild,
		"STRONG": WSStrong,
		"GALE":   WSGale,
	}
)

type UnitMovement_e int

const (
	UMUnknown UnitMovement_e = iota
	UMFleet
	UMFollows
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
		UMScouts:  "Scout",
		UMStill:   "Still",
		UMTribe:   "Tribe",
	}
	// helper map for unmarshalling the enum
	unitMoveStringToEnum = map[string]UnitMovement_e{
		"N/A":     UMUnknown,
		"Fleet":   UMFleet,
		"Follows": UMFollows,
		"Scout":   UMScouts,
		"Still":   UMStill,
		"Tribe":   UMTribe,
	}
)
