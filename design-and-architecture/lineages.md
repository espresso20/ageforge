# AgeForge — Building Lineages

## Overview

Every production building belongs to a **lineage** — a chain of age-specific incarnations
of the same role. On age advance, buildings transform in-place (count preserved, stats
upgraded). See age-transitions.md for the transformation mechanics.

**13 lineages total:**
1. Housing (all ages, no workers)
2. Food Production (all ages)
3. Lumber Production (all ages)
4. Masonry/Stone (all ages)
5. Knowledge (all ages)
6. Faith (all ages)
7. Military (all ages)
8. Trade (Bronze → Quantum)
9. Engineering (Bronze → Quantum)
10. Culture/Arts (Classical → Quantum, no workers — see resources.md)
11. Metals/Smelting (Iron → Quantum)
12. Energy (Industrial → Quantum)
13. Digital (Information → Quantum)

**Ageless:** Storage buildings (stack cumulatively, never transform), Wonders (unique, permanent).

---

## Lineage 1 — Housing
No workers. Pop cap scales with housing tier. ~5× per tier target.

| Age | Building | Pop per building |
|-----|----------|-----------------|
| Primitive | Hut | +10 |
| Stone | Longhouse | +25 |
| Bronze | House | +50 |
| Iron | Townhouse | +80 |
| Classical | Villa | +120 |
| Medieval | Manor | +200 |
| Renaissance | Estate | +350 |
| Colonial | Settlement Block | +600 |
| Industrial | Tenement | +1,000 |
| Victorian | Row House | +1,800 |
| Electric | Apartment Block | +3,200 |
| Atomic | Housing Project | +5,500 |
| Modern | Tower Block | +10,000 |
| Information | Smart Complex | +18,000 |
| Digital | Megaplex | +32,000 |
| Cyberpunk | Arcology Pod | +55,000 |
| Fusion | Habitat Ring | +100,000 |
| Space | Orbital Habitat | +180,000 |
| Interstellar | Generation Ship | +320,000 |
| Galactic | Dyson Sphere Habitat | +600,000 |
| Quantum | Reality Fold | +1,000,000 |

---

## Lineage 2 — Food Production
Worker domain: **Food** (Gatherer → Reality Cultivator). See workers.md.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Gathering Camp | 3 |
| Stone | Forager Post | 4 |
| Bronze | Farm | 5 |
| Iron | Field Works | 5 |
| Classical | Estate Farm | 6 |
| Medieval | Demesne | 6 |
| Renaissance | Market Garden | 7 |
| Colonial | Plantation | 8 |
| Industrial | Agricultural Works | 10 |
| Victorian | Mechanized Farm | 10 |
| Electric | Industrial Farm | 12 |
| Atomic | Agricultural Complex | 12 |
| Modern | Agri-Complex | 14 |
| Information | Smart Farm | 15 |
| Digital | Nano-Farm | 16 |
| Cyberpunk | Vat Farm | 18 |
| Fusion | Bio-Reactor Farm | 20 |
| Space | Hydroponic Bay | 20 |
| Interstellar | Protein Synthesizer | 25 |
| Galactic | Matter Converter | 25 |
| Quantum | Quantum Cultivator | 30 |

---

## Lineage 3 — Lumber Production
Worker domain: **Lumber** (Wood Gatherer → Reality Lumberjack). See workers.md.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Wood Camp | 3 |
| Stone | Woodcutter's Camp | 4 |
| Bronze | Lumber Mill | 5 |
| Iron | Timber Yard | 5 |
| Classical | Wood Workshop | 6 |
| Medieval | Sawmill | 6 |
| Renaissance | Lumber Works | 7 |
| Colonial | Timber Plantation | 8 |
| Industrial | Steam Sawmill | 10 |
| Victorian | Lumber Mill Complex | 10 |
| Electric | Automated Sawmill | 12 |
| Atomic | Chemical Pulp Mill | 12 |
| Modern | Composite Factory | 14 |
| Information | Smart Lumber Yard | 15 |
| Digital | Nano-Wood Processor | 16 |
| Cyberpunk | Synthetic Wood Vat | 18 |
| Fusion | Molecular Synthesizer | 20 |
| Space | Carbon Extractor | 20 |
| Interstellar | Matter Weaver | 25 |
| Galactic | Quantum Lumber Works | 25 |
| Quantum | Reality Wood Works | 30 |

---

