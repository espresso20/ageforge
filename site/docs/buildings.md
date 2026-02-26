# Buildings

AgeForge has 80 buildings: 58 standard buildings grouped by category, and 22 wonders. Buildings unlock per age and have scaling costs — each additional copy costs more than the last.

For the full wonder list see [Wonders](wonders.md).

---

## Building mechanics

- Only **one building** can be in the queue at a time
- Construction takes a fixed number of ticks (`BuildTicks`)
- Cost scales per copy: `cost × scale^count`
- Some buildings can be **upgraded** to a more advanced version (costs 25% of the destination's base cost)
- Press `F1` or type `e` for the Economy tab — buildings are grouped by category on the right side

```
build <key>
build cancel
```

---

## Housing
*Raises population cap. More pop = more workers.*

| Building | Key | Age | Base Cost | Effect |
|---|---|---|---|---|
| Hut | `hut` | Primitive | 10 wood | +3 pop cap |
| House | `house` | Bronze | wood + stone | +5 pop cap |
| Manor | `manor` | Medieval | stone + steel | +10 pop cap |
| Apartment | `apartment` | Industrial | steel + stone | +20 pop cap |
| Skyscraper | `skyscraper` | Modern | steel + electricity | +50 pop cap |
| Neon Tower | `neon_tower` | Cyberpunk | steel + crypto | +100 pop cap |
| Orbital Habitat | `orbital_habitat` | Space | titanium + plasma | +200 pop cap |

---

## Storage
*Raises resource caps so production doesn't go to waste. All storage buildings increase the cap for every resource.*

| Building | Key | Age | Effect | Max |
|---|---|---|---|---|
| Stash | `stash` | Primitive | +300 all storage | 30 |
| Storage Pit | `storage_pit` | Stone | +500 all storage | — |
| Warehouse | `warehouse` | Bronze | +3,000 all storage | — |
| Granary | `granary` | Iron | +12,000 all storage | — |
| Classical Vault | `classical_vault` | Classical | +25,000 all storage | — |
| Keep | `keep` | Medieval | +60,000 all storage | — |
| Renaissance Vault | `renaissance_vault` | Renaissance | +500,000 all storage | — |
| Colonial Warehouse | `colonial_warehouse` | Colonial | +10M all storage | — |
| Industrial Depot | `industrial_depot` | Industrial | +50M all storage | — |
| Victorian Vault | `victorian_vault` | Victorian | +350M all storage | — |
| Electric Warehouse | `electric_warehouse` | Electric | +3.5B all storage | — |
| Atomic Vault | `atomic_vault` | Atomic | +15B all storage | — |
| Modern Depot | `modern_depot` | Modern | +45B all storage | — |
| Info Vault | `info_vault` | Information | +250B all storage | — |
| Digital Archive | `digital_archive` | Digital | +1.5T all storage | — |
| Cyber Vault | `cyber_vault` | Cyberpunk | +5T all storage | — |
| Fusion Vault | `fusion_vault` | Fusion | +30T all storage | — |
| Orbital Depot | `orbital_depot` | Space | +200T all storage | — |
| Stellar Vault | `stellar_vault` | Interstellar | +500T all storage | — |
| Galactic Vault | `galactic_vault` | Galactic | +2Q all storage | — |
| Quantum Vault | `quantum_vault` | Quantum | +5Q all storage | — |

> **Tip:** Stash is capped at 30. After that, transition to Storage Pit (Stone Age) and Warehouse (Bronze Age). Each tier's storage building is the bottleneck unlock — build one as soon as you enter a new age.

---

## Knowledge & Religion

| Building | Key | Age | Effect |
|---|---|---|---|
| Altar | `altar` | Primitive | +0.5 knowledge/t |
| Firepit | `firepit` | Stone | +0.8 knowledge/t, +0.2 food/t |
| Library | `library` | Bronze | +1.5 knowledge/t, +200 knowledge storage |
| University | `university` | Medieval | +4.0 knowledge/t |
| Cathedral | `cathedral` | Medieval | +2.0 faith/t |
| Observatory | `observatory` | Renaissance | +3.0 knowledge/t, +1.0 culture/t |
| Research Lab | `research_lab` | Atomic | +8.0 knowledge/t |
| AI Lab | `ai_lab` | Digital | +20.0 knowledge/t, +10.0 data/t |

---

## Food Production

| Building | Key | Age | Effect |
|---|---|---|---|
| Gathering Camp | `gathering_camp` | Stone | +0.5 food/t, +0.3 wood/t |
| Farm | `farm` | Bronze | +2.0 food/t |
| Granary | `granary` | Iron | +0.5 food/t, +500 food storage |
| Plantation | `plantation` | Colonial | +5.0 food/t |

---

## Wood & Stone Production

| Building | Key | Age | Effect |
|---|---|---|---|
| Woodcutter Camp | `woodcutter_camp` | Stone | +0.8 wood/t |
| Lumber Mill | `lumber_mill` | Bronze | +2.0 wood/t |
| Stone Pit | `stone_pit` | Stone | +0.6 stone/t |
| Quarry | `quarry` | Bronze | +2.0 stone/t |

---

## Metal & Industry

| Building | Key | Age | Effect |
|---|---|---|---|
| Mine | `mine` | Bronze | +1.0 iron/t |
| Coal Mine | `coal_mine` | Iron | +1.5 coal/t |
| Smithy | `smithy` | Iron | +0.5 steel/t, +0.3 iron/t |
| Steam Engine | `steam_engine` | Industrial | +10% all production |
| Factory | `factory` | Industrial | +5.0 steel/t |
| Steel Mill | `steel_mill` | Industrial | +8.0 steel/t |
| Oil Well | `oil_well` | Victorian | +2.0 oil/t |
| Power Plant | `power_plant` | Electric | +10.0 electricity/t |
| Power Grid | `power_grid` | Electric | +5.0 electricity/t, grid bonus |
| Nuclear Plant | `nuclear_plant` | Atomic | +50.0 electricity/t, +1.0 uranium/t |
| Fusion Reactor | `fusion_reactor` | Fusion | +200.0 electricity/t, +5.0 plasma/t |

---

## Commerce & Gold

| Building | Key | Age | Effect |
|---|---|---|---|
| Market | `market` | Bronze | +0.5 gold/t |
| Forum | `forum` | Classical | +1.0 gold/t, +0.5 culture/t |
| Amphitheater | `amphitheater` | Classical | +1.5 culture/t |
| Bank | `bank` | Renaissance | +3.0 gold/t |
| Stock Exchange | `stock_exchange` | Victorian | +10.0 gold/t |
| Global Bank | `global_bank` | Modern | +50.0 gold/t |
| Crypto Exchange | `crypto_exchange` | Digital | +5.0 crypto/t |
| Black Market | `black_market` | Cyberpunk | +10.0 crypto/t |

---

## Military

| Building | Key | Age | Effect |
|---|---|---|---|
| Barracks | `barracks` | Iron | Enables soldiers; +0.5 defense |
| Castle | `castle` | Medieval | +2.0 defense, +20 soldier cap |
| Keep | `keep` | Medieval | +1.5 defense |
| Missile Silo | `missile_silo` | Atomic | +10.0 defense |
| Bunker | `bunker` | Atomic | +5.0 defense |

---

## Culture

| Building | Key | Age | Effect |
|---|---|---|---|
| Sacred Grove | `sacred_grove` | Primitive | +0.1 culture/t *(also a wonder)* |
| Colosseum | `colosseum` | Iron | +2.0 culture/t *(also a wonder)* |
| Printing Press | `printing_press` | Renaissance | +2.0 culture/t |
| Grand Museum | `grand_museum` | Victorian | +5.0 culture/t |

---

## Late-game buildings (Space+)

| Building | Key | Age | Effect |
|---|---|---|---|
| Launch Pad | `launch_pad` | Space | Enables orbital expeditions |
| Space Station | `space_station` | Space | +10.0 titanium/t |
| Moon Base | `moon_base` | Space | +20.0 titanium/t |
| Warp Gate | `warp_gate` | Interstellar | Enables warp trade |
| Galactic Hub | `galactic_hub` | Galactic | +5.0 dark matter/t |
| Quantum Computer | `quantum_computer` | Quantum | +20.0 quantum flux/t |
| Reality Engine | `reality_engine` | Quantum | +50.0 quantum flux/t |
