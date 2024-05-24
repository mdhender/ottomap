// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import "log"

// MergeHexes merges the hexes into the consolidated map, creating new grids as necessary.
// It returns the first error encountered merging the new hexes.
func (w *WXX) MergeHexes(turnId string, hexes []*Hex) error {
	for _, hex := range hexes {
		if err := w.mergeHex(turnId, hex); err != nil {
			return err
		}
	}
	log.Printf("wxx: %s: merge: hexes %6d: grids %6d\n", turnId, len(hexes), w.totalGrids)
	return nil
}

// mergeHex merges the hex into the consolidated map, creating new grids and tiles as necessary.
// It returns the first error encountered merging the new hex.
func (w *WXX) mergeHex(turnId string, hex *Hex) error {
	gridId := hex.GridId
	gridRow, gridColumn := gridIdToRowColumn(gridId)

	// create a new grid if necessary
	g := w.grids[gridRow][gridColumn]
	if g == nil {
		if w.totalGrids == 0 {
			// this is the first grid we've seen, so initialize the min and max grid coordinates
			w.minGridRow, w.minGridColumn = gridRow, gridColumn
			w.maxGridRow, w.maxGridColumn = gridRow, gridColumn
		}

		w.grids[gridRow][gridColumn] = w.newGrid(gridId)
		g = w.grids[gridRow][gridColumn]
		w.totalGrids++

		// track the bounds of the populated grids on the map
		if gridRow < w.minGridRow {
			w.minGridRow = gridRow
		} else if gridRow > w.maxGridRow {
			w.maxGridRow = gridRow
		}
		if gridColumn < w.minGridColumn {
			w.minGridColumn = gridColumn
		} else if gridColumn > w.maxGridColumn {
			w.maxGridColumn = gridColumn
		}
	}

	// add the hex to the grid as a tile, returning any error
	return g.addTile(turnId, hex)
}

func (w *WXX) addGridCoords() {
	for row := 0; row < 26; row++ {
		for col := 0; col < 26; col++ {
			if w.grids[row][col] == nil {
				continue
			}
			w.grids[row][col].addCoords()
		}
	}
}

func (w *WXX) addGridNumbers() {
	for row := 0; row < 26; row++ {
		for col := 0; col < 26; col++ {
			if w.grids[row][col] == nil {
				continue
			}
			w.grids[row][col].addNumbers()
		}
	}
}
