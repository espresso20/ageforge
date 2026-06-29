# Changelog

All notable changes to AgeForge are documented here.

---

## [Unreleased]

### Added
- **Humor & personality layer (buildings + ages)** — buildings now carry an optional cosmetic *flavor* line, rendered as a dim italic note beneath the functional description in both the Buildings overlay and the Economy tab (costs, rates, and worker counts are untouched — flavor is purely additive). All 300 buildings got a line, age-appropriately voiced: primitive-age field notes ("Four walls and the ambition of someday having five.") drifting toward late-game existential dread ("Stores resources in a superposition of 'we have it' and 'oh no.'"). Every one of the 22 ages also gained a one-liner shown on the age-advance splash beneath the formal title, scaling from caveman observation to cosmic dread. Events, milestones, and log lines get the same treatment in a follow-up.
- **Full Diplomacy System — Civilization Encounters** — the six NPC factions become the backbone of an **11-civilization roster** spanning every epoch. Each civ has a **personality** (peaceful / aggressive / mercantile / isolationist), a **backstory**, and a **strength** rating. Civilizations are met through **first-contact events** (age/epoch-gated; the founding civs appear in the early eras, more each epoch). Opinion now **drifts by personality** — aggressive civs trend hostile, peaceful trend friendly, mercantile warm to your trade activity, isolationists hover near neutral. **Peaceful, high-opinion civs lend you workers** (temporary loans, or permanent above opinion 80). **War** is event-driven (no combat): a civ declares war only when its opinion is below -75 *and* a provocation threshold is crossed (raid their trade route via `diplomacy raid`, or embargo them — two provocations trips it), then launches periodic resource **raids** scaled to its strength. Make peace with `diplomacy tribute <civ>` (gold + culture) or by waiting them out. The Diplomacy overlay now shows personality, backstory, war banners, and lent-worker status; new civ state (discovery, opinion, war, provocations, lent batches) persists in saves and resets per prestige run.
- **Embassy buildings & Diplomacy overlay** — two new diplomacy buildings: the **Embassy** (Colonial Age, costs gold + iron) and the **Grand Embassy** (Industrial Age, costs gold + steel, twice the rate). Staffed with trade-domain workers, they passively generate opinion each tick toward all of your non-hostile factions (scaled by worker fill, capped at +100). A new **Diplomacy overlay** — opened by typing `diplomacy` (or `dip`) with no arguments — lists every civilization with opinion bars, color-coded status, the active trade-rate bonus, a threshold indicator (e.g. *+8 to friendly*), and the actions available per civ; `diplomacy <action> <civ>` still performs the action directly.
- **Theme system** — nine switchable themes with a live picker (preview + palette swatches) and a `theme` command. Forge (default) plus three accessibility themes — Deuteranopia-safe, Protanopia-safe, High Contrast — unlocked from the start; five flavor themes (Parchment, Bronze, Cyberpunk, Monochrome, Cosmic) unlocked by reaching milestones. Active theme and unlocks persist per account.
- **Multiple accounts** — each account gets its own data slot (`data/accounts/<id>/`), with a one-time, non-destructive migration of existing data. A start-screen Accounts panel lists every account and lets you switch, create, export, import, back up, recover, and wipe. Name-derived identity with `AGEF-` recovery codes; signed, ID-bound export/import that lands in its own slot and can't clobber another account; lifetime stats and achievements.
- **Account backups** — a full-slot snapshot (account.json + saves) is taken before a wipe, on export, and on demand (`account backup` or the panel), keeping the newest ten; plus a one-time pre-migration snapshot of the old data layout.
- **Save lineage and Load Game browser** — procedurally-named saves you can branch into new lines, shown as a lineage tree; the browser adds delete / rename / duplicate, richer save metadata, and account attribution.
- **Multiplier resolver** — a single engine resolver that every bonus source emits into; the Active Multipliers panel renders from its breakdown, color-coded by sign with every contributing source shown.
- **Morale rework** — morale is now a managed two-way resource with restoring buildings, a continuous curve, a history graph, and banded displays; faith production rate lifts morale.
- **Expeditions and Army, split** — scouting Expeditions and military Campaigns are now separate systems; soldiers are a real produced/stored resource, with concurrent per-category expedition slots.
- **Help panel** — `help` opens a panel instead of dumping inline text.
- **Main-screen UI overhaul** — a framed command bar, a scannable ✓/✗ age-progress strip, a cleaner log, lifted secondary-text contrast, an early-game onboarding panel, and rebalanced panel widths.
- Nanobots are a real resource now — a Modern-age producer building (Nano Foundry, +80 nanobots/tick), 3 nanobot techs (Nanofabrication cuts build costs −8%, Medical Nanobots boosts population and food, Self-Replication ramps nanobot output), and several digital/cyberpunk buildings now cost nanobots to construct.
- Culture has sinks now — Cultural Monuments (spend culture for a permanent production bonus) and a `festival` command (spend a lump of culture for a temporary multi-resource buff); prestige gates remain the primary long-term sink.
- **Trade System Expansion** — six new trade routes fill the colonial → industrial gap (`mercantile_convoy`, `triangular_trade`, `tea_clippers`, `coal_barges`, `cotton_exchange`, `steamship_line`), bringing the total to **21**. A new **harbour lineage** (`harbor` → `harbor_authority` → `seaport` → `container_terminal` → `logistics_hub`, Colonial → Digital, trade-domain workers) boosts the income of **every active trade route** (+5% to +25% per tier, stacking). **Trade disruption** ties into diplomacy: routes whose imports include a resource specialised in by a civilization you're **at war with or have embargoed** earn nothing until the conflict ends — flagged in the log and the Trade overlay, and resuming automatically on peace. A new **black market** (`blackmarket` / `trade black <resource>`, Colonial Age+) lets you spend a lump of **culture** on a high-risk deal: a 55% chance of a large resource haul (2.5× the stake) or losing the culture, on an ~8-minute cooldown. The Trade chain capstone **Maritime Empire** extends the chain to **6** milestones.
- **Age Awakenings** — a one-time deterministic boost fires the first time you enter each epoch's signature age (7 epochs, 7 awakenings): Pottery Mastery, Discovery of Metallurgy, Steam Breakthrough, Electrification, Information Age Dawns, Cybernetic Awakening, and First Contact Signal. Each grants a modest temporary production bonus, fires at most once per prestige run, persists across save/load, and resets on prestige so the next run can earn them again.
- **Ancient Civilization Memory** — early in a new prestige run (Primitive or Stone age), a ~40% chance offers a cache of your now-extinct predecessor: accept to research one random age-appropriate tech free of prerequisites, age gate, and knowledge cost — but at half research speed (2× ticks). One cache per run; declining or reloading won't re-roll it, and the reachable tech tier scales with prestige level (one extra age of reach per two levels). Requires prestige level ≥ 1 and resets each run.

