# Atlas: the world view as a cartographic map drawn in text

## The idea

The world view becomes an **atlas**: a map of your known world, drawn the way
real maps are drawn, in type. It has a coastline, relief, rivers, labelled
regions in spaced capitals, sea names in italics, a cartouche, a compass, a
scale bar and a legend. Your realm spreads across it as a border. The peoples
you have met have seats and territories coloured by how they stand with you.
Trade routes are dotted lines with cargo moving along them, and expedition
tracks run out into terra incognita. War fronts, blockades, harbinger omens and
catastrophe scars sit on the map where they happen. The atlas changes the way
cartography changed. It starts as charcoal on a stretched hide, then becomes a
portolan chart with rhumb lines, an engraved Enlightenment map with
hand-coloured borders, a survey sheet with contours and grid references, a
satellite mosaic, and an orbital night pass where your realm is its city
lights. When the empire leaves the planet, the atlas zooms out to an orrery,
then a star chart, then the galaxy. Each era you reach adds a **plate** to the
atlas, and you can leaf back through the earlier plates.

## Why it fits a terminal, and why it avoids what went wrong before

The earlier maps rendered images and pushed them through `▄` half-blocks. The
atlas is **type**, in the sense that paper maps were typeset: every cell is a
glyph that means something, and labels share the grid with the art.

| History's diagnosis | What the atlas does |
|---|---|
| The wrong renderer: RGBA images through half-blocks | No raster. Glyphs, box-drawing for coasts, contours and water lines, `░▒▓` only as texture. One world unit per cell at the default zoom. |
| No verb, hidden | A cursor, tab through points of interest, and an inspector that ends in the game's own command (`» diplomacy accept riverlands_tribes 1`, `» trade route stop spice_trade`, `» expedition scout_party`). Enter picks the verb. |
| Reviewed as PNGs | Every capture here is a real tcell `SimulationScreen` read back cell by cell, as `.txt` and `.html`. |
| Static | Cargo moves along the routes, the active expedition's flag advances on its track, the harbinger's omen pulses, and the radar sweeps. City lights flicker, and the satellite plate has a downlink scanline. A frame costs 1.3 to 2.4 ms at 200×60. |
| Scope creep (20.7k lines of per-age art) | One renderer and nine **plates**. A plate is a table of dials (glyphs, relief mode, water lines, frame, fog, colour classes), not code. The prototype's renderer and model come to about 4.7k lines, and about 2.2k of those draw anything. |

It also fits because an idle game's world is mostly *knowledge*: who you have
met, where your goods go, what lies unexplored. A map is the native form for
that, and text maps have a long history, from the portolans to the Ordnance
Survey to the green phosphor of radar.

## What the player sees and does

- **The full view** (≥120 columns): the map in its plate's frame, plus a side
  panel with ATLAS (plate, peoples met, routes), SINCE YOU LAST LOOKED, INSPECT
  and LEGEND. A status line shows the cursor's position in the plate's own
  terms: `13 leagues by the Southwest wind from Yrogate`, `grid ref 905 375`,
  `LAT +12.40 LON -31.02 RNG 812 KM`, `RA 4h12m · 8.3 ly from Sol`.
- **Narrow** (80–119 columns): the map, a four-line inspector strip under
  it, and a one-line title slip in place of the cartouche.
- **Mini** (≤60×18, sidebar-sized, e.g. 40×15): a locator for the main
  screen. It shows the realm at a fitting zoom, the peoples, routes and alerts,
  with no furniture. Its last line is the one thing most worth knowing
  (`✕ war · Ironhold Clans`, `☄ the Town Crier 11%`, `+ met the Merchant
  Guild`).
- **Keys** (interactive prototype): arrows or hjkl move the cursor, and the
  map scrolls when you push at its edge. WASD or HJKL pan. `+`/`-` zoom about
  the cursor and `0` refits. Tab and shift-tab step through points of interest
  (capital, peoples, routes, the omen, the active expedition, scars, wonders,
  towns, old tracks). `[`/`]` leaf through the plates you have earned, and
  `<`/`>` step through the age snapshots. `t` cycles themes, `m` shows
  monochrome, `g` toggles the legend, Enter runs the verb (the prototype
  shows `would run: ...`) and `q` quits.

