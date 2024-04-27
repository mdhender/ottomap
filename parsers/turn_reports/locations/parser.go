// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package locations

import (
	"errors"
	"log"
)

// ParseLocation parses the unit's location
func ParseLocation(input []byte) (*Hex, *Hex, error) {
	if input == nil {
		return nil, nil, errors.New("locations is nil")
	}
	var prev, curr *Hex
	x, err := Parse("locations", input)
	if err != nil {
		log.Printf("locations: %q\n", string(input))
		log.Fatalf("locations: %v\n", err)
	}
	l, ok := x.(Location)
	if !ok {
		log.Fatalf("locations: %T\n", x)
	}
	// log.Printf("locations: %+v\n", l)
	if l.PreviousHex.NA {
		log.Printf("locations: warning: previous hex is n/a\n")
	} else {
		prev = &Hex{Grid: l.PreviousHex.Grid, Row: l.PreviousHex.Row, Col: l.PreviousHex.Col}
	}
	if l.CurrentHex.NA {
		log.Printf("locations: warning: previous hex is n/a\n")
	} else {
		curr = &Hex{Grid: l.CurrentHex.Grid, Row: l.CurrentHex.Row, Col: l.CurrentHex.Col}
	}
	return prev, curr, nil
}
