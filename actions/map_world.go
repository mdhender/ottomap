// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/parser"
	"github.com/mdhender/ottomap/internal/wxx"
	"log"
	"path/filepath"
)

type MapConfig struct {
	ClanId string
	Path   string
	Dump   struct {
		All bool
	}
	Show struct {
		Grid struct {
			Centers bool
			Coords  bool
			Numbers bool
		}
	}
}

func MapWorld(reports []*parser.Report_t, cfg MapConfig) error {
	if len(reports) == 0 {
		log.Fatalf("error: no reports to map\n")
	}
	log.Printf("map: collected %8d hexes\n", len(reports))

	// log the map boundaries?
	minGrid, maxGrid := FindBounds(reports)
	log.Printf("map: upper left  grid %s\n", minGrid.GridId())
	log.Printf("map: lower right grid %s\n", maxGrid.GridId())

	consolidatedMap := &wxx.WXX{}

	// world hex map is indexed by grid location
	worldHexMap := map[string]*wxx.Hex{}
	for _, report := range reports {
		gridColumn, gridRow := report.Location.GridColumnRow()
		hex := &wxx.Hex{
			Location: report.Location,
			Offset: wxx.Offset{
				Column: gridColumn,
				Row:    gridRow,
			},
			Terrain: report.Terrain,
			Features: wxx.Features{
				Created: report.TurnId,
				//Resources: report.Resources,
			},
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

	if cfg.Show.Grid.Coords {
		consolidatedMap.AddGridCoords()
	} else if cfg.Show.Grid.Numbers {
		consolidatedMap.AddGridNumbering()
	}

	// now we can create the Worldographer map!
	mapName := filepath.Join(cfg.Path, fmt.Sprintf("%s.wxx", cfg.ClanId))
	if err := consolidatedMap.Create(mapName, cfg.Show.Grid.Centers); err != nil {
		log.Printf("creating %s\n", mapName)
		log.Fatalf("error: %v\n", err)
	}
	log.Printf("created  %s\n", mapName)

	return nil
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
