# Faith

Faith is a resource that accumulates from Faith lineage buildings and faith-domain workers. It does **not** drain on its own each tick; it goes down only when something spends or removes it (see [How Faith Decreases](#how-faith-decreases)). Your faith level as a percentage of your faith storage sets your [Epoch](epochs.md) event roll odds and the chance a fated [catastrophe](catastrophe.md) strikes.

---

## Why Faith Matters

1. **Epoch roll odds.** Your faith % sets the chance of a good (rather than bad) event at each epoch transition.
2. **Catastrophe odds.** When a doom fated for your era reaches its moment, your faith % at that moment sets the chance it strikes. It sets the odds of the Last Passage at a Cosmic Era prestige the same way. The table below is the one place these odds are listed.
3. **Worker morale.** Producing faith lifts worker [morale](morale.md) a little each tick. The lift scales with your faith **production rate** (faith/tick), not your stored faith, and it is **capped** per tick, so even a huge late-game faith income can't max out morale in one step. A steady faith income is an ongoing morale source on top of the epoch odds.

---

## Faith Threshold Bands

| Faith fill | Good epoch event | Fated doom strikes | Last Passage |
|------------|------------------|--------------------|--------------|
| under 25% of storage | 40% | 90% | 18% |
| 25 to 75% of storage (or no faith storage yet) | 50% | 75% | 15% |
| over 75% of storage | 60% | 60% | 12% |

- The faith that counts is the faith you hold at the moment of the roll: when you cross into a new epoch, when a fated doom reaches its moment, or when you confirm a Cosmic Era prestige. Faith you held when the harbinger came doesn't count.
- A harbinger's **Appease** multiplies the doom and Last Passage chances by 0.6 per level (0.36 at two levels, the most), and **Invite** makes the catastrophe certain. See [The Harbinger](harbinger.md).
- The Economy panel's faith row shows your band and the epoch odds.

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
| Entering the Renaissance Age | 8.1K, alongside 180K gold, 220K knowledge and 880 steel |
| Sistine Chapel (Renaissance wonder) | 20K, alongside culture, gold and stone |

Bank faith through the Medieval Age so the Renaissance requirement doesn't hold you up. The age requirement only checks your faith, and your faith carries into the new age untouched. The Sistine Chapel, which you must build before leaving the Renaissance Age, does spend it (through `wonder collect`). That lowers your faith % until it refills, so keep an eye on the odds if an epoch boundary is close or a harbinger has come.

---

## The Faith Supply Chain

| Stage | Target Setup |
|-------|-------------|
| Early game | A few Shrines, 2 Faith workers each |
| Mid game | Multiple Temples / Cathedrals with full Faith worker assignment |
| Late game | Keep faith above 75% before every epoch boundary and while a harbinger warns you |

---

## Common Mistake

Players often neglect faith until the epoch notification appears, and by then it is too late to raise it before the roll. Build faith infrastructure at the **start** of each epoch, not the end.

The same goes for a [harbinger](harbinger.md): it comes only part of an age before its doom strikes, so faith you haven't banked by then is hard to raise in time.

---

## Epoch Preparation Checklist

Before reaching the last age of an epoch (e.g. the Bronze Age before the Iron Era):

- Faith above 75% of storage for the 60% good roll
- Culture above 40% of storage to make Major good events eligible; above 75% also gives a 15% chance at the Legendary one
- Spare materials to rebuild with, in case The Great Fire destroys up to 8 buildings (never wonders or storage)
- A gold income that can absorb Economic Crash or Merchant Betrayal, which each take half your gold and then drain more per tick

See [Epochs](epochs.md) for the event tables and [Catastrophe](catastrophe.md) for what a doom does.