## How it reacts to the game

Everything is a pure function of `(seed, GameState)`. The world comes from
`GameState.Seed`, so two players in the same age see different continents,
and their buildings, peoples and routes differ too. Placement is by **fixed
lists**: the realm's claim order, the town sites, the civ seats and the camp
marks are computed from the seed once. The state decides only how far along
each list the map has got, so adding a building never moves anything.

| Game state | On the map |
|---|---|
| Buildings (count) | How far the realm has grown along its claim order, and new towns. In the Stone Era: hunting grounds around the camp, one mark per building kind, each named while there are few. |
| Age / epoch | Which plate. The known radius grows with age and expeditions. From the Digital Era the whole planet is known from orbit. |
| Wonders | `✦` landmarks around the capital. |
| `Diplomacy.Factions` | Seats and territories appear only once a people is discovered. Standing sets the signal colour: allied green, friendly gold, neutral steel, rival amber, war red. Standing is also spelled out on the label (`+ally`, `!embargo`, `✕war`), so it survives without colour. The inspector shows personality, strength ★, a standing bar, specialty, deals and lent workers. |
| `AtWar` | A `✕` front along the shared border, or halfway between the capitals. |
| `Status == embargo` | A `▫` blockade ring around their seat. |
| `Trade.ActiveRoutes` | A route to the partner people (matched by specialty) or to a foreign market, found by A* (sea cheap, mountains dear). Cargo moves both ways, and a `✕` marks a disrupted route. |
| Expeditions | A track per expedition kind, and the three most recent are brighter. The active scout's flag advances with `TicksLeft`. Tracks reveal the fog along their way. |
| `Harbinger` / `HarbingerHistory` | The omen at the edge of the known world, pulsing, with the odds once the era prints numbers. Past omens are left faint. |
| `EpochEventHistory` catastrophes | A scar in a province of the realm, named, faded once endured. |
| `Workers.TotalIdle`, `AgeReady` | In the capital's inspector: `18 hands with nothing to do » assign`, `Renaissance Age is within reach » advance`. |
| Since the last check-in | Newly annexed provinces take a highlight wash, and new peoples get a `new` tag. The panel lists what changed: provinces, peoples met, routes opened, expeditions returned, wonders raised, scars. |

## How it looks in each epoch (the plates)

| Epoch | Plate | Dials |
|---|---|---|
| Stone (Primitive–Bronze) | Charcoal on hide | The map is cut to a stitched hide shape. Sparse strokes: `^` ranges, `♣` woods, `∴` sand, `~` water. A charcoal smudge marks where knowledge gives out. "the lands we know / drawn by the fire of X". Scale in days' walk, and `☼ sunrise` on the east edge. |
| Iron (Iron–Medieval) | Portolan chart | A parchment tint, a bold coast, blank interiors with `⋀` ranges, and a web of rhumb lines from two wind roses (black, green and red winds, from theme roles). Sea serpents and *hic svnt dracones* in the unknown. Leagues, and a double-rule frame. |
| Steel (Renaissance–Industrial) | Engraved map | Water-lining parallel to the coast, hachured slopes, `▲` peaks, hand-coloured territory washes, a graticule with degree labels on the neatline, and a compass. "A New & Accurate Map of X". |
| Electric (Victorian–Atomic) | Survey sheet | Contours every 185 m with bold index contours, spot heights, `♧` woodland, a blue water tint, a numbered grid and grid references, and hatching over unsurveyed land. |
| Digital (Modern–Digital) | Satellite mosaic | Biome colour as the photograph, with glyphs only where relief or canopy shows. HUD corners, a pass stamp, a downlink scanline and targets. |
| Neon (Cyberpunk–Fusion) | Orbital night pass | Dark land, where settled land glows as city lights in each people's colour. Satellite ground tracks and a radar sweep from the capital with blips. |
| Neon (Space) | Orrery | The home system as tilted ellipses, the belt, colonies, lanes with cargo, and the peoples beyond as arrows at the edge. |
| Cosmic (Interstellar) | Star chart | A polar chart of the neighbourhood (10 ly rings, RA spokes), named stars, constellations, colonies linked by hyperlanes, and the spheres of influence of off-world peoples. |
| Cosmic (Galactic+) | Galactic atlas | Spiral arms, the core's glow, and sectors: yours grows by age, theirs are coloured by standing. War fronts, hyperlanes, and named arms. |

