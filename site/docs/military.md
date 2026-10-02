# Military & Expeditions

Soldiers pay for campaigns, and campaigns bring back gold, resources and knowledge you couldn't produce as fast on your own. Waging a campaign costs soldiers. Scouting expeditions cost none, only a little food and wood. The soldiers you keep in stock are your garrison: they blunt raids and soften catastrophes. Ignore the army and you still carry the small garrison the age gates build for you; invest in it and raids and catastrophes take noticeably less. Both kinds of mission are also how you meet other civilizations, and soldiers count toward several milestones and prestige upgrades.

---

## 1. Overview

The military system covers four things:

- **Campaigns.** You spend stockpiled soldiers on a timed military mission (`campaign <key>`) that pays resources, gold and knowledge. Scouting expeditions (`expedition <key>`) are the soldier-free kind of mission; see [§4](#4-expeditions).
- **Defense rating.** A figure on the Army panel worked out from your soldier count and military power. Measured against the raid threat of your current age, it blunts part of what raids, war raids and an Endure take from you (at most 45%); see [§7](#7-defense-what-your-army-blunts).
- **Milestones.** Five military milestones form a chain with a title at the end. The early ones grant permanent military power; the later ones grant all production.
- **Prestige.** The `military_power` and `expedition_loot` prestige upgrades carry over through resets.

**Soldiers are a resource**, not a count of workers. Your military buildings produce and store them: staff a War Camp or Barracks with workers and it adds to the `soldiers` resource every tick, the same way a Farm adds food. Soldiers unlock in the **Iron Age**. Before that, the scouting expeditions (see [§4](#4-expeditions)) let you explore for resources without any soldiers.

There are two kinds of mission:

| Kind | Command | Costs | Available |
|---|---|---|---|
| Scouting expedition | `expedition <key>` | resources only, 0 soldiers | from the start; there are three: `scout_party`, `scout_ruins`, `naval_expedition` |
| Military campaign | `campaign <key>` | soldiers, plus resources for some | the first opens in the Bronze Age, but you have no soldiers until the Iron Age |

---

## 2. Soldiers

### What soldiers are

Soldiers are a stockpiled resource (`soldiers`) that unlocks in the **Iron Age**. You bank them the way you bank food or wood, then spend them on campaigns.

Staffed military buildings produce soldiers every tick. A fully staffed building makes about its soldier cap ÷ 50 per tick (at least 0.1/tick), so more workers and more buildings mean faster soldiers. Your soldier storage is the sum of every military building's soldier cap, plus every storage effect that raises all resources (the general storage buildings and the "all storage" techs). That second part is most of it: by the Iron Age a typical civilization can hold a couple of hundred thousand soldiers, far more than the military buildings' own caps. Campaigns spend soldiers from the stockpile at launch (see [§4](#4-expeditions)).

### Producing soldiers

```
recruit [count|max]     # recruit workers into free housing
assign war_camp 5       # staff a military building; it starts producing soldiers
assign barracks all     # fill a building for its full soldier output
```

`recruit 5` recruits 5 workers into free housing, and `recruit max` fills all of it at once. Workers you leave idle go to work by your [worker shares](workers-and-domains.md#worker-shares) a minute later, and with auto-recruit on (the default) the game recruits into empty slots on its own, military buildings included. A worker assigned to a military building produces soldiers each tick. To keep workers out of military buildings, set `workers share military 0`.

### Food drain

Every worker eats the same amount of food, whatever building it works in: the current age's food cost per worker. In the Iron Age that is about 0.08 food/tick, and it rises by ×1.12 with each age after that. Military workers cost no more food than farmers do, and soldiers themselves eat nothing once they're trained. A big military workforce still costs food because it's a big workforce, so check your food rate before you staff up.

### Losing soldiers

Soldiers leave your stockpile when you launch a campaign. The cost is spent up front, whether the campaign then succeeds or fails. The three scouting expeditions cost 0 soldiers. No random event and no setback takes soldiers.

An Endure also costs soldiers: it cuts every stored resource to the share it keeps, soldiers included, so your garrison is weaker afterwards. Your soldiers are measured before the blow lands, so the garrison still counts for that Endure.

To refill the stockpile, keep workers assigned to your military buildings.

### Viewing your army

```
army
```

This opens the **Army** panel: soldier count, defense rating, the running campaign (if any) and how many campaigns you've completed. Under the defense rating it shows what the army does for you:

- **Threat:** the raid threat of your current age.
- "Your garrison would blunt about N% of a raid.": the share it takes off raids and war raids in your current age (an Endure measures it against the age the catastrophe strikes in). After it comes what twice the garrison would blunt, and a reminder that no army blunts more than 45%. With no soldiers it reads "You have no garrison: raids hit you with full force."
- **Saved this run:** once the garrison has saved something, the buildings, workers and resources (the four largest) it kept, and how many raids it blunted. Like other run stats, it resets with the run.

The running scouting expedition shows in the separate **Expeditions** panel (`expedition`). You can run one scouting expedition and one campaign at the same time. The Loot History (total loot collected) is on the Expeditions panel.

---

## 3. Military Buildings

Military buildings give workers somewhere to work, produce soldiers every tick while staffed, and each one adds its soldier cap to your soldier storage. Your soldier storage is their combined cap plus every storage effect that raises all resources (see Soldiers above).

All military buildings belong to the **military lineage** (22 tiers, Stone Age through Transcendent Age). Each extra copy of a building costs 35% more than the one before.

A fully staffed military building produces about **its soldier cap ÷ 50 per tick** (at least 0.1/tick). A staffed Barracks (cap 20) makes about 0.4 soldiers/tick; a Castle Keep (cap 320) makes about 6.4/tick. Output scales with staffing the same way every other building's does; see [§5](#5-military-domain-workers).

| Tier | Key | Name | Age Unlocked | Soldier Cap | Worker Slots |
|------|-----|------|-------------|-------------|--------------|
| 0 | `war_camp` | War Camp | Stone Age | +10 | 3 |
| 1 | `barracks` | Barracks | Bronze Age | +20 | 4 |
| 2 | `hunting_lodge` | Hunting Lodge | Iron Age | +40 | 5 |
| 3 | `legion_fort` | Legion Fort | Iron Age | +80 | 6 |
| 4 | `military_academy` | Military Academy | Classical Age | +160 | 6 |
| 5 | `castle_keep` | Castle Keep | Medieval Age | +320 | 7 |
| 6 | `fortress` | Fortress | Renaissance Age | +640 | 7 |
| 7 | `fort` | Fort | Colonial Age | +1.28K | 8 |
| 8 | `military_base` | Military Base | Industrial Age | +2.56K | 10 |
| 9 | `garrison` | Garrison | Victorian Age | +5.12K | 10 |
| 10 | `command_post` | Command Post | Electric Age | +10.2K | 12 |
| 11 | `bunker_complex` | Bunker Complex | Atomic Age | +20.5K | 12 |
| 12 | `special_ops_hq` | Special Ops HQ | Modern Age | +41.0K | 14 |
| 13 | `cyber_command` | Cyber Command | Information Age | +81.9K | 15 |
| 14 | `drone_warfare_center` | Drone Warfare Center | Digital Age | +164K | 16 |
| 15 | `combat_aug_center` | Combat Aug Center | Cyberpunk Age | +328K | 18 |
| 16 | `plasma_command` | Plasma Command | Fusion Age | +655K | 20 |
| 17 | `space_force_base` | Space Force Base | Space Age | +1.31M | 20 |
| 18 | `fleet_command` | Fleet Command | Interstellar Age | +2.62M | 25 |
| 19 | `stellar_armada_hq` | Stellar Armada HQ | Galactic Age | +5.24M | 25 |
| 20 | `probability_war_room` | Probability War Room | Quantum Age | +10.5M | 30 |
| 21 | `omniversal_war_council` | Omniversal War Council | Transcendent Age | +21.0M | 35 |

The soldier cap counts per building: two Barracks give 40 soldier storage and, when both are staffed, twice the soldier output. Build up both before you go after the campaigns that cost many soldiers.

---

## 4. Expeditions

### What expeditions are

Expeditions are timed missions paid from your stockpiles. There are two kinds:

- **Scouting** (`scout_party`, `scout_ruins`, `naval_expedition`) costs resources and 0 soldiers. These are open before soldiers exist.
- **Military campaigns** (the other 13) cost soldiers, plus resources for some.

The cost comes out of your stockpiles the moment you launch, and there's no refund. The mission then runs for a number of ticks rolled at launch within its own range, and resolves. Success and failure differ only in the reward: a success pays full loot, a failure pays 30% of it, and the cost is gone either way. You can run one scouting expedition and one campaign at the same time, but not two of the same kind.

### Cost, age range and rewards

Campaigns spend soldiers at launch, and you can't launch one you can't afford. Some missions also charge resources up front: `scout_party` costs 30 food and 30 wood, and `naval_expedition` costs 150 food and 100 wood.

Every mission has a first age, and some have a last age after which they disappear. `scout_party` runs from the Primitive Age through the Bronze Age and is gone once you reach the Iron Age, when soldiers arrive. `scout_ruins` opens in the Bronze Age and `naval_expedition` in the Renaissance Age.

The reward is paid when the mission resolves. The soldiers and resources you paid don't come into it; they were spent at launch.

### Commands

Scouting expeditions and campaigns have separate commands:

```
expedition list         # scouting expeditions open in your current age
expedition <key>        # send a scouting expedition (e.g. expedition scout_ruins); costs resources, never soldiers
exp list                # shorthand
exp <key>               # shorthand

campaign list           # military campaigns open in your current age
campaign <key>          # wage a campaign (e.g. campaign raid_bandits); spends soldiers
```

`expedition` with no arguments opens the **Expeditions** panel, `campaign` (or `campaign list`) lists the campaigns you can wage, and `army` opens the **Army** panel. You can type spaces for the underscores in a key: `expedition scout ruins` is the same as `expedition scout_ruins`.

If you mix the two up, the game points you to the right one. `expedition raid_bandits` is refused with a note to use `campaign <key>`, and a scouting key given to `campaign` sends you to `expedition <key>`.

### Automatic dispatch: the Geographic Society

From the **Industrial Age** you can stop sending every party by hand. The **Geographic Society** is a building that sends scouting parties out on its own for the rest of the run.

It only ever scouts. A Society never wages a campaign; those stay yours to order.

It uses your one scouting slot, and only when the slot is free. Nothing is queued. If a scouting party is out, whether you sent it or the Society did, the Society waits, so a party you launch by hand always goes first.

It pays full price. Each automatic party costs the same resources you would pay. If your stores can't outfit one, the Society holds it back, tells you once, and sends it as soon as the supplies are there. It never runs up a debt.

It sends the cheapest scouting mission your age offers that you can currently afford, so it won't spend your treasury on a voyage you didn't ask for.

Its parties are ordinary parties. They succeed and fail at the same odds, pay the same loot, and roll civilization encounters, boons and setbacks exactly like the ones you send.

The more you invest, the faster it goes. A single unstaffed Society sends about one party every 900 ticks (about 30 minutes at 1x), roughly what you'd manage by remembering to send one now and then. Staffing it fully cuts that by about a third, and each extra Society shortens the wait again, down to a floor of about **one party every 100 ticks** (about 3m 20s) with six fully staffed Societies. More than that only helps when a party comes home late.

The pace is set in ticks, so the real-time figures above are approximate and shrink as your tick speed rises. The **Expeditions** panel (`expedition`) shows one line for it: the time until the next party leaves, or a warning that a party is due and your stores are too thin to outfit it. The full status (Societies built, staffing, the interval, and a countdown with a progress bar) is on the **Factions** panel (`factions`); see [The Factions panel](trade.md#the-factions-panel). Both show wall-clock time rather than ticks; see [Timers and durations](commands.md#timers-and-durations).

A Society is always slower than doing it yourself, by design. A fully invested Society runs at about **60% of the pace of a player who chains expeditions by hand and runs campaigns alongside them**. That keeps an idle empire meeting other civilizations and keeps its boons topped up, while playing by hand stays worth it.

The **Expeditions** panel (`expedition`) lists only the scouting expeditions open in your current age. The **Army** panel (`army`) lists only the campaigns, next to your soldier count and defense rating. The Expeditions panel also carries the **Loot History**, the running total of loot from expeditions and campaigns alike.

### Success chance

```
difficulty = base difficulty - (military power × 0.3)
difficulty = max(difficulty, 0.05)
success if a random roll (0 to 1) beats difficulty
```

Military power lowers the effective difficulty. With no military power, a 0.8-difficulty mission succeeds about 20% of the time. With +2.0 military power its difficulty drops to 0.2 (about 80% success), and at +2.5 it reaches the 0.05 floor, about 95%.

The soldier and resource cost is already spent at launch, so the outcome only changes the reward:

| Outcome | Reward |
|---|---|
| Success | full rewards × (1 + expedition reward bonus) |
| Failure | 30% of the base rewards |

A failure costs no extra soldiers. The launch cost is the whole cost, win or lose.

Every resolution logs a plain line: which mission, whether it succeeded, and that loot came in. About one resolution in three also gets a short gray account of how it went underneath. It's kept rare on purpose so those lines stay worth reading. The account depends on the kind of mission, whether it succeeded and your age (a Bronze Age party and a Quantum Age crew come home to different places). It won't repeat a sentence you saw in roughly the last screenful of log, and it's cosmetic only: it never contains a number that isn't on the line above it.

### Full expedition table

This table covers all 16 missions. The three scouting expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) are sent with `expedition <key>`; every other row is a campaign, waged with `campaign <key>`. The Soldier Cost column is spent from your soldiers at launch.

| Key | Name | Min Age | Soldier Cost | Resource Cost | Duration | Difficulty | Rewards on Success |
|-----|------|---------|--------------|---------------|----------|------------|-------------------|
| `scout_party` | Scout Party | Primitive Age (last: Bronze Age) | 0 | 30 food, 30 wood | 100-160t | 0.20 | 60 food, 60 wood, 20 stone |
| `scout_ruins` | Scout Nearby Ruins | Bronze Age | 0 | 40 food, 30 wood | 60-100t | 0.20 | 30 food, 20 wood, 15 stone |
| `raid_bandits` | Raid Bandit Camp | Bronze Age | 5 | none | 60-100t | 0.40 | 30 gold, 15 iron, 20 food |
| `trade_escort` | Trade Escort | Iron Age | 3 | none | 60-100t | 0.30 | 50 gold, 10 knowledge |
| `conquer_territory` | Conquer Territory | Iron Age | 10 | none | 60-100t | 0.60 | 80 gold, 40 iron, 50 food |
| `siege_castle` | Siege Enemy Castle | Medieval Age | 15 | none | 60-100t | 0.70 | 150 gold, 30 steel, 20 faith |
| `naval_expedition` | Naval Expedition | Renaissance Age | 0 | 150 food, 100 wood | 60-100t | 0.50 | 200 gold, 30 culture, 40 knowledge |
| `colonial_campaign` | Colonial Campaign | Industrial Age | 20 | none | 60-100t | 0.60 | 300 gold, 50 oil, 40 steel |
| `world_domination` | World Domination | Modern Age | 50 | none | 60-100t | 0.80 | 1K gold, 200 electricity, 500 knowledge |
| `cyber_raid` | Cyber Raid | Information Age | 30 | none | 60-100t | 0.60 | 200 data, 50 crypto, 500 gold |
| `neon_heist` | Neon Heist | Cyberpunk Age | 25 | none | 60-100t | 0.55 | 100 crypto, 150 data, 800 gold |
| `fusion_assault` | Fusion Plant Assault | Fusion Age | 35 | none | 60-100t | 0.65 | 120 plasma, 500 electricity, 50 uranium |
| `orbital_strike` | Orbital Strike | Space Age | 40 | none | 60-100t | 0.70 | 100 titanium, 80 plasma, 300 knowledge |
| `warp_invasion` | Warp Invasion | Interstellar Age | 60 | none | 65-105t | 0.75 | 50 dark matter, 200 titanium, 2K gold |
| `galactic_conquest` | Galactic Conquest | Galactic Age | 80 | none | 80-130t | 0.80 | 30 antimatter, 100 dark matter, 5K gold |
| `quantum_incursion` | Quantum Incursion | Quantum Age | 100 | none | 90-145t | 0.85 | 20 quantum flux, 50 antimatter, 5K knowledge |

That is 3 scouting expeditions and 13 campaigns.

The Duration column is each mission's range in ticks. The actual time is rolled evenly within that range at launch, so nothing resolves in under 60 ticks, and `scout_party` runs 100-160 ticks (about 130 on average). The time-left readout counts down the rolled value.

There are no soldiers before the Iron Age, so scouting fills the gap. `scout_party` costs 30 food and 30 wood, runs 100-160 ticks and pays about 60 food, 60 wood and 20 stone, a net gain worth repeating through the Primitive, Stone and Bronze ages. It disappears once you reach the Iron Age. `scout_ruins` (Bronze Age, 40 food and 30 wood) carries scouting on from there, and `naval_expedition` (Renaissance Age, 150 food and 100 wood) is the late scouting option.

`campaign trade_escort` (Iron Age, 3 soldiers, 60-100 ticks, 0.30 difficulty) is the cheap, repeatable early campaign for steady gold. For resources without soldiers, chain the scouting expeditions (`expedition scout_party`, then `expedition scout_ruins`). Durations are rolled, so keep one of each kind running rather than counting on a fixed timer.

### Civilization encounters

Every time a mission resolves, win or lose, the game rolls a chance to **encounter a civilization**. An encounter either makes **first contact** with a civilization your age makes eligible, or **meets again** one you already know. This is the main way you discover civilizations: an age only makes a civilization *eligible*, and a resolved mission is what turns it up. If you never send any, each civilization is discovered anyway about two ages after its first age, much later than an explorer would meet it.

Scouting expeditions find civilizations far more often than campaigns, and a success finds them more often than a failure. A failed run can still find someone, rarely, and what it brings home is trouble rather than a gift (see *Setbacks* below).

| Resolution | Encounter chance |
|---|---|
| Scouting success | ~18% |
| Scouting failure | ~6% |
| Campaign success | ~8% |
| Campaign failure | ~2% |

The chance is rolled once per resolved mission, and missions aren't quick: the shortest scouting run takes 60-100 ticks (about 2m to 3m 20s), rolled fresh each launch. A player who keeps a scouting party and a campaign in the field all the time meets someone about every **375 ticks** (about 12 minutes at 1x). Encounters are occasional, so the point of chaining missions is that the rewards you do get overlap.

From the Industrial Age a fully invested **Geographic Society** (see [Automatic dispatch](#automatic-dispatch-the-geographic-society)) keeps encounters coming while you're away, at about **60%** of the hands-on rate: about one meeting every 640 ticks (about 21 minutes at 1x). A single Society is slower, closer to one every 2,500 ticks (about 1h 23m). That gives three real paces: chaining missions yourself, letting the Societies work, and doing neither.

An encounter from a **successful** mission can also grant a **boon**, a reward rolled from a shared catalog: a timed boost to one resource, to all production or to knowledge; a tick-speed surge; an instant lump of resources; or temporary workers. The roll is weighted by the civilization's character (personality, specialty and strength) and by its **opinion** of you. Allies give bigger boons, and rare ones. Instant gifts scale with your age, so a caravan of supplies is still worth having in the Quantum Age. The log names the civilization and the reward.

**You can hold five boons at a time.** Only *timed* rewards take one of the five slots; an instant gift is used on arrival and holds nothing. While all five slots are full, a timed reward is turned away (the envoys are thanked, fed and sent home with their crates unopened), but an instant lump of resources or a gang of temporary workers still arrives, because neither needs a slot. A full set of boons costs you the buffs, never the goods.

Timed boons last **750-3000 ticks** each, so slots free up on their own while you explore. A player who explores without a break gets a reward from about four encounters in five, has at least one boon running over 95% of the time, and has all five slots full only **12-18%** of the time. Only the sixth timed boon at once goes to waste.

**Setbacks.** Some encounters go badly and bring a **setback** instead of a boon:

- every encounter from a **failed** mission;
- about **one in three** encounters with a civilization you are **at war** with (the rest are standoffs: contact is made and nothing is gained or lost, and an enemy never gives a gift);
- about **one in four** of the timed rewards turned away because all five boon slots are full.

A setback is one of: a few workers lost on the road home, part of one resource stockpile spoiled or carried off, a temporary drop in one resource's output, or a production dip across your whole civilization for a while. It's worse when the civilization is **strong** and milder when its **opinion** of you is high, so a friend's bad news is gentler than an enemy's, and a strong civilization you're at war with is the worst case. Setbacks are capped more tightly than boons: at most **three** timed setbacks run at once (against five boon slots), they expire sooner than a boon of the same size, and a spoilage never empties a store.

Missions stay worth running after you've met everyone, but a failed run costs you something, and a run that resolves while all five boon slots are full pays in goods rather than buffs. See [Trade & Diplomacy](trade.md#diplomacy-civilization-encounters) for the civilization roster and diplomacy.

---

## 5. Military Domain Workers

Workers in military buildings take a class name that changes with the age:

| Age | Class Name |
|-----|-----------|
| Iron Age | Soldier |
| Classical Age | Legionary |
| Medieval Age | Knight |
| Renaissance Age | Musketeer |
| Colonial Age | Colonial Marine |
| Industrial Age | Industrial Rifleman |
| Victorian Age | Victorian Guard |
| Electric Age | Electric Trooper |
| Atomic Age | Atomic Soldier |
| Modern Age | Modern Soldier |
| Information Age | Information Warrior |
| Digital Age | Digital Soldier |
| Cyberpunk Age | Cyber Warrior |
| Fusion Age | Plasma Trooper |
| Space Age | Space Marine |
| Interstellar Age | Interstellar Commando |
| Galactic Age | Galactic Guardian |
| Quantum Age | Quantum Soldier |

The class name doesn't change what they eat. Every worker eats the age's food cost per worker (see [Food drain](#food-drain)).

### Assignment

```
assign <building> [count|all]

# Examples:
assign war_camp 3
assign barracks all
assign castle_keep 7
```

Workers in military buildings produce soldiers; they aren't soldiers themselves, and the soldiers they make are a separate resource. You can only assign workers to a building you have built at least one of.

Staffing sets output: a building runs at `20% + 80% × (assigned ÷ worker slots)` of its full effect. Fill your military buildings to get their full soldier output.

---

## 6. Military Power Bonus

Military power is a running total that lowers mission difficulty (see [Success chance](#success-chance)) and feeds the defense rating:

```
defense = soldiers × 2.0 × (1 + military power)
```

The defense rating is what your garrison is measured by against raids and catastrophes (see [§7](#7-defense-what-your-army-blunts)), so military power pays twice: it raises mission success chances, and it makes the same soldiers blunt more. A +1.0 bonus doubles the defense rating of the same army.

Sources add together:

| Source | How to get it | Bonus |
|--------|-------------|---------------|
| Techs | Several military techs grant military power | +0.2 to +1.5 per tech |
| Milestones | Complete military milestones | +0.05 to +0.10 each |
| Prestige upgrade | `prestige buy military_power` (5 tiers, 2/3/5/8/10 points) | +0.05 per tier |

There's no cap on military power, but difficulty never drops below **0.05** (a 5% minimum failure chance). Past about +2.5 to +3.0, more military power stops helping on most missions.

The expedition reward bonus (from techs, milestones, the `expedition_loot` prestige upgrade and some wonders) is separate. It multiplies the loot on a success: `rewards × (1 + expedition reward bonus)`.

---

## 7. Defense: what your army blunts

Soldiers you keep in stock are your **garrison**. They are never spent by defending: the garrison takes a share off what raids and catastrophes cost you, and the soldiers stay where they are.

### The formula

Your defense rating is measured against the **raid threat** of the age you are in:

```
defense    = soldierCount × 2.0 × (1 + militaryBonus)
threat     = 160,000 × 2^(age order)      # Primitive Age = order 0
mitigation = 45% × defense / (defense + threat)
```

`mitigation` is the share of a raid's losses the garrison blunts. It is 0 with no soldiers, half the ceiling (22.5%) when your defense equals the threat, and it approaches 45% without ever reaching it. More soldiers always help, but each doubling helps less than the last:

| Defense vs threat | Share blunted |
|-------------------|---------------|
| none | 0% |
| one third of the threat | about 11% |
| half the threat | 15% |
| equal to the threat | 22.5% |
| twice the threat | 30% |
| four times the threat | 36% |
| nine times the threat | about 40% |

The threat **doubles every age**, the same rate at which each new military building's soldier cap doubles. An army that is strong for one age is ordinary for the next and a rounding error a few ages later, so a garrison has to grow with you.

| Age | Threat | Soldiers for 22.5% (no military bonus) |
|-----|--------|----------------------------------------|
| Iron Age | 1.28M | 640K |
| Classical Age | 2.56M | 1.28M |
| Medieval Age | 5.12M | 2.56M |
| Renaissance Age | 10.24M | 5.12M |
| Colonial Age | 20.48M | 10.24M |
| Industrial Age | 40.96M | 20.48M |
| Victorian Age | 81.92M | 40.96M |
| Electric Age | 163.84M | 81.92M |
| Atomic Age | 327.68M | 163.84M |
| Modern Age | 655.36M | 327.68M |
| Information Age | 1.31B | 655.36M |
| Digital Age | 2.62B | 1.31B |
| Cyberpunk Age | 5.24B | 2.62B |
| Fusion Age | 10.49B | 5.24B |
| Space Age | 20.97B | 10.49B |
| Interstellar Age | 41.94B | 20.97B |
| Galactic Age | 83.89B | 41.94B |
| Quantum Age | 167.77B | 83.89B |
| Transcendent Age | 335.54B | 167.77B |

The Primitive, Stone and Bronze Ages have threats of 160K, 320K and 640K, but soldiers don't exist until the Iron Age. A `military_power` bonus cuts the soldiers needed: at +1.0, half as many.

### The garrison you already have

Every player carries a garrison without trying. The age gates ask for military buildings (15 Hunting Lodges for the Classical Age, 15 Military Academies for the Medieval, 3 Castle Keeps for the Renaissance, 15 Bunker Complexes for the Modern, 10 Plasma Commands for the Space Age, 15 Probability War Rooms for the Transcendent). You keep those buildings into every later age, and they train soldiers even with no workers assigned, at a fifth of the staffed rate. A player who builds only what the gates ask for blunts roughly **8-19%** of a raid from the Iron Age on (measured on the smoke-test bot, which does nothing else for its army).

A deliberate army is what pushes toward the 45% ceiling: more military buildings of the current age, workers staffed into them, and `military_power` research.

### What it blunts

1. **Raid events.** Random events that are attacks by outsiders: Bandit Raid, Pirate Attack, Data Breach, The Great Breach, Corporate Espionage, and the Stone Era's Tribal Raid and Beast Stampede. The garrison cuts the resources they steal and the workers they drive off by its share (at least one worker still flees if the event takes workers). A raid's production penalty is not blunted. See [Raids and your garrison](events.md#raids-and-your-garrison). The log adds a line under the event: "Your garrison blunted about N% of the raid: you kept ...".
2. **War raids.** While a civilization is at war with you, each raid it makes takes a resource. The garrison keeps its share of that resource. A raid bigger than your stock still takes nothing, as before; the army never makes a raid that missed land. The raid's log line adds "Your garrison kept X gold from them (about N% of the raid)." (with the raid's resource in place of gold).
3. **Endure.** When you Endure a catastrophe, the garrison blunts its share of the buildings destroyed and the stock lost, after the Harbinger's Brace, measured against the threat of the age the catastrophe strikes in (see below).

Nothing else. Disasters and unrest (earthquakes, plague, mine collapses, industrial accidents, uprisings) are not raids, and soldiers do nothing against them. Endure's 25% worker loss, its production debuff and its morale hit are unchanged.

A player with **no soldiers** takes exactly the losses they always did. From the Iron Age on that is rare: the military buildings the age gates require give nearly everyone some garrison (see [The garrison you already have](#the-garrison-you-already-have)).

### Endure: Brace first, then the garrison

A catastrophe strikes at a fated moment anywhere in its era (Iron to Cosmic), so the garrison meets the threat of whatever age you are in when it strikes. Soldiers first unlock in the Iron Age, so an Iron Era doom that strikes early in the Iron Age finds little garrison to count.

1. **Brace** applies first: 20% / 15% / 10% of buildings fall and 15% / 30% / 45% of stock is kept at Brace 0 / 1 / 2.
2. **The garrison** then blunts its share of what is left: the braced share of buildings destroyed shrinks by that share, and it keeps that share of the stock Brace would have let go. Buildings saved round down, so the garrison never saves more than its share.
3. **The cap:** Brace and garrison together can cut the unbraced loss by at most **60%**. At least 8% of buildings fall and at most 66% of stock is kept. The cap only bites at Brace level 2.

| Brace | Garrison share | Buildings destroyed | Stock kept |
|-------|----------------|---------------------|------------|
| none | 20% | 16% | 32% |
| none | near the 45% ceiling | about 11% | about 53% |
| 1 | 20% | 12% | 44% |
| 2 | 20% | 8% | 56% |
| 2 | 40% or more | 8% (cap) | 66% (cap) |

The catastrophe window shows the real numbers after Brace and garrison, with a line for each; the Harbinger panel's Brace preview counts the garrison too. After an Endure the log says "Your garrison held the line: ...". See [Your garrison](catastrophe.md#your-garrison) and [Brace](harbinger.md#brace-soften-an-endure).

The Last Passage in the Cosmic Era costs prestige points, not buildings or stock, and the garrison doesn't change it.

### Worked example

You hold 2,108,000 soldiers with no `military_power` bonus: defense 2,108,000 × 2 = 4,216,000.

- **In the Classical Age** (threat 2.56M): 45% × 4.216M / (4.216M + 2.56M) = about **28%**. A war raid that would take 1,000 gold takes about 720; the Army panel says twice the garrison would blunt about 35%.
- **In the Industrial Age** (threat 40.96M): 45% × 4.216M / (4.216M + 40.96M) = about **4%**. The same army barely matters four ages later.
- **Endure at the Renaissance** (threat 10.24M) with Brace 1, 150 buildings that can fall (everything but wonders and storage) and 2,560,000 soldiers (defense 5.12M, half the threat): the garrison share is 15%. Brace 1 alone would destroy 22 buildings and keep 30% of stock; with the garrison, 19 fall (22 × 15% = 3.3, rounded down to 3 saved) and about 40% of stock is kept.

---

## 8. Military Milestones

The five military milestones form a chain. Completing all five grants a title, and between them they give **+0.25 military power** and **+25% all production**. Soldier counts in these milestones are soldiers trained over the run, not the number in stock.

| Key | Name | Requirement | Min Age | Reward |
|-----|------|-------------|----------|--------|
| `first_soldiers` | First Soldiers | train 5 soldiers | Iron Age | +0.05 military power |
| `war_machine` | War Machine | train 250 soldiers | Iron Age | +0.10 military power |
| `iron_legion` | Iron Legion | train 500 soldiers, build 10 Barracks | Classical Age | +5% all production |
| `fortress_state` | Fortress State | build 20 Castle Keeps | Medieval Age | +0.10 military power, +5% all production |
| `military_superpower` | Military Superpower | train 2,000 soldiers | Industrial Age | +15% all production |

`iron_legion`, `fortress_state` and `military_superpower` are **hidden** until you are more than halfway to them or reach the age before their minimum age, so expect them to appear mid-game.

> **Note:** `standing_army` (train 100 soldiers and build 10 Barracks, Classical Age, +0.05 military power) is a separate military milestone, not part of the chain.

---

## 9. Strategy

### Early game (Iron Age to Classical Age)

- Before the Iron Age, run `expedition scout_party` on repeat. It costs 30 food and 30 wood and pays about 60 food, 60 wood and 20 stone, with no soldiers needed.
- Build a **War Camp** in the Stone Age even though soldiers don't exist yet. It gives you soldier storage and starts producing as soon as soldiers unlock.
- Soldiers unlock in the **Iron Age**. Staff your military buildings (the Hunting Lodge costs only 25 wood) and train 5 soldiers for `first_soldiers` and its +0.05 military power.
- `campaign trade_escort` (3 soldiers, Iron Age, 60-100 ticks) is the best early campaign: cheap, with a fair gold reward.
- Workers in military buildings eat the same food as everyone else, but every worker you move there is one fewer farmer. Keep your food rate positive.

### Mid game (Classical to Industrial Age)

- Go for `iron_legion` (train 500 soldiers, build 10 Barracks) for its +5% all production.
- `campaign conquer_territory` (10 soldiers, 60-100 ticks, 0.60 difficulty) gives the most per soldier in this range. `expedition naval_expedition` is a scouting option (0 soldiers, 150 food and 100 wood, Renaissance Age, 60-100 ticks, 0.50 difficulty), so it doesn't touch your soldier stockpile.
- Build **Legion Forts** and **Military Academies** to raise your soldier storage. The cap doubles with each tier, so each new tier holds far more soldiers.
- Research military techs as they appear. Even +0.2 military power makes a visible difference against 0.6-difficulty missions.

### Late game (Modern Age onward)

- `campaign world_domination` costs 50 soldiers but pays 1K gold, worth it once your soldier production can refill the cost.
- `campaign cyber_raid` and `campaign neon_heist` are the best value in the Information to Cyberpunk range. `neon_heist` (0.55 difficulty) is easier than `cyber_raid` (0.60) for comparable loot.
- Buy `expedition_loot` prestige tiers across resets. At tier 5 (+25% rewards), stacked with research bonuses, mission rewards grow a lot.
- Late-game military techs push most missions down to the 0.05 difficulty floor. The `military_superpower` milestone adds +15% all production on top.

### Defense rating

The Army panel shows your **defense rating** (`soldiers × 2.0 × (1 + military power)`). It is measured against the raid threat of your age, and the resulting share (at most 45%) comes off raid events, war raids from civilizations at war with you, and the buildings and stock an Endure takes. It never stops a raid or a catastrophe from happening, and it does nothing against disasters such as plague or earthquakes. A civilization at war with you raids every 40 ticks and takes 50 × its strength (1-5) of its specialty resource; your garrison keeps its share of that, but only staying out of wars, or ending them with tribute, stops the raids: see [War & Peace](trade.md#war-amp-peace). See [§7](#7-defense-what-your-army-blunts) for the numbers.

- **Keep up with the age.** The threat doubles every age, so a garrison you stop growing fades fast. Adding the current age's military buildings keeps pace, since each tier's soldier cap doubles too.
- **Mind the next age.** An Endure is measured against the age the doom strikes in. The Harbinger panel's Brace preview uses the age you are in; if you advance before the doom strikes, the threat doubles and your garrison blunts less than the preview showed.
- **Brace and garrison stack, up to a point.** Together they cut an Endure's losses by at most 60%. With Brace 2, a garrison that blunts 20% already reaches the building cap (8% fall), so extra soldiers mostly buy stock kept, up to 66%.
- **Soldiers cost nothing to hold.** The stock has no upkeep; only the military workers producing it eat food and count toward the morale ratio below. You can staff up to bank a garrison, then move the workers back.
- **Research military power.** It multiplies the defense rating of the soldiers you already have.

### Army size and food

Soldiers eat nothing, and a worker in a military building eats no more than any other worker. The food cost of an army is simply that its workers are workers. Rough rules:

- Before you recruit, check `rates` and make sure your food rate stays positive with the new workers added.
- Use `recruit max` only when food is overflowing, so a big recruit doesn't tip you into a deficit.
- After an age advance, check food again: the cost per worker goes up by ×1.12 each age.

### Morale and military ratio

A large army has a second cost: **morale**. If more than **30% of your population** works in military buildings, morale drops every tick, and the further over 30% you are, the faster it drops:

| Military share | Over 30% by | Morale lost per tick |
|---------------|---------|--------------|
| 30% | 0 | none |
| 40% | 10 points | 0.3 points |
| 50% | 20 points | 0.6 points |
| 60% | 30 points | 0.9 points |

Morale multiplies the output of every worker in your civilization, on a continuous curve centered on 50%. Below 50% it cuts production, down to ×0.50 at the 10% floor. Above 50% it raises production, up to +20% at the top. An oversized army that drags morale down therefore slows every domain at once (food, knowledge, trade and the rest), and the army's food gets harder to cover as your farmers produce less. A lean army leaves room to push morale above 50% instead.

**Keep military workers at 25-28% of your population or less.** That leaves a buffer below 30% if you lose workers to an event. If you need a large soldier stockpile for an expensive campaign, staff your military buildings heavily to bank soldiers, then move the extra workers back to civilian buildings once you've launched. The stored soldiers stay, and the morale drain stops.

Morale drifts back toward 50% by itself, a little each tick, once you're back under 30%. If morale is falling and more than 30% of your workers are in military buildings, unassign some of them or recruit more workers for other buildings.

See [Morale](morale.md) for the full system.

---

## 10. Tips & Common Mistakes

**Don't recruit past your food income.** A food deficit stalls all production, because starving workers can't work. Work out the extra food before `recruit max`.

**Don't wage hard campaigns without military power.** `campaign siege_castle` (0.70) and `campaign world_domination` (0.80) fail often with no bonus. Research a few military techs first.

**A failed mission still costs its launch.** The soldiers and any resources are spent when you launch. Failure doesn't refund them; it only cuts the reward to 30%. Don't launch a hard campaign unless you can afford to lose the soldiers for a small payout.

**Workers need a built building.** You can't assign workers to a building you haven't built yet.

**Plan Castle Keeps early for `fortress_state`.** It needs 20 Castle Keeps (Medieval Age), a large stone and iron investment, so start building them as soon as you reach the Medieval Age. The +0.10 military power and +5% all production are worth the cost.

**Keep missions running.** There's no cooldown beyond the mission's own duration (60 to 160 ticks, rolled at launch). When one resolves, send the next.

**Prestige keeps military strength.** The `military_power` upgrade (5 tiers × 0.05 = +0.25 military power) and `expedition_loot` (5 tiers × 5% = +25% rewards) both carry over through resets. Buy them early in your second and third runs.
