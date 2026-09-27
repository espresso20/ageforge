# Wonders

22 unique wonders can be built exactly once per civilization. Each grants permanent civilization-wide bonuses and a **+0.5× speed boost** when completed. They are the most impactful single buildings in the game.

---

## How to build a wonder

Wonders require you to **bank resources** before construction begins. Resources in the wonder bank are reserved and cannot be used elsewhere.

```
wonder collect <resource> <amount>   # bank resources
build <wonder_key>                   # start construction once bank is full
```

Progress is shown in the **Wonders** overlay (`wonders`) with per-resource progress bars. Each completed wonder in the overlay now displays a small colour thumbnail — a 2-character half-block pixel art icon sampled from the wonder's sprite — making it easy to visually identify wonders at a glance.

---

## Viewing Wonders on the City Map

Completed wonders appear on the City Map (`map` command) as their own named, gold-tinted markers — one per wonder you've built — placed among your other buildings on the age-appropriate city layout. As you complete more across the ages, more appear. Open the City Map any time with `map` to see your wonders amid your growing settlement; the whole map (terrain, districts, and labels) also retints with your active theme.

---

## Wonders and Age Advancement

Completing a wonder is **required** to advance to the next age. Each age unlocks one wonder — you cannot type `advance` until that wonder is built.

The wonder requirement appears in the **age progress bar** at the top of the screen alongside your other advancement requirements. If the wonder is still missing, you will see a red notice: `✗ Wonder required: <name>`.

If you try to advance before completing your age's wonder, the game will tell you which wonder is blocking and remind you to use `wonder collect` then `build <key>`.

Once completed, the wonder appears as a named gold marker on the City Map (`map` command), and the city's layout, terrain, and styling shift to reflect your current age.

---

## What wonders cost

Because every wonder stands between you and the next age, each one costs about the same share of its age's economy: **40 price units** of that age. A price unit is the median price of one resource in that age (the typical first-copy price of the age's buildings in that resource). Each wonder keeps its own resource mix; only the size changed.

The flow resources (food, faith, culture) aren't priced that way, so those parts were set by hand to what your buildings actually make: the Sacred Grove takes **500 food**, the Great Monolith **1,500 food**, and the Sistine Chapel **20,000 faith** (down from 6M). The Stellar Cradle no longer costs uranium.

Wonders are banked a deposit at a time, but each part of a wonder's price still fits in the most storage you can build in its age, so you never bank at the cap in rounds. The Sistine Chapel (24M stone and 24M gold, was 27M and 19M) and the World Simulation (34T steel and 20T electricity, was 54T and 6.4T) were rebalanced to fit.

No wonder takes longer to build than a sixth of its age's target length (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)). Build times below are at 1x speed; one tick is 2 seconds.

---

## Wonder list

### 🌿 Sacred Grove
**Age:** Primitive · **Key:** `sacred_grove` · **Build:** 75 ticks (2m 30s)

| Resource | Cost |
|---|---|
| Wood | 1,000 |
| Food | 500 |

**Bonus:** +0.02 knowledge/t · +0.05 food/t

---

### 🗿 Great Monolith
**Age:** Stone · **Key:** `great_monolith` · **Build:** 225 ticks (7m 30s)

| Resource | Cost |
|---|---|
| Stone | 6,300 |
| Wood | 5,000 |
| Food | 1,500 |

**Bonus:** +0.05 knowledge/t · +5,000 all storage

---

### ⭕ Stonehenge
**Age:** Bronze · **Key:** `stonehenge` · **Build:** 450 ticks (15m)

| Resource | Cost |
|---|---|
| Stone | 34,000 |
| Wood | 19,000 |
| Iron | 3,400 |

**Bonus:** +0.8 knowledge/t · +0.6 faith/t

---

### 🏟 Colosseum
**Age:** Iron · **Key:** `colosseum` · **Build:** 750 ticks (25m)

| Resource | Cost |
|---|---|
| Stone | 320,000 |
| Iron | 72,000 |
| Gold | 64,000 |

**Bonus:** +100 population cap · +2.0 culture/t

---

