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
	} else { // convert the movement steps
		unit.Movement = &Movement{}
		unit.Movement.Steps = make([]*Step, len(moves.Steps))
		for i, m := range moves.Steps {
			unit.Movement.Steps[i] = &Step{
				Direction: m.Direction,
				Terrain:   m.Terrain,
				Edges:     m.Edges,
				// Found: ?
				Settlement: m.Settlement,
				RawText:    m.RawText,
			}
		}
		unit.Movement.Failed.Direction = moves.Failed.Direction
		unit.Movement.Failed.Edge = moves.Failed.Edge
		unit.Movement.Failed.Terrain = moves.Failed.Terrain
		unit.Movement.Failed.Text = moves.Failed.RawText
		unit.Movement.Found = make([]string, len(moves.Found))
		for i, found := range moves.Found {
			unit.Movement.Found[i] = found
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

func parseStatus(line []byte) (string, error) {
	if line == nil {
		return "", nil
	}
	return string(line), nil
}
