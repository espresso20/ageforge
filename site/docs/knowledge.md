# Scholars and Research

Research is done by **scholars**. A knowledge building comes with its scholars the moment a copy stands. They work on **one technology at a time**, and a bar fills as they do, until it reads 100%. A scholar who is not researching **teaches**, which raises what your workers make.

Knowledge is not a resource. Nothing holds it, spends it or prices a tech in it. There is no store to fill and no amount to afford: a tech needs only its prerequisites and its age to be started.

No age asks you to hold anything for research to advance. The gates do ask for knowledge buildings: 5 Story Circles for the Stone Age, 5 Elders' Halls for the Bronze Age, 5 Scriptoria for the Iron Age, 7 Agoras for the Classical Age and so on (see [Advancement Requirements](ages.md#advancement-requirements)). What an age needs of research itself is one tech: the **keystone** its wonder cannot be built without, and the techs that one stands on (see [Keystone Techs](ages.md#keystone-techs)).

---

## What a Scholar Is

A scholar is not a worker. He is not recruited, takes no housing, eats no food and needs no assignment. Build a Story Circle and its 2 scholars are at work. Knowledge buildings cannot be sold.

A scholar's **kind** is the worker class of the age of his building: a Shaman in the Primitive Age, an Elder in the Stone Age, a Scribe in the Bronze Age, and so on up to the Quantum Theorist. The `scholars` command lists your kinds and counts (see [Commands](commands.md#research)).

**Knowledge lineage** (in order):

Story Circle → Elders' Hall → Scriptorium → Agora → Library → Monastery Library → University → Natural Philosophy Hall → Research Institute → Academy → Physics Laboratory → Research Campus → Think Tank → Innovation Hub → AI Research Lab → Neuro Research Center → Theoretical Institute → Deep Space Observatory → Xenology Institute → Cosmic Research Station → Reality Academy

| Building | Age | Scholar | Scholars a copy | Most copies |
|----------|-----|---------|-----------------|-------------|
| Story Circle | Primitive | Shaman | 2 | 5 |
| Elders' Hall | Stone | Elder | 2 | 5 |
| Scriptorium | Bronze | Scribe | 3 | 5 |
| Agora | Iron | Scholar | 3 | 7 |
| Library | Classical | Philosopher | 4 | 8 |
| Monastery Library | Medieval | Friar | 4 | 5 |
| University | Renaissance | Academician | 5 | 7 |
| Natural Philosophy Hall | Colonial | Naturalist | 5 | 5 |
| Research Institute | Industrial | Engineer-Scientist | 6 | 5 |
| Academy | Victorian | Victorian Scholar | 6 | 7 |
| Physics Laboratory | Electric | Research Fellow | 7 | 8 |
| Research Campus | Atomic | Nuclear Scientist | 7 | 9 |
| Think Tank | Modern | Modern Researcher | 8 | 10 |
| Innovation Hub | Information | Data Scientist | 8 | 8 |
| AI Research Lab | Digital | AI Researcher | 10 | 8 |
| Neuro Research Center | Cyberpunk | Cyber-Scholar | 10 | 5 |
| Theoretical Institute | Fusion | Fusion Theorist | 12 | 5 |
| Deep Space Observatory | Space | Orbital Researcher | 12 | 5 |
| Xenology Institute | Interstellar | Stellar Scientist | 15 | 5 |
| Cosmic Research Station | Galactic | Galactic Researcher | 15 | 5 |
| Reality Academy | Quantum | Quantum Theorist | 20 | 10 |

**The cap.** A knowledge building cannot be built past its most copies, the count its age's gate asks for. Each copy costs 50% more than the one before it. The Buildings panel shows a knowledge building's scholars a copy and its copies against its cap (`2 scholars a copy, 3 of 5 built.`), and at the cap it reads `Building limit reached.`

**Work.** The scholars do their work whether or not you look at them. The game shows it as a time on a tech, not as a number. Later buildings do far more than earlier ones, and they all keep working. [Morale](morale.md) multiplies the work, like every building's output, and so do the bonuses under [Research Speed](#research-speed).

---

## One Tech at a Time

```
research <tech>
plan research <tech>
```

A tech has a price, the work it needs, and a **bar**. Each tick the researching scholars' work fills the bar, and the tech is done when it reads 100%. Only one technology is in progress at a time, and techs are age-gated: you must be in the right age to research them.

Starting a tech costs nothing. Cancelling it (`research cancel`) loses the work done on it.

**The bar.** The tech in progress shows a bar, its percent and the time left, wherever research is shown: the header of the Research panel, the tech's card, `research list` and the Activity panel. It is the bar a building under construction has:

```
Primitive Writing ████░░░░░░ 40% (~4m 20s left)
```

**The time.** A tech's time is its remaining work over the work the researching scholars do each tick. A tech you have not started shows the time it would take the scholars your research setting puts on research, with research speed, Ancient Knowledge and Era Mastery counted. That holds even when nothing is in progress and every scholar teaches: the time is what the tech would take once started. If the setting leaves no scholar researching, the screens say the tech needs scholars researching to go on, and the bar stays where it is. Change who researches and every time changes with it.

