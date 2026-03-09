# AgeForge — Building Lineages

## Overview

Every production building belongs to a **lineage** — a chain of age-specific incarnations
of the same role. On age advance, buildings transform in-place (count preserved, stats
upgraded). See age-transitions.md for the transformation mechanics.

**13 lineages total:**
1. Housing (all ages, no workers)
2. Food Production (all ages)
3. **Organic Extraction** (all ages) — formerly "Lumber"; produces wood → coal → oil → nanobots → dark_matter → quantum_flux
4. **Geological Extraction** (all ages) — formerly "Masonry"; produces stone → marble/iron_ore → uranium → titanium_ore → dark_matter_crystals → antimatter
5. Knowledge (all ages)
6. Faith (all ages)
7. Military (all ages)
8. Trade (Primitive → Quantum)
9. Engineering / Infrastructure (Bronze → Quantum)
10. Culture/Arts (Classical → Quantum, no workers — see resources.md)
11. **Metallurgy** (Iron → Quantum) — processes Geological ore into refined construction metals
12. Energy (Industrial → Quantum) — converts fuel to usable power resources
13. Digital (Information → Quantum)

**Ageless:** Storage buildings (stack cumulatively, never transform), Wonders (unique, permanent).

### Lineage Output by Epoch

The key design feature: Organic Extraction and Geological Extraction change what resource they
produce at each epoch boundary. The building names and worker classes also change (via the normal
age-advance transformation), but the OUTPUT RESOURCE is the epoch-defining change.

| Epoch | Organic Extraction output | Geological Extraction output | Metallurgy output |
|-------|--------------------------|-----------------------------|--------------------|
| Stone Era | **wood** | **stone** | — (unlocks Iron Era) |
| Iron Era | **wood** (declining) | **marble** + **iron_ore** | **iron** |
| Steel Era | **coal** | **iron_ore** (deep seams) | **steel** |
| Electric Era | **oil** | **uranium** | **steel** (alloys) |
| Digital Era | **oil** (refined) / **nanobots** | **titanium_ore** | **titanium** |
| Neon Era | **nanobots** | **dark_matter_crystals** | **dark_matter** |
| Cosmic Era | **quantum_flux** (partial) | **antimatter** | (retired — cosmic scale) |

> The Organic lineage transitions from wood to coal at the Steel Era (Renaissance age). The
> building name changes from "Lumber Works" → "Coal Mine" and the output resource switches
> from wood to coal. Players see this as part of the age-advance transformation screen.

> The Geological lineage produces BOTH marble and iron_ore in the Iron Era (transition period).
> Marble is used in Classical/Medieval building costs; iron_ore feeds Metallurgy. Two parallel
> outputs from the same lineage at this epoch.

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

## Lineage 3 — Organic Extraction
Worker domain: **Organic Extraction** (Wood Gatherer → Reality Lumberjack). See workers.md.
**Output resource changes per epoch** — this lineage extracts carbon-based organic matter;
what that means evolves from living trees to exotic cosmic organic compounds.

| Age | Building | Worker Capacity | Output Resource | Epoch |
|-----|----------|----------------|----------------|-------|
| Primitive | Wood Camp | 3 | wood | Stone Era |
| Stone | Woodcutter's Camp | 4 | wood | Stone Era |
| Bronze | Lumber Mill | 5 | wood | Stone Era |
| Iron | Timber Yard | 5 | wood | Iron Era |
| Classical | Wood Workshop | 6 | wood | Iron Era |
| Medieval | Sawmill | 6 | wood | Iron Era |
| Renaissance | **Coal Mine** | 7 | **coal** | Steel Era ← epoch transition |
| Colonial | Coal Works | 8 | coal | Steel Era |
| Industrial | Steam Coal Plant | 10 | coal | Steel Era |
| Victorian | **Oil Derrick** | 10 | **oil** | Electric Era ← epoch transition |
| Electric | Oil Field | 12 | oil | Electric Era |
| Atomic | Petroleum Refinery | 12 | oil | Electric Era |
| Modern | Oil Platform | 14 | oil | Digital Era |
| Information | Smart Refinery | 15 | oil | Digital Era |
| Digital | **Bio-Fabrication Lab** | 16 | **nanobots** | Digital Era ← epoch transition |
| Cyberpunk | Nanobot Vat | 18 | nanobots | Neon Era |
| Fusion | Molecular Synthesizer | 20 | nanobots | Neon Era |
| Space | **Quantum Organic Extractor** | 20 | **quantum_flux** | Cosmic Era ← epoch transition |
| Interstellar | Reality Matter Weaver | 25 | quantum_flux | Cosmic Era |
| Galactic | Cosmic Organic Works | 25 | quantum_flux | Cosmic Era |
| Quantum | Reality Harvester | 30 | quantum_flux | Cosmic Era |

> **Epoch transitions in this lineage:** At the Renaissance age advance (entering Steel Era),
> the Sawmill transforms into a Coal Mine and begins producing coal instead of wood. Players
> see this on the age advance summary screen. At Victorian (entering Electric Era), the Coal Mine
> transforms into an Oil Derrick. And so on at each epoch boundary.

---

