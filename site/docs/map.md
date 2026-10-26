# The Map

Open it with the `map` command. The Map draws your empire from your real game state: the buildings you own, the workers staffing them, your wonders, the civilizations you have met, your trade routes and wars. Build, trade or meet someone new and the map changes to match.

The Map is a glyph world seen from above. Every cell is one colored character that means something.

The Map shows only what you have reached. Civilizations you have not met are nowhere on it: no name, no town, no marker where they live. It names no era you have not reached and no age past your next one. The harbinger on the map warns of impending doom (in the Cosmic Era, it may warn of the Last Passage instead), never of the era to come.

`citymap` and `worldmap` are aliases of `map`. Both open the Map, and `worldmap` opens it on the known world (the region zoom).

---

## The Map panel

`map` opens the Map over the whole screen except the command bar, which keeps working: type any command while the Map is open, just as you would with it closed, and watch the map change. A command that opens another panel (say `research`) swaps the Map for it.

Because typing goes to the command bar, the Map takes only the keys that print nothing. These work:

| Key | Action |
|---|---|
| Arrows | Move the cursor |
| `Tab` / `Shift-Tab` | Jump to the next or previous thing to inspect |
| `PgUp` / `PgDn` (`Option+↑` / `Option+↓`), or `-` / `=` | Zoom out and in |
| `Enter` | Put the command for what the cursor is on into the prompt. It is not run: press `Enter` again to run it |
| `Esc` | Close the Map |

Once you have typed something, `Tab`, `Shift-Tab` and `Enter` act on the prompt instead (complete the command, run it), as they do everywhere else. The arrows, `PgUp`, `PgDn`, `Home` and `End` stay with the Map.

