# The Advance Panel

The Advance panel is one screen for the question "what is left before the age turns?" It shows the march of ages with a calendar year, the four pillars the next age stands on, and what the next age brings. Type `charter` to open it and `Esc` to close it. Opening it, moving about it and closing it change nothing in your game.

<figure class="screen" data-screen="advance"><figcaption>The Advance panel in the Bronze Age, partway: the march of ages with the year, the four pillars with the Knowledge pillar picked, and what the Iron Age brings.</figcaption></figure>

It is a view of the same gate the `advance` command checks: the buildings the next age lists and the wonder of the age you are in. The panel adds no requirement and takes none away. `advance` still works from the prompt, and when the gate is not met it points you at `charter`.

The top of the dashboard shows the same figures at a glance: a long bar to the next age with the percent, the four pillars as small meters, and a line naming what is still missing (see [The screen layout](how-to-play.md#the-screen-layout)). They are read from the same gate, so the percents match.

---

## The march of ages

Every age is a stop on a track across the top. The ages you have passed are filled in, the one you are in and the next one are lit, and the ones beyond the next stay unnamed until you reach them, as everywhere else in the game. Between your age and the next, a marker moves along the track as the pillars come in. The line under the track reads like this:

```
Year 3,300 BC · 43% of the way to the Iron Age
```

The year is the calendar year of your town's history. It starts at the year your age begins and moves toward the year of the next age as the four pillars fill, so the track and the year close on the next age together. It never runs backward within an age: if you sell a building, the pillars show the loss, but the year you reached stays. The last age has no year, and shows a word in its place.

The year says where your town stands in history. It is not a clock, and it says nothing about how long play takes.

## The Charter

The Charter is the four pillars the next age asks for, each with a bar and a percent, or `✓ stands` once everything in it is met.

| Pillar | What it holds |
|---|---|
| **Knowledge** | The age's keystone tech, and the research buildings the next age asks for |
| **People** | The homes, the food buildings and the faith buildings the next age asks for |
| **Industry** | Every other building the next age asks for: producers, trade, military and storage |
| **Monument** | The age's wonder: its bank filled, and the wonder built |

A pillar's bar is the average of its rows. A pillar with nothing to ask in some age reads `nothing needed this age` and counts as met. `↑` and `↓` pick a pillar, and the box under the four lists what it asks for: a building as `7 of 10`, the keystone as researched or not, the wonder's bank as a percent, the wonder as built or not. A row that is met is green. When a pillar has more rows than the screen has room for, the ones still to do come first.

The wonder's bank counts each resource at what it is worth in the age's prices, so a resource named in huge amounts does not outweigh a dear one. The Waiting on line at the top of the dashboard, the [Wonders](wonders.md) panel and the [map](map.md) show the same banked share. The wonder needs its [keystone tech](ages.md#keystone-techs) before it can be built, and the pillars show it: the keystone leads Knowledge, and the Monument's last row says what the wonder is waiting for.

## What the next age brings

The box on the right lists what the next age adds: how many new buildings, its new resources, the names it gives the four kinds of person in your town, and its wonder.

## In the way

Under it, a box lists what else stands between you and the advance, apart from the pillars:

- A pending **catastrophe** stops the advance until you choose. The line says so and names the command.
- A **harbinger**'s word is a notice: an advance that would carry you past a doom it foretold meets it first. `harbinger` answers it. See [The Harbinger](harbinger.md).
- In the final age there is no next age, and the panel says so.

## Advancing

When all four pillars stand and nothing else is in the way, the foot of the panel reads `Enter  Advance to the Iron Age`. `Enter` does what `advance` does, with the same refusals, and then the arrival opens. While you have text at the prompt, the keys belong to the prompt. `Enter` before all four stand says which pillars are still to do and changes nothing.

| Key | What it does |
|---|---|
| `↑` / `↓` | Pick a pillar |
| `Enter` | Advance, when all four pillars stand |
| `Esc` | Close the panel |

Like the other panels, it reads in every theme and on the plain glyph set, and it fits an 80 by 24 terminal.

---

## Trying the new gate (experimental)

There is a second way for the gate to ask, kept behind a switch while it is tried. It is off unless you turn it on, and nothing in your save depends on it. Start the game with the variable set:

```
AGEFORGE_GATES=v2 ./ageforge
```

On Windows, set it first (`set AGEFORGE_GATES=v2` in Command Prompt, `$env:AGEFORGE_GATES = "v2"` in PowerShell) and then start the game. It is read once, when the game starts. Start the game without it and the gate is the one described above again, with the same save.

The four pillars stay, and so does everything around them: the panel, the `advance` command, the plan's `advance` item, the notices and the refusals all follow whichever gate the game started with. What changes is what three of the pillars count. The buildings the next age lists are no longer asked for. The wonder is asked for as before.

| Pillar | What it asks under the new gate |
|---|---|
| **Knowledge** | The age's keystone tech, and one more key tech of the age |
| **People** | The population of your town, whatever mix of homes it lives in |
| **Industry** | The income a tick of the age's main materials together, whatever mix of producers makes it |
| **Monument** | Unchanged: the wonder's bank filled, and the wonder built |

The rows read like this: `Population 412 / 500`, `Iron + Coal income 210 / 300 a tick`, `Keystone: Calendar` and `Key tech: Military Tactics`.

The key tech is the tech of the age, besides the keystone and the techs the keystone stands on, that most of the age's buildings need. When no building needs one, it is the tech most other techs are built on. The main materials, two at most, are the ones that the age's producers make and the age's building prices use most.

The numbers are not set by hand. For each age they are what the town the game is balanced around has, at the moment it meets the building list: its population, and its income of the main materials. A town built the ordinary way meets them as it meets the list, and a town that reaches them some other way advances just as well.

This is a trial. The numbers, the choice of materials and key techs, and the rule itself may change or be dropped.