### 🏛 Parthenon
**Age:** Classical · **Key:** `parthenon` · **Build:** 1,050 ticks (35m)

| Resource | Cost |
|---|---|
| Stone | 1M |
| Gold | 440,000 |
| Iron | 440,000 |

**Bonus:** +2.0 culture/t · +1.2 knowledge/t

---

### 📚 Great Library
**Age:** Medieval · **Key:** `great_library` · **Build:** 1,350 ticks (45m)

| Resource | Cost |
|---|---|
| Stone | 3.5M |
| Gold | 2.7M |
| Knowledge | 840,000 |

**Bonus:** +2.0 knowledge/t · **+30% knowledge rate** (permanent multiplier)

---

### 🎨 Sistine Chapel
**Age:** Renaissance · **Key:** `sistine_chapel` · **Build:** 1,800 ticks (1h)

| Resource | Cost |
|---|---|
| Stone | 24M |
| Gold | 24M |
| Faith | 20,000 |
| Culture | 8M |

**Bonus:** +3.5 culture/t · +1.8 faith/t

---

### 🏮 Grand Lighthouse
**Age:** Colonial · **Key:** `grand_lighthouse` · **Build:** 2,100 ticks (1h 10m)

| Resource | Cost |
|---|---|
| Stone | 160M |
| Gold | 120M |
| Steel | 23M |

**Bonus:** +5.0 gold/t · **+80% expedition reward** — the best military wonder

---

### 🏗 Crystal Palace
**Age:** Industrial · **Key:** `crystal_palace` · **Build:** 2,400 ticks (1h 20m)

| Resource | Cost |
|---|---|
| Steel | 450M |
| Iron | 400M |
| Gold | 310M |
| Coal | 400M |

**Bonus:** **+15% all production** · +8.0 gold/t

---

### 🗼 Eiffel Tower
**Age:** Victorian · **Key:** `eiffel_tower` · **Build:** 2,700 ticks (1h 30m)

| Resource | Cost |
|---|---|
| Steel | 4.7B |
| Iron | 3.8B |
| Gold | 5.1B |

**Bonus:** +5.0 culture/t · +2.0 knowledge/t

---

### 🌊 Hoover Dam
**Age:** Electric · **Key:** `hoover_dam` · **Build:** 3,000 ticks (1h 40m)

| Resource | Cost |
|---|---|
| Steel | 71B |
| Stone | 50B |
| Electricity | 20B |

**Bonus:** +10.0 electricity/t · **+20% all production**

---

### ⚛️ Particle Accelerator
**Age:** Atomic · **Key:** `particle_accelerator` · **Build:** 3,600 ticks (2h)

| Resource | Cost |
|---|---|
| Steel | 100B |
| Electricity | 140B |
| Uranium | 16B |

**Bonus:** +10.0 knowledge/t · +1.5 uranium/t

---

### 🚀 Space Program
**Age:** Modern · **Key:** `space_program` · **Build:** 3,600 ticks (2h)

| Resource | Cost |
|---|---|
| Steel | 770B |
| Gold | 690B |
| Electricity | 430B |
| Knowledge | 600B |

**Bonus:** +6.0 knowledge/t · +8.0 culture/t

---

### 🌐 Global Network
**Age:** Information · **Key:** `global_network` · **Build:** 4,200 ticks (2h 20m)

| Resource | Cost |
|---|---|
| Steel | 4.8T |
| Data | 580B |
| Electricity | 960B |
| Gold | 2.4T |

**Bonus:** +30.0 data/t · **+30% knowledge rate**

---

### 💻 World Simulation
**Age:** Digital · **Key:** `world_simulation` · **Build:** 4,800 ticks (2h 40m)

| Resource | Cost |
|---|---|
| Steel | 34T |
| Data | 1T |
| Electricity | 20T |

**Bonus:** +60.0 data/t · +15.0 knowledge/t

---

### 🌆 Neon Citadel
**Age:** Cyberpunk · **Key:** `neon_citadel` · **Build:** 5,400 ticks (3h)

