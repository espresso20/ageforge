# AgeForge TODO

Status legend: [ ] pending | [~] in progress | [x] done

---

## Phase 1: Prestige Fixes ✓ DONE
## Phase 2: Wonder Resource Banking System ✓ DONE
## Phase 3: Map Improvements ✓ DONE
## Phase 4: Villager Dashboard Panel ✓ DONE
## Phase 5 (Economy Redesign): Config Foundation ✓ DONE
## Phase 6 (Economy Redesign): Worker-Building Coupling Engine ✓ DONE
## Phase 7 (Economy Redesign): Age Transition Transformation Pass ✓ DONE
## Phase 8 (Economy Redesign): Epoch System ✓ DONE
## Phase 9 (Economy Redesign): Catastrophe System ✓ DONE
## Phase 10 (Economy Redesign): Balance & Building Content ✓ DONE
## Phase 11 (Economy Redesign): UI Completion ✓ DONE
## Phase 12: Documentation Rebuild ✓ DONE
## Phase 13: Epoch-Exclusive Regular Events ✓ DONE

See DONE.md for all completion notes.

---

## Phase 14: Critical Bug Fix Pass

- [ ] Fix `gather` cap — `ui/input.go` clamps `amount = 10000` instead of `amount = 10`
- [ ] Remove duplicate `gathering_camp` definition — `config/buildings.go` has a stale stone-age entry with no WorkerDomain that shadows the lineage version; delete it
- [ ] Fix assign command — `assign` expects domain keys (`food`, `knowledge`) but autocomplete and snapshot types use legacy names (`worker`, `shaman`); align so the command and autocomplete both use domain keys
- [ ] Fix over-assignment — `AssignWorkers` allows assigning workers to a building with count 0; add building-count validation before assigning

---

## Phase 17a: Stone + Military Age Gating

- [ ] `stone_camp` produces `stone` but `stone` is not unlocked until `stone_age` — either unlock stone in `primitive_age` or move `stone_camp` to `stone_age`
- [ ] Hunting Lodge grants military cap but military domain is not available until Bronze Age — gate Hunting Lodge to bronze_age or remove military cap from primitive-age buildings

---

## Phase 17b: Balance Pass

- [ ] Food rate vs drain — at primitive age, fully-staffed gathering camp produces ~+0.05 food/tick vs -1/tick per worker; players starve immediately. Review and raise gathering camp food rate or lower worker drain
- [ ] Hut build time — 50–80 ticks is too long; players run out of food before they can shelter workers. Reduce to match early-game pacing
- [ ] Building tick times across primitive and stone age — audit all early-game BuildTicks for consistency with food drain speed
- [ ] Stone camp resource bar — if stone is kept in primitive age, add stone to the resources panel for that age

---

## Phase 15: Autocomplete Rewrite

- [ ] Replace `unlockedVillagerTypes` (returns legacy names) with domain key list throughout autocomplete
- [ ] Remove stale first-arg suggestions (e.g. `assign worker wood`) — `assign` first arg must be domain key
- [ ] Add autocomplete for `unassign all <domain>` using domain keys
- [ ] Verify `recruit`, `assign`, `unassign` all suggest the correct canonical keys

---

## Phase 16: Command History

- [ ] Add a ring-buffer command history (last 50 commands) to the input field
- [ ] Up arrow scrolls back through history; down arrow scrolls forward
- [ ] History is session-only (not persisted to disk)

---

## Phase 18: Population Screen Redesign

- [ ] Remove or repurpose the Population panel — it duplicates what the Economy tab already shows
- [ ] Add a dedicated **Under Construction** section to the Economy screen showing active build queue progress bars
- [ ] If multiple buildings of the same type are queued, show a single combined bar (e.g. `Hut x3 [████░░] 2 ticks`)
- [ ] Decide fate of population panel: either replace with a compact workers-per-domain summary or remove entirely

---

## Completed
<!-- Move items here with [x] as they are finished -->
