# AgeForge — Worker Class System

## Overview

Workers are age-tiered. Every age unlocks a new class of worker for each relevant domain.
Higher-tier workers cost more food per tick but produce significantly more per assignment.
Old-tier workers remain after an age advance but cannot be newly recruited — only the
current age's tier is available for recruitment. This creates a workforce transition dynamic:
do you keep cheap legacy workers or invest in feeding expensive new ones?

## Mechanics

- Each building has a `worker_domain` (which class fills it) and `worker_capacity` (how many)
- Buildings produce at **20% base rate** with no workers assigned
- Each assigned worker of the matching domain adds to that building's output
- Full capacity = 100% production. Over-capacity is not possible.
- A Serf (Food domain) assigned to a Smithy (Engineering domain) produces nothing — domains are strict
- Higher-tier workers have a higher `output_multiplier` — a Colonist does more per slot than a Gatherer
- **Culture buildings have no worker domain** — they auto-produce culture each tick (see resources.md)

## 12 Worker Domains

| # | Domain | Unlock Age | Starting Class | Ending Class |
|---|--------|-----------|---------------|-------------|
| 1 | Food | Primitive | Gatherer | Reality Cultivator |
| 2 | Lumber | Primitive | Wood Gatherer | Reality Lumberjack |
| 3 | Masonry | Primitive | Stone Picker | Reality Excavator |
| 4 | Knowledge | Primitive | Lorekeeper | Transcendent Mind |
| 5 | Faith | Primitive | Shaman | Transcendent |
| 6 | Military | Primitive | Hunter | Probability Assassin |
| 7 | Trade | Primitive | Trader | Probability Trader |
| 8 | Engineering | Bronze | Toolmaker | Reality Engineer |
| 9 | Metallurgy | Iron | Smelter | Quantum Metallurgist |
| 10 | Energy | Industrial | Stoker | Zero-Point Engineer |
| 11 | Hacker | Information | Programmer | Quantum Cryptographer |
| 12 | Astronaut | Space | Astronaut | Dimension Walker |

**Culture has no worker domain.** Culture/Arts buildings auto-produce culture passively.

## Tier Stats (design targets, to be tuned)

Each tier step within a domain:
- Food cost per worker: × 1.5 over previous tier
- Output per worker: × 2.0 over previous tier
- Net efficiency per food spent: × 1.33 improvement per tier (always worth upgrading)

---

## Worker Class Tables

### 1. Food Domain
Fills: gathering_camp, forager_post, farm, field_works, estate_farm, demesne,
       market_garden, plantation, agricultural_works, mechanized_farm, industrial_farm,
       agricultural_complex, agri_complex, smart_farm, nano_farm, vat_farm,
       bio_reactor_farm, hydroponic_bay, protein_synthesizer, matter_converter,
       quantum_cultivator

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Gatherer | 0.10 | 1.0× |
| Stone | Forager | 0.15 | 2.0× |
| Bronze | Farmhand | 0.22 | 4.0× |
| Iron | Peasant | 0.33 | 8.0× |
| Classical | Agrarian | 0.50 | 16× |
| Medieval | Serf | 0.75 | 32× |
| Renaissance | Yeoman | 1.10 | 64× |
| Colonial | Settler | 1.65 | 128× |
| Industrial | Farm Laborer | 2.50 | 256× |
| Victorian | Mechanized Farmer | 3.75 | 512× |
| Electric | Agricultural Worker | 5.60 | 1,024× |
| Atomic | Agricultural Technician | 8.40 | 2,048× |
| Modern | Food Scientist | 12.6 | 4,096× |
| Information | Agronomist | 19.0 | 8,192× |
| Digital | Nano-Farmer | 28.5 | 16,384× |
| Cyberpunk | Vat Tender | 42.7 | 32,768× |
| Fusion | Bio-Synthesist | 64.0 | 65,536× |
| Space | Hydro-Farmer | 96.0 | 131,072× |
| Interstellar | Protein Engineer | 144 | 262,144× |
| Galactic | Matter Cultivator | 216 | 524,288× |
| Quantum | Reality Cultivator | 325 | 1,048,576× |

---

