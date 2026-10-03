# AgeForge achievements audit and badge redesign

Read-only audit of `master` at `54c171d`. Nothing in the repo was edited. The working tree carries an uncommitted `game/devmode.go` change (`DevModeActive = true`, marked "TEMP, not for merge"); findings below are against master unless stated.

Verification: config math and account behaviour were checked with throwaway Go programs in the scratchpad (`scratchpad/audit/`, importing the repo through a `replace` directive and pointing the data root at a temp dir with `game.SetDataDirForTest`). The repo's own `data/` was not touched. `go test ./config -run 'Milestone|Feasible'` passes on master, which is part of the problem (see B2).

Contents

1. Short version
2. Inventory: everything achievement-like
3. Bugs, ranked
4. Milestone-by-milestone verdicts
5. Coverage gaps
6. Proposed design
7. Badge art, badge case and mock-ups
8. Reachability rules, CI guard and worked feasibility math
9. The catalog: 565 badges
10. Milestone retune (the in-run layer)
11. Phased plan
Appendix A to D: generated ladder tables

---

## 1. Short version

The account "achievements" system is four hard-coded entries with no notification, no locked list and no progress. It is not broken so much as barely there. The worse problems are around it:

- `account recover` overwrites the active account's file with an empty account under a different ID (verified). Achievements, stats and themes are gone, and the guard only checks themes.
- 7 of the 77 milestones cannot be completed, 3 more are practically out of reach, and 3 of the 6 chains (with their titles) can never finish. The existing feasibility test passes regardless.
- The dev console earns account achievements and permanent theme unlocks with no flag (verified).
- The in-game `account switch` and `account import` drop unflushed achievement records and keep the running game going under the new account (verified), and `Account.Save()` writes to whichever slot is globally active rather than its own.

Recommendation: keep per-run Milestones as the gameplay layer, add account-wide Badges as a separate cosmetic layer, and drive both from one data table and one evaluator. Fix the account persistence bugs before building on top of them.

---

## 2. Inventory: everything achievement-like

| System | Defined | Evaluated (when, how) | Persisted | Displayed |
|---|---|---|---|---|
| Account achievements (4: `first_prestige`, `prestige_x10`, `reached_iron`, `reached_modern`) | `game/account.go:1525` `accountAchievements`, names inline, age orders hard-coded at `:1517` (iron=3, modern=12, both still correct) | `recordEvaluateLocked` from `RecordPrestige` (`completePrestige`, `engine.go:3656`) and `RecordAgeReached` (`advanceAge`, `engine.go:1786`), under ge.mu, taking a.mu only, in memory | `account.json` `achievements []string`, flushed by `FlushIfDirty` in the autosave block (`engine.go:660`) and on `Stop` | Stats overlay "Lifetime (account)" list of names (`ui/overlay_stats.go:67`); Accounts panel count (`ui/accounts_panel.go:690`). No toast, no log line. `site/docs/account.md:222` says so. |
| Account lifetime stats | `AccountStats` (`account.go:217`): `TotalPrestiges`, `HighestAge`, `CivilizationsStarted`, `SavesCompleted` | First two by the Record* hooks. The last two are never incremented. | `account.json` `stats` | Stats overlay, Accounts panel, `account list` |
| Milestones (77) and chains (6) | `config/milestones.go` (header comment still says 74), 13 extra thresholds hard-coded by key in `game/milestones.go:151` and again in `computeProgress` | `checkMilestones` every tick in `doTick` (`engine.go:1368`), under ge.mu; rewards applied to `permanentBonuses`; publishes `milestone_completed` / `chain_completed` | Per save: `Milestones`, `ChainsCompleted`, `CurrentTitle`, `PermanentBonuses`. Reset on prestige and Succumb by design. | Milestones overlay (`ui/overlay_milestones.go`), toast and log line on completion, title in the status bar |
| Titles | Chain titles plus a count ladder (`MilestoneTitles`: 5, 12, 24, 40, 56, 62) | `recalculateTitle` every tick | Per save | Status bar, milestones header |
| Theme unlocks (5 gated themes) | `theme/themes_flavor.go`: `UnlockMilestone` = `bronze_pioneer`, `enlightened`, `information_pioneer`, `cyberpunk_milestone`, `galactic_emperor` | `processThemeUnlocks` in the dashboard refresh (UI goroutine, outside ge.mu; `ui/dashboard.go:655`). Silent on first sync, toast after. | `account.json` `unlocks.themes` via `UnlockTheme` (writes immediately) | Theme picker and `theme list` with lock hints |
| Cheater badge | `GameSave.CheaterBadge` (`game/save.go:132`), set when a save's HMAC fails | `LoadGame` | Per save, sticky through prestige, cleared by `Reset` | Dashboard "shame" line (`ui/dashboard.go:610`), "modified" in Load Game |
| Elite badge (easter egg) | `forgeMasterKey` in `game/devmode.go`, `_proof` field | `LoadGame` verifies the proof | Per save | Splash (reads only `autosave`), "elite" in Load Game |
| Account tamper flag | `Account.Tampered`, `json:"-"` | On load | Not persisted | "modified" marker in Accounts panel and `account list` |
| Per-run stats | `GameStats` (`game/stats.go`): `TotalBuilt`, `TotalRecruited`, `TotalGathered`, `SoldiersTrained`, `Defense` | Engine hooks | Per save | Stats overlay |

Bus events that exist: `building_built`, `age_advanced`, `research_done`, `milestone_completed`, `chain_completed`, `epoch_advanced`, `epoch_event_fired`, `awakening_fired`, `harbinger_arrived`, `game_loaded`. Declared but never published: `worker_added`, `resource_depleted`, `game_saved`.

---

## 3. Bugs, ranked

Severity is my call; "verified" means reproduced in the scratch harness, the rest are traced in code.

**B1. `account recover` destroys the active account (high, data loss, verified).**
`ImportRecoveryCode` (`game/account.go:1230`) builds an empty account with the recovered ID and calls `Save()`, and `Save()` writes to `accountPath()`, which resolves through the global active ID (`account.go:82`, `:866`). It never switches the active ID. Result: the active slot's `account.json` is replaced by an empty account carrying someone else's ID. Scratch run: Alice (`first_prestige`, 1 prestige) recovered Bob's code, and Alice's slot then read `id=81b637d8 name="" ach=[] prestiges=0`. The overwrite guard (`ui/input.go:1010`) only asks for `confirm` when the account has unlocked themes, so an account with achievements and stats but no themes is wiped without a prompt. Recovering your own code does the same to your own progress. The unit tests only check the returned ID, and the smoke accounts scenario happens to wipe the damaged slot a step later, which hides it.

