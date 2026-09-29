# Map Lab: organism

## The idea

Your civilisation is drawn as one tree, in the tradition of `cbonsai` and
ASCII tree art: one glyph per cell, no pixels. Each production lineage is a
limb, and it owns a wedge of the crown as wide as its share of the economy.
It grows one branch segment per building or so, and its leaves are the
lineage's own glyph (`&` food, `%` timber, `#` stone, `◊` metal, `¤` craft,
`!` energy, `*` knowledge, `♪` culture, `+` faith, `$` trade, `ω` harbor,
`^` military, `@` net, `o` orbit). How full a limb's leaf clusters are shows
how well it is staffed. Wonders hang as golden fruit, and the age's unbuilt
wonder ripens `○ ◔ ◑ ◕ ●` as its bank fills. Housing makes the trunk, so the
trunk's girth is your population. Storage grows the roots, and the roots
reach down through **strata of your history**: one coloured layer per epoch
lived, newest on top, with that epoch's events buried in it as fossils (`✧`
blessings, `†` hardships, `✹` catastrophes endured, `✕` succumbed, `◉`
harbingers). Other civs stand on the horizon as small conifers, coloured
green for friends, amber for wary, red at war, and grey otherwise. Trade
routes are pollen drifting between the crowns. A pending harbinger is a
comet with a tail as long as the odds, and a pending catastrophe is
lightning into the crown under a reddening sky. One view covers both the
city and the world.

It is a map of *what your civ is*, not *where things are*.

## What I considered, and why this one

| Candidate | What it's good at | Why not alone |
|---|---|---|
| L-system / root network | organic, stable growth, easy to read "big limb = strong lineage" | on its own it has no history and no world |
| Cellular-automaton culture | alive and hypnotic | can't answer "what am I strong in?"; its state is noise to the player |
| Reaction-diffusion or flow field | beautiful | same problem, plus the cost per frame; it's a screensaver, not a readout |
| Tree-ring / geological strata | history at a glance, very "terminal" | static, and says little about the present economy |
| **Tree above, strata below** (chosen) | the present economy (crown), population (trunk), history (strata), the world (horizon) and threats (sky) share one metaphor | stops being literal about space; see Risks |

The tree and the strata are one idea, not two: a living thing whose roots
grow through its own past. A generative CA or flow field was the tempting
"boldest" choice. I rejected it because the brief says the art must still
*communicate*, and a CA can't tell you that your metal lineage is
understaffed.

**A pivot worth recording:** the first version drew the tree with braille
(2×4 dots per cell). In a real font, braille reads as dotted mush, and it
breaks the history report's "one resolution per frame" rule (labels on a
coarser grid than the art). Cell-native glyphs were better on every axis:
crisper, readable in monochrome (leaf *shape* names the lineage), and at
home next to the rest of the UI's text.

## Why it fits a terminal, and avoids what made the old maps clunky

The history report blamed seven iterations on five things. How this concept
answers each:

1. **Wrong renderer (RGBA through `▄`).** Every cell is one glyph with a
   foreground colour: branch glyphs `| / \ ~ ─ │ ╱ ╲`, leaf glyphs, and text.
   Labels share the grid with the art. There are no images, sprites or
   half-blocks.
2. **No verb, hidden overlay.** Inspecting is the verb: `←/→` or `tab`
   selects a limb, the rest of the tree recedes, and a card lists its
   buildings *by their command keys* (`14× build ironmonger`), its workers
   against capacity, and its output. The spectrum bar, labels and ticker use
   the same lineage names the command line does.
3. **Reviewed as PNGs.** Every capture here is tcell `SimulationScreen`
   cells written to `.txt` and to inline-coloured `.html`. `go test` renders
   every age at seven sizes.
4. **Static in a ticking game.** The sun and moon track game time (a day is
   an hour of play). Sap pulses rise up the trunk faster with morale, leaves
   catch the light, pollen flows along trade arcs, stars twinkle, the comet's
   tail flickers, and lightning flashes. Growth itself happens on every
   build.
5. **Scope ballooned (~20.7k lines of per-age art).** One grammar with seven
   epoch dials (bark hue, branch pen, leaf style, stars). The whole prototype
   is about 2.8k lines including the generator, captures and tests. The
   renderer (canvas, genome, dials, render, HUD) is about 1.7k.

## What the player sees and does

