# History

The History panel is the game's metrics page. It draws one chart of your current run, can lay an earlier run behind it, and lists every run you have finished on the account. Type `history` at the prompt to open it. Close it with Esc.

---

## Ranges

Three tabs set how much of the record the chart covers. Tab cycles through them.

| Range | What it covers |
|---|---|
| **This age** | From the last age advance to now |
| **This run** | The whole current run, from its start |
| **All runs** | The current run against every finished run on the account, each lined up from its own start |

Ages appear as tinted bands behind the line and as a color ribbon along the edge. Marks for age advances, wonders, catastrophes and harbingers sit on a sign row above the chart, and each age is named inside its own band when it is wide enough. A head marker shows the live end of the current run.

<figure class="screen" data-screen="history"><figcaption>The History panel on This run.</figcaption></figure>

---

## Measures

The picker on the left lists the measures in three groups: People, Progress and Rates. Up to three measures can be shown at once, each in its own slot with its own color and shape (circle, square, triangle). Color and shape belong to the slot, so the first measure shown is always the first slot's color.

The ten core measures carry the keys 1 to 0. Pressing a key fills a free slot, or clears the measure if it is already shown. Resource rates have no key: move to one with the Up and Down keys and press Space. Rates are shown per second.

Finished runs keep only the ten core measures. In All runs, the resource rates are shown as unavailable.

Each line is drawn on its own scale. Under the chart, the legend states each measure's true range, such as "log, 1 to 300K workers", or says that the measure stays linear.

---

## Linear and Log

The chart can use a linear or a log scale. Log is the default on All runs, where the first hours of a run would otherwise flatten against a run that lasted days. Linear is the default on the other ranges. The choice is kept for each range while the panel is open. Measures with a fixed cap stay linear.

---

## The pointer

A vertical pointer reads the chart. Move it with the left and right arrow keys, or with Home and End to jump to the start and the live end. Moving the mouse over the chart moves the pointer too. The readout box shows the run time, the age and one row per shown measure. In All runs it has two columns, the current run and the ghost.

---

## The ghost run

In All runs, a finished run can be drawn behind the current one as a dotted line with a hollow marker at the pointer, faint age ticks along the bottom edge and a dotted sample in the legend. Move to a run with the Up and Down keys and press Enter, or click its row. Picking it again clears it.

---

## The runs list

Under the chart in All runs, the list shows the run in play pinned at the top, then every finished run, newest first. Scroll it with the wheel or the Up and Down keys.

| Column | Meaning |
|---|---|
| **Run** | The save name |
| **End** | How the run ended: prestige, succumbed or left |
| **Highest age** | The age reached, in its color |
| **Length** | How long the run lasted |
| **Buildings, Techs, Expeditions, Workers** | Final counts. Hidden when the window is narrower than 120 columns |
| **Pace** | A strip of the run's ages, drawn to one shared scale, so it also reads as a length bar |

A crown marks the best run: the highest age reached, the shortest among equals. A run with no kept history is shown crossed out with an empty strip and cannot be picked. With no finished runs the list shows a placeholder line.

---

## What is recorded

The game keeps the whole current run at a fixed size: recent samples at full detail and older ones thinned. When a run ends it is filed on the account with the ten core measures and the age, wonder, catastrophe and harbinger marks. Resource rates, away spans, epoch events and game speed per sample are not kept for finished runs, and wonder times are not available for runs filed earlier.

---

## Keys and mouse

| Key | Action |
|---|---|
| Left / Right, Home / End | Move the pointer |
| 1 to 0 | Show or clear a core measure |
| Up / Down | Move through the picker, or through the runs list in All runs |
| Space | Show or clear the picked measure (This age and This run) |
| Tab | Next range |
| `-` | Linear or log scale |
| Enter | Pick or clear the ghost run |
| Esc | Close |

With the mouse: hover moves the pointer, a click on a tab, a measure, the scale button or a run row acts on it, and the wheel scrolls the runs list.
