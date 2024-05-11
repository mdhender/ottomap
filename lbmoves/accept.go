// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lbmoves

import (
	"bytes"
	"regexp"
)

func acceptDirection(b []byte) bool {
	if bytes.Equal(b, []byte{'N'}) {
		return true
	} else if bytes.Equal(b, []byte{'N', 'E'}) {
		return true
	} else if bytes.Equal(b, []byte{'S', 'E'}) {
		return true
	} else if bytes.Equal(b, []byte{'S'}) {
		return true
	} else if bytes.Equal(b, []byte{'S', 'W'}) {
		return true
	} else if bytes.Equal(b, []byte{'N', 'W'}) {
		return true
	}
	return false
}

func acceptFordEdge(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'F', 'o', 'r', 'd', ' '})
}

func acceptLakeEdge(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'O', ' '})
}

func acceptOceanEdge(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'O', ' '})
}

func acceptPassEdge(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'P', 'a', 's', 's', ' '})
}

func acceptPatrolledAndFound(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'P', 'a', 't', 'r', 'o', 'l', 'l', 'e', 'd', ' ', 'a', 'n', 'd', ' ', 'f', 'o', 'u', 'n', 'd', ' '})
}

func acceptRiverEdge(b []byte) bool {
	return bytes.HasPrefix(b, []byte{'R', 'i', 'v', 'e', 'r', ' '})
}

var (
	rxUnitId = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])?$`)
)

func acceptUnitId(b []byte) bool {
	return rxUnitId.Match(b)
}
