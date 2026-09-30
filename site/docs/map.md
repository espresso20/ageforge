# The Map

Open it with the `map` command. The **Map** draws your empire from your **real game state**: the buildings you own, the workers staffing them, your wonders, the civilizations you have met, your trade routes and wars. Build, trade or meet someone new and the map changes to match.

The Map has two **styles** you can switch between, both drawn from the same map model:

- **Roguelike** (the default): a glyph world seen from above. Every cell is one colored character that means something.
- **Skyline**: your empire side-on as a panorama, one district for every age you have lived through.

`citymap` and `worldmap` still work: both open the Map. `worldmap` opens it on the known world (the roguelike style's region zoom).

---

## The Map panel

`map` opens the Map full screen. These keys work in every style:

| Key | Action |
|---|---|
| `s` | Next style (roguelike, skyline) |
| `g` | Next glyph set (ascii, unicode, nerd) |
| `Enter` | Put the command for what the cursor is on into the prompt. It is not run: press Enter again to run it |
| `Esc` | Close the Map |

The bottom row is a **key bar**: it shows the current style and glyph set, the keys you can press, and the command for the thing under the cursor.

### The inspect cursor

In both styles a cursor picks out one thing at a time: a building, a wonder, a civilization. The Map reports its **name**, a few **details** (how many you own, how many workers staff it, what changed) and the **whole command** to type for it, for example `build hut` or `diplomacy gift riverlands_tribes`. Press `Enter` to put that command into the prompt, edit it if you like, and press Enter again to run it.

### Since your last visit

A **news line** on the Map lists what happened since the save was loaded: buildings built, civilizations met, wars, new trade routes and so on. The cursor's details say it too ("+2 since your last visit"). In the roguelike style, `n` highlights everything that changed; in the skyline, `c` does.

---

## Roguelike

The roguelike style draws your settlement as a small glyph world. Your town grows out of your **real buildings**, quarter by quarter, with streets, walls, wonders and people walking the streets. The same save always grows the same town, and new buildings join the existing streets instead of reshuffling them.

This age's **wonder** appears on its plot, under scaffolding, once you start it: bank some of its cost or queue its construction. Until then its plot stays empty.

**Walls** follow your age: none at first, a palisade from the Bronze Age, then a stone wall with towers, and later a ring boulevard. The ring grows in steps as the town outgrows it. Harbor buildings sit on the shore and read as piers or jetties; that is intended.

### Three zooms

| Zoom | What you see |
|---|---|
| **Region** | The whole known world, drawn on your epoch's cartographic plate, with the civilizations you have met |
| **Settlement** | Your town, one tile per cell |
| **District** | A closer look, one tile per two cells, with building names |

The district zoom names only a handful of building types at a time (those nearest the cursor first), so it stays readable.

### Roguelike keys

| Key | Action |
|---|---|
| `←` `→` `↑` `↓` or `h` `j` `k` `l` | Move the cursor |
| `Shift`+arrow or `H` `J` `K` `L` | Move the cursor by 8 |
| `z` | Zoom in |
| `x` | Zoom out |
| `Tab` / `Shift-Tab` | Jump between buildings and wonders (and civilizations at region zoom) |
| `c` | Center on the town square |
| `f` | Flows overlay: full stores, understaffed buildings, idle workers |
| `?` | Legend: what each glyph means |
| `n` | Highlight what changed since your last visit |

---

## Skyline

The skyline style shows your empire **side-on**, as an ANSI-art panorama. It has one **district for every age** you have lived through: the oldest in the west, the present in the east, and your build queue as cranes on the frontier.

- **Windows light and chimneys smoke only where workers are staffed.** An unstaffed building stays dark.
- **The sky** follows the game clock and the weather.
- **The far ridge** holds the civilizations you have met, each drawn as a town, and the harbinger.
- **Traffic follows real state.** Trade routes travel by land, sea or air depending on the route: harbor and ship routes sail the bays, warp and stellar routes fly once your age has aircraft, and rails and caravans stay on land.
- In the **Monochrome** and **Parchment** themes the skyline is drawn as a two-tone (duotone) picture.

### Skyline keys

| Key | Action |
|---|---|
| `←` `→` or `h` `l` | Scroll (or move the cursor while inspecting) |
| `H` `L` | Scroll half a screen |
| `Home` | The oldest district |
| `End` | The present |
| `i` | Show or hide the inspect cursor |
| `↑` `↓` or `k` `j` | Move between the building rows and the ridge |
| `Tab` / `Shift-Tab` | Step through every target |
| `f` | Flows: full stores, understaffed buildings, idle workers |
| `c` | Highlight what changed since your last visit |

---

## Settings

Two settings shape the Map. Both are saved **per account**, like your theme: they carry across saves and new games, and switching accounts swaps them. With no account loaded they last for the session.

| Command | What it does |
|---|---|
| `map style` | Show the current style |
| `map style roguelike` | Use the roguelike style (the default) |
| `map style skyline` | Use the skyline style |
| `style` | Same as `map style`: `style skyline` switches too |
| `map glyphs` | Show the current glyph set |
| `map glyphs ascii` | Plain ASCII, for fonts with poor symbol coverage |
| `map glyphs unicode` | Box drawing, blocks and widely supported symbols (the default) |
| `map glyphs nerd` | Nerd Font icons. Needs a Nerd Font in your terminal; every icon has a Unicode fallback |

The `s` and `g` keys in the panel change the same settings.

---

## The mini map

The dashboard shows a compact view of the Map above the **Buildings** list, in a border titled "Map · Roguelike" or "Map · Skyline". It follows your `map style` and `map glyphs` settings and shows the since-last-visit news too (for the skyline, in its bottom border).

The mini map needs room: it appears on terminals of about 120x40 and larger. On smaller terminals (80x24, 100x30) it hides and the Buildings list gets the space. Type `map` to open the full panel at any size.

---

## Icons (Nerd Fonts)

The `nerd` glyph set draws real icons, but only if your terminal uses a **Nerd Font**. The `icons` command opens a window over the dashboard that walks you through it. The window has its own buttons: press a button's letter, or move with `Tab` or the arrow keys and press `Enter`. `Esc` closes it at any step.

1. **Check.** It shows sample Nerd Font icons next to their Unicode fallbacks. Press **I see icons** (`I`) or **I see boxes** (`B`). If you see icons, it sets `map glyphs nerd` for you.
2. **Install.** If you see boxes, it offers to install **JetBrains Mono Nerd Font** for your user: **Install** (`Y`) or **No, show me how** (`N`). Install downloads the official JetBrains Mono zip from the [Nerd Fonts v3.5.1 release](https://github.com/ryanoasis/nerd-fonts/releases) (about 130 MB), checks its SHA-256, and installs the 16 font files of the font's Nerd Font Mono variant for **your user only**, with no admin rights:
   - **macOS:** `~/Library/Fonts`
   - **Linux:** `~/.local/share/fonts` (or `$XDG_DATA_HOME/fonts`), then `fc-cache -f` if it is available
   - **Windows:** `%LOCALAPPDATA%\Microsoft\Windows\Fonts`, plus per-user font entries in the registry (under HKCU, so no admin rights)

   The install runs in the background with its progress in the window. You can close the window and keep playing: `icons` opens it again, and the log tells you when the install is done. It never runs anything else and never edits your terminal's config files. If the download fails or the checksum does not match, it says so, changes nothing and offers **Try again** (`R`). **No, show me how** points you to [nerdfonts.com/font-downloads](https://www.nerdfonts.com/font-downloads) to install one yourself.
3. **Select the font.** It detects your terminal (Terminal.app, iTerm2, VS Code, WezTerm, Windows Terminal, kitty, Alacritty, Konsole, GNOME Terminal; generic steps for anything else) and shows the two or three steps to select the font. It gives the font's exact name as your terminal lists it.
4. **Restart.** Restart your terminal, then run `icons` again to check.

The first time you open the Map it shows a one-time hint: "Want real icons on your map? Type icons." It is remembered per account.

---

## Themes

Both styles draw their colors from your active [theme](themes.md), so switching themes recolors the Map and the mini map at once, light themes included.

---

## See also

- [All Commands](commands.md#map): the `map` command reference
- [Themes & Accessibility](themes.md)
- [Wonders](wonders.md)
- [Trade & Diplomacy](trade.md): the routes and civilizations the Map draws