### 2. Lumber Domain
Fills: wood_camp, woodcutters_camp, lumber_mill, timber_yard, wood_workshop,
       sawmill, lumber_works, timber_plantation, steam_sawmill, lumber_mill_complex,
       automated_sawmill, chemical_pulp_mill, composite_factory, smart_lumber_yard,
       nano_wood_processor, synthetic_wood_vat, molecular_synthesizer, carbon_extractor,
       matter_weaver, quantum_lumber_works, reality_wood_works

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Wood Gatherer | 0.10 | 1.0× |
| Stone | Woodcutter | 0.15 | 2.0× |
| Bronze | Lumberjack | 0.22 | 4.0× |
| Iron | Timber Worker | 0.33 | 8.0× |
| Classical | Wood Artisan | 0.50 | 16× |
| Medieval | Sawyer | 0.75 | 32× |
| Renaissance | Carpenter | 1.10 | 64× |
| Colonial | Timber Merchant | 1.65 | 128× |
| Industrial | Lumber Worker | 2.50 | 256× |
| Victorian | Mill Operator | 3.75 | 512× |
| Electric | Automated Cutter | 5.60 | 1,024× |
| Atomic | Pulp Worker | 8.40 | 2,048× |
| Modern | Composite Specialist | 12.6 | 4,096× |
| Information | Forestry Technician | 19.0 | 8,192× |
| Digital | Nano-Wood Engineer | 28.5 | 16,384× |
| Cyberpunk | Synth-Wood Operator | 42.7 | 32,768× |
| Fusion | Molecular Weaver | 64.0 | 65,536× |
| Space | Carbon Extractor | 96.0 | 131,072× |
| Interstellar | Matter Weaver | 144 | 262,144× |
| Galactic | Quantum Lumberjack | 216 | 524,288× |
| Quantum | Reality Lumberjack | 325 | 1,048,576× |

---

### 3. Masonry Domain
Fills: stone_camp, stone_pit, quarry, deep_quarry, marble_works, stonemason_guild,
       quarry_complex, mining_settlement, steam_quarry, rock_processing_plant,
       automated_quarry, blast_mining_works, open_pit_mine, smart_quarry,
       nano_drill_complex, augmented_mine, plasma_cutter_mine, asteroid_quarry,
       planetary_core_drill, neutron_star_mine, reality_excavator_works

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Stone Picker | 0.10 | 1.0× |
| Stone | Stone Cutter | 0.15 | 2.0× |
| Bronze | Quarryman | 0.22 | 4.0× |
| Iron | Stone Mason | 0.33 | 8.0× |
| Classical | Marble Cutter | 0.50 | 16× |
| Medieval | Master Mason | 0.75 | 32× |
| Renaissance | Quarry Worker | 1.10 | 64× |
| Colonial | Mining Settler | 1.65 | 128× |
| Industrial | Steam Driller | 2.50 | 256× |
| Victorian | Rock Processor | 3.75 | 512× |
| Electric | Rock Blaster | 5.60 | 1,024× |
| Atomic | Blast Miner | 8.40 | 2,048× |
| Modern | Mining Engineer | 12.6 | 4,096× |
| Information | Smart Miner | 19.0 | 8,192× |
| Digital | Nano-Driller | 28.5 | 16,384× |
| Cyberpunk | Augmented Miner | 42.7 | 32,768× |
| Fusion | Plasma Cutter | 64.0 | 65,536× |
| Space | Asteroid Miner | 96.0 | 131,072× |
| Interstellar | Core Driller | 144 | 262,144× |
| Galactic | Neutron Miner | 216 | 524,288× |
| Quantum | Reality Excavator | 325 | 1,048,576× |

---

### 4. Knowledge Domain
Fills: story_circle, elders_hall, scriptorium, agora, library, monastery_library,
       university, natural_philosophy_hall, research_institute, academy, physics_laboratory,
       research_campus, think_tank, innovation_hub, ai_research_lab, neuro_research_center,
       theoretical_institute, deep_space_observatory, xenology_institute,
       cosmic_research_station, reality_academy

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Lorekeeper | 0.30 | 1.0× |
| Stone | Elder Scholar | 0.45 | 2.0× |
| Bronze | Scribe | 0.68 | 4.0× |
| Iron | Philosopher | 1.02 | 8.0× |
| Classical | Rhetorician | 1.53 | 16× |
| Medieval | Monk | 2.30 | 32× |
| Renaissance | Polymath | 3.45 | 64× |
| Colonial | Naturalist | 5.17 | 128× |
| Industrial | Inventor | 7.76 | 256× |
| Victorian | Academic | 11.6 | 512× |
| Electric | Physicist | 17.5 | 1,024× |
| Atomic | Nuclear Scientist | 26.2 | 2,048× |
| Modern | Researcher | 39.3 | 4,096× |
| Information | Data Scientist | 59.0 | 8,192× |
| Digital | AI Engineer | 88.5 | 16,384× |
| Cyberpunk | Neural Theorist | 133 | 32,768× |
| Fusion | Plasma Theorist | 199 | 65,536× |
| Space | Astrophysicist | 299 | 131,072× |
| Interstellar | Xenologist | 448 | 262,144× |
| Galactic | Cosmic Scholar | 672 | 524,288× |
| Quantum | Transcendent Mind | 1,008 | 1,048,576× |