## Lineage 4 — Masonry / Stone
Worker domain: **Masonry** (Stone Picker → Reality Excavator). See workers.md.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Stone Camp | 3 |
| Stone | Stone Pit | 4 |
| Bronze | Quarry | 5 |
| Iron | Deep Quarry | 5 |
| Classical | Marble Works | 6 |
| Medieval | Stonemason's Guild | 6 |
| Renaissance | Quarry Complex | 7 |
| Colonial | Mining Settlement | 8 |
| Industrial | Steam Quarry | 10 |
| Victorian | Rock Processing Plant | 10 |
| Electric | Automated Quarry | 12 |
| Atomic | Blast Mining Works | 12 |
| Modern | Open Pit Mine | 14 |
| Information | Smart Quarry | 15 |
| Digital | Nano-Drill Complex | 16 |
| Cyberpunk | Augmented Mine | 18 |
| Fusion | Plasma Cutter Mine | 20 |
| Space | Asteroid Quarry | 20 |
| Interstellar | Planetary Core Drill | 25 |
| Galactic | Neutron Star Mine | 25 |
| Quantum | Reality Excavator Works | 30 |

---

## Lineage 5 — Knowledge
Worker domain: **Knowledge** (Lorekeeper → Transcendent Mind). See workers.md.
Produces: `knowledge` resource (fuels research, tech unlocks).

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Story Circle | 2 |
| Stone | Elder's Hall | 2 |
| Bronze | Scriptorium | 3 |
| Iron | Agora | 3 |
| Classical | Library | 4 |
| Medieval | Monastery Library | 4 |
| Renaissance | University | 5 |
| Colonial | Natural Philosophy Hall | 5 |
| Industrial | Research Institute | 6 |
| Victorian | Academy | 6 |
| Electric | Physics Laboratory | 7 |
| Atomic | Research Campus | 7 |
| Modern | Think Tank | 8 |
| Information | Innovation Hub | 8 |
| Digital | AI Research Lab | 10 |
| Cyberpunk | Neuro-Research Center | 10 |
| Fusion | Theoretical Institute | 12 |
| Space | Deep Space Observatory | 12 |
| Interstellar | Xenology Institute | 15 |
| Galactic | Cosmic Research Station | 15 |
| Quantum | Reality Academy | 20 |

---

## Lineage 6 — Faith
Worker domain: **Faith** (Shaman → Transcendent). See workers.md.
Produces: `faith` resource. See resources.md for faith uses (morale, cohesion, diplomacy, prestige).
Faith IS a draining resource — must be actively maintained, unlike culture.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Shrine | 2 |
| Stone | Standing Stones | 2 |
| Bronze | Altar | 3 |
| Iron | Temple | 3 |
| Classical | Oracle House | 4 |
| Medieval | Cathedral | 5 |
| Renaissance | Basilica | 5 |
| Colonial | Mission | 5 |
| Industrial | Church | 5 |
| Victorian | Grand Cathedral | 6 |
| Electric | Revival Hall | 6 |
| Atomic | Spiritual Center | 6 |
| Modern | Meditation Center | 7 |
| Information | Digital Temple | 7 |
| Digital | Cyber Shrine | 8 |
| Cyberpunk | Neon Sanctuary | 8 |
| Fusion | Quantum Chapel | 9 |
| Space | Orbital Sanctuary | 9 |
| Interstellar | Void Monastery | 10 |
| Galactic | Stellar Shrine | 10 |
| Quantum | Transcendence Hall | 12 |

---

## Lineage 7 — Military
Worker domain: **Military** (Hunter → Probability Assassin). See workers.md.
Produces: soldiers, defense rating, expedition capacity.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Primitive | Hunting Lodge | 3 |
| Stone | War Camp | 4 |
| Bronze | Barracks | 5 |
| Iron | Legion Fort | 6 |
| Classical | Military Academy | 6 |
| Medieval | Castle Keep | 7 |
| Renaissance | Fortress | 7 |
| Colonial | Fort | 8 |
| Industrial | Military Base | 10 |
| Victorian | Garrison | 10 |
| Electric | Command Post | 12 |
| Atomic | Bunker Complex | 12 |
| Modern | Special Ops HQ | 14 |
| Information | Cyber Command | 15 |
| Digital | Drone Warfare Center | 16 |
| Cyberpunk | Combat Aug Center | 18 |
| Fusion | Plasma Command | 20 |
| Space | Space Force Base | 20 |
| Interstellar | Fleet Command | 25 |
| Galactic | Stellar Armada HQ | 25 |
| Quantum | Probability War Room | 30 |

---

