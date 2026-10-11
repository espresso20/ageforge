# 🏛 AgeForge Wiki

> A terminal idle civilization game: 22 ages in 7 epochs, from Primitive survival to a Transcendent civilization.

> **Version note:** this wiki describes version 4.0, the game on the master branch, which is not released yet. Until it is, the downloads are the 3.6 release, which still has some older systems (for example the `speed` command and the old city map).

AgeForge runs full-screen in your terminal as a single binary, built with Go and tview/tcell. You steer by typing commands at a prompt, and your civilization keeps working between them, including while the game is closed.

---

## At a glance

- **22 ages in 7 epochs**, from the Primitive Age to the Transcendent Age.
- **About a week for a first run**: that is real time, to the Modern Age and a full run's prestige. The Primitive Age takes about 15 minutes and the Stone Age about 45; from the Bronze Age on, each age takes hours. The game is paced for checking in a few times a day.
- **Offline progress**: when you load a save, the game credits the time since you saved, up to 24 hours at 50% of your production. Construction and research finish, the build plan runs and your workers keep staffing as that time passes.
- **Workers that staff themselves**: workers arrive on their own while housing and food allow (auto-recruit) and go to work by your roster across 12 worker domains. You can still recruit and assign by hand.
- **The build plan**: queue up to 60 builds, techs, trades and an advance. The game starts each one as the resources come in, while you play and while you are away.
- **301 buildings**: a 14-lineage production system plus storage, monuments and 22 wonders, one per age. You need each age's wonder to advance.
- **26 resources**, from food and wood up to Quantum Flux (soldiers included), and **202 technologies** in prerequisite chains.
- **The Map**: your empire drawn from your real buildings, in two styles: a roguelike glyph world with three zooms, and a side-on skyline with one district per age. An inspect cursor gives you the command for whatever it is on, and `minimap on` puts a mini map on the dashboard.
- **Catastrophes and harbingers**: from the Iron Era on, an era may hide a fated catastrophe. A harbinger always comes to warn you first, and when the doom strikes you choose to Endure or Succumb.
- **Trade**: a market exchange, 21 trade routes, harbors and a black market.
- **Factions**: an 11-civilization roster to meet, with opinion, diplomacy, trade deals, boons and setbacks, embassies and war.
- **The army**: soldiers defend you against raids and fight campaigns. There are 16 expeditions in all (scouting missions and military campaigns), each a gamble of cost against reward.
- **77 milestones** in 6 chains, with civilization titles.
- **Prestige** from the Medieval Age on: start over with prestige points, which pay more the deeper the run went, to spend on 3 legacy kit items that carry your plan, roster and the civilizations you met into every later run.
- **Era Mastery**: every age a run completes runs faster on your later runs, twice as fast after one completion and up to 4.2x after ten (4x at once for ages 6 or more behind your record), so later runs reach the Modern Age much sooner than the first.

---

## Quick navigation

| Section | Pages |
|---|---|
| Getting Started | [Installation](getting-started.md) · [Your First Age](first-ten-minutes.md) · [How to Play](how-to-play.md) |
| Core Loop | [Resources](resources.md) · [Workers](workers-and-domains.md) · [Buildings](buildings.md) · [The Build Plan](plan.md) · [Technologies](technologies.md) · [The Map](map.md) |
| Progression | [The 22 Ages](ages.md) · [Wonders](wonders.md) · [Milestones](milestones.md) · [Epochs](epochs.md) · [Prestige](prestige.md) |
| Eras and Dooms | [Catastrophes](catastrophe.md) · [The Harbinger](harbinger.md) · [Faith](faith.md) |
| The World | [Trade](trade.md) · [Factions & Diplomacy](factions.md) · [Army & Missions](military.md) · [Events](events.md) |
| Systems | [Morale](morale.md) · [Knowledge](knowledge.md) |
| Reference | [All Commands](commands.md) · [Civilization History](history.md) |
| Your Game | [Saving & Loading](saving-and-loading.md) · [Accounts](account.md) · [Badges](badges.md) · [Themes](themes.md) |

New to the game? Install it, then follow [Your First Age](first-ten-minutes.md).

---

## Controls

| Key | Action |
|---|---|
| Type a command, then `Enter` | Run it. Commands such as `research`, `trade`, `logs` and `help` open a panel. See [All Commands](commands.md) |
| `Tab` or `→` | Take the completion shown in dim text after the cursor (`→` only at the end of the line). Press `Tab` again for the next suggestion, `Shift+Tab` for the previous one |
| `↑` / `↓` | Step through your command history |
| `PgUp` / `PgDn` | Turn the Buildings list a page when no panel is open (in the Map, they zoom) |
| `Esc` | Close the open panel. With no panel open: save, stop the game and return to the [main menu](how-to-play.md#the-main-menu), where Continue opens the game again |

Type commands at the `>` prompt at the bottom of the screen.
