# The AgeForge Compendium

> A terminal idle civilization game: 22 ages in 7 epochs, from the Primitive Age to the Transcendent Age.

> Version note: this Compendium describes version 4.0, the game on the master branch, which has no release yet. The downloads are version 3.6, which has the `speed` command and the old city map. To play version 4.0, build it from source (see [Installation](getting-started.md)).

AgeForge runs full-screen in your terminal as a single binary, built with Go and tview/tcell. You steer by typing commands at a prompt. Your civilization works on its own between commands, and when you open a save again the game credits the time you were away.

---

## At a glance

- 22 ages in 7 epochs, from the Primitive Age to the Transcendent Age.
- Offline progress: when you load a save, the game credits the time since you saved, up to 24 hours at 50% of your production. Construction and research finish, the build plan runs and your workers keep staffing as that time passes.
- Workers staff themselves. They move in toward your housing while food allows (auto-recruit), and the ones without a job wait in the reserve and go to work by your roster across 11 worker domains. You can still recruit and assign by hand.
- The build plan holds up to 60 items (builds, upgrades, trade deals and an advance), plus any research you queue. The game starts each item as the resources come in, while you play and while you are away.
- 302 buildings: 251 in 14 production lineages, plus storage buildings, monuments, standalone buildings and 22 wonders, one per age. You need an age's wonder to advance. Each storage building adds 1% to all production.
- 24 resources, from food and wood up to Quantum Flux, and 203 technologies in prerequisite chains. Scholars come with your knowledge buildings and do the research.
- The Advance panel (`charter`) shows what the next age asks for as four pillars: Knowledge, People, Industry and Monument. The same figures sit in the Ledger at the top of the dashboard.
- The Map draws your town from your real buildings in two styles: a roguelike glyph world with three zooms, and a side-on skyline with one district per age. An inspect cursor puts the command for whatever it is on in the prompt, and `minimap on` puts a mini map above the Buildings list.
- Catastrophes and harbingers: from the Iron Era on, an era may hold a fated catastrophe. A harbinger warns you before it strikes, and when it does you choose to Endure or Succumb.
- Trade: 21 trade routes, harbors and a Black Market where culture buys the same worth in any resource, five times in six.
- Factions: an 11-civilization roster to meet, with opinion, diplomacy, trade deals, boons and setbacks, embassies and war.
- The army: soldiers come with your military buildings, soften raids and march with every mission. The Expeditions panel shows the mission board, with short, long and grand missions that are each a gamble of cost against reward.
- 77 milestones in 6 chains, with civilization titles.
- Prestige from the Medieval Age on: start over with points that become a lasting production bonus in every age. Plan Template, Roster and Old Friends carry over after your first prestige.
- Era Mastery: when a run ends (by a prestige or a Succumb), each age it finished runs faster in later runs, 2x after one completion and up to 4.2x after ten. An age six or more ages behind your record runs at 4x at once.
- Every panel opens full screen. VibeMode (`vibemode on`) swaps the game's words for a second script in GenZ slang.

---

## Quick navigation

| Section | Pages |
|---|---|
| Getting Started | [Installation](getting-started.md) · [Your First Age](first-ten-minutes.md) · [How to Play](how-to-play.md) |
| Core Loop | [Resources](resources.md) · [People](workers-and-domains.md) · [Buildings](buildings.md) · [The Build Plan](plan.md) · [Technologies](technologies.md) · [The Map](map.md) |
| Progression | [The 22 Ages](ages.md) · [The Advance Panel](advance.md) · [Wonders](wonders.md) · [Milestones & Chains](milestones.md) · [Epochs](epochs.md) · [Prestige](prestige.md) |
| Eras and Dooms | [Catastrophes](catastrophe.md) · [The Harbinger](harbinger.md) · [Faith](faith.md) |
| The World | [Trade](trade.md) · [Factions & Diplomacy](factions.md) · [Army & Missions](military.md) · [Events](events.md) |
| Systems | [Morale](morale.md) · [Scholars and Research](knowledge.md) |
| Reference | [All Commands](commands.md) · [Civilization History](history.md) |
| Your Game | [Saving & Loading](saving-and-loading.md) · [Accounts & Recovery](account.md) · [Badges](badges.md) · [Themes & Accessibility](themes.md) |

New to the game? Install it, then follow [Your First Age](first-ten-minutes.md).

---

## Controls

Type commands at the prompt at the bottom of the screen.

| Key | Action |
|---|---|
| Type a command, then `Enter` | Run it. Commands such as `research`, `trade`, `logs` and `help` open a panel. See [All Commands](commands.md) |
| `Tab` or `→` | Take the completion shown in dim text after the cursor (`→` only at the end of the line). Press `Tab` again for the next suggestion, `Shift+Tab` for the previous one |
| `↑` / `↓` | Step through your command history (in a text panel with the prompt empty, they scroll it) |
| `PgUp` / `PgDn` | Turn the Buildings list a page when no panel is open (in the Map, they zoom; in a text panel, they scroll it). `Option+↑` / `Option+↓` do the same on a keyboard with no Page keys, and `Option+←` / `Option+→` are `Home` / `End` |
| `Ctrl+R` | Show the next page of the Resources box, when it has more rows than fit |
| `Esc` | Close the open panel. With no panel open: save, stop the game and return to the [main menu](how-to-play.md#the-main-menu), where Continue opens the game again |

Every panel opens over the whole screen, and the prompt keeps working in most of them. The Wonders, Harbinger and Plan panels take letter keys, so press `Esc` before you type a command there.
