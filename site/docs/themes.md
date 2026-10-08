# 🎨 Themes & Accessibility

AgeForge's interface colors come from swappable **themes**: 16 themes in all, **dark** and **light**. You can switch at any time and the change applies at once. Your choice is saved with your **account**, not with a game save, so it carries across every save and every new game. Loading an old save never changes your theme.

Themes come in three groups: **Standard** (the default dark **Forge** and the light **Daylight**), **Accessibility** (colorblind-safe and high-contrast, in dark and light), and **Unlockable** flavor themes that come with a [badge](badges.md): five for reaching an age, five for other badges. Standard and Accessibility themes are always unlocked.

---

## 🎨 The themes

| Theme | Key | Group | Variant | How to get it |
|---|---|---|---|---|
| **Forge** | `forge` | Standard | Dark | The classic dark and gold look, active by default |
| **Daylight** | `daylight` | Standard | Light | Unlocked from the start. Charcoal ink on an off-white page, white panels, deep-amber accents |
| **Deuteranopia-safe** | `deuteranopia` | Accessibility | Dark | Unlocked from the start |
| **Protanopia-safe** | `protanopia` | Accessibility | Dark | Unlocked from the start |
| **High Contrast** | `high_contrast` | Accessibility | Dark | Unlocked from the start |
| **High Contrast Light** | `high_contrast_light` | Accessibility | Light | Unlocked from the start |
| **Parchment** | `parchment` | Unlockable | Light | Reach the Renaissance Age |
| **Bronze** | `bronze` | Unlockable | Dark | Reach the Bronze Age |
| **Cyberpunk** | `cyberpunk` | Unlockable | Dark | Reach the Cyberpunk Age |
| **Monochrome** | `monochrome` | Unlockable | Dark | Reach the Information Age |
| **Cosmic** | `cosmic` | Unlockable | Dark | Reach the Galactic Age |
| **Source** | `source` | Unlockable | Dark | Given by a secret badge. Phosphor green on black, with code rain |
| **Glitch** | `glitch` | Unlockable | Dark | Given by a secret badge. Cyan and magenta on violet black, with tears of static |
| **Ashfall** | `ashfall` | Unlockable | Dark | Succumb to every era's catastrophe. Ember orange on soot, with sparks rising |
| **Ledger** | `ledger` | Unlockable | Light | Take 100 deals from civilizations. Ink and bookkeeper's green on green-bar paper |
| **Prismatic** | `prismatic` | Unlockable | Dark | Earn 400 badges. Starlight on blue-black, with glints that turn through the spectrum |

### Every theme paints its own background

A theme paints the **whole surface**, text included: the page background, panels, borders, selection, text tiers, and the good / warn / bad colors. A light theme looks right on a dark terminal and a dark theme looks right on a light one, so you don't need to match your terminal's colors to the theme.

<div class="screen-pair">
<figure class="screen" data-screen="theme-forge"><figcaption>The Workers panel in Forge, the default theme.</figcaption></figure>
<figure class="screen" data-screen="theme-daylight"><figcaption>The same panel in Daylight.</figcaption></figure>
</div>

A **truecolor** terminal is recommended for exact colors.

### Motion

