# Changelog

All notable changes to AgeForge are documented here.

---

## [Unreleased]

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
