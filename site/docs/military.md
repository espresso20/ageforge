# Army & Missions

## Road events

Parties can meet road events while an expedition is out. Storms delay them, helpers save time, finds increase the haul, tolls cost gold, foes lower the odds, ambushes cost soldiers or loot, and fair weather raises the odds. A party can also mark ruins, adding a short offer to the board. Events are rolled when a mission launches and appear in the log and its Expedition Report.

Soldiers are a resource your military buildings train. Military grand offers cost soldiers; short, long and civil grand offers cost materials. Mission hauls scale with the current age's income. What missions are really for is meeting the other civilizations: every resolved mission can bring an encounter, which is how you make first contact and collect boons (or, when a run fails, setbacks). The soldiers you keep in stock are your garrison: they blunt raids and soften catastrophes. Ignore the army and you still carry the small garrison the age gates build for you; invest in it and raids and catastrophes take noticeably less. The soldiers you train also count toward the military milestones.

---

## Overview

The military system covers four things:

- **Missions.** The mission board has short, long and grand offers, each with rolled length, odds and haul; see [Missions](#missions).
- **Defense rating.** A figure on the Army panel worked out from your soldier count and military power. Measured against the raid threat of your current age, it blunts part of what raids, war raids and an Endure take from you (at most 45%); see [Defense](#defense-what-your-army-blunts).
- **Milestones.** Five military milestones form a chain with a title at the end. The early ones grant permanent military power; the later ones grant all production.
- **Prestige.** The army resets with the run. Prestige points give a lasting production bonus, not military power or mission rewards. From the first prestige, [Old Friends](prestige.md#old-friends) meets every civilization you have met again on later runs without a mission.

**Soldiers are a resource**, not a count of workers. Your military buildings produce and store them: staff a War Camp or Barracks with workers and it adds to the `soldiers` resource every tick, the same way a Farm adds food. Soldiers begin with your first military building: a War Camp can be built in the Stone Age and trains them from the day it stands. A town with no military building has none until the **Iron Age**, which brings them to every town. Short, long and civil grand offers (see [Missions](#missions)) let you explore without soldiers.

The mission board has three tiers. A grand offer is either military or civil. See [The board](#the-board) for the costs and odds.

---

## Soldiers

### What soldiers are

Soldiers are a stockpiled resource (`soldiers`). They unlock when your first military building stands, whatever the age (a War Camp in the Stone Age, a Barracks in the Bronze), and in the **Iron Age** for a town that built none. You bank them the way you bank food or wood, then spend them on campaigns. Until then they stand as your garrison, which blunts raids.

Staffed military buildings produce soldiers every tick. A fully staffed building makes about its soldier cap ÷ 50 per tick (at least 0.1/tick), so more workers and more buildings mean faster soldiers. Your soldier storage is the sum of every military building's soldier cap, plus every storage effect that raises all resources (the general storage buildings and the "all storage" techs). That second part is most of it: by the Iron Age a typical civilization can hold a couple of hundred thousand soldiers, far more than the military buildings' own caps. The **Military-Industrial Complex**, the Atomic Age's Military capstone, adds 20% to the whole of it, and **Augmented Soldiers** (Cyberpunk Age) 10% more. Campaigns spend soldiers from the stockpile at launch (see [Missions](#missions)).

### Producing soldiers

```
recruit [count|max]     # recruit workers into free housing
assign war_camp 5       # staff a military building; it starts producing soldiers
assign barracks all     # fill a building for its full soldier output
```

`recruit 5` recruits 5 workers into free housing, and `recruit max` fills all of it at once. Workers you leave idle go to work by your [roster](workers-and-domains.md#the-roster) a minute later, and with auto-recruit on (the default) the game recruits into empty slots on its own, military buildings included. A worker assigned to a military building produces soldiers each tick. To keep workers out of military buildings, set `roster military 0`.

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

- **Training:** soldiers per tick. Before your first military building it reads "not yet. A military building trains soldiers."
- **Threat:** the raid threat of your current age.
- "Your garrison would blunt about N% of a raid.": the share it takes off raids and war raids in your current age (an Endure measures it against the age the catastrophe strikes in). After it comes what twice the garrison would blunt, and a reminder that no army blunts more than 45%. With no soldiers it reads "You have no garrison: raids hit you with full force."
- **Saved this run:** once the garrison has saved something, the buildings, workers and resources (the four largest) it kept, and how many raids it blunted. Like other run stats, it resets with the run.

The running missions show in the separate **Expeditions** panel (`expedition`), which also displays the board and the Geographic Society.

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

### The board

The mission board holds nine offers, three short, three long and three grand. Each offer has a title and description made from the current age's words. The board is saved with the run and uses its seeded random generator.

| Tier | Length | Base odds | Haul |
|---|---:|---:|---:|
| Short | 5 to 30 minutes | 85%, plus or minus 10 points | 2 times the rolled length in minutes of income |
| Long | 30 to 60 minutes | 75%, plus or minus 10 points | 3 times the rolled length in minutes of income |
| Grand | 60 to 120 minutes | 65%, plus or minus 10 points | 4 times the rolled length in minutes of income |

The board's offer shows the rolled length, success odds, haul and launch cost. The haul is based on the current age's income and reward bonuses. Failure pays 30% of the haul. A Legendary offer is longer, has lower odds and pays four times the ordinary haul. About 5% of offers are Legendary, never two within 10 offers of each other, with at least one in every 41 offers.

Short and long offers cost the two materials the current age's buildings charge most, sized to about 5 and 15 minutes of their income. A military grand offer costs soldiers and uses the army in its odds. A civil grand offer costs about 45 minutes of income in the age's two most-charged materials and uses plain odds. The offer shows its cost before you send it.

| Age | Short | Long | Grand, military | Grand, civil |
|---|---|---|---|---|
| Primitive | Forage | Hunt | Raid | Great Migration |
| Stone | Forage | Trek | Raid | Hill Fort Raising |
| Bronze | Scout | Trade Trek | Skirmish | Trade Mission |
| Iron | Survey | Caravan | Siege | Founding Expedition |
| Classical | Survey | Trade Voyage | Legion March | Embassy |
| Medieval | Patrol | Pilgrimage | Siege | Grand Pilgrimage |
| Renaissance | Exploration | Voyage | Campaign | Grand Voyage |
| Colonial | Survey | Colony Run | Campaign | Found a Colony |
| Industrial | Prospect | Rail Expedition | Offensive | Great Railway |
| Victorian | Prospect | Expedition | Offensive | Polar Expedition |
| Electric | Aerial Survey | Expedition | Offensive | Grand Exhibition |
| Atomic | Recon Flight | Research Expedition | Operation | Manhattan-scale Project |
| Modern | Recon | Field Op | Strike | Moonshot |
| Information | Data Run | Satellite Op | Cyber Strike | Global Launch |
| Digital | Data Run | Network Op | Cyber Strike | Network Founding |
| Cyberpunk | Recon | Heist | Assault | Arcology Raising |
| Fusion | Sweep | Salvage | Assault | Fusion City |
| Space | Probe | Orbital Mission | Orbital Strike | Colony Ship |
| Interstellar | Probe | Deep-Space Mission | Invasion | Generation Ship |
| Galactic | Survey Drone | Star Voyage | Invasion | Stellar Settlement |
| Quantum | Phase Probe | Rift Expedition | Incursion | Pocket Universe |
| Transcendent | Glimpse | Pilgrimage of Light | Reality Incursion | Ascension Rite |

Sending an offer removes it temporarily. A replacement arrives after five minutes. Untaken offers expire after an hour, and entering a new age creates a new board. You can have one mission of each tier under way at once. The panel and `expedition list` show why an offer cannot be sent now. A taken slot shows when its replacement arrives.

### The Expeditions panel

<figure class="screen" data-screen="expeditions"><figcaption>The Expeditions panel: the Ready Room, a map of the age's lands with every offer as a pin, and the briefing for the one you pick.</figcaption></figure>

`expedition` opens the **Expeditions** panel. It has two views, and `Tab` switches between them. The soldiers at home are on the right of the top line.

The **Ready Room** is where you read the map and send a party. The map shows the age's lands with your town in the middle and every offer on the board as a pin:

| Pin | Offer |
|---|---|
| ○ | Short |
| ◇ | Long |
| ▲ | Grand, military |
| ■ | Grand, civil |
| ◈ | Legendary, in gold |

Short offers lie near the town, long ones farther out and grand ones at the edge of the map. The land is drawn from your run and the age, so it holds still while the age lasts and changes with the next. A mission under way shows a dotted route from the town, an `@` where its party is and a ◆ at its goal.

The highlighted pin has a dotted route from the town and its place name beside it. Next to the map the briefing for that offer gives its twist, its odds, how long it is out and back, how long the offer stays up, its haul in minutes of income and in resources, and its cost (green where you can pay it, red where you cannot). A military grand offer also lists the soldiers it takes as its party. The card ends with `Enter send`, or with the reason the offer cannot be sent now. Under the map are the missions out now, a key to the pins and the time until the board next changes.

The arrow keys hop to the nearest pin in that direction. `Enter` opens a box that asks before the party goes, and `Enter` again sends it. `Esc` takes the question back, and closes the panel when there is no question.

The **Expedition Report** follows the parties out. It shows the same map with every mission under way, and a block for each: where it goes, a progress bar, the time left, its odds and the soldiers of a military mission. `Up` and `Down` pick a party, and its route lights up on the map. Below, **Came home** lists the last ten missions to return, newest first, with a ● on the ones that came home since you last looked, and the Geographic Society's line. With nobody out, the Report says so and how many offers are waiting in the Ready Room.

The panel opens on the Report while a party is out or one has come home since you last looked at it, and on the Ready Room otherwise. Looking at the Report clears the `● new` count beside its tab. The prompt keeps working while the panel is open; with something typed, `Tab` and `Enter` act on the prompt instead.

### Commands

```
expedition              # open the Expeditions panel
expedition list         # print the mission board
expedition <name>       # send the offer with that name
campaign                # alias for expedition
campaign <name>         # alias for expedition <name>
```

Offers have no numbers. To send one from the prompt, type part of its title after `expedition`. Case does not matter, and "the", "of", "a" and "an" are left out. An exact title wins; otherwise a title that starts with what you typed; otherwise any title in which every word you typed begins one of its words. If one offer matches, it is sent. If several do, the game lists them and asks you to type more of the name. If none does, it says so.

A mission's resource or soldier cost is spent at launch, whether it succeeds or fails. One mission per tier can be under way at once.

### Automatic dispatch: the Geographic Society

From the **Industrial Age** the **Geographic Society** sends short or long offers on its own for the rest of the run. It waits for the **Geographic Societies** tech of that age, which needs Cartography.

It only ever scouts. A Society never wages a campaign; those stay yours to order.

It uses the offer's tier slot, and only when that slot is free. Nothing is queued. A party you launch by hand always goes first for its tier.

It pays full price. Each automatic mission costs the same materials you would pay. If your stores can't outfit one, the Society holds it back, tells you once, and sends it as soon as the supplies are there. It never runs up a debt.

It sends the affordable short or long offer with the best haul per minute, so it won't spend your treasury on a grand mission.

Its missions are ordinary missions. They use the same odds, pay the same haul and roll civilization encounters, boons and setbacks exactly like the ones you send.

The more you invest, the faster it goes. A single unstaffed Society sends about one party every 2,340 ticks (about 1h 18m), roughly what you'd manage by remembering to send one now and then. Staffing it fully cuts that by about a third, and each extra Society shortens the wait again, down to a floor of about **one party every 260 ticks** (about 8m 40s) with six fully staffed Societies. More than that only helps when a party comes home late. Societies are staffed by military-domain workers, so they count toward the [military share of your population](#morale-and-military-ratio).

The **Expedition Report** (the second view of the `expedition` panel) shows one line for it: the time until the next party leaves, or a warning that a party is due and your stores are too thin to outfit it. The full status (Societies built, staffing, the interval, and a countdown with a progress bar) is on the **Factions** panel (`factions`); see [The Factions panel](factions.md#the-factions-panel). Both show approximate wall-clock time rather than ticks.

A Society is always slower than doing it yourself, by design. The Society keeps missions moving while you are away. Sending offers by hand gives you control over the odds, haul and cost.

The **Expeditions** panel (`expedition`) shows every offer and all missions under way. The **Army** panel (`army`) shows your soldier count and defense rating.

### Success chance

Each offer rolls its success odds when it appears. A military grand offer also adjusts its odds for your army. The panel shows the odds that apply now. The army's military power and its effects on raids are covered in [Defense](#defense-what-your-army-blunts).

### What missions are worth

**Hauls scale with income.** The board shows the expected resource haul before launch. Longer missions and Legendary offers can pay more, while failure pays 30% of the haul.

**Encounters are one payoff.** Every resolved mission can meet a civilization. An encounter is how you make first contact, and it usually brings a boon: a timed production boost, a lump of resources or a crew of workers. It brings a setback instead when the run failed, and sometimes when the civilization is at war with you. The full rules (what the boons and setbacks are, the five-boon limit, war) are on [Factions & Diplomacy](factions.md#encounters-boons-and-setbacks).

One mission per tier can be under way at once. Compare its odds and haul before sending it.

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
threat   = 160,000 × 2^(age order) × the age's threat scale      # from the Iron Age (order 3) on
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

A `military_power` bonus cuts the soldiers needed: at +1.0, half as many.

### Before the Iron Age

The doubling curve was measured from the Iron Age on, where every town carries the military buildings the age gates asked for. Before it, the only military building a gate asks for is the War Camp (eight of them, to leave the Stone Age), so a garrison is mostly the War Camps and Barracks you chose to build, and the three early ages have raid threats sized to that:

| Age | Threat | Soldiers for 22.5% (no military bonus) | What trains them |
|-----|--------|----------------------------------------|------------------|
| Primitive Age | 1K | none can train | no military building yet |
| Stone Age | 4.8K | 2,400 | War Camp: 0.2 soldiers/tick with its 3 posts staffed, a fifth of that with none |
| Bronze Age | 69K | 34.5K | Barracks (after Military Tactics): 0.4 soldiers/tick with its 4 posts staffed |

Eight staffed War Camps train 1.6 soldiers a tick. Over a Stone Age of ordinary length that garrison comes to blunt about a fifth of a raid, and half of it about an eighth. In the Bronze Age the same holds for eight War Camps and eight Barracks. A garrison carried into the next age meets that age's larger threat, so it has to keep growing, as it does later.

The raids of these ages are Tribal Raid and Beast Stampede (Stone Era) and Bandit Raid (Bronze Age on). See [Raids and your garrison](events.md#raids-and-your-garrison).

### The garrison you already have

Every player carries a garrison without trying. The age gates ask for military buildings (8 War Camps for the Bronze Age, 15 Hunting Lodges for the Classical Age, 15 Military Academies for the Medieval, 8 Castle Keeps for the Renaissance, 18 Bunker Complexes for the Modern, 10 Plasma Commands for the Space Age, 15 Probability War Rooms for the Transcendent, and eight of each age's own military building in most of the ages between; see [Advancement Requirements](ages.md#advancement-requirements)). You keep those buildings into every later age, and they train soldiers even with no workers assigned, at a fifth of the staffed rate. A player who builds only what the gates ask for blunts roughly **8-19%** of a raid from the Iron Age on (measured on the smoke-test bot, which does nothing else for its army).

A deliberate army is what pushes toward the 45% ceiling: more military buildings of the current age, workers staffed into them, and `military_power` research.

### What it blunts

1. **Raid events.** Random events that are attacks by outsiders: Bandit Raid, Pirate Attack, Data Breach, The Great Breach, Corporate Espionage, and the Stone Era's Tribal Raid and Beast Stampede. The garrison cuts the resources they steal and the workers they drive off by its share (at least one worker still flees if the event takes workers). A raid's production penalty is not blunted. See [Raids and your garrison](events.md#raids-and-your-garrison). The log adds a line under the event: "Your garrison blunted about N% of the raid: you kept ...".
2. **War raids.** While a civilization is at war with you, each raid it makes takes a resource. The garrison keeps its share of that resource. A raid bigger than your stock still takes nothing, as before; the army never makes a raid that missed land. The raid's log line adds "Your garrison kept X gold from them (about N% of the raid)." (with the raid's resource in place of gold). See [War & Peace](factions.md#war-amp-peace).
3. **Endure.** When you Endure a catastrophe, the garrison blunts its share of the buildings destroyed and the stock lost, after the Harbinger's Brace, measured against the threat of the age the catastrophe strikes in (see below).

**Four techs make raids smaller before the garrison meets them.** Imperial Legions, Nuclear Deterrence and Orbital Defense each take 10% off what a raid event or a war raid would steal, and Fortification 15% (38% with all four: the cuts multiply). The garrison then blunts its share of what is left. They do nothing for an Endure.

Nothing else. Disasters and unrest (earthquakes, plague, mine collapses, industrial accidents, uprisings) are not raids, and soldiers do nothing against them. Endure's 25% worker loss, its production debuff and its morale hit are unchanged.

A player with **no soldiers** takes exactly the losses they always did. From the Iron Age on that is rare: the military buildings the age gates require give nearly everyone some garrison (see [The garrison you already have](#the-garrison-you-already-have)).

### Endure: Brace first, then the garrison

A catastrophe strikes at a fated moment anywhere in its era (Iron to Cosmic), so the garrison meets the threat of whatever age you are in when it strikes. A town that built no War Camp or Barracks has its first soldiers in the Iron Age, so an Iron Era doom that strikes early in the Iron Age finds little garrison to count.

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
- **Endure at the Renaissance** (threat 7.31M) with Brace 1, 150 buildings that can fall (everything but wonders, storage and knowledge buildings) and 1,827,500 soldiers (defense 3.655M, half the threat): the garrison share is 15%. Brace 1 alone would destroy 22 buildings and keep 30% of stock; with the garrison, 19 fall (22 × 15% = 3.3, rounded down to 3 saved) and about 40% of stock is kept.

---

## Military Milestones

The five military milestones form a chain. Completing all five grants a title, and between them they give **+0.25 military power** and **+25% all production** (which adds into [the all-production pool](resources.md#the-all-production-cap)). Soldier counts in these milestones are soldiers trained over the run, not the number in stock; missions don't count toward any of them.

| Key | Name | Requirement | Min Age | Reward |
|-----|------|-------------|----------|--------|
| `first_soldiers` | First Soldiers | train 5 soldiers | Iron Age | +0.05 military power |
| `war_machine` | War Machine | train 250 soldiers | Iron Age | +0.10 military power |
| `iron_legion` | Iron Legion | train 500 soldiers, build 10 Barracks | Classical Age | +5% all production |
| `fortress_state` | Fortress State | build 13 Castle Keeps | Medieval Age | +0.10 military power, +5% all production |
| `military_superpower` | Military Superpower | train 2,000 soldiers | Industrial Age | +15% all production |

`iron_legion`, `fortress_state` and `military_superpower` are **hidden** until you are more than halfway to them or reach the age before their minimum age, so expect them to appear mid-game.

> **Note:** `standing_army` (train 100 soldiers and build 10 Barracks, Classical Age, +0.05 military power) is a separate military milestone, not part of the chain.

---

## Strategy

### Early game (Stone Age to Classical Age)

- Use short offers when you want a quick, lower-cost mission. Grand military offers use your army, while civil grand offers need materials.
- War Camps and Barracks train nothing before the Iron Age, when soldiers unlock. Build Barracks for `iron_legion` (10 Barracks), not for early soldiers.
- Once soldiers unlock, staff your military buildings (the Hunting Lodge costs only wood) and train 5 soldiers for `first_soldiers` and its +0.05 military power.
- `campaign trade_escort` (3 soldiers, Iron Age, 156-260 ticks, 0.30 difficulty) is the easiest early campaign: keep it running next to a scouting party for the encounters.
- Workers in military buildings eat the same food as everyone else, but every worker you move there is one fewer farmer. Keep your food rate positive.

### Mid game (Classical to Industrial Age)

- Go for `iron_legion` (train 500 soldiers, build 10 Barracks) for its +5% all production.
- Pick the easiest campaign you can afford rather than the one with the best loot. With no military power, `conquer_territory` (0.60) fails more often than it wins, and each failure that meets someone brings a setback.
- The Geographic Society sends the offer with the best haul per minute that you can afford.
- The age gates ask for military buildings (15 Hunting Lodges, 15 Military Academies, 5 Castle Keeps), and staffed they build your garrison too.
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

Morale multiplies the output of every worker in your civilization, on a continuous curve centered on 50%. Below 50% it cuts production, down to ×0.50 at the 10% floor. Above 50% it raises production, up to +20% at the top. An oversized army that drags morale down therefore slows every domain at once (food, research, trade and the rest), and the army's food gets harder to cover as your farmers produce less. A lean army leaves room to push morale above 50% instead.

**Keep military workers at 25-28% of your population or less.** That leaves a buffer below 30% if you lose workers to an event. If you need a large soldier stockpile, staff your military buildings heavily to bank soldiers, then move the extra workers back to civilian buildings. The stored soldiers stay, and the morale drain stops.

Morale drifts back toward 50% by itself, a little each tick, once you're back under 30%. If morale is falling and more than 30% of your workers are in military buildings, unassign some of them or recruit more workers for other buildings.

See [Morale](morale.md) for the full system.

---

## Tips & Common Mistakes

**Don't recruit past your food income.** A food deficit stalls all production, because starving workers can't work. Work out the extra food before `recruit max`.

**Don't wage hard campaigns without military power.** `campaign siege_castle` (0.70) and `campaign world_domination` (0.80) fail often with no bonus, and a failure meets fewer civilizations and brings setbacks home. Research a few military techs first.

**A failed mission still costs its launch.** The soldiers and any resources are spent when you launch. Failure doesn't refund them; it only cuts the reward to 30%.

**Workers need a built building.** You can't assign workers to a building you haven't built yet.

**Plan Castle Keeps early for `fortress_state`.** It needs 13 Castle Keeps (Medieval Age), a large stone and iron investment, so start building them as soon as you reach the Medieval Age. The +0.10 military power and +5% all production are worth the cost.

**Keep missions running.** There's no cooldown beyond the mission's own duration (100 to 160 ticks in the Primitive and Stone Ages, 156 to 416 from the Bronze Age on, rolled at launch). When one resolves, send the next.

**Prestige resets the army.** Soldiers, military buildings and the military power from techs and milestones all start over with the run. From the first prestige, [Old Friends](prestige.md#old-friends) meets civilizations found by missions again as soon as you reach their age.
