// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package hexes

import (
	"fmt"
	"strconv"
	"strings"
)

type GridCoords struct {
	Row        int // range A .. Z
	Column     int // range A .. Z
	GridColumn int // range 01 .. 30
	GridRow    int // range 01 .. 21
}

func (gc GridCoords) String() string {
	return fmt.Sprintf("%c%c %02d%02d", 'A'+gc.Row, 'A'+gc.Column, gc.GridColumn+1, gc.GridRow+1)
}

type MapCoords struct {
	Row    int
	Column int
}

func GridCoordsFromString(s string) (GridCoords, bool) {
	var gc GridCoords
	if !(len(s) == 7 && s[2] == ' ') {
		return GridCoords{}, false
	}
	if strings.HasPrefix(s, "##") {
		gc.Row, gc.Column = 0, 0
	} else {
		if gc.Row = int(s[0] - 'A'); !(0 <= gc.Row && gc.Row < 26) {
			return GridCoords{}, false
		}
		if gc.Column = int(s[1] - 'A'); !(0 <= gc.Column && gc.Column < 26) {
			return GridCoords{}, false
		}
	}
	var err error
	if gc.GridColumn, err = strconv.Atoi(s[3:5]); err != nil {
		return GridCoords{}, false
	} else if gc.GridColumn = gc.GridColumn - 1; !(0 <= gc.GridColumn && gc.GridColumn < 30) {
		return GridCoords{}, false
	}
	if gc.GridRow, err = strconv.Atoi(s[5:]); err != nil {
		return GridCoords{}, false
	} else if gc.GridRow = gc.GridRow - 1; !(0 <= gc.GridRow && gc.GridRow < 21) {
		return GridCoords{}, false
	}
	return gc, true
}

func (gc GridCoords) ToMapCoords(s string) (MapCoords, error) {
	return MapCoords{
		Row:    gc.Row*31 + gc.GridRow,
		Column: gc.Column*21 + gc.GridColumn,
	}, nil
}

func (mc MapCoords) String() string {
	return fmt.Sprintf("(%d, %d)", mc.Row, mc.Column)
}
