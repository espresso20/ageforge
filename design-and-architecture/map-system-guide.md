# Map System Developer Guide

This guide covers the procedural city renderer used for the Map tab and the dashboard minimap.
All code lives in `ui/mapgen.go`, `ui/tab_map.go`, and `ui/minimap.go`.

---

## 1. System Overview

The map system generates a procedural pixel-art city image from the current game state:
which buildings exist, how many of each, and the player's current age.
The output is an `*image.RGBA` that is handed to a `tview.Image` widget for display.

### TrueColor half-block rendering

tview renders images using Unicode half-block characters (`▄`, `▀`).
Each terminal cell is 1 pixel wide and 2 pixels tall in image space.
This means:

- A 200-column terminal area requires a 200-pixel-wide image.
- That same area is only 100 terminal rows tall, but the image height should be 200 pixels
  (100 rows × 2 px/row) to fill it at full resolution.

The Map tab and minimap both account for this:

```go
// tab_map.go — full-screen map tab
pixW := w       // terminal columns = image pixel width
pixH := ht * 2  // terminal rows × 2 = image pixel height

// minimap.go — dashboard widget (upscaled 2x for more detail)
pixW := w * 2
pixH := ht * 4
```

### Entry points

`GenerateMapImage(cfg MapGenConfig) *image.RGBA` is the single entry point for both the
full-screen map and the minimap. The `DetailLevel` field in `MapGenConfig` controls which
rendering path is used:

- `DetailLevel: 1` — full map (called by `MapTab.Refresh`)
- `DetailLevel: 0` — minimap (called by `MiniMap.UpdateState`)

There is no separate `GenerateMinimapImage` function. Both callers use `GenerateMapImage`
with different `DetailLevel` values and different pixel dimensions.

### MapGenConfig fields

```go
type MapGenConfig struct {
    Width, Height int                        // image dimensions in pixels
    DetailLevel   int                        // 0 = minimap quality, 1 = full map quality
    Buildings     map[string]game.BuildingState // all building state from game engine
    AgeKey        string                     // current age key, e.g. "medieval_age"
}
```

| Field | Description |
|-------|-------------|
| `Width` | Image width in pixels (equals terminal column count for the full map) |
| `Height` | Image height in pixels (equals terminal row count × 2 for the full map) |
| `DetailLevel` | Controls road width, tree radius, building size, sprite scale, wonder glow |
| `Buildings` | All `BuildingState` values; only `Unlocked == true && Count > 0` are drawn |
| `AgeKey` | Used to select the era integer, palette, era name, and RNG seed |

---

## 2. Rendering Pipeline

`GenerateMapImage` draws the image in nine sequential stages.
Each stage paints over what came before, so later stages sit visually on top.

| Stage | Function / inline code | What it draws |
|-------|------------------------|---------------|
| 1 | inline loop | Base terrain: every pixel is colored using two octaves of hash-based noise, producing ground and hill variation |
| 2 | inline switch | Era-specific background: stars + nebulae (space/cosmic), grid lines (cyberpunk), circuit traces (digital) |
| 3 | inline loop | River: a sinusoidal strip of water pixels running vertically through the left-center of the map (skipped for space/cosmic) |
| 4 | `drawVegetation` | Tree canopies scattered by noise density; density decreases each era; absent from digital era onward |
| 5 | `collectBuildingPlacements` | Determines (x, y) positions for all buildings using the era layout algorithm; returns a sorted `[]bldInfo` |
| 6 | `drawSurroundings` | Per-building context rings: farmland (early eras), soot ground (industrial), parking lots (modern) |
| 7 | `drawInfrastructure` | Roads, railways, highways, or neon trails connecting each building to the city center |
| 8 | `drawBuildings` | Pixel-art sprites for every building, centred on its placement position; wonders also get a glow corona |
| 9 | `drawDecorations` | Era-specific overlays: smokestacks (industrial), power lines (modern/digital), holographic billboards (cyberpunk) |

The placement sort in stage 5 orders buildings furthest-from-center first so that buildings
closer to the viewer (city center) are drawn on top in stages 8 and 9.

---

## 3. Terrain and Palettes

### TerrainPalette fields

