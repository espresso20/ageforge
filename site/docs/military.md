# Military & Expeditions

The military system is one of AgeForge's primary engines of wealth. Soldiers defend your empire from hostile events, unlock progressively richer expeditions, and gate several milestones and prestige bonuses. If you ignore it, catastrophe events will eat your stockpiles. If you invest in it, expeditions flood you with resources you couldn't produce fast enough on your own.

---

## 1. Overview

The military system does four things:

- **Expeditions** — send soldiers out for timed missions that return resources, gold, and knowledge on success.
- **Defense rating** — a passive stat derived from soldier count and military bonuses that reduces damage from hostile epoch events.
- **Milestones** — five military milestones grant permanent `military_power` bonuses and chain into a title reward.
- **Prestige** — prestige upgrades `military_power` and `expedition_loot` carry over through resets, compounding across runs.

Soldiers are workers from the **military domain**. They are recruited the same way as any worker, then assigned to military buildings. The military domain unlocks at **Iron Age** — you cannot assign workers to military buildings before then.

---

## 2. Soldiers

### What soldiers are

Soldiers are workers assigned to military-domain buildings. Your soldier count equals the number of workers the game reports under `WorkerManager.GetDomainCount("military")`. They consume food every tick at the current military worker class food cost (base **2.0/tick** at Iron Age, scaling by ×1.5 per tier as you advance through ages).

### Recruiting

```
recruit [count|max]
```

- `recruit 5` — recruits 5 new generic workers from available housing capacity.
- `recruit max` — fills every available housing slot instantly.
- Workers remain generic until assigned to a building; assignment to a military building makes them soldiers.

After recruiting, assign them:

```
assign war_camp 5
assign barracks all
```

### Food drain

Military workers eat at the base rate for their current class. At **Iron Age** that is 2.0 food/tick per soldier. The rate scales geometrically (×1.5 per tier) as you advance through ages. A large standing army in the late game has a meaningful food cost — plan your food production accordingly before enlisting.

### Losing soldiers

Soldiers can be lost in two ways:

- **Expedition failure** — a failed expedition costs 1–2 soldiers (random). Even a success has a small chance of 1 soldier lost (probability: `difficulty × 0.3`).
- **Catastrophe events** — certain epoch events that destroy buildings can also reduce your workforce.

Soldiers who die are simply removed from the worker pool. They do not need to be "buried" or replaced automatically — you recruit replacements manually.

### Viewing your army

```
army
```

Opens the army overlay showing current soldier count, defense rating, active expedition (if any), and completed expedition count.

---

## 3. Military Buildings

Military buildings do two things: they provide **worker slots** (which determines how many soldiers you can have) and they each grant a **+N military capacity** effect. You cannot field more soldiers than your total military capacity allows.

All military buildings belong to the **military lineage** (22 tiers, stone_age through transcendent_age). CostScale is 1.15 — costs rise gently with each additional copy, so scaling up an army stays affordable.

| Tier | Key | Name | Age Unlocked | Soldier Cap | Worker Slots |
|------|-----|------|-------------|-------------|--------------|
| 0 | `war_camp` | War Camp | Stone Age | +10 | 3 |
| 1 | `barracks` | Barracks | Bronze Age | +20 | 4 |
| 2 | `hunting_lodge` | Hunting Lodge | Iron Age | +40 | 5 |
| 3 | `legion_fort` | Legion Fort | Iron Age | +80 | 6 |
| 4 | `military_academy` | Military Academy | Classical Age | +160 | 6 |
| 5 | `castle_keep` | Castle Keep | Medieval Age | +320 | 7 |
| 6 | `fortress` | Fortress | Renaissance Age | +640 | 7 |
| 7 | `fort` | Fort | Colonial Age | +1,280 | 8 |
| 8 | `military_base` | Military Base | Industrial Age | +2,560 | 10 |
| 9 | `garrison` | Garrison | Victorian Age | +5,120 | 10 |
| 10 | `command_post` | Command Post | Electric Age | +10,240 | 12 |
| 11 | `bunker_complex` | Bunker Complex | Atomic Age | +20,480 | 12 |
| 12 | `special_ops_hq` | Special Ops HQ | Modern Age | +40,960 | 14 |
| 13 | `cyber_command` | Cyber Command | Information Age | +81,920 | 15 |
| 14 | `drone_warfare_center` | Drone Warfare Center | Digital Age | +163,840 | 16 |
| 15 | `combat_aug_center` | Combat Aug Center | Cyberpunk Age | +327,680 | 18 |
| 16 | `plasma_command` | Plasma Command | Fusion Age | +655,360 | 20 |
| 17 | `space_force_base` | Space Force Base | Space Age | +1,310,720 | 20 |
| 18 | `fleet_command` | Fleet Command | Interstellar Age | +2,621,440 | 25 |
| 19 | `stellar_armada_hq` | Stellar Armada HQ | Galactic Age | +5,242,880 | 25 |
| 20 | `probability_war_room` | Probability War Room | Quantum Age | +10,485,760 | 30 |
| 21 | `omniversal_war_council` | Omniversal War Council | Transcendent Age | +20,971,520 | 35 |

