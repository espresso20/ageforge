# The 22 ages

AgeForge has 22 ages, from primitive survival to transcendence. Each age opens new buildings, and most bring a new resource or a new worker domain. Your town changes with the age on the [Map](map.md). To leave an age you need the buildings the next age's table asks for, the age's wonder and the wonder's keystone tech. No age asks you to hold an amount of any resource.

From the Information Age on, first-copy prices step up by a factor of 1.4 with each age. The Modern Age keeps its written prices. Wonder prices and tech prices follow the same step. What buildings produce does not. In every age, a copy of any building but a wonder or a monument costs 1.43 times the price it was written at.

<figure class="screen" data-screen="age-arrival"><figcaption>The arrival screen for the Bronze Age: the age's name in large letters, its description and a one-line quip, and the town below.</figcaption></figure>

An advance opens the arrival screen. It takes the whole screen, in your theme and glyph set. The first part is a short celebration. It moves on by itself, or at any key (`Esc` counts as a key). The second part says what the age opens. A key closes it, or turns the page when the text needs more than one. The screen closes by itself 20 seconds after it opens. The game keeps running underneath the whole time. With `motion off` the celebration is a single still frame.

<figure class="screen" data-screen="age-advance"><figcaption>What the Bronze Age opens, shown after the arrival: its buildings (some wait for research), its resources, its wonder and the upgrades on offer.</figcaption></figure>

If more than one age goes by before the screen can show (a queued [`plan advance`](plan.md) goes by itself), there is one arrival, for the furthest age, with the ages passed listed under its name.

## Wonder requirement

Each age opens exactly one wonder, and you must build it before you can advance to the next age. The [Advance panel](advance.md) shows it in the Monument pillar as two rows: the wonder's bank, and the wonder itself with a note on what it waits for (`bank it first`, `needs <tech> first`, `ready to build` or `building, 40%`). The Waiting on line in the Ledger at the top of the game lists the same items while they are missing. See [Wonders](wonders.md) for prices and how to build one. The Advance panel sorts the whole gate into four pillars.

## Keystone techs

From the Stone Age on, each age's wonder needs one tech of that age before you can build it: its keystone. The wonder's bank is open from the start of the age, so you can fill it while you research. Only `build <wonder>` waits for the tech.