```go
type TerrainPalette struct {
    Ground, GroundAlt            color.RGBA // base land color and noise-blended variant
    Water, WaterLight, WaterDeep color.RGBA // river: center, edge, deep
    Tree, TreeDark, TreeLight    color.RGBA // canopy shading from south to north
    Road, RoadEdge               color.RGBA // road fill and kerb/edge color
    Hill, HillLight              color.RGBA // elevated terrain colors
    Farmland, FarmAlt            color.RGBA // crop-row alternation colors
    Accent1, Accent2             color.RGBA // era-specific: campfire, neon, energy glow, etc.
}
```

### Era-to-palette selection

The era integer is computed from the age's `Order` field by `eraFromAge(ageKey string) int`,
and `getTerrainPalette(era int)` returns the matching palette. The mapping is:

| Era | Name | Age orders |
|-----|------|-----------|
| 0 | primitive | order <= 1 |
| 1 | ancient | order <= 4 |
| 2 | medieval | order <= 7 |
| 3 | industrial | order <= 10 |
| 4 | modern | order <= 13 |
| 5 | digital | order <= 15 |
| 6 | cyberpunk | order == 16 |
| 7 | space | order <= 19 |
| 8 | cosmic | order > 19 |

### How to add or modify a palette

Palettes are returned by `getTerrainPalette` in `mapgen.go`. To add a new era or change
existing colors, edit the corresponding case. The helper `c(r, g, b)` builds a fully-opaque
`color.RGBA`:

```go
case 9: // hypothetical post-cosmic era
    return TerrainPalette{
        Ground: c(20, 5, 30), GroundAlt: c(28, 8, 40),
        Water: c(80, 0, 180), WaterLight: c(100, 10, 200), WaterDeep: c(55, 0, 130),
        Tree: c(0, 0, 0), TreeDark: c(0, 0, 0), TreeLight: c(0, 0, 0),
        Road: c(120, 80, 30), RoadEdge: c(90, 60, 20),
        Hill: c(30, 10, 50), HillLight: c(40, 15, 65),
        Farmland: c(0, 0, 0), FarmAlt: c(0, 0, 0),
        Accent1: c(255, 220, 80), Accent2: c(180, 80, 255),
    }
```

### Noise functions

Terrain variation uses two noise calls:

- `noise2D(x, y, seed)` — fine-grain noise (full pixel resolution), blended 60/40 from two
  FNV-64a hashes; drives ground color variation and tree placement threshold.
- `noise2D(x/4, y/4, seed+200)` — coarse noise (1/4 resolution) drives hill elevation.

Both functions are deterministic given the same `AgeKey`-derived seed.

---

## 4. Era System

### Era integer constants

There are no named constants in the code; the era is an `int` computed inline:

| Value | Informal name | Used by |
|-------|---------------|---------|
| 0 | primitive | palette, layout, decorations |
| 1 | ancient | palette, layout |
| 2 | medieval | palette, layout |
| 3 | industrial | palette, layout, decorations (smokestacks, railways) |
| 4 | modern | palette, layout, decorations (power lines) |
| 5 | digital | palette, layout, decorations (power lines) |
| 6 | cyberpunk | palette, layout, decorations (neon billboards) |
| 7 | space | palette, layout, background (stars) |
| 8 | cosmic | palette, layout, background (stars + nebulae) |

### getEraName mapping

`getEraName(ageKey string) string` maps age keys to era name strings used in layout
dispatch and sprite selection:

```go
func getEraName(ageKey string) string {
    switch ageKey {
    case "primitive_age":                            return "primitive"
    case "stone_age":                                return "stone"
    case "bronze_age":                               return "bronze"
    case "iron_age":                                 return "iron"
    case "classical_age":                            return "classical"
    case "medieval_age":                             return "medieval"
    case "renaissance_age":                          return "renaissance"
    case "colonial_age":                             return "colonial"
    case "industrial_age", "victorian_age":          return "industrial"
    case "electric_age":                             return "industrial"
    case "atomic_age":                               return "atomic"
    case "modern_age", "information_age":            return "modern"
    case "digital_age":                              return "digital"
    case "cyberpunk_age", "fusion_age":              return "nano"
    case "space_age", "interstellar_age":            return "space"
    case "galactic_age", "quantum_age", "transcendent_age": return "galactic"
    default:                                         return "stone"
    }
}
```

