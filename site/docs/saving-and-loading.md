# Saving & Loading

AgeForge keeps your civilization in named saves, with a regular autosave and a save browser. This page covers where saves live, the commands that manage them, which game the main menu's Continue opens, how to read the Load Game browser, and how to delete saves.

---

## Starting a new game

When you choose **New game** on the [main menu](how-to-play.md#the-main-menu), the game asks you to name your civilization and fills in a random name. Press **Enter** to accept it, type your own, or press **Tab** for a new suggestion. That name becomes your active save, and the game autosaves into it from then on.

---

## Where saves live

Save files go in **your active account's** saves folder, `data/accounts/<account_id>/saves/*.json`, inside the `data/` folder next to the `ageforge` binary. There is one file per save. The `save`/`load` commands and the **Load Game** browser all read and write the same files, so a save made one way shows up the other.

Saves are **per-account**: each account keeps its own `saves/` folder, so switching accounts changes which saves you see, and one account's saves never mix with another's. See [Account & Recovery](account.md) for how accounts and their slots are laid out.

> **Saves from an older version:** older builds kept a single flat `data/saves/` folder. On first launch the game moves those saves into your account's `data/accounts/<id>/saves/` slot. Nothing is deleted, and you don't have to do anything.

> **Saves from before the research changes:** a game saved before wonders needed a keystone tech, or before some commands waited for a tech, loads as usual. Everything it researched and built stays. In the age it was in nothing new is locked, and any command it was already using (a campaign under way, a route running, the black market) stays open for the rest of that run. The log says so once on the first load. Its techs now give what today's techs give: smaller bonuses that no cap holds back (see [How Tech Bonuses Stack](technologies.md#how-tech-bonuses-stack)).

> **Saves from before the one-week pacing:** a game saved before a first run took about a week loads as usual and carries on where you left it. On that first load the log says once that from the Bronze Age on each age, and its timers, now run about 2.6 times as long. Timers already running in the save finish at their old length.

---

## Commands

| Command | Description |
|---|---|
| `save` | Opens an **Overwrite / Branch** prompt for your current run (see below) |
| `save <name>` | **Branches** a new save with that name off your current run; autosave then follows it |
| `load` | Opens the **Load Game** browser (your save tree) to pick which save or branch to load |
| `load <name>` | Loads that named save directly |
| `saves` (or `save list`) | List all save files |
| `Esc` | Closes the open panel. With no panel open, saves to your active save, stops the game and returns to the main menu |

```
save           # prompt: Overwrite this run, or Branch a new save?
save hero      # branch a new save named "hero" off the current run
load           # open the Load Game browser to pick a save or branch
load hero      # load the save named "hero" directly
saves
```

A bare `load` (no name) **opens the Load Game browser**, the tree of every save, so you choose which save or branch to load. You can open it mid-game; pressing `Esc` in the browser returns you to your current run without loading anything. `load <name>` skips the browser and loads that save directly.

---

## Active save slot

The game remembers the last slot you saved to or `load`ed by name. That is your **active** save.

- Until you've named a save or loaded one this session, the active slot is `autosave`.
- Once you branch to `hero` or `load hero`, the active slot is `hero` and autosave follows it.

The periodic autosave and the save on `Esc` both write to your **active** save, overwriting it each time. Your current game *is* the autosave; there is no separate hidden slot.

---

## Branching your save

A bare `save` opens a prompt with two choices:

- **Overwrite** writes your current run to the active slot now (what autosave does, on demand).
- **Branch new** starts a new save. You get a generated name (edit it, press **Tab** for another, **Enter** to confirm). The new save's *parent* is the save you branched from, and **autosave switches to the new branch**, so the old save stays exactly as it was at the branch point.

`save <name>` skips the prompt and branches straight to that name.

Branching keeps a moment without stopping play: branch before a prestige, a risky catastrophe, or any decision you might want to revisit. Your old save stays as it was while you keep playing on the new branch. Branched saves are ordinary files and appear in the `saves` list and the Load Game browser alongside everything else. Save names must be valid and unique; branching to a name that's already taken is refused.

---

## Autosave

The game autosaves every 60 seconds to your **active** save, and saves again when `Esc` takes you back to the main menu, so the file on disk is never more than about a minute behind your game. To keep a point you don't want overwritten, save it under a new name (or duplicate it with `c` in the Load Game browser).

---

## Continue and your main game

**Continue** on the main menu opens your **current game**, and the menu draws that game's town behind it. Which save that is follows three rules, in order:

1. **Your main game**, if you marked one. Highlight a save in the Load Game browser and press `m`. It stays the current game, whatever else you play, until you press `m` on it again or delete it.
2. Otherwise, **the game you played last**: the save you last loaded or saved on this account.
3. If neither is on record, **the save written most recently**. This is what an account from an older version of the game gets until you play once, and what you get if the save on record is gone.

A save that cannot be read is never the current game. An account with no save has no Continue.

The Load Game browser shows which save Continue opens: its subtitle names it, its row carries `◆ main game` or `▸ continue`, and its detail pane says why. Renaming a save keeps it your main game; deleting it clears the mark.

Both records are kept in the account's `settings.json`, beside your display settings (see [The settings file](account.md#the-settings-file)). They are part of the account, not of any save, so each account has its own current game.

---

## The Load Game browser

Choosing **Load Game** from the main menu, or typing a bare `load` mid-game, opens a browser that lists every save belonging to your **active account** (under `data/accounts/<id>/saves/`). Highlighting a save fills a **detail pane** with what you need to judge it before loading: its age and epoch, population, buildings, wonders, milestones, techs, soldiers, prestige, [morale](morale.md), which account made it (*this account* / *another account* / *pre-account*), a ⚠ warning if a catastrophe is pending, the exact save time, and for branched saves a **Branched from** line naming its parent. Opened mid-game, `Esc` returns you to your current run without loading anything.

**The save tree.** Saves are arranged as a **tree**, not a flat list. When you [branch](saving-and-loading.md#branching-your-save) a new save off your current run, it appears **indented beneath its parent** with tree connectors (`├─`, `└─`), so you can see which saves descend from which. Top-level saves (ones you started fresh, plus any orphans) are ordered most recent first, and so are each parent's children.

A **● active** marker shows which save your game is autosaving into, and **◆ main game** or **▸ continue** marks the save the main menu's Continue opens (see [Continue and your main game](#continue-and-your-main-game)).

If a save's parent is **deleted**, the child becomes an **orphan** and moves to the top level of the tree (its detail pane marks the lost parent as *detached*). **Renaming** a save keeps its children attached: they are re-pointed at the new name and re-signed, so they don't load flagged as modified. Children that are already flagged *modified* are left alone, so a rename never clears a tampered save's badge.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move the highlight between saves |
| `Enter` | Load the highlighted save |
| `d` | Delete the highlighted save (asks you to confirm first) |
| `r` | Rename the highlighted save |
| `c` | Duplicate the highlighted save |
| `m` | Mark the highlighted save as your main game, or unmark it |
| `Esc` | Return to where you opened it from: the main menu, or your current run if opened mid-game |

---

## Save badges

Saves can carry a tag in the list, explained on screen in a bordered **Key** box:

| Tag | Meaning |
|---|---|
| ★ auto | The automatic save slot |
| ● active | The save your game is autosaving into |
| ◆ main game | The save you marked as your main game: Continue on the main menu opens it |
| ▸ continue | With no main game marked, the save Continue opens (the game you played last) |
| ⚠ modified | The save file was edited outside the game (integrity check failed) |
| ⚠ corrupt | The file could not be read. It is still listed but dimmed, and cannot be loaded |

---

## Save integrity

Saves are signed. If a save file is edited outside the game, the integrity check fails and the browser flags the file `⚠ modified`. A `⚠ corrupt` file is one the game couldn't read at all; it stays in the list, dimmed, but cannot be loaded.

---

## Deleting saves

To delete one save, highlight it in the Load Game browser and press `d`; the game asks you to confirm first.

To delete them all, open **Accounts** on the main menu, highlight the account and press `x` (**Delete saves**). After you confirm, it deletes **every save of that account**: the autosave, your named saves and every branch. The runs in them, and the prestige they carried, are gone. **No backup is made, and it cannot be undone.** (It used to be an entry on the main menu itself, next to Quit.)

The account itself is kept: its name, theme unlocks, lifetime stats and badges stay, and so do any backups already in `data/backups/`. Other accounts' saves are not touched. To keep a copy, back the account up first with `account backup` (or `b` in the Accounts panel). An account left with no save sees the main menu's first page again, with New game first.

Compare **wiping an account** (`w` in the Accounts panel). A wipe deletes the account itself, identity and unlocks included, along with every save in its slot, but it backs the whole slot up to `data/backups/` first. See [Wiping an account](account.md#wiping-an-account).

| | Delete saves | Wipe an account |
|---|---|---|
| Where | Accounts panel (`x`), behind a confirm | Accounts panel (`w`), behind a type-the-name confirm |
| Saves | Every save of the highlighted account | Every save of that account |
| Account name, unlocks, lifetime stats, badges | Kept | Deleted |
| Backup first | No | Yes, to `data/backups/` |

---

## Tips

- **Branch at milestones.** `save iron_age_start` makes a new save at that moment and moves autosave onto it. The old run stays where it was, and autosave can't overwrite it.
- **Duplicate before risky moves.** Press `c` in the Load Game browser to copy a save before a prestige, a catastrophe, or any decision you might want to undo. Branching does much the same, but you keep playing on the new copy instead of the old one.
- **Your active save is overwritten all the time.** Autosave keeps it matching your current game, so it is not a snapshot. To keep a run exactly as it is now, duplicate it (`c`) or save it under a new name.
