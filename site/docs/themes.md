# 🎨 Themes & Accessibility

AgeForge's interface colors are driven by a set of swappable **themes** — 11 in all, both **dark** and **light**. You can switch between them **live** at any time, and your choice sticks: it's saved with your **account**, not with any individual game save, so it travels across every save and every new game. Loading an old save never changes your theme.

Themes come in three groups: **Standard** (the default dark **Forge** and the light **Daylight**), **Accessibility** (colorblind-safe and high-contrast, in dark and light), and **Unlockable** flavor themes you earn by reaching later ages. Standard and Accessibility themes are always unlocked.

---

## 🎨 The themes

| Theme | Key | Group | Variant | How to get it |
|---|---|---|---|---|
| **Forge** | `forge` | Standard | Dark | The classic dark + gold look — active by default |
| **Daylight** | `daylight` | Standard | Light | Unlocked from the start — charcoal ink on an off-white page, white panels, deep-amber accents |
| **Deuteranopia-safe** | `deuteranopia` | Accessibility | Dark | Unlocked from the start |
| **Protanopia-safe** | `protanopia` | Accessibility | Dark | Unlocked from the start |
| **High Contrast** | `high_contrast` | Accessibility | Dark | Unlocked from the start |
| **High Contrast Light** | `high_contrast_light` | Accessibility | Light | Unlocked from the start |
| **Parchment** | `parchment` | Unlockable | Light | Reach the Renaissance Age |
| **Bronze** | `bronze` | Unlockable | Dark | Reach the Bronze Age |
| **Cyberpunk** | `cyberpunk` | Unlockable | Dark | Reach the Cyberpunk Age |
| **Monochrome** | `monochrome` | Unlockable | Dark | Reach the Information Age |
| **Cosmic** | `cosmic` | Unlockable | Dark | Reach the Galactic Age |

### Every theme paints its own background

A theme doesn't just recolor the text — it paints the **whole surface**: the page background, panels, borders, selection, text tiers, and the good / warn / bad colors. So a light theme looks right on a dark terminal, and a dark theme looks right on a light terminal; you don't need to match your terminal's own colors to the theme.

A **truecolor** terminal is recommended for exact colors.

---

## 🔀 How to switch

You can change theme from the **`theme` command** at the `>` prompt, or from the **Themes** entry on the main menu. Both open the same picker.

| Command | What it does |
|---|---|
| `theme` | Open the live theme picker |
| `theme list` | List every theme by name and key, marking the active one and showing each theme's light/dark variant, which are accessible, and the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`). A locked theme is refused with its unlock hint |

```
theme
theme list
theme daylight
```

### The picker

The picker is a full-screen, **live-preview** screen. Themes are grouped under **Standard**, **Accessibility**, and **Unlockable** headings, and each row is tagged **light** or **dark**, **accessible**, or **locked** (`🔒`).

| Key | Action |
|---|---|
| `↑` / `↓` | Preview a theme — the whole UI retints instantly so you can judge it in place |
| `Enter` | Keep the highlighted theme (only if it's unlocked) |
| `Esc` / `q` | Cancel and revert to whatever you had before |

Beside the list, a **details pane** shows the highlighted theme's:

- one-line description, its **Light/Dark** variant and group;
- the `▲` / `▼` gain/loss glyphs, for accessible themes;
- the **unlock condition**, for locked themes;
- **palette swatches** — a colored block for each color role;
- a small **sample panel** painted in that theme's own colors, so you can see it on its real background.

A **locked** theme can still be previewed, but `Enter` won't keep one you haven't earned yet.

---

## ♿ Accessibility

Color is never the only signal. Four accessibility themes ship **unlocked from the start** and are **never gated** — they're always available:

- **Deuteranopia-safe** and **Protanopia-safe** — for red-green color vision deficiency.
- **High Contrast** — maximum legibility on a near-black background, for low-vision players and high-glare terminals.
- **High Contrast Light** — the light counterpart: black ink on a white page, with text at **AAA (7:1)** contrast.

Because AgeForge normally uses **green for gains and red for losses**, the accessible themes drop that pairing entirely. Instead:

- Gains are **blue**, losses are **orange** — the standard colorblind-safe opposition, so red-green colorblind players can tell `+` from `−`.
- `▲` (gain) and `▼` (loss) glyphs mark the sign by **shape as well as color**, a redundant non-color cue so the direction reads even if the hues don't.

When an accessible theme is active, the usual "green = gain / red = loss" UI legends adapt to **blue / orange + glyphs** to match.

Every shipped theme is **contrast-checked** (WCAG AA for text on its background). The accessibility themes additionally pass a **colorblind-distinguishability** check.

---

## 🗺️ Maps on light themes

The maps follow your theme too, including light ones:

- **City Map** — re-keys itself for a light page: lighter ground, streets lighter still, drop shadows and shaded walls darker, and building labels darkened so they stay readable. Space-age cities stay dark — a starfield is dark by nature.
- **World Map** — civilization markers stay marker-bright against the map's own canvas, whatever the theme. Space-age star-maps stay dark.

---

## 🔓 Unlocking flavor themes

Beyond the Standard and Accessibility groups, AgeForge ships **flavor themes** — purely cosmetic looks you unlock by reaching an age:

| Theme | Variant | Unlocks when you… |
|---|---|---|
| **Bronze** | Dark | Reach the Bronze Age |
| **Parchment** | Light | Reach the Renaissance Age |
| **Monochrome** | Dark | Reach the Information Age |
| **Cyberpunk** | Dark | Reach the Cyberpunk Age |
| **Cosmic** | Dark | Reach the Galactic Age |

Unlocks are **account-wide and permanent**: earn a theme on one empire and it's yours on **every save and every future new game** — exactly like your accessibility themes.

Until you've earned it, a flavor theme shows in the picker (and in `theme list`) with a `🔒` and its unlock condition. You can still preview a locked theme, but you can't make it your active theme until you reach the age that unlocks it.

---

## See also

- [All Commands](commands.md) — the full `theme` command reference
- [Account & Recovery](account.md) — how theme unlocks (and other account-wide progress) persist and travel between machines
- [The 22 Ages](ages.md) — the ages that unlock the flavor themes
- [The City Map](city-map.md) and [The World Map](world-map.md) — the map views