### Balance
- Flattened building cost curves and raised age-advancement requirements.
- Capped age-transition carryover to a starter head-start.
- `gather` yield raised 10 → 25 and disabled past the Medieval age.
- Storage lineage: every tier stays affordable to its build cap — the cost of copy N never outruns the storage those copies provide (the old stash-deadlock class of bug, lineage-wide). Storage buildings cap at 25 copies (stash 50), and several under-provisioned vault caps were raised. Guarded by a regression test.
- Milestone chains rebalanced: completion speed-boosts normalized across all six chains (Military was 18× weaker than Settlement — now in line); the Trade chain expanded 3 → 6 milestones (5 via the milestone revamp, then a sixth — Maritime Empire — with the Trade System Expansion); and Military/Trade/Scholar capstones now pay out broad `production_all` instead of domain-only bonuses, so finishing any chain helps your whole economy.

### Fixed
- Theme picker no longer deadlocks the app on arrow/Esc.
- Procedural save names: flattened the distribution (no more "Grand Duchy" clustering) and added ~10× more names across every bank.
- Account Wipe modal shows the exact name to type and has a cleaner confirm UI.
- Save-name modal: readable input contrast, opaque background, sized to content.
- Load Game: bare `load` opens the browser; renaming re-parents child saves; footer hotkeys render as keycaps.
- Modifiers: negative production debuffs apply correctly and build-cost modifiers are wired in.
- Stash is buildable again (its first-copy cost had exceeded the base wood cap).

---

## [v3.6.4] — 2026-03-25

### Added
- add buildings overlay showing full age history

### Fixed
- reconstruct pending upgrades on load so 'upgrade' works after save/reload
- correct age advancement requirements to reference current-age buildings
- city map shows all-age buildings rendered in current age style
- lock build command to current age only
- economy tab Buildings panel only shows current age buildings

---

## [v3.6.3] — 2026-03-24

