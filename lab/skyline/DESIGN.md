# Skyline: the empire as a BBS-art panorama

## The idea in one paragraph

Stop drawing a map. Draw the city side-on, the way a 1990s BBS login screen drew
a city: a wide ANSI panorama in CP437 blocks, shade glyphs (`░▒▓█`) and line art,
with a dithered sky, far ridges, a hazy distant town and the real city in front.
It reads left to right like a timeline. Each age you have lived through is a
quarter of the skyline (huts and a longhouse in the west, then ziggurats, a
temple row, the castle town, domes, brick and smokestacks, deco towers, glass,
neon, arcologies, saucer habitats and spires), with the present at the east end
and your build queue as cranes on the frontier. Every silhouette is a building
you own. The skyline's density and height are your economy at a glance. Light
means labour: windows light up and chimneys smoke only where workers are staffed,
so a dark mill is an idle mill. The sky runs a day/night cycle on the game tick,
with weather from active events. The harbinger stands on the ridge, and a
catastrophe sets the roofs on fire.

## Why it fits a terminal game, and why it avoids the old maps' problems

The history report names five failures. This is how the skyline answers each.

| What went wrong before | Skyline |
|---|---|
| RGBA images pushed through `▄` half-blocks | Everything is drawn on the character grid: CP437 blocks, shade and box glyphs, and ~30 building forms written as cell drawings. Half-blocks are used as ANSI artists use them, for roof slopes, domes and ridgelines, never as a pixel framebuffer. The chrome, the harbinger's name and the inspect labels share that grid. |
| No verb, hidden, passive | `i` puts a cursor on the skyline. The status line names the building under it with its count, its workers (`12/15`, or `idle`) and the exact `build <key>` command with its main cost. The minimap strip scrolls you anywhere in one keystroke, and the compact view sits on the main screen. |
| Reviewed as PNGs | Every capture here is a tcell `SimulationScreen` read back cell by cell into `.txt` and `.html`. |
| Static in a ticking game | The clock comes from the tick. Smoke and windows follow staffing, and traffic follows population, trade routes and wars. Beacons blink, windmills turn, ships sail the bay, and weather follows events. |
| 20.7k lines of bespoke per-age art | One grammar: ~30 parametric forms plus 22 wonder routines. An age is a set of dials: material family, sky keyframes, road skin, mid-town roofline and form table. The render core is ~5.5k lines, of which ~1.7k is art (forms + wonders); the lab adds ~1k lines of tooling (generator, captures, viewer, tests). That is over the 5k target; the forms and wonders are where it would be trimmed. |

It's terminal-native in a way a top-down map isn't. ANSI art always was city
skylines, logos and sunsets on a 16-colour palette, so this is the medium's own
genre. A skyline also degrades well. Squeeze it to 40 columns and it's still a
skyline; drain the colour and the silhouettes still say "castle", "factory",
"tower".

## What the player sees and does

**Full view** (`citymap`, or a key from the dashboard):

- A header bar in BBS style (`░▒▓█ INDUSTRIAL AGE █▓▒░`) showing the era, the day and clock, population, the building count, **"▼ 14 new since your last visit"**, the harbinger's name, and any pending catastrophe.
- The panorama, in back-to-front layers: sky, stars and sun/moon, clouds, far ridge, then near hills or the distant town (at 0.4–0.6 parallax), three depth rows of real buildings, smoke, road, bay, and traffic.
- A **minimap strip**: the whole skyline as one row of `▁▂▃▄▅▆▇█`, with the viewport bracketed, age boundaries ticked, and new buildings in the positive colour.
- A status line with key help, or the inspected building.

**Keys** (all working in the prototype): `← →`/`h l` scroll, `H L`/shift jump
half a screen, `Home`/`End` go to the oldest quarter or the present, `[ ]` step
through saved states (ages), `i` inspects (`← →` then move the cursor), `n` sets
the time (auto, dawn, noon, dusk, night), `w` sets the weather, `t` changes the
theme, `6` switches to 16 colours, `c` toggles the change markers, `m` toggles
the compact mini view, `space` pauses, and `q` quits.

**Compact view** (40×15, sidebar-sized). The whole city is squeezed into a
sparkline skyline: eighth-block tops give each column eight height steps, and
facades use the tallest building's real material. It carries the same signals as
the full view: lit windows (staffed), smoke (producing), `▼` (new), `♦`
(wonders), the harbinger on the horizon, fire in a catastrophe, caravans on the
road, and a one-line header with the age, clock, new count and **idle producer
count**.

## How it reacts to the game

