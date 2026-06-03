// Package ottomap is a Go-idiomatic library for building and manipulating
// hex-grid maps. The in-memory model is independent of any file format;
// readers and writers for specific formats live in subpackages
// (for example, github.com/mdhender/ottomap/wog for Worldographer).
//
// A zero-configuration Map is immediately valid:
//
//	m := ottomap.NewMap()
//	m.SetTerrain(hex.At(0, 0), "Grassland")
//	wog.Write(out, m, wog.WriteOptions{Version: wog.V2025})
//
// Coordinates are axial (hex.Axial) so the same map can be written by any
// adapter without orientation-dependent recomputation. The adapter selects
// the on-disk orientation at write time.
package ottomap
