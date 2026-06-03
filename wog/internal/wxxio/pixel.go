package wxxio

import (
	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/hex"
)

// DefaultCellWidth and DefaultCellHeight are the conventional pixel
// dimensions used by the pixel<->axial conversion. Worldographer's true
// hex size depends on hexWidth/hexHeight on <map>; we approximate with a
// fixed 40-pixel rectangular cell.
const (
	DefaultCellWidth  = 40.0
	DefaultCellHeight = 40.0
)

// PixelToAxial converts on-disk pixel coordinates into an anchor hex plus
// a sub-tile pixel offset. The result is approximate but symmetric with
// AxialToPixel, so round-tripping a value produced by AxialToPixel
// reconstructs the original pixel coordinates exactly.
func PixelToAxial(px, py float64, orientation string) (hex.Axial, ottomap.Offset) {
	col := int(px / DefaultCellWidth)
	row := int(py / DefaultCellHeight)
	ox := px - float64(col)*DefaultCellWidth
	oy := py - float64(row)*DefaultCellHeight
	layout := hex.OddQ
	if orientation == "ROWS" {
		layout = hex.OddR
	}
	return hex.FromOffset(hex.OffsetCoord{Col: col, Row: row}, layout), ottomap.Offset{DX: ox, DY: oy}
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
	return float64(oc.Col)*DefaultCellWidth + off.DX, float64(oc.Row)*DefaultCellHeight + off.DY
}

// TileCoord converts an on-disk (rowIdx, tileIdx) position within a
// <tiles> grid into axial coordinates. The meaning of rowIdx vs tileIdx
// depends on orientation: in COLUMNS mode rowIdx is the column index, in
// ROWS mode it is the row index.
func TileCoord(orientation string, rowIdx, tileIdx int) hex.Axial {
	if orientation == "ROWS" {
		return hex.FromOffset(hex.OffsetCoord{Col: tileIdx, Row: rowIdx}, hex.OddR)
	}
	return hex.FromOffset(hex.OffsetCoord{Col: rowIdx, Row: tileIdx}, hex.OddQ)
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
