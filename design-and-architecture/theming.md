# Theming & Accessibility

Status: design / implementation-ready
Owner: UI
Related: `design-and-architecture/accounts.md` (account-wide unlock + settings state, written in parallel)

A player-selectable theme system for AgeForge's terminal UI. Ships with a default
"Forge" look plus required accessibility themes (colorblind-safe, high-contrast)
unlocked from day one, and flavor themes that unlock via milestones. The hard part
isn't the picker. It's that the UI's color is currently hardcoded across ~957
sites in two completely different code paths. This doc resolves how to drive all
of them from a single ~9-role palette, including a definitive feasibility verdict
on the name-remap trick.

**Don't under-budget Phase 1 from the headline.** The name-remap trick retints the
~915 *named* inline tags (`[gold]`, `[gray]`, …) with **zero edits**, and that part is
free. But the ~42 direct `tcell` chrome calls (§2 Path B) and the ~23 hex
tags (§3.4) are **not** free: they're real, hand-applied edits, and they are the actual
bulk of Phase-1 work. "Retint 957 sites from one palette" is true as an *outcome*; it is
not true that all 957 sites cost nothing. Budget Phase 1 for the ~65 edits, not for zero.

> **Light-theme overhaul (2026-09, card dzjsC9cY).** Sections §3.1, §3.7, §3.8, §4,
> §7, §8 and the audit appendix describe the current system. The short version:
> a theme now specifies its whole surface (17 roles, not 9); every theme paints its
> own canvas; widget chrome late-binds through `theme.Ref` + `theme.WrapScreen`, so
> no widget can keep a stale color after a live switch; map derivations read roles
> through a luminance-aware adapter; raw colors outside `theme/` fail a test. The
> historical Phase-1..4 plan below is kept for context.

---

## 1. Goals / Non-Goals

### Goals
- One source of truth: every color in the UI derives from a small set of **semantic
  role colors** (Accent, Dim, Label, Positive, Negative, Highlight, Text, Background,
  Selection/Border).
- Ship multiple themes, switchable live without restarting the game.
- **Accessibility is not optional.** Colorblind-safe (deuteranopia + protanopia) and
  high-contrast themes ship in the first release and are **never gated** behind
  milestones.
- A contrast guard that makes it structurally hard to ship an unreadable theme.
- Theme choice and unlock state persist **account-wide**, not per-save.
- Minimal churn on the existing 957 color sites. We do not want to hand-edit every
  `[gold]` tag, and the feasibility work below shows the ~915 *named* tags need zero
  edits. The ~42 direct `tcell` calls and ~23 hex tags still require real edits (that's
  Phase-1 work, not free); "minimal churn" means ~65 sites, not zero.

### Non-Goals
- Not redesigning the account/settings system. We *depend on*
  `design-and-architecture/accounts.md` for where unlock + active-theme state lives;
  we do not specify that layer here.
- Not a full per-widget styling engine or user-authored custom themes (a stretch,
  §9). Shipped themes are curated and code-defined.
- Not touching the map renderer's bespoke RGB terrain palette (`ui/map.go`). The map
  uses geography-driven colors, not semantic roles; it stays out of scope except for
  its border/title chrome, which it shares with everything else.
- Not changing save format semantics. Themes never live in `data/saves/*.json`.

---

## 2. Current State

There is no central theme. Color is hardcoded across two independent paths.

### Path A: inline tview color tags (text content)
`SetDynamicColors(true)` text carries inline tags like `[gold]TITLE[-]`,
`[green]+12.5[-]`, `[#8b949e]hint[-]`. Real counts from `grep` over `ui/*.go`
(42 files):

| Tag | Count | Conventional role |
|------|------:|-------------------|
| `[gray]` | 154 | dim / secondary |
| `[gold]` | 134 | accent / titles / borders-in-text |
| `[cyan]` | 127 | labels / values |
| `[green]` | 79 | positive / gains |
| `[white]` | 66 | primary text |
| `[red]` | 60 | negative / losses |
| `[yellow]` | 52 | highlight / numbers |
| `[lime]` `[aqua]` `[orange]` `[blue]` | 7 total | stray one-offs |
| `[#8b949e]` | 21 | dim hint (hex) |
| `[#cc88ff]` `[#ff66cc]` | 2 | one-off accents (hex) |

~915 inline tags total. The named colors are semantic **by convention only**:
nothing enforces it, but the convention is consistent enough to remap on.

### Path B: direct tcell color calls (widget chrome)
Borders, modal backgrounds, list selection, input fields, age palettes. ~168
`tcell.Color*` / `tcell.NewRGBColor` references plus 42 `Set*Color(...)` widget
calls:

| Call | Count |
|------|------:|
| `SetTitleColor` | 16 |
| `SetBackgroundColor` | 10 |
| `SetBorderColor` | 9 |
| `SetFieldBackgroundColor` | 3 |
| `SetSelectedBackgroundColor` | 2 |
| `SetFieldTextColor` | 2 |

These do **not** flow through color-name tags. They take `tcell.Color` values
directly and apply once at widget construction.

### Existing prior art: `ui/theme.go`
There is already a `ui/theme.go` with global `tcell.Color` vars (`ColorTitle`,
`ColorAccent`, `ColorSuccess`, …) and an `AgePalette` / `ApplyAgePalette(ageKey)`
system that mutates those globals as the player advances ages. This is **only Path B
prior art**. It never touches the inline `[gold]` tags, which is why advancing an
age recolors some chrome but not the body text. Our design subsumes this: the
age-palette globals become one consumer of the theme palette, and the
epoch-adaptive idea (§9) is the generalization of `ApplyAgePalette`.

### tview's own theme
tview has a global `tview.Styles` (`Theme` struct: `BorderColor`, `TitleColor`,
`PrimitiveBackgroundColor`, …) read inside `NewBox()` at **construction time**.
Setting it changes defaults for *newly created* widgets but does not retint existing
ones. Useful as a baseline but not a live-retint mechanism.

### Persistence today
None for preferences. `data/` holds only `saves/` and `logs/`. There is no settings
file. The only cross-save state precedent is `CheaterBadge` / `EliteBadge`, peeked
from the autosave via `game.PeekSaveBadges`. Account-wide state is new ground, owned
by `accounts.md`.

