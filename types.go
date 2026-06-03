package ottomap

import "image/color"

// Terrain identifies the terrain of a Tile. It is an opaque string whose
// meaning is fixed by the adapter that will read or write the map. The
// Worldographer adapter expects names like "Grassland" or "Forest";
// game-specific adapters (e.g. Olympia) may give those same names different
// semantic meaning.
type Terrain string

// Layer is a free-form layer name. Most adapters ship with a small set of
// well-known layer names; callers are free to invent more.
type Layer string

// Common Worldographer layer names. Provided as a convenience; any string
// is a valid Layer.
const (
	LayerAboveTerrain Layer = "Above Terrain"
	LayerBelowTerrain Layer = "Below Terrain"
	LayerAboveAll     Layer = "Above All"
	LayerBelowAll     Layer = "Below All"
)

// Projection is how the map is laid out on the surface of a world.
type Projection int

const (
	// Flat is a planar, rectangular projection.
	Flat Projection = iota
	// Icosahedral wraps the grid onto an icosahedron.
	Icosahedral
)

// Scope is the visibility scope of a label. It mirrors Worldographer's four
// nested zoom levels.
type Scope int

const (
	ScopeWorld Scope = iota
	ScopeContinent
	ScopeKingdom
	ScopeProvince
)

// Offset is a sub-tile placement offset. Units are arbitrary screen units
// interpreted by the renderer; (0, 0) is the tile's anchor.
type Offset struct {
	DX, DY float64
}

// FontSpec describes a label font.
type FontSpec struct {
	Family string
	Size   float64
	Bold   bool
	Italic bool
}

// Outline describes a label outline.
type Outline struct {
	Color color.RGBA
	Size  float64
}

// Resources records how much of each renewable or non-renewable resource is
// present on a tile. Each component is a percentage (0..100). The zero value
// means "no resources".
type Resources struct {
	Animal uint8
	Brick  uint8
	Crops  uint8
	Gems   uint8
	Lumber uint8
	Metals uint8
	Rock   uint8
}
