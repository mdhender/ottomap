// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package adapters

import "github.com/mdhender/ottomap/domain"

// this files defines adapters for fetching the raw text from a report section.
// not needed or used at the moment.

// RawLocationLine is a helper function, just in case we ever change the type of LocationLine.
func RawLocationLine(rs *domain.ReportSection) []byte {
	return []byte(rs.LocationLine)
}

// RawMovementLine is a helper function, just in case we ever change the type of MovementLine.
func RawMovementLine(rs *domain.ReportSection) []byte {
	return []byte(rs.MovementLine)
}

// RawScoutLines is a helper function, just in case we ever change the type of ScoutLines
func RawScoutLines(rs *domain.ReportSection) [][]byte {
	lines := make([][]byte, len(rs.ScoutLines))
	for i, line := range rs.ScoutLines {
		lines[i] = []byte(line)
	}
	return lines
}

// RawStatusLine is a helper function, just in case we ever change the type of Status.
func RawStatusLine(rs *domain.ReportSection) []byte {
	return []byte(rs.StatusLine)
}
