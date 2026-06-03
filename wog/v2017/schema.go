package v2017

import "encoding/xml"

// schemaMap is the on-disk XML root for the Worldographer 2017 (Hexographer)
// schema. It differs from v2025 in three ways: there is no release
// attribute on <map>, <maplayer> lacks the opacity attribute, and there is
// no <blurTerrainBG> or <extraTerrain> element.
type schemaMap struct {
	XMLName xml.Name `xml:"map"`

	Type                      string  `xml:"type,attr"`
	Version                   string  `xml:"version,attr"`
	LastViewLevel             string  `xml:"lastViewLevel,attr"`
	ContinentFactor           int     `xml:"continentFactor,attr"`
	KingdomFactor             int     `xml:"kingdomFactor,attr"`
	ProvinceFactor            int     `xml:"provinceFactor,attr"`
	WorldToContinentHOffset   float64 `xml:"worldToContinentHOffset,attr"`
	ContinentToKingdomHOffset float64 `xml:"continentToKingdomHOffset,attr"`
	KingdomToProvinceHOffset  float64 `xml:"kingdomToProvinceHOffset,attr"`
	WorldToContinentVOffset   float64 `xml:"worldToContinentVOffset,attr"`
	ContinentToKingdomVOffset float64 `xml:"continentToKingdomVOffset,attr"`
	KingdomToProvinceVOffset  float64 `xml:"kingdomToProvinceVOffset,attr"`
	HexWidth                  float64 `xml:"hexWidth,attr"`
	HexHeight                 float64 `xml:"hexHeight,attr"`
	HexOrientation            string  `xml:"hexOrientation,attr"`
	MapProjection             string  `xml:"mapProjection,attr"`
	ShowNotes                 bool    `xml:"showNotes,attr"`
	ShowGMOnly                bool    `xml:"showGMOnly,attr"`
	ShowGMOnlyGlow            bool    `xml:"showGMOnlyGlow,attr"`
	ShowFeatureLabels         bool    `xml:"showFeatureLabels,attr"`
	ShowGrid                  bool    `xml:"showGrid,attr"`
	ShowGridNumbers           bool    `xml:"showGridNumbers,attr"`
	ShowShadows               bool    `xml:"showShadows,attr"`
	TriangleSize              int     `xml:"triangleSize,attr"`

	TerrainMap schemaTerrainMap `xml:"terrainmap"`
	MapLayers  []schemaMapLayer `xml:"maplayer"`
	Tiles      schemaTiles      `xml:"tiles"`
	Features   schemaFeatures   `xml:"features"`
	Labels     schemaLabels     `xml:"labels"`
	Notes      schemaNotes      `xml:"notes"`
}

type schemaTerrainMap struct {
	InnerText string `xml:",chardata"`
}

type schemaMapLayer struct {
	Name      string `xml:"name,attr"`
	IsVisible bool   `xml:"isVisible,attr"`
}

type schemaTiles struct {
	ViewLevel string          `xml:"viewLevel,attr"`
	TilesWide int             `xml:"tilesWide,attr"`
	TilesHigh int             `xml:"tilesHigh,attr"`
	TileRows  []schemaTileRow `xml:"tilerow"`
}

type schemaTileRow struct {
	InnerText string `xml:",chardata"`
}

type schemaFeatures struct {
	Features []schemaFeature `xml:"feature"`
}

type schemaFeature struct {
	Type              string  `xml:"type,attr"`
	Rotate            float64 `xml:"rotate,attr"`
	UUID              string  `xml:"uuid,attr"`
	MapLayer          string  `xml:"mapLayer,attr"`
	IsFlipHorizontal  bool    `xml:"isFlipHorizontal,attr"`
	IsFlipVertical    bool    `xml:"isFlipVertical,attr"`
	Scale             float64 `xml:"scale,attr"`
	ScaleHt           float64 `xml:"scaleHt,attr"`
	Tags              string  `xml:"tags,attr"`
	Color             string  `xml:"color,attr"`
	RingColor         string  `xml:"ringcolor,attr"`
	IsGMOnly          bool    `xml:"isGMOnly,attr"`
	IsPlaceFreely     bool    `xml:"isPlaceFreely,attr"`
	LabelPosition     string  `xml:"labelPosition,attr"`
	LabelDistance     int     `xml:"labelDistance,attr"`
	IsWorld           bool    `xml:"isWorld,attr"`
	IsContinent       bool    `xml:"isContinent,attr"`
	IsKingdom         bool    `xml:"isKingdom,attr"`
	IsProvince        bool    `xml:"isProvince,attr"`
	IsFillHexBottom   bool    `xml:"isFillHexBottom,attr"`
	IsHideTerrainIcon bool    `xml:"isHideTerrainIcon,attr"`

	Location schemaFeatureLocation `xml:"location"`
	Label    *schemaLabel          `xml:"label,omitempty"`
}

type schemaFeatureLocation struct {
	ViewLevel string  `xml:"viewLevel,attr"`
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
}

type schemaLabels struct {
	Labels []schemaLabel `xml:"label"`
}

type schemaLabel struct {
	MapLayer        string  `xml:"mapLayer,attr"`
	Style           string  `xml:"style,attr"`
	FontFace        string  `xml:"fontFace,attr"`
	Color           string  `xml:"color,attr"`
	OutlineColor    string  `xml:"outlineColor,attr"`
	OutlineSize     float64 `xml:"outlineSize,attr"`
	Rotate          float64 `xml:"rotate,attr"`
	IsBold          bool    `xml:"isBold,attr"`
	IsItalic        bool    `xml:"isItalic,attr"`
	IsWorld         bool    `xml:"isWorld,attr"`
	IsContinent     bool    `xml:"isContinent,attr"`
	IsKingdom       bool    `xml:"isKingdom,attr"`
	IsProvince      bool    `xml:"isProvince,attr"`
	IsGMOnly        bool    `xml:"isGMOnly,attr"`
	Tags            string  `xml:"tags,attr"`
	BackgroundColor string  `xml:"backgroundColor,attr"`

	Location  schemaLabelLocation `xml:"location"`
	InnerText string              `xml:",chardata"`
}

type schemaLabelLocation struct {
	ViewLevel string  `xml:"viewLevel,attr"`
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
	Scale     float64 `xml:"scale,attr"`
}

type schemaNotes struct {
	Notes []schemaNote `xml:"note"`
}

type schemaNote struct {
	Key       string  `xml:"key,attr"`
	ViewLevel string  `xml:"viewLevel,attr"`
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
	Filename  string  `xml:"filename,attr"`
	Parent    string  `xml:"parent,attr"`
	Color     string  `xml:"color,attr"`
	Title     string  `xml:"title,attr"`
	IsGMOnly  bool    `xml:"isGMOnly,attr"`
	NoteText  string  `xml:"notetext"`
}
