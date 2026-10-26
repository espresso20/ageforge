# Scholars and Research

Scholars do your research. A knowledge building comes with its scholars the moment a copy stands. They work on one technology at a time, and a bar fills as they work until it reads 100%. A scholar who is not researching teaches, which raises what your workers make.

Knowledge is not a resource. Nothing holds it, spends it or prices a tech in it. There is no store to fill and no amount to afford: a tech needs only its prerequisites and its age to be started.

No gate asks you to hold knowledge. The gates do ask for knowledge buildings: you leave the Primitive Age with 5 Story Circles, the Stone Age with 5 Elders' Halls, the Bronze Age with 5 Scriptoria, the Iron Age with 7 Agoras and so on (see [Advancement Requirements](ages.md#advancement-requirements) and the [Advance panel](advance.md)). Of the research itself, an age asks for one tech: the keystone its wonder cannot be built without, and the techs that one stands on (see [Keystone Techs](ages.md#keystone-techs)).

---

## What a scholar is

A scholar is a person, but not a worker. Scholars are not recruited and need no assignment: build a Story Circle and its 2 scholars move in and start work. They live in your homes and eat like anyone (see [People](workers-and-domains.md#the-people-of-your-town)). If your homes are full, they still move in, and the reserve stops growing until there is room again. Knowledge buildings cannot be sold.

Each age has its own word for its scholars, and the game shows it. A scholar's kind is the worker class of the age of its building: a Shaman in the Primitive Age, an Elder in the Stone Age, a Scribe in the Bronze Age, and so on up to the Quantum Theorist. The `scholars` command lists your kinds and counts (see [Commands](commands.md#research)).

The knowledge lineage, in order:

Story Circle → Elders' Hall → Scriptorium → Agora → Library → Monastery Library → University → Natural Philosophy Hall → Research Institute → Academy → Physics Laboratory → Research Campus → Think Tank → Innovation Hub → AI Research Lab → Neuro Research Center → Theoretical Institute → Deep Space Observatory → Xenology Institute → Cosmic Research Station → Reality Academy

| Building | Age | Scholar | Scholars a copy | Copies at full rate |
|----------|-----|---------|-----------------|---------------------|
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

### The soft cap

A knowledge building researches at full rate up to the count the gate out of its age asks for, the last column of the table. Past that count nothing stops you. Another copy builds at the same climb in price (each copy costs 50% more than the one before it), its scholars move in as before, and each copy past the count researches at half rate.

The Buildings panel says it beside the building (`2 scholars a copy. Past 5 copies, each researches at half rate.`). The Scholars tab of the People panel lists each knowledge building with its scholars, and says how many of its copies work at half rate.

### The work

The scholars work whether or not you look at them. Each tech shows its time, and the Scholars tab gives the work a tick. Later buildings do far more than earlier ones, and all of them keep working. [Morale](morale.md) multiplies the work, like every building's output, and so do the bonuses under [Research speed](#research-speed).

---

## One tech at a time

```
research <tech>
plan research <tech>
```

A tech has a price, the work it needs, and a bar. Each tick the researching scholars' work fills the bar, and the tech is done when it reads 100%. Only one technology is in progress at a time. Techs are age-gated: you must be in the right age to research them, and you need their prerequisites.

Starting a tech costs nothing. Cancelling it (`research cancel`) loses the work done on it.

### The bar

The tech in progress shows a bar, its percent and the time left, wherever research is shown: the header of the Research panel, the tech's card, `research list`, the Logs panel and the Scholars tab of the People panel. The bar is drawn the way a building under construction is:

```
Primitive Writing ████░░░░░░ 40% (~4m 20s left)
```

### The time

A tech's time is its remaining work over the work the researching scholars do each tick. A tech you have not started shows the time it would take the scholars your research setting puts on research, with research speed, Ancient Knowledge and Era Mastery counted. That holds even when nothing is in progress and every scholar teaches: the time is what the tech would take once started. If the setting leaves no scholar researching, the screens say the tech needs scholars researching to go on, and the bar stays where it is. When you change who researches, every time changes with it.

The [tech tables](technologies.md#tech-tree-by-age) give each tech's time for a full set of scholars at its age: every knowledge building up to that age built to its full-rate count, researching, before any bonus. Inside an age, each tech costs about 1.3 times the one before it.

### Research takes more than a stay

From the Stone Age on, an age's techs add up to a little more than the work the scholars of a typical town do while the town stays in that age (about 1.2 times). A typical town does not finish all of an age's techs before it leaves the age, and what is left stays open in the ages after. The Primitive Age asks for less than a stay's work (about 0.8 times), so a new player sees a tech finish early. The keystone, the tech the age's wonder needs, is priced to finish inside its age, because you cannot leave without it.

---

## The queue

To line techs up, add them to the [Build Plan](plan.md) with `plan research <tech>` (or `plan res`). The plan's research items are the queue. When the tech in progress is done, the next one starts by itself, in plan order, and keeps doing so while you are away. A tech's prerequisites must be researched, in progress or planned above it, and a planned tech that loses one waits for it. You can also plan the next age's techs. They wait for the advance.

The plan is the only research queue, and the game never chooses what you research next. `research list` shows the queue in order. The techs you plan are remembered with the plan, and [Plan Template](prestige.md#plan-template) plans them again on later runs after your first prestige, in the age you planned them in.

Nothing carries over from one tech to the next. Work past a finished tech's price is discarded, and the next tech you start begins at zero. Work that arrives while nothing is in progress is lost.

---

## Research and teaching

One setting divides your scholars: the share that researches. It is 100% to begin with. The rest teach.

```
scholars              the kinds, how many research and teach, the bonus
scholars <percent>    set the share that researches
```

The split goes to whole scholars. With nothing in progress and nothing queued, every scholar teaches, whatever the setting says; start a tech and the researching share goes to work.

Teaching raises what your workers make. With every scholar teaching, the bonus is:

| Ages | Worker output |
|------|---------------|
| Primitive to Bronze | +10% |
| Iron to Medieval | +8% |
| Renaissance to Industrial | +6% |
| Victorian to Digital | +5% |
| Cyberpunk and later | +4% |

The bonus is in proportion to the share teaching: if half of your scholars teach, you get half of it. It is a source in the worker output bonus, with a line of its own in `stats` and `rates`.

---

## Research speed

Several things shorten a tech's time or raise the scholars' work:

- Research speed comes from milestones. Five milestones, three of them in the Scholar Chain, add +50% in total. It takes that share off every tech's time.
- Research output raises the scholars' work. Fifteen techs add to it (see [Output of one resource](technologies.md#output-of-one-resource)), and some wonders add research of their own.
- Scholar output, from milestones and wonders, raises the work too.
- Thirteen research techs, from Scholasticism to the Timeless Archive, take 3% to 6% each off the time. Between them they never take it under half.
- Ancient Knowledge, the Succumb reward, multiplies the time by 0.8 for each distinct epoch you have succumbed in. You keep it through prestige (see [Ancient Knowledge](catastrophe.md#ancient-knowledge)).
- Era Mastery: on known ground, the age's speed multiplies the work (see [Era Mastery](prestige.md#era-mastery)).
- The [Cosmic Legacy](prestige.md#cosmic-legacy) multiplies the work by 1.1.

The Research section of the Stats panel lists the research speed, the research techs' cut and Ancient Knowledge, each when you have it. See [Technologies](technologies.md#research-speed-sources) for the sources in full and [Technologies](technologies.md) for the tech tree.

---

## Other sources of research

Events, loot, milestone rewards, the Ancient Cache and two trade routes (the colonial exchange and the cotton exchange) give research. A gain is progress on the tech in progress, worth the minutes of your scholars' work it is sized in. A gain adds at most half of that tech's price, so a big one cannot finish a tech by itself unless the bar is already past halfway. With nothing in progress the gain is lost. The log says which: `Gained 12% of Banking.`, `Gained enough to complete Banking.` or `Your scholars had nothing to study, so the insight was lost.` When the cap cut a gain, the log adds `Your scholars gained some ground on Banking.`

---

## Epoch event interactions

Three epoch transition events affect research directly:

| Event | Type | Effect |
|-------|------|--------|
| The Grand Discovery | Good (Major) | Completes up to 3 techs you could start now (age reached, prerequisites done), cheapest first, free |
| Political Instability | Challenging | Research speed -30% for a spell (plus 60% of your faith lost) |
| The Dark Age | Challenging | Cancels your current research (its work is lost), then research speed -40% for a spell |

A high faith strength makes a good event more likely at an epoch transition, and so a bad one less likely, the Dark Age included. A [culture strength](epochs.md#culture-strength) over 40% at an epoch transition makes Major good events like The Grand Discovery possible. See [Faith](faith.md) and [Epochs](epochs.md).

A new run can also bring free research. Early in a run after your first prestige (in the Primitive or Stone Age), the Ancient Civilization Memory cache can offer a single tech with no prerequisites, researched at half speed. See [Prestige](prestige.md#ancient-civilization-memory).

---

## The last age

The Transcendent Age has no knowledge building of its own. Your scholars from the ages before keep working, and they research its four techs.

---

## Old saves

A save made before scholars loads clean. The knowledge it held becomes a gift of work for the tech in progress (at most half its price), and is lost when nothing is in progress. A tech that was paid for and waiting out its time completes on load. The log says once: `Knowledge is no longer stored. Your scholars now work on one technology at a time.` See [Saving and Loading](saving-and-loading.md#saves-from-before-scholars).

---

## Tips

- Build a Story Circle early: its 2 scholars start the first tech at once. The gate out of the Primitive Age asks for 5 of them. The Stone Age's wonder needs Stoneworking, which stands on Tool Making.
- In each new age, find the keystone first (the ★ on the tech tree). The wonder waits for it, and the advance waits for the wonder.
- Queue techs with `plan research` to keep the bar filling while you are away. With nothing in progress and nothing queued, your scholars teach.
- To teach more while a long tech runs, lower the share with `scholars <percent>`. The tech takes longer, and your workers make more.
- After a Dark Age, restart the research it canceled.