### Era controls both layout and palette

The era integer (from `eraFromAge`) drives `getTerrainPalette` and the stage-2 background.
The era name string (from `getEraName`) drives `placeAllBuildings` (layout algorithm
selection) and `getBuildingSprite` (sprite type selection).
These two paths are independent: `eraFromAge` uses the age `Order` field; `getEraName`
uses the age key string directly.

### How to add a new era

1. Add a new case to `getTerrainPalette` for the new era integer.
2. Update `eraFromAge` to return the new integer for the relevant age orders.
3. Add a new case to `getEraName` for any new age keys.
4. Add the era name to the `placeAllBuildings` switch if it needs a distinct layout.
5. Update `drawVegetation`, `drawInfrastructure`, and `drawDecorations` switches if the
   era needs distinct vegetation, road type, or decorations.

---

## 5. City Layout Algorithms

Buildings are positioned by `collectBuildingPlacements`, which separates wonders (placed
at fixed angles in the outer 60-80% zone) from ordinary buildings (passed to
`placeAllBuildings`). The latter dispatches to one of seven layout functions based on
`eraName`.

### plotGrid collision system

All layout functions use a shared `plotGrid` to prevent buildings from overlapping.

```go
type plotGrid struct {
    occupied map[[2]int]bool
    cellSize int
}
```

- `cellSize` sets the grid granularity in pixels. Larger values create more spacing between
  buildings because each building claim covers more cells.
- `isFree(px, py, w, h int) bool` checks a 2-cell padding around the bounding box — the
  effective exclusion zone is the building footprint plus `cellSize*2` pixels on each side.
- `claim(px, py, w, h int)` marks the cells covered by a building as occupied.

To increase building spacing, raise `cellSize` in the layout function's `newPlotGrid` call.
To decrease spacing (pack buildings tighter), lower it. Typical values are 8–14.

```go
// Typical call pattern in every layout function:
grid := newPlotGrid(12)   // adjust this number to change density
// ...
if grid.isFree(bx, by, size, size) {
    grid.claim(bx, by, size, size)
    placements = append(placements, bldInfo{...})
}
```

### placeBuildingsOrganic (eras: primitive, stone)

`cellSize: 12`, `maxR: 40% of min(w,h)`.

Buildings are placed using a random-walk anchor system. The first building goes at a random
point within `maxR` of the center. Subsequent buildings have a 2-in-3 chance of being placed
near an existing anchor (within ±15 px horizontal, ±10 px vertical), simulating organic
settlement growth. Water pixels are rejected. Up to 30 placement attempts per building.

Key parameters: `maxR = int(float64(min(w, h)) * 0.40)`, anchor jitter `rng.Intn(31)-15` / `rng.Intn(21)-10`.

### placeBuildingsVillage (eras: bronze, iron, classical)

`cellSize: 10`, `maxR: 35% of min(w,h)`, 6 radial spokes.

Six road spokes radiate from the city center. Buildings are assigned to spokes round-robin
and placed at a random distance along the spoke with a small perpendicular offset (3–10 px).
This creates a hub-and-spoke settlement pattern typical of ancient market towns.

Key parameters: `numSpokes = 6`, spoke color is a blend of `pal.Road` and `pal.Ground`,
offset `rng.Intn(8)+3`.

### placeBuildingsMedieval (eras: medieval, renaissance)

`cellSize: 10`, `wallR: 30 px`.

A roughly square castle wall is drawn at radius 30 from center. Four cardinal roads lead
outward. Buildings are divided into four quarter-groups and placed in a 4-column grid within
each quarter starting just outside the wall. Grid wraps at 4 columns per row.

Key parameters: `wallR = 30`, `spacing = 10`, quarter grid starts at `cx ± (wallR+10)`.

### placeBuildingsIndustrial (eras: colonial, industrial)

`cellSize: 14`, `cellSpacing: 14`.

