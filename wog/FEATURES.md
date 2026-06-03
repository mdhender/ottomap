# Worldographer 2025 (V2025) Feature Tracking

Status of each Worldographer 2025 schema feature in the `wog/v2025` adapter.

Legend:
- ✅ implemented (read and/or write as noted)
- ⚠️ partial — implemented with documented caveats
- ❌ not yet implemented

| Feature                                                                                                              | Read | Write | Notes                                                           |
|----------------------------------------------------------------------------------------------------------------------|:----:|:-----:|-----------------------------------------------------------------|
| `<map>` root element + dispatch                                                                                      |  ✅   |   ✅   | Writer regenerates release/version/schema                       |
| `hexOrientation` (COLUMNS / ROWS)                                                                                    |  ✅   |   ✅   | Picked by `WriteOptions.Orientation`                            |
| `mapProjection` (FLAT / ICOSAHEDRAL)                                                                                 |  ✅   |   ✅   | Picked by `WriteOptions.Projection`                             |
| Map-level UI flags (`showNotes`, `showGrid`, `triangleSize`, etc.)                                                   |  ⚠️  |  ⚠️   | Decoded but dropped from domain; writer emits sensible defaults |
| Map-level factors (`continentFactor`, view offsets, `hexWidth`/`hexHeight`)                                          |  ⚠️  |  ⚠️   | Decoded but dropped; writer emits defaults                      |
| `<gridandnumbering>` (30 attributes)                                                                                 |  ⚠️  |  ⚠️   | Reader ignores, writer emits Worldographer defaults             |
| `<terrainmap>` (slot table)                                                                                          |  ✅   |   ✅   | Built lazily from terrains actually used                        |
| `<maplayer>` (with `opacity`)                                                                                        |  ✅   |   ✅   | Standard layer set emitted on write                             |
| `<tiles>` / `<tilerow>` tile grid                                                                                    |  ✅   |   ✅   | Honors `Map.SetBounds`; otherwise derived                       |
| Tile `terrain` (slot index)                                                                                          |  ✅   |   ✅   | Mapped through `Terrain` string                                 |
| Tile `elevation`                                                                                                     |  ✅   |   ✅   | Stored as float; written as truncated int                       |
| Tile `isIcy` / `isGMOnly`                                                                                            |  ✅   |   ✅   |                                                                 |
| Tile resources (full 6-resource form)                                                                                |  ✅   |   ✅   |                                                                 |
| Tile resources (`Z`-sentinel compressed form)                                                                        |  ✅   |   ✅   | Writer auto-compresses when non-Animal resources are all zero   |
| Tile `customBackgroundColor` (optional float-rgba)                                                                   |  ✅   |   ✅   | `color.RGBA` zero value = no override                           |
| `<features>` / `<feature>`                                                                                           |  ✅   |   ✅   | Feature `Kind` is an opaque string                              |
| Feature `<location>` (x/y, no scale)                                                                                 |  ⚠️  |  ⚠️   | Pixel→axial uses a 40-px cell approximation                     |
| Feature inline `<label>` (optional)                                                                                  |  ✅   |   ✅   | Stored as `Feature.Label` text                                  |
| Feature `tags` (comma-separated)                                                                                     |  ✅   |   ✅   | Split into `[]string` on read                                   |
| Feature `color`, `ringcolor`                                                                                         |  ✅   |  ⚠️   | `ringcolor` always emitted as `null`                            |
| Feature scope flags (`isWorld`/`isContinent`/`isKingdom`/`isProvince`)                                               |  ⚠️  |  ⚠️   | Writer emits all-true; domain doesn't track per-feature scope   |
| `<labels>` / `<label>` (standalone)                                                                                  |  ✅   |   ✅   |                                                                 |
| Label `<location>` with `scale`                                                                                      |  ✅   |   ✅   | Scale stored in `Label.Font.Size`                               |
| Label `style` (named text style)                                                                                     |  ❌   |   ❌   | Always emitted as empty string                                  |
| Label `backgroundColor`                                                                                              |  ⚠️  |   ❌   | Decoded but not preserved on the domain `Label`                 |
| Label scope flags                                                                                                    |  ✅   |   ✅   | Maps to `Label.Scope` enum                                      |
| `<notes>` / `<note>`                                                                                                 |  ✅   |   ✅   | Full CDATA body preserved (current `wxx` package loses this)    |
| Note location (x, y)                                                                                                 |  ⚠️  |  ⚠️   | Anchored to a hex via 40-px cell approximation                  |
| `<shapes>` / `<shape>`                                                                                               |  ❌   |   ❌   | Needs a `Shape` domain type — tracked in TODO.md                |
| `<informations>` (lore tree)                                                                                         |  ❌   |   ❌   | Reader ignores; writer emits empty element                      |
| `<mapkey>`                                                                                                           |  ❌   |   ❌   | Reader ignores; writer emits Worldographer default block        |
| `<configuration>` sub-elements (`terrain-config`, `feature-config`, `texture-config`, `text-config`, `shape-config`) |  ❌   |   ❌   | Reader ignores; writer emits empty placeholders                 |
| `<blurTerrainBG>`                                                                                                    |  ❌   |   ❌   | Reader ignores; writer omits                                    |
| `<extraTerrain>`                                                                                                     |  ❌   |  ⚠️   | Writer emits empty element to keep file shape consistent        |

## Notes

- "Decoded but dropped" means the XML parser accepts the attribute (so
  unknown attributes never break a read) but the value is not surfaced on
  the `ottomap.Map` domain object. This is the "creator-first" trade-off:
  callers wanting fidelity for these fields should request follow-up work.
- The pixel↔axial caveat is shared with v2017 and is tracked in TODO.md.
