# Army & Missions

Soldiers are a resource your military buildings train. You spend them on **campaigns**; **scouting expeditions** cost a little food and wood instead. Mission loot is a fixed bundle that never grows with the age, so it matters early and hardly at all later. What missions are really for is meeting the other civilizations: every resolved mission can bring an encounter, which is how you make first contact and collect boons (or, when a run fails, setbacks). The soldiers you keep in stock are your garrison: they blunt raids and soften catastrophes. Ignore the army and you still carry the small garrison the age gates build for you; invest in it and raids and catastrophes take noticeably less. The soldiers you train also count toward the military milestones.

---

## Overview

The military system covers four things:

- **Missions.** Scouting expeditions (`expedition <key>`) cost resources; military campaigns (`campaign <key>`) cost soldiers. Each one resolves after a rolled time, pays a fixed bundle of loot and may bring an encounter with a civilization; see [Missions](#missions).
- **Defense rating.** A figure on the Army panel worked out from your soldier count and military power. Measured against the raid threat of your current age, it blunts part of what raids, war raids and an Endure take from you (at most 45%); see [Defense](#defense-what-your-army-blunts).
- **Milestones.** Five military milestones form a chain with a title at the end. The early ones grant permanent military power; the later ones grant all production.
- **Prestige.** The army resets with the run, and nothing you buy with prestige points raises military power or mission rewards. What carries over is who your missions found: with the legacy kit's [Old Friends](prestige.md#old-friends), every civilization you have met is met again on later runs without a mission.

**Soldiers are a resource**, not a count of workers. Your military buildings produce and store them: staff a War Camp or Barracks with workers and it adds to the `soldiers` resource every tick, the same way a Farm adds food. Soldiers unlock in the **Iron Age**. Before that, the scouting expeditions (see [Missions](#missions)) let you explore without any soldiers.

There are two kinds of mission:

| Kind | Command | Costs | Available |
|---|---|---|---|
| Scouting expedition | `expedition <key>` | resources only, 0 soldiers | from the start; there are three: `scout_party`, `scout_ruins`, `naval_expedition`. The last two wait for the Exploration tech, and the Naval Expedition for Navigation as well |
| Military campaign | `campaign <key>` | soldiers | the first opens in the Bronze Age, once Military Tactics is researched, but you have no soldiers until the Iron Age |

---

## Soldiers

### What soldiers are

Soldiers are a stockpiled resource (`soldiers`) that unlocks in the **Iron Age**. You bank them the way you bank food or wood, then spend them on campaigns.

Staffed military buildings produce soldiers every tick. A fully staffed building makes about its soldier cap ÷ 50 per tick (at least 0.1/tick), so more workers and more buildings mean faster soldiers. Your soldier storage is the sum of every military building's soldier cap, plus every storage effect that raises all resources (the general storage buildings and the "all storage" techs). That second part is most of it: by the Iron Age a typical civilization can hold a couple of hundred thousand soldiers, far more than the military buildings' own caps. The **Military-Industrial Complex**, the Atomic Age's Military capstone, adds 20% to the whole of it, and **Augmented Soldiers** (Cyberpunk Age) 10% more. Campaigns spend soldiers from the stockpile at launch (see [Missions](#missions)).

### Producing soldiers

```
recruit [count|max]     # recruit workers into free housing
assign war_camp 5       # staff a military building; it starts producing soldiers
assign barracks all     # fill a building for its full soldier output
```

`recruit 5` recruits 5 workers into free housing, and `recruit max` fills all of it at once. Workers you leave idle go to work by your [worker shares](workers-and-domains.md#worker-shares) a minute later, and with auto-recruit on (the default) the game recruits into empty slots on its own, military buildings included. A worker assigned to a military building produces soldiers each tick. To keep workers out of military buildings, set `workers share military 0`.

### Food drain

Every worker eats the same amount of food, whatever building it works in (see [Food drain](workers-and-domains.md#food-drain)). Military workers cost no more food than farmers do, and soldiers themselves eat nothing once they're trained. A big military workforce still costs food because it's a big workforce, so check your food rate before you staff up.

### Losing soldiers

Soldiers leave your stockpile when you launch a campaign. The cost is spent up front, whether the campaign then succeeds or fails. The three scouting expeditions cost 0 soldiers. No random event and no setback takes soldiers.

An Endure also costs soldiers: it cuts every stored resource to the share it keeps, soldiers included, so your garrison is weaker afterwards. Your soldiers are measured before the blow lands, so the garrison still counts for that Endure.

To refill the stockpile, keep workers assigned to your military buildings.

### Viewing your army

```
army
```

This opens the **Army** panel: soldier count, defense rating, the running campaign (if any) and how many campaigns you've completed. Under the defense rating it shows what the army does for you:

- **Training:** soldiers per tick. A War Camp or a Barracks can stand before the Iron Age brings soldiers; until then nothing trains in it, and the line reads "not yet. Soldiers arrive in the Iron Age."
- **Threat:** the raid threat of your current age.
- "Your garrison would blunt about N% of a raid.": the share it takes off raids and war raids in your current age (an Endure measures it against the age the catastrophe strikes in). After it comes what twice the garrison would blunt, and a reminder that no army blunts more than 45%. With no soldiers it reads "You have no garrison: raids hit you with full force."
- **Saved this run:** once the garrison has saved something, the buildings, workers and resources (the four largest) it kept, and how many raids it blunted. Like other run stats, it resets with the run.

The running scouting expedition shows in the separate **Expeditions** panel (`expedition`). You can run one scouting expedition and one campaign at the same time. The Loot History (total loot collected) is on the Expeditions panel.

---

## Military Buildings

Military buildings give workers somewhere to work, produce soldiers every tick while staffed, and each one adds its soldier cap to your soldier storage. Your soldier storage is their combined cap plus every storage effect that raises all resources (see Soldiers above).

All military buildings belong to the **military lineage** (22 tiers, Stone Age through Transcendent Age). Each extra copy of a building costs 15% more than the one before.

A fully staffed military building produces about **its soldier cap ÷ 50 per tick** (at least 0.1/tick). A staffed Barracks (cap 20) makes about 0.4 soldiers/tick; a Castle Keep (cap 320) makes about 6.4/tick. Output scales with staffing the same way every other building's does; see [Military Domain Workers](#military-domain-workers).

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

The soldier cap counts per building: two Barracks give 40 soldier storage and, when both are staffed, twice the soldier output.

Five of them wait for a tech of their own age: the Barracks (Military Tactics), the Legion Fort (Siege Warfare), the Cyber Command (Cybersecurity), the Combat Aug Center (Cybernetics) and the Omniversal War Council (Omniversal Command). The others open with their age.

---

## Missions

### What missions are

Missions are timed and paid from your stockpiles. There are two kinds:

- **Scouting expeditions** (`scout_party`, `scout_ruins`, `naval_expedition`) cost resources and 0 soldiers. These are open before soldiers exist.
- **Military campaigns** (the other 13) cost soldiers.

The cost comes out of your stockpiles the moment you launch, and there's no refund. The mission then runs for a number of ticks rolled at launch within its own range, and resolves. A success pays full loot and a failure 30% of it; the cost is gone either way. A failure is also less likely to meet a civilization, and a meeting on a failed run never brings a gift, usually a setback (see [What missions are worth](#what-missions-are-worth)). You can run one scouting expedition and one campaign at the same time, but not two of the same kind.

<figure class="screen" data-screen="expeditions"><figcaption>The Expeditions panel with scouts on the road: the expedition under way, then the ones this age offers, each with its time, risk and cost.</figcaption></figure>

Times on this page are at the base tick of 2 seconds. Tick-speed bonuses (some techs, a Time Dilation boon) make every tick a little shorter.

### Cost, age range and rewards

Campaigns spend soldiers at launch, and you can't launch one you can't afford. The scouting expeditions charge resources up front: `scout_party` costs 30 food and 30 wood, `scout_ruins` 40 food and 30 wood, and `naval_expedition` 150 food and 100 wood.

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

**Three techs open missions.** Campaigns wait for **Military Tactics** (Bronze Age), every scouting expedition past the Scout Party for **Exploration** (Iron Age, after Map Making or Boatbuilding), and the Naval Expedition for **Navigation** (Renaissance Age) on top. Until the tech is researched the command is refused and names it: `Campaigns need Military Tactics first. Research it to send one.` The Scout Party needs no tech. Scout Nearby Ruins is listed from the Bronze Age, an age before Exploration can be researched, and the Expeditions panel says so. Two military buildings wait for a tech too: the Barracks for Military Tactics and the Legion Fort for Siege Warfare. A game saved before this rule keeps what it was already sending for the rest of that run (see [Commands a Tech Opens](technologies.md#commands-a-tech-opens)).

### Automatic dispatch: the Geographic Society

From the **Industrial Age** you can stop sending every party by hand. The **Geographic Society** is a building that sends scouting parties out on its own for the rest of the run. It waits for the **Geographic Societies** tech of that age, which needs Cartography.

It only ever scouts. A Society never wages a campaign; those stay yours to order.

It uses your one scouting slot, and only when the slot is free. Nothing is queued. If a scouting party is out, whether you sent it or the Society did, the Society waits, so a party you launch by hand always goes first.

It pays full price. Each automatic party costs the same resources you would pay. If your stores can't outfit one, the Society holds it back, tells you once, and sends it as soon as the supplies are there. It never runs up a debt.

It sends the cheapest scouting mission your age offers that you can currently afford, so it won't spend your treasury on a voyage you didn't ask for.

Its parties are ordinary parties. They succeed and fail at the same odds, pay the same loot, and roll civilization encounters, boons and setbacks exactly like the ones you send.

The more you invest, the faster it goes. A single unstaffed Society sends about one party every 2,340 ticks (about 1h 18m), roughly what you'd manage by remembering to send one now and then. Staffing it fully cuts that by about a third, and each extra Society shortens the wait again, down to a floor of about **one party every 260 ticks** (about 8m 40s) with six fully staffed Societies. More than that only helps when a party comes home late. Societies are staffed by military-domain workers, so they count toward the [military share of your population](#morale-and-military-ratio).

The **Expeditions** panel (`expedition`) shows one line for it: the time until the next party leaves, or a warning that a party is due and your stores are too thin to outfit it. The full status (Societies built, staffing, the interval, and a countdown with a progress bar) is on the **Factions** panel (`factions`); see [The Factions panel](factions.md#the-factions-panel). Both show approximate wall-clock time rather than ticks.

A Society is always slower than doing it yourself, by design. A fully invested Society runs at about **60% of the pace of a player who chains expeditions by hand and runs campaigns alongside them**. That keeps an idle empire meeting other civilizations and keeps its boons topped up, while playing by hand stays worth it.

The **Expeditions** panel (`expedition`) lists only the scouting expeditions open in your current age. The **Army** panel (`army`) lists only the campaigns, next to your soldier count and defense rating. The Expeditions panel also carries the **Loot History**, the running total of loot from expeditions and campaigns alike.

### Success chance

```
mission power = military power × the age's mission scale
difficulty    = base difficulty - (mission power × 0.3)
difficulty    = max(difficulty, 0.05)
success if a random roll (0 to 1) beats difficulty
```

Military power lowers the effective difficulty. With no military power, a 0.8-difficulty mission succeeds about 20% of the time. The Army panel prints the chance of failure for every campaign, and the chance with your army beside it whenever the army lowers it.

**The mission scale** is the age's [military yardstick](#the-military-yardstick). The missions were sized when techs gave far more military power than the finished tree does (+670% by the Cyberpunk Age, where the tree now gives +176%). The scale puts the difference back: a player who holds every military tech up to their age and the two milestones any garrison earns has the odds the missions were built for, which from the Industrial Age on is the 5% floor on every campaign. With fewer military techs you are proportionally short of it.

| Age | Typical military power | Mission scale | Hardest campaign open | It fails, for the typical player |
|-----|------------------------|---------------|-----------------------|-------------------------------|
| Bronze Age | +15% | 1.33 | `raid_bandits` (0.40) | 34% |
| Iron Age | +45% | 1.44 | `conquer_territory` (0.60) | 40% |
| Classical Age | +60% | 1.75 | `conquer_territory` (0.60) | 28% |
| Medieval Age | +70% | 1.50 | `siege_castle` (0.70) | 38% |
| Renaissance Age | +82% | 1.89 | `siege_castle` (0.70) | 24% |
| Colonial Age | +94% | 1.97 | `siege_castle` (0.70) | 15% |
| Industrial to Electric Age | +106% to +116% | 2.22 to 2.03 | `siege_castle` (0.70) | 5% |
| Atomic and Modern Age | +136% | 3.57 | `world_domination` (0.80, Modern) | 5% |
| Information to Transcendent Age | +146% to +216% | 4.01 to 3.17 | `quantum_incursion` (0.85, Quantum) | 5% |

The soldier and resource cost is already spent at launch, so the outcome only changes the reward and the encounter:

| Outcome | Reward |
|---|---|
| Success | full rewards × (1 + expedition reward bonus) |
| Failure | 30% of the base rewards |

A failure costs no extra soldiers. The launch cost is the whole cost, win or lose. With the **Military-Industrial Complex** researched a campaign brings back 20% more of either row, and with **Probability Warfare** 15% more again; scouting expeditions are not campaigns and keep theirs.

Every resolution logs a plain line: which mission, whether it succeeded, and that loot came in. About one resolution in three also gets a short gray account of how it went underneath. It's kept rare on purpose so those lines stay worth reading. The account depends on the kind of mission, whether it succeeded and your age (a Bronze Age party and a Quantum Age crew come home to different places). It won't repeat a sentence you saw in roughly the last screenful of log, and it's cosmetic only: it never contains a number that isn't on the line above it.

### Full mission table

This table covers all 16 missions. The three scouting expeditions (`scout_party`, `scout_ruins`, `naval_expedition`) are sent with `expedition <key>`; every other row is a campaign, waged with `campaign <key>`. The Soldier Cost column is spent from your soldiers at launch.

| Key | Name | Min Age | Soldier Cost | Resource Cost | Duration | Difficulty | Rewards on Success |
|-----|------|---------|--------------|---------------|----------|------------|-------------------|
| `scout_party` | Scout Party | Primitive Age (last: Bronze Age) | 0 | 30 food, 30 wood | 100-160t (260-416t in the Bronze Age) | 0.20 | 60 food, 60 wood, 20 stone |
| `scout_ruins` | Scout Nearby Ruins | Bronze Age | 0 | 40 food, 30 wood | 156-260t | 0.20 | 30 food, 20 wood, 15 stone |
| `raid_bandits` | Raid Bandit Camp | Bronze Age | 5 | none | 156-260t | 0.40 | 30 gold, 15 iron, 20 food |
| `trade_escort` | Trade Escort | Iron Age | 3 | none | 156-260t | 0.30 | 50 gold, 10 knowledge |
| `conquer_territory` | Conquer Territory | Iron Age | 10 | none | 156-260t | 0.60 | 80 gold, 40 iron, 50 food |
| `siege_castle` | Siege Enemy Castle | Medieval Age | 15 | none | 156-260t | 0.70 | 150 gold, 30 steel, 20 faith |
| `naval_expedition` | Naval Expedition | Renaissance Age | 0 | 150 food, 100 wood | 156-260t | 0.50 | 200 gold, 30 culture, 40 knowledge |
| `colonial_campaign` | Colonial Campaign | Industrial Age | 20 | none | 156-260t | 0.60 | 300 gold, 50 oil, 40 steel |
| `world_domination` | World Domination | Modern Age | 50 | none | 156-260t | 0.80 | 1K gold, 200 electricity, 500 knowledge |
| `cyber_raid` | Cyber Raid | Information Age | 30 | none | 156-260t | 0.60 | 200 data, 50 crypto, 500 gold |
| `neon_heist` | Neon Heist | Cyberpunk Age | 25 | none | 156-260t | 0.55 | 100 crypto, 150 data, 800 gold |
| `fusion_assault` | Fusion Plant Assault | Fusion Age | 35 | none | 156-260t | 0.65 | 120 plasma, 500 electricity, 50 uranium |
| `orbital_strike` | Orbital Strike | Space Age | 40 | none | 156-260t | 0.70 | 100 titanium, 80 plasma, 300 knowledge |
| `warp_invasion` | Warp Invasion | Interstellar Age | 60 | none | 169-273t | 0.75 | 50 dark matter, 200 titanium, 2K gold |
| `galactic_conquest` | Galactic Conquest | Galactic Age | 80 | none | 208-338t | 0.80 | 30 antimatter, 100 dark matter, 5K gold |
| `quantum_incursion` | Quantum Incursion | Quantum Age | 100 | none | 234-377t | 0.85 | 20 quantum flux, 50 antimatter, 5K knowledge |

That is 3 scouting expeditions and 13 campaigns.

The Duration column is each mission's range in ticks, as the panel shows it. From the Bronze Age on, where ages and their timers run 2.6 times as long, every range is 2.6 times its base length, so the column already includes that. The actual time is rolled evenly within that range at launch, so nothing resolves in under 100 ticks (156 from the Bronze Age on), and `scout_party` runs 100-160 ticks (about 130 on average; 260-416 in the Bronze Age). The time-left readout counts down the rolled value. Techs shorten the range, and the panels list it as it stands: Cartography, Aviation, Reusable Launchers and Fusion Drives take 10% each off scouting expeditions, Wormholes 15% and Stellar Cartography 20% (45% of the listed time with all six), and a **General Staff** (Victorian Age) and **Special Forces** (Modern Age) take 15% each off campaigns.

There are no soldiers before the Iron Age, so scouting fills the gap. `scout_party` costs 30 food and 30 wood, runs 100-160 ticks (260-416 in the Bronze Age) and pays about 60 food, 60 wood and 20 stone, a net gain worth repeating through the Primitive, Stone and Bronze ages. It disappears once you reach the Iron Age. `scout_ruins` (Bronze Age, 40 food and 30 wood) carries scouting on from there, and `naval_expedition` (Renaissance Age, 150 food and 100 wood) is the late scouting option, harder (0.50) than `scout_ruins` (0.20).

`campaign trade_escort` (Iron Age, 3 soldiers, 156-260 ticks, 0.30 difficulty) is the cheapest, easiest campaign to keep running. Durations are rolled, so keep one of each kind running rather than counting on a fixed timer.

### What missions are worth

**Loot is fixed.** Every mission pays the same bundle in every age it runs, and the bundles were sized for a much smaller economy. World Domination's 1K gold arrives in the Modern Age, where a typical building costs tens of billions of gold. Early on, `scout_party` and the first campaigns are a real help; from about the Iron Age, loot is small change.

**Encounters are the payoff.** Every resolved mission rolls a chance to meet a civilization: about 18% for a successful scouting run and 8% for a successful campaign, and much less for failures (6% and 2%). An encounter is how you make first contact, and it usually brings a boon: a timed production boost, a lump of resources or a crew of workers. It brings a setback instead when the run failed, and sometimes when the civilization is at war with you. A player who keeps a scouting party and a campaign in the field all the time meets someone about every 975 ticks (about 32 minutes). The full rules (what the boons and setbacks are, the five-boon limit, war) are on [Factions & Diplomacy](factions.md#encounters-boons-and-setbacks).

So keep one scouting expedition and one campaign running after you have met everyone, and prefer missions you're likely to win: a success is three to four times as likely to meet someone, and a failure that does meet someone brings trouble home.

---

## Military Domain Workers

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

## Military Power Bonus

Military power is a running total that lowers mission difficulty (see [Success chance](#success-chance)) and feeds the defense rating:

```
defense = soldiers × 2.0 × (1 + military power)
```

The defense rating is what your garrison is measured by against raids and catastrophes (see [Defense](#defense-what-your-army-blunts)), so military power pays twice: it raises mission success chances, and it makes the same soldiers blunt more. A +1.0 bonus doubles the defense rating of the same army.

Sources add together:

| Source | How to get it | Bonus |
|--------|-------------|---------------|
| Techs | Eighteen military techs grant military power | +0.10 to +0.15 per tech, +2.01 with all eighteen |
| Milestones | Complete military milestones | +0.05 to +0.10 each |

There's no cap on military power, but difficulty never drops below **0.05** (a 5% minimum failure chance). A mission reads military power through the age's mission scale (see [Success chance](#success-chance)), so a player who keeps up with the military techs reaches that floor on every campaign from the Industrial Age on, and a player who skips them does not.

### The military yardstick

Two things are measured against military power: the raid threat (through the defense rating) and a mission's difficulty. Both were sized against a tree that gave much more of it than the finished one. Each age has a yardstick that puts the difference back, so the **typical player** stands where the game was measured: one who holds the military techs of every age up to their own, plus First Soldiers and War Machine (+15% together, which the buildings the age gates ask for earn unstaffed).

- The **threat scale** multiplies the age's raid threat: (1 + typical power now) ÷ (1 + typical power then). The typical player's garrison blunts what it blunted when the threat was measured.
- The **mission scale** multiplies military power where a mission reads it: typical power then ÷ typical power now. The typical player's missions fail as often as they did.

More military power than the typical player's still helps, and less still costs, in the same proportion as before. The yardstick is read off the tech tree, so it moves with it.

The expedition reward bonus (from techs, milestones and some wonders) is separate. It multiplies the loot on a success: `rewards × (1 + expedition reward bonus)`.

---

## Defense: what your army blunts

Soldiers you keep in stock are your **garrison**. They are never spent by defending: the garrison takes a share off what raids and catastrophes cost you, and the soldiers stay where they are.

### The formula

Your defense rating is measured against the **raid threat** of the age you are in:

```
defense  = soldiers × 2.0 × (1 + military power)
threat   = 160,000 × 2^(age order) × the age's threat scale      # Primitive Age = order 0
blunted  = 45% × defense ÷ (defense + threat)
```

The blunted share is the part of a raid's losses the garrison takes off. It is 0 with no soldiers, half the ceiling (22.5%) when your defense equals the threat, and it approaches 45% without ever reaching it. More soldiers always help, but each doubling helps less than the last:

| Defense vs threat | Share blunted |
|-------------------|---------------|
| none | 0% |
| one third of the threat | about 11% |
| half the threat | 15% |
| equal to the threat | 22.5% |
| twice the threat | 30% |
| four times the threat | 36% |
| nine times the threat | about 40% |

The threat **about doubles every age**, the same rate at which each new military building's soldier cap doubles (the Atomic Age is the exception: its threat is a quarter above the Electric Age's). An army that is strong for one age is ordinary for the next and a rounding error a few ages later, so a garrison has to grow with you. The threat scale is the age's [military yardstick](#the-military-yardstick).

| Age | Threat scale | Threat | Soldiers for 22.5% (no military bonus) |
|-----|--------------|--------|----------------------------------------|
| Iron Age | 0.88 | 1.12M | 562K |
| Classical Age | 0.78 | 2M | 999K |
| Medieval Age | 0.83 | 4.25M | 2.12M |
| Renaissance Age | 0.71 | 7.31M | 3.65M |
| Colonial Age | 0.68 | 13.9M | 6.97M |
| Industrial Age | 0.61 | 25.2M | 12.6M |
| Victorian Age | 0.61 | 50.4M | 25.2M |
| Electric Age | 0.64 | 106M | 52.8M |
| Atomic Age | 0.40 | 132M | 66.1M |
| Modern Age | 0.40 | 264M | 132M |
| Information Age | 0.36 | 471M | 235M |
| Digital Age | 0.37 | 980M | 490M |
| Cyberpunk Age | 0.35 | 1.84B | 922M |
| Fusion Age | 0.36 | 3.82B | 1.91B |
| Space Age | 0.36 | 7.64B | 3.82B |
| Interstellar Age | 0.38 | 15.8B | 7.91B |
| Galactic Age | 0.39 | 32.7B | 16.3B |
| Quantum Age | 0.40 | 67.5B | 33.8B |
| Transcendent Age | 0.40 | 135B | 67.5B |

The Primitive, Stone and Bronze Ages have threats of 160K, 320K and 613K, but soldiers don't exist until the Iron Age. A `military_power` bonus cuts the soldiers needed: at +1.0, half as many.

### The garrison you already have

Every player carries a garrison without trying. The age gates ask for military buildings (15 Hunting Lodges for the Classical Age, 15 Military Academies for the Medieval, 3 Castle Keeps for the Renaissance, 15 Bunker Complexes for the Modern, 10 Plasma Commands for the Space Age, 15 Probability War Rooms for the Transcendent). You keep those buildings into every later age, and they train soldiers even with no workers assigned, at a fifth of the staffed rate. A player who builds only what the gates ask for blunts roughly **8-19%** of a raid from the Iron Age on (measured on the smoke-test bot, which does nothing else for its army).

A deliberate army is what pushes toward the 45% ceiling: more military buildings of the current age, workers staffed into them, and `military_power` research.

### What it blunts

1. **Raid events.** Random events that are attacks by outsiders: Bandit Raid, Pirate Attack, Data Breach, The Great Breach, Corporate Espionage, and the Stone Era's Tribal Raid and Beast Stampede. The garrison cuts the resources they steal and the workers they drive off by its share (at least one worker still flees if the event takes workers). A raid's production penalty is not blunted. See [Raids and your garrison](events.md#raids-and-your-garrison). The log adds a line under the event: "Your garrison blunted about N% of the raid: you kept ...".
2. **War raids.** While a civilization is at war with you, each raid it makes takes a resource. The garrison keeps its share of that resource. A raid bigger than your stock still takes nothing, as before; the army never makes a raid that missed land. The raid's log line adds "Your garrison kept X gold from them (about N% of the raid)." (with the raid's resource in place of gold). See [War & Peace](factions.md#war-amp-peace).
3. **Endure.** When you Endure a catastrophe, the garrison blunts its share of the buildings destroyed and the stock lost, after the Harbinger's Brace, measured against the threat of the age the catastrophe strikes in (see below).

**Four techs make raids smaller before the garrison meets them.** Imperial Legions, Nuclear Deterrence and Orbital Defense each take 10% off what a raid event or a war raid would steal, and Fortification 15% (38% with all four: the cuts multiply). The garrison then blunts its share of what is left. They do nothing for an Endure.

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

The catastrophe window shows the real numbers after Brace and garrison, with a line for each; the Harbinger panel's Brace preview counts the garrison too. After an Endure the log says "Your garrison held the line: ...". See [Endure](catastrophe.md#endure) and [Brace](harbinger.md#brace-soften-an-endure).

The Last Passage in the Cosmic Era costs prestige points, not buildings or stock, and the garrison doesn't change it.

### Worked example

You hold 2,108,000 soldiers with no `military_power` bonus: defense 2,108,000 × 2 = 4,216,000.

- **In the Classical Age** (threat 2M): 45% × 4.216M / (4.216M + 2M) = about **31%**. A war raid that would take 1,000 gold takes about 690; the Army panel says twice the garrison would blunt about 36%.
- **In the Industrial Age** (threat 25.2M): 45% × 4.216M / (4.216M + 25.2M) = about **6%**. The same army barely matters four ages later.
- **Endure at the Renaissance** (threat 7.31M) with Brace 1, 150 buildings that can fall (everything but wonders and storage) and 1,827,500 soldiers (defense 3.655M, half the threat): the garrison share is 15%. Brace 1 alone would destroy 22 buildings and keep 30% of stock; with the garrison, 19 fall (22 × 15% = 3.3, rounded down to 3 saved) and about 40% of stock is kept.

---

## Military Milestones

The five military milestones form a chain. Completing all five grants a title, and between them they give **+0.25 military power** and **+25% all production** (which adds into [the all-production pool](resources.md#the-all-production-cap)). Soldier counts in these milestones are soldiers trained over the run, not the number in stock; missions don't count toward any of them.

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

## Strategy

### Early game (Stone Age to Classical Age)

- Before the Iron Age, run `expedition scout_party` on repeat. It costs 30 food and 30 wood and pays about 60 food, 60 wood and 20 stone, with no soldiers needed. From the Bronze Age its encounters can also meet your first civilization.
- War Camps and Barracks train nothing before the Iron Age, when soldiers unlock. Build Barracks for `iron_legion` (10 Barracks), not for early soldiers.
- Once soldiers unlock, staff your military buildings (the Hunting Lodge costs only wood) and train 5 soldiers for `first_soldiers` and its +0.05 military power.
- `campaign trade_escort` (3 soldiers, Iron Age, 156-260 ticks, 0.30 difficulty) is the easiest early campaign: keep it running next to a scouting party for the encounters.
- Workers in military buildings eat the same food as everyone else, but every worker you move there is one fewer farmer. Keep your food rate positive.

### Mid game (Classical to Industrial Age)

- Go for `iron_legion` (train 500 soldiers, build 10 Barracks) for its +5% all production.
- Pick the easiest campaign you can afford rather than the one with the best loot. With no military power, `conquer_territory` (0.60) fails more often than it wins, and each failure that meets someone brings a setback.
- For scouting, `scout_ruins` (0.20) is cheaper and easier than `naval_expedition` (0.50), so it meets more civilizations per run. The Geographic Society sends the cheapest one on its own.
- The age gates ask for military buildings (15 Hunting Lodges, 15 Military Academies, 3 Castle Keeps), and staffed they build your garrison too.
- Research military techs as they appear. Each one counts for more than its figure: a mission reads military power through the age's mission scale, so +0.15 in the Iron Age takes about 6 points off a campaign's chance to fail.

### Late game (Modern Age onward)

- Campaign soldier costs (at most 100) are tiny next to what your buildings train by now, so wage whatever campaign is open. The loot doesn't matter; the encounter roll does.
- With the military techs of each age researched, every campaign sits at the 0.05 difficulty floor from the Industrial Age on, so nearly every run succeeds and its encounters can bring boons. Skip them and the late campaigns are a gamble again.
- The `military_superpower` milestone (train 2,000 soldiers) adds +15% all production.

### Defense rating

The Army panel shows your **defense rating** (soldiers × 2.0 × (1 + military power)). It is measured against the raid threat of your age, and the resulting share (at most 45%) comes off raid events, war raids from civilizations at war with you, and the buildings and stock an Endure takes. It never stops a raid or a catastrophe from happening, and it does nothing against disasters such as plague or earthquakes. A civilization at war with you raids every 104 ticks and takes 50 × its strength (1-5) of its specialty resource; your garrison keeps its share of that, but only staying out of wars, or ending them with tribute, stops the raids: see [War & Peace](factions.md#war-amp-peace). See [Defense](#defense-what-your-army-blunts) for the numbers.

- **Keep up with the age.** The threat about doubles every age, so a garrison you stop growing fades fast. Adding the current age's military buildings keeps pace, since each tier's soldier cap doubles too.
- **Mind the next age.** An Endure is measured against the age the doom strikes in. The Harbinger panel's Brace preview uses the age you are in; if you advance before the doom strikes, the threat about doubles and your garrison blunts less than the preview showed.
- **Brace and garrison stack, up to a point.** Together they cut an Endure's losses by at most 60%. With Brace 2, a garrison that blunts 20% already reaches the building cap (8% fall), so extra soldiers mostly buy stock kept, up to 66%.
- **Soldiers cost nothing to hold.** The stock has no upkeep; only the military workers producing it eat food and count toward the morale ratio below. You can staff up to bank a garrison, then move the workers back.
- **Research military power.** It multiplies the defense rating of the soldiers you already have.

### Army size and food

Soldiers eat nothing, and a worker in a military building eats no more than any other worker. The food cost of an army is simply that its workers are workers. Rough rules:

- Before you recruit, check `rates` and make sure your food rate stays positive with the new workers added.
- Use `recruit max` only when food is overflowing, so a big recruit doesn't tip you into a deficit.
- After an age advance, check food again: the cost per worker goes up by ×1.12 each age.

### Morale and military ratio

A large army has a second cost: **morale**. If more than **30% of your population** works in military buildings (Geographic Societies included), morale drops every tick, and the further over 30% you are, the faster it drops:

| Military share | Over 30% by | Morale lost per tick |
|---------------|---------|--------------|
| 30% | 0 | none |
| 40% | 10 points | 0.3 points |
| 50% | 20 points | 0.6 points |
| 60% | 30 points | 0.9 points |

Morale multiplies the output of every worker in your civilization, on a continuous curve centered on 50%. Below 50% it cuts production, down to ×0.50 at the 10% floor. Above 50% it raises production, up to +20% at the top. An oversized army that drags morale down therefore slows every domain at once (food, knowledge, trade and the rest), and the army's food gets harder to cover as your farmers produce less. A lean army leaves room to push morale above 50% instead.

**Keep military workers at 25-28% of your population or less.** That leaves a buffer below 30% if you lose workers to an event. If you need a large soldier stockpile, staff your military buildings heavily to bank soldiers, then move the extra workers back to civilian buildings. The stored soldiers stay, and the morale drain stops.

Morale drifts back toward 50% by itself, a little each tick, once you're back under 30%. If morale is falling and more than 30% of your workers are in military buildings, unassign some of them or recruit more workers for other buildings.

See [Morale](morale.md) for the full system.

---

## Tips & Common Mistakes

**Don't recruit past your food income.** A food deficit stalls all production, because starving workers can't work. Work out the extra food before `recruit max`.

**Don't wage hard campaigns without military power.** `campaign siege_castle` (0.70) and `campaign world_domination` (0.80) fail often with no bonus, and a failure meets fewer civilizations and brings setbacks home. Research a few military techs first.

**A failed mission still costs its launch.** The soldiers and any resources are spent when you launch. Failure doesn't refund them; it only cuts the reward to 30%.

**Workers need a built building.** You can't assign workers to a building you haven't built yet.

**Plan Castle Keeps early for `fortress_state`.** It needs 20 Castle Keeps (Medieval Age), a large stone and iron investment, so start building them as soon as you reach the Medieval Age. The +0.10 military power and +5% all production are worth the cost.

**Keep missions running.** There's no cooldown beyond the mission's own duration (100 to 160 ticks in the Primitive and Stone Ages, 156 to 416 from the Bronze Age on, rolled at launch). When one resolves, send the next.

**Prestige resets the army.** Soldiers, military buildings and the military power from techs and milestones all start over with the run. What a prestige keeps from your missions is the civilizations they found: with [Old Friends](prestige.md#old-friends) from the legacy kit, each is met again as soon as you reach its age.
