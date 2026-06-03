# TODO

Follow-up work that is not blocking but should be revisited.

## Round-trip tests for real sample files

`TestWriteThenRead` builds a map from scratch and round-trips it. Add a test
that performs `Read → Write → Read` against each `testdata/input/*.wxx` file
so encoder/decoder asymmetries on real-world content are caught early.

## Shapes

Worldographer `<shape>` elements (polylines for rivers, roads, borders, etc.)
are skipped in both directions. They need a domain type — likely something
like:

```go
type Shape struct {
    ID     string
    Style  string         // opaque, e.g. "River", "Road"
    Points []hex.Axial    // or sub-tile pixel offsets
    Layer  Layer
    Tags   []string
}
```

Once defined, add reader/writer support to `wog/v2017` and `wog/v2025`.

## Pixel ↔ axial precision

`wog/internal/wxxio.PixelToAxial` / `AxialToPixel` use a fixed 40-pixel cell.
A round-tripped pixel position is reconstructed within that cell. For maps
authored entirely through the ottomap API this is exact, but a `.wxx`
authored in Worldographer with non-default `hexWidth`/`hexHeight` will round
features to the nearest 40-pixel grid intersection.

Two ways to fix when a use case demands it:

1. Add `Map.HexSize` (or per-call options on `wog.Read`/`wog.Write`) so the
   conversion is exact.
2. Keep the on-disk pixel position verbatim on `Feature`/`Label` (as a
   "raw position" optional field) and reconstruct the axial anchor only when
   asked.

## v2017 schema variations

The three v2017 sample files (1.73, 1.74, 1.77) all decode cleanly with the
current single schema struct. If a later file reveals a 1.x-only attribute we
need, the cheapest fix is to relax the v2017 schema struct (Go's encoding/xml
silently ignores unknown attributes — but missing required attributes parse
as zero values, so confirm behavior).

## More expressive `wog.WriteOptions`

Currently `WriteOptions` exposes `Version`, `Orientation`, `Projection`,
`GMView`, `Compress`, `UTF16BE`. We may eventually want:

- `HexWidth` / `HexHeight` (pairs with the pixel-precision item above).
- `MapName` override that becomes the Worldographer title.
- A way to choose between Worldographer's odd-q and even-q offset
  (currently hardcoded odd-q).
