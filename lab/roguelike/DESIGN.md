# Glyph World: the map as a roguelike

## The idea

The map is a living roguelike map in the lineage of Brogue, Dwarf Fortress and
Caves of Qud. Every cell is one coloured character that means something:
`♣` is forest, `≈` is water, `§` is a knowledge building, `$` is trade, `☺` is a
worker walking to their job, and `@` is your scouts heading into the fog. A
seeded world of terrain (coasts, rivers, forests, hills) surrounds your
settlement, and the settlement grows out of your real buildings, tile by tile,
in quarters per lineage, with streets, walls and wonders. A cursor inspects
any tile in the game's own words ("♦ Forge ×4 · 12/16 workers · ~+3.2 iron/s ·
build forge"). One view and one grammar cover both the city and the world:
zoom out and the same glyphs become the region, with the civs you have met and
the roads to them. Zoom in and tiles double in width, with counts and
building names on the grid.

```
go run ./lab/roguelike -age medieval_age                 # interactive (q quits)
go run ./lab/roguelike -age galactic_age -theme daylight
go run ./lab/roguelike -age industrial_age -catastrophe  # catastrophe overlay on
go run ./lab/roguelike -captures                         # rewrite captures/
go run ./lab/roguelike -gen medieval_age -seed 11        # new state via smoke bot
```

Keys: arrows or `hjkl` move the cursor (Shift moves in steps of 8), `z`/`x`
zoom in or out, `Tab`/`Shift-Tab` jump between buildings, `c` returns to the
square, `?` toggles the legend, `q` quits.

## Why it fits, and why it avoids what went wrong before

The history report found seven failures. Here is how this concept answers each:

| Past failure | This concept |
|---|---|
| Wrong renderer: RGBA images pushed through `▄` half-blocks | No images at all. Every cell is a glyph set with `SetContent`. Labels, counts and art share one grid at one resolution. |
| No verb; a passive overlay | A cursor with an inspector that speaks the command vocabulary (`build forge`, `assign`, `diplomacy <civ>`), plus `Tab` to jump between buildings. |
| Reviewed as PNGs | Every capture in `captures/` is a real tcell `SimulationScreen` dump (`.txt` plus `.html` built from the cells). |
| Static in a game that ticks | Water shimmers, crops sway, smoke rises from staffed works, workers walk real paths, caravans travel the roads, scouts walk into the fog, the hearth flickers and the harbinger blinks. It all runs off one frame counter and costs nothing to simulate. |
| Scope ballooned (20.7k lines of per-age art) | One grammar with per-epoch dials (`glyphs.go`: road style, wall style, a few glyphs, a hue, night mode, space mode). The whole prototype, including the state generator and tests, is about 3.5k lines; the renderer itself is about 2.6k. |

The requirements checklist from the history report:

- **Glyph-native, with labels on the grid.** Yes. Wonder names at settlement
  zoom, civ names at region zoom, and building names plus counts (`♦⁵`) at
  district zoom are all placed in free cells of the same grid.
- **Legible at 80×24; degrades to 256/16 colours; meaning survives
  monochrome.** Yes. See `medieval_age_80x24.*` and `industrial_age_100x30.*`.
  Shape carries the meaning: the `.txt` captures are the monochrome test, and
  they read. Colours are plain RGB `tcell` colours, so tcell's palette
  fitting handles 256 and 16 colours.
- **Theme colours from a few role classes.** 17 classes (`Class` in
  `glyphs.go`), each derived from a theme role, nudged toward an identity hue
  (water leans blue) and clamped with `theme.Legible`. Daylight, Parchment and
  Cyberpunk captures need no special-casing.
- **A verb.** Inspect, jump and read the command to type. The production
  version would let `Enter` prefill the command line (`build forge`).
- **Glanceable compact mode.** `DrawCompact` gives a 40×15 mini-map (see
  `*_mini.*`) with one header line (idle workers shown in the warning colour)
  and one news line ("▲ +2 Guildhall, +1 Castle Keep" or "⚠ Barbarian
  Invasion").
- **Shows what the rest of the UI doesn't.** It shows idle workers standing in
  the square (`☻`, amber), understaffed buildings dimmed, what changed since
  your last visit (green background on new tiles plus a summary line), roads
  and caravans to civs, war (`×` on the road, the civ shown in red),
  catastrophe pressure (a ring of hazard glyphs outside the walls), scouts and
  raiders out, and the harbinger waiting at the town's edge.
- **Changes on ticks.** See above; `medieval_age_anim.html` has 8 frames.
- **Two players in the same age see different maps.** The seed sets the
  terrain, the river, the town's name, the quarter order and the street
  stagger. The state sets everything built. Compare `medieval_age.*` with
  `medieval_age_seed11.*`. `TestDeterministicAndPersonal` checks both.
- **One grammar, well under 5k lines.** Yes, as above.
- **The Primitive Age with 1–3 buildings looks good.** Every instance is one
  tile while counts are small: a few `∩` huts around a flickering `*` hearth by
  a river, fields, a woodcutter at the forest edge, people walking out to
  gather, a torchlit circle of fog. See `primitive_age.*` and
  `primitive_age_anim.html`.
- **Deterministic, stable placement.** The world and the lot plan are pure
  functions of the seed. Buildings take lots by replaying a build history
  (older tiers first, tile by tile), so the current age's growth only adds
  tiles at the edges. `TestGrowthNeverMovesNeighbours` checks this. The one
  honest gap: an out-of-order build within the same tier can nudge a
  neighbour one lot. Production should persist lot claims in the save as an
  append-only list (a few hundred ints), which makes placement strictly
  stable.
- **One metaphor for city and world.** The region zoom is the world map: the
  same terrain glyphs down-sampled, civ settlements (their initial letter in
  reverse video, coloured by stance), roads to them and fog. It replaces
  `worldmap` as well as `citymap`.

## What the player sees and does

It is interactive but works fine passively.

- **Region zoom** fits the whole world to the terminal. It shows the
  settlement as a dense glyph cluster, the civs you have met (named on the
  grid), dotted roads with `&` caravans, and fog over what nobody has scouted.
- **Settlement zoom** (the default) shows one tile per cell: the town, its
  walls, the fields and mines outside, walkers and smoke.
- **District zoom** makes tiles two cells wide (square, in the Brogue style).
  Each building cluster's first tile shows its count as a superscript (`♦⁵`,
  `⁺` for 10 or more), and building names label the streets like a street
  map.
