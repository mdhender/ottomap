package ottomap

import (
	"image/color"

	"github.com/mdhender/ottomap/hex"
)

// Label is a piece of text placed on the map.
type Label struct {
	// ID uniquely identifies this Label within a Map. Assigned by AddLabel
	// if empty.
	ID string

	// Text is the label content.
	Text string

	// Location is the hex the label is anchored to.
	Location hex.Axial

	// Offset displaces the label from its tile anchor.
	Offset Offset

	// Scope controls which zoom level the label appears at.
	Scope Scope

	// Font describes the label font.
	Font FontSpec

	// Color is the text color. Zero value means "use a default".
	Color color.RGBA

	// Outline describes an optional text outline.
	Outline Outline

	// Rotation is rotation in degrees.
	Rotation float64

	// Layer controls draw order.
	Layer Layer

	// Tags are free-form user tags.
	Tags []string

	// GMOnly hides the label from non-GM views.
	GMOnly bool
}

func (l Label) clone() Label {
	out := l
	if l.Tags != nil {
		out.Tags = append([]string(nil), l.Tags...)
	}
	return out
}
