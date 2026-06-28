# Theming & Accessibility

Status: design / implementation-ready
Owner: UI
Related: `design-and-architecture/accounts.md` (account-wide unlock + settings state — being written in parallel)

A player-selectable theme system for AgeForge's terminal UI. Ships with a default
"Forge" look plus required accessibility themes (colorblind-safe, high-contrast)
unlocked from day one, and flavor themes that unlock via milestones. The hard part
isn't the picker — it's that the UI's color is currently hardcoded across ~957
sites in two completely different code paths. This doc resolves how to drive all
of them from a single ~9-role palette, including a definitive feasibility verdict
on the name-remap trick.

**Don't under-budget Phase 1 from the headline.** The name-remap trick retints the
~915 *named* inline tags (`[gold]`, `[gray]`, …) with **zero edits** — that part is
genuinely free. But the ~42 direct `tcell` chrome calls (§2 Path B) and the ~23 hex
tags (§3.4) are **not** free: they're real, hand-applied edits, and they are the actual
bulk of Phase-1 work. "Retint 957 sites from one palette" is true as an *outcome*; it is
not true that all 957 sites cost nothing. Budget Phase 1 for the ~65 edits, not for zero.

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
  `[gold]` tag — and the feasibility work below shows the ~915 *named* tags need zero
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

### Path A — inline tview color tags (text content)
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

~915 inline tags total. The named colors are semantic **by convention only** —
nothing enforces it, but the convention is consistent enough to remap on.

### Path B — direct tcell color calls (widget chrome)
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
prior art** — it never touches the inline `[gold]` tags, which is why advancing an
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
names — `Positive` is "the gains color," whatever hue the active theme picks.

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
    Name        string            // "Forge" — shown in picker
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

### 3.2 Retinting Path A (inline tags) — the name-remap strategy

**This is the load-bearing decision, so the verdict is spelled out, not asserted.**

#### Verdict: name-remap is VIABLE for named tags. Caveat: hex tags are not.

I read the dependency source (`rivo/tview@v0.42.0`, `gdamore/tcell/v2@v2.13.10` in
`$(go env GOMODCACHE)`; no vendor dir).

What actually happens when tview draws `[gold]TEXT[-]`:

1. `TextView.Draw` iterates visible lines and, **for every line on every Draw**,
   walks the raw text (tags still embedded) via `step()` → `parseTag()` in
   `tview/strings.go`. The line index caches only each line's *starting* state and
   byte offset — **not** the resolved per-character colors. Colors are re-resolved
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
draw cycle, every existing `[gold]` tag in the entire UI retints** — no edits to the
957 sites required for the named path. Because resolution is per-Draw, a theme
switch followed by `app.Draw()` / `app.QueueUpdateDraw()` retints everything live.

**The caveat (do not over-claim "zero edits"):**

- **Hex tags bypass the map.** A `[#8b949e]` tag resolves via `tcell.GetColor("#8b949e")`,
  which parses the literal hex and never reads `ColorNames`. The 23 hex tags
  (`#8b949e`×21, plus two one-offs) are **frozen** under remap. We fix this by
  converting them to named role tokens (§3.4) — a small, bounded edit.
- The remap is **global process state, and its blast radius is wider than "inline
  text tags."** `tcell.ColorNames` is not a tview-tag-only table — `tcell.GetColor("name")`
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
helpers that emit the active hex at format time. That's the "honest" long-term form
anyway; remap is the cheap shortcut that lets us ship retinting now without touching
957 sites. The guard test (§3.6) tells us if/when we're forced onto the fallback.

### 3.3 Routing Path B (direct widget calls) — palette routing

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

**Registration discipline — a registry someone must *remember* to populate will rot.**
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
  so `Restyle()` is fully data-driven — it re-applies exactly the roles each widget
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
  highlight) — judgment call at each of the two sites.

