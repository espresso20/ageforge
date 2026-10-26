# Faith

Faith is a resource that **Faith lineage buildings** make, with **faith-domain workers** in them. It never drains on its own. It goes down only when something spends or removes it (see [How faith decreases](#how-faith-decreases)). Your **faith strength** measures how much your town has put into faith. It sets the odds of the [Epoch](epochs.md) event roll and the chance a fated [catastrophe](catastrophe.md) strikes.

---

## Why faith matters

1. Epoch roll odds. Your faith strength sets the chance of a good (rather than bad) event at each epoch transition.
2. Catastrophe odds. When a doom fated for your era reaches its moment, your faith strength at that moment sets the chance it strikes. It sets the odds of the Last Passage at a Cosmic Era prestige the same way. The table below lists these odds.
3. Worker morale. Producing faith lifts worker [morale](morale.md) a little each tick. The lift scales with your faith **production rate** (faith per tick), not the faith you hold, and it stops growing at 20 faith per tick, so even a huge late-game faith income can't max out morale in one step. A steady faith income is an ongoing morale source on top of the epoch odds.
4. Spending. [Appease](harbinger.md) pays for a lower doom chance in faith (and culture), and one wonder costs faith (see [Faith costs](#faith-costs)).

---

## Faith threshold bands

Your faith strength is your **devotion** times **the share of your faith you have kept**, divided by four and a half.

- **Devotion** is what your own faith buildings have made this run, against what a moderate set of them would have made in your town. A moderate set is eight fully staffed copies of every faith building your age has. A moderate town has a devotion of 1.0x.
- **The share kept** is the faith you hold against the faith your income has made this run. It is 100% until you spend or lose faith.

| Faith strength | Band | Good epoch event | Fated doom strikes | Last Passage |
|----------------|------|------------------|--------------------|--------------|
| under 25% | bottom | 40% | 90% | 18% |
| 25% to 75% | middle | 50% | 75% | 15% |
| over 75% | top | 60% | 60% | 12% |

A typical town sits in the bottom band. A moderate town that never spends its faith reads 22%, in every age, and a town with fewer faith buildings reads less. The bottom row is the odds the game is tuned around. The middle and top rows are what leaning into faith buys: more faith buildings, and fuller crews in them.

| Your faith buildings, against a moderate set | Faith strength, with all your faith kept | Band |
|----------------------------------------------|------------------------------------------|------|
| none | 0% | bottom |
| 1x (a moderate town) | 22% | bottom |
| 1.125x | 25% | the middle band begins |
| 2x | 44% | middle |
| 3x | 67% | middle |
| 3.375x | 75% | the top band begins just above |
| 3.5x | 78% | top |
| 4.5x or more | 100% | top |

- Devotion is counted over the whole run, so it moves slowly. Each age's faith buildings make about twice the last age's, so what you build now soon outweighs what you built before, but faith buildings put up just before a roll change little.
- Staffing counts. An empty faith building makes a fifth of what a full one does. Eight full copies are worth forty empty ones.
- Faith from wonders and events does not count. Stonehenge makes 0.6 faith a tick and the Sistine Chapel 1.8, and every town builds both to advance, so they raise no band. That faith is yours to spend.
- Your town's bonuses cancel out. The moderate set is measured in your town, with your morale, your production bonuses (the faith techs among them), your Era Mastery speed and your time away. Only the faith buildings and their crews make the difference.
- Nothing changes when you advance. The measure follows your run, not your age, so the epoch roll at an advance reads the faith strength you were shown before it.
- Spending lowers it by the share spent. Spend half the faith your income has made and your faith strength halves, and it climbs back as you make more. A windfall, such as an event's gift of faith, can make up for faith you spent but counts for nothing beyond that.
- The measure only reads your faith. It takes nothing from you.
- A harbinger's **Appease** multiplies the doom and Last Passage chances by 0.6 per level (0.36 at two levels, the most), and **Invite** makes the catastrophe certain. Appease costs faith, so it lowers the share you have kept. See [The Harbinger](harbinger.md).
- The faith row of the Resources box shows your faith strength as a percentage and a bar. The label beside it names the band: No faith when you hold none, Dim faith in the bottom band, Low faith in the middle band up to 50% strength and Faith above that, Strong faith in the top band, and Faith full at 100%. A roomy box adds a line under the row with the band in full and the epoch odds, for example "Dim faith (epoch: 40% good)". The `catastrophe` command and the Harbinger panel print the strength with its two parts, for example "faith strength 22% (devotion 1.0x, 100% of your faith kept)".

A moderate set by age. The last two columns are what your faith buildings must average over the run, before bonuses, to reach each band with all your faith kept:

| Age | Faith buildings in a moderate set | It makes, per tick | Middle band from | Top band over |
|-----|-----------------------------------|--------------------|------------------|---------------|
| Primitive Age | 8 | 0.016 | 0.018 | 0.054 |
| Stone Age | 16 | 0.048 | 0.054 | 0.162 |
| Bronze Age | 24 | 0.112 | 0.126 | 0.378 |
| Iron Age | 32 | 0.24 | 0.27 | 0.81 |
| Classical Age | 40 | 0.496 | 0.558 | 1.67 |
| Medieval Age | 48 | 1.01 | 1.13 | 3.4 |
| Renaissance Age | 56 | 2.03 | 2.29 | 6.86 |
| Colonial Age | 64 | 4.08 | 4.59 | 13.8 |
| Industrial Age | 72 | 8.18 | 9.2 | 27.6 |
| Victorian Age | 80 | 16.4 | 18.4 | 55.2 |
| Electric Age | 88 | 32.8 | 36.8 | 111 |
| Atomic Age | 96 | 65.5 | 73.7 | 221 |
| Modern Age | 104 | 131 | 147 | 442 |
| Information Age | 112 | 262 | 295 | 885 |
| Digital Age | 120 | 524 | 590 | 1.77K |
| Cyberpunk Age | 128 | 1.05K | 1.18K | 3.54K |
| Fusion Age | 136 | 2.1K | 2.36K | 7.08K |
| Space Age | 144 | 4.19K | 4.72K | 14.2K |
| Interstellar Age | 152 | 8.39K | 9.44K | 28.3K |
| Galactic Age | 160 | 16.8K | 18.9K | 56.6K |
| Quantum Age | 168 | 33.6K | 37.7K | 113K |
| Transcendent Age | 168 | 33.6K | 37.7K | 113K |

Each faith building makes twice what the one before it does, so the newest one carries about half of a moderate set's faith: sixteen fully staffed copies of it make about what the whole set does. The per-building rates are under [How to produce faith](#how-to-produce-faith).

The bands don't read a share of what you hold. Stock has no limit, so a pile of faith says nothing about how devoted a town is. They read what your faith buildings have made against a moderate set, times the share of your faith you have kept. A town that works at it can reach the middle and the top rows.

High faith pays off over a run. A doom is fated in 27% of the eras from the Iron Era on (see [When it triggers](catastrophe.md#when-it-triggers)). A first run to a Modern Age prestige lives through three eras that can hold a doom (the Iron, Steel and Electric Eras), so it can expect about 0.73 catastrophes at low faith, 0.61 at mid faith and 0.49 at high faith. High faith saves you about a quarter of a catastrophe per run, and brings more good epoch events. A deep run to a Quantum Age prestige lives through all six such eras: about 1.46 catastrophes at low faith, plus the Last Passage roll.

---

## How faith decreases

Faith does not drain on its own. It goes down only when:

- the **Political Instability** epoch event fires (you lose 60% of the faith you hold),
- a random event takes some: Workers' Uprising and Industrial Blight (Steel Era events) take 15% and 10% of the faith you hold, within limits that scale with your age and never more than a quarter of it, and Heresy (from the Medieval Age on) cuts faith production by half for 31 ticks,
- you **Appease** a [harbinger](harbinger.md), which costs faith and, from the Steel Era on, culture,
- you build the **Sistine Chapel** (see [Faith costs](#faith-costs)),
- you **Endure** a catastrophe, which cuts every resource you hold, faith included, to 15% (more with Brace and a garrison, see [Endure](catastrophe.md#endure)).

Otherwise your faith total only grows.

Some things add faith in a burst. The Cultural Festival epoch event adds 20% of the faith you hold at once, then raises faith production by 50% for 374 ticks. The First Shrine milestone pays 10 faith, and Devout Settlement (16 Shrines) adds 5% to faith production. A few random events give faith or raise its production for a while: Humming Grove and Wandering Sage in the Stone Era, Oracle's Prophecy in the Iron Era and The Broadcast in the Electric Era.

---

## How to produce faith

Faith comes from **Faith lineage buildings** with **faith-domain workers** assigned.

Faith lineage, in order:

Shrine → Standing Stones → Altar → Temple → Oracle House → Cathedral → Basilica → Mission → Church → Grand Cathedral → Revival Hall → Spiritual Center → Meditation Center → Digital Temple → Cyber Shrine → Neon Sanctuary → Quantum Chapel → Orbital Sanctuary → Void Monastery → Stellar Shrine → Transcendence Hall

Faith rates, fully staffed, per building:

| Building | Age | Faith/tick | Workers |
|----------|-----|------------|---------|
| Shrine | Primitive | 0.002 | 2 |
| Standing Stones | Stone | 0.004 | 2 |
| Altar | Bronze | 0.008 | 3 |
| Temple | Iron | 0.016 | 3 |
| Oracle House | Classical | 0.032 | 4 |
| Cathedral | Medieval | 0.064 | 5 |
| Basilica | Renaissance | 0.128 | 5 |
| Mission | Colonial | 0.256 | 5 |
| Church | Industrial | 0.512 | 5 |
| Grand Cathedral | Victorian | 1.02 | 6 |
| Revival Hall | Electric | 2.05 | 6 |
| Spiritual Center | Atomic | 4.1 | 6 |
| Meditation Center | Modern | 8.19 | 7 |
| Digital Temple | Information | 16.4 | 7 |
| Cyber Shrine | Digital | 32.8 | 8 |
| Neon Sanctuary | Cyberpunk | 65.5 | 8 |
| Quantum Chapel | Fusion | 131 | 9 |
| Orbital Sanctuary | Space | 262 | 9 |
| Void Monastery | Interstellar | 524 | 10 |
| Stellar Shrine | Galactic | 1.05K | 10 |
| Transcendence Hall | Quantum | 2.1K | 12 |

Faith is a **flow resource**. Unlike construction resources, whose rates follow the [Payback Rule](buildings.md#how-production-rates-are-set), its rates are set by hand and double each tier. Two wonders make faith of their own: Stonehenge (Bronze Age) adds 0.6 faith a tick and the Sistine Chapel (Renaissance Age) adds 1.8. In the Bronze Age that is more than five times what a moderate set of faith buildings makes. Three techs raise faith production: Ritual and Calendar by 10% each, and Theology by 8%.

A Shrine holds 2 workers, and later faith buildings hold more (the Workers column above). By default the [roster](workers-and-domains.md#the-roster) staffs your faith buildings along with everything else, so a new Shrine fills on its own while housing and food allow. To push more of your workforce into faith, give the domain a share, or assign by hand:

```
roster faith 20    # keep 20% of your workers on faith buildings
assign shrine 2    # put 2 idle workers on your Shrines
```

Workers aren't recruited into a domain. A worker counts as a faith worker while it staffs a faith building.

---

## Faith costs

No age gate asks you to hold faith, and advancing spends none. The gates do ask for faith buildings, as they ask for every other building that makes something (see [Advancement requirements](ages.md#advancement-requirements)). One building costs faith:

| What | Faith needed |
|------|--------------|
| Sistine Chapel (Renaissance wonder) | 32K, alongside culture, gold and stone |

Your faith carries into a new age untouched. The Sistine Chapel, which you must build before leaving the Renaissance Age, spends it (through `wonder collect`). That lowers the share of your faith you have kept, and your faith strength with it, until your income makes the faith back. If you have worked your way into the middle or top band, watch the odds when an epoch boundary is close or a harbinger has come.

---

## Building faith up

| Stage | Target setup |
|-------|--------------|
| Early game | Shrines with 2 faith workers each |
| Mid game | Temples and Cathedrals with full faith crews (3 and 5 workers) |
| Late game | A moderate set, fully staffed, keeps you in the bottom band. The middle band begins at 1.125 times that set, and the top band just above 3.375 times |

---

## Build faith early

Faith buildings put up at the last minute barely move your faith strength, because devotion is counted over the whole run. By the time the epoch notification appears it is too late to raise it before the roll. If you want better odds than a typical town's, build faith infrastructure at the **start** of each epoch, not the end.

The same goes for a [harbinger](harbinger.md). It comes only part of an age before its doom strikes, so devotion you haven't built by then is hard to raise in time.

---

## Epoch preparation checklist

Before reaching the last age of an epoch (for example the Bronze Age before the Iron Era):

- If you are leaning into faith: a faith strength of 25% or more for the 50% good roll, and over 75% for the 60% one. A typical town rolls at 40%.
- A [culture strength](epochs.md#culture-strength) above 40% makes Major good events eligible. Above 75% it also gives a 15% chance at the Legendary one.
- Spare materials to rebuild with, in case The Great Fire destroys up to 8 buildings (never wonders, storage or knowledge buildings).
- A gold income that can absorb Economic Crash or Merchant Betrayal, which each take half your gold and then cut gold production for a while.

See [Epochs](epochs.md) for the event tables and [Catastrophes](catastrophe.md) for what a doom does.
