// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var (
	rxCourierSection  = regexp.MustCompile(`^Courier \d{4}c\d, ,`)
	rxElementSection  = regexp.MustCompile(`^Element \d{4}e\d, ,`)
	rxFleetSection    = regexp.MustCompile(`^Fleet \d{4}f\d, ,`)
	rxGarrisonSection = regexp.MustCompile(`^Garrison \d{4}g\d, ,`)
	rxScoutLine       = regexp.MustCompile(`^Scout \d:Scout `)
	rxTribeSection    = regexp.MustCompile(`^Tribe \d{4}, ,`)
)

// ImportReport imports a report from a file into the database.
func ImportReport(db *sqlc.DB, path string, debug bool) (id int, err error) {
	debugf := func(format string, args ...any) {
		if debug {
			log.Printf(format, args...)
		}
	}

	debugf("importReport: %s\n", path)

	// split the file name into parts
	importPath, importName := filepath.Split(path)
	if importPath == "" {
		importPath = "."
	}
	debugf("importReport: %s: %s\n", importPath, importName)

	// verify that the report file exists.
	if sb, err := os.Stat(path); err != nil {
		debugf("importReport: %s: %v\n", path, err)
		return 0, err
	} else if sb.IsDir() {
		debugf("importReport: %s: %v\n", path, cerrs.ErrNotAFile)
		return 0, cerrs.ErrNotAFile
	} else if !sb.Mode().IsRegular() {
		debugf("importReport: %s: %v\n", path, cerrs.ErrNotAFile)
		return 0, cerrs.ErrNotAFile
	}

	// load the report file into memory
	var data []byte
	if data, err = os.ReadFile(path); err != nil {
		return 0, err
	}
	debugf("importReport: %s: loaded %d bytes\n", importName, len(data))
	if len(data) < 128 || !bytes.HasPrefix(data, []byte("Tribe 0")) {
		return 0, cerrs.ErrNotATurnReport
	}

	// calculate the checksum of the data.
	cksum := fmt.Sprintf("%x", sha256.Sum256(data))
	debugf("importReport: %s: cksum %s\n", importName, cksum)

	// return an error if there are records with the same checksum.
	// the caller will need to delete the file and try again.
	duplicateRows, err := db.Queries.ReadInputMetadataByChecksum(db.Ctx, cksum)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// there was some error other than no rows found
			return 0, errors.Join(fmt.Errorf("read input metadata by checksum"), err)
		}
	}
	if len(duplicateRows) != 0 {
		// rows with this checksum exist, so we cannot insert this report
		for _, dup := range duplicateRows {
			debugf("importInput: duplicate %s\n", dup.Name)
		}
		return 0, cerrs.ErrDuplicateChecksum
	}

	// section off the input
	type sline struct {
		no   int
		line string
	}
	var sections [][]sline
	var section []sline
	var elementId []byte
	var elementStatusPrefix []byte
	for n, line := range bytes.Split(data, []byte{'\n'}) {
		no := n + 1
		if rxCourierSection.Match(line) {
			elementId = line[8 : 8+6]
			debugf("importReport: %5d: found %q %q\n", no, line[:14], elementId)
			if section != nil {
				sections = append(sections, section)
			}
			section = []sline{sline{no: no, line: string(line)}}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxElementSection.Match(line) {
			elementId = line[8 : 8+6]
			debugf("importReport: %5d: found %q %q\n", no, line[:14], elementId)
			if section != nil {
				sections = append(sections, section)
			}
			section = []sline{sline{no: no, line: string(line)}}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxFleetSection.Match(line) {
			elementId = line[6 : 6+6]
			debugf("importReport: %5d: found %q %q\n", no, line[:12], elementId)
			if section != nil {
				sections = append(sections, section)
			}
			section = []sline{sline{no: no, line: string(line)}}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxGarrisonSection.Match(line) {
			elementId = line[9 : 9+6]
			debugf("importReport: %5d: found %q %q\n", no, line[:15], elementId)
			if section != nil {
				sections = append(sections, section)
			}
			section = []sline{sline{no: no, line: string(line)}}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if rxTribeSection.Match(line) {
			elementId = line[6 : 6+4]
			debugf("importReport: %5d: found %q %q\n", no, line[:10], elementId)
			if section != nil {
				sections = append(sections, section)
			}
			section = []sline{sline{no: no, line: string(line)}}
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", string(elementId)))
		} else if section == nil {
			// ignore
		} else if len(section) == 1 && bytes.HasPrefix(line, []byte("Current Turn ")) {
			debugf("importReport: %5d: found %q\n", no, line[:12])
			section = append(section, sline{no: no, line: string(line)})
		} else if bytes.HasPrefix(line, []byte("Tribe Follows: ")) {
			debugf("importReport: %5d: found %q\n", no, line[:13])
			section = append(section, sline{no: no, line: string(line)})
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			debugf("importReport: %5d: found %q\n", no, line[:14])
			section = append(section, sline{no: no, line: string(line)})
		} else if rxScoutLine.Match(line) {
			debugf("importReport: %5d: found %q\n", no, line[:14])
			section = append(section, sline{no: no, line: string(line)})
		} else if bytes.HasPrefix(line, elementStatusPrefix) {
			debugf("importReport: %5d: found %q\n", no, line[:len(elementStatusPrefix)])
			section = append(section, sline{no: no, line: string(line)})
		}
	}
	if len(section) != 0 {
		sections = append(sections, section)
	}
	for n, s := range sections {
		for _, ss := range s {
			if len(ss.line) < 55 {
				debugf("section %3d: line %5d: %s\n", n+1, ss.no, ss.line)
			} else {
				debugf("section %3d: line %5d: %s...\n", n+1, ss.no, ss.line[:55])
			}
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	qtx := db.Queries.WithTx(tx)

	if n, err := qtx.InsertInput(db.Ctx, sqlc.InsertInputParams{
		Path:  importPath,
		Name:  importName,
		Cksum: cksum,
	}); err != nil {
		return 0, errors.Join(fmt.Errorf("create input"), err)
	} else {
		id = int(n)
	}
	debugf("importInput: %s: %s\n", importName, cksum)

	for n, s := range sections {
		sectNo := n + 1
		for _, ss := range s {
			if err = qtx.InsertInputLine(db.Ctx, sqlc.InsertInputLineParams{
				Iid:    int64(id),
				SectNo: int64(sectNo),
				LineNo: int64(ss.no),
				Line:   ss.line,
			}); err != nil {
				return 0, errors.Join(fmt.Errorf("insert input lines"), err)
			}
		}
	}

	debugf("importInput: %s: lines: %d\n", importName, len(bytes.Split(data, []byte{'\n'})))

	if err = tx.Commit(); err != nil {
		debugf("sqlc: importInput: commit: %v\n", err)
		return id, err
	}

	// return no errors
	return id, nil
}
