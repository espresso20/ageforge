# Epochs

Epochs are the meta-progression layer that wraps your entire civilisation arc. You progress through 22 ages linearly, but those ages are grouped into 7 epochs — each one a distinct era of human (and post-human) history with its own resources, events, catastrophe, and flavour.

The critical difference between ages and epochs: **ages are about what you build; epochs are about what happens to you**. An age transition is a technology gate you push through. An epoch transition is a moment of reckoning — a dice roll shaped by your faith and culture that fires exactly one large event with lasting consequences.

---

## The 7 Epochs

| # | Icon | Epoch | Ages | Primary Resource | Energy Resource | Catastrophe |
|---|------|-------|------|-----------------|-----------------|-------------|
| 1 | ◈ | Stone Era | Primitive → Stone → Bronze | wood | food | none (catastrophes start in the Iron Era) |
| 2 | ⚔ | Iron Era | Iron → Classical → Medieval | iron | coal | The Great Plague |
| 3 | ⚙ | Steel Era | Renaissance → Colonial → Industrial | steel | coal | The World War |
| 4 | ⚡ | Electric Era | Victorian → Electric → Atomic | steel | electricity | The Nuclear Exchange |
| 5 | ▣ | Digital Era | Modern → Information → Digital | data | electricity | The Great Hack |
| 6 | ◉ | Neon Era | Cyberpunk → Fusion → Space | plasma | plasma | Corporate Armageddon |
| 7 | ✦ | Cosmic Era | Interstellar → Galactic → Quantum → Transcendent | dark_matter | antimatter | The Reality Tear |

The Cosmic Era is the only epoch with 4 ages instead of 3.

### Epoch Flavours

**◈ Stone Era** — *"Humanity's first steps — wood, stone, and fire."*
Your settlement scratches out survival. Food is the bottleneck; wood is the building block. Events are raw and primal — river floods, wandering sages, tribal raids. Every resource counts, and your faith capacity is tiny. No catastrophe can strike here.

**⚔ Iron Era** — *"Empires of iron and faith rise and fall."*
Iron is king. Your armies grow, trade routes lengthen, and the church exerts real influence. The Great Plague is the first catastrophe that can strike, and the cheapest one to Succumb to. Oracle prophecies and imperial roads can dramatically accelerate your mid-game.

**⚙ Steel Era** — *"Steam, steel, and global ambition."*
Industrial scale changes everything. Workers' uprisings can gut your production. The colonial bounty (+5000 gold) is one of the biggest windfall events in the game. Coal seam discoveries extend your energy runway significantly.

**⚡ Electric Era** — *"Electricity and the atom reshape civilisation."*
Power grids, oil strikes, and nuclear theory accelerate research dramatically. This era's catastrophe is The Nuclear Exchange. Nuclear scares and labour movements are short, bearable setbacks compared to what comes later.

**▣ Digital Era** — *"Data flows become the rivers of power."*
Data replaces iron as the critical bottleneck. Server outages are crippling — 50% data production loss for 120 ticks hits hard. AI breakthrough (+50% knowledge production) is the top pick for research-focused runs. The Great Hack catastrophe can erase critical data reserves.

**◉ Neon Era** — *"Augmented reality and the conquest of the solar system."*
Plasma is everything — both primary and energy resource. The neural uprising is the nastiest non-catastrophe event in the game (20% worker loss, food drain, food stolen simultaneously). Corporate espionage can steal 10,000 gold and 8,000 data at once. Plan defensively.

**✦ Cosmic Era** — *"Between stars and beyond time itself."*
Dark matter and antimatter at scales that make earlier resources feel quaint. The Transcendence Signal event (+100,000 knowledge, +50,000 culture) is the most valuable event in the entire game. Reality fractures and entropy waves are manageable if you have strong production. The Reality Tear catastrophe is the hardest reset decision of the run.

---

## Age Awakenings (One Per Epoch)