- **Glance:** crown size shows how big the civ is. Wedge widths show what it
  is strong in, read off the labels (`& food 29`) and the proportional
  spectrum bar underneath. Sparse leaves mean understaffed, wilting amber
  leaves mean starving, and grey means morale has collapsed. The header shows
  pop, idle workers (amber when above zero, as a bottleneck), food rate,
  morale, built and techs.
- **What changed since your last visit:** growth since the previous check-in
  is drawn in a bright, bold tint. The layout is prefix-stable (see below), so
  the new segments are exactly the ones at the end of each limb's growth
  order. The ticker says it in words:
  `» since your last visit: +7 food +2 timber … · +84 pop`.
- **What just happened:** the ticker leads with a pending catastrophe
  (`☄ catastrophe pending: type catastrophe to endure or succumb`), then
  active events, then the newest log headline.
- **Inspect (the verb):** select a limb and it lights up while the rest fade
  to 20%. The card shows the limb's buildings by `build` key, its workers
  against capacity (amber under 80%), its tier and its output.
- Prototype keys: `←/→` or `tab` inspect, `esc` clear, `[`/`]` previous or
  next age's save, `v` since-last-visit marks, `t` next theme, `space` pause
  the wind, `q` quit.

## How it reacts to the game

| Game event | Organism |
|---|---|
| New building | the next segment of that lineage's limb appears, with a leaf cluster |
| New lineage | a new wedge opens in its fixed place in crown order |
| Workers assigned | leaf clusters fill in (vigor = workers / capacity) |
| Population | trunk girth, and sap pulse count |
| Storage | the root system grows |
| Wonder built / banking | golden fruit / ripening fruit |
| Age advance | the tree's stature steps up; the strata deepen; at an epoch change the bark, branch pen and leaf style change |
| Epoch events | fossils in that epoch's stratum, forever |
| Harbinger | comet with the harbinger's name; tail length shows its probability |
| Catastrophe pending | red sky, lightning, the crown sheds leaves |
| Food negative / low morale | leaves wilt toward amber / grey; leaves fall |
| Civ met | a conifer on the horizon, sized by strength |
| Relations | conifer colour: friend, wary, war (with `×` sparks on the ground between) |
| Trade route | a dotted arc from the crown's edge to the partner, with pollen moving both ways (exports in highlight, imports in accent); red and still when disrupted |
| Idle time | sun and moon move with game ticks; on return, "since your last visit" growth is lit |

**Stability.** A limb's genome, meaning every segment it could ever grow in
the order it grows them, depends only on the save's seed and the lineage
key. More buildings reveal more of it and never move what is already drawn
(`TestGrowthIsPrefixStable`). The crown's fit to the sky uses a count-free
bound, so one more farm never rescales the tree. The honest exception is the
wedges: they are proportional to share, so when the balance shifts a limb
*leans* a few degrees. Limbs never swap places; crown order is fixed.

**Two players in the same age differ.** The seed shapes every genome, the
lineage mix sets the wedges, and the civs met, trades, wonders and history
are theirs.

## How it looks at each epoch

The grammar never changes; seven dials do (`style.go`):

| Epoch | Bark / branches | Leaves | Sky |
|---|---|---|---|
| Stone | brown, bowed organic strokes `/ \ | ~`, grass tufts | foliage clusters | sun or moon, stars at night |
| Iron | darker wood, same pen | foliage | same |
| Steel | grey, stiffer pen, `─` for horizontals, `║` trunk | foliage | same |
| Electric | brass | foliage | same |
| Digital | cyan circuit traces `│ ─ ╱ ╲` routed at 45°, with `•` vias; `┃╂` trunk | the same glyphs, read as pads | same |
| Neon | magenta traces | pads | stars by day |
| Cosmic | starlight: dotted constellation lines, `✦`/`+` stars at every joint, a `║` beam of a trunk | leaves glint `✦` | deep sky, no sun |

The primitive camp starts as a sapling: the first sixty buildings grow the
tree to its age's stature, and a camp with no production lineage yet shows
seed leaves (`00_primitive_seedling`: four huts, and it already looks
intentional).

## Themes and small terminals

- **Colour comes from theme roles.** Sky and background, text, dim,
  highlight, positive, negative, warning, accent and border are roles, so a
  theme switch retints everything, and light themes work
  (`10_medieval_age_light`, `11_digital_age_light`, the parchment and cosmic
  captures). The only fixed hues are identities (lineages, epochs, bark),
  and every one goes through `theme.Legible` against the theme background:
  darker on light themes, untouched on dark ones. Soil is the epoch hue
  washed into the background.
