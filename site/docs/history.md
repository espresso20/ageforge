# Civilization History

The Civilization History panel graphs how your civilization has changed over time. Type `history` at the prompt to open it.

---

## What it shows

The game tracks and graphs seven metrics:

| Metric | What it measures |
|---|---|
| **Population** | Total workers alive |
| **Food Rate** | Net food per tick (positive = surplus, negative = deficit) |
| **Knowledge** | Knowledge production per tick |
| **Faith** | Total faith accumulated |
| **Morale** | Civilization morale as a percentage (the production multiplier's input) |
| **Prod Bonus** | Your permanent bonus to all production, as a percentage |
| **Tick Speed** | Your tick speed bonus from techs, prestige and events |

Each metric gets its own braille line graph covering the whole stored history. Beside each graph is the current value, a trend arrow (↑ growing, ↓ shrinking, → stable), and the recorded min/max.

---

## Reading the graphs

```
Population     ↑ 163workers  min:10.0    max:198.0
45.0  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⠤⠤⠒⠒⠉
      ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│⠀⠀⠀⠀⠀⠀⣀⣀⠤⠤⠒⠒⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
25.0  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣠⠤⠒⠒⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│
      ⣀⣤⠤⠒⠒⠋⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀│
10.0  ────────────────────────────────────────────────
```

- The labels on the left show the max, midpoint and min of the visible window.
- The `│` markers show where you advanced an age. All 7 graphs share the same markers, so you can see how each metric responded to the advance.
- Time runs left to right, from the oldest sample to now.

---

## How history is collected

The game records one sample every **10 ticks** (about 20 seconds at 1x) and keeps the last **300 samples**, about 100 minutes of history at 1x. When the store is full, each new sample replaces the oldest one. Age advances are stored as markers and drawn across all graphs. The history is saved with your game, so it survives restarts. The graphs appear once two samples exist.

---

## Tips

- After an age advance, the `│` marker shows which metrics jumped or dipped at the change.
- A **Food Rate** line that drops below zero is a food deficit. Catch it before your workers start to starve.
- A flat **Prod Bonus** line means no milestone or prestige upgrade has added to your all-production bonus recently.
- **Tick Speed** steps up when you finish a tech that grants tick speed. Wonders do not move it.
- A taller terminal window fits all 7 graphs at once.
