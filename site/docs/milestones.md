# Milestones & Chains

AgeForge has 77 milestones, and 33 of them form 6 chains. Completing every milestone in a chain grants your civilization a **title** and a **temporary game speed boost**. Some milestones stay hidden until you make progress toward them.

Milestones belong to one run: they pay inside it and start over with the next, at prestige and at Succumb. What you keep across runs is on your account, as [badges](badges.md). Milestones feed some of those: two badge ladders count the milestones and the chains you complete across all your runs, and a few one-off badges ask for milestones or chains finished in a single run.

Every milestone and every chain can be completed. See [Can every milestone be done?](#can-every-milestone-be-done) below.

> When a milestone or chain completes, the log shows the achievement and its reward, then a short flavor quip on a dim line below ("You built a Wonder. Your neighbors are impressed. One is drafting a strongly worded letter."). The quip is only there for character and changes nothing.

Type `milestones` (or `ms`) to open the Milestones panel. It opens full screen. Scroll with the arrow keys or PgUp and PgDn, and `Esc` closes it.

<figure class="screen" data-screen="milestones"><figcaption>The Milestones panel: each chain with its progress, and every open milestone with what it asks for and what it pays.</figcaption></figure>

---

## Civilization titles

A civilization title belongs to the run you are playing. It shows next to your progress count at the top of the Milestones panel, and in quotes on the status ribbon when the terminal is wide enough. It is not the title your account wears, which comes from your badges (see [Titles](badges.md#titles)).

When a chain is completed, its title overrides any count-based fallback. If you finish more than one chain, the title of the one lower in this table shows. The boost adds to game speed, so +300% runs the game four times as fast for its duration:

| Chain | Title | Game speed boost |
|---|---|---|
| Settlement | "The Founders" | +300% for 180 ticks |
| Scholar | "The Enlightened" | +300% for 180 ticks |
| Builder | "The Architects" | +250% for 150 ticks |
| Military | "The Conquerors" | +250% for 150 ticks |
| Trade | "The Merchants" | +250% for 150 ticks |
| Ancient Ages | "The Ancients" | +250% for 150 ticks |

These are the boosts' base lengths. From the Bronze Age on the game's clocks run 2.6 times as long, so a chain completed in the Bronze Age or later boosts for 468 or 390 ticks instead.

If no chain is completed, fallback titles apply:

| Milestones completed | Title |
|---|---|
| 5+ | Aspiring |
| 12+ | Rising Power |
| 24+ | Established |
| 40+ | Dominant Force |
| 56+ | Legend |
| 62+ | Transcendent |

---

## Settlement Chain

*Focus: growing your population*

| Milestone | Condition | Reward |
|---|---|---|
| First Shelter | Build 1 Hut | +10 food |
| Small Village | Population: 250 | +50 wood |
| Bustling Town | Population: 1,500 | +5% food production |
| Growing City | Population: 5,000 (Bronze Age) | +10% food production |
| Metropolis *(hidden)* | Population: 15,000 (Iron Age) | +10% all production |
| Megalopolis *(hidden)* | Population: 250,000 (Classical Age) | +20% all production |

Housing caps your population, so the larger rungs also name the age they need.

**Chain reward:** Title "The Founders" · game speed +300% for 180 ticks

---

## Scholar Chain

*Focus: research and scholars*

| Milestone | Condition | Reward |
|---|---|---|
| Knowledge Seeker | Accumulate 10,000 research: the work your scholars put into techs this run | +5% scholar output |
| First Research | Research 1 tech | research progress |
| Tech Pioneer | Research 22 techs | +5% research speed |
| Scholar's Haven | Build 3 Libraries and have 50 scholars | +10% scholar output |
| Renaissance Mind *(hidden)* | Research 83 techs (Renaissance Age) | +10% research speed |
| Tech Master *(hidden)* | Research 114 techs (Information Age) | +10% research speed, +5% all production |

Tech Master asks for more techs than the tree holds before the Information Age, so it completes only from that age on, and finishing this chain means playing past the Modern Age before you prestige. The tech-count milestones come within reach in these ages: Tech Pioneer in the Iron Age, Deep Thinker in the Renaissance Age, Philosophes in the Industrial Age and Renaissance Mind in the Electric Age.

The two milestones that pay "research progress" put it on the tech you are researching, up to half that tech's price. With no tech in progress it is lost.

Research speed counts only up to +100% in total, because a tech always takes at least one tick. Five milestones add +50% of it. The panel tags a reward that a limit holds back with "capped".

**Chain reward:** Title "The Enlightened" · game speed +300% for 180 ticks

---

## Builder Chain

*Focus: construction and wonders*

| Milestone | Condition | Reward |
|---|---|---|
| Stone Mason | Build 22 Stone Pits (Stone Age) | +10% stone production |
| Master Builder | Build 1,000 total structures | -5% building costs |
| Wonder Builder *(hidden)* | Build 1 wonder | +5% all production |
| Grand Architect *(hidden)* | Build 1,365 structures (Medieval Age) | -5% building costs, +5% all production |
| Wonder Collector *(hidden)* | Build 8 wonders (Colonial Age) | +10% all production |

Stone Pits can only be built in the Stone Age, so build them before you advance. Structures count every building you finish this run, wonders included (upgrades don't count, and the count starts over when you prestige).

The `-5% building costs` rewards here (Master Builder, Grand Architect) are real discounts on every building you queue, and the cost the `build` command lists already includes them. Five milestones cut costs, by 19% in all. Eleven techs cut them too, each taking its share of what is left, and with every milestone and tech the cut reaches about 42%. Costs never fall below 10% of the listed price. See [Buildings](buildings.md#build-cost-reductions).

**Chain reward:** Title "The Architects" · game speed +250% for 150 ticks

---

## Military Chain

*Focus: soldiers and combat*

| Milestone | Condition | Reward |
|---|---|---|
| First Soldiers | Train 5 soldiers (Iron Age) | +5% military power |
| War Machine | Train 250 soldiers (Iron Age) | +10% military power |
| Iron Legion *(hidden)* | Train 500 soldiers and build 10 Barracks (Classical Age) | +5% all production |
| Fortress State *(hidden)* | Build 13 Castle Keeps (Medieval Age) | +10% military power, +5% all production |
| Military Superpower *(hidden)* | Train 2,000 soldiers (Industrial Age) | +15% all production |

Soldier counts are the soldiers who took a post at your military buildings this run, not the size of your current army. Losing soldiers never takes the count back.

**Chain reward:** Title "The Conquerors" · game speed +250% for 150 ticks

The late tiers pay out all production rather than military power alone, so finishing the Military chain speeds up your whole economy as well as campaigns and combat.

One more military milestone sits outside the chain: train 100 soldiers and build 10 Barracks (Classical Age) for +5% military power.

---

## Trade Chain

*Focus: commerce and gold*

| Milestone | Condition | Reward |
|---|---|---|
| First Market | Build 1 Market (Bronze Age) | +25 gold |
| Market Town | Build 8 Markets (Iron Age) | +5% gold production |
| Caravan Network | Build 5 Trading Posts (Classical Age) | +10% gold production |
| Merchant Princes *(hidden)* | Build 12 Trading Posts and 4 Merchant Quarters (Medieval Age) | +5% all production |
| Trade Empire *(hidden)* | Build 20 Trading Posts and Merchant Quarters in any mix (Renaissance Age) | +10% all production |
| Maritime Empire *(hidden)* | Build 5 Harbors, 5 Ports and 2 Seaports (Modern Age) | +10% all production, +10% gold production |

**Chain reward:** Title "The Merchants" · game speed +250% for 150 ticks

Trade Empire counts Trading Posts and Merchant Quarters together, so upgrading posts into quarters keeps your progress. Maritime Empire needs Seaports, a Modern Age building, so this chain finishes in the Modern Age.

The last three milestones pay out all production, so completing the chain lifts every resource, not only your gold.

---

## Ancient Ages Chain

*Complete the first 5 age milestones (through the Renaissance) to finish this chain.*

| Milestone | Age required | Reward |
|---|---|---|
| Bronze Age Pioneer | Bronze Age | +30 iron, +30 gold |
| Iron Forged | Iron Age | +40 iron, +10% iron production |
| Classical Scholar | Classical Age | research progress, +10% scholar output |
| Medieval Lord | Medieval Age | +50 faith, +25 steel |
| Enlightened | Renaissance Age | +75 culture, +15% scholar output |

**Chain reward:** Title "The Ancients" · game speed +250% for 150 ticks

---

## Age milestones (beyond the chain)

*These age milestones are standalone rewards, hidden until you approach them.*

| Milestone | Age required | Reward |
|---|---|---|
| Colonial Power *(hidden)* | Colonial Age | +15% gold production, +10% expedition rewards |
| Industrial Revolution *(hidden)* | Industrial Age | +15% all production |
| Victorian Innovation *(hidden)* | Victorian Age | +10% all production |
| Electric Dawn *(hidden)* | Electric Age | +15% all production |
| Atomic Pioneer *(hidden)* | Atomic Age | +15% all production |
| Modern Era *(hidden)* | Modern Age | +20% all production |
| Information Pioneer *(hidden)* | Information Age | +20% scholar output |
| Digital Native *(hidden)* | Digital Age | +20% all production |
| Jacked In *(hidden)* | Cyberpunk Age | +15% worker output |
| Fusion Pioneer *(hidden)* | Fusion Age | +20% all production |
| Space Explorer *(hidden)* | Space Age | +20% all production, +20% expedition rewards |
| Star Voyager *(hidden)* | Interstellar Age | +20% all production |
| Galactic Emperor *(hidden)* | Galactic Age | +20% all production |
| Quantum Master *(hidden)* | Quantum Age | +20% all production |
| Transcended *(hidden)* | Transcendent Age | +50% all production |

Most of these rewards, and many of the chain rewards above, are all production. Every all-production bonus in the game but a tech's shares one pool. The pool applies in full, however large it grows (see [The all-production pool](resources.md#the-all-production-pool)). So every one of these rewards raises your output by its whole percentage, whenever you earn it. An example: with +320% earned, a +20% milestone takes the pool to +340%, twenty points more. The per-resource, research and expedition rewards are separate.

---

## Other milestones

These milestones sit outside the chains and don't count toward any chain title.

**Faith**

| Milestone | Condition | Reward |
|---|---|---|
| First Shrine | Build 1 Shrine | +10 faith |
| Devout Settlement | Build 16 Shrines (Stone Age) | +5% faith production |
| Temple City *(hidden)* | Build 18 Temples (Iron Age) | +10% faith production |
| Cathedral Age *(hidden)* | Build 10 Cathedrals (Medieval Age) | +10% faith production, +5% scholar output |

**Epoch and longevity** (the Milestones panel's Epoch group)

| Milestone | Condition | Reward |
|---|---|---|
| First Farmers | Build 3 Gathering Camps and reach tick 30 | +30 food |
| Survivor | Reach tick 26,000 in a run (the panel words it as about 14h 26m of play) | +5% all production |
| Enduring Civilization *(hidden)* | Reach tick 130,000 in a run (the panel words it as about 3 days of play) | +5% all production |
| Age Hopper | Reach the Classical Age and research 10 techs | +5% all production |
| Industrial Titan *(hidden)* | Stockpile 10K coal and 5K iron ore (Industrial Age) | +10% all production |
| Power Grid *(hidden)* | Build 13 Coal Plants and 10 Steam Turbines (Victorian Age) | +10% all production |

**Builder**

| Milestone | Condition | Reward |
|---|---|---|
| First Storehouse | Build 1 Stash | +20 food, +20 wood |
| Storage Network | Build 5 Storage Pits (Stone Age) | +5% food production |
| Granary Keeper | Build 5 Granaries (Bronze Age) | +5% food production |
| Lumber Operation | Build 25 Wood Camps (Stone Age) | +10% wood production |
| Early Builder | Build 250 structures (Bronze Age) | -3% building costs |
| Mining Syndicate | Build 22 Stone Pits and 10 Iron Mines (Iron Age) | +10% iron production |
| Forge Master | Build 15 Smithies (Iron Age) | +5% iron production, -3% building costs |
| Seasoned Builder | Build 500 structures (Iron Age) | -3% building costs |
| Wonder Empire *(hidden)* | Build 15 wonders (Digital Age) | +15% all production |

**Scholar**

| Milestone | Condition | Reward |
|---|---|---|
| Deep Thinker | Research 46 techs (Bronze Age) | +5% scholar output |
| Philosophes | Research 64 techs (Classical Age) | +5% research speed |
| Grand Library Built *(hidden)* | Build the Great Library (Classical Age) | +15% scholar output |
| Tech Ascendant *(hidden)* | Research all 203 techs (Transcendent Age) | +20% research speed |

**Settlement**

| Milestone | Condition | Reward |
|---|---|---|
| Urban Sprawl *(hidden)* | Population: 50,000 (Medieval Age) | +15% all production |
| Global City *(hidden)* | Population: 1,000,000 (Digital Age) | +20% all production |

**Trade**

| Milestone | Condition | Reward |
|---|---|---|
| Guildhall Master *(hidden)* | Build 10 Guildhalls (Renaissance Age) | +10% gold production |
| Colonial Trade Network *(hidden)* | Build 3 Ports and 5 Colonial Warehouses (Colonial Age) | +10% gold production, +10% expedition rewards |
| Gold Hoard *(hidden)* | Accumulate 1,000,000 gold (Renaissance Age) | +10% gold production |

---

## Can every milestone be done?

Yes. A milestone that doesn't name a later age fits inside a normal run: everything it asks for can be reached by the end of the Atomic Age, before a full run's prestige in the Modern Age (an early taste from the Medieval Age ends the run before some of them). A milestone that names a later age fits in that age. The game's test suite checks this against the game's numbers, so a balance change can't quietly strand one.

| Chain | Can be finished by |
|---|---|
| Settlement, Builder, Military, Ancient Ages | the Atomic Age (a normal run) |
| Trade | the Modern Age (Maritime Empire's Seaports) |
| Scholar | the Information Age (Tech Master's 114 techs, counted from the Information Age on) |

Some milestones count buildings you can only build in their own age: Stone Pits, Temples, Trading Posts, Coal Plants, Castle Keeps and others. Build them before you advance. A milestone that asks for a building also counts the later buildings of its line, so upgrading Shrines or Coal Plants to the newer building never takes that progress back.

---

## Tips

- The Scholar and Settlement chains carry the biggest boosts (+300% game speed for 180 ticks). Builder, Military, Trade and Ancient Ages each grant +250% for 150 ticks.
- The Milestones panel groups every milestone by category: Settlement, Builder, Scholar, Military, Trade, Faith, Epoch and Ages. A chain's progress sits on its group's heading. Each open milestone lists what it asks for, with a bar for each condition, and what it pays.
- A hidden milestone shows up once your overall progress toward it passes 50% (the average of its conditions). A hidden milestone that names an age also shows up as soon as you are in the age before it.
- A milestone that names an age past your next one stays hidden, whatever your progress. The section counts it in "+ N hidden milestones" until that age is your next one, or until you have reached it before on your account. The panel names an age only once it is your next one or you have reached it.
- Chain boosts add to any other game speed bonuses you hold, so two chains finishing close together overlap.
- The Military, Trade and Scholar capstones pay out all production bonuses, so completing any of those chains helps your whole economy, not only its own domain.