Each epoch has a single **Awakening** — a deterministic, one-time production boost that fires the first time you enter that epoch's signature age, marking its arrival. These are distinct from the gambled epoch-event roll documented below: an awakening always fires and is always positive (the roll can come up bad), but the boost is temporary and decays. An awakening fires at most once per prestige run and resets on prestige so the next run can earn it again.

| Epoch | Awakening | Triggers On | Effect |
|-------|-----------|-------------|--------|
| ◈ Stone Era | Pottery Mastery | Stone Age | +1 food/tick, +0.5 stone/tick for ~8 min |
| ⚔ Iron Era | Discovery of Metallurgy | Iron Age | +2 iron/tick for ~16 min |
| ⚙ Steel Era | Steam Breakthrough | Industrial Age | +25% all production for ~6.5 min |
| ⚡ Electric Era | Electrification | Victorian Age | +2 electricity/tick, +10% all production for ~10 min |
| ▣ Digital Era | Information Age Dawns | Modern Age | +2 data/tick, +1 knowledge/tick for ~10 min |
| ◉ Neon Era | Cybernetic Awakening | Cyberpunk Age | +20% all production for ~8 min |
| ✦ Cosmic Era | First Contact Signal | Interstellar Age | +1.5 dark matter/tick, +10% all production for ~13 min |

