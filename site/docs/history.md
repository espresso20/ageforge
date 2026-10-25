# Civilization History

The Civilization History panel graphs how your civilization has changed during the current run. Type `history` at the prompt to open it. The panel opens full screen. Scroll it with the arrow keys, Page Up and Page Down, and close it with Esc.

---

## What it shows

The panel graphs seven metrics:

| Metric | What it measures |
|---|---|
| **Population** | Everyone in your civilization: workers, scholars and soldiers |
| **Food rate** | Net food per tick (positive = surplus, negative = deficit) |
| **Research rate** | The work your scholars do per tick |
| **Faith** | The faith you hold |
| **Morale** | Civilization morale as a percentage (see [Morale](morale.md)) |
| **All production** | Your all-production bonus as the game applies it, as a percentage: every source added together, in full |
| **Game speed** | Your game speed bonus (0.15 means ticks come 15% faster) |

Each metric gets its own braille line graph that covers the whole stored history. Above each graph are the metric's name, a trend arrow, its current value, and the lowest and highest values in the window. The arrow compares the newest sample with the one before it: ↑ means it grew by more than 2%, ↓ means it fell by more than 2%, and → means it stayed within 2%. The line above the first graph gives the tick range and the number of samples.

<figure class="screen" data-screen="history"><figcaption>The Civilization History panel in the Bronze Age: population, food rate and research rate over the last three thousand ticks.</figcaption></figure>

**All production** adds up every source: wonders, milestones, monuments, events and boons, in full, however large the pool grows (see [The all-production pool](resources.md#the-all-production-pool)). A pool that has earned +320% reads +320% here, and the line keeps rising as you earn more. The Stats panel's Active multipliers section lists every source.

**Game speed** comes from three techs (Chronometry adds 0.05, Clockwork Automation adds 0.10 and Quantum Computing adds 0.15) and from timed boosts, such as the one a completed milestone chain grants or a faction boon that speeds the game. A timed boost lifts the line for as long as it lasts. Wonders do not move it.

---

## Reading the graphs

```
Population      ↑ 198 workers  min:10       max:198
  198    ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⣀⣀⡤⠤⠖⠚⠉
         ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣤⠤⠶⠒⠚⠉⠉⠀⠀⠀⠀⠀⠀
  104    ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⣀⣀⣀⡤⠤⠤⠖⠒⠚⠉⠉⠁⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
         ⣀⣀⣀⣀⣀⣀⣀⣀⣀⡤⠤⠤⠤⠤⠤⠴⠒⠒⠒⠒⠛⠉⠉⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  10     ────────────────────────────────────────────────────
```

- The labels on the left show the highest value, the midpoint and the lowest value in the window. A label that would only repeat the one under it is left out, so a flat line carries its value once, on the base line.
- The `│` markers show where you advanced an age. All 7 graphs share the same markers, so you can see how each metric responded to the advance. The ages you advanced into are listed under the graphs, by name.
- Time runs left to right, from the oldest sample to now.
- Values use short forms, such as 950, 12.5K or 1.23M, like the rest of the game.

---

## How history is collected

The game records one sample every 10 ticks and keeps the last 300 samples, which is the last 3,000 ticks of the run. A tick takes 2 seconds before any game speed bonus, so at the base speed the graphs cover the last 100 minutes. They show recent trends, not the whole run. When the history is full, each new sample replaces the oldest one. Age advances are stored as markers and drawn across all graphs. Markers from before the window are dropped the next time you advance.

Time you spend away from the game is sampled the same way, so the samples stay evenly spaced. The history is saved with your game, so it survives restarts. It starts over with each run: when you prestige, when you succumb to a catastrophe, and when you start a new game. The graphs appear once two samples exist. Until then the panel says there is no history yet.

---

## Tips

- After an age advance, the `│` marker shows which metrics jumped or dipped at the change.
- A **Food rate** line that drops below zero is a food deficit. Catch it before your workers start to starve.
- A flat **All production** line means nothing new has joined the pool during the window.
- A taller terminal window fits more graphs on the screen at once. Scroll to reach the rest.