Production-domain buildings (`food`, `lumber`, `masonry`, `metallurgy`, `energy`) are
separated onto the left half; all other buildings go on the right. A road grid is drawn
with horizontal roads every 3 cells and vertical roads every 4 cells. Each group fills a
5-column grid expanding from `cx ± cellSpacing`.

Key parameters: `cellSpacing = 14`, grid roads at `cellSpacing*3` and `cellSpacing*4` intervals.

### placeBuildingsModern (eras: atomic, modern)

`cellSize: 12`, city block system.

Buildings are packed into city blocks of `blockW=6` columns and `blockH=4` rows with
`cellSize=10` pixels per slot and `streetW=3` pixel gutters between blocks. Streets are
drawn around each block as it is filled. Blocks advance left-to-right within `cityR =
42% of min(w,h)`, then wrap to the next row. This produces a grid-plan city layout.

Key parameters: `blockW = 6`, `blockH = 4`, `cellSize = 10`, `streetW = 3`.

### placeBuildingsCampus (eras: digital, nano)

`cellSize: 8`, cluster-based hexagonal arrangement.

Buildings are grouped into clusters of 8. The first cluster is centered on the map center;
subsequent clusters are placed at hexagonal grid positions at `clusterR = 32% of min(w,h)`.
Within each cluster, buildings orbit the cluster center at increasing radii. Winding paths
connect adjacent cluster centers.

Key parameters: `clusterSize = 8`, `clusterR = int(float64(min(w, h)) * 0.32)`,
per-building orbit distance `5 + i%3*clusterCellSize` where `clusterCellSize = 8`.

### placeBuildingsOrbital (eras: space, galactic)

`cellSize: 10`, three orbital rings at radii 25, 45, and 70 pixels.

The first building in the list is placed at the exact center as a hub. Remaining buildings
are distributed across three concentric rings drawn as dotted ellipses (0.70 vertical
compression). Buildings within each ring are equally spaced by angle. Failed placements
are retried with small angular offsets.

Key parameters: `ringRadii = []int{25, 45, 70}`, ellipse y-compression factor `0.70`.

---

## 6. Building Sprites

### The spriteType enum

```go
type spriteType int

const (
    spriteHut          spriteType = iota // thatched hut, primitive housing
    spriteFarm                           // striped crop field
    spriteMill                           // peaked mill building
    spriteLumberCamp                     // tree-stump silhouette
    spriteMine                           // open-pit entrance
    spriteFortress                       // crenellated fort
    spriteBarracks                       // long rectangular barracks
    spriteTemple                         // peaked temple with colonnade
    spriteLibrary                        // rectangular building with portico
    spriteMarket                         // awning-topped stall row
    spriteFactory                        // industrial block with chimney stubs
    spriteWorkshop                       // small workshop
    spritePalace                         // (alias: not explicitly listed, maps to spriteMill default)
    spriteObservatory                    // domed observatory
    spriteDome                           // energy reactor dome
    spriteSkyscraper                     // tall tower with antenna
    spriteServer                         // server rack stack
    spriteSpaceStation                   // cross-shaped station with solar panels
    spriteWonder                         // large ornate monument
)
```

### getBuildingSprite — sprite selection logic

`getBuildingSprite(domain, buildingKey, eraName string) spriteType` applies rules in order:

1. Wonder keys are checked first by building key — any wonder always returns `spriteWonder`.
2. Three era-group booleans are set: `isEarlyEra` (primitive through classical),
   `isLateEra` (space, galactic, nano), `isDigitalEra` (digital, nano).
3. A switch on `domain` maps each worker domain to a sprite, with era overrides.
4. A fallback switch on `buildingKey` handles specific non-domain buildings
   (skyscrapers, reactors, space stations, etc.).
5. Default: `spriteHut`.

```go
// Domain-to-sprite summary:
// "food"             → spriteHut (early) / spriteFarm (later)
// "lumber"           → spriteLumberCamp
// "masonry"          → spriteMine
// "military"         → spriteBarracks (early) / spriteFortress (later)
// "knowledge"        → spriteObservatory (late) / spriteServer (digital) / spriteLibrary
// "faith"            → spriteTemple
// "trade"            → spriteMarket
// "engineering","metallurgy" → spriteWorkshop (early) / spriteFactory (later)
// "energy"           → spriteDome (late) / spriteWorkshop
// "hacker"           → spriteServer
// "astronaut"        → spriteSpaceStation
```