This is ~23 targeted edits, all in the same direction (hex → role token), and it
makes those tags theme-aware for free. The two progress-bar hex constants
(`BarFillColor = "#9370DB"`, `BarEmptyColor = "#444444"`) become role-derived: fill
= Accent/Highlight, empty = Dim/Background. After this pass, **zero hex literals
remain in retintable text** — the only fixed colors left are intentional ones (map
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

---

## 4. Shipped Themes

All themes are code-defined RGB. Accessibility themes are `Accessible: true` and
**unlocked by default, never milestone-gated.**

### Forge (default) — `forge`
The current dark + gold look, formalized. Dark near-black background, warm gold
accent, gray dim, cyan labels, green/red for ±, yellow highlights, off-white text.
This is what ships selected.

### Deuteranopia-safe — `deuteranopia` *(accessible, default-unlocked)*
### Protanopia-safe — `protanopia` *(accessible, default-unlocked)*
Critical constraint: in AgeForge **green = gains and red = losses everywhere** —
resource deltas, rates, combat, trade. Red-green deficiency makes that distinction
collapse. So the accessible palettes **must not encode ± with red vs green.**

- **Positive → blue** (e.g. `#3B9EFF`), **Negative → orange** (e.g. `#FF8C42`).
  Blue/orange is the canonical deutan/protan-safe opposition and stays distinct under
  both simulations.
- **Belt-and-suspenders: signed glyphs.** Accessible themes set `GainGlyph`/`LossGlyph`
  (`▲`/`▼`, or `+`/`-`) so the sign is encoded by **shape as well as hue**. Delta
  formatting helpers consult the active theme's glyphs; non-accessible themes can
  leave them empty. This is the redundant-encoding principle — never rely on color
  alone for meaning.
- Accent/Highlight/Label chosen to stay mutually distinguishable under deutan/protan
  simulation (favor blue/yellow/white spread; avoid accent≈positive collisions).
- Deutan and protan ship as separate themes because their safe hues differ slightly;
  one "colorblind" catch-all under-serves both.

### High Contrast — `high_contrast` *(accessible, default-unlocked)*
Maximum legibility: pure/near-pure background, white text, saturated unambiguous role
colors, every role pair comfortably above the WCAG AA contrast floor (§8 enforces
this). For low-vision players and high-glare terminals.

### Flavor themes (milestone-gated, §5)
Curated, code-defined, **cosmetic only** — they never alter the ± encoding semantics
in a way that breaks accessibility expectations (and still pass the contrast guard).
Candidates:
- **Parchment** — light sepia background, ink-brown text, wax-red/forest-green ±.
  (Our first light-background theme; a real contrast-guard exercise.)
- **Bronze Age** — burnished metallics.
- **Cyberpunk** — magenta/cyan neon on black (riffs on the existing `cyberpunk_age`
  age palette).
- **Monochrome Terminal** — amber-on-black or green-on-black retro CRT.
- **Cosmic** — deep indigo with starlight accents (riffs on `galactic_age`).

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
  `data/saves/*.json` — earn Cyberpunk on one empire and it's yours on every save and
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
  `SetActiveTheme(key) error`. (Earlier drafts of this doc said `IsThemeUnlocked` — the
  account layer names it `HasTheme`; we use `HasTheme` here too.) `accounts.md` §8 is the
  authoritative signature list.

---

## 6. Persistence

- **Active theme** and **unlocked-theme set** live in the **account/settings layer**
  — concretely `data/account.json`, HMAC-signed for consistency with saves. (`accounts.md`
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
- `theme` — opens the picker (returns `CommandResult{OverlayName: "theme"}`).
- `theme list` — prints unlocked vs locked themes (locked ones show their unlock
  condition, e.g. "Cyberpunk — reach the Cyberpunk Age").
- `theme <name>` — switches directly to a theme by key/name if unlocked; errors with
  the unlock hint if locked, errors "unknown theme" otherwise.

Add a `case "theme":` to the dispatch switch returning the above.

### Live preview (apply-on-highlight, revert-on-cancel)
Exactly the load-game detail-pane pattern: the picker list's `SetChangedFunc` fires
on highlight. On highlight we **apply the theme for real** (remap + restyle + redraw)
so the player sees the whole UI in that theme immediately — the picker is itself a
live sample of the running UI. We remember the theme that was active on open.
- **Confirm** (Enter / select): persist the highlighted theme as active via the
  account layer; close.
- **Cancel** (Esc): re-apply the remembered original theme; close. No persistence.

Because retinting is a global remap, "preview" and "apply" are the same operation —
the only difference is whether we persist and whether cancel reverts. Clean.

### Palette swatches
The picker detail pane shows the theme's blurb plus a swatch row — one colored block
per role rendered with that role's color tag, labeled (Accent / Positive / Negative /
…). For accessible themes, show the gain/loss **glyphs** next to the ± swatches so the
redundant encoding is visible in the picker itself. Locked themes show swatches dimmed
with a lock marker and the unlock condition.

### Unlock notification
When a milestone grants a theme (§5), surface it through the existing toast/log
channel (the same path milestone completions already use), e.g. *"New theme unlocked:
Cyberpunk — switch in Themes (`theme cyberpunk`)."* Non-modal; don't interrupt play.
**Gate the toast on `account.UnlockTheme` returning `newly == true`** — `UnlockTheme` is
idempotent (re-unlocking an owned theme is a no-op that returns `newly == false`), so a
milestone re-fire or a redundant chain check must not produce a duplicate "unlocked!"
toast. Only a genuinely-new unlock notifies.

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

### Check: colorblind distinguishability (simulation, not luminance)

Luminance contrast and colorblind distinguishability are **different properties** — two
hues can clear 4.5:1 against the background yet be nearly identical to a deuteranope.
The accessible palettes (§4) are *designed* to keep Accent/Positive/Negative/Highlight
mutually distinct under deutan/protan deficiency, but "designed to" must be backed by a
test, not by a sighted developer's eyeball.

So the guard ALSO runs the accessible palettes through **deuteranopia and protanopia
simulation** (a standard CVD transform — Brettel/Viénot or Machado — applied to each
role's RGB), then asserts that the post-simulation role colors stay separated by a
minimum perceptual distance (e.g. a ΔE floor in a perceptually-uniform space). This is
what actually catches an **Accent ≈ Positive collision** under simulated deficiency — a
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
nearest palette slot — and a pair that clears 4.5:1 in truecolor can collapse below the
floor once both ends snap to neighboring palette entries. This is worst for
**light-background themes like Parchment**, where the foreground/background luminances
are already close and quantization has more room to flip the ratio. So the strong claim
("cannot reach players") only holds on truecolor terminals; on 256-color terminals a
"passing" theme can still be marginal.

We have two honest options, and should pick one rather than wave at it:

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
  entirely and is the honest long-term form. Large but mechanical; do it lineage-area
  by lineage-area. Mandatory only if a tview bump breaks remap.
- **User-authored themes.** Read extra themes from `data/themes/*.json`, run them
  through the same contrast guard at load, reject ones that fail. Lets the community
  share palettes without code changes.

---

## 10. Phased Implementation Plan

Ordered so something visible lands early. Each phase is independently shippable.

### Phase 1 — Theme spine + Forge + accessibility themes (visible win)
- Add the `theme` package: `Theme`/`Role` model, palette accessors, the name-remap
  apply/restore (`remap.go`), the restylable-widget registry (`restyle.go`).
- Define **Forge**, **Deuteranopia**, **Protanopia**, **High Contrast** (accessible,
  default-unlocked).
- Convert the 23 hex tags + the two bar-color constants to role tokens (§3.4).
- Add the `theme` command + main-menu Themes entry + picker with live preview and
  swatches.
- Add the contrast guard test (§8) and the remap guard test (§3.6).
- **Theme choice persists in-memory / process-local for now** if `accounts.md` isn't
  landed yet — wire to a stub `ActiveTheme/SetActiveTheme` so the picker works end to
  end. Visible result: a player can switch between four legible themes live.
- **Cross-doc dependency: Phase 1 has NO dependency on `accounts.md`.** It ships against a
  stub account API (`HasTheme`/`UnlockTheme`/`ActiveTheme`/`SetActiveTheme`), so theming
  Phase 1 can land *before* the account system exists. This is deliberate — it keeps the
  two tracks from being scheduled into a deadlock.

### Phase 2 — Account-wide persistence
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

### Phase 3 — Flavor themes + milestone unlocks
- Define flavor themes (Parchment, Bronze, Cyberpunk, Monochrome, Cosmic), each
  tuned to pass the contrast guard.
- Add the theme-unlock resolver hooked to `CheckMilestones`/`CheckChains`; persist
  unlocks account-wide; fire the unlock notification.
- Picker shows locked themes with unlock conditions; `theme list` reflects state.

### Phase 4 — Stretch
- Epoch-adaptive Adaptive theme (generalize `ApplyAgePalette`).
- Begin the semantic-token cleanup (`theme.Tag`) — and keep it on the shelf as the
  hard fallback if a dependency bump ever breaks the remap.
- Optionally: user-authored `data/themes/*.json` with contrast-gated loading.

---

## Appendix — wiki sync

Per project rules, the shipping change updates `site/`:
- `site/docs/commands.md` — document the `theme` command (`theme`, `theme list`,
  `theme <name>`).
- A new/updated accessibility section in the wiki noting colorblind + high-contrast
  themes ship unlocked.
- Any wiki page that asserts "green = gain / red = loss" should note accessible
  themes use blue/orange + glyphs instead.
