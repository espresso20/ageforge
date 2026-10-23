# The 22 Ages

AgeForge spans 22 ages from primitive survival to transcendence. Each age unlocks new buildings, resources and worker domains, and adds a new district to your [skyline](map.md#skyline). Advancement requires the buildings each age asks for **and** completing the age's wonder. No age asks you to hold an amount of any resource.

From the Modern Age on, first-copy prices rise by 1.4 times each age. The Modern Age keeps its authored prices. Wonders and research scale with the price level; resource outputs do not. In every age a new copy of a building (wonders and monuments aside) also costs 1.43 times what it was written at.

<figure class="screen" data-screen="age-arrival"><figcaption>Reaching a new age, here the Bronze Age: its name is struck in iron, sparks fly, and the town turns from the age you left into the new one.</figcaption></figure>

An advance opens on the arrival: the whole screen, in your map style, theme and glyph set. Any key moves on to what the age opens, and a second key closes it. `Esc` counts as a key. Left alone, the arrival moves on by itself and the screen closes after 20 seconds. The game keeps running underneath the whole time. With `motion off` the arrival is a single still frame.

<figure class="screen" data-screen="age-advance"><figcaption>What the new age opens, shown after the arrival: its buildings and resources, its wonder, and the buildings ready to upgrade.</figcaption></figure>

If more than one age goes by before the screen can show (a queued [`plan advance`](plan.md) goes by itself), there is one arrival, for the furthest age, with the ages passed listed under its name.

## Wonder Requirement

Each age unlocks exactly one wonder building. You must **build that wonder** before you can advance to the next age, on top of the building requirements listed below. The age progress bar shows a red `✗ Wonder: <name>` notice until it is complete. See [Wonders](wonders.md) for build costs and instructions. The whole gate, sorted into four pillars, is on the [Advance panel](advance.md).

## Keystone Techs

From the Stone Age on, each age's wonder needs one technology of that age before it can be built: its **keystone**. The wonder's bank is open from the first tick of the age and fills while you research, by deposit. Only `build <wonder>` waits for the tech. Since the next age needs the wonder, the keystone is the one tech an age asks of you, together with the techs it stands on. Everything else in the tree is your choice.

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

**No age asks you to hold anything for research.** Your scholars do the research, and the keystone is what the advance needs of it. See [how research is priced](technologies.md#how-research-is-priced).

**A game saved before keystones** keeps the age it was in exempt: there the wonder builds without its keystone, and the lock starts when you next advance. The log says so once, on the first load. Every tech you researched stays researched and every wonder you built stays built.

## Advancement Requirements

An age asks for the town it is priced for. **No age asks you to hold an amount of any resource**, and advancing spends nothing. To leave an age you need:

- **Eight copies of every building of that age that makes something.** That means food, wood, stone, metals, gold, power, faith or culture. A military building counts too: it makes nothing, but it comes with soldiers.
- **Fewer copies of the age's knowledge building**, because each copy costs 50% more than the one before it (a producer's costs 15% more). The tables below give the count, from 5 to 10. See [Scholars and Research](knowledge.md#what-a-scholar-is).
- **Every building count the age asked for already.** A building the age asked for in more copies keeps that count.
- **Five copies of the age's storage building.**
- **The age's wonder**, built, and **its keystone tech**, researched. See [Wonder Requirement](#wonder-requirement) and [Keystone Techs](#keystone-techs).

The table under each age below lists the building counts the gate into it asks for. The Waiting on line at the top of the game lists what is still missing from the same list, with your count beside each one.

**What a gate leaves out.**
- A building that waits for a tech is asked for only when you cannot leave the age without that tech: it is the keystone of that age's wonder or of an earlier one, or a tech one stands on. Standing Stones (after Ritual), Barracks (after Military Tactics), Smelter (after Iron Smelting), Steam Works (after Electrification), Cyber Command (after Cybersecurity), Pulsar Tap (after Stellar Engineering) and Quasar Tap (after Antimatter Synthesis) are not asked for, nor is any other building behind a tech you can finish the age without.
- Housing is asked for only where a table lists it. Monuments and embassies make nothing, so no gate asks for them.
- The Transcendent Age is the last, so it has no gate and none of its buildings is asked for.

Every age can be finished, and an automated test checks this in every release. The test holds each age to these rules:

- Every building requirement names a building you can build in the age you are advancing **from**, never an older one you can no longer build.
- A town that passes a gate arrives with five copies of the age's storage building, because the gate asks for them, so it already has the production boost they give (see [Storage](buildings.md#storage-buildings-21-tiers)).
- A town that passes a gate has the age's knowledge building built to its full-rate count, because the gate asks for that many copies, and the scholars that come with them (see [Scholars and Research](knowledge.md#what-a-scholar-is)).
- The wonder's keystone tech, and every tech it stands on, can be researched by the wonder's own age.
- Every resource a required building, the wonder or the keystone costs has a source in the age itself, even for a town that skipped every building no earlier gate asked for. See [The Maker Rule](buildings.md#the-maker-rule).
- A food, faith or culture cost is something a moderate economy makes in the age.
- A town that arrives in a new age with nothing in stock, and only the buildings the gates asked for, can buy the first copy of every building the age opens, except those that wait for a tech you may never research.

## What Happens on Age Advance

When your civilization crosses into a new age:

1. **Carryover is capped.** Each resource the new age *uses as a build cost* is capped to roughly **8× the cheapest new-age starter building's cost** in that resource, enough to pay for a handful of opening buildings. If you were already below that cap, your stockpile carries over untouched. Anything above the cap is lost, so stockpiling before an advance doesn't pay. Resources the new age does **not** build with (mainly food) keep **10%** of what you had, and **faith is exempt** (it carries over in full). You enter each age with a small head start.
2. **Buildings of earlier ages become older buildings.** Wherever a building's line has a newer building for the new age, the old one is an older building. The newer building is in the same line and makes the same resource (see [The 39 Building Lines](buildings.md#the-39-building-lines)), so a Sawmill is never offered a Coal Mine. Nothing is converted automatically. Storage buildings are in no line: every stash, storage pit or vault you built keeps counting toward your storage for the rest of the run.
3. **Production continues at the old rate** until you upgrade. Leaving older buildings as they are costs nothing; the reason to upgrade is the new building's higher output, and an upgraded copy counts toward the next age's building requirements.
4. **New-tier buildings are available at once** in the Buildings panel, so you can build fresh copies of the new tier right away.
5. **The age's speed is set.** On known ground the log says how much faster the age runs (`Known ground: the Bronze Age runs 2.4x faster (mastery 2).`). Stepping from known ground onto an age no run has completed logs that it runs at 1x.

### Upgrading buildings after an advance

Use the `upgrade` command to turn older buildings into the building their line has for your age. Each copy costs half of what a new copy of that building costs, and the command upgrades as many copies as you can pay for:

```
upgrade                      # list the older buildings you hold
upgrade gathering_camp       # upgrade as many Gathering Camps as you can pay for
upgrade gathering_camp 3     # upgrade up to 3 copies
```

An upgraded copy keeps its workers, and it counts as the new building for the next age's gate. A building from several ages back goes straight to the building of the current age. An upgrade never takes the new building past its max count.

See [Buildings: Building Upgrades](buildings.md#building-upgrades) for the price rule, the limits and worker moves.

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
- Resources: Food, Wood, Faith
- Worker domains: food, lumber, faith

---

## Age 1: Stone Age 🪓

> *Stone tools and the first permanent camps.*

| Building | Copies |
|---|---|
| Gathering Camp | 8 |
| Hut | 10 |
| Shrine | 8 |
| Stash | 5 |
| Story Circle | 5 |
| Wood Camp | 8 |

**Unlocks:** Longhouse, Storage Pit, Forager Post, Woodcutter Camp, Stone Camp, Stone Pit, Elders' Hall, Standing Stones (after Ritual), War Camp, Great Monolith · **Resource:** Stone · **New domains:** masonry, military

---

## Age 2: Bronze Age 🛡

> *Metalworking arrives, and with it trade.*

| Building | Copies |
|---|---|
| Elders' Hall | 5 |
| Forager Post | 8 |
| Longhouse | 15 |
| Stone Camp | 8 |
| Stone Pit | 8 |
| Storage Pit | 5 |
| War Camp | 8 |
| Woodcutter Camp | 8 |

**Unlocks:** House, Warehouse, Farm, Lumber Mill, Quarry, Scriptorium, Altar (after Calendar), Barracks (after Military Tactics), Market, Smithy, Stonehenge · **Resources:** Iron, Gold · **New domains:** trade, engineering

---

## Age 3: Iron Age ⚔️

> *Iron tools and weapons spread.*

| Building | Copies |
|---|---|
| Altar | 8 |
| Farm | 8 |
| Lumber Mill | 8 |
| Market | 8 |
| Quarry | 8 |
| Scriptorium | 5 |
| Smithy | 8 |
| Warehouse | 5 |

**Unlocks:** Townhouse, Granary, Field Works, Timber Yard, Marble Quarry, Agora, Temple, Hunting Lodge, Legion Fort (after Siege Warfare), Trading Post, Ironworks, Smelter (after Iron Smelting), Colosseum · **Resources:** Marble, Iron Ore, Soldiers · **New domain:** metallurgy

---

## Age 4: Classical Age 🏛

> *Great empires are built and philosophy flourishes.*

| Building | Copies |
|---|---|
| Agora | 7 |
| Field Works | 8 |
| Granary | 5 |
| Hunting Lodge | 15 |
| Ironworks | 8 |
| Marble Quarry | 8 |
| Temple | 8 |
| Timber Yard | 8 |
| Trading Post | 10 |

**Unlocks:** Villa, Classical Vault, Terrace Farm, Wood Workshop, Marble Works, Library, Oracle House, Military Academy, Merchant Quarter, Aqueduct, Forge (after Metal Casting), Amphitheater, Parthenon, Cultural Obelisk · **Resource:** Culture

---

## Age 5: Medieval Age 🏰

> *Kingdoms rise and feudalism takes hold.*

| Building | Copies |
|---|---|
| Amphitheater | 8 |
| Aqueduct | 8 |
| Classical Vault | 5 |
| Library | 8 |
| Marble Works | 8 |
| Merchant Quarter | 8 |
| Military Academy | 15 |
| Oracle House | 8 |
| Terrace Farm | 8 |
| Wood Workshop | 8 |

Prestige becomes available at this age. A prestige from here to the Atomic Age is an early taste: it pays 9 points from the Medieval Age, enough for the Plan Template, while a run to the Modern Age pays 120. Type `prestige` to see what a prestige pays now and from the next age. See [Prestige System](prestige.md).

**Unlocks:** Manor, Strongroom, Demesne, Sawmill, Stonemason's Guild, Monastery Library, Cathedral (after Theology), Castle Keep, Guildhall, Workshop, Ironmonger, Great Hall, Great Library, Grand Amphitheater · **Resource:** Steel

---

## Age 6: Renaissance Age 🎨

> *Art, science, and exploration flourish.*

| Building | Copies |
|---|---|
| Castle Keep | 8 |
| Cathedral | 8 |
| Demesne | 8 |
| Great Hall | 8 |
| Guildhall | 10 |
| Ironmonger | 8 |
| Monastery Library | 5 |
| Sawmill | 8 |
| Stonemason's Guild | 8 |
| Strongroom | 5 |
| Workshop | 8 |

**Unlocks:** Estate, Renaissance Vault, Market Garden, Coal Mine, Iron Mine, University, Basilica, Fortress, Exchange, Mill, Foundry (after Blast Furnace), Art Studio, Sistine Chapel · **Resource:** Coal

---

## Age 7: Colonial Age ⚓

> *Exploration and trade span the globe.*

| Building | Copies |
|---|---|
| Art Studio | 12 |
| Basilica | 8 |
| Coal Mine | 8 |
| Exchange | 12 |
| Fortress | 8 |
| Iron Mine | 8 |
| Market Garden | 8 |
| Mill | 8 |
| Renaissance Vault | 5 |
| University | 7 |

**Unlocks:** Settlement Block, Colonial Warehouse, Plantation, Coal Works, Deep Iron Mine, Natural Philosophy Hall, Mission, Fort, Port, Harbor (after Mercantilism), Dockyard, Colonial Steelworks (after Coke Smelting), Concert Hall, Embassy (after Embassies), Grand Lighthouse

---

## Age 8: Industrial Age 🏭

> *Machines take over production.*

| Building | Copies |
|---|---|
| Coal Works | 8 |
| Colonial Warehouse | 5 |
| Concert Hall | 8 |
| Deep Iron Mine | 8 |
| Dockyard | 8 |
| Fort | 8 |
| Mission | 8 |
| Natural Philosophy Hall | 5 |
| Plantation | 8 |
| Port | 10 |

**Unlocks:** Tenement, Industrial Depot, Agricultural Works, Steam Colliery, Steam Mine, Research Institute, Church, Military Base, Stock Exchange, Harbor Authority, Integrated Steelworks, Steel Mill, Coal Plant (after Steam Power), Opera House, Grand Embassy (after the Concert of Nations), Geographic Society (after Geographic Societies), Crystal Palace, Eternal Library · **Resource:** Oil · **New domain:** energy

---

## Age 9: Victorian Age 🎩

> *Steam and innovation drive progress.*

| Building | Copies |
|---|---|
| Agricultural Works | 8 |
| Church | 8 |
| Coal Plant | 8 |
| Harbor Authority | 8 |
| Industrial Depot | 5 |
| Integrated Steelworks | 8 |
| Military Base | 8 |
| Opera House | 8 |
| Research Institute | 5 |
| Steam Colliery | 8 |
| Steam Mine | 8 |
| Steel Mill | 8 |
| Stock Exchange | 8 |
| Tenement | 30 |

**Unlocks:** Row House, Victorian Vault, Mechanized Farm, Oil Derrick, Uranium Mine, Academy, Grand Cathedral, Garrison, Bank, Steam Works (after Electrification), Bessemer Plant, Steam Turbine, Grand Museum, Eiffel Tower · **Resource:** Electricity

---

## Age 10: Electric Age ⚡

> *Electric light and power reach daily life.*

| Building | Copies |
|---|---|
| Academy | 7 |
| Bank | 8 |
| Bessemer Plant | 12 |
| Garrison | 8 |
| Grand Cathedral | 8 |
| Grand Museum | 8 |
| Mechanized Farm | 8 |
| Oil Derrick | 8 |
| Steam Turbine | 12 |
| Uranium Mine | 8 |
| Victorian Vault | 5 |

**Unlocks:** Apartment Block, Electric Warehouse, Industrial Farm, Oil Field, Nuclear Extraction Plant, Physics Laboratory, Revival Hall, Command Post, Financial District, Power Station, Electric Arc Furnace, Dynamo Hall (after Power Distribution), Radio Station, Hoover Dam

---

## Age 11: Atomic Age ☢️

> *Nuclear power, for better and worse.*

| Building | Copies |
|---|---|
| Command Post | 8 |
| Dynamo Hall | 8 |
| Electric Arc Furnace | 15 |
| Electric Warehouse | 5 |
| Financial District | 8 |
| Industrial Farm | 8 |
| Nuclear Extraction Plant | 8 |
| Oil Field | 8 |
| Physics Laboratory | 8 |
| Power Station | 15 |
| Radio Station | 8 |
| Revival Hall | 8 |

**Unlocks:** Housing Project, Atomic Vault, Agricultural Complex, Petroleum Refinery, Uranium Processing Works, Research Campus, Spiritual Center, Bunker Complex, Corporate HQ, Nuclear Plant (after Civilian Reactors), Advanced Alloy Plant, Breeder Reactor, Cinema, Particle Accelerator · **Resource:** Uranium

---

## Age 12: Modern Age 🌐

> *Technology and innovation define the era.*

| Building | Copies |
|---|---|
| Advanced Alloy Plant | 8 |
| Agricultural Complex | 8 |
| Atomic Vault | 5 |
| Breeder Reactor | 18 |
| Bunker Complex | 18 |
| Cinema | 8 |
| Corporate HQ | 8 |
| Petroleum Refinery | 8 |
| Research Campus | 9 |
| Spiritual Center | 8 |
| Uranium Processing Works | 8 |

From this age a prestige counts as a full run: it pays 120 points, against 93 from the Atomic Age before it. Type `prestige confirm yes` to reset and add those points to the lasting production bonus. See [Prestige System](prestige.md).

**Unlocks:** Tower Block, Modern Depot, Agritech Campus, Oil Platform, Titanium Mine, Think Tank, Meditation Center, Special Ops HQ, Investment Firm, Seaport (after Containerization), Power Grid Hub (after Advanced Electrics), Titanium Smelter, Oil Refinery, Mainframe Hall, TV Studio, Space Program, Nano Foundry, Monument of Ages (after Television) · **Resources:** Data, Nanobots, Titanium Ore

---

## Age 13: Information Age 📡

> *The Internet connects the world.*

| Building | Copies |
|---|---|
| Agritech Campus | 8 |
| Investment Firm | 8 |
| Mainframe Hall | 8 |
| Meditation Center | 8 |
| Modern Depot | 5 |
| Nano Foundry | 8 |
| Oil Platform | 8 |
| Oil Refinery | 15 |
| Power Grid Hub | 8 |
| Special Ops HQ | 8 |
| Think Tank | 10 |
| Titanium Mine | 8 |
| Titanium Smelter | 8 |
| Tower Block | 30 |
| TV Studio | 8 |

**Unlocks:** Smart Complex (after Internet of Things), Info Vault, Smart Farm (after Internet of Things), Smart Refinery, Precision Mine, Innovation Hub, Digital Temple, Cyber Command (after Cybersecurity), Venture Hub, Container Terminal (after E-commerce), Smart Grid Node, Aerospace Foundry, Microgrid Array (after Smart Grid), Server Farm, Media Center, Global Network · **New domain:** hacker

---

## Age 14: Digital Age 💻

> *Full digitization of civilization.*

| Building | Copies |
|---|---|
| Aerospace Foundry | 8 |
| Digital Temple | 8 |
| Info Vault | 5 |
| Innovation Hub | 8 |
| Media Center | 15 |
| Precision Mine | 8 |
| Server Farm | 10 |
| Smart Grid Node | 8 |
| Smart Refinery | 8 |
| Venture Hub | 8 |

**Unlocks:** Megaplex, Digital Archive, Nano Farm, Bio Fabrication Lab, Nano Drill Complex, AI Research Lab, Cyber Shrine, Drone Warfare Center, Crypto Exchange, Logistics Hub (after Automated Logistics), Neural Grid, Nano Alloy Plant, Quantum Battery Array (after Grid Storage), Data Center, VR Studio, World Simulation

---

## Age 15: Cyberpunk Age 🤖

> *Neon lights and cybernetic augmentation.*

| Building | Copies |
|---|---|
| AI Research Lab | 8 |
| Bio Fabrication Lab | 8 |
| Crypto Exchange | 8 |
| Cyber Shrine | 8 |
| Data Center | 15 |
| Digital Archive | 5 |
| Drone Warfare Center | 8 |
| Nano Alloy Plant | 8 |
| Nano Drill Complex | 8 |
| Nano Farm | 8 |
| Neural Grid | 15 |
| VR Studio | 8 |

**Unlocks:** Arcology Pod, Cyber Vault, Vat Farm, Nanobot Vat, Dark Crystal Mine, Neuro Research Center, Neon Sanctuary, Combat Aug Center (after Cybernetics), Black Market, Augmentation Foundry, Dark Matter Refinery, Dark Energy Tap (after Dark Energy), Cyber Hub, Holographic Theater (after Holography), Neon Citadel · **Resources:** Crypto, Dark Matter Crystals

---

## Age 16: Fusion Age 🔬

> *Fusion brings clean, plentiful energy.*

| Building | Copies |
|---|---|
| Arcology Pod | 25 |
| Augmentation Foundry | 15 |
| Black Market Hub | 15 |
| Combat Aug Center | 8 |
| Cyber Hub | 8 |
| Cyber Vault | 5 |
| Dark Crystal Mine | 8 |
| Dark Matter Refinery | 8 |
| Nanobot Vat | 8 |
| Neon Sanctuary | 8 |
| Neuro Research Center | 5 |
| Vat Farm | 8 |

**Unlocks:** Habitat Ring, Fusion Vault, Bio Reactor Farm, Molecular Synthesizer, Exotic Mineral Extractor, Theoretical Institute, Quantum Chapel, Plasma Command, Energy Exchange (after Maglev Transit), Fusion Reactor, Exotic Matter Forge, Tokamak Array, Quantum Server Farm, Neural Art Complex, Stellar Cradle · **Resource:** Plasma

---

## Age 17: Space Age 🚀

> *Orbital expansion begins.*

| Building | Copies |
|---|---|
| Bio Reactor Farm | 8 |
| Exotic Matter Forge | 8 |
| Exotic Mineral Extractor | 8 |
| Fusion Reactor | 10 |
| Fusion Vault | 5 |
| Molecular Synthesizer | 8 |
| Neural Art Complex | 8 |
| Plasma Command | 10 |
| Quantum Chapel | 8 |
| Quantum Server Farm | 8 |
| Theoretical Institute | 5 |
| Tokamak Array | 10 |

**Unlocks:** Orbital Habitat, Orbital Depot, Hydroponic Bay, Quantum Organic Extractor, Asteroid Crystal Mine, Deep Space Observatory, Orbital Sanctuary, Space Force Base, Asteroid Market, Launch Complex, Orbital Refinery, Solar Collector Array, Orbital Data Relay, Zero G Gallery, Dyson Scaffold · **Resource:** Titanium

---

## Age 18: Interstellar Age 🛸

> *Ships cross the space between stars.*

| Building | Copies |
|---|---|
| Asteroid Crystal Mine | 8 |
| Asteroid Market | 8 |
| Deep Space Observatory | 5 |
| Hydroponic Bay | 8 |
| Launch Complex | 10 |
| Orbital Data Relay | 8 |
| Orbital Depot | 5 |
| Orbital Habitat | 15 |
| Orbital Refinery | 8 |
| Orbital Sanctuary | 8 |
| Quantum Organic Extractor | 8 |
| Solar Collector Array | 10 |
| Space Force Base | 8 |
| Zero G Gallery | 8 |

**Unlocks:** Generation Ship, Stellar Vault, Protein Synthesizer, Reality Matter Weaver, Stellar Core Drill (after Stellar Core Mining), Xenology Institute, Void Monastery, Fleet Command, Galactic Trade Hub, Warp Drive Plant, Antimatter Forge, Pulsar Tap (after Stellar Engineering), Galactic Network Node, Cultural Beacon, Warp Nexus · **Resource:** Dark Matter

---

## Age 19: Galactic Age 🌌

> *Galactic civilization spans the cosmos.*

| Building | Copies |
|---|---|
| Antimatter Forge | 15 |
| Cultural Beacon | 8 |
| Fleet Command | 8 |
| Galactic Network Node | 8 |
| Galactic Trade Hub | 8 |
| Generation Ship | 30 |
| Protein Synthesizer | 8 |
| Reality Matter Weaver | 8 |
| Stellar Vault | 5 |
| Void Monastery | 8 |
| Warp Drive Plant | 15 |
| Xenology Institute | 5 |

**Unlocks:** Dyson Sphere Habitat, Galactic Vault, Matter Converter, Cosmic Organic Works, Neutron Star Mine (after Neutron Mining), Cosmic Research Station, Stellar Shrine, Stellar Armada HQ, Stellar Exchange, Dyson Assembly, Stellar Metallurgy, Quasar Tap (after Antimatter Synthesis), Consciousness Upload Hub, Civilization Archive, Cosmic Beacon · **Resource:** Antimatter

---

## Age 20: Quantum Age ⚛️

> *Reality bends to quantum mastery.*

| Building | Copies |
|---|---|
| Civilization Archive | 8 |
| Consciousness Upload Hub | 8 |
| Cosmic Organic Works | 8 |
| Cosmic Research Station | 5 |
| Dyson Assembly | 8 |
| Dyson Sphere Habitat | 30 |
| Galactic Vault | 5 |
| Matter Converter | 8 |
| Stellar Armada HQ | 8 |
| Stellar Exchange | 15 |
| Stellar Metallurgy | 15 |
| Stellar Shrine | 8 |

**Unlocks:** Reality Fold, Quantum Vault, Quantum Cultivator, Reality Harvester, Reality Excavator, Reality Academy, Transcendence Hall, Probability War Room, Probability Market, Reality Forge, Quantum Metal Works (after Quantum Metallurgy), Zero Point Generator (after Zero-Point Energy), Reality Processor, Reality Art Engine, Reality Anchor · **Resource:** Quantum Flux

---

## Age 21: Transcendent Age ✨

> *The last age. There is nowhere further to go.*

| Building | Copies |
|---|---|
| Probability Market | 8 |
| Probability War Room | 15 |
| Quantum Cultivator | 8 |
| Quantum Vault | 5 |
| Reality Academy | 10 |
| Reality Art Engine | 8 |
| Reality Excavator | 8 |
| Reality Forge | 15 |
| Reality Harvester | 8 |
| Reality Processor | 8 |
| Transcendence Hall | 8 |

The end of the progression. Prestige has been available since the Medieval Age (Age 5), and a full run since the Modern Age (Age 12). A prestige from the Transcendent Age pays the most points a single run can, 3,279: each Cosmic Era age the run completes adds 729 (see [Prestige Points Formula](prestige.md#prestige-points-formula)).

**Unlocks:** Singularity Core, Transcendent Nexus, Omniversal War Council (after Omniversal Command), Omniversal Bazaar (after Omniversal Exchange), Singularity Engine (after Singularity Engineering)

```
prestige confirm yes
```

See [Prestige System](prestige.md) for details.
