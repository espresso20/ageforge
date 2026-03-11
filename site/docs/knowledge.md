# Knowledge

Knowledge is the fuel for all research. It accumulates over time and is consumed when you start researching a technology. Your knowledge generation rate and storage cap are the primary drivers of how quickly you can advance through the tech tree.

---

## How Knowledge is Produced

Knowledge is produced by **Knowledge lineage buildings** with **Knowledge-domain workers** assigned.

**Knowledge lineage** (in order):

Story Circle → Elders' Hall → Scriptorium → Agora → Library → Monastery Library → University → Natural Philosophy Hall → Research Institute → Academy → Physics Laboratory → Research Campus → Think Tank → Innovation Hub → AI Research Lab → Neuro Research Center → Theoretical Institute → Deep Space Observatory → Xenology Institute → Cosmic Research Station → Reality Academy

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

---

## Ancient Knowledge (Epoch Succumb Reward)

Succumbing to a catastrophe grants **Ancient Knowledge** — a permanent +25% research speed bonus that persists through prestige resets. Players who plan to Succumb early gain a significant compounding research advantage across all future runs.

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

- Get your first Story Circle + 2 Knowledge workers before your first age advance — early research unlocks compound fast
- Keep knowledge capped before starting research; don't let it drain below the tech cost mid-research
- Library and University unlock tiers have large capacity bonuses — prioritize them when they become available
- The culture thresholds at 500, 2,500, and 10,000 each give +5/+10/+15% knowledge rate — culture investment directly pays off in research speed
- If a Dark Age event fires, cancel non-essential production assignments temporarily to recover knowledge quickly
