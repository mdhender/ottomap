package wog

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/wog/v2017"
	"github.com/mdhender/ottomap/wog/v2025"
)

// Read parses a Worldographer .wxx file from r and returns the resulting
// domain model along with the detected major Version.
//
// The returned Version is informational: callers that want to write the map
// back out are expected to pick a target version explicitly; nothing about
// the on-disk metadata is preserved verbatim.
func Read(r io.Reader) (*ottomap.Map, Version, error) {
	xmlBytes, err := readPipeline(r)
	if err != nil {
		return nil, 0, err
	}
	v, err := detectVersion(xmlBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("detect version: %w", err)
	}
	switch v {
	case V2017:
		m, err := v2017.Decode(xmlBytes)
		if err != nil {
			return nil, V2017, fmt.Errorf("v2017 decode: %w", err)
		}
		return m, V2017, nil
	case V2025:
		m, err := v2025.Decode(xmlBytes)
		if err != nil {
			return nil, V2025, fmt.Errorf("v2025 decode: %w", err)
		}
		return m, V2025, nil
	default:
		return nil, 0, fmt.Errorf("unsupported schema")
	}
}

// detectVersion scans the XML until it finds the root <map> element and
// classifies it. The 2025 schema family carries an explicit release="2025"
// attribute; the 2017 schema family does not, and is identified by a
// version attribute beginning with "1." (and the absence of release).
func detectVersion(xmlBytes []byte) (Version, error) {
	dec := xml.NewDecoder(bytes.NewReader(xmlBytes))
	for {
		tok, err := dec.Token()
		if err != nil {
			return 0, fmt.Errorf("read token: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "map" {
			return 0, fmt.Errorf("expected <map> root, got <%s>", se.Name.Local)
		}
		var release, version string
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "release":
				release = a.Value
			case "version":
				version = a.Value
			}
		}
		switch {
		case release == "2025":
			return V2025, nil
		case release == "2017":
			return V2017, nil
		case release == "" && len(version) >= 2 && version[:2] == "1.":
			return V2017, nil
		}
		return 0, fmt.Errorf("cannot classify <map release=%q version=%q>", release, version)
	}
}
