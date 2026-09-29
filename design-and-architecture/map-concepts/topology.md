# Map Lab concept: topology

## The idea

Stop drawing the empire as a place and draw it as a system, the way a sysadmin sees a network in btop or a schematic. Every lineage you have built becomes a **producer** box on the left. Its wires carry packets into the **resource bus** in the middle, a single component with one port per store. From the bus, wires run to the city's **services** (citizens, research, construction, the wonder, the garrison, the market), and past them to the **external hosts**: civilisations, trade partners, expeditions, the harbinger and the next age. Packets move along the wires at a speed set by the real rate. When a store is full its inbound packets back up against the port. A war cuts a link, and a catastrophe sends an alarm pulse rolling out from the bus. A cursor can select any node, and the inspector then lists the node's parts and up to three commands, in the game's own command vocabulary, that act on it. The map shows the state of your economy and gives you something to do about it.

## Why it fits, and why it avoids the old failures

The map history (`map-history.md`) lists five things that went wrong. Here is how this concept handles each one.

| What went wrong before | Topology |
|---|---|
| Pixels pushed through `▄`, which never looked right | Everything is drawn in glyphs: box drawing, block and braille sparklines, gauges and arrowheads. Labels and art share one grid at one resolution. No images anywhere. |
| No verb, and the map was hidden | You can select any node with the arrow keys and press ⏎ to inspect it. The inspector lists up to three commands (`assign farm 3`, `build granary`, `diplomacy tribute ironhold_clans`, `trade route start …`, `advance`), and `1`–`3` stage one in the input line. A 40×15 **mini** form is meant to sit on the main screen as a sidebar. |
| Reviewed as PNGs | Every capture here comes from a tcell `SimulationScreen`. The `.txt` files are the exact cells and the `.html` files are those cells with their colours. |
| Static in a game that ticks | Packets move every frame, gauges and sparklines change every tick, the ticker scrolls and the alarm pulses. |
| 20.7k lines of bespoke per-age art | The prototype is one grammar in about 4k lines (tests, generator and capture tooling included). The ages differ only through seven sets of settings (`style.go`): line weight, packet glyph, gauge glyph, sparkline glyph, and so on. |

Checklist from the history report:

- **Glyph-native, labels on the grid:** yes.
- **Legible at 80×24:** yes, in the 3-column layout. At 100×30 it is still a readable diagram (`medieval_age-100x30`). Below 60×16 it switches to the mini bus view (`*-mini-40x15`).
- **256 and 16 colours, and monochrome:** colours are only emphasis. Meaning is carried by glyph shape: `▲`/`‼` for an alarm, `△`/`!` for a warning, `╳` for a cut link, stacked packets for backflow, and the text tags `FULL`, `IDLE`, `DRAIN`, `EMPTY`, `CUT`, `AT WAR` and `UNDERSTAFFED`. Proof: every `.txt` capture, plus `digital_age-mono.html`, which renders all colour classes as Text.
- **Theme colours from a few role classes:** every cell carries one of 12 foreground classes and 5 background classes (`canvas.go`), which map onto theme roles. The epoch tint is a single `Mix` into Accent and Border only. Light themes need no re-keying (`*-light.html`).
- **A verb:** select, inspect, and stage commands, as described above.
- **Glanceable plus a full view:** the mini bus view for the sidebar, the full diagram for the `map` command, and `w` for the world view.
- **Shows what the rest of the UI doesn't:** flows and where they go, backflow at full stores, understaffed lineages, idle workers, unused trade routes, wars, blockades, the harbinger, catastrophe pressure, and a "since you last looked" diff (`Δ last 13m: +37 built +68 pop` in the header, plus `+3` and `NEW` badges on nodes).
- **Changes on ticks:** yes.
- **Two players in the same age see different maps:** yes. The nodes are *your* lineages, stores, trade partners and wars. A player who never built Commerce has no Commerce box, and a player at war has a red cut link.
- **Under 5k lines, one grammar with per-age settings:** yes.
- **The Primitive age with 1–3 buildings looks good:** `primitive_age.html` has four big boxes on hand-drawn ASCII wires with pebble packets. Three or fewer lineages are always drawn as large boxes.
- **Deterministic placement that never reshuffles:** the bus lists stores in canonical resource order. Producers and services are placed by where they connect, and the order breaks ties stably. The same state always renders cell for cell the same (`TestRenderSizes` checks this). Nodes can slide a few rows as new ones appear, but they never change order.
- **One metaphor for city and world:** `w` zooms out. The whole city folds into one node, and every civilisation (met or uncharted), trade partner, expedition and harbinger links to it with the same wire grammar (`*-world.html`). This view replaces the worldmap.

