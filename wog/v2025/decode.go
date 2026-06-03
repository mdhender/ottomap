package v2025

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/hex"
	"github.com/mdhender/ottomap/wog/internal/wxxio"
)

// Decode parses a UTF-8 XML document (Worldographer 2025 schema family) and
// returns the resulting domain Map.
func Decode(xmlBytes []byte) (*ottomap.Map, error) {
	var s schemaMap
	dec := xml.NewDecoder(bytes.NewReader(xmlBytes))
	dec.Strict = false
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("xml decode: %w", err)
	}

	m := ottomap.NewMap()

	terrainBySlot, err := wxxio.ParseTerrainMap(s.TerrainMap.InnerText)
	if err != nil {
		return nil, fmt.Errorf("terrainmap: %w", err)
	}

	orientation := strings.ToUpper(s.HexOrientation)
	if orientation == "" {
		orientation = "COLUMNS"
	}
	if err := decodeTiles(m, &s.Tiles, terrainBySlot, orientation); err != nil {
		return nil, fmt.Errorf("tiles: %w", err)
	}
	decodeFeatures(m, s.Features.Features, orientation)
	decodeLabels(m, s.Labels.Labels, orientation)
	decodeNotes(m, s.Notes.Notes)

	return m, nil
}

func decodeTiles(m *ottomap.Map, t *schemaTiles, terrains map[int]ottomap.Terrain, orientation string) error {
	for rowIdx, tr := range t.TileRows {
		lines := wxxio.SplitTileLines(tr.InnerText)
		for tileIdx, line := range lines {
			pt, err := wxxio.ParseTileLine(line)
			if err != nil {
				return fmt.Errorf("tilerow %d tile %d: %w", rowIdx, tileIdx, err)
			}
			coord := wxxio.TileCoord(orientation, rowIdx, tileIdx)
			m.SetTile(coord, ottomap.Tile{
				Terrain:    terrains[pt.TerrainSlot],
				Elevation:  pt.Elevation,
				Icy:        pt.Icy,
				GMOnly:     pt.GMOnly,
				Resources:  pt.Resources,
				Background: pt.Background,
			})
		}
	}
	min, max := tileExtent(t, orientation)
	m.SetBounds(min, max)
	return nil
}

func tileExtent(t *schemaTiles, orientation string) (hex.Axial, hex.Axial) {
	if t.TilesWide <= 0 || t.TilesHigh <= 0 || len(t.TileRows) == 0 {
		return hex.Axial{}, hex.Axial{}
	}
	min := wxxio.TileCoord(orientation, 0, 0)
	last := wxxio.SplitTileLines(t.TileRows[len(t.TileRows)-1].InnerText)
	lastIdx := 0
	if len(last) > 0 {
		lastIdx = len(last) - 1
	}
	max := wxxio.TileCoord(orientation, len(t.TileRows)-1, lastIdx)
	if max.Q < min.Q {
		min.Q, max.Q = max.Q, min.Q
	}
	if max.R < min.R {
		min.R, max.R = max.R, min.R
	}
	return min, max
}

func decodeFeatures(m *ottomap.Map, fs []schemaFeature, orientation string) {
	for _, sf := range fs {
		c, off := wxxio.PixelToAxial(sf.Location.X, sf.Location.Y, orientation)
		clr, _ := wxxio.ParseFloatRGBA(sf.Color)
		f := ottomap.Feature{
			ID:       sf.UUID,
			Kind:     sf.Type,
			Location: c,
			Offset:   off,
			Color:    clr,
			Scale:    sf.Scale,
			Rotation: sf.Rotate,
			Layer:    ottomap.Layer(sf.MapLayer),
			Tags:     wxxio.SplitTags(sf.Tags),
			GMOnly:   sf.IsGMOnly,
		}
		if sf.Label != nil {
			f.Label = strings.TrimSpace(sf.Label.InnerText)
		}
		m.AddFeature(f)
	}
}

func decodeLabels(m *ottomap.Map, ls []schemaLabel, orientation string) {
	for _, sl := range ls {
		c, off := wxxio.PixelToAxial(sl.Location.X, sl.Location.Y, orientation)
		clr, _ := wxxio.ParseFloatRGBA(sl.Color)
		outlineColor, _ := wxxio.ParseFloatRGBA(sl.OutlineColor)
		l := ottomap.Label{
			Text:     strings.TrimSpace(sl.InnerText),
			Location: c,
			Offset:   off,
			Scope:    wxxio.ScopeFor(sl.IsContinent, sl.IsKingdom, sl.IsProvince),
			Font: ottomap.FontSpec{
				Family: sl.FontFace,
				Size:   sl.Location.Scale,
				Bold:   sl.IsBold,
				Italic: sl.IsItalic,
			},
			Color:    clr,
			Outline:  ottomap.Outline{Color: outlineColor, Size: sl.OutlineSize},
			Rotation: sl.Rotate,
			Layer:    ottomap.Layer(sl.MapLayer),
			Tags:     wxxio.SplitTags(sl.Tags),
			GMOnly:   sl.IsGMOnly,
		}
		m.AddLabel(l)
	}
}

func decodeNotes(m *ottomap.Map, ns []schemaNote) {
	for _, sn := range ns {
		m.AddNote(ottomap.Note{
			ID:     sn.Key,
			Title:  sn.Title,
			Body:   sn.NoteText,
			GMOnly: sn.IsGMOnly,
		})
	}
}
