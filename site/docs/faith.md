# Faith

Faith is a resource that accumulates from Faith lineage buildings and faith-domain workers. It does **not** drain automatically each tick — it only decreases through specific epoch events (e.g. Political Instability removes 60% of current faith). Your faith level as a percentage of your faith storage cap determines your [Epoch](epochs.md) event roll odds.

---

## Why Faith Matters

1. **Epoch Roll Odds** — Faith % directly determines your chances of getting a good (vs bad) event at each epoch transition.
2. **Worker Morale** — Producing faith now also lifts worker [morale](morale.md). An active faith economy keeps spirits up: a small morale lift each tick scales with your faith **production rate** (faith/tick), not your stored faith. The per-tick lift is **capped**, so even a late-game faith firehose can't peg morale in one step — but a steady faith income is a passive, ongoing morale source on top of the epoch odds.

---

## Faith Threshold Bands

| Condition | Epoch Good-Roll Odds |
|-----------|---------------------|
| Faith < 25% of cap | 40% |
| Faith 25–75% of cap | 50% |
| Faith > 75% of cap | 60% |

The faith % display in the Economy tab shows your current percentage alongside the epoch odds.

---

## How Faith Decreases

Faith does not drain automatically. It only decreases when specific epoch events fire — for example, Political Instability removes 60% of your current faith, and Cultural Festival adds a temporary faith production bonus that stops after the event duration. Outside of events, your faith total is stable.

---

## How to Produce Faith

Faith is produced by **Faith lineage buildings** with **Faith-domain workers** assigned.

**Faith lineage** (in order):

Shrine → Standing Stones → Altar → Temple → Oracle House → Cathedral → Basilica → Mission → Church → Grand Cathedral → Revival Hall → Spiritual Center → Meditation Center → Digital Temple → Cyber Shrine → Neon Sanctuary → Quantum Chapel → Orbital Sanctuary → Void Monastery → Stellar Shrine → Transcendence Hall

**Recruit and assign Faith workers:**

```
recruit 3
assign shrine 3
assign temple 5
```

Workers are recruited generically (no domain argument). They become Faith workers when assigned to a faith-domain building.

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

- Faith at 76%+ for 60% good roll
- Culture at 10,000+ for Major event eligibility, 250,000+ for Legendary
- Buildings diversified (Great Fire cannot destroy what you haven't over-concentrated)
- Gold reserves for Economic Crash mitigation

See [Epochs](epochs.md) for the full event table and Catastrophe mechanics.
