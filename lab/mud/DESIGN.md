# Map Lab: mud, the city you walk through

## The idea

The settlement is a small text world. Fifteen places (the square, the homes, the fields, the woods, the quarry, the forge quarter, the works, the temple, the academy, the market, the harbour, the stores, the barracks, the wonder site and the gate) sit on a fixed compass grid joined by roads. Each place collects the buildings of one or two lineages. You stand in one place at a time. `look` prints a room in the MUD tradition: a paragraph built from the real state, then what stands here, what it makes, who is about, and the exits. Above the text are a small ASCII scene of the place and a box-drawn map of the whole graph, `♣Timber───⌂Lanes────@Forum────$Guilds`. You move with the arrow keys or by typing `north`, `go forge`, `visit great library`. Places appear when their first building goes up, and they are renamed each epoch (the Village Green, the Forum, the Piazza, the Plaza, the Concourse, the Arcology Atrium, the Orbital Ring). After orbit the compass becomes hubward, spinward, rimward and antispinward. Past the gate, `out` shows the known world: every civilization you have met, with bearing, distance and mood. That makes one metaphor for both the city and the world map.

## Why it suits a terminal, and why the old maps felt clunky

Every earlier map tried to be a picture: RGBA through `▄` half-blocks, sprites, photos, 20k lines of per-age art. Each one was a passive overlay you looked at and closed. This concept doesn't try to be a picture. Text is the medium the terminal is best at, and MUDs and interactive fiction proved decades ago that a place described in prose, with a few lines of art, stays with a player better than a blurry bitmap. The map here is a graph with labels on the same character grid as everything else, and nothing is rasterised.

It also has verbs. The walk speaks the prompt's language: names resolve the way `build <key>` does, `visit farm` walks you there, and `x ironworks` shows the building's card (flavor line, description, hands, legacy status, key). The history agent's review found that "no verb" was one of the two root failures.

## What the player sees and does

Captures are listed in `captures/index.html`; paths below are relative to `lab/mud/`.

- **The room view** (the full panel, e.g. `captures/walk_medieval.html`):
  - title bar: place, age, epoch, time of day and weather;
  - scene (left): sky, a row of building stamps, the ground;
  - map (right): the room graph, `@` for you, goods moving along the roads;
  - the room paragraph;
  - `Here:` buildings by count, legacy ones marked `(old)`, ruins in red, and a `Hands` total;
  - `Makes:` the resource the place produces, its rate and a fill gauge `▰▰▰▱▱` (`full` shows in warning colour, so a capped resource shows up in the place that is wasting it);
  - per-place extras: Folk and morale in the square, the current study in the academy, shelf fill in the stores, the garrison, the wonder being raised, routes at the harbour;
  - `Also here:` the harbinger, envoys and raiders at the gate, cargo, idle workers;
  - `Exits:`;
  - a transcript of recent commands and replies, which fades as it ages, as in a MUD client;
  - the prompt, and key hints.
- **Commands:** `look`/`l`, `n e s w` and arrows, `go <direction|place>`, `visit <building>` (walks the shortest road and examines it; the reply names the places you pass through), `x <building>`, `survey` (every place on one screen: totals, newest building, short-handed, ruined), `out` (the world, from the gate), `away` (what changed), `help`, `q`.
- **The compact form** (`captures/mini_medieval_40x15.txt`, 40×15): the graph with a count and marks under each place (`!` short-handed, `x` ruins, `?` harbinger, `×` war, `→` expedition out, `*` changed since your last look), plus one line of news. It fits a dashboard corner. Enter opens the full walk.

## How it reacts to the game

