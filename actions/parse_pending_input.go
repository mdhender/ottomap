// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"database/sql"
	"fmt"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"log"
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

	if err = q.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
		ID:       pendingRow.ID,
		Status:   "pending",
		Status_2: "parsing",
	}); err != nil {
		return dbs.Close(err)
	}
	log.Printf("parsePending: name %s: id %5d: set status to 'parsing'\n", pendingRow.Name, pendingRow.ID)

	sectionRows, err := q.ReadInputSections(db.Ctx, pendingRow.ID)
	if err != nil {
		log.Printf("parsePending: name %s: id %5d: readInputSections: %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Close(err)
	} else if len(sectionRows) == 0 {
		log.Printf("parsePending: name %s: id %5d: readInputSections: %d sections\n", pendingRow.Name, pendingRow.ID, len(sectionRows))
		if err = db.Queries.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
			ID:       pendingRow.ID,
			Status:   "parsing",
			Status_2: "completed",
		}); err != nil {
			log.Printf("parsePending: name %s: id %5d: updatingStatus: %v\n", pendingRow.Name, pendingRow.ID, err)
			return dbs.Close(err)
		}
		return sql.ErrNoRows
	}
	log.Printf("parsePending: name %s: id %5d: readInputSections: %d sections\n", pendingRow.Name, pendingRow.ID, len(sectionRows))

	for _, sectionRow := range sectionRows {
		rows, err := q.ReadInputLines(db.Ctx, sqlc.ReadInputLinesParams{
			ID:     sectionRow.ID,
			SectNo: sectionRow.SectNo,
		})
		if err != nil {
			log.Printf("parsePending: name %s: id %5d: readInputLines: %v\n", pendingRow.Name, pendingRow.ID, err)
			return dbs.Close(err)
		}
		_ = rows
	}

	if err = q.UpdateInputStatus(db.Ctx, sqlc.UpdateInputStatusParams{
		ID:       pendingRow.ID,
		Status:   "parsing",
		Status_2: "completed",
	}); err != nil {
		log.Printf("parsePending: name %s: id %5d: setting status to complete: %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Close(err)
	}

	if err = q.CreateLogMessage(db.Ctx, sqlc.CreateLogMessageParams{
		Arg1:    "parsePending",
		Arg2:    "sectioning",
		Arg3:    fmt.Sprintf("%d", pendingRow.ID),
		Message: "completed",
	}); err != nil {
		log.Printf("parsePending: name %s: id %5d: createLogMessage: %v\n", pendingRow.Name, pendingRow.ID, err)
		return dbs.Close(err)
	}

	dbs.Abort()
	log.Printf("parsePending: name %s: id %5d: aborting transaction!\n", pendingRow.Name, pendingRow.ID)

	if err = dbs.Commit(); err != nil {
		log.Printf("parsePending: name %s: id %5d: commiting: %v\n", pendingRow.Name, pendingRow.ID, err)
		return err
	}

	return nil
}
