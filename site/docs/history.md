# Civilization History

The Civilization History overlay gives you a live graphical view of how your civilization has grown and changed over time. Open it at any point by typing `history` at the command prompt.

---

## What it shows

Six metrics are tracked and graphed continuously:

| Metric | What it measures |
|---|---|
| **Population** | Total workers alive |
| **Food Rate** | Net food per tick (positive = surplus, negative = deficit) |
| **Knowledge Rate** | Knowledge production per tick |
| **Faith** | Total faith accumulated |
| **Prod Bonus** | Your current `production_all` permanent bonus percentage |
| **Tick Speed** | Current tick speed multiplier |

Each metric gets its own braille line graph showing the full rolling history. Beside each graph you'll see the current value, a trend arrow (↑ growing, ↓ shrinking, → stable), and the recorded min/max.

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

- **Y-axis labels** on the left show max, midpoint, and min values for the visible window
- **`│` vertical markers** show where age advances occurred — all 6 graphs share the same markers so you can see how each metric responded to an age transition
- **X-axis** represents cumulative ticks from oldest to newest sample (left = oldest, right = now)

---

## How history is collected

- One sample is recorded every **10 ticks** (~13–20 seconds of real time depending on speed)
- Up to **300 samples** are stored — roughly 1 hour of rolling history at 1.5× speed
- When the buffer is full, the oldest sample is dropped to make room for the new one
- Age advance events are stored as markers and displayed across all graphs
- History is **saved and restored** automatically with your game save — it survives restarts

---

## Tips

- **Check history after an age advance** — the `│` marker makes it easy to see which metrics spiked or dipped at the transition
- **Food Rate dipping below zero** shows up clearly as the line crossing the midpoint — useful for catching worker starvation before it becomes critical
- **Prod Bonus flat-lining** means no new milestones or prestige upgrades have fired recently
- **Tick Speed** shows the effect of wonders — each wonder built adds a visible step up in the graph
- Scroll up/down in the overlay with arrow keys if all 6 graphs don't fit on screen
