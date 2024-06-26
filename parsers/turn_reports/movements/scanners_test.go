// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements_test

import (
	"github.com/mdhender/ottomap/parsers/turn_reports/movements"
	"testing"
)

func TestAcceptDirectionCode(t *testing.T) {
	tests := []struct {
		id          int
		input       string
		token, rest string
	}{
		// valid directions followed by EOF terminator
		{1001, "N", "N", ""},
		{1002, "NE", "NE", ""},
		{1003, "SE", "SE", ""},
		{1004, "S", "S", ""},
		{1005, "SW", "SW", ""},
		{1006, "NW", "NW", ""},
		// valid direction followed by other terminators
		{2001, "N ", "N", " "},
		{2002, "NE,", "NE", ","},
		{2003, "SE\\", "SE", "\\"},
		{2004, "S\t", "S", "\t"},
		{2005, "NW-PR", "NW", "-PR"},
		// valid direction not followed by a terminator
		{3001, "N\n", "", "N\n"},
		{3002, "NNE ", "", "NNE "},
		{3003, "South", "", "South"},
		{3004, "SWS ", "", "SWS "},
		// invalid direction
		{4001, "Howdy", "", "Howdy"},
		{4001, " N ", "", " N "},
		// empty input
		{9001, "", "", ""},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptDirectionCode([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptMove(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		{id: 1001, input: "Move", token: "Move", rest: ""},
		{id: 1002, input: "Move ", token: "Move", rest: " "},
		{id: 1003, input: "Move\t", token: "Move", rest: "\t"},
		{id: 2001, input: "MoveX", token: "", rest: "MoveX"},
		{id: 2002, input: "NotMove", token: "", rest: "NotMove"},
		{id: 9001, input: "", token: "", rest: ""},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptMove([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptMovementStep(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		// SW-D,  \ 1230,  0230\SW-PR,  \  why are there numbers?
		{id: 1001, input: `Move SW-PR,  River S\SW-PR,  River SE`, token: `Move SW-PR,  River S`, rest: `\SW-PR,  River SE`},
		{id: 1002, input: `Move SW-D, \ 1230,  0230\SW-PR,  \`, token: `Move SW-D, \ 1230,  0230`, rest: `\SW-PR,  \`},
		{id: 9001, input: ``, token: ``, rest: ``},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptMovementStep([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptSpaces(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		{1001, "  text ", "  ", "text "},
		{1002, "\t\ttext\t", "\t\t", "text\t"},
		{1003, " \ttext\t ", " \t", "text\t "},
		{2001, "text  ", "", "text  "},
		{3001, "text", "", "text"},
		{9001, "", "", ""},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptSpaces([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest : got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptCantMoveToEndOfStep(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		{id: 1001, input: `Can't Move on Ocean to NW of HEX`, token: `Can't Move on Ocean to NW of HEX`, rest: ``},
		{id: 2001, input: `Can't Moove`, token: ``, rest: `Can't Moove`},
		{id: 9001, input: ``, token: ``, rest: ``},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptCantMoveToEndOfStep([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptDirectionToEndOfStep(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		// Move NW-PR,  \NW-PR,  O NW,  N\Can't Move on Ocean to NW of HEX
		{id: 1001, input: `NW-PR,  \NW-PR`, token: `NW-PR,  `, rest: `\NW-PR`},
		{id: 1002, input: `NW-PR,  O NW,  N\Can't Move on Ocean to NW of HEX`, token: `NW-PR,  O NW,  N`, rest: `\Can't Move on Ocean to NW of HEX`},
		{id: 2001, input: `\Hello`, token: ``, rest: `\Hello`},
		{id: 9001, input: ``, token: ``, rest: ``},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptDirectionToEndOfStep([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}

func TestAcceptStepText(t *testing.T) {
	tests := []struct {
		id    int
		input string
		token string
		rest  string
	}{
		// Move NW-PR,  \NW-PR,  O NW,  N\Can't Move on Ocean to NW of HEX
		{id: 1001, input: `,  \NW-PR`, token: `,  `, rest: `\NW-PR`},
		{id: 1002, input: `,  O NW,  N\Can't Move on Ocean to NW of HEX`, token: `,  O NW,  N`, rest: `\Can't Move on Ocean to NW of HEX`},
		{id: 1003, input: `,  \ 1230,  0230\SW-D,  \  why are there numbers?`, token: `,  \ 1230,  0230`, rest: `\SW-D,  \  why are there numbers?`},
		{id: 5001, input: `foo\bar`, token: `foo\bar`, rest: ``},
		{id: 9001, input: ``, token: ``, rest: ``},
	}

	for _, tt := range tests {
		gotToken, gotRest := movements.AcceptStepText([]byte(tt.input))
		token, rest := string(gotToken), string(gotRest)
		if token != tt.token {
			t.Errorf("%d: %q: token: got %q, want %q\n", tt.id, tt.input, token, tt.token)
		}
		if rest != tt.rest {
			t.Errorf("%d: %q: rest:  got %q, want %q\n", tt.id, tt.input, rest, tt.rest)
		}
	}
}
