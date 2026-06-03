package wog

import (
	"fmt"
	"io"

	"github.com/mdhender/ottomap"
	"github.com/mdhender/ottomap/wog/v2017"
	"github.com/mdhender/ottomap/wog/v2025"
)

// Write serializes m to w as a Worldographer .wxx file.
//
// opts.Version is required. The writer regenerates schema/version metadata
// for the target; nothing from a previously-read file is preserved verbatim.
func Write(w io.Writer, m *ottomap.Map, opts WriteOptions) error {
	if m == nil {
		return fmt.Errorf("nil map")
	}
	var (
		xmlBytes []byte
		err      error
	)
	switch opts.Version {
	case V2017:
		xmlBytes, err = v2017.Encode(m, v2017.Options{
			Orientation: int(opts.Orientation),
			Projection:  int(opts.Projection),
			GMView:      opts.GMView,
		})
	case V2025:
		xmlBytes, err = v2025.Encode(m, v2025.Options{
			Orientation: int(opts.Orientation),
			Projection:  int(opts.Projection),
			GMView:      opts.GMView,
		})
	case 0:
		return fmt.Errorf("WriteOptions.Version is required")
	default:
		return fmt.Errorf("unsupported version %d", opts.Version)
	}
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return writePipeline(xmlBytes, opts, w)
}
