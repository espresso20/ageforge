# Knowledge

Knowledge pays for all research. It accumulates over time and is spent when you start researching a technology. Your knowledge rate and storage decide how quickly you can move through the tech tree. A few buildings also cost knowledge, most of all the Modern Age wonder, the Space Program.

---

## How Knowledge is Produced

Knowledge is produced by **Knowledge lineage buildings**, at full rate when their worker slots are filled. Workers in them are Knowledge workers; see [Workers](workers-and-domains.md) for recruiting, worker shares and the domain table.

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

The first five tiers have hand-set rates. From the **Medieval to the Colonial Age**, buildings also cost knowledge, so knowledge counts as a construction resource and its producers follow the [Payback Rule](buildings.md#how-production-rates-are-set): fully staffed, each earns back its first copy's price within the age's payback time. That is why the Monastery Library, University and Natural Philosophy Hall jump so far ahead. From the Industrial Age on, research is the only regular use of knowledge, and the lineage goes back to fixed rates that double each age. Your older halls and universities keep producing through all of this, so don't sell them.

Buildings with no workers still produce 20% of their base rate; each filled slot adds its share of the other 80%:

```
knowledge/tick = base rate × buildings × (0.20 + 0.80 × workers assigned / worker slots)
```

[Morale](morale.md) multiplies that output, like every building's.

---

## Knowledge Storage

Knowledge is limited by your storage. Build **Storage lineage** buildings (Stash → Storage Pit → Warehouse → Classical Vault → ...) to raise it. Once knowledge hits its storage limit, further production is wasted. You can only pay for a tech with knowledge you hold, so your storage must be at least the tech's cost. Raise storage before you save up for an expensive one.

---

## Researching and the Research Queue

```
research <tech>
plan research <tech>
```

The full knowledge cost is paid **up front**, the moment research starts. Nothing is taken per tick while it runs, so your stored knowledge can rise or fall afterwards without affecting it. Only one technology can be researched at a time, and techs are age-gated: you must be in the right age to research them.

To queue techs, add them to the [Build Plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts them one at a time, in order, as the knowledge comes in, and keeps doing so while you are away. A tech's prerequisites must be researched, in progress or planned above it. You can also plan the next age's techs; they wait for the advance.

Research speed shortens how long each tech takes. It comes from milestones (five in the scholar chain add +50% in total) and from Ancient Knowledge, the Succumb reward: +25% per distinct epoch you have succumbed to, kept through prestige (see the [Legacy Bonus Table](catastrophe.md#legacy-bonus-table)). See [Technologies](technologies.md) for the full tech tree.

---

## Knowledge at the Market

The market trades knowledge two ways, depending on the age:

| Ages | How the market trades knowledge |
|------|---------------------------------|
| Medieval to Colonial | Knowledge is a construction resource, so it trades at parity with the age's other construction resources, less the market's fee. |
| Industrial on | Only gold buys knowledge, at a flat **5 knowledge per gold**. |

Each trade pushes that pair's rate down a little, by the same step whatever its size, and the rate recovers over time, so a few large trades get a better price than many small ones. See [Trade](trade.md) for market rates and pressure.

### Paying for the Space Program

The Modern Age wonder, the **Space Program**, costs **600B knowledge** along with 770B steel, 690B gold and 430B electricity. Your knowledge buildings make thousands per tick at that point, nowhere near enough. The practical way to pay is the market's gold to knowledge pair: at 5 knowledge per gold, 600B knowledge costs about 120B gold, less than a fifth of the gold the wonder asks for anyway.

1. Raise your knowledge storage to hold what you buy: knowledge bought past the cap is lost. Each Modern Depot adds 90B to every store.
2. Trade gold for knowledge in a few large lumps. Amounts can be typed in e-notation: `trade gold knowledge 2e10` sells 20B gold for about 100B knowledge.
3. Bank each lump into the wonder with `wonder bank knowledge`, and repeat.

`plan trade gold knowledge` does step 2 for you as gold comes in, spacing its trades so the rate stays near the market's, but it never buys past your storage cap, so you still bank the knowledge yourself.

---

## Epoch Event Interactions

Three epoch transition events affect knowledge directly:

| Event | Type | Effect |
|-------|------|--------|
| The Grand Discovery | Good (Major) | Completes up to 3 techs of your current age or earlier that you haven't researched yet, free |
| Political Instability | Challenging | Knowledge -2/tick for 156 ticks (plus 60% of your faith lost) |
| The Dark Age | Challenging | Cancels your current research (the knowledge paid for it is lost), removes 80% of your stored knowledge, then knowledge -3/tick for 374 ticks |

The Dark Age is the most punishing event for knowledge-heavy civilizations. High faith lowers the chance of every bad epoch event, the Dark Age included. Culture does not raise your knowledge rate, but culture over 40% of its storage at an epoch transition makes Major good events like The Grand Discovery possible. See [Faith](faith.md) and [Epochs](epochs.md).

There is also a prestige-run path to free research: early in a new run (Primitive or Stone Age), the **Ancient Civilization Memory** cache can offer a single tech with no prerequisites and no knowledge cost, researched at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## Tips

- Get your first Story Circle and 2 Knowledge workers before your first age advance, because early research unlocks pay off quickly. The Stone Age asks for 150 knowledge and 5 Story Circles, the Bronze Age for 1.5K knowledge and 5 Elders' Halls.
- Research is paid up front, so start a tech as soon as you can afford it, or queue it with `plan research`. Knowledge that sits at its storage limit is wasted.
- Plan for the Space Program before the Modern Age: build up gold income and storage, then buy the knowledge.
- After a Dark Age, move spare workers onto knowledge buildings for a while to rebuild your stock, then restart the research it canceled.
