# The Advance Panel

The Advance panel is one screen for the question "what is left before the age turns?" It shows the march of ages with a calendar year, the four pillars the next age stands on, and what the next age brings. Type `charter` to open it and `Esc` to close it. Opening it, moving about it and closing it change nothing in your game.

<figure class="screen" data-screen="advance"><figcaption>The Advance panel in the Bronze Age, partway: the march of ages with the year, the four pillars with the Knowledge pillar picked, and what the Iron Age brings.</figcaption></figure>

It is a view of the same gate the `advance` command checks: the buildings the next age lists and the wonder of the age you are in. The panel adds no requirement and takes none away. `advance` works from the prompt too, and when the gate is not met it points you at `charter`.

The top of the dashboard shows the same figures at a glance: a long bar to the next age with the percent, the four pillars as small meters, and a line naming what is still missing (see [The screen layout](how-to-play.md#the-screen-layout)). On a small terminal the top keeps the long bar and the line, and the pillars appear as the terminal grows. All of them are read from the same gate, so the percents match.

---

## The march of ages

Every age is a stop on a track across the top. The ages you have passed are filled in, the one you are in and the next one are lit, and the ones beyond the next stay unnamed (`?`) until you reach them, as everywhere else in the game. On a narrow screen the track scrolls, and a `◂` or `▸` at its end shows there are more ages that way. Between your age and the next, a marker moves along the track as the pillars come in. The line under the track reads like this:

```
Year c. 2,300 BC · 49% of the way to the Iron Age
```

The year is the calendar year of your town's history, rounded: Primitive begins 100 million years ago, in the Late Cretaceous, and the deep past shows a geologic epoch instead of a number (Paleocene, Eocene and on to the Pleistocene). Closer years read as "c. 40,000 BC" or "c. 3,300 BC", then "c. 1450 AD", then a decade ("1840s"), and from 2100 a century ("22nd century"). It starts at the year your age begins and moves toward the year of the next age as the four pillars fill, so the track and the year close on the next age together. While you play an age the year never runs backward: if you sell a building, the pillars show the loss, but the year you reached stays. Loading a game or starting a new age sets it from where the pillars stand. The last age has no number. It shows the words "Beyond time" in place of a year.

The year says where your town stands in history. It is not a clock, and it says nothing about how long play takes.

## The Charter

The Charter is the four pillars the next age asks for. Each has a bar and a percent, or `✓ stands` once everything in it is met.

| Pillar | What it holds |
|---|---|
| Knowledge | The age's keystone tech, and the research buildings the next age asks for |
| People | The homes, the food buildings and the faith buildings the next age asks for |
| Industry | Every other building the next age asks for: producers, trade, military and storage |
| Monument | The age's wonder: its bank filled, and the wonder built |

A pillar's bar is the average of its rows. A pillar with nothing to ask in some age reads `nothing needed this age` and counts as met. `↑` and `↓` pick a pillar, and the box under the four lists what it asks for: a building as `7 of 10`, the keystone as researched or not, the wonder's bank as a percent, the wonder as built or not. A row that is met is green. When a pillar has more rows than the screen has room for, the ones still to do come first. Under the rows, Knowledge says how many scholars are at work, and People gives your workers, your reserve and your homes.

The wonder's bank counts each resource at what it is worth in the age's prices, so a resource named in huge amounts does not outweigh a dear one. A surplus of one resource does not pay for another. The Waiting on line at the top of the dashboard, the [Wonders](wonders.md) panel and the [map](map.md) show the same banked share. The wonder needs its [keystone tech](ages.md#keystone-techs) before it can be built, and the pillars show it: the keystone leads Knowledge, and the Monument's last row says what the wonder is waiting for.

## What the next age brings

The box on the right lists what the next age adds: how many buildings it unlocks, the resources it unlocks, the names it gives the four kinds of person in your town, and its wonder.

## In the way

Under it, a box lists what else stands between you and the advance, apart from the pillars:

- A pending catastrophe stops the advance until you choose. The line says so and names the command.
- A harbinger's warning shows as a notice when it foretells a doom for this age, or when the advance would leave the era with a doom still open. An advance meets that doom first. `harbinger` answers it. See [The Harbinger](harbinger.md).
- In the final age there is no next age, and the panel says so.

Without any of these the box reads `Nothing else stands in the way.`

## Advancing

When all four pillars stand and nothing else is in the way, the foot of the panel reads `Enter  Advance to the Iron Age`. `Enter` does what `advance` does, with the same refusals, and then the arrival screen opens. While you have text at the prompt, the keys belong to the prompt. `Enter` before all four stand names the pillars still to do and changes nothing.

| Key | What it does |
|---|---|
| `↑` / `↓` | Pick a pillar |
| `Enter` | Advance, when all four pillars stand |
| `Esc` | Close the panel |

Like the other panels, it reads in every theme and on the plain glyph set, and it fits an 80 by 24 terminal.

---

## The measured gate (experimental)

A second gate rule is under trial behind a switch. It is off unless you turn it on, and nothing in your save depends on it. Start the game with the variable set:

```
AGEFORGE_GATES=v2 ./ageforge
```

On Windows, set it first (`set AGEFORGE_GATES=v2` in Command Prompt, `$env:AGEFORGE_GATES = "v2"` in PowerShell) and then start the game. The game reads it once, when it starts, and any value other than `v2` leaves the gate described above on. Start the game without it and the gate is the one described above, with the same save.

The four pillars stay, and so does everything around them: the panel, the `advance` command, the plan's `advance` item, the notices and the refusals all follow whichever gate the game started with. Three of the pillars count different things. The buildings the next age lists are not asked for, and the wonder is asked for as before.

| Pillar | What it asks under the measured gate |
|---|---|
| Knowledge | The age's keystone tech, and one more tech of the age, the key tech |
| People | The population of your town: everyone who lives in it (workers, the reserve, scholars and soldiers), in whatever mix of homes |
| Industry | The income a tick of the age's main materials together, from whatever mix of producers |
| Monument | The wonder's bank filled, and the wonder built, as before |

In the Iron Age the rows read like this: `Population 412 / 669`, `Stone + Iron income 210 / 539 a tick`, `Keystone: Mathematics` and `Key tech: Iron Smelting`.

The key tech is the tech of the age, besides the keystone and the techs the keystone stands on, that most of the age's buildings need. When no building needs one, it is the tech most other techs are built on. The main materials, two at most, are the ones that the age's producers make and the age's building prices use most.

The numbers are not set by hand. They come from a model of an ordinary town: for each age, its population and its income of the main materials at the moment it meets the building list. Any mix of homes and producers that reaches the numbers counts.

This is a trial. The numbers, the choice of materials and key techs, and the rule itself may change or be dropped.
