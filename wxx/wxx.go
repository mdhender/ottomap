// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
)

type WXX struct {
	buffer *bytes.Buffer

	// range of grids is AA to ZZ.
	// grids are stored in row-column format.
	grids                     [26][26]*Grid
	totalGrids                int
	minGridRow, minGridColumn int
	maxGridRow, maxGridColumn int
}
