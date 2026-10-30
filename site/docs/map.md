# The Map

Open it with the `map` command. The Map draws your empire from your real game state: the buildings you own, the workers staffing them and your wonders. Build something new and the map changes to match.

The Map is a glyph world seen from above. Every cell is one colored character, and the land, the roads, the buildings and the light all come from your save.

The Map shows only what you have reached. It names no civilization you have not met, no era you have not reached and no age past your next one.

`citymap` and `worldmap` are aliases of `map`. Both open the Map, and `worldmap` opens it on the whole town (the region zoom).

---

## The Map panel

`map` opens the Map over the whole screen except the command bar, which keeps working: type any command while the Map is open, just as you would with it closed, and watch the map change. A command that opens another panel (say `research`) swaps the Map for it.

Because typing goes to the command bar, the Map takes only the keys that print nothing. These work:

| Key | Action |
|---|---|
| Arrows | Move the cursor |
| `Tab` / `Shift-Tab` | Jump to the next or previous building, wonder or civilization (to the next or previous district, in the district zoom) |
| `PgUp` / `PgDn` (`Option+↑` / `Option+↓`), or `-` / `=` | Zoom out and in a step: the whole town, the close-up, a district |
| `Home` | Center on the town square |
| `Enter` | Put the command for what the cursor is on into the prompt. It is not run: press `Enter` again to run it |
| `Esc` | Close the Map |

Once you have typed something, `Tab`, `Shift-Tab` and `Enter` act on the prompt instead (complete the command, run it), as they do everywhere else. The arrows, `PgUp`, `PgDn`, `Home` and `End` stay with the Map.

