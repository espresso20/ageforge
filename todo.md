# AgeForge TODO

Status legend: [ ] pending | [~] in progress | [x] done

---

## Phase 1: Prestige Fixes
- [x] Increase all prestige shop costs by at least 200% (focus on late-game items like Temporal Mastery)
- [x] Lock prestige availability until Modern Age

---

## Phase 2: Wonder Resource Banking System
- [x] Add a per-wonder resource bank (tracked separately from player storage)
- [x] New command: `Wonder Collect <RESOURCE> <AMOUNT|all>` — pulls from player storage into the wonder's bank immediately
- [x] `Build <WONDER_KEY>` only succeeds once the wonder's bank is fully funded
- [x] Display wonder bank fill progress in the Wonders tab and Wonder panel (bar + % per resource)

---

## Phase 3: Map Improvements
- [x] Draw the active wonder on the map away from the city (separate visual area)
- [x] Spread the city layout to avoid overcrowding at high tick counts (100k+ ticks / 100+ huts)
- [x] Each age set should use a unique city layout pattern — not just the Primitive-age spiral

---

## Phase 4: Villager Dashboard Panel
- [x] Add a panel to the main game screen showing:
  - Villager types recruitable in the current age
  - Their current gathering rate per type
  - Any active bonuses applied to them

---

## Phase 5 (Economy Redesign): Config Foundation — Data Structures ✓ DONE
See: DONE.md for completion notes. All items complete except MaxCount removal (deferred to Phase 10).

---

## Phase 6 (Economy Redesign): Worker-Building Coupling Engine ✓ DONE
See: DONE.md for completion notes.

---

## Phase 7 (Economy Redesign): Age Transition Transformation Pass ✓ DONE
See: DONE.md for completion notes. UI modal deferred to Phase 11.

---

## Phase 8 (Economy Redesign): Epoch System ✓ DONE
See: DONE.md for completion notes.

---

## Phase 9 (Economy Redesign): Catastrophe System ✓ DONE
See: DONE.md for completion notes.

---

## Phase 10 (Economy Redesign): Balance & Building Content ✓ DONE
See: DONE.md for completion notes.

---

## Phase 11 (Economy Redesign): UI Completion
- [ ] Economy tab: per-building worker assignment (assigned/capacity + domain name + +/- buttons)
- [ ] Villager panel: domain-grouped workers; current-tier prominent, legacy tiers collapsed
- [ ] Age advance modal: transformation summary + epoch event reveal
- [ ] Culture: progress bar toward next threshold
- [ ] Faith: threshold tier indicator
- [ ] Epoch tab (new): epoch history, legacy bonuses, voluntary catastrophe, civilization log
- [ ] Stats tab: epoch/catastrophe/legacy bonus fields

---

## Completed
<!-- Move items here with [x] as they are finished -->