**The tables.** The [tech tables](technologies.md#tech-tree-by-age) give each tech's time for a full set of scholars at its age: every knowledge building up to that age built to its cap, researching, before any bonus. A tech costs 1.3 times the one before it inside an age.

---

## The Queue

To line techs up, add them to the [Build Plan](plan.md) with `plan research <tech>` (or `plan res`). The plan's research items are the queue: when the tech in progress is done, the next one starts by itself, in plan order, and keeps doing so while you are away. A tech's prerequisites must be researched, in progress or planned above it; a planned tech that loses one waits for it. You can also plan the next age's techs; they wait for the advance.

The plan is the only research queue: the game never chooses what you research next. `research list` shows the queue in order. The techs you plan are remembered with the plan, and with the prestige legacy kit's [Plan Template](prestige.md#plan-template) they are planned again on later runs, in the age you planned them in.

**A head start.** Work that has no tech to land on is kept for the next tech you start, never more than that tech's price. `research list` shows it (`Head start: 12% of Banking`).

---

## Research and Teaching

One setting divides your scholars: the share that **researches**. It is 100% to begin with. The rest **teach**.

```
scholars              the kinds, how many research and teach, the bonus
scholars <percent>    set the share that researches
```

The split goes to whole scholars. With nothing in progress and nothing queued, every scholar teaches, whatever the setting says; start a tech and the researching share goes to work.

**Teaching** raises what your workers make. With every scholar teaching, the bonus is:

| Ages | Worker output |
|------|---------------|
| Primitive to Bronze | +10% |
| Iron to Medieval | +8% |
| Renaissance to Industrial | +6% |
| Victorian to Digital | +5% |
| Cyberpunk and later | +4% |

The bonus is in proportion to the share teaching: half of your scholars teaching give half of it. It is a source in the worker output bonus, with a line of its own in `stats` and `rates`.

---

## Research Speed

Research speed is what shortens a tech's time. It is made of:

- **Research speed** from milestones (five milestones, three of them in the Scholar chain, add +50% in total).
- **Research output** from techs and wonders, which raises the scholars' work. Fifteen techs add to it (see [Output of one resource](technologies.md#output-of-one-resource)).
- **Research techs**: seven techs take another 3% to 6% each off the time (Scholasticism, Printing Press, Scientific Method, Public Education, Big Science, Computers and Machine Learning).
- **Ancient Knowledge**, the Succumb reward, multiplies the time by 0.8 for each distinct epoch you have succumbed in, kept through prestige (see [Ancient Knowledge](catastrophe.md#ancient-knowledge)).
- **Era Mastery**: on known ground the age's speed multiplies the work (see [Era Mastery](prestige.md#era-mastery)).

The Stats panel lists each of these that you have. See [Technologies](technologies.md) for the full tech tree.

---

## What Used to Give Knowledge

Events, loot, milestone rewards, the Ancient Cache and two trade routes (the colonial exchange and the cotton exchange) still give research. A gain is **progress on the tech in progress**, worth the minutes of your scholars' work it is sized in. With nothing in progress it is a **head start** on the next tech you start. It is never more than that tech's work, and a gain that fills the bar finishes the tech. The log says which: `Gained 12% of Banking.`, `Gained enough to complete Banking.` or `Gained a head start on the next tech.`

---

## Epoch Event Interactions

Three epoch transition events affect research directly:

| Event | Type | Effect |
|-------|------|--------|
| The Grand Discovery | Good (Major) | Completes up to 3 techs of your current age or earlier that you haven't researched yet, free |
| Political Instability | Challenging | Research speed -30% for a spell (plus 60% of your faith lost) |
| The Dark Age | Challenging | Cancels your current research (its work is lost), takes the head start you had kept, then research speed -40% for a spell |

The Dark Age is the most punishing event for a research-heavy civilization. High faith lowers the chance of every bad epoch event, the Dark Age included. A [culture strength](epochs.md#culture-strength) over 40% at an epoch transition makes Major good events like The Grand Discovery possible. See [Faith](faith.md) and [Epochs](epochs.md).

There is also a prestige-run path to free research: early in a new run (Primitive or Stone Age), the **Ancient Civilization Memory** cache can offer a single tech with no prerequisites, researched at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## The Last Age

The Transcendent Age has no knowledge building of its own. Your scholars from the ages before keep working, and they research its four techs.

---

## Old Saves

A save made before scholars loads clean. Stored knowledge becomes a head start. A tech that was paid for and waiting out its time completes on load. Workers on knowledge buildings return to the idle pool. The log says once: `Knowledge is no longer stored. Your scholars now work on one technology at a time.` See [Saving and Loading](saving-and-loading.md).

---

## Tips

- Build your first Story Circle before your first age advance: its 2 scholars start the first tech at once, and early research unlocks pay off quickly. The Stone Age asks for 5 Story Circles and the Bronze Age for 5 Elders' Halls, and the Stone Age's wonder needs Stoneworking, which stands on Tool Making.
- In each new age, find the keystone first (the ★ on the tech tree). The wonder waits for it, and the advance waits for the wonder.
- Queue techs with `plan research` so the bar never stands empty: with nothing in progress your scholars teach.
- To teach more while a long tech runs, lower the share with `scholars <percent>`. The tech takes longer, and your workers make more.
- After a Dark Age, restart the research it canceled.