## What the player sees and does

- **Layout (≥120 columns):** four columns: `producers │ bus │ services │ hosts`. Wires run only between neighbouring columns, through channels between them, so the diagram always reads left to right: *made → stored → used → outside*. Imports and diplomacy flow right to left.
- **Layout (60–119 columns):** three columns. The hosts fold into a "World" list box in the services column.
- **Below 60×16:** the mini view is the bus alone. Each store has a four-cell flow lane in front of it: packets crawl in at the inflow speed, pile up when the store is full, and run backwards while it drains. The worst alarm sits on the bottom line.
- **Nodes:** producer boxes are sized by their share of your buildings: 5, 4 or 3 rows, or a one-row pill. Each is aligned so its output wire runs level into the port it feeds most. A box shows a sparkline of its attributed output, the rate, a staffing gauge (`93/93`) and its newest building's name.
- **Wires:** packets move along the real path. Their speed is log-scaled from the real rate, so both a trickle and a flood read correctly in any age. A crowded channel bundles wires onto a few shared trunks, like a schematic bus. Each wire's packets still follow their own path. Minor side outputs (under 10% of a lineage's output, into a store someone else also feeds) are left to the inspector.
- **Signals:**
  - A full store: amber `FULL`, with packets queued against the port. The inspector says whether the overflow banks into the wonder or is lost.
  - A store draining: red `DRAIN`, or `EMPTY` if it runs dry within a minute.
  - Understaffed lineage: amber box and `UNDERSTAFFED`.
  - Idle workers: amber `IDLE`.
  - A market with startable routes and none running: `2 ROUTES UNUSED`.
  - A civ at war: red box, and the link cut with `╳`.
  - A route blockaded: `CUT`.
  - A pending catastrophe: a red ticker, and an alarm pulse that leaves the bus and rolls outward through producers and services to the hosts every 12 frames.
- **Keys (prototype):** `←↑↓→` select, `⏎` inspect, `1`–`3` stage a command, `w` world view, `d` toggle the diff, `t` cycle themes, `a`/`A` cycle ages, `q` quit.

## How it reacts to the game

- **New building:** the lineage's count and size change, and on the first build of a lineage a new box appears with `NEW`.
- **Age advance:** the settings switch at epoch boundaries (see below). New resources add ports, and the Next-age host resets.
- **Catastrophe:** the alarm cascade and ticker described above.
- **Trade:** export wires run from store to market, imports come back from market to store, and route hosts hang off the market.
- **Diplomacy:** civs appear as hosts when first met. The opinion gauge runs from −100 to +100, and allies carry packets.
- **Expeditions:** a host linked to the garrison while the expedition is away.
- **Idle time:** the header diff and the `+N`/`NEW` badges say what changed since your last look. The prototype diffs against the state 300 ticks earlier. In the game, the diff would be taken against the snapshot from your last visit.

## Each epoch

| Epoch | Name on screen | Wires / boxes | Packets | Gauges | Sparklines |
|---|---|---|---|---|---|
| Stone | TALLY STONES | ASCII `+-|` | pebbles `.o` | `#.` | `_.-~^` |
| Iron | SCRIBE'S LEDGER | rounded ink `╭─╮` | `·•` | `▰▱` | blocks |
| Steel | TELEGRAPH EXCHANGE | light wires in double-ruled cabinets `╔═╗` | punched `▪` | `▮▯` | blocks |
| Electric | SWITCHBOARD | heavy panels `┏━┓`, light cable | current `●∙` | `█░` with partial cells | blocks |
| Digital | NETWORK TOPOLOGY | thin fibre `┌─┐` | lit pulses (`━━━` travelling along the fibre) | fractional blocks | braille |
| Neon | NEON GRID | rounded wires, double boxes | `◉•` | `▰▱` | braille |
| Cosmic | QUANTUM LATTICE | dashed entanglement `╌╎` | `◆` in superposition (a ghost `◇` also appears further along, flickering) | `━╌` | braille |

