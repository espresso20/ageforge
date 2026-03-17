# Changelog

All notable changes to AgeForge are documented here.

---

## [Unreleased]

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
