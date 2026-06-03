package ottomap

import "image/color"

// Tile is a single hex on the map. It is a value type; pass it by value, and
// mutate the Map through SetTile or the convenience setters on Map.
//
// The zero value is a valid blank Tile.
type Tile struct {
	Terrain    Terrain
	Elevation  float64
	Icy        bool
	GMOnly     bool
	Resources  Resources
	Background color.RGBA // zero value means "use the terrain's default color"
}