**Building for capacity:** The soldier cap from a building applies per instance. Two Barracks = 40 military capacity total. Stack them to increase your army ceiling before attempting high-soldier expeditions.

---

## 4. Expeditions

### What expeditions are

Expeditions are timed missions that consume no resources to launch (only soldier presence is required). After a set number of ticks, they resolve — either succeeding for full loot or failing for partial loot plus soldier losses. Only **one expedition can be active at a time**.

### Commands

```
expedition list         # Show all expeditions available in your current age
expedition <key>        # Launch an expedition (e.g., expedition scout_ruins)
exp list                # Shorthand
exp <key>               # Shorthand
```

Keys with underscores can be typed with spaces and are joined: `expedition scout ruins` is equivalent to `expedition scout_ruins`.

### Success probability formula

From `game/military.go`:

```
difficulty = DifficultyBase - (militaryBonus × 0.3)
difficulty = max(difficulty, 0.05)
success = rand() > difficulty
```

A higher `militaryBonus` reduces the effective difficulty. With zero bonus, a 0.8-difficulty expedition succeeds only ~20% of the time. With a +2.0 military bonus, effective difficulty is clamped to 0.05, giving ~95% success.

**On success:** Full rewards × (1 + expeditionBonus). Small chance of 1 soldier lost: `rand() < difficulty × 0.3`.

