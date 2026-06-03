package wxxio

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/mdhender/ottomap"
)

// ParsedTile is the parsed form of one line of <tilerow>.
type ParsedTile struct {
	TerrainSlot int
	Elevation   float64
	Icy         bool
	GMOnly      bool
	Resources   ottomap.Resources
	Background  color.RGBA
}

// ParseTileLine decodes one tab-separated tile line. Accepted field counts
// are 6, 7, 11, and 12. The 6/7 forms use the "Z" sentinel at field 5 to
// indicate that all non-Animal resources are zero. The 7/12 forms append a
// trailing float-RGBA custom background color.
func ParseTileLine(line string) (ParsedTile, error) {
	var r ParsedTile
	fields := strings.Split(line, "\t")
	if n := len(fields); n != 6 && n != 7 && n != 11 && n != 12 {
		return r, fmt.Errorf("expected 6/7/11/12 fields, got %d", n)
	}
	var err error
	if r.TerrainSlot, err = strconv.Atoi(strings.TrimSpace(fields[0])); err != nil {
		return r, fmt.Errorf("terrain slot: %w", err)
	}
	if r.Elevation, err = strconv.ParseFloat(strings.TrimSpace(fields[1]), 64); err != nil {
		return r, fmt.Errorf("elevation: %w", err)
	}
	r.Icy = strings.TrimSpace(fields[2]) == "1"
	r.GMOnly = strings.TrimSpace(fields[3]) == "1"
	animal, err := strconv.Atoi(strings.TrimSpace(fields[4]))
	if err != nil {
		return r, fmt.Errorf("resources.animal: %w", err)
	}
	r.Resources.Animal = ClampU8(animal)

	compressed := len(fields) == 6 || len(fields) == 7
	if compressed {
		if strings.TrimSpace(fields[5]) != "Z" {
			return r, fmt.Errorf("expected 'Z' sentinel at field 5, got %q", fields[5])
		}
	} else {
		ptrs := []*uint8{
			&r.Resources.Brick, &r.Resources.Crops, &r.Resources.Gems,
			&r.Resources.Lumber, &r.Resources.Metals, &r.Resources.Rock,
		}
		for i, p := range ptrs {
			v, err := strconv.Atoi(strings.TrimSpace(fields[5+i]))
			if err != nil {
				return r, fmt.Errorf("resource[%d]: %w", i, err)
			}
			*p = ClampU8(v)
		}
	}
	if len(fields) == 7 || len(fields) == 12 {
		bg, err := ParseFloatRGBA(strings.TrimSpace(fields[len(fields)-1]))
		if err != nil {
			return r, fmt.Errorf("background: %w", err)
		}
		r.Background = bg
	}
	return r, nil
}

// FormatTileLine emits one line of <tilerow> chardata for the given Tile.
// The line ends with a newline.
func FormatTileLine(b *strings.Builder, t ottomap.Tile, slot int) {
	icy := BoolDigit(t.Icy)
	gm := BoolDigit(t.GMOnly)
	r := t.Resources
	hasOtherResources := r.Brick|r.Crops|r.Gems|r.Lumber|r.Metals|r.Rock != 0
	hasBg := t.Background != (color.RGBA{})
	if !hasOtherResources {
		fmt.Fprintf(b, "%d\t%d\t%d\t%d\t%d\tZ", slot, int(t.Elevation), icy, gm, r.Animal)
	} else {
		fmt.Fprintf(b, "%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d",
			slot, int(t.Elevation), icy, gm,
			r.Animal, r.Brick, r.Crops, r.Gems, r.Lumber, r.Metals, r.Rock)
	}
	if hasBg {
		fmt.Fprintf(b, "\t%s", FormatFloatRGBA(t.Background))
	}
	b.WriteByte('\n')
}

// SplitTileLines splits a <tilerow> chardata block into one tile per
// element, dropping blank lines.
func SplitTileLines(s string) []string {
	parts := strings.Split(s, "\n")
	out := parts[:0]
	for _, l := range parts {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}

// ClampU8 clamps an int to the uint8 range and returns the result.
func ClampU8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// BoolDigit returns 1 for true, 0 for false.
func BoolDigit(b bool) int {
	if b {
		return 1
	}
	return 0
}
