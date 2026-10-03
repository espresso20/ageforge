# The 22 Ages

AgeForge spans 22 ages from primitive survival to transcendence. Each age unlocks new buildings, resources and worker domains, and adds a new district to your [skyline](map.md#skyline). Advancement requires meeting all resource and building requirements **and** completing the age's wonder.

## Wonder Requirement

Each age unlocks exactly one wonder building. You must **build that wonder** before you can advance to the next age, on top of the resource and building requirements listed below. The age progress bar shows a red `✗ Wonder required: <name>` notice until it is complete. See [Wonders](wonders.md) for build costs and instructions.

## Advancement Requirements

Every age can be finished, and an automated test checks this in every release. The test holds each age to these rules:

- Every building requirement names a building you can build in the age you are advancing **from**, never an older one you can no longer build.
- The last required copy of each building costs at most **half** the storage you can build by then (every storage building up to your current age at its build cap), so a full set of storage always leaves room to afford it.
- Each part of the age's wonder fits in that storage too.
- Every resource a requirement asks for has a source in the age itself, even for a player who skipped every building no earlier requirement asked for.
- A requirement in faith, food or culture is something a moderate economy makes within the age's target length, or can buy for a small share of it.

## How Long Each Age Takes

Each age is tuned to a target length in game time at 1x speed. Building output, prices, wonder costs and requirements are all sized to it: a staffed production building earns back the price of its first copy in a fraction of the age's target, and nothing takes longer to build than a sixth of it.

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

That is about a week (167 hours) from a fresh start to the Modern Age, where prestige unlocks. From the Bronze Age on, every age runs 2.6 times as long as it did on the earlier three-day curve, and the game's timers (events, raids, trade routes, expeditions, cooldowns) stretch with it, so each age holds as many of them as before. You don't have to sit through it: the game grants up to 24 hours of offline progress when you come back.

These are first-run lengths. On later runs an age a past run completed is **known ground** and runs faster: 2x after one completion, up to 4.2x after ten, with production and storage multiplied and build and research times divided by the same factor. Ages 6 or more behind the deepest age you have ever entered run at least 4x. See [Era Mastery](prestige.md#era-mastery).

## What Happens on Age Advance

When your civilization crosses into a new age:

1. **Carryover is capped.** Each resource the new age *uses as a build cost* is capped to roughly **8× the cheapest new-age starter building's cost** in that resource, enough to pay for a handful of opening buildings. If you were already below that cap, your stockpile carries over untouched. Anything above the cap is lost, so stockpiling before an advance doesn't pay. Resources the new age does **not** build with (mainly food) keep **10%** of what you had, and **faith is exempt** (it carries over in full). You enter each age with a small head start.
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

Every 3 ages you cross an **epoch boundary**. Epoch transitions trigger an event roll (see [Epochs](epochs.md)) and may shift your buildings' output resources.

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
- Worker domains: food, knowledge, faith

---

## Age 1: Stone Age 🪓

> *Stone tools and the first settled villages.*

| Requirement | Amount |
|---|---|
| Food | 1K |
| Wood | 1K |
| Knowledge | 150 |
| Huts | 10 |
| Story Circles | 5 |

**Unlocks:** Longhouse, Storage Pit, Forager Post, Woodcutter Camp, Stone Camp, Stone Pit, Elders' Hall, Standing Stones, War Camp, Great Monolith · **Resource:** Stone · **New domains:** lumber, masonry

---

## Age 2: Bronze Age 🛡

> *Metalworking begins.*

| Requirement | Amount |
|---|---|
| Food | 4K |
| Wood | 8K |
| Stone | 4K |
| Knowledge | 1.5K |
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
| Knowledge | 20K |
| Lumber Mills | 8 |
| Quarries | 8 |
| Scriptoria | 5 |

**Unlocks:** Townhouse, Granary, Field Works, Timber Yard, Marble Quarry, Agora, Temple, Hunting Lodge, Legion Fort, Trading Post, Ironworks, Smelter, Colosseum · **Resources:** Marble, Iron Ore · **New domains:** military, metallurgy

---

## Age 4: Classical Age 🏛

> *Great empires are built and philosophy flourishes.*

| Requirement | Amount |
|---|---|
| Stone | 130K |
| Iron | 26K |
| Gold | 14K |
| Knowledge | 35K |
| Hunting Lodges | 15 |
| Agoras | 12 |
| Trading Posts | 10 |

**Unlocks:** Villa, Classical Vault, Terrace Farm, Wood Workshop, Marble Works, Library, Oracle House, Military Academy, Merchant Quarter, Aqueduct, Forge, Amphitheater, Parthenon · **Resource:** Culture

---

## Age 5: Medieval Age 🏰

> *Kingdoms rise and feudalism takes hold.*

| Requirement | Amount |
|---|---|
| Stone | 220K |
| Iron | 53K |
| Gold | 35K |
| Knowledge | 88K |
| Merchant Quarters | 5 |
| Libraries | 15 |
| Military Academies | 15 |

**Unlocks:** Manor, Strongroom, Demesne, Sawmill, Stonemasons' Guild, Monastery Library, Cathedral, Castle Keep, Guildhall, Workshop, Ironmonger, Great Hall, Great Library · **Resource:** Steel

---

## Age 6: Renaissance Age 🎨

> *Art, science, and exploration flourish.*

| Requirement | Amount |
|---|---|
| Gold | 180K |
| Knowledge | 220K |
| Steel | 880 |
| Faith | 8.1K |
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
| Knowledge | 30M |
| Steel | 110K |
| Culture | 300K |
| Exchanges | 8 |
| Universities | 8 |
| Art Studios | 8 |

**Unlocks:** Settlement Block, Colonial Warehouse, Plantation, Coal Works, Deep Iron Mine, Natural Philosophy Hall, Mission, Fort, Port, Dockyard, Colonial Steelworks, Concert Hall, Grand Lighthouse

---

## Age 8: Industrial Age 🏭

> *Machines take over production.*

| Requirement | Amount |
|---|---|
| Steel | 470K |
| Gold | 3.8M |
| Knowledge | 3M |
| Plantations | 8 |
| Ports | 10 |

**Unlocks:** Tenement, Industrial Depot, Agricultural Works, Steam Colliery, Steam Mine, Research Institute, Church, Military Base, Stock Exchange, Integrated Steelworks, Steel Mill, Coal Plant, Opera House, Geographic Society, Crystal Palace · **Resource:** Oil

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

**Unlocks:** Row House, Victorian Vault, Mechanized Farm, Oil Derrick, Uranium Mine, Academy, Grand Cathedral, Garrison, Bank, Steam Works, Bessemer Plant, Steam Turbine, Grand Museum, Eiffel Tower · **Resource:** Electricity · **New domain:** energy

---

## Age 10: Electric Age ⚡

> *Electric power reaches every home.*

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

> *Nuclear power, for good and ill.*

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

Prestige becomes available at this age. Type `prestige confirm yes` to reset with permanent upgrades. See [Prestige System](prestige.md).

**Unlocks:** Tower Block, Modern Depot, Agritech Campus, Oil Platform, Titanium Mine, Think Tank, Meditation Center, Special Ops HQ, Investment Firm, Power Grid Hub, Titanium Smelter, Oil Refinery, TV Studio, Space Program · **Resources:** Data, Nanobots, Titanium Ore

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

**Unlocks:** Smart Complex (after Internet of Things), Info Vault, Smart Farm (after Internet of Things), Smart Refinery, Precision Mine, Innovation Hub, Digital Temple, Cyber Command, Venture Hub, Smart Grid Node, Aerospace Foundry, Microgrid Array, Server Farm, Media Center, Global Network · **New domain:** hacker

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

**Unlocks:** Megaplex, Digital Archive, Nano Farm, Bio Fabrication Lab, Nano Drill Complex, AI Research Lab, Cyber Shrine, Drone Warfare Center, Crypto Exchange, Neural Grid, Nano Alloy Plant, Quantum Battery Array, Data Center, VR Studio, World Simulation

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

> *Fusion brings cheap, clean energy.*

| Requirement | Amount |
|---|---|
| Electricity | 490B |
| Crypto | 25B |
| Data | 78B |
| Augmentation Foundries | 15 |
| Arcology Pods | 25 |
| Black Markets | 15 |

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

**Unlocks:** Orbital Habitat, Orbital Depot, Hydroponic Bay, Quantum Organic Extractor, Asteroid Crystal Mine, Deep Space Observatory, Orbital Sanctuary, Space Force Base, Asteroid Market, Launch Complex, Orbital Refinery, Solar Collector Array, Orbital Data Relay, Zero-G Gallery, Dyson Scaffold · **Resource:** Titanium · **New domain:** astronaut

---

## Age 18: Interstellar Age 🛸

> *Ships set out between the stars.*

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

> *The final age.*

| Requirement | Amount |
|---|---|
| Quantum Flux | 190T |
| Antimatter | 310T |
| Reality Academies | 20 |
| Reality Forges | 15 |
| Probability War Rooms | 15 |

The end of the progression. Prestige has been available since Modern Age (Age 12). Reaching the Transcendent Age gives the most prestige points, so it is the best time to prestige if you haven't already.

**Unlocks:** Singularity Core, Transcendent Nexus, Omniversal War Council, Omniversal Bazaar, Singularity Engine

```
prestige confirm yes
```

See [Prestige System](prestige.md) for details.
