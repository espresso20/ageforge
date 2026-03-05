# Buildings

AgeForge uses a **13-lineage system** with 284 total buildings: 241 production buildings, 21 storage buildings, and 22 wonders. Buildings belong to lineages that span the full 22-age arc. When you advance an age, buildings in your current lineage tier automatically transform into the next tier — you keep your count and gain the upgraded building.

For the full wonder list see [Wonders](wonders.md).

---

## Lineage System

Buildings no longer unlock individually per age. Instead, each lineage starts at a specific age and gains one new tier per age from there. When your civilization enters a new age:

1. All buildings in active lineages that have a next tier **transform** automatically.
2. The old tier becomes **legacy** — still producing, cannot be built again.
3. The new tier becomes the buildable version at its upgraded stats.

You do not lose progress. A Gathering Camp in the Food lineage may become a Farm, then a Plantation, then an Agri-Complex — your existing count carries forward.

---

## The 13 Lineages

| # | Lineage | Domain | Primary Output | Tiers | First Building | Final Building |
|---|---------|--------|---------------|-------|----------------|----------------|
| 1 | Housing | None | Population cap | 21 | Hut (+10 pop) | Reality Fold (+1M pop) |
| 2 | Food | food | food | 21 | Gathering Camp | Quantum Cultivator |
| 3 | Organic Extraction | lumber | wood → coal → oil → nanobots → quantum flux | 21 | Wood Camp | Reality Harvester |
| 4 | Geological Extraction | masonry | stone → iron ore → uranium → titanium ore → dark matter crystals → antimatter | 21 | Stone Camp | Reality Excavator |
| 5 | Knowledge | knowledge | knowledge | 21 | Story Circle | Reality Academy |
| 6 | Faith | faith | faith | 21 | Shrine | Transcendence Hall |
| 7 | Military | military | soldiers | 21 | Hunting Lodge | Probability War Room |
| 8 | Trade | trade | gold | 19 | Market | Probability Market |
| 9 | Engineering | engineering | iron → steel → electricity → plasma → dark matter → quantum flux | 19 | Smithy | Reality Forge |
| 10 | Culture/Arts | None | culture | 17 | Amphitheater | Reality Art Engine |
| 11 | Metallurgy | metallurgy | iron → steel → titanium → dark matter → antimatter → quantum flux | 18 | Smelter | Quantum Metal Works |
| 12 | Energy | energy | coal → electricity → plasma → dark matter → quantum flux | 13 | Coal Plant | Zero Point Generator |
| 13 | Hacker/Digital | hacker | data | 8 | Server Farm | Reality Processor |

> Lineages 8–13 start in later ages (Bronze, Iron, Victorian, Information) so they have fewer tiers.

---

## Lineage Output Progression

### Organic Extraction (Lineage 3)

The resource this lineage outputs changes as you advance through epochs:

| Era | Output Resource |
|-----|----------------|
| Stone / Iron Era (tiers 1–5) | wood |
| Steel Era (tier 6+) | coal |
| Electric Era | oil |
| Digital Era+ | nanobots |
| Neon / Cosmic Era | quantum flux |

### Geological Extraction (Lineage 4)

| Era | Output Resource |
|-----|----------------|
| Stone / Bronze Era | stone |
| Iron Era | iron ore |
| Steel / Industrial Era | uranium |
| Modern Era | titanium ore |
| Neon Era | dark matter crystals |
| Cosmic Era | antimatter |

### Engineering (Lineage 9)

| Era | Output Resource |
|-----|----------------|
| Bronze / Iron Era | iron |
| Medieval / Steel Era | steel |
| Victorian / Electric Era | electricity |
| Fusion Era | plasma |
| Galactic Era | dark matter |
| Quantum Era | quantum flux |

### Metallurgy (Lineage 11)

Consumes raw ore from Geological Extraction and refines it into usable metals:

| Era | Ore In | Metal Out |
|-----|--------|----------|
| Iron / Medieval Era | iron ore | iron |
| Industrial / Electric Era | iron ore | steel |
| Modern / Space Era | titanium ore | titanium |
| Interstellar / Neon Era | dark matter crystals | dark matter |
| Cosmic / Galactic Era | dark matter crystals | antimatter |
| Quantum Era | — | quantum flux |

---

## Legacy Buildings

When a lineage tier upgrades, the previous-tier buildings become legacy. Legacy buildings:

- Continue producing at their normal rate
- Cannot be built again (greyed out in the build menu)
- Appear with a ☒ indicator in the Economy tab
- Count toward building totals in your civilization stats

You never lose production from an age advance. Your 20 Farms become 20 Plantations (or whatever the next tier is) automatically.

---

## Storage Buildings (21 tiers)

Storage buildings form their own lineage and expand your resource cap for **every** resource simultaneously. Build one as soon as you enter each new age — storage bottlenecks stop all late-game resource accumulation.

| Building | Age | Effect |
|----------|-----|--------|
| Stash | Primitive | +300 all storage (max 30) |
| Storage Pit | Stone | +500 all storage |
| Warehouse | Bronze | +3,000 all storage |
| Granary | Iron | +12,000 all storage |
| Classical Vault | Classical | +25,000 all storage |
| Keep | Medieval | +60,000 all storage |
| Renaissance Vault | Renaissance | +500,000 all storage |
| Colonial Warehouse | Colonial | +10M all storage |
| Industrial Depot | Industrial | +50M all storage |
| Victorian Vault | Victorian | +350M all storage |
| Electric Warehouse | Electric | +3.5B all storage |
| Atomic Vault | Atomic | +15B all storage |
| Modern Depot | Modern | +45B all storage |
| Info Vault | Information | +250B all storage |
| Digital Archive | Digital | +1.5T all storage |
| Cyber Vault | Cyberpunk | +5T all storage |
| Fusion Vault | Fusion | +30T all storage |
| Orbital Depot | Space | +200T all storage |
| Stellar Vault | Interstellar | +500T all storage |
| Galactic Vault | Galactic | +2Q all storage |
| Quantum Vault | Quantum | +5Q all storage |

> **Tip:** Stash is hard-capped at 30. Transition to Storage Pit the moment you enter the Stone Age.

---

## Wonders (22)

Wonders are unique one-of-a-kind buildings — you can only build each wonder once. They require resource banking before construction can begin. See [Wonders](wonders.md) for the full list.

```
wonder collect <resource> <amount|all>   — bank resources toward the active wonder
build <wonder_key>                       — begin construction once fully funded
```

Wonders provide powerful civilization-wide bonuses and are required for some milestone chains.

---

## Building Commands

```
build <key>            — queue one building
build <key> <count>    — queue multiple copies
build cancel           — cancel the current queue item
```

Only one building can be under construction at a time. Costs scale per copy: `cost × scale^(current_count)`. Building transforms on age advance happen automatically — no player action required.

---

## Worker Capacity

Every production building has a worker capacity slot count. The total available worker slots for a domain equals:

```
total_capacity = building_count × WorkerCapacity
```

More buildings mean more worker slots, which means a higher production ceiling. Building the same lineage tier repeatedly is how you scale output in a given age before the next transformation.

---

## Culture Buildings (Lineage 10)

Culture buildings have no worker domain. They automatically produce culture each tick based on their base rate, regardless of worker assignment. They also contribute to your culture storage cap. Build them to accumulate culture passively and unlock culture threshold bonuses.

---

## Economy Tab

Press `e` or `F1` to open the Economy tab. Buildings are displayed by lineage on the right panel. Legacy buildings show ☒. The left panel shows your current resource rates and worker assignments.
