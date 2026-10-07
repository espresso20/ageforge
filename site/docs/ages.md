# The 22 Ages

AgeForge spans 22 ages from primitive survival to transcendence. Each age unlocks new buildings, resources and worker domains, and adds a new district to your [skyline](map.md#skyline). Advancement requires meeting all resource and building requirements **and** completing the age's wonder.

<figure class="screen" data-screen="age-advance"><figcaption>The splash that greets an advance, here into the Bronze Age: what the new age unlocks, its wonder, and the buildings ready to upgrade.</figcaption></figure>

## Wonder Requirement

Each age unlocks exactly one wonder building. You must **build that wonder** before you can advance to the next age, on top of the resource and building requirements listed below. The age progress bar shows a red `✗ Wonder: <name>` notice until it is complete. See [Wonders](wonders.md) for build costs and instructions.

## Keystone Techs

From the Stone Age on, each age's wonder needs one technology of that age before it can be built: its **keystone**. The wonder's bank is open from the first tick of the age and fills while you research, by deposit and by overflow. Only `build <wonder>` waits for the tech. Since the next age needs the wonder, the keystone is the one tech an age asks of you, together with the techs it stands on. Everything else in the tree is your choice.

| Age | Wonder | Keystone tech | It stands on |
|---|---|---|---|
| Primitive | Sacred Grove | none | |
| Stone | Great Monolith | Stoneworking | Tool Making |
| Bronze | Stonehenge | Calendar | Ritual, Language |
| Iron | Colosseum | Mathematics | Primitive Writing, Language |
| Classical | Parthenon | Philosophy | Mathematics |
| Medieval | Great Library | Theology | Philosophy |
| Renaissance | Sistine Chapel | Patronage | Banking (Currency, Mathematics) |
| Colonial | Grand Lighthouse | Cartography | Navigation |
| Industrial | Crystal Palace | Industrialization | Steam Power (Steel Forging) |
| Victorian | Eiffel Tower | Mass Production | Industrialization |
| Electric | Hoover Dam | Power Distribution | Electrification |
| Atomic | Particle Accelerator | Nuclear Fission | Power Distribution, Chemical Engineering |
| Modern | Space Program | Satellite Technology | Rocketry, Advanced Electrics |
| Information | Global Network | Internet | Computers, Satellite Technology |
| Digital | World Simulation | Machine Learning | Internet, Cybersecurity |
| Cyberpunk | Neon Citadel | Cybernetics | Neural Interface |
| Fusion | Stellar Cradle | Fusion Power | Nuclear Fission, Cybernetics |
| Space | Dyson Scaffold | Orbital Mechanics | Rocketry, Plasma Physics |
| Interstellar | Warp Nexus | Warp Drive | Space Mining, Zero-G Manufacturing |
| Galactic | Cosmic Beacon | Galactic Navigation | Warp Drive, Stellar Engineering |
| Quantum | Reality Anchor | Quantum Mechanics | Antimatter Synthesis |
| Transcendent | Singularity Core | Transcendence | Reality Manipulation |

The Sacred Grove needs no tech, so nothing stands between a new game and the Stone Age.

The age progress bar shows `✗ Keystone: <tech>` beside the wonder until the tech is researched, the Wonders panel and `wonder` show the keystone and whether it is done, and the Research panel marks every keystone with a ★. A wonder in your [build plan](plan.md) waits for its keystone and starts once it is researched.

**No age asks for knowledge.** The requirements from the Stone Age to the Industrial Age used to include a knowledge amount, which had you saving the knowledge your research wanted. They are gone: an age's knowledge goes to research, and the keystone is what the advance needs of it. See [how research is priced](technologies.md#how-research-is-priced).

**A game saved before keystones** keeps the age it was in exempt: there the wonder builds without its keystone, and the lock starts when you next advance. The log says so once, on the first load. Every tech you researched stays researched and every wonder you built stays built.

## Advancement Requirements

Every age can be finished, and an automated test checks this in every release. The test holds each age to these rules:

- Every building requirement names a building you can build in the age you are advancing **from**, never an older one you can no longer build.
- The last required copy of each building costs at most **half** the storage you can build by then (every storage building up to your current age at its build cap), so a full set of storage always leaves room to afford it.
- Each part of the age's wonder fits in that storage too.
- The wonder's keystone tech, and every tech it stands on, can be researched by the wonder's own age, and each one's knowledge cost fits the knowledge storage you can build in that age with room to spare.
- The keystone is affordable well inside the age: the knowledge it costs (with the techs of its own age it stands on), the time to research them and the time to build the wonder fit inside the age's target length together, with at least a fifth of the age to spare.
- The cheapest tech of an age fits the knowledge storage you are sure to enter the age with, or does within four of the age's storage buildings (four in the Renaissance, two at most elsewhere).
- Every resource a requirement asks for has a source in the age itself, even for a player who skipped every building no earlier requirement asked for.
- A requirement in faith, food or culture is something a moderate economy makes within the age's target length, or can buy for a small share of it.

## How Long Each Age Takes

Each age is tuned to a target length in real time. Building output, prices, wonder costs and requirements are all sized to it: a staffed production building earns back the price of its first copy in a fraction of the age's target, and nothing takes longer to build than a sixth of it.

| Age | Target | Age | Target |
|---|---|---|---|
| Primitive | 15m | Modern | 31h 12m |
| Stone | 45m | Information | 36h 24m |
| Bronze | 3h 54m | Digital | 41h 36m |
| Iron | 6h 30m | Cyberpunk | 46h 48m |
| Classical | 9h 6m | Fusion | 52h |
| Medieval | 11h 42m | Space | 57h 12m |
| Renaissance | 15h 36m | Interstellar | 62h 24m |
| Colonial | 18h 12m | Galactic | 62h 24m |
| Industrial | 20h 48m | Quantum | 62h 24m |
| Victorian | 23h 24m | | |
| Electric | 26h | | |
| Atomic | 31h 12m | | |

That is about a week (167 hours) from a fresh start to the Modern Age, where a full run's prestige comes. Prestige opens earlier, at the Medieval Age (about 20 hours in), as an early taste that pays little. From the Bronze Age on, every age runs 2.6 times as long as it did on the earlier three-day curve, and the game's timers (events, raids, trade routes, expeditions, cooldowns) stretch with it, so each age holds as many of them as before. You don't have to sit through it: the game keeps playing while you are away, for up to 24 hours at 50% of your normal production.

These are first-run lengths. On later runs an age a past run completed is **known ground** and runs faster: 2x after one completion, up to 4.2x after ten, with production and storage multiplied and build and research times divided by the same factor. Ages 6 or more behind the deepest age you have ever entered run at least 4x. See [Era Mastery](prestige.md#era-mastery).

## What Happens on Age Advance

When your civilization crosses into a new age:

1. **Carryover is capped.** Each resource the new age *uses as a build cost* is capped to roughly **8× the cheapest new-age starter building's cost** in that resource, enough to pay for a handful of opening buildings. If you were already below that cap, your stockpile carries over untouched. Anything above the cap is lost, so stockpiling before an advance doesn't pay. Resources the new age does **not** build with (mainly food) keep **10%** of what you had, and **faith is exempt** (it carries over in full). Whatever the [build plan](plan.md) had banked from overflow goes back to your stores first, so it is trimmed like the rest. You enter each age with a small head start.
2. **Buildings whose next tier arrives in the new age are marked as upgradeable.** They are not converted automatically. Storage buildings never upgrade: every stash, storage pit or vault you built keeps counting toward your storage for the rest of the run. Each upgradeable building shows a gold hint in the Economy panel:
   ```
   ↑ Upgrade available → Forager Post  type: upgrade gathering_camp
   ```
3. **Production continues at the old rate** until you upgrade. Leaving buildings un-upgraded costs nothing; the reason to upgrade is the new tier's higher output.
4. **New-tier buildings are available at once** in the Buildings panel, so you can build fresh copies of the new tier right away.
5. **The age's speed is set.** On known ground the log says how much faster the age runs (`Known ground: the Bronze Age runs 2.4x faster (mastery 2).`). Stepping from known ground onto an age no run has completed logs that it runs at 1x; storage shrinks back with the speed, but stock already above the new cap stays until you spend it (see [The grace rule](prestige.md#the-grace-rule)).

### Upgrading buildings after an advance

Use the `upgrade` command to convert old copies to new ones, paying only the cost delta (new build cost minus 50% of the old building's value):

```
upgrade gathering_camp       # upgrade all Gathering Camps to Forager Posts
upgrade gathering_camp 3     # upgrade exactly 3 copies
```

Workers transfer automatically when all copies of a building are upgraded. For partial upgrades, workers stay on the old building. An upgrade never takes the new building past its max count; copies beyond that stay as they are.

See [Buildings: Building Upgrades](buildings.md#building-upgrades) for the full cost formula, strategic timing advice, and worker transfer rules.

## Epoch Overview

Ages are grouped into epochs of 3 (the Cosmic Era has 4), and the advance into a new epoch crosses an **epoch boundary**. An epoch transition triggers an event roll (see [Epochs](epochs.md)); the stockpile trim above applies to it like any other advance.

| Ages | Epoch | Symbol |
|------|-------|--------|
| Primitive (0), Stone (1), Bronze (2) | Stone Era | ◈ |
| Iron (3), Classical (4), Medieval (5) | Iron Era | ⚔ |
| Renaissance (6), Colonial (7), Industrial (8) | Steel Era | ⚙ |
| Victorian (9), Electric (10), Atomic (11) | Electric Era | ⚡ |
| Modern (12), Information (13), Digital (14) | Digital Era | ▣ |
| Cyberpunk (15), Fusion (16), Space (17) | Neon Era | ◉ |
| Interstellar (18), Galactic (19), Quantum (20), Transcendent (21) | Cosmic Era | ✦ |

---

## Age 0: Primitive Age 🪨

> *Survival. Nothing but your hands and wits.*

Starting age. No requirements.

**Unlocks:**
- Buildings: Hut, Stash, Gathering Camp, Wood Camp, Story Circle, Shrine, Sacred Grove
- Resources: Food, Wood, Knowledge, Faith
- Worker domains: food, lumber, knowledge, faith

---

## Age 1: Stone Age 🪓

> *Stone tools and the first permanent camps.*

| Requirement | Amount |
|---|---|
| Food | 1K |
| Wood | 1K |
| Huts | 10 |
| Story Circles | 5 |

**Unlocks:** Longhouse, Storage Pit, Forager Post, Woodcutter Camp, Stone Camp, Stone Pit, Elders' Hall, Standing Stones, War Camp, Great Monolith · **Resource:** Stone · **New domains:** masonry, military

---

## Age 2: Bronze Age 🛡

> *Metalworking arrives, and with it trade.*

| Requirement | Amount |
|---|---|
| Food | 4K |
| Wood | 8K |
| Stone | 4K |
| Longhouses | 15 |
| Stone Pits | 5 |
| Elders' Halls | 5 |

**Unlocks:** House, Warehouse, Farm, Lumber Mill, Quarry, Scriptorium, Altar, Barracks, Market, Smithy, Stonehenge · **Resources:** Iron, Gold · **New domains:** trade, engineering

---

## Age 3: Iron Age ⚔️

> *Iron tools and weapons spread.*

| Requirement | Amount |
|---|---|
| Food | 80K |
| Wood | 40K |
| Stone | 16K |
| Iron | 8K |
| Lumber Mills | 8 |
| Quarries | 8 |
| Scriptoria | 5 |

**Unlocks:** Townhouse, Granary, Field Works, Timber Yard, Marble Quarry, Agora, Temple, Hunting Lodge, Legion Fort, Trading Post, Ironworks, Smelter, Colosseum · **Resources:** Marble, Iron Ore, Soldiers · **New domain:** metallurgy

---

## Age 4: Classical Age 🏛

> *Great empires are built and philosophy flourishes.*

| Requirement | Amount |
|---|---|
| Stone | 130K |
| Iron | 26K |
| Gold | 14K |
| Hunting Lodges | 15 |
| Agoras | 12 |
| Trading Posts | 10 |

**Unlocks:** Villa, Classical Vault, Terrace Farm, Wood Workshop, Marble Works, Library, Oracle House, Military Academy, Merchant Quarter, Aqueduct, Forge, Amphitheater, Parthenon, Cultural Obelisk · **Resource:** Culture

---

## Age 5: Medieval Age 🏰

> *Kingdoms rise and feudalism takes hold.*

| Requirement | Amount |
|---|---|
| Stone | 220K |
| Iron | 53K |
| Gold | 35K |
| Merchant Quarters | 5 |
| Libraries | 15 |
| Military Academies | 15 |

Prestige becomes available at this age. A prestige from here to the Atomic Age is an early taste: it pays 9 points from the Medieval Age, enough for the Plan Template, while a run to the Modern Age pays 120. Type `prestige` to see what a prestige pays now and from the next age. See [Prestige System](prestige.md).

**Unlocks:** Manor, Strongroom, Demesne, Sawmill, Stonemason's Guild, Monastery Library, Cathedral, Castle Keep, Guildhall, Workshop, Ironmonger, Great Hall, Great Library, Grand Amphitheater · **Resource:** Steel

---

## Age 6: Renaissance Age 🎨

> *Art, science, and exploration flourish.*

| Requirement | Amount |
|---|---|
| Gold | 180K |
| Steel | 880 |
| Faith | 9.5K |
| Monastery Libraries | 5 |
| Guildhalls | 10 |
| Castle Keeps | 5 |

**Unlocks:** Estate, Renaissance Vault, Market Garden, Coal Mine, Iron Mine, University, Basilica, Fortress, Exchange, Mill, Foundry, Art Studio, Sistine Chapel · **Resource:** Coal

---

## Age 7: Colonial Age ⚓

> *Exploration and trade span the globe.*

| Requirement | Amount |
|---|---|
| Gold | 710K |
| Steel | 110K |
| Culture | 300K |
| Exchanges | 8 |
| Universities | 8 |
| Art Studios | 8 |

**Unlocks:** Settlement Block, Colonial Warehouse, Plantation, Coal Works, Deep Iron Mine, Natural Philosophy Hall, Mission, Fort, Port, Harbor, Dockyard, Colonial Steelworks, Concert Hall, Embassy, Grand Lighthouse

---

## Age 8: Industrial Age 🏭

> *Machines take over production.*

| Requirement | Amount |
|---|---|
| Steel | 470K |
| Gold | 3.8M |
| Plantations | 8 |
| Ports | 10 |

**Unlocks:** Tenement, Industrial Depot, Agricultural Works, Steam Colliery, Steam Mine, Research Institute, Church, Military Base, Stock Exchange, Harbor Authority, Integrated Steelworks, Steel Mill, Coal Plant, Opera House, Grand Embassy, Geographic Society, Crystal Palace, Eternal Library · **Resource:** Oil · **New domain:** energy

---

## Age 9: Victorian Age 🎩

> *Steam and innovation drive progress.*

| Requirement | Amount |
|---|---|
| Steel | 2.4M |
| Gold | 15M |
| Steel Mills | 5 |
| Integrated Steelworks | 5 |
| Tenements | 30 |

**Unlocks:** Row House, Victorian Vault, Mechanized Farm, Oil Derrick, Uranium Mine, Academy, Grand Cathedral, Garrison, Bank, Steam Works, Bessemer Plant, Steam Turbine, Grand Museum, Eiffel Tower · **Resource:** Electricity

---

## Age 10: Electric Age ⚡

> *Electric light and power reach daily life.*

| Requirement | Amount |
|---|---|
| Steel | 11M |
| Oil | 3.3M |
| Electricity | 1.1M |
| Steam Turbines | 10 |
| Academies | 10 |
| Bessemer Plants | 10 |

**Unlocks:** Apartment Block, Electric Warehouse, Industrial Farm, Oil Field, Nuclear Extraction Plant, Physics Laboratory, Revival Hall, Command Post, Financial District, Power Station, Electric Arc Furnace, Dynamo Hall, Radio Station, Hoover Dam

---

## Age 11: Atomic Age ☢️

> *Nuclear power, for better and worse.*

| Requirement | Amount |
|---|---|
| Steel | 110M |
| Electricity | 12M |
| Oil | 7.7M |
| Electric Arc Furnaces | 15 |
| Power Stations | 15 |
| Physics Laboratories | 15 |

**Unlocks:** Housing Project, Atomic Vault, Agricultural Complex, Petroleum Refinery, Uranium Processing Works, Research Campus, Spiritual Center, Bunker Complex, Corporate HQ, Nuclear Plant (after Civilian Reactors), Advanced Alloy Plant, Breeder Reactor, Cinema, Particle Accelerator · **Resource:** Uranium

---

## Age 12: Modern Age 🌐

> *Technology and innovation define the era.*

| Requirement | Amount |
|---|---|
| Electricity | 33M |
| Uranium | 6.9M |
| Steel | 470M |
| Breeder Reactors | 15 |
| Bunker Complexes | 15 |
| Research Campuses | 15 |

From this age a prestige counts as a full run: it pays 120 points, against 93 from the Atomic Age before it. Type `prestige confirm yes` to reset for prestige points to spend on the legacy kit. See [Prestige System](prestige.md).

**Unlocks:** Tower Block, Modern Depot, Agritech Campus, Oil Platform, Titanium Mine, Think Tank, Meditation Center, Special Ops HQ, Investment Firm, Seaport, Power Grid Hub, Titanium Smelter, Oil Refinery, TV Studio, Space Program, Nano Foundry, Monument of Ages · **Resources:** Data, Nanobots

---

## Age 13: Information Age 📡

> *The Internet connects the world.*

| Requirement | Amount |
|---|---|
| Electricity | 660M |
| Data | 69M |
| Gold | 1.3B |
| Think Tanks | 20 |
| Tower Blocks | 30 |
| Oil Refineries | 15 |

**Unlocks:** Smart Complex (after Internet of Things), Info Vault, Smart Farm (after Internet of Things), Smart Refinery, Precision Mine, Innovation Hub, Digital Temple, Cyber Command, Venture Hub, Container Terminal, Smart Grid Node, Aerospace Foundry, Microgrid Array, Server Farm, Media Center, Global Network · **New domain:** hacker

---

## Age 14: Digital Age 💻

> *Full digitization of civilization.*

| Requirement | Amount |
|---|---|
| Data | 3.1B |
| Electricity | 20B |
| Server Farms | 10 |
| Media Centers | 15 |
| Innovation Hubs | 15 |

**Unlocks:** Megaplex, Digital Archive, Nano Farm, Bio Fabrication Lab, Nano Drill Complex, AI Research Lab, Cyber Shrine, Drone Warfare Center, Crypto Exchange, Logistics Hub, Neural Grid, Nano Alloy Plant, Quantum Battery Array, Data Center, VR Studio, World Simulation

---

## Age 15: Cyberpunk Age 🤖

> *Neon lights and cybernetic augmentation.*

| Requirement | Amount |
|---|---|
| Data | 160B |
| Electricity | 980B |
| AI Research Labs | 15 |
| Data Centers | 15 |
| Neural Grids | 15 |

**Unlocks:** Arcology Pod, Cyber Vault, Vat Farm, Nanobot Vat, Dark Crystal Mine, Neuro Research Center, Neon Sanctuary, Combat Aug Center, Black Market, Augmentation Foundry, Dark Matter Refinery, Dark Energy Tap, Cyber Hub, Holographic Theater (after Holography), Neon Citadel · **Resources:** Crypto, Dark Matter Crystals

---

## Age 16: Fusion Age 🔬

> *Fusion brings clean, plentiful energy.*

| Requirement | Amount |
|---|---|
| Electricity | 490B |
| Crypto | 25B |
| Data | 78B |
| Augmentation Foundries | 15 |
| Arcology Pods | 25 |
| Black Market Hubs | 15 |

**Unlocks:** Habitat Ring, Fusion Vault, Bio Reactor Farm, Molecular Synthesizer, Exotic Mineral Extractor, Theoretical Institute, Quantum Chapel, Plasma Command, Energy Exchange (after Maglev Transit), Fusion Reactor, Exotic Matter Forge, Tokamak Array, Quantum Server Farm, Neural Art Complex, Stellar Cradle · **Resource:** Plasma

---

## Age 17: Space Age 🚀

> *Orbital expansion begins.*

| Requirement | Amount |
|---|---|
| Plasma | 63B |
| Electricity | 2.4T |
| Data | 390B |
| Fusion Reactors | 10 |
| Tokamak Arrays | 10 |
| Plasma Commands | 10 |

**Unlocks:** Orbital Habitat, Orbital Depot, Hydroponic Bay, Quantum Organic Extractor, Asteroid Crystal Mine, Deep Space Observatory, Orbital Sanctuary, Space Force Base, Asteroid Market, Launch Complex, Orbital Refinery, Solar Collector Array, Orbital Data Relay, Zero G Gallery, Dyson Scaffold · **Resources:** Titanium, Titanium Ore

---

## Age 18: Interstellar Age 🛸

> *Ships cross the space between stars.*

| Requirement | Amount |
|---|---|
| Titanium | 130B |
| Plasma | 310B |
| Launch Complexes | 10 |
| Orbital Habitats | 15 |
| Solar Collector Arrays | 10 |

**Unlocks:** Generation Ship, Stellar Vault, Protein Synthesizer, Reality Matter Weaver, Stellar Core Drill, Xenology Institute, Void Monastery, Fleet Command, Galactic Trade Hub, Warp Drive Plant, Antimatter Forge, Pulsar Tap, Galactic Network Node, Cultural Beacon, Warp Nexus · **Resource:** Dark Matter

---

## Age 19: Galactic Age 🌌

> *Galactic civilization spans the cosmos.*

| Requirement | Amount |
|---|---|
| Dark Matter | 250B |
| Titanium | 630B |
| Warp Drive Plants | 15 |
| Generation Ships | 30 |
| Antimatter Forges | 15 |

**Unlocks:** Dyson Sphere Habitat, Galactic Vault, Matter Converter, Cosmic Organic Works, Neutron Star Mine, Cosmic Research Station, Stellar Shrine, Stellar Armada HQ, Stellar Exchange, Dyson Assembly, Stellar Metallurgy, Quasar Tap, Consciousness Upload Hub, Civilization Archive, Cosmic Beacon · **Resource:** Antimatter

---

## Age 20: Quantum Age ⚛️

> *Reality bends to quantum mastery.*

| Requirement | Amount |
|---|---|
| Antimatter | 6.3T |
| Dark Matter | 13T |
| Stellar Exchanges | 15 |
| Stellar Metallurgy | 15 |
| Dyson Sphere Habitats | 30 |

**Unlocks:** Reality Fold, Quantum Vault, Quantum Cultivator, Reality Harvester, Reality Excavator, Reality Academy, Transcendence Hall, Probability War Room, Probability Market, Reality Forge, Quantum Metal Works, Zero Point Generator, Reality Processor, Reality Art Engine, Reality Anchor · **Resource:** Quantum Flux

---

## Age 21: Transcendent Age ✨

> *The last age. There is nowhere further to go.*

| Requirement | Amount |
|---|---|
| Quantum Flux | 190T |
| Antimatter | 310T |
| Reality Academies | 20 |
| Reality Forges | 15 |
| Probability War Rooms | 15 |

The end of the progression. Prestige has been available since the Medieval Age (Age 5), and a full run since the Modern Age (Age 12). A prestige from the Transcendent Age pays the most points a single run can, 3,279: each Cosmic Era age the run completes adds 729 (see [Prestige Points Formula](prestige.md#prestige-points-formula)).

**Unlocks:** Singularity Core, Transcendent Nexus, Omniversal War Council, Omniversal Bazaar, Singularity Engine

```
prestige confirm yes
```

See [Prestige System](prestige.md) for details.
