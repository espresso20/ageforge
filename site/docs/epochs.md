# Epochs

Epochs are the 7 meta-progression layers that span your entire civilization arc. Every 3 ages, you cross into a new epoch. Epoch transitions are the most dramatic moments in a run — they change what resources matter, shift building outputs, trigger powerful events, and may force a catastrophic choice.

---

## The 7 Epochs

| # | Epoch | Key | Ages | Primary Resource | Icon |
|---|-------|-----|------|-----------------|------|
| 1 | Stone Era | `stone_era` | Primitive, Stone, Bronze | stone / food | ✧ |
| 2 | Iron Era | `iron_era` | Iron, Classical, Medieval | iron / knowledge | ⚒ |
| 3 | Steel Era | `steel_era` | Renaissance, Colonial, Industrial | steel / coal | ⚙ |
| 4 | Electric Era | `electric_era` | Victorian, Electric, Modern | electricity / oil | ⚡ |
| 5 | Digital Era | `digital_era` | Information, Digital, Cyberpunk | data / crypto | ◈ |
| 6 | Neon Era | `neon_era` | Fusion, Space, Interstellar | plasma / dark_matter | ✦ |
| 7 | Cosmic Era | `cosmic_era` | Galactic, Quantum, Transcendent | antimatter / quantum_flux | ★ |

---

## What happens at an epoch transition?

1. The **Epoch Event roll** fires (see below)
2. Building lineages may shift their output (e.g. Organic Extraction switches from wood to coal at Steel Era)
3. Your status bar updates with the new epoch icon and color
4. The **Epoch tab (F10)** records the transition

---

## Epoch Event Roll

At every epoch transition, one event fires. The type is determined by your **faith level**:

| Faith Level | Chance of Good Event |
|-------------|---------------------|
| 0–25% (No Faith / Dim Faith) | 40% good |
| 26–75% (Low Faith / Mid Faith) | 50% good |
| 76–100% (Strong Faith / Faith Full) | 60% good |

If the roll is **good**, your **culture level** determines the tier:

- Below 10,000 culture → Minor event
- 10,000–250,000 culture → Major event
- Above 250,000 culture → Legendary event (rare)

If the roll is **bad** (60 / 50 / 40%): 70% chance of a Challenging event, 30% chance of a **Catastrophe**.

---

## Good Epoch Events

10 total, tiered by culture level.

**Minor events:**

| Event | Effect |
|-------|--------|
| Age of Plenty | +20% all production for 216 ticks |
| Population Surge | +15% all workers instantly |
| Ancient Cache | Fills 40% of storage with the primary resource |
| Trade Winds | Large gold influx |
| Cultural Festival | Instant culture + faith boost, timed production bonus |

**Major events:**

| Event | Effect |
|-------|--------|
| Grand Discovery | 3 technologies completed instantly (free) |
| Worker Innovation | Permanent +10% worker output multiplier |
| Architect's Gift | 10 buildings constructed instantly (free) |
| Peaceful Century | +20% all production for extended duration |

**Legendary events:**

| Event | Effect |
|-------|--------|
| Epoch Blessing | Permanent +15% all production |

---

## Challenging Events

8 total. These fire when the epoch roll lands bad and the Catastrophe roll is not triggered.

| Event | Effect |
|-------|--------|
| Famine | Food production debuffed for extended period |
| Merchant's Betrayal | Gold stolen, commerce debuff |
| Great Fire | 8 buildings destroyed randomly |
| Epidemic | 20% of workers removed, population debuff |
| Resource Drought | Primary resource production halved |
| Political Instability | Faith and knowledge stolen, knowledge debuff |
| Economic Crash | Gold stolen, trade debuff |
| Dark Age | Active research cancelled, knowledge stolen, research debuff for 144 ticks |

---

## Catastrophe

If a catastrophe is rolled — or you invoke one voluntarily — you face a three-way choice:

### Endure — Pay the price, survive

- 20% of your buildings destroyed randomly
- Resources reduced to 15%
- Workers reduced by 25%
- Production debuff −10% for 216 ticks
- Epoch marked as "Survived" (✓) in your history

### Succumb — Reset but gain permanent power

- Full civilization reset (resources, buildings, workers, and research reset)
- 8 **ruins** placed — ancient remnants that produce at 50% base rate with no workers needed
- **Ancient Knowledge** — permanent +25% research speed
- **Legacy Bonus** — permanent production boost for this epoch's primary resource(s) for all future runs, including prestige resets
- Epoch marked as "Succumbed" in your history

### Defer — Wait and decide later

Close the modal without choosing. The catastrophe stays pending — you will be prompted again on your next login. **You cannot advance ages while a catastrophe is pending.**

---

## Legacy Bonuses by Epoch

Earned by Succumbing to a catastrophe in that epoch. Permanent — they survive prestige resets and stack across multiple runs.

| Epoch | Legacy Bonus |
|-------|-------------|
| Stone Era | wood +20%, stone +20% |
| Iron Era | iron +20% |
| Steel Era | steel +20%, coal +20% |
| Electric Era | electricity +20% |
| Digital Era | data +20%, oil +20% |
| Neon Era | dark_matter +20% |
| Cosmic Era | dark_matter +35% |

---

## Voluntary Catastrophe

You can trigger a catastrophe yourself at any time during an epoch:

```
catastrophe invoke
```

This is useful if you want to Succumb early — for example in the Stone Era when your civilization is small — to lock in the legacy bonus at low reset cost. One catastrophe per epoch maximum.

---

## How to Prepare for Epoch Transitions

- **Faith** — Invest in Faith lineage buildings before each transition. 76%+ faith = 60% good roll. This is the single highest-impact preparation step. See [Faith](faith.md).
- **Culture** — Reach culture thresholds before crossing the epoch boundary. 250,000+ culture = Legendary event eligibility.
- **Storage** — High storage means Challenging events that drain resources hurt less.
- **Research** — Complete Endure-mitigation techs before risky transitions.
- **Voluntary timing** — If you plan to Succumb, do it early in an epoch while your civilization is small. The reset is less painful.

---

## The Epoch Tab (F10)

Press **F10** to open the Epoch tab. It shows:

- Current epoch name, icon, and primary resources
- Last event outcome and catastrophe status
- Full epoch history across all runs
- Legacy bonuses earned so far
- Civilization event log