The next age needs the wonder, so the keystone is the one tech an age asks of you, together with the techs it stands on. The 21 keystones and the techs under them are 48 of the 203 techs, and they are the only ones a run has to research. Everything else in the tree is your choice. Research costs no resources: your scholars do the work (see [how research is priced](technologies.md#how-research-is-priced)).

The last column lists the techs a keystone needs first, with the techs each of those needs in brackets.

| Age | Wonder | Keystone tech | Needs first |
|---|---|---|---|
| Primitive | Sacred Grove | none | |
| Stone | Great Monolith | Stoneworking | Tool Making |
| Bronze | Stonehenge | Calendar | Ritual (Language) |
| Iron | Colosseum | Mathematics | Primitive Writing (Language) |
| Classical | Parthenon | Philosophy | Mathematics (Primitive Writing) |
| Medieval | Great Library | Theology | Philosophy (Mathematics) |
| Renaissance | Sistine Chapel | Patronage | Banking (Currency, Mathematics) |
| Colonial | Grand Lighthouse | Cartography | Navigation (Exploration, Mathematics) |
| Industrial | Crystal Palace | Industrialization | Steam Power (Steel Forging) |
| Victorian | Eiffel Tower | Mass Production | Industrialization (Steam Power) |
| Electric | Hoover Dam | Power Distribution | Electrification (Industrialization) |
| Atomic | Particle Accelerator | Nuclear Fission | Power Distribution (Electrification), Chemical Engineering (Mass Production) |
| Modern | Space Program | Satellite Technology | Rocketry (Aviation), Advanced Electrics (Nuclear Fission) |
| Information | Global Network | Internet | Computers (Advanced Electrics), Satellite Technology (Rocketry, Advanced Electrics) |
| Digital | World Simulation | Machine Learning | Internet (Computers, Satellite Technology), Cybersecurity (Computers) |
| Cyberpunk | Neon Citadel | Cybernetics | Neural Interface (Machine Learning) |
| Fusion | Stellar Cradle | Fusion Power | Nuclear Fission (Power Distribution, Chemical Engineering), Cybernetics (Neural Interface) |
| Space | Dyson Scaffold | Orbital Mechanics | Rocketry (Aviation), Plasma Physics (Fusion Power) |
| Interstellar | Warp Nexus | Warp Drive | Space Mining (Orbital Mechanics), Zero-G Manufacturing (Orbital Mechanics, Superconductors) |
| Galactic | Cosmic Beacon | Galactic Navigation | Warp Drive (Space Mining, Zero-G Manufacturing), Stellar Engineering (Warp Drive) |
| Quantum | Reality Anchor | Quantum Mechanics | Antimatter Synthesis (Galactic Navigation) |
| Transcendent | Singularity Core | Transcendence | Reality Manipulation (Quantum Mechanics) |

The Sacred Grove needs no tech, so the advance out of the Primitive Age asks for no research.

Until the keystone is researched, the Advance panel and the Waiting on line show it as `Keystone: <tech>` with its progress (`not started`, `queued` or `researching, 40%`). The Wonders panel and `wonder` show the keystone and whether it is researched, and the Research panel marks every keystone with a ★. A wonder in your [build plan](plan.md) waits for its keystone.

If you load a save from before keystones existed, the age it was in is exempt. There the wonder builds without its keystone, and the lock starts when you next advance. The log says so once, on the first load. Every tech you researched stays researched, and every wonder you built stays built.

## Advancement requirements

To leave an age you need all of these:

- The buildings that the next age's table lists, in the counts shown. A copy you have upgraded counts as the building it became, and a gate that asks for a building also counts the later buildings of its line.
- The age's wonder, built.
- The wonder's keystone tech, researched. The wonder cannot be built without it. See [Wonder requirement](#wonder-requirement) and [Keystone techs](#keystone-techs).

No age asks you to hold an amount of any resource, and advancing spends nothing. The log tells you when the gate is met. You advance by typing `advance`, or by queuing `plan advance`. A pending catastrophe stops the advance until you choose, and a doom foretold for the age is met first (see [In the way](advance.md#in-the-way)).

The table under each age below lists the building counts that the gate into it asks for. The Waiting on line in the Ledger at the top of the game lists what is still missing from the same list, with your count beside each one (`Gathering Camp 3/8`).

### How the counts are set

- Every building the age opened that makes something is asked for in eight copies. That means food, wood, stone, metals, gold, power, faith or culture. A military building counts too. It makes nothing, but it comes with soldiers.
- Some tables ask for more than eight of a building, such as 15 Hunting Lodges or 30 Tenements.
- The age's knowledge building is asked for in 5 to 10 copies. That is fewer, because each copy costs 50% more than the one before it (a producer's next copy costs 15% more). See [Scholars and Research](knowledge.md#what-a-scholar-is).
- Housing is asked for only where a table lists it.
- The age's storage building is asked for in five copies.

### What a gate leaves out

- A building that waits for a tech is asked for only when you cannot leave the age without that tech: it is the keystone of that age's wonder or of an earlier one, or a tech one of them stands on. The Altar (after Calendar), the Cathedral (after Theology) and the Coal Plant (after Steam Power) pass that test, so they are asked for.
- The Standing Stones (after Ritual), Barracks (after Military Tactics), Smelter (after Iron Smelting), Steam Works (after Electrification), Cyber Command (after Cybersecurity), Pulsar Tap (after Stellar Engineering) and Quasar Tap (after Antimatter Synthesis) are not asked for, and neither is any other building behind a tech you can finish the age without.
- Monuments and embassies make no resource, so no gate asks for them.
- The Transcendent Age is the last, so no gate leaves it and none of its buildings is asked for.

### How the gates are tested

Every age can be finished. Automated tests hold each gate to these rules:

- Every building requirement names a building you can build in the age you are advancing from, never an older one you can no longer build.
- A town that passes a gate arrives with five copies of the age's storage building, because the gate asks for them, so it already has the production boost they give (see [Storage](buildings.md#storage-buildings-21-tiers)).
- A town that passes a gate has the age's knowledge building at its full-rate count, because the gate asks for that many copies, and the scholars that come with them (see [Scholars and Research](knowledge.md#what-a-scholar-is)).
- The wonder's keystone tech, and every tech it stands on, can be researched by the wonder's own age.
- Every resource that a required building, the wonder or the keystone costs has a source in the age itself, even for a town that skipped every building no earlier gate asked for. See [The Maker Rule](buildings.md#the-maker-rule).
- A food, faith, culture or nanobot cost is something a moderate economy makes in the age.
- A town that arrives in an age with nothing in stock, and only the buildings the gates asked for, can buy the first copy of every building the age opens, except those that wait for a tech you may never research.

## The measured gate (experimental)

A second gate rule is under trial. It is off by default. To try it, start the game with the environment variable `AGEFORGE_GATES=v2`. The game reads the variable once, when it starts. Under this rule the wonder is asked for as before, and the building counts in the tables on this page are replaced by three measures: your town's population, the income a tick of the age's main materials, and one more tech of the age beside the keystone. The rule, its numbers and its choices of materials and techs may change. See [The Advance panel](advance.md#the-measured-gate-experimental) for what the panel shows under it.

## What happens on age advance

When your civilization crosses into an age:

1. **Carryover is capped.** Each resource the age uses as a build cost is capped at 8 times the smallest amount of that resource any building of the age costs (a wonder aside). That is enough for a handful of opening buildings. If you hold less than the cap, you keep all of it. Anything above the cap is lost, so stockpiling before an advance does not pay. A resource that no building of the age costs, mainly food, keeps 10% of what you held. Faith carries over in full. The log says so in a `Stockpiles trimmed for the new age` line.
2. **Buildings of earlier ages become older buildings.** Wherever a building's line has a newer building for the age you enter, the old one is an older building, and the log lists each one with its `upgrade` command. The newer building is in the same line and makes the same resource (see [The 39 Building Lines](buildings.md#the-39-building-lines)), so a Sawmill is never offered a Coal Mine. Nothing is converted by itself. Storage buildings, knowledge buildings, wonders, monuments and the standalone buildings are in no line, so they are never older buildings. Every storage building you built keeps adding its share to your production for the rest of the run. A building of an earlier age can no longer be built.
3. **Production continues at the old rate** until you upgrade. Leaving older buildings as they are costs nothing. The reason to upgrade is the newer building's higher output, and an upgraded copy counts toward the next gate.
4. **Buildings of the age are available at once** in the Buildings panel, so you can build copies of them right away.
5. **The age's speed is set.** On known ground the log says how much faster the age runs (`Known ground: the Bronze Age runs 2.4x faster (mastery 2).`). Stepping from known ground onto an age no run has completed logs that it runs at 1x. See [Era Mastery](prestige.md#era-mastery).

### Upgrading buildings after an advance

Use the `upgrade` command to turn older buildings into the building their line has for your age. Each copy costs half of what a copy of the newer building costs. Bare `upgrade` lists the older buildings you hold. With a building, it upgrades as many copies as you can pay for, or up to the count you give:

```
upgrade                      # list the older buildings you hold
upgrade gathering_camp       # upgrade as many Gathering Camps as you can pay for
upgrade gathering_camp 3     # upgrade up to 3 copies
```

An upgraded copy takes its workers with it, as many as the newer building has room for, and it counts as the newer building toward the next gate. A building from several ages back goes straight to the building of the current age. An upgrade stops at the newer building's max count, and it waits for the tech if the newer building needs one.

See [Buildings: Building Upgrades](buildings.md#building-upgrades) for the price rule, the limits and worker moves.

## Epoch overview

Ages come in epochs of 3 (the Cosmic Era has 4). An advance into the first age of an epoch crosses an epoch boundary and rolls an epoch event (see [Epochs](epochs.md)). The stockpile trim above applies to it like any other advance.

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

**Unlocks:** Longhouse, Storage Pit, Forager Post, Woodcutter Camp, Stone Camp, Stone Pit, Elders' Hall, Standing Stones (after Ritual), War Camp, Great Monolith · **Resource:** Stone · **Worker domain:** masonry

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

**Unlocks:** House, Warehouse, Farm, Lumber Mill, Quarry, Scriptorium, Altar (after Calendar), Barracks (after Military Tactics), Market, Smithy, Stonehenge · **Resources:** Iron, Gold · **Worker domains:** trade, engineering

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

**Unlocks:** Townhouse, Granary, Field Works, Timber Yard, Marble Quarry, Agora, Temple, Hunting Lodge, Legion Fort (after Siege Warfare), Trading Post, Ironworks, Smelter (after Iron Smelting), Colosseum · **Resources:** Marble, Iron Ore · **Worker domain:** metallurgy

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

Prestige opens at this age. A prestige from the Medieval Age through the Atomic Age is an early taste: it pays 9 points from the Medieval Age, against 120 from the Modern Age. Type `prestige` to see what a prestige pays from your age and what one more age would add. See [Prestige System](prestige.md).

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

**Unlocks:** Tenement, Industrial Depot, Agricultural Works, Steam Colliery, Steam Mine, Research Institute, Church, Military Base, Stock Exchange, Harbor Authority, Integrated Steelworks, Steel Mill, Coal Plant (after Steam Power), Opera House, Grand Embassy (after the Concert of Nations), Geographic Society (after Geographic Societies), Crystal Palace, Eternal Library · **Resource:** Oil · **Worker domains:** energy, military

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

From this age a prestige counts as a full run. It pays 120 points, against 93 from the Atomic Age. Type `prestige confirm yes` to reset your run and add those points to the lasting production bonus. See [Prestige System](prestige.md).

**Unlocks:** Tower Block, Modern Depot, Agritech Campus, Oil Platform, Titanium Mine, Think Tank, Meditation Center, Special Ops HQ, Investment Firm, Seaport (after Containerization), Power Grid Hub (after Advanced Electrics), Titanium Smelter, Oil Refinery, Mainframe Hall, TV Studio, Space Program, Nano Foundry, Monument of Ages (after Television) · **Resources:** Data, Nanobots, Titanium Ore · **Worker domain:** hacker

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

**Unlocks:** Smart Complex (after Internet of Things), Info Vault, Smart Farm (after Internet of Things), Smart Refinery, Precision Mine, Innovation Hub, Digital Temple, Cyber Command (after Cybersecurity), Venture Hub, Container Terminal (after E-commerce), Smart Grid Node, Aerospace Foundry, Microgrid Array (after Smart Grid), Server Farm, Media Center, Global Network

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

**Unlocks:** Arcology Pod, Cyber Vault, Vat Farm, Nanobot Vat, Dark Crystal Mine, Neuro Research Center, Neon Sanctuary, Combat Aug Center (after Cybernetics), Black Market Hub, Augmentation Foundry, Dark Matter Refinery, Dark Energy Tap (after Dark Energy), Cyber Hub, Holographic Theater (after Holography), Neon Citadel · **Resources:** Crypto, Dark Matter Crystals

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

Prestige has been open since the Medieval Age (Age 5), and counts as a full run since the Modern Age (Age 12). A prestige from the Transcendent Age earns the most points a single run can, 3,279: each Cosmic Era age the run completes adds 729 (see [Prestige points formula](prestige.md#prestige-points-formula)). In the Cosmic Era the [Last Passage](prestige.md#the-last-passage) can take part of a run's points. Type `prestige confirm yes` to prestige. See [Prestige System](prestige.md) for details.

**Unlocks:** Singularity Core, Transcendent Nexus, Omniversal War Council (after Omniversal Command), Omniversal Bazaar (after Omniversal Exchange), Singularity Engine (after Singularity Engineering)
