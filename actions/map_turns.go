// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"github.com/mdhender/ottomap/internal/parser"
	"log"
)

type MapConfig struct {
	Clan string
	Path string
	Dump struct {
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

	//consolidatedMap := &wxx.WXX{}

	log.Printf("map: collected %d hexes\n", len(reports))

	return nil
}
