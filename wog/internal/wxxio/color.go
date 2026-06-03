package wxxio

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// ParseFloatRGBA parses Worldographer's comma-separated float RGBA syntax,
// e.g. "0.5,0.5,0.5,1.0". Returns the zero color.RGBA when s is empty,
// "null", or the "0.0,0.0,0.0,1.0" sentinel (Worldographer encodes "no
// color" as opaque black in many fields).
func ParseFloatRGBA(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" || s == "0.0,0.0,0.0,1.0" {
		return color.RGBA{}, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return color.RGBA{}, fmt.Errorf("expected 4 components, got %d: %q", len(parts), s)
	}
	var c [4]uint8
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("component %d: %w", i, err)
		}
		if f < 0 {
			f = 0
		} else if f > 1 {
			f = 1
		}
		c[i] = uint8(f*255 + 0.5)
	}
	return color.RGBA{R: c[0], G: c[1], B: c[2], A: c[3]}, nil
}

// FormatFloatRGBA renders a color.RGBA into Worldographer's float-rgba
// syntax. The zero RGBA renders as "null" (matching ParseFloatRGBA).
func FormatFloatRGBA(c color.RGBA) string {
	if c == (color.RGBA{}) {
		return "null"
	}
	return fmt.Sprintf("%g,%g,%g,%g",
		float64(c.R)/255,
		float64(c.G)/255,
		float64(c.B)/255,
		float64(c.A)/255,
	)
}
