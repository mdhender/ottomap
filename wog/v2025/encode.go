package v2025

import (
	"fmt"
	"strings"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/hex"
	"github.com/mdhender/ottomap/wog/internal/wxxio"
)

// Encode renders m as a UTF-8 XML document targeting the Worldographer 2025
// schema family.
func Encode(m *ottomap.Map, opts Options) ([]byte, error) {
	orientation := "COLUMNS"
	if opts.Orientation == 1 {
		orientation = "ROWS"
	}
	projection := "FLAT"
	if opts.Projection == 1 {
		projection = "ICOSAHEDRAL"
	}

	// Build the terrain registry from terrains actually used.
	registry := wxxio.NewTerrainRegistry()
	for _, t := range m.Tiles() {
		registry.Intern(t.Terrain)
	}

	min, max, _ := m.Bounds()
	offMin, offMax, layout := wxxio.AxialRangeToOffset(min, max, orientation)
	cols := offMax.Col - offMin.Col + 1
	rows := offMax.Row - offMin.Row + 1
	if cols <= 0 {
		cols = 1
	}
	if rows <= 0 {
		rows = 1
	}

	var b strings.Builder
	writeHeader(&b, orientation, projection, opts.GMView)
	writeGridAndNumbering(&b)
	b.WriteString("<terrainmap>")
	registry.WriteTerrainMap(&b)
	b.WriteString("</terrainmap>\n")
	writeMapLayers(&b)

	// Worldographer emits cols outer <tilerow> elements, each containing
	// rows tile lines, for both orientations. tilesWide always reports the
	// column count and tilesHigh the row count.
	fmt.Fprintf(&b, `<tiles viewLevel="WORLD" tilesWide="%d" tilesHigh="%d">`+"\n", cols, rows)
	writeTileGrid(&b, m, registry, offMin, cols, rows, layout)
	b.WriteString("</tiles>\n")

	writeMapKey(&b)
	writeFeatures(&b, m, orientation)
	b.WriteString("<extraTerrain>\n</extraTerrain>\n")
	writeLabels(&b, m, orientation)
	b.WriteString("<shapes>\n</shapes>\n")
	writeNotes(&b, m, orientation)
	b.WriteString("<informations>\n</informations>\n")
	writeConfiguration(&b)
	b.WriteString("</map>\n")

	return []byte(b.String()), nil
}

func writeHeader(b *strings.Builder, orientation, projection string, gmView bool) {
	fmt.Fprintf(b,
		`<?xml version="1.0" encoding="utf-8"?>`+"\n"+
			`<map type="WORLD" release="2025" version="2.06" schema="1.10" lastViewLevel="WORLD" `+
			`continentFactor="0" kingdomFactor="0" provinceFactor="0" `+
			`worldToContinentHOffset="0.0" continentToKingdomHOffset="0.0" kingdomToProvinceHOffset="0.0" `+
			`worldToContinentVOffset="0.0" continentToKingdomVOffset="0.0" kingdomToProvinceVOffset="0.0" `+
			`hScrollbarPos="0.0" vScrollbarPos="0.0" `+
			`hexWidth="46.18" hexHeight="40.0" hexOrientation="%s" mapProjection="%s" `+
			`showNotes="false" showGMOnly="%t" showGMOnlyGlow="false" showFeatureLabels="true" `+
			`showGrid="true" showGridNumbers="true" showShadows="true" triangleSize="12">`+"\n",
		orientation, projection, gmView,
	)
}

func writeGridAndNumbering(b *strings.Builder) {
	b.WriteString(`<gridandnumbering color0="0x00000040" color1="0x00000040" color2="0x00000040" ` +
		`color3="0x00000040" color4="0x00000040" width0="1.0" width1="2.0" width2="3.0" width3="4.0" ` +
		`width4="1.0" gridOffsetContinentKingdomX="0.0" gridOffsetContinentKingdomY="0.0" ` +
		`gridOffsetWorldContinentX="0.0" gridOffsetWorldContinentY="0.0" gridOffsetWorldKingdomX="0.0" ` +
		`gridOffsetWorldKingdomY="0.0" gridSquare="0" gridSquareHeight="-1.0" gridSquareWidth="-1.0" ` +
		`gridOffsetX="0.0" gridOffsetY="0.0" numberFont="Arial" numberColor="0x000000ff" numberSize="20" ` +
		`numberStyle="PLAIN" numberFirstCol="0" numberFirstRow="0" numberOrder="COL_ROW" ` +
		`numberPosition="BOTTOM" numberPrePad="DOUBLE_ZERO" numberSeparator="."/>` + "\n")
}

func writeMapLayers(b *strings.Builder) {
	layers := []string{
		"Labels", "Grid", "Features", "Above Terrain", "Terrain Land",
		"Above Water", "Terrain Water", "Below All",
	}
	for _, l := range layers {
		fmt.Fprintf(b, `<maplayer name="%s" isVisible="true" opacity="1.0"/>`+"\n", l)
	}
}

func writeMapKey(b *strings.Builder) {
	b.WriteString(`<mapkey positionx="0.0" positiony="0.0" viewlevel="null" height="-1" ` +
		`backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" ` +
		`backgroundopacity="50" titleText="Map Key" titleFontFace="Arial" ` +
		`titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" ` +
		`scaleText="1 Hex = ? units" scaleFontFace="Arial" scaleFontColor="0.0,0.0,0.0,1.0" ` +
		`scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial" ` +
		`entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55">` + "\n" +
		`</mapkey>` + "\n")
}