---

## 3. Architecture

### 3.1 The Theme model

A theme is ~9 semantic **role** colors plus metadata. Roles, not literal color
names: `Positive` is "the gains color," whatever hue the active theme picks.

```go
// package theme

type Role int

const (
    RoleBackground Role = iota // canvas / primitive background
    RoleText                   // primary readable text
    RoleDim                    // secondary / hints / disabled
    RoleLabel                  // field labels, values
    RoleAccent                 // titles, borders, brand highlights
    RoleHighlight              // numbers, attention, "look here"
    RolePositive               // gains, success, +deltas
    RoleNegative               // losses, errors, -deltas
    RoleSelection              // selected list-row background
    numRoles
)

type Theme struct {
    Key         string            // "forge", "deuteranopia", "high_contrast", ...
    Name        string            // "Forge", shown in picker
    Blurb       string            // one-line flavor for the picker detail pane
    Accessible  bool              // true => never gated, always unlocked
    Colors      [numRoles]tcell.Color
    // Signed sentinel for the ± distinction in accessible themes (see §4):
    GainGlyph   string            // e.g. "▲" / "+"
    LossGlyph   string            // e.g. "▼" / "-"
}
```

`tcell.Color` carries true RGB (`tcell.NewRGBColor(r,g,b)`), so themes are defined as
explicit RGB and are not at the mercy of the terminal's 16-color palette on
truecolor-capable terminals.

#### Current role model (light-theme overhaul)

Nine roles described text on a canvas. They could not describe a light theme,
because the UI also draws panels, borders, selected rows, keycaps and danger
panels, and each of those needs a paired "text that reads on it" color. The model
is now 17 roles, split into the original core and an extended set:

| Role | Tag name | Used for | Derived default (if a theme leaves it unset) |
|------|----------|----------|------|
| Background | `bg` | canvas | none (core) |
| Text | `text` | primary text | none (core) |
| Dim | `dim` | secondary text, hints | none (core) |
| Label | `label` | field labels, values | none (core) |
| Accent | `accent` | titles, brand | none (core) |
| Highlight | `highlight` | numbers, attention | none (core) |
| Positive | `positive` | good / gains | none (core) |
| Negative | `negative` | bad / losses | none (core) |
| Selection | `selection` | selected-row fill | none (core) |
| Surface | `surface` | modal / overlay / panel fill | Background |
| Border | `border` | box borders, rules | Accent |
| SelectionText | `seltext` | text on Selection | Text |
| Bright | `bright` | emphasis inside body copy | Text |
| Warning | `warning` | warn | Highlight |
| OnAccent | `onaccent` | text on an Accent fill (keycaps, primary buttons) | black or white, whichever reads |
| Chip | `chip` | keycap-label / chip fill | Background 15% toward Text |
| OnNegative | `onnegative` | text on a Negative fill (danger panels) | Text if ≥ 3:1 on Negative, else black/white |

The derivations reproduce exactly what the UI drew before the extended roles
existed, which is how the dark themes kept their look (pinned by
`TestDerivedRoles_PreserveDarkThemes`). Forge pins `Chip = #30363d`, the keycap gray
the footer used to hard-code.

Semantic good / warn / bad are Positive / Warning / Negative.

**Light/dark** is derived, not stored: `Theme.IsLight()` is true when Background's
relative luminance is ≥ 0.25, and `Variant()` returns "Light"/"Dark". Deriving it
means it cannot drift from the palette.

**Groups** (picker sections): `Theme.Group()` is Accessibility if `Accessible`,
Standard if `Standard` (always unlocked, e.g. Forge and Daylight), otherwise
Unlockable. `AlwaysAvailable()` is `Accessible || Standard`; the registry test
requires every other theme to declare exactly one unlock key.

Themes register through package-level var initialization
(`var _ = register(...)`), which Go completes before any `init()` runs. The old
`init()` registration relied on file-name order and in practice ran *after*
`palette.go`'s seeding `init()`.

### 3.2 Retinting Path A (inline tags): the name-remap strategy

**This is the load-bearing decision, so the verdict is spelled out, not asserted.**

#### Verdict: name-remap is VIABLE for named tags. Caveat: hex tags are not.

I read the dependency source (`rivo/tview@v0.42.0`, `gdamore/tcell/v2@v2.13.10` in
`$(go env GOMODCACHE)`; no vendor dir).

What actually happens when tview draws `[gold]TEXT[-]`:

1. `TextView.Draw` iterates visible lines and, **for every line on every Draw**,
   walks the raw text (tags still embedded) via `step()` → `parseTag()` in
   `tview/strings.go`. The line index caches only each line's *starting* state and
   byte offset, **not** the resolved per-character colors. Colors are re-resolved
   from the raw string on each frame.

2. For a **named** foreground tag, `parseTag` resolves the color with a direct map
   lookup:
   ```go
   tStyle = tStyle.Foreground(tcell.ColorNames[name])   // strings.go
   ```
   `tcell.ColorNames` is an **exported, mutable** `map[string]tcell.Color`.

3. The named entries already carry true RGB, e.g.
   `ColorGold = ColorIsRGB | ColorValid | 0xFFD700`. So `[gold]` already paints a
   real RGB color, not a fragile palette index.

Therefore: **overwrite `tcell.ColorNames["gold"] = <theme RGB>` and, on the next
draw cycle, every existing `[gold]` tag in the entire UI retints**, with no edits to the
957 sites required for the named path. Because resolution is per-Draw, a theme
switch followed by `app.Draw()` / `app.QueueUpdateDraw()` retints everything live.

**The caveat (do not over-claim "zero edits"):**

- **Hex tags bypass the map.** A `[#8b949e]` tag resolves via `tcell.GetColor("#8b949e")`,
  which parses the literal hex and never reads `ColorNames`. The 23 hex tags
  (`#8b949e`×21, plus two one-offs) are **frozen** under remap. We fix this by
  converting them to named role tokens (§3.4), a small, bounded edit.