A few keys that print are the Map's too, while the prompt is empty (no command begins with one): `+` and `=` zoom in, `-` and `_` zoom out. A keyboard with no Page keys (a Mac's, for one) can zoom with `-` and `=` alone. `Option+↑` and `Option+↓` also work as `PgUp` and `PgDn` here, and `Option+←` and `Option+→` as `Home` and `End` (see [Commands](commands.md#the-prompt)).

The settings are commands you can type with the Map open: `map glyphs`, `map flows` and `minimap` (see [Settings](#settings)). With VibeMode on, the words on the Map come in its second script too (see [VibeMode](commands.md#vibemode)).

The bottom row is a key bar. It shows what `Enter` will do, the current glyph set, and the reply to the last command you typed (the log is hidden behind the Map). The first time you open the Map it also shows a one-time hint, "Want real icons on your map? Type icons." (see [Icons](#icons-nerd-fonts)).

### The inspect cursor

A cursor picks out one thing at a time: a building, a wonder, a civilization. The cursor is a highlighted cell: a solid block in a color the map itself does not use in your theme (an orange on the dark themes), with the glyph beneath it still showing. The cursor blinks slowly, and holds steady when `motion` is off.

The Map reports the cursor's name, a few details (how many you own, how many workers staff it, what changed) and the whole command to type for it, for example `build hut` or `diplomacy gift riverlands_tribes`. Press `Enter` to put that command into the prompt, edit it if you like, and press `Enter` again to run it. The command depends on what the cursor is on:

| The cursor is on | `Enter` types |
|---|---|
| A building you own that can be upgraded | `upgrade <building>` |
| A building short of workers while some are idle | `assign <building>` |
| Any other building | `build <building>` |
| A wonder | `wonder collect all` while it is rising, otherwise `wonders` |
| A civilization | `diplomacy gift <civ>`, or `diplomacy tribute <civ>` when it is at war with you, or `diplomacy neutral <civ>` under an embargo or a rivalry |
| A trail to a civilization you have a trade route with | `trade route list` |
| The town square | `advance` when your age is ready, otherwise `status` |
| The harbinger | `harbinger` |

An old building with nothing left to do has no command, and `Enter` says so.

### Since your last visit

When you load a save, the Map remembers your town as the save left it, before any time away was played out, and shows what changed since. A news line lists what happened: buildings built, civilizations met, wars, new trade routes and so on. The cursor's details say it too ("+2 since your last visit"), and the map marks what is new.

---

## The town

The Map draws your settlement as a small glyph world. Your town grows out of your real buildings, quarter by quarter, with streets, walls, wonders and people walking the streets. The same save always grows the same town, and new buildings join the existing streets instead of reshuffling them.

<figure class="screen" data-screen="map"><figcaption>The map of a Bronze Age town: the palisade, the quarters inside it and the wonders named on their plots, with what the cursor is on spelled out beneath.</figcaption></figure>

The land comes from your save's seed: hills and forests, lakes and coasts ringed with shallows (`~`) round deeper water (`≈`), and rivers that wind from the high ground down to the sea or a lake. Your town is placed on flat ground near a river.

This age's wonder appears on its plot, under scaffolding, once you start it: bank some of its cost or queue its construction. Until then its plot stays empty. Wonders you have built stand on their plots.

Walls follow your age: none at first, a palisade from the Bronze Age, then a stone wall with towers, and later a ring boulevard. The ring grows in steps as the town outgrows it. Harbor buildings sit on the shore and read as piers or jetties.

### Traffic

Every era brings its own traffic to the town. Nothing appears before the age that introduces it, and older traffic retires as its era passes, so the streets stay readable.

| Era | What moves |
|---|---|
| Stone | Workers walking to work; hunters heading out to the woods and hills |
| Iron | Ox carts and riders on the streets; rowboats on the river and lakes |
| Steel | Wagons; sailing ships. From the Industrial Age a railway crosses the town, with steam trains trailing smoke |
| Electric | Trams and steamships; from the Electric Age, the first motor cars |
| Digital | Cars and trucks, streaming down the highways, planes overhead, container ships and freight trains; satellites cross the night sky; from the Information Age a news helicopter circles the town, and from the Digital Age delivery drones buzz over the streets |
| Neon | Maglevs on the old railway and hovercars; in the Cyberpunk Age sky trains on every elevated line, drone swarms and crowds in the alleys; from the Fusion Age maglevs on the skyways and climbers riding the space elevator's tether |
| Space Age and later | The map has left the ground: see [The sky](#the-sky) |

People stroll. Carts are a little quicker, cars quicker still, and trains and maglevs the quickest. How much moves follows how big and busy your town is. The railway runs along one of the town's street rows, so it never moves as the town grows, and a street it crosses stays a street. Traffic draws on streets, rails, water and open sky, never over a building or a name. The legend lists what is about, and the inspect cursor says what anything moving is ("A steam train, hauling ore.").

Keep an eye on the sky in the late game.

### The city, from the Modern Age to the Fusion Age

From the Modern Age the land round your town becomes a city. Your own buildings still make the town in the middle. The city is everything between and beyond them, and each age draws its own, with its own colors, a signature structure, a signature mover and one effect that moves:

| Age | The city | Signature mover | Moving effect |
|---|---|---|---|
| Modern | A glass belt round the ring boulevard, suburbs of houses on green lawns along asphalt streets, parks still green but fenced into their blocks, and highways out of town. Daylight blues and grays | Cars streaming down the highways | Headlights and tail lights on the highways at night |
| Information | Office blocks over most of the land, server farms, satellite dishes and fiber lines along the streets. About half the green is left, as parks and rooftop gardens, and the first smog gathers at the city's edge | A news helicopter circling the town | Screens flickering in the windows |
| Digital | The last green: one walled reserve, and nothing else. The rest is pale concrete and data halls, and past the city, landfill under the smog | Delivery drones | The data glow, pulsing round the data halls and hacker dens |
| Cyberpunk | The megacity. Every cell is built or derelict: megablocks, megacorp towers with their initials in neon, arcologies, flickering neon signs and holo-ads, steam vents, scrap and toxic ground. The water runs in toxic canals, and three to five elevated sky rails cross it all. Night falls on the map, neon on black | Sky trains, a couple on every rail, with drone swarms, hovercars and crowds in the alleys: the busiest map in the game | Acid rain, and the neon flickering |
| Fusion | The megacity powered: deep blue towers, fusion reactors in their plasma rings, power conduits running in to the town, maglev skyways, launch towers, clean cooling water, and the space elevator's tether rising from the town square to the top of the map | Climbers riding the tether | Plasma pulsing round the reactors and down the conduits |

The green fades by a rule. The Modern Age keeps all its parks and gardens, the Information Age keeps about half, the Digital Age keeps only its reserve (about 15% of the Modern green), and from the Cyberpunk Age there is none at all. It is measured over the settlement view round the town square.

The city is laid from your save's world and the age alone, so the same save always grows the same city, and a new building never moves any of it. Nothing of a later age shows early: the legend lists, and the inspect cursor names, only what your age has built. The city's features list under land in the legend (a busy legend pairs its short labels two to a row so the traffic still fits), and the inspect cursor says what each one is ("Megacorp tower", "Server farm", "Space elevator"). At the region zoom the satellite mosaic shows the city's concrete, green and waste, and from the Cyberpunk Age the night pass shows the megacity as a carpet of lights.

From the Space Age the settlement and district zooms show the sky instead of the town (see [The sky](#the-sky)). The region zoom stays the known world: `PgUp` from the sky climbs to it, and `PgDn` comes back down.

### Three zooms

| Zoom | What you see |
|---|---|
| Region | The whole known world, drawn on your epoch's cartographic plate, with the civilizations you have met |
| Settlement | Your town, one tile per cell |
| District | A closer look, one tile per two cells, with building names |

The district zoom names only a handful of building types at a time (those nearest the cursor first), so it stays readable.

### Map keys

| Key | Action |
|---|---|
| `←` `→` `↑` `↓` | Move the cursor |
| `Shift`+arrow | Move the cursor by 8 |
| `PgUp` or `-` | Zoom out (settlement, then region) |
| `PgDn` or `=` | Zoom in (settlement, then district) |
| `Tab` / `Shift-Tab` | Jump between buildings and wonders (and civilizations at region zoom) |
| `Home` | Center on the town square |

The legend (what each glyph means) shows beside the map when there is room.

---

## The sky

From the Space Age the Map leaves the ground. Each of the last ages has a sky of its own, with its own colors, its own landmark, its own traffic and something always gently moving. It is still your civilization: every building you own becomes something up there, each lineage its own kind of thing (homes are habitat modules in orbit, for instance), and the picture grows as you build. Like the town, it never reshuffles: what is drawn stays where it is. The legend names what you see. The inspect cursor and `Tab` work on all of it (a module tells you its building, its workers and the command for it), the flows overlay marks what is short of hands, and the mini map follows along.

The Space Age is actual space. Your planet's curved limb fills the bottom of the map. Its night side is your own world, land and sea, with cloud swirls drifting over it, amber city lights where your town stands and the civilizations you have met glowing in their colors. The space elevator's tether rises from the town to the station. You see a ring station from above: your buildings dock round the ring, a sector for each lineage, growing outward ring by ring. Solar wings run out along the truss, hulls rise on the shipyard's slips, foundry tanks glow, relay satellites ride a low orbit, mining outposts sit on the rocks of the asteroid belt, and your military holds a base on the moon. Climbers ride the tether, shuttles climb from the town, and satellites and mining drones go about their rounds.

The ages of the Cosmic Era each look like no other, and every one is built from your own buildings. In the first, the home system shrinks to a little orrery while your lineages settle colony worlds and a warp gate rises. The next is something worthy of a starship captain, and the one after is a place where reality bends. At the very end, it is calm. If you have glimpsed a visitor in the sky before, you may find it is not so rare out there.

Each sky has a hub: `Home` takes the cursor there, and `PgDn` gives a closer view with names.

---

## Settings

Two settings shape the Map: the glyph set and the mini map. They are saved per account, like your theme: they carry across saves and new games, and switching accounts swaps them. With no account loaded they last for the session.

| Command | What it does |
|---|---|
| `map glyphs` | Show the current glyph set |
| `map glyphs ascii` | Plain ASCII, for fonts with poor symbol coverage |
| `map glyphs unicode` | Box drawing, blocks and widely supported symbols (the default) |
| `map glyphs nerd` | Nerd Font icons. Needs a Nerd Font in your terminal; every icon has a Unicode fallback |
| `minimap` | Show whether the dashboard's mini map is on |
| `minimap off` | Hide the mini map, so the Buildings list gets the whole column |
| `minimap on` | Show the mini map above the Buildings list (the default) |
| `map flows` | The flows overlay: where your economy is stuck (see [The flows overlay](#the-flows-overlay)). `map flows on` and `map flows off` set it; bare, it switches. It lasts for the session |

All of them work while the Map is open, so you see the change at once. The `motion` setting (`motion on` and `motion off`) holds the maps, the badge case and a theme's ambient effect still when it is off.

### The flows overlay

`map flows` marks where your economy is stuck: stocks that are falling and will run dry within about a minute, buildings short of workers, and idle workers. A flows panel lists each of those, then the five lineages that make most of your output, with a bar for each one's share, so you can see what your settlement actually runs on. The inspect cursor on the town square gives a one-line summary.

---

## The mini map

The mini map is on by default. The dashboard shows a compact view of the Map above the Buildings list, in a border titled "Map". The setting is remembered with your account. It follows your `map glyphs` setting, and its last row is the since-last-visit news.

It is kept short, at most 9 rows inside its border and about a quarter of the column, so the Buildings list keeps most of the room. `minimap off` hides it and gives the list the whole column; `minimap on` brings it back. It marks the tile the Map panel's cursor is on with the same highlighted cell.

The mini map needs room: it appears on terminals of 105 columns by 34 rows and larger. On smaller terminals (80x24, 100x30) it hides and the Buildings list gets the space. Type `map` to open the full panel at any size.

---

## Icons (Nerd Fonts)

The `nerd` glyph set draws real icons, but only if your terminal uses a Nerd Font. The `icons` command opens a window over the dashboard that walks you through it. The window has its own buttons: press a button's letter, or move with `Tab` or the arrow keys and press `Enter`. `Esc` closes it at any step.

1. Check. It shows sample Nerd Font icons next to their Unicode fallbacks. Press "I see icons" (`I`) or "I see boxes" (`B`). If you see icons, it sets `map glyphs nerd` for you.
2. Install. If you see boxes, it offers to install JetBrains Mono Nerd Font for your user: "Install" (`Y`) or "No, show me how" (`N`). Install downloads the official JetBrains Mono zip from the [Nerd Fonts v3.5.1 release](https://github.com/ryanoasis/nerd-fonts/releases) (about 130 MB), checks its SHA-256, and installs the font files of the font's Nerd Font Mono variant for your user only, with no admin rights:
   - macOS: `~/Library/Fonts`
   - Linux: `~/.local/share/fonts` (or `$XDG_DATA_HOME/fonts`), then `fc-cache -f` if it is available
   - Windows: `%LOCALAPPDATA%\Microsoft\Windows\Fonts`, plus per-user font entries in the registry (under HKCU, so no admin rights)

   The install runs in the background with its progress in the window. You can close the window and keep playing: `icons` opens it again, and the log tells you when the install is done. It never runs anything else and never edits your terminal's config files. If the download fails or the checksum does not match, it says so, changes nothing and offers "Try again" (`R`). "No, show me how" points you to [nerdfonts.com/font-downloads](https://www.nerdfonts.com/font-downloads) to install one yourself.
3. Select the font. It detects your terminal (Terminal.app, iTerm2, VS Code, WezTerm, Windows Terminal, kitty, Alacritty, Konsole, GNOME Terminal; generic steps for anything else) and shows the two or three steps to select the font. It gives the font's exact name as your terminal lists it.
4. Restart. Restart your terminal, then run `icons` again to check.

The first time you open the Map it shows a one-time hint: "Want real icons on your map? Type icons." It is remembered per account.

---

## Themes

The Map draws its colors from your active [theme](themes.md), so switching themes recolors the Map and the mini map at once, light themes included.

---

## See also

- [Commands](commands.md#map): the `map` command reference
- [Themes & accessibility](themes.md)
- [Wonders](wonders.md)
- [Trade](trade.md): the routes the Map draws
- [Factions & Diplomacy](factions.md): the civilizations on the Map
