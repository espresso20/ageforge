# Military & Expeditions

The military system is one of AgeForge's primary engines of wealth. Soldiers defend your empire from hostile events, unlock progressively richer expeditions, and gate several milestones and prestige bonuses. If you ignore it, catastrophe events will eat your stockpiles. If you invest in it, expeditions flood you with resources you couldn't produce fast enough on your own.

---

## 1. Overview

The military system does four things:

- **Campaigns** — spend stockpiled soldiers on timed military missions that return resources, gold, and knowledge on success (waged with `campaign <key>`). Soldier-free **scouting expeditions** (`expedition <key>`) cover the same ground without troops — see [§4](#4-expeditions).
- **Defense rating** — a passive stat derived from soldier count and military bonuses that reduces damage from hostile epoch events.
- **Milestones** — five military milestones chain into a title reward; the early tiers grant permanent `military_power`, while the late tiers now pay out broad `all production`.
- **Prestige** — prestige upgrades `military_power` and `expedition_loot` carry over through resets, compounding across runs.

**Soldiers are a resource** (the 26th), not a worker domain count. They are *produced and stored by your military buildings*: assign military-domain workers to a War Camp or Barracks and it generates the `soldiers` resource every tick, the same way a Farm generates food. The military domain — and the soldiers resource — unlock at the **Iron Age**. Before then, the **scouting** expeditions (see [§4](#4-expeditions)) let you explore for resources without any soldiers at all.

Expeditions come in **two kinds**:

- **Scouting** — cost only resources (no soldiers), and are available early, before the soldiers resource exists at the Iron Age. There are exactly three: `scout_party`, `scout_ruins`, and `naval_expedition`.
- **Military campaigns** — cost soldiers (plus any resource `Cost`), and are gated behind the Iron Age and beyond.

---

## 2. Soldiers

### What soldiers are

Soldiers are a stockpiled **resource** (`soldiers`), unlocked at the **Iron Age**. You don't count heads in a worker pool — you bank soldiers the way you bank food or wood, then spend them waging campaigns.

- **Production:** Military buildings produce soldiers every tick when military-domain workers are assigned to them. Production is worker-scaled — a fully-staffed military building produces roughly its **soldier cap ÷ 50** soldiers per tick (minimum 0.1/tick). Assign more military workers, build more military buildings → soldiers accrue faster.
- **Storage:** Your soldier cap equals the **sum of every military building's soldier cap**. Building or upgrading military buildings is the *only* way to raise the ceiling — the Storage lineage does not hold soldiers.
- **Spending:** Military campaigns deduct their soldier cost from your stockpile at launch (see [§4](#4-expeditions)).

### Producing soldiers

The soldiers resource is produced by the military workers you staff into military buildings. The workflow:

```
recruit [count|max]     # recruit generic workers from housing capacity
assign war_camp 5       # staff them into a military building → it starts producing soldiers
assign barracks all     # fill a building to capacity for maximum soldier output
```

- `recruit 5` — recruits 5 new generic workers from available housing capacity.
- `recruit max` — fills every available housing slot instantly.
- Workers remain generic until assigned to a building. Assigning them to a military building turns them into military-class workers, who consume food and produce the `soldiers` resource each tick.

### Food drain (military workers)

The military *workers* who produce soldiers eat food at the base rate for their current class — **2.0 food/tick** each at Iron Age, scaling geometrically (×1.5 per tier) as you advance through ages. The soldiers resource itself has no upkeep once produced; only the workforce generating it costs food. A large soldier-producing operation in the late game has a meaningful food cost, so plan your food production accordingly before staffing up.

### Losing soldiers

Soldiers leave your stockpile in two ways:

- **Expedition launch** — every *military campaign* spends its soldier cost up front, deducted whether the run later succeeds or fails. The three **scouting** expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) cost **0** soldiers and instead charge a resource `Cost` — see [§4](#4-expeditions).
- **Catastrophe events** — certain hostile epoch events can drain stored soldiers along with other losses.

Spent or lost soldiers are simply removed from the stockpile. To replenish, keep military workers assigned to your military buildings so production continues.

### Viewing your army

```
army
```

Opens the **Army** panel showing current soldier count, defense rating, the active military **campaign** (if any), and completed campaign count. The active scouting expedition lives in the separate **Expeditions** panel (`expedition`) — one scouting expedition and one military campaign can run concurrently, one of each category.

---

## 3. Military Buildings

Military buildings do three things: they provide **worker slots** for military-domain workers, they **produce the soldiers resource** every tick while staffed, and their per-instance **soldier cap** sets how many soldiers you can store. Your total soldier storage is the sum of all your military buildings' soldier caps — you cannot bank more soldiers than that combined cap allows.

All military buildings belong to the **military lineage** (22 tiers, stone_age through transcendent_age). CostScale is 1.15 — costs rise gently with each additional copy, so scaling up an army stays affordable.

**Soldier production per building:** A fully-staffed military building produces about **its soldier cap ÷ 50 per tick** (minimum 0.1/tick). A staffed Barracks (cap 20) yields ~0.4 soldiers/tick; a Castle Keep (cap 320) yields ~6.4/tick. Production scales with worker assignment exactly like every other domain — see [§5](#5-military-domain-workers).

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

**Building for capacity and output:** The soldier cap applies per instance — two Barracks = 40 soldier storage total *and* doubled soldier production when both are staffed. Stack military buildings to raise both your soldier ceiling and your soldiers-per-tick rate before attempting high-soldier expeditions.

---

## 4. Expeditions

### What expeditions are

Expeditions are timed missions launched from your stockpiles. They come in **two kinds**:

- **Scouting** (`scout_party`, `scout_ruins`, `naval_expedition`) — cost only a resource `Cost`, **0 soldiers**. These are available early, before the `soldiers` resource exists at the Iron Age.
- **Military campaigns** (everything else — 13 of them) — **spend the soldiers resource** to launch, plus any additional resource `Cost`.

Whatever the kind, the cost is deducted from your stockpiles the moment you launch — there's no refund. After a set number of ticks, the mission resolves. **Success and failure differ only in the size of the reward, not the cost:** the soldiers and/or resources are spent either way. A successful run pays full loot; a failed run pays a reduced amount. **One mission of each category can be active at a time:** a scouting expedition and a military campaign can run concurrently, but you can't launch a second of the same category while one is still in progress.

### Cost, gating, and rewards

- **Soldier cost** — military campaigns spend it from the `soldiers` resource at launch, and you cannot launch one you can't afford. The three scouting expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) cost **0 soldiers** — they charge only a resource `Cost`.
- **Resource `Cost`** — many expeditions charge resources up front (e.g. `scout_party` costs 30 food + 30 wood; `naval_expedition` costs 150 food + 100 wood). Deducted at launch alongside any soldier cost.
- **Age gating** — every expedition has a `MinAge`. Some also have a `MaxAge` and disappear once you advance past it. `scout_party` is gated to primitive → bronze ages (it vanishes at Iron Age, when real soldiers become available); `scout_ruins` opens at Bronze Age and `naval_expedition` at Renaissance Age — both scouting, both 0 soldiers.
- **Reward** — paid on resolution: full amount on success, reduced amount on failure. The soldier and resource costs are *not* part of the reward calculation; they're already gone.

### Commands

Scouting and military missions live on two separate command surfaces:

```
expedition list         # Show scouting expeditions available in your current age
expedition <key>        # Send a scouting expedition (e.g., expedition scout_ruins) — costs resources, never soldiers
exp list                # Shorthand
exp <key>               # Shorthand

campaign list           # Show military campaigns available in your current age
campaign <key>          # Wage a military campaign (e.g., campaign raid_bandits) — spends soldiers
```

`expedition` with no args opens the **Expeditions** panel; `campaign` (or `campaign list`) lists the available campaigns; `army` opens the **Army** panel. Keys with underscores can be typed with spaces and are joined: `expedition scout ruins` is equivalent to `expedition scout_ruins`.

The two surfaces are kept distinct, and the game cross-redirects if you mix them up: running a military key through `expedition` (e.g. `expedition raid_bandits`) is refused with a note to use `campaign <key>` instead, and running a scouting key through `campaign` redirects you to `expedition <key>`.

The **Expeditions** panel (`expedition`) lists only the **scouting** expeditions available in your current age — the resource-only missions. The **Army** panel (`army`) lists only the **military campaigns** — the ones that spend soldiers — alongside your soldier count and defense rating. So you can see at a glance which missions are which: scouting on the Expeditions panel, campaigns on the Army panel.

### Success probability formula

From `game/military.go`:

```
difficulty = DifficultyBase - (militaryBonus × 0.3)
difficulty = max(difficulty, 0.05)
success = rand() > difficulty
```

A higher `militaryBonus` reduces the effective difficulty. With zero bonus, a 0.8-difficulty expedition succeeds only ~20% of the time. With a +2.0 military bonus, effective difficulty is clamped to 0.05, giving ~95% success.

The soldier (and resource) cost is already spent at launch, so the outcome only scales the **reward**:

**On success:** Full rewards × (1 + expeditionBonus).

**On failure:** A reduced fraction of the rewards is awarded. No additional soldiers are deducted on failure — the launch cost is the entire cost, win or lose.

### Full expedition table

This is a reference table of **all** missions, scouting and campaigns alike. The three **scouting** expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) are launched with `expedition <key>`; every other mission is a **campaign**, launched with `campaign <key>`. The **Soldier Cost** column is spent from the `soldiers` resource at launch (not a soldier headcount requirement). The scouting rows show **0** in that column and instead charge a resource `Cost`; the campaigns spend soldiers plus any resource `Cost`.

| Key | Name | Min Age | Soldier Cost | Resource Cost | Duration | Difficulty | Rewards on Success |
|-----|------|---------|--------------|---------------|----------|------------|-------------------|
| `scout_party` | Scout Party | Primitive Age (max Bronze Age) | 0 | 30 food, 30 wood | 20t | — | 60 food, 60 wood, 20 stone |
| `scout_ruins` | Scout Nearby Ruins | Bronze Age | 0 | 40 food, 30 wood | 10t | 0.20 | 30 food, 20 wood, 15 stone |
| `raid_bandits` | Raid Bandit Camp | Bronze Age | 5 | — | 15t | 0.40 | 30 gold, 15 iron, 20 food |
| `trade_escort` | Trade Escort | Iron Age | 3 | — | 12t | 0.30 | 50 gold, 10 knowledge |
| `conquer_territory` | Conquer Territory | Iron Age | 10 | — | 25t | 0.60 | 80 gold, 40 iron, 50 food |
| `siege_castle` | Siege Enemy Castle | Medieval Age | 15 | — | 30t | 0.70 | 150 gold, 30 steel, 20 faith |
| `naval_expedition` | Naval Expedition | Renaissance Age | 0 | 150 food, 100 wood | 35t | 0.50 | 200 gold, 30 culture, 40 knowledge |
| `colonial_campaign` | Colonial Campaign | Industrial Age | 20 | — | 40t | 0.60 | 300 gold, 50 oil, 40 steel |
| `world_domination` | World Domination | Modern Age | 50 | — | 60t | 0.80 | 1,000 gold, 200 electricity, 500 knowledge |
| `cyber_raid` | Cyber Raid | Information Age | 30 | — | 45t | 0.60 | 200 data, 50 crypto, 500 gold |
| `neon_heist` | Neon Heist | Cyberpunk Age | 25 | — | 35t | 0.55 | 100 crypto, 150 data, 800 gold |
| `fusion_assault` | Fusion Plant Assault | Fusion Age | 35 | — | 40t | 0.65 | 120 plasma, 500 electricity, 50 uranium |
| `orbital_strike` | Orbital Strike | Space Age | 40 | — | 50t | 0.70 | 100 titanium, 80 plasma, 300 knowledge |
| `warp_invasion` | Warp Invasion | Interstellar Age | 60 | — | 65t | 0.75 | 50 dark matter, 200 titanium, 2,000 gold |
| `galactic_conquest` | Galactic Conquest | Galactic Age | 80 | — | 80t | 0.80 | 30 antimatter, 100 dark matter, 5,000 gold |
| `quantum_incursion` | Quantum Incursion | Quantum Age | 100 | — | 90t | 0.85 | 20 quantum flux, 50 antimatter, 5,000 knowledge |

That's **16 missions** in total — 3 scouting expeditions and 13 military campaigns.

**Scouting — the soldier-free expeditions.** Before the Iron Age there is no military worker domain and no `soldiers` resource, so the military campaigns aren't available. The **scouting** expeditions fill that gap — `scout_party` (primitive → bronze) and then `scout_ruins` (bronze age) are available without soldiers, charging only resources. `scout_party`: *"A small band of foragers scouts nearby territory for resources."* It costs **30 food + 30 wood**, needs **0 soldiers**, runs ~20 ticks, and rewards roughly **60 food / 60 wood / 20 stone** — a net resource gain worth running on repeat through the Primitive, Stone, and Bronze ages. It has a `MaxAge` of Bronze Age and disappears once you reach the Iron Age. `scout_ruins` (Bronze Age, 0 soldiers, 40 food + 30 wood) carries scouting forward, and `naval_expedition` (Renaissance Age, 0 soldiers, 150 food + 100 wood) is the late scouting option.

**Tip:** `campaign trade_escort` (Iron Age, 3 soldiers, 12t, 0.30 difficulty) is the workhorse of early *soldier* play — cheap, short, and chainable for consistent gold. For soldier-free resource throughput, chain the scouting expeditions (`expedition scout_party`, then `expedition scout_ruins`) — low difficulty, short duration, and no soldiers required.

### Faction encounters

Every time an expedition resolves — win or lose — the game rolls a chance to **encounter a civilization**. An encounter either makes **first contact** with a new civ you're old enough to meet or **re-encounters** one you already know. This is the main way you discover the diplomacy roster: an age only makes a civ *eligible*; a resolved expedition is what actually turns someone up. (Never run expeditions and a civ is auto-discovered anyway about two ages past its minimum age — the fallback, not the fast path.)

**Scouting** expeditions turn up civilizations far more readily than **military** campaigns, and success beats failure — a failed run can still find someone, just rarely. Roughly:

| Resolution | Encounter chance |
|---|---|
| Scouting success | ~35% |
| Scouting failure | ~12% |
| Military success | ~15% |
| Military failure | ~5% |

An encounter can also grant a temporary **boon** — a boost to that civ's **specialty resource**, currently **+8–20% for 3,000–6,000 ticks**, announced with a flavour line (e.g. *"The Ironhold Clans share their forge-craft — +15% iron for 4,200 ticks."*). So expeditions stay worth running even after you've met everyone. See [Trade & Diplomacy](trade.md#diplomacy-civilization-encounters) for the civilization roster and diplomacy mechanics.

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
| **Permanent bonuses** (milestones) | Complete military milestones | +0.05 to +0.10 each (`military_power`) |
| **Prestige upgrade** | `prestige buy military_power` (5 tiers, 2/3/5/8/10 pts each) | +5% per tier |

There is no hard cap on `military_power`, but effective expedition difficulty is floored at **0.05** (5% chance of failure minimum), so stacking beyond ~2.5–3.0 bonus yields diminishing returns against failure rates.

The `expedition_reward` bonus (from research, prestige `expedition_loot`, and certain wonders) is separate and multiplies the loot amount on success: `rewards × (1 + expeditionBonus)`.

---

## 7. Military Milestones

The five military milestones form a chain. Completing the full chain grants a permanent title plus, cumulatively, **+0.25 military_power** and **+0.25 production_all** — the late tiers now broaden into all-production bonuses so the chain lifts your whole economy, not just combat.

| Key | Name | Requirement | Age Gate | Reward |
|-----|------|-------------|----------|--------|
| `first_soldiers` | First Soldiers | 5 soldiers | Iron Age | +0.05 military_power |
| `war_machine` | War Machine | 250 soldiers | Iron Age | +0.10 military_power |
| `iron_legion` | Iron Legion | 500 soldiers + 10 Barracks | Classical Age | +0.05 production_all |
| `fortress_state` | Fortress State | 20 Castle Keeps | Medieval Age | +0.10 military_power, +0.05 production_all |
| `military_superpower` | Military Superpower | 2,000 soldiers | Industrial Age | +0.15 production_all |

`iron_legion`, `fortress_state`, and `military_superpower` are **hidden** until their prerequisites are visible (progress > 50% or you're in the preceding age). Don't be surprised when they appear mid-game.

> **Note:** Standing Army (100 soldiers + 10 Barracks, Classical Age, +0.05 military_power) is a standalone military milestone — it is not part of the chain.

---

## 8. Strategy

### Early game (Iron Age to Classical Age)

- Before Iron Age, run `expedition scout_party` on repeat — 0 soldiers, just 30 food + 30 wood for a ~60 food / 60 wood / 20 stone payout. It's free resource throughput while you wait for the military domain.
- Build a **War Camp** in the Stone Age even though soldiers aren't possible yet — it prepares your soldier storage and starts producing the moment the domain unlocks.
- Your first soldiers become available in **Iron Age** via the Hunting Lodge. Stockpile 5 quickly to land `first_soldiers` for the free +0.05 bonus.
- `campaign trade_escort` (spends 3 soldiers, Iron Age) is your best early campaign — cheap, fast duration (12t), reasonable gold reward.
- Keep food workers prioritised. Military workers eat 2.0 food/tick each at this stage — a 10-soldier army demands 20 food/tick just to sustain itself.

### Mid game (Classical to Industrial Age)

- Push `iron_legion` (300 soldiers + 5 Barracks) — the +0.10 bonus noticeably improves expedition success on harder missions.
- `campaign conquer_territory` (10 soldiers, 25t, 0.60 difficulty) is the best soldier-spend bang-for-tick in this range. `expedition naval_expedition` is a **scouting** option — 0 soldiers, just 150 food + 100 wood (Renaissance Age, 35t, 0.50 difficulty) — so you can run it without dipping into your soldier stockpile.
- Build **Legion Forts** and **Military Academies** to raise your soldier ceiling. The capacity doubling per tier means each new building unlocks dramatically more troops.
- Research military-flavored techs as they appear — even +0.2 military_power makes a visible difference against 0.6-difficulty expeditions.

### Late game (Modern Age onward)

- `campaign world_domination` spends 50 soldiers but pays 1,000 gold — worth the queue for gold-hungry ages once your soldier production can refill the cost.
- `campaign cyber_raid` and `campaign neon_heist` are the best value in the digital/cyberpunk range. `neon_heist` (0.55 difficulty) is easier than `cyber_raid` (0.60) for comparable loot.
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

Maintaining a large standing army has a second cost beyond food: **morale**. If military workers exceed **30% of your total population**, morale drains every tick — and the **further over the threshold you are, the faster it drains**:

| Military ratio | Overage | Morale drain |
|---------------|---------|--------------|
| 30% (at threshold) | 0% | — (none) |
| 40% | +10% | mild |
| 50% | +20% | moderate |
| 60% | +30% | steep |

Morale is a civilization-wide percentage that multiplies **all** worker-driven output. It sits in three bands: in the **neutral band (25–75%)** it has no effect, but **below 25%** it penalises production — ramping down toward **×0.50 at the 10% floor**. So letting an oversized army drag morale into the low band suppresses every domain at once — food, knowledge, trade, everything — creating a feedback loop where the army's food cost gets harder to cover as your food workers produce less. **Above 75%**, morale instead *boosts* output up to **+20%** near the cap, so a lean military leaves headroom to push morale into the bonus band rather than spending it fighting an over-large army.

**Recommended target: keep military workers below 25–28% of total population.** This provides a comfortable buffer against the threshold even if population fluctuates from worker loss events. If you need a large soldier stockpile for an expensive expedition, staff your military buildings heavily to bank soldiers quickly, then unassign the excess military workers back to civilian buildings once you've launched — the stored soldiers remain, and the morale and food drain from the idle workforce goes away.

Recovery is automatic: morale **drifts back toward 50% neutral** each tick once you shed the excess military, so the low-band penalty self-heals as soon as the ratio is fixed — you don't have to do anything beyond getting back under the threshold. If you see morale trending downward and your military ratio is over 30%, unassign some soldiers or recruit more civilians to halt the drain.

See [Morale](morale.md) for the full banded system.

---

## 9. Tips & Common Mistakes

**Don't recruit past your food income.** A food deficit stalls all production (workers can't work when starving). Calculate the drain before `recruit max`.

**Don't wage high-difficulty campaigns without military power.** `campaign siege_castle` (0.70) and `campaign world_domination` (0.80) fail frequently with zero bonus. Research a few military techs first and watch the success probability change.

**Expedition failure still costs the launch.** The soldiers (and any resource cost) are spent the moment you launch — failure doesn't refund them, it only shrinks the reward. Don't launch a high-difficulty run unless you can afford to lose the soldier cost for a reduced payout.

**Workers must be assigned to built buildings.** The game blocks assignment if the building count is zero. Build the structure before trying to assign soldiers to it.

**Castle Keep timing matters for `fortress_state`.** You need 10 Castle Keeps (Medieval Age). That's a serious stone and iron investment — start queuing them as soon as you hit Medieval. The +0.10 bonus is worth the build cost several times over in improved expedition returns.

**Run expeditions continuously.** There's no cooldown between expeditions beyond the active duration. The moment one resolves, launch the next. Idle military is lost throughput.

**Prestige compounds military strength.** `military_power` prestige upgrade (5 tiers × 5% = +25%) and `expedition_loot` (5 tiers × 5% = +25%) both persist through resets. Prioritise these in the prestige shop on your second and third runs.
