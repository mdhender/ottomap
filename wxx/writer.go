// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/mdhender/ottomap/domain"
	"log"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

// Hex is a hex on the Tribenet map.
type Hex struct {
	Grid    string // AA ... ZZ
	Coords  Offset // coordinates in a grid hex are one-based
	Terrain domain.Terrain
}

// Tile is a hex on the Worldographer map.
type Tile struct {
	Terrain   domain.Terrain
	Elevation int
	IsIcy     bool
	IsGMOnly  bool
	Resources Resources
	Label     *Label
}

type Resources struct {
	Animal int
	Brick  int
	Crops  int
	Gems   int
	Lumber int
	Metals int
	Rock   int
}

// Offset captures the layout.
type Offset struct {
	Column int
	Row    int
}

func (w *WXX) Create(path string, hexes []Hex, showGridNumbering bool) error {
	// wmap is the consolidated Worldographer map.
	// It is indexed by column then row.
	var wmap [][]Tile

	// one grid on the Worldographer map is 30 columns wide by 21 rows high.
	columns, rows := 30, 21

	// allocate memory for the wmap and populate with blank tiles.
	wmap = make([][]Tile, columns)
	for column := 0; column < columns; column++ {
		wmap[column] = make([]Tile, rows)
	}

	// convert the grid hexes to tiles
	for _, hex := range hexes {
		tile := &wmap[hex.Coords.Column-1][hex.Coords.Row-1]
		tile.Terrain = hex.Terrain
		switch tile.Terrain {
		case domain.TLake:
			tile.Elevation = -1
		case domain.TOcean:
			tile.Elevation = -3
		case domain.TPrairie:
			tile.Elevation = 1_000
		default:
			tile.Elevation = 1
		}
		tile.Label = &Label{Text: fmt.Sprintf("%s %02d%02d", hex.Grid, hex.Coords.Column, hex.Coords.Row)}
	}

	w.buffer = &bytes.Buffer{}

	// create the slice that maps our terrains to the Worldographer terrain names.
	// todo: this is a hack and should be extracted into the domain package.
	var terrains []string
	var maxTerrain domain.Terrain
	for k := range domain.TileTerrainNames {
		if k > maxTerrain {
			maxTerrain = k
		}
	}
	terrains = make([]string, maxTerrain+1)
	for k, v := range domain.TileTerrainNames {
		terrains[k] = v
	}
	// the first row must be the Blank terrain
	if terrains[0] != "Blank" {
		panic(`assert(terrains[0] == "Blank")`)
	}
	// all rows must have a value
	for k, v := range terrains {
		if v == "" {
			log.Fatalf("map: terrain %q has no associated terrain type!\n", domain.Terrain(k).String())
		}
	}

	// start writing the XML

	w.Println(`<?xml version='1.0' encoding='utf-16'?>`)

	// hexWidth and hexHeight are used to control the initial "zoom" on the map.
	const hexWidth, hexHeight = 46.18, 40.0 // standard, unzoomed map scale

	w.Println(`<map type="WORLD" version="1.74" lastViewLevel="WORLD" continentFactor="0" kingdomFactor="0" provinceFactor="0" worldToContinentHOffset="0.0" continentToKingdomHOffset="0.0" kingdomToProvinceHOffset="0.0" worldToContinentVOffset="0.0" continentToKingdomVOffset="0.0" kingdomToProvinceVOffset="0.0" `)
	w.Println(`hexWidth="%g" hexHeight="%g" hexOrientation="COLUMNS" mapProjection="FLAT" showNotes="true" showGMOnly="true" showGMOnlyGlow="false" showFeatureLabels="true" showGrid="true" showGridNumbers="false" showShadows="true"  triangleSize="12">`, hexWidth, hexHeight)

	w.Println(`<gridandnumbering color0="0x00000040" color1="0x00000040" color2="0x00000040" color3="0x00000040" color4="0x00000040" width0="1.0" width1="2.0" width2="3.0" width3="4.0" width4="1.0" gridOffsetContinentKingdomX="0.0" gridOffsetContinentKingdomY="0.0" gridOffsetWorldContinentX="0.0" gridOffsetWorldContinentY="0.0" gridOffsetWorldKingdomX="0.0" gridOffsetWorldKingdomY="0.0" gridSquare="0" gridSquareHeight="-1.0" gridSquareWidth="-1.0" gridOffsetX="0.0" gridOffsetY="0.0" numberFont="Arial" numberColor="0x000000ff" numberSize="20" numberStyle="PLAIN" numberFirstCol="0" numberFirstRow="0" numberOrder="COL_ROW" numberPosition="BOTTOM" numberPrePad="DOUBLE_ZERO" numberSeparator="." />`)

	w.Printf("<terrainmap>")
	for n, terrain := range terrains {
		if n == 0 {
			w.Printf("%s\t%d", terrain, n)
		} else {
			w.Printf("\t%s\t%d", terrain, n)
		}
	}
	w.Printf("</terrainmap>\n")

	w.Println(`<maplayer name="Labels" isVisible="true"/>`)
	w.Println(`<maplayer name="Grid" isVisible="true"/>`)
	w.Println(`<maplayer name="Features" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Terrain" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Land" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Below All" isVisible="true"/>`)

	// width is the number of columns, height is the number of rows.
	w.Println(`<tiles viewLevel="WORLD" tilesWide="%d" tilesHigh="%d">`, columns, rows)

	// NB: the element is named "tilerow" but it holds the tiles for a single column.
	for col := 0; col < columns; col++ {
		w.Printf("<tilerow>\n")
		for row := 0; row < rows; row++ {
			tr := wmap[col][row]
			w.Printf("%d\t%d", int(tr.Terrain), tr.Elevation)
			if tr.IsIcy {
				w.Printf("\t1")
			} else {
				w.Printf("\t0")
			}
			if tr.IsGMOnly {
				w.Printf("\t1")
			} else {
				w.Printf("\t0")
			}
			// todo: implement resources. for now, just set them to 0 Z.
			w.Printf("\t%d\t%s\n", tr.Resources.Animal, "Z")
		}
		w.Printf("</tilerow>\n")
	}

	w.Println(`</tiles>`)

	w.Println(`<mapkey positionx="0.0" positiony="0.0" viewlevel="WORLD" height="-1" backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" backgroundopacity="50" titleText="Map Key" titleFontFace="Arial"  titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" scaleText="1 Hex = ? units" scaleFontFace="Arial"  scaleFontColor="0.0,0.0,0.0,1.0" scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial"  entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55"  >`)
	w.Println(`</mapkey>`)

	w.Println(`<features>`)

	w.Println(`<feature type="Building Cottage" rotate="0.0" uuid="2be6a8ee-4fad-421c-8bee-80335405e11d" mapLayer="Features" isFlipHorizontal="false" isFlipVertical="false" scale="-1.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="150.0" y="150.0" /><label  mapLayer="Features" style="City" fontFace=".AppleSystemUIFont" color="0.0,0.0,0.0,1.0" outlineColor="0.0,0.0,0.0,1.0" outlineSize="2.0" rotate="0.0" isBold="false" isItalic="false" isWorld="false" isContinent="false" isKingdom="false" isProvince="false" isGMOnly="false" tags=""><location viewLevel="WORLD" x="150.0" y="150.0" scale="25.0" /></label>`)
	w.Println(`</feature>`)

	w.Println(`<feature type="Settlement City" rotate="0.0" uuid="fcb48970-c74d-4984-a920-6819d56a9e25" mapLayer="Features" isFlipHorizontal="false" isFlipVertical="false" scale="-1.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="2400.0" y="2250.0" /><label  mapLayer="Features" style="City" fontFace=".AppleSystemUIFont" color="0.0,0.0,0.0,1.0" outlineColor="0.0,0.0,0.0,1.0" outlineSize="2.0" rotate="0.0" isBold="false" isItalic="false" isWorld="false" isContinent="false" isKingdom="false" isProvince="false" isGMOnly="false" tags=""><location viewLevel="WORLD" x="2400.0" y="2250.0" scale="25.0" /></label>`)
	w.Println(`</feature>`)

	w.Println(`</features>`)

	w.Printf("<labels>\n")

	if showGridNumbering {
		for col := 0; col < columns; col++ {
			for row := 0; row < rows; row++ {
				w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
				p := crs_to_pixel(col, row)
				w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.75" />`, p.X, p.Y)
				w.Printf("%2d,%2d", col, row)
				w.Printf("</label>\n")
			}
		}
	}

	// add labels to tiles when needed.
	for col := 0; col < columns; col++ {
		for row := 0; row < rows; row++ {
			tile := wmap[col][row]
			if tile.Label == nil {
				continue
			}
			w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
			p := crs_to_pixel(col, row)
			p.X += 25  // nudge labels to the center of the tile.
			p.Y += 125 // nudge labels down to the bottom of the tile.
			w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, p.X, p.Y)
			w.Printf("%s", tile.Label.Text)
			w.Printf("</label>\n")
		}
	}

	w.Printf("</labels>\n")

	w.Println(`<shapes>`)
	w.Println(`</shapes>`)
	w.Println(`<notes>`)
	w.Println(`</notes>`)
	w.Println(`<informations>`)
	w.Println(`</informations>`)
	w.Println(`<configuration>`)
	w.Println(`  <terrain-config>`)
	w.Println(`  </terrain-config>`)
	w.Println(`  <feature-config>`)
	w.Println(`  </feature-config>`)
	w.Println(`  <texture-config>`)
	w.Println(`  </texture-config>`)
	w.Println(`  <text-config>`)
	w.Println(`  </text-config>`)
	w.Println(`  <shape-config>`)
	w.Println(`  </shape-config>`)
	w.Println(`  </configuration>`)
	w.Println(`</map>`)
	w.Println(``)

	//fmt.Printf("%s\n", w.buffer.String())

	// convert the source from UTF-8 to UTF-16
	var buf16 bytes.Buffer
	buf16.Write([]byte{0xfe, 0xff}) // write the BOM
	for src := w.buffer.Bytes(); len(src) > 0; {
		// extract next rune from the source
		r, w := utf8.DecodeRune(src)
		if r == utf8.RuneError {
			return fmt.Errorf("invalid utf8 data")
		}
		// consume that rune
		src = src[w:]
		// convert the rune to UTF-16 and write it to the results
		for _, v := range utf16.Encode([]rune{r}) {
			if err := binary.Write(&buf16, binary.BigEndian, v); err != nil {
				return err
			}
		}
	}
	w.buffer = nil

	// convert the UTF-16 to a gzip stream
	var bufGZ bytes.Buffer
	gz := gzip.NewWriter(&bufGZ)
	if _, err := gz.Write(buf16.Bytes()); err != nil {
		return err
	} else if err = gz.Close(); err != nil {
		return err
	}

	// write the compressed data to the output file
	if err := os.WriteFile(path, bufGZ.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// ok. the world map doesn't draw perfect hexagons. they're flattened slightly.
// It's easier to call them "tiles" since they aren't regular hexagons. Anyway, I
// estimated the size of the tiles by looking at the output for labels in different
// scenarios.
//
// I came up a tile size of about 300 pixels for height and 225 pixels for width.
// The displayed tiles are also offset by 130 pixels for the left margin and 165
// pixels for the top margin.
//
// We will need to revisit the sizing when creating larger maps because the center
// point of the tiles drifts.

// crs_to_pixel converts a column, row to the pixel at the center of the corresponding tile.
func crs_to_pixel(column, row int) Point {
	const height, width = 300, 225
	const leftMargin, topMargin = 130, 165

	x, y := float64(column)*width, float64(row)*height
	if column%2 == 1 { // shove odd rows down half the height of a tile
		y += height / 2
	}

	// offset final point by the margins
	return Point{X: x + leftMargin, Y: y + topMargin}
}

// NB: most of the code below is derived from https://www.redblobgames.com/grids/hexagons/.
// It turns out that it isn't used because Worldographer doesn't output regular hexagons.

// There are four types of layouts for Offset coordinates.
//
// 1. EvenQ is a vertical   layout where columns with an even Q value are shoved "down."
// 2. OddQ  is a vertical   layout where columns with an odd  Q value are shoved "down."
// 3. EvenR is a horizontal layout where rows    with an even R value are shoved "right."
// 4. OddR  is a horizontal layout where rows    with an odd  R value are shoved "right."
//
// A vertical layout contains flat-top hexes. A horizontal layout contains pointy-top hexes.

type Layout int

const (
	EvenQ Layout = iota
	EvenR
	OddQ
	OddR
)

type Label struct {
	Text string
}

// Cube are the coordinates of a hex in a cube.
// They have the constraint Q + R + S = 0.
type Cube struct {
	Q int // q is the north-south axis
	R int // r is the northwest-southeast axis
	S int // s is the northeast-southwest axis
}

// The "size" of a hex is the distance from the center of the hex to a vertex.
// The "apothem" is the distance from the center of the hex to the midpoint of a side.
// The apothem is sqrt(3) times the size divided by 2.

func cube_to_evenq(c Cube) Offset {
	return Offset{
		Column: c.Q,
		Row:    c.R + (c.Q+(c.Q&1))/2,
	}
}

func cube_to_oddq(c Cube) Offset {
	return Offset{
		Column: c.Q,
		Row:    c.R + (c.Q-(c.Q&1))/2,
	}
}

func oddq_to_cube(o Offset) Cube {
	q := o.Column
	r := o.Row - (o.Column-(o.Column&1))/2
	return Cube{Q: q, R: r, S: -q - r}
}

func evenq_to_cube(o Offset) Cube {
	q := o.Column
	r := o.Row - (o.Column+(o.Column&1))/2
	return Cube{Q: q, R: r, S: -q - r}
}

type Axial struct {
	Q float64
	R float64
}

func cube_to_axial(c Cube) Axial {
	return Axial{Q: float64(c.Q), R: float64(c.R)}
}

func evenq_to_axial(o Offset) Axial {
	return Axial{
		Q: float64(o.Column),
		R: float64(o.Row - (o.Column+(o.Column&1))/2),
	}
}

func oddq_to_axial(o Offset) Axial {
	return Axial{
		Q: float64(o.Column),
		R: float64(o.Row - (o.Column-(o.Column&1))/2),
	}
}

type Point struct {
	X float64
	Y float64
}

func (p Point) Scale(s float64) Point {
	return Point{
		X: p.X * s,
		Y: p.Y * s,
	}
}

const (
	sqrt3 = 1.73205080757 // math.Sqrt(3)
)

func flat_hex_to_pixel(a Axial, size float64) Point {
	apothem := size * sqrt3 / 2
	p := Point{
		X: size * (3.0 * a.Q / 2.0),
		Y: size * (sqrt3*a.Q/2.0 + sqrt3*a.R),
	}
	// bump down and over
	p.X, p.Y = p.X+size, p.Y+apothem
	return p
}

func pointy_hex_to_pixel(a Axial) Point {
	return Point{
		X: sqrt3*a.Q + sqrt3*a.R/2,
		Y: 3 * a.R / 2,
	}
}
