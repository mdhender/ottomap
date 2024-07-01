// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/internal/direction"
	"log"
	"strconv"
	"strings"
)

func Move(hex string, d direction.Direction_e) string {
	if hex == "N/A" || strings.HasPrefix(hex, "##") {
		log.Printf("error: hex %q direction %q: %v\n", hex, d, fmt.Errorf("bad hex"))
		return hex
	}
	from, err := HexToMap(hex)
	if err != nil {
		log.Printf("error: hex %q direction %q: %v\n", hex, d, err)
		panic(err)
	}
	to := from.Add(d)
	//log.Printf("from %s to %s\n", from, to)
	return to.ToHex()
}

func HexToMap(hex string) (Map, error) {
	if hex == "N/" || strings.HasPrefix(hex, "##") {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if !(len(hex) == 7 && hex[2] == ' ') {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	}
	grid, digits, ok := strings.Cut(hex, " ")
	if !ok {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if len(grid) != 2 {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if strings.TrimRight(grid, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if len(digits) != 4 {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if strings.TrimRight(digits, "0123456789") != "" {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	}
	bigMapRow, bigMapColumn := int(grid[0]-'A'), int(grid[1]-'A')
	littleMapColumn, err := strconv.Atoi(digits[:2])
	if err != nil {
		panic(err)
	}
	littleMapRow, err := strconv.Atoi(digits[2:])
	if err != nil {
		panic(err)
	}
	// log.Printf("hex %q brow %2d bcol %2d mcol %2d mrow %2d\n", hex, bigMapRow, bigMapColumn, littleMapColumn, littleMapRow)
	return Map{
		Column: bigMapColumn*30 + littleMapColumn - 1,
		Row:    bigMapRow*21 + littleMapRow - 1,
	}, nil

	return Map{}, nil
}
