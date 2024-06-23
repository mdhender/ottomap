// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/directions"
	"github.com/mdhender/ottomap/domain"
	"log"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

// Hex is a hex on the Tribenet map.
type Hex struct {
	GridId     string // AA ... ZZ
	GridCoords string // original grid coordinates
	Offset     Offset // coordinates in a grid hex are one-based
	Terrain    domain.Terrain
	Scouted    bool
	Features   Features
}

func (h *Hex) Grid() string {
	return h.GridCoords[:2]
	//id, column, row := h.Grid, h.Coords.Column, h.Coords.Row
	//return fmt.Sprintf("%s %02d%02d", id, column, row)
}

// Tile is a hex on the Worldographer map.
type Tile struct {
	created    string // turn id when the tile was created
	updated    string // turn id when the tile was updated
	GridCoords string // original grid coordinates
	Terrain    domain.Terrain
	Elevation  int
	IsIcy      bool
	IsGMOnly   bool
	Resources  Resources
	Features   Features
}

// Features are things to display on the map
type Features struct {
	Edges struct {
		Ford      []directions.Direction
		Pass      []directions.Direction
		River     []directions.Direction
		StoneRoad []directions.Direction
	}

	// set label for either Coords or Numbers, not both
	CoordsLabel  string
	NumbersLabel string

	IsOrigin   bool // true for the clan's origin hex
	Label      *Label
	Resources  domain.Resource
	Settlement *Settlement // name of settlement

	Created string // turn id when the hex was created
	Updated string // turn id when the hex was updated
	Visited string // turn id when the hex was last visited
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

func (w *WXX) AddGridCoords() {
	w.addGridCoords()
}

func (w *WXX) AddGridNumbering() {
	w.addGridNumbers()
}

func (w *WXX) Create(path string, showGridCenters bool) error {
	if w.totalGrids == 0 {
		return fmt.Errorf("wxx: create: no grids")
	}

	// handy way to figure out offset for features and labels
	//origin := coordsToPoints(0, 0)
	//log.Printf("origin (%f, %f)\n", origin[0].X, origin[0].Y)
	//x, y := 148.14830212120597, 241.81408953094206
	//log.Printf("delta (%f, %f)\n", x-origin[0].X, y-origin[0].Y)

	// create any missing grids
	for gridRow := w.minGridRow; gridRow <= w.maxGridRow; gridRow++ {
		for gridColumn := w.minGridColumn; gridColumn <= w.maxGridColumn; gridColumn++ {
			if w.grids[gridRow][gridColumn] != nil {
				continue
			}
			// create a new grid
			gridId := gridRowColumnToId(gridRow, gridColumn)
			w.newGrid(gridId)
		}
	}

	// calculate the size of the consolidated map
	gridsWide, gridsHigh := w.maxGridColumn-w.minGridColumn+1, w.maxGridRow-w.minGridRow+1
	tilesWide, tilesHigh := columnsPerGrid*gridsWide, rowsPerGrid*gridsHigh
	//log.Printf("map: grid columns %4d rows %4d", gridsWide, gridsHigh)
	//log.Printf("map: tile columns %4d rows %4d", tilesWide, tilesHigh)

	// create the slice that maps our terrains to the Worldographer terrain names.
	// todo: this is a hack and should be extracted into the domain package.
	var terrainSlice []string // the first row must be the Blank terrain
	for n := 0; n < domain.NumberOfTerrainTypes; n++ {
		value, ok := domain.TileTerrainNames[domain.Terrain(n)]
		// all rows must have a value
		if !ok {
			panic(fmt.Sprintf(`assert(terrains[%d] != false)`, n))
		} else if value == "" {
			panic(fmt.Sprintf(`assert(terrains[%d] != "")`, n))
		}
		terrainSlice = append(terrainSlice, value)
	}
	//log.Printf("terrains: %d: %v\n", len(terrainSlice), terrainSlice)

	// start writing the XML
	w.buffer = &bytes.Buffer{}

	w.Println(`<?xml version='1.0' encoding='utf-16'?>`)

	// hexWidth and hexHeight are used to control the initial "zoom" on the map.
	const hexWidth, hexHeight = 46.18, 40.0

	w.Println(`<map type="WORLD" version="1.74" lastViewLevel="WORLD" continentFactor="0" kingdomFactor="0" provinceFactor="0" worldToContinentHOffset="0.0" continentToKingdomHOffset="0.0" kingdomToProvinceHOffset="0.0" worldToContinentVOffset="0.0" continentToKingdomVOffset="0.0" kingdomToProvinceVOffset="0.0" `)
	w.Println(`hexWidth="%g" hexHeight="%g" hexOrientation="COLUMNS" mapProjection="FLAT" showNotes="true" showGMOnly="true" showGMOnlyGlow="false" showFeatureLabels="true" showGrid="true" showGridNumbers="false" showShadows="true"  triangleSize="12">`, hexWidth, hexHeight)

	w.Println(`<gridandnumbering color0="0x00000040" color1="0x00000040" color2="0x00000040" color3="0x00000040" color4="0x00000040" width0="1.0" width1="2.0" width2="3.0" width3="4.0" width4="1.0" gridOffsetContinentKingdomX="0.0" gridOffsetContinentKingdomY="0.0" gridOffsetWorldContinentX="0.0" gridOffsetWorldContinentY="0.0" gridOffsetWorldKingdomX="0.0" gridOffsetWorldKingdomY="0.0" gridSquare="0" gridSquareHeight="-1.0" gridSquareWidth="-1.0" gridOffsetX="0.0" gridOffsetY="0.0" numberFont="Arial" numberColor="0x000000ff" numberSize="20" numberStyle="PLAIN" numberFirstCol="0" numberFirstRow="0" numberOrder="COL_ROW" numberPosition="BOTTOM" numberPrePad="DOUBLE_ZERO" numberSeparator="." />`)

	w.Printf("<terrainmap>")
	for n, terrain := range terrainSlice {
		if n == 0 {
			w.Printf("%s\t%d", terrain, n)
		} else {
			w.Printf("\t%s\t%d", terrain, n)
		}
	}
	w.Printf("</terrainmap>\n")

	w.Println(`<maplayer name="Tribenet Coords" isVisible="true"/>`)
	w.Println(`<maplayer name="Tribenet Origin" isVisible="false"/>`)
	w.Println(`<maplayer name="Tribenet Resources" isVisible="true"/>`)
	w.Println(`<maplayer name="Tribenet Settlements" isVisible="true"/>`)
	w.Println(`<maplayer name="Tribenet Unvisited" isVisible="true"/>`)
	w.Println(`<maplayer name="Labels" isVisible="true"/>`)
	w.Println(`<maplayer name="Grid" isVisible="true"/>`)
	w.Println(`<maplayer name="Features" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Terrain" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Land" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Below All" isVisible="true"/>`)

	// width is the number of columns, height is the number of rows.
	w.Println(`<tiles viewLevel="WORLD" tilesWide="%d" tilesHigh="%d">`, tilesWide, tilesHigh)

	// NB: the element is named "tilerow" but it holds the tiles for a single column because we're using COLUMNS orientation.
	//
	// this is hard for me to understand.
	//
	// we have to sweep through the columns from left to right.
	// for each column, we have to sweep through the rows from top to bottom.
	// the rows have to jump across grids as we process each tile row.
	for gridColumn := w.minGridColumn; gridColumn <= w.maxGridColumn; gridColumn++ {
		// we are creating tile-row elements, but they actually contain all the tiles for a single column.
		// the name is confusing because we're using COLUMNS orientation.
		for tileColumn := 0; tileColumn < columnsPerGrid; tileColumn++ {
			// the tile-row element holds the tiles for a single column.
			// these rows are going to come from multiple grids.
			w.Printf("<tilerow>\n")
			// sweep through all the grids that contain this column, from top to bottom.
			for gridRow := w.minGridRow; gridRow <= w.maxGridRow; gridRow++ {
				g := w.grids[gridRow][gridColumn]
				if g == nil {
					// todo: this could be updated to punch out blank tiles instead of panicking.
					//log.Printf("bug: minGridColumn %d: gridColumn %d: maxGridColumn %d\n", w.minGridColumn, gridColumn, w.maxGridColumn)
					//log.Printf("bug: tileColumn %d\n", tileColumn)
					//log.Printf("bug: minGridRow %d: gridRow %d: maxGridRow %d\n", w.minGridRow, gridRow, w.maxGridRow)
					gridId := gridRowColumnToId(gridRow, gridColumn)
					log.Printf("bug: gridId %q is missing\n", gridId)
					g = w.newGrid(gridId)
					//panic("assert(g != nil)\nplease report this bug")
					w.grids[gridRow][gridColumn] = g
				}
				for tileRow := 0; tileRow < rowsPerGrid; tileRow++ {
					// process all the tiles in this row of this grid.
					tile := g.tiles[tileColumn][tileRow]
					w.Printf("%d\t%d", int(tile.Terrain), tile.Elevation)
					if tile.IsIcy {
						w.Printf("\t1")
					} else {
						w.Printf("\t0")
					}
					if tile.IsGMOnly {
						w.Printf("\t1")
					} else {
						w.Printf("\t0")
					}
					// todo: implement resources. for now, just set them to 0 Z.
					w.Printf("\t%d\t%s\n", tile.Resources.Animal, "Z")
				}
			}
			w.Printf("</tilerow>\n")
		}
	}

	w.Println(`</tiles>`)

	w.Println(`<mapkey positionx="0.0" positiony="0.0" viewlevel="WORLD" height="-1" backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" backgroundopacity="50" titleText="Map Key" titleFontFace="Arial"  titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" scaleText="1 Hex = ? units" scaleFontFace="Arial"  scaleFontColor="0.0,0.0,0.0,1.0" scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial"  entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55"  >`)
	w.Println(`</mapkey>`)

	// add features
	w.Println(`<features>`)

	for gridRow := w.minGridRow; gridRow <= w.maxGridRow; gridRow++ {
		gridRowOffset := (gridRow - w.minGridRow) * rowsPerGrid
		for gridColumn := w.minGridColumn; gridColumn <= w.maxGridColumn; gridColumn++ {
			g := w.grids[gridRow][gridColumn]
			if g == nil {
				continue
			}
			gridColumnOffset := (gridColumn - w.minGridColumn) * columnsPerGrid
			for column := 0; column < columnsPerGrid; column++ {
				for row := 0; row < rowsPerGrid; row++ {
					tile := g.tiles[column][row]
					points := coordsToPoints(gridColumnOffset+column, gridRowOffset+row)

					if tile.Features.IsOrigin {
						origin := points[0]
						w.Printf(`<feature type="Three Dots" rotate="0.0" uuid="%s" mapLayer="Tribenet Origin" isFlipHorizontal="false" isFlipVertical="false" scale="-1.0" scaleHt="-1.0" tags="" color="0.800000011920929,0.800000011920929,0.800000011920929,1.0" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false">`, uuid.New().String())
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" />`, origin.X, origin.Y)
						w.Printf(`<label  mapLayer="Tribenet Origin" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" scale="25.0" />`, origin.X, origin.Y)
						w.Printf(`</label>`)
						w.Printf("</feature>\n")
					}

					if tile.Terrain == domain.TPrairiePlateau {
						origin := points[0]
						w.Printf(`<feature type="Semi-Real Hill Jagged" rotate="0.0" uuid="%s" mapLayer="Features" isFlipHorizontal="false" isFlipVertical="false" scale="90.0" scaleHt="-1.0" tags="" color="0.800000011920929,0.800000011920929,0.800000011920929,1.0" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false">`, uuid.New().String())
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" />`, origin.X, origin.Y)
						w.Printf(`<label  mapLayer="Features" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" scale="25.0" />`, origin.X, origin.Y)
						w.Printf(`</label>`)
						w.Printf("</feature>\n")
					}

					if tile.Features.Resources != domain.RNone {
						origin := points[0]
						w.Printf(`<feature type="Resource Mines" rotate="0.0" uuid="%s" mapLayer="Tribenet Resources" isFlipHorizontal="false" isFlipVertical="false" scale="35.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false">`, uuid.New().String())
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" />`, origin.X, origin.Y)
						w.Printf(`<label  mapLayer="Tribenet Resources" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, origin.X, origin.Y)
						w.Printf("%s", tile.Features.Resources)
						w.Printf(`</label>`)
						w.Println(`</feature>`)
					}

					if tile.Features.Settlement != nil {
						settlement := points[0]
						w.Printf(`<feature type="Settlement City" rotate="0.0" uuid="%s" mapLayer="Tribenet Settlements" isFlipHorizontal="false" isFlipVertical="false" scale="35.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="%f" y="%f" />`, uuid.New().String(), settlement.X, settlement.Y)
						w.Println(`</feature>`)
					}
				}
			}
		}
	}

	w.Println(`</features>`)

	w.Printf("<labels>\n")

	for gridRow := w.minGridRow; gridRow <= w.maxGridRow; gridRow++ {
		gridRowOffset := (gridRow - w.minGridRow) * rowsPerGrid
		for gridColumn := w.minGridColumn; gridColumn <= w.maxGridColumn; gridColumn++ {
			g := w.grids[gridRow][gridColumn]
			if g == nil {
				continue
			}
			gridColumnOffset := (gridColumn - w.minGridColumn) * columnsPerGrid
			for column := 0; column < columnsPerGrid; column++ {
				for row := 0; row < rowsPerGrid; row++ {
					tile := g.tiles[column][row]
					points := coordsToPoints(gridColumnOffset+column, gridRowOffset+row)

					if showGridCenters {
						w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, points[0].X, points[0].Y)
						if column&1 == 0 {
							w.Printf("%d", column&1)
						} else {
							w.Printf("%d", column&1)
						}
						w.Printf("</label>\n")
					}

					if tile.Features.Created != "" && tile.Features.Visited == "" {
						labelXY := points[0].Translate(Point{-1.851698, 91.814090})
						w.Printf(`<label  mapLayer="Tribenet Unvisited" style="null" fontFace="null" color="0.7019608020782471,0.7019608020782471,0.7019608020782471,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%f" y="%f" scale="90.0" />`, labelXY.X, labelXY.Y)
						w.Printf("X")
						w.Printf("</label>/n")
					}

					if tile.Features.CoordsLabel != "" {
						labelXY := bottomLeftCenter(points).Translate(Point{-9, -2.5})
						w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, labelXY.X, labelXY.Y)
						w.Printf("%s", tile.Features.CoordsLabel)
						w.Printf("</label>\n")
					} else if tile.Features.NumbersLabel != "" {
						labelXY := bottomLeftCenter(points).Translate(Point{-15, -2.5})
						w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, labelXY.X, labelXY.Y)
						w.Printf("%s", tile.Features.NumbersLabel)
						w.Printf("</label>\n")
					}

					if tile.Features.Label != nil {
						labelXY := points[0]
						w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, labelXY.X, labelXY.Y)
						w.Printf("%s", tile.Features.Label.Text)
						w.Printf("</label>\n")
					}

					if tile.Features.Settlement != nil {
						label := tile.Features.Settlement.Name
						labelXY := settlementLabelXY(label, points)
						w.Printf(`<label  mapLayer="Tribenet Settlements" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
						w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, labelXY.X, labelXY.Y)
						w.Printf("%s", tile.Features.Settlement.Name)
						w.Printf("</label>\n")
					}
				}
			}
		}
	}

	w.Printf("</labels>\n")

	w.Println(`<shapes>`)

	const riverWidth = 0.0625
	for gridRow := w.minGridRow; gridRow <= w.maxGridRow; gridRow++ {
		gridRowOffset := (gridRow - w.minGridRow) * rowsPerGrid
		for gridColumn := w.minGridColumn; gridColumn <= w.maxGridColumn; gridColumn++ {
			g := w.grids[gridRow][gridColumn]
			if g == nil {
				continue
			}
			gridColumnOffset := (gridColumn - w.minGridColumn) * columnsPerGrid
			for column := 0; column < columnsPerGrid; column++ {
				for row := 0; row < rowsPerGrid; row++ {
					tile := g.tiles[column][row]
					points := coordsToPoints(gridColumnOffset+column, gridRowOffset+row)

					// detect edges that are both Ford and River
					fordEdges := map[directions.Direction]bool{}

					var from, to Point

					for _, dir := range tile.Features.Edges.Ford {
						switch dir {
						case directions.DNorth:
							from, to = points[2], points[3]
						case directions.DNorthEast:
							from, to = points[3], points[4]
						case directions.DSouthEast:
							from, to = points[4], points[5]
						case directions.DSouth:
							from, to = points[5], points[6]
						case directions.DSouthWest:
							from, to = points[6], points[1]
						case directions.DNorthWest:
							from, to = points[1], points[2]
						default:
							panic(fmt.Sprintf("assert(direction != %d)", dir))
						}
						fordEdges[dir] = true

						ford := edgeCenter(dir, points)
						midpointFrom := midpoint(from, ford)
						midpointTo := midpoint(to, ford)

						w.Printf(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.6000000238418579,0.800000011920929,1.0,1.0" strokeWidth="%f" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`, riverWidth)
						w.Printf(` <p type="m" x="%f" y="%f"/>`, from.X, from.Y)
						w.Printf(` <p x="%f" y="%f"/>`, midpointFrom.X, midpointFrom.Y)
						w.Println(`</shape>`)

						w.Printf(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.6000000238418579,0.800000011920929,1.0,1.0" strokeWidth="%f" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`, riverWidth)
						w.Printf(` <p type="m" x="%f" y="%f"/>`, midpointTo.X, midpointTo.Y)
						w.Printf(` <p x="%f" y="%f"/>`, to.X, to.Y)
						w.Println(`</shape>`)
					}

					for _, dir := range tile.Features.Edges.River {
						// if we have both a ford and a river, honor the ford
						if fordEdges[dir] {
							continue
						}
						switch dir {
						case directions.DNorth:
							from, to = points[2], points[3]
						case directions.DNorthEast:
							from, to = points[3], points[4]
						case directions.DSouthEast:
							from, to = points[4], points[5]
						case directions.DSouth:
							from, to = points[5], points[6]
						case directions.DSouthWest:
							from, to = points[6], points[1]
						case directions.DNorthWest:
							from, to = points[1], points[2]
						default:
							panic(fmt.Sprintf("assert(direction != %d)", dir))
						}
						w.Printf(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.6000000238418579,0.800000011920929,1.0,1.0" strokeWidth="%f" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`, riverWidth)
						w.Printf(` <p type="m" x="%f" y="%f"/>`, from.X, from.Y)
						w.Printf(` <p x="%f" y="%f"/>`, to.X, to.Y)
						w.Println(`</shape>`)
					}

					for _, dir := range tile.Features.Edges.StoneRoad {
						// get the center of the hex we're in
						center := points[0]

						// get the midpoint of the segment from the center to the edge
						segmentEnd := edgeCenter(dir, points)
						segmentStart := midpoint(midpoint(center, segmentEnd), segmentEnd)

						w.Printf(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" fillColor="0.7019608020782471,0.7019608020782471,0.7019608020782471,1.0" strokeColor="0.7019608020782471,0.7019608020782471,0.7019608020782471,1.0" strokeWidth="0.05" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
						w.Printf(` <p type="m" x="%f" y="%f"/>`, segmentStart.X, segmentStart.Y)
						w.Printf(` <p x="%f" y="%f"/>`, segmentEnd.X, segmentEnd.Y)
						w.Println(`</shape>`)
					}
				}
			}
		}
	}

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

// crs_to_pixel converts a column, row to the pixel at the center of the corresponding tile.
//
// ok. the world map doesn't draw regular hexagons. they're flattened slightly.
// It's easier to call them "tiles" since they aren't regular hexagons. Anyway, I
// estimated the size of the tiles by looking at the output for labels in different
// scenarios.
//
// I came up a tile size of 300 pixels for height and 300 pixels for width.
func crs_to_pixel(column, row int, _ bool) Point {
	const height, width = 300, 300
	const halfHeight, threeQuarterWidth = height / 2, width * 3 / 4
	const leftMargin, topMargin = 0, 0

	var x, y float64

	x = float64(column) * threeQuarterWidth
	if column&2 == 1 { // shove odd rows down half the height of a tile
		y = float64(row)*halfHeight + halfHeight
	} else {
		y = float64(row) * halfHeight
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

type Settlement struct {
	UUID string
	Name string
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