- **Cursor and inspector.** The camera stays put and scrolls only when the
  cursor nears an edge, as in a roguelike. The first inspector line describes
  the tile under the cursor; the second holds contextual advice or the build
  key. Alerts (catastrophe, since-last-visit) sit at the right, or take the
  hint line on narrow terminals.
- **Legend.** A sidebar at 110 columns or wider, toggled with `?`. It is built
  from what is actually drawn in the current frame, so it never lists
  glyphs you can't see and never misses one you can.

## How it reacts to the game

| Event | On the map |
|---|---|
| New building | New tiles at the edge of its quarter, green background until your next visit; the news line says "+2 Guildhall". |
| Age advance within an epoch | Buildings keep their lots and upgrade in place (lineage tier changes, the glyph stays). |
| Epoch change | Dials turn: trodden paths become light streets, then heavy avenues, then neon boulevards; the palisade becomes a stone wall, then a ring boulevard where the wall stood; housing goes `∩` → `⌂` → `▪` → `▓` → `◘`; fields go `"` → `≡` → `▤` → `▦`; the neon era goes to night palette; in the cosmic era the region view pulls back to show the planet's limb, stars and an orbital ring with stations. |
| Wonders | 3×2 prefabs on the best free plot at the time they were built. They keep their era's look (a ziggurat stays a ziggurat in the neon city). |
| Catastrophe pending | A pulsing ring of epoch-specific hazard glyphs just outside the walls: meteors, a barbarian host `Ж`, smog, fallout, glitch characters, solar flares, reality fractures. A red ⚠ line appears in the inspector and the mini view. |
| Ruins | `%` tiles, kept where the building stood. |
| Trade and civs | A road is drawn to every civ you have met; caravans `&` walk it when you trade; at war, the road shows a blinking `×` and the civ turns red. |
| Expeditions | The fog radius grows with age and with completed expeditions; the lands of civs you have met and the roads to them are revealed; an active scouting party `@` walks to the edge of the known world; raiders `»` walk toward your least-liked civ. |
| Idle workers | `☻`/`☺` in amber standing in the square, one per ~6 idle, up to 9. |
| Understaffed buildings | Glyph dimmed (under 50% staffed). |
| Harbinger | `Ψ` waiting just outside the walls, blinking. |
| Time passing while idle | The animations above; "since last visit" highlights and a news line are what an idle player sees first. |

## How it looks at each epoch

(These are the dial settings; see `captures/<age>.html`.)

- **Stone Era:** no streets, only trampled ground; `∩` huts round a
  flickering `*` hearth; a `#` palisade from the Bronze Age; a torchlit fog
  circle.
- **Iron Era:** a staggered grid of light-line streets (`┼`), double-line
  stone walls with `■` towers, `Π` hall, colonnade wonders `╓╥╖`.
- **Steel Era:** the same walls, `≡` ploughed fields, `Φ` town hall, smoke
  over the works, brick-red industry.
- **Electric Era:** the walls come down and a ring avenue replaces them;
  heavy `╋` avenues; amber light.
- **Digital Era:** `▓` block housing, cyan works.
- **Neon Era:** night palette (terrain dimmed, structures glow), double-line
  neon boulevards, magenta works.
- **Cosmic Era:** `◘` habitats, `▦` hydroponics, a violet palette; region zoom
  shows the planet in space with an orbital ring.

## Light and dark themes, small terminals

- The classes are derived from theme roles and clamped with `theme.Legible`
  against the actual background, so light themes darken and dark themes
  brighten automatically. Water gets a faint blue background tint (mixed from
  the theme background), which keeps coastlines readable in both. See
  `medieval_age_light.html`, `industrial_age_light.html`,
  `galactic_age_region_light.html`, `medieval_age_parchment.html` and
  `cyberpunk_age_cyberpunk_theme.html`.