- The remap is **global process state, and its blast radius is wider than "inline
  text tags."** `tcell.ColorNames` is not a tview-tag-only table: `tcell.GetColor("name")`
  reads the *same* map. So overwriting `ColorNames["gold"]` changes **every** named-color
  resolution process-wide: not just `[gold]` tags, but any `tcell.GetColor("gold")` call
  in our code *and* any `tview.Styles` field that was assigned via a named color. Treat
  this as a process-global side effect, not a text-tag trick. The theme module owns those
  map keys and must restore/overwrite them atomically on switch. Document it loudly so
  nobody reaches for `ColorNames["gold"]` expecting tcell's gold.
  - **Pre-ship implementation step (do this before the first remap lands):** audit every
    `tcell.GetColor("<word>")` call and every `tview.Styles` field set via a named color
    (vs a literal `tcell.Color*` / `NewRGBColor`). For each, **decide intentionally**
    whether it *should* track the theme (leave it reading the remapped name) or *must stay
    fixed* (switch it to an explicit RGB literal so the remap can't drag it). Colors that
    silently follow the theme by accident are a bug waiting to surface on the first
    light-bg theme. This audit is Phase-1 work, not a footnote.
- We must remap a **fixed, known set** of names and never leave a name pointing at a
  stale value. The set is exactly the named colors in use:
  `gold, gray, cyan, green, white, red, yellow` plus the strays we fold in.
- This is a **deliberate use of a library implementation detail.** Pin the tview /
  tcell versions (they already are in `go.mod`) and add a tiny guard test (§3.6) so
  a future bump can't silently break retinting.

**Mapping convention name → role.** A `[name]` tag is just the role color, looked up
through the role:

| tag name | role it maps to |
|----------|-----------------|
| `gold`   | Accent |
| `gray`   | Dim |
| `cyan`   | Label |
| `green`  | Positive |
| `red`    | Negative |
| `yellow` | Highlight |
| `white`  | Text |

On theme switch:
```go
tcell.ColorNames["gold"]   = active.Colors[RoleAccent]
tcell.ColorNames["gray"]   = active.Colors[RoleDim]
tcell.ColorNames["cyan"]   = active.Colors[RoleLabel]
tcell.ColorNames["green"]  = active.Colors[RolePositive]
tcell.ColorNames["red"]    = active.Colors[RoleNegative]
tcell.ColorNames["yellow"] = active.Colors[RoleHighlight]
tcell.ColorNames["white"]  = active.Colors[RoleText]
// ... then app.QueueUpdateDraw(func(){})
```

#### Fallback if a future tview rewrites this
If a later tview version caches resolved colors at parse time (so map mutation no
longer retints live), the fallback is the **semantic-helper cleanup** from §9
promoted to mandatory: replace inline `[gold]` literals with `theme.Tag(RoleAccent)`
helpers that emit the active hex at format time. That's the cleaner long-term form
anyway; remap is the cheap shortcut that lets us ship retinting now without touching
957 sites. The guard test (§3.6) tells us if/when we're forced onto the fallback.

### 3.3 Routing Path B (direct widget calls): palette routing

Inline-tag remap does nothing for `SetBorderColor`, `SetBackgroundColor`,
`SetTitleColor`, list selection, input fields. Those read a `tcell.Color` once at
construction. We route them through the theme palette instead of literals.

1. The theme module exposes `theme.Color(RoleAccent)` etc., returning the active
   theme's `tcell.Color`.
2. Replace literal `tcell.ColorGold` / `tcell.NewRGBColor(...)` chrome calls with
   `theme.Color(role)`. This is a bounded set: 42 `Set*Color` call sites plus the
   age-palette globals in `ui/theme.go`.
3. Because these apply at construction, a **live** theme switch needs widgets to
   re-pull. Two-tier approach:
   - **Chrome defaults** also push into `tview.Styles` (BorderColor, TitleColor,
     PrimitiveBackgroundColor, ContrastBackgroundColor) so *future* widgets are
     correct.
   - **Existing** widgets are re-styled by a `Restyle()` pass: a small registry of
     "restylable" widgets (borders, modal bgs, the dashboard frame, list selection)
     that the theme module walks and re-applies `Set*Color` on switch, then
     redraws. The registry is the handful of long-lived chrome widgets, not every
     text view (text views retint for free via §3.2).

**Registration discipline: a registry someone must *remember* to populate will rot.**
A bare "add your widget to the slice" registry guarantees that some future widget ships
unthemed because nobody touched the registry. Make registration the *default path*, not
an optional afterthought:

- Expose a `theme.Track(widget, roleMap)` helper (and/or thin constructor wrappers like
  `theme.NewFramedBox(...)`) that both applies the current palette *and* enrolls the
  widget in the restyle registry in one call. Widget creation and theme enrollment
  become a single operation, so a new chrome widget can't be created without being
  tracked.
- `roleMap` declares which `Set*Color` setter maps to which `Role` (e.g.
  `{BorderColor: RoleAccent, TitleColor: RoleAccent, BackgroundColor: RoleBackground}`),
  so `Restyle()` is fully data-driven: it re-applies exactly the roles each widget
  declared, with no per-widget switch statement to keep in sync.
- A lint/grep guard in CI (or a code-review checklist item) flags raw `Set*Color`
  chrome calls outside the `theme` package and the `Track` wrappers, so the
  "construct-without-tracking" path is caught rather than trusted.

The existing `ui/theme.go` globals (`ColorTitle`, `ColorAccent`, …) become thin
aliases over `theme.Color(role)` so existing call sites keep compiling while we
migrate. `ApplyAgePalette` is reframed as an *optional* epoch-adaptive theme (§9),
not a competing color authority.

### 3.4 Stray hex tags

The 23 hex tags (`#8b949e` dim hints, `#cc88ff`, `#ff66cc`) are frozen under remap.
Resolve by converting them to **named role tokens**:
- `[#8b949e]` → `[gray]` (it's already a dim hint; same role).
- `[#cc88ff]` / `[#ff66cc]` → `[gold]` or `[yellow]` per intent (accent vs
  highlight), a judgment call at each of the two sites.

This is ~23 targeted edits, all in the same direction (hex → role token), and it
makes those tags theme-aware for free. The two progress-bar hex constants
(`BarFillColor = "#9370DB"`, `BarEmptyColor = "#444444"`) become role-derived: fill
= Accent/Highlight, empty = Dim/Background. After this pass, **zero hex literals
remain in retintable text**. The only fixed colors left are intentional ones (map
terrain, splash art accents) which we explicitly accept.

### 3.5 Module shape

```
theme/
  theme.go      // Theme struct, Role consts, registry of built-in themes
  palette.go    // Color(role), Tag(role) helpers; current active theme
  remap.go      // name-remap apply + restore of tcell.ColorNames
  restyle.go    // restylable-widget registry + Restyle() pass
  contrast.go   // luminance + contrast-ratio guard (§8)
  themes_*.go   // forge.go, accessible.go (deutan/protan/high_contrast), flavor.go
```

`theme` is a leaf package (depends only on tcell). `ui` depends on `theme`. No
import cycle; no engine dependency (themes are pure presentation).

### 3.6 Guard test
A unit test asserts the remap assumption so a dependency bump can't break us
silently:
```go
// Render "[gold]X" to a SimulationScreen, remap ColorNames["gold"] to a sentinel
// RGB, redraw, and assert the cell's foreground is the sentinel. If tview ever
// caches at parse time, this fails loudly and we switch to the §9 fallback.
```
Plus a contrast test (§8) over every shipped theme.

### 3.7 Background painting and late-bound chrome (light-theme overhaul)

**Decision: every theme paints an explicit background.** A light theme has to look
light on a black terminal, and a dark theme dark on a white one, so nothing may
fall through to the terminal's default colors. There is no per-theme
"transparent background" option. Nothing needed one: Forge was already painting
`#0d1117` on every widget through `tview.Styles.PrimitiveBackgroundColor`. The
only places that showed the terminal default were `StyleDefault` draws (splash
starfield, spacer cells) and unpainted screen area, and those were accidents.
Terminal transparency would be a new feature, not something we are preserving.

**The real bug was staleness, not painting.** tview reads `tview.Styles` once, when
a widget is constructed. Before the overhaul, `Styles` held concrete RGB, so the
~130 widgets not enrolled in `theme.Track` (146 constructors, 17 Track closures)
kept whatever canvas and text color were active when they were built. A live
switch from Forge to a light theme therefore left dark boxes under dark ink. The
remap retinted the named tags, but nothing retinted the boxes behind them.

**Mechanism.**

- `theme.Ref(role)` returns a sentinel `tcell.Color`: a `ColorValid` palette index
  far outside 0-255 that never has `IsRGB` set, so it cannot collide with a real
  theme color. `applyRemap` fills **every** `tview.Styles` field with Refs,
  including the ones tview uses for inverse states (list selection, button
  activation, dropdowns).
- `theme.WrapScreen(s)` wraps the terminal screen. Every write path
  (`SetContent`, `SetCell`, `Put`, `PutStr*`, `Fill`, `SetStyle`, `Clear`) resolves
  styles against the active theme: Refs become their role color, `ColorDefault` fg
  becomes Text and `ColorDefault` bg becomes Background. `Clear` fills with the
  theme canvas. The active palette sits behind an `atomic.Pointer`, so per-cell
  resolution takes no lock.
- `ui.App.Run` installs the wrapper before `tview.Application.Run`. Tests draw
  through `theme.WrapScreen(tcell.NewSimulationScreen(...))`.

The result: a theme switch is just a redraw. Existing `theme.Track` closures still
work (they set concrete colors and re-run on switch), but new chrome should
prefer `theme.Ref(role)`, which needs no enrollment at all.

**Inline tags** are already late-bound: tview re-resolves `[name]` through
`tcell.ColorNames` on every Draw (§3.2). So `theme.Tag(role)` now returns a *named*
tag (`[accent]`), no longer a hex literal. A hex tag computed at format time
freezes the color of whatever theme was active then. That was a latent bug in
`BarFillColor` for any static text. `theme.HexTag` survives only for colors that
deliberately are not the active theme's, like the picker's swatches of other
themes. Helpers: `Tag`, `TagFgBg`, `TagFgBgAttr`, `Keycap`, `KeycapButton`,
`Selected`, `Paint`, `NameTag` (for data-supplied color names such as
`config/epochs.go`'s, keeping the hue but correcting it with `Legible`) and
`LegibleTag`. Every role has a tag name (the table in §3.1). The seven legacy
aliases (`gold`, `gray`, `cyan`, `green`, `red`, `yellow`, `white`) stay remapped,
so the ~930 existing tags keep working without edits.

**Why the legacy aliases were not mass-rewritten to role names.** They already route
through the theme via the remap and are late-bound, and the raw-color guard
(§3.8) forbids any name outside the owned vocabulary. Rewriting them would be about
930 edits across the largest files in `ui/` with zero visual change. That is a
merge-conflict magnet for other work in flight. New code should use role names.

**Maps.** The citymap/worldmap recipes (hundreds of blends in `ui/citymap`) assume a
dark canvas. Map code reads roles through `citymap.mapColor(role)`: on a dark
theme it returns the theme's own role, byte-for-byte. On a light theme it returns
a dark-polarity *proxy* built from the theme's hues: the ink becomes the canvas,
the page becomes the light pole, Dim is lifted to mid-gray, and semantic roles are
raised to marker brightness (HSL L ≥ 0.62, then until ≥ 4.5:1 on the proxy
canvas). The recipes therefore compose a correct map. The citymap then re-keys it
for the light page with `liftForLight`, a monotonic concave lightness curve
(`l' = 0.15 + 0.85·l^0.7`), which preserves every lighter-than/darker-than
relationship (streets over ground, lit roof over shaded wall). On light themes the
drop-shadow tone is anchored near black so the 28% shadow blend always darkens.
Labels on light themes use `theme.Legible(…, banner, 4.5)`, and "bright" means
more ink. Space ages keep the dark proxy unlifted. World-map mediums (charcoal,
clay, parchment, blueprint, satellite, neon) paint their own canvases, and their
civ markers use the proxy so they stay marker-bright on those canvases.
`TestMapPolarity_AllThemes` pins the contract.

### 3.8 Audit rules for future UI code: "no raw colors outside `theme/`"

Enforced by `ui/theme_guard_test.go` (an AST walk over `ui/`, `game/`, `config/`):

1. **No inline tag with a raw color.** Tag fg/bg must be a role name or legacy alias.
   `[#rrggbb]`, `[aqua]`, `[black:gold]`, `[white:#30363d]` all fail. Use
   `theme.Tag(role)` / `theme.TagFgBg` / `theme.Keycap*`.
2. **No `tcell.Color<Name>` constants.** Use `theme.Color(role)` for a concrete color
   or `theme.Ref(role)` for widget chrome.
3. **No `tcell.NewRGBColor` / `NewHexColor` / `GetColor`** outside a short allow-list
   of pixel-streaming and color-math files (`ui/citymap/citymap.go`,
   `ui/citymap/overlay.go`, `ui/splash_canvas.go` for its dark-theme art).
4. **Fills come with their "on" role.** Text drawn on Accent uses OnAccent, on
   Negative uses OnNegative, on Selection uses SelectionText, on Chip uses Text. Use
   `styleFilledButton` / `styleDangerModal` / `dangerTag` (`ui/theme_widgets.go`).
   Canvas roles such as `[red]` or `[gray]` inside a Negative panel are a bug. The
   account-wipe panel shipped `[red]` on its red fill.
5. **Panels paint Surface, and so do their children.** Modals and overlays use
   `RoleSurface`, and every child primitive sets it too, so the panel reads as one
   piece on themes where Surface ≠ Background.
6. **Identity hues** (epochs, lineages, factions) keep their hue through
   `theme.Legible` / `NameTag`, never a raw tag.
7. **Map colors** go through `mapColor`, never `rgba(theme.Color(...))` directly.

The render sweep (`ui/theme_render_test.go`) is the backstop for these rules. It
draws the dashboard, several overlays including both maps, the picker and a danger
modal under every theme, then fails on any glyph with fg == bg and on any cell
without a concrete RGB background.

---

## 4. Shipped Themes

All themes are code-defined RGB. Accessibility themes are `Accessible: true` and
**unlocked by default, never milestone-gated.**

### Forge (default): `forge`
The current dark + gold look, formalized. Dark near-black background, warm gold
accent, gray dim, cyan labels, green/red for ±, yellow highlights, off-white text.
This is what ships selected.

### Deuteranopia-safe: `deuteranopia` *(accessible, default-unlocked)*
### Protanopia-safe: `protanopia` *(accessible, default-unlocked)*
Critical constraint: in AgeForge **green = gains and red = losses everywhere**:
resource deltas, rates, combat, trade. Red-green deficiency makes that distinction
collapse. So the accessible palettes **must not encode ± with red vs green.**

- **Positive → blue** (e.g. `#3B9EFF`), **Negative → orange** (e.g. `#FF8C42`).
  Blue/orange is the canonical deutan/protan-safe opposition and stays distinct under
  both simulations.
- **Belt-and-suspenders: signed glyphs.** Accessible themes set `GainGlyph`/`LossGlyph`
  (`▲`/`▼`, or `+`/`-`) so the sign is encoded by **shape as well as hue**. Delta
  formatting helpers consult the active theme's glyphs; non-accessible themes can
  leave them empty. This is the redundant-encoding principle: never rely on color
  alone for meaning.
- Accent/Highlight/Label chosen to stay mutually distinguishable under deutan/protan
  simulation (favor blue/yellow/white spread; avoid accent≈positive collisions).
- Deutan and protan ship as separate themes because their safe hues differ slightly;
  one "colorblind" catch-all under-serves both.

### High Contrast: `high_contrast` *(accessible, default-unlocked)*
Maximum legibility: pure/near-pure background, white text, saturated unambiguous role
colors, every role pair comfortably above the WCAG AA contrast floor (§8 enforces
this). For low-vision players and high-glare terminals.

### Daylight: `daylight` *(Standard, default-unlocked, light)*
The clean light default, added by the light-theme overhaul. Off-white canvas
`#f6f7f9`, true-white Surface panels, charcoal ink `#1f2328`, slate Dim, deep-teal
labels, and deep-amber accent `#9a6700` that echoes Forge's gold. Positive and
Negative are forest green and brick red. Every foreground role is a *dark* variant
of its Forge counterpart, because on a light page the roles must be darker than the
page. Listed under Standard next to Forge.

### High Contrast Light: `high_contrast_light` *(accessible, default-unlocked, light)*
White page, black ink, black borders. Every text role clears AAA (7:1) where the
matrix demands it. It keeps the colorblind-safe blue gain / dark-orange loss and the
▲/▼ glyphs. The new role model made it nearly free, so it ships.

### Group summary

| Group | Themes |
|-------|--------|
| Standard | Forge (dark, default), Daylight (light) |
| Accessibility | Deuteranopia-safe, Protanopia-safe, High Contrast (dark); High Contrast Light (light) |
| Unlockable | Parchment (light), Bronze, Cyberpunk, Monochrome, Cosmic (dark) |

### Flavor themes (milestone-gated, §5)
Curated, code-defined, **cosmetic only**. They never alter the ± encoding semantics
in a way that breaks accessibility expectations (and still pass the contrast guard).
Candidates:
- **Parchment**: light sepia background, ink-brown text, wax-red/forest-green ±.
  (The first light-background theme. Before the overhaul it only half-worked:
  widgets built under another theme kept their dark canvas after a live switch.)
- **Bronze Age**: burnished metallics.
- **Cyberpunk**: magenta/cyan neon on black (riffs on the existing `cyberpunk_age`
  age palette).
- **Monochrome Terminal**: amber-on-black or green-on-black retro CRT.
- **Cosmic**: deep indigo with starlight accents (riffs on `galactic_age`).

Exact RGB values are filled in during Phase 2 against the contrast guard; the guard
is the acceptance gate, not a designer's eyeball.

---

## 5. Milestone-Gated Unlocks

Flavor themes unlock through the existing milestone system; accessibility + Forge are
always available.

- Each gated theme declares an unlock condition: a milestone key or chain key (e.g.
  Cyberpunk unlocks on reaching `cyberpunk_age`; Parchment on a renaissance
  milestone). The mapping lives in the theme registry, not scattered in milestone
  code.
- **Unlock state is account-wide.** When a milestone/chain completes, the game records
  the unlock in the **account/settings layer** (`accounts.md`). It is *not* stored in
  `data/saves/*.json`. Earn Cyberpunk on one empire and it's yours on every save and
  every future new game. This mirrors how badges feel permanent, but lives in the
  proper account store rather than being peeked from a save.
- Hook point: `MilestoneManager.CheckMilestones` / `CheckChains` already return
  newly-completed milestones/chains. On a new completion, the UI layer asks the theme
  registry "does this unlock a theme?" and, if so, calls `account.UnlockTheme(key)`. That
  call returns `(newly bool, err)`: **we fire the unlock notification (§7) only when
  `newly == true`**, so replaying a milestone (or re-running a chain check) never re-toasts
  an already-owned theme. We add a thin theme-unlock resolver; we do **not** bake theme
  keys into engine code.
- This doc treats the account layer as a dependency, and uses **`accounts.md`'s exact
  names** so the two docs can't drift: `UnlockTheme(key) (newly bool, err error)`,
  `HasTheme(key) bool`, `UnlockedThemes() []string`, `ActiveTheme() string`,
  `SetActiveTheme(key) error`. (Earlier drafts of this doc said `IsThemeUnlocked`; the
  account layer names it `HasTheme`; we use `HasTheme` here too.) `accounts.md` §8 is the
  authoritative signature list.

---

## 6. Persistence

- **Active theme** and **unlocked-theme set** live in the **account/settings layer**:
  concretely `data/account.json`, HMAC-signed for consistency with saves. (`accounts.md`
  §3 makes this a firm decision: one signed `account.json`, not a to-be-decided choice.)
  **Never** in per-save JSON.
- Rationale: theme is a player preference and a player-account achievement, not
  empire state. Loading an old save must not change your theme; starting a new game
  must not relock your earned themes.
- The save format (`game/save.go` `GameSave`) is **untouched**. No new fields.
- On startup the UI reads the account layer, resolves the active theme (default
  `forge` if unset or if the stored key is unknown), and applies it (remap + restyle)
  **before** the first Draw, so the splash already wears the chosen theme.
- Defensive: if the stored active theme is somehow locked or missing, fall back to
  Forge and don't crash.

---

## 7. UX

### Main-menu Theme picker
A "Themes" entry on the splash `mainList` (alongside Load / New Game / Quit). Opens a
picker page modeled directly on the **load-game browser** (`ui/load_game.go`): a list
on the left, a detail/preview pane on the right.

### `theme` command
Available from the in-game `>` prompt via `HandleCommand` (`ui/input.go`):
- `theme`: opens the picker (returns `CommandResult{OverlayName: "theme"}`).
- `theme list`: prints unlocked vs locked themes (locked ones show their unlock
  condition, e.g. "Cyberpunk: Reach the Cyberpunk Age").
- `theme <name>`: switches directly to a theme by key/name if unlocked; errors with
  the unlock hint if locked, errors "unknown theme" otherwise.

Add a `case "theme":` to the dispatch switch returning the above.

### Live preview (apply-on-highlight, revert-on-cancel)
Exactly the load-game detail-pane pattern: the picker list's `SetChangedFunc` fires
on highlight. On highlight we **apply the theme for real** (remap + restyle + redraw)
so the player sees the whole UI in that theme immediately. The picker is itself a
live sample of the running UI. We remember the theme that was active on open.
- **Confirm** (Enter / select): persist the highlighted theme as active via the
  account layer; close.
- **Cancel** (Esc): re-apply the remembered original theme; close. No persistence.

Because retinting is a global remap, "preview" and "apply" are the same operation;
the only difference is whether we persist and whether cancel reverts. Clean.

### Picker layout (light-theme overhaul)

- **Grouped list.** Section headings Standard / Accessibility / Unlockable
  (`theme.Groups`). Headings are list items that the changed handler skips in the
  direction of travel, wrapping like `tview.List`. The skip uses only
  `List.SetCurrentItem`, which just re-enters the handler: no lock is taken, no
  `QueueUpdateDraw` (the TCGiSWYX deadlock guard, `TestThemePicker_NoDeadlockOnNavigate`).
- **Row tags.** `(current)`, `light`/`dark`, `accessible`, `🔒 locked`.
- **Details pane** (Surface): name, "Light theme · Standard", blurb, accessible
  note with glyphs, lock line with the unlock hint, and swatches for Background,
  Surface, Text, Accent, Positive, Negative, Highlight, Dim and Selection. Each
  swatch sits on a contrast rim (`▐███▌`), so a Background swatch doesn't vanish
  into the pane.
- **Sample panel.** A Box whose draw function paints a miniature UI in the
  candidate's *own literal colors*: canvas, Surface panel with Border and Accent
  title, label/number/± rows, body plus dim text, a selected row, a keycap with its
  chip label, and a danger chip. It is independent of the active theme, so the
  preview is correct even without live apply. Live apply still runs, so the whole
  UI behind the picker shows the candidate too.
- Locked themes still preview (list, details, sample, live apply); only Enter is
  gated.
- `theme list` shows each theme's light/dark variant.

### Palette swatches
The picker detail pane shows the theme's blurb plus a swatch row: one colored block
per role rendered with that role's color tag, labeled (Accent / Positive / Negative /
…). For accessible themes, show the gain/loss **glyphs** next to the ± swatches so the
redundant encoding is visible in the picker itself. Locked themes show swatches dimmed
with a lock marker and the unlock condition.

### Unlock notification
When a milestone grants a theme (§5), surface it through the existing toast/log
channel (the same path milestone completions already use): a toast that names the
theme and points to the `theme` command. Non-modal; don't interrupt play.
**Gate the toast on `account.UnlockTheme` returning `newly == true`**. `UnlockTheme` is
idempotent (re-unlocking an owned theme is a no-op that returns `newly == false`), so a
milestone re-fire or a redundant chain check must not produce a duplicate "unlocked"
toast. Only a new unlock notifies.

---

## 8. Contrast-Safety Guard

No unreadable theme ships. The Parchment (light-bg) theme and the contrast bug we
already hit are the motivation: a modal input once used a light field background with
white text and was effectively invisible until we set an explicit dark field bg
(`ui/newgame_modal.go`). A theme that does that to the whole UI is unacceptable.

### Check: WCAG luminance contrast ratio
For each theme, compute the contrast ratio between every **foreground role** (Text,
Dim, Label, Accent, Highlight, Positive, Negative) and the theme's **Background**
(and, for Selection, against text drawn on the selection bg):

```
L = relative luminance per WCAG (sRGB linearization, 0.2126 R + 0.7152 G + 0.0722 B)
ratio = (Llighter + 0.05) / (Ldarker + 0.05)   // 1.0 .. 21.0
```

Thresholds:
- Body roles (Text, Label, Positive, Negative, Highlight) vs Background: **>= 4.5**
  (WCAG AA normal text).
- Dim vs Background: **>= 3.0** (it's intentionally secondary, AA large-text floor).
- Accent vs Background: **>= 3.0** (often borders/titles, large glyphs).
- Text on Selection background: **>= 4.5**.

**Full role matrix (light-theme overhaul, `theme/contrast_roles_test.go`).** Every pair
the UI actually draws is checked, on both Background *and* Surface:

| fg on bg | Standard | Accessible |
|----------|---------:|-----------:|
| Text, Bright on Background/Surface | 4.5 | 7.0 |
| Label, Highlight, Positive, Negative, Warning on Background/Surface | 4.5 | 4.5 |
| Dim, Accent on Background/Surface | 3.0 | 4.5 |
| Border on Background/Surface (non-text UI) | 3.0 | 3.0 |
| SelectionText on Selection | 4.5 | 7.0 |
| OnAccent on Accent (keycaps) | 4.5 | 4.5 |
| Text on Chip | 4.5 | 7.0 |
| OnNegative on Negative (bold danger text) | 3.0 | 4.5 |

`TestRoles_AllSet` asserts that no role is unset or non-RGB, so an unset role can't
fall through to the terminal default. The render sweep (§3.8) is the end-to-end check.

### Check: colorblind distinguishability (simulation, not luminance)

Luminance contrast and colorblind distinguishability are **different properties**. Two
hues can clear 4.5:1 against the background yet be nearly identical to a deuteranope.
The accessible palettes (§4) are *designed* to keep Accent/Positive/Negative/Highlight
mutually distinct under deutan/protan deficiency, but "designed to" must be backed by a
test, not by a sighted developer's eyeball.

So the guard ALSO runs the accessible palettes through **deuteranopia and protanopia
simulation** (a standard CVD model such as Brettel/Viénot or Machado, applied to each
role's RGB), then asserts that the post-simulation role colors stay separated by a
minimum perceptual distance (e.g. a ΔE floor in a perceptually-uniform space). This is
what actually catches an **Accent ≈ Positive collision** under simulated deficiency, a
class of bug luminance contrast is blind to. Run it at least on the accessible themes;
running it on all themes is cheap and worthwhile.

### Enforcement
A `theme_contrast_test.go` iterates every shipped theme and fails the build if any
pair is under its floor. Themes are tuned against the test, not by eye. This is also
where the §3.6 remap guard lives. Net effect: **on truecolor terminals, an unreadable
theme cannot reach players** because it can't pass CI.

### The limitation, stated honestly: 256-color terminals can still erode contrast

The guard computes WCAG luminance on the theme's **declared truecolor RGB**. But on a
256-color (or 16-color) terminal, tcell **down-samples** each declared RGB to the
nearest palette slot, and a pair that clears 4.5:1 in truecolor can collapse below the
floor once both ends snap to neighboring palette entries. This is worst for
**light-background themes like Parchment**, where the foreground/background luminances
are already close and quantization has more room to flip the ratio. So the strong claim
("cannot reach players") only holds on truecolor terminals; on 256-color terminals a
"passing" theme can still be marginal.

There are two options, and we should pick one:

- **Scope the guarantee:** state plainly that the contrast guarantee applies to
  truecolor terminals, and document that 256-color rendering is best-effort. Simplest;
  acceptable if truecolor is effectively required.
- **Add a 256-color-quantized check:** for light-background themes (and ideally all
  themes), quantize each role color through tcell's own 256-color down-sample, then
  re-run the contrast ratio on the *quantized* values against the same floors. This
  catches the Parchment-style collapse in CI instead of in a player's terminal.

Recommended: ship the truecolor check for everything *and* the quantized check for
light-bg themes, since those are where quantization actually bites.

---

## 9. Stretch / Future

- **Epoch-adaptive auto-theme.** Generalize the existing `ApplyAgePalette` into a
  proper theme: an "Adaptive" pseudo-theme that retints role colors as the player
  advances epochs (Bronze warmth → Industrial grime → Cosmic indigo). Now that all
  color flows through one palette, this is "swap the active palette on age-change"
  rather than the current half-measure that only touches chrome. Must still pass the
  contrast guard at every age step.
- **Semantic-token cleanup (also the §3.2 fallback).** Replace inline `[gold]`
  literals with `theme.Tag(RoleAccent)` helpers that emit the active color at format
  time. This removes the reliance on the tcell-map-mutation implementation detail
  entirely and is the cleaner long-term form. Large but mechanical; do it lineage-area
  by lineage-area. Mandatory only if a tview bump breaks remap.
- **User-authored themes.** Read extra themes from `data/themes/*.json`, run them
  through the same contrast guard at load, reject ones that fail. Lets the community
  share palettes without code changes.

---

## 10. Phased Implementation Plan

Ordered so something visible lands early. Each phase is independently shippable.

### Phase 1: Theme spine + Forge + accessibility themes (visible win)
- Add the `theme` package: `Theme`/`Role` model, palette accessors, the name-remap
  apply/restore (`remap.go`), the restylable-widget registry (`restyle.go`).
- Define **Forge**, **Deuteranopia**, **Protanopia**, **High Contrast** (accessible,
  default-unlocked).
- Convert the 23 hex tags + the two bar-color constants to role tokens (§3.4).
- Add the `theme` command + main-menu Themes entry + picker with live preview and
  swatches.
- Add the contrast guard test (§8) and the remap guard test (§3.6).
- **Theme choice persists in-memory / process-local for now** if `accounts.md` isn't
  landed yet. Wire to a stub `ActiveTheme/SetActiveTheme` so the picker works end to
  end. Visible result: a player can switch between four legible themes live.
- **Cross-doc dependency: Phase 1 has NO dependency on `accounts.md`.** It ships against a
  stub account API (`HasTheme`/`UnlockTheme`/`ActiveTheme`/`SetActiveTheme`), so theming
  Phase 1 can land *before* the account system exists. This is deliberate: it keeps the
  two tracks from being scheduled into a deadlock.

### Phase 2: Account-wide persistence
- **Cross-doc dependency: this phase REQUIRES `accounts.md` Phase 3 (the unlock API).**
  Theming Phase 2 replaces the Phase-1 stub with the real `account.HasTheme` /
  `UnlockTheme` / `ActiveTheme` / `SetActiveTheme`, so it cannot start until accounts
  Phase 3 (`HasTheme`/`UnlockTheme`/`UnlockedThemes`/`ActiveTheme`/`SetActiveTheme`) has
  landed. Named here and in `accounts.md` §9 so neither plan can be scheduled to block the
  other.
- Integrate with the account/settings layer (`accounts.md`): persist active theme +
  unlocked set; load and apply before first Draw.
- Default-unlock all accessible themes + Forge; everything flavor starts locked.
- Migrate the `ui/theme.go` age-palette globals to thin aliases over `theme.Color`.

### Phase 3: Flavor themes + milestone unlocks
- Define flavor themes (Parchment, Bronze, Cyberpunk, Monochrome, Cosmic), each
  tuned to pass the contrast guard.
- Add the theme-unlock resolver hooked to `CheckMilestones`/`CheckChains`; persist
  unlocks account-wide; fire the unlock notification.
- Picker shows locked themes with unlock conditions; `theme list` reflects state.

### Phase 4: Stretch
- Epoch-adaptive Adaptive theme (generalize `ApplyAgePalette`).
- Begin the semantic-token cleanup (`theme.Tag`), and keep it on the shelf as the
  hard fallback if a dependency bump ever breaks the remap.
- Optionally: user-authored `data/themes/*.json` with contrast-gated loading.

---

## Appendix: wiki sync

Per project rules, the shipping change updates `site/`:
- `site/docs/commands.md`: document the `theme` command (`theme`, `theme list`,
  `theme <name>`).
- A new/updated accessibility section in the wiki noting colorblind + high-contrast
  themes ship unlocked.
- Any wiki page that asserts "green = gain / red = loss" should note accessible
  themes use blue/orange + glyphs instead.

---

## Appendix: color audit catalog (light-theme overhaul, base `06f6d24`)

Everything that bypassed the theme, or that would break on a light background,
and what happened to it.

| Category | Count at base | Resolution |
|----------|--------------:|------------|
| Widgets built with construction-time `tview.Styles` colors, not in `theme.Track` | ~129 of 146 constructors (17 Track closures) | `tview.Styles` now holds `theme.Ref` sentinels, resolved per cell by `theme.WrapScreen` (§3.7) |
| tview defaults never set (inverse text, graphics, tertiary, more-contrast bg) | 5 `Styles` fields | all 11 fields set to Refs |
| `tcell.StyleDefault` draws (terminal-default bg) | 8 (splash canvas 5, citymap 3) | wrapper resolves ColorDefault to Background/Text; `Clear` paints the canvas |
| `tcell.Color<Name>` constants | 110, all in the inert `AgePalettes` (`ui/theme.go`) | moved to `theme/agepalettes.go` |
| `tcell.NewRGBColor` outside theme | 10 (citymap 5, splash 4, picker 1) | splash tagline/badge → roles; starfield/title → light-aware; citymap pixel streaming + math allow-listed |
| Keycap tags `[black:gold:b]` + `[white:#30363d:b]`, sidebar `[black:gold]` | 2 sites | `theme.KeycapButton`, `theme.Selected` (OnAccent/Accent, Text/Chip) |
| Stray unowned tag names `[aqua]` ×6, `[lime]` ×3, `[orange]`, `[blue]` | 11 | `[label]`, `[positive]`, `[warning]`, `[label]` |
| Hex tags | 1 (`[#808080]` picker fallback) | `theme.HexTag` (falls back to Dim) |
| Data-supplied color names (`config/epochs.go`: `lightblue`, `blue`, `magenta`) | 3 names, 4 render sites | `theme.NameTag` (Legible-corrected hue) |
| Legacy alias tags `[gold]` 167, `[gray]` 265, `[cyan]` 175, `[green]` 85, `[red]` 82, `[yellow]` 71, `[white]` 87 | 932 | unchanged; already late-bound via remap, now part of the guarded vocabulary |
| `SetBackgroundColor(Negative)` danger panels with canvas-colored text | 11 sites; the account-wipe panel drew `[red]` on the red fill | `styleDangerModal`, OnNegative tags, `styleFilledButton` |
| Buttons with a label color that doesn't pair with the fill (Accent+Background, Negative+Text/Highlight, Positive+Text) | 4 | `styleFilledButton` → OnAccent / OnNegative / BestOn |
| Modal/overlay panels on Background | 13 sites | Surface (children too) |
| `theme.Tag` emitting frozen hex (used by progress bars) | 2 helpers | named late-bound tags |
| Map derivations `rgba(theme.Color(...))` | 29 (palette 12, topdown 7, worldmap 9, worldmedium 1) | `mapColor` light proxy + `liftForLight` + light shadow anchor + legible labels (§3.7) |
| Progress bars, sparklines, morale graph, age ✓/✗ strip, onboarding panel | n/a | already role tags; covered by the render sweep |
| Wonder icon half-block pixel art (`ui/wonder_icon.go`) | 1 | intentional art: self-contained sprite tiles with their own backgrounds |

Visible changes on dark themes, all deliberate: the four stray names take their
role's hue (for example aqua `#00ffff` becomes Forge's Label `#39c5cf`); the splash
tagline and prestige badge use Dim/Label; the Succumb button label is OnNegative,
no longer yellow-on-red; danger text on the orange Negative of the colorblind
themes is black, where it used to be white at about 2:1; the keycap-label Chip on
non-Forge dark themes is tinted from the theme (Forge keeps `#30363d`). Map
renders on dark themes are byte-identical (`TestMapColor_DarkThemesUnchanged`).
