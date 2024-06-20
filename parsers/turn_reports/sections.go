// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turn_reports

// Split splits the input into sections. It returns the sections along
// with the section separator. We trim leading and trailing new-lines
// from each section and then force the section to end with a new-line.
//
// We check for a few types of separators and use the
// first one that we find. If we can't find a separator,
// we return the entire input as the first value and nil
// for the separator.
//
// NB: The first turn report (the "setup" turn) might have just
// one section, so we wouldn't find a section separator. The
// instructions should tell the user to manually add one. Or the
// caller should have logic to handle.
func Split(input []byte) ([][]byte, []byte) {
	panic("replaced by sections.Split")
}
