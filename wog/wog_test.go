package wog_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/hex"
	"github.com/mdhender/ottomap/wog"
)

func axial(q, r int) hex.Axial { return hex.At(q, r) }

func TestReadAllSamples(t *testing.T) {
	matches, err := filepath.Glob("../testdata/input/*.wxx")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Skip("no sample wxx files")
	}
	for _, path := range matches {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			m, v, err := wog.Read(f)
			if err != nil {
				t.Fatalf("Read failed (version=%v): %v", v, err)
			}
			if m == nil {
				t.Fatal("Read returned nil map")
			}
			t.Logf("read %s: version=%v tiles=%d features=%d labels=%d notes=%d",
				filepath.Base(path), v, m.TileCount(),
				len(m.Features()), len(m.Labels()), len(m.Notes()))
		})
	}
}

func TestReadOrientationAndBounds(t *testing.T) {
	cases := []struct {
		path        string
		wantLayout  hex.Layout
		wantCols    int // tiles wide
		wantRows    int // tiles high
	}{
		{"../testdata/input/2025-2.06-columns-13x11.wxx", hex.OddQ, 13, 11},
		{"../testdata/input/2025-2.06-rows-13x11.wxx", hex.OddR, 13, 11},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(filepath.Base(tc.path), func(t *testing.T) {
			f, err := os.Open(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			m, _, err := wog.Read(f)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if got := m.Layout(); got != tc.wantLayout {
				t.Errorf("Layout: got %v want %v", got, tc.wantLayout)
			}
			min, max, _ := m.BoundsOffset()
			cols := max.Col - min.Col + 1
			rows := max.Row - min.Row + 1
			if cols != tc.wantCols || rows != tc.wantRows {
				t.Errorf("BoundsOffset: got %dx%d (min=%+v max=%+v) want %dx%d",
					cols, rows, min, max, tc.wantCols, tc.wantRows)
			}
		})
	}
}

func TestWriteRequiresVersion(t *testing.T) {
	m := ottomap.NewMap()
	var buf bytes.Buffer
	if err := wog.Write(&buf, m, wog.WriteOptions{}); err == nil {
		t.Fatal("expected error for missing Version")
	}
}

func TestWriteThenRead(t *testing.T) {
	versions := []wog.Version{wog.V2017, wog.V2025}
	for _, version := range versions {
		version := version
		t.Run(version.String(), func(t *testing.T) {
			m := ottomap.NewMap()
			for q := 0; q < 3; q++ {
				for r := 0; r < 2; r++ {
					m.SetTerrain(axial(q, r), ottomap.Terrain("Grassland"))
				}
			}
			m.AddFeature(ottomap.Feature{Kind: "City", Location: axial(1, 0), Label: "Foo"})
			m.AddLabel(ottomap.Label{Text: "Hello", Location: axial(2, 1)})

			var buf bytes.Buffer
			if err := wog.Write(&buf, m, wog.WriteOptions{Version: version}); err != nil {
				t.Fatalf("Write: %v", err)
			}
			got, gotV, err := wog.Read(&buf)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if gotV != version {
				t.Errorf("version: got %v want %v", gotV, version)
			}
			if got.TileCount() == 0 {
				t.Fatal("round-trip lost all tiles")
			}
			if len(got.Features()) != 1 {
				t.Errorf("features: got %d want 1", len(got.Features()))
			}
			if len(got.Labels()) != 1 {
				t.Errorf("labels: got %d want 1", len(got.Labels()))
			}
		})
	}
}