| Resource | Cost |
|---|---|
| Steel | 100T |
| Electricity | 77T |
| Crypto | 12T |
| Data | 6.4T |

**Bonus:** +10.0 crypto/t · **+500 population cap**

---

### ☀️ Stellar Cradle
**Age:** Fusion · **Key:** `stellar_cradle` · **Build:** 6,000 ticks (3h 20m)

| Resource | Cost |
|---|---|
| Steel | 430T |
| Plasma | 340T |
| Electricity | 460T |

**Bonus:** +15.0 plasma/t · +200.0 electricity/t

---

### 🛰 Dyson Scaffold
**Age:** Space · **Key:** `dyson_scaffold` · **Build:** 6,600 ticks (3h 40m)

| Resource | Cost |
|---|---|
| Titanium | 910T |
| Plasma | 660T |
| Steel | 5.6Q |

**Bonus:** +200.0 electricity/t · +30.0 plasma/t

---

### 🌀 Warp Nexus
**Age:** Interstellar · **Key:** `warp_nexus` · **Build:** 7,200 ticks (4h)

| Resource | Cost |
|---|---|
| Titanium | 31Q |
| Dark Matter | 2.2Q |
| Plasma | 29Q |

**Bonus:** +8.0 dark matter/t · **+80% all production**

---

### 🌌 Cosmic Beacon
**Age:** Galactic · **Key:** `cosmic_beacon` · **Build:** 7,200 ticks (4h)

| Resource | Cost |
|---|---|
| Dark Matter | 23Q |
| Antimatter | 17Q |
| Titanium | 65Q |

**Bonus:** +10.0 antimatter/t · **+50% all production**

---

### ⚡ Reality Anchor
**Age:** Quantum · **Key:** `reality_anchor` · **Build:** 7,200 ticks (4h)

| Resource | Cost |
|---|---|
| Quantum Flux | 24Q |
| Antimatter | 39Q |
| Dark Matter | 44Q |

**Bonus:** +15.0 quantum flux/t · **+50% all production**

---

### ✨ Singularity Core
**Age:** Transcendent · **Key:** `singularity_core` · **Build:** 7,200 ticks (4h)

| Resource | Cost |
|---|---|
| Quantum Flux | 360Q |
| Antimatter | 370Q |
| Dark Matter | 370Q |

**Bonus:** **+200% all production** · +20.0 quantum flux/t

> The Singularity Core is the ultimate wonder: the biggest bill in the game and the biggest single production bonus.

---

## Wonders and Morale

Every wonder you build raises your **morale cap** by +5%. Morale is a civilization-wide multiplier on all worker output; it starts at 50% neutral, only produces a bonus once it climbs above 75%, and reaches up to **+20%** to all production as it approaches the cap. The base cap (0 wonders) is **100%**, and each wonder lifts that ceiling — so a higher cap means a higher reachable bonus.

- 5 wonders built → morale cap **125%**
- 10 wonders built → morale cap **150%**
- 22 wonders built → morale cap **210%**

These figures are the morale **ceiling** — how high the morale percentage can climb — not a direct output multiplier. The production bonus from the high band still tops out at +20% as morale nears that ceiling; a higher cap doesn't raise the +20% bonus, but it makes the high-band bonus zone larger and easier to sit in. Morale starts at 50% neutral and must be raised into the high band with worship and culture buildings — wonders simply set how high it can go. Full-completion runs (all 22 wonders) unlock the maximum 210% morale ceiling.

---

## Tips

- Always bank resources across multiple ticks — don't try to dump everything at once
- **Great Library** and **Global Network** both grant +30% knowledge rate — research becomes exponentially faster with both built
- **Grand Lighthouse** + Rocketry tech + prestige `expedition_loot` = absurd expedition loot multipliers
- Build wonders as early as possible in each age — the speed boost fires when construction completes and helps you hit the next age faster
- **Crystal Palace** (+15% all production) at the Industrial Age is often the single biggest inflection point in the game
- Each wonder also raises the morale cap by +5% (100% base → 210% with all 22) — prioritising wonder completion pays dividends in both direct bonuses and morale headroom. See [Morale](morale.md)
