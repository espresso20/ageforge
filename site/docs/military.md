# Army & Missions

Soldiers come with your military buildings, and every mission takes a party of them out of the army at home. Short, long and civil grand missions also cost materials, and every haul scales with your income in the current age. A resolved mission can meet another civilization: that is how you make first contact and collect boons (or setbacks, when the mission fails). The soldiers at their posts are your garrison, and they blunt raids and soften an Endure. If you ignore the army, you still carry the small garrison the age gates build for you. If you invest in it, raids and catastrophes take less from you. The soldiers who arrive with new military buildings also count toward the military milestones.

---

## Overview

The military system covers four things:

- The mission board offers short, long and grand missions, each with its own length, odds and haul. See [Missions](#missions).
- The defense rating is a figure on the Army panel, worked out from the soldiers at home and your military power. Measured against the raid threat of your current age, it blunts part of what raids, war raids and an Endure take from you, by at most 45%. See [Defense](#defense-what-your-army-blunts).
- Five military milestones form a chain with a title at the end. The early ones grant permanent military power, and the later ones grant all production.
- The army resets with the run. Prestige points give a lasting production bonus, not military power or mission rewards. From the first prestige, [Old Friends](prestige.md#old-friends) meets again every civilization you met on earlier runs, without a mission.

Soldiers come with your military buildings. A military building does not train soldiers and has no soldier store: each copy has posts, and soldiers take them from the reserve first and then as new arrivals, as the food allows. The soldiers are people, so they live in your homes and eat like anyone (see [People](workers-and-domains.md#the-people-of-your-town)). A War Camp can be built in the Stone Age, and its soldiers move in as the food allows. A town with no military building has no soldiers. From the Stone Age on, every mission takes a party of them, so a town with no army cannot send one.

The mission board has three tiers, and a grand offer is either military or civil. See [The board](#the-board) for the costs and odds.

---

## Soldiers

### What soldiers are

Soldiers are people at the posts of your military buildings. A copy of a military building comes with a fixed number of them (see [Military buildings](#military-buildings) for the number in each age), and they leave with the copy if it is lost. They take their posts from the reserve first and then as new arrivals, whether or not your homes have room. The People box and the People panel's Army tab list them under the age's own word for them. They are not in the Resources box.

The army is the soldiers at their posts, and that is the number the rest of this page calls your soldiers. The defense rating, raids and the Brace preview count the soldiers at home, and a mission's party is taken from them.

Posts fill only while the food margin holds: net food must stay at or above one person's food, or 5% of the food made, whichever is more, after the newcomer eats. Soldiers arriving never starve the town. A post that food cannot fill waits and fills as the food allows. The Army tab shows posted over posts, and the log says once for each building that its posts are waiting. Nothing trains soldiers and nothing stores them, and a military building has no worker slots.

Two technologies change the posts. The Military-Industrial Complex and Augmented Soldiers each give every military building a tenth more posts (`posts a copy = ceil(base x (1 + 0.1 x techs))`), and the new posts fill as the food allows.

### Food drain

Every soldier eats the same amount of food as a worker (see [Food drain](workers-and-domains.md#food-drain)). A big army costs food because it is a lot of people, so check your food rate after you build a military building. Check your homes too: soldiers take a place in them like everyone else, and a town whose army fills the homes stops growing until you build more.

### Parties and losses

Every mission takes a party of soldiers. The party is a share of the army an ordinary town holds in the age you send it, so it is the same for every player in that age, whatever the size of your own army. A short offer takes 4% of that army, a long one 10%, a military grand offer 30% and a civil grand offer 12%, rounded up and at least one. The Primitive Age has no army and no military building, so its missions take no party.

| Age | Army | Short | Long | Grand, military | Grand, civil |
|---|---:|---:|---:|---:|---:|
| Primitive | 0 | 0 | 0 | 0 | 0 |
| Stone | 72 | 3 | 8 | 22 | 9 |
| Bronze | 200 | 8 | 20 | 60 | 24 |
| Iron | 430 | 18 | 43 | 129 | 52 |
| Classical | 790 | 32 | 79 | 237 | 95 |
| Medieval | 1,406 | 57 | 141 | 422 | 169 |
| Renaissance | 2,566 | 103 | 257 | 770 | 308 |
| Colonial | 4,798 | 192 | 480 | 1,440 | 576 |
| Industrial | 15,094 | 604 | 1,510 | 4,529 | 1,812 |
| Victorian | 29,694 | 1,188 | 2,970 | 8,909 | 3,564 |
| Electric | 52,894 | 2,116 | 5,290 | 15,869 | 6,348 |
| Atomic | 93,304 | 3,733 | 9,331 | 27,992 | 11,197 |
| Modern | 262,728 | 10,510 | 26,273 | 78,819 | 31,528 |
| Information | 500,968 | 20,039 | 50,097 | 150,291 | 60,117 |
| Digital | 876,832 | 35,074 | 87,684 | 263,050 | 105,220 |
| Cyberpunk | 2,112,856 | 84,515 | 211,286 | 633,857 | 253,543 |
| Fusion | 3,899,386 | 155,976 | 389,939 | 1,169,816 | 467,927 |
| Space | 7,750,298 | 310,012 | 775,030 | 2,325,090 | 930,036 |
| Interstellar | 19,858,746 | 794,350 | 1,985,875 | 5,957,624 | 2,383,050 |
| Galactic | 48,482,266 | 1,939,291 | 4,848,227 | 14,544,680 | 5,817,872 |
| Quantum | 85,913,821 | 3,436,553 | 8,591,383 | 25,774,147 | 10,309,659 |
| Transcendent | 140,961,453 | 5,638,459 | 14,096,146 | 42,288,436 | 16,915,375 |

The party leaves its posts while it is out. Soldiers away are not at home: the defense rating, raids and every other launch count only the soldiers at home, and the Ready Room header shows the soldiers at home out of all those posted. They are still people, so they still take a place in your homes and eat. Nothing but the army limits how many missions are out at once. You can send a second short mission, or a third, while the soldiers at home still cover another party. An offer you cannot staff says so, naming the party and the soldiers at home.

Soldiers fall in four ways:

- On a failure, between 40% and 70% of the party that came home (60% to 90% on a Legendary mission), rounded up.
- On a success, three times in ten, up to 15% of the party, rounded up.
- On the road, an ambush takes a tenth of the party of any mission, at least one.
- At home, when a raid lands and your garrison blunts it, a twentieth of the soldiers at home fall, at least one. A raid that finds nobody at home costs no soldiers.

The returns log and the mission's message say how many fell, for example "36 of 60 soldiers fell", or that all of them came home. The fallen leave the population and their posts stay empty. Empty posts fill again from the reserve, a few at a time (1% of all your posts in each staffing pass, at least one), while the food margin holds. Nobody arrives from outside to take them. The Army tab shows posted over posts while they train. The same rules run while you are away, so a mission that comes home while you are gone loses its soldiers, and the posts retrain just as they would at the screen.

### Viewing your army

```
army
```

This opens the **Army** panel, full screen like every panel. It shows:

- Soldiers: the soldiers at home over all your posts.
- Posts: the soldiers at their posts building by building. Before your first military building it reads "none yet. A military building comes with its soldiers."
- Defense: your defense rating.
- Threat: the raid threat of your current age.
- "Your garrison would blunt about N% of a raid.": the share it takes off raids and war raids in your current age (an Endure measures it against the age the catastrophe strikes in). After it comes what twice the garrison would blunt, and a reminder that no army blunts more than 45%. With no soldiers it reads "You have no garrison: raids hit you with full force."
- Saved this run: once the garrison has saved something, the buildings, workers and resources (the four largest) it kept, and how many raids it blunted. Like other run stats, it resets with the run.
- Military Bonus and Expedition Bonus lines, once either is above zero.
- The running military grand mission (if any) and how many missions you have completed.
- The military grand offers on the board, each with its length, party, difficulty and cost, or a message saying none are open yet.

The other missions show in the separate **Expeditions** panel (`expedition`), which also displays the whole board and the Geographic Society.

---

## Military buildings

Each copy of a military building comes with its soldiers (see Soldiers above). They have no worker slots and make and store nothing else.

All military buildings belong to the military lineage (22 tiers, Stone Age through Transcendent Age). Each extra copy of a building costs 15% more than the one before.

The soldiers a copy comes with come from the town an age is priced for. The army that town holds when it opens the gate out of the age is about 14% of its homes, so a copy comes with

```
soldiers a copy = ceil(14% of the homes at the end of the age ÷ copies of the age's military buildings that town holds)
```

The town holds eight copies of each of the age's military buildings, or the gate's count where it asks for more (a building the gate leaves out is in the town all the same). Its homes are eight copies of every housing building so far, or the gate's count where it asks for more.

| Tier | Key | Name | Age unlocked | Soldiers a copy |
|------|-----|------|-------------|-----------------|
| 0 | `war_camp` | War Camp | Stone Age | 9 |
| 1 | `barracks` | Barracks | Bronze Age | 16 |
| 2 | `hunting_lodge` | Hunting Lodge | Iron Age | 10 |
| 3 | `legion_fort` | Legion Fort | Iron Age | 10 |
| 4 | `military_academy` | Military Academy | Classical Age | 24 |
| 5 | `castle_keep` | Castle Keep | Medieval Age | 77 |
| 6 | `fortress` | Fortress | Renaissance Age | 145 |
| 7 | `fort` | Fort | Colonial Age | 279 |
| 8 | `military_base` | Military Base | Industrial Age | 1,287 |
| 9 | `garrison` | Garrison | Victorian Age | 1,825 |
| 10 | `command_post` | Command Post | Electric Age | 2,900 |
| 11 | `bunker_complex` | Bunker Complex | Atomic Age | 2,245 |
| 12 | `special_ops_hq` | Special Ops HQ | Modern Age | 21,178 |
| 13 | `cyber_command` | Cyber Command | Information Age | 29,780 |
| 14 | `drone_warfare_center` | Drone Warfare Center | Digital Age | 46,983 |
| 15 | `combat_aug_center` | Combat Aug Center | Cyberpunk Age | 154,503 |
| 16 | `plasma_command` | Plasma Command | Fusion Age | 178,653 |
| 17 | `space_force_base` | Space Force Base | Space Age | 481,364 |
| 18 | `fleet_command` | Fleet Command | Interstellar Age | 1,513,556 |
| 19 | `stellar_armada_hq` | Stellar Armada HQ | Galactic Age | 3,577,940 |
| 20 | `probability_war_room` | Probability War Room | Quantum Age | 2,495,437 |
| 21 | `omniversal_war_council` | Omniversal War Council | Transcendent Age | 6,880,954 |

The soldiers count per building: two Barracks bring twice the soldiers of one.

Five of them wait for a tech of their own age: the Barracks (Military Tactics), the Legion Fort (Siege Warfare), the Cyber Command (Cybersecurity), the Combat Aug Center (Cybernetics) and the Omniversal War Council (Omniversal Command). The others open with their age.

---

## Missions

A mission sends a party of soldiers out for a while and brings back a haul. Missions come from the mission board, and you send them all with the same `expedition` command. The game groups short, long and civil grand missions as scouting and military grand missions as campaigns. The two groups use different tech bonuses and meet civilizations at different odds (see [What missions are worth](#what-missions-are-worth)).

### The board

The mission board holds three offers of each tier that is open, so it holds three, six or nine offers. Each offer has a title and description made from the current age's words. The board is saved with the run and uses its seeded random generator.

Tiers open by era. No tech locks a tier:

| Era | Ages | Tiers on the board |
|---|---|---|
| Stone Era | Primitive, Stone, Bronze | Short (3 offers) |
| Iron Era | Iron, Classical, Medieval | Short and long (6 offers) |
| Steel Era and every later era | Renaissance and on | Short, long and grand (9 offers) |

A Veteran army or better sees the next tier an era early: long offers in the Stone Era, grand offers in the Iron Era (see [Army experience](#army-experience)). A tier that opens brings three new offers. One that closes again takes its offers with it, and the other tiers keep theirs.

| Tier | Length | Base odds | Haul, in minutes of income for each minute out |
|---|---|---:|---:|
| Short | 5 to 30 minutes | 85%, plus or minus 10 points | 2 |
| Long | 30 to 60 minutes | 75%, plus or minus 10 points | 3 |
| Grand, military | 60 to 120 minutes | 65%, plus or minus 10 points | 4 |
| Grand, civil | 60 to 120 minutes | 60%, plus or minus 10 points | 4 |

Each offer rolls its length, its odds and its haul when it appears. The haul comes out between 0.6 and 1.6 times the rate in the table. The lengths are in minutes at the base game speed, and game speed shortens the wait in real time.

The board shows each offer's rolled length, odds, haul and cost. The haul is in minutes of your income, paid in two of the age's materials (the ones your town earns the most of for their price), and it grows with the army's rank. A success also adds your expedition reward bonuses. Failure pays 30% of the haul.

A Legendary offer is half again as long, has odds 15 points lower (never under 10%) and pays four times the ordinary haul. About 5% of new offers are Legendary. At least 10 ordinary offers come between two Legendary ones, and after 39 ordinary offers in a row the next one is Legendary.

Every offer takes its party of soldiers (see [Parties and losses](#parties-and-losses)). A military grand offer costs nothing else, and its odds count your army. A short, long or civil grand offer also costs materials. They are the two materials the current age's buildings charge most, worth 5 minutes of your income for a short offer, 15 for a long one and 45 for a civil grand one, half of it in each material. A material you make none of is left out. The offer shows its cost before you send it.

The age's own words name each tier:

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

### Army experience

Missions that come home give your army experience, and a failure takes some back. The experience is saved with the run and starts again at 0 when you prestige.

| Mission | Success | Failure |
|---|---:|---:|
| Short | +1 | -0.5 |
| Long | +3 | -1.5 |
| Grand | +8 | -4 |

Experience never falls below 0. The army's rank follows from it, and every rank above Green adds to what the army does:

| Rank | Experience | Odds | Haul | Legendary offers |
|---|---:|---:|---:|---:|
| Green | 0 | no change | no change | no change |
| Seasoned | 10 | +2 points | +5% | +25% |
| Veteran | 30 | +4 points | +10% | +50% |
| Elite | 80 | +6 points | +15% | +75% |
| Legendary | 200 | +8 points | +20% | +100% |

The odds bonus applies to every offer you send, up to the 95% ceiling. The haul bonus applies to what a mission brings home. The Legendary bonus raises the chance that a new offer is Legendary (the usual 5% becomes 6.25% for a Seasoned army, and so on), and the spacing rules for Legendary offers stay. From Veteran, the next tier of the board opens an era early.

The log says `Your army is now <Rank>.` when the army rises a rank. The top border of the Expeditions panel shows the rank and the experience toward the next one (for example `Army: Veteran · 42/80`). The People panel's Army tab shows the rank, the experience and what the rank gives. Each line of the Expedition Report says how much experience the mission gave or took.

Sending an offer takes it off the board. A replacement arrives after five minutes. An untaken offer expires after an hour and is replaced at once, and entering a new age creates a new board. The panel and `expedition list` show why an offer cannot be sent now. A taken slot shows when its replacement arrives.

### Road events

A party can meet road events while it is out. The events are rolled when the mission launches and fire at set points along the trip. They appear in the log and in the Expedition Report. Longer missions meet more of them: a short mission about one in five, a long one about one, a grand one about two.

| Event | What it does |
|---|---|
| Storm | Delays the party by 15% of the trip's length. |
| Helpers | Save a quarter of the time left. |
| Find | Raises the haul by 40% (80% on a Legendary mission). |
| Toll | Costs two minutes of your gold income. If you cannot pay, it becomes a foe encounter. |
| Foes | Lower the odds by 8 points, to no less than 5%. |
| Fair weather | Raises the odds by 6 points, to no more than 95%. |
| Ambush | A tenth of the party falls, at least one soldier. A short, long or civil grand mission also loses 15% of its haul. A mission with no party (the Primitive Age) loses the 15% of the haul alone. |
| Ruins | Puts a new short offer on the board, in a free place, or in place of the first offer when the board is full. |

### The Expeditions panel

<figure class="screen" data-screen="expeditions"><figcaption>The Expeditions panel: the Ready Room, a map of the age's lands with every offer as a pin, and the briefing for the one you pick.</figcaption></figure>

`expedition` opens the **Expeditions** panel. It fills the screen except for the command bar and needs at least 60 columns and 14 rows. It has two views, and `Tab` (or `Shift+Tab`) switches between them. The army's rank is on the right of the top border, and the soldiers at home are on the right of the tab row, for example `Soldiers 72 of 72 at home`.

The **Ready Room** is where you read the map and send a party. The map shows the age's lands with your town in the middle and every offer on the board as a pin:

| Pin | Offer |
|---|---|
| ○ | Short |
| ◇ | Long |
| ▲ | Grand, military |
| ■ | Grand, civil |
| ◈ | Legendary, in gold |

Short offers lie near the town, long ones in the middle distance and grand ones far out. The land is drawn from your run and the age, so it holds still while the age lasts and changes with the next. A mission under way shows a dotted route from the town, an `@` where its party is and a ◆ at its goal.

The selected pin has a dotted route from the town and its place name beside it. Next to the map, the briefing for that offer gives its tier, its twist, its party, its odds, how long it is out and back, how long the offer stays up, its haul in minutes of income and in resources, and its cost. The party and the cost are green where you can pay them and red where you cannot. The card ends with `Enter send the party`, or with the reason the offer cannot be sent now. Under the map are the missions out now, a key to the pins in the age's own words, and the time until the board next changes.

The arrow keys hop to the nearest pin in that direction. `Enter` opens a box that asks before the party goes, and `Enter` again sends it. On an offer you cannot send, `Enter` says why on the bottom line. `Esc` takes the question back, and closes the panel when there is no question.

The **Expedition Report** follows the parties out. Its map is titled Parties in the field and shows every mission under way. On a wide terminal a list at the right has one line for each party (its place, a progress bar and the time left), with a field report under it for the selected party: its tier, its soldiers, its odds (with the launch odds when a road event has moved them), its progress, its expected haul and every road event so far, newest first. On a narrow terminal a card lists each party with its progress, soldiers, odds and haul in minutes, and the road events of the selected one. `Up` and `Down` pick a party, and its route lights up on the map. The **Came home** box lists the last ten missions to return, newest first, with a ● on the ones that came home since you last looked. Each line says whether the mission succeeded, what it brought back, how many road events it met, how much experience it gave or took, and how many soldiers fell. The Geographic Society's line is on the Report too. With nobody out, the Report says so and how many offers are waiting in the Ready Room.

The panel opens on the Report while a party is out or one has come home since you last looked at it, and on the Ready Room otherwise. Looking at the Report clears the `● new` count beside its tab. The prompt keeps working while the panel is open. With something typed, the arrow keys, `Tab` and `Enter` act on the prompt instead.

### Commands

```
expedition              # open the Expeditions panel (shorthand: exp)
expedition list         # print the mission board
expedition <name>       # send the offer with that name
campaign                # alias for expedition
campaign list           # alias for expedition list
campaign <name>         # alias for expedition <name>
```

Offers have no numbers. To send one from the prompt, type part of its title after `expedition`. Case does not matter, and "the", "of", "a" and "an" are left out. An exact title wins. Otherwise a title that starts with what you typed wins. Otherwise any title in which every word you typed begins one of its words matches. If one offer matches, it is sent. If several do, the game lists them and asks you to type more of the name. If none does, it says so. Offers that share one title count as one match, and the first is sent.

`expedition list` prints each offer with a ✓ if you can send it now or a ✗ and the reason if you cannot. A Legendary offer has a star before its title.

A mission's resource cost is spent at launch, whether it succeeds or fails. Every mission needs its party at home and takes it out of the posts until it returns.

### Automatic dispatch: the Geographic Society

From the **Industrial Age** the **Geographic Society** sends short offers on its own for the rest of the run. It waits for the **Geographic Societies** tech of that age, which needs Cartography.

It only ever sends short offers. Long and grand missions stay yours to order.

It sends a party only when the soldiers at home can make one up, as a hand-sent mission does. Nothing is queued.

It pays full price. Each automatic mission costs the same materials you would pay. If you do not hold the materials to outfit one, or the soldiers at home cannot make up the party, the Society holds it back, tells you once, and sends it as soon as it can. It never runs up a debt.

Among the short offers it can send, it takes the one with the best haul per minute of length. Ties go to the bigger haul, then to the offer that comes first on the board. The log says `Geographic Society dispatches a survey party: <title>.` when it sends one.

Its missions are ordinary missions. They use the same odds, pay the same haul and roll civilization encounters, boons and setbacks exactly like the ones you send.

The more you invest, the faster it goes. The wait between parties is

```
wait = 900 x (1 - 0.35 x staffed share) ÷ Societies, never under 300, then times 2.6
```

The staffed share is the workers assigned to your Societies over their 8 slots each. The 2.6 is the stretch every age from the Bronze Age on has for its timers. A single unstaffed Society sends about one party every 2,340 ticks (about 1h 18m). Staffing it fully cuts that to 1,521 ticks, and each extra Society shortens the wait again. The floor is one party every 780 ticks (about 26m), reached with two fully staffed Societies or three unstaffed ones. A new Society or new staff shortens a wait already running at once. Societies are staffed by military-domain workers, so they count toward the [military share of your workers](#morale-and-military-ratio).

The **Expedition Report** (the second view of the `expedition` panel) shows one line for it: the time until the next party leaves, or a warning that a party is due and you cannot pay for it. The full status (Societies built, staffing, the wait between parties, and a countdown with a progress bar) is on the **Factions** panel (`factions`); see [The Factions panel](factions.md#the-factions-panel). Both show approximate wall-clock time rather than ticks.

The Society is slower than sending parties yourself, on purpose. It keeps missions moving while you are away, and sending offers by hand gives you control over the odds, haul and cost.

### Success chance

Each offer rolls its base odds when it appears (see [The board](#the-board)). When you send it, three things move the odds:

- Your army's rank adds its points to every offer (see [Army experience](#army-experience)).
- A military grand offer also counts your military power against the typical player's of that age (see [The military yardstick](#the-military-yardstick)). The same as the typical player's changes nothing, twice as much adds 10 points, and none at all takes 10 away. The change stops at 15 points either way, and it starts in the Bronze Age. A civil grand offer, and a short or long one, ignore military power.
- Road events move the odds of a mission under way (see [Road events](#road-events)).

The odds you send with stay between 10% and 95%. The panel shows the odds that apply now. The army's military power and its effects on raids are covered in [Defense](#defense-what-your-army-blunts).

### What missions are worth

Hauls scale with income. The board shows the expected resource haul before launch. Longer missions and Legendary offers can pay more, while failure pays 30% of the haul.

Techs can change a mission, and they count for missions sent after you learn them:

- Cartography, Aviation, Reusable Launchers and Fusion Drives each cut the time of short, long and civil grand missions by 10%, Wormholes by 15% and Stellar Cartography by 20%. The cuts multiply, and none takes a time under 40% of its length.
- General Staff and Special Forces each cut the time of military grand missions by 15%, on the same terms.
- The Military-Industrial Complex makes military grand missions bring back 20% more, won or lost, and Probability Warfare 15% more again.
- The expedition reward bonus (from techs, milestones and the Grand Lighthouse wonder) multiplies the loot on a success: `rewards × (1 + expedition reward bonus)`.

Every resolved mission can meet a civilization. A short, long or civil grand mission meets one 36% of the time when it succeeds and 12% of the time when it fails. A military grand mission meets one 16% of the time when it succeeds and 4% when it fails. An encounter is how you make first contact, and it usually brings a boon: a timed production boost, a lump of resources or a crew of workers. It brings a setback instead when the mission failed, and sometimes when the civilization is at war with you. The full rules (what the boons and setbacks are, the five-boon limit, war) are on [Factions & Diplomacy](factions.md#encounters-boons-and-setbacks).

Compare an offer's odds, haul and party before sending it.

---

## Military domain workers

Military buildings have no worker slots. The one building with military worker slots is the Geographic Society (see [Automatic dispatch](#automatic-dispatch-the-geographic-society)). Workers in it take a class name that changes with the age:

| Age | Class name |
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
assign geographic_society [count|all]
```

You can only assign workers to a building you have built at least one of. Staffing a Society shortens the wait between its parties. It changes no soldier count.

---

## Military power bonus

Military power is a running total that raises the odds of military grand offers (see [Success chance](#success-chance)) and feeds the defense rating:

```
defense = soldiers at home × 2.0 × (1 + military power)
```

The defense rating is what your garrison is measured by against raids and catastrophes (see [Defense](#defense-what-your-army-blunts)), so military power pays twice: it helps military grand offers, and it makes the same soldiers blunt more. A +1.0 bonus doubles the defense rating of the same army.

Sources add together:

| Source | How to get it | Bonus |
|--------|-------------|---------------|
| Techs | Eighteen military techs grant military power (see [Military power](technologies.md#military-power)) | +0.10 to +0.15 per tech, +2.01 with all eighteen |
| Milestones | Complete military milestones | +0.05 to +0.10 each |

Nothing caps military power. The army adds at most 15 points to a military grand offer's odds, and every offer's odds stay between 10% and 95%.

### The military yardstick

Each age has a yardstick: the military power a typical player holds there. The typical player holds the military techs of every age up to their own, plus First Soldiers and War Machine (+0.15 together, which the military buildings the age gates ask for earn on their own). The yardstick is used in two places:

- A military grand offer compares your military power with the typical player's of the age (see [Success chance](#success-chance)). More than the typical player's still helps, and less still costs.
- The threat scale adjusts the age's raid threat for the military power the tech tree gives. It is the Threat scale column in the table under [Defense](#defense-what-your-army-blunts).

The typical military power of every age is in the same tables. The yardstick is read off the tech tree, so it moves with it.

---

## Defense: what your army blunts

The soldiers at your posts are your garrison. Defending does not use them up, except that a raid your garrison blunts costs a twentieth of the soldiers at home. The garrison takes a share off what raids and catastrophes cost you.

### The formula

Your defense rating is measured against the raid threat of the age you are in. The threat is in the same units as the rating, and the Army panel shows both:

```
defense  = soldiers at home × 2.0 × (1 + military power)
threat   = the age's raid threat                       # the tables below
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

The threat follows the army of the town an age is priced for (see [Military buildings](#military-buildings)). It rises with the army the age's buildings bring, and dips once, in the Atomic Age, where the gate asks for more copies of a building with fewer posts each. A garrison you stop growing therefore blunts less as you advance.

| Age | Typical military power | Threat scale | Threat | Soldiers for 22.5% (no military bonus) |
|-----|------------------------|--------------|--------|----------------------------------------|
| Iron Age | 0.45 | 0.88 | 1.72K | 861 |
| Classical Age | 0.60 | 0.78 | 1.94K | 971 |
| Medieval Age | 0.70 | 0.83 | 2.81K | 1.41K |
| Renaissance Age | 0.82 | 0.71 | 3.29K | 1.64K |
| Colonial Age | 0.94 | 0.68 | 5K | 2.5K |
| Industrial Age | 1.06 | 0.61 | 6.39K | 3.2K |
| Victorian Age | 1.06 | 0.61 | 10.6K | 5.3K |
| Electric Age | 1.16 | 0.64 | 15.1K | 7.55K |
| Atomic Age | 1.36 | 0.40 | 13K | 6.52K |
| Modern Age | 1.36 | 0.40 | 33.3K | 16.6K |
| Information Age | 1.46 | 0.36 | 45.9K | 22.9K |
| Digital Age | 1.56 | 0.37 | 68.3K | 34.2K |
| Cyberpunk Age | 1.76 | 0.35 | 138K | 68.8K |
| Fusion Age | 1.86 | 0.36 | 234K | 117K |
| Space Age | 1.86 | 0.36 | 420K | 210K |
| Interstellar Age | 1.96 | 0.38 | 1.01M | 504K |
| Galactic Age | 2.06 | 0.39 | 2.42M | 1.21M |
| Quantum Age | 2.16 | 0.40 | 4.27M | 2.13M |
| Transcendent Age | 2.16 | 0.40 | 6.38M | 3.19M |

A `military_power` bonus cuts the soldiers needed: at +1.0, half as many.

### Before the Iron Age

The three early ages have raid threats of their own. Before the Iron Age the only military building a gate asks for is the War Camp (eight of them, to leave the Stone Age), so a garrison is the War Camps and Barracks you chose to build:

| Age | Typical military power | Threat | Soldiers for 22.5% (no military bonus) | Where soldiers come from |
|-----|------------------------|--------|----------------------------------------|--------------------------|
| Primitive Age | 0 | 33 | none | no military building yet |
| Stone Age | 0 | 160 | 80 | War Camp: 9 soldiers a copy, 72 with the eight the gate asks for |
| Bronze Age | 0.15 | 410 | 205 | Barracks (after Military Tactics): 16 soldiers a copy |

The reference town's army (72 soldiers in the Stone Age, 200 with its eight Barracks in the Bronze) blunts about a fifth of a raid, and half of it about 14%. A garrison carried into the next age meets that age's larger threat, so it has to keep growing, as it does later.

The raids of these ages are Tribal Raid and Beast Stampede (Stone Era) and Bandit Raid (Bronze Age on). See [Raids and your garrison](events.md#raids-and-your-garrison).

### The garrison you already have

Every player carries a garrison without trying. The age gates ask for military buildings: 8 War Camps for the Bronze Age, 15 Hunting Lodges for the Classical Age, 15 Military Academies for the Medieval, 8 Castle Keeps for the Renaissance, 8 of each age's own military building for the Colonial, Industrial, Victorian, Electric, Atomic, Information, Cyberpunk, Fusion, Interstellar, Galactic and Quantum Ages, 18 Bunker Complexes for the Modern, 10 Plasma Commands for the Space Age and 15 Probability War Rooms for the Transcendent. The Iron and Digital Age gates ask for none. See [Advancement Requirements](ages.md#advancement-requirements). You keep those buildings into every later age, and their soldiers stay with them. The soldiers a gate's worth of buildings brings is about 14% of the homes of the town that age is priced for, which is how many soldiers each building comes with (see [Military buildings](#military-buildings)).

The town an age is priced for holds eight copies of each military building so far, or the gate's count where it asks for more. With every post filled and the typical player's military power, its garrison blunts about 19% of a raid in the Iron Age, 33% in the Renaissance Age, 41% in the Industrial Age and 44% or more from the Atomic Age on. The gates ask for fewer buildings than that town holds, so the military buildings of every age, not only the ones the gates ask for, bring you closer to it. More copies and military power research push toward the 45% ceiling.

### What it blunts

1. **Raid events.** Random events that are attacks by outsiders: Bandit Raid, Pirate Attack, Data Breach, The Great Breach, Corporate Espionage, and the Stone Era's Tribal Raid and Beast Stampede. The garrison cuts the resources they steal and the workers they drive off by its share. A raid's production penalty is not blunted. See [Raids and your garrison](events.md#raids-and-your-garrison). The log adds a line under the event: "Your garrison blunted about N% of the raid: you kept ...".
2. **War raids.** While a civilization is at war with you, each raid it makes takes a resource. The garrison keeps its share of that resource. A raid bigger than your stock takes nothing, and the army never turns a raid that missed into one that lands. The raid's log line adds "Your garrison kept X gold from them (about N% of the raid)." (with the raid's resource in place of gold). See [War & Peace](factions.md#war-amp-peace).
3. **Endure.** When you Endure a catastrophe, the garrison blunts its share of the buildings destroyed and the stock lost, after the Harbinger's Brace, measured against the threat of your age when you Endure (see below).

Four techs make raids smaller before the garrison meets them. Imperial Legions, Nuclear Deterrence and Orbital Defense each take 10% off what a raid event or a war raid would steal, and Fortification 15% (38% with all four: the cuts multiply). The garrison then blunts its share of what is left. They do nothing for an Endure.

Nothing else. Disasters and unrest (earthquakes, plague, mine collapses, industrial accidents, uprisings) are not raids, and soldiers do nothing against them. Endure's 25% worker loss, its production debuff and its morale hit are not blunted.

A player with no soldiers takes the full losses. From the Iron Age on that is rare: the military buildings the age gates require give nearly everyone some garrison (see [The garrison you already have](#the-garrison-you-already-have)).

### Endure: Brace first, then the garrison

A catastrophe strikes at a fated moment anywhere in its era (Iron to Cosmic), and a pending catastrophe keeps you in the age it struck in until you choose. The garrison meets the threat of that age. A town that built no military building has no soldiers, so an Iron Era doom that strikes before it builds one finds no garrison to count.

1. **Brace** applies first: 20% / 15% / 10% of buildings fall and 15% / 30% / 45% of stock is kept at Brace 0 / 1 / 2.
2. **The garrison** then blunts its share of what is left: the braced share of buildings destroyed shrinks by that share, and it keeps that share of the stock Brace would have let go. Buildings saved round down, so the garrison never saves more than its share.
3. **The cap:** Brace and garrison together can cut the unbraced loss by at most 60%. At least 8% of buildings fall and at most 66% of stock is kept. The cap only bites at Brace level 2.

| Brace | Garrison share | Buildings destroyed | Stock kept |
|-------|----------------|---------------------|------------|
| none | 20% | 16% | 32% |
| none | near the 45% ceiling | about 11% | about 53% |
| 1 | 20% | 12% | 44% |
| 2 | 20% | 8% | 56% |
| 2 | 40% or more | 8% (cap) | 66% (cap) |

The catastrophe window shows the real numbers after Brace and garrison, with a line for each, and the Harbinger panel's Brace preview counts the garrison too. After an Endure the log says "Your garrison held the line: ...". See [Endure](catastrophe.md#endure) and [Brace](harbinger.md#brace-soften-an-endure).

The Last Passage in the Cosmic Era costs prestige points, not buildings or stock, and the garrison doesn't change it.

### Worked example

You hold 2,000 soldiers at home with no military power bonus: defense 2,000 × 2 = 4,000.

- In the Classical Age (threat 1.94K): 45% × 4,000 / (4,000 + 1,943) = about 30%. A war raid that would take 1,000 gold takes about 700. The Army panel says twice the garrison would blunt about 36%.
- In the Industrial Age (threat 6.39K): 45% × 4,000 / (4,000 + 6,394) = about 17%. The same army matters less four ages later.
- Endure at the Renaissance (threat 3.29K) with Brace 1, 150 buildings that can fall (everything but wonders, storage and knowledge buildings) and 822 soldiers (defense 1,644, half the threat): the garrison share is 15%. Brace 1 alone would destroy 22 buildings and keep 30% of stock. With the garrison, 19 fall (22 × 15% = 3.3, rounded down to 3 saved) and about 40% of stock is kept.

---

## Military milestones

The five military milestones form a chain. Completing all five grants the title "The Conquerors", and between them they give +0.25 military power and +25% all production (which adds into [the all-production pool](resources.md#the-all-production-pool)). Soldier counts in these milestones are the soldiers who have come with your military buildings over the run, not the number in stock. Missions don't count toward any of them.

| Key | Name | Requirement | Min age | Reward |
|-----|------|-------------|----------|--------|
| `first_soldiers` | First Soldiers | train 5 soldiers | Iron Age | +0.05 military power |
| `war_machine` | War Machine | train 250 soldiers | Iron Age | +0.10 military power |
| `iron_legion` | Iron Legion | train 500 soldiers, build 10 Barracks | Classical Age | +5% all production |
| `fortress_state` | Fortress State | build 13 Castle Keeps | Medieval Age | +0.10 military power, +5% all production |
| `military_superpower` | Military Superpower | train 2,000 soldiers | Industrial Age | +15% all production |

`iron_legion`, `fortress_state` and `military_superpower` are hidden. They show up once your overall progress toward them passes 50% (the average of their conditions), or as soon as you are in the age before their minimum age.

A sixth military milestone sits outside the chain: `standing_army` (train 100 soldiers and build 10 Barracks, Classical Age, +0.05 military power).

---

## Strategy

### Early game (Stone Age to Classical Age)

- A War Camp can be built in the Stone Age. Its 9 soldiers move in as the food allows, and the gate into the Bronze Age asks for eight of them. A Barracks needs Military Tactics. Build 10 Barracks for `iron_legion`.
- Your first military building already brings the 5 soldiers `first_soldiers` asks for. The milestone opens in the Iron Age and gives +0.05 military power. The Hunting Lodge costs only wood.
- Short offers cost the least: the smallest party and the least material. Grand military offers use only soldiers, while civil grand offers need materials as well.
- Soldiers eat the same food as everyone else and take places in your homes. Keep your food rate positive.

### Mid game (Classical to Industrial Age)

- Go for `iron_legion` (train 500 soldiers, build 10 Barracks) for its +5% all production.
- Compare an offer's odds, haul and party before you send it. A failure costs between 40% and 70% of the party that came home and some army experience, and a failed mission that meets a civilization brings a setback.
- The age gates ask for military buildings (15 Hunting Lodges, 15 Military Academies, 8 Castle Keeps), and their soldiers build your garrison too.
- Research military techs as they appear. Each one adds to your defense rating and to the odds of military grand offers.
- The Geographic Society (Industrial Age) sends the short offer with the best haul per minute that you can pay for and send.

### Late game (Modern Age onward)

- The parties of the late ages are large (see [Parties and losses](#parties-and-losses)). Keep enough soldiers at home to cover the missions you want out at once.
- Add the current age's military buildings as you go, since the threat rises with the age.
- The `military_superpower` milestone (train 2,000 soldiers) adds +15% all production.

### Defense rating

The Army panel shows your defense rating (soldiers at home × 2.0 × (1 + military power)). It is measured against the raid threat of your age, and the resulting share (at most 45%) comes off raid events, war raids from civilizations at war with you, and the buildings and stock an Endure takes. It never stops a raid or a catastrophe from happening, and it does nothing against disasters such as plague or earthquakes. A civilization at war with you raids every 104 ticks from the Bronze Age on (every 40 before it) and takes 50 × its strength (1 to 5) of its specialty resource. Your garrison keeps its share of that, but only staying out of wars, or ending them with tribute, stops the raids: see [War & Peace](factions.md#war-amp-peace). See [Defense](#defense-what-your-army-blunts) for the numbers.

- Keep up with the age. The threat rises with the army each age's buildings bring, so a garrison you stop growing fades. Adding the current age's military buildings is the way to keep pace, because most of them come with more soldiers per copy than the last age's.
- Mind the next age. The Harbinger panel's Brace preview uses the age you are in. If you advance before the doom strikes, the threat of the new age applies, and it is usually higher, so your garrison blunts less than the preview showed.
- Brace and garrison stack, up to a point. Together they cut an Endure's losses by at most 60%. With Brace 2, a garrison that blunts 20% already reaches the building cap (8% fall), so extra soldiers mostly buy stock kept, up to 66%.
- Parties out on missions are not at home. While they are away your defense rating is lower.
- Soldiers cost food and homes. Every soldier eats like a worker and takes a place in your homes.
- Research military power. It multiplies the defense rating of the soldiers you already have.

### Army size and food

Soldiers eat like any worker, so the food cost of an army is the food cost of its people. After you build a military building, check `rates` and make sure your food rate stays positive with the new soldiers added. After an age advance, check food again: the cost per worker goes up by ×1.12 each age.

### Morale and military ratio

A large force of military workers has a second cost: morale. If more than 30% of your workers are assigned to military-domain buildings, morale drops every tick, and the further over 30% you are, the faster it drops:

| Military share | Over 30% by | Morale lost per tick |
|---------------|---------|--------------|
| 30% | 0 | none |
| 40% | 10 points | 0.3 points |
| 50% | 20 points | 0.6 points |
| 60% | 30 points | 0.9 points |

Morale scales what your workers make, on a continuous curve centered on 50%. Below 50% it cuts production, down to ×0.50 at the 10% floor. Above 50% it raises production, up to +20% at the top. An oversized force that drags morale down therefore slows every domain at once (food, research, trade and the rest), and the army's food gets harder to cover as your farmers produce less.

Only the Geographic Society has military worker slots. Soldiers at their posts are not military workers and do not count toward this share, so the drain needs a very large number of Societies, all staffed. If morale is falling and more than 30% of your workers are in Societies, unassign some of them or recruit more workers for other buildings.

Morale also drifts back toward 50% by itself, a little each tick.

See [Morale](morale.md) for the full system.

---

## Tips and common mistakes

- Watch your food after you build a military building. A food deficit stalls all production, because starving workers can't work. Soldiers eat like workers.
- Check the odds before you send a military grand offer. A failure costs a large share of the party, some army experience and often a setback instead of a boon.
- A failed mission still costs its launch. Any resources are spent when you launch, and failure doesn't refund them. It only cuts the reward to 30%.
- Workers need a built building. You can't assign workers to a building you haven't built yet.
- Plan Castle Keeps early for `fortress_state`. It needs 13 Castle Keeps (Medieval Age), a large stone and iron investment, so start building them as soon as you reach the Medieval Age. The +0.10 military power and +5% all production are worth the cost.
- Keep missions running. There is no cooldown beyond the mission's own length (5 to 120 minutes at the base game speed, half as long again for a Legendary offer). When one resolves, send the next. A taken offer's replacement arrives after five minutes.
- Prestige resets the army. Soldiers, military buildings and the military power from techs and milestones all start over with the run. From the first prestige, [Old Friends](prestige.md#old-friends) meets civilizations found by missions again as soon as you reach their age.