### drawBuildingSprite — scale behavior

`drawBuildingSprite(img, imgW, imgH, px, py, stype, primary, accent, scale)` renders the
pixel-art pattern from `spriteRows`:

- `scale <= 1` (minimap): draws a solid 3×3 block of `primary` color. No sprite detail.
- `scale == 2` (full map, regular building): each pixel in the pattern becomes a 2×2 block.
- `scale == 3` (full map, wonder): each pixel becomes a 3×3 block.

Scale is chosen in `drawBuildings`:

```go
scale := 1
if dl > 0 {
    if b.category == "wonder" {
        scale = 3
    } else {
        scale = 2
    }
}
```

### How to add a new sprite

Three changes are required:

**Step 1 — Add a constant to the `spriteType` block:**

```go
const (
    spriteHut spriteType = iota
    // ... existing constants ...
    spriteWonder
    spriteWindmill  // <-- add here, after the last existing constant
)
```

**Step 2 — Add a case in `getBuildingSprite`:**

```go
// Inside the domain switch, or as a buildingKey fallback:
case "windmill_key":
    return spriteWindmill
```

Or inside the domain switch:

```go
case "lumber":
    if eraName == "colonial" || eraName == "industrial" {
        return spriteWindmill
    }
    return spriteLumberCamp
```

**Step 3 — Add a pixel pattern in `spriteRows`:**

```go
case spriteWindmill:
    return []string{
        "..A..",   // top sail
        ".A.A.",   // upper arms
        "PPPPP",   // body top
        "P.P.P",   // body middle
        "PPPPP",   // body base
        "..A..",   // base post
    }
```

### Pattern character reference

| Character | Meaning |
|-----------|---------|
| `P` | Primary color (main building material) |
| `A` or `I` | Accent color (roof, trim, detail, glow) |
| `.` | Transparent — pixel is not drawn |

Each row is a string of characters. All rows should be the same width or the widest row
determines the sprite width. The sprite is centered on the placement point.

---

## 7. Building Colors (buildingVisuals)

### BuildingVisual struct

```go
type BuildingVisual struct {
    Shape   BuildingShape  // geometric shape template (used by drawBuildingShape, a legacy path)
    Primary color.RGBA     // main body color
    Accent  color.RGBA     // roof, trim, windows, or glow color
}
```

`Shape` is still defined and has a complete set of shape functions (`shapeCircle`,
`shapeSquare`, etc.) called via `testShape` and `drawBuildingShape`. However, the active
rendering path in `drawBuildings` uses `drawBuildingSprite` (the pixel-art sprite system),
not `drawBuildingShape`. The `Shape` field is currently unused at runtime but is available
if you want to add a shape-based rendering path for a new era or mode.

### The buildingVisuals map

`var buildingVisuals map[string]BuildingVisual` in `mapgen.go` contains an entry for every
named building key (284 buildings). Keys match `config/buildings_lineage_*.go` building keys.

To look up or edit a building's colors:

```go
// Example entry:
"cathedral": {ShapeCross, c(160, 155, 140), c(200, 170, 40)},
//             ^shape      ^primary gray    ^accent gold
```

Searching by building key in `mapgen.go` (around line 219) will find the full map.
The fallback `getBuildingVisual` function returns category-based defaults for any key
not in the map.

### Pattern for overriding colors

To change a building's colors, find its key in `buildingVisuals` and edit the `Primary`
and `Accent` color.RGBA values using the `c(r, g, b)` helper:

```go
// Before:
"server_farm": {ShapeSquare, c(30, 50, 100), c(40, 160, 60)},

// After (make it look more electric):
"server_farm": {ShapeSquare, c(20, 30, 80), c(0, 220, 180)},
```

---

## 8. Adding a New Building to the Map

When you add a building in `config/buildings_lineage_*.go`, the map will render it
immediately using the category fallback in `getBuildingVisual` — no required changes.
However, for distinct colors and a fitting sprite, do the following:

### Step 1 — Add a BuildingVisual entry