---

### 5. Faith Domain
Fills: shrine, standing_stones, altar, temple, oracle_house, cathedral, basilica,
       mission, church, grand_cathedral, revival_hall, spiritual_center, meditation_center,
       digital_temple, cyber_shrine, neon_sanctuary, quantum_chapel, orbital_sanctuary,
       void_monastery, stellar_shrine, transcendence_hall

Faith workers are MORE expensive than raw material workers — spiritual leadership is a premium.

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Shaman | 0.50 | 1.0× |
| Stone | Elder | 0.75 | 2.0× |
| Bronze | Priest | 1.10 | 4.0× |
| Iron | Oracle | 1.65 | 8.0× |
| Classical | High Priest | 2.50 | 16× |
| Medieval | Bishop | 3.75 | 32× |
| Renaissance | Theologian | 5.60 | 64× |
| Colonial | Missionary | 8.40 | 128× |
| Industrial | Preacher | 12.6 | 256× |
| Victorian | Evangelist | 19.0 | 512× |
| Electric | Revival Pastor | 28.5 | 1,024× |
| Atomic | Spiritual Counselor | 42.7 | 2,048× |
| Modern | Meditation Guide | 64.0 | 4,096× |
| Information | Digital Prophet | 96.0 | 8,192× |
| Digital | Cyber Shaman | 144 | 16,384× |
| Cyberpunk | Neon Mystic | 216 | 32,768× |
| Fusion | Quantum Shaman | 325 | 65,536× |
| Space | Void Priest | 487 | 131,072× |
| Interstellar | Cosmic Mystic | 730 | 262,144× |
| Galactic | Stellar Prophet | 1,096 | 524,288× |
| Quantum | Transcendent | 1,644 | 1,048,576× |

---

### 6. Military Domain
Fills: hunting_lodge, war_camp, barracks, legion_fort, military_academy, castle_keep,
       fortress, fort, military_base, garrison, command_post, bunker_complex,
       special_ops_hq, cyber_command, drone_warfare_center, combat_aug_center,
       plasma_command, space_force_base, fleet_command, stellar_armada_hq, probability_war_room

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Hunter | 0.25 | 1.0× |
| Stone | Warrior | 0.37 | 2.0× |
| Bronze | Spearman | 0.56 | 4.0× |
| Iron | Legionnaire | 0.84 | 8.0× |
| Classical | Hoplite | 1.26 | 16× |
| Medieval | Knight | 1.90 | 32× |
| Renaissance | Musketeer | 2.85 | 64× |
| Colonial | Militiaman | 4.27 | 128× |
| Industrial | Infantry | 6.40 | 256× |
| Victorian | Rifleman | 9.60 | 512× |
| Electric | Doughboy | 14.4 | 1,024× |
| Atomic | Grunt | 21.6 | 2,048× |
| Modern | Special Ops | 32.4 | 4,096× |
| Information | Cyber Warrior | 48.6 | 8,192× |
| Digital | Drone Pilot | 72.9 | 16,384× |
| Cyberpunk | Street Samurai | 109 | 32,768× |
| Fusion | Plasma Trooper | 164 | 65,536× |
| Space | Space Marine | 246 | 131,072× |
| Interstellar | Void Ranger | 369 | 262,144× |
| Galactic | Stellar Guardian | 553 | 524,288× |
| Quantum | Probability Assassin | 830 | 1,048,576× |

---

### 7. Trade Domain
Fills: market, trading_post, merchant_quarter, guildhall, exchange, port,
       stock_exchange, bank, financial_district, corporate_hq, investment_firm,
       venture_hub, crypto_exchange, black_market, energy_exchange, asteroid_market,
       galactic_trade_hub, stellar_exchange, probability_market

