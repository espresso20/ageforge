# Faith

Faith is a resource that accumulates from Faith lineage buildings and faith-domain workers. It does **not** drain on its own each tick; it goes down only when something spends or removes it (see [How Faith Decreases](#how-faith-decreases)). Your **faith strength**, which measures how much your town has put into faith, sets your [Epoch](epochs.md) event roll odds and the chance a fated [catastrophe](catastrophe.md) strikes.

---

## Why Faith Matters

1. **Epoch roll odds.** Your faith strength sets the chance of a good (rather than bad) event at each epoch transition.
2. **Catastrophe odds.** When a doom fated for your era reaches its moment, your faith strength at that moment sets the chance it strikes. It sets the odds of the Last Passage at a Cosmic Era prestige the same way. The table below is the one place these odds are listed.
3. **Worker morale.** Producing faith lifts worker [morale](morale.md) a little each tick. The lift scales with your faith **production rate** (faith/tick), not your stored faith, and it is **capped** per tick, so even a huge late-game faith income can't max out morale in one step. A steady faith income is an ongoing morale source on top of the epoch odds.

---

## Faith Threshold Bands

Your faith strength is your **devotion** times **the share of your faith you have kept**, divided by four and a half.

- **Devotion** is what your own faith buildings have made this run, against what a moderate set of them would have made in your town. A moderate set is five fully staffed copies of every faith building your age has. A moderate town has a devotion of 1.0x.
- **The share kept** is the faith you hold against the faith your income has made this run. It is 100% until you spend or lose faith.

| Faith strength | Band | Good epoch event | Fated doom strikes | Last Passage |
|----------------|------|------------------|--------------------|--------------|
| under 25% | bottom | 40% | 90% | 18% |
| 25% to 75% | middle | 50% | 75% | 15% |
| over 75% | top | 60% | 60% | 12% |

**A typical town sits in the bottom band.** A moderate town that never spends its faith reads 22%, in every age, and a town with fewer faith buildings reads less. The bottom row is what the game is balanced around. The middle and top rows are what leaning into faith buys: more faith buildings, and more priests in them.

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

- **Devotion is counted over the whole run**, so it moves slowly. Each age's faith buildings make about twice the last age's, so what you build now soon outweighs what you built before, but faith buildings put up just before a roll change little.
- **Staffing counts.** An empty faith building makes a fifth of what a full one does. Five full copies are worth twenty-five empty ones.
- **Faith from wonders, techs and events does not count.** Stonehenge and the Sistine Chapel each make more faith when they go up than a moderate set of faith buildings does, the Theology tech adds more, and every town gets them. That faith is yours to spend, but no building the game requires of you can raise your band.
- **Nothing changes when you advance.** The measure follows your run, not your age, so the epoch roll at an advance reads the faith strength you were shown before it.
- **Your town's bonuses cancel out.** The moderate set is measured in your town: with your morale, your production bonuses, your Era Mastery speed and your time away. Only the faith buildings and their crews make the difference.
- **Spending lowers it by the share spent.** Spend half the faith your income has made and your faith strength halves; it climbs back as you make more. A windfall, such as an event's gift of faith, can make up for faith you spent but counts for nothing beyond that.
- **Nothing is taken.** The measure only reads your faith.
- A harbinger's **Appease** multiplies the doom and Last Passage chances by 0.6 per level (0.36 at two levels, the most), and **Invite** makes the catastrophe certain. Appease costs faith, so it lowers the share you have kept. See [The Harbinger](harbinger.md).
- The Economy panel's faith row shows your faith strength as a percentage and a bar, with your band beside it. On a terminal with room, a line under the row gives the band in full and the epoch odds. The `catastrophe` command and the Harbinger panel print it with its two parts, for example "faith strength 22% (devotion 1.0x, 100% of your faith kept)".

A moderate set by age. The last two columns are what your faith buildings must average over the run, before bonuses, to reach each band with all your faith kept:

| Age | Faith buildings in a moderate set | It makes, per tick | Middle band from | Top band over |
|-----|-----------------------------------|--------------------|------------------|---------------|
| Primitive Age | 5 | 0.01 | 0.0112 | 0.0338 |
| Stone Age | 10 | 0.03 | 0.0338 | 0.101 |
| Bronze Age | 15 | 0.07 | 0.0788 | 0.236 |
| Iron Age | 20 | 0.15 | 0.169 | 0.506 |
| Classical Age | 25 | 0.31 | 0.349 | 1.05 |
| Medieval Age | 30 | 0.63 | 0.709 | 2.13 |
| Renaissance Age | 35 | 1.27 | 1.43 | 4.29 |
| Colonial Age | 40 | 2.55 | 2.87 | 8.61 |
| Industrial Age | 45 | 5.11 | 5.75 | 17.2 |
| Victorian Age | 50 | 10.2 | 11.5 | 34.5 |
| Electric Age | 55 | 20.5 | 23 | 69.1 |
| Atomic Age | 60 | 41 | 46.1 | 138 |
| Modern Age | 65 | 81.9 | 92.1 | 276 |
| Information Age | 70 | 164 | 184 | 553 |
| Digital Age | 75 | 328 | 369 | 1.11K |
| Cyberpunk Age | 80 | 655 | 737 | 2.21K |
| Fusion Age | 85 | 1.31K | 1.47K | 4.42K |
| Space Age | 90 | 2.62K | 2.95K | 8.85K |
| Interstellar Age | 95 | 5.24K | 5.9K | 17.7K |
| Galactic Age | 100 | 10.5K | 11.8K | 35.4K |
| Quantum Age | 105 | 21K | 23.6K | 70.8K |
| Transcendent Age | 105 | 21K | 23.6K | 70.8K |

Each faith building makes twice what the one before it does, so the newest one carries about half of a moderate set's faith: ten fully staffed copies of it make about what the whole set does. The per-building rates are under [How to Produce Faith](#how-to-produce-faith).

**Why it isn't a share of storage.** It used to be: the bands read your faith as a percentage of your faith storage. But faith has no store of its own. It is kept in the general store, which is sized for building materials and grows about tenfold an age while faith income doubles. Against the least storage the age gates force on anyone, an economy three times as devoted as the moderate one, saving through a doom's longest warning, reached 20% in the Iron Age, 21% in the Classical Age and under 5% from the Medieval Age on, and by the Cosmic Era a whole age's faith was millionths of the store. So every roll read the bottom band, whatever you did, and those are the odds every game so far was played at. They still are for a typical town. What changed is that the middle and top rows can now be reached, by a town that works at it.

**What high faith is worth.** A doom is fated in 27% of the eras from the Iron Era on (see [When It Triggers](catastrophe.md#when-it-triggers)). A first run to a Modern Age prestige lives through three eras that can hold a doom (the Iron, Steel and Electric Eras; the Digital Era's doom rarely strikes before you prestige there), so it can expect about 0.73 catastrophes at low faith, 0.61 at mid faith and 0.49 at high faith. High faith saves you about a quarter of a catastrophe per run, and brings more good epoch events. A deep run to a Quantum Age prestige lives through all six such eras: about 1.46 catastrophes at low faith, plus the Last Passage roll.

---

## How Faith Decreases

Faith does not drain on its own. It goes down only when:

- the **Political Instability** epoch event fires (you lose 60% of your current faith),
- a random event takes some: **Workers' Uprising** (500 faith) and **Industrial Blight** (300 faith) in the Steel Era, or **Heresy** (-0.5 faith/tick for 31 ticks) from the Medieval Age on,
- you **Appease** a [harbinger](harbinger.md), which costs faith and culture,
- you build the **Sistine Chapel** (see [Faith Costs](#faith-costs)),
- you **Endure** a catastrophe, which cuts every stored resource, faith included (see [Catastrophe](catastrophe.md#endure)).

Otherwise your faith total only grows. The Cultural Festival epoch event goes the other way: it adds 20% to your faith at once and +1 faith/tick for 374 ticks.

---

## How to Produce Faith

Faith is produced by **Faith lineage buildings** with **Faith-domain workers** assigned.

**Faith lineage** (in order):

Shrine → Standing Stones → Altar → Temple → Oracle House → Cathedral → Basilica → Mission → Church → Grand Cathedral → Revival Hall → Spiritual Center → Meditation Center → Digital Temple → Cyber Shrine → Neon Sanctuary → Quantum Chapel → Orbital Sanctuary → Void Monastery → Stellar Shrine → Transcendence Hall

**Faith rates (fully staffed, per building):**

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

Faith is a **flow resource**: unlike construction resources, whose rates follow the [Payback Rule](buildings.md#how-production-rates-are-set), faith rates are set by hand and double each tier, and the requirements that ask for faith are sized to them. Faith cannot be bought at the market. Early on, the big faith sources are not the lineage buildings: the **Stonehenge** wonder (Bronze Age) adds 0.6 faith/tick, the **Theology** tech (Medieval Age) 0.3 faith/tick and the **Sistine Chapel** wonder (Renaissance Age) 1.8 faith/tick.

**Staffing faith buildings.** A Shrine holds 2 workers; later faith buildings hold more (the Workers column above). By default [worker shares](workers-and-domains.md#worker-shares) staff your faith buildings along with everything else, so a new Shrine fills on its own while housing and food allow. To push more of your workforce into faith, give the domain a share, or assign by hand:

```
workers share faith 20    # keep 20% of your workers on faith buildings
assign shrine 2           # put 2 workers on your Shrines
```

Workers aren't recruited into a domain. A worker counts as a Faith worker while it staffs a faith building.

---

## Faith Costs

Two things ask for faith directly:

| What | Faith needed |
|------|--------------|
| Entering the Renaissance Age | 9.5K, alongside 180K gold and 880 steel |
| Sistine Chapel (Renaissance wonder) | 20K, alongside culture, gold and stone |

Bank faith through the Medieval Age so the Renaissance requirement doesn't hold you up. The age requirement only checks your faith, and your faith carries into the new age untouched. The Sistine Chapel, which you must build before leaving the Renaissance Age, does spend it (through `wonder collect`). That lowers the share of your faith you have kept, and your faith strength with it, until your income makes the faith back. If you have worked your way into the middle or top band, keep an eye on the odds when an epoch boundary is close or a harbinger has come.

---

## The Faith Supply Chain

| Stage | Target Setup |
|-------|-------------|
| Early game | A few Shrines, 2 Faith workers each |
| Mid game | Multiple Temples / Cathedrals with full Faith worker assignment |
| Late game | A moderate set and no more keeps you in the bottom band. For the middle band keep about twice the set, fully staffed; for the top band about three and a half times |

---

## Common Mistake

Players often neglect faith until the epoch notification appears, and by then it is too late to raise it before the roll: devotion is counted over the whole run, so faith buildings put up at the last minute barely move it. If you want better odds than a typical town's, build faith infrastructure at the **start** of each epoch, not the end.

The same goes for a [harbinger](harbinger.md): it comes only part of an age before its doom strikes, so devotion you haven't built by then is hard to raise in time.

---

## Epoch Preparation Checklist

Before reaching the last age of an epoch (e.g. the Bronze Age before the Iron Era):

- If you are leaning into faith: a faith strength of 25% or more for the 50% good roll, over 75% for the 60% one. A typical town rolls at 40%
- A [culture strength](epochs.md#culture-strength) above 40% (about twice a moderate set of culture buildings) to make Major good events eligible; above 75% also gives a 15% chance at the Legendary one
- Spare materials to rebuild with, in case The Great Fire destroys up to 8 buildings (never wonders or storage)
- A gold income that can absorb Economic Crash or Merchant Betrayal, which each take half your gold and then drain more per tick

See [Epochs](epochs.md) for the event tables and [Catastrophe](catastrophe.md) for what a doom does.
