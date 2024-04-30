// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

import "bytes"

// this file implements scanners. they're not quite the same as parsers,
// they return only slices and never types.

// AcceptMovementStep returns the step and the remainder of the input.
// If there is no step, returns nil, original input.
//
// Step is MOVE? DIRECTION_CODE DASH TERRAIN_CODE STEP_TEXT? &TERMINATOR
//
// NB: That '&' means that the token is expected but never consumed.
func AcceptMovementStep(input []byte) ([]byte, []byte) {
	token, rest := AcceptMove(input)
	length := len(token)
	if token != nil {
		// accept SPACE*
		token, rest = AcceptSpaces(rest)
		length += len(token)
	}
	// expect DIRECTION_CODE
	token, rest = AcceptDirectionCode(rest)
	if token == nil {
		return nil, input
	}
	length += len(token)
	// expect DASH
	token, rest = AcceptDash(rest)
	if token == nil {
		return nil, input
	}
	length += len(token)
	// expect TERRAIN_CODE
	token, rest = AcceptTerrainCode(rest)
	if token == nil {
		return nil, input
	}
	length += len(token)
	// accept STEP_TEXT
	token, rest = AcceptStepText(rest)
	length += len(token)

	if length == len(input) {
		return input, nil
	}
	return input[:length], input[length:]
}

func AcceptCantMoveToEndOfStep(input []byte) ([]byte, []byte) {
	if !bytes.HasPrefix(input, []byte{'C', 'a', 'n', '\'', 't', ' ', 'M', 'o', 'v', 'e'}) {
		return nil, input
	}
	return input, nil
}

// AcceptDirectionCode is a function that checks the beginning of the input for a valid direction pattern.
// The function iterates over a predefined list of directions ['N', 'NE', 'SE', 'S', 'SW', 'NW'] and
// checks if the input starts with any of these directions.
// If a match is found, the function ensures the following character is a valid terminator, i.e.,
// either a space, tab, comma, dash, or a backslash ('\\') or the End Of File (EOF).
// If these conditions are met, the function returns the matched direction and the remaining part of the input.
// If no matches are found, the function returns nil and the original input.
//
// Parameters:
// - input ([]byte): The input byte array. This is typically the string we want to scan for directions.
//
// Returns:
//   - ([]byte, []byte): Two arrays of bytes. The first one represents the scanned direction.
//     The second one represents the remaining part of the input not scanned.
func AcceptDirectionCode(input []byte) ([]byte, []byte) {
	for _, dir := range [][]byte{
		{'N'}, {'N', 'E'}, {'S', 'E'}, {'S'}, {'S', 'W'}, {'N', 'W'},
	} {
		if bytes.HasPrefix(input, dir) {
			// must be followed by a valid terminator, which is
			// a space, comma, backslash, or EOF
			rest := input[len(dir):]
			if len(rest) == 0 || bytes.IndexByte([]byte{' ', '\t', '-', ',', '\\'}, rest[0]) != -1 {
				return dir, rest
			}
		}
	}
	return nil, input
}

func AcceptDash(input []byte) ([]byte, []byte) {
	if len(input) == 0 || input[0] != '-' {
		return nil, input
	}
	return input[:1], input[1:]
}

// AcceptDirectionToEndOfStep scans the input byte slice for a step token.
// The step token starts with direction code followed by a dash and then a
// terrain code. It continues until the next step code (signaled by a
// backslash and the step code) or the end of the input.
//
// The function returns the entire step token and the remaining input. The
// backslash, if present, is never included in the returned text.
//
// If the input does not start with a valid direction token, the function
// returns nil, input.
func AcceptDirectionToEndOfStep(input []byte) ([]byte, []byte) {
	if len(input) == 0 {
		return nil, nil
	}

	length := 0

	// expect DIRECTION DASH TERRAIN_CODE
	token, rest := AcceptDirectionCode(input)
	if token == nil {
		return nil, input
	}
	length += len(token)
	token, rest = AcceptDash(rest)
	if token == nil {
		return nil, input
	}
	length += len(token)
	token, rest = AcceptTerrainCode(rest)
	if token == nil {
		return nil, input
	}
	length += len(token)

	// step continues to a valid terminator or EOF.
	for len(rest) != 0 {
		if rest[0] == '\\' {
			// check for valid terminator
			if bytes.Equal(rest, []byte{'\\'}) { // terminated by EOF
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'N', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'N', 'E', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'S', 'E', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'S', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'S', 'W', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'N', 'W', '-'}) { // terminated by a direction
				break
			} else if bytes.HasPrefix(rest, []byte{'\\', 'C', 'a', 'n', '\'', 't', ' ', 'M', 'o', 'v', 'e'}) { // terminated by failed movement
				break
			}
		}
		rest, length = rest[1:], length+1
	}
	if len(rest) == 0 { // terminated by EOF
		return nil, input
	}

	return input[:length], input[length:]
}

