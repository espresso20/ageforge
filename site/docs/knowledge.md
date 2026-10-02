# Knowledge

Knowledge pays for all research. It accumulates over time and is spent when you start researching a technology. Your knowledge rate and storage decide how quickly you can move through the tech tree.

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
| Monastery Library | Medieval | 30.1 | 4 |
| University | Renaissance | 79.9 | 5 |
| Natural Philosophy Hall | Colonial | 297 | 5 |
| Research Institute | Industrial | 12.8 | 6 |
| Academy | Victorian | 25.6 | 6 |
| Physics Laboratory | Electric | 51.2 | 7 |
| Research Campus | Atomic | 102 | 7 |
| Think Tank | Modern | 205 | 8 |
| Innovation Hub | Information | 410 | 8 |
| AI Research Lab | Digital | 819 | 10 |
| Neuro Research Center | Cyberpunk | 1.64K | 10 |
| Theoretical Institute | Fusion | 3.28K | 12 |
| Deep Space Observatory | Space | 6.55K | 12 |
| Xenology Institute | Interstellar | 13.1K | 15 |
| Cosmic Research Station | Galactic | 26.2K | 15 |
| Reality Academy | Quantum | 52.4K | 20 |

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

Buildings with no workers assigned still produce 20% of their base rate. Full assignment gives full output.

---

## Knowledge Storage

Knowledge is limited by your storage. Build **Storage lineage** buildings (Stash → Storage Pit → Warehouse → Classical Vault → ...) to raise it. Once knowledge hits its storage limit, further production is wasted. You can only pay for a tech with knowledge you hold, so your storage must be at least the tech's cost. Raise storage before you save up for an expensive one.

---

## The Research Queue

```
research <tech>
```

The full knowledge cost is paid **up front**, the moment you issue the `research` command. Nothing is taken per tick while the research runs, so your stored knowledge can rise or fall afterwards without affecting it. Only one technology can be researched at a time, and techs are age-gated: you must be in the right age to research them.

See [Technologies](technologies.md) for the full tech tree.

---

## Epoch Event Interactions

Three epoch transition events affect knowledge directly:

| Event | Type | Effect |
|-------|------|--------|
| The Grand Discovery | Good (Major) | 3 technologies completed at once, free |
| Political Instability | Challenging | Knowledge -2/tick for 156 ticks (plus 60% of your faith lost) |
| The Dark Age | Challenging | Cancels your current research (the knowledge paid for it is lost), removes 80% of your stored knowledge, then knowledge -3/tick for 374 ticks |

The Dark Age is the most punishing event for knowledge-heavy civilizations. High faith lowers the chance of every bad epoch event, the Dark Age included. See [Faith](faith.md) and [Epochs](epochs.md).

There is also a prestige-run path to free research: early in a new run (Primitive or Stone Age), the **Ancient Civilization Memory** cache can offer a single tech with no prerequisites and no knowledge cost, researched at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## Ancient Knowledge (Epoch Succumb Reward)

Succumbing to a catastrophe grants **Ancient Knowledge**: a permanent +25% research speed per distinct epoch succumbed (up to +150% from Iron to Cosmic) that persists through prestige. Players who plan to Succumb early build a research advantage that carries into every later run.

---

## Knowledge and Culture

Culture does not raise your knowledge rate. The link is indirect: culture fill above 40% at an epoch transition makes Major good events eligible, and one of them, The Grand Discovery, completes 3 techs for free. See [Epochs](epochs.md).

---

## Tips

- Get your first Story Circle and 2 Knowledge workers before your first age advance, because early research unlocks pay off quickly. The Stone Age asks for 150 knowledge and 5 Story Circles, the Bronze Age for 1.5K knowledge and 5 Elders' Halls.
- Research is paid up front, so start a tech as soon as you can afford it. Knowledge that sits at its storage limit is wasted.
- After a Dark Age, move spare workers onto knowledge buildings for a while to rebuild your stock, then restart the research it canceled.