## Light and dark themes, small terminals, poor colour

- **Colour classes, not per-cell colour.** Everything is drawn in a few dozen
  classes per frame, and most of them are theme roles outright: Text, Dim,
  Label, Accent (you), Border, Positive (ally), Negative (war, omen), Warning
  (rival), Highlight (friendly, new). The rest are a fixed hue lifted for dark
  backgrounds or sunk for light ones (water, relief, forest, sand, snow,
  lights), plus mixes of these with the paper. They are computed once per
  frame and memoised (`Pal.mix`). The paper is the theme's background, with a
  touch of the plate's tint (parchment, hide). On a light theme the portolan
  and engraved plates look like printed maps. See `light_*`.
- **Monochrome**: glyph shape carries the meaning. Coasts are box-drawing,
  and relief is `^ ⋀ ▲`, hachures or contours. Standing is on the label, and
  war is `✕`. See `mono_industrial_age`. **16 colours**: see
  `ansi16_medieval_age`, where every colour is snapped to the xterm 16. tcell
  degrades to 256 colours on its own.
- **Small terminals**: 80×24 and 100×30 use the narrow layout. The mini view
  is sidebar-sized, and the tests draw every plate at 40×15, 80×24, 100×30,
  160×48 and 220×64, on a dark and a light theme, with exact output size and
  no panics.

## Performance

A full 200×60 frame takes 1.3 to 2.4 ms on an M-series laptop
(`go test ./lab/atlas -bench Draw`), with the world model generated once per
seed (about 50 ms). The only per-cell work is noise lookups, a knowledge test
and a colour-class lookup. The first version mixed colours per cell and took
22 ms. The model is rebuilt from the snapshot; in production it would be cached
on a state signature (age, building count, civ set, routes, journal length).

## Integration

- **Replaces `worldmap`** (and `worldmedium.go`, `worldcosmic.go`,
  `worldmodel.go`) outright. The command is `map` (`worldmap` stays as an
  alias), and `map <plate>` or `map plates` leafs through plates.
- **Mini view on the main screen**: the 40×15 locator can sit in the
  dashboard sidebar. Its status line is the glance.
- **Citymap**: the honest answer is that the atlas is the world view. The
  same grammar extends to one more plate at the deepest zoom, *a plan of the
  capital*: streets in box-drawing and blocks by lineage, the classic town
  plan in an atlas. That would give one metaphor for both views. I have not
  built it here, and it is the natural next step if the atlas is picked.
- **Verbs**: the inspector's last line is always a real command. In the game,
  Enter would put it on the command line, so the atlas teaches the command
  vocabulary instead of bypassing it.
- **Engine additions** (small):
  1. An **expedition journal** in the save: per expedition key, the runs, the
     last tick and the first age. The engine keeps only a count today. The
     prototype's driver records it; about 60 lines in `game/`.
  2. A **last-look digest** for "since you last looked": provinces held, civs
     known, route keys, journal length and wonders at the last time the map
     was opened. It is a UI-side file or a save field. The prototype diffs
     against the previous age's snapshot instead.
  3. Nothing else. Everything else is already in `GameState`.

## Effort to productionise

About two weeks of focused work, or five or six agent phases:
1. `ui/atlas` package: port world, model, plates, canvas and scene, and wire
   it as an overlay widget (Refresh stores the snapshot, Draw reads it, cache
   by signature). *3–4 days*
