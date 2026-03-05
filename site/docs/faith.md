# Faith

Faith is a resource that *drains* — unlike culture which accumulates, faith naturally declines unless you continuously produce it. Your faith level as a percentage of your faith storage cap has powerful effects on your civilization, most critically on your [Epoch](epochs.md) event roll odds.

---

## Why Faith Matters

1. **Epoch Roll Odds** — The most important effect. Faith % directly determines your chances of getting a good (vs bad) event at each epoch transition.
2. **Morale** — High faith boosts soldier effectiveness in expeditions.
3. **Cohesion** — Mid-to-high faith reduces the severity of negative events.
4. **Prestige Bonus** — Faith at 100% cap at prestige time grants a bonus prestige multiplier.
5. **Wonder Requirements** — Some wonders require minimum faith thresholds.

---

## Faith Threshold Bands

| Band | Condition | Epoch Good-Roll Odds | Other Effects |
|------|-----------|---------------------|---------------|
| No Faith | 0% | 40% | Morale penalty |
| Dim Faith | 1–25% | 40% | — |
| Low Faith | 26–50% | 50% | Minor cohesion |
| Mid Faith | 51–75% | 50% | Cohesion |
| Strong Faith | 76–99% | 60% | Cohesion + morale |
| Faith Full | 100% | 60% | All above + prestige bonus |

The faith % display in the Economy tab shows your current band alongside the epoch odds.

---

## How Faith Drains

Faith naturally declines each tick. The drain scales with your population — larger civilizations need proportionally more faith investment to maintain their band. If you stop producing faith, it will fall to zero.

---

## How to Produce Faith

Faith is produced by **Faith lineage buildings** with **Faith-domain workers** assigned.

**Faith lineage** (in order):

Shrine → Standing Stones → Altar → Temple → Oracle House → Cathedral → Basilica → Mission → Church → Grand Cathedral → Revival Hall → Spiritual Center → Meditation Center → Digital Temple → Cyber Shrine → Neon Sanctuary → Quantum Chapel → Orbital Sanctuary → Void Monastery → Stellar Shrine → Transcendence Hall

**Recruit and assign Faith workers:**

```
recruit faith
assign faith shrine 3
assign faith temple 5
```

---

## The Faith Supply Chain

| Stage | Target Setup |
|-------|-------------|
| Early game | 1 Shrine + 2–3 Faith workers |
| Mid game | Multiple Temples / Cathedrals with full Faith worker assignment |
| Late game | Maintain 76%+ faith before every epoch boundary |

---

## Common Mistake

Players often neglect faith until the epoch notification appears — but by then it is too late to raise it before the roll fires. Build faith infrastructure at the **start** of each epoch, not the end.

---

## Epoch Preparation Checklist

Before reaching the last age of an epoch (e.g. Bronze Age before Iron Era):

- Faith at 76%+ (Strong Faith band) for 60% good roll
- Culture at 10,000+ for Major event eligibility, 250,000+ for Legendary
- Buildings diversified (Great Fire cannot destroy what you haven't over-concentrated)
- Gold reserves for Economic Crash mitigation

See [Epochs](epochs.md) for the full event table and Catastrophe mechanics.
