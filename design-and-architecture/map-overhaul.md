# Map Overhaul — the `citymap` rewrite

**Status:** active (2026-06). Supersedes `map-system-guide.md`,
`map-layout-strategy-specs.md`, `map-rendering-experiments.md`, and
`map-background-prompts.md` — all aspirational or stale, none realized by the
shipped MapV4.

## Why (the teardown)

The current map (MapV4, `ui/map.go`, 2,734 lines) is being wiped. It fails on
every axis:

- **Fights the medium.** It composites a full RGBA image and streams it through
  half-blocks (`▄`, bg=upper px / fg=lower px), so a 200×50 terminal is a
  ~200×100-pixel canvas. Into that it stuffs **51MB of Midjourney-generated
  realistic overhead photos**. Realistic photography at ~20k coarse cells is
  mush.
- **No coherent identity.** The 24 terrains were each AI-generated
  independently — primitive is a noisy green satellite photo, quantum a slick
  dark hex-topo. The look *lurches* between ages; clean pixel icons get stapled
  on top. Two unrelated games in one frame.
- **The map means nothing.** Docs specced 7–8 era-specific layouts with roads;
  MapV4 uses **one jittered grid for every age, zero roads**, wonders dumped in
  a bottom strip. A primitive camp and a quantum metropolis are *topologically
  identical*.
- **Heavy and rotten.** ~51MB embedded (incl. a dead 8.2MB `primitive_age2.png`)
  + **2.9MB of 752 sprite PNGs that silently fail to load** and fall back to
  in-code art — pure dead weight. A 2,734-line monolith with no separation. And
  it is **invisible to the theme system**.

## What a map is FOR (in an idle game)

Not navigation. Three jobs: **(1)** feel your empire grow and change by age,
**(2)** ambient identity / delight, **(3)** glance-and-know "this is my
civilization." MapV4 serves none.

## Decision

**D1 spine + D2 terrain + D3 weave + 2.5D depth.** Full wipe of the 51MB photos,
the 2.9MB dead sprite PNGs, and the MapV4 monolith.

### Isometric → 2.5D (the resolution reality)

Half-blocks give a ~200×100px canvas. A literal iso tile grid means ~16×8px
diamonds — iso's depth/overlap/height collapse at that size and it burns
vertical space. So the *part of iso we want is depth*, delivered as **2.5D**:
buildings as solid volumes (lit roof cell + shaded wall cell + drop shadow),
slightly staggered, reading as dimensional without an iso projection the
resolution can't carry. Fuller iso is a later experiment if 2.5D lands.

### Rendering model (hybrid)

- **Terrain:** procedural (simplex/FBM → elevation/biome), soft, drawn via
  half-blocks, **tinted from the active theme palette** → retints live.
- **Structure:** buildings, roads, districts as crisp theme-colored
  glyphs / blocks / box-drawing *overlaid* on the terrain; buildings get the
  2.5D depth treatment.
- The **whole map is theme-aware** — switching themes retints it live. The
  differentiator nothing else in the genre has.
- **Per-age layout strategies + roads** — so structure and silhouette evolve.
- **Systems-weave (D3):** real trade routes drawn as lines to civ-edge markers;
  the 11 diplomacy civs as relationship-colored edge markers; lineage districts.

## Architecture

New `ui/citymap` package, separated modules:

- `terrain` — procedural elevation/biome, theme-tinted half-block fill.
- `layout` — per-age strategy dispatch + road generation.
- `entities` — buildings → 2.5D glyph-volumes colored by lineage/theme-role;
  civ markers; trade-route lines.
- `render` — composite (soft half-block terrain + crisp glyph/cell structure),
  stream to the screen.
- `themebridge` — pull `theme.Color(role)`; redraw on theme switch.

**Wipe:** `assets/maps` (51MB) + its `//go:embed`; `assets/sprites/buildings`
(2.9MB dead); `ui/map.go` (MapV4). Binary −~54MB.
**Keep:** a clean half-block render primitive (reimplemented in `citymap`),
`pkg/sprites` (in-code sprite gen, optional for structure), the age/era model.

### Per-age layout strategies

organic scatter (primitive/stone) → hub-and-spoke roads
(bronze/iron/classical) → castle + quarters (medieval/renaissance) → zoned grid
(colonial/industrial/victorian) → city blocks (modern/atomic) → campus clusters
(digital/cyberpunk) → orbital rings (space/galactic/quantum). Roads per strategy.

### Minimap

Deferred — overlay-first (the `map` view), given the recent main-screen
declutter. A compact always-visible minimap is a later option if wanted.

## Phases

- **P1 — Foundation + wipe:** the `citymap` package; theme-aware procedural
  terrain + basic building markers; register as the `map` view; **delete** the
  51MB embed + dead sprite PNGs + MapV4. Result: a light, theme-retinting map;
  binary −54MB.
- **P2 — Structure & meaning:** per-age layout strategies + roads + 2.5D
  buildings colored by theme role per lineage.
- **P3 — Weave & polish:** trade routes + civ-edge markers + lineage districts +
  per-age styling; final cleanup.