## Lineage 8 — Trade / Commerce
Worker domain: **Trade** (Merchant → Probability Trader). Unlocks at Bronze Age.
Produces: `gold`, `culture` (secondary).

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Bronze | Market | 3 |
| Iron | Trading Post | 3 |
| Classical | Merchant Quarter | 4 |
| Medieval | Guildhall | 4 |
| Renaissance | Exchange | 5 |
| Colonial | Port | 5 |
| Industrial | Stock Exchange | 6 |
| Victorian | Bank | 6 |
| Electric | Financial District | 7 |
| Atomic | Corporate HQ | 7 |
| Modern | Investment Firm | 8 |
| Information | Venture Hub | 8 |
| Digital | Crypto Exchange | 10 |
| Cyberpunk | Black Market | 10 |
| Fusion | Energy Exchange | 12 |
| Space | Asteroid Market | 12 |
| Interstellar | Galactic Trade Hub | 15 |
| Galactic | Stellar Exchange | 15 |
| Quantum | Probability Market | 18 |

---

## Lineage 9 — Engineering / Infrastructure
Worker domain: **Engineering** (Toolmaker → Reality Engineer). Unlocks at Bronze Age.
Produces: secondary resources (iron, steel, electricity) depending on age tier.

| Age | Building | Worker Capacity | Primary Output |
|-----|----------|----------------|----------------|
| Bronze | Smithy | 4 | iron |
| Iron | Ironworks | 5 | iron |
| Classical | Aqueduct | 5 | stone (bonus) |
| Medieval | Workshop | 6 | iron/wood (bonus) |
| Renaissance | Mill | 6 | mixed |
| Colonial | Dockyard | 7 | gold/wood |
| Industrial | Iron Works Complex | 8 | iron/steel |
| Victorian | Steam Works | 9 | steel/coal |
| Electric | Power Station | 10 | electricity |
| Atomic | Nuclear Plant | 11 | electricity/uranium |
| Modern | Power Grid Hub | 12 | electricity/oil |
| Information | Smart Grid Node | 13 | electricity/data |
| Digital | Neural Grid | 14 | data/electricity |
| Cyberpunk | Augmentation Foundry | 15 | steel/data |
| Fusion | Fusion Reactor | 18 | plasma/electricity |
| Space | Launch Complex | 20 | titanium/plasma |
| Interstellar | Warp Drive Plant | 22 | dark_matter/plasma |
| Galactic | Dyson Assembly | 25 | dark_matter/antimatter |
| Quantum | Reality Forge | 30 | quantum_flux |

---

## Lineage 10 — Culture / Arts
**No workers. Auto-produces culture per tick.**
Culture accumulates (does not drain). Has a max cap set by number of culture buildings.
See resources.md for full culture accumulation rules, thresholds, and uses.

| Age | Building | Culture/tick | Adds to culture cap |
|-----|----------|-------------|---------------------|
| Classical | Amphitheater | 0.5 | +500 |
| Medieval | Great Hall | 1.0 | +1,000 |
| Renaissance | Art Studio | 2.0 | +2,500 |
| Colonial | Concert Hall | 4.0 | +5,000 |
| Industrial | Opera House | 8.0 | +10,000 |
| Victorian | Grand Museum | 15 | +25,000 |
| Electric | Radio Station | 30 | +50,000 |
| Atomic | Cinema | 60 | +100,000 |
| Modern | TV Studio | 120 | +250,000 |
| Information | Media Center | 250 | +500,000 |
| Digital | VR Studio | 500 | +1,000,000 |
| Cyberpunk | Holographic Theater | 1,000 | +2,500,000 |
| Fusion | Neural Art Complex | 2,000 | +5,000,000 |
| Space | Zero-G Gallery | 4,000 | +10,000,000 |
| Interstellar | Cultural Beacon | 8,000 | +25,000,000 |
| Galactic | Civilization Archive | 16,000 | +50,000,000 |
| Quantum | Reality Art Engine | 32,000 | +100,000,000 |

---

## Lineage 11 — Metals / Smelting
Worker domain: **Metallurgy** (Smelter → Quantum Metallurgist). Unlocks at Iron Age.
Produces: `iron`, `steel`, `titanium` depending on tier.