See [Events](events.md#age-awakenings) for full flavor, exact durations, and how awakenings surface in the active-events panel.

---

## How Epoch Events Work

### The Roll

Every time you cross into a **new epoch** (the first age advance that crosses an epoch boundary), the engine rolls your epoch transition event exactly once. This roll never repeats for the same epoch in the same civilisation cycle.

The outcome is determined in two steps:

**Step 1 — Good or bad?** Your faith fill percentage determines the probability:

| Faith % of Storage Cap | Good Event Chance | Bad Roll Chance |
|------------------------|-------------------|-----------------|
| 0–24% (Low Faith) | 40% | 60% |
| 25–75% (Mid Faith) | 50% | 50% |
| 76–100% (High Faith) | 60% | 40% |

**Step 2 — If the roll is bad:** there's a 30% chance it escalates to a **Catastrophe** (modal prompt, your choice). The remaining 70% of bad rolls produce a Challenging event (applied immediately, no choice required). That makes the catastrophe chance 18% / 15% / 12% at low / mid / high faith. A bad roll never escalates on a transition into an epoch before the Iron Era, in an epoch that already had its catastrophe this run, or while another catastrophe is pending. See [Catastrophe](catastrophe.md).

**Step 3 — If the roll is good:** your culture fill percentage gates which tier of event you can receive:

| Culture % of Storage Cap | Eligible Tiers |
|--------------------------|----------------|
| 0–39% | Minor only |
| 40–74% | Minor + Major |
| 75–100% (with 15% chance) | All tiers (Legendary eligible) |

The Legendary roll is a 15% sub-chance within the ≥75% culture bracket — it doesn't fire automatically even if you max culture. The remaining 85% of those rolls draw from Minor + Major.

> **Key insight:** Faith controls your luck; culture controls your upside. You need both to consistently get the best outcomes.

### Cooldown and Anti-Streak

The epoch transition event itself has no cooldown — it fires once per epoch, full stop. However, the **regular random event system** (the events that fire during normal gameplay, not at transitions) does have both a per-event cooldown and an anti-streak system:

- After 3 consecutive good events, the next is forced bad.
- After 2 consecutive bad events, the next is forced good.
- Each event has its own cooldown (minimum ticks between occurrences of that specific event).

This prevents both lucky streaks and brutal punishment spirals during normal play.

### Reading the Epoch Tab (`epoch`)

Press **`epoch`** to open the Epoch tab. It displays:

- Current epoch name, icon, and primary/energy resources
- The result of your last epoch transition roll
- Your catastrophe status for this epoch (pending, survived, succumbed), and the catastrophe odds for the next transition
- Full epoch event history for the current civilisation cycle
- Your legacy bonuses earned across all runs
- Civilisation history log (catastrophe decisions, Succumb/Endure records)

---

## Good Epoch Events

10 events across three tiers. You get exactly one per epoch transition (when the roll is good).

### Minor Events — any culture level

| Event | Effect | Duration |
|-------|--------|----------|
| Age of Plenty | ×2 all production (+100%) | 216 ticks (~7 min) |
| Population Surge | +15% workers added | Instant |
| Ancient Cache | Fills 40% of every resource's storage | Instant |
| Trade Winds | +5.0 gold/tick flat | 144 ticks (~5 min) |
| Cultural Festival | +30% culture, +20% faith (instant) + culture +1.0/tick, faith +1.0/tick | 144 ticks |

### Major Events — 40%+ culture fill required

| Event | Effect | Duration |
|-------|--------|----------|
| Grand Discovery | 3 technologies completed for free | Instant |
| Worker Innovation | Permanent +10% production (all domains) | Permanent |
| Architect's Gift | 10 buildings constructed for free | Instant |
| Peaceful Century | +20% all production | 288 ticks (~10 min) |

### Legendary Event — 75%+ culture fill, 15% sub-chance

| Event | Effect | Duration |
|-------|--------|----------|
| Epoch Blessing | Permanent +15% all production, recorded in history | Permanent |

> Worker Innovation and Epoch Blessing are the two permanent production multipliers in the entire game. They stack. A run that lands both will feel noticeably faster for every subsequent age.

---

## Challenging Epoch Events

8 bad events that fire when the bad roll doesn't escalate to a catastrophe. Applied immediately — no choice, no deferral.

| Event | Effect | Duration |
|-------|--------|----------|
| The Famine | Food production -3.0/tick | 120 ticks |
| Merchant Betrayal | -50% current gold (instant) + gold -2.0/tick | 72 ticks |
| The Great Fire | 8 buildings destroyed randomly | Instant |
| Epidemic | -20% workers (instant) + food -1.5/tick | 180 ticks |
| Resource Drought | Epoch's primary resource -3.0/tick | 90 ticks |
| Political Instability | -60% current faith (instant) + knowledge -2.0/tick | 60 ticks |
| Economic Crash | -50% current gold (instant) + gold -3.0/tick | 216 ticks |
| The Dark Age | Research cancelled, -80% current knowledge (instant) + knowledge -3.0/tick | 144 ticks |

The Great Fire and Epidemic are the two you least want to see. Eight lost buildings in the early game can set you back significantly; 20% worker loss in the Neon or Cosmic era is brutal given how long workers take to replace.

---

## Epoch-Exclusive Random Events

Beyond the transition roll, each epoch has 5 events that only appear in the random event pool while you're in that epoch. These fire during normal gameplay, not at transitions.

### Stone Era ◈

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Tribal Raid | Bad | Food production -0.15/tick, food stolen, -10% workers — lasts 60 ticks |
| Sacred Grove | Good | Faith +0.20/tick + 200 wood — lasts 120 ticks |
| Beast Stampede | Bad | -30 wood, -20 food (instant) |
| River Blessing | Good | Food production +0.25/tick — lasts 144 ticks |
| Wandering Sage | Good | +500 knowledge, +100 faith (instant) |

### Iron Era ⚔

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Iron Vein Strike | Good | Iron production +0.30/tick — lasts 180 ticks |
| Locust Swarm | Bad | Food -0.35/tick, -12% workers — lasts 120 ticks |
| Conquered Village | Good | +2000 gold (instant) |
| Imperial Road | Good | Gold production +0.20/tick — lasts 216 ticks |
| Oracle's Prophecy | Good | Faith +0.30/tick, knowledge +0.15/tick — lasts 144 ticks |

### Steel Era ⚙

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Coal Seam Discovery | Good | Coal production +0.40/tick — lasts 180 ticks |
| Workers' Uprising | Bad | Food -0.15/tick, -500 faith stolen, -8% workers — lasts 120 ticks |
| Colonial Bounty | Good | +5000 gold (instant) |
| Steam Inventor | Good | +2000 knowledge + knowledge production +0.20/tick — lasts 144 ticks |
| Industrial Blight | Bad | Food -0.20/tick, -300 faith stolen — lasts 144 ticks |

### Electric Era ⚡

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Power Surge | Good | Electricity +0.35/tick — lasts 144 ticks |
| Oil Strike | Good | Oil +0.50/tick + 3000 gold — lasts 180 ticks |
| The Broadcast | Good | +5000 culture + faith +0.20/tick — lasts 180 ticks |
| Labour Movement | Bad | Food -0.10/tick, gold -0.10/tick — lasts 60 ticks |
| Nuclear Theory | Good | +8000 knowledge + knowledge production +0.25/tick — lasts 180 ticks |

### Digital Era ▣

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Data Breach | Bad | -5000 data stolen, knowledge -0.20/tick — lasts 120 ticks |
| Viral Moment | Good | +20,000 culture (instant) |
| Tech Monopoly | Good | Gold +0.40/tick — lasts 180 ticks |
| Server Outage | Bad | Data production -0.50/tick — lasts 120 ticks |
| AI Breakthrough | Good | Knowledge +0.50/tick, data +0.20/tick — lasts 216 ticks |

### Neon Era ◉

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Plasma Storm | Good | Plasma +0.50/tick, electricity +0.30/tick — lasts 180 ticks |
| Void Rift | Good | +5000 dark matter (instant) |
| Neural Uprising | Bad | -500 food stolen, food -0.10/tick, -20% workers — lasts 120 ticks |
| Corporate Espionage | Bad | -10,000 gold, -8000 data (instant) |
| Stellar Migration | Mixed | +1000 food (instant), food -0.15/tick — lasts 144 ticks |

### Cosmic Era ✦

| Event | Sentiment | Effect |
|-------|-----------|--------|
| Reality Fracture | Bad | Quantum flux -0.40/tick, knowledge -0.10/tick — lasts 120 ticks |
| Dimensional Harvest | Good | +2000 antimatter, +5000 quantum flux (instant) |
| Galactic Council | Good | Gold +0.20/tick + 20,000 gold — lasts 216 ticks |
| Entropy Wave | Bad | Quantum flux -0.20/tick, knowledge -0.20/tick — lasts 144 ticks |
| Transcendence Signal | Good | +100,000 knowledge, +50,000 culture (instant) |

---

## Endure vs Succumb

When a catastrophe hits, nothing happens until you choose. The game keeps running, but **you can't advance ages or prestige while a catastrophe is pending**. Press Esc to close the choice and look around; a status-bar badge reminds you it is waiting, and typing `catastrophe` reopens it. There is no Defer button.

### Endure — Pay the price and survive

- **20% of your non-wonder buildings** destroyed at random (wonders are spared and don't count)
- Workers of destroyed buildings go back to the idle pool
- **All resources** reduced to 15% of current amounts
- **25% of workers** lost, the same share from every building
- **-10% all production** for 216 ticks (reconstruction), and morale -10
- Earn the "Survived" marker for that epoch, recorded in the civilisation log

Best when: you've built a large, mature civilisation that would be painful to restart, or you already hold this epoch's legacy bonus.

### Succumb — Reset, earn permanent power

- **Up to 8 ruins** from your current buildings (50% base rate in later runs, no workers). Ruins are capped at 24 in total; past the cap the lowest-value ruins crumble first
- **Legacy Bonus** — permanent production bonus for this epoch's primary resource(s), active in all future runs including after prestige
- **Ancient Knowledge** — permanent +25% research speed per distinct epoch succumbed (a second Succumb in the same epoch adds nothing)
- Full reset to the Primitive Age: resources, buildings, workers and research. No prestige points are earned; prestige level and upgrades are kept

**Legacy bonuses by epoch:**

| Epoch | Legacy Bonus |
|-------|-------------|
| ◈ Stone Era | wood +20%, stone +20% (not reachable any more; kept by saves that earned it) |
| ⚔ Iron Era | iron +20% |
| ⚙ Steel Era | steel +25%, coal +25% |
| ⚡ Electric Era | electricity +25%, uranium +25% |
| ▣ Digital Era | data +30%, titanium_ore +30% |
| ◉ Neon Era | plasma +30%, dark_matter_crystals +30% |
| ✦ Cosmic Era | dark_matter +35% |

Best when: you just entered the epoch (low reset cost) and don't hold its legacy bonus yet. Right after entering the Iron Era is the cheapest window.

**The stacking math:** six epochs can be succumbed in (Iron to Cosmic), so Ancient Knowledge tops out at +150% research speed.

---

## Random Event Types (Reference)

These are the effect types that events can apply:

| Effect Type | What It Does |
|-------------|-------------|
| `instant_resource` | One-time add to a resource (no duration, no active event entry) |
| `production` | Multiplier bonus or penalty to a specific resource's production rate |
| `production_all` | Multiplier applied to all production (used by Endure debuff, permanent bonuses) |
| `steal_resource` | Removes a fixed amount from a resource |
| `worker_loss` | Removes a percentage of the total worker pool |

Duration of 0 means instant/one-shot. Duration > 0 means the effect is tracked as an Active Event and ticks down, disappearing when it expires.

---

## Gameplay Strategy by Playstyle

### The Balanced Approach

Keep faith at 50–70% of cap. Invest in culture buildings at a moderate pace. Take transition events as they come without over-optimising.

This works because the math at 50% faith is already coin-flip territory. With decent culture (40%+ fill) you're eligible for Major events. You won't hit Legendary, but Grand Discovery and Worker Innovation are both transformative. The reliable 50% good rate means over the 6 transitions of a run you expect about 3 good events, 2–3 challenging ones, and maybe one catastrophe.

Best for: first or second run, players who don't want to commit hard to any single strategy, relaxed sessions.

### Faith Maximiser

Stack faith production buildings aggressively. Keep faith at 75%+ of cap going into every epoch transition.

Going from 50% to 76%+ faith flips your odds from 50/50 to 60/40. Over 6 transitions that's about 0.6 extra good events compared to a neutral run. It also cuts your catastrophe exposure: catastrophes are 30% of bad rolls, so the chance per transition drops from 15% to 12%.

Trade-offs: faith buildings typically draw on food workers. You're competing with gathering/farming capacity. Don't let food go critical in the early Stone Era chasing faith.

Best for: players who want consistent income and hate variance. Also great for runs where you plan to Endure rather than Succumb — good events let you build strength before the hit.

### Culture Rusher

Prioritise culture buildings to push culture fill as high as possible, aiming for the 75%+ bracket before each epoch boundary.

The Legendary tier is the goal. Epoch Blessing (+15% permanent all production) and Worker Innovation (+10% permanent) are the two most powerful events in the game. Landing Epoch Blessing in the Iron Era or Steel Era compounds across 4–5 more epochs of play. Even if you only hit Legendary once or twice across the whole run, the compounding production boost pays for all the culture investment.

Culture also unlocks Major events (40%+ fill), which are strictly better than Minor events. Grand Discovery (3 free techs) and Architect's Gift (10 free buildings) can leapfrog you through entire age progressions.

Trade-offs: culture buildings require significant investment. This approach is slower to production in the early epochs but accelerates dramatically in the mid-game once Major/Legendary events start landing.

Best for: research-focused runs, players who understand the tech tree well, longer sessions where the late-game payoff matters.

### Speed Runner

Minimise faith and culture investment. Build production infrastructure only. Accept bad events as the cost of faster ages.

The logic: the time spent building faith and culture buildings is time not spent building production buildings. Bad events (Challenging ones, not catastrophes) are mostly temporary debuffs. A Drought or Merchant Betrayal is annoying but recoverable. If your production rate is high enough, you shrug off events that would cripple a weaker economy.

This approach becomes dangerous in the Neon and Cosmic eras where event magnitudes scale up. A Neural Uprising (-20% workers, food stolen) or Corporate Espionage (-10,000 gold, -8,000 data) can cascade badly without buffers.

Best for: experienced players who know the age gates, speedrun-minded sessions, players who plan to Succumb quickly anyway.

### Epoch Endurance (Farming)

Deliberately delay advancing to the next age — staying in the current epoch longer to accumulate more random events, more resources, and more time to build before facing the next epoch transition roll.

When is this worth it?
- You're approaching a transition with low faith and low culture. Stay in the epoch, build faith/culture, then cross the boundary when you're ready.
- You're one or two techs away from unlocking a building that will significantly improve your odds for the next epoch.
- The next epoch introduces a resource you're not yet equipped to produce at scale (e.g. transitioning into Digital Era without data infrastructure in place).

When is it a trap?
- Age gates exist because later ages have better production buildings. Staying in an earlier age means slower resource accumulation overall.
- The longer you stay, the more random events fire — and bad events can damage a stagnant, non-growing civilisation more than a growing one.
- Epoch-exclusive events for your current epoch stop being as useful once you've outgrown their resource amounts.

The sweet spot is usually one full age's worth of extra time (enough ticks to build a few faith/culture structures and fill their caps), not three.

---

## Tips and Common Mistakes

**Don't let faith drain to zero.** At 0% faith fill, you're at 40% good odds. Every point of faith capacity and production matters. Even basic faith buildings provide meaningful transition insurance. Check your faith storage cap — it's easy to underinvest in faith storage while faith production looks "fine."

**Culture storage matters as much as culture production.** Culture gates on fill percentage, not raw amount. A small culture storage at 90% fill beats a massive storage at 10% fill. Don't build enormous culture storage you can't fill before the transition.

**The anti-streak system is your safety net during normal play.** Two consecutive bad events forces a good one. Three consecutive good events forces a bad one. Don't panic after two rough events — a good one is statistically guaranteed next.

**Storage capacity softens Challenging events.** Resource Drought and Economic Crash steal fixed amounts. If you have deep reserves, the impact is proportionally smaller. Storage buildings (granaries, warehouses, vaults) are underrated as catastrophe insurance.

**Know your epoch's primary resource before transitioning in.** The Digital Era wants data infrastructure online before you arrive. The Neon Era wants plasma reactors. Don't cross an epoch boundary and find out you can't produce the era's core resource.

**Succumb early, Endure late.** In the Iron and Steel eras, reset cost is low and the legacy bonus + ruins + research speed compound over many more epochs. In the Neon and Cosmic eras, your civilisation represents enormous investment — Enduring is usually worth the hit.

**Each epoch you Succumb in adds +25% research speed permanently.** Repeat Succumbs in the same epoch add nothing, so the value is in collecting different epochs. This is the main argument for a deliberate early Succumb in the Iron Era.

**One catastrophe per epoch per run.** You cannot chain-catastrophe your way through an epoch, and none can strike before the Iron Era.

---

## Epoch Transitions: What Carries, What Resets

**At epoch transitions (normal age advance into new epoch):**
- Your civilisation continues uninterrupted — no reset, no resources lost
- The epoch event fires once (the transition roll)
- Active events from the previous epoch continue ticking down
- The random event pool shifts to include new epoch-exclusive events
- The Epoch tab records the transition and its outcome
- Your status bar icon and colour update

**After Succumb:**
- Resources, buildings, workers, research: reset to zero
- Ruins (up to 8 new from the last run, 24 in total) placed in your fresh civilisation — produce passively
- All legacy bonuses active and applied
- Ancient Knowledge active: +25% research speed per epoch succumbed
- Epoch event history and catastrophe history preserved
- Prestige bonuses preserved

**After Prestige (end of full run):**
- Similar to Succumb but chosen deliberately, from the Modern Age on, and it earns prestige points
- Refused while a catastrophe is pending
- Legacy bonuses, Ancient Knowledge and ruins carry
- The civilisation log carries; the per-run epoch event history is cleared
- Prestige upgrades available

The epoch framework is designed so that each run builds on the last. A player three runs in has meaningful advantages — ruins giving free passive production, stacked research speed, and legacy bonuses on the resources that matter most — while still needing to play through all 22 ages.
