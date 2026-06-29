# Resources

AgeForge has **26 resources** that unlock progressively as you advance through the 22 ages. Resources accumulate passively based on your buildings and worker assignments. Each resource has a base storage cap that is intentionally small — building storage lineage buildings is required to hold meaningful quantities.

---

## Full Resource Table

### Basic Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Food | `food` | Primitive Age | 50 | Feeds all workers every tick — must stay positive |
| Wood | `wood` | Primitive Age | 50 | Primary early building material |
| Knowledge | `knowledge` | Primitive Age | 30 | Powers all research; starts capped low — build Knowledge lineage early |
| Faith | `faith` | Primitive Age | 50 | Drains over time; level as % of cap affects epoch roll odds |
| Stone | `stone` | Stone Age | 50 | Durable construction material |
| Iron | `iron` | Bronze Age | 50 | Metal for tools, weapons, and early engineering |
| Gold | `gold` | Bronze Age | 50 | Currency for trade, diplomacy, and mid-game buildings |

### Knowledge & Faith

See dedicated sections below — both have special mechanics beyond simple accumulation.

### Culture

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Culture | `culture` | Classical Age | 50 | Accumulates permanently; gates prestige bonuses at thresholds |

See dedicated section below.

### Ore Intermediates

These resources are produced by Geological Extraction buildings and consumed by the Metallurgy lineage. They are intermediate materials, not used directly in buildings or tech.

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Marble | `marble` | Iron Age | 30 | Refined stone for monumental construction |
| Iron Ore | `iron_ore` | Iron Age | 30 | Raw ore before smelting — feeds Metallurgy lineage |
| Titanium Ore | `titanium_ore` | Modern Age | 20 | Raw titanium ore — refines into titanium via Metallurgy |
| Dark Matter Crystals | `dark_matter_crystals` | Cyberpunk Age | 10 | Crystallised dark matter — refines into dark matter |
| Nanobots | `nanobots` | Modern Age | 20 | Microscopic machines. Built by the **Nano Foundry** (Modern Age) and Organic Extraction from Digital Era+; consumed as a build material by several digital/cyberpunk buildings |

### Industrial Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Coal | `coal` | Iron Age | 50 | Fuel for smelting; Organic Extraction output in Steel Era |
| Steel | `steel` | Medieval Age | 30 | Refined metal for advanced construction |
| Oil | `oil` | Industrial Age | 50 | Fuel for machines; Organic Extraction output in Electric Era |
| Electricity | `electricity` | Victorian Age | 50 | Powers modern infrastructure |
| Uranium | `uranium` | Atomic Age | 30 | Radioactive reactor fuel |

### Advanced Resources

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Data | `data` | Modern Age | 50 | Digital information and analytics; produced by Hacker domain |
| Crypto | `crypto` | Cyberpunk Age | 50 | Decentralized digital currency |
| Plasma | `plasma` | Fusion Age | 30 | Superheated ionized gas for energy |
| Titanium | `titanium` | Space Age | 30 | Lightweight refined metal for space construction |
| Dark Matter | `dark_matter` | Interstellar Age | 20 | Exotic refined matter for warp technology |
| Antimatter | `antimatter` | Galactic Age | 20 | Annihilation fuel for megastructures |
| Quantum Flux | `quantum_flux` | Quantum Age | 10 | Unstable quantum energy for reality manipulation |

### Military

| Resource | Key | Unlocks | Base Storage | Notes |
|----------|-----|---------|--------------|-------|
| Soldiers | `soldiers` | Iron Age | Sum of military buildings' soldier caps | Produced by military buildings when workers are assigned; spent to launch expeditions |

