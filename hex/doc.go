// Package hex provides axial coordinates and hex-grid math.
//
// Axial coordinates address tiles on an infinite hexagonal grid with a pair
// of integers (Q, R). They are independent of how the grid is rendered
// (flat-top vs pointy-top, odd vs even offset rows). Renderers and file
// formats that need offset coordinates can convert with ToOffset / FromOffset.
//
// References: https://www.redblobgames.com/grids/hexagons/
package hex