- **Small terminals:** the legend drops below 110 columns; below 100 the key
  hints shorten; alerts take the hint line; below 60×16 `Draw` falls back to
  the compact mini view. Every zoom draws at any size from 1×1 up
  (`TestDrawAnySize`).

## Performance

Measured on an M5 with the Galactic state (90 building types, about 1,000
tiles):

- world and plan generation: about 13 ms, once per seed;
- `Build` (layout, walls, roads, fog, walker paths): about 33 ms, run only
  when the snapshot's buildings, civs, routes or expeditions change, keyed on
  a hash (not every tick);
- one 160×48 frame: about 2.5 ms, including the legend.

At the 280 ms animation cadence that is under 1% of a core. The frame cost is
linear in visible cells; nothing is proportional to building counts at draw
time.

## Integration

- Replace both `citymap` and `worldmap` with one `ui/glyphmap` package: a
  `tview.Box` whose `Draw` calls `View.Draw` (full panel) or
  `View.DrawCompact` (dashboard sidebar). `Refresh` stores the snapshot and
  rebuilds the scene only when its layout hash changes. The existing lock
  discipline stays: never call `GetState` from a draw.
- Commands: `map` opens the full view (`map region`, `map district` start at a
  zoom); the dashboard gets the mini view in place of the minimap.
  `citymap`/`worldmap` become aliases of `map`/`map region`.
- `MapView` (snapshot.go) is the contract: a narrow, JSON-able view of
  `GameState`, which is also why the captures can be regenerated without the
  engine.
- **Persistence:** save the seed (it already is) and, for strict stability, a
  per-building list of claimed lots.
- **Dev tooling:** `-gen` builds states with the smoke bot, and `-captures`
  rewrites every capture. This is the "review in a real terminal" workflow the
  history report asked for.

## Effort to productionise

About 2–3 weeks for one engineer:

1. Move the prototype to `ui/glyphmap`, wire it to `Refresh`/`Draw`, add the
   layout-hash cache and persisted lot claims, and cover exact size and panic
   safety with the tests that exist here (3–4 days).
2. Dashboard mini view and the `map` commands; `Enter` prefills `build <key>`
   (2 days).
3. Tune per-epoch dials across all 22 ages; per-lineage glyph review with
   Adam; civ placement from `FactionInfo` rather than slots (3–4 days).
4. Real routes and the harbinger and catastrophe mapped to their actual
   engine events; "since last visit" keyed on the real check-in timestamp
   (2 days).
5. Docs and site updates, `commands.md`, and a wiki page with the legend
   (1–2 days).

## Risks, and what would make it feel wrong

- **Glyph soup.** At Galactic, 90 building types in about 1,000 tiles gets
  busy. Mitigations: sub-linear tile counts (`tilesFor`), quarters per
  lineage, the legend built from the current frame, and district zoom for
  names. If it still reads as noise, collapse legacy tiers into one
  "old town" texture.
- **Font coverage.** The glyph set sticks to CP437 and box-drawing (`♣ ♠ ♦ §
  Ω Φ Π ☺ ☻ ≈ ▲ ▼ ◘ ◙ ▓`) plus `λ`, `τ`, `Ж` and superscript digits. Those
  render in every mainstream monospace font. Emoji-width symbols are avoided
  entirely, but a player with an odd font could still see tofu.
- **The street grid can read as a spreadsheet** in dense eras. The stagger
  per band, minor versus major streets, and dropping streets that run
  alongside the ring help. More would come from occasional plazas and parks
  in empty lots.
- **Walls move** as the town grows past the 85th percentile. That is historically
  right, but a player watching closely will see the ring jump outward. It
  could be pinned to the maximum radius per epoch.
- **Wrong if:** the map stops matching the numbers (it must never draw a
  building you don't have), the animation gets busy or flashy (it should
  stay at a background shimmer), or the camera moves on its own.
- **Prototype honesty:** the state generator's bot rarely starts trade routes
  (routes need a market, port or harbor; markets upgrade away on age-up and the bot
  never rebuilds one), so caravans appear on roads to civs you have traded
  with (`TradeCount`), not on routes. The world has 11 civ slots placed by
  seed, not by any engine geography. The catastrophe capture forces the
  overlay on (it says so in its title).

## Files

- `snapshot.go`: the `MapView` contract, plus a smoke-bot state generator
  (with `explore.go` adding scouting and trade the bot skips).
- `world.go`: seeded terrain, rivers, centre, civ sites and A*.
- `layout.go`: the static lot plan and the stable replayed placement.
- `scene.go`: streets, walls and ring, roads to civs, fog, walkers,
  hazards and change tracking.
- `glyphs.go`: epoch dials, lineage glyphs and theme-role colour classes.
- `render.go`: zooms, animation, labels, legend, inspector and the mini view.
- `capture.go`, `captures_all.go`: SimulationScreen to `.txt` and `.html`
  (animated when there are multiple frames).
- `states/`: cached real states (seed 7 for seven ages, plus Medieval with
  seed 11).
- `captures/`: all captures; open `captures/index.html`.
