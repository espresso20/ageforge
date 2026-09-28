# Knowledge

Knowledge is the fuel for all research. It accumulates over time and is consumed when you start researching a technology. Your knowledge generation rate and storage cap are the primary drivers of how quickly you can advance through the tech tree.

---

## How Knowledge is Produced

Knowledge is produced by **Knowledge lineage buildings** with **Knowledge-domain workers** assigned.

**Knowledge lineage** (in order):

Story Circle → Elders' Hall → Scriptorium → Agora → Library → Monastery Library → University → Natural Philosophy Hall → Research Institute → Academy → Physics Laboratory → Research Campus → Think Tank → Innovation Hub → AI Research Lab → Neuro Research Center → Theoretical Institute → Deep Space Observatory → Xenology Institute → Cosmic Research Station → Reality Academy

**Knowledge rates (fully staffed, per building):**

| Building | Age | Knowledge/tick | Workers |
|----------|-----|----------------|---------|
| Story Circle | Primitive | 0.2 | 2 |
| Elders' Hall | Stone | 0.6 | 2 |
| Scriptorium | Bronze | 2.0 | 3 |
| Agora | Iron | 1.6 | 3 |
| Library | Classical | 3.2 | 4 |
| Monastery Library | Medieval | 78.3 | 4 |
| University | Renaissance | 208 | 5 |
| Natural Philosophy Hall | Colonial | 772 | 5 |
| Research Institute | Industrial | 12.8 | 6 |
| Academy | Victorian | 25.6 | 6 |
| Physics Laboratory | Electric | 51.2 | 7 |
| Research Campus | Atomic | 102 | 7 |
| Think Tank | Modern | 205 | 8 |
| Innovation Hub | Information | 410 | 8 |
| AI Research Lab | Digital | 819 | 10 |
| Neuro Research Center | Cyberpunk | 1,640 | 10 |
| Theoretical Institute | Fusion | 3,280 | 12 |
| Deep Space Observatory | Space | 6,550 | 12 |
| Xenology Institute | Interstellar | 13,100 | 15 |
| Cosmic Research Station | Galactic | 26,200 | 15 |
| Reality Academy | Quantum | 52,400 | 20 |

The first five tiers have hand-set rates. From the **Medieval to the Colonial Age**, buildings also cost knowledge, so knowledge counts as a construction resource and its producers follow the [Payback Rule](buildings.md#how-production-rates-are-set): fully staffed, each earns back its first copy's price within the age's payback time. That is why the Monastery Library, University and Natural Philosophy Hall jump so far ahead. From the Industrial Age on, nothing but research costs knowledge, and the lineage goes back to fixed rates that double each age. Your older halls and universities keep producing through all of this, so don't sell them. While knowledge is a construction resource (Medieval to Colonial), the market also trades it at parity with the age's other construction resources; see [Resources](resources.md#buying-at-the-market).

**Recruit and assign Knowledge workers:**

```
recruit 3
assign story_circle 2
assign library 5
assign university 3
```

Workers are recruited generically (no domain argument). They become Knowledge workers when assigned to a knowledge-domain building.

**Worker efficiency formula:**

```
knowledge/tick = base_rate × building_count × (0.20 + 0.80 × assigned / total_capacity)
```

Buildings with zero workers assigned still contribute 20% of their base rate. Full assignment maximizes output.

---

## Knowledge Storage Cap

Knowledge is capped by your storage buildings. Build **Storage lineage** buildings (Stash → Storage Pit → Warehouse → Classical Vault → ...) to increase your cap. Without enough storage, knowledge production is wasted when the cap is reached. Always expand storage before starting a long research project.

---

## The Research Queue

```
research <tech_key>
```

Knowledge is deducted **upfront** (immediately when you issue the `research` command) — not per tick. Only one technology can be researched at a time. Research is age-gated — you must be in the correct age to unlock certain techs.

See [Technologies](technologies.md) for the full tech tree.

---

## Epoch Event Interactions

Two epoch events directly affect knowledge:

| Event | Type | Effect |
|-------|------|--------|
| Grand Discovery | Good — Major | 3 technologies completed instantly (free) |
| Dark Age | Challenging | Cancels current research, steals knowledge, applies −research debuff for 144 ticks |

The Dark Age is the most punishing event for knowledge-heavy civilizations. Maintaining high faith reduces the probability of all bad epoch events, including Dark Age. See [Faith](faith.md) and [Epochs](epochs.md).

There is also a prestige-run path to free research: early in a new run (Primitive or Stone age), the **Ancient Civilization Memory** cache can offer a single tech researched free of prerequisites and knowledge cost — at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## Ancient Knowledge (Epoch Succumb Reward)

Succumbing to a catastrophe grants **Ancient Knowledge** — a permanent +25% research speed bonus per distinct epoch succumbed (up to +150% from Iron to Cosmic) that persists through prestige resets. Players who plan to Succumb early gain a significant compounding research advantage across all future runs.

---

## Knowledge vs Culture

Both are "soft power" resources that interact with each other:

- **Knowledge** fuels technologies → permanent multipliers
- **Culture** unlocks cultural thresholds → knowledge rate bonuses

They synergize: high culture increases your knowledge production rate, which accelerates research. Investing in culture pays back in research speed.

**Culture knowledge-rate bonuses:**

| Culture Threshold | Knowledge Rate Bonus |
|-------------------|---------------------|
| 500 | +5% |
| 2,500 | +10% |
| 10,000 | +15% |

---

## Tips

- Get your first Story Circle + 2 Knowledge workers before your first age advance — early research unlocks compound fast. The Stone Age itself asks for 150 knowledge and 5 Story Circles, the Bronze Age for 1.5K knowledge and 5 Elders' Halls
- Keep knowledge capped before starting research; don't let it drain below the tech cost mid-research
- Library and University unlock tiers have large capacity bonuses — prioritize them when they become available
- The culture thresholds at 500, 2,500, and 10,000 each give +5/+10/+15% knowledge rate — culture investment directly pays off in research speed
- If a Dark Age event fires, cancel non-essential production assignments temporarily to recover knowledge quickly
