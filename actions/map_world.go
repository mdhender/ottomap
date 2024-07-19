// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"fmt"
	"github.com/mdhender/ottomap/internal/coords"
	"github.com/mdhender/ottomap/internal/direction"
	"github.com/mdhender/ottomap/internal/edges"
	"github.com/mdhender/ottomap/internal/tiles"
	"github.com/mdhender/ottomap/internal/wxx"
	"log"
)

type MapConfig struct {
	Dump struct {
		All          bool
		BorderCounts bool
	}
	Origin coords.Map
	Show   struct {
		Origin bool // if set, put a marker in the origin hex
	}
}

func MapWorld(allTiles *tiles.Map_t, cfg MapConfig) (*wxx.WXX, error) {
	if allTiles.Length() == 0 {
		log.Fatalf("error: no tiles to map\n")
	}
	log.Printf("map: collected %8d tiles\n", allTiles.Length())

	if cfg.Dump.BorderCounts {
		panic("border counts not implemented")
		//for _, report := range allTiles {
		//	gridCoords := report.Location.GridString()
		//	gridColumn, gridRow := report.Location.GridColumnRow()
		//	log.Printf("%s: %4d %4d: %6d\n", gridCoords, gridColumn, gridRow, len(report.Borders))
		//}
	}

	consolidatedMap := wxx.NewWXX()

	// create an offset that will shift the map to about 4 hexes from the upper left.
	var renderOffset coords.Map
	upperLeft, lowerRight := allTiles.Bounds()
	log.Printf("map: upper left  grid %s\n", upperLeft.GridString())
	log.Printf("map: lower right grid %s\n", lowerRight.GridString())
	log.Printf("map: todo: move grid creation from merge to here\n")
	if upperLeft.Column > 4 {
		renderOffset.Column = upperLeft.Column - 4
	}
	if upperLeft.Row > 4 {
		renderOffset.Row = upperLeft.Row - 4
	}

	// world hex map is indexed by render location, not true location
	worldHexMap := map[coords.Map]*wxx.Hex{}
	for _, t := range allTiles.Tiles {
		hex := &wxx.Hex{
			Location: t.Location,
			RenderAt: coords.Map{
				Column: t.Location.Column - renderOffset.Column,
				Row:    t.Location.Row - renderOffset.Row,
			},
			Terrain: t.Terrain,
			Features: wxx.Features{
				IsOrigin: cfg.Show.Origin && t.Location == cfg.Origin,
				//Resources: report.Resources,
			},
			WasVisited: t.Visited != "",
			WasScouted: t.Scouted != "",
		}

		// todo: one way fords and one way passes?
		for _, d := range direction.Directions {
			for _, edge := range t.Edges[d] {
				switch edge {
				case edges.None:
				case edges.Ford:
					hex.Features.Edges.Ford = append(hex.Features.Edges.Ford, d)
				case edges.Pass:
					hex.Features.Edges.Pass = append(hex.Features.Edges.Pass, d)
				case edges.River:
					hex.Features.Edges.River = append(hex.Features.Edges.River, d)
				case edges.StoneRoad:
					hex.Features.Edges.StoneRoad = append(hex.Features.Edges.StoneRoad, d)
				default:
					panic(fmt.Sprintf("assert(edge != %d)", edge))
				}
			}
		}

		for _, resource := range t.Resources {
			hex.Features.Resources = append(hex.Features.Resources, resource)
		}

		for _, settlement := range t.Settlements {
			hex.Features.Settlements = append(hex.Features.Settlements, settlement)
		}

		worldHexMap[hex.RenderAt] = hex

		if err := consolidatedMap.MergeHex(hex); err != nil {
			log.Fatalf("error: wxx: mergeHexes: newHexes: %v\n", err)
		}
	}

	//for _, report := range allTiles {
	//	gridCoords := report.Location.GridString()
	//	gridColumn, gridRow := report.Location.GridColumnRow()
	//	hex := &wxx.Hex{
	//		Location: report.Location,
	//		Offset: wxx.Offset{
	//			Column: gridColumn,
	//			Row:    gridRow,
	//		},
	//		Terrain: report.Terrain,
	//		Features: wxx.Features{
	//			IsOrigin: cfg.Show.Origin == gridCoords,
	//			//Resources: report.Resources,
	//		},
	//		WasScouted: report.ScoutedTurnId != "",
	//	}
	//	for _, settlement := range report.Settlements {
	//		hex.Features.Settlements = append(hex.Features.Settlements, settlement)
	//	}
	//	worldHexMap[hex.Location.GridString()] = hex
	//
	//	//for _, border := range report.Borders {
	//	//	switch border.Edge {
	//	//	case edges.None:
	//	//	case edges.Ford:
	//	//		hex.Features.Edges.Ford = append(hex.Features.Edges.Ford, border.Direction)
	//	//	case edges.Pass:
	//	//		hex.Features.Edges.Pass = append(hex.Features.Edges.Pass, border.Direction)
	//	//	case edges.River:
	//	//		hex.Features.Edges.River = append(hex.Features.Edges.River, border.Direction)
	//	//	case edges.StoneRoad:
	//	//		hex.Features.Edges.StoneRoad = append(hex.Features.Edges.StoneRoad, border.Direction)
	//	//	default:
	//	//		panic(fmt.Sprintf("assert(edge != %d)", border.Edge))
	//	//	}
	//	//}
	//
	//	if err := consolidatedMap.MergeHex(report.TurnId, hex); err != nil {
	//		log.Fatalf("error: wxx: mergeHexes: newHexes: %v\n", err)
	//	}
	//}

	log.Printf("map: collected %8d new     hexes\n", len(worldHexMap))

	return consolidatedMap, nil
}
