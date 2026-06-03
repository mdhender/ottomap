package wog

import "github.com/mdhender/ottomap"

// Version identifies a major Worldographer release.
//
// Schema details within a major release (e.g. 2017 schema 1.73 vs 1.77) are
// handled internally; callers choose the major bucket.
type Version int

const (
	// V2017 selects the Worldographer 2017 (a.k.a. Hexographer) schema family.
	V2017 Version = 2017
	// V2025 selects the Worldographer 2025 schema family.
	V2025 Version = 2025
)

func (v Version) String() string {
	switch v {
	case V2017:
		return "Worldographer 2017"
	case V2025:
		return "Worldographer 2025"
	default:
		return "unknown"
	}
}

// Orientation is the hex orientation used in the on-disk file.
//
//   - Columns is "flat-top" hexes (Worldographer "COLUMNS").
//   - Rows is "pointy-top" hexes (Worldographer "ROWS").
type Orientation int

const (
	Columns Orientation = iota
	Rows
)

func (o Orientation) String() string {
	switch o {
	case Columns:
		return "COLUMNS"
	case Rows:
		return "ROWS"
	default:
		return "unknown"
	}
}

// WriteOptions controls how a Map is serialized to a .wxx file.
//
// Version is the only required field; the zero value of WriteOptions is
// invalid because no default Version makes sense.
type WriteOptions struct {
	// Version is the target Worldographer release. Required.
	Version Version

	// Orientation is the hex orientation on disk. Default: Columns.
	Orientation Orientation

	// Projection is the map projection. Default: ottomap.Flat.
	Projection ottomap.Projection

	// GMView causes the writer to emit the file in "GM view" mode (showGMOnly
	// is true, so GM-only features and labels are visible by default in the
	// editor).
	GMView bool

	// Compress controls gzip compression of the output. Default: true. Set
	// to false to write raw UTF-16 XML (useful for debugging).
	Compress *bool

	// UTF16BE controls big-endian UTF-16 output. Default: true (the format
	// Worldographer reads). Set to false to write little-endian UTF-16.
	UTF16BE *bool
}

func (o WriteOptions) compress() bool {
	if o.Compress == nil {
		return true
	}
	return *o.Compress
}

func (o WriteOptions) utf16BE() bool {
	if o.UTF16BE == nil {
		return true
	}
	return *o.UTF16BE
}