func writeConfiguration(b *strings.Builder) {
	b.WriteString(`<configuration>` + "\n" +
		`<terrain-config></terrain-config>` + "\n" +
		`<feature-config></feature-config>` + "\n" +
		`<texture-config></texture-config>` + "\n" +
		`<text-config></text-config>` + "\n" +
		`<shape-config></shape-config>` + "\n" +
		`</configuration>` + "\n")
}

func writeTileGrid(
	b *strings.Builder,
	m *ottomap.Map,
	reg *wxxio.TerrainRegistry,
	offMin hex.OffsetCoord,
	cols, rows int,
	layout hex.Layout,
) {
	for col := 0; col < cols; col++ {
		b.WriteString("<tilerow>\n")
		for row := 0; row < rows; row++ {
			oc := hex.OffsetCoord{Col: offMin.Col + col, Row: offMin.Row + row}
			c := hex.FromOffset(oc, layout)
			t, _ := m.Tile(c)
			wxxio.FormatTileLine(b, t, reg.Slot(t.Terrain))
		}
		b.WriteString("</tilerow>\n")
	}
}

func writeFeatures(b *strings.Builder, m *ottomap.Map, orientation string) {
	b.WriteString("<features>\n")
	for _, f := range m.Features() {
		writeFeature(b, f, orientation)
	}
	b.WriteString("</features>\n")
}

func writeFeature(b *strings.Builder, f ottomap.Feature, orientation string) {
	px, py := wxxio.AxialToPixel(f.Location, f.Offset, orientation)
	scale := f.Scale
	if scale == 0 {
		scale = 1
	}
	fmt.Fprintf(b,
		`<feature type="%s" rotate="%g" uuid="%s" mapLayer="%s" `+
			`isFlipHorizontal="false" isFlipVertical="false" scale="%g" scaleHt="-1.0" `+
			`tags="%s" color="%s" ringcolor="null" isGMOnly="%t" isPlaceFreely="false" `+
			`labelPosition="180.0" labelDistance="-40" isWorld="true" isContinent="true" `+
			`isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false">`+"\n"+
			`<location viewLevel="WORLD" x="%g" y="%g"/>`+"\n",
		wxxio.Escape(f.Kind), f.Rotation, wxxio.Escape(f.ID), wxxio.Escape(string(f.Layer)),
		scale,
		wxxio.Escape(strings.Join(f.Tags, ",")),
		wxxio.FormatFloatRGBA(f.Color),
		f.GMOnly,
		px, py,
	)
	if f.Label != "" {
		fmt.Fprintf(b,
			`<label mapLayer="Labels" style="" fontFace="" color="0.0,0.0,0.0,1.0" `+
				`outlineColor="null" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" `+
				`isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`+"\n"+
				`<location viewLevel="WORLD" x="0.0" y="0.0" scale="12.5"/>%s</label>`+"\n",
			wxxio.XMLChars(f.Label),
		)
	}
	b.WriteString("</feature>\n")
}

func writeLabels(b *strings.Builder, m *ottomap.Map, orientation string) {
	b.WriteString("<labels>\n")
	for _, l := range m.Labels() {
		writeLabel(b, l, orientation)
	}
	b.WriteString("</labels>\n")
}

func writeLabel(b *strings.Builder, l ottomap.Label, orientation string) {
	px, py := wxxio.AxialToPixel(l.Location, l.Offset, orientation)
	font := l.Font.Family
	if font == "" {
		font = "Arial"
	}
	scale := l.Font.Size
	if scale == 0 {
		scale = 55
	}
	fmt.Fprintf(b,
		`<label mapLayer="%s" style="" fontFace="%s" color="%s" outlineColor="%s" `+
			`outlineSize="%g" rotate="%g" isBold="%t" isItalic="%t" `+
			`isWorld="%s" isContinent="%s" isKingdom="%s" isProvince="%s" isGMOnly="%t" tags="%s">`+"\n"+
			`<location viewLevel="WORLD" x="%g" y="%g" scale="%g"/>%s</label>`+"\n",
		wxxio.Escape(string(l.Layer)),
		wxxio.Escape(font),
		wxxio.FormatFloatRGBA(l.Color),
		wxxio.FormatFloatRGBA(l.Outline.Color),
		l.Outline.Size,
		l.Rotation,
		l.Font.Bold, l.Font.Italic,
		wxxio.ScopeFlagString(l.Scope, "world"),
		wxxio.ScopeFlagString(l.Scope, "continent"),
		wxxio.ScopeFlagString(l.Scope, "kingdom"),
		wxxio.ScopeFlagString(l.Scope, "province"),
		l.GMOnly,
		wxxio.Escape(strings.Join(l.Tags, ",")),
		px, py, scale,
		wxxio.XMLChars(l.Text),
	)
}

func writeNotes(b *strings.Builder, m *ottomap.Map, orientation string) {
	notes := m.Notes()
	if len(notes) == 0 {
		b.WriteString("<notes>\n</notes>\n")
		return
	}
	b.WriteString("<notes>\n")
	for _, n := range notes {
		px, py := wxxio.AxialToPixel(n.Location, ottomap.Offset{}, orientation)
		fmt.Fprintf(b,
			`<note key="%s" viewLevel="WORLD" x="%g" y="%g" filename="" parent="" `+
				`color="null" title="%s" isGMOnly="%t"><notetext>%s</notetext></note>`+"\n",
			wxxio.Escape(n.ID), px, py, wxxio.Escape(n.Title), n.GMOnly, wxxio.XMLChars(n.Body),
		)
	}
	b.WriteString("</notes>\n")
}