**On failure:** 30% of rewards awarded. 1–2 soldiers lost (random, capped at expedition's soldier count).

### Full expedition table

| Key | Name | Min Age | Soldiers | Duration | Difficulty | Rewards on Success |
|-----|------|---------|----------|----------|------------|-------------------|
| `scout_ruins` | Scout Nearby Ruins | Bronze Age | 2 | 10t | 0.20 | 30 food, 20 wood, 15 stone |
| `raid_bandits` | Raid Bandit Camp | Bronze Age | 5 | 15t | 0.40 | 30 gold, 15 iron, 20 food |
| `trade_escort` | Trade Escort | Iron Age | 3 | 12t | 0.30 | 50 gold, 10 knowledge |
| `conquer_territory` | Conquer Territory | Iron Age | 10 | 25t | 0.60 | 80 gold, 40 iron, 50 food |
| `siege_castle` | Siege Enemy Castle | Medieval Age | 15 | 30t | 0.70 | 150 gold, 30 steel, 20 faith |
| `naval_expedition` | Naval Expedition | Renaissance Age | 10 | 35t | 0.50 | 200 gold, 30 culture, 40 knowledge |
| `colonial_campaign` | Colonial Campaign | Industrial Age | 20 | 40t | 0.60 | 300 gold, 50 oil, 40 steel |
| `world_domination` | World Domination | Modern Age | 50 | 60t | 0.80 | 1,000 gold, 200 electricity, 500 knowledge |
| `cyber_raid` | Cyber Raid | Information Age | 30 | 45t | 0.60 | 200 data, 50 crypto, 500 gold |
| `neon_heist` | Neon Heist | Cyberpunk Age | 25 | 35t | 0.55 | 100 crypto, 150 data, 800 gold |
| `fusion_assault` | Fusion Plant Assault | Fusion Age | 35 | 40t | 0.65 | 120 plasma, 500 electricity, 50 uranium |
| `orbital_strike` | Orbital Strike | Space Age | 40 | 50t | 0.70 | 100 titanium, 80 plasma, 300 knowledge |
| `warp_invasion` | Warp Invasion | Interstellar Age | 60 | 65t | 0.75 | 50 dark matter, 200 titanium, 2,000 gold |
| `galactic_conquest` | Galactic Conquest | Galactic Age | 80 | 80t | 0.80 | 30 antimatter, 100 dark matter, 5,000 gold |
| `quantum_incursion` | Quantum Incursion | Quantum Age | 100 | 90t | 0.85 | 20 quantum flux, 50 antimatter, 5,000 knowledge |

**Tip:** `scout_ruins` and `trade_escort` are the workhorses of early play — low difficulty, short duration, and they can be chained rapidly for consistent loot.

---

## 5. Military Domain Workers

Military workers evolve their class name as you advance through ages. The base food cost is **2.0/tick** at Iron Age and scales ×1.5 per tier:

| Age | Class Name | Food Cost/tick |
|-----|-----------|----------------|
| Iron Age | Soldier | 2.00 |
| Classical Age | Legionary | 3.00 |
| Medieval Age | Knight | 4.50 |
| Renaissance Age | Musketeer | 6.75 |
| Colonial Age | Colonial Marine | 10.13 |
| Industrial Age | Industrial Rifleman | 15.19 |
| Victorian Age | Victorian Guard | 22.78 |
| Electric Age | Electric Trooper | 34.17 |
| Atomic Age | Atomic Soldier | 51.26 |
| Modern Age | Modern Soldier | 76.89 |
| Information Age | Information Warrior | 115.33 |
| Digital Age | Digital Soldier | 172.99 |
| Cyberpunk Age | Cyber Warrior | 259.49 |
| Fusion Age | Plasma Trooper | 389.23 |
| Space Age | Space Marine | 583.85 |
| Interstellar Age | Interstellar Commando | 875.77 |
| Galactic Age | Galactic Guardian | 1,313.66 |
| Quantum Age | Quantum Soldier | 1,970.49 |

### Assignment

```
assign <building_key> [count|all]

# Examples:
assign war_camp 3
assign barracks all
assign castle_keep 7
```

Workers assigned to military buildings count as soldiers. A building must exist (count > 0) before workers can be assigned — the game blocks assignment to unbuilt structures.

Worker assignment affects **capacity scaling**: buildings run at `20% + 80% × (assigned / totalCapacity)` of their effect value. Fill your military buildings to full worker capacity to maximise military effectiveness.

---

## 6. Military Power Bonus

`military_power` is a cumulative float that feeds directly into the expedition success formula and the defense rating calculation:

```
defense = soldierCount × 2.0 × (1 + militaryBonus)
```

Sources, stacked additively:

| Source | How to get it | Bonus per step |
|--------|-------------|---------------|
| **Research techs** | Various military-flavored techs grant `military_power` bonus | +0.2 to +1.5 per tech |
| **Permanent bonuses** (milestones) | Complete military milestones | +0.05 to +0.15 each |
| **Prestige upgrade** | `prestige buy military_power` (5 tiers, 2/3/5/8/10 pts each) | +5% per tier |

There is no hard cap on `military_power`, but effective expedition difficulty is floored at **0.05** (5% chance of failure minimum), so stacking beyond ~2.5–3.0 bonus yields diminishing returns against failure rates.

The `expedition_reward` bonus (from research, prestige `expedition_loot`, and certain wonders) is separate and multiplies the loot amount on success: `rewards × (1 + expeditionBonus)`.

---

## 7. Military Milestones

The five military milestones form a chain. Completing the full chain grants a permanent title and a cumulative **+0.50 military_power** bonus.

| Key | Name | Requirement | Age Gate | Reward |
|-----|------|-------------|----------|--------|
| `first_soldiers` | First Soldiers | 5 soldiers | Iron Age | +0.05 military_power |
| `war_machine` | War Machine | 250 soldiers | Iron Age | +0.10 military_power |
| `iron_legion` | Iron Legion | 500 soldiers + 10 Barracks | Classical Age | +0.10 military_power |
| `fortress_state` | Fortress State | 20 Castle Keeps | Medieval Age | +0.10 military_power |
| `military_superpower` | Military Superpower | 2,000 soldiers | Industrial Age | +0.15 military_power |

`iron_legion`, `fortress_state`, and `military_superpower` are **hidden** until their prerequisites are visible (progress > 50% or you're in the preceding age). Don't be surprised when they appear mid-game.

> **Note:** Standing Army (100 soldiers + 10 Barracks, Classical Age, +0.05 military_power) is a standalone military milestone — it is not part of the chain.

---

## 8. Strategy

### Early game (Iron Age to Classical Age)

- Build a **War Camp** in the Stone Age even though soldiers aren't possible yet — it prepares your capacity.
- Your first soldiers become available in **Iron Age** via the Hunting Lodge. Train 5 quickly to land `first_soldiers` for the free +0.05 bonus.
- `trade_escort` (3 soldiers, Iron Age) is your best early expedition — low soldiers needed, fast duration (12t), reasonable gold reward.
- Keep food workers prioritised. Military workers eat 2.0 food/tick each at this stage — a 10-soldier army demands 20 food/tick just to sustain itself.

### Mid game (Classical to Industrial Age)

- Push `iron_legion` (300 soldiers + 5 Barracks) — the +0.10 bonus noticeably improves expedition success on harder missions.
- `conquer_territory` (10 soldiers, 25t, 0.60 difficulty) and `naval_expedition` (10 soldiers, 35t, 0.50 difficulty) are the best bang-for-tick in this range.
- Build **Legion Forts** and **Military Academies** to raise your soldier ceiling. The capacity doubling per tier means each new building unlocks dramatically more troops.
- Research military-flavored techs as they appear — even +0.2 military_power makes a visible difference against 0.6-difficulty expeditions.

### Late game (Modern Age onward)

- `world_domination` requires 50 soldiers but pays 1,000 gold — worth the queue for gold-hungry ages.
- `cyber_raid` and `neon_heist` are the best value in the digital/cyberpunk range. `neon_heist` (0.55 difficulty) is easier than `cyber_raid` (0.60) for comparable loot.
- Buy `expedition_loot` prestige upgrades across resets — at tier 5 (+25% rewards) stacked with research bonuses, expedition returns scale dramatically.
- The `military_superpower` milestone (+0.15 bonus) combined with late-game research can push effective difficulty on most expeditions to the 0.05 floor.

### Catastrophe defense

Your **defense rating** (`soldierCount × 2.0 × (1 + militaryBonus)`) is checked against hostile epoch events. A higher defense rating reduces resource losses from events like Bandit Raid, Pirate Attack, and rival aggression. It does not prevent catastrophe events outright — but it significantly reduces the damage.

### Balancing army size vs food drain

Each soldier eats at the current military class food cost (2.0/tick at Iron Age, scaling up aggressively). A rough rule of thumb:

- Before recruiting, confirm your food net rate (check `rates`) is positive with the projected soldier drain added.
- Use `recruit max` only when food is overflowing and storage is near cap — don't spike your army into a food deficit.
- After major age advances, review food drain: the ×1.5 per tier scaling means your army costs 50% more food per tick in each successive age.

### Morale and Military Ratio

Maintaining a large standing army has a second cost beyond food: **morale drain**. If military workers exceed **30% of your total population**, morale decays at −0.003/tick for every 10% over the threshold:

| Military ratio | Overage | Morale drain/tick |
|---------------|---------|------------------|
| 30% (at threshold) | 0% | 0 |
| 40% | +10% | −0.003/tick |
| 50% | +20% | −0.006/tick |
| 60% | +30% | −0.009/tick |

Morale is a multiplier on all worker output (floor 0.10). Sustained morale drain from an oversized army gradually suppresses production across every domain — food, knowledge, trade, everything — creating a feedback loop where the army's food cost becomes even harder to cover as food workers produce less.

**Recommended target: keep military workers below 25–28% of total population.** This provides a comfortable buffer against the threshold even if population fluctuates from worker loss events. If you need a large army for a high-soldier expedition, build it temporarily, run the expedition, then unassign excess soldiers back to civilian buildings.

Large armies require strong food production and a robust civilian workforce to offset the morale penalty. If you see morale trending downward and your military ratio is over 30%, unassign some soldiers or recruit more civilians before the drain compounds further.

---

## 9. Tips & Common Mistakes

**Don't recruit past your food income.** A food deficit stalls all production (workers can't work when starving). Calculate the drain before `recruit max`.

**Don't launch high-difficulty expeditions without military power.** `siege_castle` (0.70) and `world_domination` (0.80) fail frequently with zero bonus. Research a few military techs first and watch the success probability change.

**Expedition failure isn't free.** A failed run returns only 30% of rewards and costs you 1–2 soldiers. At 2.0 food/tick per soldier, a dead soldier is a food drain removed — but replacing them costs recruit time and building capacity.

**Workers must be assigned to built buildings.** The game blocks assignment if the building count is zero. Build the structure before trying to assign soldiers to it.

**Castle Keep timing matters for `fortress_state`.** You need 10 Castle Keeps (Medieval Age). That's a serious stone and iron investment — start queuing them as soon as you hit Medieval. The +0.10 bonus is worth the build cost several times over in improved expedition returns.

**Run expeditions continuously.** There's no cooldown between expeditions beyond the active duration. The moment one resolves, launch the next. Idle military is lost throughput.

**Prestige compounds military strength.** `military_power` prestige upgrade (5 tiers × 5% = +25%) and `expedition_loot` (5 tiers × 5% = +25%) both persist through resets. Prioritise these in the prestige shop on your second and third runs.
