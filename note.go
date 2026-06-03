package ottomap

import "github.com/mdhender/ottomap/hex"

// Note is a free-form note attached to a map. Notes have a title and body
// (the body is preserved verbatim, including newlines) and may optionally be
// anchored to a hex.
type Note struct {
	// ID uniquely identifies this Note within a Map. Assigned by AddNote
	// if empty.
	ID string

	// Title is the short header shown in note lists.
	Title string

	// Body is the note's content.
	Body string

	// Location is the hex the note is anchored to. The zero value of
	// hex.Axial means "unanchored".
	Location hex.Axial

	// GMOnly hides the note from non-GM views.
	GMOnly bool
}