Unlocks at Primitive Age (Barter Post precursor); full lineage starts at Bronze.

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Primitive | Trader | 0.20 | 1.0× |
| Stone | Peddler | 0.30 | 2.0× |
| Bronze | Merchant | 0.45 | 4.0× |
| Iron | Caravanner | 0.68 | 8.0× |
| Classical | Agora Keeper | 1.02 | 16× |
| Medieval | Guilder | 1.53 | 32× |
| Renaissance | Financier | 2.30 | 64× |
| Colonial | Merchant Captain | 3.45 | 128× |
| Industrial | Industrialist | 5.17 | 256× |
| Victorian | Magnate | 7.76 | 512× |
| Electric | Tycoon | 11.6 | 1,024× |
| Atomic | Executive | 17.5 | 2,048× |
| Modern | Broker | 26.2 | 4,096× |
| Information | Venture Capitalist | 39.3 | 8,192× |
| Digital | Crypto Whale | 59.0 | 16,384× |
| Cyberpunk | Fixer | 88.5 | 32,768× |
| Fusion | Energy Baron | 133 | 65,536× |
| Space | Asteroid Trader | 199 | 131,072× |
| Interstellar | Trade Emissary | 299 | 262,144× |
| Galactic | Galactic Merchant | 448 | 524,288× |
| Quantum | Probability Trader | 672 | 1,048,576× |

---

### 8. Engineering Domain
Unlocks at Bronze Age.
Fills: smithy, ironworks, aqueduct, workshop, mill, dockyard, iron_works_complex,
       steam_works, power_station, nuclear_plant, power_grid_hub, smart_grid_node,
       neural_grid, augmentation_foundry, fusion_reactor, launch_complex,
       warp_drive_plant, dyson_assembly, reality_forge

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Bronze | Toolmaker | 0.22 | 1.0× |
| Iron | Smith | 0.33 | 2.0× |
| Classical | Architect | 0.50 | 4.0× |
| Medieval | Master Builder | 0.75 | 8.0× |
| Renaissance | Engineer | 1.10 | 16× |
| Colonial | Mechanist | 1.65 | 32× |
| Industrial | Machinist | 2.50 | 64× |
| Victorian | Steam Engineer | 3.75 | 128× |
| Electric | Electrical Engineer | 5.60 | 256× |
| Atomic | Nuclear Engineer | 8.40 | 512× |
| Modern | Systems Engineer | 12.6 | 1,024× |
| Information | Software Engineer | 19.0 | 2,048× |
| Digital | Systems Architect | 28.5 | 4,096× |
| Cyberpunk | Augmentation Specialist | 42.7 | 8,192× |
| Fusion | Fusion Engineer | 64.0 | 16,384× |
| Space | Rocket Scientist | 96.0 | 32,768× |
| Interstellar | Warp Engineer | 144 | 65,536× |
| Galactic | Dyson Architect | 216 | 131,072× |
| Quantum | Reality Engineer | 325 | 262,144× |

---

### 9. Metallurgy Domain
Unlocks at Iron Age.
Fills: smelter, forge, ironmonger, foundry, iron_works, steel_mill, bessemer_plant,
       electric_arc_furnace, titanium_works, advanced_alloy_plant, nano_materials_lab,
       molecular_foundry, augmented_metal_works, plasma_forge, orbital_smelter,
       stellar_forge, neutron_forge, quantum_metal_works

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Iron | Smelter | 0.33 | 1.0× |
| Classical | Forge Keeper | 0.50 | 2.0× |
| Medieval | Ironmonger | 0.75 | 4.0× |
| Renaissance | Foundry Worker | 1.10 | 8.0× |
| Colonial | Iron Smith | 1.65 | 16× |
| Industrial | Steel Worker | 2.50 | 32× |
| Victorian | Bessemer Operator | 3.75 | 64× |
| Electric | Arc Furnace Operator | 5.60 | 128× |
| Atomic | Titanium Worker | 8.40 | 256× |
| Modern | Alloy Specialist | 12.6 | 512× |
| Information | Nano-Materials Tech | 19.0 | 1,024× |
| Digital | Molecular Smith | 28.5 | 2,048× |
| Cyberpunk | Augmented Metallurgist | 42.7 | 4,096× |
| Fusion | Plasma Smith | 64.0 | 8,192× |
| Space | Orbital Metallurgist | 96.0 | 16,384× |
| Interstellar | Stellar Forger | 144 | 32,768× |
| Galactic | Neutron Smith | 216 | 65,536× |
| Quantum | Quantum Metallurgist | 325 | 131,072× |

---