2. Engine journal and last-look digest, with save migration and tests.
   *1 day*
3. The mini view in the dashboard sidebar, and the `map` command family.
   *1–2 days*
4. Contrast tests for the colour classes on all 11 themes, and panic-safe
   sizes. *1 day*
5. Retire `worldmap*` and update the wiki (`site/docs/`, commands). *1 day*
6. Optional: the capital-plan plate, to unify it with the citymap.
   *3–4 days*

## Risks, and what would make it feel wrong

- **Label clutter** is the failure mode of every text map. The prototype
  places labels with collision avoidance and priorities, and drops towns
  when zoomed out, but at 80×24 the engraved plate is busy. Production needs
  a label budget per plate and zoom.
- **Font coverage.** `⋀ ♜ ☄ ⚑ ♧ ≀` render in common terminal fonts, but a
  poor font could draw some of them wide. The HTML captures pin every symbol
  to one cell; a terminal cannot. There should be an ASCII fallback table per
  plate (one dial).
- **The world is invented.** Civ seats, trade partners, expedition
  destinations and scar sites are placed by the seed, not by the game, so
  the map is honest about *what* (a war, a route, a discovery) and only
  plausible about *where*. If the geography ever needs to mean something
  mechanically, the engine would have to own it.
- **Plate changes are a jolt by design.** An epoch shift redraws the whole
  map in a new medium. Leafing back through earned plates is what turns that
  into a collection rather than a loss.
- **Late-game sameness.** From Galactic on, the galaxy plate only grows your
  sector. Quantum and Transcendent could use their own dials (a fractured
  chart, a map of possible worlds).

## Captures (`captures/index.html` lists them all)

All of these come from smoke-bot states for seed 42 (`states/`), and seed 11
(`states-alt/`) is the second player.

- Plates by epoch: `primitive_age`, `bronze_age` (hide), `medieval_age` (portolan, with the Ironhold war front),
  `industrial_age` (engraved), `victorian_age`, `atomic_age` (survey, with the Nuclear Exchange scar), `digital_age`
  (satellite), `cyberpunk_age` (night pass), `space_age` (orrery), `interstellar_age` (star chart), `galactic_age`, `quantum_age`.
- Light themes: `light_medieval_age`, `light_industrial_age` (daylight), `light_primitive_age` (parchment).
- Small terminals: `small_100x30_*`, `small_80x24_industrial_age`. Mini view: `mini_40x15_*`.
- Animated: `anim_medieval_age`, `anim_digital_age`, `anim_cyberpunk_age`, `anim_space_age`.
- Colour degradation: `mono_industrial_age`, `ansi16_medieval_age`.
- The verb: `inspect_civ_industrial_age`, `inspect_route_victorian_age`.
- Plates and identity: `plates_industrial_on_portolan` (an earlier plate, leafed back to), `other_player_*`.

## Files

- `main.go`: flags, snapshot library. `gen.go`: the smoke-bot driver
  (greedy bot, deals on, scouting expeditions, a statesman for routes,
  alliances, a feud and embargoes). `states/`: the real game states it wrote.
- `world.go`, `noise.go`: the seeded planet. `model.go`: the atlas model
  from a state. `style.go`: the plates (dials) and colour classes.
- `render.go`: the planet renderer. `cosmic.go`: the sky plates. `canvas.go`:
  cells, z-order and labels. `scene.go`: layout, chrome, inspector, legend,
  mini view.
- `app.go`: the interactive tview app. `capture.go`, `captures.go`: the
  SimulationScreen captures. `atlas_test.go`: sizes, determinism, the
  interactive loop driven through a simulation screen, and a benchmark.

Run it:

```bash
go run ./lab/atlas -age medieval_age           # interactive
go run ./lab/atlas -age cyberpunk_age -theme daylight
go run ./lab/atlas -print -age atomic_age -w 100 -h 30
go run ./lab/atlas -capture lab/atlas/captures # every capture
go run ./lab/atlas -gen -seed 42               # regenerate states/ (states-alt/ is seed 11)
```
