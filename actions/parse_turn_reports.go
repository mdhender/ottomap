// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package actions

import (
	"bytes"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/wbmoves"
	"log"
	"os"
	"unicode"
)

// ParseTurnReports assumes that reports are sorted by turn and clan.
func ParseTurnReports(clanId string, debug DebugTurnReports, reports ...string) error {
	log.Printf("parseTurnReports: clanId %q: debugSteps %v: debugNodes %v\n", clanId, debug.Steps, debug.Nodes)

	for _, report := range reports {
		log.Printf("parseTurnReports: report %q\n", report)
		data, err := os.ReadFile(report)
		if err != nil {
			return err
		}
		log.Printf("parseTurnReports: read %d bytes\n", len(data))
		var lines []*wbmoves.Line_t
		for n, line := range bytes.Split(data, []byte{'\n'}) {
			lines = append(lines, &wbmoves.Line_t{
				LineNo: n + 1,
				Text:   bytes.TrimRightFunc(line, unicode.IsSpace),
			})
		}
		log.Printf("parseTurnReports: read %d lines\n", len(lines))
		err = ParseTurnReport(lines, debug.Steps, debug.Nodes)
		if err != nil {
			return err
		}
	}

	return nil
}

func ParseTurnReport(lines []*wbmoves.Line_t, debugSteps, debugNodes bool) error {
	log.Printf("parseTurnReport: debugSteps %v: debugNodes %v\n", debugSteps, debugNodes)

	if len(lines) == 0 {
		return cerrs.ErrNotATurnReport
	}

	panic("!implemented")
}

type DebugTurnReports struct {
	Nodes bool
	Steps bool
}
