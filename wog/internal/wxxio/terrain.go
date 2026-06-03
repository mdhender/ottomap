package wxxio

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mdhender/ottomap"
)

// ParseTerrainMap parses the tab-delimited "Name<TAB>Index<TAB>..."
// content of a <terrainmap> element.
func ParseTerrainMap(s string) (map[int]ottomap.Terrain, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[int]ottomap.Terrain{}, nil
	}
	fields := strings.Split(s, "\t")
	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("odd number of fields: %d", len(fields))
	}
	out := make(map[int]ottomap.Terrain, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		name := strings.TrimSpace(fields[i])
		idx, err := strconv.Atoi(strings.TrimSpace(fields[i+1]))
		if err != nil {
			return nil, fmt.Errorf("slot %q: bad index %q", name, fields[i+1])
		}
		out[idx] = ottomap.Terrain(name)
	}
	return out, nil
}

// TerrainRegistry assigns numeric slots to Terrain names so that the
// encoder can produce a <terrainmap> and emit tile-row terrain indices.
// Slot 0 is always "Blank".
type TerrainRegistry struct {
	slots map[ottomap.Terrain]int
	order []ottomap.Terrain
}

// NewTerrainRegistry returns a registry pre-seeded with "Blank" in slot 0.
func NewTerrainRegistry() *TerrainRegistry {
	r := &TerrainRegistry{slots: map[ottomap.Terrain]int{}}
	r.Intern("Blank")
	return r
}

// Intern assigns a slot to t if it doesn't already have one and returns it.
// Empty Terrain names are treated as "Blank".
func (r *TerrainRegistry) Intern(t ottomap.Terrain) int {
	if t == "" {
		t = "Blank"
	}
	if idx, ok := r.slots[t]; ok {
		return idx
	}
	idx := len(r.order)
	r.slots[t] = idx
	r.order = append(r.order, t)
	return idx
}

// Slot returns the slot for t, interning it if needed.
func (r *TerrainRegistry) Slot(t ottomap.Terrain) int { return r.Intern(t) }

// WriteTerrainMap writes the tab-delimited terrainmap content into b.
func (r *TerrainRegistry) WriteTerrainMap(b *strings.Builder) {
	type entry struct {
		name string
		idx  int
	}
	es := make([]entry, 0, len(r.order))
	for i, n := range r.order {
		es = append(es, entry{string(n), i})
	}
	sort.Slice(es, func(i, j int) bool { return es[i].idx < es[j].idx })
	for i, e := range es {
		if i == 0 {
			fmt.Fprintf(b, "%s\t%d", e.name, e.idx)
		} else {
			fmt.Fprintf(b, "\t%s\t%d", e.name, e.idx)
		}
	}
}
