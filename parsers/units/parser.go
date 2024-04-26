// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package units

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/parsers/units/locations"
	"github.com/mdhender/ottomap/parsers/units/movements"
	"os"
	"strconv"
)

// Parse extracts unit information from the input.
// Returns a structure containing the unit details.
func Parse(input InputFile) (*Unit, error) {

	unit := &Unit{
		Id: input.Unit,
	}
	data, err := os.ReadFile(input.File)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte{'\n'})
	//log.Printf("parse: unit: data %d bytes, %d lines\n", len(data), len(lines))

	var locationLine, movementLine, statusLine []byte
	var scoutingLines [][]byte
	for n, line := range lines {
		if n == 0 {
			locationLine = line
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			movementLine = line
		} else if bytes.HasPrefix(line, []byte("Scout ")) {
			scoutingLines = append(scoutingLines, line)
		} else if bytes.HasPrefix(line, []byte(unit.Id+" Status: ")) {
			statusLine = line
		}
	}

	if start, finish, err := locations.ParseLocation(locationLine); err != nil {
		return nil, fmt.Errorf("unit: locations: %w", err)
	} else {
		if start != nil {
			unit.Started = &Hex{Grid: start.Grid}
			if unit.Started.Row, err = strconv.Atoi(start.Row); err != nil {
				return nil, fmt.Errorf("unit: locations: row: %w", err)
			}
			if unit.Started.Col, err = strconv.Atoi(start.Col); err != nil {
				return nil, fmt.Errorf("unit: locations: col: %w", err)
			}
		}
		if finish != nil {
			unit.Finished = &Hex{Grid: start.Grid}
			if unit.Finished.Row, err = strconv.Atoi(finish.Row); err != nil {
				return nil, fmt.Errorf("unit: locations: row: %w", err)
			}
			if unit.Finished.Col, err = strconv.Atoi(finish.Col); err != nil {
				return nil, fmt.Errorf("unit: locations: col: %w", err)
			}
		}
	}

	if moves, err := movements.ParseMovements(movementLine); err != nil {
		return nil, fmt.Errorf("unit: movement: %w", err)
	} else { // convert the movements
		unit.Movements = &Movements{}
		unit.Movements.Moves = make([]*Movement, len(moves.Moves))
		for i, m := range moves.Moves {
			unit.Movements.Moves[i] = &Movement{
				Direction: m.Direction,
				Result:    m.Result,
				Raw:       m.Raw,
			}
		}
		unit.Movements.Failed.Direction = moves.Failed.Direction
		unit.Movements.Failed.Edge = moves.Failed.Edge
		unit.Movements.Failed.Terrain = moves.Failed.Terrain
		unit.Movements.Failed.Text = moves.Failed.Text
		unit.Movements.Found = make([]string, len(moves.Found))
		for i, found := range moves.Found {
			unit.Movements.Found[i] = found
		}
	}

	if err = parseScoutingResults(scoutingLines); err != nil {
		return nil, fmt.Errorf("unit: scouting: %w", err)
	} else if unit.Status, err = parseStatus(statusLine); err != nil {
		return nil, fmt.Errorf("unit: status: %w", err)
	}
	return unit, nil
}

func parseScoutingResults(lines [][]byte) error {
	return nil
}

/*
0138 Status: PRAIRIE, 0138

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA SPACE UNIT (COMMA SPACE UNIT)*

1138 Status: CONIFER HILLS, O SE, SW, S, 1138, 1138e1

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA OCEAN SPACE DIRECTION (COMMA SPACE DIRECTION)* SPACE UNIT (COMMA SPACE UNIT)*

0138e1 Status: PRAIRIE,,River S 0138e1

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA COMMA EDGE DIRECTION SPACE UNIT (COMMA SPACE UNIT)*

2138 Status: PRAIRIE, O S,,Ford SE 2138, 0138

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA OCEAN SPACE DIRECTION (COMMA SPACE DIRECTION)* COMMA EDGE DIRECTION SPACE UNIT (COMMA SPACE UNIT)*

0138c1 Status: GRASSY HILLS, Lothal, O SW,,River N,Ford NW 0138c1, 0108c1, 0117e1, 0108g1, 0199c1

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA SPACE SETTLEMENT COMMA SPACE OCEAN DIRECTION COMMA SPACE DIRECTION COMMA UNIT (COMMA SPACE UNIT)*

0138c1 Status: GRASSY HILLS, Lothal, O SW,,River N,Ford NW 0138c1, 0108c1, 0117e1, 0108g1, 0199c1

UNIT SPACE "Status" COLON SPACE TERRAIN COMMA SETTLEMENT COMMA OCEAN SPACE DIRECTION (COMMA SPACE DIRECTION)* EDGE SPACE DIRECTION (COMMA EDGE SPACE DIRECTION)* SPACE UNIT (COMMA SPACE UNIT)*
*/
func parseStatus(line []byte) (string, error) {
	if line == nil {
		return "", nil
	}
	return string(line), nil
}
