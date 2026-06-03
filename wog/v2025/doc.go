// Package v2025 decodes and encodes Worldographer 2025 .wxx schemas
// (release="2025"). The package handles the full family of 2025 schemas
// (currently 1.01 through 1.10) at the granularity needed to map to and
// from the ottomap domain model.
package v2025

// Options controls how Encode writes a map.
type Options struct {
	Orientation int  // 0 = COLUMNS (flat-top), 1 = ROWS (pointy-top)
	Projection  int  // 0 = FLAT, 1 = ICOSAHEDRAL
	GMView      bool // emit showGMOnly=true in the editor view
}
