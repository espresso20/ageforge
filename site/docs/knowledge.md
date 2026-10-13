# Knowledge

Knowledge pays for all research. It accumulates over time and is spent when you start researching a technology. Your knowledge rate and your knowledge store decide how quickly you can move through the tech tree. Buildings cost it too from the Medieval to the Colonial Age.

Knowledge is not a physical thing, so it has a **store of its own**, and only knowledge buildings raise it. Storage buildings do not hold it. See [The Knowledge Store](#the-knowledge-store).

No age asks you to hold knowledge to advance. The gates do ask for knowledge buildings: 5 Story Circles for the Stone Age, 5 Elders' Halls for the Bronze Age, 5 Scriptoria for the Iron Age, 7 Agoras for the Classical Age and so on (see [Advancement Requirements](ages.md#advancement-requirements)). What an age needs of research itself is one tech: the **keystone** its wonder cannot be built without, and the techs that one stands on (see [Keystone Techs](ages.md#keystone-techs)). The rest of your knowledge is yours to spend on the techs you choose.

---

## How Knowledge is Produced

Knowledge is produced by **Knowledge lineage buildings**, at full rate when their worker slots are filled. Workers in them are Knowledge workers; see [Workers](workers-and-domains.md) for recruiting, the roster and the domain table.

**Knowledge lineage** (in order):

Story Circle → Elders' Hall → Scriptorium → Agora → Library → Monastery Library → University → Natural Philosophy Hall → Research Institute → Academy → Physics Laboratory → Research Campus → Think Tank → Innovation Hub → AI Research Lab → Neuro Research Center → Theoretical Institute → Deep Space Observatory → Xenology Institute → Cosmic Research Station → Reality Academy

**Knowledge buildings (rates fully staffed, per building; each copy costs 50% more than the one before):**

| Building | Age | Knowledge/tick | Workers | Adds to the store | Copies the gate asks |
|----------|-----|----------------|---------|-------------------|----------------------|
| Story Circle | Primitive | 0.2 | 2 | +140 | 5 |
| Elders' Hall | Stone | 0.6 | 2 | +370 | 5 |
| Scriptorium | Bronze | 2.0 | 3 | +3.1K | 5 |
| Agora | Iron | 1.6 | 3 | +6.4K | 7 |
| Library | Classical | 3.2 | 4 | +19K | 8 |
| Monastery Library | Medieval | 42.4 | 4 | +330K | 5 |
| University | Renaissance | 76.3 | 5 | +820K | 7 |
| Natural Philosophy Hall | Colonial | 452 | 5 | +8.4M | 5 |
| Research Institute | Industrial | 12.8 | 6 | +910K | 5 |
| Academy | Victorian | 25.6 | 6 | +4.5M | 7 |
| Physics Laboratory | Electric | 51.2 | 7 | +11M | 8 |
| Research Campus | Atomic | 102 | 7 | +16M | 9 |
| Think Tank | Modern | 205 | 8 | +13M | 10 |
| Innovation Hub | Information | 410 | 8 | +15M | 8 |
| AI Research Lab | Digital | 819 | 10 | +94M | 8 |
| Neuro Research Center | Cyberpunk | 1.64K | 10 | +310M | 5 |
| Theoretical Institute | Fusion | 3.28K | 12 | +1.5B | 5 |
| Deep Space Observatory | Space | 6.55K | 12 | +3.4B | 5 |
| Xenology Institute | Interstellar | 13.1K | 15 | +4.9B | 5 |
| Cosmic Research Station | Galactic | 26.2K | 15 | +1.4B | 5 |
| Reality Academy | Quantum | 52.4K | 20 | +5.6B | 10 |

The first five tiers have hand-set rates. From the **Medieval to the Colonial Age**, buildings also cost knowledge, so knowledge counts as a construction resource and its producers follow the [Payback Rule](buildings.md#how-production-rates-are-set): fully staffed, each earns back its first copy's price within the age's payback time. That is why the Monastery Library, University and Natural Philosophy Hall jump so far ahead. From the Industrial Age on, research is the only regular use of knowledge, and the lineage goes back to fixed rates that double each age. Your older halls and universities keep producing through all of this, so don't sell them.

Buildings with no workers still produce 20% of their base rate; each filled slot adds its share of the other 80%:

```
knowledge/tick = base rate × buildings × (0.20 + 0.80 × workers assigned / worker slots)
```

[Morale](morale.md) multiplies that output, like every building's.

---

## The Knowledge Store

Knowledge is not a physical thing, so it does not sit in the stores that hold food, wood and stone. It has a **store of its own**:

```
knowledge store = 240 + the copies of each knowledge building × what one copy adds
```

The 240 is yours with no knowledge building. Each copy of a knowledge building adds its figure from the table above. The Resources box shows knowledge first, set apart from the goods, as the amount held over this store. The Buildings panel shows what a copy adds on each knowledge building.

**Only knowledge buildings raise it.** Storage buildings do not, nor do the techs and wonders that add storage to every resource but knowledge, Era Mastery or prestige. Once knowledge hits the store, further production is wasted.

**The wall.** A tech priced over the knowledge store cannot be started, however long you wait. The `research` command refuses it, a planned tech waits as blocked ("needs more knowledge storage"), and the tech tree marks its price with a red ✗. The refusal names the knowledge building of your age to build:

```
Fire Mastery costs 398 knowledge, more than your knowledge storage holds (240). Build more storage first: Elders' Hall.
```

A building priced in knowledge (Medieval to Colonial) meets the same wall. Knowledge you already hold above the store, from a save made before knowledge had its own store, can still pay.

**The gate sizes it.** The gate asks for copies of the age's knowledge building, and the store they hold is sized to the techs: an age's dearest tech fits with a quarter to spare, and the second-cheapest tech of the next age fits with a fifth to spare. Each knowledge building climbs at 1.5 a copy, so the gate asks for 5 to 10 of them, not the 8 to 20 it did.

**An age advance** trims what you hold, as it does every stock but faith (see [What Happens on Age Advance](ages.md#what-happens-on-age-advance)). Catastrophes spare knowledge buildings, as they spare storage and wonders (see [Catastrophe](catastrophe.md)).

---

## What Research Costs

Inside an age each tech costs **1.3 times the one before it**, from the cheapest to the dearest. An age's techs together cost 80% of the knowledge a well-run town's knowledge buildings make over the age, and the store above is sized so the dearest one fits. The Stone Age's six techs cost about 7.3K knowledge between them, the Medieval Age's nine about 4.2M, the Renaissance's seven about 21M, the Atomic Age's eleven about 1.0B. The Transcendent Age has no knowledge building of its own, so its four techs keep the prices they had. See [How Research is Priced](technologies.md#how-research-is-priced) and the [tech tables](technologies.md#tech-tree-by-age).

---

## Researching and the Research Queue

```
research <tech>
plan research <tech>
```

The full knowledge cost is paid **up front**, the moment research starts. Nothing is taken per tick while it runs, so your stored knowledge can rise or fall afterwards without affecting it. Only one technology can be researched at a time, and techs are age-gated: you must be in the right age to research them.

To queue techs, add them to the [Build Plan](plan.md) with `plan research <tech>` (or `plan res`). The plan starts them one at a time, in order, as the knowledge comes in, and keeps doing so while you are away. A tech's prerequisites must be researched, in progress or planned above it; a planned tech that loses one waits for it. You can also plan the next age's techs; they wait for the advance.

The plan is the only research queue: the game never chooses what you research next. The techs you plan are remembered with the plan, and with the prestige legacy kit's [Plan Template](prestige.md#plan-template) they are planned again on later runs, in the age you planned them in.

Research speed shortens how long each tech takes. It comes from milestones (five milestones, three of them in the Scholar chain, add +50% in total). Seven techs take another 3% to 6% each off the time that leaves (Scholasticism, Printing Press, Scientific Method, Public Education, Big Science, Computers and Machine Learning; a quarter with all of them). Ancient Knowledge, the Succumb reward, then multiplies the time that is left by 0.8 for each distinct epoch you have succumbed in, kept through prestige (see [Ancient Knowledge](catastrophe.md#ancient-knowledge)). On known ground, [Era Mastery](prestige.md#era-mastery) then divides the time again by the age's speed. See [Technologies](technologies.md) for the full tech tree.

---

## Epoch Event Interactions

Three epoch transition events affect knowledge directly:

| Event | Type | Effect |
|-------|------|--------|
| The Grand Discovery | Good (Major) | Completes up to 3 techs of your current age or earlier that you haven't researched yet, free |
| Political Instability | Challenging | Knowledge -2/tick for 156 ticks (plus 60% of your faith lost) |
| The Dark Age | Challenging | Cancels your current research (the knowledge paid for it is lost), removes 80% of your stored knowledge, then knowledge -3/tick for 374 ticks |

The Dark Age is the most punishing event for knowledge-heavy civilizations. High faith lowers the chance of every bad epoch event, the Dark Age included. Culture does not raise your knowledge rate, but a [culture strength](epochs.md#culture-strength) over 40% at an epoch transition makes Major good events like The Grand Discovery possible. See [Faith](faith.md) and [Epochs](epochs.md).

There is also a prestige-run path to free research: early in a new run (Primitive or Stone Age), the **Ancient Civilization Memory** cache can offer a single tech with no prerequisites and no knowledge cost, researched at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## Tips

- Get your first Story Circle and 2 Knowledge workers before your first age advance, because early research unlocks pay off quickly. The Stone Age asks for 5 Story Circles and the Bronze Age for 5 Elders' Halls, and the Stone Age's wonder needs Stoneworking, which stands on Tool Making.
- In each new age, find the keystone first (the ★ on the tech tree). The wonder waits for it, and the advance waits for the wonder.
- Research is paid up front, so start a tech as soon as you can afford it, or queue it with `plan research`. Knowledge that sits at the store's limit is wasted, and a tech priced over the store waits for a knowledge building.
- After a Dark Age, move spare workers onto knowledge buildings for a while to rebuild your stock, then restart the research it canceled.