A few keys that print are the Map's too, while the prompt is empty (no command begins with one): `+` and `=` zoom in, `-` and `_` zoom out. A keyboard with no Page keys (a Mac's, for one) can zoom with `-` and `=` alone. `Option+↑` and `Option+↓` also work as `PgUp` and `PgDn` here, and `Option+←` and `Option+→` as `Home` and `End` (see [Commands](commands.md#the-prompt)).

The settings are commands you can type with the Map open: `map glyphs`, `map flows` and `minimap` (see [Settings](#settings)). With VibeMode on, the words on the Map come in its second script too (see [VibeMode](commands.md#vibemode)).

The top row names your town, its age and era, the weather and the time of day, your people, your idle workers, how many civilizations you have met and the zoom you are in. The rows at the foot hold the news line, what the cursor is on, the command to type for it and the keys. Below them is a key bar. It shows what `Enter` will do, the current glyph set, and the reply to the last command you typed (the log is hidden behind the Map). The first time you open the Map it also shows a one-time hint, "Want real icons on your map? Type icons." (see [Icons](#icons-nerd-fonts)).

### The inspect cursor

A cursor picks out one thing at a time: a building, a wonder, a road, the town square or the open ground. The cursor is a highlighted cell: a solid block in a color the map does not use in your theme (an orange on the dark themes), with the glyph beneath it still showing. Over a ground that color would melt into, it takes another that stands out. The cursor blinks slowly, and holds steady when `motion` is off.

The Map reports the cursor's name, a few details (how many you own, how many workers staff it, what it makes, what changed) and the whole command to type for it, for example `build hut`. Press `Enter` to put that command into the prompt, edit it if you like, and press `Enter` again to run it. The command depends on what the cursor is on:

| The cursor is on | `Enter` types |
|---|---|
| A building you own that can be upgraded | `upgrade <building>` |
| A building short of workers while some are idle | `assign <building>` |
| Any other building | `build <building>` |
| A wonder | `wonder collect all` while it is rising, otherwise `wonders` |
| The town square | `advance` when your age is ready, otherwise `status` |

An old building with nothing left to do has no command, and `Enter` says so. A road or open ground has none either: the Map names it and says what it is.

Everything that moves is named for what it is: the worker, the ox cart, the wagon, the car, the hovercar, the shuttle, the mining drone, the generation ship, the starship. In the space ages the cursor names the homeworld, the home sun, the other planets, the stars, the elevator's tether, the rings and orbits and the routes between stars, the empty sky, and the heart of the town as a ring station, a home system, a starbase, an anchor or a core. A building is a part of the scene it stands in: a habitat module in orbit, a colony ark among the planets, a deck on the starbase. The spaceport on the ground is named too.

### Since your last visit

When you load a save, the Map remembers your town as the save left it, before any time away was played out, and shows what changed since. A news line lists what happened: buildings built, civilizations met, wars, new trade routes and so on. The cursor's details say it too ("+2 since your last visit"). When there is no news the line says what is stuck, or that all is quiet.

The ground under every kind of building that grew since your last visit, and under a wonder you finished, is lit for a while when the Map first shows it, and clears once you have looked at the Map. The lighting changes only the ground, never a glyph. A later visit with new news lights again.

---

## The town

The Map builds your town from the land up. The land comes from your save's seed: sea and shallows, shores, grassland, forest, hills and mountains, and rivers that wind from the high ground down to the sea. The square sits on flat dry ground near the water, and the roads grow out from it: a main road first, then branches. Where a road meets a river it bridges it.

<figure class="screen" data-screen="map"><figcaption>The map of a Bronze Age town in the close-up: the streets that grow from the square, the quarters along them and the legend at the right, with what the cursor is on spelled out beneath.</figcaption></figure>

Every building stands on a plot beside a road. Each kind of building has its own part of town, so the quarters read apart, and the legend names the sign each kind wears. A building keeps its plot for good. A new building takes a plot that is free and never moves one that is there, so the same save always grows the same town. As you own more of a kind, its plot grows with it, from a hut to a house to a block to a tower, in the materials of the age. A building lost to ruin leaves rubble on its plot.

This age's wonder appears on its own plot, rising from the ground under scaffolding, once you start it: bank some of its cost or queue its construction. Until then its plot stays empty. Wonders you have built stand finished, each one a small hand-drawn shape of its own, dressed in the materials of its age.

Around the buildings the land is dressed too: trees and rocks, crops, chimney smoke, and people and traffic on the roads. Nothing moves when `motion` is off. At night the light fades and the windows and lamps glow.

### The world around the town

The Map draws what the header and the news report. Each civilization you have met stands at the edge of the picture, in a direction of its own, marked in a color for its opinion of you: allied, friendly, neutral, rival or at war. A trade route is a line of dots from the square toward its partner, with caravans going back and forth on it. A route that is disrupted has one caravan stalled on it.

When scouts are out, a small party walks out along a road and back. An army that is out marches toward the civilization that likes you least. The harbinger, while there is one, stands by the square, dressed in the colors of the age, and idle workers loiter round the square: one for every six who have nothing to do, up to nine. Nothing here covers a building.

From the Space Age the same things appear as stars at the edge of the picture, with lanes to them and ships on the lanes. On the star map the civilizations live at the systems themselves.

Put the cursor on any of them and the Map names it: the civilization with its opinion of you, its strength and its trades, the trail or lane to it with what its caravans bring, the harbinger with its warning, the party that is out, or the idle workers with the command to put them to work. `Tab` goes on from the buildings and wonders to the civilizations at the edge.

### Dinosaurs and the asteroid

In the Primitive Age the Late Cretaceous is still on: now and then a long-necked sauropod plods across open ground near the town, a small raptor dashes past, or a pterodactyl flaps over the upper rows. They are only scenery. They keep off buildings, can't be clicked, and thin out as you close on the Stone Age, until none are left. Near the end of the Late Cretaceous a fireball streaks down across the sky every few seconds for a short stretch of the age. With `motion` off, whatever is on screen holds still.

### The rare visitor

Now and then someone comes to the Map from somewhere else. On the land a small visitor walks in along one of the trunk roads from beyond the edge of the view, stops a few cells from the square and is gone. In orbit and the solar system it is a ship, flying in from the edge of the map to one of the lanes. Rarer still, an alien's craft crosses over the town, or drops in over open ground, hangs there and shoots off.

A visitor never covers a building, and it shows in the whole-town view too. It is rare before the Space Age and more common from it; from the Galactic Age on there are none, because the aliens are ordinary traffic by then. It comes and goes on the world's clock, so with `motion` off it still arrives, and stands still while it is there. Put the cursor on it and the Map says something of its own for a visitor and for an alien's craft. Looking at one earns a secret badge.

### The look of each era

Each era of the game has its own look, and every age in it sets itself apart by its roads, its materials, its tallest buildings and its light:

| Era | What the town is made of |
|---|---|
| Stone | Huts of thatch and mud, dirt tracks and hearths |
| Iron | Stone and timber, tile roofs, cobbled streets, towers and walls |
| Steel | Brick and slate, domes, docks and the first great chimneys, with a railway from the Industrial Age |
| Electric | Concrete, wire and light, with lamps along the streets |
| Digital | Glass and asphalt, screens and the last parks |
| Neon | Dark towers hung with light, rain on black streets and, at the edge of the city, a spaceport |

People walk the streets of every era. Carts roll through the first three eras, cars arrive with the Electric Era, and the Neon Era has hovercars. How much moves follows how big and busy your town is, and traffic keeps to the roads.

### Three zooms

| Zoom | What you see |
|---|---|
| Region | The whole town squeezed into the space: each cell stands for a block of the world and shows the most important thing in it, a building as the sign of its kind and a wonder as a star |
| Settlement | The close-up: your town round the cursor, one cell of the world to a cell of the screen |
| District | The close-up framed on one kind of building's part of town, with every other kind dimmed so the district stands out. The top row names it |

A large terminal opens the Map on the close-up. A smaller one, and the mini map, show the whole town. `PgDn` or `=` zooms in a step, from the region to the settlement to a district, and `PgUp` or `-` zooms out a step. In a district, `Tab` and `Shift-Tab` go on to the next or previous district, and the cursor lands on a building in it. The close-up scrolls only when the cursor nears its edge. In the region zoom the arrows move the cursor a block at a time, and the cursor still tells you what it is on.

<figure class="screen" data-screen="map-region"><figcaption>The same town in the region zoom: the whole town in one view, each building the sign of its kind.</figcaption></figure>

### Map keys

| Key | Action |
|---|---|
| `←` `→` `↑` `↓` | Move the cursor |
| `Shift`+arrow | Move the cursor by 8 |
| `PgUp` or `-` | Zoom out a step |
| `PgDn` or `=` | Zoom in a step |
| `Tab` / `Shift-Tab` | Jump between buildings, wonders and civilizations (between districts, in the district zoom) |
| `Home` | Center on the town square |

The legend (the sign of each kind of building your town has) shows beside the map when there is room.

---

## The sky

From the Space Age the Map leaves the ground, and each step out pulls the view further back. It is still your civilization: every building you own becomes something up there, and the picture grows as you build. Like the town, it never reshuffles: what is drawn stays where it is. The inspect cursor and `Tab` work on all of it, the flows overlay lists what is short of hands, and the mini map follows along.

The Space Age looks down on your planet from orbit. The planet is a small disc in the middle, its land and sea the world your town stood on, with clouds drifting over it and the lights of its cities by night. Round it run the rings of orbit, and your buildings stand on them as stations, growing outward ring by ring. A space elevator joins the ground to the rings, and shuttles and freighters go about their rounds.

The Interstellar Age pulls back to the solar system. The sun sits in the middle with the planets on their orbits and the lanes between them. Your buildings are marks along the orbits, in the order your people reached them, and ships travel the lanes.

The last three ages, the Galactic, the Quantum and the Transcendent, pull back to a spiral galaxy of arms, dust and a bright core, different for every save. Your civilization is territory on it: each kind of building has its own sector along your home arm, a hub with a cluster of systems that grows as you build, and a faint border runs round the lot. Lanes join the hubs, thicker where more traffic runs, and the civilizations you have met hold clusters of their own on the other arms, in the color of how they stand toward you. A harbinger opens a red rift at the edge. Your wonders stand as bright stars beside the systems.

The sky is alive. Stars twinkle, a comet crosses every so often, a star goes nova now and then, the dust shimmers in the whole-galaxy view, and the hubs and the core slowly breathe. The Quantum Age adds a few ghost systems that flicker in and out at the edge of your territory, and the Transcendent Age rings your home sectors with a halo of small circles. With `motion` off the sky is held still, and the picture is whole.

`Home` takes the cursor to the middle of the picture, and `PgUp` and `PgDn` switch between the whole picture and a closer view.

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

`map flows` shows where your economy is stuck: stocks that are falling and about to run dry, buildings short of workers, and idle workers. A flows panel lists each of those, then the five lineages that make most of your output, with a bar for each one's share, so you can see what your settlement actually runs on. On a terminal too narrow for a side panel the list opens in a box over the map. The inspect cursor on the town square gives a one-line summary.

---

## The mini map

The mini map is on by default. The dashboard shows a compact view of the Map above the Buildings list, in a border titled "Map". The setting is remembered with your account. It follows your `map glyphs` setting, and its last row is the since-last-visit news.

It is kept short, at most 9 rows inside its border and about a quarter of the column, so the Buildings list keeps most of the room. `minimap off` hides it and gives the list the whole column; `minimap on` brings it back. It shows the whole town, and it marks the cell the Map panel's cursor is on with the same highlighted cell.

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

The Map draws its colors from your active [theme](themes.md), so switching themes recolors the Map and the mini map at once, light themes included. In the one-ink themes, Monochrome and Parchment, it is drawn in one ink on one paper.

---

## See also

- [Commands](commands.md#map): the `map` command reference
- [Themes & accessibility](themes.md)
- [Wonders](wonders.md)
- [Trade](trade.md): the routes that shape your economy
- [Factions & Diplomacy](factions.md): the civilizations you meet