- **New buildings:** the count moves and a stamp is added (stamps grow sub-linearly: 1 + log2(n)). The room says "The guildhall is new; the scaffolding is still leaning against it" (taken from the log's `Build complete` lines).
- **Age changes:** places get renamed. Stamps are chosen by the era the building comes from, not the current one, so the old huts keep their thatch next to the new tenements (see `captures/whatif_war_ruins_industrial.txt` frame 3). The square changes its landmark every epoch: fire, forum, fountain, clock tower, glass concourse, neon atrium, ring hub. Each wonder has its own silhouette, so the wonder walk reads as your own history.
- **Catastrophes:** ruined buildings are listed in red, a ruin stamp appears, and the room mentions the shell the catastrophe left. The harbinger stands in the square as a figure, and its own flavor lines (`HarbingerView.Lines`, from the flavor generator) become part of the square's paragraph.
- **Trade, factions, war:** active routes show as a caravan at the gate and as cargo at the harbour or market. Goods flow along the market–harbour road. Envoys wait at the gate. At war, the gate gets a barricade, turns red on the map, gets `×` in the compact form, and the room says the watch counts everyone twice. `out` lists every civ you know.
- **Flows and bottlenecks:** a `•` travels each road from a producing place toward the stores. A place staffed below 30% sends nothing, so a still road is a bottleneck you can see. Short-handed places turn amber.
- **Time passing:** the hour and the weather are cosmetic. The hour comes from the tick (a 40-minute day), and the weather is seeded per day and bent by events (drought, festival); the sun crosses the sky. Smoke rises, clouds drift, water moves and stars twinkle, all per frame. The room's sentence choice is stable for about ten minutes, then changes, so a second `look` does not reword itself and a later one does.
- **Idle return:** on opening, the walk shows *While you were away (2h 08m)*: new places, renamed places, per-place deltas (`The Forge Quarter: +11 Forge`), new ruins, harbinger arrivals and the log's headlines. Changed places carry `*` on the map until you visit them, and each room gets a `Since you last looked` line (`captures/away_medieval.html`).

## Each epoch

One grammar with four stamp bands (ancient, industrial, modern, orbital) plus per-place ground textures. Only the square and the wonders have landmarks per epoch or per wonder. Captures: `primitive_age_first` (four minutes in: a fire, the Wild Man, huts faint to the west), `bronze_age`, `medieval_age`, `industrial_age`, `digital_age`, `cyberpunk_age`, `interstellar_age`, `galactic_age`. In orbit the ground becomes hull plating, the sky is always stars with a planet's limb, weather becomes solar storms and meteor showers, and the compass words change.

## Themes and small terminals

All colour comes from theme roles through a handful of glyph classes: roofs get Accent, bodies Text, lights and fire Highlight, water Label, plants Positive, stone and ground Dim, ruins Negative, and lit windows switch from Dim to Highlight at night. Chrome uses Border, Label and Selection. A theme switch needs nothing special. Light themes are in `medieval_age_light` (Daylight) and `industrial_age_light` (Parchment). Meaning survives monochrome: every place has a distinct glyph (`⌂ ≡ ♣ ▲ ¤ # † § $ ≈ ▪ ‡ ★ ∩ ○`), marks are glyphs (`! x ? × * @`), and gauges are shapes (`▰▱`). The `.txt` captures are the monochrome proof. Emoji-width glyphs were avoided on purpose. tcell maps truecolor down to 256 or 16 colours, and since the roles are ordinary theme colours this degrades like the rest of the UI.

Layout adapts: map labels shrink from 7 to 3 characters to fit, and the scene takes whatever width is left. `medieval_age_80x24` and `industrial_age_100x30` show the room view at small sizes. Below 60×18 the panel falls back to the compact form. A test draws every state at every place at sizes from 20×8 to 200×60.

## Performance

Building a `City` from a `GameState` is one pass over the building map, then a BFS over fifteen nodes. Drawing is a few thousand `SetContent` calls with no allocation-heavy work (the scene grid is at most a few KB). Build the `City` once per state refresh and redraw it per frame. That is far cheaper than the half-block renderer.

## Integration

- **Replace both `citymap` and `worldmap`.** The worldmap becomes `out` from the gate, and a later version can make each civ a room you visit (backstory, deals, opinion).
- **New panel:** `walk` (aliases `city`, `map`, `citymap`), and the compact form as a dashboard widget where the minimap sits now.
- **Coexisting with the prompt.** The dashboard prompt already owns `s` (status), and `n/e/w` would be fragile there, so movement lives inside the walk panel, which has its own input line and arrow keys and gives Esc back to the dashboard. The dashboard prompt gains `look [place]` and `visit <building>`. Both open the panel at that place and use a new `ArgPlace` kind in `ui/commands.go` for completion. `x`/`examine` in the panel reuses `ArgBuiltBuilding`. Nothing shadows an existing command.
- **Prose:** move `catalog.go` into the `flavor` package as new Moments (`RoomAmbience` per place and epoch, `RoomCondition`), so the prose gets the package's burstiness, register and no-AI-tells tests and its `Stream` for anti-repeat. The prototype already follows that package's contract: whole authored sentences, at most one noun-phrase slot, no markup, flat by default, and never a sentence about the mechanics. Mechanics are printed separately in `Here`/`Makes`. It also calls `flavor.Line` directly for returned expeditions at the gate.

## Effort to productionise

About 2–3 weeks for one person:

- **Panel, dashboard widget and command wiring:** 4 days.
- **Moving the prose into `flavor`, expanding it to three to five sentences per place and epoch, and tests:** 5 days. This is where the quality lives.
- **Place visits for civs (the world view):** 2 days.
- **Stamp review with Adam in a real terminal, plus light-theme passes:** 2 days.
- **Docs (`site/docs/commands.md` and a wiki page):** 1 day.

The prototype is about 3.6k lines including capture tooling and the generator. The production code should land well under 3k.

## Risks, and what would make it feel wrong

- **Prose fatigue.** The whole idea rests on the sentences. If the catalog is thin, the fifth `look` reads like a slot machine. The current catalog has one establishing line per place and epoch and small condition banks: enough to judge the idea, not enough to ship. It needs a writer's pass and the flavor package's `Stream`.
- **It is slower than a glance.** Walking is a verb, and some players want a picture. The compact form and `survey` cover the glance, but if Adam wants a spatial city he can watch grow, this is not that.
- **Fixed geography.** Every player's graph has the same shape, so two players in the same age differ only by which places exist, their counts, names, stamps and prose. Their layouts are the same. A seeded shuffle of the non-spine places would help, but not at the cost of stable muscle memory.
- **Dead rooms.** The bot never builds a harbour or a temple in some runs, and a few late-game places list only `(old)` buildings. The prose handles it, but the absence is visible. That may be a feature, since it is your civ.
- **Numbers.** `Makes` lines at 940.6M/t read like a spreadsheet inside a story. Keep them one line and in Dim/Label so the paragraph stays the room.

## Requirements checklist (from the map-history review)

| Requirement | How |
|---|---|
| Glyph/box-drawing native, labels on the grid | Everything is characters: stamps, graph and text share one grid. |
| Legible at 80×24, degrades to 16 colours and monochrome | `medieval_age_80x24`; distinct glyphs, marks and shape gauges; `.txt` captures. |
| Theme colour from a few role classes | Glyph classes mapped to roles (`roleFor`); two light captures. |
| A verb in the command vocabulary | `go`, `visit`, `x`, `survey`, `out`, `away`; names resolve like `build <key>`. |
| Glanceable compact form plus a full view | `mini_*_40x15` and the room view. |
| Shows what the rest of the UI doesn't | Goods flows, still roads, short-handed places, `full` gauges where the waste happens, war at the gate, and what changed since your last check-in. |
| Changes on ticks | Smoke, clouds, water, stars and flow pulses per frame; prose rotates about every ten minutes; the hour moves. |
| Two players in one age see different maps | Different places exist, with different counts, legacy mixes, harbingers, civs, weather and prose. The graph shape is the same (see Risks). |
| One grammar, per-age dials, well under 5k lines | Four stamp bands plus landmarks; about 3.6k lines with tooling. |
| Primitive age with 1–3 buildings looks good | `primitive_age_first`: fire, harbinger, huts in the distance, a three-place map. |
| Deterministic, stable placement | Fixed grid positions. Places only appear, never move; the frame grows. |
| One metaphor for city and world | The world is the room past the gate. |

## Running it

```
go run ./lab/mud -gen                          # smoke bot, seed 7, writes states/*.json.gz (about 8 min)
go run ./lab/mud -age medieval_age             # walk it (arrows, type commands, Tab toggles the compact form)
go run ./lab/mud -age medieval_age -since medieval_age_early   # idle return
go run ./lab/mud -age galactic_age -theme daylight
go run ./lab/mud -capture                      # rewrite captures/
```

States are real: the smoke bot (`smoke.NewBot`) plays seed 7 with endure-on-catastrophe. A small layer adds the two things a real player does that the greedy bot doesn't: it sends a scout when none is out and opens trade routes, through the same engine calls the commands use. The run never reached war or ruins, so `whatif_war_ruins_industrial` edits the real Industrial state (one civ at war, a third of three building types ruined) to show those paths. It is labelled as a what-if.
