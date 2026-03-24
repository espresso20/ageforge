# Map Rendering Experiments

Three parallel map rendering experiments, each driven by the same `GameState`,
accessible via separate overlay commands (`mapv1`, `mapv2`, `mapv3`).
Goal: determine which approach produces the best visual result for a permanent
"civilization map" feature before committing to a final implementation.

---

## Shared Principles (all three versions)

- **Same data source**: all three read `game.GameState` — buildings, age, resources, wonders
- **Same overlay system**: registered via `overlayMgr.Register("mapv1", ...)` etc.
- **Same refresh cadence**: re-rendered on overlay open + on every engine tick while open
- **No shared code between versions**: each is self-contained in its own file(s)
- **No permanent UI changes**: these are overlays, not new tabs. Player types `mapv1`, `mapv2`, or `mapv3`

---

## mapv1 — Character Native (Civ 1 Authentic)

**Philosophy**: Don't fight the terminal. Every cell is a `(rune, fg, bg)` triplet — draw
directly via tcell. Zero image conversion, zero rendering artifacts. This is structurally
identical to how Civilization 1 (DOS, 1991) rendered its map.

**Files**: `ui/mapv1.go`

### Rendering approach

Use a `tview.Box` with `SetDrawFunc` to get raw tcell screen access:

```go
box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
    for ty := 0; ty < h; ty++ {
        for tx := 0; tx < w; tx++ {
            r, style := mapTile(state, tx, ty, w, h)
            screen.SetContent(x+tx, y+ty, r, nil, style)
        }
    }
    return box.GetInnerRect()
})
```

Each tile is computed by `mapTile(state, tx, ty, mapW, mapH)` which returns a rune and
a tcell.Style with explicit fg + bg color. The map fills the entire overlay area.

### Tile system

The map is a logical grid. Each cell maps to one terminal character.
Terrain is determined by a seeded hash of (x, y, civSeed) — same seed every game,
derived from the player's civilization name or a stored seed in GameState.

**Terrain tiles:**

| Terrain     | Rune | BG color (approx)      | FG color     |
|-------------|------|------------------------|--------------|
| Grassland   | `·`  | #4a7c3f (muted green)  | #3a6a2f      |
| Plains      | `·`  | #8a9a4a (yellow-green) | #7a8a3a      |
| Forest      | `♣`  | #2d5a1b (dark green)   | #1d4a0b      |
| Hills       | `n`  | #6b7a3a                | #8a9a5a      |
| Mountains   | `▲`  | #4a3a2a                | #8a7a6a      |
| Ocean       | `~`  | #1a3a6b (deep blue)    | #2a5a9b      |
| Coast       | `~`  | #2a5a8a                | #4a8abf      |
| Desert      | `·`  | #b8a060 (tan)          | #d0b870      |
| Tundra      | `·`  | #7a8a80                | #9aaba0      |
| Snow        | `*`  | #c0c8cc                | #e0e8ec      |
| River       | `~`  | (terrain bg)           | #4a8aff      |
| Road        | `·`  | (terrain bg)           | #a09070      |

**City tiles (single cell, placed at city center):**

| Age group         | Rune | Style                        |
|-------------------|------|------------------------------|
| primitive/stone   | `⌂`  | white on brown               |
| bronze–classical  | `⌂`  | bright white on grey         |
| medieval–colonial | `♜`  | bright white on dark grey    |
| industrial+       | `▣`  | cyan on dark blue            |
| digital+          | `▣`  | bright cyan on near-black    |
| space+            | `◈`  | bright white on black        |

**Building tiles (radiate outward from city center):**

Buildings are placed in concentric rings around the city center.
Each building type maps to a domain-appropriate rune:

| Domain       | Rune | Color hint      |
|--------------|------|-----------------|
| food         | `⌂`  | warm green      |
| lumber       | `♣`  | dark green      |
| masonry      | `▪`  | grey            |
| metallurgy   | `⚒`  | orange-red      |
| energy       | `⚡`  | yellow          |
| military     | `⚔`  | red             |
| knowledge    | `◎`  | light blue      |
| faith        | `✚`  | white/gold      |
| trade        | `$`  | gold            |
| engineering  | `⚙`  | cyan            |
| hacker       | `#`  | bright green    |
| astronaut    | `◆`  | bright white    |
| wonders      | `★`  | gold glow (bold)|

### City growth

The city radius in tiles grows with building count:
- 0–10 buildings: radius 2 (tiny village)
- 11–30: radius 4
- 31–80: radius 6
- 81–180: radius 8
- 181+: radius 10

Buildings are placed deterministically by key hash — same building always appears
in the same relative position so the city doesn't jump on each refresh.

### Age palette shifts