| Age | Building | Worker Capacity | Output |
|-----|----------|----------------|--------|
| Iron | Smelter | 4 | iron |
| Classical | Forge | 4 | iron/gold |
| Medieval | Ironmonger | 5 | iron |
| Renaissance | Foundry | 5 | iron/steel |
| Colonial | Iron Works | 6 | steel |
| Industrial | Steel Mill | 7 | steel |
| Victorian | Bessemer Plant | 8 | steel |
| Electric | Electric Arc Furnace | 9 | steel |
| Atomic | Titanium Works | 10 | titanium/steel |
| Modern | Advanced Alloy Plant | 11 | titanium |
| Information | Nano-Materials Lab | 12 | titanium |
| Digital | Molecular Foundry | 13 | titanium/data |
| Cyberpunk | Augmented Metal Works | 14 | titanium |
| Fusion | Plasma Forge | 16 | titanium/plasma |
| Space | Orbital Smelter | 18 | titanium/dark_matter |
| Interstellar | Stellar Forge | 20 | dark_matter |
| Galactic | Neutron Forge | 22 | dark_matter/antimatter |
| Quantum | Quantum Metal Works | 25 | quantum_flux |

---

## Lineage 12 — Energy
Worker domain: **Energy** (Stoker → Zero-Point Engineer). Unlocks at Industrial Age.
Produces: `coal`, `oil`, `electricity`, `plasma`, `uranium` depending on tier.

| Age | Building | Worker Capacity | Output |
|-----|----------|----------------|--------|
| Industrial | Coal Plant | 6 | coal |
| Victorian | Steam Turbine | 7 | coal/electricity |
| Electric | Power Generator | 8 | electricity |
| Atomic | Nuclear Reactor | 9 | electricity/uranium |
| Modern | Oil Refinery | 10 | oil/electricity |
| Information | Smart Energy Grid | 11 | electricity |
| Digital | Quantum Battery Array | 12 | electricity/data |
| Cyberpunk | Dark Energy Tap | 13 | electricity/data |
| Fusion | Fusion Reactor Array | 15 | plasma/electricity |
| Space | Solar Collector Array | 16 | plasma/electricity |
| Interstellar | Pulsar Tap | 18 | plasma/dark_matter |
| Galactic | Quasar Tap | 20 | dark_matter/antimatter |
| Quantum | Zero-Point Generator | 25 | quantum_flux |

---

## Lineage 13 — Digital
Worker domain: **Hacker** (Programmer → Quantum Cryptographer). Unlocks at Information Age.
Produces: `data`, `crypto`.

| Age | Building | Worker Capacity |
|-----|----------|----------------|
| Information | Server Farm | 8 |
| Digital | Data Center | 10 |
| Cyberpunk | Cyber Hub | 12 |
| Fusion | Quantum Server Farm | 14 |
| Space | Orbital Data Relay | 16 |
| Interstellar | Galactic Network Node | 18 |
| Galactic | Consciousness Upload Hub | 20 |
| Quantum | Reality Processor | 25 |

---

## Storage Buildings (Ageless — Cumulative, Not Transformed)

Storage buildings stack across ages. A player keeps all storage they've built.
MaxCount is enforced on storage buildings only.

| Age | Building | Storage bonus | Max Count |
|-----|----------|--------------|-----------|
| Primitive | Stash | +300 all | 50 |
| Stone | Storage Pit | +800 all | 40 |
| Bronze | Warehouse | +2,000 all | 35 |
| Iron | Granary | +5,000 food only | 30 |
| Classical | Vault | +10,000 gold/knowledge | 25 |
| Medieval | Castle Vault | +25,000 all | 20 |
| Renaissance | Treasury | +60,000 all | 18 |
| Colonial | Customs House | +150,000 all | 15 |
| Industrial | Industrial Depot | +400,000 all | 12 |
| Victorian | Victorian Vault | +1,000,000 all | 10 |
| Electric | Electric Warehouse | +2,500,000 all | 8 |
| Atomic | Atomic Vault | +6,000,000 all | 7 |
| Modern | Modern Depot | +15,000,000 all | 6 |
| Information | Info Vault | +40,000,000 all | 5 |
| Digital | Digital Archive | +100,000,000 all | 4 |
| Cyberpunk | Cyber Vault | +250,000,000 all | 4 |
| Fusion | Fusion Vault | +600,000,000 all | 3 |
| Space | Orbital Depot | +1.5B all | 3 |
| Interstellar | Stellar Vault | +4B all | 2 |
| Galactic | Galactic Vault | +10B all | 2 |
| Quantum | Quantum Vault | +25B all | 2 |

**Storage Covenant check:** At each age, max possible storage must be ≥ 2× cost of most
expensive building in that age. This must be verified during balance tuning (Phase 3).

---

## Wonders (Ageless — Unique, Never Transform)

Wonders are permanent landmarks built once. They provide major bonuses and gate prestige/
victory conditions. Full wonder list in config/buildings.go. Key rule: wonders with faith
or culture requirements should gate on those resource thresholds (see resources.md).