Soldiers are a true stockpiled resource (the 26th), produced and stored by your military buildings. See [Soldiers](#soldiers) below and [Military & Expeditions](military.md) for the full mechanics.

---

## Food

Food is the most critical resource. Every worker in every domain drains food each tick — the rate scales with their class tier (`baseFoodCost × 1.12^tier`). If food hits zero, workers begin to die at a rate of 1 per 5 ticks until food is restored.

**How to increase food:**
- Assign food-domain workers to Food lineage buildings (`assign gathering_camp 5`)
- Build more Food lineage buildings to increase total worker capacity
- Research Agriculture, Crop Rotation, and related techs for food multipliers

**Watch the food rate** in the Economy tab — keep the rate (`+N/t`) positive before recruiting new workers from any domain.

---

## Knowledge

Knowledge powers all research. Technologies cost knowledge to unlock, and your knowledge storage sets the ceiling on how much you can bank before it caps out.

**How to increase knowledge:**
- Assign knowledge-domain workers to Knowledge lineage buildings
- Build Knowledge lineage buildings (Story Circle → Library → University → ...)
- Knowledge storage starts at 30 — build Knowledge buildings early to raise the cap

Knowledge workers have the highest production multiplier progression of any domain: at Quantum tier (tier 20), a single Quantum Theorist produces `2^20 × base_rate` — massively more than a Primitive-age knowledge worker (tier 0).

---

## Faith

Faith accumulates from Faith lineage buildings and faith-domain workers. It does not drain automatically each tick — it only decreases through specific epoch events (e.g. Political Instability removes 60% of current faith). The faith resource unlocks at the Primitive Age. Early-age faith workers (Devotee, Believer, Worshipper, Celebrant, Initiate) exist from Primitive through Classical ages with lower food costs. The formal high-cost Faith domain tier (Acolyte, base 2.0 food/t) begins at Medieval Age.

Your faith level as a **percentage of your storage cap** determines epoch event roll odds:

| Faith % of Cap | Epoch Roll (Good Chance) |
|----------------|--------------------------|
| Less than 25% | 40% |
| 25% to 75% | 50% |
| More than 75% | 60% |

**Advice:** Assign Faith workers (Acolytes and their successors) well before you expect an epoch transition. Faith production from Faith lineage buildings auto-runs, but workers accelerate accumulation significantly.

---

## Culture

Culture **accumulates permanently** rather than draining. Your culture storage cap grows as you build Culture/Arts lineage buildings. Once accumulated culture passes a threshold, the bonus is permanent — it does not require maintaining a balance.

| Threshold | Bonus |
|-----------|-------|
| 500 | +5% knowledge rate |
| 2,500 | +10% knowledge rate |
| 10,000 | +15% knowledge rate, unlocks a wonder tier |
| 50,000 | +20% knowledge rate |
| 250,000 | +25% knowledge rate, enables culture events |
| 1,000,000+ | +30% knowledge rate and beyond |

**On prestige:** Culture is reduced to 20% of its current value. Bonuses from thresholds you already passed remain permanently — only the current culture total is cut.

Culture buildings (Lineage 10: Amphitheater → ...) auto-produce culture without requiring worker assignment. Build them passively while focusing workers on higher-priority domains.

---

## Ore Processing Chain

The metallurgy pipeline requires two lineages working together:

```
Geological Extraction buildings  →  raw ore  →  Metallurgy buildings  →  refined metal
     (masonry workers)                              (metallurgy workers)
```

| Raw Ore | Refined Output | Notes |
|---------|---------------|-------|
| Iron Ore | Iron / Steel | Classical Age ore; Metallurgy lineage starts Iron Age |
| Titanium Ore | Titanium | Modern Age ore |
| Dark Matter Crystals | Dark Matter | Cyberpunk Age ore |

You must have both the Geological Extraction lineage buildings (to produce ore) and the Metallurgy lineage buildings (to refine it) staffed with appropriate workers to keep metal flowing. An ore surplus with no smelters, or smelters with no ore supply, both result in zero output.

---

## Resource Storage

The 25 civilian resources share storage from the Storage lineage buildings. Each Storage tier raises the cap for every one of them simultaneously. The base caps are very small (10–50 units) — you will hit them early. (Soldiers are the exception: their cap comes from your military buildings, not the Storage lineage — see [Soldiers](#soldiers).)

**Priority:** Build a new Storage lineage building as your first or second action when entering any new age.

Late-game resources (Plasma, Titanium, Dark Matter, Antimatter, Quantum Flux) have especially low base storage (10–30 units) and will cap out instantly without dedicated storage investment. See [Buildings](buildings.md) for the full Storage lineage progression.

---

## Gold

Gold enables trade, diplomacy gifts, and many mid-to-late game building costs.

**How to increase gold:**
- Assign trade-domain workers to Trade lineage buildings (Market, Bank, Stock Exchange, ...)
- Complete trade routes that export surplus resources
- Research Currency, Mercantilism, and related technologies

---

## Soldiers

Soldiers are a stockpiled resource — the 26th — that unlocks at the **Iron Age**. Unlike civilian resources, soldiers are both **produced** and **stored** exclusively by your military buildings.

**How soldiers are produced:** Each military building you own (War Camp, Barracks, and successors) produces soldiers every tick when military-domain workers are assigned to it, exactly like food or wood production scales with assigned workers. A fully-staffed military building produces roughly its soldier cap ÷ 50 soldiers per tick (minimum 0.1/tick).

**How soldiers are stored:** Your soldier storage cap is the **sum of every military building's soldier cap**. There is no separate Storage-lineage contribution — building more (and higher-tier) military buildings is the only way to raise the ceiling. A single War Camp caps you at 10 soldiers; a Barracks adds 20; caps double per tier up the military lineage.

**How soldiers are spent:** Launching an [expedition](military.md#expeditions) deducts that expedition's soldier cost from your stockpile at launch — whether the run succeeds or fails. Hostile catastrophe events can also reduce your soldier count.

See [Military & Expeditions](military.md) for production formulas, per-building soldier caps, and the full expedition table.
