package ottomap

import (
	"image/color"

	"github.com/mdhender/ottomap/hex"
)

// Feature is a placed map element (city icon, tree, tower, ...). The Kind
// field is an opaque string; the meaning is fixed by the adapter that will
// write the map. Worldographer ships hundreds of named icons, and the names
// change between versions, so callers normalize Kind in their own code.
//
// The zero value of Feature is not directly useful; at minimum, Kind and
// Location should be set before AddFeature.
type Feature struct {
	// ID uniquely identifies this Feature within a Map. If empty when passed
	// to Map.AddFeature, the Map assigns a stable identifier.
	ID string

	// Kind is the feature's icon or type name (opaque to ottomap).
	Kind string

	// Location is the hex this Feature is anchored to.
	Location hex.Axial

	// Offset displaces the feature from its tile anchor.
	Offset Offset

	// Label is an optional text label rendered with the feature.
	Label string

	// Color is the primary feature color. Zero value means "use a default".
	Color color.RGBA

	// Scale is the rendered size, where 1.0 is "natural" size.
	Scale float64

	// Rotation is rotation in degrees.
	Rotation float64

	// Layer controls draw order. Empty means the writer picks a default.
	Layer Layer

	// Tags are free-form user tags. The adapter joins/splits them as needed
	// when serializing.
	Tags []string

	// GMOnly hides the feature from non-GM views.
	GMOnly bool
}

func (f Feature) clone() Feature {
	out := f
	if f.Tags != nil {
		out.Tags = append([]string(nil), f.Tags...)
	}
	return out
}