A few things in the game move on their own: the [Map](map.md) and the mini map, some badges in the [badge case](account.md#the-badge-case), and the ambient effect of the five themes that have one. One setting turns all of it off:

| Command | What it does |
|---|---|
| `motion` | Show whether motion is on |
| `motion off` | Hold everything still: each of those is drawn as its first frame, and nothing is animated |
| `motion on` | Bring it back (the default) |

Nothing is hidden with motion off, and nothing in a game changes: the same badges, the same map, drawn still. The setting is saved per account; with no account loaded it lasts for the session.

---

## 🔀 How to switch

You can change theme from the **`theme` command** at the `>` prompt, or from the **Themes** entry on the main menu. Both open the same picker.

| Command | What it does |
|---|---|
| `theme` | Open the live theme picker |
| `theme list` | List every theme by name and key. It marks the active one and shows each theme's light/dark variant, which ones are accessible, and the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`). A locked theme is refused with its unlock hint |

```
theme
theme list
theme daylight
```

### The picker

The picker and `theme list` show a locked theme's unlock condition. One earned in an age past your next one reads "Reach a later age" instead of naming the age, so the list doesn't spoil ages you haven't reached (the table above has them all). A theme whose badge is still hidden from you says only what kind of badge gives it.

The picker fills the window and **previews** each theme as you move to it. Themes are grouped under **Standard**, **Accessibility**, and **Unlockable** headings, and each row is tagged **light** or **dark**, **accessible**, or **locked** (`🔒`).

| Key | Action |
|---|---|
| `↑` / `↓` | Preview a theme. The whole interface retints at once so you can judge it in place |
| `Enter` | Keep the highlighted theme (only if it's unlocked) |
| `Esc` / `q` | Cancel and revert to whatever you had before |

Beside the list, a **details pane** shows the highlighted theme's:

- one-line description, its **Light/Dark** variant and group;
- the `▲` / `▼` gain/loss glyphs, for accessible themes;
- the **unlock condition**, for locked themes;
- **palette swatches**, a colored block for each color role;
- a small **sample panel** painted in that theme's own colors, so you can see it on its real background.

A **locked** theme can still be previewed, but `Enter` won't keep one you haven't earned yet.

---

## ♿ Accessibility

Color is never the only signal. Four accessibility themes are **unlocked from the start** for everyone:

- **Deuteranopia-safe** and **Protanopia-safe**, for red-green color vision deficiency.
- **High Contrast**, for the most legible text on a near-black background. It suits low-vision players and high-glare screens.
- **High Contrast Light**, the light counterpart: black ink on a white page, with text at **AAA (7:1)** contrast.

AgeForge normally uses **green for gains and red for losses**. The accessible themes drop that pairing:

- Gains are **blue** and losses **orange**, the standard colorblind-safe pair, so red-green colorblind players can tell `+` from `−`.
- `▲` (gain) and `▼` (loss) glyphs mark the sign by **shape as well as color**, so the direction reads even if the hues don't.

When an accessible theme is active, the usual "green = gain / red = loss" UI legends adapt to **blue / orange + glyphs** to match.

Every shipped theme is **contrast-checked** (WCAG AA for text on its background). The accessibility themes additionally pass a **colorblind-distinguishability** check.

---

## 🗺️ The Map and your theme

The [Map](map.md) follows your theme too, including light ones. Both styles, roguelike and skyline, take their colors from the theme, so switching themes recolors the Map and the dashboard's mini map at once. In the **Monochrome** and **Parchment** themes the skyline is drawn as a two-tone (duotone) picture.

---

## 🔓 Unlocking flavor themes

Beyond the Standard and Accessibility groups, AgeForge has ten **flavor themes**. Each is the reward of a [badge](badges.md#themes-from-badges): earn the badge and the theme is unlocked.

| Theme | Variant | Unlocks when you… |
|---|---|---|
| **Bronze** | Dark | Reach the Bronze Age |
| **Parchment** | Light | Reach the Renaissance Age |
| **Monochrome** | Dark | Reach the Information Age |
| **Cyberpunk** | Dark | Reach the Cyberpunk Age |
| **Cosmic** | Dark | Reach the Galactic Age |
| **Ledger** | Light | Take 100 deals from civilizations, across all your runs |
| **Ashfall** | Dark | Succumb to every era's catastrophe |
| **Prismatic** | Dark | Earn 400 badges |
| **Source** | Dark | Earn a secret badge |
| **Glitch** | Dark | Earn another secret badge |

Unlocks are **account-wide and permanent**: earn a theme on one empire and it's yours on **every save and every future new game**, just like the accessibility themes. When a badge unlocks a theme, the log says so.

The five age themes come with the badge for reaching their age. They used to come from milestones; the condition is the same, and an account that had already reached an age keeps its theme, or gets it the first time this version opens the account.

Until you've earned it, a flavor theme shows in the picker (and in `theme list`) with a `🔒` and its unlock condition. You can still preview a locked theme, but you can't make it your active theme until you earn the badge that unlocks it. The picker never says more than the badge case does: Source and Glitch say only that a secret badge gives them, and Ashfall and Ledger say what kind of badge gives them until that badge is in sight.

### Ambient effects

Five themes carry an ambient effect on the dashboard:

| Theme | Effect |
|---|---|
| **Source** | Code falls down the empty columns |
| **Glitch** | A tear of static crosses the screen now and then |
| **Ashfall** | Sparks drift up the empty columns |
| **Ledger** | Every other row of the empty space is tinted, like green-bar paper |
| **Prismatic** | Glints in the empty space turn slowly through the spectrum |

An effect is only ever drawn in cells that are otherwise empty, with clear space between it and any text, and never in the command bar, so nothing you read or type is covered. `motion off` turns the effects off (see [Motion](#motion)); the colors stay.

---

## See also

- [All Commands](commands.md): the full `theme` command reference
- [Badges](badges.md): the badges that unlock the flavor themes
- [Account & Recovery](account.md): how theme unlocks (and other account-wide progress) are kept and moved between machines
- [The 22 Ages](ages.md): the ages whose badges unlock five of them
- [The Map](map.md): the map styles and glyph sets
