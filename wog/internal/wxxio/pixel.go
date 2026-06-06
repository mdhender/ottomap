package wxxio

import (
	"math"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/hex"
)

// Worldographer stores feature and label coordinates against an "ideal"
// hex of 300x300 units, independent of the hexWidth/hexHeight the editor
// renders with. Flat-top columns advance 3/4 of a hex (225 units) and
// stagger odd columns down by half a hex (150 units); pointy-top rows are
// the mirror image. A feature anchored to a tile sits at that tile's
// center, which is half a hex (150 units) in from the grid origin.
const (
	idealHexSize = 300.0            // ideal tile width and height in file units
	idealStep    = idealHexSize * 3 / 4 // 225: spacing along the staggered axis
	idealHalf    = idealHexSize / 2  // 150: tile center offset and stagger
)

// parity returns n mod 2 in {0,1}, correct for negative n.
func parity(n int) int { return ((n % 2) + 2) % 2 }

// PixelToAxial converts on-disk pixel coordinates into an anchor hex plus
// a sub-tile pixel offset. Symmetric with AxialToPixel, so round-tripping
// a value produced by AxialToPixel reconstructs the original coordinates.
func PixelToAxial(px, py float64, orientation string) (hex.Axial, ottomap.Offset) {
	var col, row int
	if orientation == "ROWS" { // pointy-top: rows staggered
		row = int(math.Round((py - idealHalf) / idealStep))
		col = int(math.Round((px - idealHalf - idealHalf*float64(parity(row))) / idealHexSize))
	} else { // COLUMNS flat-top: columns staggered
		col = int(math.Round((px - idealHalf) / idealStep))
		row = int(math.Round((py - idealHalf - idealHalf*float64(parity(col))) / idealHexSize))
	}
	cx, cy := offsetToPixel(col, row, orientation)
	layout := hex.OddQ
	if orientation == "ROWS" {
		layout = hex.OddR
	}
	return hex.FromOffset(hex.OffsetCoord{Col: col, Row: row}, layout),
		ottomap.Offset{DX: px - cx, DY: py - cy}
}

// AxialToPixel converts a hex anchor plus sub-tile offset back into the
// absolute pixel coordinates Worldographer uses. Symmetric with
// PixelToAxial.
func AxialToPixel(a hex.Axial, off ottomap.Offset, orientation string) (float64, float64) {
	layout := hex.OddQ
	if orientation == "ROWS" {
		layout = hex.OddR
	}
	oc := a.ToOffset(layout)
	cx, cy := offsetToPixel(oc.Col, oc.Row, orientation)
	return cx + off.DX, cy + off.DY
}

// offsetToPixel returns the pixel center of the tile at (col, row) in the
// ideal coordinate system Worldographer uses for features and labels.
func offsetToPixel(col, row int, orientation string) (float64, float64) {
	if orientation == "ROWS" { // pointy-top: rows staggered
		x := float64(col)*idealHexSize + idealHalf*float64(parity(row)) + idealHalf
		y := float64(row)*idealStep + idealHalf
		return x, y
	}
	// COLUMNS flat-top: columns staggered
	x := float64(col)*idealStep + idealHalf
	y := float64(row)*idealHexSize + idealHalf*float64(parity(col)) + idealHalf
	return x, y
}

// TileCoord converts an on-disk (rowIdx, tileIdx) position within a
// <tiles> grid into axial coordinates. Worldographer emits the same
// column-major layout for both orientations: there are tilesWide
// <tilerow> elements (despite the name), each containing tilesHigh tile
// lines. So rowIdx is always the column and tileIdx is always the row;
// only the offset layout (OddQ vs OddR) changes with orientation.
func TileCoord(orientation string, rowIdx, tileIdx int) hex.Axial {
	layout := hex.OddQ
	if orientation == "ROWS" {
		layout = hex.OddR
	}
	return hex.FromOffset(hex.OffsetCoord{Col: rowIdx, Row: tileIdx}, layout)
}

// AxialRangeToOffset projects an axial bounding box onto the offset grid
// implied by orientation and returns the resulting offset bounding box
// along with the layout used.
func AxialRangeToOffset(min, max hex.Axial, orientation string) (hex.OffsetCoord, hex.OffsetCoord, hex.Layout) {
	layout := hex.OddQ
	if orientation == "ROWS" {
		layout = hex.OddR
	}
	corners := []hex.Axial{
		min,
		{Q: max.Q, R: min.R},
		{Q: min.Q, R: max.R},
		max,
	}
	first := corners[0].ToOffset(layout)
	mn, mx := first, first
	for _, c := range corners[1:] {
		o := c.ToOffset(layout)
		if o.Col < mn.Col {
			mn.Col = o.Col
		}
		if o.Row < mn.Row {
			mn.Row = o.Row
		}
		if o.Col > mx.Col {
			mx.Col = o.Col
		}
		if o.Row > mx.Row {
			mx.Row = o.Row
		}
	}
	return mn, mx, layout
}
