// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
)

var (
	DebugBuffer   = &bytes.Buffer{}
	rxRiverEdge   *regexp.Regexp
	rxTerrainCost *regexp.Regexp
	rxWaterEdge   *regexp.Regexp
)

// ParseMovements parses the unit's movements
func ParseMovements(input []byte) (*Movements, error) {
	if rxRiverEdge == nil || rxTerrainCost == nil || rxWaterEdge == nil {
		// No Ford on River to NW of HEX
		if rx, err := regexp.Compile(`^No Ford on River to ([A-Z]+) of HEX$`); err != nil {
			return nil, fmt.Errorf("regex: river: %w", err)
		} else {
			rxRiverEdge = rx
		}
		// Not enough M.P's to move to NE into ROCKY HILLS
		if rx, err := regexp.Compile(`^Not enough M\.P's to move to ([A-Z]+) into (.+)$`); err != nil {
			return nil, fmt.Errorf("regex: terrain: %w", err)
		} else {
			rxTerrainCost = rx
		}
		// Can't Move on Lake to S of HEX
		// Can't Move on Ocean to NW of HEX
		if rx, err := regexp.Compile(`^Can't Move on ([LO][a-z]+) to ([A-Z]+) of HEX$`); err != nil {
			return nil, fmt.Errorf("regex: water: %w", err)
		} else {
			rxWaterEdge = rx
		}
	}

	m := &Movements{}
	if input == nil {
		// this normally happens only on the setup report.
		return m, nil
	} else if bytes.Equal(input, []byte(`Tribe Movement: Move \`)) {
		// unit did not move this turn.
		return m, nil
	}

	// skip the prefix
	input = input[21:]
	DebugBuffer.WriteString(fmt.Sprintf("input_ `%s`\n", string(input)))

	// MOVES is MOVE (SPACE SPACE?)? BACKSLASH MOVE)* BACKSLASH FAIL_MSG?
	moves, results, ok := bsSplit(input)
	DebugBuffer.WriteString(fmt.Sprintf(" moves `%s`\n", string(moves)))
	DebugBuffer.WriteString(fmt.Sprintf("  stat `%s`\n", string(results)))
	DebugBuffer.WriteString(fmt.Sprintf("    ok  %v\n\n", ok))

	//log.Printf("het, moves be `%s`\n", string(moves))
	for _, data := range bytes.Split(moves, []byte{'\\'}) {
		rawText := string(data)
		//log.Printf("het, data  be `%s`\n", string(data))
		data = bytes.TrimRight(bytes.TrimSpace(data), ",")
		//log.Printf("het, data  be `%s`\n", string(data))
		if prslt, err := Parse("movements.step", data); err != nil {
			log.Fatalf("%s: %v\n", string(data), err)
		} else if prslt == nil {
			log.Fatalf("%s: parse result is nil\n", string(data))
		} else if step, ok := prslt.(*Step); !ok {
			log.Fatalf("%s: parse result is %T\n", string(data), step)
		} else {
			//log.Printf("het, steps be %+v\n", *step)
			step.RawText = rawText
			m.Steps = append(m.Steps, step)
		}
	}

	if len(results) == 0 {
		//
	} else {
		m.Failed.RawText = string(results)
		if bytes.HasPrefix(results, []byte("Not enough M.P's")) {
			// Not enough M.P's to move to SW into GRASSY HILLS
			if matches := rxTerrainCost.FindStringSubmatch(m.Failed.RawText); len(matches) == 0 {
				log.Printf("parse: unit: terrain: failed but no terrain found\n")
			} else {
				// log.Printf("parse: unit: terrain: found %d %v\n", len(matches), matches)
				m.Failed.Direction = matches[1] // SW
				m.Failed.Terrain = matches[2]   // GRASSY HILLS
			}
		} else if bytes.HasPrefix(results, []byte("Can't Move on ")) {
			// Can't Move on Lake to S of HEX
			// Can't Move on Ocean to NW of HEX
			if matches := rxWaterEdge.FindStringSubmatch(m.Failed.RawText); len(matches) == 0 {
				log.Printf("parse: unit: water: failed but no water found\n")
			} else {
				// log.Printf("parse: unit: water: found %d %v\n", len(matches), matches)
				m.Failed.Edge = matches[1]      // Lake or Ocean
				m.Failed.Direction = matches[2] // S or NW
			}
		} else if bytes.HasPrefix(results, []byte("No Ford on River to ")) {
			// No Ford on River to NW of HEX
			if matches := rxRiverEdge.FindStringSubmatch(m.Failed.RawText); len(matches) == 0 {
				log.Printf("parse: unit: river: failed but no river found\n")
			} else {
				// log.Printf("parse: unit: river: found %d %v\n", len(matches), matches)
				m.Failed.Edge = "River"
				m.Failed.Direction = matches[1] // NW
			}
		}
	}

	return m, nil
}

// the input looks something like STUFF BACKSLASH STATUS.
// we'd like to use bytes.Cut, but STUFF can contain back-slashes.
// so we need to find the last back-slash and split there.
func bsSplit(input []byte) ([]byte, []byte, bool) {
	if pos := bytes.LastIndexByte(input, '\\'); pos >= 0 {
		return input[:pos], input[pos+1:], true
	}
	return input, nil, false
}
