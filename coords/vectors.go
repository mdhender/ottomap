// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import "github.com/mdhender/ottomap/directions"

// column direction vectors defines the vectors used to determine the coordinates
// of the neighboring column based on the direction and the odd/even column
// property of the starting hex.
//
// NB: grids start at 0101 and hexes at (0,0), so "odd" and "even" are based
//     on the hex coordinates, not the grid.

var OddColumnVectors = map[directions.Direction][2]int{
	directions.DNorth:     [2]int{+0, -1}, // ## 1206 -> (11, 05) -> (11, 04) -> ## 1205
	directions.DNorthEast: [2]int{+1, +0}, // ## 1206 -> (11, 05) -> (12, 05) -> ## 1306
	directions.DSouthEast: [2]int{+1, +1}, // ## 1206 -> (11, 05) -> (12, 06) -> ## 1307
	directions.DSouth:     [2]int{+0, +1}, // ## 1206 -> (11, 05) -> (11, 06) -> ## 1207
	directions.DSouthWest: [2]int{-1, +1}, // ## 1206 -> (11, 05) -> (10, 06) -> ## 1107
	directions.DNorthWest: [2]int{-1, +0}, // ## 1206 -> (11, 05) -> (10, 05) -> ## 1106
}

var EvenColumnVectors = map[directions.Direction][2]int{
	directions.DNorth:     [2]int{+0, -1}, // ## 1306 -> (12, 05) -> (12, 04) -> ## 1305
	directions.DNorthEast: [2]int{+1, -1}, // ## 1306 -> (12, 05) -> (13, 04) -> ## 1405
	directions.DSouthEast: [2]int{+1, +0}, // ## 1306 -> (12, 05) -> (13, 05) -> ## 1406
	directions.DSouth:     [2]int{+0, +1}, // ## 1306 -> (12, 05) -> (12, 06) -> ## 1307
	directions.DSouthWest: [2]int{-1, +0}, // ## 1306 -> (12, 05) -> (11, 05) -> ## 1206
	directions.DNorthWest: [2]int{-1, -1}, // ## 1306 -> (12, 05) -> (11, 04) -> ## 1205
}
