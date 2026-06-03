package ottomap

import (
	"fmt"
	"iter"
	"sort"
	"time"

	"github.com/mdhender/ottomap/hex"
)

// Map is a hex-grid map.
//
// The zero value of *Map (constructed by NewMap) is a valid, empty,
// immediately-writable map. Tiles are sparse: only tiles you set exist.
// Writers compute a bounding box from the tiles, or honor SetBounds if the
// caller pinned one.
type Map struct {
	Name    string
	Created time.Time

	tiles    map[hex.Axial]Tile
	features []Feature
	labels   []Label
	notes    []Note

	layout     hex.Layout // offset convention for col/row projection
	bounds     *bounds    // nil = derive from tiles on write
	nextEntity uint64     // monotonic counter for auto-assigned IDs
}

type bounds struct {
	min, max hex.Axial
}

// NewMap returns an empty Map with the current time as Created.
func NewMap() *Map {
	return &Map{
		Created: time.Now().UTC(),
		tiles:   make(map[hex.Axial]Tile),
	}
}

// --- tiles --------------------------------------------------------------

// Tile returns the Tile at c. The boolean is false if no tile has been set
// at c.
func (m *Map) Tile(c hex.Axial) (Tile, bool) {
	t, ok := m.tiles[c]
	return t, ok
}

// SetTile stores t at c.
func (m *Map) SetTile(c hex.Axial, t Tile) {
	m.tiles[c] = t
}

// DeleteTile removes the tile at c. No-op if no tile exists.
func (m *Map) DeleteTile(c hex.Axial) {
	delete(m.tiles, c)
}

// SetTerrain sets only the terrain on the tile at c, creating a blank tile
// if none exists.
func (m *Map) SetTerrain(c hex.Axial, t Terrain) {
	tile := m.tiles[c]
	tile.Terrain = t
	m.tiles[c] = tile
}

// SetElevation sets only the elevation on the tile at c, creating a blank
// tile if none exists.
func (m *Map) SetElevation(c hex.Axial, e float64) {
	tile := m.tiles[c]
	tile.Elevation = e
	m.tiles[c] = tile
}

// TileCount returns the number of tiles set on the map.
func (m *Map) TileCount() int { return len(m.tiles) }

// Tiles returns an iterator over all tiles, in deterministic order (sorted by
// R, then Q).
func (m *Map) Tiles() iter.Seq2[hex.Axial, Tile] {
	keys := make([]hex.Axial, 0, len(m.tiles))
	for k := range m.tiles {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].R != keys[j].R {
			return keys[i].R < keys[j].R
		}
		return keys[i].Q < keys[j].Q
	})
	return func(yield func(hex.Axial, Tile) bool) {
		for _, k := range keys {
			if !yield(k, m.tiles[k]) {
				return
			}
		}
	}
}

// --- layout -------------------------------------------------------------

// Layout returns the offset-coordinate convention used when projecting this
// map's axial coordinates onto a (column, row) grid. The zero value is
// hex.OddQ (flat-top columns), which matches Worldographer's "COLUMNS"
// orientation; use hex.OddR for "ROWS" (pointy-top rows).
func (m *Map) Layout() hex.Layout { return m.layout }

// SetLayout records the offset convention used by this map. It does not
// move any tiles; it only affects how Bounds are projected to offset space
// and how on-disk formats interpret the (column, row) grid.
func (m *Map) SetLayout(l hex.Layout) { m.layout = l }

// --- bounds -------------------------------------------------------------

// SetBounds pins the map's bounding box. Writers will emit exactly this
// rectangle (tiles outside are dropped; missing tiles inside become blank).
// Pass two corners; they will be normalized.
func (m *Map) SetBounds(a, b hex.Axial) {
	min, max := a, b
	if b.Q < min.Q {
		min.Q, max.Q = b.Q, a.Q
	}
	if b.R < min.R {
		min.R, max.R = b.R, a.R
	}
	m.bounds = &bounds{min: min, max: max}
}

// ClearBounds removes any pinned bounding box; writers will derive bounds
// from the present tiles instead.
func (m *Map) ClearBounds() { m.bounds = nil }

// Bounds returns the map's bounding box. The third return is true if the
// caller pinned it with SetBounds, false if it was derived from the tiles.
// If the map has no tiles and no pinned bounds, returns the zero box and
// false.
func (m *Map) Bounds() (min, max hex.Axial, explicit bool) {
	if m.bounds != nil {
		return m.bounds.min, m.bounds.max, true
	}
	if len(m.tiles) == 0 {
		return hex.Axial{}, hex.Axial{}, false
	}
	first := true
	for c := range m.tiles {
		if first {
			min, max = c, c
			first = false
			continue
		}
		if c.Q < min.Q {
			min.Q = c.Q
		}
		if c.R < min.R {
			min.R = c.R
		}
		if c.Q > max.Q {
			max.Q = c.Q
		}
		if c.R > max.R {
			max.R = c.R
		}
	}
	return min, max, false
}

