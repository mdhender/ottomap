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
	Grid     string // AA ... ZZ
	Coords   Offset // coordinates in a grid hex are one-based
	Terrain  domain.Terrain
	Visited  bool
	Scouted  bool
	Features Features
}

// Tile is a hex on the Worldographer map.
type Tile struct {
	Terrain   domain.Terrain
	Elevation int
	IsIcy     bool
	IsGMOnly  bool
	Resources Resources
	Features  Features
}

// Features are things to display on the map
type Features struct {
	Edges struct {
		Ford  [6]bool
		Pass  [6]bool
		River [6]bool
	}
	Label      *Label
	Settlement *Settlement // name of settlement
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

func (w *WXX) Create(path string, hexes []*Hex, showGridNumbering, showGridCenter bool) error {
	// we must track the minimum and maximum grid coordinates so that we can create the larger map
	var minGridRow, maxGridRow, minGridColumn, maxGridColumn = 26, 0, 26, 0

	// group the hexes by grid
	grids := map[string][]*Hex{}
	for gridId, grid := range grids {
		log.Printf("map: grid %q: %4d hexes\n", gridId, len(grid))
	}
	for _, hex := range hexes {
		if hex.Grid == "" {
			panic(fmt.Sprintf("assert(hex.Grid!= %q)", hex.Grid))
		}

		grids[hex.Grid] = append(grids[hex.Grid], hex)

		// update the minimum and maximum grid coordinates
		gridRow, gridColumn := int(hex.Grid[0]-'A'), int(hex.Grid[1]-'A')
		if gridRow < minGridRow {
			minGridRow = gridRow
		}
		if gridRow > maxGridRow {
			maxGridRow = gridRow
		}
		if gridColumn < minGridColumn {
			minGridColumn = gridColumn
		}
		if gridColumn > maxGridColumn {
			maxGridColumn = gridColumn
		}
	}
	for gridId, grid := range grids {
		log.Printf("map: grid %q: %4d hexes\n", gridId, len(grid))
	}

	// create any missing grids
	for gridRow := minGridRow; gridRow <= maxGridRow; gridRow++ {
		for gridColumn := minGridColumn; gridColumn <= maxGridColumn; gridColumn++ {
			gridId := fmt.Sprintf("%c%c", gridRow+'A', gridColumn+'A')
			if _, ok := grids[gridId]; !ok {
				grids[gridId] = []*Hex{}
			}
		}
	}
	for gridId, grid := range grids {
		log.Printf("map: grid %q: %4d hexes\n", gridId, len(grid))
	}
	log.Printf("map: grid (%c%c %c%c) (%c%c %c%c)", minGridRow+'A', minGridColumn+'A', maxGridRow+'A', maxGridColumn+'A', minGridRow+'A', minGridColumn+'A', maxGridRow+'A', maxGridColumn+'A')

	// one grid on the consolidated map is 30 columns wide by 21 rows high.
	const columnsPerGrid, rowsPerGrid = 30, 21

	// calculate the size of the consolidated map
	gridsWide, gridsHigh := int(maxGridColumn-minGridColumn)+1, int(maxGridRow-minGridRow)+1
	tilesWide, tilesHigh := columnsPerGrid*gridsWide, rowsPerGrid*gridsHigh
	log.Printf("map: grid columns %4d rows %4d", gridsWide, gridsHigh)
	log.Printf("map: tile columns %4d rows %4d", tilesWide, tilesHigh)

	// create a consolidated map of tiles
	var wmap [][]Tile
	for column := 0; column < tilesWide; column++ {
		wmap = append(wmap, make([]Tile, tilesHigh))
	}

	// convert the grids to tiles and then update the consolidated map.
	// It is indexed by column then row.
	for gridId, grid := range grids {
		log.Printf("map: grid %q: %4d hexes\n", gridId, len(grid))
		// extract this grid's coordinates and determine where it fits in the consolidated map.
		gridRow, gridColumn := int(gridId[0]-'A'), int(gridId[1]-'A')
		log.Printf("map: grid %q: row %4d col %4d: minRow %4d minColum %4d\n", gridId, gridRow, gridColumn, minGridRow, minGridColumn)
		gridRow, gridColumn = gridRow-minGridRow, gridColumn-minGridColumn
		log.Printf("map: grid %q: row %4d col %4d\n", gridId, gridRow, gridColumn)
		mapColumn, mapRow := gridColumn*columnsPerGrid, gridRow*rowsPerGrid
		log.Printf("map: grid %q: row %4d col %4d: map (%4d, %4d)", gridId, gridRow, gridColumn, mapColumn, mapRow)

		gridTiles, err := w.CreateGrid(grid, showGridNumbering)
		if err != nil {
			return err
		}

		for column, tiles := range gridTiles {
			for row, tile := range tiles {
				// log.Printf("map: grid %s: (%4d, %4d): tile (%4d, %4d)", gridId, mapColumn, mapRow, column, row)
				wmap[mapColumn+column][mapRow+row] = tile
			}
		}
	}

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
	w.Println(`<maplayer name="Tribenet Settlements" isVisible="true"/>`)
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
	for col := 0; col < tilesWide; col++ {
		w.Printf("<tilerow>\n")
		for row := 0; row < tilesHigh; row++ {
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

	// add features
	w.Println(`<features>`)

	for column := 0; column < tilesWide; column++ {
		for row := 0; row < tilesHigh; row++ {
			tile := wmap[column][row]

			points := coordsToPoints(column, row)

			if tile.Features.Settlement != nil {
				settlement := points[0]
				//w.Printf(`<feature type="Settlement City" rotate="0.0" uuid="%s" mapLayer="Tribenet Settlements" isFlipHorizontal="false" isFlipVertical="false" scale="35.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="%f" y="%f" /><label  mapLayer="Tribenet Settlements" style="City" fontFace=".AppleSystemUIFont" color="0.0,0.0,0.0,1.0" outlineColor="0.0,0.0,0.0,1.0" outlineSize="2.0" rotate="0.0" isBold="false" isItalic="false" isWorld="false" isContinent="false" isKingdom="false" isProvince="false" isGMOnly="false" tags=""><location viewLevel="WORLD" x="%f" y="%f" scale="12.5" /></label>`, tile.Features.Settlement.UUID, settlement.X, settlement.Y, settlement.X, settlement.Y)
				//w.Printf(`<feature type="Settlement City" rotate="0.0" uuid="%s" mapLayer="Tribenet Settlements" isFlipHorizontal="false" isFlipVertical="false" scale="35.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="%f" y="%f" /><label  mapLayer="Tribenet Settlements" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="0.0,0.0,0.0,1.0" outlineSize="2.0" rotate="0.0" isBold="false" isItalic="false" isWorld="false" isContinent="false" isKingdom="false" isProvince="false" isGMOnly="false" tags=""><location viewLevel="WORLD" x="%f" y="%f" scale="12.5" /></label>`, tile.Features.Settlement.UUID, settlement.X, settlement.Y, settlement.X, settlement.Y)
				w.Printf(`<feature type="Settlement City" rotate="0.0" uuid="%s" mapLayer="Tribenet Settlements" isFlipHorizontal="false" isFlipVertical="false" scale="35.0" scaleHt="-1.0" tags="" color="null" ringcolor="null" isGMOnly="false" isPlaceFreely="false" labelPosition="6:00" labelDistance="0" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isFillHexBottom="false" isHideTerrainIcon="false"><location viewLevel="WORLD" x="%f" y="%f" />`, tile.Features.Settlement.UUID, settlement.X, settlement.Y)
				w.Println(`</feature>`)
			}
		}
	}

	w.Println(`</features>`)

	w.Printf("<labels>\n")

	for column := 0; column < tilesWide; column++ {
		for row := 0; row < tilesHigh; row++ {
			points := coordsToPoints(column, row)

			if showGridCenter {
				w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
				w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, points[0].X, points[0].Y)
				if column&1 == 0 {
					w.Printf("%d", column&1)
				} else {
					w.Printf("%d", column&1)
				}
				w.Printf("</label>\n")
			}

			if showGridNumbering {
				label := fmt.Sprintf("%02d%02d", (column%columnsPerGrid)+1, (row%rowsPerGrid)+1)
				labelXY := bottomLeftCenter(points)
				w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
				w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, labelXY.X-15, labelXY.Y-2.5)
				w.Printf("%s", label)
				w.Printf("</label>\n")
			}

			// add labels to tiles when needed.
			tile := wmap[column][row]
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

	//x, y := 0, 0
	//center := Point{float64(x) * 300, float64(y)*300 + 150}
	//if x%2 == 0 {
	//	center.X += 150.0
	//} else {
	//	center.Y += 150.0
	//}
	//w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, center.X, center.Y)
	//w.Printf("+")
	//w.Printf("</label>\n")
	//
	//x, y = x+1, y+0
	//center = Point{float64(x) * 300, float64(y)*300 + 150}
	//if x%2 == 0 {
	//} else {
	//	center.X += 75
	//	center.Y += 150.0
	//}
	//w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, center.X, center.Y)
	//w.Printf("+")
	//w.Printf("</label>\n")
	//
	//x, y = x+1, y+0
	//center = Point{float64(x) * 300, float64(y)*300 + 150}
	//if x%2 == 0 {
	//	center.X += 0.0
	//} else {
	//	center.Y += 150.0
	//}
	//w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, center.X, center.Y)
	//w.Printf("+")
	//w.Printf("</label>\n")
	//
	//x, y = x+1, y+0
	//center = Point{float64(x)*300 + 75 - 75 - 75, float64(y)*300 + 150}
	//if x%2 == 0 {
	//} else {
	//	center.X += 0.0
	//	center.Y += 150.0
	//}
	//w.Printf(`<label  mapLayer="Labels" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
	//w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="12.5" />`, center.X, center.Y)
	//w.Printf("+")
	//w.Printf("</label>\n")

	w.Printf("</labels>\n")

	w.Println(`<shapes>`)
	//w.Println(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.0,0.0,0.0,1.0" strokeWidth="0.03" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
	//w.Println(` <p type="m" x="225.0" y = "0.0"/>`)
	//w.Println(`</shape>`)

	var line [2]Point

	line = [2]Point{{75.0, 0.0}, {225.0, 0.0}}
	for seg := 0; seg < 3; seg++ {
		w.Println(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="1.0,1.0,1.0,1.0" strokeWidth="0.03" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
		for n, p := range line {
			if n == 0 {
				w.Printf(` <p type="m" x="%f" y="%f"/>`+"\n", p.X, p.Y)
			} else {
				w.Printf(` <p x="%f" y="%f"/>`+"\n", p.X, p.Y)
			}
		}
		line[0].Y, line[1].Y = line[0].Y+150.0, line[1].Y+150.0
		w.Println(`</shape>`)
	}

	w.Println(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="1.0,1.0,1.0,1.0" strokeWidth="0.03" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
	w.Println(` <p type="m" x="75.0" y = "0.0"/>`)
	w.Println(` <p x="225.0" y="0.0"/>`)
	w.Println(`</shape>`)

	w.Println(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.6000000238418579,0.800000011920929,1.0,1.0" strokeWidth="0.03" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
	w.Println(` <p type="m" x="450.0" y="0.0"/>`)
	w.Println(` <p x="750.0" y="0.0"/>`)
	w.Println(`</shape>`)

	w.Println(`<shape  type="Path" isCurve="false" isGMOnly="false" isSnapVertices="true" isMatchTileBorders="false" tags="" creationType="BASIC" isDropShadow="false" isInnerShadow="false" isBoxBlur="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" dsSpread="0.2" dsRadius="50.0" dsOffsetX="0.0" dsOffsetY="0.0" insChoke="0.2" insRadius="50.0" insOffsetX="0.0" insOffsetY="0.0" bbWidth="10.0" bbHeight="10.0" bbIterations="3" mapLayer="Above Terrain" fillTexture="" strokeTexture="" strokeType="SIMPLE" highestViewLevel="WORLD" currentShapeViewLevel="WORLD" lineCap="ROUND" lineJoin="ROUND" opacity="1.0" fillRule="NON_ZERO" strokeColor="0.6000000238418579,0.800000011920929,1.0,1.0" strokeWidth="0.03" dsColor="1.0,0.8941176533699036,0.7686274647712708,1.0" insColor="1.0,0.8941176533699036,0.7686274647712708,1.0">`)
	w.Println(` <p type="m" x="450.0" y="150.0"/>`)
	w.Println(` <p x="750.0" y="150.0"/>`)
	w.Println(`</shape>`)

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

var (
	// Define the offsets based on the flattened hexagon dimensions
	flattenedHexOffsets = [6][2]float64{
		{-150, 0},   // left vertex
		{-75, -150}, // top-left vertex
		{75, -150},  // top-right vertex
		{150, 0},    // right vertex
		{75, 150},   // bottom-right vertex
		{-75, 150},  // bottom-left vertex
	}
)

// coordsToPoints returns the center point and vertices of a hexagon centered at the
// given column and row. It converts the column, row to the pixel at the center of the
// corresponding tile, then calculates the vertices based on that point.
// The center point is the first value in the returned slice.
func coordsToPoints(column, row int) [7]Point {
	const height, width = 300, 300
	const halfHeight, oneQuarterWidth, threeQuarterWidth = height / 2, width / 4, width * 3 / 4
	const leftMargin, topMargin = width / 2, halfHeight

	// points is the center plus the six vertices
	var points [7]Point

	points[0].X = float64(column)*threeQuarterWidth + leftMargin
	if column&1 == 1 { // shove odd rows down half the height of a tile
		points[0].Y = float64(row)*height + halfHeight + topMargin
	} else {
		points[0].Y = float64(row)*height + topMargin
	}

	// Calculate vertices based on offsets from center
	for i, offset := range flattenedHexOffsets {
		points[i+1] = Point{
			X: points[0].X + offset[0],
			Y: points[0].Y + offset[1],
		}
	}

	return points
}

func bottomCenter(v [7]Point) Point {
	return Point{X: (v[5].X + v[6].X) / 2, Y: (v[5].Y + v[6].Y) / 2}
}

func bottomLeft(v [7]Point) Point {
	return Point{X: v[6].X, Y: v[6].Y}
}

func bottomLeftCenter(v [7]Point) Point {
	bc := bottomCenter(v)
	return Point{X: (v[6].X + bc.X) / 2, Y: bc.Y}
}

func settlementLabelXY(label string, v [7]Point) Point {
	return bottomCenter(v).Translate(Point{X: float64(-3 * len(label)), Y: -25})
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

func (p Point) Translate(t Point) Point {
	return Point{
		X: p.X + t.X,
		Y: p.Y + t.Y,
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
