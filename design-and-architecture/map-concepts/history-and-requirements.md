# AgeForge live map: history and diagnosis

Read-only research, 2026-09-28. Sources: `git log --all` (844 commits, ~110 touch the map), PRs #26/#99/#100/#103, `design-and-architecture/map-*.md` + `city-synthesis.md`, `site/docs/city-map.md` + `world-map.md`, CHANGELOG, `map_demos/*.png`, the public Trello board (cards Wl35J3q1, BO5NDDAD, 3WaCRAzw, 3iqROH9M, NREUHI7m, d5DPYuBm, xqHcTxTo), the one surviving long Claude transcript for `~/Documents/ageforge` (session 8e71f7f5, 2026-07-06 to 08-10), the memory files `project_citymap_phase2.md`, `project_go_worldmap.md`, `project_worldmap_pictorial.md` (web port), and fresh PNG/ANSI dumps of the current renderers (`CITYMAP_PNG_DUMP`, written to `scratchpad/dump/`).

> **Note:** `map_demos/` and `ui/citymap` were deleted in phase 2 (#153). The paths cited below remain in git history before that PR. `site/docs/city-map.md` and `world-map.md` are now stubs pointing to `site/docs/map.md`.

Owner quotes are verbatim, typos included. Where I could only find an agent's paraphrase on a Trello card, I say so.

---

## 1. Timeline

Seven iterations in seven months. Each one was a rewrite, not a tune.

| # | Era | Dates | Approach | Fate |
|---|---|---|---|---|
| 0 | ASCII minimap + Map tab | 2026-02-19 (`920958a`) | Procedural ASCII settlement minimap + an F6 Map tab | Replaced the same day |
| 1 | `mapgen.go` pixel city | 2026-02-19 → 03-09 | `image.RGBA` → tview.Image half-blocks, 9 era palettes, 7 layout strategies, 19 sprite types, minimap | Deleted in full on 03-09 ("did not meet quality standards") |
| 2 | mapv1/v2/v3 experiments (PR #26) | 2026-03-23 → 03-24 | Three parallel overlays: glyph-native, half-block image, noise terrain | v1 absorbed v3; all three were replaced by v4 inside 24h |
| 3 | Sprites + MapV4 + AI photo backgrounds | 2026-03-23 → 06-30 | 16×16 PNG sprites (spritegen, `pkg/sprites`), 51MB of Midjourney "80,000-ft" backgrounds, jittered grid | Wiped: "a mess and not worth incremental work" |
| 4 | citymap rewrite P1–P4 + worldmap v1 (PR #99) | 2026-06-30 → 07-02 | Theme-tinted FBM terrain, 2.5D volumes, A* roads, then a pivot to a top-down Voronoi roof city; Game-of-Life worldmap | City pivoted mid-PR ("a city is not a planet") |
| 5 | Per-age material (PR #100) | 2026-07-02 → 07-07 | Each of the 22 ages gets its own roofs, ground, walls and wonder motif | Owner: the shape is still "the same odd oval/circle" |
| 6 | Per-age form + cosmic scenes + worldmap mediums (PR #103) | 2026-07-07 → 08-09 | Footprint shapes, space-mode, planet/galaxy scenes, 17 cartographic "mediums" + 5 strategic star maps | Shipped; this is the map the owner now calls "very clunky" |
| 7 | Light-theme maintenance | 2026-09-26 (`d7820d7`) | Luminance-aware re-keying for light themes | Upkeep only |

### Era 0: ASCII settlement minimap (2026-02-19)
- `920958a` "Phase 7 — UI polish, procedural map": a procedural **ASCII** settlement mini-map in the sidebar and a full Map tab (F6) with era terrain themes.
- **Fate:** the same day, `6d4aab9` "upgrade map to pixel rendering via tview.Image" replaced the ASCII with half-block pixels. That set the course for the next seven months. The project dropped the glyph approach within hours and never really tested it until mapv1, a month later.
- **What was good:** it was native to the medium, and it was always on screen.

### Era 1: `mapgen.go` pixel city + minimap (2026-02-19 → 03-09)
- `6d4aab9`, `0d519c3` (villagers removed, terrain depth added), `989d4cc` (era evolution), `63a1fb0` (building shapes), `2586418` (pixel-art wonder gallery), `e3fb8ed` (8 era layouts, wonders isolated in an outer zone), `ae58c21`/`34b325b` (wonder size and glow up, then down), `6dd1454` (7 era layout strategies, 19 sprite types, collision grid), `ece79ea`, `5eda9e5`, `dfc98b6` ("fix map city circular blob"), `69bfdb7`.
- **What it looked like:** see `map_demos/04_medieval_age.png` and `09_cyberpunk_age.png`. A tiny, unreadable cluster of coloured specks around 1/15th of the frame, a sine-wave river, and an ocean of flat noise. The city occupies about 3% of the canvas.
- A "Phase 20: Living City Map" (zone grid, 11 scene files, a 12fps animated overlay with particle shimmer, and a `citypulse` sidebar) is logged as done on 2026-03-08 in DONE.md (`2029e6c`). **No commit in `git log --all` contains those files.** It was either never committed or was lost. That was the only animated map ever attempted.
- **Fate:** `989ed15` (2026-03-09) deleted `mapgen.go` (2,640 lines), `minimap.go`, `tab_map.go` and `cmd/mapdemo`: "experimental visual map rendering code that was built and iterated but did not meet quality standards … will be revisited in a future PR." v3.0.0 shipped the same day with no map.
- **What was good:** the always-visible minimap; the 7 era layout strategies (organic → spokes → castle quarters → zoned grid → blocks → campus → orbital), which every later spec re-derived.

### Era 2: mapv1 / mapv2 / mapv3 experiments (PR #26, 2026-03-23 → 03-24)
- `d074765` added three isolated overlays, and `design-and-architecture/map-rendering-experiments.md` specced them:
  - **mapv1, "Character Native (Civ 1 Authentic)":** each cell is a `(rune, fg, bg)` via tcell `SetContent`, with domain runes `⌂ ♣ ▪ ⚒ ⚡ ⚔ ◎ ✚ $ ⚙ # ◆ ★`. The doc's pitch: "Don't fight the terminal … Zero image conversion, zero rendering artifacts … crisp at all font sizes."
  - **mapv2:** an RGBA image through `▄` half-blocks.
  - **mapv3:** Whittaker biomes + river carving, rendered as characters.
- Fixes in the PR show the medium pushing back: per-cell hashing gave "pure noise (TV static)" (`70bf1c6`); ocean cells caused "horizontal scan line artifacts from terminal font row gaps" (`a61b1ba`, `8fe63a2`); a golden-angle spiral collapsed every building onto 1–2 cells (`2fdf856`); mapv2 flashed until it was cached.
- `0dec129` merged v3's terrain into v1. Then, **the same evening**, `30e4f2f` added **mapv4** ("16x16 sprites and age-shifting terrain"), and the next day `c1b5f53` replaced v1, v2 and v4 with one `map` command built on the v4 approach.
- **Fate:** the glyph-native branch (mapv1) was abandoned within a day. No owner verdict is recorded. The commit record shows only the pivot to sprites.
- **What was good:** mapv1 is the only iteration that took the terminal as its medium, and the doc's arguments for it still hold. Its v3 terrain pipeline (FBM elevation + moisture → Whittaker biomes → rivers → coast) was reused in every later era.

### Era 3: sprites, AI photo backgrounds, MapV4 monolith (2026-03-23 → 06-30)
- **Sprites:** `3094763` (spritegen, 8×8 domain icons) → `5133da7` (16×16) → `95669b2` (per-building sprites for all 284 buildings) → `d2ed00f` (`pkg/sprites`) → `841e67e` "building sprites no longer look humanoid" → `ef5d75e` "regenerate all 735 building sprites".
- **Backgrounds:** `bb30363` produced `map-background-prompts.md`, which holds Midjourney prompts for "overhead map view in pixel art format, from 80,000 feet … in the old 2000's game view of Civilization 1". `c1b5f53` added 22 of these PNGs (plus a stray 8.6MB `primitive_age2.png`), downscaled into half-blocks, with an "organic jittered-grid city layout" and a density system. `15c1645` embedded them all in the binary.
- **Fixes and bugs:** `0e94705`/`c6dedeb` "city map only renders buildings from the current age" (Trello xqHcTxTo). The owner's own card NREUHI7m reversed that: "Buildings should continue to show from the previous age. but should be migrated from their ages look to the new ages look. I.E a hut in prim age would be transitioned to the sprite for a long house". `bcde55b` implemented it. Also `fcc65f6` gave the 22 wonders unique 16×16 sprites (card d5DPYuBm: every wonder had been the same generic pyramid).
- **Fate:** on 2026-06-29 (per the Trello comment on Wl35J3q1, an agent paraphrase) the owner said "the current map (MapV4, ui/map.go ~2734 lines) is considered a mess and not worth incremental work … a full FROM-SCRATCH map overhaul". `map-overhaul.md` is the post-mortem:
  - "**Fights the medium.** … 51MB of Midjourney-generated realistic overhead photos. Realistic photography at ~20k coarse cells is mush."
  - "**No coherent identity.** The 24 terrains were each AI-generated independently … The look *lurches* between ages; clean pixel icons get stapled on top. Two unrelated games in one frame."
  - "**The map means nothing.** … one jittered grid for every age, zero roads, wonders dumped in a bottom strip."
  - "**Heavy and rotten.** … 2.9MB of 752 sprite PNGs that silently fail to load … invisible to the theme system."
- **What was good:** the per-wonder silhouettes, which survive as `ui/wonder_icon.go` via `pkg/sprites`, and the owner's rule that old buildings persist and re-skin.

### Era 4: citymap rewrite P1–P4 + worldmap v1 (PR #99, 2026-06-30 → 07-02; +12,245 / −2,759)
- **P1 `19464aa`:** new `ui/citymap` package; FBM terrain tinted from **theme roles**, so the map retints when the theme changes; a 54MB wipe. P2 `258a80a` (per-age layouts, roads, 2.5D lineage districts), P3 `c6b70a3` (trade routes to the border, civ-edge markers, labels), P4 (biomes, A* roads). The design doc defines what the map is for: "Not navigation. Three jobs: (1) feel your empire grow and change by age, (2) ambient identity / delight, (3) glance-and-know". It also concedes the resolution: "a 200×50 terminal is a ~200×100-pixel canvas … iso's depth/overlap/height collapse at that size", hence "2.5D".
- **Review passes** in the same PR: `1261693` "show actual buildings, not resource districts", `4172bf7` "declutter citymap roads + de-snowstorm worldmap", `4e1e790` added the **worldmap** (a Game-of-Life settlement backdrop, your civ as a dot, diplomacy civs as relationship-coloured dots), and `2d9badd` put real terrain on the worldmap.
- **Pivot:** `2ae5e7e` "count-driven city synthesis" (terrain-gated, isometric-ish) failed playtest. The Trello comment on Wl35J3q1 (2026-07-01, agent paraphrase) reads: "PIVOT (playtest): the terrain-gated isometric citymap failed review — a city is not a planet." A 14-question design pass locked `city-synthesis.md`: a **top-down pixel-art roof city** ("Stardew, top-down Zelda"), with no world terrain, fill-frame, count-driven, a persistent layout and wonder anchors. `d6556c6` built the engine (V3-A). Then came a string of fixes to how it looked: "roofs read as roofs — kill yellow dots", "kill the pinwheel", "Voronoi block/ward citymap — streets are the gaps between blocks", "no more wagon-wheel everywhere", "organic village is a real mesh, not a plaza-hub wheel".
- **What was good:** theme-role tinting, deterministic seeded layout, stable incremental placement ("adding a building never moves an existing one"), Voronoi wards, drawing real building counts, and the city/world split.

### Era 5: per-age material (PR #100, 2026-07-02 → 07-07; +6,544)
- The owner on 2026-07-06: "fyi i just checked the citymap on a new "make" and every single age even into digital all look exactly like primitive.. is this expected given our current state? or ... is there a bug?"
- Decision on Trello, 2026-07-06: "EVERY AGE FULLY DISTINCT (not 7 shared bands)". Eleven styling commits followed (`3f29895` … `acc7a17`), each "PNG-reviewed by hand before commit". Background workers repeatedly died on network errors and the 600s watchdog.
- **What was good:** a real material progression (thatch → clay tile → slate → brick → glass → neon → metal); every age has its own wonder centrepiece.

### Era 6: per-age form, cosmic scenes, worldmap mediums (PR #103, 2026-07-10 → 08-09; +12,069)
- The owner on 2026-07-07: "the city stays the same odd oval/circle thorugh each age that i tested all the way up to trancendance … a circle village shape in modern or digital age... kinda wierd … a victorian era city like london is going to be sprawling and wide … imaging cyber punk age... with neon glimmer and sharp digitally defined streets... and megaroads … it was kinda wierd seeing a background of "ground" and trees. in the galatic and galazy eras … i agree the flourishes are all mismatched … i want real feel for each distinct age. plus a full and complete "wow" change per epoch era... 4. the backgrounds could use some better work as well."
- Also 2026-07-07: "when a player leaves the digital era epoch period and enters space.. i wan thtere to be a full "holy shit" kind of excitment."
- **2a** (per-age form) hit a structural wall. The Trello finding: "the near-centre RADIAL spokes are STRUCTURAL — a large open wonder-plaza in a Voronoi disc makes the plaza-adjacent wards radiate regardless of scatter form … Seed-scatter alone cannot de-radialize it." The team skipped to **2b** (footprint shapes: sprawl, round-rect, core+halo), then **2c** (space-mode void + starfield).
- The owner on 2026-07-08: "can you generate a view of that here? its alot of work to setup the game, use dev mode and find the citymap version..." When an attempt was shown inline, on 2026-07-09: "uhmmm im not sure what. you did. i just see a giant blob of raw code. just update teh game.. ill open the terminal".
- Also 2026-07-09: "ok so ... the galactic and transcendant should not be cities. space should be a view of the planet in pixels with sattelittes. galactic should be a view of a pixel rendered galaxy environment". The five cosmic ages then **abandoned the city renderer** for bespoke scenes (planet, star system, spiral galaxy, cosmic web, mandala). `6e4b027` had to squash the planet vertically so it "reads round in-terminal".
- **Worldmap:** the owner on 2026-07-10: "the "worldmap" it doesnt really work with the current static background being used across every scene... Take a look at what you are doing in the ~/Documents/ageforge-online project with world maps … obsiouvly its written for a fancier web view... i just mean the ideas..." That produced a seeded world model (`24665a2`) plus **17 cartographic "mediums"** (charcoal, petroglyph, clay tablet, hide, mosaic, parchment, copperplate, nautical, ordnance, lithograph, blueprint, retro atlas, satellite, vector, pixel, neon, hologram), and **5 cosmic strategic views** driven by faction state. The owner's one recorded reaction: "yep that looks way better."
- Parallel context: in the web port (canvas, `ageforge-online`), the same worldmap idea went fully pictorial, with oblique pen-and-ink mountains and watercolour washes. The owner reviewed it closely and approved it (memory `project_worldmap_pictorial.md`). **The idea this Go worldmap imitates was built for a real canvas.**

### Era 7: light-theme upkeep (2026-09-26)
- `d7820d7` "luminance-aware citymap and worldmap derivations", `27bc04a` (contrast matrix and a raw-colour guard). Every theme change now carries map re-keying work. Space ages "stay dark on every theme".

### The current state (2026-09-28)
- `ui/citymap/`: **20,717 non-test lines + 8,985 test lines**. `topdown.go` is 9,544 lines with 143 funcs; `worldmedium.go` is 3,722; `worldcosmic.go` 2,209; `worldmap.go` 1,243; `worldmodel.go` 898. For scale, all of `ui/*.go` is 21,511 lines and `game/`+`config/` non-test is 30,925. **The map is roughly as big as the rest of the UI and two-thirds the size of the game logic.**
- **Rendering:** each map builds an `image.RGBA` of cols × rows·2 and `streamHalfBlocks` writes `▄` with bg = upper pixel, fg = lower pixel, always via `tcell.NewRGBColor` (24-bit, **no 256-colour fallback**). Then `stampOverlay` draws crisp text labels on top (the "hybrid" model).
- **How it's opened:** only as a full-screen overlay via the `citymap`/`map`/`worldmap` commands (`ui/input.go:122-128`, `ui/dashboard.go:175-182`). There is no always-visible map: the minimap was removed in March, and later deferred "given the recent main-screen declutter".
- **Redraw:** the dashboard ticks every 500ms (`dashboard.go:469`). `Refresh` only stores the snapshot, and the image is cached on (w, h, age, total building count, theme key, plus civ signature for the worldmap). A full regeneration happens on every build, resize, age-up or theme change. A render takes tens of milliseconds; the whole citymap test suite runs in about 1.2s. **Performance is not the problem.**
- **Interaction:** none. There is no input capture, cursor, selection, tooltip, pan or zoom. The citymap reads building counts, age, trade routes and factions. The worldmap reads age, building count and factions. Neither shows anything you can't already read elsewhere, and neither lets you do anything.
- **What a new player sees** (`scratchpad/dump/forge_overlay_citymap.png`, primitive age, starting state): a flat grey-brown field, four huge pale-grey chunky road blobs radiating from a disc, one dark building blot, and a yellow "City Center" pill. The worldmap (`forge_overlay_worldmap.png`) is a pixelated dark continent with a jagged white stipple coast and a big yellow blob labelled "Your Empire". At terminal resolution each "pixel" is half a cell, so every edge is a staircase.
- The working tree still has `game/devmode.go` modified. That's the dev-lock bypass the map review loop needed ("User wants it LEFT bypassed for now (still testing)"). It's a small sign of how expensive it has been to look at the map.

---

## 2. Diagnosis: why no iteration ever felt right

### Technical
1. **Half-block pixel painting on a ~200×100 canvas, seven times over.** Eras 1, 2 (v2), 3, 4, 5 and 6 all stream an RGBA image through `▄`. Every doc concedes the ceiling: "Realistic photography at ~20k coarse cells is mush", "iso's depth/overlap/height collapse at that size", planets need an aspect squash. The fixes keep fighting the same artifacts: "TV static", scan lines from font row gaps, pinwheels, wagon wheels, blobs. Each rewrite changed *what* was painted and never *how*. The one glyph-native attempt (mapv1) lived less than a day.
2. **What gets reviewed is not what the player sees.** Per-age work was "PNG-reviewed by hand": square-pixel PNGs at 160×100 to 440×300, viewed as thumbnails or inline JPEGs. In the terminal the same art becomes blocky, depends on the font, needs 24-bit colour, and is re-tinted by the theme. The owner found the gap himself ("every single age … look exactly like primitive"; "its alot of work to setup the game, use dev mode and find the citymap version"). Review happened in the wrong medium, so "approved" never meant "feels right in the game".
3. **Theme retinting as a constraint on every pixel.** Every colour flows through theme roles, and then light themes needed luminance re-keying, a contrast matrix and a raw-colour guard. It's a real differentiator, but it multiplies each art decision by 11 themes and fights era identity. Cosmic ages had to opt out ("space is dark").
4. **Truecolor-only output with no fallback.** `tcell.NewRGBColor` everywhere, with no `Colors()` check. On a 256-colour terminal the subtle era palettes quantise badly. I can't confirm the owner's terminal, but the dependency is real.

### Design
5. **The map has no verb.** From `mapgen` to `worldcosmic`, the map is a read-only picture. There's no cursor, selection, inspect, or link to `build`/`assign`/`trade`/`expedition`, and it shows no numbers. In an idle game played through typed commands, a picture you open with a command and then close is a detour. `map-overhaul.md` defined its jobs as "feel growth / delight / glance-and-know", but the map can't be glanced at because it's hidden behind a full-screen overlay. The only glanceable version, the minimap, was deleted in March and deferred ever since.
6. **Static in a game about ticking.** Resources change every 500ms, but the map image changes only on a build or age-up. There's no motion, no workers, no flows, no events, no catastrophes. The only animated version ("Phase 20" living map, 12fps) never made it into git.
7. **Scope treadmill: 22 ages × 2 maps × bespoke.** The owner's standard ("every age fully distinct", "a full and complete 'wow' change per epoch") turned the map into 44 hand-tuned renderers and about 20.7k lines of code. Each arc solved the last complaint (material → shape → background → cosmic scenes → worldmap mediums), and the next complaint appeared one level up. Growth by age was the goal, but the work became per-age *restyling*, not showing *your* civilization's shape. Two players in the same age get near-identical pictures.
8. **Mixed metaphors between views.** The citymap is a Stardew-style roof city, then a planet and a galaxy from space. The worldmap is 17 pastiches of analog media (clay, hide, copperplate, satellite …), then a star map. Each age is internally coherent, but there's no through-line. This is the same problem `map-overhaul.md` called "two unrelated games in one frame", now spread across time instead of layers.

### Aesthetic
9. **Pixel-art imitation, not terminal-native.** Midjourney photos, 16×16 sprites, a "Stardew / top-down Zelda" roof city, parchment and satellite pastiche: each was trying to *look like a graphical game* at 1/20th of a graphical game's resolution. The rest of AgeForge is crisp text, box-drawing, glyphs, keycaps, bars and ▲/▼. The map is the one surface that isn't made of characters, which is exactly why it reads as "not appropriate for a terminal-esque game". The web port's success with pictorial maps shows the *ideas* are fine; the *medium* can't carry them.
10. **The hybrid look.** Soft blurry half-block ground with crisp text pills on top ("pill labels", "Empire — Primitive Age") gives two resolutions in one frame. Labels look pasted on because they are.

### The top 5
1. It paints pixels (▄) instead of drawing glyphs, and that ceiling has been hit seven times.
2. It's a passive overlay with no verb, no numbers, no link to commands, and it's hidden. Nobody glances at it, so it doesn't earn its keep.
3. It's reviewed as PNGs, not in the terminal, so "done" never matched how it felt in play.
4. It's static in a game that ticks.
5. The per-age bespoke scope (22 ages × 2 maps, about 20.7k lines) produced restyling, not expression of *your* civilization, and it's too costly to change course.

---

## 3. Requirements for a successful map (checklist)

**Medium**
- [ ] Built from **characters**: glyphs, box-drawing, braille/sextants where they help. It must look intentional at 1 cell = 1 unit. No RGBA images through `▄`, no photos, no sprite PNGs.
- [ ] Legible at **80×24** and still good at 200×50. It must degrade gracefully to **256 colours and 16 colours**, and its meaning must survive monochrome (glyph shape carries meaning, colour only helps).
- [ ] Crisp everywhere: one resolution per frame. Labels and art share the grid.
- [ ] Theme-aware through **roles, not per-pixel math**: a small set of role-coloured glyph classes, so a theme switch is free and light themes need no special re-keying.

**Purpose**
- [ ] Has a **verb**: the player can do or learn something there that's faster than typing elsewhere (select a building/district → count, output, workers, upgrade; or it mirrors the command vocabulary, e.g. names match `build <key>`).
- [ ] **Glanceable**: a compact form that lives on the main screen (or one keystroke away), as well as an optional full view.
- [ ] Shows **state the rest of the UI doesn't show well**: flows, bottlenecks, idle workers, trade and war lines, catastrophe pressure, what changed since the last check-in (important for an idler's return visit).
- [ ] **Alive**: changes on ticks (cheap glyph animation, flows, counters), not just on builds and age-ups.
- [ ] **Your** civ, not just your age: two players in the same age should see different maps, because their buildings, lineages, trades and factions differ.

**Scope and process**
- [ ] One **rendering grammar** across all 22 ages. Age identity comes from a small set of dials (glyph set, palette roles, density, a few landmark glyphs), not 22 bespoke renderers. The target is well under 5k lines.
- [ ] Early game must already look good: primitive age with 1–3 buildings is the first impression.
- [ ] **Reviewed in a real terminal**: txt/ANSI captures per age plus a dev command to jump ages without the dev-lock bypass. No PNG-only approval.
- [ ] One metaphor that holds from primitive to transcendent (or an explicit, designed epoch shift), shared or reconciled between the city and world views. Or merge them into one view.
- [ ] Deterministic, stable placement (never reshuffles), panic-safe, exact output size.

---

## 4. Reusable from past iterations

- **World model** (`ui/citymap/worldmodel.go`, `noise.go`, `terrain.go`): seeded FBM elevation + moisture → biomes, coastlines, rivers, account-stable seed. It's pure data, independent of the renderer, and could be rendered as glyphs (mapv3 already did this in March).
- **Stable incremental placement** (from `city-synthesis.md` revisions): a lot is a pure function of (building type, instance index, seed), so adding a building never moves another. This is a key invariant for any city view.
- **Count-driven synthesis rules:** near 1:1 at low counts, sub-linear at high counts; old buildings persist and re-skin on age-up (owner rule, card NREUHI7m).
- **Era layout vocabulary:** organic → radial → walled quarters → survey grid → blocks → campus → rings. It's been re-derived three times and is consistent, so keep it as data.
- **mapv1's glyph tables** (`map-rendering-experiments.md`: terrain runes and domain runes `⌂ ♣ ▪ ⚒ ⚡ ⚔ ◎ ✚ $ ⚙ # ◆ ★`), plus its rationale.
- **Faction/strategic data plumbing** (`worldcosmic.go` `strategicEmpires()`, `worldmap.go` `worldCivs`, `civStrengthBucket`) and the fixed relationship signal colours (war red / ally green / mercantile gold / neutral steel-blue).
- **Trade-route and civ-edge weave idea** (P3): systems drawn as lines into and out of the city.
- **Overlay/widget contract and lock discipline** (`Refresh` stores the snapshot, the draw reads it, the cache is keyed on state) and the `panic-safe exact size` test pattern.
- **Per-age landmark and wonder identities** (ziggurat, keep, temple, cathedral, dome, factory, setback tower, space needle, data hub, reactor core, launchpad, rings), which work as a list of landmark glyphs or mini-diagrams.
- **Per-epoch "medium" idea for the worldmap:** keep it as a dial (glyph set + border chrome per epoch), not a raster pastiche.
- **Capture tooling:** `theme_dump_test.go` already writes `.ans`/`.html` terminal captures, so extend that to per-age captures for review in terminal form.
- **Not reusable:** half-block streaming as the primary renderer, `topdown.go` roof sprites, raster "medium" painters, `pkg/sprites` for the map (keep it for `wonder_icon.go` only), and the Midjourney prompt doc.
