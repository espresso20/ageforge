# Factions & Diplomacy

The world beyond your borders holds an **11-civilization roster** of other powers. You meet them by sending missions. Each one holds an **opinion** of you, and opinion decides what it offers: trade deals, gifts after an encounter, an alliance that raises your production of its specialty, workers on loan. Provoke one while it hates you and it goes to war. Everything they are doing to you is on the **Factions** panel (`factions`).

The market, trade routes, harbors and the black market are on [Trade](trade.md).

---

## Meeting civilizations

Every civilization has a first age. Reaching it makes the civilization *eligible*; it doesn't come to you on its own. Sending missions is what finds it: whenever a **scouting expedition** or a **campaign** resolves, win or lose, the game rolls a chance to **encounter** a civilization (see [Army & Missions](military.md#missions)).

| Resolution | Encounter chance |
|---|---|
| Scouting success | ~18% |
| Scouting failure | ~6% |
| Campaign success | ~8% |
| Campaign failure | ~2% |

An encounter meets someone new if it can: a civilization your age makes eligible that you haven't met yet (first contact). Once you know everyone within reach, it meets one you already know again. Either way the pick leans toward stronger civilizations. Before the Bronze Age there is nobody out there to meet.

Missions aren't quick (the shortest scouting run takes 156 to 260 ticks from the Bronze Age on), so encounters are occasional. A player who keeps a scouting party and a campaign in the field all the time meets someone about every **975 ticks** (about 32 minutes). From the Industrial Age a fully invested **Geographic Society** keeps parties going while you're away, at about 60% of that pace: about one meeting every 1,660 ticks (about 55 minutes). See [Automatic dispatch](military.md#automatic-dispatch-the-geographic-society).

If you never send a mission, each civilization is discovered anyway **two ages after** its first age, far later than an explorer would meet it.

On first contact the log introduces the civilization: its name, personality and backstory. Until then it doesn't exist as far as the game is concerned. No panel, log line or refusal names a civilization you haven't met, its age, specialty or personality. `diplomacy` and `plan deal` refuse an unmet civilization the same way they refuse a name that doesn't exist, and completion only offers the civilizations you know.

**Old friends after a prestige.** Prestige and Succumb end every relationship: the next run starts having met no one. The game remembers every civilization you have met, in any run, and with **Old Friends** from the prestige [legacy kit](prestige.md#old-friends) (`prestige buy legacy_factions`, 54 points) each of them is met again as soon as your age reaches its own, with no mission and no two-age wait, at neutral opinion. The log says so: "Old friends: the Riverlands Tribes remember your people and make contact again." Civilizations you have never met still have to be found by the rules above.

---

## The Factions panel

`factions` opens the **Factions** panel, one screen for everything the other civilizations are doing to you. `diplomacy` and `dip` with no arguments open the same panel, and `diplomacy <action> <civ>` performs an action directly. From top to bottom it shows:

1. The title, with how many civilizations you have met and how many are still undiscovered.
2. **Boons and setbacks**: every timed effect a civilization has on you right now. Boons are marked `✦` and setbacks `⚠`. Each line names the civilization, the effect, its size (`+13% food`, `+8% all prod`, `+9% tick speed`) and the time left. The header shows how full both pools are (`boons 2/5 · setbacks 1/3`), because both are hard limits (see [Limits](#limits)). Workers on loan from a peaceful civilization are listed here too, marked temporary or permanent. So is each crew of temporary workers a boon lent you (*↳ Merchant Guild: crew on loan · 5 workers · ~1h 47m*): who sent it, how many and the time left. Crews take no boon slot, so the header doesn't count them.
3. The **Geographic Society**, in one of three states: nothing built (a prompt to build one, which names its age once you can see it); starved (a party is due but your stores can't pay for it); or running, with the number of Societies, their staffing, the dispatch interval, and a countdown with a progress bar to the next party.
4. A card for each civilization you have met: name, personality, specialty and a **strength rating** of 1 to 5 stars; a line of backstory; a war banner if you are at war; the opinion bar; the status (color-coded) with any active ally bonus and how many trades you've done; how far it is to the next threshold (*+8 opinion to friendly*, *+12 opinion to ally-eligible*, or the ally command once you qualify); workers it has lent you; how many boons and setbacks it is applying right now; and its current **trade deals**, one numbered line per offer (see [Trade deals](#trade-deals)). A civilization that won't trade says why instead.
5. **Not yet met**: how many civilizations you haven't discovered, and nothing else (`5 civilizations not yet discovered. Send expeditions to find them.`).

Every duration on the panel shows as approximate wall-clock time, not ticks.

---

## Commands

| Command | What it does |
|---|---|
| `factions` | Open the Factions panel |
| `diplomacy` (or `dip`) | Open the same panel |
| `diplomacy ally <civ>` | Ally with a civilization. Needs opinion 50 or more and costs 500 gold |
| `diplomacy rival <civ>` | Declare a civilization your rival. Free. Opinion drains, and it offers no deals |
| `diplomacy embargo <civ>` | Embargo a civilization. Free. Opinion drains, no deals, your routes that bring in its specialty stop, and it counts as a provocation |
| `diplomacy gift <civ>` | Send a gift: 200 gold for +15 opinion |
| `diplomacy neutral <civ>` | Return to neutral. Free; an alliance's 500 gold isn't refunded |
| `diplomacy tribute <civ>` | End a war at once: 300 gold and 50 culture per point of the civilization's strength |
| `diplomacy raid <civ>` | Raid its trade route: -20 opinion, and a provocation |
| `diplomacy deals [civ]` | List trade deals (one civilization, or every one you have met) |
| `diplomacy accept <civ> <n>` | Take deal `n` |
| `plan deal <civ> <n>` | Take deal `n` in the [build plan](plan.md), once you can pay for it |

`dip` works in place of `diplomacy` everywhere. Civilizations are named by key (`riverlands_tribes`), and only the ones you have met are accepted.

---

## Opinion and status

Each civilization you have met holds an **opinion** of you from -100 to +100, starting at 0, and a **status** that you set (or that a gift or deal raises to friendly).

| Status | Meaning |
|---|---|
| `neutral` | The default. No bonuses, no penalties. |
| `friendly` | Set when a gift or deal brings a neutral civilization's opinion to 25 or more. Friendly civilizations offer more deals at better rates, give bigger boons and gentler setbacks. |
| `allied` | Needs opinion 50+ and costs 500 gold. Adds the civilization's bonus to your whole production of its specialty and to route income of it (see [Allied Bonuses](#allied-bonuses)), and gives the best deals and boons. |
| `rival` | Free to declare. Opinion drops an extra 5 every 130 ticks. No deals, no ally bonus, smaller boons, harsher setbacks. |
| `embargo` | Free to declare. The same as rival, and it counts as a provocation. Your routes that bring in its specialty are disrupted. |

### What moves opinion

From the Bronze Age on (which is when the first civilization becomes eligible) every clock below runs at the same stretched length.

| Source | Change | Notes |
|---|---|---|
| Trade routes | +1 per completed route cycle | With every civilization you have met that isn't at war with you, all at once. The biggest source for most players: see [What routes are worth](trade.md#what-routes-are-worth). |
| Embassies | a steady trickle | Spread across your non-hostile civilizations (see [Embassy Buildings](#embassy-buildings)). |
| Gifts | +15 per gift | `diplomacy gift <civ>` costs 200 gold. |
| Deals | +1 per Buy, Sell or Rare deal; +5 per Goodwill deal | Only up to opinion 50. |
| Personality | ±1 every 65 ticks | See [Personalities](#personalities). |
| Natural drift | 1 point toward 0 every 260 ticks | Not while at war. For an ally this is a slow leak. |
| Rival or embargo | -5 every 130 ticks | A civilization at -100 stays there. |
| Raiding its route | -20 at once | `diplomacy raid`, also a provocation. |
| Tribute | +25, but never above 0 | Ends a war. |

130 ticks is about 4 minutes, 260 about 9.

### Personalities

Every civilization has one of four personalities, which sets how its opinion drifts and how it behaves toward you:

| Personality | Opinion drift (every 65 ticks) | Behavior |
|---|---|---|
| **peaceful** | +1 (not while at war) | Lends you workers when its opinion is 40 or more (see [Worker Lending](#worker-lending)) |
| **aggressive** | -1 | Drifts hostile; can be provoked into **war** when it hates you |
| **mercantile** | +1 while any of your trade routes is running (not while at war); -1 while none is and its opinion is above 0 | Warms to an active trader |
| **isolationist** | 1 toward 0 | Slow to befriend, slow to anger |

Drift runs alongside natural drift and the rival or embargo drain, and opinion stays within -100 to +100.

---

## Trade deals

Every civilization you have met offers a small, rotating set of **trade deals**. Nothing is written by hand for each civilization: its specialty, personality and strength, and its opinion of you, decide what it offers. The Factions panel shows each civilization's deals on its card. `diplomacy deals [civ]` lists them, `diplomacy accept <civ> <n>` takes one, and `plan deal <civ> <n>` queues one in the [build plan](plan.md).

Every deal is worded from your side, the same way on the panel, in `diplomacy deals`, in the plan and in the log: `<kind>: give <price> → get <goods>`, for example `Buy: give 876M coal → get 966K food`.

| Kind | You give | You get |
|---|---|---|
| **Buy** | one of your construction resources | the civilization's specialty, sized by the goods |
| **Sell** | one of the three resources whose stores are fullest | the civilization's specialty, a little more than a Buy pays |
| **Goodwill** | one of the three resources whose stores are fullest | **+5 opinion** instead of goods (only offered below opinion 50) |
| **Rare** | one of your construction resources | a construction resource of the **next** age that this age's market doesn't sell, at a steep price (0.6 of its next-age parity) |

Next to each line the panel shows how much better than the market the deal pays (`+15% vs market`), `next age's goods` for a Rare deal, or `not sold at the market` when the market doesn't trade the pair. A price you can't pay yet shows in red, and deals you've taken are dimmed.

Where the market trades the pair, a deal pays **5% to 25% better than the market** (it trades at 0.84 to 1.0 of parity, where the market pays 0.8). Where the market doesn't trade the pair, the rate comes from the two resources' price levels (food and culture are valued by what the age's producers of them make). A deal never pays better than parity, so trading still never beats building. Deals ignore market pressure: each one is a fixed contract.

A civilization's personality sets how many deals it offers and which kinds:

| Personality | Deals |
|---|---|
| **peaceful** | 2 when neutral, 3 friendly, 4 allied; mostly Buy deals |
| **mercantile** | one more than peaceful, 2 points better rates, 20% bigger lots, and more Sell deals (it asks for your goods) |
| **aggressive** | one fewer than peaceful (at least one), 4 points worse rates, and more Goodwill deals |
| **isolationist** | one deal (two when allied): a **Rare** deal when the next age has goods for it to sell, its specialty otherwise |

Opinion matters too. Friendly civilizations (friendly status, or opinion 25+) and allies offer more deals, at better rates (0.88 of parity when neutral, 0.92 friendly, 0.96 allied, plus 0.03 on a Sell) and in bigger lots. A civilization **at war** with you, under your **embargo**, your **rival**, or with opinion **-50 or lower** offers nothing and won't honor the offers it already made until that changes. Taking a Buy, Sell or Rare deal adds +1 opinion and counts as a trade; a Goodwill deal adds +5. Deals can raise opinion only up to **50**: they can bring a civilization to the edge of an alliance but not past it.

A deal moves about 1.5 median building prices of the age (×0.9 plus 0.1 per point of the civilization's strength, ×1.25 friendly, ×1.5 allied, ×1.2 mercantile, ×2 for a Rare deal, and a roll between ×0.75 and ×1.25). It's capped so the goods fit in half your storage and the price in 80% of it. Amounts are rounded to three figures, never in your favor. A deal can't be taken while its goods wouldn't fit in the room left in their store.

Offers rotate after **4,680 ticks of play** (about 2h 36m) and when you advance an age. The timer only runs while the game does: offline catch-up doesn't advance it, so when you come back you find the offers you left, and a deal you planned is still there for the plan to take while you're away. The card shows the time to the next set.

---

## Allied Bonuses

An ally raises your production of its specialty resource:

```
production rate = normal rate × (1 + ally bonus)
route income    = base income × (1 + harbor bonus + ally bonus)
```

The bonus applies to your whole per-tick rate of that resource (buildings, workers and everything else), and to that resource on every trade route that brings it in. It is its own multiplier, outside [the all-production cap](resources.md#the-all-production-cap). Two allies with the same specialty would add their bonuses together. An ally at war with you gives nothing.

The bonus shows in the **Active Multipliers** section of the Stats panel as a `Diplomacy` line on the affected resource, and the Trade panel lists the same bonuses under **Allied bonuses**, one line per ally (for example *Merchant Guild: +20% gold*).

Each civilization's bonus, from +15% to +30%, is in the [Civilization Reference](#civilization-reference).

An alliance lasts until you change the status. Going back to neutral is free, but allying again costs another 500 gold.

---

## Worker Lending

A peaceful civilization with opinion 40 or more, not at war with you, now and then **lends you** 3 to 6 workers, announced in the log (a 12% chance every 65 ticks). Lent workers join your pool at once, idle, even past your housing, and go home after 520 ticks (about 17 minutes); idle workers leave first. If the civilization's opinion is **above 80** when it lends, the loan is **permanent**: the workers stay for the rest of the run. A civilization has one loan out at a time, so a permanent loan is the last it makes. Loans show on the Factions panel and on the lender's card.

---

## Encounters: boons and setbacks

An encounter, first contact or a repeat meeting, usually brings something home. What it brings is settled in this order:

1. **At war.** About **one in three** meetings with a civilization you are at war with brings a **setback**. The rest are **standoffs**: the two parties see each other and withdraw, and you come home with nothing but the sighting. A civilization you are fighting never gives a gift.
2. **The mission failed.** A failed run that still met someone always brings a setback.
3. **All five boon slots are full.** The game rolls a boon anyway. An instant gift (a lump of resources, a crew of workers) still arrives, since it needs no slot. A timed boon is turned away: the envoys are thanked and sent home with their crates unopened, and about **one in four** of those refusals turns into a setback instead.
4. **Otherwise** the encounter grants a **boon**.

### Boons

A boon is rolled from one catalog shared by every civilization:

| Boon | Effect | Size | Lasts |
|---|---|---|---|
| Specialty Windfall | more output of the civilization's specialty | +8% to +20% | 3,900 to 7,800 ticks (2h 10m to 4h 20m) |
| Resource Surge | more output of a random resource of your age or earlier | +8% to +18% | 3,900 to 6,500 ticks (2h 10m to 3h 37m) |
| Enlightenment | more knowledge | +12% to +25% | 3,900 to 7,800 ticks |
| Industrious Spell | more of all production | +5% to +12% | 2,600 to 5,200 ticks (1h 27m to 2h 53m) |
| Time Dilation | faster ticks: the whole game runs faster | +8% to +15% | 1,950 to 3,900 ticks (1h 5m to 2h 10m) |
| Supply Drop | an instant lump of a random resource of your age or earlier | grows with the age | instant |
| Grand Cache (rare) | a bigger instant lump of a rarer resource (knowledge, gold, uranium, plasma, ...) | grows with the age | instant |
| Extra Hands | 3 to 8 temporary workers | | 2,600 to 5,200 ticks |

The sizes are before scaling. Stronger civilizations give bigger boons (×0.98 at strength 1 up to ×1.30 at strength 5), and so do better relations: ×1.15 friendly, ×1.4 allied, ×0.7 from a rival or a civilization under embargo. Industrious Spell adds into [the all-production cap](resources.md#the-all-production-cap) like any all-production bonus.

Which boon you get depends on the civilization's character. Every civilization favors sharing its specialty. A peaceful one leans toward gentle production and knowledge gifts and rarely speeds time; an aggressive one toward Time Dilation and lumps of spoils; a mercantile one toward lumps of goods and resource surges; an isolationist one toward lumps and the rare Grand Cache. Allies make the Grand Cache more likely too.

Instant lumps grow with your age, but far more slowly than prices do. A Supply Drop is worth having in the early ages and small change late in the run.

Timed boons, like setbacks and crews, count down only while you play. The log names the civilization and the reward.

### Extra Hands crews

A crew of 3 to 8 workers joins your pool at once, idle, past your housing if need be, and goes home when the time the log gives is up: 2,600 to 5,200 ticks (about 1h 27m to 2h 53m), counted only while you play. Idle workers leave first. If the crew was staffing buildings, those buildings lose the workers, and the log says how many. Until then the crew has its own line on the Factions panel under Boons and setbacks, with who sent it and the time it has left.

### Setbacks

Setbacks come from their own table:

| Setback | Effect | Lasts |
|---|---|---|
| Dysentery | 1 to 3 workers lost on the way home | instant |
| Spoiled Supplies | 5% to 15% of one resource's stock spoiled or stolen | instant |
| Cursed Relic | one random resource's output falls 8% to 15% | 1,950 to 3,900 ticks (1h 5m to 2h 10m) |
| Bad Omen | all production falls 5% to 10% | 1,300 to 3,250 ticks (43m to 1h 48m) |
| Lost Scouts | 2% to 6% of one resource's stock lost, and all production falls 3% to 6% | 780 to 1,560 ticks (26m to 52m) |

The sizes are before scaling. Setbacks (all but Dysentery's head count) get worse with the civilization's **strength** (×0.90 at strength 1 up to ×1.10 at strength 5) and milder with your relations: ×0.85 friendly, ×0.75 allied, ×1.15 from a rival or under embargo, and ×1.35 at war. So an ally's bad news is gentler than a rival's, and a war with a strong civilization is the worst case. A spoilage never takes more than half a store. Setbacks never touch soldiers.

### Limits

- You can hold **five boons** at once. Only the timed boons take a slot; an instant lump or a crew is used on arrival and holds nothing, so a full set of boons costs you buffs, never goods.
- At most **three timed setbacks** run at once. At that limit a setback can only be one of the instant kinds (Lost Scouts then loses its production dip).
- A setback expires sooner than a boon of the same size.

A player who explores without a break has at least one boon running almost all the time, and all five slots full only 12% to 18% of the time.

---

## Embassy Buildings

Two diplomacy buildings turn workers into a steady source of opinion. Staff an embassy and, each tick, it raises opinion with every **non-hostile** civilization you have met (neutral, friendly or allied; rivals, civilizations under embargo and civilizations at war with you get nothing). Output follows the same `0.20 + 0.80 × fill` staffing curve as production buildings, so a fully staffed embassy runs at full rate and an empty one still trickles. Opinion tops out at +100.

| Building | Unlocks | Cost | Workers | Opinion Rate |
|---|---|---|---|---|
| **Embassy** | Colonial Age | gold + iron | 5 (trade domain) | +0.05 opinion / worker / tick |
| **Grand Embassy** | Industrial Age | gold + steel | 8 (trade domain) | +0.10 opinion / worker / tick (2× the Embassy) |

Embassies use the **trade** worker domain, the same workers who staff markets and harbors. The gain each tick is split across all the non-hostile civilizations you have met, so each embassy does less for each civilization as you meet more of them.

---

## War & Peace

A civilization declares **war** at the moment you provoke it, if **both** conditions hold then: its opinion is **below -75**, *and* you have provoked it at least twice in all, this time included. **Raiding its trade route** (`diplomacy raid`) is one provocation and **putting it under embargo** is another, so a raid plus an embargo, or two raids, can start a war. Provocations are counted until a war ends, so old ones count too: two raids made while it liked you mean the next provocation, once it hates you, starts a war. Anger alone never starts one.

While at war, the civilization **raids** you every 104 ticks (about 3.5 minutes) and takes 50 × its strength (1 to 5) of its specialty resource. War is purely a matter of these raids; there's no tactical combat. Each raid is logged with what you lost. A raid takes its whole amount or nothing: if you hold less than it would carry off, the log says so and you lose nothing. Your garrison blunts part of each raid that lands; see [Defense: what your army blunts](military.md#defense-what-your-army-blunts).

The raids themselves are fixed, small amounts. What a war really costs is everything else: no deals, no ally bonus, your routes that bring in its specialty disrupted ([Trade Disruption](trade.md#trade-disruption-war-amp-embargo)), route cycles that no longer raise its opinion, and encounters with it that bring setbacks or nothing.

There are two ways to make **peace**:

1. **Tribute.** `diplomacy tribute <civ>` pays 300 gold and 50 culture per point of the civilization's strength and ends the war at once. Opinion rises by 25, but never above 0, and a rival or embargo status goes back to neutral.
2. **Wait it out.** A war burns out by itself once 780 ticks (about 26 minutes) have passed since your last raid on that civilization. Stop provoking them and it ends.

Either way the provocation count starts again from zero.

---

## Civilization Reference

The game introduces each civilization when you meet it and never names one before. This table lists the whole roster; skip it if you'd rather find them yourself.

| Key | Name | Eligible From | Personality | Strength | Specialty | Allied Bonus |
|---|---|---|---|---|---|---|
| `riverlands_tribes` | Riverlands Tribes | Bronze Age | peaceful | 1 | food | +15% food |
| `ironhold_clans` | Ironhold Clans | Medieval Age | aggressive | 3 | iron | +20% iron |
| `merchant_guild` | Merchant Guild | Colonial Age | mercantile | 2 | gold | +20% gold |
| `artisan_league` | Artisan League | Industrial Age | peaceful | 1 | culture | +15% culture |
| `atomic_directorate` | Atomic Directorate | Atomic Age | isolationist | 4 | steel | +20% steel |
| `tech_consortium` | Tech Consortium | Information Age | mercantile | 2 | data | +20% data |
| `shadow_syndicate` | Shadow Syndicate | Cyberpunk Age | aggressive | 3 | crypto | +25% crypto |
| `plasma_nomads` | Plasma Nomads | Fusion Age | peaceful | 2 | plasma | +22% plasma |
| `stellar_federation` | Stellar Federation | Space Age | isolationist | 4 | dark\_matter | +20% dark matter |
| `void_reavers` | Void Reavers | Galactic Age | aggressive | 5 | antimatter | +28% antimatter |
| `quantum_collective` | Quantum Collective | Quantum Age | isolationist | 5 | quantum\_flux | +30% quantum flux |

The peaceful civilizations (Riverlands Tribes, Artisan League, Plasma Nomads) are the ones that lend workers; keep their opinion above 80 for a permanent loan. The aggressive ones (Ironhold Clans, Shadow Syndicate, Void Reavers) drift hostile and will declare war if you provoke them while they hate you. The isolationists (Atomic Directorate, Stellar Federation, Quantum Collective) stay near neutral, hard to befriend and hard to anger. The mercantile ones (Merchant Guild, Tech Consortium) warm up while you keep trade routes running.

---

## Strategy

### Ally early and widely

Gifts, alliances and tribute have fixed gold prices that never grow with the age. In the Bronze Age 200 gold is real money; from the Iron Age on it is small change. Four gifts (800 gold) take a civilization from 0 to 60 opinion, and the alliance costs 500 gold more. You can only gift civilizations you have met.

An alliance pays every tick: its bonus covers your whole production of the ally's specialty, and the ally gives the best deals, the biggest boons and the gentlest setbacks. Ally with every civilization whose specialty you produce. The later civilizations carry the biggest bonuses, up to +30%.

### Keep opinion up for free

A single cheap trade route raises opinion with everyone you have met faster than drift lowers it (see [What routes are worth](trade.md#what-routes-are-worth)). With routes running, opinion climbs toward +100 on its own, which keeps peaceful civilizations lending and keeps deals at their best rates. Embassies do the same job without routes, split across the civilizations you know.

### Keep missions running

Encounters are the only way to collect boons, and the main way to meet civilizations early. Chain scouting expeditions and campaigns, and from the Industrial Age let a [Geographic Society](military.md#automatic-dispatch-the-geographic-society) keep parties going while you're away. A failed run brings a setback home, so prefer missions you're likely to win.

### Don't start wars

An embargo or a raid hurts you more than it hurts them: you lose that civilization's deals, ally bonus and boons, and your routes for its specialty. If you don't want to invest in a civilization, leave it neutral.

---

## Tips & Common Mistakes

- **An alliance costs 500 gold each time.** If you drop to neutral and want to ally again, you pay again. Opinion stays when you go neutral, so the gold is all you need.
- **Deals stop at opinion 50.** To get past 50 use gifts, routes and embassies.
- **A full set of boons costs buffs, not goods.** With five timed boons running, an encounter can still bring a lump of resources or a crew.
- **Gifts add up fast.** Six gifts (1.2K gold) take a civilization from 0 to 90 opinion.
- **Opinion drifts toward zero.** Every 260 ticks each civilization not at war with you moves 1 point toward 0, so an alliance left without routes, gifts or embassies slowly cools (the status stays).
