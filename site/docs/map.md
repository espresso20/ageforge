# The Map

Open it with the `map` command. The **Map** draws your empire from your **real game state**: the buildings you own, the workers staffing them, your wonders, the civilizations you have met, your trade routes and wars. Build, trade or meet someone new and the map changes to match.

The Map has two **styles** you can switch between, both drawn from the same map model:

- **Roguelike** (the default): a glyph world seen from above. Every cell is one colored character that means something.
- **Skyline**: your empire side-on as a panorama, one district for every age you have lived through.

The Map shows only what you have reached. Civilizations you have not met are nowhere on it: no name, no town, no marker where they live. It names no age or era you have not reached either (the next age appears once it is within reach), and the harbinger on the map only "warns of impending doom".

`citymap` and `worldmap` still work: both open the Map. `worldmap` opens it on the known world (the roguelike style's region zoom).

---

## The Map panel

`map` opens the Map over the whole screen except the **command bar**, which keeps working: type any command while the Map is open, just as you would with it closed, and watch the map change. A command that opens another panel (say `research`) swaps the Map for it.

Because typing goes to the command bar, the Map takes only the keys that print nothing. These work in every style:

| Key | Action |
|---|---|
| Arrows | Move the cursor or scroll (see each style below) |
| `Tab` / `Shift-Tab` | Jump to the next or previous thing to inspect |
| `PgUp` / `PgDn` | Zoom out and in (roguelike) or scroll half a screen (skyline) |
| `Enter` | Put the command for what the cursor is on into the prompt. It is not run: press Enter again to run it |
| `Esc` | Close the Map |

Once you have typed something, `Tab`, `Shift-Tab` and `Enter` act on the prompt instead (complete the command, run it), as they do everywhere else. The arrows stay with the Map.

The settings are commands you can type with the Map open: `map style`, `map glyphs`, `map flows` and `minimap` (see Settings below).

The bottom row is a **key bar**: it shows what `Enter` will do, the current style and glyph set, and the reply to the last command you typed (the log is hidden behind the Map).

### The inspect cursor

In both styles a cursor picks out one thing at a time: a building, a wonder, a civilization. The Map reports its **name**, a few **details** (how many you own, how many workers staff it, what changed) and the **whole command** to type for it, for example `build hut` or `diplomacy gift riverlands_tribes`. Press `Enter` to put that command into the prompt, edit it if you like, and press Enter again to run it.

### Since your last visit

A **news line** on the Map lists what happened since the save was loaded: buildings built, civilizations met, wars, new trade routes and so on. The cursor's details say it too ("+2 since your last visit"), and both styles highlight what changed.

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
| `←` `→` `↑` `↓` | Move the cursor |
| `Shift`+arrow | Move the cursor by 8 |
| `PgUp` | Zoom out (settlement, then region) |
| `PgDn` | Zoom in (settlement, then district) |
| `Tab` / `Shift-Tab` | Jump between buildings and wonders (and civilizations at region zoom) |
| `Home` | Center on the town square |

The legend (what each glyph means) shows beside the map when there is room.

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
| `←` `→` | Scroll (or move the cursor while inspecting) |
| `Shift`+`←` `→`, `PgUp` `PgDn` | Scroll half a screen |
| `Home` | The oldest district |
| `End` | The present |
| `Tab` / `Shift-Tab` | Put the inspect cursor out, then step through every target |
| `↑` `↓` | Move between the building rows and the ridge (while inspecting) |

---

## Settings

Three settings shape the Map. They are saved **per account**, like your theme: they carry across saves and new games, and switching accounts swaps them. With no account loaded they last for the session.

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
| `minimap` | Show whether the dashboard's mini map is on |
| `minimap off` | Hide the mini map, so the Buildings list gets the whole column |
| `minimap on` | Show the mini map again (the default) |
| `map flows` | The flows overlay: full stores, understaffed buildings, idle workers. `map flows on` and `map flows off` set it; bare, it switches. It lasts for the session |

All of them work while the Map is open, so you see the change at once.

---

## The mini map

The dashboard shows a compact view of the Map above the **Buildings** list, in a border titled "Map · Roguelike" or "Map · Skyline". It follows your `map style` and `map glyphs` settings and shows the since-last-visit news too (for the skyline, in its bottom border).

It is kept short, at most 9 rows inside its border and about a quarter of the column, so the Buildings list keeps most of the room. `minimap off` hides it and gives the list the whole column; `minimap on` brings it back.

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