| Game state | On the skyline |
|---|---|
| Buildings (`Buildings[k].Count`) | Silhouettes grow sub-linearly: 1:1 up to 3 copies, then about √n×1.15, capped at 6 per type. A type's height grows with log(count). |
| Workers (`WorkersAssigned/WorkerCapacity`) | Staffing sets the fraction of lit windows at night and whether a producer's chimney smokes. For housing, the population fill sets its lights. An idle producer is visibly dark and smokeless, and `inspect` labels it `idle`. |
| Age advance | A new quarter opens at the east end, and the camera's "present" moves there. Old quarters keep their period architecture, but the road through the whole city is re-skinned to the current age (dirt, cobbles, rails, asphalt, neon, maglev), so history stands on today's street. Street lamps arrive in the Victorian age. |
| Build queue | Cranes and scaffolding on the frontier, one per queued item. |
| Trade routes (`ActiveRoutes`) | A caravan, train, truck or maglev in trade gold on the road for each route, plus ships in the bays. |
| Factions | Every civilisation you have met stands on the far ridge as a small town of its own, sized by its strength, with a pennant in a fixed relation colour (war red, rival orange, ally green, friendly/trading gold, neutral steel) and its name on the grid. A civ at war also sends a red-bannered warband marching in from the west. |
| Harbinger | A cloaked figure on the far ridge with its name on the grid. The horizon bruises red as the catastrophe odds rise (`Pressure`). |
| Pending catastrophe | An overlay per epoch: roof fires with black smoke (barbarians, industrial collapse), cracks in the sky (meltdown, reality fracture), or glitch bands with the grid down, so almost every window goes dark (digital collapse, solar event). From the Digital Era the harbinger itself is a flickering cyan hologram. |
| Events | Storms bring rain and lightning; floods, plague and blight bring rain; drought clears the sky. Otherwise the weather is seeded per in-game day. |
| Idle time | The clock is `tick`-driven (a day is 1,800 ticks, one hour at 1x). A returning player sees the time of day move on, and the `▼` markers show what was built while they were away (diffed against a check-in snapshot). |
| The save's seed | Placement jitter, material variants, mirroring and ridge shapes all hash from `GameState.Seed`, so two players in the same age get different skylines from different building mixes, counts and seeds. |

**Stable placement.** A lot is a pure function of (seed, age quarter, catalogue
slot of the type, copy index). Each quarter's catalogue is sorted tallest-first
and spirals out from the quarter's centre, so the first copy of each type fills
the front row and later copies go behind. Building something never moves anything
else (enforced by `TestPlacementIsStable`). A building that grows taller stays
centred on its lot.

## How it looks at each epoch

- **Stone:** a camp in a big landscape. Conical thatch huts with lit doorways, tents with cooking fires, a story circle of `☻` round a fire, standing stones, conifer woods on the hills, the Sacred Grove's giant tree and the black Monolith.
- **Iron:** mudbrick blocks, a stepped ziggurat, marble temples with pediments and columns, the ruined-tier Colosseum and Parthenon, and a curtain wall on the far hill. Then the castle town: towers, keeps, timber-framed houses and cathedral spires.
- **Steel:** ochre and copper domes, windmills, a bay with sailing ships and the striped Grand Lighthouse, then brick, sawtooth roofs and smokestacks under a smoggy sky, the glass Crystal Palace, and trains on the rails.
- **Electric:** the lattice Eiffel tower, deco setback towers with spires and beacons, cooling towers breathing steam, oil derricks and pylons.
- **Digital:** glass towers with banded floors, masts, data centres with LED rows, satellites at night, and the World Simulation cube with data rain.
- **Neon:** dark megatowers with neon edges and vertical signs, hover traffic threading between towers, a magenta night, the Citadel, launch gantries, the orbital ring and a space elevator.
- **Cosmic:** saucer habitats on stems, arcologies, spires with light rings, a ringed planet in a sky that shows stars by day, the Warp Nexus ring gate, and the singularity core.

## Light themes, monochrome, 16 colours, and small terminals

- **Colour is roles, not pixels.** Each frame builds one small palette table (`Pal`): material slots × depth, cached. Cells name a (material, slot, depth) and look up their colour, so a theme switch, a time of day, or 16-colour mode costs nothing per cell. The chrome uses theme roles only: Accent, Label, Dim, Highlight, Positive, Warning, Surface and Chip.
- **Light themes** get a daylight sky lifted toward the page colour. Night becomes a lavender blue hour, not a black hole in a white UI, with silhouettes in ink and lit windows still warm. See `light_daylight_*` and `light_hcl_medieval`.
- **Monochrome and Parchment** remap the whole scene to a duotone between the theme's background and text colours. Shape carries everything (`theme_monochrome_digital`).
- **16 colours** quantise the palette to the CGA/VGA ANSI set, the palette this art style was invented for. The sky's `░▒▓` band transitions become true ANSI dithers (`ansi16_*`). 256-colour terminals get tcell's own downsampling.
- **Small terminals.** Building heights scale with the scene height, and wonders are procedural and shrink with the scale. At 80×24 it is still a skyline with a sky; at 40×15 the compact renderer takes over.

