# The 22 Ages

AgeForge spans 22 ages from primitive survival to transcendence. Each age unlocks new buildings, resources, and worker domains — and gives your settlement a distinct look on the [City Map](city-map.md), which re-skins from thatch huts to neon megablocks as you advance. Advancement requires meeting all resource and building requirements **and** completing the age's wonder.

## Wonder Requirement

Each age unlocks exactly one wonder building. You must **build that wonder** before you can advance to the next age — it is a hard requirement alongside the resource and building thresholds listed below. The age progress bar will show a red `✗ Wonder required: <name>` notice until it is complete. See [Wonders](wonders.md) for build costs and instructions.

## Advancement Requirements

Every age can be finished, and an automated check (the "Gate Covenant") keeps it that way in every release. Every building requirement names a building you can build in the age you are advancing **from**, never an older one you can no longer build. And the last required copy of each building always costs at most **half** the storage you can build by then (every storage building up to your current age at its build cap), so a full set of storage always leaves room to afford it. Each part of the age's wonder fits in that storage too, every resource a gate needs has a source in the age itself, and a requirement in faith, food or culture is something a moderate economy makes within the age's target length (or can buy for a small share of it).

## How Long Each Age Takes

Each age is tuned to a target length in game time at 1x speed. Building output, prices, wonder costs and requirements are all sized to it: a staffed production building earns back the price of its first copy in a fraction of the age's target, and nothing takes longer to build than a sixth of it.

| Age | Target | Age | Target |
|---|---|---|---|
| Primitive | 15m | Modern | 12h |
| Stone | 45m | Information | 14h |
| Bronze | 1.5h | Digital | 16h |
| Iron | 2.5h | Cyberpunk | 18h |
| Classical | 3.5h | Fusion | 20h |
| Medieval | 4.5h | Space | 22h |
| Renaissance | 6h | Interstellar | 24h |
| Colonial | 7h | Galactic | 24h |
| Industrial | 8h | Quantum | 24h |
| Victorian | 9h | | |
| Electric | 10h | | |
| Atomic | 12h | | |

That is about three days from a fresh start to the Modern Age, where prestige unlocks. You don't have to sit through it: the game grants up to 24 hours of offline progress when you come back.

## What Happens on Age Advance

When your civilization crosses into a new age:

1. **Carryover is capped, not wiped.** Each resource the new age actually *uses as a build cost* is capped to roughly **8× the cheapest new-age starter building's cost** in that resource — enough to bankroll a handful of opening buildings, no more. If you were already below that cap, your stockpile carries over untouched. Hoarding past the cap is wasted, so there's no reward for over-banking before a transition. Resources the new age does **not** build with (mainly food) keep the old **10%** carryover, and **Faith is exempt entirely** (it accumulates across ages). Net effect: you enter each age with a small head start, not a full war chest.
2. **Buildings with a next-tier equivalent receive a pending upgrade marker.** They are NOT automatically transformed. Storage buildings are the exception: they never upgrade, and every stash, storage pit or vault you built keeps counting toward your storage for the rest of the run. Each upgradeable building shows a gold hint in the Economy tab:
   ```
   ↑ Upgrade available → Farm   type: upgrade gathering_camp
   ```
3. **Production continues at the old rate** until you manually upgrade. There is no production penalty for leaving buildings pending — the incentive to upgrade is the higher output of the new tier.
4. **New-tier buildings become available** in the build menu immediately — you can start building fresh copies of the new tier right away.

### Upgrading buildings after an advance