### 10. Energy Domain
Unlocks at Industrial Age.
Fills: coal_plant, steam_turbine, power_generator, nuclear_reactor, oil_refinery,
       smart_energy_grid, quantum_battery_array, dark_energy_tap, fusion_reactor_array,
       solar_collector_array, pulsar_tap, quasar_tap, zero_point_generator

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Industrial | Stoker | 2.50 | 1.0× |
| Victorian | Turbine Operator | 3.75 | 2.0× |
| Electric | Electrician | 5.60 | 4.0× |
| Atomic | Reactor Technician | 8.40 | 8.0× |
| Modern | Power Engineer | 12.6 | 16× |
| Information | Grid Engineer | 19.0 | 32× |
| Digital | Quantum Battery Tech | 28.5 | 64× |
| Cyberpunk | Dark Energy Tap Tech | 42.7 | 128× |
| Fusion | Fusion Reactor Tech | 64.0 | 256× |
| Space | Solar Array Engineer | 96.0 | 512× |
| Interstellar | Pulsar Technician | 144 | 1,024× |
| Galactic | Quasar Engineer | 216 | 2,048× |
| Quantum | Zero-Point Engineer | 325 | 4,096× |

---

### 11. Hacker Domain
Unlocks at Information Age.
Fills: server_farm, data_center, cyber_hub, quantum_server_farm, orbital_data_relay,
       galactic_network_node, consciousness_upload_hub, reality_processor

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Information | Programmer | 0.30 | 1.0× |
| Digital | Hacker | 0.45 | 2.0× |
| Cyberpunk | Ghost | 0.68 | 4.0× |
| Fusion | Data Architect | 1.02 | 8.0× |
| Space | Signal Breaker | 1.53 | 16× |
| Interstellar | Memetic Hacker | 2.30 | 32× |
| Galactic | Consciousness Coder | 3.45 | 64× |
| Quantum | Quantum Cryptographer | 5.17 | 128× |

---

### 12. Astronaut Domain
Unlocks at Space Age.
Fills: space_station, orbital_habitat, warp_gate, colony_ship, star_forge,
       stellar_cradle, dyson_scaffold, cosmic_beacon, reality_anchor, singularity_core

| Age | Class Name | Food/tick | Output Multiplier |
|-----|-----------|-----------|-------------------|
| Space | Astronaut | 0.40 | 1.0× |
| Interstellar | Cosmonaut | 0.60 | 2.0× |
| Galactic | Void Traveler | 0.90 | 4.0× |
| Quantum | Dimension Walker | 1.35 | 8.0× |

---

## Recruitment Rules

1. **Only current-age tier** can be newly recruited. Advancing an age immediately unlocks
   the new tier for each domain.
2. **Legacy workers persist** — existing workers don't disappear, but cost their original
   food rate. Players may keep them or dismiss them.
3. **Dismissal**: workers can be dismissed (removed from population) to free up food budget.
4. **Assignment is cross-tier**: a building's worker capacity accepts any tier of the
   correct domain. Higher tiers produce more per slot.
5. **Over-assignment is impossible**: can't assign more workers to a building than its
   `worker_capacity` allows.
6. **Culture domain exception**: Culture/Arts buildings accept no workers — they produce
   culture automatically. See resources.md.

## UI Implications

- Population panel shows active tier prominently; legacy tiers shown collapsed or grayed
- Each worker class name appears in the UI (seeing "Serf" instead of "Worker" rewards lore engagement)
- Worker assignment uses domain name, not individual class name, to avoid UI churn on advance
  (you always "assign to Food" — the class name is flavor on top)
- Three raw material domains (Food, Lumber, Masonry) are grouped under "Materials" in the UI
  to reduce cognitive load — but they remain mechanically separate

## Worker Domain Summary Reference

| Domain | Buildings Filled | Unlock Age | Starting Food/tick |
|--------|-----------------|-----------|-------------------|
| Food | Food production lineage | Primitive | 0.10 |
| Lumber | Lumber production lineage | Primitive | 0.10 |
| Masonry | Stone/Masonry lineage | Primitive | 0.10 |
| Knowledge | Knowledge/Research lineage | Primitive | 0.30 |
| Faith | Faith/Spiritual lineage | Primitive | 0.50 |
| Military | Military/Defense lineage | Primitive | 0.25 |
| Trade | Trade/Commerce lineage (Bronze+) | Primitive | 0.20 |
| Engineering | Engineering/Infrastructure lineage | Bronze | 0.22 |
| Metallurgy | Metals/Smelting lineage | Iron | 0.33 |
| Energy | Energy lineage | Industrial | 2.50 |
| Hacker | Digital lineage | Information | 0.30 |
| Astronaut | Space-faring buildings | Space | 0.40 |
| *(none)* | Culture/Arts lineage | Classical | — |
