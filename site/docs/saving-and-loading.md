# Saving & loading

AgeForge keeps your civilization in named saves, with a regular autosave and a save browser. This page covers where saves live, the commands that manage them, which game the main menu's Continue opens, how to read the Load Game browser, how to delete saves and what happens when you load a save from an older version.

---

## Starting a new game

When you choose **New game** on the [main menu](how-to-play.md#the-main-menu), the game asks you to name your civilization and fills in a random name. Press **Enter** to accept it, type your own, or press **Tab** for a new suggestion. **Esc** cancels and returns to the menu. That name becomes your active save, the game writes its first save under it at once, and it autosaves into it from then on. A name that an existing save already has replaces that save.

---

## Where saves live

Save files go in **your active account's** saves folder, `data/accounts/<account_id>/saves/*.json`, inside the `data/` folder next to the `ageforge` binary. There is one file per save. The `save` and `load` commands and the **Load Game** browser all read and write the same files, so a save made one way shows up the other.

Saves are **per-account**: each account keeps its own `saves/` folder, so switching accounts changes which saves you see, and one account's saves never mix with another's. See [Account & Recovery](account.md) for how accounts and their slots are laid out.

A save name can be up to 64 characters. It can't be empty, contain `/` or `\`, contain `..` or start with a dot.

---

## Commands

| Command | Description |
|---|---|
| `save` | Opens an **Overwrite / Branch new / Cancel** prompt for your current run (see below) |
| `save <name>` | **Branches** a new save with that name off your current run; autosave then follows it |
| `load` | Opens the **Load Game** browser (your save tree) to pick which save or branch to load |
| `load <name>` | Loads that named save directly |
| `saves` (or `save list`) | List all save files, each with its save time and age |
| `Esc` | Closes the open panel. With no panel open, saves to your active save, stops the game and returns to the main menu |

```
save           # prompt: Overwrite this run, or Branch a new save?
save hero      # branch a new save named "hero" off the current run
load           # open the Load Game browser to pick a save or branch
load hero      # load the save named "hero" directly
saves
```

A name typed after `save` or `load` is one word (`save my_run`). A name with spaces, like the suggested ones, works in the prompts and the browser.

A bare `load` (no name) **opens the Load Game browser**, the tree of every save, so you choose which save or branch to load. You can open it mid-game, and pressing `Esc` in the browser returns you to your current run without loading anything. `load <name>` skips the browser and loads that save directly. Loading replaces the game in play and does not save it first, so branch or save before you load if you want to keep where you are.

---

## Active save slot

The game remembers the last slot you saved to or loaded. That is your **active** save.

- Until you've named a save or loaded one this session, the active slot is `autosave`.
- Once you start a new game, branch to `hero` or load `hero`, the active slot is `hero` and autosave follows it.

The periodic autosave and the save on `Esc` both write to your **active** save, overwriting it each time. Your current game is that save. There is no separate hidden slot.

---

## Branching your save

A bare `save` opens a prompt, `Save "<active save>"?`, with three buttons:

- **Overwrite** writes your current run to the active slot now (what autosave does, on demand).
- **Branch new** starts a new save. You get a generated name (edit it, press **Tab** for another, **Enter** to confirm, **Esc** to cancel). The new save's *parent* is the save you branched from, and **autosave switches to the new branch**, so the old save stays exactly as it was at the branch point.
- **Cancel** closes the prompt and saves nothing.

`save <name>` skips the prompt and branches straight to that name.

Branching keeps a moment without stopping play: branch before a prestige, a risky catastrophe, or any decision you might want to revisit. Your old save stays as it was while you keep playing on the new branch. Branched saves are ordinary files and appear in the `saves` list and the Load Game browser alongside everything else. Branching to a name that's already taken is refused.

---

## Autosave

The game autosaves every 60 seconds to your **active** save, and saves again when `Esc` takes you back to the main menu, so the file on disk is never more than about a minute behind your game. Leaving the program from inside a game, with `quit` or with Ctrl+C, saves that game to its own save first. Leaving from the main menu, or from any page it opens, writes nothing: there is no game in play to save. To keep a point you don't want overwritten, save it under a new name (or duplicate it with `c` in the Load Game browser).

---

## Continue and your main game

**Continue** on the main menu opens your **current game**, and the menu draws that game's town behind it. Which save that is follows three rules, in order:

1. **Your main game**, if you marked one. Select a save in the Load Game browser and press `m`. It stays the current game, whatever else you play, until you press `m` on it again or delete it.
2. Otherwise, **the game you played last**: the save you last loaded or saved on this account.
3. If neither is on record, **the save written most recently**. This is what an account that has never recorded a game gets until you play once, and what you get if the save on record is gone.

A save that cannot be read is never the current game. An account with no save has no Continue.

The Load Game browser shows which save Continue opens: its subtitle names it, its row carries `◆ main game` or `▸ continue`, and its detail pane says why. Renaming a save keeps it your main game. Deleting it clears the mark.

Both records are kept in the account's `settings.json`, beside your display settings (see [The settings file](account.md#the-settings-file)). They are part of the account, not of any save, so each account has its own current game.

---

## The Load Game browser

Choosing **Load Game** from the main menu, or typing a bare `load` mid-game, opens a browser that lists every save belonging to your **active account** (under `data/accounts/<id>/saves/`). The subtitle gives the number of saves and names the one Continue opens. Selecting a save fills a **detail pane** with what you need to judge it before loading: its civilization title in quotes (once you have one), age and epoch, population, buildings, wonders, milestones, techs, soldiers, prestige level and points, [morale](morale.md), which account made it (*this account* / *another account* / *pre-account*), a ⚠ warning if a catastrophe is pending, the exact save time, and for branched saves a **Branched from** line naming its parent.

**The save tree.** Saves are arranged as a **tree**, not a flat list. When you [branch](#branching-your-save) a new save off your current run, it appears **indented beneath its parent** with tree connectors (`├─`, `└─`), so you can see which saves descend from which. Top-level saves (ones you started fresh, plus any orphans) are ordered most recent first, and so are each parent's children.

A **● active** marker shows which save your game is autosaving into, and **◆ main game** or **▸ continue** marks the save the main menu's Continue opens (see [Continue and your main game](#continue-and-your-main-game)).

If a save's parent is **deleted**, the child becomes an **orphan** and moves to the top level of the tree (its detail pane marks the lost parent as *detached*). **Renaming** a save keeps its children attached: they are re-pointed at the new name and re-signed, so they don't load flagged as modified. A child edited outside the game (its signature no longer checks) is left alone, so a rename never clears a modified save's mark.

**Keys inside the browser:**

| Key | Action |
|---|---|
| `↑` / `↓` | Move between saves |
| `Enter` | Load the selected save |
| `d` | Delete the selected save (asks you to confirm first) |
| `r` | Rename the selected save |
| `c` | Duplicate the selected save, as `<name>-copy` (then `-copy-2`, `-copy-3` and so on) |
| `m` | Mark the selected save as your main game, or unmark it |
| `Esc` / `q` | Return to where you opened it from: the main menu, or your current run if opened mid-game |

A corrupt save can't be loaded or made your main game. Loading one shows a message in the detail pane and stays in the browser.

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

Saves are signed. When you load a save that was edited outside the game, the integrity check fails and the game is marked modified. The saves you make from it carry the mark, and the browser shows it as `⚠ modified`. A `⚠ corrupt` file is one the game couldn't read at all. It stays in the list, dimmed, but cannot be loaded.

---

## Deleting saves

To delete one save, select it in the Load Game browser and press `d`. The game asks you to confirm first.

To delete them all, open **Accounts** on the main menu, select the account and press `x` (**Delete saves**). After you confirm, it deletes **every save of that account**: the autosave, your named saves and every branch. The runs in them, and the prestige they carried, are gone. **No backup is made, and it cannot be undone.**

The account itself is kept: its name, theme unlocks, lifetime stats and badges stay, and so do any backups already in `data/backups/`. Other accounts' saves are not touched. To keep a copy, back the account up first with `account backup` (or `b` in the Accounts panel). An account left with no save sees the main menu's first page again, with New game first.

Compare **wiping an account** (`w` in the Accounts panel). A wipe deletes the account itself, identity and unlocks included, along with every save in its slot, but it backs the whole slot up to `data/backups/` first. See [Wiping an account](account.md#wiping-an-account).

| | Delete saves | Wipe an account |
|---|---|---|
| Where | Accounts panel (`x`), behind a confirm | Accounts panel (`w`), behind a type-the-name confirm |
| Saves | Every save of the selected account | Every save of that account |
| Account name, unlocks, lifetime stats, badges | Kept | Deleted |
| Backup first | No | Yes, to `data/backups/` |

---

## Tips

- **Branch at milestones.** `save iron_age_start` makes a new save at that moment and moves autosave onto it. The old run stays where it was, and autosave can't overwrite it.
- **Duplicate before risky moves.** Press `c` in the Load Game browser to copy a save before a prestige, a catastrophe, or any decision you might want to undo. Branching does much the same, but you keep playing on the new copy instead of the old one.
- **Your active save is overwritten all the time.** Autosave keeps it matching your current game, so it is not a snapshot. To keep a run exactly as it is now, duplicate it (`c`) or save it under a new name.

---

## Saves from older versions

An older save loads as usual. For each change below except the folder move, the log says so once on the first load of a save from before it, and the next save you make doesn't repeat it.

### Saves from the flat save folder

Older builds kept every save in a single flat `data/saves/` folder. When the game starts and finds that layout, it moves the saves into your account's `data/accounts/<id>/saves/` folder. Nothing is deleted, and you don't have to do anything.

### Saves from before the research changes

A game saved before wonders needed a keystone tech, or before some commands waited for a tech, keeps everything it researched and built. In the age it was in nothing new is locked, and a command it was already using (trade routes, festivals, the black market or diplomacy) stays open for the rest of that run. If the time away already carried the game into the next age, that grace is over and the log says only what changed. Techs give the bonuses of the current tree: smaller, and always counted in full (see [How Tech Bonuses Stack](technologies.md#how-tech-bonuses-stack)).

### Saves from before the pacing update

A game saved before the pacing update carries on where you left it. The log says once that the ages ahead are longer and that events, raids, trade routes and other timers stretch with them.

### Saves from before the storage wall

A game saved before storage became a wall keeps every building copy and all of its stock, and stock has no limit. What your plan had banked from overflow is given back to you. The log says so:

```
Storage changed. A storage building's price now climbs much faster with each copy, and stock has no ceiling: nothing you make is wasted. You keep every building you have, and what your plan had banked was given back to you.
```

### Saves from before scholars

A game saved before scholars keeps every building copy, even above a knowledge building's cap. Knowledge was a stock before, and it is not any more:

- Knowledge you held becomes a gift of work for the tech in progress, at most half of that tech's price. With nothing in progress it is lost.
- A tech that was paid for and waiting out its research time completes on load.
- Workers on knowledge buildings return to the idle pool, and the knowledge buildings come with their scholars.
- A share you had set for the knowledge domain is ignored.

The log says:

```
Knowledge is no longer stored. Your scholars now work on one technology at a time.
```

### Saves from before the market ended

A game saved while the market could swap resources keeps everything it built, researched and holds. If its build plan held sell orders, they are removed. The log says:

```
The market no longer swaps resources. The sell orders in your plan were removed.
```