In `mapgen.go`, inside the `buildingVisuals` map, add an entry for your building key:

```go
// In the appropriate category comment group:
"tidal_generator": {ShapeWide, c(60, 100, 140), c(80, 180, 220)},
//                  ^shape     ^steel blue      ^water teal accent
```

The shape is cosmetically unused in the current rendering path but documents intent and
ensures forward compatibility if the shape-based path is ever wired in.

### Step 2 — Optionally add a getBuildingSprite case

If the building's domain is already handled by `getBuildingSprite`, no change is needed —
it will inherit the domain's sprite. If you want a specific sprite (e.g. the building is a
new domain or a key-level override), add a case to the `buildingKey` fallback switch:

```go
case "tidal_generator":
    return spriteDome
```

Or add a new `spriteType` constant and pattern following the three-step process in Section 6.

### Step 3 — No other changes needed

`collectBuildingPlacements` iterates `cfg.Buildings` and skips entries where
`bs.Unlocked == false || bs.Count == 0`. As soon as a building is built in game,
it appears on the map at the next `Refresh` or `UpdateState` call. The layout algorithm,
road connections, and wonder placement all handle new buildings automatically.

---

## 9. Minimap Differences

The minimap (`ui/minimap.go`) uses the same `GenerateMapImage` function with two differences:

### DetailLevel: 0

Setting `DetailLevel: 0` (vs `1` for the full map) has these effects throughout the pipeline:

- River width: `5 + dl*4` → 5 px (vs 9 px at dl=1)
- River bank width: `2 + dl` → 2 px (vs 3 px)
- Tree radius: `2 + dl` → 2 px (vs 3 px)
- Building base size: starts at 3 (vs 5 for high eras at dl=1)
- Wonder size: `7 + dl*3` → 7 px (vs 10 px)
- Road width in `drawRoad`: `1 + dl` → 1 px
- Highway width: `2 + dl` → 2 px
- Shadow offset: `1 + dl` → 1 px
- Sprite scale: `dl > 0` is false → `scale = 1` → solid 3×3 blocks instead of detailed sprites

### Sprite rendering at scale 1

At `scale <= 1`, `drawBuildingSprite` skips the pattern lookup entirely and paints a
solid 3×3 block using `primary` color. This is intentional: at minimap resolution,
a 5×7 sprite scaled to 2×2 terminal cells is indistinguishable from a solid dot,
and the solid block is faster to render.

### Pixel resolution

The minimap doubles both dimensions for more density:

```go
pixW := w * 2   // 2x terminal width
pixH := ht * 4  // 2x terminal height (already doubled for half-blocks)
```

The full map tab uses `pixW = w` and `pixH = ht * 2`.

---

## 10. Quick Reference — Common Tweaks

| Goal | What to change |
|------|----------------|
| Change a building's color | `buildingVisuals` map in `mapgen.go` |
| Make buildings larger | Raise `baseSize` in `collectBuildingPlacements` or raise `wsz` for wonders |
| Change city spacing | `cellSize` in `newPlotGrid(...)` inside the relevant layout function |
| Change which era uses which layout | `placeAllBuildings` switch in `mapgen.go` |
| Add a new city layout style | New `placeBuildingsXxx` function + case in `placeAllBuildings` |
| Change terrain colors | `TerrainPalette` struct literal for the relevant era in `getTerrainPalette` |
| Add a new age key mapping | `getEraName` switch in `mapgen.go` |
| Adjust wonder size | `wsz := 7 + dl*3` in `collectBuildingPlacements` |
| Change orbital ring radii | `ringRadii := []int{25, 45, 70}` in `placeBuildingsOrbital` |
| Change road color | `pal.Road` / `pal.RoadEdge` in the palette, or color literals in `drawInfrastructure` |
| Add a new sprite | Three steps: constant in `spriteType`, case in `getBuildingSprite`, pattern in `spriteRows` |
| Change minimap resolution | `pixW = w * 2` / `pixH = ht * 4` multipliers in `minimap.go` |
| Suppress a decoration type | Remove or gate the relevant case in `drawDecorations` |
| Change vegetation density | `density` variable per era in `drawVegetation` |