Background terrain colors shift subtly with age to reflect the era:
- primitive–stone: saturated natural greens
- medieval: slightly darker, more muted
- industrial: desaturated, slight grey cast, smoke-colored sky (dark bg)
- digital: dark background, neon-tinted terrain
- space: near-black bg, terrain replaced with grey/crater tiles, stars in empty space

### Layout

```
┌─────────────────────────────────────────────┐
│ ~ ~ ~ ♣ ♣ · · ⌂ · · ♣ ♣ ~ ~ ~            │
│ ~ · · ♣ · ⌂ ⌂ ▣ ⌂ ⌂ · · · ~ ~            │
│ · · · · · ⌂ ★ ◉ ★ ⌂ · n n · ·            │  ← city center (◉)
│ · · n · · ⌂ ⌂ ▣ ⌂ · · n n · ·            │
│ · · n n · · · · · · · · · · ·              │
│ [Medieval Age — Civilization of Echo]       │  ← status line at bottom
└─────────────────────────────────────────────┘
```

### What makes this version good

- Zero image conversion — no flicker, no artifacts
- Scales to any terminal size automatically
- Every character is crisp at all font sizes
- Authentic to the Civ 1 aesthetic the player wants
- Fastest to implement and iterate

---

## mapv2 — Image Generation + pixterm

**Philosophy**: Generate a proper `image.RGBA` with real drawing code, then render it to
the terminal via `ansimage` (pixterm). The difference from the failed previous attempt:
the *source image* will be high quality (proper algorithms, no hand-rolled noise),
and the rendering pipeline will be tested for tview/tcell compatibility first.

**Files**: `ui/mapv2.go`

### Rendering approach

The key open question from the previous failure: **does ansimage output conflict with tview?**

The safe integration path is a `tview.TextView` with `SetDynamicColors(false)` and direct
ANSI string injection. The ansimage `Render()` method returns a string of ANSI escape codes —
this must be tested in isolation before building the full map.

**Spike test (must pass before any other mapv2 work):**

```go
// Render a 40×20 solid-color image through ansimage into a TextView.
// If it displays cleanly with correct colors and no artifacts → proceed.
// If it conflicts with tview's tcell layer → mapv2 is not viable as a tview overlay.
```

If the spike fails, mapv2 will use a standalone `tcell.Screen` layer drawn over tview
for the map area only (more complex but guaranteed to work).

### Image generation pipeline

```
GameState
    │
    ▼
buildMapImage(state, pixW, pixH int) *image.RGBA
    │   ├─ drawTerrain()      — base terrain layer
    │   ├─ drawWater()        — rivers, coast, ocean
    │   ├─ drawVegetation()   — forest/jungle coverage
    │   ├─ drawRoads()        — era-appropriate paths
    │   ├─ drawBuildings()    — building sprites (larger than v1 attempt)
    │   └─ drawLabels()       — city name, wonder names
    │
    ▼
ansimage.NewScaledFromImage(img, termH*2, termW, ...)
    │
    ▼
ansi.Render() → string → tview.TextView.SetText()
```

### Source image quality improvements over previous attempt

The previous attempt failed because the source image was built with basic hand-rolled
noise and tiny 3×3 pixel sprites. mapv2 will use:

**Terrain generation:**
- `github.com/ojrac/opensimplex-go` for multi-octave simplex noise
  (or `github.com/aquilax/go-perlin` — evaluate both)
- Two noise layers: elevation + moisture → biome lookup table
- Smooth bilinear blending at biome borders (no hard edges)

**City/building rendering:**
- Buildings rendered at minimum 12×12 pixels (not 3×3)
- Sprites drawn at 2× scale minimum
- City center rendered at 24×24 with distinct silhouette per age
- No circular layout — rectilinear city blocks for modern eras, organic for early

**Color palette:**
- Civ 1 inspired: flat colors, limited palette, high contrast
- Per-era palette swap (same as mapv1 eras, expressed as RGBA)

### tview integration risk

ansimage produces raw ANSI codes. tview's `SetText` passes strings through tcell which
may re-interpret or strip escape sequences. Known mitigation options:

1. `TextView.SetDynamicColors(false).SetText(ansiStr)` — may work if tcell passes ANSI through
2. Write directly to `os.Stdout` before tview renders — dirty hack, avoid
3. Use a `tview.Box` with `SetDrawFunc` and translate ansimage pixel data to tcell styles manually
   (most reliable — avoids ANSI entirely, uses tcell's native color API)

Option 3 is the fallback if the spike test fails. It means iterating ansimage's pixel
matrix and calling `screen.SetContent` with computed fg/bg colors — same approach as
mapv1 but with the source image being a proper generated `image.RGBA`.

---

## mapv3 — Noise Terrain (Greenfield)

**Philosophy**: Forget city layout for now — focus purely on making the *terrain* look
extraordinary. A proper procedural world map with realistic biomes, elevation,
rivers, and coastlines. Buildings and city overlays added later once terrain is solid.
Rendering method chosen *after* terrain generation works (likely mapv1-style chars or
option 3 from mapv2).

**Files**: `ui/mapv3.go`, `ui/mapv3_terrain.go`

### Terrain generation

Uses a full noise pipeline:

```
1. Elevation map    — simplex noise, 4 octaves, persistence 0.5
2. Moisture map     — simplex noise, different seed/frequency
3. Temperature map  — latitude gradient + noise perturbation
4. Biome lookup     — Whittaker biome diagram (elevation × moisture × temp)
5. River generation — flow from high elevation to coast following steepest descent
6. Erosion pass     — smooth sharp transitions between elevation bands
```

**Biome table (Whittaker-inspired):**

| Elevation | Moisture low | Moisture mid | Moisture high |
|-----------|-------------|--------------|---------------|
| High      | Mountains   | Mountains    | Snow          |
| Mid-high  | Hills/Desert| Hills/Plains | Forest/Hills  |
| Mid       | Desert      | Plains       | Forest        |
| Mid-low   | Plains      | Grassland    | Jungle        |
| Low       | Coast       | Coast        | Swamp         |
| Sea level | Ocean       | Ocean        | Ocean         |

**River generation:**
- Place sources at high-elevation local maxima
- Flow downhill following gradient, add moisture to adjacent cells
- Rivers widen as they flow (1px → 3px at coast)
- Create river deltas at ocean boundary

### World seed

The world is generated once per civilization from a seed derived from the player's
save slot or a stored `MapSeed uint64` in GameState (added to save JSON).
Same seed → same world every session. Age transitions don't regenerate terrain —
they only change the visual palette and what city overlay is drawn on top.

### City placement

Unlike v1/v2 which place the city at map center, v3 places the city at a biome-appropriate
starting location (computed from seed):
- primitive/stone: near river, flat terrain, forest adjacent
- later ages: city grows outward from this same origin point

### Rendering

mapv3 uses the character-native approach from mapv1 (tcell SetContent) because:
1. The terrain detail is in the *data* (elevation, moisture, biome), not the pixel art
2. Character rendering is reliable and immediate
3. The terrain rune+color set can be richer than v1 (more biome types, elevation shading)

A future mapv4 could combine v3 terrain generation with v2 image rendering.

### What makes this version different

- The *world* feels real — it looks like an actual map, not a generated game board
- Terrain persists across ages (same geography, evolving city)
- Rivers, coastlines, and elevation create natural strategic geography
- This is the most ambitious version and should be built last

---

## Implementation Order

```
mapv1  →  mapv2 spike test  →  mapv2 full  →  mapv3
  │              │                              │
  │         if spike fails:                     │
  │         mapv2 uses tcell option 3           │
  │                                             │
  └─────── all three registered as overlays ───┘
           mapv1 / mapv2 / mapv3
```

mapv1 first — it's the fastest to build and will immediately tell us if the
character-native approach feels good. If it does, that informs how mapv3 renders.
mapv2 depends on the spike test result before committing to implementation.
mapv3 is the most complex terrain generation and should be built last.

---

## Commands

```
mapv1   — character-native Civ 1 style map
mapv2   — image-generated + pixterm rendered map
mapv3   — noise terrain + character rendering
```

All three are overlays. All three can be open sequentially to compare.
No tab changes, no layout changes, no sidebar changes required.

---

## Files to create

| File                  | Version | Purpose                                     |
|-----------------------|---------|---------------------------------------------|
| `ui/mapv1.go`         | v1      | Full implementation — tcell char rendering  |
| `ui/mapv2.go`         | v2      | Spike test + full image pipeline            |
| `ui/mapv3_terrain.go` | v3      | Noise terrain generation (no rendering)     |
| `ui/mapv3.go`         | v3      | Rendering layer — calls mapv3_terrain       |

Register in `ui/dashboard.go`:
```go
d.overlayMgr.Register("mapv1", "Map v1 (Character)", mapV1Provider)
d.overlayMgr.Register("mapv2", "Map v2 (Image)", mapV2Provider)
d.overlayMgr.Register("mapv3", "Map v3 (Terrain)", mapV3Provider)
```

Add to autocomplete and sidebar.

---

## Success Criteria

Each version is evaluated against:
- [ ] No flicker or artifacts on refresh
- [ ] City grows visibly as buildings are added
- [ ] Age transitions produce a clearly different look
- [ ] Wonders are visually distinct
- [ ] Readable at 80-column and 120-column terminals
- [ ] Feels like a civilization map, not a debug view
- [ ] Refresh does not cause noticeable lag

The winning approach (or hybrid of approaches) becomes the permanent `map` command.
The losing approaches are removed after evaluation.
