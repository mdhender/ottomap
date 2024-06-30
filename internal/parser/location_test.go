// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser_test

import (
	"github.com/mdhender/ottomap/internal/compass"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/results"
	"github.com/mdhender/ottomap/internal/terrain"
	"testing"
)

func TestCompassPoint(t *testing.T) {
	var pt compass.Point_e
	for _, tc := range []struct {
		id    string
		line  string
		point compass.Point_e
	}{
		{id: "N/N", line: "N/N", point: compass.North},
		{id: "N/NE", line: "N/NE", point: compass.NorthNorthEast},
		{id: "NE/NE", line: "NE/NE", point: compass.NorthEast},
		{id: "NE/SE", line: "NE/SE", point: compass.East},
		{id: "SE/SE", line: "SE/SE", point: compass.SouthEast},
		{id: "S/SE", line: "S/SE", point: compass.SouthSouthEast},
		{id: "S/S", line: "S/S", point: compass.South},
		{id: "S/SW", line: "S/SW", point: compass.SouthSouthWest},
		{id: "SW/SW", line: "SW/SW", point: compass.SouthWest},
		{id: "SW/NW", line: "SW/NW", point: compass.West},
		{id: "NW/NW", line: "NW/NW", point: compass.NorthWest},
		{id: "N/NW", line: "N/NW", point: compass.NorthNorthWest},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("COMPASSPOINT"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		point, ok := va.(compass.Point_e)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, pt, va)
			continue
		}
		if tc.point != point {
			t.Errorf("id %q: point: want %q, got %q\n", tc.id, tc.point, point)
		}
	}
}

func TestCrowsNestObservation(t *testing.T) {
	var fh parser.FarHorizon_t
	for _, tc := range []struct {
		id     string
		line   string
		point  compass.Point_e
		isLand bool
	}{
		{id: "land", line: "Sight Land - N/N", point: compass.North, isLand: true},
		{id: "water", line: "Sight Water - N/NE", point: compass.NorthNorthEast, isLand: false},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("CrowsNestObservation"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		cno, ok := va.(parser.FarHorizon_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, fh, va)
			continue
		}
		if tc.point != cno.Point {
			t.Errorf("id %q: point: want %q, got %q\n", tc.id, tc.point, cno.Point)
		}
		if tc.isLand != cno.IsLand {
			t.Errorf("id %q: terrain: want %v, got %v\n", tc.id, tc.isLand, cno.IsLand)
		}
	}
}

func TestFleetMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id        string
		line      string
		unitId    parser.UnitId_t
		moveCount int
		debug     bool
	}{
		{id: "900-06.0138f4",
			line:      `MILD NW Fleet Movement: Move NE-LCM,  Lcm NE, SE, S,\NE-LCM,  Lcm NE, SE, SW, S,\NE-LCM,  Lcm NE, SE, SW, S,\`,
			unitId:    "0138f4",
			moveCount: 3,
		},
		{id: "900-06.0138f1",
			line:      "MILD NW Fleet Movement: Move SE-O,-(NE O,  SE LCM,  N O,  S LCM,  SW O,  NW O,  )(Sight Water - N/N, Sight Land - N/NE)",
			unitId:    "0138f1",
			moveCount: 3,
		},
		{id: "900-06.0138f2",
			line:      "STRONG S Fleet Movement: Move NW-GH,",
			unitId:    "0138f2",
			moveCount: 3,
		},
		{id: "900-06.1138f2",
			line:      `MILD N Fleet Movement: Move SW-PR The Dirty Squirrel-(NE GH,  SE O, N GH, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Land - N/NW,Sight Water - NE/NE,Sight Water - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\NW-O, -(NE GH, SE PR, N SW, S O, SW O, NW O, )(Sight Water - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\NW-O, -(NE SW, SE O, N O, S O, SW O, NW O, )(Sight Water - N/N,Sight Water - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-O, -(NE O, SE SW, N O, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-O,  Lcm NE, N,-(NE LCM, SE O, N LCM, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Land - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-LCM,  Lcm NE, SE,  Ensalada sin Tomate\`,
			unitId:    "0138f2",
			moveCount: 3,
		},
	} {
		fm, err := parser.ParseFleetMovementLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug, tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.moveCount != len(fm) {
			t.Errorf("id %q: moveCount: want %d, got %d\n", tc.id, tc.moveCount, len(fm))
		}
	}
}

func TestLocationParse(t *testing.T) {
	var lt parser.Location_t
	for _, tc := range []struct {
		id      string
		line    string
		unitId  parser.UnitId_t
		msg     string
		currHex string
		prevHex string
	}{
		{id: "0138", line: "Tribe 0138, , Current Hex = ## 1108, (Previous Hex = OO 1615)", unitId: "0138", msg: "", currHex: "## 1108", prevHex: "OO 1615"},
		{id: "1138", line: "Tribe 1138, , Current Hex = ## 0709, (Previous Hex = ## 0709)", unitId: "1138", msg: "", currHex: "## 0709", prevHex: "## 0709"},
		{id: "2138", line: "Tribe 2138, , Current Hex = ## 0709, (Previous Hex = ## 0709)", unitId: "2138", msg: "", currHex: "## 0709", prevHex: "## 0709"},
		{id: "3138", line: "Tribe 3138, , Current Hex = ## 0708, (Previous Hex = ## 0708)", unitId: "3138", msg: "", currHex: "## 0708", prevHex: "## 0708"},
		{id: "4138", line: "Tribe 4138, , Current Hex = ## 0709, (Previous Hex = OO 0709)", unitId: "4138", msg: "", currHex: "## 0709", prevHex: "OO 0709"},
		{id: "0138c1", line: "Courier 0138c1, , Current Hex = ## 0709, (Previous Hex = ## 1010)", unitId: "0138c1", msg: "", currHex: "## 0709", prevHex: "## 1010"},
		{id: "0138c2", line: "Courier 0138c2, , Current Hex = ## 1103, (Previous Hex = ## 0709)", unitId: "0138c2", msg: "", currHex: "## 1103", prevHex: "## 0709"},
		{id: "0138c3", line: "Courier 0138c3, , Current Hex = ## 1103, (Previous Hex = ## 0709)", unitId: "0138c3", msg: "", currHex: "## 1103", prevHex: "## 0709"},
		{id: "0138e1", line: "Element 0138e1, , Current Hex = ## 1106, (Previous Hex = ## 2002)", unitId: "0138e1", msg: "", currHex: "## 1106", prevHex: "## 2002"},
		{id: "0138e9", line: "Element 0138e9, , Current Hex = OO 0602, (Previous Hex = OO 0302)", unitId: "0138e9", msg: "", currHex: "OO 0602", prevHex: "OO 0302"},
		{id: "1138e1", line: "Element 1138e1, , Current Hex = ## 0709, (Previous Hex = ## 1010)", unitId: "1138e1", msg: "", currHex: "## 0709", prevHex: "## 1010"},
		{id: "2138e1", line: "Element 2138e1, , Current Hex = ## 0904, (Previous Hex = ## 1507)", unitId: "2138e1", msg: "", currHex: "## 0904", prevHex: "## 1507"},
		{id: "0138f1", line: "Fleet 0138f1, , Current Hex = OO 1508, (Previous Hex = OO 1508)", unitId: "0138f1", msg: "", currHex: "OO 1508", prevHex: "OO 1508"},
		{id: "0138f3", line: "Fleet 0138f3, , Current Hex = OQ 1210, (Previous Hex = OQ 0713)", unitId: "0138f3", msg: "", currHex: "OQ 1210", prevHex: "OQ 0713"},
		{id: "0138f8", line: "Fleet 0138f8, , Current Hex = QP 1210, (Previous Hex = QP 0713)", unitId: "0138f8", msg: "", currHex: "QP 1210", prevHex: "QP 0713"},
		{id: "1138f2", line: "Fleet 1138f2, , Current Hex = RO 2415, (Previous Hex = RO 2415)", unitId: "1138f2", msg: "", currHex: "RO 2415", prevHex: "RO 2415"},
		{id: "3138g1", line: "Garrison 3138g1, , Current Hex = ## 0708, (Previous Hex = OO 0708)", unitId: "3138g1", msg: "", currHex: "## 0708", prevHex: "OO 0708"},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("Location"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		location, ok := va.(parser.Location_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, lt, va)
			continue
		}
		if tc.unitId != location.UnitId {
			t.Errorf("id %q: follows: want %q, got %q\n", tc.id, tc.unitId, location.UnitId)
		}
		if tc.msg != location.Message {
			t.Errorf("id %q: message: want %q, got %q\n", tc.id, tc.msg, location.Message)
		}
		if tc.currHex != location.CurrentHex {
			t.Errorf("id %q: currentHex: want %q, got %q\n", tc.id, tc.currHex, location.CurrentHex)
		}
		if tc.prevHex != location.PreviousHex {
			t.Errorf("id %q: previousHex: want %q, got %q\n", tc.id, tc.prevHex, location.PreviousHex)
		}
	}
}

func TestScoutMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id        string
		line      string
		unitId    parser.UnitId_t
		scoutNo   int
		moveCount int
		debug     bool
	}{
		{id: "900-05.0138e1s1", line: `Scout 1:Scout N-PR,  \N-GH,  \N-RH,  O NW,  N, Find Iron Ore, 1190,  0138c2,  0138c3\ Can't Move on Ocean to N of HEX,  Patrolled and found 1190,  0138c2,  0138c3`, unitId: "0138e1s1", scoutNo: 1},
		{id: "900-05.0138e1s2", line: `Scout 2:Scout NE-RH,  \N-PR,  \N-CH,  O NE\ Not enough M.P's to move to N into CONIFER HILLS,  Nothing of interest found`, unitId: "0138e1s2", scoutNo: 2},
		{id: "900-05.0138e1s3", line: `Scout 3:Scout SE-PR,  \SE-RH,  \SE-PR,  River S, 0190\ Not enough M.P's to move to SE into ROCKY HILLS,  Patrolled and found 0190`, unitId: "0138e1s3", scoutNo: 3},
		{id: "900-05.0138e1s4", line: `Scout 4:Scout SE-PR,  \SE-RH,  \NE-PR,  \NE-PR,  \ Not enough M.P's to move to NE into PRAIRIE,  Nothing of interest found`, unitId: "0138e1s4", scoutNo: 4},
		{id: "900-05.0138e1s5", line: `Scout 5:Scout SE-PR,  \SE-RH,  \SE-PR,  River S, 0190\N-PR,  \ Not enough M.P's to move to N into PRAIRIE,  Nothing of interest found`, unitId: "0138e1s5", scoutNo: 5},
		{id: "900-05.0138e1s6", line: `Scout 6:Scout SE-PR,  \SE-RH,  \N-PR,  \N-PR,  \ Not enough M.P's to move to N into PRAIRIE,  Nothing of interest found`, unitId: "0138e1s6", scoutNo: 6},
		{id: "900-05.0138e1s7", line: `Scout 7:Scout NW-RH,  \N-GH,  \N-PR,  O NW,  N, 3138\ Can't Move on Ocean to N of HEX,  Patrolled and found 3138`, unitId: "0138e1s7", scoutNo: 7},
		{id: "900-05.0138e1s8", line: `Scout 8:Scout SW-GH,  \NW-PR,  \NW-PR,  \NW-PR,  \ Not enough M.P's to move to NW into PRAIRIE,  Nothing of interest found`, unitId: "0138e1s8", scoutNo: 8},
	} {
		sm, err := parser.ParseScoutMovementLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug, tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.scoutNo != sm.No {
			t.Errorf("id %q: scoutNo: want %d, got %d\n", tc.id, tc.scoutNo, sm.No)
		}
		if tc.moveCount != len(sm.Moves) {
			t.Errorf("id %q: moveCount: want %d, got %d\n", tc.id, tc.moveCount, len(sm.Moves))
		}
	}
}

func TestStatusLine(t *testing.T) {
	for _, tc := range []struct {
		id      string
		line    string
		unitId  parser.UnitId_t
		terrain terrain.Terrain_e
		borders []*parser.Border_t
		units   []parser.UnitId_t
		debug   bool
	}{
		{id: "899-12.0138.0138",
			line:    `0138 Status: PRAIRIE, 0138`,
			unitId:  "0138",
			terrain: terrain.Prairie,
			units:   []parser.UnitId_t{"0138"},
		},
		{id: "900-01.0138.0138e1",
			line:    `0138e1 Status: PRAIRIE,River S, 0138e1`,
			unitId:  "0138e1",
			terrain: terrain.Prairie,
			borders: []*parser.Border_t{
				{Direction: direction.South, Edge: edges.River},
			},
			units: []parser.UnitId_t{"0138e1"},
		},
		{id: "900-02.0138.0138",
			line:    `0138 Status: PRAIRIE, O S,Ford SE, 2138, 0138`,
			unitId:  "0138",
			terrain: terrain.Prairie,
			borders: []*parser.Border_t{
				{Direction: direction.South, Terrain: terrain.Ocean},
				{Direction: direction.SouthEast, Edge: edges.Ford},
			},
			units: []parser.UnitId_t{"0138", "2138"},
		},
		{id: "900-02.0138.0138e1",
			line:    `0138e1 Status: PRAIRIE, O NW, 0138e1`,
			unitId:  "0138e1",
			terrain: terrain.Prairie,
			borders: []*parser.Border_t{
				{Direction: direction.NorthWest, Terrain: terrain.Ocean},
			},
			units: []parser.UnitId_t{"0138e1"},
		},
		{id: "900-04.0138.0138",
			line:    `0138 Status: CONIFER HILLS, O SW, NW, S, 2138, 0138c1, 0138, 1138`,
			unitId:  "0138",
			terrain: terrain.ConiferHills,
			borders: []*parser.Border_t{
				{Direction: direction.NorthWest, Terrain: terrain.Ocean},
				{Direction: direction.South, Terrain: terrain.Ocean},
				{Direction: direction.SouthWest, Terrain: terrain.Ocean},
			},
			units: []parser.UnitId_t{"0138", "0138c1", "1138", "2138"},
		},
	} {
		sl, err := parser.ParseStatusLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug, tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if len(sl) != 1 {
			t.Errorf("id %q: moveCount: want 1, got %d\n", tc.id, len(sl))
			if len(sl) == 0 {
				continue
			}
		}
		move := sl[0]
		if results.StayedInPlace != move.Result {
			t.Errorf("id %q: result: want %q, got %q\n", tc.id, results.StayedInPlace, move.Result)
		}
		report := move.Report
		if tc.terrain != report.Terrain {
			t.Errorf("id %q: terrain: want %q, got %q\n", tc.id, tc.terrain, report.Terrain)
		}
		// borders
		if len(tc.borders) != len(report.Borders) {
			t.Errorf("id %q: borders: want %d, got %d\n", tc.id, len(tc.borders), len(report.Borders))
		}
		b1, b2 := diffBorderSets(tc.borders, report.Borders)
		for _, b := range b1 {
			t.Errorf("id %q: borders: want %q, got nil\n", tc.id, b)
		}
		for _, b := range b2 {
			t.Errorf("id %q: borders: want nil, got %q\n", tc.id, b)
		}
		// encounters
		if len(tc.units) != len(report.Encounters) {
			t.Errorf("id %q: units: want %d, got %d\n", tc.id, len(tc.units), len(report.Encounters))
		}
		u1, u2 := diffUnitSets(tc.units, report.Encounters)
		for _, u := range u1 {
			t.Errorf("id %q: units: want %q, got nil\n", tc.id, u)
		}
		for _, u := range u2 {
			t.Errorf("id %q: units: want nil, got %q\n", tc.id, u)
		}
	}
}

func TestTribeFollowsParse(t *testing.T) {
	for _, tc := range []struct {
		id      string
		line    string
		unitId  parser.UnitId_t
		follows parser.UnitId_t
		debug   bool
	}{
		{id: "1812", line: "Tribe Follows 1812", follows: "1812"},
		{id: "1812f3", line: "Tribe Follows 1812f3", follows: "1812f3"},
	} {
		tf, err := parser.ParseTribeFollowsLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.follows != tf.Follows {
			t.Errorf("id %q: follows: want %q, got %q\n", tc.id, tc.follows, tf.Follows)
		}
	}
}

func TestTribeGoesParse(t *testing.T) {
	for _, tc := range []struct {
		id     string
		line   string
		unitId parser.UnitId_t
		goesTo string
		debug  bool
	}{
		{id: "1", line: "Tribe Goes to DT 1812", goesTo: "DT 1812"},
		{id: "2", line: "Tribe Goes to ## 1812", goesTo: "## 1812"},
		{id: "3", line: "Tribe Goes to N/A", goesTo: "N/A"},
	} {
		gt, err := parser.ParseTribeGoesToLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.goesTo != gt.GoesTo {
			t.Errorf("id %q: goesTo: want %q, got %q\n", tc.id, tc.goesTo, gt.GoesTo)
		}
	}
}

func TestTribeMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id        string
		line      string
		unitId    parser.UnitId_t
		moveCount int
		debug     bool
	}{
		{id: "1",
			line: `Tribe Movement: Move \`,
		},
		{id: "2",
			line: "Tribe Movement: Move NW-GH,",
		},
		{id: "3",
			line: `Tribe Movement: Move SW-PR The Dirty Squirrel\N-LCM,  Lcm NE, SE,  Ensalada sin Tomate\`,
		},
	} {
		tm, err := parser.ParseTribeMovementLine(tc.id, tc.unitId, 1, []byte(tc.line), tc.debug, tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.moveCount != len(tm) {
			t.Errorf("id %q: moveCount: want %d, got %d\n", tc.id, tc.moveCount, len(tm))
		}
	}
}

func TestTurnInfoParse(t *testing.T) {
	var ti parser.TurnInfo_t
	for _, tc := range []struct {
		id        string
		line      string
		thisYear  int
		thisMonth int
		nextYear  int
		nextMonth int
	}{
		{id: "900-01", line: "Current Turn 900-01 (#1), Spring, FINE\tNext Turn 900-02 (#2), 12/11/2023", thisYear: 900, thisMonth: 1, nextYear: 900, nextMonth: 2},
		{id: "900-02", line: "Current Turn 900-02 (#2), Spring, FINE", thisYear: 900, thisMonth: 2},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("TurnInfo"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		turnInfo, ok := va.(parser.TurnInfo_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, ti, va)
			continue
		}
		if tc.thisYear != turnInfo.CurrentTurn.Year {
			t.Errorf("id %q: thisYear: want %d, got %d\n", tc.id, tc.thisYear, turnInfo.CurrentTurn.Year)
		}
		if tc.thisMonth != turnInfo.CurrentTurn.Month {
			t.Errorf("id %q: thisMonth: want %d, got %d\n", tc.id, tc.thisMonth, turnInfo.CurrentTurn.Month)
		}
		if tc.nextYear == 0 && tc.nextMonth == 0 && !turnInfo.NextTurn.IsZero() {
			t.Errorf("id %q: nextTurn: want %v, got %v\n", tc.id, parser.Date_t{}, turnInfo.NextTurn)
		} else {
			if tc.nextYear != turnInfo.NextTurn.Year {
				t.Errorf("id %q: nextYear: want %d, got %d\n", tc.id, tc.nextYear, turnInfo.NextTurn.Year)
			}
			if tc.nextMonth != turnInfo.NextTurn.Month {
				t.Errorf("id %q: nextMonth: want %d, got %d\n", tc.id, tc.nextMonth, turnInfo.NextTurn.Month)
			}
		}
	}
}

func diffBorderSets(set1, set2 []*parser.Border_t) ([]*parser.Border_t, []*parser.Border_t) {
	map1 := map[string]*parser.Border_t{}
	map2 := map[string]*parser.Border_t{}

	// Fill map1 with elements from set1
	for _, i := range set1 {
		map1[i.String()] = i
	}

	// Fill map2 with elements from set2
	for _, i := range set2 {
		map2[i.String()] = i
	}

	// Find elements in set1 but not in set2
	var onlyInSet1 []*parser.Border_t
	for k, v := range map1 {
		if _, found := map2[k]; !found {
			onlyInSet1 = append(onlyInSet1, v)
		}
	}

	// Find elements in set2 but not in set1
	var onlyInSet2 []*parser.Border_t
	for k, v := range map2 {
		if _, found := map1[k]; !found {
			onlyInSet2 = append(onlyInSet2, v)
		}
	}

	return onlyInSet1, onlyInSet2
}

func diffUnitSets(set1, set2 []parser.UnitId_t) ([]parser.UnitId_t, []parser.UnitId_t) {
	// Create maps to store the presence of UnitIDs in each slice
	map1 := map[parser.UnitId_t]bool{}
	map2 := map[parser.UnitId_t]bool{}

	// Fill map1 with elements from set1
	for _, id := range set1 {
		map1[id] = true
	}

	// Fill map2 with elements from set2
	for _, id := range set2 {
		map2[id] = true
	}

	// Find elements in set1 but not in set2
	var onlyInSet1 []parser.UnitId_t
	for id := range map1 {
		if _, found := map2[id]; !found {
			onlyInSet1 = append(onlyInSet1, id)
		}
	}

	// Find elements in set2 but not in set1
	var onlyInSet2 []parser.UnitId_t
	for id := range map2 {
		if _, found := map1[id]; !found {
			onlyInSet2 = append(onlyInSet2, id)
		}
	}

	return onlyInSet1, onlyInSet2
}