**B2. Seven milestones are impossible, three more are practically out of reach, and three chains can never finish (high, verified from config math).**
The age lock (`engine.go:2770`: you can only build the current age's buildings), the storage ceilings (`smoke.MaxStorage`) and the cost curves put hard caps on counts. No build-cost discount exists before the Classical Age (`civil_engineering`, -5%), and the cost floor is 0.10 (`engine.go:77`).

| Milestone | Asks | Hard limit | Verdict |
|---|---|---|---|
| `stone_mason` | 50 Stone Pits (Stone Age only) | 50th costs 282.7K wood, 169.6K stone; Stone Age max storage 80.1K; ceiling 40 | impossible |
| `temple_city` | 50 Temples (Iron Age only) | 50th costs 14.1M stone vs 1.0M; even the 0.10 floor does not fit; ceiling 31 | impossible |
| `trade_empire` | 30 Trading Posts (Iron Age only) + 12 Merchant Quarters | 30th post costs 1.84M stone vs 1.0M; ceiling 25 | impossible |
| `power_grid` | 50 Coal Plants (Industrial only) | 50th costs 87.6B steel vs 5.16B; below the floor; ceiling 29 | impossible |
| `megalopolis` | 1B population | housing capacity with every housing building of all 22 ages maxed is 942M | impossible |
| `global_city` | 10B population | same, 942M | impossible |
| `grand_architect` | 20,000 structures built in one run | per-run build ceiling for the whole game is 7,561; the nightly bot builds about 550 per run | impossible without sell-and-rebuild churn (B13) |
| `master_builder` | 5,000 built | ceiling passes 5,000 only at the Information Age with every building maxed | practically impossible |
| `metropolis` | 10M population (MinAge iron) | housing ceiling reaches 10.1M at the Digital Age, all housing maxed | practically impossible |
| `urban_sprawl` | 100M population | ceiling 140M at Interstellar, all housing maxed | practically impossible |

Chains blocked: Settlement (`megalopolis`), Builder (`stone_mason`, `grand_architect`, `master_builder`), Trade (`trade_empire`). Their titles ("The Founders", "The Architects", "The Merchants") are unreachable. `TestMilestonesAreFeasible` (`config/effect_text_test.go:228`) only compares counts to `MaxCount` and techs to the tech total, so it passes.

**B3. The dev console earns achievements and permanent theme unlocks, unflagged (high, verified).**
`/age <key>` (`devmode.go`, master) sets `ge.age` and applies unlocks without `advanceAge`, so it records nothing directly, but `/age modern_age` followed by `prestige` runs the real `completePrestige` and grants `first_prestige` (scratch: `prestiges=1 highest="" ach=[first_prestige]`, and `HighestAge` stays empty, an inconsistent account). `/age galactic_age` completes all five theme-gating milestones on the next tick (verified), and `processThemeUnlocks` (`ui/dashboard.go:573`) has no dev check, so the five flavor themes are written to the account for good. That contradicts `ui/theme_account.go:34`, which promises dev mode "writes nothing to the account's unlock set" (true only for the preview path). No cheater badge is set (`CheaterBadge=false` in the scratch run). The uncommitted `DevModeActive = true` would hand all of this, and the dev autocomplete, to every player; revert it before any merge.

**B4. In-game account switch and import lose records and cross-wire accounts (high, verified).**
- `engine.SwitchAccount` (`engine.go:892`) swaps `ge.account` without flushing the old one. A prestige or age-up since the last autosave is lost (scratch: Carol's `RecordPrestige` gone after switching to Dave).
- The running game is not stopped or reset. After `account switch` in play, autosave writes the current run into the new account's `saves/`, and later prestiges and age-ups are credited to the new account. The main-menu switch is safe because Esc saves and `Stop()` flushes first; the typed command is not.
- `Account.Save()` writes to the globally active slot, not `accountDir(a.AccountID)`. Any stale holder that flushes after a switch writes its data into another account's file (scratch: Dave's slot ended up holding Carol's name and achievements). The autosave block reads `ge.Account()` and then flushes, so a switch landing between those two calls does exactly this.
- `ImportAccountExport` (`account.go:1446`) temporarily points the global active ID at the imported slot around its `Save()`. An autosave on the tick goroutine in that window writes the live save and account into the imported slot.
- Importing the active account's own backup replaces `ge.account` with a fresh object and drops the old one's unflushed records.

**B5. Tamper laundering (medium, verified).** A hand-edited `account.json` loads with `Tampered=true`, but `Tampered` is not persisted and any `Save()` (a theme change, a theme unlock, the next flush) re-signs the file. Scratch: added `prestige_x10`, loaded (tampered), called `SetActiveTheme`, reloaded: `Tampered=false`, fake achievement kept. The save path deliberately refuses to launder (`reparentSaveFile`); the account path does not.

**B6. Six milestones never appear in the Milestones panel (medium, display).** `ui/overlay_milestones.go:39` hard-codes the category order without `"epoch"`, despite the comment saying it must match `config.MilestoneCategoryOrder()`. `first_farmers`, `survivor`, `enduring_civilization`, `age_hopper`, `industrial_titan` and `power_grid` never render, completed or not, but they count toward "Progress: X/77".

**B7. Account achievements are invisible when earned (medium, display).** `RecordPrestige` and `RecordAgeReached` return nothing, so there is no toast or log line; the Stats panel shows earned names only (no locked entries, no progress). `AchievementName` falls back to the raw key for unknown keys (for example from a newer export), and `AccountSummary.Achievements` counts unknown keys too.

**B8. Late milestone rewards do nothing (medium, balance).** Milestone `permanent_bonus` values sit in the same `production_all` pool that `recalculateRates` clamps at x3 (`engine.go:1543`, `config.ProductionAllCap`). Techs and earlier wonders alone reach +205% at the Electric Age (`config/pacing.go:459` says as much), and milestone rewards push a normal player over the cap by the Victorian. Every `production_all` reward from Electric Dawn to Transcended (+50%) is swallowed, along with the late `<res>_rate` rewards under the same cap.

**B9. Trivial and trap milestones (medium, design).**
- Soldier thresholds are tiny next to flow income: Military Superpower wants 2,000 soldiers trained at the Industrial Age, where typical soldier income is 920 a tick.
- Eleven milestones need old-tier buildings to still be standing ages after the age lock stopped you building them, while the game prompts you to upgrade them on every advance: `lumber_operation`, `devout_settlement`, `forge_master`, `merchant_guild`, `caravan_network`, `standing_army`, `iron_legion`, `merchant_princes`, `guildhall_master`, `mining_syndicate`, `maritime_empire`. The nightly bot upgrades about 850 times a run (16,379 upgrades across 8 seeds), so it would never earn them.
- Many are auto-completed by age gates: `caravan_network`, `guildhall_master`, `colonial_trade`, `knowledge_seeker`, `gold_hoard`, `industrial_titan`, `wonder_builder`/`wonder_collector`/`wonder_empire` (each age's wonder is required to advance).
- Stale MinAges: `granary_keeper` (granary is Iron, not Bronze), `mining_syndicate` (iron mines are Renaissance), `grand_library_built` (Great Library is Medieval), `wonder_empire` (15 wonders means Digital). The tech-count milestones assume the old tree: Deep Thinker (25) needs the Renaissance, Philosophes (35) the Industrial, Renaissance Mind (42) the Electric, and Tech Master (50) the Information Age, past the point where prestige opens. So the Scholar chain is out of reach for anyone who prestiges on arrival at Modern, which is what the bot does.
- Offline catch-up advances `ge.tick`, so `survivor` and `enduring_civilization` finish while the game is closed (two offline days cover Enduring Civilization).

**B10. Spoilers (medium).** Locked themes print "Reach the Cyberpunk Age", "Reach the Information Age" and "Reach the Galactic Age" to a new player (`theme/themes_flavor.go:102`, `:133`, `:161`, shown by `theme list` at `ui/input.go:1293` and the picker). Non-hidden milestones name the Bronze through Renaissance Ages at game start, and their age progress rows print age indices ("0/3 Age: Iron Age", `game/milestones.go:294`).

**B11. No retroactive grants (low).** Account achievements are evaluated only on a live advance or prestige. A save already past Modern, or an account created after the fact, gets nothing until the next age-up. Themes get a first-sync; achievements do not. `/age` and `LoadGame` never update `HighestAge`. Import merge keeps the local `HighestAge` unless it is empty, even when the backup's is higher (`account.go:1429`).

**B12. Dead stats (low).** `CivilizationsStarted` and `SavesCompleted` are exported, merged and snapshotted but never incremented.

**B13. TotalBuilt can be farmed (low).** It counts raw build completions (`engine.go:2600`). Selling (allowed from the Stone Age, `engine.go:3188`) and rebuilding raises it, which feeds `master_builder`, `grand_architect` and prestige points (+1 per 50 builds, `game/prestige.go:61`).

**B14. The elite splash line never shows for named saves (low).** `ui/splash.go:26` peeks only `autosave`, but every new game is named (`StartNewNamedGame`) and autosaves under that name (`engine.go:646`).

**B15. Latent (low).** Completed milestone keys from a save are not validated on load, so a renamed or removed key would still count toward `CompletedCount` (prestige points, titles). Side note: the `astronaut` worker domain has classes but no staffable building (`buildingMeta` lists `launch_pad` and `warp_gate`, which do not exist), which matters for any per-domain badge.

Locks and determinism: today's hooks are correct. `RecordPrestige`/`RecordAgeReached` take only `a.mu` under ge.mu and do no I/O, and theme unlocks run on the UI goroutine. Lock order is always ge.mu then a.mu, and nothing draws from the run's RNG. The races are the global active-ID ones in B4.

---

## 4. Milestone-by-milestone verdicts

"Hold" means old-tier buildings must survive later ages without being upgraded.

| Group | Milestones |
|---|---|
| Impossible | stone_mason, temple_city, trade_empire, power_grid, megalopolis, global_city, grand_architect |
| Practically impossible | master_builder, metropolis, urban_sprawl |
| Feasible but mis-scaled population | small_village (5K: Iron at best), bustling_town (50K: Colonial at best), growing_city (500K, MinAge Bronze: Electric at best) |
| Trap (hold) | lumber_operation, devout_settlement, forge_master, merchant_guild, standing_army, iron_legion, merchant_princes, mining_syndicate, maritime_empire |
| Hold, but the gate hands it to you | caravan_network, guildhall_master |
| Tight | fortress_state (20th Castle Keep needs 94.5% of max Medieval storage), granary_keeper (MaxCount in its age) |
| Auto-completed by gates or flow income | all 20 age milestones, wonder_builder, wonder_collector, wonder_empire, knowledge_seeker, first_research, tech_pioneer, age_hopper, colonial_trade, gold_hoard, industrial_titan, every soldier milestone's soldier count |
| Tutorial | first_shelter, first_storehouse, first_shrine, first_farmers, storage_network, first_market |
| Stale MinAge or tech count | granary_keeper, mining_syndicate, grand_library_built, wonder_empire, deep_thinker, philosophes, renaissance_mind, tech_master, tech_ascendant (needs the Transcendent tech, not Quantum) |
| Dead reward (x3 cap) | every production_all reward from victorian_innovation on, plus metropolis, wonder_empire, tech_master, military_superpower, maritime_empire |
| Time-gated, offline counts | survivor, enduring_civilization |
| Fine as is | deep_thinker's reward, scholars_haven, cathedral_age, first_soldiers (as a gate on Iron), war_machine and the rest once rescaled |

Gaps inside the milestone set: the Stone Age is the only age without an age milestone, and no milestone marks an epoch.

---

## 5. Coverage gaps

No milestone or achievement touches any of these:

- Catastrophes (Endure, Succumb, Brace effect, legacy bonuses), and the Stone Era meteor, which is gated off entirely
- Harbingers: Appease, Brace, Invite, false prophets, the four outcomes (fulfilled, vindicated, discredited, spared)
- The Last Passage and Cosmic Legacy
- Army defense: garrison, raids blunted, buildings saved in an Endure
- Diplomacy: first contact, alliance, war, tribute, wait-out, gifts, embargo, workers on loan
- Faction deals, trade routes, the market, the black market
- Expeditions (16) and auto-expeditions; faction boons
- The build plan, wonder overflow, idle and offline catch-up, session marks
- Festivals, morale, the 4 cultural monuments; awakenings (7); good and challenging epoch events; random events
- Ancient Memory, prestige upgrades (9), ruins
- Worker domains (only knowledge workers appear, once); upgrades and selling
- Storage beyond three early milestones
- 22 of 26 resources (only knowledge, gold, coal and iron ore appear, as stockpiles)
- Individual wonders; epochs as such; the Stone Age
- Map styles, glyph tiers, themes in use, accounts and backups; speed

---

## 6. Proposed design

### 6.1 Two layers, one engine

Keep them separate in purpose and storage, and share the machinery:

- **Milestones** stay per run. They carry gameplay rewards, feed prestige points and titles, and reset every run. They get retuned (section 10) and move onto the shared definition format.
- **Badges** are account-wide and permanent. Rewards are cosmetic only: the art itself, titles, theme unlocks. A badge never changes a run's numbers. That keeps a run's outcome independent of which account plays it (the determinism tests and smoke seeds depend on this) and avoids power creep across prestiges.
- The 4 old achievements migrate into badges through aliases. Theme unlocks move from milestone keys to age badges (same conditions). The age milestones and age badges then overlap in trigger only: one pays out inside the run, the other goes in the case.

Merging the two into one list was the alternative. I'd rather not: they differ in lifetime, reward and anti-cheat needs. A merged list would let a gameplay reward depend on account history, which is the one thing a per-run layer must not do.

### 6.2 Definitions as data (`config/badges.go`)

```go
type BadgeDef struct {
    Key       string     // "age.iron_age", "lineage.housing.3", "special.hut_hoarder"
    Family    string     // ages, epochs, wonders, maximalist, techs, lineage, domain,
                         // resource, civ, harbinger, doom, expedition, theme, map,
                         // awakening, ladder, special
    Subject   string     // the config key it is generated from
    Name, Desc string    // generated from a family template, overridable per subject
    Tier      Tier       // Bronze, Silver, Gold, Platinum, Legendary: art and points
    Rarity    Rarity     // Common..Legendary: design label, defaults from Tier
    Scope     Scope      // Lifetime (counter), Run (run facts), Moment (event predicate)
    Counter   string     // e.g. "built.lineage.housing", "produced.steel", "endured"
    Threshold float64
    Event     EventKind  // for Moment and Run badges
    Pred      string     // named predicate for challenges ("run.no_endure")
    Reveal    Reveal     // Visible, UntilAge(age), UntilCivMet, UntilHarbingerMet, UntilUnlocked, Secret
    Emblem    Emblem     // symbol id resolved per glyph tier
    Reward    Reward     // Theme key or Title; cosmetic only
    Proof     Proof      // static:<rule> | bot:<style> | integrity (exempt from the guard)
    Aliases   []string   // legacy keys: "first_prestige", "reached_iron", ...
}
```

Generated families are built once at init from config tables (ages, epochs, wonders, lineages, domains, resources, factions, harbingers, epochs' catastrophes, expeditions, themes, map styles, awakenings), the way `theme.buildUnlockIndex` builds its index. Hand-written specials sit in the same file. Challenge predicates are a small registry of named functions on the game side, so config stays pure data.

### 6.3 One evaluator, keyed by event

- The engine calls `ge.badges.Record(ev)` at the existing hook sites: `advanceAge`, `completePrestige`, `Endure`, `Succumb`, `resolveLastPassage`, `resolveHarbinger`, `HarbingerAppease/Brace/Invite`, first contact, status changes, war start and end, `AcceptFactionDeal`, expedition resolution, build completion, research completion, the per-tick rate application, `applyOfflineProgress`, plan starts, festival, black market, Ancient Memory, epoch events, awakenings, raids blunted, upgrades, sells, hand gathering, starvation. UI-side events (theme or map style in use, export, recovery code viewed, dev unlock) call the account directly from the UI goroutine.
- This is not the UI Bus. Bus handlers exist for the UI, run in subscription order and must not do I/O. Badge logic should not depend on who subscribed first.
- `Record` is in-memory only: it takes `a.mu`, touches no file and calls nothing on the engine, so it is safe under ge.mu. This is the discipline `RecordPrestige` already follows.
- Indexing: `byCounter map[string][]*BadgeDef` sorted by threshold plus a cursor per counter. An event adds to one counter and compares against one "next threshold", so each event is O(1) amortized. Moment badges sit in `byEvent map[EventKind][]*BadgeDef`, short lists. Nothing scans hundreds of conditions per tick.
- Production counters: the tick already computes every rate. Add the positive credited amount into 26 floats and emit one `produced` batch every 10 ticks (and one per offline step).
- Run facts for challenges live in the save (`RunFacts{DevTouched, Endured, Succumbed, Traded, PlanStarts, SoldiersTrained, Sessions, AwayTicks, PlayTicks, ...}`, omitempty) and travel in the event, so a challenge like "Reach the Modern Age without choosing Endure" is one predicate over the run facts at the moment of the advance.
- Determinism: `Record` never draws from `ge.rng` or `ge.quip`. Badge log text is fixed per badge (no random quip), so earning a badge cannot shift a run's streams.
- Earned badges go into `a.pendingEarned`. The dashboard calls `engine.DrainEarnedBadges()` from its refresh, outside the lock, then shows the toast and writes the log line with `AddLog`, the same pattern the theme unlock toast uses.
- Theme unlocks obey the same gate as badges (no writes from a tainted run), which closes B3's theme leak.

### 6.4 Persistence and migration

`account.json` schema v2, all new fields `omitempty` and covered by the existing HMAC:

```json
"badges":   {"lineage.housing.2": {"at": 1759190400, "run": "Rome", "f": 0}},
"counters": {"built.lineage.housing": 812, "produced.steel": 7.9e12, "endured": 6},
"days":     ["2026-09-27", "2026-09-28"]
```

- Migration on load: `achievements` keys map through aliases (`first_prestige` to the prestige ladder rung 1, `prestige_x10` to rung 3, `reached_iron` to `age.iron_age`, `reached_modern` to `age.modern_age`) with `at: 0` meaning "earned before badges". Keep reading the old field for a version.
- Retroactive seeding on first v2 load: `HighestAge` grants age badges up to it and `TotalPrestiges` grants prestige rungs, silently (no toast storm, same idea as the theme first-sync).
- Build counters are churn-proof: a build counts only when the building's net built count (builds minus sells, upgrades excluded) passes its previous high in the run.
- `Save()` writes to `accountDir(a.AccountID)`, never the active slot. That one change removes the whole class of B1 and B4 cross-writes. `ImportRecoveryCode` then writes into its own slot and switches to it; the `recover` guard checks any data, not only themes.
- In-game `account switch` and `import` flush the old account, stop the tick loop and return to the menu (or refuse while a run is live).
- Tamper: persist `tampered: true` once a signature fails and keep it through re-signing. Badges earned while it is set get the crossed-rim flag.
- Import and export carry badges and counters. Merge = union of badges (earliest timestamp wins), max of counters (not sum, so importing your own backup cannot double anything), `HighestAge` by age order.
- Drop or wire up `CivilizationsStarted` and `SavesCompleted`.

### 6.5 Anti-cheat policy

Recommendation: **not awarded.** While `DevModeActive` is on, or in any run a dev command has touched, nothing is earned and no counter moves. `DevTouched` is a sticky run fact set by every dev command (`/age`, `/give`, `/fill`, `/build`, `/techs`, `/prestige`, `/speed`, `/god`, `/catastrophe`, `/harbinger`, `/lastpassage`) and saved with the run, so reloading does not clear it.

Why not a crossed rim for dev play: dev mode is the owner's test tool, and anything that persists pollutes the test account for good; not awarding keeps it clean. The one exception is "Hand in the Cookie Jar", awarded on unlocking the console, 0 points, crossed rim, excluded from completion.

Crossed rim is for what can't be prevented, only marked: badges earned in a save flagged modified (`CheaterBadge`) and badges earned while the account is flagged tampered. They show a diagonal strike and are excluded from the score. A forced catastrophe (`source == forced`) never counts either way.

### 6.6 Spoiler rules

- Future ages: per-age entries (age reached, wonder, maximalist, tech tier) are silhouettes until reached. The one exception is the next age's "reached" badge, whose name the Next Age bar already shows.
- Civilizations and harbingers: silhouettes until met. Resources, lineages and domains: until unlocked. Awakenings: until they fire. Catastrophe entries: until the first warning names that passage or it strikes.
- Secret specials stay silhouettes with a one-line hint ("Something about huts").
- Locked theme hints say "Unlocked by reaching a later age" until the age is next or reached (fixes B10). Visible badge descriptions may only name ages up to the reveal age, and the CI guard checks this.
- The counter reads "23 of 135 · ??? hidden". Hidden entries are not counted in the visible denominator. A settings toggle, or reaching the final age, turns "???" into the number.

---

## 7. Badge art, badge case and mock-ups

### 7.1 One grammar, per-tier dials

Following the map lesson (one rendering grammar, tiers as dials, not one-off pieces): a badge is `frame(tier) + rim(tier) + ribbon(tier) + emblem(badge)`.

| Tier | Size (rows x cols) | Frame | Rim | Ribbon | Animation |
|---|---|---|---|---|---|
| Bronze | 3 x 5 | rounded box | none | notch | none |
| Silver | 5 x 9 | octagon | sparkle dots | two tails | none |
| Gold | 7 x 13 | half-block ring | three-stop gradient (shade blocks) | split tails | none |
| Platinum | 7 x 13 | gold frame, cooler hues | gradient with corner stars | split tails | glint on the corner stars |
| Legendary | 9 x 17 | double ring | gradient plus highlight band | long tails | shimmer sweep, glints, pulsing rim |

The emblem is one glyph chosen per family and subject: the lineage's map glyph (reuse `mapmodel` symbols such as `SymHut`, `SymKnowledge`, `SymMine`, `SymHarbor`, `SymEnergy`, `SymHacker`, `SymLaunch`), the epoch icon for epoch badges, a flag for harbingers, a comet for catastrophes, a skull for Succumb, a crown for civilizations, a scroll for techs, a star for meta badges.

Hand-crafted art only for the rarest few: the legendary specials (Stopped Clock, Six for Six, Box Ticker, Museum Piece, Twenty-Two Wonders, Rogues' Gallery, Connoisseur of Endings, Unkillable, Fully Invested) and the Transcendent Age badge get literal 9 x 17 sprites in `badge_art_special.go`. They go through the same colour and glyph pipeline and are folded to ASCII by `mapmodel.Fold`. The wonder half-block sprite renderer (`ui/wonder_icon.go`) could supply these if the owner wants pixel art for the top tier.

### 7.2 Colour

- Metals are fixed identity hues: bronze `#B87333`, silver `#C0C0C0`, gold `#D4AF37`, platinum `#E5E4E2`, legendary a prismatic cycle. Each passes through `theme.Legible(c, Background, 3.0)` for art and 4.5 for text, the way lineage and epoch colours do, so pale metals darken on Daylight and the High Contrast themes.
- Gold and legendary rims are three-stop gradients (dark, base, pale), each stop run through `Legible`.
- Frames of locked and hidden badges, the selection highlight and all text use theme roles (`RoleDim`, `RoleAccent`, `RoleText`), so every theme and the colourblind-safe ones work without special cases.
- The crossed rim strike uses `RoleNegative`.

### 7.3 Glyph tiers

Respect the account's `map glyphs ascii|unicode|nerd` setting (`mapmodel.GlyphTier`). A `badgeGlyphs` table mirrors `mapmodel.Glyph{ASCII, Unicode, Nerd}` for emblems. Nerd tier uses the Font Awesome icons from the Nerd Font private use area (trophy, skull, flag, crown, scroll), with Unicode fallbacks as the map does. Frames and rims are drawn with box and block characters and folded to ASCII (`#` `=` `:` `_` `'` `|`). No emoji-presentation code points (double width).

### 7.4 Animation

Only while the badge case is open, and only the selected badge plus legendary badges on the visible page. A UI-side frame counter (not the engine tick, which pauses with the game and would couple art to game speed) drives a pure function `frame(badgeKey, n)`, so the same frame always looks the same. Shimmer sweep: a highlight band crosses the rim diagonally over 8 frames. Glints: the four corner cells cycle `·` `✧` `✦` `✧` with offset phases. Pulse: the outer ring steps through the gradient. In ASCII the band is a glyph swap instead of a tint.

### 7.5 The badge case

- Command `badges` (alias `achievements`), plus `badges <family>`, `badges next` (closest rungs), `badges <name>`. Opens a dedicated page like the theme picker, with its own key handling.
- Grid of mini badges (every badge at 3 x 5; the frame style shows the tier), grouped by family tabs. Header: earned of visible, "??? hidden", points, current title. The All tab adds a rarity line under the tabs ("common 18 · uncommon 4 · rare 1 · epic 0 · legendary 0") and per-family completion. Selecting a badge shows it at full size with name, description, tier, rarity, points, date earned, the run (save name) it was earned in, and a progress bar for counted goals.
- Small terminals: at 80 x 24 the grid is 7 x 4 minis beside a 33-column detail pane; below 80 columns the detail pane moves under the grid; below 60 the grid shows names in a list.
- Keys: arrows move, Tab cycles family, Enter zooms, `/` searches, Esc closes.
- Toast: one line (the toast bar is one row), with a mini emblem in the tier colour, plus a fixed log line.

### 7.6 Mock-ups (as drawn in the terminal; every row width-checked)

Bronze, 3 x 5 (Housing Hobbyist), then its ASCII fold, then locked, hidden and crossed:

```
╭───╮   .---.   ╭┄┄┄╮   ▗▄▄▄▖   ╭──╱╮
│ ⌂ │   | ^ |   ┆ ? ┆   ▐███▌   │ ⌂╱│
╰─▾─╯   '-v-'   ╰┄┄┄╯   ▝▀▀▀▘   ╰╱──╯
```

Silver, 5 x 9 (Engineering Contractor):

```
 ╭─────╮ 
╭╯ · · ╰╮
│   ⚒   │
╰╮ · · ╭╯
 ╰┬───┬╯ 
```

Gold, 7 x 13, gradient rim (Military Magnate), and its ASCII fold:

```
   ▄▓▓▓▓▓▄         _#####_   
 ▄▓▒░░░░░▒▓▄     _#=:::::=#_ 
▐▓░  ✦ ✦  ░▓▌   |#:  * *  :#|
▐▓░   ♜   ░▓▌   |#:   R   :#|
▐▓░  ✦ ✦  ░▓▌   |#:  * *  :#|
 ▀▓▒░░░░░▒▓▀     '#=:::::=#' 
   ▀█▀ ▀█▀         '#' '#'   
```

Legendary, 9 x 17, animated (Stopped Clock). Frame 1: rim at rest, glints at two corners. Frame 2: the shimmer band crosses the upper left rim, glints move. Frame 3: the band reaches the lower right, the comet pulses to a star.

```
    ▄▄█████▄▄            ▄▄█████▄▄            ▄▄█████▄▄    
  ▄█▓▒░░░░░▒▓█▄        ▄█▓█░░░░░▒▓█▄        ▄█▓▒░░░░░▒▓█▄  
 █▓░ ✧     · ░▓█      █▓█ ·     ✧ ░▓█      █▓░ ·     · ░▓█ 
▐█▒     ☄     ▒█▌    ▐██     ☄     ▒█▌    ▐█▒     ✶     ▒█▌
▐█▒   ✦ ◆ ✦   ▒█▌    ▐█▒   ✧ ◆ ✧   ▒█▌    ▐█▒   ✦ ◆ ✦   ▒██
▐█▒           ▒█▌    ▐█▒           ▒█▌    ▐█▒           ███
 █▓░ ·     ✧ ░▓█      █▓░ ✧     · ░▓█      █▓░ ✦     ✦ ░██ 
  ▀█▓▒░░░░░▒▓█▀        ▀█▓▒░░░░░▒▓█▀        ▀█▓▒░░░░░██▓▀  
    ▀▀█▀▀▀█▀▀            ▀▀█▀▀▀█▀▀            ▀▀█▀▀▀█▀▀    
```

The badge case at 80 x 24. Row 1 of the grid: gold, silver and bronze housing rungs, a bronze medical emblem, two locked, one hidden. Row 4 starts with an earned legendary and a crossed-rim badge earned in a modified save.

```
┌ Badges ───────────────────── 23 of 135 · ??? hidden · 1,245 pts · Magistrate ┐
│ ◂ All  [Ages]  Wonders  Lineages  Resources  Civs  Harbingers  Doom  More ▸  │
│ ▛▀▀▀▜ ╔═══╗ ╭───╮ ╭───╮ ╭┄┄┄╮ ╭┄┄┄╮ ▗▄▄▄▖  │          ▄▓▓▓▓▓▄                │
│ ▌ ⌂ ▐ ║ ⌂ ║ │ ⌂ │ │ ✚ │ ┆ ? ┆ ┆ ? ┆ ▐███▌  │        ▄▓▒░░░░░▒▓▄              │
│ ▙▄▄▄▟ ╚═▾═╝ ╰─▾─╯ ╰─▾─╯ ╰┄┄┄╯ ╰┄┄┄╯ ▝▀▀▀▘  │       ▐▓░  ✦ ✦  ░▓▌             │
│                                            │       ▐▓░   ⌂   ░▓▌             │
│ ╭───╮ ╔═══╗ ╭┄┄┄╮ ▗▄▄▄▖ ▗▄▄▄▖ ▗▄▄▄▖ ▗▄▄▄▖  │       ▐▓░  ✦ ✦  ░▓▌             │
│ │ ⚒ │ ║ ⚒ ║ ┆ ? ┆ ▐███▌ ▐███▌ ▐███▌ ▐███▌  │        ▀▓▒░░░░░▒▓▀              │
│ ╰─▾─╯ ╚═▾═╝ ╰┄┄┄╯ ▝▀▀▀▘ ▝▀▀▀▘ ▝▀▀▀▘ ▝▀▀▀▘  │          ▀█▀ ▀█▀                │
│                                            │ Hut Hoarder                     │
│ ✦▀▀▀✦ ▛▀▀▀▜ ╭───╮ ╭───╮ ╭┄┄┄╮ ▗▄▄▄▖ ▗▄▄▄▖  │ Gold · Epic · 25 pts            │
│ ▌ ☄ ▐ ▌ ⚑ ▐ │ ⚑ │ │ ⚑ │ ┆ ? ┆ ▐███▌ ▐███▌  │ 60 huts before you leave the    │
│ ✦▄▄▄✦ ▙▄▄▄▟ ╰─▾─╯ ╰─▾─╯ ╰┄┄┄╯ ▝▀▀▀▘ ▝▀▀▀▘  │ Primitive Age. The 60th costs   │
│                                            │ 18,956 wood, 1,354 times the    │
│ ▗▟█▙▖ ╭──╱╮ ╭┄┄┄╮ ╭┄┄┄╮ ▗▄▄▄▖ ▗▄▄▄▖ ▗▄▄▄▖  │ first.                          │
│ █ ✶ █ │ ☠╱│ ┆ ? ┆ ┆ ? ┆ ▐███▌ ▐███▌ ▐███▌  │ Earned 29 Sep 2026 · run Rome   │
│ ▝▜█▛▘ ╰╱──╯ ╰┄┄┄╯ ╰┄┄┄╯ ▝▀▀▀▘ ▝▀▀▀▘ ▝▀▀▀▘  │                                 │
│                                            │                                 │
├──────────────────────────────────────────────────────────────────────────────┤
│ Next rung: Housing Contractor  ████████████░░░░░░  41 / 57                   │
│ ←→↑↓ move · Tab family · Enter zoom · / search · Esc close                   │
└──────────────────────────────────────────────────────────────────────────────┘
```

Toast (one row, emblem in the tier colour):

```
◖✦◗ Badge earned: Hut Hoarder (gold). 60 huts before leaving the Primitive Age.
```

### 7.7 Rarity, points, titles and theme rewards

- Points by tier: bronze 5, silver 10, gold 25, platinum 50, legendary 100. Integrity badges are 0. Maximum score with the catalog below: 13,345.
- Rarity is a design label (no telemetry): common (about a quarter of a run or less), uncommon (one run), rare (4 to 5 runs), epic (about 10 runs, or a hard one-run challenge), legendary (25 runs, or luck plus skill). It defaults from the tier and can be overridden (the age maximalists are gold tier, epic rarity).
- Score titles, shown in the badge case header and optionally beside the run title in the status bar: Settler (0), Headman (250), Magistrate (1,000), Sovereign (3,000), Paragon (6,000), Eternal (10,000), Completionist (every countable badge).
- Badge titles: Survivor (Endured ladder rung 3), Harbinger Whisperer (every provable Heeded badge), Merchant Prince (deals ladder gold), The Undying (Unkillable), Cosmic Heir (Last Passage Succumbed), Hut Magnate (Zoning Board).
- Theme rewards: the five existing themes move to age badges with the same conditions (Bronze: Bronze Age; Parchment: Renaissance; Monochrome: Information; Cyberpunk: Cyberpunk; Cosmic: Galactic). Three new ones worth making: Ashfall for Connoisseur of Endings, Ledger for the deals ladder gold, Prismatic for Museum Piece.

---

## 8. Reachability rules, CI guard and worked feasibility math

### 8.1 Rules

1. **One run** means Primitive through Atomic fully played, prestiging on reaching Modern. That is the nightly bot's cycle (2.2 to 2.3 days at 1x, the same for runs 1 to 3). A **deep run** continues past Modern. Every lifetime rung is stated in runs, never days, so it can be recalibrated when run lengths change.
2. **Per-run counts** use the Gate Covenant's math: `copy n price = base x scale^(n-1)` against `smoke.MaxStorage(RequiredAge, res)`, only in the building's own age (age lock), no discounts. The threshold is at most 97% of that ceiling for single buildings and 75% for "max out an age".
3. **Lifetime counts** use lineage-wide or all-building totals, never a single tier. One run's worth R = 10% of the lineage's per-run static ceiling, which matches the nightly bot (about 1,570 buildings over 2.3 cycles against a static ceiling near 14,000, so about 11%). Rungs sit at 0.25, 1, 4, 10 and 25 runs.
4. **Resource ladders** count production credited to you by buildings, workers and techs (not trades, loot, refunds or gifts, which are easy to farm). One run's production P = sum over the run's ages of `config.TypicalIncome(res, age) x AgeTargetTicks(age)`. Rungs sit at 0.25, 1, 5 and 25 runs; resources that only produce after Modern use deep runs.
5. **Time thresholds** are multiples of `config.AgeTargets` (converted to 1x ticks), or counts of runs or sessions, never hours. The owner's possible 4x slowdown then scales them for free. The one exception is the calendar-day streaks, which measure a habit, not pacing.
6. The top rung of any ladder is at most 25 runs of normal play.

### 8.2 The CI guard

`config/badges_feasibility_test.go`, plus a smoke deep-tier assertion. Every `BadgeDef` must carry a `Proof`, and the test fails on:

- **static:copies**: threshold > 0.97 x the building's own-age storage ceiling (0.75 x for age maximalists).
- **static:lifetime**: per-run increment is 0, or threshold / per-run increment > 25 runs.
- **existence**: the subject cannot occur. For example a catastrophe in an epoch where `CatastropheAllowed` is false (this rejects the Great Meteor), a staffing badge for a domain with no staffable building (rejects `astronaut`), a Heeded badge whose passage the harbinger bot has never afforded (rejects 10 figures today), a tech tier count above `TechsByAge`.
- **bot:\<style\>**: the deep tier must see at least one seed of that bot style earn the badge (rush, pacifist, autarky, harbinger, idle and cosmic styles).
- **spoilers**: a visible badge's description names an age later than its reveal age.
- **text**: an em dash or exclamation mark in badge text (extend `TestConfigPlayerTextStyle`).
- **integrity**: exempt, listed explicitly.

The same machinery replaces `TestMilestonesAreFeasible`. With the age lock, storage ceilings and population ceilings it would have failed on all seven impossible milestones.

### 8.3 Worked feasibility math, five ladders

**1. Hut Hoarder (per run, single building).** After `normalizeCostCurves`, a hut costs 14 wood x 1.13^(n-1) (the raw 15 x 1.12 is rewritten to hold the 10th copy's price). Max Primitive wood storage is 50 + 50 stashes x 500 = 25,050. The 62nd hut costs 24,205 and fits; the 63rd costs 27,352 and does not. Ceiling 62; each 10x more wood buys only 18.8 more huts.

| hut | cost (wood) | x the first |
|---|---|---|
| 10th | 42 | 3x |
| 30th | 485 | 35x |
| 50th | 5,584 | 399x |
| 60th | 18,956 | 1,354x |
| 62nd | 24,205 | 1,729x |
| 100th | 2,517,062 | 179,790x |

Threshold 60 (97% of the ceiling). The path: 38 stashes to hold 19,050 wood (27,727 wood in total), 40 wood camps at 0.569 wood a tick each (28,465 wood, about 23 wood a tick staffed), and 164,667 wood for the huts. About 221,000 wood in all, roughly 21 times the Primitive Age's pacing target at that output. A real flex, still reachable. Zoning Board (the 62nd) needs 49 stashes and 210,293 wood for the huts.

**2. Housing ladder (lifetime, lineage-wide).** The 12 housing tiers from Primitive to Atomic have per-run ceilings 62, 54, 53, 43, 38, 36, 41, 51, 51, 50, 46, 45, for C = 570. One run R = 57. Rungs: 14, 57, 230, 570, 1,400, which is 0.25, 1, 4, 10 and 25 runs. Legendary means building every housing tier of every age to its storage limit two and a half times, or ordinary play for 25 runs.

**3. Steel ladder (lifetime production).** Typical steel production from the Medieval to the Atomic Age sums to P = 4.32T per run (the gates ask 71B steel at Electric to Atomic and 1.28T at Modern to Information, so this is the right order). Rungs: 1.1T, 4.3T, 22T, 110T at 0.25, 1, 5 and 25 runs.

**4. Catastrophes Endured (lifetime event count).** A run to Modern can meet at most four catastrophes, one per passage into the Iron, Steel, Electric and Digital Eras (the Stone Era meteor is gated off by `CatastropheGateEpoch`). A full run meets six plus the Last Passage. The nightly bot, Endure policy and no invites, saw 18 catastrophes across 8 seeds of about 2.4 runs each, a little under 1 per run. Rungs 1, 5, 15, 30: about 1, 5, 15 and 30 runs naturally, or 1, 2, 4 and 8 runs if you Invite every passage. Under the redesign (random strikes within an epoch, harbinger only when doom is fated) the count is per strike, so the math holds as long as it stays at most one per epoch.

**5. Stone Age Maximalist (per run, whole age).** The Stone Age's nine buildings against its 80,080 storage: longhouse 54, stone camp 54, stone pit 40, standing stones 40, elders hall 39, woodcutter camp 39, forager post 38, war camp 34, storage pit 25. Ceiling 363, threshold 272 (75%). You may stay in an age as long as you like, so the only hard bound is storage, and every building's price has a source in its own age (the static gate check's `dead_building` rule).

---

## 9. The catalog: 565 badges

### 9.1 Count

| Part | Entries |
|---|---|
| Generated families (15) | 417 |
| Lifetime system ladders (23 ladders) | 73 |
| Hand-written specials | 75 |
| **Total** | **565** |
| Integrity badges (0 points, excluded from completion) | 3 |
| **Countable toward completion** | **562** |
| Visible to a brand new account | 135 (76 generated, 32 ladder rungs, 27 specials) |
| Hidden at start | 430, shown as "???" |

Why 565: it sits inside the requested 300 to 600 and every entry maps to one config row or one engine event. About 74% are generated from config tables, so they grow with content, and the CI guard proves each one reachable. The largest family (resources, 94) is still browsable on one tab. Anything bigger would mean padding with near-duplicates, such as per-building ladders, which the owner has already ruled out.

| Family | Rule | Entries |
|---|---|---|
| F1 Ages reached | one per age after Primitive | 21 |
| F2 Swift epochs | one per epoch | 7 |
| F3 Wonders | one per wonder | 22 |
| F4 Age maximalist | one per age | 22 |
| F5 Tech tiers | one per age | 22 |
| F6 Lineage ladders | 15 lineages (14 production + storage) x 5 rungs | 75 |
| F7 Domain staffing | 11 staffable domains x 2 rungs (astronaut has no staffable building) | 22 |
| F8 Resource ladders | 16 resources x 4 rungs + 10 late resources x 3 rungs | 94 |
| F9 Civilizations | 11 civs x 4 (met, allied, regular, peace terms) | 44 |
| F10 Harbingers | 22 met + 12 heeded (the provable ones) | 34 |
| F11 Doom | 6 catastrophes x 2 (endured, succumbed) + Last Passage x 3 | 15 |
| F12 Expeditions | one per expedition | 16 |
| F13 Themes | one per theme | 11 |
| F14 Map | 2 styles + 3 glyph tiers | 5 |
| F15 Awakenings | one per awakening | 7 |
| **Sum** | | **417** |

### 9.2 Generated families: templates and examples

Names come from the template unless a per-subject override exists. Descriptions come from the template. All text is plain, no em dashes, no exclamation marks.

**F1 Ages reached** (`age.<age>`). Template: name override or the age's name; "Reach the {Age}." Tier by epoch: Stone and Iron Eras bronze, Steel and Electric silver, Digital and Neon gold, Cosmic platinum, Transcendent legendary. Reveal: silhouette until reached, except the next age. Proof: static (the gate is the smoke-proven path). Aliases: `reached_iron`, `reached_modern`. Theme rewards as in 7.7.

| key | name | description | tier |
|---|---|---|---|
| age.stone_age | Rock Solid | Reach the Stone Age. | bronze |
| age.renaissance_age | Born Again, Slightly | Reach the Renaissance Age. Unlocks the Parchment theme. | silver |
| age.transcendent_age | Nothing Left to Prove | Reach the Transcendent Age. | legendary |

**F2 Swift epochs** (`swift.<epoch>`). "Leave the {Epoch} within its pacing target." Threshold: the sum of `AgeTargetTicks` over the epoch's ages, in 1x ticks; "leaving" the Cosmic Era means prestiging from it. Gold. Proof: bot. The current bot already does the Iron Era in 0.98x, the Steel Era in 0.82x and the Electric in 0.76x of target; the Stone Era takes it 1.17x, so that one needs a rush style to prove.

| key | name | description |
|---|---|---|
| swift.stone_era | Out of the Cave Early | Leave the Stone Era within its pacing target. |
| swift.steel_era | Ahead of Schedule | Leave the Steel Era within its pacing target. |
| swift.cosmic_era | Speed of Light, Roughly | Prestige out of the Cosmic Era within its pacing target. |

**F3 Wonders** (`wonder.<key>`). Name: the wonder's name; "Raise the {Wonder}." The detail pane reuses the building's existing Flavor line. Tier by epoch as in F1 (Primitive included). Reveal: until its age is reached. Proof: static (each age's wonder is required to advance).

| key | name | description | tier |
|---|---|---|---|
| wonder.sacred_grove | Sacred Grove | Raise the Sacred Grove. | bronze |
| wonder.eiffel_tower | Eiffel Tower | Raise the Eiffel Tower. | silver |
| wonder.singularity_core | Singularity Core | Raise the Singularity Core. | legendary |

**F4 Age maximalist** (`maximalist.<age>`). "{Age} Maximalist": "Have {N} of the {Age}'s buildings standing before you leave it." N = 75% of the storage ceiling (Appendix C). Gold, epic. Reveal: until reached. Proof: static:copies.

| key | name | N of ceiling |
|---|---|---|
| maximalist.primitive_age | Primitive Maximalist | 229 of 306 |
| maximalist.industrial_age | Industrial Maximalist | 362 of 483 |
| maximalist.transcendent_age | Transcendent Maximalist (before you prestige) | 27 of 36 |

**F5 Tech tiers** (`techs.<age>`). "{Age} Syllabus": "Research all {n} {Age} techs in one run." Silver. Reveal: until reached. Proof: static (older techs have no age lock; count from `TechsByAge`).

| key | name | description |
|---|---|---|
| techs.primitive_age | Primitive Syllabus | Research both Primitive Age techs in one run. |
| techs.medieval_age | Medieval Syllabus | Research all 6 Medieval Age techs in one run. |
| techs.transcendent_age | Transcendent Syllabus | Research the one Transcendent Age tech. |

**F6 Lineage ladders** (`lineage.<lineage>.<1-5>`). Rung names Hobbyist, Contractor, Magnate, Tycoon, Dynasty: "Build {N} {lineage} buildings across all your runs. Sold and rebuilt copies count once." Bronze to legendary. Reveal: when the lineage's first building unlocks. Proof: static:lifetime. Full table in Appendix A.

| key | name | N | runs |
|---|---|---|---|
| lineage.housing.1 | Housing Hobbyist | 14 | 0.25 |
| lineage.food.3 | Food Magnate | 150 | 4 |
| lineage.hacker.5 | Hacker Dynasty | 430 | 25 deep runs |

**F7 Domain staffing** (`domain.<domain>.<1-2>`). "{Domain} Payroll" and "{Domain} Payroll II": "Staff {N} {domain} workers at the same time." N = 10% and 20% of the domain's staffing ceiling by Atomic (Appendix D). Silver, gold. Proof: static:copies plus a population check.

| key | name | N |
|---|---|---|
| domain.food.1 | Food Payroll | 270 |
| domain.metallurgy.2 | Metallurgy Payroll II | 290 |
| domain.hacker.1 | Hacker Payroll | 260 (deep run) |

**F8 Resource ladders** (`resource.<res>.<rung>`). Rung names Trickle, Stream, River, Flood: "Produce {N} {resource} across all your runs." Late resources skip Trickle. Reveal: when the resource unlocks. Proof: static:lifetime. Full table in Appendix B.

| key | name | N | runs |
|---|---|---|---|
| resource.wood.1 | Wood Trickle | 49M | 0.25 |
| resource.steel.4 | Steel Flood | 110T | 25 |
| resource.antimatter.3 | Antimatter Flood | 200,000Q | 25 deep runs |

**F9 Civilizations** (`civ.<key>.<met|allied|regular|peace>`). "{Civ}: First Contact" (Meet the {Civ}), "{Civ}: Allies" (Ally with the {Civ}), "{Civ}: Regular" (Take 5 deals from the {Civ} across your runs), "{Civ}: Peace Terms" (End a war with the {Civ}, by tribute or by waiting it out). Bronze, silver, silver, silver. Reveal: until met. Proof: static (war needs opinion below -75 and two provocations, which any personality allows; alliance needs opinion and 500 gold) plus a bot check. There is no combat, so "war won" becomes "war ended".

| key | name |
|---|---|
| civ.riverlands_tribes.met | Riverlands Tribes: First Contact |
| civ.ironhold_clans.peace | Ironhold Clans: Peace Terms |
| civ.quantum_collective.allied | Quantum Collective: Allies |

**F10 Harbingers** (`harbinger.<key>.met`, `.heeded`). Met: the figure's name, "Hear {figure} out." Heeded: "Heeded {Figure}", "Appease while {figure} is speaking." Bronze, silver. Reveal: until met. Heeded exists only for the 12 figures of the first four passages (Wild Man to Civil Defense Broadcast), where the harbinger-style bot afforded both Appease levels. The other 10 fail the existence check today (the report shows Appease never affordable in the Digital to Neon thread) and join when the bot proves them.

| key | name | description |
|---|---|---|
| harbinger.wild_man.met | The Wild Man | Hear the Wild Man out. |
| harbinger.oracle.heeded | Heeded the Oracle | Appease while the Oracle is speaking. |
| harbinger.future_self.met | Your Future Self | Hear your future self out. |

**F11 Doom** (`doom.<epoch>.endured`, `.succumbed`; `lastpassage.<endured|succumbed|spared>`). "Endured: {Catastrophe}" (Choose Endure against {Catastrophe} and keep going with what is left), "Succumbed: {Catastrophe}" (Succumb to {Catastrophe} and start over with its legacy). Gold, silver. Last Passage: endured platinum, succumbed legendary (grants Cosmic Legacy), spared gold (The Last Passage opens and nothing comes through). Six epochs, Iron to Cosmic; the Stone Era meteor fails the existence check. Keyed to the Endure and Succumb outcomes, not to transitions, so random strikes within an epoch change nothing.

| key | name |
|---|---|
| doom.iron_era.endured | Endured: The Great Plague |
| doom.digital_era.succumbed | Succumbed: The Great Hack |
| lastpassage.succumbed | Succumbed: The Last Passage |

**F12 Expeditions** (`expedition.<key>`). Name: the expedition; "Come back from a successful {expedition}." Silver. Reveal: when available. Proof: bot.

| key | name |
|---|---|
| expedition.scout_party | Scout Party |
| expedition.siege_castle | Siege Enemy Castle |
| expedition.quantum_incursion | Quantum Incursion |

**F13 Themes** (`theme.<key>`). "Dressed as {Theme}": "Advance an age while wearing the {Theme} theme." Tied to progress, not hours, so switching and switching back does not count. Bronze.

| key | name |
|---|---|
| theme.forge | Dressed as Forge |
| theme.high_contrast | Dressed as High Contrast |
| theme.cosmic | Dressed as Cosmic |

**F14 Map** (`map.style.<key>`, `map.glyphs.<tier>`). "Advance an age with the {style} map style" and "Advance an age with {tier} map glyphs." Bronze.

| key | name |
|---|---|
| map.style.roguelike | Dungeon Crawler |
| map.style.skyline | Skyline Watcher |
| map.glyphs.nerd | Font Snob |

**F15 Awakenings** (`awakening.<key>`). Name: the awakening; "Witness {Awakening}." Silver. Reveal: until it fires.

| key | name |
|---|---|
| awakening.awakening_pottery_mastery | Pottery Mastery |
| awakening.awakening_electrification | The Grid Wakes |
| awakening.awakening_first_contact | First Contact Signal |

### 9.3 Lifetime system ladders (73 entries)

Rungs are bronze, silver, gold, and legendary where there is a fourth. "Runs" is the approximate number of normal runs per rung.

| # | Ladder | Rungs | Runs per rung, and the basis | Reveal |
|---|---|---|---|---|
| 1 | Prestiges | 1, 3, 10, 25 | 1, 3, 10, 25 (aliases `first_prestige`, `prestige_x10`) | visible |
| 2 | Catastrophes endured | 1, 5, 15, 30 | about 1 per run naturally, up to 4 with Invite | first warning |
| 3 | Catastrophes succumbed | 1, 5, 15 | each costs a partial run | first warning |
| 4 | Harbinger threads resolved | 1, 10, 40, 100 | 4 threads per run: 0.25, 2.5, 10, 25 | first harbinger |
| 5 | Appeasements bought | 1, 10, 50 | up to 8 per run with a faith focus | first harbinger |
| 6 | Braces bought | 1, 10, 50 | up to 8 per run | first harbinger |
| 7 | Invitations | 1, 5, 20 | up to 4 per run | first harbinger |
| 8 | False prophets exposed | 1, 3, 6 | expected 0.23 per run (12.5% + 7.8% + 3.1% on the thread-starting ages): about 4, 13, 26 | secret |
| 9 | Wonders raised | 25, 100, 300 | 13 per run: 2, 8, 23 | visible |
| 10 | Techs researched | 100, 500, 1,200 | 45 to 49 per run: 2, 10, 25 | visible |
| 11 | Milestones completed | 100, 400, 1,000 | about 40 per run (bot: 106 over 2.4 runs): 2.5, 10, 25 | visible |
| 12 | Chains completed | 1, 10, 30 | 2 to 4 per run once retuned | visible |
| 13 | Expeditions returned | 10, 50, 200, 500 | calibrated from a new smoke column | visible |
| 14 | Faction deals taken | 5, 25, 100 | bot-calibrated | first civ |
| 15 | Raids blunted | 10, 100, 500 | about 30 per run (bot): 0.3, 3, 17 | first raid |
| 16 | Returns after time away | 1, 25, 100 | sessions that ran offline catch-up | visible |
| 17 | Plan items started | 50, 500, 5,000 | bot-calibrated | visible |
| 18 | Buildings upgraded | 250, 2,500, 20,000 | bot does about 850 per run: 0.3, 3, 24 | visible |
| 19 | Festivals held | 1, 10, 50 | cooldown 300 ticks, culture from Classical | first festival |
| 20 | Black market deals | 1, 10, 50 | opens at Colonial, cooldown 240 ticks | when it opens |
| 21 | Ancient Memories accepted | 1, 5, 12 | about 0.5 per prestige (bot): 2, 10, 24 | first offer |
| 22 | Challenging epoch events weathered | 1, 10, 30 | about 1.4 per run (bot): 1, 7, 21 | visible |
| 23 | Market trades | 50, 500, 5,000 | human-scaled (the bot trades far more) | first trade building |

4-rung ladders: 1, 2, 4, 13 (16 entries). 3-rung ladders: the other 19 (57 entries). Total 73.

### 9.4 Hand-written specials (all 75)

Columns: key, name, description, trigger, tier, rarity, reveal (V visible at start, H hidden until relevant, S secret), proof.

**Flex**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| hut_hoarder | Hut Hoarder | 60 huts before you leave the Primitive Age. The 60th costs 18,956 wood, 1,354 times the first. | building_built hut, age primitive, count 60 | gold / epic | V | static:copies |
| zoning_board | Zoning Board | Build the 62nd hut in the Primitive Age. There is no 63rd; the storage will not hold its price. | count 62 in primitive | platinum / legendary | S | static:copies |
| every_nook | Every Nook | Build all 50 stashes before you leave the Primitive Age. | stash count 50 in primitive | silver / uncommon | V | static (MaxCount) |

**Challenges**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| nothing_to_endure | Nothing to Endure | Reach the Modern Age without choosing Endure once in the run. | age_reached modern, run.Endured false | gold / rare | H | bot (appease or luck; one nightly seed saw no catastrophe in 8 days) |
| conscientious_objector | Conscientious Objector | Prestige without training a single soldier in the run. | prestige, run.SoldiersTrained 0 | gold / rare | H | static (no gate asks for soldiers) + bot:pacifist |
| autarky | Autarky | Reach the Modern Age without the market, a trade route or a faction deal. | age_reached modern, run.Traded false | platinum / epic | H | static (cold start with the market off) + bot:autarky |
| hands_on | Hands On | Reach the Industrial Age without the build plan starting anything. | age_reached industrial, run.PlanStarts 0 | silver / uncommon | H | bot |
| exact_change | Exact Change | Raise five wonders in one run with wonder overflow off. | wonder_raised, overflow off all run, count 5 | silver / uncommon | V | static |
| family_heirlooms | Family Heirlooms | Leave the Bronze Age without upgrading a single building. | age_reached iron, run.Upgrades 0 | silver / uncommon | H | static (upgrades are optional) |
| express_lane | Express Lane | Reach the Modern Age within four fifths of the combined pacing target of the twelve ages before it. | age_reached modern, run ticks at most 0.8 x sum of targets | platinum / epic | H | bot:rush (the current bot manages 0.83x) |
| early_riser | Early Riser | Reach the Iron Age within the combined pacing target of the first three ages. | age_reached iron, ticks at most the sum of the three targets | silver / uncommon | H | bot:rush (current bot 1.17x) |
| fashionably_late | Fashionably Late | Spend ten times the Primitive Age's pacing target in the Primitive Age. | tick in primitive at 10 x AgeTargetTicks | bronze / common | V | static |

**Catastrophes and harbingers** (keyed to outcomes and actions, so they survive the random-strike redesign; if an outcome disappears, the existence check retires the badge)

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| sandbagged | Sandbagged | Endure a catastrophe with Brace at level 2. | endure, brace 2 | silver / uncommon | H | bot:harbinger |
| bare_hands | Bare Hands | Endure a catastrophe with no Brace and no garrison. | endure, brace 0, garrison 0 | gold / rare | H | static |
| standing_guard | Standing Guard | Have the garrison save buildings during an Endure. | endure, defense tally buildings saved above 0 | bronze / common | H | bot (seen in 3 of 8 nightly seeds) |
| asked_for_it | Asked For It | Invite a catastrophe, then Endure it. | endure on an invited strike | silver / uncommon | H | static |
| signed_the_guest_book | Signed the Guest Book | Invite a catastrophe, then Succumb to it. | succumb on an invited strike | silver / uncommon | H | static |
| paid_in_full | Paid in Full | Appease a harbinger twice and have the catastrophe come anyway. | harbinger resolved vindicated, appease 2 | gold / rare | H | bot:harbinger |
| tithe_to_a_liar | Tithe to a Liar | Appease a harbinger who turns out to have made the whole thing up. | resolved discredited, appease 1 or more | gold / rare | S | bot:harbinger |
| self_fulfilling | Self-Fulfilling | Invite doom on a false prophet's word and get it anyway. | resolved fulfilled, false prophet | platinum / epic | S | static (Invite guarantees the strike) |
| stopped_clock | Stopped Clock | A false prophet's catastrophe arrives, uninvited. The lie was accurate. | resolved vindicated, false prophet, not invited | legendary / legendary | S | luck; static existence |
| prepared_for_nothing | Prepared for Nothing | Brace to level 2 and be spared. | resolved spared, brace 2 | silver / uncommon | H | bot:harbinger |
| full_chorus | Full Chorus | Hear all three voices of one epoch's warning in a single thread. | resolved, chain length 3 | bronze / common | H | bot (threads span three ages) |
| selective_hearing | Selective Hearing | Ignore a harbinger completely and be spared. | resolved spared, no answers | bronze / common | H | bot (the "ignore" policy is spared about 70% of the time) |
| note_to_self | Note to Self | Appease while your future self is speaking. | appease while the Quantum Age figure speaks | platinum / epic | S | bot:harbinger deep (typical faith and culture income covers the price inside the Quantum Age) |

**Last Passage and Cosmic Legacy**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| knock_knock | Knock Knock | Invite the Last Passage. | invite in the Cosmic Era thread | gold / rare | H | static |
| walked_out_whole | Walked Out Whole | Endure the Last Passage with Brace at level 2. | lastpassage endured, brace 2 | platinum / epic | H | bot:cosmic |
| heirloom_universe | Heirloom Universe | Prestige five times while holding Cosmic Legacy. | prestige with cosmicLegacy, count 5 | platinum / epic | H | static:lifetime |

**Army, diplomacy and trade**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| not_today | Not Today | Blunt 30 raids in one run. | raids blunted in run, 30 | silver / uncommon | H | bot (78 to 100 over about 3 runs) |
| protection_money | Protection Money | End a war by paying tribute. | war_ended, tribute | bronze / common | H | static |
| cold_shoulder | Cold Shoulder | Let a war burn out without paying anyone. | war_ended, wait-out | silver / uncommon | H | static |
| popular | Popular | Be at war with two civilizations at once. | war count 2 | gold / rare | S | static |
| everybodys_friend | Everybody's Friend | Be allied with every civilization you have met, at least four of them. | allied count equals met count, met 4 or more | gold / rare | H | bot |
| regifting | Regifting | Receive workers on loan from a civilization. | lend received | bronze / common | H | bot |
| loyal_customer | Loyal Customer | Take 10 deals from one civilization in one run. | deals with one civ in run, 10 | silver / uncommon | H | bot |
| no_questions_asked | No Questions Asked | Make a deal at the black market. | black market deal | bronze / common | H | static |

**Idle, check-in and the build plan**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| while_you_were_out | While You Were Out | Come back after the full offline allowance to find a plan of 10 or more items built and the plan empty. | offline return at the cap, plan emptied by starts, 10 or more items | gold / rare | V | bot:idle |
| maximum_leave | Maximum Leave | Stay away past the offline allowance. The game stopped counting before you came back. | offline return with elapsed above MaxOfflineTime | bronze / common | V | static |
| regular | Regular | Check in on 7 different days within 10 days. | distinct calendar days | silver / uncommon | V | static (calendar habit, not pacing) |
| standing_appointment | Standing Appointment | Check in on 30 different days. | distinct calendar days | gold / rare | V | static |
| nothing_wasted | Nothing Wasted | Let wonder overflow finish banking a wonder while you are away. | wonder bank full during offline catch-up | silver / uncommon | V | bot:idle |
| long_game | Long Game | Have 20 items waiting in the build plan at once. | plan size 20 | bronze / common | V | static |
| absentee_landlord | Absentee Landlord | Prestige in a run where you spent at least three times as long away as you did playing. Only credited offline time counts. | prestige, run.AwayTicks at least 3 x run.PlayTicks | gold / rare | V | bot:idle-8h (first prestige in about 6.5 days) |

**Milestones and meta**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| chain_reaction | Chain Reaction | Complete a milestone chain. | chain_completed | bronze / common | V | static (after the retune) |
| six_for_six | Six for Six | Complete all six milestone chains in one run. | chains in run, 6 | legendary / legendary | V | static (after the retune) + bot deep |
| box_ticker | Box Ticker | Complete every milestone in one run. | milestones in run equals total | legendary / legendary | S | static (after the retune) |
| collector | Collector | Earn 25 badges. | badge count | bronze / common | V | derived |
| curator | Curator | Earn 100 badges. | badge count | silver / uncommon | V | derived |
| archivist | Archivist | Earn 250 badges. | badge count | gold / rare | V | derived |
| museum_piece | Museum Piece | Earn 400 badges. Unlocks the Prismatic theme. | badge count | legendary / legendary | V | derived |
| full_set | Full Set | Finish every rung of one lineage ladder. | a lineage ladder complete | gold / rare | V | derived |
| twenty_two_wonders | Twenty-Two Wonders | Raise every wonder at least once. | F3 complete | legendary / legendary | V | derived |
| rogues_gallery | Rogues' Gallery | Meet all 22 harbingers. | F10 met complete | legendary / legendary | H | derived |
| small_world | Small World | Meet all 11 civilizations. | F9 met complete | platinum / epic | H | derived |
| connoisseur_of_endings | Connoisseur of Endings | Succumb to all six catastrophes. Unlocks the Ashfall theme. | F11 succumbed complete | legendary / legendary | H | derived |
| unkillable | Unkillable | Endure all six catastrophes and the Last Passage. | F11 endured complete plus lastpassage endured | legendary / legendary | H | derived |

**Integrity** (0 points, excluded from completion and from the Proof guard)

| key | name | description | trigger | rev |
|---|---|---|---|---|
| hand_in_the_cookie_jar | Hand in the Cookie Jar | Unlock the developer console. Nothing else will be earned while it is open. | dev unlock | S, crossed rim |
| touched_by_the_source | Touched by the Source | Load a save that carries the forge master's proof. | LoadGame with a valid elite proof | S |
| creative_accounting | Creative Accounting | Load a save that was edited outside the game. | LoadGame with a failed signature | S, crossed rim |

**Odds and ends**

| key | name | description | trigger | tier / rarity | rev | proof |
|---|---|---|---|---|---|---|
| portion_control | Portion Control | Lose a worker to starvation. | starvation death | bronze / common | S | bot (30 to 143 deaths per seed) |
| grumbling | Grumbling | Let morale fall below 40% and hear about it. | low morale warning | bronze / common | S | static |
| every_day_a_holiday | Every Day Is a Holiday | Hold 10 festivals in one run. | festivals in run, 10 | silver / uncommon | H | static (cooldown 300 ticks) |
| liquidation_sale | Liquidation Sale | Sell 100 buildings in one run. | sells in run, 100 | bronze / common | S | static |
| artisanal | Artisanal | Gather by hand 1,000 times in one run. | gathers in run, 1,000 | bronze / common | V | bot |
| floor_it | Floor It | Play at the highest speed your wonders allow. | speed equals MaxSpeedForAge | bronze / common | V | static |
| forgetful | Forgetful | Turn down an Ancient Memory. | memory declined | bronze / common | S | bot |
| deja_vu | Déjà Vu | Accept an Ancient Memory. | memory accepted | bronze / common | H | bot |
| belt_and_braces | Belt and Braces | Export an account backup. | account export | bronze / common | V | static |
| written_down_somewhere | Written Down Somewhere | Look at your recovery code. | recovery code shown | bronze / common | V | static |
| maxed_out | Maxed Out | Buy every tier of one prestige upgrade. | a prestige upgrade at MaxTier | silver / uncommon | V | static (about 3 runs of points) |
| fully_invested | Fully Invested | Buy every tier of every prestige upgrade. | all upgrades at MaxTier | legendary / legendary | V | static (about 25 runs of points at 30 / sqrt(level+1) per run) |
| lucky_break | Lucky Break | Get a legendary good epoch event. | epoch event good_legendary | silver / uncommon | V | luck; static existence |
| character_building | Character Building | Weather three challenging epoch events in one run. | challenging events in run, 3 | silver / uncommon | V | bot |
| everything_full | Everything Full | Have every capped resource at its storage limit at once, in the Industrial Age or later. | all capped resources at cap | gold / rare | H | static |
| old_bones | Old Bones | Start a run carrying ruins from a civilization that fell. | new run with ruins | bronze / common | H | static |

Specials by tier: legendary 9, platinum 8, gold 15, silver 19, bronze 21, integrity 3. Total 75, 1,970 points.

---

## 10. Milestone retune (the in-run layer)

Proposed numbers use the same ceilings (Appendix C and the population table); confirm each with the new guard.

| Milestone | Now | Proposed |
|---|---|---|
| stone_mason | 50 stone pits | 30 (ceiling 40) |
| temple_city | 50 temples | 25 (ceiling 31) |
| trade_empire | 30 posts + 12 quarters | 20 trade-lineage buildings of the Iron and Classical tiers standing in the Renaissance (lineage-wide, so upgrading does not forfeit it) |
| power_grid | 50 coal plants + 10 turbines | 20 coal plants + 10 turbines (ceilings 29 and 27) |
| population ladder | 5K, 50K, 500K, 10M, 100M, 1B, 10B | 250 (Primitive), 1,500 (Bronze), 5,000 (Classical), 15,000 (Renaissance), 50,000 (Industrial), 250,000 (Atomic), 1M (Digital): at most 40% of each age's housing ceiling, to be confirmed with bot population telemetry |
| build counts | 500, 2,000, 5,000, 20,000 | 250 by Bronze, 500 by Iron, 1,000 by Colonial, 2,000 by Modern, churn-proof |
| soldier counts | 5, 100, 250, 500, 2,000 | scale to FlowIncome: 100, 20K, 50K, 150K, 10M |
| the eleven "hold" milestones | a specific old tier | lineage-wide counts |
| stale MinAges | as listed in B9 | match the buildings' ages and the tech tree (derive tech-count gates from `TechsByAge`) |
| survivor, enduring_civilization | fixed MinTick | derive from `AgeTargets` (for example the sum of targets through Iron and through Colonial) |
| late production_all rewards | swallowed by the x3 cap | switch to uncapped effects (storage, tick speed, expedition reward, build cost), or leave `permanentBonuses` outside the cap |
| overlay | `catOrder` without epoch | use `config.MilestoneCategoryOrder()` |
| Stone Age | no age milestone | add one |

---

## 11. Phased plan

1. **Fix first (small, independent):** B1 (`Save()` by own ID, recover into its own slot and switch, broader guard), B4 (flush and stop on in-game switch and import, no global-ID swap in import), B5 (sticky tamper), B6 (overlay category), B3 (dev gate on theme unlocks, revert the local devmode change), B14.
2. **Guard:** land the feasibility test for milestones, then the milestone retune (section 10).
3. **Engine:** `BadgeDef` table, `Record` evaluator, run facts in the save, account v2 persistence and migration, drain-and-toast.
4. **UI:** badge case page, tier art grammar, glyph tiers, animation, toast.
5. **Content:** generated families, ladders and specials, each landing with its Proof; the new bot styles (rush, pacifist, autarky) and the smoke report columns (expeditions, deals, plan starts, peak population, production per resource).
6. **Docs:** `site/docs/achievements.md` (new), `account.md`, `commands.md`, `milestones.md`, `themes.md`, and the landing page stats, per the wiki sync rule.

---

## Appendix A. Lineage ladders (F6)

One run = Primitive to Atomic. C = sum of each tier's own-age storage ceiling. R = 0.10 C. Bronze is at least 3 to 5 for the small lineages.

| lineage | tiers by Atomic | ceiling C | one run R | bronze | silver | gold | platinum | legendary |
|---|---|---|---|---|---|---|---|---|
| culture_arts | 8 | 217 | 22 | 6 | 22 | 88 | 220 | 550 |
| energy | 4 | 103 | 10 | 5 | 10 | 40 | 100 | 250 |
| engineering | 10 | 260 | 26 | 7 | 26 | 100 | 260 | 650 |
| faith | 12 | 379 | 38 | 10 | 38 | 150 | 380 | 950 |
| food | 12 | 383 | 38 | 10 | 38 | 150 | 380 | 950 |
| geological_extraction | 12 | 395 | 40 | 10 | 40 | 160 | 400 | 1,000 |
| hacker | 0 | 172 (deep run, all tiers) | 17 | 5 | 17 | 68 | 170 | 430 |
| harbor | 2 | 51 | 5 | 3 | 5 | 20 | 50 | 130 |
| housing | 12 | 570 | 57 | 14 | 57 | 230 | 570 | 1,400 |
| knowledge | 12 | 376 | 38 | 10 | 38 | 150 | 380 | 950 |
| metallurgy | 9 | 227 | 23 | 6 | 23 | 92 | 230 | 580 |
| military | 12 | 353 | 35 | 9 | 35 | 140 | 350 | 880 |
| organic_extraction | 12 | 392 | 39 | 10 | 39 | 160 | 390 | 980 |
| storage | 12 | 325 | 33 | 8 | 33 | 130 | 330 | 830 |
| trade | 10 | 236 | 24 | 6 | 24 | 96 | 240 | 600 |

## Appendix B. Resource ladders (F8)

Production only. Rungs at 0.25, 1, 5 and 25 runs; deep-run resources start at silver. Data and nanobots unlock at Modern, but a run that prestiges on arrival produces none, so they count as deep-run resources.

| resource | unlocks | run | one run P | bronze | silver | gold | legendary |
|---|---|---|---|---|---|---|---|
| food | primitive | Primitive to Atomic | 1.09B | 270M | 1.1B | 5.5B | 27B |
| wood | primitive | Primitive to Atomic | 198M | 49M | 200M | 990M | 4.9B |
| knowledge | primitive | Primitive to Atomic | 1.18B | 290M | 1.2B | 5.9B | 29B |
| faith | primitive | Primitive to Atomic | 4.93M | 1.2M | 4.9M | 25M | 120M |
| stone | stone | Primitive to Atomic | 15M | 3.7M | 15M | 75M | 370M |
| iron | bronze | Primitive to Atomic | 810M | 200M | 810M | 4.1B | 20B |
| gold | bronze | Primitive to Atomic | 4.12T | 1T | 4.1T | 21T | 100T |
| coal | renaissance | Primitive to Atomic | 257B | 64B | 260B | 1.3T | 6.4T |
| soldiers | iron | Primitive to Atomic | 437M | 110M | 440M | 2.2B | 11B |
| marble | iron | Primitive to Atomic | 2.16M | 540K | 2.2M | 11M | 54M |
| iron_ore | iron | Primitive to Atomic | 39.5M | 9.9M | 40M | 200M | 990M |
| steel | medieval | Primitive to Atomic | 4.32T | 1.1T | 4.3T | 22T | 110T |
| culture | classical | Primitive to Atomic | 67M | 17M | 67M | 340M | 1.7B |
| oil | industrial | Primitive to Atomic | 214B | 53B | 210B | 1.1T | 5.3T |
| electricity | victorian | Primitive to Atomic | 3.02T | 750B | 3T | 15T | 75T |
| uranium | atomic | Primitive to Atomic | 145B | 36B | 140B | 720B | 3.6T |
| data | modern | Modern + Information | 3.78T | - | 3.8T | 19T | 95T |
| nanobots | modern | Modern + Information | 56.2M | - | 56M | 280M | 1.4B |
| titanium_ore | space | Space + Interstellar | 2.85B | - | 2.8B | 14B | 71B |
| crypto | cyberpunk | Cyberpunk + Fusion | 1.49M | - | 1.5M | 7.5M | 37M |
| dark_matter_crystals | cyberpunk | Cyberpunk + Fusion | 5.52B | - | 5.5B | 28B | 140B |
| plasma | fusion | Fusion + Space | 21.6Q | - | 22Q | 110Q | 540Q |
| titanium | space | Space + Interstellar | 63Q | - | 63Q | 310Q | 1,600Q |
| dark_matter | interstellar | Interstellar + Galactic | 582Q | - | 580Q | 2,900Q | 15,000Q |
| antimatter | galactic | Galactic + Quantum | 7,880Q | - | 7,900Q | 39,000Q | 200,000Q |
| quantum_flux | quantum | Quantum + Transcendent | 894Q | - | 890Q | 4,500Q | 22,000Q |

Old resources such as stone are mostly bought at the market late in the game, so their production ladders are small next to the gate costs (241B stone at Atomic). If the owner wants gate-sized numbers for those, add a separate "gained by trade" ladder rather than folding trades into production, which round-trip trades could farm.

## Appendix C. Age maximalist (F4)

| age | buildings | ceiling | threshold | age | buildings | ceiling | threshold |
|---|---|---|---|---|---|---|---|
| primitive | 6 | 306 | 229 | atomic | 13 | 343 | 257 |
| stone | 9 | 363 | 272 | modern | 16 | 369 | 276 |
| bronze | 10 | 359 | 269 | information | 15 | 415 | 311 |
| iron | 12 | 396 | 297 | digital | 15 | 376 | 282 |
| classical | 13 | 315 | 236 | cyberpunk | 14 | 363 | 272 |
| medieval | 13 | 286 | 214 | fusion | 14 | 290 | 217 |
| renaissance | 12 | 329 | 246 | space | 14 | 303 | 227 |
| colonial | 14 | 435 | 326 | interstellar | 14 | 327 | 245 |
| industrial | 17 | 483 | 362 | galactic | 14 | 365 | 273 |
| victorian | 13 | 387 | 290 | quantum | 14 | 346 | 259 |
| electric | 13 | 347 | 260 | transcendent | 4 | 36 | 27 |

## Appendix D. Domain staffing (F7) and population ceilings

| domain | staffing ceiling by Atomic | silver (10%) | gold (20%) |
|---|---|---|---|
| food | 2,656 | 270 | 530 |
| lumber | 2,720 | 270 | 540 |
| masonry | 2,728 | 270 | 550 |
| knowledge | 1,614 | 160 | 320 |
| faith | 1,564 | 160 | 310 |
| military | 2,722 | 270 | 540 |
| trade | 1,706 | 170 | 340 |
| engineering | 1,816 | 180 | 360 |
| metallurgy | 1,454 | 150 | 290 |
| energy | 762 | 76 | 150 |
| hacker | 2,609 (all ages; deep run) | 260 | 520 |

Housing capacity with every housing building maxed at its own age's storage (upper bound on population): Primitive 620, Stone 1,970, Bronze 4,620, Iron 8,160, Classical 12,700, Medieval 21,400, Renaissance 41,000, Colonial 90,000, Industrial 188,000, Victorian 380,000, Electric 733,000, Atomic 1.42M, Modern 2.71M, Information 5.42M, Digital 10.1M, Cyberpunk 18.4M, Fusion 32.7M, Space 61.2M, Interstellar 140M, Galactic 297M, Quantum 612M, Transcendent 942M.

---

## Addendum (2026-10-01, owner requests)

- **Integrity badge art.** Each of the three integrity badges gets a hand-drawn 9×17 sprite with its own glitch. Prototypes are in the Badge Lab (https://claude.ai/artifact/RbX37vTcFi33sqrkBR4F3d).
  - **Hand in the Cookie Jar:** a glitched cookie jar. The label flips to C00K1E5, sudo nom or a /dev cursor, and a red strike flickers across it.
  - **Creative Accounting:** a cyber-glitch ledger with impossible sums, cyan and magenta row tears, and a flickering VOID stamp.
  - **Touched by the Source:** Matrix code rain inside a SOURCE ring that resolves into the forge master's key.
  - Every frame is a pure function of (badge, frame), with noise drawn from a hash.
- **Hacker themes.** The owner chose both:
  - **Source** (phosphor green with code rain) unlocks with Touched by the Source.
  - **Glitch** (cyan and magenta on violet black, with screen tears) unlocks with Creative Accounting.
- **New theme palettes.** Ashfall, Ledger and Prismatic now have palettes in the Badge Lab. Themes may carry optional ambient effects (rain, embers, tears, a turning accent) drawn in empty cells only. The motion setting switches them off.
- **New secret special, "We Are Not Alone"** (silver, secret): inspect an alien visitor on the map. The roguelike movers PR exposes the alien's kind in the inspection result as the hook for this badge.
