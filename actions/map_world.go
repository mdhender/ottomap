// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/wxx"
	"log"
)

type MapConfig struct {
	Dump struct {
		All bool
	}
	Show struct {
		Origin string // if set, flag this location as the "origin"
	}
}

func MapWorld(reports []*parser.Report_t, cfg MapConfig) (*wxx.WXX, error) {
	if len(reports) == 0 {
		log.Fatalf("error: no reports to map\n")
	}
	log.Printf("map: collected %8d hexes\n", len(reports))

	log.Printf("hey, resources disabled\n")
	log.Printf("hey, borders   disabled\n")

	consolidatedMap := &wxx.WXX{}

	// create the grids within the existing bounds
	minGrid, maxGrid := FindBounds(reports)
	log.Printf("map: upper left  grid %s\n", minGrid.GridId())
	log.Printf("map: lower right grid %s\n", maxGrid.GridId())
	log.Printf("map: todo: move grid creation from merge to here\n")

	// world hex map is indexed by grid location
	worldHexMap := map[string]*wxx.Hex{}
	for _, report := range reports {
		gridCoords := report.Location.GridString()
		gridColumn, gridRow := report.Location.GridColumnRow()
		hex := &wxx.Hex{
			Location: report.Location,
			Offset: wxx.Offset{
				Column: gridColumn,
				Row:    gridRow,
			},
			Terrain: report.Terrain,
			Features: wxx.Features{
				IsOrigin: cfg.Show.Origin == gridCoords,
				//Resources: report.Resources,
			},
			WasScouted: report.ScoutedTurnId != "",
		}
		worldHexMap[hex.Location.GridString()] = hex

		//for _, border := range report.Borders {
		//	switch border.Edge {
		//	case edges.None:
		//	case edges.Ford:
		//		hex.Features.Edges.Ford = append(hex.Features.Edges.Ford, border.Direction)
		//	case edges.Pass:
		//		hex.Features.Edges.Pass = append(hex.Features.Edges.Pass, border.Direction)
		//	case edges.River:
		//		hex.Features.Edges.River = append(hex.Features.Edges.River, border.Direction)
		//	case edges.StoneRoad:
		//		hex.Features.Edges.StoneRoad = append(hex.Features.Edges.StoneRoad, border.Direction)
		//	default:
		//		panic(fmt.Sprintf("assert(edge != %d)", border.Edge))
		//	}
		//}

		if err := consolidatedMap.MergeHex(report.TurnId, hex); err != nil {
			log.Fatalf("error: wxx: mergeHexes: newHexes: %v\n", err)
		}
	}

	log.Printf("map: collected %8d new     hexes\n", len(worldHexMap))

	return consolidatedMap, nil
}

func FindBounds(reports []*parser.Report_t) (minLocation, maxLocation coords.Map) {
	if len(reports) == 0 {
		return coords.Map{}, coords.Map{}
	}

	minColumn, minRow := reports[0].Location.Column, reports[0].Location.Row
	maxColumn, maxRow := reports[0].Location.Column, reports[0].Location.Row

	for _, report := range reports {
		if report.Location.Column < minColumn {
			minColumn = report.Location.Column
		}
		if report.Location.Row < minRow {
			minRow = report.Location.Row
		}
		if maxColumn < report.Location.Column {
			maxColumn = report.Location.Column
		}
		if maxRow < report.Location.Row {
			maxRow = report.Location.Row
		}
	}

	return coords.ColumnRowToMap(minColumn, minRow), coords.ColumnRowToMap(maxColumn, maxRow)
}
