# Civilization History

The Civilization History panel graphs how your civilization has changed over time. Type `history` at the prompt to open it.

---

## What it shows

The game tracks and graphs seven metrics:

| Metric | What it measures |
|---|---|
| **Population** | Total workers alive |
| **Food rate** | Net food per tick (positive = surplus, negative = deficit) |
| **Knowledge rate** | Knowledge production per tick |
| **Faith** | The faith you have stored |
| **Morale** | Civilization morale as a percentage (see [Morale](morale.md)) |
| **All production** | Your all-production bonus as the game applies it, as a percentage: every source added together, held to the cap |
| **Game speed** | Your game speed bonus from techs and events (0.15 means ticks come 15% faster) |

Each metric gets its own braille line graph covering the whole stored history. Beside each graph is the current value, a trend arrow (↑ growing, ↓ shrinking, → stable), and the recorded min/max.

<figure class="screen" data-screen="history"><figcaption>The Civilization History panel in the Bronze Age: population, food rate and knowledge rate over the last three thousand ticks.</figcaption></figure>

**All production** is the bonus that counts: techs, wonders, milestones, monuments, events and boons added together and held to the cap, so it never reads above +200% (see [The all-production cap](resources.md#the-all-production-cap)). The Stats panel's Active Multipliers lists every source. Samples a save recorded before this change show only the milestone and epoch event part.

---

## Reading the graphs

```
Population      ↑ 163 workers  min:10.0    max:198.0
  45.0   ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⠤⠤⠒⠒⠉
         ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⣀⣀⠤⠤⠒⠒⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  25.0   ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣠⠤⠒⠒⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│
         ⣀⣤⠤⠒⠒⠋⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│
  10.0   ────────────────────────────────────────────────
```

- The labels on the left show the max, midpoint and min of the visible window.
- The `│` markers show where you advanced an age. All 7 graphs share the same markers, so you can see how each metric responded to the advance. The ages you passed through are listed under the graphs.
- Time runs left to right, from the oldest sample to now.

---

## How history is collected

The game records one sample every **10 ticks** (about 20 seconds) and keeps the last **300 samples**. That is about the last **100 minutes** of play, not your whole run: on a run that lasts days, the graphs show recent trends. When the store is full, each new sample replaces the oldest one. Age advances are stored as markers and drawn across all graphs. The history is saved with your game, so it survives restarts. The graphs appear once two samples exist.

---

## Tips

- After an age advance, the `│` marker shows which metrics jumped or dipped at the change.
- A **Food rate** line that drops below zero is a food deficit. Catch it before your workers start to starve.
- A flat **All production** line means no milestone or epoch event has added to your permanent all-production bonus recently.
- **Game speed** steps up when you finish a tech that grants game speed (Chronometry, Clockwork Automation, Quantum Computing), and rises for as long as an event that speeds the game lasts. Wonders do not move it.
- A taller terminal window fits all 7 graphs at once.