// AcceptMove scans the input byte slice for the word "Move" and returns the
// word as a token if it is present and properly terminated. If the input starts
// with "Move", it checks if the word is terminated by either EOF, a space, or a
// tab character. If it is terminated by one of these, it returns "Move" as the
// token and the rest of the input after the terminator. If the word "Move" is
// not present or not followed by a valid terminator, it returns nil for the
// token and the original input as the rest.
func AcceptMove(input []byte) ([]byte, []byte) {
	// Check if the input starts with the word "Move"
	if !bytes.HasPrefix(input, []byte{'M', 'o', 'v', 'e'}) {
		// If it doesn't, return nil for the token and the original input as the rest
		return nil, input
	}

	// If the input is exactly "Move" and terminated by EOF
	if len(input) == 4 { // terminated by EOF
		// Return "Move" as the token and nil as the rest
		return input, nil
	} else if input[4] == ' ' || input[4] == '\t' { // terminated by space or tab
		// If the input is "Move" followed by a space or tab
		// Return "Move" as the token and the rest of the input after the space/tab
		return input[:4], input[4:]
	}

	// If the input is "Move" followed by something other than a space, tab, or EOF
	// Return nil for the token and the original input as the rest
	return nil, input
}

// AcceptSpaces is a utility function that processes input, a slice of byte.
// If the input has leading spaces, returns a token containing those spaces
// and remainder of the input. Otherwise, returns nil and the original input.
func AcceptSpaces(input []byte) ([]byte, []byte) {
	pos := 0
	for pos < len(input) && (input[pos] == ' ' || input[pos] == '\t') {
		pos++
	}
	if pos == 0 { // no leading spaces
		return nil, input
	} else if pos == len(input) { // entirely spaces
		return input, nil
	}
	return input[:pos], input[pos:]
}

// AcceptStepText returns the text token and the remainder of the input.
// The text token runs to the terminator or EOF. It never includes the
// terminating backslash.
func AcceptStepText(input []byte) ([]byte, []byte) {
	for length, rest := 0, input; len(rest) != 0; rest, length = rest[1:], length+1 {
		if rest[0] != '\\' {
			continue
		}

		bs := rest[1:]

		// terminated by backslash EOF
		if len(bs) == 0 {
			return input[:length], input[length:]
		}

		// terminated by a direction
		if bytes.HasPrefix(bs, []byte{'N', '-'}) {
			return input[:length], input[length:]
		} else if bytes.HasPrefix(bs, []byte{'N', 'E', '-'}) {
			return input[:length], input[length:]
		} else if bytes.HasPrefix(bs, []byte{'S', 'E', '-'}) {
			return input[:length], input[length:]
		} else if bytes.HasPrefix(bs, []byte{'S', '-'}) {
			return input[:length], input[length:]
		} else if bytes.HasPrefix(bs, []byte{'S', 'W', '-'}) {
			return input[:length], input[length:]
		} else if bytes.HasPrefix(bs, []byte{'N', 'W', '-'}) {
			return input[:length], input[length:]
		}

		// terminated by a failed move
		if bytes.HasPrefix(bs, []byte{'C', 'a', 'n', '\'', 't', ' ', 'M', 'o', 'v', 'e'}) {
			return input[:length], input[length:]
		}
	}

	// text runs to EOF
	return input, nil
}

func AcceptTerrainCode(input []byte) ([]byte, []byte) {
	for _, code := range [][]byte{
		// three character terrain codes
		{'L', 'V', 'M'}, {'H', 'S', 'M'},
		// two character terrain codes
		{'C', 'H'}, {'P', 'R'},
		// one character terrain codes
		{'D'}, {'O'}, {'R'},
	} {
		if bytes.HasPrefix(input, code) {
			// must be followed by a space, comma, backslash, or EOF
			rest := input[len(code):]
			if len(rest) == 0 || bytes.IndexByte([]byte{' ', '\t', '-', ',', '\\'}, rest[0]) != -1 {
				return code, rest
			}
		}
	}
	return nil, input
}