The bus is renamed each epoch too: STORES, GRANARY LEDGER, COUNTING HOUSE, BUS BAR, CORE SWITCH, FUSION CORE, SINGULARITY. The epoch accent from `theme.AgePalettes` is mixed into Accent and Border, 35–50%, checked with `theme.Legible`.

## Themes and small terminals

All colour comes from `theme` roles through classes. `daylight` and `high_contrast_light` work unchanged (`medieval_age-light`, `digital_age-light`, `industrial_incident-light-100x30`). The layout scales from 200×60 down to 80×24 in three-column mode, then switches to the mini view below 60×16. The test renders every fixture at 200×60, 160×48, 120×36, 100×30, 80×24, 40×15 and 30×10.

## Performance

A full rebuild (model, layout, routing, draw) takes about **0.68 ms** at 160×48, and an animation frame on a cached layout about **0.38 ms** (Apple M5, `BenchmarkFrame`/`BenchmarkAnimate`). In the game the model and layout would be rebuilt only when the state's signature changes (lineages, counts, ports, hosts or size), and each 100–150 ms animation tick would only redraw. That is cheap enough to animate on the main screen.

## Integration

- **Replaces both `citymap` and `worldmap`.** `map`/`citymap` opens the full topology, and `worldmap` opens it zoomed out (`w`). One package of roughly 3k lines would replace the 20.7k in `ui/citymap/`.
- **Main screen:** the mini bus view as an optional sidebar panel (40×15), or as a replacement for the resources panel, since it is the resource list plus flow.
- **Engine hooks needed (read-only, all small):**
  - a per-lineage production breakdown (the prototype re-derives it from building counts and staffing with the engine's formula, `0.20 + 0.80 × fill`);
  - the snapshot from the player's last visit, for the diff;
  - a trade-route → civ link, if routes should hang off their civ rather than the market.
- **Commands:** the inspector's suggestions are real commands. In the game, `1`–`3` pre-fill the command input, and the player still presses Enter.

## Effort to productionise

About 2–3 weeks for one engineer:

- Move `lab/topology` to `ui/topology` behind the overlay/widget contract (`Refresh` stores the snapshot, the draw reads it), with a layout cache keyed on the state's signature.
- Wire in input, the sidebar mini view and staging commands into the input line.
- Add the engine's per-lineage breakdown and the last-visit snapshot.
- Write golden tests from the capture tool.
- Tune per-epoch glyphs against real fonts: `◉`, `◆` and braille are the risky ones on Windows Terminal.
- Update the docs.

Deleting `ui/citymap` is part of the win.

## Risks, and what would make it feel wrong

- **"It's a spreadsheet, not a map."** Some players want to *see their town*. This concept deliberately trades that for information. Its delight has to come from motion, era settings and seeing your own economy's shape. If the owner wants a *place*, this concept isn't that.
- **Wire clutter at late ages.** 15+ lineages into 17 stores produce crossings. Bundling, dropping minor outputs and level alignment keep it readable at 160×48 (`cyberpunk_age.html`), but at 100×30 the late game gets dense. A production version might collapse the smallest producers into an "other" pill.
- **Sizing and attribution.** Box sizes follow building counts, because output isn't comparable across resources (a gold/t is not a food/t). Per-lineage output is re-derived with the engine's formula but ignores bonuses. It is honest for *share*, not for exact totals.
- **Font support.** Box drawing and blocks are safe. Braille and a few geometric glyphs (`◉ ◆ ◈`) depend on the font. Each epoch's settings can fall back to ASCII (the Stone era already is ASCII).
- **Stability vs. optimal layout.** Order never changes, but inserting a node can shift its neighbours by a few rows. Hard slotting (fixed rows per lineage) would remove that at the cost of wasted space.
- **What's staged in the fixtures:** the `industrial_age_incident` fixture is the bot's real run plus two real player actions: raiding a civ's route until it declares war (`diplomacy raid`), then the dev console's `/catastrophe`. Everything else is exactly what the smoke bot produced (seed 7).

## Running it

```bash
go run ./lab/topology -gen                    # replay the smoke bot into lab/topology/fixtures (a few minutes)
go run ./lab/topology -age medieval_age       # interactive; a/A cycles ages, t cycles themes
go run ./lab/topology -capture lab/topology/captures
go test ./lab/topology -bench .
```
