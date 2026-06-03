package ottomap

import (
	"testing"

	"github.com/mdhender/ottomap/hex"
)

func TestNewMapIsValidAndEmpty(t *testing.T) {
	m := NewMap()
	if m.TileCount() != 0 {
		t.Fatalf("expected 0 tiles, got %d", m.TileCount())
	}
	if _, _, explicit := m.Bounds(); explicit {
		t.Fatalf("expected unpinned bounds on new map")
	}
	if len(m.Features()) != 0 || len(m.Labels()) != 0 || len(m.Notes()) != 0 {
		t.Fatalf("expected empty entity lists")
	}
}

func TestTileSetGetDelete(t *testing.T) {
	m := NewMap()
	c := hex.At(3, 4)
	if _, ok := m.Tile(c); ok {
		t.Fatal("unset tile should not be present")
	}
	m.SetTile(c, Tile{Terrain: "Grassland", Elevation: 100})
	got, ok := m.Tile(c)
	if !ok {
		t.Fatal("tile missing after SetTile")
	}
	if got.Terrain != "Grassland" || got.Elevation != 100 {
		t.Fatalf("tile fields wrong: %+v", got)
	}
	m.DeleteTile(c)
	if _, ok := m.Tile(c); ok {
		t.Fatal("tile present after DeleteTile")
	}
}

func TestSetTerrainCreatesBlankTile(t *testing.T) {
	m := NewMap()
	c := hex.At(0, 0)
	m.SetTerrain(c, "Forest")
	got, ok := m.Tile(c)
	if !ok || got.Terrain != "Forest" {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
	m.SetElevation(c, 500)
	got, _ = m.Tile(c)
	if got.Elevation != 500 || got.Terrain != "Forest" {
		t.Fatalf("elevation update lost terrain: %+v", got)
	}
}

func TestTilesIterationIsDeterministic(t *testing.T) {
	m := NewMap()
	coords := []hex.Axial{
		hex.At(2, 0), hex.At(0, 1), hex.At(1, 1), hex.At(0, 0),
	}
	for i, c := range coords {
		m.SetTile(c, Tile{Elevation: float64(i)})
	}
	want := []hex.Axial{
		hex.At(0, 0), hex.At(2, 0), hex.At(0, 1), hex.At(1, 1),
	}
	var got []hex.Axial
	for c := range m.Tiles() {
		got = append(got, c)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tiles, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("position %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestDerivedBounds(t *testing.T) {
	m := NewMap()
	m.SetTerrain(hex.At(-2, 5), "X")
	m.SetTerrain(hex.At(3, -1), "Y")
	min, max, explicit := m.Bounds()
	if explicit {
		t.Fatal("bounds should be derived, not explicit")
	}
	if min != (hex.Axial{Q: -2, R: -1}) || max != (hex.Axial{Q: 3, R: 5}) {
		t.Errorf("bounds: min=%v max=%v", min, max)
	}
}

func TestExplicitBoundsNormalized(t *testing.T) {
	m := NewMap()
	m.SetBounds(hex.At(5, 5), hex.At(0, 0))
	min, max, explicit := m.Bounds()
	if !explicit {
		t.Fatal("bounds should be explicit")
	}
	if min != (hex.Axial{Q: 0, R: 0}) || max != (hex.Axial{Q: 5, R: 5}) {
		t.Errorf("bounds: min=%v max=%v", min, max)
	}
}

func TestFeatureCRUD(t *testing.T) {
	m := NewMap()
	id := m.AddFeature(Feature{Kind: "City", Location: hex.At(1, 2), Label: "Foo"})
	if id == "" {
		t.Fatal("expected auto-assigned ID")
	}
	f, ok := m.Feature(id)
	if !ok || f.Label != "Foo" {
		t.Fatalf("got %+v ok=%v", f, ok)
	}
	f.Label = "Bar"
	if !m.UpdateFeature(f) {
		t.Fatal("UpdateFeature returned false")
	}
	again, _ := m.Feature(id)
	if again.Label != "Bar" {
		t.Errorf("update lost: %+v", again)
	}
	if !m.RemoveFeature(id) {
		t.Fatal("RemoveFeature returned false")
	}
	if _, ok := m.Feature(id); ok {
		t.Fatal("feature present after remove")
	}
}

func TestAddFeaturePreservesExplicitID(t *testing.T) {
	m := NewMap()
	id := m.AddFeature(Feature{ID: "my-id", Kind: "X"})
	if id != "my-id" {
		t.Errorf("ID changed: %q", id)
	}
}

func TestFeatureTagsAreCopied(t *testing.T) {
	m := NewMap()
	tags := []string{"a", "b"}
	id := m.AddFeature(Feature{Kind: "X", Tags: tags})
	tags[0] = "MUTATED"
	got, _ := m.Feature(id)
	if got.Tags[0] != "a" {
		t.Errorf("tag mutation leaked through: %v", got.Tags)
	}
	got.Tags[1] = "ALSO MUTATED"
	again, _ := m.Feature(id)
	if again.Tags[1] != "b" {
		t.Errorf("tag mutation leaked back: %v", again.Tags)
	}
}

func TestCloneIsIndependent(t *testing.T) {
	m := NewMap()
	m.SetTerrain(hex.At(0, 0), "A")
	m.AddFeature(Feature{Kind: "City", Location: hex.At(0, 0), Tags: []string{"x"}})

	c := m.Clone()
	c.SetTerrain(hex.At(0, 0), "B")
	c.AddFeature(Feature{Kind: "Town"})

	orig, _ := m.Tile(hex.At(0, 0))
	if orig.Terrain != "A" {
		t.Errorf("clone leaked into original tile: %v", orig.Terrain)
	}
	if len(m.Features()) != 1 {
		t.Errorf("clone leaked into original features: %d", len(m.Features()))
	}
	if len(c.Features()) != 2 {
		t.Errorf("clone feature count: %d", len(c.Features()))
	}
}
