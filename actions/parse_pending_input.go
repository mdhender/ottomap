// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"database/sql"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/lbmoves"
	ploc "github.com/mdhender/ottomap/parsers/locations"
	pturn "github.com/mdhender/ottomap/parsers/turns"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"log"
	"strings"
)

// ParsePendingInput reads a pending input and parses it into sections.
// If there are errors parsing, the status of the input is set to "error"
// and an error message is created and added to the log.
// If there were no errors, a report header is created with status of "pending"
// and the sections are inserted into the report sections table.
// Any errors with the insert cause the status of the input to "error" and
// the log is updated.
// If everything worked, the status is set to "complete" and the transaction
// is committed.
func ParsePendingInput(db *sqlc.DB, pendingRow sqlc.ReadPendingInputMetadataRow) (err error) {
	// this could be a race condition
	log.Printf("parsePending: name %s: id %d\n", pendingRow.Name, pendingRow.ID)

	dbs, err := db.OpenSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = dbs.Close(nil)
	}()
	q := dbs.Queries

	if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
		Arg1:    "parsePending",
		Arg2:    pendingRow.Name,
		Arg3:    fmt.Sprintf("input %d", pendingRow.ID),
		Message: "starting parse",
	}); err != nil {
		log.Printf("parsePending: name %s: id %5d: clm %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Abort(err)
	}

	if err = q.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
		ID:       pendingRow.ID,
		Status:   "pending",
		Status_2: "parsing",
	}); err != nil {
		log.Printf("parsePending: name %s: id %5d: uis %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Abort(err)
	}
	log.Printf("parsePending: name %s: id %5d: set status to 'parsing'\n", pendingRow.Name, pendingRow.ID)

	sectionRows, err := q.ReadInputSections(db.Ctx, pendingRow.ID)
	if err != nil {
		log.Printf("parsePending: name %s: id %5d: readInputSections: %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Abort(err)
	} else if len(sectionRows) == 0 {
		log.Printf("parsePending: name %s: id %5d: readInputSections: %d sections\n", pendingRow.Name, pendingRow.ID, len(sectionRows))
		if err = q.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
			ID:       pendingRow.ID,
			Status:   "parsing",
			Status_2: "completed",
		}); err != nil {
			log.Printf("parsePending: name %s: id %5d: updatingStatus: %v\n", pendingRow.Name, pendingRow.ID, err)
			return dbs.Abort(err)
		}
		return sql.ErrNoRows
	}
	log.Printf("parsePending: name %s: id %5d: sections %5d\n", pendingRow.Name, pendingRow.ID, len(sectionRows))

	//var reportClan, reportTurn string
	var currentClan string
	var currentTurn *pturn.TurnInfo
	errorCount := 0
	for _, sectionRow := range sectionRows {
		log.Printf("parsePending: name %s: id %5d: section %6d/%5d\n", pendingRow.Name, sectionRow.ID, sectionRow.SectNo, len(sectionRows))
		rows, err := q.ReadInputLines(db.Ctx, sqlc.ReadInputLinesParams{
			ID:     sectionRow.ID,
			SectNo: sectionRow.SectNo,
		})
		if err != nil {
			log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
			return dbs.Abort(err)
		}
		log.Printf("parsePending: name %s: id %5d: section %d: rows %d\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, len(rows))

		// input must contain at least the element, turn, and status lines
		if len(rows) < 3 {
			errorCount++
			log.Printf("parsePending: name %s: id %5d: section %d: rows %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, len(rows), cerrs.ErrNotATurnReport)
			if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
				Arg1:    "parsePending",
				Arg2:    fmt.Sprintf("input %d", sectionRow.ID),
				Arg3:    fmt.Sprintf("section %d", sectionRow.SectNo),
				Message: "missing element, turn, movement, or status lines",
			}); err != nil {
				log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
				return dbs.Abort(err)
			}
			continue
		}

		var ul *ploc.Location
		var ok bool

		// first row is the element header. parse the element and location
		if v, err := ploc.Parse("location", []byte(rows[0].Line)); err != nil {
			log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
			log.Printf("parsePending: name %s: parsing error: line %d\n", pendingRow.Name, rows[0].LineNo)
			log.Printf("parsePending: input: %q\n", rows[0].Line)
			log.Printf("parsePending: error: %v\n", err)
			if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
				Arg1:    "parsePending",
				Arg2:    fmt.Sprintf("input %d", sectionRow.ID),
				Arg3:    fmt.Sprintf("line %d", rows[0].LineNo),
				Message: fmt.Sprintf("parse: %v", err),
			}); err != nil {
				log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
				return dbs.Abort(err)
			}
			errorCount++
			continue
		} else if ul, ok = v.(*ploc.Location); !ok {
			panic(fmt.Sprintf("expected *locations.Location, got %T", v))
		}
		if currentClan == "" {
			currentClan = ul.UnitId
		}
		log.Printf("parsePending: name %s: unit %s (%s %s)\n", pendingRow.Name, ul.UnitId, ul.PrevCoords, ul.CurrCoords)

		// verify that this section belongs to the same clan
		if ul.UnitId[1:4] != currentClan[1:4] {
			log.Printf("parsePending: name %s: id %5d: section %d: element %q: not the same clan as section 1: %q\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, ul.UnitId, currentClan)
			if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
				Arg1:    "parsePending",
				Arg2:    fmt.Sprintf("input %d", sectionRow.ID),
				Arg3:    fmt.Sprintf("line %d", rows[1].LineNo),
				Message: "not the same turn as section 1",
			}); err != nil {
				log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
				return dbs.Abort(err)
			}
			errorCount++
			continue
		}
		elementStatusPrefix := fmt.Sprintf("%s Status: ", ul.UnitId)

		// second row must be the turn header
		var ti *pturn.TurnInfo
		if v, err := pturn.Parse("turnInfo", []byte(rows[1].Line)); err != nil {
			log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
			log.Printf("parsePending: name %s: parsing error: line %d\n", pendingRow.Name, rows[1].LineNo)
			log.Printf("parsePending: input: %q\n", rows[1].Line)
			log.Printf("parsePending: error: %v\n", err)
			if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
				Arg1:    "parsePending",
				Arg2:    fmt.Sprintf("input %d", sectionRow.ID),
				Arg3:    fmt.Sprintf("line %d", rows[1].LineNo),
				Message: fmt.Sprintf("parse: %v", err),
			}); err != nil {
				log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
				return dbs.Abort(err)
			}
			errorCount++
			continue
		} else if ti, ok = v.(*pturn.TurnInfo); !ok {
			panic(fmt.Sprintf("expected *turns.TurnInfo, got %T", v))
		}
		log.Printf("parsePending: name %s: unit %s: turn %04d-%02d\n", pendingRow.Name, ul.UnitId, ti.TurnDate.Year, ti.TurnDate.Month)
		if currentTurn == nil {
			currentTurn = ti
		}

		// verify that this section belongs to the same turn
		if ti.TurnDate.Year != currentTurn.TurnDate.Year || ti.TurnDate.Month != currentTurn.TurnDate.Month {
			log.Printf("parsePending: name %s: id %5d: section %d: turn %04d-%02d: not the same turn as section 1: %04d-%02d\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, ti.TurnDate.Year, ti.TurnDate.Month, currentTurn.TurnDate.Year, currentTurn.TurnDate.Month)
			if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
				Arg1:    "parsePending",
				Arg2:    fmt.Sprintf("input %d", sectionRow.ID),
				Arg3:    fmt.Sprintf("line %d", rows[1].LineNo),
				Message: "not the same turn as section 1",
			}); err != nil {
				log.Printf("parsePending: name %s: id %5d: section %d: %v\n", pendingRow.Name, pendingRow.ID, sectionRow.SectNo, err)
				return dbs.Abort(err)
			}
			errorCount++
			continue
		}
		turnId := fmt.Sprintf("%04d-%02d", ti.TurnDate.Year, ti.TurnDate.Month)

		showSteps := true
		// remaining rows are movement or status lines; process them
		for n, row := range rows {
			if n == 0 || n == 1 { // already processed these rows
				continue
			}
			slug := row.Line
			if len(slug) > 18 {
				slug = slug[:18] + "..."
			}
			// parse and report, breaking on the first error we encounter
			if strings.HasPrefix(row.Line, "Tribe Follows: ") {
				// parse the tribe follows line
				log.Printf("parsePending: name %s: unit %s: element follows: %q\n", pendingRow.Name, ul.UnitId, slug)
			} else if strings.HasPrefix(row.Line, "Tribe Movement: ") {
				// parse the tribe movement line
				log.Printf("parsePending: name %s: unit %s: element movements: %q\n", pendingRow.Name, ul.UnitId, slug)
			} else if strings.HasPrefix(row.Line, "Scout ") {
				// parse the scout line
				log.Printf("parsePending: name %s: unit %s: element scouts: %q\n", pendingRow.Name, ul.UnitId, slug)
			} else if strings.HasPrefix(row.Line, elementStatusPrefix) {
				// parse the element status line
				log.Printf("parsePending: name %s: unit %s: element status: %q\n", pendingRow.Name, ul.UnitId, slug)
				steps, err := lbmoves.ParseMoveResults(turnId, ul.UnitId, []byte(row.Line), showSteps)
				if err != nil {
					log.Fatalf("parsePending: name %s: unit %s: line %d: %v\n", pendingRow.Name, ul.UnitId, row.LineNo, err)
				} else if len(steps) != 1 {
					log.Fatalf("parsePending: name %s: unit %s: line %d: want 1 step, got %d\n", pendingRow.Name, ul.UnitId, row.LineNo, len(steps))
				}
				log.Printf("do something with %v\n", *steps[0])
			} else {
				panic(fmt.Sprintf("section %d: row %d: unknown line type: %q", sectionRow.SectNo, row.LineNo, row.Line))
			}
		}
	}

	status := "completed"
	if errorCount > 0 {
		status = "error"
	}
	if err = q.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
		ID:       pendingRow.ID,
		Status:   "parsing",
		Status_2: status,
	}); err != nil {
		log.Printf("parsePending: name %s: id %5d: setting status to complete: %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Abort(err)
	}

	_ = dbs.Abort(nil)
	log.Printf("parsePending: name %s: id %5d: aborting transaction!\n", pendingRow.Name, pendingRow.ID)

	if err = dbs.Commit(); err != nil {
		log.Printf("parsePending: name %s: id %5d: commiting: %v\n", pendingRow.Name, pendingRow.ID, err)
		return err
	}

	return nil
}