## Lineage 4 — Geological Extraction
Worker domain: **Geological Extraction** (Stone Picker → Reality Excavator). See workers.md.
**Output resource changes per epoch** — this lineage mines progressively deeper geological
formations, from surface stone to stellar-core antimatter.

In the Iron Era, this lineage produces **two outputs**: marble (for construction) AND iron_ore
(fed to Metallurgy). This dual output is the transitional period between stone and full iron
economy. In implementation, two building slots or a split output rate handles this.

| Age | Building | Worker Capacity | Output Resource | Epoch |
|-----|----------|----------------|----------------|-------|
| Primitive | Stone Camp | 3 | stone | Stone Era |
| Stone | Stone Pit | 4 | stone | Stone Era |
| Bronze | Quarry | 5 | stone | Stone Era |
| Iron | **Marble Quarry** | 5 | marble + iron_ore | Iron Era ← epoch transition (dual) |
| Classical | Marble Works | 6 | marble + iron_ore | Iron Era |
| Medieval | Stonemason's Guild | 6 | iron_ore | Iron Era |
| Renaissance | **Iron Mine** | 7 | **iron_ore** | Steel Era ← epoch transition |
| Colonial | Deep Iron Mine | 8 | iron_ore | Steel Era |
| Industrial | Steam Mine | 10 | iron_ore | Steel Era |
| Victorian | **Uranium Mine** | 10 | **uranium** | Electric Era ← epoch transition |
| Electric | Nuclear Extraction Plant | 12 | uranium | Electric Era |
| Atomic | Uranium Processing Works | 12 | uranium | Electric Era |
| Modern | **Titanium Mine** | 14 | **titanium_ore** | Digital Era ← epoch transition |
| Information | Precision Mine | 15 | titanium_ore | Digital Era |
| Digital | Nano-Drill Complex | 16 | titanium_ore | Digital Era |
| Cyberpunk | **Dark Crystal Mine** | 18 | **dark_matter_crystals** | Neon Era ← epoch transition |
| Fusion | Exotic Mineral Extractor | 20 | dark_matter_crystals | Neon Era |
| Space | Asteroid Crystal Mine | 20 | dark_matter_crystals | Neon Era |
| Interstellar | **Stellar Core Drill** | 25 | **antimatter** | Cosmic Era ← epoch transition |
| Galactic | Neutron Star Mine | 25 | antimatter | Cosmic Era |
| Quantum | Reality Excavator | 30 | antimatter | Cosmic Era |

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

## Lineage 11 — Metallurgy (Processing Chain)
Worker domain: **Metallurgy** (Smelter → Quantum Metallurgist). Unlocks at Iron Age.

**2-stage dependency:** Metallurgy consumes raw ore from Geological Extraction and refines it
into construction metals. It does NOT produce metals from thin air — it requires ore input.
If ore supply drops below Metallurgy throughput, Metallurgy buildings under-produce.

```
Geological Extraction → [ore] → Metallurgy → [refined metal] → building costs
```

| Age | Building | Workers | Ore Input | Metal Output | Epoch |
|-----|----------|---------|-----------|-------------|-------|
| Iron | Smelter | 4 | iron_ore | **iron** | Iron Era |
| Classical | Forge | 4 | iron_ore | **iron** | Iron Era |
| Medieval | Ironmonger | 5 | iron_ore | **iron** | Iron Era |
| Renaissance | Foundry | 5 | **iron** | **steel** | Steel Era ← now processes iron into steel |
| Colonial | Iron Works | 6 | iron | steel | Steel Era |
| Industrial | Steel Mill | 7 | iron | steel | Steel Era |
| Victorian | Bessemer Plant | 8 | iron | steel | Electric Era |
| Electric | Electric Arc Furnace | 9 | iron | steel | Electric Era |
| Atomic | Advanced Alloy Plant | 10 | iron | steel | Electric Era |
| Modern | **Titanium Smelter** | 11 | **titanium_ore** | **titanium** | Digital Era ← new ore type |
| Information | Aerospace Foundry | 12 | titanium_ore | titanium | Digital Era |
| Digital | Nano-Alloy Plant | 13 | titanium_ore | titanium | Digital Era |
| Cyberpunk | **Dark Matter Refinery** | 14 | **dark_matter_crystals** | **dark_matter** | Neon Era ← new ore type |
| Fusion | Exotic Matter Forge | 16 | dark_matter_crystals | dark_matter | Neon Era |
| Space | Orbital Refinery | 18 | dark_matter_crystals | dark_matter | Neon Era |
| Interstellar | **Antimatter Forge** | 20 | **antimatter** | antimatter (refined) | Cosmic Era |
| Galactic | Stellar Metallurgy Works | 22 | antimatter | antimatter (refined) | Cosmic Era |
| Quantum | Quantum Metal Works | 25 | antimatter | **quantum_flux** | Cosmic Era |

> **Note on the Steel Era transition:** At Renaissance, Metallurgy no longer takes iron_ore as input.
> It now takes iron (the output of the Iron Era Metallurgy tier itself) and produces steel. This means
> iron becomes an intermediate resource: Geological → iron_ore → Metallurgy (Iron tier) → iron →
> Metallurgy (Steel tier) → steel → building costs. Full chain.

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