// BoundsOffset returns the map's bounding box projected onto the
// (column, row) grid implied by the map's Layout. The third return is true
// if the box came from SetBounds, false if it was derived from the tiles.
// If the map has no tiles and no pinned bounds, returns the zero box and
// false.
//
// When the bounds are derived from tiles, BoundsOffset scans every tile in
// offset space rather than converting only the axial extremes (which, for
// hex grids, are not the same rectangle). The result is therefore the
// tight (column, row) bounding box of the present tiles.
func (m *Map) BoundsOffset() (min, max hex.OffsetCoord, explicit bool) {
	if m.bounds != nil {
		min = m.bounds.min.ToOffset(m.layout)
		max = m.bounds.max.ToOffset(m.layout)
		if max.Col < min.Col {
			min.Col, max.Col = max.Col, min.Col
		}
		if max.Row < min.Row {
			min.Row, max.Row = max.Row, min.Row
		}
		return min, max, true
	}
	if len(m.tiles) == 0 {
		return hex.OffsetCoord{}, hex.OffsetCoord{}, false
	}
	first := true
	for c := range m.tiles {
		o := c.ToOffset(m.layout)
		if first {
			min, max = o, o
			first = false
			continue
		}
		if o.Col < min.Col {
			min.Col = o.Col
		}
		if o.Row < min.Row {
			min.Row = o.Row
		}
		if o.Col > max.Col {
			max.Col = o.Col
		}
		if o.Row > max.Row {
			max.Row = o.Row
		}
	}
	return min, max, false
}

// --- features / labels / notes -----------------------------------------

// AddFeature stores f on the map. If f.ID is empty, the Map assigns one.
// Returns the (possibly newly assigned) ID.
func (m *Map) AddFeature(f Feature) string {
	if f.ID == "" {
		f.ID = m.newID("f")
	}
	m.features = append(m.features, f.clone())
	return f.ID
}

// Feature returns the Feature with the given ID.
func (m *Map) Feature(id string) (Feature, bool) {
	for _, f := range m.features {
		if f.ID == id {
			return f.clone(), true
		}
	}
	return Feature{}, false
}

// UpdateFeature replaces the Feature with id == f.ID. Returns false if no
// such feature exists.
func (m *Map) UpdateFeature(f Feature) bool {
	for i := range m.features {
		if m.features[i].ID == f.ID {
			m.features[i] = f.clone()
			return true
		}
	}
	return false
}

// RemoveFeature removes the feature with the given ID. Returns true if a
// feature was removed.
func (m *Map) RemoveFeature(id string) bool {
	for i, f := range m.features {
		if f.ID == id {
			m.features = append(m.features[:i], m.features[i+1:]...)
			return true
		}
	}
	return false
}

// Features returns a copy of the feature list.
func (m *Map) Features() []Feature {
	out := make([]Feature, len(m.features))
	for i, f := range m.features {
		out[i] = f.clone()
	}
	return out
}

// AddLabel stores l on the map. If l.ID is empty, the Map assigns one.
func (m *Map) AddLabel(l Label) string {
	if l.ID == "" {
		l.ID = m.newID("l")
	}
	m.labels = append(m.labels, l.clone())
	return l.ID
}

// Label returns the Label with the given ID.
func (m *Map) Label(id string) (Label, bool) {
	for _, l := range m.labels {
		if l.ID == id {
			return l.clone(), true
		}
	}
	return Label{}, false
}

// UpdateLabel replaces the Label with id == l.ID.
func (m *Map) UpdateLabel(l Label) bool {
	for i := range m.labels {
		if m.labels[i].ID == l.ID {
			m.labels[i] = l.clone()
			return true
		}
	}
	return false
}

// RemoveLabel removes the label with the given ID.
func (m *Map) RemoveLabel(id string) bool {
	for i, l := range m.labels {
		if l.ID == id {
			m.labels = append(m.labels[:i], m.labels[i+1:]...)
			return true
		}
	}
	return false
}

// Labels returns a copy of the label list.
func (m *Map) Labels() []Label {
	out := make([]Label, len(m.labels))
	for i, l := range m.labels {
		out[i] = l.clone()
	}
	return out
}

// AddNote stores n on the map. If n.ID is empty, the Map assigns one.
func (m *Map) AddNote(n Note) string {
	if n.ID == "" {
		n.ID = m.newID("n")
	}
	m.notes = append(m.notes, n)
	return n.ID
}

// Note returns the Note with the given ID.
func (m *Map) Note(id string) (Note, bool) {
	for _, n := range m.notes {
		if n.ID == id {
			return n, true
		}
	}
	return Note{}, false
}

// UpdateNote replaces the Note with id == n.ID.
func (m *Map) UpdateNote(n Note) bool {
	for i := range m.notes {
		if m.notes[i].ID == n.ID {
			m.notes[i] = n
			return true
		}
	}
	return false
}

// RemoveNote removes the note with the given ID.
func (m *Map) RemoveNote(id string) bool {
	for i, n := range m.notes {
		if n.ID == id {
			m.notes = append(m.notes[:i], m.notes[i+1:]...)
			return true
		}
	}
	return false
}

// Notes returns a copy of the note list.
func (m *Map) Notes() []Note {
	return append([]Note(nil), m.notes...)
}

// --- clone --------------------------------------------------------------

// Clone returns a deep copy of m. Use this to make per-player variants
// without affecting the master.
func (m *Map) Clone() *Map {
	out := &Map{
		Name:       m.Name,
		Created:    m.Created,
		tiles:      make(map[hex.Axial]Tile, len(m.tiles)),
		features:   make([]Feature, len(m.features)),
		labels:     make([]Label, len(m.labels)),
		notes:      append([]Note(nil), m.notes...),
		layout:     m.layout,
		nextEntity: m.nextEntity,
	}
	for k, v := range m.tiles {
		out.tiles[k] = v
	}
	for i, f := range m.features {
		out.features[i] = f.clone()
	}
	for i, l := range m.labels {
		out.labels[i] = l.clone()
	}
	if m.bounds != nil {
		b := *m.bounds
		out.bounds = &b
	}
	return out
}

// --- internal -----------------------------------------------------------

func (m *Map) newID(prefix string) string {
	m.nextEntity++
	return fmt.Sprintf("%s-%d", prefix, m.nextEntity)
}