### Fixed
- resolve undefined variable compile errors in overlay_stats (#32)

---

## [v3.6.2] — 2026-03-24

### Fixed
- expose permanentBonuses in GameState and fix Active Multipliers display (#27)
- correct energy and metallurgy lineage production rates on upgrade (#28)
- city map only renders buildings from the current age (#29)
- unique 16x16 sprites for all 22 age wonders (#30)
- wonder multiplier bonuses now appear in Active Multipliers (#31)

---

## [v3.6.1] — 2026-03-24

### Fixed
- embed assets/maps and assets/sprites into binary

---

## [v3.6.0] — 2026-03-24

### Added
- map rendering experiments (mapv1 / mapv2 / mapv3) (#26)

---

## [v3.5.0] — 2026-03-23

### Added
- player-driven building upgrade system (#24)

### Fixed
- prevent electricity regression when upgrading energy lineage tiers 4 and 9 (#25)
- resolve CSP violation on changelog page

---

## [v3.4.0] — 2026-03-23

### Fixed
- changelog page — fetch CHANGELOG.md instead of GitHub Releases API (#23)
- add morale panel to workers overlay (#22)

---

## [v3.3.0] — 2026-03-18

### Added
- morale system — worker output multiplier (#21)
- declutter main screen — workers/wonder panels to overlays (#20)

---

## [v3.2.5] — 2026-03-17

### Added
- wonder completion required to advance ages (#19)

### Fixed
- remove full-screen dev console, route dev commands through main input
- age splash keypress regression and faith preserved on age advance

---

## [v3.2.4] — 2026-03-17

### Added
- dynamic overlay width + fix all providers to accept terminal size
- civilization history overlay with braille line graphs

---

## [v3.2.3] — 2026-03-17

### Fixed
- building cost scaling now correct for batch, max, and queued builds

---

## [v3.2.2] — 2026-03-16

### Fixed
- three UI display fixes — wonder rates, building names, speed multiplier

---

## [v3.2.1] — 2026-03-16

### Fixed
- four config bug fixes — event collision, tech rates, cost cliff, titanium age

---

## [v3.2.0] — 2026-03-16

### Fixed
- age advancement splash ignores keypresses until auto-dismiss

---

## [v3.1.0] — 2026-03-11

### Fixed
- wiki update and fix exponential worker growth bug

---

## [v3.0.0] — 2026-03-09

### Added
- expand 33 → 74 milestones and harden all thresholds
- fix map building scale, density throttle, and 249-building sprite coverage
- revamp map system with age-appropriate city layouts and building sprites
- add research as a panel view and command option
- expressive resource bars with fill-level colour and status glyphs
- implement worker_loss effect + ended message damage summary
- worker panel grouping, map fix, stats rate breakdown
- replace F-tab system with command-summoned overlay panels
- purple fill color for all progress bars
- rewrite worker panel with per-domain assignment groups
- add floating overlay framework + milestones panel
- collapse 12 domain pools into single generic worker pool
- simplify assign/unassign/recruit command API
- rename Villager→Worker across UI layer
- rename Villager→Worker across game core and engine
- strip legacy alias system from WorkerManager
- comprehensive autocomplete revamp
- population screen redesign
- add command history ring buffer
- rewrite autocomplete to use domain keys
- Phase 13 — epoch-exclusive regular events
- Phase 11e+11f — Epoch tab (F10) + Stats tab epoch/legacy fields
- Phase 11c — age advance modal transformation summary + epoch reveal
- Phase 11a+11b — economy tab worker display + villager panel rewrite
- Phase 11d — culture progress bar + faith threshold indicator
- Phase 10 economy redesign — full 13-lineage building content overhaul
- implement catastrophe system (Phase 9 economy redesign)
- implement epoch system (Phase 8 economy redesign)
- Phase 7 economy redesign — age transition building transformation pass
- Phase 6 economy redesign — worker-building coupling engine
- Phase 5 economy redesign — config foundation data structures
- update Savefile system with a new configuration
- add logo to website nav, hero, footer, and og:image

### Fixed
- refactor changes for wiki and add updated screenshots
- remove all F-key references, fix commands at a glance and controls
- radical threshold hardening + fix blank progress bars
- fix milestone progress flickering, missing categories, and autocomplete gaps
- correct military lineage tier order, bad upgrade, and wonder description
- research_speed permanentBonus now actually reduces research duration
- restore building scale=2, remove primitive-era wonder circle artifact
- fix map city circular blob — remove anchor-walk clustering and expand zone
- fix map building shadow artifact and add era-aware city label
- wire master_builder and scholars_haven to correct game state
- worker panel always resolves class name for any domain+age
- Scholar's Haven now requires 5 knowledge workers assigned
- war_machine no longer fires in primitive age
- capture ESC on overlay TextView to close panel
- recruit log shows worker(s) not domain key; fix dev tab focus
- assign/unassign gracefully skip legacy domain prefix arg
- recruit autocomplete shows class names; save-compat; class name resolution
- worker alias, wood camp timing, UI panel layout
- Phase 17a — stone_camp and hunting_lodge age gating
- Phase 14 — worker assignment, gather cap, domain key alignment
- remove dashboard bleed-through on wipe confirmation
- three ticker freeze bugs
- resolve save directory relative to binary, not CWD
- fix docsify security for mobile use

### Balance
- full age pass renaissance→quantum — rates, times, food costs
- city spread ratio sized to 20
- medieval age pass — rates, build times, food cost
- classical age pass — rates, build times, food cost
- iron age pass — rates, build times, food cost
- bronze age pass — rates, build times, food cost
- stone age pass — buildings, rates, costs
- wood camp rate up, food drain nudged higher
- stash cost up, gathering camp cheaper, food drain reduced
- Phase 17b — fix primitive age economy (food rate vs drain)
- game balances, ages and research

### Changed
- delete dead tab_*.go files, migrate helpers to overlays
- simplify recruit command — remove domain arg
- replace DefaultWorkerTypes usage with live worker state
- remove all backward-compat shims
- remove assign/unassign backward-compat shims
- split buildings_new*.go into per-lineage files

### Other
- Change floor value for buildings to NOT produce if no workers in the building

---

## [v2.5.2] — 2026-03-02

### Fixed
- package declaration typo in config/buildings.go (confi → config)

---

## [v2.5.1] — 2026-03-02

### Fixed
- clarify bullet point prompt wording in commit helper
- commit helper adds balance type, length enforcement, and bullet body

### Changed
- stash now has max count of 50, all buildings in primitive take longer to build, altar production raised from .004 to .008 knowledge

---

## [v2.5.0] — 2026-03-02

### Added
- manual age advancement — type 'advance' when ready
- add interactive conventional commit helper (make commit)

---

## [v2.4.7] — 2026-03-02

### Added
- show version + async update badge

---

## [v2.4.6] — 2026-03-01

### Fixed
- rich Discord embeds + fix blank release notes

### Other
- patch release notes

### Fixed
- release workflow: awk now skips [Unreleased] section and targets versioned entries only — release notes no longer blank
- release workflow: removed redundant github-release-to-discord.yml (GITHUB_TOKEN releases don't trigger release: published in other workflows)
- release workflow: cleaned up dead commented-out SethCohen job
- discord notification: rich embed with per-section fields (Added/Fixed/Changed/Other), thumbnail, timestamp, and download link

---

## [v2.4.5] — 2026-03-01

### Other
- new release notes updates

---

## [v2.4.4] — 2026-03-01

### Fixed
- replace Python heredoc with jq+curl for Discord notification

---

## [v2.4.3] — 2026-03-01

### Fixed
- Discord notify step - pass values via env vars not heredoc interpolation

---

## [v2.4.2] — 2026-03-01

### Added
- screenshot lightbox on click + smaller cards showing 2 at a time

### Fixed
- move Discord notification into release.yml as a final step

### Other
- update screen shots mechanism
- new screenshots

---

## [v2.4.1] — 2026-03-01

### Fixed
- explicitly mark GitHub releases as published to trigger Discord webhook
- screenshot caption renders below image, not overlapping it

---

## [v2.4.0] — 2026-03-01

### Added
- add auto-discovering screenshots carousel to site

### Fixed
- show correct precision for small resource rates in economy tab
- build max now correctly queues multiple buildings with a MaxCount
- update screenshots section heading and subtitle

### Other
- add webook url for discord connect
- add screenshot

---

## [v2.3.0] — 2026-03-01

### Added
- auto-generate release notes from commits + changelog page on site

---

## [v2.2.0] — 2026-03-01

---

## [v2.1.0] — 2026-03-01

---

## [v2.0.0] — 2026-03-01

---

## [v1.1.0] — 2026-02-27

---

## [v1.0.0] — 2026-02-26

### Added
- 22 playable ages from Primitive to Transcendent
- 80 buildings (58 standard + 22 wonders) with scaling costs
- 52 researches with prerequisites and permanent bonuses
- 33 milestones across 5 chains with titles and speed boosts
- 21 resource types with rates, storage caps, and breakdowns
- 8 villager types with food drain and assignment system
- 15 military expeditions with risk/reward mechanics
- 28 random events with sentiment streaks and timed effects
- 15 trade routes with supply/demand pressure
- 6 diplomacy factions with opinion and status tracking
- Prestige system with 9 upgrades across 5 tiers
- In-game wiki server (port 7891) with 10 reference pages
- 9 UI tabs: Economy, Research, Military, Trade, Stats, Wiki, Map, Wonders, Logs
- Save/load system with multiple named slots
- Full keyboard navigation and command parser

---

## [1.0.1] — 2025-02-19

### Fixed
- Minor balance adjustments to early-game resource rates
- Hut cost and pop cap rebalanced for smoother ramp

---

## [1.0.0] — 2025-02-18

### Added
- Initial public release
- Core idle loop with tick-based resource production
- Primitive and Stone Age content
- Basic building queue and villager assignment
