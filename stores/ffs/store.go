// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package ffs implements a file-based flat file system.
package ffs

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

func New(path string) (*FFS, error) {
	return &FFS{
		path: path,
		rx:   regexp.MustCompile(`^([0-9]{4}-[0-9]{2})\.[0-9]{4}\.report\.txt`),
	}, nil
}

type FFS struct {
	path string
	rx   *regexp.Regexp
}

type Turn_t struct {
	Id string
}

// GetAllTurns scan the data path for turn reports and adds them to the list
func (f *FFS) GetAllTurns(id string) (list []Turn_t) {
	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getAllTurns: %v\n", err)
		return nil
	}

	// add all turn reports to the list
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := f.rx.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		list = append(list, Turn_t{Id: matches[1]})
	}

	// sort the list, not sure why.
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})

	return list
}