## Performance

Measured on an M-series Mac at 160×45 in the Cyberpunk age (~615 buildings): about
**0.3–0.9 ms per frame** and ~0.5 MB allocated per frame (the frame buffer and
palette cache), plus **2.7 ms** to rebuild the world. The world is rebuilt only
when the state's buildings change or the terminal resizes; per tick, only the
snapshot fields change. The animation runs at about 8 fps, independent of the
game tick. Easy wins for production: reuse the frame buffer, and cache the
static layers per (camera, time bucket).

## Integration

- **Replace both `citymap` and `worldmap` with one metaphor.** The skyline is the city view, and the world view lives on its horizon: discovered civs are towns on the far ridge in relation colours (implemented), warbands and caravans come and go along the road (implemented). A production world view would widen this into its own panorama of those ridge towns. There is one renderer and one grammar.
- **Commands.** `citymap` (or `skyline`) opens the full view. `citymap <building>` jumps the cursor to that building. On the main screen, the 40×15 compact view replaces the current map panel. The inspect status line always prints the real `build <key>`, so the view teaches the command vocabulary.
- **Data.** Everything comes from `GameState`, and the renderer takes the snapshot only (no engine calls, so it's lock-safe). "New since your last visit" needs one thing the game doesn't have yet: a building-count snapshot saved at session end (a small addition to the save), which the renderer diffs against.
- **Code.** `ui/skyline/` with `palette.go`, `sprite.go`, `arch.go`, `catalog.go`, `wonders.go`, `layout.go`, `render*.go`, `compact.go` and `chrome.go`, and a tview `Primitive` wrapping `render()` into the screen.

## Effort to productionise

About **2–3 weeks for one engineer**:

1. Port to `ui/skyline` as a tview widget with the lock and refresh contract (2 days).
2. An art pass on the forms that are still weak: the dam, the accelerator, the late energy buildings, and more cosmic variety (4–5 days).
3. Harbour placement, far-ridge factions for the world view, and the save-side check-in snapshot (3 days).
4. Swap the dashboard panel for the compact view, add `citymap <key>`, and wire the key bindings into the help (2 days).
5. Tests: exact-size, panic, determinism and stability (already written here), plus golden captures per age (1–2 days).

## Risks, and what would make it feel wrong

- **Art quality is the whole bet.** A skyline with ugly forms is worse than a list. The weakest pieces now are a few wonders (Hoover Dam, particle accelerator) and the sameness of late-era towers. It needs a real art pass and a reviewer with taste, using the sprite sheets (`captures/sprites_*.html`).
- **Chronological quarters mean the present can be thin.** Just after an age advance, the east end shows only a few new buildings. The camera opens on the present, so the first impression after an advance is "small". Mitigations: the frontier cranes, the minimap, and `Home` for the old town. An alternative is a "current-age-centred" camera that also takes in the previous quarter.
- **Fonts.** The HTML captures show hairline seams between half-block cells. Real terminals mostly don't, but some fonts render `▀▄` a pixel short. The art relies on half-blocks for slopes.
- **Scale.** A max-level empire is ~1,000 buildings and ~5 screens of panorama. That's good for scrolling but not glanceable at full size, which is why the compact view and the minimap exist.
- **It is a picture first.** The verb (inspect, `build <key>`, idle markers) makes it useful, but it will never be as dense as the economy tab. It should sell the fantasy and flag problems (dark districts, the harbinger, fires), not replace the numbers.

## Running the prototype

```bash
go run ./lab/skyline -gen                 # replay the smoke bot (seed 7) and save states/ (about 10 min)
go run ./lab/skyline -age industrial_age  # interactive viewer (keys above; [ ] walks the ages)
go run ./lab/skyline -captures            # rewrite captures/ (index.html lists them)
go run ./lab/skyline -sheet lab/skyline/captures/sprites   # sprite sheets of every form and wonder
go test ./lab/skyline                     # exact size, determinism, stable placement
```

The states in `states/` are real `GameState` snapshots, taken from one greedy-bot
playthrough (seed 7, trade routes opened as a player would). Each is taken at the
moment its age became ready to advance, so it is as built-up as that age gets.
Alongside them are `prev_<age>` (about two hours of play earlier, standing in for
the last check-in), `primitive_early` (3 buildings), harbinger sightings, and one
Digital Era catastrophe triggered through the dev force hook.
