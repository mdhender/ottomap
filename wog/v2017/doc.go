// Package v2017 decodes and encodes Worldographer 2017 (Hexographer) .wxx
// schemas (release="2017"). At the time of writing this package is a stub:
// Decode and Encode return errors. It exists so that the wog dispatcher
// reports a meaningful unsupported error and so its future drop-in won't
// require changes to the top-level wog package API.
package v2017

// Options controls how Encode writes a map.
type Options struct {
	Orientation int  // 0 = COLUMNS (flat-top), 1 = ROWS (pointy-top)
	Projection  int  // 0 = FLAT, 1 = ICOSAHEDRAL
	GMView      bool
}