Use the `upgrade` command to convert old copies to new ones, paying only the cost delta (new build cost minus 50% of the old building's value):

```
upgrade gathering_camp       — upgrade all Gathering Camps to Forager Posts
upgrade gathering_camp 3     — upgrade exactly 3 copies
```

Workers transfer automatically when all copies of a building are upgraded. For partial upgrades, workers stay on the old building. An upgrade never takes the new building past its max count; copies beyond that stay as they are.

See [Buildings — Building Upgrades](buildings.md#building-upgrades) for the full cost formula, strategic timing advice, and worker transfer rules.

## Epoch Overview

Every 3 ages you cross an **epoch boundary**. Epoch transitions trigger an event roll (see [Epochs](epochs.md)) and may shift your buildings' output resources.

| Ages | Epoch | Symbol |
|------|-------|--------|
| 0 — Primitive, 1 — Stone, 2 — Bronze | Stone Era | ◈ |
| 3 — Iron, 4 — Classical, 5 — Medieval | Iron Era | ⚔ |
| 6 — Renaissance, 7 — Colonial, 8 — Industrial | Steel Era | ⚙ |
| 9 — Victorian, 10 — Electric, 11 — Atomic | Electric Era | ⚡ |
| 12 — Modern, 13 — Information, 14 — Digital | Digital Era | ▣ |
| 15 — Cyberpunk, 16 — Fusion, 17 — Space | Neon Era | ◉ |
| 18 — Interstellar, 19 — Galactic, 20 — Quantum, 21 — Transcendent | Cosmic Era | ✦ |

---

## Age 0 — Primitive Age 🪨

> *Survival. Nothing but your hands and wits.*

Starting age. No requirements.

**Unlocks:**
- Buildings: Hut, Stash, Gathering Camp, Wood Camp, Story Circle, Shrine, Sacred Grove
- Resources: Food, Wood, Knowledge, Faith
- Worker domains: food, knowledge

---

## Age 1 — Stone Age 🪓

> *Tools of stone change everything.*

| Requirement | Amount |
|---|---|
| Food | 1,000 |
| Wood | 1,000 |
| Knowledge | 150 |
| Huts | 10 |
| Story Circles | 5 |

**Unlocks:** Longhouse, Storage Pit, Forager Post, Woodcutter Camp, Stone Camp, Stone Pit, Elders' Hall, Standing Stones, War Camp, Great Monolith · **Resource:** Stone

---

## Age 2 — Bronze Age 🛡

> *Discovery of metalworking changes everything.*

| Requirement | Amount |
|---|---|
| Food | 4,000 |
| Wood | 8,000 |
| Stone | 4,000 |
| Knowledge | 1,500 |
| Longhouses | 15 |
| Stone Pits | 5 |
| Elders' Halls | 5 |

**Unlocks:** House, Warehouse, Farm, Lumber Mill, Quarry, Scriptorium, Altar, Barracks, Market, Smithy, Stonehenge · **Resources:** Iron, Gold · **New domain:** trade

---

## Age 3 — Iron Age ⚔️

> *Iron tools and weapons transform society.*

| Requirement | Amount |
|---|---|
| Food | 80,000 |
| Wood | 40,000 |
| Stone | 16,000 |
| Iron | 8,000 |
| Knowledge | 20,000 |
| Lumber Mills | 8 |
| Quarries | 8 |
| Scriptoria | 5 |

**Unlocks:** Townhouse, Granary, Field Works, Timber Yard, Marble Quarry, Agora, Temple, Hunting Lodge, Legion Fort, Trading Post, Ironworks, Smelter, Colosseum · **Resources:** Marble, Iron Ore · **New domains:** military, metallurgy

---

## Age 4 — Classical Age 🏛

> *Great empires are built and philosophy flourishes.*

| Requirement | Amount |
|---|---|
| Stone | 130,000 |
| Iron | 26,000 |
| Gold | 14,000 |
| Knowledge | 35,000 |
| Hunting Lodges | 15 |
| Agoras | 12 |
| Trading Posts | 10 |

**Unlocks:** Villa, Classical Vault, Estate Farm, Wood Workshop, Marble Works, Library, Oracle House, Military Academy, Merchant Quarter, Aqueduct, Forge, Amphitheater, Parthenon · **Resource:** Culture

---

## Age 5 — Medieval Age 🏰

> *Kingdoms rise and feudalism takes hold.*

| Requirement | Amount |
|---|---|
| Stone | 220,000 |
| Iron | 53,000 |
| Gold | 35,000 |
| Knowledge | 88,000 |
| Merchant Quarters | 5 |
| Libraries | 15 |
| Military Academies | 15 |

**Unlocks:** Manor, Keep, Demesne, Sawmill, Stonemasons' Guild, Monastery Library, Cathedral, Castle Keep, Guildhall, Workshop, Ironmonger, Great Hall, Great Library · **Resource:** Steel · **New domain:** faith

---

## Age 6 — Renaissance Age 🎨

> *Art, science, and exploration flourish.*

| Requirement | Amount |
|---|---|
| Gold | 180,000 |
| Knowledge | 220,000 |
| Steel | 880 |
| Faith | 8,100 |
| Monastery Libraries | 5 |
| Guildhalls | 10 |
| Castle Keeps | 5 |

**Unlocks:** Estate, Renaissance Vault, Market Garden, Coal Mine, Iron Mine, University, Basilica, Fortress, Exchange, Mill, Foundry, Art Studio, Sistine Chapel · **Resource:** Coal

---

## Age 7 — Colonial Age ⚓

> *Exploration and trade span the globe.*

| Requirement | Amount |
|---|---|
| Gold | 710,000 |
| Knowledge | 940,000 |
| Steel | 110,000 |
| Culture | 300,000 |
| Exchanges | 8 |
| Universities | 8 |
| Art Studios | 8 |

**Unlocks:** Settlement Block, Colonial Warehouse, Plantation, Coal Works, Deep Iron Mine, Natural Philosophy Hall, Mission, Fort, Port, Dockyard, Iron Works, Concert Hall, Grand Lighthouse

---

## Age 8 — Industrial Age 🏭

> *Machines revolutionize production.*

| Requirement | Amount |
|---|---|
| Steel | 470,000 |
| Gold | 3,800,000 |
| Knowledge | 3,000,000 |
| Plantations | 8 |
| Ports | 10 |

**Unlocks:** Tenement, Industrial Depot, Agricultural Works, Steam Coal Plant, Steam Mine, Research Institute, Church, Military Base, Stock Exchange, Iron Works Complex, Steel Mill, Coal Plant, Opera House, Geographic Society, Crystal Palace · **Resource:** Oil

---

## Age 9 — Victorian Age 🎩

> *Steam and innovation drive progress.*

| Requirement | Amount |
|---|---|
| Steel | 2,400,000 |
| Gold | 15,000,000 |
| Steel Mills | 5 |
| Iron Works Complexes | 5 |
| Tenements | 30 |

**Unlocks:** Row House, Victorian Vault, Mechanized Farm, Oil Derrick, Uranium Mine, Academy, Grand Cathedral, Garrison, Bank, Steam Works, Bessemer Plant, Steam Turbine, Grand Museum, Eiffel Tower · **Resource:** Electricity · **New domains:** engineering, energy

---

## Age 10 — Electric Age ⚡

> *Electrification transforms daily life.*

| Requirement | Amount |
|---|---|
| Steel | 11,000,000 |
| Oil | 3,300,000 |
| Electricity | 1,100,000 |
| Steam Turbines | 10 |
| Academies | 10 |
| Bessemer Plants | 10 |

**Unlocks:** Apartment Block, Electric Warehouse, Industrial Farm, Oil Field, Nuclear Extraction Plant, Physics Laboratory, Revival Hall, Command Post, Financial District, Power Station, Electric Arc Furnace, Power Generator, Radio Station, Hoover Dam

---

## Age 11 — Atomic Age ☢️

> *Nuclear power unleashes terrifying potential.*

| Requirement | Amount |
|---|---|
| Steel | 110,000,000 |
| Electricity | 12,000,000 |
| Oil | 7,700,000 |
| Electric Arc Furnaces | 15 |
| Power Stations | 15 |
| Physics Laboratories | 15 |

**Unlocks:** Housing Project, Atomic Vault, Agricultural Complex, Petroleum Refinery, Uranium Processing Works, Research Campus, Spiritual Center, Bunker Complex, Corporate HQ, Nuclear Plant, Advanced Alloy Plant, Nuclear Reactor, Cinema, Particle Accelerator · **Resource:** Uranium

---

## Age 12 — Modern Age 🌐

> *Technology and innovation define the era.*

| Requirement | Amount |
|---|---|
| Electricity | 33,000,000 |
| Uranium | 6,900,000 |
| Steel | 470,000,000 |
| Nuclear Reactors | 15 |
| Bunker Complexes | 15 |
| Research Campuses | 15 |

Prestige becomes available at this age. Type `prestige confirm yes` to reset with permanent upgrades. See [Prestige System](prestige.md).

**Unlocks:** Tower Block, Modern Depot, Agri-Complex, Oil Platform, Titanium Mine, Think Tank, Meditation Center, Special Ops HQ, Investment Firm, Power Grid Hub, Titanium Smelter, Oil Refinery, TV Studio, Space Program · **Resources:** Data, Nanobots, Titanium Ore

---

## Age 13 — Information Age 📡

> *The Internet connects the world.*

| Requirement | Amount |
|---|---|
| Electricity | 660,000,000 |
| Data | 69,000,000 |
| Gold | 1,300,000,000 |
| Think Tanks | 20 |
| Tower Blocks | 30 |
| Oil Refineries | 15 |

**Unlocks:** Smart Complex, Info Vault, Smart Farm, Smart Refinery, Precision Mine, Innovation Hub, Digital Temple, Cyber Command, Venture Hub, Smart Grid Node, Aerospace Foundry, Smart Energy Grid, Server Farm, Media Center, Global Network · **New domain:** hacker

---

## Age 14 — Digital Age 💻

> *Full digitization of civilization.*

| Requirement | Amount |
|---|---|
| Data | 3,100,000,000 |
| Electricity | 20,000,000,000 |
| Server Farms | 10 |
| Media Centers | 15 |
| Innovation Hubs | 15 |

**Unlocks:** Megaplex, Digital Archive, Nano Farm, Bio Fabrication Lab, Nano Drill Complex, AI Research Lab, Cyber Shrine, Drone Warfare Center, Crypto Exchange, Neural Grid, Nano Alloy Plant, Quantum Battery Array, Data Center, VR Studio, World Simulation

---

## Age 15 — Cyberpunk Age 🤖

> *Neon lights and cybernetic augmentation.*

| Requirement | Amount |
|---|---|
| Data | 160,000,000,000 |
| Electricity | 980,000,000,000 |
| AI Research Labs | 15 |
| Data Centers | 15 |
| Neural Grids | 15 |

**Unlocks:** Arcology Pod, Cyber Vault, Vat Farm, Nanobot Vat, Dark Crystal Mine, Neuro Research Center, Neon Sanctuary, Combat Aug Center, Black Market, Augmentation Foundry, Dark Matter Refinery, Dark Energy Tap, Cyber Hub, Holographic Theater, Neon Citadel · **Resources:** Crypto, Dark Matter Crystals

---

## Age 16 — Fusion Age 🔬

> *Clean energy breakthrough changes everything.*

| Requirement | Amount |
|---|---|
| Electricity | 490,000,000,000 |
| Crypto | 25,000,000,000 |
| Data | 78,000,000,000 |
| Augmentation Foundries | 15 |
| Arcology Pods | 25 |
| Black Markets | 15 |

**Unlocks:** Habitat Ring, Fusion Vault, Bio Reactor Farm, Molecular Synthesizer, Exotic Mineral Extractor, Theoretical Institute, Quantum Chapel, Plasma Command, Energy Exchange, Fusion Reactor, Exotic Matter Forge, Fusion Reactor Array, Quantum Server Farm, Neural Art Complex, Stellar Cradle · **Resource:** Plasma

---

## Age 17 — Space Age 🚀

> *Orbital expansion begins.*

| Requirement | Amount |
|---|---|
| Plasma | 63,000,000,000 |
| Electricity | 2,400,000,000,000 |
| Data | 390,000,000,000 |
| Fusion Reactors | 10 |
| Fusion Reactor Arrays | 10 |
| Plasma Commands | 10 |

**Unlocks:** Orbital Habitat, Orbital Depot, Hydroponic Bay, Quantum Organic Extractor, Asteroid Crystal Mine, Deep Space Observatory, Orbital Sanctuary, Space Force Base, Asteroid Market, Launch Complex, Orbital Refinery, Solar Collector Array, Orbital Data Relay, Zero-G Gallery, Dyson Scaffold · **Resource:** Titanium · **New domain:** astronaut

---

## Age 18 — Interstellar Age 🛸

> *Between the stars, new frontiers await.*

| Requirement | Amount |
|---|---|
| Titanium | 130,000,000,000 |
| Plasma | 310,000,000,000 |
| Launch Complexes | 10 |
| Orbital Habitats | 15 |
| Solar Collector Arrays | 10 |

**Unlocks:** Generation Ship, Stellar Vault, Protein Synthesizer, Reality Matter Weaver, Stellar Core Drill, Xenology Institute, Void Monastery, Fleet Command, Galactic Trade Hub, Warp Drive Plant, Antimatter Forge, Pulsar Tap, Galactic Network Node, Cultural Beacon, Warp Nexus · **Resource:** Dark Matter

---

## Age 19 — Galactic Age 🌌

> *Galactic civilization spans the cosmos.*

| Requirement | Amount |
|---|---|
| Dark Matter | 250,000,000,000 |
| Titanium | 630,000,000,000 |
| Warp Drive Plants | 15 |
| Generation Ships | 30 |
| Antimatter Forges | 15 |

**Unlocks:** Dyson Sphere Habitat, Galactic Vault, Matter Converter, Cosmic Organic Works, Neutron Star Mine, Cosmic Research Station, Stellar Shrine, Stellar Armada HQ, Stellar Exchange, Dyson Assembly, Stellar Metallurgy, Quasar Tap, Consciousness Upload Hub, Civilization Archive, Cosmic Beacon · **Resource:** Antimatter

---

## Age 20 — Quantum Age ⚛️

> *Reality bends to quantum mastery.*

| Requirement | Amount |
|---|---|
| Antimatter | 6,300,000,000,000 |
| Dark Matter | 13,000,000,000,000 |
| Stellar Exchanges | 15 |
| Stellar Metallurgy | 15 |
| Dyson Sphere Habitats | 30 |

**Unlocks:** Reality Fold, Quantum Vault, Quantum Cultivator, Reality Harvester, Reality Excavator, Reality Academy, Transcendence Hall, Probability War Room, Probability Market, Reality Forge, Quantum Metal Works, Zero Point Generator, Reality Processor, Reality Art Engine, Reality Anchor · **Resource:** Quantum Flux

---

## Age 21 — Transcendent Age ✨

> *Final ascension. The ultimate civilization.*

| Requirement | Amount |
|---|---|
| Quantum Flux | 190,000,000,000,000 |
| Antimatter | 310,000,000,000,000 |
| Reality Academies | 20 |
| Reality Forges | 15 |
| Probability War Rooms | 15 |

The end of the progression. Prestige has been available since Modern Age (Age 12). Reaching the Transcendent Age yields maximum prestige points — the ideal moment to prestige if you haven't already.

**Unlocks:** Singularity Core, Transcendent Nexus, Omniversal War Council, Omniversal Bazaar, Singularity Engine

```
prestige confirm yes
```

See [Prestige System](prestige.md) for details.
