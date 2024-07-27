// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package ffs implements a file-based flat file system.
package ffs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

func New(path string) (*FFS, error) {
	return &FFS{
		path:          path,
		rxTurnReports: regexp.MustCompile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.report\.txt`),
	}, nil
}

type FFS struct {
	path          string
	rxTurnReports *regexp.Regexp
}

type Turn_t struct {
	Id string
}

// GetTurnListing scan the data path for turn reports and adds them to the list
func (f *FFS) GetTurnListing(id string) (list []Turn_t, err error) {
	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getTurnListing: %v\n", err)
		return nil, nil
	}

	// add all turn reports to the list
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := f.rxTurnReports.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		list = append(list, Turn_t{Id: matches[1]})
	}

	// sort the list, not sure why.
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})

	return list, nil
}

type TurnDetail_t struct {
	Id    string
	Clans []string
}

func (f *FFS) GetTurnDetails(id string, turnId string) (row TurnDetail_t, err error) {
	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getTurnDetails: %v\n", err)
		return row, nil
	}

	row.Id = turnId

	// find all turn reports for this turn and collect the clan names.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := f.rxTurnReports.FindStringSubmatch(entry.Name())
		if len(matches) != 3 || matches[1] != turnId {
			continue
		}
		row.Clans = append(row.Clans, matches[2])
	}

	// sort the list, not sure why.
	sort.Slice(row.Clans, func(i, j int) bool {
		return row.Clans[i] < row.Clans[j]
	})

	return row, nil
}

type TurnSection_t struct {
	Id          string
	Clan        string
	Unit        string
	CurrentHex  string
	PreviousHex string
}

func (f *FFS) GetTurnSections(id string, turnId, clanId string) (sections []TurnSection_t, err error) {
	rxCourierSection := regexp.MustCompile(`^Courier (\d{4}c)\d, `)
	rxElementSection := regexp.MustCompile(`^Element (\d{4}e)\d, `)
	rxFleetSection := regexp.MustCompile(`^Fleet (\d{4}f)\d, `)
	rxGarrisonSection := regexp.MustCompile(`^Garrison (\d{4}g)\d, `)
	rxTribeSection := regexp.MustCompile(`^Tribe (\d{4}), `)

	turnReportFile := filepath.Join(f.path, id, fmt.Sprintf("%s.%s.report.txt", turnId, clanId))
	if data, err := os.ReadFile(turnReportFile); err != nil {
		log.Printf("getTurnSections: %s: %v\n", turnReportFile, err)
	} else {
		for _, line := range bytes.Split(data, []byte("\n")) {
			if matches := rxCourierSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				sections = append(sections, TurnSection_t{
					Id:          turnId,
					Clan:        clanId,
					Unit:        matches[1],
					CurrentHex:  "",
					PreviousHex: "",
				})
			} else if matches = rxElementSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				sections = append(sections, TurnSection_t{
					Id:          turnId,
					Clan:        clanId,
					Unit:        matches[1],
					CurrentHex:  "",
					PreviousHex: "",
				})
			} else if matches = rxFleetSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				sections = append(sections, TurnSection_t{
					Id:          turnId,
					Clan:        clanId,
					Unit:        matches[1],
					CurrentHex:  "",
					PreviousHex: "",
				})
			} else if matches = rxGarrisonSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				sections = append(sections, TurnSection_t{
					Id:          turnId,
					Clan:        clanId,
					Unit:        matches[1],
					CurrentHex:  "",
					PreviousHex: "",
				})
			} else if matches = rxTribeSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				sections = append(sections, TurnSection_t{
					Id:          turnId,
					Clan:        clanId,
					Unit:        matches[1],
					CurrentHex:  "",
					PreviousHex: "",
				})
			}
		}
	}

	// sort the list, not sure why.
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Unit < sections[j].Unit
	})

	return sections, nil
}