- **Monochrome.** The `.txt` captures are the monochrome test. Leaf *shapes*
  name lineages, fossils are distinct glyphs, and labels name limbs.
- **16 colours** (`42_industrial_age_16color`): the tree, labels and HUD
  survive. The strata backgrounds collapse to black, so the history reads
  only through its labels and fossils there.
- **Small terminals:** it lays out for any size. 100×30 and 80×24 keep
  everything except some labels. Below 64×20 a **mini mode** (the sidebar
  view, `32_…_mini_40x15`) drops the footer, keeps the three biggest labels
  and trims the HUD to `pop · built▲ · idle`.

## Performance

Genomes are grown once per state and cached on the model. A full 200×60
frame costs **0.72 ms** and about 1k allocations
(`BenchmarkRender200x60`, Apple M5), so redrawing it at 8 fps is noise. The
render is a pure function of (state, previous state, view, size, time).

## Integration

- **Replace both `citymap` and `worldmap`** with one `organism` view: the
  crown is the city, the horizon is the world. Keep the `map` command as an
  alias.
- **Main-screen mini view:** the 40×15 mode as a sidebar panel. It is the
  glanceable form, one keystroke from the full view.
- **Full view:** `map` / `organism` opens it full screen. Inspect keys as
  above; `enter` on a selected limb could pre-fill `build <key>` for its
  cheapest next building.
- Wire it into the existing overlay contract: `Refresh` stores the snapshot
  (plus the snapshot from the last check-in, for the growth marks), the draw
  reads it, and it never calls `GetState` under the bus lock.
- Code lands as `ui/organism/` (model, genome, render, hud); `lab/` goes
  away.

## Effort to productionise

About **1.5–2 weeks** for one developer:

- 2–3 days: move into `ui/organism`, the tview primitive, the overlay
  contract, and a mini panel on the main screen;
- 2 days: "last check-in" snapshot persistence (the account already knows
  the last session end), and the ticker fed from the log;
- 2 days: interaction polish (enter-to-build, mouse click on a limb, a
  legend popup on `?`);
- 1–2 days: 256/16-colour pass (strata as `░▒` texture when backgrounds
  collapse), a wide-glyph audit (`♪ ω ◊ ✦` in common fonts);
- 2 days: tests (exact size, panic-safety, stability, theme contrast sweep),
  the wiki page, and deleting about 20k lines of the old maps.

## Risks, and what would make it feel wrong

- **It is not a map.** Players who want to *see their city* won't find
  streets or buildings here. The owner has to decide whether a literal place
  is part of the fantasy. (The honest read of the history: seven literal
  maps failed.)
- **The glyph legend must be learned.** Labels and the spectrum bar carry it
  at first, but `ω`, `◊` and `¤` are not self-explanatory. A `?` legend is
  needed.
- **Late game saturates.** By design, growth is sub-linear, and in the bot's
  run galactic and transcendent look alike. Later ages need their own
  landmark (the trunk becoming a beam is a start), or the epoch dials need a
  bigger swing.
- **Wedges lean** as shares shift. That's accurate, but it is movement, and
  someone watching closely will see it.
- **The horizon is a picket fence** with ten or more civs. Wars and trade
  are only shown as colour and arcs, with no geography. If diplomacy grows,
  the horizon may need its own inspect mode.
- **Untested against real war and weather:** the bot never goes to war,
  and events were rare in these runs. War sparks and event motes are
  implemented but have no real-state capture. Trade routes were opened on
  real saves with the player's route command (`-trade`) because the bot
  never opens them.
- **Fonts:** a few glyphs (`♪ ω ◊ ✦ ☼ ☾`) are narrow in every font I know,
  but CJK-ambiguous-width terminals could double them.
- **Colour noise:** a crown of fourteen hues can tip into a rainbow. The
  palette is tuned so neighbours in crown order contrast, but it needs a
  designer's pass.

## Running it

```bash
go run ./lab/organism -gen                     # smoke bot, seed 7: one save per age (+ _prev, _catastrophe, _trade)
go run ./lab/organism -age digital_age         # interactive
go run ./lab/organism -capture                 # writes captures/ and captures/index.html
go run ./lab/organism -dump bronze_age -w 100 -h 30 -theme daylight
go test ./lab/organism/                        # every age × 7 sizes, determinism, stability, interaction
```
