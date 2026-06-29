# 🎨 Themes & Accessibility

AgeForge's interface colors are driven by a set of swappable **themes** — 9 in all. You can switch between them **live** at any time, and your choice sticks: it's saved with your **account**, not with any individual game save, so it travels across every save and every new game. Loading an old save never changes your theme.

Themes come in three flavors: the default **Forge** look, a set of **accessibility** themes (colorblind-safe and high-contrast, always unlocked), and **flavor** themes you unlock by reaching later ages.

---

## 🎨 The themes

| Theme | Key | Type | How to get it |
|---|---|---|---|
| **Forge** | `forge` | Default | The classic dark + gold look — active by default |
| **Deuteranopia-safe** | `deuteranopia` | Accessibility | Unlocked from the start |
| **Protanopia-safe** | `protanopia` | Accessibility | Unlocked from the start |
| **High Contrast** | `high_contrast` | Accessibility | Unlocked from the start |
| **Bronze** | `bronze` | Flavor | Reach the Bronze Age |
| **Parchment** | `parchment` | Flavor | Reach the Renaissance Age |
| **Monochrome** | `monochrome` | Flavor | Reach the Information Age |
| **Cyberpunk** | `cyberpunk` | Flavor | Reach the Cyberpunk Age |
| **Cosmic** | `cosmic` | Flavor | Reach the Galactic Age |

---

## 🔀 How to switch

You can change theme from the **`theme` command** at the `>` prompt, or from the **Themes** entry on the main menu. Both open the same picker.

| Command | What it does |
|---|---|
| `theme` | Open the live theme picker |
| `theme list` | List every theme by name and key, marking the active one and showing the lock status of any theme you haven't unlocked |
| `theme <key>` | Switch directly to a theme by key (e.g. `theme high_contrast`) |

```
theme
theme list
theme high_contrast
```

### The picker

The picker is a full-screen, **live-preview** screen:

| Key | Action |
|---|---|
| `↑` / `↓` | Preview a theme — the whole UI retints instantly so you can judge it in place |
| `Enter` | Keep the highlighted theme |
| `Esc` | Cancel and revert to whatever you had before |

It shows **palette swatches** (colored blocks for each color role) and, for the accessibility themes, the `▲` / `▼` glyphs. A **locked** flavor theme can still be previewed, but `Enter` won't keep one you haven't earned yet.

---

## ♿ Accessibility

Color is never the only signal. Three accessibility themes ship **unlocked from the start** and are **never gated** — they're always available:

- **Deuteranopia-safe** and **Protanopia-safe** — for red-green color vision deficiency.
- **High Contrast** — maximum legibility on a near-black background, for low-vision players and high-glare terminals.

Because AgeForge normally uses **green for gains and red for losses**, the accessible themes drop that pairing entirely. Instead:

- Gains are **blue**, losses are **orange** — the standard colorblind-safe opposition, so red-green colorblind players can tell `+` from `−`.
- `▲` (gain) and `▼` (loss) glyphs mark the sign by **shape as well as color**, a redundant non-color cue so the direction reads even if the hues don't.

When an accessible theme is active, the usual "green = gain / red = loss" UI legends adapt to **blue / orange + glyphs** to match.

Every shipped theme is **contrast-checked** (WCAG AA for text on its background). The accessibility themes additionally pass a **colorblind-distinguishability** check.

---

## 🔓 Unlocking flavor themes

Beyond Forge and the accessibility set, AgeForge ships **flavor themes** — purely cosmetic looks you unlock by reaching an age:

| Theme | Unlocks when you… |
|---|---|
| **Bronze** | Reach the Bronze Age |
| **Parchment** | Reach the Renaissance Age |
| **Monochrome** | Reach the Information Age |
| **Cyberpunk** | Reach the Cyberpunk Age |
| **Cosmic** | Reach the Galactic Age |

Unlocks are **account-wide and permanent**: earn a theme on one empire and it's yours on **every save and every future new game** — exactly like your accessibility themes.

Until you've earned it, a flavor theme shows in the picker (and in `theme list`) with a `🔒` and its unlock condition. You can still preview a locked theme, but you can't make it your active theme until you reach the age that unlocks it.

---

## See also

- [All Commands](commands.md) — the full `theme` command reference
- [Account & Recovery](account.md) — how theme unlocks (and other account-wide progress) persist and travel between machines
- [The 22 Ages](ages.md) — the ages that unlock the flavor themes
