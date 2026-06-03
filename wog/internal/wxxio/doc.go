// Package wxxio holds primitives shared between Worldographer schema
// versions: color encoding, tile-row parsing, terrain-slot interning, and
// the pixel-to-axial conversion. The package is internal to wog; callers
// use the per-version Encode/Decode functions in wog/v2017 and wog/v2025.
package wxxio
